package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSConfig 定义了腾讯云COS的配置参数。
type COSConfig jcbaseGo.COSStruct

// cosSignHeaders 是 COS 签名算法会主动参与签名的标准请求头集合（小写）。
// 不在此集合且不以 x-cos- 开头的自定义元数据，需自动添加 x-cos-meta- 前缀。
var cosSignHeaders = map[string]struct{}{
	"host":                           {},
	"range":                          {},
	"x-cos-acl":                      {},
	"x-cos-grant-read":               {},
	"x-cos-grant-write":              {},
	"x-cos-grant-full-control":       {},
	"cache-control":                  {},
	"content-disposition":            {},
	"content-encoding":               {},
	"content-type":                   {},
	"content-length":                 {},
	"content-md5":                    {},
	"transfer-encoding":              {},
	"expect":                         {},
	"expires":                        {},
	"x-cos-content-sha1":             {},
	"x-cos-storage-class":            {},
	"if-match":                       {},
	"if-modified-since":              {},
	"if-none-match":                  {},
	"if-unmodified-since":            {},
	"origin":                         {},
	"access-control-request-method":  {},
	"access-control-request-headers": {},
	"x-cos-object-type":              {},
	"pic-operations":                 {},
}

// COSClient 实现了腾讯云COS存储的客户端。
// 注意：COSClient是并发安全的，因为底层COS SDK客户端本身是并发安全的。
type COSClient struct {
	client *cos.Client
}

// NewCOSClient 创建一个新的COS客户端。
//
// 参数：
//   - config COSConfig: COS 配置，Url 为空时会使用 Bucket 与 Region 自动构造
//
// 返回值：
//   - *COSClient: 初始化后的 COS 客户端
//   - error: 配置错误时返回错误
//
// 使用示例：
//
//	client, err := NewCOSClient(COSConfig{SecretId: "xxx", SecretKey: "xxx", Bucket: "bucket-1250000000", Region: "ap-guangzhou"})
func NewCOSClient(config COSConfig) (*COSClient, error) {
	bucketURL := config.Url
	if bucketURL == "" {
		if config.Bucket == "" || config.Region == "" {
			return nil, &Error{Op: "NewCOSClient", Err: errors.New("cos url 为空时，bucket 与 region 必须同时填写")}
		}
		bucketURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, config.Region)
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, &Error{Op: "ParseBucketURL", Err: err}
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     config.SecretId,
			SecretKey:    config.SecretKey,
			SessionToken: config.Token,
		},
	})

	return &COSClient{client: client}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *COSClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	select {
	case <-ctx.Done():
		return &Error{Op: "Upload", Err: ctx.Err()}
	default:
	}

	_, err := c.client.Object.Put(ctx, remotePath, bytes.NewReader(data), nil)
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}

// Download 实现了Client接口的Download方法。
func (c *COSClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, &Error{Op: "Download", Err: ctx.Err()}
	default:
	}

	resp, err := c.client.Object.Get(ctx, remotePath, nil)
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
func (c *COSClient) Delete(ctx context.Context, remotePath string) error {
	select {
	case <-ctx.Done():
		return &Error{Op: "Delete", Err: ctx.Err()}
	default:
	}

	_, err := c.client.Object.Delete(ctx, remotePath)
	if err != nil {
		return &Error{Op: "Delete", Err: err}
	}
	return nil
}

// List 实现了Client接口的List方法。
func (c *COSClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	select {
	case <-ctx.Done():
		return ListResult{}, &Error{Op: "List", Err: ctx.Err()}
	default:
	}

	opt := &cos.BucketGetOptions{
		Prefix:  options.Prefix,
		Marker:  options.Marker,
		MaxKeys: options.MaxKeys,
	}
	result, _, err := c.client.Bucket.Get(ctx, opt)
	if err != nil {
		return ListResult{}, &Error{Op: "List", Err: err}
	}

	var files []FileInfo // 使用 var 声明空切片
	for _, content := range result.Contents {
		modTime, err := time.Parse(time.RFC3339, content.LastModified)
		if err != nil {
			return ListResult{}, &Error{Op: "List", Err: err}
		}

		files = append(files, FileInfo{
			Name:    content.Key,
			Size:    content.Size,
			ModTime: modTime,
			IsDir:   false,
		})
	}

	return ListResult{
		Files:       files,
		NextMarker:  result.NextMarker,
		IsTruncated: result.IsTruncated,
	}, nil
}

// Close 实现了Client接口的Close方法。
func (c *COSClient) Close() error {
	// COS客户端不需要显式关闭连接
	return nil
}

