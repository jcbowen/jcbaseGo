package remote

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// newLocalSFTPConfig 构造一份指向本地内存 SFTP 服务器的测试配置。
// 参数：
//   - t: 测试对象
//
// 返回值：
//   - SFTPConfig: 已填充本地服务器地址与账号的配置
//   - *sftpTestServer: 本次测试专用的本地服务器，随测试结束自动关闭
//
// 异常：
//   - 服务器启动失败时由 newSFTPTestServer 调用 t.Fatalf 终止测试
//
// 使用示例：
//
//	config, srv := newLocalSFTPConfig(t)
func newLocalSFTPConfig(t *testing.T) (SFTPConfig, *sftpTestServer) {
	t.Helper()

	srv := newSFTPTestServer(t)
	return SFTPConfig(jcbaseGo.SFTPStruct{
		Address:  srv.Address(),
		Username: "test",
		Password: "test",
		Timeout:  10 * time.Second,
	}), srv
}

// TestSFTPAddressNormalization 以表驱动方式验证 SFTP/FTP 地址的归一化处理。
// 该用例为纯逻辑校验，不发起任何网络连接。
func TestSFTPAddressNormalization(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		defaultPort string
		want        string
		wantErr     bool
	}{
		{"主机与端口", "127.0.0.1:2222", "22", "127.0.0.1:2222", false},
		{"缺端口时补默认端口", "127.0.0.1", "22", "127.0.0.1:22", false},
		{"去除 sftp scheme", "sftp://127.0.0.1:2222", "22", "127.0.0.1:2222", false},
		{"去除 ftp scheme", "ftp://example.com:2121", "21", "example.com:2121", false},
		{"去除 scheme 后补端口", "sftp://example.com", "22", "example.com:22", false},
		{"空地址", "", "22", "", true},
		{"仅空白字符", "   ", "22", "", true},
		{"仅有 scheme", "sftp://", "22", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAddress(tt.address, tt.defaultPort)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("normalizeAddress(%q) = %q, want %q", tt.address, got, tt.want)
			}
		})
	}
}

// TestSFTPFullFlow 对 SFTP 远程附件封装进行全流程测试，全部在本地内存服务器上完成。
// 覆盖：连接 -> 上传（自动创建多级父目录） -> 下载校验 -> 列出目录 -> 删除 -> 关闭。
func TestSFTPFullFlow(t *testing.T) {
	config, _ := newLocalSFTPConfig(t)
	ctx := context.Background()
	remoteDir := "/jcbaseGo_test/" + time.Now().Format("20060102_150405")
	remotePath := remoteDir + "/hello.txt"
	content := []byte("Hello, jcbaseGo SFTP remote attachment test!")

	client, err := NewSFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 SFTP 服务器失败: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("关闭 SFTP 客户端失败: %v", err)
		}
	}()

	// 上传：父目录不存在，应由客户端自动递归创建
	if err := client.Upload(ctx, remotePath, content); err != nil {
		t.Fatalf("上传文件失败: %v", err)
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

	// 删除后再次下载应失败
	if err := client.Delete(ctx, remotePath); err != nil {
		t.Fatalf("删除文件失败: %v", err)
	}
	if _, err := client.Download(ctx, remotePath); err == nil {
		t.Errorf("删除后下载仍成功，不符合预期")
	}
}

// TestSFTPAddressWithScheme 验证 Address 带有 sftp:// scheme 时能够被自动清洗并正常连接。
func TestSFTPAddressWithScheme(t *testing.T) {
	config, _ := newLocalSFTPConfig(t)
	config.Address = "sftp://" + config.Address

	client, err := NewSFTPClient(config)
	if err != nil {
		t.Fatalf("传入带 scheme 的地址应被清洗后成功连接，实际失败: %v", err)
	}
	defer func() { _ = client.Close() }()
}

// TestSFTPWrongPassword 验证密码错误时返回错误，不会建立可用客户端。
func TestSFTPWrongPassword(t *testing.T) {
	config, _ := newLocalSFTPConfig(t)
	config.Password = "wrong-password"

	client, err := NewSFTPClient(config)
	if err == nil {
		_ = client.Close()
		t.Fatalf("密码错误时 NewSFTPClient 应返回错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "Dial") {
		t.Errorf("错误信息应包含 Dial 操作名，实际: %v", err)
	}
}

// TestSFTPNoAuthMethod 验证缺少密码与私钥时直接返回错误，不发起连接。
func TestSFTPNoAuthMethod(t *testing.T) {
	config, _ := newLocalSFTPConfig(t)
	config.Password = ""

	client, err := NewSFTPClient(config)
	if err == nil {
		_ = client.Close()
		t.Fatalf("缺少认证方式时应返回错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "认证方式") {
		t.Errorf("错误信息应提示缺少认证方式，实际: %v", err)
	}
}

// TestSFTPListPaging 验证 List 在 MaxKeys 限制下按文件名顺序分页返回。
func TestSFTPListPaging(t *testing.T) {
	config, _ := newLocalSFTPConfig(t)
	ctx := context.Background()
	remoteDir := "/jcbaseGo_paging"

	client, err := NewSFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 SFTP 服务器失败: %v", err)
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

	second, err := client.List(ctx, ListOptions{Prefix: remoteDir, Marker: first.NextMarker, MaxKeys: 2})
	if err != nil {
		t.Fatalf("第二页列表失败: %v", err)
	}
	if len(second.Files) != 1 || second.Files[0].Name != "c.txt" {
		t.Fatalf("第二页内容不符合预期: %+v", second.Files)
	}
}

// TestSFTPOperationsAfterClose 验证客户端关闭后再执行操作会返回错误。
func TestSFTPOperationsAfterClose(t *testing.T) {
	config, _ := newLocalSFTPConfig(t)

	client, err := NewSFTPClient(config)
	if err != nil {
		t.Fatalf("连接本地 SFTP 服务器失败: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("关闭 SFTP 客户端失败: %v", err)
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
