package remote

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/jcbowen/jcbaseGo"
)

// TestOSSClient_PresignUpload 验证 OSSClient 能够生成预签名上传 URL。
func TestOSSClient_PresignUpload(t *testing.T) {
	config := OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	})

	client, err := NewOSSClient(config)
	if err != nil {
		t.Fatalf("NewOSSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// 断言 OSSClient 实现了 PresignUploader 接口
	_, ok := interface{}(client).(PresignUploader)
	if !ok {
		t.Fatalf("OSSClient does not implement PresignUploader")
	}

	ctx := context.Background()
	url, headers, err := client.PresignUpload(ctx, "images/test.jpg", &PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if url == "" {
		t.Errorf("PresignUpload() url is empty")
	}

	// 预签名 URL 中应包含桶名、Endpoint 和对象路径
	if !strings.Contains(url, "test-bucket") {
		t.Errorf("PresignUpload() url does not contain bucket name")
	}
	if !strings.Contains(url, "images/test.jpg") {
		t.Errorf("PresignUpload() url does not contain object key")
	}

	// headers 可能为空，但不应为 nil（ SDK 返回 map 或空 map）
	if headers == nil {
		t.Errorf("PresignUpload() headers is nil")
	}
}

// TestOSSClient_PresignUpload_WithHeaders 验证指定 ContentType 与 Metadata 时签名头正确返回。
func TestOSSClient_PresignUpload_WithHeaders(t *testing.T) {
	config := OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	})

	client, err := NewOSSClient(config)
	if err != nil {
		t.Fatalf("NewOSSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	opts := &PresignOptions{
		Expires:     10 * time.Minute,
		ContentType: "image/jpeg",
		Metadata: map[string]string{
			"x-oss-meta-uid": "12345",
		},
	}
	url, headers, err := client.PresignUpload(ctx, "images/test.jpg", opts)
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	if url == "" {
		t.Errorf("PresignUpload() url is empty")
	}

	// 指定 ContentType 后，SignedHeaders 中应包含 Content-Type
	if headers["Content-Type"] != "image/jpeg" {
		t.Errorf("PresignUpload() Content-Type header = %q, want %q", headers["Content-Type"], "image/jpeg")
	}
}

// TestOSSClient_PresignUpload_ContextCanceled 验证上下文取消时返回错误。
func TestOSSClient_PresignUpload_ContextCanceled(t *testing.T) {
	config := OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	})

	client, err := NewOSSClient(config)
	if err != nil {
		t.Fatalf("NewOSSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = client.PresignUpload(ctx, "images/test.jpg", &PresignOptions{Expires: 10 * time.Minute})
	if err == nil {
		t.Fatalf("PresignUpload() expected error when context canceled, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("PresignUpload() error = %v, want context canceled", err)
	}
}

// newOSSTestClient 构造指向本地 httptest 服务的 OSSClient，仅用于测试。
// 使用 path-style 访问，避免虚拟主机域名无法解析到本地测试服务。
func newOSSTestClient(t *testing.T, handler http.HandlerFunc) *OSSClient {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	endpoint := server.URL
	region := "cn-hangzhou"
	usePathStyle := true
	retryMaxAttempts := 1

	sdkClient := oss.NewClient(&oss.Config{
		Endpoint:            &endpoint,
		Region:              &region,
		CredentialsProvider: credentials.NewStaticCredentialsProvider("test-ak", "test-sk"),
		UsePathStyle:        &usePathStyle,
		RetryMaxAttempts:    &retryMaxAttempts,
	})

	return &OSSClient{client: sdkClient, bucket: "test-bucket"}
}

// TestOSSClient_PresignGet 验证预签名下载 URL 的签名版本、有效期与响应头覆盖参数。
func TestOSSClient_PresignGet(t *testing.T) {
	client, err := NewOSSClient(OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}))
	if err != nil {
		t.Fatalf("NewOSSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// 断言 OSSClient 实现了新增接口
	if _, ok := interface{}(client).(PresignDownloader); !ok {
		t.Fatalf("OSSClient does not implement PresignDownloader")
	}
	if _, ok := interface{}(client).(ObjectMetaReader); !ok {
		t.Fatalf("OSSClient does not implement ObjectMetaReader")
	}
	if _, ok := interface{}(client).(StreamDownloader); !ok {
		t.Fatalf("OSSClient does not implement StreamDownloader")
	}

	disposition := "attachment;filename=package.zip"
	rawURL, headers, err := client.PresignGet(context.Background(), "releases/package.zip", &PresignDownloadOptions{
		Expires:                    60 * time.Minute,
		ResponseContentDisposition: disposition,
		ResponseContentType:        "application/zip",
	})
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("PresignGet() returned invalid url %q: %v", rawURL, err)
	}

	if !strings.Contains(parsed.Host, "test-bucket") {
		t.Errorf("PresignGet() host = %q, want bucket in host", parsed.Host)
	}
	if !strings.HasSuffix(parsed.Path, "/releases/package.zip") {
		t.Errorf("PresignGet() path = %q, want object key", parsed.Path)
	}

	query := parsed.Query()
	if got := query.Get("x-oss-signature-version"); got != "OSS4-HMAC-SHA256" {
		t.Errorf("PresignGet() signature version = %q, want OSS4-HMAC-SHA256", got)
	}
	if got := query.Get("x-oss-expires"); got != "3600" {
		t.Errorf("PresignGet() expires = %q, want 3600", got)
	}
	if got := query.Get("response-content-disposition"); got != disposition {
		t.Errorf("PresignGet() response-content-disposition = %q, want %q", got, disposition)
	}
	if got := query.Get("response-content-type"); got != "application/zip" {
		t.Errorf("PresignGet() response-content-type = %q, want application/zip", got)
	}
	if headers == nil {
		t.Errorf("PresignGet() headers is nil")
	}
}

// TestOSSClient_PresignGet_DefaultExpires 验证未指定有效期时使用默认 10 分钟。
func TestOSSClient_PresignGet_DefaultExpires(t *testing.T) {
	client, err := NewOSSClient(OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}))
	if err != nil {
		t.Fatalf("NewOSSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	rawURL, _, err := client.PresignGet(context.Background(), "releases/package.zip", nil)
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("PresignGet() returned invalid url %q: %v", rawURL, err)
	}
	if got := parsed.Query().Get("x-oss-expires"); got != "600" {
		t.Errorf("PresignGet() expires = %q, want 600", got)
	}
}

