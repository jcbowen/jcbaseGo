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

// TestAttachment_RemoveDomain 验证可从附件 URL 中去除访问域名并得到相对路径。
// 逻辑对齐 PHP 端 Util::removeMediaDomain：命中域名后从文件类型目录处截取，
// 本地 /static/ 资源保留 static 起点。
func TestAttachment_RemoveDomain(t *testing.T) {
	att := &Attachment{
		BaseConfig: &jcbaseGo.AttachmentStruct{
			VisitDomain:      "https://cdn.example.com/",
			LocalVisitDomain: "https://localhost:8080/",
			StorageType:      "local",
			LocalDir:         "uploads",
		},
	}

	// 本地域名下的普通附件：LocalDir 前缀随类型目录定位一并剥离
	if got := att.ToSource("https://localhost:8080/uploads/images/logo.png"); got != "images/logo.png" {
		t.Fatalf("ToSource() = %q, want %q", got, "images/logo.png")
	}

	if got := att.ToSource("https://localhost:8080/uploads/images/logo.png?v=123"); got != "images/logo.png" {
		t.Fatalf("ToSource() = %q, want %q", got, "images/logo.png")
	}

	if got := att.ToSource("https://localhost:8080/uploads/images/logo.png?v=123", false); got != "images/logo.png?v=123" {
		t.Fatalf("ToSource() = %q, want %q", got, "images/logo.png?v=123")
	}

	// 远程域名：同样从类型目录处截取，bucket/path 前缀被剥离（对应 PHP 的 attachment 域名分支）
	if got := att.ToSource("https://cdn.example.com/bucket/videos/2026/01/demo.mp4"); got != "videos/2026/01/demo.mp4" {
		t.Fatalf("ToSource() = %q, want %q", got, "videos/2026/01/demo.mp4")
	}

	// zips 类型目录：正常截取（与 PHP 端 {$type}s/ 存储目录命名一致）
	if got := att.ToSource("https://cdn.example.com/zips/2026/01/demo.zip"); got != "zips/2026/01/demo.zip" {
		t.Fatalf("ToSource() = %q, want %q", got, "zips/2026/01/demo.zip")
	}

	// 路径中仅含 zip 子串但非 zips 目录时不应被误截断
	if got := att.ToSource("https://cdn.example.com/uploads/myzipfile/x.png"); got != "uploads/myzipfile/x.png" {
		t.Fatalf("ToSource() = %q, want %q", got, "uploads/myzipfile/x.png")
	}

	// 本地 /static/ 内部资源：保留 static 起始路径
	if got := att.ToSource("https://localhost:8080/static/js/app.js"); got != "static/js/app.js" {
		t.Fatalf("ToSource() = %q, want %q", got, "static/js/app.js")
	}

	// 未命中任何域名时原样返回相对路径
	if got := att.ToSource("images/logo.png"); got != "images/logo.png" {
		t.Fatalf("ToSource() = %q, want %q", got, "images/logo.png")
	}
}

// TestAttachment_ToMediaToSourceRoundTrip 验证 ToMedia 与 ToSource 本地存储下互逆，
// 还原结果与 FileAttachment 保存的相对路径格式一致（不含 LocalDir 前缀）。
func TestAttachment_ToMediaToSourceRoundTrip(t *testing.T) {
	att := &Attachment{
		BaseConfig: &jcbaseGo.AttachmentStruct{
			VisitDomain:      "https://localhost:8080/",
			LocalVisitDomain: "https://localhost:8080/",
			StorageType:      "local",
			LocalDir:         "uploads",
		},
	}

	src := "images/2026/01/logo.png"
	if got := att.ToSource(att.ToMedia(src)); got != src {
		t.Fatalf("ToSource(ToMedia(%q)) = %q, want %q", src, got, src)
	}
}

// TestAttachment_ToMedia 验证 ToMedia 的 static/ 资源拼接、本地完整 URL 重组与缓存参数语义，
// 逻辑对齐 PHP 端 Util::toMedia。
func TestAttachment_ToMedia(t *testing.T) {
	att := &Attachment{
		BaseConfig: &jcbaseGo.AttachmentStruct{
			VisitDomain:      "https://cdn.example.com/",
			LocalVisitDomain: "https://localhost:8080/",
			StorageType:      "remote",
			LocalDir:         "uploads",
		},
	}

	// 相对路径含 static/ 时：从 static/ 处拼接本地资源域名（不追加 LocalDir）
	if got := att.ToMedia("static/js/app.js"); got != "https://localhost:8080/static/js/app.js" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://localhost:8080/static/js/app.js")
	}
	if got := att.ToMedia("/static/js/app.js"); got != "https://localhost:8080/static/js/app.js" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://localhost:8080/static/js/app.js")
	}

	// 完整本地附件 URL 重组：远程存储下剥离域名与 LocalDir 前缀后切换到远程域名
	if got := att.ToMedia("https://localhost:8080/uploads/images/2026/01/logo.png"); got != "https://cdn.example.com/images/2026/01/logo.png" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://cdn.example.com/images/2026/01/logo.png")
	}

	// 完整本地 /static/ URL：原样返回，不做重组
	if got := att.ToMedia("https://localhost:8080/static/js/app.js"); got != "https://localhost:8080/static/js/app.js" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://localhost:8080/static/js/app.js")
	}

	// 完整本地 URL 但未命中类型目录：原样返回
	if got := att.ToMedia("https://localhost:8080/uploads/logo.png"); got != "https://localhost:8080/uploads/logo.png" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://localhost:8080/uploads/logo.png")
	}

	// 本地存储下重组后按 LocalDir 拼回（与原 URL 一致）
	attLocal := &Attachment{BaseConfig: &jcbaseGo.AttachmentStruct{
		VisitDomain:      "https://cdn.example.com/",
		LocalVisitDomain: "https://localhost:8080/",
		StorageType:      "local",
		LocalDir:         "uploads",
	}}
	if got := attLocal.ToMedia("https://localhost:8080/uploads/images/logo.png"); got != "https://localhost:8080/uploads/images/logo.png" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://localhost:8080/uploads/images/logo.png")
	}

	// 缓存参数：isCache=false 追加时间戳，默认不追加
	if got := att.ToMedia("images/logo.png", false, false); !strings.HasPrefix(got, "https://cdn.example.com/images/logo.png?v=") {
		t.Fatalf("ToMedia() = %q, want prefix %q", got, "https://cdn.example.com/images/logo.png?v=")
	}
	if got := att.ToMedia("images/logo.png"); got != "https://cdn.example.com/images/logo.png" {
		t.Fatalf("ToMedia() = %q, want %q", got, "https://cdn.example.com/images/logo.png")
	}
}
