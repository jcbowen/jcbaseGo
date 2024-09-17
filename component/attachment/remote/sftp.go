package remote

import (
	"errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"time"
)

// SFTPConfig 定义了SFTP存储的配置参数。
type SFTPConfig struct {
	Address         string              // SFTP服务器地址，格式为 "host:port"
	Username        string              // SFTP登录用户名
	Password        string              // SFTP登录密码（与PrivateKey二选一）
	PrivateKey      []byte              // SFTP登录私钥（与Password二选一）
	HostKeyCallback ssh.HostKeyCallback // 主机密钥回调，可选，默认不验证
	Timeout         time.Duration       // 连接超时时间，可选，默认5秒
}

// SFTPClient 实现了SFTP存储的客户端。
type SFTPClient struct {
	client *sftp.Client // SFTP客户端
}

// NewSFTPClient 创建一个新的SFTP客户端。
// config：SFTP配置参数
func NewSFTPClient(config SFTPConfig) (*SFTPClient, error) {
	var authMethods []ssh.AuthMethod
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}
	if config.PrivateKey != nil {
		signer, err := ssh.ParsePrivateKey(config.PrivateKey)
		if err != nil {
			return nil, err
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, errors.New("no authentication method provided")
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
		return nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}

	return &SFTPClient{client: client}, nil
}

// Upload 实现了Client接口的Upload方法。
// 将数据上传到指定的远程路径。
func (c *SFTPClient) Upload(remotePath string, data []byte) error {
	remoteFile, err := c.client.Create(remotePath)
	if err != nil {
		return err
	}
	defer func(remoteFile *sftp.File) {
		_ = remoteFile.Close()
	}(remoteFile)

	_, err = remoteFile.Write(data)
	return err
}

// Delete 实现了Client接口的Delete方法。
// 删除指定的远程文件。
func (c *SFTPClient) Delete(remotePath string) error {
	return c.client.Remove(remotePath)
}

// List 实现了Client接口的List方法。
// 列举指定远程目录下的文件。
func (c *SFTPClient) List(remoteDir string) ([]string, error) {
	infos, err := c.client.ReadDir(remoteDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, info := range infos {
		files = append(files, info.Name())
	}
	return files, nil
}

// Close 实现了Client接口的Close方法。
// 关闭SFTP连接。
func (c *SFTPClient) Close() error {
	return c.client.Close()
}
