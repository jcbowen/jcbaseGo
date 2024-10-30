// Package remote 提供多种远程存储方式的统一接口。
package remote

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

// 存储类型常量定义
const (
	TypeFTP  = "ftp"  // FTP存储类型
	TypeSFTP = "sftp" // SFTP存储类型
	TypeCOS  = "cos"  // 腾讯云COS存储类型
	TypeOSS  = "oss"  // 阿里云OSS存储类型
)

// Error 定义了统一的错误类型
type Error struct {
	Op  string // 操作名称，如"Upload"
	Err error  // 实际的错误
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s operation failed: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// ListOptions 定义了分页和过滤选项
type ListOptions struct {
	Prefix  string // 文件名前缀过滤，可选
	Marker  string // 分页标记，可选
	MaxKeys int    // 每次返回的最大文件数量，可选
}

// FileInfo 定义了文件的元数据信息
type FileInfo struct {
	Name    string    // 文件名
	Size    int64     // 文件大小（字节）
	ModTime time.Time // 修改时间
	IsDir   bool      // 是否为目录
}

// ListResult 定义了List方法的返回结果
type ListResult struct {
	Files       []FileInfo // 文件信息列表
	NextMarker  string     // 下一个分页标记
	IsTruncated bool       // 是否还有更多数据
}

// Client 定义了远程存储的统一接口。
// 注意：所有方法都应是并发安全的。
type Client interface {
	Upload(ctx context.Context, remotePath string, data []byte) error
	Download(ctx context.Context, remotePath string) ([]byte, error)
	Delete(ctx context.Context, remotePath string) error
	List(ctx context.Context, options ListOptions) (ListResult, error)
	Close() error
}

// NewClient 创建一个新的远程存储客户端。
// - storageType 远程附件类型
// - config 远程附件连接配置
// - hostKeyCallback 主机密钥回调，选填，仅sftp有效
func NewClient(storageType string, config interface{}, args ...ssh.HostKeyCallback) (Client, error) {
	// 仅sftp需要
	var hostKeyCallback ssh.HostKeyCallback
	if storageType == TypeSFTP && len(args) > 0 {
		hostKeyCallback = args[0]
	}

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
		return NewSFTPClient(sftpConfig, hostKeyCallback)
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