// TestOSSClient_PresignGet_ContextCanceled 验证上下文取消时返回错误。
func TestOSSClient_PresignGet_ContextCanceled(t *testing.T) {
	client, err := NewOSSClient(OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}))
	if err != nil {
		t.Fatalf("NewOSSClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = client.PresignGet(ctx, "releases/package.zip", nil)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("PresignGet() error = %v, want context canceled", err)
	}
}

// TestOSSClient_HeadObjectMeta 验证元信息解析，包括 x-oss-meta-sha256。
func TestOSSClient_HeadObjectMeta(t *testing.T) {
	client := newOSSTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("method = %q, want HEAD", r.Method)
		}
		if r.URL.Path != "/test-bucket/releases/package.zip" {
			t.Errorf("path = %q, want /test-bucket/releases/package.zip", r.URL.Path)
		}

		w.Header().Set("Content-Length", "12345")
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		w.Header().Set("ETag", `"etag-123"`)
		w.Header().Set("x-oss-meta-sha256", "abc123def456")
		w.WriteHeader(http.StatusOK)
	})

	meta, err := client.HeadObjectMeta(context.Background(), "releases/package.zip")
	if err != nil {
		t.Fatalf("HeadObjectMeta() error = %v", err)
	}
	if meta.Size != 12345 {
		t.Errorf("HeadObjectMeta() size = %d, want 12345", meta.Size)
	}
	if meta.ContentType != "application/zip" {
		t.Errorf("HeadObjectMeta() content type = %q, want application/zip", meta.ContentType)
	}
	wantTime := time.Date(2015, 10, 21, 7, 28, 0, 0, time.UTC)
	if !meta.LastModified.Equal(wantTime) {
		t.Errorf("HeadObjectMeta() last modified = %v, want %v", meta.LastModified, wantTime)
	}
	if meta.ETag != `"etag-123"` {
		t.Errorf("HeadObjectMeta() etag = %q, want %q", meta.ETag, `"etag-123"`)
	}
	if meta.Metadata["sha256"] != "abc123def456" {
		t.Errorf("HeadObjectMeta() metadata sha256 = %q, want abc123def456", meta.Metadata["sha256"])
	}
}

