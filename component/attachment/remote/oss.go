package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/jcbowen/jcbaseGo"
)

// OSSConfig 定义了阿里云OSS的配置参数。
type OSSConfig jcbaseGo.OSSStruct

// OSSClient 实现了阿里云OSS存储的客户端。
// 注意：OSSClient是并发安全的，因为底层OSS SDK客户端本身是并发安全的。
type OSSClient struct {
	client *oss.Client
	bucket string
}

// regionFromEndpoint 从 OSS 访问域名推导地域 ID。
// V4 签名的 credential 需要地域 ID（如 cn-shanghai），
// 而配置中的 Endpoint 形如 oss-cn-shanghai.aliyuncs.com，需去掉 oss- 前缀与 -internal 后缀。
// 参数：
//   - endpoint string: OSS 访问域名，允许带协议头与端口
//
// 返回值：
//   - string: 地域 ID；域名不符合预期格式时返回原主机名的第一段
//
// 使用示例：
//
//	regionFromEndpoint("oss-cn-shanghai.aliyuncs.com")          // "cn-shanghai"
//	regionFromEndpoint("oss-cn-shanghai-internal.aliyuncs.com") // "cn-shanghai"
func regionFromEndpoint(endpoint string) string {
	host := endpoint

	// 去掉协议头
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	// 去掉端口
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	// 取主机名的第一段
	if i := strings.Index(host, "."); i > 0 {
		host = host[:i]
	}

	return strings.TrimSuffix(strings.TrimPrefix(host, "oss-"), "-internal")
}

// NewOSSClient 创建一个新的OSS客户端。
func NewOSSClient(config OSSConfig) (*OSSClient, error) {
	// 创建凭证提供器
	cred := credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret)

	// 创建客户端配置
	// Region 必填：V4 签名的 credential 需要地域ID，缺失会导致生成的预签名URL被OSS拒绝
	region := regionFromEndpoint(config.Endpoint)
	clientConfig := &oss.Config{
		Endpoint:            &config.Endpoint,
		Region:              &region,
		CredentialsProvider: cred,
	}

	// 创建客户端
	client := oss.NewClient(clientConfig)

	return &OSSClient{client: client, bucket: config.BucketName}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *OSSClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	select {
	case <-ctx.Done():
		return &Error{Op: "Upload", Err: ctx.Err()}
	default:
	}

	reader := bytes.NewReader(data)
	_, err := c.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
		Body:   reader,
	})
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}

// Download 实现了Client接口的Download方法。
func (c *OSSClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, &Error{Op: "Download", Err: ctx.Err()}
	default:
	}

	resp, err := c.client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	})
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	return data, nil
}

// Delete 实现了Client接口的Delete方法。
func (c *OSSClient) Delete(ctx context.Context, remotePath string) error {
	select {
	case <-ctx.Done():
		return &Error{Op: "Delete", Err: ctx.Err()}
	default:
	}

	_, err := c.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	})
	if err != nil {
		return &Error{Op: "Delete", Err: err}
	}
	return nil
}

// List 实现了Client接口的List方法。
func (c *OSSClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	select {
	case <-ctx.Done():
		return ListResult{}, &Error{Op: "List", Err: ctx.Err()}
	default:
	}

	var files []FileInfo
	continuationToken := options.Marker
	prefix := options.Prefix
	maxKeys := int32(options.MaxKeys)

	// 构建ListObjectsV2请求
	req := &oss.ListObjectsV2Request{
		Bucket:            &c.bucket,
		Prefix:            &prefix,
		ContinuationToken: &continuationToken,
	}
	if maxKeys > 0 {
		req.MaxKeys = maxKeys
	}

	// 调用ListObjectsV2 API
	lsRes, err := c.client.ListObjectsV2(ctx, req)
	if err != nil {
		return ListResult{}, &Error{Op: "List", Err: err}
	}

	// 处理返回的对象列表
	for _, object := range lsRes.Contents {
		files = append(files, FileInfo{
			Name:    *object.Key,
			Size:    object.Size,
			ModTime: *object.LastModified,
			IsDir:   false,
		})
	}

	isTruncated := lsRes.IsTruncated
	nextMarker := ""
	if lsRes.NextContinuationToken != nil {
		nextMarker = *lsRes.NextContinuationToken
	}

	return ListResult{
		Files:       files,
		NextMarker:  nextMarker,
		IsTruncated: isTruncated,
	}, nil
}

// Close 实现了Client接口的Close方法。
func (c *OSSClient) Close() error {
	// OSS客户端不需要显式关闭连接
	return nil
}

