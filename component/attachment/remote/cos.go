package remote

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSConfig 定义了腾讯云COS的配置参数。
type COSConfig jcbaseGo.COSStruct

// COSClient 实现了腾讯云COS存储的客户端。
// 注意：COSClient是并发安全的，因为底层COS SDK客户端本身是并发安全的。
type COSClient struct {
	client *cos.Client
}

// NewCOSClient 创建一个新的COS客户端。
func NewCOSClient(config COSConfig) (*COSClient, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, &Error{Op: "ParseBucketURL", Err: err}
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretId,
			SecretKey: config.SecretKey,
		},
	})

	return &COSClient{client: client}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *COSClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	_, err := c.client.Object.Put(ctx, remotePath, bytes.NewReader(data), nil)
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}

// Download 实现了Client接口的Download方法。
func (c *COSClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
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
	_, err := c.client.Object.Delete(ctx, remotePath)
	if err != nil {
		return &Error{Op: "Delete", Err: err}
	}
	return nil
}

// List 实现了Client接口的List方法。
func (c *COSClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
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
// 使用示例：
//
//	opt, headers := buildCOSPresignHeader(&PresignOptions{ContentType: "image/jpeg"})
func buildCOSPresignHeader(opts *PresignOptions) (*cos.PresignedURLOptions, map[string]string) {
	headers := make(map[string]string)
	signHeader := make(http.Header)
	if opts == nil {
		return &cos.PresignedURLOptions{Header: &signHeader}, headers
	}

	if opts.ContentType != "" {
		signHeader.Set("Content-Type", opts.ContentType)
		headers["Content-Type"] = opts.ContentType
	}

	for key, value := range opts.Metadata {
		signHeader.Set(key, value)
		headers[key] = value
	}

	return &cos.PresignedURLOptions{Header: &signHeader}, headers
}
