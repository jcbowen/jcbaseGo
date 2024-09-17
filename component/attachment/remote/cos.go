package remote

import (
	"bytes"
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
)

// COSConfig 定义了腾讯云COS的配置参数。
type COSConfig struct {
	BucketURL string // COS存储桶URL，例如 "https://your-bucket.cos.region.myqcloud.com"
	SecretID  string // 腾讯云API密钥ID
	SecretKey string // 腾讯云API密钥Key
	// 可添加更多配置项，例如超时等
}

// COSClient 实现了腾讯云COS存储的客户端。
type COSClient struct {
	client *cos.Client // COS客户端
}

// NewCOSClient 创建一个新的COS客户端。
// config：COS配置参数
func NewCOSClient(config COSConfig) (*COSClient, error) {
	u, err := url.Parse(config.BucketURL)
	if err != nil {
		return nil, err
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
	})

	return &COSClient{client: client}, nil
}

// Upload 实现了Client接口的Upload方法。
// 将数据上传到指定的远程路径。
func (c *COSClient) Upload(remotePath string, data []byte) error {
	_, err := c.client.Object.Put(context.Background(), remotePath, bytes.NewReader(data), nil)
	return err
}

// Delete 实现了Client接口的Delete方法。
// 删除指定的远程文件。
func (c *COSClient) Delete(remotePath string) error {
	_, err := c.client.Object.Delete(context.Background(), remotePath)
	return err
}

// List 实现了Client接口的List方法。
// 列举指定远程目录下的文件。
func (c *COSClient) List(remoteDir string) ([]string, error) {
	opt := &cos.BucketGetOptions{
		Prefix:  remoteDir,
		MaxKeys: 1000,
	}
	result, _, err := c.client.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, content := range result.Contents {
		files = append(files, content.Key)
	}
	return files, nil
}

// Close 实现了Client接口的Close方法。
// COS客户端不需要显式关闭连接，因此该方法为空实现。
func (c *COSClient) Close() error {
	return nil
}
