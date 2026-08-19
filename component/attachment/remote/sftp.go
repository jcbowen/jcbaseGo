package remote

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPConfig 定义了SFTP存储的配置参数。
type SFTPConfig jcbaseGo.SFTPStruct

// SFTPClient 实现了SFTP存储的客户端。
// 注意：SFTPClient是并发安全的。
type SFTPClient struct {
	client  *sftp.Client
	sshConn *ssh.Client
	config  SFTPConfig
	closed  bool
	mu      sync.Mutex
}

// NewSFTPClient 创建一个新的SFTP客户端。
// - config SFTP 连接配置
// - args 可选的主机密钥回调，未提供时使用 ssh.InsecureIgnoreHostKey
// 返回值：
//   - 创建成功返回 *SFTPClient
//   - 地址格式错误、认证信息缺失或连接失败时返回错误
func NewSFTPClient(config SFTPConfig, args ...ssh.HostKeyCallback) (*SFTPClient, error) {
	var hostKeyCallback ssh.HostKeyCallback
	if len(args) > 0 {
		hostKeyCallback = args[0]
	}

	var authMethods []ssh.AuthMethod
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}
	if config.PrivateKey != nil {
		signer, err := ssh.ParsePrivateKey(config.PrivateKey)
		if err != nil {
			return nil, &Error{Op: "ParsePrivateKey", Err: err}
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, &Error{Op: "NewSFTPClient", Err: errors.New("未提供 SFTP 认证方式")}
	}

	if hostKeyCallback == nil {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	address, err := normalizeAddress(config.Address, "22")
	if err != nil {
		return nil, &Error{Op: "NewSFTPClient", Err: err}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, &Error{Op: "Dial", Err: err}
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		_ = conn.Close()
		return nil, &Error{Op: "NewSFTPClient", Err: err}
	}

	return &SFTPClient{client: client, sshConn: conn, config: config}, nil
}

// Upload 实现了Client接口的Upload方法。
// 上传前会自动创建 remotePath 所在的多级父目录，并检测/恢复连接。
func (c *SFTPClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Upload", func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}

		// SFTP Create 不会自动创建父目录，需先确保目录存在
		if err := c.client.MkdirAll(path.Dir(remotePath)); err != nil {
			return err
		}

		remoteFile, err := c.client.Create(remotePath)
		if err != nil {
			return err
		}
		defer func() { _ = remoteFile.Close() }()

		_, err = remoteFile.Write(data)
		return err
	})
}

// Download 实现了Client接口的Download方法。
// 下载前会检测/恢复连接。
func (c *SFTPClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeout(ctx, "Download", func() ([]byte, error) {
		if err := c.ensureConnected(); err != nil {
			return nil, err
		}

		remoteFile, err := c.client.Open(remotePath)
		if err != nil {
			return nil, err
		}
		defer func() { _ = remoteFile.Close() }()

		return io.ReadAll(remoteFile)
	})
}

// Delete 实现了Client接口的Delete方法。
// 删除前会检测/恢复连接。
func (c *SFTPClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Delete", func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}
		return c.client.Remove(remotePath)
	})
}

// List 实现了Client接口的List方法。
// 注意：SFTP 的 Prefix 表示待列出内容的目录路径，非对象前缀。
// 列出前会检测/恢复连接，返回结果按文件名排序后做本地分页。
func (c *SFTPClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := withContextTimeout(ctx, "List", func() ([]os.FileInfo, error) {
		if err := c.ensureConnected(); err != nil {
			return nil, err
		}
		return c.client.ReadDir(options.Prefix)
	})
	if err != nil {
		return ListResult{}, err
	}

	var files []FileInfo
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			ModTime: entry.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	// 按文件名排序，保证 Marker 分页可预期
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	// 本地模拟分页
	start := 0
	if options.Marker != "" {
		for i, file := range files {
			if file.Name == options.Marker {
				start = i + 1
				break
			}
		}
	}

	end := len(files)
	if options.MaxKeys > 0 && start+options.MaxKeys < end {
		end = start + options.MaxKeys
	}

	resultFiles := files[start:end]
	isTruncated := end < len(files)
	nextMarker := ""
	if isTruncated {
		nextMarker = resultFiles[len(resultFiles)-1].Name
	}

	return ListResult{
		Files:       resultFiles,
		NextMarker:  nextMarker,
		IsTruncated: isTruncated,
	}, nil
}

// Close 实现了Client接口的Close方法。
// 关闭 SFTP 客户端和底层 SSH 连接，并将内部引用置空，标记客户端为已关闭状态。
func (c *SFTPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	var err1, err2 error
	// 先关闭SFTP客户端，再关闭SSH连接
	if c.client != nil {
		err1 = c.client.Close()
		c.client = nil
	}
	if c.sshConn != nil {
		err2 = c.sshConn.Close()
		c.sshConn = nil
	}

	if err1 != nil {
		return &Error{Op: "Close", Err: err1}
	}
	if err2 != nil {
		return &Error{Op: "Close", Err: err2}
	}
	return nil
}

// ensureConnected 检测当前 SFTP 连接是否可用，若不可用则尝试重连。
// 调用方必须已经持有 c.mu 锁。
func (c *SFTPClient) ensureConnected() error {
	if c.closed {
		return errors.New("SFTP 客户端已关闭")
	}
	if c.client == nil || c.sshConn == nil {
		return c.reconnect()
	}
	// 通过 Getwd 轻量检测连接活性
	if _, err := c.client.Getwd(); err == nil {
		return nil
	}
	return c.reconnect()
}

// reconnect 使用创建时保存的配置重新建立 SFTP 连接。
// 调用方必须已经持有 c.mu 锁。
func (c *SFTPClient) reconnect() error {
	newClient, err := NewSFTPClient(c.config)
	if err != nil {
		return err
	}

	// 关闭旧连接，忽略错误
	if c.client != nil {
		_ = c.client.Close()
	}
	if c.sshConn != nil {
		_ = c.sshConn.Close()
	}

	c.client = newClient.client
	c.sshConn = newClient.sshConn
	return nil
}
