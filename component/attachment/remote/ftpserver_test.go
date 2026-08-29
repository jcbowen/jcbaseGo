package remote

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// localTestHost 本地测试服务器绑定的回环地址。
// 数据连接同样监听在该地址上，保证与客户端发起连接时使用的 host 一致。
const localTestHost = "127.0.0.1"

// ftpMemNode 描述内存文件系统中的一个节点（文件或目录）。
// 作为值类型在 ftpMemFS 外部传递，避免并发读写底层数据产生数据竞争。
type ftpMemNode struct {
	name    string    // 节点名称（不含路径）
	path    string    // 规范化后的绝对路径
	isDir   bool      // 是否为目录
	content []byte    // 文件内容，目录为空
	modTime time.Time // 最后修改时间
}

// ftpMemFS 是一个并发安全的内存文件系统，用于模拟 FTP 服务器上的目录树。
type ftpMemFS struct {
	mu    sync.Mutex
	nodes map[string]*ftpMemNode
}

// newFTPMemFS 创建一个仅包含根目录的内存文件系统。
// 参数：
//   - 无
//
// 返回值：
//   - *ftpMemFS: 已初始化根目录的文件系统实例
//
// 异常：
//   - 无
//
// 使用示例：
//
//	fs := newFTPMemFS()
//	_ = fs.mkdir("/a/b")
func newFTPMemFS() *ftpMemFS {
	return &ftpMemFS{
		nodes: map[string]*ftpMemNode{
			"/": {name: "/", path: "/", isDir: true, modTime: time.Now()},
		},
	}
}

// clean 将任意路径规范化为绝对路径形式。
// 参数：
//   - p: 原始路径，允许为空或相对路径
//
// 返回值：
//   - string: 以 "/" 开头的规范化路径
//
// 异常：
//   - 无
//
// 使用示例：
//
//	fs.clean("a/../b") // "/b"
func (fs *ftpMemFS) clean(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return path.Clean(p)
}

// stat 查询指定路径的节点快照。
// 参数：
//   - p: 待查询的路径
//
// 返回值：
//   - ftpMemNode: 节点副本，避免调用方持有内部指针
//   - bool: 节点是否存在
//
// 异常：
//   - 无
//
// 使用示例：
//
//	node, ok := fs.stat("/a/hello.txt")
func (fs *ftpMemFS) stat(p string) (ftpMemNode, bool) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	node, ok := fs.nodes[fs.clean(p)]
	if !ok {
		return ftpMemNode{}, false
	}
	return *node, true
}

// mkdir 创建单层目录。与真实 FTP 服务器一致，父目录不存在时创建失败。
// 参数：
//   - p: 待创建的目录路径
//
// 返回值：
//   - error: 目录已存在或父目录不存在时返回错误
//
// 异常：
//   - 无
//
// 使用示例：
//
//	if err := fs.mkdir("/a/b"); err != nil { ... }
func (fs *ftpMemFS) mkdir(p string) error {
	target := fs.clean(p)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, ok := fs.nodes[target]; ok {
		return fmt.Errorf("目录已存在: %s", target)
	}
	if parent, ok := fs.nodes[path.Dir(target)]; !ok || !parent.isDir {
		return fmt.Errorf("父目录不存在: %s", path.Dir(target))
	}

	fs.nodes[target] = &ftpMemNode{
		name:    path.Base(target),
		path:    target,
		isDir:   true,
		modTime: time.Now(),
	}
	return nil
}

// writeFile 写入或覆盖文件内容。
// 参数：
//   - p: 文件绝对路径
//   - data: 文件内容
//
// 返回值：
//   - error: 目标为目录时返回错误
//
// 异常：
//   - 无
//
// 使用示例：
//
//	if err := fs.writeFile("/a/hello.txt", []byte("hi")); err != nil { ... }
func (fs *ftpMemFS) writeFile(p string, data []byte) error {
	target := fs.clean(p)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if node, ok := fs.nodes[target]; ok && node.isDir {
		return fmt.Errorf("目标为目录: %s", target)
	}

	fs.nodes[target] = &ftpMemNode{
		name:    path.Base(target),
		path:    target,
		content: data,
		modTime: time.Now(),
	}
	return nil
}

