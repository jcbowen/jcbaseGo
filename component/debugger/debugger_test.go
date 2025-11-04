package debugger

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestLoggerInterface 测试Logger接口的基本功能
func TestLoggerInterface(t *testing.T) {
	// 创建调试器
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelDebug, // 设置为调试级别，记录所有日志
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 获取Logger实例
	logger := debugger.GetLogger()

	// 测试Debug级别日志
	t.Run("Debug日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Debug("测试Debug日志")
		})

		// 检查日志消息是否包含在输出中
		assert.Contains(t, output, "测试Debug日志")
	})

	// 测试Info级别日志
	t.Run("Info日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Info("测试Info日志", map[string]interface{}{
				"user_id": 123,
				"action":  "login",
			})
		})

		// 检查日志消息是否包含在输出中
		assert.Contains(t, output, "测试Info日志")
	})

	// 测试Warn级别日志
	t.Run("Warn日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Warn("测试Warn日志", map[string]interface{}{
				"warning_type": "validation",
				"field":        "email",
			})
		})

		// 检查日志消息是否包含在输出中
		assert.Contains(t, output, "测试Warn日志")
	})

	// 测试Error级别日志
	t.Run("Error日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Error("测试Error日志", map[string]interface{}{
				"error_code": "DB_CONNECTION_FAILED",
				"details":    "数据库连接超时",
			})
		})

		// 检查日志消息是否包含在输出中
		assert.Contains(t, output, "测试Error日志")
	})
}

// TestLoggerLevelFiltering 测试日志级别过滤功能
func TestLoggerLevelFiltering(t *testing.T) {
	testCases := []struct {
		name        string
		configLevel string
		logLevel    string
		shouldLog   bool
	}{
		{"Debug级别记录Debug日志", LevelDebug, LevelDebug, true},
		{"Debug级别记录Info日志", LevelDebug, LevelInfo, true},
		{"Debug级别记录Warn日志", LevelDebug, LevelWarn, true},
		{"Debug级别记录Error日志", LevelDebug, LevelError, true},
		{"Info级别记录Debug日志", LevelInfo, LevelDebug, false},
		{"Info级别记录Info日志", LevelInfo, LevelInfo, true},
		{"Info级别记录Warn日志", LevelInfo, LevelWarn, true},
		{"Info级别记录Error日志", LevelInfo, LevelError, true},
		{"Warn级别记录Debug日志", LevelWarn, LevelDebug, false},
		{"Warn级别记录Info日志", LevelWarn, LevelInfo, false},
		{"Warn级别记录Warn日志", LevelWarn, LevelWarn, true},
		{"Warn级别记录Error日志", LevelWarn, LevelError, true},
		{"Error级别记录Debug日志", LevelError, LevelDebug, false},
		{"Error级别记录Info日志", LevelError, LevelInfo, false},
		{"Error级别记录Warn日志", LevelError, LevelWarn, false},
		{"Error级别记录Error日志", LevelError, LevelError, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建调试器
			memoryStorage, _ := NewMemoryStorage(100)
			config := &Config{
				Enabled: true,
				Storage: memoryStorage,
				Level:   tc.configLevel,
			}
			debugger, err := New(config)
			assert.NoError(t, err)

			// 获取Logger实例
			logger := debugger.GetLogger()

			output := captureOutput(func() {
				// 根据日志级别调用相应的方法
				switch tc.logLevel {
				case LevelDebug:
					logger.Debug("测试日志")
				case LevelInfo:
					logger.Info("测试日志")
				case LevelWarn:
					logger.Warn("测试日志")
				case LevelError:
					logger.Error("测试日志")
				}
			})

			if tc.shouldLog {
				assert.NotEmpty(t, output, "应该记录日志但输出为空")
				assert.Contains(t, output, "测试日志")
			} else {
				assert.Empty(t, output, "不应该记录日志但输出了内容")
			}
		})
	}
}

