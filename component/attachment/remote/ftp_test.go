package remote

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// newLocalFTPConfig 构造一份指向本地内存 FTP 服务器的测试配置。
// 参数：
//   - t: 测试对象
//
// 返回值：
//   - FTPConfig: 已填充本地服务器地址与账号的配置
//   - *ftpTestServer: 本次测试专用的本地服务器，随测试结束自动关闭
//
// 异常：
//   - 服务器启动失败时由 newFTPTestServer 调用 t.Fatalf 终止测试
//
// 使用示例：
//
//	config, srv := newLocalFTPConfig(t)
func newLocalFTPConfig(t *testing.T) (FTPConfig, *ftpTestServer) {
	t.Helper()

	srv := newFTPTestServer(t)
	return FTPConfig(jcbaseGo.FTPStruct{
		Address:  srv.Address(),
		Username: "test",
		Password: "test",
		Timeout:  10 * time.Second,
	}), srv
}

// TestFTPFullFlow 对 FTP 远程附件封装进行全流程测试，全部在本地内存服务器上完成。
// 覆盖：连接 -> 上传（自动创建多级父目录） -> 下载校验 -> 列出目录 -> 删除 -> 关闭。
func TestFTPFullFlow(t *testing.T) {
	config, srv := newLocalFTPConfig(t)
	ctx := context.Background()
	remoteDir := "/jcbaseGo_test/" + time.Now().Format("20060102_150405")
	remotePath := remoteDir + "/hello.txt"
	content := []byte("Hello, jcbaseGo FTP remote attachment test!")

	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 FTP 服务器失败: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("关闭 FTP 客户端失败: %v", err)
		}
	}()

	// 上传：父目录不存在，应由客户端自动递归创建
	if err := client.Upload(ctx, remotePath, content); err != nil {
		t.Fatalf("上传文件失败: %v", err)
	}
	node, ok := srv.Files().stat(remotePath)
	if !ok {
		t.Fatalf("服务端未找到刚上传的文件: %s", remotePath)
	}
	if string(node.content) != string(content) {
		t.Fatalf("服务端文件内容不匹配，期望 %q，实际 %q", string(content), string(node.content))
	}

	// 下载并校验内容
	downloaded, err := client.Download(ctx, remotePath)
	if err != nil {
		t.Fatalf("下载文件失败: %v", err)
	}
	if string(downloaded) != string(content) {
		t.Fatalf("下载内容不匹配，期望 %q，实际 %q", string(content), string(downloaded))
	}

	// 列出目录并确认文件元数据正确
	listResult, err := client.List(ctx, ListOptions{Prefix: remoteDir, MaxKeys: 10})
	if err != nil {
		t.Fatalf("列出目录失败: %v", err)
	}
	var found bool
	for _, f := range listResult.Files {
		if f.Name != "hello.txt" {
			continue
		}
		found = true
		if f.Size != int64(len(content)) {
			t.Errorf("文件大小 = %d，期望 %d", f.Size, len(content))
		}
		if f.IsDir {
			t.Errorf("hello.txt 应为文件，实际为目录")
		}
	}
	if !found {
		t.Fatalf("目录 %s 中未找到预期文件 hello.txt", remoteDir)
	}

	// 删除后服务端不应再保留该文件
	if err := client.Delete(ctx, remotePath); err != nil {
		t.Fatalf("删除文件失败: %v", err)
	}
	if _, ok := srv.Files().stat(remotePath); ok {
		t.Errorf("删除后服务端仍存在文件: %s", remotePath)
	}
}

// TestFTPAddressWithScheme 验证 Address 带有 ftp:// scheme 时能够被自动清洗并正常连接。
func TestFTPAddressWithScheme(t *testing.T) {
	config, _ := newLocalFTPConfig(t)
	config.Address = "ftp://" + config.Address

	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("传入带 scheme 的地址应被清洗后成功连接，实际失败: %v", err)
	}
	defer func() { _ = client.Close() }()
}

// TestFTPWrongPassword 验证密码错误时返回错误，不会建立可用客户端。
func TestFTPWrongPassword(t *testing.T) {
	config, _ := newLocalFTPConfig(t)
	config.Password = "wrong-password"

	client, err := NewFTPClient(config)
	if err == nil {
		_ = client.Close()
		t.Fatalf("密码错误时 NewFTPClient 应返回错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "Login") {
		t.Errorf("错误信息应包含 Login 操作名，实际: %v", err)
	}
}

// TestFTPDownloadNotExist 验证下载不存在的文件时返回错误。
func TestFTPDownloadNotExist(t *testing.T) {
	config, _ := newLocalFTPConfig(t)

	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 FTP 服务器失败: %v", err)
	}
	defer func() { _ = client.Close() }()

	if _, err := client.Download(context.Background(), "/not-exist.txt"); err == nil {
		t.Fatalf("下载不存在的文件应返回错误，实际为 nil")
	}
}

// TestFTPListPaging 验证 List 在 MaxKeys 限制下按文件名顺序分页返回。
func TestFTPListPaging(t *testing.T) {
	config, _ := newLocalFTPConfig(t)
	ctx := context.Background()
	remoteDir := "/jcbaseGo_paging"

	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 FTP 服务器失败: %v", err)
	}
	defer func() { _ = client.Close() }()

	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := client.Upload(ctx, remoteDir+"/"+name, []byte(name)); err != nil {
			t.Fatalf("上传 %s 失败: %v", name, err)
		}
	}

	first, err := client.List(ctx, ListOptions{Prefix: remoteDir, MaxKeys: 2})
	if err != nil {
		t.Fatalf("第一页列表失败: %v", err)
	}
	if len(first.Files) != 2 {
		t.Fatalf("第一页文件数 = %d，期望 2", len(first.Files))
	}
	if !first.IsTruncated {
		t.Fatalf("第一页应存在更多数据，IsTruncated = false")
	}
	if first.NextMarker != "b.txt" {
		t.Fatalf("NextMarker = %q，期望 %q", first.NextMarker, "b.txt")
	}

	second, err := client.List(ctx, ListOptions{Prefix: remoteDir, Marker: first.NextMarker, MaxKeys: 2})
	if err != nil {
		t.Fatalf("第二页列表失败: %v", err)
	}
	if len(second.Files) != 1 || second.Files[0].Name != "c.txt" {
		t.Fatalf("第二页内容不符合预期: %+v", second.Files)
	}
	if second.IsTruncated {
		t.Fatalf("第二页不应存在更多数据，IsTruncated = true")
	}
}

// TestFTPOperationsAfterClose 验证客户端关闭后再执行操作会返回错误。
func TestFTPOperationsAfterClose(t *testing.T) {
	config, _ := newLocalFTPConfig(t)

	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 FTP 服务器失败: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("关闭 FTP 客户端失败: %v", err)
	}

	ctx := context.Background()
	if err := client.Upload(ctx, "/closed.txt", []byte("data")); err == nil {
		t.Errorf("关闭后 Upload 应返回错误，实际为 nil")
	}
	if _, err := client.Download(ctx, "/closed.txt"); err == nil {
		t.Errorf("关闭后 Download 应返回错误，实际为 nil")
	}
	if _, err := client.List(ctx, ListOptions{Prefix: "/"}); err == nil {
		t.Errorf("关闭后 List 应返回错误，实际为 nil")
	}
	if err := client.Delete(ctx, "/closed.txt"); err == nil {
		t.Errorf("关闭后 Delete 应返回错误，实际为 nil")
	}
}