// remove 删除指定文件，目录不可通过该方法删除。
// 参数：
//   - p: 文件绝对路径
//
// 返回值：
//   - error: 文件不存在或目标为目录时返回错误
//
// 异常：
//   - 无
//
// 使用示例：
//
//	if err := fs.remove("/a/hello.txt"); err != nil { ... }
func (fs *ftpMemFS) remove(p string) error {
	target := fs.clean(p)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	node, ok := fs.nodes[target]
	if !ok {
		return fmt.Errorf("文件不存在: %s", target)
	}
	if node.isDir {
		return fmt.Errorf("目标为目录: %s", target)
	}

	delete(fs.nodes, target)
	return nil
}

// listDir 列出指定目录下的直接子节点，按名称升序返回。
// 参数：
//   - p: 目录绝对路径
//
// 返回值：
//   - []ftpMemNode: 子节点快照列表，目录不存在时返回空列表
//
// 异常：
//   - 无
//
// 使用示例：
//
//	nodes := fs.listDir("/a")
func (fs *ftpMemFS) listDir(p string) []ftpMemNode {
	target := fs.clean(p)

	fs.mu.Lock()
	nodes := make([]ftpMemNode, 0, len(fs.nodes))
	for _, node := range fs.nodes {
		if node.path == "/" || path.Dir(node.path) != target {
			continue
		}
		nodes = append(nodes, *node)
	}
	fs.mu.Unlock()

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].name < nodes[j].name })
	return nodes
}

// ftpTestServer 是一个仅供测试使用的内存 FTP 服务器。
// 它监听回环地址的随机端口，文件全部保存在内存中，不会读写磁盘，也不访问外部网络。
type ftpTestServer struct {
	ln       net.Listener
	addr     string
	fs       *ftpMemFS
	username string
	password string

	mu     sync.Mutex
	conns  map[net.Conn]struct{}
	closed bool
}

// newFTPTestServer 启动一个本地内存 FTP 服务器，并注册 t.Cleanup 自动关闭。
// 参数：
//   - t: 测试对象，用于注册清理逻辑与记录地址
//
// 返回值：
//   - *ftpTestServer: 已启动并可接受连接的服务器实例
//
// 异常：
//   - 监听回环地址失败时调用 t.Fatalf 终止测试
//
// 使用示例：
//
//	srv := newFTPTestServer(t)
//	client, err := NewFTPClient(FTPConfig(jcbaseGo.FTPStruct{Address: srv.Address(), ...}))
func newFTPTestServer(t *testing.T) *ftpTestServer {
	t.Helper()

	ln, err := net.Listen("tcp", net.JoinHostPort(localTestHost, "0"))
	if err != nil {
		t.Fatalf("启动本地 FTP 测试服务器失败: %v", err)
	}

	srv := &ftpTestServer{
		ln:       ln,
		addr:     ln.Addr().String(),
		fs:       newFTPMemFS(),
		username: "test",
		password: "test",
		conns:    make(map[net.Conn]struct{}),
	}
	t.Cleanup(srv.Close)

	go srv.acceptLoop()
	return srv
}

// Address 返回服务器的 host:port 地址，可直接用于 FTPConfig.Address。
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
func (s *ftpTestServer) Address() string {
	return s.addr
}

// Files 返回底层内存文件系统，便于测试直接校验服务端状态。
// 参数：
//   - 无
//
// 返回值：
//   - *ftpMemFS: 内存文件系统实例
//
// 异常：
//   - 无
//
// 使用示例：
//
//	if _, ok := srv.Files().stat("/a/hello.txt"); !ok { ... }
func (s *ftpTestServer) Files() *ftpMemFS {
	return s.fs
}

// Close 关闭监听器与所有已建立的控制连接，可重复调用。
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
func (s *ftpTestServer) Close() {
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

// acceptLoop 循环接受控制连接，每条连接交由独立 goroutine 处理。
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
func (s *ftpTestServer) acceptLoop() {
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

		session := &ftpSession{srv: s, conn: conn}
		go session.run()
	}
}

// ftpSession 表示一条 FTP 控制连接上的会话状态。
type ftpSession struct {
	srv      *ftpTestServer
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	loggedIn bool
	cwd      string
	dataLn   net.Listener
}

