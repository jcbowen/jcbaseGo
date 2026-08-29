package remote

import (
	"bytes"
	"context"
	"io"
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

// NewOSSClient 创建一个新的OSS客户端。
func NewOSSClient(config OSSConfig) (*OSSClient, error) {
	// 创建凭证提供器
	cred := credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret)

	// 创建客户端配置
	clientConfig := &oss.Config{
		Endpoint:            &config.Endpoint,
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
