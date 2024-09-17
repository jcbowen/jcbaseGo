package remote

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPConfig 定义了SFTP存储的配置参数。
type SFTPConfig struct {
	Address         string              // SFTP服务器地址
	Username        string              // SFTP登录用户名
	Password        string              // SFTP登录密码
	PrivateKey      []byte              // SFTP登录私钥
	HostKeyCallback ssh.HostKeyCallback // 主机密钥回调，可选
	Timeout         time.Duration       // 连接超时时间，可选
}

// SFTPClient 实现了SFTP存储的客户端。
// 注意：SFTPClient是并发安全的。
type SFTPClient struct {
	client *sftp.Client
	mu     sync.Mutex
}

// NewSFTPClient 创建一个新的SFTP客户端。
func NewSFTPClient(config SFTPConfig) (*SFTPClient, error) {
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
		return nil, &Error{Op: "NewSFTPClient", Err: errors.New("no authentication method provided")}
	}

	hostKeyCallback := config.HostKeyCallback
	if hostKeyCallback == nil {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
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

	conn, err := ssh.Dial("tcp", config.Address, sshConfig)
	if err != nil {
		return nil, &Error{Op: "Dial", Err: err}
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, &Error{Op: "NewSFTPClient", Err: err}
	}

	return &SFTPClient{client: client}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *SFTPClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return &Error{Op: "Upload", Err: ctx.Err()}
	default:
	}

	remoteFile, err := c.client.Create(remotePath)
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	defer func() { _ = remoteFile.Close() }()

	_, err = remoteFile.Write(data)
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}

// Download 实现了Client接口的Download方法。
func (c *SFTPClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, &Error{Op: "Download", Err: ctx.Err()}
	default:
	}

	remoteFile, err := c.client.Open(remotePath)
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	defer func() { _ = remoteFile.Close() }()

	data, err := io.ReadAll(remoteFile)
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	return data, nil
}

// Delete 实现了Client接口的Delete方法。
func (c *SFTPClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return &Error{Op: "Delete", Err: ctx.Err()}
	default:
	}

	err := c.client.Remove(remotePath)
	if err != nil {
		return &Error{Op: "Delete", Err: err}
	}
	return nil
}

// List 实现了Client接口的List方法。
func (c *SFTPClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return ListResult{}, &Error{Op: "List", Err: ctx.Err()}
	default:
	}

	entries, err := c.client.ReadDir(options.Prefix)
	if err != nil {
		return ListResult{}, &Error{Op: "List", Err: err}
	}

	var files []FileInfo // 使用 var 声明空切片
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			ModTime: entry.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	// 模拟分页
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
func (c *SFTPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.client.Close()
	if err != nil {
		return &Error{Op: "Close", Err: err}
	}
	return nil
}
