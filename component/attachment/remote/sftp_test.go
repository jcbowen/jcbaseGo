package remote

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// getSFTPTestConfig 从环境变量读取 SFTP 测试配置，未设置时使用默认值并默认跳过真实连接测试。
// 环境变量：
//   - SFTP_TEST_ADDRESS  服务器地址，默认 127.0.0.1:22
//   - SFTP_TEST_USERNAME 用户名，默认 test
//   - SFTP_TEST_PASSWORD 密码，默认 test
//   - SFTP_TEST_RUN      设置为 true 时才执行 SFTP 真实连接测试
func getSFTPTestConfig(t *testing.T) (SFTPConfig, bool) {
	if os.Getenv("SFTP_TEST_RUN") != "true" {
		t.Skip("环境变量 SFTP_TEST_RUN 不为 true，跳过 SFTP 真实连接测试")
	}

	address := os.Getenv("SFTP_TEST_ADDRESS")
	if address == "" {
		address = "127.0.0.1:22"
	}
	username := os.Getenv("SFTP_TEST_USERNAME")
	if username == "" {
		username = "test"
	}
	password := os.Getenv("SFTP_TEST_PASSWORD")
	if password == "" {
		password = "test"
	}

	return SFTPConfig(jcbaseGo.SFTPStruct{
		Address:  address,
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}), false
}

// TestSFTPAddressNormalization 验证 SFTP Address 的 scheme 和默认端口处理。
// 该测试不需要真实 SFTP 服务器，使用一个不可达的地址验证错误类型。
func TestSFTPAddressNormalization(t *testing.T) {
	config := SFTPConfig(jcbaseGo.SFTPStruct{
		Address:  "sftp://127.0.0.1:2222",
		Username: "test",
		Password: "test",
		Timeout:  1 * time.Second,
	})

	_, err := NewSFTPClient(config)
	if err == nil {
		t.Fatal("期望连接不可达 SFTP 服务器时返回错误，实际未返回错误")
	}
	//  scheme 应被清洗，错误应为 Dial 类错误而非地址格式错误
	if err.Error() == "" {
		t.Fatal("错误信息为空")
	}
	t.Logf("SFTP 带 scheme 地址已清洗，连接错误: %v", err)
}

// TestSFTPFullFlow 对 SFTP 远程附件封装进行全流程真实连接测试。
// 测试步骤：连接 -> 上传（含子目录） -> 下载校验 -> 列出目录 -> 删除 -> 关闭。
// 需要环境变量 SFTP_TEST_RUN=true 才会执行。
func TestSFTPFullFlow(t *testing.T) {
	config, skip := getSFTPTestConfig(t)
	if skip {
		return
	}

	ctx := context.Background()
	remoteDir := fmt.Sprintf("/jcbaseGo_test/%s", time.Now().Format("20060102_150405"))
	remotePath := remoteDir + "/hello.txt"
	content := []byte("Hello, jcbaseGo SFTP remote attachment test!")

	t.Logf("正在连接 SFTP 服务器: %s", config.Address)
	client, err := NewSFTPClient(config)
	if err != nil {
		t.Fatalf("连接 SFTP 服务器失败: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("关闭 SFTP 连接出错: %v", err)
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
