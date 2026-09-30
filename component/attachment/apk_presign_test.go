package attachment

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo/component/attachment/remote"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// TestApkPresignUpload 端到端验证：apk 文件经 zip 类型白名单校验后，
// 通过预签名 URL 直传 OSS 的完整链路（模拟 ssjk 预签名上传的对象路径规则）。
//
// 需要真实 OSS 配置，通过环境变量提供，缺失时自动跳过，不影响常规单测：
//	OSS_TEST_AK / OSS_TEST_SK / OSS_TEST_ENDPOINT / OSS_TEST_BUCKET
//
// 运行示例：
//
//	OSS_TEST_AK=xxx OSS_TEST_SK=xxx OSS_TEST_ENDPOINT=oss-cn-shanghai.aliyuncs.com \
//	OSS_TEST_BUCKET=ssjk2 go test ./component/attachment -run TestApkPresignUpload -v
func TestApkPresignUpload(t *testing.T) {
	ak := os.Getenv("OSS_TEST_AK")
	sk := os.Getenv("OSS_TEST_SK")
	endpoint := os.Getenv("OSS_TEST_ENDPOINT")
	bucket := os.Getenv("OSS_TEST_BUCKET")
	if ak == "" || sk == "" || endpoint == "" || bucket == "" {
		t.Skip("缺少 OSS_TEST_* 环境变量，跳过真实 OSS 集成测试")
	}

	// 1. 类型白名单：apk 必须命中 zip 类型的 AllowExt
	typeInfo, ok := Types["zip"]
	if !ok {
		t.Fatal("Types 中不存在 zip 类型")
	}
	found := false
	for _, ext := range typeInfo.AllowExt {
		if ext == ".apk" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("zip 白名单未包含 .apk，当前为: %v", typeInfo.AllowExt)
	}
	t.Logf("白名单校验通过：zip 类型允许 .apk（当前白名单: %v）", typeInfo.AllowExt)

	// 2. 构造一个合法 zip 结构的文件（apk 本质是 zip），命名为 .apk
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entry, err := zw.Create("AndroidManifest.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = entry.Write([]byte("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<manifest/>\n")); err != nil {
		t.Fatal(err)
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	apkBytes := buf.Bytes()
	t.Logf("构造测试 apk：%d 字节（zip 结构，含 AndroidManifest.xml）", len(apkBytes))

	// 3. 初始化 OSS 客户端，按 ssjk 预签名的对象路径规则签发 URL：zips/{Y/m}/{日时分秒}{随机}{.apk}
	client, err := remote.NewOSSClient(remote.OSSConfig{
		AccessKeyId:     ak,
		AccessKeySecret: sk,
		Endpoint:        endpoint,
		BucketName:      bucket,
	})
	if err != nil {
		t.Fatalf("创建 OSS 客户端失败: %v", err)
	}

	// 对象路径模拟 ssjk 真实规则：{type}s/{Y/m}/{日+时分秒}{随机串}{.ext}。
	// 注意：文件名前缀「日(d)与时分秒(His)」是有意拼接、无分隔符，
	// 与 PHP File.php::file_random_name 的 date('dHis') 及本库 fileRandomName 的 02150405 保持一致，
	// 此处不可在日和时间之间加 "/"，否则会多出目录层级、偏离真实路径规则。
	now := time.Now()
	objectName := fmt.Sprintf("zips/%s/%s%s.apk", now.Format("2006/01"), now.Format("02150405"), helper.Random(6))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const contentType = "application/vnd.android.package-archive"
	presignURL, headers, err := client.PresignUpload(ctx, objectName, &remote.PresignOptions{
		Expires:     10 * time.Minute,
		ContentType: contentType,
	})
	if err != nil {
		t.Fatalf("签发预签名URL失败: %v", err)
	}
	t.Logf("预签名URL已签发，对象路径: %s", objectName)

	// 4. PUT 直传（模拟客户端：原样携带签名返回的 headers，body 为文件二进制）
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presignURL, bytes.NewReader(apkBytes))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT 直传请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT 直传被拒绝，HTTP %d: %s", resp.StatusCode, string(body))
	}
	t.Logf("PUT 直传成功，HTTP %d，Content-Type=%s", resp.StatusCode, req.Header.Get("Content-Type"))

	// 5. HEAD 探测对象真实存在
	exists, err := client.Exists(ctx, objectName)
	if err != nil {
		t.Fatalf("探测对象失败: %v", err)
	}
	if !exists {
		t.Fatal("上传后对象不存在")
	}
	t.Log("HEAD 探测：对象已存在于 OSS")

	// 6. 清理测试对象，不遗留脏数据
	if err = client.Delete(ctx, objectName); err != nil {
		t.Fatalf("清理测试对象失败: %v", err)
	}
	t.Log("测试对象已清理，链路验证完成：apk 经 zip 类型预签名直传 OSS 可行")
}
