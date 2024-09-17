package remote

import (
	"bytes"
	"github.com/jlaffaye/ftp"
	"time"
)

// FTPConfig 定义了FTP存储的配置参数。
type FTPConfig struct {
	Address  string        // FTP服务器地址，格式为 "host:port"
	Username string        // FTP登录用户名
	Password string        // FTP登录密码
	Timeout  time.Duration // 连接超时时间，可选，默认5秒
}

// FTPClient 实现了FTP存储的客户端。
type FTPClient struct {
	conn *ftp.ServerConn // FTP服务器连接
}

// NewFTPClient 创建一个新的FTP客户端。
// config：FTP配置参数
func NewFTPClient(config FTPConfig) (*FTPClient, error) {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	conn, err := ftp.Dial(config.Address, ftp.DialWithTimeout(timeout))
	if err != nil {
		return nil, err
	}

	err = conn.Login(config.Username, config.Password)
	if err != nil {
		return nil, err
	}

	return &FTPClient{conn: conn}, nil
}

// Upload 实现了Client接口的Upload方法。
// 将数据上传到指定的远程路径。
func (c *FTPClient) Upload(remotePath string, data []byte) error {
	return c.conn.Stor(remotePath, bytes.NewBuffer(data))
}

// Delete 实现了Client接口的Delete方法。
// 删除指定的远程文件。
func (c *FTPClient) Delete(remotePath string) error {
	return c.conn.Delete(remotePath)
}

// List 实现了Client接口的List方法。
// 列举指定远程目录下的文件。
func (c *FTPClient) List(remoteDir string) ([]string, error) {
	entries, err := c.conn.List(remoteDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name)
	}
	return files, nil
}

// Close 实现了Client接口的Close方法。
// 关闭FTP连接。
func (c *FTPClient) Close() error {
	return c.conn.Quit()
}
