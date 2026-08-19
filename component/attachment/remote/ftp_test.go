package remote

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// getFTPTestConfig 从环境变量读取 FTP 测试配置，未设置时使用默认值。
// 环境变量：
//   - FTP_TEST_ADDRESS  服务器地址，默认 192.168.10.132:21
//   - FTP_TEST_USERNAME 用户名，默认 test
//   - FTP_TEST_PASSWORD 密码，默认 1tpHAGsMk2Dr
//   - FTP_TEST_SKIP     设置为 true 时跳过 FTP 真实连接测试
func getFTPTestConfig(t *testing.T) (FTPConfig, bool) {
	if os.Getenv("FTP_TEST_SKIP") == "true" {
		t.Skip("环境变量 FTP_TEST_SKIP=true，跳过 FTP 真实连接测试")
	}

	address := os.Getenv("FTP_TEST_ADDRESS")
	if address == "" {
		address = "192.168.10.132:21"
	}
	username := os.Getenv("FTP_TEST_USERNAME")
	if username == "" {
		username = "test"
	}
	password := os.Getenv("FTP_TEST_PASSWORD")
	if password == "" {
		password = "1tpHAGsMk2Dr"
	}

	return FTPConfig(jcbaseGo.FTPStruct{
		Address:  address,
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}), false
}

// TestFTPFullFlow 对 FTP 远程附件封装进行全流程真实连接测试。
// 测试步骤：连接 -> 上传（含子目录） -> 下载校验 -> 列出目录 -> 删除 -> 关闭。
func TestFTPFullFlow(t *testing.T) {
	config, skip := getFTPTestConfig(t)
	if skip {
		return
	}

	ctx := context.Background()
	remoteDir := fmt.Sprintf("/jcbaseGo_test/%s", time.Now().Format("20060102_150405"))
	remotePath := remoteDir + "/hello.txt"
	content := []byte("Hello, jcbaseGo FTP remote attachment test!")

	t.Logf("正在连接 FTP 服务器: %s", config.Address)
	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("连接 FTP 服务器失败: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("关闭 FTP 连接出错: %v", err)
		}
	}()

	t.Logf("开始上传文件到: %s", remotePath)
	if err := client.Upload(ctx, remotePath, content); err != nil {
		t.Fatalf("上传文件失败: %v", err)
	}
	t.Log("文件上传成功")

	t.Logf("开始下载文件: %s", remotePath)
	downloaded, err := client.Download(ctx, remotePath)
	if err != nil {
		t.Fatalf("下载文件失败: %v", err)
	}
	if string(downloaded) != string(content) {
		t.Fatalf("下载内容不匹配，期望 %q，实际 %q", string(content), string(downloaded))
	}
	t.Log("文件下载成功，内容校验通过")

	t.Logf("开始列出目录: %s", remoteDir)
	listResult, err := client.List(ctx, ListOptions{Prefix: remoteDir, MaxKeys: 10})
	if err != nil {
		t.Fatalf("列出目录失败: %v", err)
	}
	found := false
	for _, f := range listResult.Files {
		t.Logf("  文件: name=%s, size=%d, isDir=%v", f.Name, f.Size, f.IsDir)
		if f.Name == "hello.txt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("在目录 %s 中未找到预期文件 hello.txt", remoteDir)
	}
	t.Log("目录列出成功")

	t.Logf("开始删除文件: %s", remotePath)
	if err := client.Delete(ctx, remotePath); err != nil {
		t.Fatalf("删除文件失败: %v", err)
	}
	t.Log("文件删除成功")
}

// TestFTPAddressWithScheme 验证 Address 带有 ftp:// scheme 时能够被自动清洗并正常连接。
func TestFTPAddressWithScheme(t *testing.T) {
	config, skip := getFTPTestConfig(t)
	if skip {
		return
	}
	config.Address = "ftp://192.168.10.132:21"

	client, err := NewFTPClient(config)
	if err != nil {
		t.Fatalf("传入带 scheme 的地址应被清洗后成功连接，实际失败: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("关闭 FTP 连接出错: %v", err)
		}
	}()
	t.Log("带 scheme 的 FTP 地址已正确清洗并连接")
}
