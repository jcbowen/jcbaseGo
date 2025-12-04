package remote

import (
	"context"
	"errors"
	"io"
	"os"
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
	mu      sync.Mutex
}

// NewSFTPClient 创建一个新的SFTP客户端。
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
		return nil, &Error{Op: "NewSFTPClient", Err: errors.New("no authentication method provided")}
	}

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
		_ = conn.Close()
		return nil, &Error{Op: "NewSFTPClient", Err: err}
	}

	return &SFTPClient{client: client, sshConn: conn}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *SFTPClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Upload", func() error {
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
func (c *SFTPClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeout(ctx, "Download", func() ([]byte, error) {
		remoteFile, err := c.client.Open(remotePath)
		if err != nil {
			return nil, err
		}
		defer func() { _ = remoteFile.Close() }()

		return io.ReadAll(remoteFile)
	})
}

// Delete 实现了Client接口的Delete方法。
func (c *SFTPClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Delete", func() error {
		return c.client.Remove(remotePath)
	})
}

// List 实现了Client接口的List方法。
func (c *SFTPClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取文件列表
	entries, err := withContextTimeout(ctx, "List", func() ([]os.FileInfo, error) {
		return c.client.ReadDir(options.Prefix)
	})
	if err != nil {
		return ListResult{}, err
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

	var err1, err2 error
	// 先关闭SFTP客户端，再关闭SSH连接
	if c.client != nil {
		err1 = c.client.Close()
	}
	if c.sshConn != nil {
		err2 = c.sshConn.Close()
	}

	if err1 != nil {
		return &Error{Op: "Close", Err: err1}
	}
	if err2 != nil {
		return &Error{Op: "Close", Err: err2}
	}
	return nil
}
