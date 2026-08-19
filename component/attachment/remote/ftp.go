package remote

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path"
	"sort"
	"strings"
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
	conn   *ftp.ServerConn
	config FTPConfig
	closed bool
	mu     sync.Mutex
}

// NewFTPClient 创建一个新的FTP客户端。
// 返回值：
//   - 创建成功返回 *FTPClient
//   - 地址格式错误或登录失败时返回错误
func NewFTPClient(config FTPConfig) (*FTPClient, error) {
	address, err := normalizeAddress(config.Address, "21")
	if err != nil {
		return nil, &Error{Op: "NewFTPClient", Err: err}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	conn, err := ftp.Dial(address, ftp.DialWithTimeout(timeout))
	if err != nil {
		return nil, &Error{Op: "Dial", Err: err}
	}

	err = conn.Login(config.Username, config.Password)
	if err != nil {
		return nil, &Error{Op: "Login", Err: err}
	}

	return &FTPClient{conn: conn, config: config}, nil
}

// Upload 实现了Client接口的Upload方法。
// 上传前会自动创建 remotePath 所在的多级父目录，并检测/恢复连接。
func (c *FTPClient) Upload(ctx context.Context, remotePath string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Upload", func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}

		// FTP 协议不会自动创建父目录，上传前先确保目录存在
		if err := c.ensureDir(path.Dir(remotePath)); err != nil {
			return err
		}
		return c.conn.Stor(remotePath, bytes.NewReader(data))
	})
}

// ensureDir 递归创建 FTP 远端目录。
// - dir: 目标目录路径，使用 POSIX 风格的分隔符 "/"
// 返回值：
//   - 创建成功或目录已存在时返回 nil
//   - 创建失败时返回底层错误
func (c *FTPClient) ensureDir(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" || dir == "/" || dir == "." {
		return nil
	}

	// 先尝试切换目录，成功则说明目录已存在
	if err := c.conn.ChangeDir(dir); err == nil {
		return nil
	}

	// 递归创建父目录
	parent := path.Dir(dir)
	if parent != dir && parent != "/" {
		if err := c.ensureDir(parent); err != nil {
			return err
		}
	}

	// 创建当前目录，若因并发等原因已存在则忽略错误
	if err := c.conn.MakeDir(dir); err != nil {
		if err := c.conn.ChangeDir(dir); err == nil {
			return nil
		}
		return err
	}
	return nil
}

// Download 实现了Client接口的Download方法。
// 下载前会检测/恢复连接。
func (c *FTPClient) Download(ctx context.Context, remotePath string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 第一步：获取文件响应
	resp, err := withContextTimeout(ctx, "Download", func() (*ftp.Response, error) {
		if err := c.ensureConnected(); err != nil {
			return nil, err
		}
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
// 删除前会检测/恢复连接。
func (c *FTPClient) Delete(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return withContextTimeoutVoid(ctx, "Delete", func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}
		return c.conn.Delete(remotePath)
	})
}

// List 实现了Client接口的List方法。
// 注意：FTP 的 Prefix 表示待列出内容的目录路径，非对象前缀。
// 列出前会检测/恢复连接，返回结果按文件名排序后做本地分页。
func (c *FTPClient) List(ctx context.Context, options ListOptions) (ListResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := withContextTimeout(ctx, "List", func() ([]*ftp.Entry, error) {
		if err := c.ensureConnected(); err != nil {
			return nil, err
		}
		return c.conn.List(options.Prefix)
	})
	if err != nil {
		return ListResult{}, err
	}

	var files []FileInfo
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:    entry.Name,
			Size:    int64(entry.Size),
			ModTime: entry.Time,
			IsDir:   entry.Type == ftp.EntryTypeFolder,
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
// 关闭 FTP 连接并将内部引用置空，标记客户端为已关闭状态。
func (c *FTPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	if c.conn == nil {
		return nil
	}

	err := c.conn.Quit()
	c.conn = nil
	if err != nil {
		return &Error{Op: "Close", Err: err}
	}
	return nil
}

// ensureConnected 检测当前 FTP 连接是否可用，若不可用则尝试重连。
// 调用方必须已经持有 c.mu 锁。
func (c *FTPClient) ensureConnected() error {
	if c.closed {
		return errors.New("FTP 客户端已关闭")
	}
	if c.conn == nil {
		return c.reconnect()
	}
	// 通过 NOOP 轻量检测连接活性
	if err := c.conn.NoOp(); err == nil {
		return nil
	}
	return c.reconnect()
}

// reconnect 使用创建时保存的配置重新建立 FTP 连接。
// 调用方必须已经持有 c.mu 锁。
func (c *FTPClient) reconnect() error {
	newClient, err := NewFTPClient(c.config)
	if err != nil {
		return err
	}

	// 关闭旧连接，忽略错误
	if c.conn != nil {
		_ = c.conn.Quit()
	}

	c.conn = newClient.conn
	return nil
}
