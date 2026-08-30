// Package remote 提供多种远程存储方式的统一接口。
package remote

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
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

// ErrPresignNotSupported 表示当前存储类型不支持预签名上传。
var ErrPresignNotSupported = errors.New("current storage type does not support presigned upload")

// PresignOptions 定义了生成预签名 URL 的可选参数。
// 不同存储类型对字段的支持程度可能不同，不支持的字段会被忽略。
type PresignOptions struct {
	Expires     time.Duration     // 预签名 URL 的有效期，为零时使用存储类型的默认有效期
	ContentType string            // 上传文件的 Content-Type，为空时不参与签名
	Metadata    map[string]string // 自定义元数据，为空时不参与签名
}

// withContextTimeout 包装一个操作，使其能够响应上下文的取消信号。
// 注意：该函数同步执行 fn，仅在操作前后检查 ctx 状态。由于底层 FTP/SFTP SDK 不支持
// context，无法中断正在进行的 IO 操作，但可保证在 ctx 已取消时不发起新操作，并在操作
// 完成后及时发现取消状态。此实现消除了早期 goroutine 方案带来的并发安全隐患。
func withContextTimeout[T any](ctx context.Context, op string, fn func() (T, error)) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, &Error{Op: op, Err: ctx.Err()}
	default:
	}

	result, err := fn()
	if err != nil {
		var zero T
		return zero, &Error{Op: op, Err: err}
	}

	select {
	case <-ctx.Done():
		var zero T
		return zero, &Error{Op: op, Err: ctx.Err()}
	default:
		return result, nil
	}
}

// withContextTimeoutVoid 包装一个无返回值的操作，使其能够响应上下文的取消信号。
// 实现逻辑与 withContextTimeout 一致，均为同步执行并在操作前后检查 ctx 状态。
func withContextTimeoutVoid(ctx context.Context, op string, fn func() error) error {
	select {
	case <-ctx.Done():
		return &Error{Op: op, Err: ctx.Err()}
	default:
	}

	if err := fn(); err != nil {
		return &Error{Op: op, Err: err}
	}

	select {
	case <-ctx.Done():
		return &Error{Op: op, Err: ctx.Err()}
	default:
		return nil
	}
}

// normalizeAddress 校验并规范化 FTP/SFTP 服务器地址。
// - 去除可能存在的 scheme（如 ftp:// / sftp://）
// - 若未指定端口，则追加 defaultPort
// 返回值：
//   - 规范化后的 host:port 地址
//   - 地址非法时返回错误
func normalizeAddress(address, defaultPort string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", errors.New("服务器地址不能为空")
	}

	// 去除 scheme
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return "", fmt.Errorf("服务器地址格式错误: %w", err)
		}
		address = u.Host
	}

	if address == "" {
		return "", errors.New("服务器地址不能为空")
	}

	// 未指定端口时追加默认端口
	if !strings.Contains(address, ":") {
		address = net.JoinHostPort(address, defaultPort)
	}

	return address, nil
}

// ListOptions 定义了分页和过滤选项
// 注意：不同存储类型对 Prefix 的语义不同：
//   - COS/OSS：对象键前缀过滤
//   - FTP/SFTP：待列出内容的目录路径
type ListOptions struct {
	Prefix  string // 文件名前缀过滤或目录路径，可选
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

// PresignUploader 定义了支持预签名上传的存储类型需要实现的接口。
// 注意：该接口是可选的，FTP/SFTP 等不支持预签名 URL 的存储类型无需实现。
type PresignUploader interface {
	// PresignUpload 生成用于上传指定对象的预签名 URL。
	// - remotePath: 对象在存储桶中的路径（Key）。
	// - opts: 预签名选项，nil 时表示使用默认选项。
	// 返回值：
	//   - string: 预签名 URL。
	//   - map[string]string: 需要在上传请求中携带的已签名请求头。
	//   - error: 生成过程中出现的错误。
	PresignUpload(ctx context.Context, remotePath string, opts *PresignOptions) (string, map[string]string, error)
}

// ObjectExister 定义了支持对象存在性探测的存储类型需要实现的接口。
// 注意：该接口是可选的，FTP/SFTP 等不支持低成本探测的存储类型无需实现；
// 未实现时上层会按「对象存在」处理，以保持原有秒传行为不被改变。
//
// 该接口与 PresignUploader 成对出现：预签名上传存在「客户端拿到URL却未真正上传」的情况，
// 秒传命中时必须探测对象是否真实存在。服务端直传的存储类型（FTP/SFTP）不存在这个问题。
// 目前 OSS 是唯一实现 PresignUploader 的类型，因此也只有它实现了本接口；
// 若后续为 COS 等类型补充 PresignUploader，请同步实现本接口。
type ObjectExister interface {
	// Exists 判断指定对象是否已存在于存储中。
	// - remotePath: 对象在存储中的相对路径（Key）。
	// 返回值：
	//   - bool: 对象存在返回 true，不存在返回 false。
	//   - error: 探测过程中出现异常时返回错误。
	Exists(ctx context.Context, remotePath string) (bool, error)
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
		c, err := NewFTPClient(ftpConfig)
		if err != nil {
			return nil, err
		}
		return c, nil
	case TypeSFTP:
		sftpConfig, ok := config.(SFTPConfig)
		if !ok {
			return nil, errors.New("invalid config for SFTP")
		}
		c, err := NewSFTPClient(sftpConfig, hostKeyCallback)
		if err != nil {
			return nil, err
		}
		return c, nil
	case TypeCOS:
		cosConfig, ok := config.(COSConfig)
		if !ok {
			return nil, errors.New("invalid config for COS")
		}
		c, err := NewCOSClient(cosConfig)
		if err != nil {
			return nil, err
		}
		return c, nil
	case TypeOSS:
		ossConfig, ok := config.(OSSConfig)
		if !ok {
			return nil, errors.New("invalid config for OSS")
		}
		c, err := NewOSSClient(ossConfig)
		if err != nil {
			return nil, err
		}
		return c, nil
	default:
		return nil, errors.New("unsupported storage type")
	}
}