// run 处理当前控制连接上的命令，直到客户端断开或发送 QUIT。
// 参数：
//   - 无
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（连接错误时静默结束会话）
//
// 使用示例：
//
//	go session.run()
func (s *ftpSession) run() {
	defer s.close()

	s.reader = bufio.NewReader(s.conn)
	s.writer = bufio.NewWriter(s.conn)
	s.cwd = "/"

	s.reply(220, "jcbaseGo local test FTP server ready")

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return
		}

		cmd, arg := parseFTPCommand(line)
		switch {
		case cmd == "":
			continue
		case cmd == "QUIT":
			s.reply(221, "Goodbye")
			return
		case !s.loggedIn && !isPreLoginCommand(cmd):
			s.reply(530, "Please login with USER and PASS")
		default:
			s.dispatch(cmd, arg)
		}
	}
}

// parseFTPCommand 从一行文本中解析出命令与参数。
// 参数：
//   - line: 未去除换行符的原始输入行
//
// 返回值：
//   - string: 大写化的命令名，空行时返回空字符串
//   - string: 命令参数，无参数时为空字符串
//
// 异常：
//   - 无
//
// 使用示例：
//
//	cmd, arg := parseFTPCommand("STOR /a/hello.txt\r\n") // "STOR", "/a/hello.txt"
func parseFTPCommand(line string) (string, string) {
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) == "" {
		return "", ""
	}

	fields := strings.SplitN(line, " ", 2)
	arg := ""
	if len(fields) > 1 {
		arg = fields[1]
	}
	return strings.ToUpper(fields[0]), arg
}

// isPreLoginCommand 判断命令是否允许在登录前执行。
// 参数：
//   - cmd: 大写化的命令名
//
// 返回值：
//   - bool: 允许在登录前执行时返回 true
//
// 异常：
//   - 无
//
// 使用示例：
//
//	isPreLoginCommand("USER") // true
func isPreLoginCommand(cmd string) bool {
	switch cmd {
	case "USER", "PASS", "FEAT", "SYST", "NOOP":
		return true
	}
	return false
}

// dispatch 按命令名分发到具体的处理逻辑。
// 参数：
//   - cmd: 大写化的命令名
//   - arg: 命令参数
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（未知命令统一返回 502）
//
// 使用示例：
//
//	s.dispatch("STOR", "/a/hello.txt")
func (s *ftpSession) dispatch(cmd, arg string) {
	switch cmd {
	case "USER":
		s.reply(331, "User name okay, need password")
	case "PASS":
		s.handlePass(arg)
	case "FEAT":
		s.replyLines(211, []string{" MLST Type*;Size*;Modify*;Perm*;", " SIZE", " MDTM"}, "End")
	case "TYPE", "MODE", "STRU", "OPTS":
		s.reply(200, "OK")
	case "NOOP":
		s.reply(200, "NOOP ok")
	case "SYST":
		s.reply(215, "UNIX Type: L8")
	case "PWD":
		s.reply(257, strconv.Quote(s.cwd))
	case "CWD":
		s.handleCwd(arg)
	case "MKD":
		s.handleMkd(arg)
	case "DELE":
		s.handleDele(arg)
	case "SIZE":
		s.handleSize(arg)
	case "MDTM":
		s.handleMdtm(arg)
	case "EPSV", "PASV":
		s.handlePassive(cmd)
	case "STOR":
		s.handleStor(arg)
	case "RETR":
		s.handleRetr(arg)
	case "MLSD":
		s.handleMlsd(arg)
	default:
		s.reply(502, "Command not implemented")
	}
}

// handlePass 校验密码并标记会话为已登录。
// 参数：
//   - password: 客户端传入的密码
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（密码错误时回复 530 并结束会话）
//
// 使用示例：
//
//	s.handlePass("test")
func (s *ftpSession) handlePass(password string) {
	if password != s.srv.password {
		s.reply(530, "Login incorrect")
		return
	}
	s.loggedIn = true
	s.reply(230, "User logged in")
}

// handleCwd 切换当前工作目录，目标目录不存在时返回 550。
// 参数：
//   - arg: 目标目录路径，支持相对路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无
//
// 使用示例：
//
//	s.handleCwd("/a/b")
func (s *ftpSession) handleCwd(arg string) {
	node, ok := s.srv.fs.stat(s.resolve(arg))
	if !ok || !node.isDir {
		s.reply(550, "No such directory")
		return
	}
	s.cwd = node.path
	s.reply(250, "Directory changed")
}

