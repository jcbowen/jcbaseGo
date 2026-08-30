package remote

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

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

// PresignUpload 生成用于上传指定对象的预签名 URL。
// - remotePath: 对象在存储桶中的路径（Key）。
// - opts: 预签名选项，nil 时表示使用默认选项（仅签名 bucket 与 key，默认有效期 10 分钟）。
// 返回值：
//   - string: 预签名 URL。
//   - map[string]string: 需要在上传请求中携带的已签名请求头。
//   - error: 生成过程中出现的错误。
func (c *OSSClient) PresignUpload(ctx context.Context, remotePath string, opts *PresignOptions) (string, map[string]string, error) {
	select {
	case <-ctx.Done():
		return "", nil, &Error{Op: "PresignUpload", Err: ctx.Err()}
	default:
	}

	expires := 10 * time.Minute
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