// TestLoggerWithFields 测试WithFields功能
func TestLoggerWithFields(t *testing.T) {
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelDebug,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建带有基础字段的Logger
	baseLogger := debugger.GetLoggerWithFields(map[string]interface{}{
		"app_name": "test_app",
		"version":  "1.0.0",
	})

	output := captureOutput(func() {
		// 记录日志
		baseLogger.Info("带有基础字段的日志")
	})

	// 检查日志消息是否包含在输出中
	assert.Contains(t, output, "带有基础字段的日志")

	// 测试字段合并
	enhancedLogger := baseLogger.WithFields(map[string]interface{}{
		"user_id": 456,
		"action":  "purchase",
	})

	output = captureOutput(func() {
		enhancedLogger.Info("合并字段后的日志")
	})

	// 检查日志消息是否包含在输出中
	assert.Contains(t, output, "合并字段后的日志")
}

// TestLoggerInContext 测试在Gin上下文中使用Logger
func TestLoggerInContext(t *testing.T) {
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelDebug,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin上下文
	c := &gin.Context{}

	// 模拟中间件设置Logger到上下文
	logger := debugger.GetLoggerWithFields(map[string]interface{}{
		"request_id": "test-request-123",
		"method":     "GET",
		"url":        "/api/test",
	})
	c.Set("debugger_logger", logger)

	// 从上下文中获取Logger
	contextLogger := GetLoggerFromContext(c)

	output := captureOutput(func() {
		// 使用从上下文获取的Logger记录日志
		contextLogger.Info("从上下文获取的Logger测试")
	})

	// 检查日志消息是否包含在输出中
	assert.Contains(t, output, "从上下文获取的Logger测试")
}

// TestLoggerWithoutContext 测试在没有Logger的上下文中获取Logger
func TestLoggerWithoutContext(t *testing.T) {
	// 创建空的Gin上下文
	c := &gin.Context{}

	// 从上下文中获取Logger（应该返回默认Logger）
	logger := GetLoggerFromContext(c)

	output := captureOutput(func() {
		// 记录日志
		logger.Warn("在没有Logger的上下文中测试")
	})

	// 验证输出包含日志消息
	assert.Contains(t, output, "在没有Logger的上下文中测试")
}

// TestCustomLogger 测试自定义Logger功能
// TODO: 需要定义CustomLogger结构体后才能启用此测试
// func TestCustomLogger(t *testing.T) {
// 	// 创建自定义Logger
// 	customLogger := &CustomLogger{
// 		prefix: "[TEST] ",
// 	}
//
// 	// 创建调试器并传入自定义Logger
// 	memoryStorage, _ := NewMemoryStorage(100)
// 	config := &Config{
// 		Enabled: true,
// 		Storage: memoryStorage,
// 	}
// 	config.Logger = customLogger
// 	debugger, err := New(config)
// 	assert.NoError(t, err)
//
// 	// 获取Logger实例
// 	logger := debugger.GetLogger()
//
// 	output := captureOutput(func() {
// 		// 使用Logger记录日志
// 		logger.Info("自定义Logger测试")
// 	})
//
// 	// 验证自定义Logger的输出格式
// 	assert.Contains(t, output, "[TEST]")
// 	assert.Contains(t, output, "[INFO]")
// 	assert.Contains(t, output, "自定义Logger测试")
// }

// TestLoggerJSONErrorHandling 测试JSON序列化错误处理
func TestLoggerJSONErrorHandling(t *testing.T) {
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelDebug,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	logger := debugger.GetLogger()

	output := captureOutput(func() {
		// 测试包含无法序列化的字段（函数类型）
		logger.Info("测试无法序列化的字段", map[string]interface{}{
			"valid_field":   "正常字段",
			"invalid_field": func() {}, // 函数无法序列化为JSON
		})
	})

	// 验证输出包含日志消息
	assert.Contains(t, output, "测试无法序列化的字段")
	// 由于当前实现使用简单的log.Println，不包含字段信息
	// 但至少应该包含基本的日志消息
}

