package remote

import (
	"bytes"
	"context"
	"github.com/jcbowen/jcbaseGo"
	"io"
	"sync"
	"time"

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

	select {
	case <-ctx.Done():
		return &Error{Op: "Upload", Err: ctx.Err()}
	default:
	}

	err := c.conn.Stor(remotePath, bytes.NewReader(data))
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}

// Download 实现了Client接口的Download方法。
func (c *FTPClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, &Error{Op: "Download", Err: ctx.Err()}
	default:
	}

	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	defer func() { _ = resp.Close() }()

	data, err := io.ReadAll(resp)
	if err != nil {
		return nil, &Error{Op: "Download", Err: err}
	}
	return data, nil
}

// Delete 实现了Client接口的Delete方法。
func (c *FTPClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return &Error{Op: "Delete", Err: ctx.Err()}
	default:
	}

	err := c.conn.Delete(remotePath)
	if err != nil {
		return &Error{Op: "Delete", Err: err}
	}
	return nil
}

// List 实现了Client接口的List方法。
func (c *FTPClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return ListResult{}, &Error{Op: "List", Err: ctx.Err()}
	default:
	}

	entries, err := c.conn.List(options.Prefix)
	if err != nil {
		return ListResult{}, &Error{Op: "List", Err: err}
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

	err := c.conn.Quit()
	if err != nil {
		return &Error{Op: "Close", Err: err}
	}
	return nil
}
