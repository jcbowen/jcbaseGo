package attachment

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/attachment/remote"
)

// TestAttachment_GetPresignURL_OSS 验证 StorageType 为 oss 时可以生成预签名 URL。
func TestAttachment_GetPresignURL_OSS(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "oss",
		LocalDir:    "uploads",
	}

	ossConfig := jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}

	att := New(nil, baseConfig, ossConfig)

	url, headers, err := att.GetPresignURL(context.Background(), "images/test.jpg", &remote.PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	if url == "" {
		t.Errorf("GetPresignURL() url is empty")
	}

	if !strings.Contains(url, "test-bucket") {
		t.Errorf("GetPresignURL() url does not contain bucket name")
	}

	if headers == nil {
		t.Errorf("GetPresignURL() headers is nil")
	}
}

// TestAttachment_GetPresignURL_WithHeaders 验证指定 ContentType 时返回对应的签名头。
func TestAttachment_GetPresignURL_WithHeaders(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "oss",
		LocalDir:    "uploads",
	}

	ossConfig := jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}

	att := New(nil, baseConfig, ossConfig)

	opts := &remote.PresignOptions{
		Expires:     10 * time.Minute,
		ContentType: "image/jpeg",
	}
	_, headers, err := att.GetPresignURL(context.Background(), "images/test.jpg", opts)
	if err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	if headers["Content-Type"] != "image/jpeg" {
		t.Errorf("GetPresignURL() Content-Type header = %q, want %q", headers["Content-Type"], "image/jpeg")
	}
}

// TestAttachment_GetPresignURL_Unsupported 验证非远程存储类型（local）返回错误。
// local 不是远程存储类型，在 getRemoteClient 阶段即返回"不支持的远程存储类型"错误，
// 不会进入 PresignUploader 接口断言。
func TestAttachment_GetPresignURL_Unsupported(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "local",
		LocalDir:    "uploads",
	}

	att := New(nil, baseConfig, nil)

	_, _, err := att.GetPresignURL(context.Background(), "images/test.jpg", nil)
	if err == nil {
		t.Fatalf("GetPresignURL() expected error for local storage, got nil")
	}

	if !strings.Contains(err.Error(), "不支持的远程存储类型") {
		t.Errorf("GetPresignURL() error = %v, want 不支持的远程存储类型", err)
	}
}

// TestAttachment_GetPresignURL_COS 验证 StorageType 为 cos 时可以生成预签名 URL。
func TestAttachment_GetPresignURL_COS(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "cos",
		LocalDir:    "uploads",
	}

	cosConfig := jcbaseGo.COSStruct{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Region:    "ap-guangzhou",
		Bucket:    "test-bucket-1250000000",
		Url:       "https://test-bucket-1250000000.cos.ap-guangzhou.myqcloud.com",
	}

	att := New(nil, baseConfig, cosConfig)

	url, headers, err := att.GetPresignURL(context.Background(), "images/test.jpg", &remote.PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	if url == "" {
		t.Errorf("GetPresignURL() url is empty")
	}

	if !strings.Contains(url, "test-bucket-1250000000") {
		t.Errorf("GetPresignURL() url does not contain bucket name")
	}

	if headers == nil {
		t.Errorf("GetPresignURL() headers is nil")
	}
}

// TestAttachment_GetPresignURL_COSWithHeaders 验证 COS 指定 ContentType 与 x-cos-meta-* 元数据时返回对应签名头。
func TestAttachment_GetPresignURL_COSWithHeaders(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "cos",
		LocalDir:    "uploads",
	}

	cosConfig := jcbaseGo.COSStruct{
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
		Region:    "ap-guangzhou",
		Bucket:    "test-bucket-1250000000",
		Url:       "https://test-bucket-1250000000.cos.ap-guangzhou.myqcloud.com",
	}

	att := New(nil, baseConfig, cosConfig)

	opts := &remote.PresignOptions{
		Expires:     10 * time.Minute,
		ContentType: "image/jpeg",
		Metadata:    map[string]string{"x-cos-meta-uid": "12345"},
	}
	_, headers, err := att.GetPresignURL(context.Background(), "images/test.jpg", opts)
	if err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	if headers["Content-Type"] != "image/jpeg" {
		t.Errorf("GetPresignURL() Content-Type header = %q, want %q", headers["Content-Type"], "image/jpeg")
	}
	if headers["X-Cos-Meta-Uid"] != "12345" {
		t.Errorf("GetPresignURL() X-Cos-Meta-Uid header = %q, want %q", headers["X-Cos-Meta-Uid"], "12345")
	}
}

// TestAttachment_GetPresignURL_InvalidConfig 验证 OSS 配置类型错误时返回错误。
func TestAttachment_GetPresignURL_InvalidConfig(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "oss",
		LocalDir:    "uploads",
	}

	att := New(nil, baseConfig, "invalid-config")

	_, _, err := att.GetPresignURL(context.Background(), "images/test.jpg", nil)
	if err == nil {
		t.Fatalf("GetPresignURL() expected error for invalid config, got nil")
	}
}

// TestAttachment_GetPresignURL_ClientCache 验证同一 Attachment 实例重复调用 GetPresignURL 时复用缓存的客户端实例。
func TestAttachment_GetPresignURL_ClientCache(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "oss",
		LocalDir:    "uploads",
	}

	ossConfig := jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}

	att := New(nil, baseConfig, ossConfig)

	if _, _, err := att.GetPresignURL(context.Background(), "images/test1.jpg", &remote.PresignOptions{Expires: 10 * time.Minute}); err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	first := att.remoteClient
	if first == nil {
		t.Fatalf("remoteClient should be cached after first call")
	}

	if _, _, err := att.GetPresignURL(context.Background(), "images/test2.jpg", &remote.PresignOptions{Expires: 10 * time.Minute}); err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	if att.remoteClient != first {
		t.Errorf("remoteClient should be reused across GetPresignURL calls")
	}
}

// TestAttachment_GetPresignURL_RebuildOnConfigChange 验证 RemoteConfig 在运行期间变化时会重建客户端实例。
func TestAttachment_GetPresignURL_RebuildOnConfigChange(t *testing.T) {
	baseConfig := &jcbaseGo.AttachmentStruct{
		StorageType: "oss",
		LocalDir:    "uploads",
	}

	ossConfig := jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	}

	att := New(nil, baseConfig, ossConfig)

	if _, _, err := att.GetPresignURL(context.Background(), "images/test1.jpg", &remote.PresignOptions{Expires: 10 * time.Minute}); err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}
	first := att.remoteClient

	// 模拟运行期间外部修改远程配置
	newConfig := att.RemoteConfig.(jcbaseGo.OSSStruct)
	newConfig.BucketName = "test-bucket-2"
	att.RemoteConfig = newConfig

	url, _, err := att.GetPresignURL(context.Background(), "images/test2.jpg", &remote.PresignOptions{Expires: 10 * time.Minute})
	if err != nil {
		t.Fatalf("GetPresignURL() error = %v", err)
	}

	if att.remoteClient == first {
		t.Errorf("remoteClient should be rebuilt when RemoteConfig changes")
	}

	if !strings.Contains(url, "test-bucket-2") {
		t.Errorf("GetPresignURL() url should use the new bucket, url = %s", url)
	}
}