// TestMaxRecordsMemoryStorage 测试MemoryStorage的最大记录数量限制
func TestMaxRecordsMemoryStorage(t *testing.T) {
	// 创建MemoryStorage，设置最大记录数为5
	storage, err := NewMemoryStorage(5)
	assert.NoError(t, err)

	// 保存6条记录，应该只保留最新的5条
	for i := 1; i <= 6; i++ {
		entry := &LogEntry{
			ID:         fmt.Sprintf("test-%d", i),
			Timestamp:  time.Now().Add(time.Duration(i) * time.Second),
			Method:     "GET",
			URL:        fmt.Sprintf("/test/%d", i),
			StatusCode: 200,
			Duration:   time.Duration(i) * time.Millisecond,
			ClientIP:   "127.0.0.1",
			UserAgent:  "test-agent",
			RequestID:  fmt.Sprintf("req-%d", i),
		}
		err := storage.Save(entry)
		assert.NoError(t, err)
	}

	// 获取所有记录
	entries, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, total, "总记录数应该为5（最大限制）")

	// 验证记录数量不超过最大限制
	assert.LessOrEqual(t, len(entries), 5)

	// 验证最早的记录（ID="test-1"）已被删除
	found := false
	for _, entry := range entries {
		if entry.ID == "test-1" {
			found = true
			break
		}
	}
	assert.False(t, found, "最早的记录应该被删除")

	// 验证最新的记录（ID="test-6"）存在
	found = false
	for _, entry := range entries {
		if entry.ID == "test-6" {
			found = true
			break
		}
	}
	assert.True(t, found, "最新的记录应该存在")
}

// TestMaxRecordsFileStorage 测试FileStorage的最大记录数量限制
func TestMaxRecordsFileStorage(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "debugger_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建FileStorage，设置最大记录数为3
	storage, err := NewFileStorage(tempDir, 3)
	assert.NoError(t, err)

	// 保存4条记录，应该只保留最新的3条
	for i := 1; i <= 4; i++ {
		entry := &LogEntry{
			ID:         fmt.Sprintf("test-%d", i),
			Timestamp:  time.Now().Add(time.Duration(i) * time.Second),
			Method:     "GET",
			URL:        fmt.Sprintf("/test/%d", i),
			StatusCode: 200,
			Duration:   time.Duration(i) * time.Millisecond,
			ClientIP:   "127.0.0.1",
			UserAgent:  "test-agent",
			RequestID:  fmt.Sprintf("req-%d", i),
		}
		err := storage.Save(entry)
		assert.NoError(t, err)
	}

	// 获取所有记录
	entries, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, total, "总记录数应该为3（最大限制）")

	// 验证记录数量不超过最大限制
	assert.LessOrEqual(t, len(entries), 3)

	// 验证最早的记录（ID="test-1"）已被删除
	found := false
	for _, entry := range entries {
		if entry.ID == "test-1" {
			found = true
			break
		}
	}
	assert.False(t, found, "最早的记录应该被删除")

	// 验证最新的记录（ID="test-4"）存在
	found = false
	for _, entry := range entries {
		if entry.ID == "test-4" {
			found = true
			break
		}
	}
	assert.True(t, found, "最新的记录应该存在")
}

