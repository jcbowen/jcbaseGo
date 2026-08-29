package remote

import (
	"context"
	"strings"
	"testing"
	"time"

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
