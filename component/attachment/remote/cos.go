package remote

import (
	"bytes"
	"context"
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
