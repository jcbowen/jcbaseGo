package remote

import (
	"context"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// TestNewClient 测试NewClient函数能否正确创建不同类型的客户端
func TestNewClient(t *testing.T) {
	// 测试用配置
	ftpConfig := FTPConfig(jcbaseGo.FTPStruct{
		Address:  "localhost:21",
		Username: "test",
		Password: "test",
		Timeout:  5 * time.Second,
	})

	sftpConfig := SFTPConfig(jcbaseGo.SFTPStruct{
		Address:  "localhost:22",
		Username: "test",
		Password: "test",
		Timeout:  5 * time.Second,
	})

	cosConfig := COSConfig(jcbaseGo.COSStruct{
		Url:       "https://test.cos.ap-guangzhou.myqcloud.com",
		SecretId:  "test-secret-id",
		SecretKey: "test-secret-key",
	})

	ossConfig := OSSConfig(jcbaseGo.OSSStruct{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyId:     "test-access-key-id",
		AccessKeySecret: "test-access-key-secret",
		BucketName:      "test-bucket",
	})

	// 测试用例
	tests := []struct {
		name        string
		storageType string
		config      interface{}
		wantErr     bool
	}{
		{"FTP", TypeFTP, ftpConfig, true},  // 预期失败，因为没有实际的FTP服务器
		{"SFTP", TypeSFTP, sftpConfig, true}, // 预期失败，因为没有实际的SFTP服务器
		{"COS", TypeCOS, cosConfig, true},  // 预期失败，因为没有实际的COS配置
		{"OSS", TypeOSS, ossConfig, true},  // 预期失败，因为没有实际的OSS配置
		{"Unknown", "unknown", nil, true},   // 预期失败，因为类型未知
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.storageType, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 只有当client不为nil时才调用Close()
			if client != nil {
				// 如果成功创建了客户端，确保能正确关闭
				if err := client.Close(); err != nil {
					t.Errorf("Client.Close() error = %v", err)
				}
			}
		})
	}
}

// TestWithContextTimeout 测试withContextTimeout函数
func TestWithContextTimeout(t *testing.T) {
	ctx := context.Background()

	// 测试正常执行
	result, err := withContextTimeout(ctx, "Test", func() (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Errorf("withContextTimeout() error = %v, want nil", err)
	}
	if result != 42 {
		t.Errorf("withContextTimeout() result = %v, want 42", result)
	}

	// 测试错误执行
	_, err = withContextTimeout(ctx, "Test", func() (int, error) {
		return 0, context.Canceled
	})
	if err == nil {
		t.Errorf("withContextTimeout() expected error, got nil")
	}

	// 测试上下文取消
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = withContextTimeout(ctx, "Test", func() (int, error) {
		time.Sleep(1 * time.Second)
		return 42, nil
	})
	if err == nil {
		t.Errorf("withContextTimeout() expected error when context canceled, got nil")
	}
}

// TestWithContextTimeoutVoid 测试withContextTimeoutVoid函数
func TestWithContextTimeoutVoid(t *testing.T) {
	ctx := context.Background()

	// 测试正常执行
	err := withContextTimeoutVoid(ctx, "Test", func() error {
		return nil
	})
	if err != nil {
		t.Errorf("withContextTimeoutVoid() error = %v, want nil", err)
	}

	// 测试错误执行
	err = withContextTimeoutVoid(ctx, "Test", func() error {
		return context.Canceled
	})
	if err == nil {
		t.Errorf("withContextTimeoutVoid() expected error, got nil")
	}

	// 测试上下文取消
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	err = withContextTimeoutVoid(ctx, "Test", func() error {
		time.Sleep(1 * time.Second)
		return nil
	})
	if err == nil {
		t.Errorf("withContextTimeoutVoid() expected error when context canceled, got nil")
	}
}

// TestError 测试Error类型的Error和Unwrap方法
func TestError(t *testing.T) {
	origErr := context.Canceled
	err := &Error{Op: "Test", Err: origErr}

	// 测试Error方法
	expectedMsg := "Test operation failed: context canceled"
	if err.Error() != expectedMsg {
		t.Errorf("Error.Error() = %v, want %v", err.Error(), expectedMsg)
	}

	// 测试Unwrap方法
	if err.Unwrap() != origErr {
		t.Errorf("Error.Unwrap() = %v, want %v", err.Unwrap(), origErr)
	}
}

// TestListOptions 测试ListOptions结构体
func TestListOptions(t *testing.T) {
	// 测试默认值
	options := ListOptions{}
	if options.Prefix != "" {
		t.Errorf("ListOptions.Prefix = %v, want \"\"", options.Prefix)
	}
	if options.Marker != "" {
		t.Errorf("ListOptions.Marker = %v, want \"\"", options.Marker)
	}
	if options.MaxKeys != 0 {
		t.Errorf("ListOptions.MaxKeys = %v, want 0", options.MaxKeys)
	}

	// 测试设置值
	options = ListOptions{
		Prefix:  "test",
		Marker:  "marker",
		MaxKeys: 100,
	}
	if options.Prefix != "test" {
		t.Errorf("ListOptions.Prefix = %v, want \"test\"", options.Prefix)
	}
	if options.Marker != "marker" {
		t.Errorf("ListOptions.Marker = %v, want \"marker\"", options.Marker)
	}
	if options.MaxKeys != 100 {
		t.Errorf("ListOptions.MaxKeys = %v, want 100", options.MaxKeys)
	}
}

// TestFileInfo 测试FileInfo结构体
func TestFileInfo(t *testing.T) {
	modTime := time.Now()
	fileInfo := FileInfo{
		Name:    "test.txt",
		Size:    1024,
		ModTime: modTime,
		IsDir:   false,
	}

	if fileInfo.Name != "test.txt" {
		t.Errorf("FileInfo.Name = %v, want \"test.txt\"", fileInfo.Name)
	}
	if fileInfo.Size != 1024 {
		t.Errorf("FileInfo.Size = %v, want 1024", fileInfo.Size)
	}
	if !fileInfo.ModTime.Equal(modTime) {
		t.Errorf("FileInfo.ModTime = %v, want %v", fileInfo.ModTime, modTime)
	}
	if fileInfo.IsDir != false {
		t.Errorf("FileInfo.IsDir = %v, want false", fileInfo.IsDir)
	}
}

// TestListResult 测试ListResult结构体
func TestListResult(t *testing.T) {
	modTime := time.Now()
	files := []FileInfo{
		{
			Name:    "test1.txt",
			Size:    1024,
			ModTime: modTime,
			IsDir:   false,
		},
		{
			Name:    "test2.txt",
			Size:    2048,
			ModTime: modTime,
			IsDir:   false,
		},
	}

	listResult := ListResult{
		Files:       files,
		NextMarker:  "test2.txt",
		IsTruncated: true,
	}

	if len(listResult.Files) != 2 {
		t.Errorf("ListResult.Files length = %v, want 2", len(listResult.Files))
	}
	if listResult.NextMarker != "test2.txt" {
		t.Errorf("ListResult.NextMarker = %v, want \"test2.txt\"", listResult.NextMarker)
	}
	if listResult.IsTruncated != true {
		t.Errorf("ListResult.IsTruncated = %v, want true", listResult.IsTruncated)
	}
}