// handleMkd 创建单层目录，父目录不存在时返回 550。
// 参数：
//   - arg: 待创建的目录路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无
//
// 使用示例：
//
//	s.handleMkd("/a/b")
func (s *ftpSession) handleMkd(arg string) {
	target := s.resolve(arg)
	if err := s.srv.fs.mkdir(target); err != nil {
		s.reply(550, "Create directory operation failed")
		return
	}
	s.reply(257, strconv.Quote(target)+" created")
}

// handleDele 删除文件，文件不存在时返回 550。
// 参数：
//   - arg: 待删除的文件路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无
//
// 使用示例：
//
//	s.handleDele("/a/hello.txt")
func (s *ftpSession) handleDele(arg string) {
	if err := s.srv.fs.remove(s.resolve(arg)); err != nil {
		s.reply(550, "Delete operation failed")
		return
	}
	s.reply(250, "File deleted")
}

// handleSize 返回文件大小，文件不存在时返回 550。
// 参数：
//   - arg: 文件路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无
//
// 使用示例：
//
//	s.handleSize("/a/hello.txt")
func (s *ftpSession) handleSize(arg string) {
	node, ok := s.srv.fs.stat(s.resolve(arg))
	if !ok {
		s.reply(550, "No such file")
		return
	}
	s.reply(213, strconv.FormatInt(int64(len(node.content)), 10))
}

// handleMdtm 返回文件的最后修改时间，格式为 YYYYMMDDHHMMSS。
// 参数：
//   - arg: 文件路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无
//
// 使用示例：
//
//	s.handleMdtm("/a/hello.txt")
func (s *ftpSession) handleMdtm(arg string) {
	node, ok := s.srv.fs.stat(s.resolve(arg))
	if !ok {
		s.reply(550, "No such file")
		return
	}
	s.reply(213, node.modTime.UTC().Format("20060102150405"))
}

// handlePassive 进入被动模式，回复数据连接监听端口。
// 参数：
//   - cmd: "EPSV" 或 "PASV"
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（监听创建失败时回复 425）
//
// 使用示例：
//
//	s.handlePassive("EPSV")
func (s *ftpSession) handlePassive(cmd string) {
	ln, ok := s.openDataListener()
	if !ok {
		s.reply(425, "Can't open data connection")
		return
	}

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		s.reply(425, "Can't resolve data connection address")
		return
	}

	if cmd == "EPSV" {
		s.reply(229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", addr.Port))
		return
	}

	ip := addr.IP.To4()
	if ip == nil {
		s.reply(425, "Can't resolve data connection address")
		return
	}
	s.reply(227, fmt.Sprintf("Entering Passive Mode (%d,%d,%d,%d,%d,%d)",
		ip[0], ip[1], ip[2], ip[3], addr.Port>>8, addr.Port&0xFF))
}

// handleStor 接收数据连接上的内容并写入内存文件系统。
// 参数：
//   - arg: 目标文件路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（父目录不存在时回复 550，传输失败时回复 451）
//
// 使用示例：
//
//	s.handleStor("/a/hello.txt")
func (s *ftpSession) handleStor(arg string) {
	target := s.resolve(arg)
	if node, ok := s.srv.fs.stat(path.Dir(target)); !ok || !node.isDir {
		s.reply(550, "No such directory")
		return
	}

	s.transfer(func(conn net.Conn) error {
		data, err := io.ReadAll(conn)
		if err != nil {
			return err
		}
		return s.srv.fs.writeFile(target, data)
	})
}

// handleRetr 通过数据连接下发文件内容。
// 参数：
//   - arg: 待下载的文件路径
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（文件不存在时回复 550，传输失败时回复 451）
//
// 使用示例：
//
//	s.handleRetr("/a/hello.txt")
func (s *ftpSession) handleRetr(arg string) {
	node, ok := s.srv.fs.stat(s.resolve(arg))
	if !ok {
		s.reply(550, "No such file")
		return
	}

	content := node.content
	s.transfer(func(conn net.Conn) error {
		_, err := conn.Write(content)
		return err
	})
}

// handleMlsd 通过数据连接下发目录列表，采用 RFC 3659 的机器可读格式。
// 参数：
//   - arg: 待列出的目录路径，为空时列出当前目录
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（传输失败时回复 451）
//
// 使用示例：
//
//	s.handleMlsd("/a")
func (s *ftpSession) handleMlsd(arg string) {
	target := s.resolve(arg)
	nodes := s.srv.fs.listDir(target)

	lines := make([]string, 0, len(nodes))
	for _, node := range nodes {
		entryType := "file"
		if node.isDir {
			entryType = "dir"
		}
		lines = append(lines, fmt.Sprintf("Type=%s;Size=%d;Modify=%s;Perm=rw; %s\r\n",
			entryType, len(node.content), node.modTime.UTC().Format("20060102150405"), node.name))
	}

	s.transfer(func(conn net.Conn) error {
		for _, line := range lines {
			if _, err := conn.Write([]byte(line)); err != nil {
				return err
			}
		}
		return nil
	})
}