// TestMaxRecordsDebuggerConfig 测试调试器配置中的最大记录数量
func TestMaxRecordsDebuggerConfig(t *testing.T) {
	// 创建调试器，设置最大记录数为2
	memoryStorage, _ := NewMemoryStorage(2)
	config := &Config{
		Enabled:    true,
		Storage:    memoryStorage,
		Level:      LevelDebug,
		MaxRecords: 2, // 设置最大记录数为2
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 获取存储实例
	storage := debugger.GetStorage()

	// 保存3条记录，应该只保留最新的2条
	for i := 1; i <= 3; i++ {
		entry := &LogEntry{
			ID:         fmt.Sprintf("test-%d", i),
			Timestamp:  time.Now().Add(time.Duration(i) * time.Second),
			Method:     "GET",
			URL:        fmt.Sprintf("/test/%d", i),
			StatusCode: 200,
			Duration:   time.Duration(i) * time.Millisecond,
			ClientIP:   "127.0.0.1",
			UserAgent:  "test-agent",
			RequestID:  fmt.Sprintf("req-%d", i),
		}
		err := storage.Save(entry)
		assert.NoError(t, err)
	}

	// 获取所有记录
	entries, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)

	// 验证记录数量不超过最大限制
	assert.LessOrEqual(t, len(entries), 2)
	assert.Equal(t, 2, total, "总记录数应该为2（最大限制）")

	// 验证最早的记录（ID="test-1"）已被删除
	found := false
	for _, entry := range entries {
		if entry.ID == "test-1" {
			found = true
			break
		}
	}
	assert.False(t, found, "最早的记录应该被删除")

	// 验证最新的记录（ID="test-3"）存在
	found = false
	for _, entry := range entries {
		if entry.ID == "test-3" {
			found = true
			break
		}
	}
	assert.True(t, found, "最新的记录应该存在")
}

// TestMaxRecordsZeroUnlimited 测试最大记录数为0时表示无限制
func TestMaxRecordsZeroUnlimited(t *testing.T) {
	// 创建MemoryStorage，设置最大记录数为0（无限制）
	storage, err := NewMemoryStorage(0)
	assert.NoError(t, err)

	// 保存大量记录
	for i := 1; i <= 100; i++ {
		entry := &LogEntry{
			ID:         fmt.Sprintf("test-%d", i),
			Timestamp:  time.Now().Add(time.Duration(i) * time.Second),
			Method:     "GET",
			URL:        fmt.Sprintf("/test/%d", i),
			StatusCode: 200,
			Duration:   time.Duration(i) * time.Millisecond,
			ClientIP:   "127.0.0.1",
			UserAgent:  "test-agent",
			RequestID:  fmt.Sprintf("req-%d", i),
		}
		err := storage.Save(entry)
		assert.NoError(t, err)
	}

	// 获取所有记录
	entries, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)

	// 验证所有记录都存在（无限制）
	assert.Equal(t, 100, total)

	// 验证所有记录ID都存在
	for i := 1; i <= 100; i++ {
		found := false
		for _, entry := range entries {
			if entry.ID == fmt.Sprintf("test-%d", i) {
				found = true
				break
			}
		}
		assert.True(t, found, "记录ID=test-%d应该存在", i)
	}
}

