// Package remote 提供了多种远程存储方式的统一接口，包括FTP、SFTP、腾讯云COS、阿里云OSS等。
package remote

import (
	"errors"
)

// 存储类型常量定义
const (
	TypeFTP  = "ftp"  // FTP存储类型
	TypeSFTP = "sftp" // SFTP存储类型
	TypeCOS  = "cos"  // 腾讯云COS存储类型
	TypeOSS  = "oss"  // 阿里云OSS存储类型
	// 可以根据需要扩展其他类型
)

// Client 定义了远程存储的统一接口。
// 无论使用何种存储类型，都可以通过该接口进行文件的上传、删除和列举操作。
type Client interface {
	// Upload 上传文件到远程存储。
	// remotePath：远程存储的文件路径
	// data：要上传的文件数据
	Upload(remotePath string, data []byte) error

	// Delete 从远程存储中删除文件。
	// remotePath：要删除的远程文件路径
	Delete(remotePath string) error

	// List 列举远程存储中指定目录下的文件列表。
	// remoteDir：要列举的远程目录路径
	List(remoteDir string) ([]string, error)

	// Close 关闭与远程存储的连接。
	Close() error
}

// NewClient 创建一个新的远程存储客户端。
// storageType：存储类型，例如 "ftp"、"sftp"、"cos"、"oss"
// config：对应存储类型的配置结构体
func NewClient(storageType string, config interface{}) (Client, error) {
	switch storageType {
	case TypeFTP:
		ftpConfig, ok := config.(FTPConfig)
		if !ok {
			return nil, errors.New("invalid config for FTP")
		}
		return NewFTPClient(ftpConfig)
	case TypeSFTP:
		sftpConfig, ok := config.(SFTPConfig)
		if !ok {
			return nil, errors.New("invalid config for SFTP")
		}
		return NewSFTPClient(sftpConfig)
	case TypeCOS:
		cosConfig, ok := config.(COSConfig)
		if !ok {
			return nil, errors.New("invalid config for COS")
		}
		return NewCOSClient(cosConfig)
	case TypeOSS:
		ossConfig, ok := config.(OSSConfig)
		if !ok {
			return nil, errors.New("invalid config for OSS")
		}
		return NewOSSClient(ossConfig)
	default:
		return nil, errors.New("unsupported storage type")
	}
}