// resolve 将客户端传入的路径解析为服务端绝对路径。
// 参数：
//   - p: 原始路径，相对路径基于当前工作目录解析
//
// 返回值：
//   - string: 规范化后的绝对路径
//
// 异常：
//   - 无
//
// 使用示例：
//
//	target := s.resolve("hello.txt")
func (s *ftpSession) resolve(p string) string {
	if p == "" {
		return s.cwd
	}
	if strings.HasPrefix(p, "/") {
		return s.srv.fs.clean(p)
	}
	return s.srv.fs.clean(path.Join(s.cwd, p))
}

// openDataListener 获取被动模式下用于数据传输的监听器，不存在时创建。
// 参数：
//   - 无
//
// 返回值：
//   - net.Listener: 数据连接监听器
//   - bool: 创建成功返回 true，否则返回 false
//
// 异常：
//   - 无
//
// 使用示例：
//
//	ln, ok := s.openDataListener()
func (s *ftpSession) openDataListener() (net.Listener, bool) {
	if s.dataLn != nil {
		return s.dataLn, true
	}

	ln, err := net.Listen("tcp", net.JoinHostPort(localTestHost, "0"))
	if err != nil {
		return nil, false
	}
	s.dataLn = ln
	return ln, true
}

// closeDataListener 关闭当前数据连接监听器。
// 参数：
//   - 无
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（关闭错误被忽略）
//
// 使用示例：
//
//	s.closeDataListener()
func (s *ftpSession) closeDataListener() {
	if s.dataLn == nil {
		return
	}
	_ = s.dataLn.Close()
	s.dataLn = nil
}

// transfer 执行一次完整的数据传输：回复 150、接受连接、执行 fn、回复 226。
// 参数：
//   - fn: 在已建立的数据连接上执行的传输逻辑
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（fn 返回错误时回复 451）
//
// 使用示例：
//
//	s.transfer(func(conn net.Conn) error { _, err := conn.Write(data); return err })
func (s *ftpSession) transfer(fn func(net.Conn) error) {
	ln, ok := s.openDataListener()
	if !ok {
		s.reply(425, "Can't open data connection")
		return
	}

	s.reply(150, "Opening data connection")

	conn, err := ln.Accept()
	if err != nil {
		s.reply(426, "Data connection failed")
		return
	}

	err = fn(conn)
	_ = conn.Close()
	s.closeDataListener()

	if err != nil {
		s.reply(451, "Transfer aborted: "+err.Error())
		return
	}
	s.reply(226, "Transfer complete")
}

// reply 向客户端发送单行响应。
// 参数：
//   - code: FTP 状态码
//   - msg: 响应文本
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（写入失败被忽略，由上层读取错误时发现）
//
// 使用示例：
//
//	s.reply(200, "OK")
func (s *ftpSession) reply(code int, msg string) {
	_, _ = fmt.Fprintf(s.writer, "%d %s\r\n", code, msg)
	_ = s.writer.Flush()
}

// replyLines 向客户端发送多行响应。
// 参数：
//   - code: FTP 状态码
//   - lines: 中间行内容，需自带前导空格
//   - end: 结束行的文本
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（写入失败被忽略）
//
// 使用示例：
//
//	s.replyLines(211, []string{" MLST"}, "End")
func (s *ftpSession) replyLines(code int, lines []string, end string) {
	for _, line := range lines {
		_, _ = fmt.Fprintf(s.writer, "%d-%s\r\n", code, line)
	}
	_, _ = fmt.Fprintf(s.writer, "%d %s\r\n", code, end)
	_ = s.writer.Flush()
}

// close 释放会话占用的数据监听与控制连接。
// 参数：
//   - 无
//
// 返回值：
//   - 无
//
// 异常：
//   - 无（关闭错误被忽略）
//
// 使用示例：
//
//	defer s.close()
func (s *ftpSession) close() {
	s.closeDataListener()
	_ = s.conn.Close()
}
