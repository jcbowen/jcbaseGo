package remote

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sftpTestServer 是一个仅供测试使用的本地 SFTP 服务器。
// 它基于 golang.org/x/crypto/ssh 提供 SSH 传输层，文件系统使用 pkg/sftp 自带的内存实现，
// 全部数据保存在进程内存中，不会读写磁盘，也不访问外部网络。
type sftpTestServer struct {
	ln       net.Listener
	addr     string
	config   *ssh.ServerConfig
	handler  sftp.Handlers
	username string
	password string

	mu     sync.Mutex
	conns  map[net.Conn]struct{}
	closed bool
}

// newSFTPTestServer 启动一个本地内存 SFTP 服务器，并注册 t.Cleanup 自动关闭。
// 参数：
//   - t: 测试对象，用于注册清理逻辑
//
// 返回值：
//   - *sftpTestServer: 已启动并可接受连接的服务器实例
//
// 异常：
//   - 生成主机密钥或监听回环地址失败时调用 t.Fatalf 终止测试
//
// 使用示例：
//
//	srv := newSFTPTestServer(t)
//	client, err := NewSFTPClient(SFTPConfig(jcbaseGo.SFTPStruct{Address: srv.Address(), ...}))
func newSFTPTestServer(t *testing.T) *sftpTestServer {
	t.Helper()

	signer, err := generateTestHostKey()
	if err != nil {
		t.Fatalf("生成 SFTP 测试服务器主机密钥失败: %v", err)
	}

	ln, err := net.Listen("tcp", net.JoinHostPort(localTestHost, "0"))
	if err != nil {
		t.Fatalf("启动本地 SFTP 测试服务器失败: %v", err)
	}

	srv := &sftpTestServer{
		ln:       ln,
		addr:     ln.Addr().String(),
		handler:  sftp.InMemHandler(),
		username: "test",
		password: "test",
		conns:    make(map[net.Conn]struct{}),
	}

	srv.config = &ssh.ServerConfig{PasswordCallback: srv.authenticate}
	srv.config.AddHostKey(signer)

	t.Cleanup(srv.Close)

	go srv.acceptLoop()
	return srv
}

// generateTestHostKey 生成一对仅用于测试的 ed25519 主机密钥。
// 参数：
//   - 无
//
// 返回值：
//   - ssh.Signer: 可被 SSH 服务端使用的签名器
//   - error: 密钥生成或转换失败时返回错误
//
// 异常：
//   - 无
//
// 使用示例：
//
//	signer, err := generateTestHostKey()
func generateTestHostKey() (ssh.Signer, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成 ed25519 密钥失败: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, fmt.Errorf("转换 SSH 签名器失败: %w", err)
	}
	return signer, nil
}

// Address 返回服务器的 host:port 地址，可直接用于 SFTPConfig.Address。
// 参数：
//   - 无
//
// 返回值：
//   - string: "127.0.0.1:<随机端口>"
//
// 异常：
//   - 无
//
// 使用示例：
//
//	config.Address = srv.Address()
func (s *sftpTestServer) Address() string {
	return s.addr
}

// authenticate 校验客户端的用户名与密码。
// 参数：
//   - conn: 连接元信息，可获取用户名等信息
//   - password: 客户端传入的密码
//
// 返回值：
//   - *ssh.Permissions: 认证通过时返回 nil
//   - error: 用户名或密码不匹配时返回错误
//
// 异常：
//   - 无
//
// 使用示例：
//
//	perms, err := srv.authenticate(connMeta, []byte("test"))
func (s *sftpTestServer) authenticate(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	if conn.User() == s.username && string(password) == s.password {
		return nil, nil
	}
	return nil, fmt.Errorf("用户名或密码错误")
}

// Close 关闭监听器与所有已建立的连接，可重复调用。
// 参数：
//   - 无
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（关闭过程中的错误被忽略）
//
// 使用示例：
//
//	defer srv.Close()
func (s *sftpTestServer) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true

	conns := make([]net.Conn, 0, len(s.conns))
	for conn := range s.conns {
		conns = append(conns, conn)
	}
	s.conns = make(map[net.Conn]struct{})
	s.mu.Unlock()

	_ = s.ln.Close()
	for _, conn := range conns {
		_ = conn.Close()
	}
}

// acceptLoop 循环接受连接，每条连接交由独立 goroutine 完成 SSH 握手。
// 参数：
//   - 无
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（监听器关闭时静默退出）
//
// 使用示例：
//
//	go srv.acceptLoop()
func (s *sftpTestServer) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}

		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			_ = conn.Close()
			return
		}
		s.conns[conn] = struct{}{}
		s.mu.Unlock()

		go s.handleConn(conn)
	}
}

// handleConn 完成 SSH 握手并处理客户端打开的会话通道。
// 参数：
//   - conn: 已接受的 TCP 连接
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（握手失败时静默关闭连接）
//
// 使用示例：
//
//	go srv.handleConn(conn)
func (s *sftpTestServer) handleConn(conn net.Conn) {
	sshConn, channels, requests, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		return
	}
	defer func() { _ = sshConn.Close() }()

	go ssh.DiscardRequests(requests)

	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "仅支持 session 通道")
			continue
		}

		channel, channelRequests, err := newChannel.Accept()
		if err != nil {
			return
		}
		go s.serveSFTP(channel, channelRequests)
	}
}

// serveSFTP 在会话通道上启动 SFTP 子系统。
// 参数：
//   - channel: 已接受的 session 通道
//   - requests: 该通道上的 SSH 请求流
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（子系统退出时清理通道）
//
// 使用示例：
//
//	go srv.serveSFTP(channel, channelRequests)
func (s *sftpTestServer) serveSFTP(channel ssh.Channel, requests <-chan *ssh.Request) {
	for req := range requests {
		accepted := req.Type == "subsystem" && len(req.Payload) >= 4 && string(req.Payload[4:]) == "sftp"
		if req.WantReply {
			_ = req.Reply(accepted, nil)
		}
		if !accepted {
			continue
		}

		server := sftp.NewRequestServer(channel, s.handler)
		if err := server.Serve(); err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("本地 SFTP 测试服务器会话结束: %v\n", err)
		}
		_ = server.Close()
		return
	}
}
