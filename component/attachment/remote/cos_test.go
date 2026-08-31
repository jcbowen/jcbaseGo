package remote

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// newTestCOSClient 构造用于本地测试的 COS 客户端。
// 客户端仅用于签名计算与结构校验，测试用例不会发起任何网络请求。
// 参数：
//   - t *testing.T: 测试上下文
//
// 返回值：
//   - *COSClient: 已初始化的 COS 客户端
//
// 使用示例：
//
//	client := newTestCOSClient(t)
//	defer func() { _ = client.Close() }()
func newTestCOSClient(t *testing.T) *COSClient {
	t.Helper()

	client, err := NewCOSClient(COSConfig(jcbaseGo.COSStruct{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Region:    "ap-guangzhou",
		Bucket:    "test-bucket-1250000000",
		Url:       "https://test-bucket-1250000000.cos.ap-guangzhou.myqcloud.com",
	}))
	if err != nil {
		t.Fatalf("NewCOSClient() error = %v", err)
	}

	return client
}

// TestCOSClient_PresignUpload 验证 COSClient 能够生成预签名上传 URL。
func TestCOSClient_PresignUpload(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	// 断言 COSClient 实现了 PresignUploader 与 ObjectExister 接口
	if _, ok := interface{}(client).(PresignUploader); !ok {
		t.Fatalf("COSClient does not implement PresignUploader")
	}
	if _, ok := interface{}(client).(ObjectExister); !ok {
		t.Fatalf("COSClient does not implement ObjectExister")
	}

	url, headers, err := client.PresignUpload(context.Background(), "images/test.jpg", &PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if url == "" {
		t.Errorf("PresignUpload() url is empty")
	}

	// 预签名 URL 中应包含桶域名和对象路径
	if !strings.Contains(url, "test-bucket-1250000000") {
		t.Errorf("PresignUpload() url does not contain bucket name")
	}
	if !strings.Contains(url, "images/test.jpg") {
		t.Errorf("PresignUpload() url does not contain object key")
	}

	// 预签名 URL 应携带签名信息
	if !strings.Contains(url, "q-signature=") {
		t.Errorf("PresignUpload() url does not contain signature")
	}

	// 未签名额外请求头时，headers 应为空 map 而非 nil
	if headers == nil {
		t.Errorf("PresignUpload() headers is nil")
	}
	if len(headers) != 0 {
		t.Errorf("PresignUpload() headers = %v, want empty", headers)
	}
}

// TestCOSClient_PresignUpload_WithHeaders 验证指定 ContentType 与 Metadata 时签名头正确返回。
func TestCOSClient_PresignUpload_WithHeaders(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	opts := &PresignOptions{
		Expires:     10 * time.Minute,
		ContentType: "image/jpeg",
		Metadata: map[string]string{
			"x-cos-meta-uid": "12345",
		},
	}
	url, headers, err := client.PresignUpload(context.Background(), "images/test.jpg", opts)
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if url == "" {
		t.Errorf("PresignUpload() url is empty")
	}

	if headers["Content-Type"] != "image/jpeg" {
		t.Errorf("PresignUpload() Content-Type header = %q, want %q", headers["Content-Type"], "image/jpeg")
	}
	if headers["x-cos-meta-uid"] != "12345" {
		t.Errorf("PresignUpload() x-cos-meta-uid header = %q, want %q", headers["x-cos-meta-uid"], "12345")
	}

	// 参与签名的请求头名应出现在签名的 header-list 中
	if !strings.Contains(url, "content-type") {
		t.Errorf("PresignUpload() url does not sign content-type header")
	}
	if !strings.Contains(url, "x-cos-meta-uid") {
		t.Errorf("PresignUpload() url does not sign x-cos-meta-uid header")
	}
}

// TestCOSClient_PresignUpload_DefaultExpires 验证未指定有效期时使用默认有效期。
func TestCOSClient_PresignUpload_DefaultExpires(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	presignedURL, _, err := client.PresignUpload(context.Background(), "images/test.jpg", nil)
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	// 从签名参数中解析签名有效期的起止时间戳
	parsedURL, err := url.Parse(presignedURL)
	if err != nil {
		t.Fatalf("parse presigned url error = %v", err)
	}

	signTime := parsedURL.Query().Get("q-sign-time")
	start, end, found := strings.Cut(signTime, ";")
	if !found {
		t.Fatalf("PresignUpload() q-sign-time = %q, want start;end", signTime)
	}

	startUnix, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		t.Fatalf("parse sign start time error = %v", err)
	}
	endUnix, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		t.Fatalf("parse sign end time error = %v", err)
	}

	if got := time.Duration(endUnix-startUnix) * time.Second; got != defaultPresignExpires {
		t.Errorf("PresignUpload() expires = %v, want %v", got, defaultPresignExpires)
	}
}

// TestCOSClient_PresignUpload_ContextCanceled 验证上下文取消时返回错误。
func TestCOSClient_PresignUpload_ContextCanceled(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.PresignUpload(ctx, "images/test.jpg", &PresignOptions{Expires: 10 * time.Minute})
	if err == nil {
		t.Fatalf("PresignUpload() expected error when context canceled, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("PresignUpload() error = %v, want context canceled", err)
	}
}

// TestCOSClient_Exists_ContextCanceled 验证上下文取消时 Exists 返回错误而非发起请求。
func TestCOSClient_Exists_ContextCanceled(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exists, err := client.Exists(ctx, "images/test.jpg")
	if err == nil {
		t.Fatalf("Exists() expected error when context canceled, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Exists() error = %v, want context canceled", err)
	}
	if exists {
		t.Errorf("Exists() = true, want false when error occurred")
	}
}