// Exists 判断指定对象是否已存在于OSS中。
// 底层使用 HeadObject，只取对象元信息不下载内容，开销较低。
// 参数：
//   - ctx context.Context: 上下文，可用于控制探测超时
//   - remotePath string: 对象在存储桶中的路径（Key）
//
// 返回值：
//   - bool: 对象存在返回 true；对象不存在返回 false
//   - error: 网络异常或权限不足时返回错误
//
// 使用示例：
//
//	ok, err := client.Exists(ctx, "images/2026/08/xxx.png")
func (c *OSSClient) Exists(ctx context.Context, remotePath string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, &Error{Op: "Exists", Err: ctx.Err()}
	default:
	}

	_, err := c.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	})
	if err == nil {
		return true, nil
	}

	// 404 表示对象确实不存在，属于正常的判定结果而非错误
	var svcErr *oss.ServiceError
	if errors.As(err, &svcErr) && svcErr.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, &Error{Op: "Exists", Err: err}
}

// PresignUpload 生成用于上传指定对象的预签名 URL（OSS 为 V4 签名 OSS4-HMAC-SHA256）。
// - remotePath: 对象在存储桶中的路径（Key）。
// - opts: 预签名选项，nil 时表示使用默认选项（仅签名 bucket 与 key，默认有效期 10 分钟）。
// 返回值：
//   - string: 预签名 URL。
//   - map[string]string: 需要在上传请求中携带的已签名请求头。
//   - error: 生成过程中出现的错误。
//
// 签名契约（浏览器直传必读）：
//
//  1. 客户端发起 PUT 时必须原样携带返回的 SignedHeaders（键名与值都不能改），
//     且不得额外出现任何被签名的头。被签名的头缺失或值不一致，OSS 会返回 403
//     SignatureDoesNotMatch；未被签名的头（如 Accept、Origin）不影响校验。
//
//  2. opts.ContentType 为空时 Content-Type 不参与签名，此时若客户端是浏览器 fetch
//     且 body 为带 type 的 Blob/File，浏览器会自动补发一个未签名的 Content-Type 头，
//     同样触发 403 SignatureDoesNotMatch。规避方式二选一：
//     - 签发时显式传入 ContentType 并将返回的 SignedHeaders 交由前端原样携带
//     （fetch/XHR 显式设置该头后浏览器不再自动补）；
//     - 或保证上传 body 为无 type 的 Blob（浏览器不自动补头）。
//
// 使用示例：
//
//	url, headers, err := client.PresignUpload(ctx, "images/2026/08/xxx.png",
//		&PresignOptions{Expires: 10 * time.Minute, ContentType: "image/jpeg"})
//	// 前端：fetch(url, { method: "PUT", headers, body: file }) —— headers 原样透传
func (c *OSSClient) PresignUpload(ctx context.Context, remotePath string, opts *PresignOptions) (string, map[string]string, error) {
	select {
	case <-ctx.Done():
		return "", nil, &Error{Op: "PresignUpload", Err: ctx.Err()}
	default:
	}

	expires := defaultPresignExpires
	req := &oss.PutObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	}

	if opts != nil {
		if opts.Expires > 0 {
			expires = opts.Expires
		}
		if opts.ContentType != "" {
			req.ContentType = &opts.ContentType
		}
		if len(opts.Metadata) > 0 {
			req.Metadata = opts.Metadata
		}
	}

	result, err := c.client.Presign(ctx, req, oss.PresignExpires(expires))
	if err != nil {
		return "", nil, &Error{Op: "PresignUpload", Err: err}
	}

	headers := result.SignedHeaders
	if headers == nil {
		headers = make(map[string]string)
	}

	return result.URL, headers, nil
}

// PresignGet 生成用于下载指定对象的预签名 URL。
// 底层使用 OSS Go SDK V2 的 GetObject 预签名能力，签名版本为 V4（OSS4-HMAC-SHA256）。
// - remotePath: 对象在存储桶中的路径（Key）。
// - opts: 下载选项，nil 时表示使用默认选项（仅签名 bucket 与 key，默认有效期 10 分钟）。
//
// 返回值：
//   - string: 预签名 URL。
//   - map[string]string: 使用该 URL 发起请求时需要携带的已签名请求头。
//   - error: 生成过程中出现的错误。
//
// 签名契约（与 PresignUpload 相同）：请求必须原样携带返回的 SignedHeaders，
// 被签名的头缺失或值不一致时服务端返回 403 SignatureDoesNotMatch。
//
// 使用示例：
//
//	url, headers, err := client.PresignGet(ctx, "releases/pkg.zip", &PresignDownloadOptions{
//		Expires:                    60 * time.Minute,
//		ResponseContentDisposition: "attachment;filename=pkg.zip",
//	})
func (c *OSSClient) PresignGet(ctx context.Context, remotePath string, opts *PresignDownloadOptions) (string, map[string]string, error) {
	select {
	case <-ctx.Done():
		return "", nil, &Error{Op: "PresignGet", Err: ctx.Err()}
	default:
	}

	expires := defaultPresignExpires
	req := &oss.GetObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	}

	if opts != nil {
		if opts.Expires > 0 {
			expires = opts.Expires
		}
		if opts.ResponseContentDisposition != "" {
			req.ResponseContentDisposition = &opts.ResponseContentDisposition
		}
		if opts.ResponseContentType != "" {
			req.ResponseContentType = &opts.ResponseContentType
		}
		if opts.ResponseCacheControl != "" {
			req.ResponseCacheControl = &opts.ResponseCacheControl
		}
	}

	result, err := c.client.Presign(ctx, req, oss.PresignExpires(expires))
	if err != nil {
		return "", nil, &Error{Op: "PresignGet", Err: err}
	}

	headers := result.SignedHeaders
	if headers == nil {
		headers = make(map[string]string)
	}

	return result.URL, headers, nil
}