// Exists 判断指定对象是否已存在于COS中。
// 底层使用 HEAD Object，只取对象元信息不下载内容，开销较低。
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
func (c *COSClient) Exists(ctx context.Context, remotePath string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, &Error{Op: "Exists", Err: ctx.Err()}
	default:
	}

	_, err := c.client.Object.Head(ctx, remotePath, nil)
	if err == nil {
		return true, nil
	}

	// 404 表示对象确实不存在，属于正常的判定结果而非错误
	var respErr *cos.ErrorResponse
	if errors.As(err, &respErr) && respErr.Response != nil && respErr.Response.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, &Error{Op: "Exists", Err: err}
}

// PresignUpload 生成用于上传指定对象的预签名 URL，客户端需使用 PUT 方法上传文件内容。
// 参数：
//   - ctx context.Context: 上下文
//   - remotePath string: 对象在存储桶中的路径（Key）
//   - opts *PresignOptions: 预签名选项，nil 时表示使用默认选项（默认有效期 10 分钟，不签名额外请求头）
//
// 返回值：
//   - string: 预签名 URL
//   - map[string]string: 需要在上传请求中原样携带的已签名请求头，恒为非 nil
//   - error: 生成过程中出现的错误
//
// 说明：
//   - COS 仅对 Content-Type 等特定请求头以及 x-cos- 前缀的请求头参与签名，
//     因此自定义元数据需使用 x-cos-meta- 前缀才会真正生效，其他 key 不会参与签名。
//   - 签名默认签入 Host，客户端必须使用返回的域名访问，不可改写。
//
// 使用示例：
//
//	url, headers, err := client.PresignUpload(ctx, "images/2026/08/xxx.png", &PresignOptions{
//	    Expires:     10 * time.Minute,
//	    ContentType: "image/jpeg",
//	    Metadata:    map[string]string{"x-cos-meta-uid": "12345"},
//	})
func (c *COSClient) PresignUpload(ctx context.Context, remotePath string, opts *PresignOptions) (string, map[string]string, error) {
	select {
	case <-ctx.Done():
		return "", nil, &Error{Op: "PresignUpload", Err: ctx.Err()}
	default:
	}

	expires := defaultPresignExpires
	if opts != nil && opts.Expires > 0 {
		expires = opts.Expires
	}

	presignOpt, signedHeaders := buildCOSPresignHeader(opts)
	presignedURL, err := c.client.Object.GetPresignedURL2(ctx, http.MethodPut, remotePath, expires, presignOpt)
	if err != nil {
		return "", nil, &Error{Op: "PresignUpload", Err: err}
	}

	return presignedURL.String(), signedHeaders, nil
}

// buildCOSPresignHeader 将预签名选项转换为参与签名的请求头。
// 参数：
//   - opts *PresignOptions: 预签名选项，nil 时不签名任何额外请求头
//
// 返回值：
//   - *cos.PresignedURLOptions: 供 COS SDK 使用的预签名选项，恒为非 nil
//   - map[string]string: 需要在上传请求中原样携带的已签名请求头，恒为非 nil
//
// 说明：
//   - 自定义元数据若未使用 x-cos-meta- 前缀，会自动补齐该前缀，确保参与签名。
//   - 返回的 header key 已规范化，调用方可直接设置到 HTTP 请求中。
//
// 使用示例：
//
//	opt, headers := buildCOSPresignHeader(&PresignOptions{ContentType: "image/jpeg"})
func buildCOSPresignHeader(opts *PresignOptions) (*cos.PresignedURLOptions, map[string]string) {
	signHeader := make(http.Header)
	headers := make(map[string]string)
	if opts == nil {
		return &cos.PresignedURLOptions{Header: &signHeader}, headers
	}

	if opts.ContentType != "" {
		key := textproto.CanonicalMIMEHeaderKey("Content-Type")
		signHeader.Set(key, opts.ContentType)
		headers[key] = opts.ContentType
	}

	for key, value := range opts.Metadata {
		signKey := cosSignHeaderKey(key)
		if signKey == "" {
			continue
		}
		signHeader.Set(signKey, value)
		headers[signKey] = value
	}

	return &cos.PresignedURLOptions{Header: &signHeader}, headers
}

// cosSignHeaderKey 返回参与 COS 签名的 header key。
// 不在标准签名头集合中且不以 x-cos- 开头的 key，会自动添加 x-cos-meta- 前缀。
//
// 参数：
//   - key string: 原始 header key
//
// 返回值：
//   - string: 处理后的 header key；空字符串表示输入非法
func cosSignHeaderKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	lower := strings.ToLower(key)
	if _, ok := cosSignHeaders[lower]; ok {
		return textproto.CanonicalMIMEHeaderKey(key)
	}
	if strings.HasPrefix(lower, "x-cos-") {
		return textproto.CanonicalMIMEHeaderKey(key)
	}
	return textproto.CanonicalMIMEHeaderKey("x-cos-meta-" + key)
}
