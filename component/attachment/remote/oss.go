package remote

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSConfig 定义了阿里云OSS的配置参数。
type OSSConfig struct {
	Endpoint        string // OSS服务的Endpoint，例如 "oss-cn-hangzhou.aliyuncs.com"
	AccessKeyID     string // 阿里云访问密钥ID
	AccessKeySecret string // 阿里云访问密钥Secret
	BucketName      string // OSS存储桶名称
	// 可添加更多配置项，例如超时等
}

// OSSClient 实现了阿里云OSS存储的客户端。
type OSSClient struct {
	client *oss.Bucket // OSS存储桶客户端
}

// NewOSSClient 创建一个新的OSS客户端。
// config：OSS配置参数
func NewOSSClient(config OSSConfig) (*OSSClient, error) {
	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		return nil, err
	}

	return &OSSClient{client: bucket}, nil
}

// Upload 实现了Client接口的Upload方法。
// 将数据上传到指定的远程路径。
func (c *OSSClient) Upload(remotePath string, data []byte) error {
	return c.client.PutObject(remotePath, bytes.NewReader(data))
}

// Delete 实现了Client接口的Delete方法。
// 删除指定的远程文件。
func (c *OSSClient) Delete(remotePath string) error {
	return c.client.DeleteObject(remotePath)
}

// List 实现了Client接口的List方法。
// 列举指定远程目录下的文件。
func (c *OSSClient) List(remoteDir string) ([]string, error) {
	var files []string
	marker := oss.Marker("")
	prefix := oss.Prefix(remoteDir)
	for {
		lsRes, err := c.client.ListObjects(marker, prefix)
		if err != nil {
			return nil, err
		}
		for _, object := range lsRes.Objects {
			files = append(files, object.Key)
		}
		if lsRes.IsTruncated {
			marker = oss.Marker(lsRes.NextMarker)
		} else {
			break
		}
	}
	return files, nil
}

// Close 实现了Client接口的Close方法。
// OSS客户端不需要显式关闭连接，因此该方法为空实现。
func (c *OSSClient) Close() error {
	return nil
}
