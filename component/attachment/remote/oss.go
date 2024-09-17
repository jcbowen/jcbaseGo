package remote

import (
	"bytes"
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"net/http"
	"sync"
)

// OSSConfig 定义了阿里云OSS的配置参数。
type OSSConfig struct {
	Endpoint        string // OSS服务的Endpoint
	AccessKeyID     string // 阿里云访问密钥ID
	AccessKeySecret string // 阿里云访问密钥Secret
	BucketName      string // OSS存储桶名称
}

// OSSClient 实现了阿里云OSS存储的客户端。
// 注意：OSSClient是并发安全的。
type OSSClient struct {
	client *oss.Bucket
	mu     sync.Mutex
}

// NewOSSClient 创建一个新的OSS客户端。
func NewOSSClient(config OSSConfig) (*OSSClient, error) {
	// 创建自定义HTTP客户端，支持context
	httpClient := &http.Client{
		Transport: &http.Transport{},
	}

	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret, oss.HTTPClient(httpClient))
	if err != nil {
		return nil, &Error{Op: "NewOSSClient", Err: err}
	}

	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return nil, &Error{Op: "GetBucket", Err: err}
	}

	return &OSSClient{client: bucket}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *OSSClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return &Error{Op: "Upload", Err: ctx.Err()}
	default:
	}

	reader := bytes.NewReader(data)
	err := c.client.PutObject(remotePath, reader, oss.WithContext(ctx))
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}

// Download 实现了Client接口的Download方法。
func (c *OSSClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, &Error{Op: "Download", Err: ctx.Err()}
	default:
	}

	resp, err := c.client.GetObject(remotePath, oss.WithContext(ctx))
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	defer func() { _ = resp.Close() }()

	data, err := io.ReadAll(resp)
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	return data, nil
}

// Delete 实现了Client接口的Delete方法。
func (c *OSSClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return &Error{Op: "Delete", Err: ctx.Err()}
	default:
	}

	err := c.client.DeleteObject(remotePath, oss.WithContext(ctx))
	if err != nil {
		return &Error{Op: "Delete", Err: err}
	}
	return nil
}

// List 实现了Client接口的List方法。
func (c *OSSClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return ListResult{}, &Error{Op: "List", Err: ctx.Err()}
	default:
	}

	var files []FileInfo
	marker := options.Marker
	prefix := options.Prefix
	maxKeys := options.MaxKeys

	opt := []oss.Option{
		oss.Prefix(prefix),
		oss.Marker(marker),
		oss.WithContext(ctx),
	}
	if maxKeys > 0 {
		opt = append(opt, oss.MaxKeys(maxKeys))
	}

	lsRes, err := c.client.ListObjects(opt...)
	if err != nil {
		return ListResult{}, &Error{Op: "List", Err: err}
	}

	for _, object := range lsRes.Objects {
		files = append(files, FileInfo{
			Name:    object.Key,
			Size:    object.Size,
			ModTime: object.LastModified,
			IsDir:   false,
		})
	}

	isTruncated := lsRes.IsTruncated
	nextMarker := lsRes.NextMarker

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
