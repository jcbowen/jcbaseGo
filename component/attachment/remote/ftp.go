package remote

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jlaffaye/ftp"
)

// FTPConfig 定义了FTP存储的配置参数。
type FTPConfig jcbaseGo.FTPStruct

// FTPClient 实现了FTP存储的客户端。
// 注意：FTPClient是并发安全的。
type FTPClient struct {
	conn *ftp.ServerConn
	mu   sync.Mutex
}

// NewFTPClient 创建一个新的FTP客户端。
func NewFTPClient(config FTPConfig) (*FTPClient, error) {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	conn, err := ftp.Dial(config.Address, ftp.DialWithTimeout(timeout))
	if err != nil {
		return nil, &Error{Op: "Dial", Err: err}
	}

	err = conn.Login(config.Username, config.Password)
	if err != nil {
		return nil, &Error{Op: "Login", Err: err}
	}

	return &FTPClient{conn: conn}, nil
}

// Upload 实现了Client接口的Upload方法。
func (c *FTPClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Upload", func() error {
		return c.conn.Stor(remotePath, bytes.NewReader(data))
	})
}

// Download 实现了Client接口的Download方法。
func (c *FTPClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 第一步：获取文件响应
	resp, err := withContextTimeout(ctx, "Download", func() (*ftp.Response, error) {
		return c.conn.Retr(remotePath)
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Close() }()

	// 第二步：读取文件数据
	return withContextTimeout(ctx, "Download", func() ([]byte, error) {
		return io.ReadAll(resp)
	})
}

// Delete 实现了Client接口的Delete方法。
func (c *FTPClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Delete", func() error {
		return c.conn.Delete(remotePath)
	})
}

// List 实现了Client接口的List方法。
func (c *FTPClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取文件列表
	entries, err := withContextTimeout(ctx, "List", func() ([]*ftp.Entry, error) {
		return c.conn.List(options.Prefix)
	})
	if err != nil {
		return ListResult{}, err
	}

	var files []FileInfo // 使用 var 声明空切片
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:    entry.Name,
			Size:    int64(entry.Size),
			ModTime: entry.Time,
			IsDir:   entry.Type == ftp.EntryTypeFolder,
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
func (c *FTPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Quit()
	if err != nil {
		return &Error{Op: "Close", Err: err}
	}
	return nil
}