// TestOSSClient_HeadObjectMeta_NotFound 验证 404 可被识别为对象不存在。
func TestOSSClient_HeadObjectMeta_NotFound(t *testing.T) {
	client := newOSSTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Error><Code>NoSuchKey</Code><Message>The specified key does not exist.</Message><RequestId>req-1</RequestId><HostId>host-1</HostId></Error>`))
	})

	meta, err := client.HeadObjectMeta(context.Background(), "releases/missing.zip")
	if meta != nil {
		t.Errorf("HeadObjectMeta() meta = %v, want nil", meta)
	}
	if !errors.Is(err, ErrObjectNotFound) {
		t.Errorf("HeadObjectMeta() error = %v, want ErrObjectNotFound", err)
	}
}

// TestOSSClient_HeadObjectMeta_ServerError 验证非 404 错误不会被误判为对象不存在。
func TestOSSClient_HeadObjectMeta_ServerError(t *testing.T) {
	client := newOSSTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.HeadObjectMeta(context.Background(), "releases/package.zip")
	if err == nil {
		t.Fatalf("HeadObjectMeta() expected error, got nil")
	}
	if errors.Is(err, ErrObjectNotFound) {
		t.Errorf("HeadObjectMeta() error = %v, must not be ErrObjectNotFound", err)
	}
}

// TestOSSClient_OpenDownload 验证流式下载返回 io.ReadCloser，并透传 Range 请求头。
func TestOSSClient_OpenDownload(t *testing.T) {
	var gotRange string
	client := newOSSTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		gotRange = r.Header.Get("Range")

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello-stream"))
	})

	ctx := context.Background()
	reader, err := client.OpenDownload(ctx, "releases/package.zip", &OpenDownloadOptions{Range: "bytes=0-11"})
	if err != nil {
		t.Fatalf("OpenDownload() error = %v", err)
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("OpenDownload() read error = %v", err)
	}
	if string(data) != "hello-stream" {
		t.Errorf("OpenDownload() body = %q, want hello-stream", string(data))
	}
	if gotRange != "bytes=0-11" {
		t.Errorf("OpenDownload() Range = %q, want bytes=0-11", gotRange)
	}
}

// TestOSSClient_OpenDownload_NotFound 验证下载流建立阶段的 404 识别。
func TestOSSClient_OpenDownload_NotFound(t *testing.T) {
	client := newOSSTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`<Error><Code>NoSuchKey</Code><Message>not found</Message></Error>`))
	})

	reader, err := client.OpenDownload(context.Background(), "releases/missing.zip", nil)
	if reader != nil {
		_ = reader.Close()
		t.Errorf("OpenDownload() reader = %v, want nil", reader)
	}
	if !errors.Is(err, ErrObjectNotFound) {
		t.Errorf("OpenDownload() error = %v, want ErrObjectNotFound", err)
	}
}

// TestOSSClient_OpenDownload_ContextCanceled 验证上下文取消时返回错误。
func TestOSSClient_OpenDownload_ContextCanceled(t *testing.T) {
	client := newOSSTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reader, err := client.OpenDownload(ctx, "releases/package.zip", nil)
	if reader != nil {
		_ = reader.Close()
		t.Errorf("OpenDownload() reader = %v, want nil", reader)
	}
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("OpenDownload() error = %v, want context canceled", err)
	}
	_ = client.Close()
}