// HeadObjectMeta 读取指定对象的元信息，不下载对象内容。
// 底层使用 HeadObject，元数据中的自定义字段（如 x-oss-meta-sha256）会去除前缀后放入 Metadata。
// - remotePath: 对象在存储桶中的路径（Key）。
//
// 返回值：
//   - *ObjectMeta: 对象元信息；Metadata 一定非 nil。
//   - error: 对象不存在时返回包装了 ErrObjectNotFound 的错误；其它异常返回请求错误。
//
// 使用示例：
//
//	meta, err := client.HeadObjectMeta(ctx, "releases/pkg.zip")
//	if errors.Is(err, ErrObjectNotFound) { /* 对象不存在 */ }
//	sha256 := meta.Metadata["sha256"]
func (c *OSSClient) HeadObjectMeta(ctx context.Context, remotePath string) (*ObjectMeta, error) {
	select {
	case <-ctx.Done():
		return nil, &Error{Op: "HeadObjectMeta", Err: ctx.Err()}
	default:
	}

	resp, err := c.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	})
	if err != nil {
		if isObjectNotFound(err) {
			return nil, &Error{Op: "HeadObjectMeta", Err: fmt.Errorf("%w: %s", ErrObjectNotFound, remotePath)}
		}
		return nil, &Error{Op: "HeadObjectMeta", Err: err}
	}

	meta := &ObjectMeta{
		Size:     resp.ContentLength,
		Metadata: resp.Metadata,
	}
	if resp.ContentType != nil {
		meta.ContentType = *resp.ContentType
	}
	if resp.LastModified != nil {
		meta.LastModified = *resp.LastModified
	}
	if resp.ETag != nil {
		meta.ETag = *resp.ETag
	}
	if meta.Metadata == nil {
		meta.Metadata = make(map[string]string)
	}

	return meta, nil
}

// OpenDownload 打开指定对象的下载流，不在内存中缓存完整对象。
// 底层使用 GetObject，返回的 io.ReadCloser 由调用方负责关闭。
// - remotePath: 对象在存储桶中的路径（Key）。
// - opts: 下载选项，nil 时表示下载完整对象；设置 Range 时只读取指定范围。
//
// 返回值：
//   - io.ReadCloser: 流式读取器；读取阶段的错误会在 Read 时返回。
//   - error: 建立下载流失败时返回错误；对象不存在时返回包装了 ErrObjectNotFound 的错误。
//
// 使用示例：
//
//	reader, err := client.OpenDownload(ctx, "releases/pkg.zip", nil)
//	if err != nil { return err }
//	defer reader.Close()
//	_, _ = io.Copy(io.MultiWriter(file, hasher), reader)
func (c *OSSClient) OpenDownload(ctx context.Context, remotePath string, opts *OpenDownloadOptions) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, &Error{Op: "OpenDownload", Err: ctx.Err()}
	default:
	}

	req := &oss.GetObjectRequest{
		Bucket: &c.bucket,
		Key:    &remotePath,
	}
	if opts != nil && opts.Range != "" {
		req.Range = &opts.Range
	}

	resp, err := c.client.GetObject(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			return nil, &Error{Op: "OpenDownload", Err: fmt.Errorf("%w: %s", ErrObjectNotFound, remotePath)}
		}
		return nil, &Error{Op: "OpenDownload", Err: err}
	}
	if resp.Body == nil {
		return nil, &Error{Op: "OpenDownload", Err: errors.New("OSS 返回空响应体")}
	}

	return resp.Body, nil
}

// isObjectNotFound 判断错误是否为 OSS 返回的 404（对象不存在）。
// 参数：
//   - err error: HeadObject / GetObject 返回的错误
//
// 返回值：
//   - bool: 对象不存在返回 true；权限不足、网络异常等其它错误返回 false
func isObjectNotFound(err error) bool {
	var svcErr *oss.ServiceError
	return errors.As(err, &svcErr) && svcErr.StatusCode == http.StatusNotFound
}