// TestFileUploadBinaryDataHandling 测试文件上传二进制数据处理功能
func TestFileUploadBinaryDataHandling(t *testing.T) {
	// 创建调试器
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:     true,
		Storage:     memoryStorage,
		Level:       LevelDebug,
		MaxBodySize: 1024, // 1MB
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 测试1: 普通文本数据
	t.Run("普通文本数据", func(t *testing.T) {
		c := &gin.Context{}
		c.Request = &http.Request{
			Method: "POST",
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"name":"test","value":123}`)),
		}

		result, err := debugger.extractRequestBody(c)
		assert.NoError(t, err)
		assert.Equal(t, `{"name":"test","value":123}`, result)
	})

	// 测试2: 二进制数据（JPEG图片）
	t.Run("二进制数据-JPEG", func(t *testing.T) {
		c := &gin.Context{}
		c.Request = &http.Request{
			Method: "POST",
			Header: http.Header{
				"Content-Type": []string{"application/octet-stream"},
			},
			Body: io.NopCloser(bytes.NewBuffer([]byte{
				0xFF, 0xD8, 0xFF, 0xE0, // JPEG文件头
				0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
			})),
		}

		result, err := debugger.extractRequestBody(c)
		assert.NoError(t, err)
		assert.Contains(t, result, "[Binary Data:")
		assert.Contains(t, result, "JPEG")
	})

	// 测试3: multipart/form-data文件上传
	t.Run("multipart文件上传", func(t *testing.T) {
		// 创建multipart表单数据
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		// 添加文本字段
		writer.WriteField("username", "testuser")
		writer.WriteField("email", "test@example.com")

		// 添加文件字段
		fileWriter, _ := writer.CreateFormFile("avatar", "avatar.jpg")
		fileWriter.Write([]byte{
			0xFF, 0xD8, 0xFF, 0xE0, // JPEG文件头
			0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		})

		writer.Close()

		c := &gin.Context{}
		c.Request = &http.Request{
			Method: "POST",
			Header: http.Header{
				"Content-Type": []string{writer.FormDataContentType()},
			},
			Body: io.NopCloser(bytes.NewBuffer(body.Bytes())),
		}

		result, err := debugger.extractRequestBody(c)
		assert.NoError(t, err)
		assert.Contains(t, result, "[Multipart Form Data]")
		assert.Contains(t, result, "Total Size:")
		assert.Contains(t, result, "Parts:")
		assert.Contains(t, result, "[File] Name: avatar.jpg")
		assert.Contains(t, result, "[Field] Name: username")
		assert.Contains(t, result, "testuser")
	})

	// 测试4: 大文件二进制数据
	t.Run("大文件二进制数据", func(t *testing.T) {
		// 创建大文件数据（超过512字节）
		largeData := make([]byte, 1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		c := &gin.Context{}
		c.Request = &http.Request{
			Method: "POST",
			Header: http.Header{
				"Content-Type": []string{"application/octet-stream"},
			},
			Body: io.NopCloser(bytes.NewBuffer(largeData)),
		}

		result, err := debugger.extractRequestBody(c)
		assert.NoError(t, err)
		assert.Contains(t, result, "[Binary File:")
		assert.Contains(t, result, "1024 bytes")
	})

	// 测试5: 响应体二进制数据处理
	t.Run("响应体二进制数据", func(t *testing.T) {
		// 创建调试器
		debugger, err := New(config)
		assert.NoError(t, err)

		// 创建日志条目
		entry := &LogEntry{
			ID:        "test-response-binary",
			Timestamp: time.Now(),
			Method:    "GET",
			URL:       "/download",
		}

		// 创建响应写入器
		writer := &responseWriter{
			ResponseWriter: &mockResponseWriter{},
			body:           &bytes.Buffer{},
		}

		// 写入二进制数据
		binaryData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG文件头
		}
		writer.body.Write(binaryData)

		// 记录响应信息
		debugger.recordResponseInfo(entry, writer, time.Now())

		// 验证响应体被正确格式化
		assert.Contains(t, entry.ResponseBody, "[Binary Data:")
		assert.Contains(t, entry.ResponseBody, "PNG")
	})
}

// mockResponseWriter 模拟响应写入器
type mockResponseWriter struct {
	headers http.Header
	status  int
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func (m *mockResponseWriter) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func (m *mockResponseWriter) Flush() {
	// 空实现
}

func (m *mockResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *mockResponseWriter) Pusher() http.Pusher {
	return nil
}

func (m *mockResponseWriter) Size() int {
	return 0
}

func (m *mockResponseWriter) Status() int {
	return m.status
}

func (m *mockResponseWriter) WriteHeaderNow() {
	// 空实现
}

func (m *mockResponseWriter) WriteString(s string) (int, error) {
	return len(s), nil
}

func (m *mockResponseWriter) Written() bool {
	return true
}

// captureOutput 捕获标准输出和标准错误
func captureOutput(f func()) string {
	// 捕获标准输出
	oldStdout := os.Stdout
	rStdout, wStdout, _ := os.Pipe()
	os.Stdout = wStdout

	// 捕获标准错误
	oldStderr := os.Stderr
	rStderr, wStderr, _ := os.Pipe()
	os.Stderr = wStderr

	// 设置log包输出到我们重定向的标准错误
	oldLogOutput := log.Writer()
	log.SetOutput(wStderr)

	f()

	wStdout.Close()
	wStderr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	log.SetOutput(oldLogOutput)

	var buf bytes.Buffer
	buf.ReadFrom(rStdout)
	buf.ReadFrom(rStderr)
	return buf.String()
}
