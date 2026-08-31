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
	if headers["X-Cos-Meta-Uid"] != "12345" {
		t.Errorf("PresignUpload() X-Cos-Meta-Uid header = %q, want %q", headers["X-Cos-Meta-Uid"], "12345")
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

// TestCOSClient_PresignUpload_Token 验证临时密钥会追加到预签名 URL 中。
func TestCOSClient_PresignUpload_Token(t *testing.T) {
	client, err := NewCOSClient(COSConfig(jcbaseGo.COSStruct{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Token:     "test-session-token",
		Region:    "ap-guangzhou",
		Bucket:    "test-bucket-1250000000",
		Url:       "https://test-bucket-1250000000.cos.ap-guangzhou.myqcloud.com",
	}))
	if err != nil {
		t.Fatalf("NewCOSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	url, _, err := client.PresignUpload(context.Background(), "images/test.jpg", &PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if !strings.Contains(url, "x-cos-security-token=test-session-token") {
		t.Errorf("PresignUpload() url does not contain session token")
	}
}

// TestCOSClient_PresignUpload_AutoMetaPrefix 验证自定义元数据缺少 x-cos-meta- 前缀时会自动补齐。
func TestCOSClient_PresignUpload_AutoMetaPrefix(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	opts := &PresignOptions{
		Expires: 10 * time.Minute,
		Metadata: map[string]string{
			"uid": "12345",
		},
	}
	url, headers, err := client.PresignUpload(context.Background(), "images/test.jpg", opts)
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if headers["X-Cos-Meta-Uid"] != "12345" {
		t.Errorf("PresignUpload() header = %v, want X-Cos-Meta-Uid=12345", headers)
	}
	if !strings.Contains(url, "x-cos-meta-uid") {
		t.Errorf("PresignUpload() url does not sign x-cos-meta-uid header")
	}
}

// TestCOSClient_NewCOSClient_AutoURL 验证 Url 为空时可使用 Bucket 与 Region 自动构造。
func TestCOSClient_NewCOSClient_AutoURL(t *testing.T) {
	client, err := NewCOSClient(COSConfig(jcbaseGo.COSStruct{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Region:    "ap-guangzhou",
		Bucket:    "test-bucket-1250000000",
	}))
	if err != nil {
		t.Fatalf("NewCOSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	url, _, err := client.PresignUpload(context.Background(), "images/test.jpg", &PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if !strings.HasPrefix(url, "https://test-bucket-1250000000.cos.ap-guangzhou.myqcloud.com/") {
		t.Errorf("PresignUpload() url = %q, want auto constructed bucket url", url)
	}
}

// TestCOSClient_NewCOSClient_MissingURLAndRegion 验证 Url 为空且缺少 Bucket 或 Region 时报错。
func TestCOSClient_NewCOSClient_MissingURLAndRegion(t *testing.T) {
	_, err := NewCOSClient(COSConfig(jcbaseGo.COSStruct{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Bucket:    "test-bucket-1250000000",
	}))
	if err == nil {
		t.Fatalf("NewCOSClient() expected error when url and region are empty, got nil")
	}
}

// TestCOSClient_ContextCanceled 验证上下文取消时 Upload/Download/Delete/List 均返回错误。
func TestCOSClient_ContextCanceled(t *testing.T) {
	client := newTestCOSClient(t)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name string
		op   func() error
	}{
		{
			name: "Upload",
			op: func() error {
				return client.Upload(ctx, "canceled/upload.txt", []byte("test"))
			},
		},
		{
			name: "Download",
			op: func() error {
				_, err := client.Download(ctx, "canceled/download.txt")
				return err
			},
		},
		{
			name: "Delete",
			op: func() error {
				return client.Delete(ctx, "canceled/delete.txt")
			},
		},
		{
			name: "List",
			op: func() error {
				_, err := client.List(ctx, ListOptions{})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.op()
			if err == nil {
				t.Fatalf("%s() expected error when context canceled, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), "context canceled") {
				t.Errorf("%s() error = %v, want context canceled", tc.name, err)
			}
		})
	}
}
