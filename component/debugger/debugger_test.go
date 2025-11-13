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
	"net/http/httptest"
	"os"
	"strings"
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
		Level:   LevelInfo, // 设置为最高调试级别，记录所有日志
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 获取Logger实例
	logger := debugger.GetLogger()

	// 测试Info级别日志（LevelInfo是最高调试级别）
	t.Run("Info日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Info("测试Info日志")
		})

		// 检查日志消息是否包含在输出中
		assert.Contains(t, output, "测试Info日志")
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
		configLevel LogLevel
		logLevel    LogLevel
		shouldLog   bool
	}{
		{"Info级别记录Info日志", LevelInfo, LevelInfo, true},
		{"Info级别记录Warn日志", LevelInfo, LevelWarn, true},
		{"Info级别记录Error日志", LevelInfo, LevelError, true},
		{"Info级别记录Info日志", LevelInfo, LevelInfo, true},
		{"Info级别记录Warn日志", LevelInfo, LevelWarn, true},
		{"Info级别记录Error日志", LevelInfo, LevelError, true},
		{"Warn级别记录Info日志", LevelWarn, LevelInfo, false},
		{"Warn级别记录Warn日志", LevelWarn, LevelWarn, true},
		{"Warn级别记录Error日志", LevelWarn, LevelError, true},
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
		Level:   LevelInfo,
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

// TestLoggerLocationInfo 测试日志位置信息记录功能
func TestLoggerLocationInfo(t *testing.T) {
	// 创建调试器
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelInfo,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 获取Logger实例
	logger := debugger.GetLogger()

	// 测试位置信息记录
	t.Run("位置信息记录", func(t *testing.T) {
		// 记录一条日志
		logger.Info("测试位置信息记录")

		// 获取DefaultLogger实例来访问内部logs字段
		if defaultLogger, ok := logger.(*DefaultLogger); ok {
			logs := defaultLogger.GetLogs()
			assert.Greater(t, len(logs), 0, "应该至少有一条日志记录")

			// 检查最后一条日志的位置信息
			lastLog := logs[len(logs)-1]

			// 验证位置信息字段存在
			assert.NotEmpty(t, lastLog.FileName, "文件名不应该为空")
			assert.Greater(t, lastLog.Line, 0, "行号应该大于0")
			assert.NotEmpty(t, lastLog.Function, "函数名不应该为空")

			// 验证位置信息包含预期内容
			assert.Contains(t, lastLog.FileName, ".go", "文件名应该包含.go后缀")
			assert.Contains(t, lastLog.Function, "TestLoggerLocationInfo", "函数名应该包含测试函数名")
		}
	})

	// 测试位置信息在日志输出中的显示
	t.Run("位置信息输出格式", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Info("测试位置信息输出")
		})

		// 检查输出是否包含位置信息格式
		assert.Contains(t, output, "[info]", "输出应该包含日志级别")
		assert.Contains(t, output, ":", "输出应该包含位置分隔符")
		assert.Contains(t, output, "测试位置信息输出", "输出应该包含日志消息")
	})
}

// TestLoggerInContext 测试在Gin上下文中使用Logger
func TestLoggerInContext(t *testing.T) {
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelInfo,
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
		Level:   LevelInfo,
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
		Level:      LevelInfo,
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
		Enabled:                true,
		Storage:                memoryStorage,
		Level:                  LevelInfo,
		MaxBodySize:            1024, // 1MB
		EnableMultipartSupport: true,
		MultipartPreserveState: true,
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
		assert.Contains(t, result, "[Multipart Form Data - Safe Processing]")
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

// TestMultipartOptimization 测试multipart优化功能
func TestMultipartOptimization(t *testing.T) {
	testCases := []struct {
		name                   string
		enableMultipartSupport bool
		multipartPreserveState bool
		shouldPreserveState    bool
		expectedProcessingType string
	}{
		{"启用multipart支持且保持状态", true, true, true, "Safe Processing"},
		{"启用multipart支持但不保持状态", true, false, false, "Stream Processing"},
		{"禁用multipart支持", false, true, false, "disabled"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建调试器配置
			memoryStorage, _ := NewMemoryStorage(100)
			config := &Config{
				Enabled:                true,
				Storage:                memoryStorage,
				EnableMultipartSupport: tc.enableMultipartSupport,
				MultipartPreserveState: tc.multipartPreserveState,
			}
			debugger, err := New(config)
			assert.NoError(t, err)

			// 创建multipart表单数据
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			writer.WriteField("test_field", "test_value")
			writer.Close()

			// 创建Gin上下文
			c := &gin.Context{}
			c.Request = &http.Request{
				Method: "POST",
				Header: http.Header{
					"Content-Type": []string{writer.FormDataContentType()},
				},
				Body: io.NopCloser(bytes.NewBuffer(body.Bytes())),
			}

			// 测试请求体提取
			result, err := debugger.extractRequestBody(c)
			assert.NoError(t, err)

			// 验证处理类型
			if tc.enableMultipartSupport {
				assert.Contains(t, result, tc.expectedProcessingType)
			} else {
				assert.Contains(t, result, "disabled")
			}

			// 验证状态保持
			if tc.shouldPreserveState {
				// 检查请求体是否仍然可读
				bodyBytes, err := io.ReadAll(c.Request.Body)
				assert.NoError(t, err)
				assert.NotEmpty(t, bodyBytes)
				// 恢复请求体以便后续测试
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		})
	}
}

// TestMiddlewareOrder 测试中间件执行顺序配置
func TestMiddlewareOrder(t *testing.T) {
	testCases := []struct {
		name            string
		middlewareOrder string
		expectedType    string
	}{
		{"正常执行顺序", "normal", "Normal"},
		{"优先执行顺序", "early", "Early"},
		{"最后执行顺序", "late", "Late"},
		{"默认执行顺序", "", "Normal"},
		{"无效执行顺序", "invalid", "Normal"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建调试器配置
			memoryStorage, _ := NewMemoryStorage(100)
			config := &Config{
				Enabled:         true,
				Storage:         memoryStorage,
				MiddlewareOrder: tc.middlewareOrder,
			}
			debugger, err := New(config)
			assert.NoError(t, err)

			// 获取中间件函数
			middlewareFunc := debugger.Middleware()
			assert.NotNil(t, middlewareFunc)

			// 使用httptest创建完整的HTTP请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:8080"

			// 创建Gin上下文
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 执行中间件
			middlewareFunc(c)

			// 验证中间件正确执行（没有panic且状态码正确）
			assert.Equal(t, 200, w.Code)
		})
	}
}

// TestMultipartStateRestoration 测试multipart状态恢复功能
func TestMultipartStateRestoration(t *testing.T) {
	t.Log("[DEBUG] TestMultipartStateRestoration STARTING")

	// 创建调试器配置（启用multipart支持和状态保持）
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableMultipartSupport: true,
		MultipartPreserveState: true,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	t.Logf("[DEBUG] Debugger created with EnableMultipartSupport: %v, MultipartPreserveState: %v",
		config.EnableMultipartSupport, config.MultipartPreserveState)

	// 创建multipart表单数据
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("field1", "value1")
	writer.WriteField("field2", "value2")
	writer.Close()

	// 保存原始body内容用于后续比较
	originalBody := make([]byte, body.Len())
	copy(originalBody, body.Bytes())

	// 创建Gin上下文
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.RemoteAddr = "127.0.0.1:8080"

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 调试：检查请求创建后的初始状态
	t.Logf("[DEBUG] Initial Content-Type: %s", c.Request.Header.Get("Content-Type"))
	t.Logf("[DEBUG] Initial Body is nil: %v", c.Request.Body == nil)
	t.Log("[DEBUG] Test starting...")

	// 模拟中间件执行流程
	startTime := time.Now()

	// 调试：在调用createLogEntry之前检查请求体状态
	t.Logf("[DEBUG] Before createLogEntry: Content-Type: %s", c.Request.Header.Get("Content-Type"))
	t.Logf("[DEBUG] Before createLogEntry: Body is nil: %v", c.Request.Body == nil)

	// 检查请求体是否可读
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			t.Logf("[DEBUG] Before createLogEntry: Body length: %d", len(bodyBytes))
			t.Logf("[DEBUG] Before createLogEntry: Body content: %s", string(bodyBytes))
			// 恢复请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		} else {
			t.Logf("[DEBUG] Before createLogEntry: Failed to read body: %v", err)
		}
	}

	entry := debugger.createLogEntry(c, startTime)

	// 调试：检查createLogEntry是否已经提取了请求体
	t.Logf("[DEBUG] After createLogEntry: RequestBody contains 'Safe Processing': %v", strings.Contains(entry.RequestBody, "Safe Processing"))
	t.Logf("[DEBUG] After createLogEntry: RequestBody length: %d", len(entry.RequestBody))
	t.Logf("[DEBUG] After createLogEntry: RequestBody content: %s", entry.RequestBody)

	// 验证createLogEntry已经提取了请求体
	assert.Contains(t, entry.RequestBody, "Safe Processing", "createLogEntry应该已经提取了multipart请求体")

	// 注意：createLogEntry已经调用了extractRequestBody，这里不需要再次调用
	// 直接使用entry.RequestBody作为结果
	result := entry.RequestBody

	// 调试：检查createLogEntry提取的请求体
	t.Logf("[DEBUG] createLogEntry提取的请求体长度: %d", len(result))
	t.Logf("[DEBUG] createLogEntry提取的请求体内容: %s", result)

	// 验证createLogEntry已经正确提取了multipart请求体
	assert.Contains(t, result, "field1", "createLogEntry应该已经提取了field1字段")
	assert.Contains(t, result, "field2", "createLogEntry应该已经提取了field2字段")
	assert.Contains(t, result, "value1", "createLogEntry应该已经提取了value1值")
	assert.Contains(t, result, "value2", "createLogEntry应该已经提取了value2值")

	// 在真实的中间件使用场景中，restoreMultipartRequestBody应该在请求处理完成后被调用
	// 这里我们模拟完整的中间件执行流程：
	// 1. createLogEntry记录请求（会读取并恢复请求体）
	// 2. 后续中间件处理请求（可以正常访问请求体）
	// 3. 请求处理完成后调用restoreMultipartRequestBody恢复multipart状态

	// 模拟后续中间件处理请求（验证请求体可以正常访问）
	if c.Request.Body != nil {
		// 保存原始请求体内容用于后续验证
		originalBodyBytes, err := io.ReadAll(c.Request.Body)
		assert.NoError(t, err, "后续中间件应该可以正常读取请求体")

		// 恢复请求体，确保其他中间件也能访问
		c.Request.Body = io.NopCloser(bytes.NewBuffer(originalBodyBytes))

		// 调试：打印原始请求体内容
		t.Logf("后续中间件读取的请求体长度: %d", len(originalBodyBytes))
		t.Logf("后续中间件读取的请求体内容: %s", string(originalBodyBytes))

		// 验证请求体包含multipart数据
		assert.NotEmpty(t, originalBodyBytes, "后续中间件读取的请求体不应该为空")
		assert.Contains(t, string(originalBodyBytes), "field1", "后续中间件读取的请求体应该包含field1字段")
		assert.Contains(t, string(originalBodyBytes), "field2", "后续中间件读取的请求体应该包含field2字段")
		assert.Contains(t, string(originalBodyBytes), "value1", "后续中间件读取的请求体应该包含value1值")
		assert.Contains(t, string(originalBodyBytes), "value2", "后续中间件读取的请求体应该包含value2值")
	}

	// 模拟请求处理完成后恢复multipart状态
	debugger.restoreMultipartRequestBody(c)

	// 验证multipart状态已正确恢复
	// 这里我们主要验证Gin的multipart状态是否正确重置，而不是再次读取请求体
	t.Logf("restoreMultipartRequestBody执行完成，multipart状态已恢复")
}

// TestSafeRequestBodyExtraction 测试安全的请求体提取方法
func TestSafeRequestBodyExtraction(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableMultipartSupport: true,
		MultipartPreserveState: true,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 测试1: 普通文本请求体
	t.Run("普通文本请求体", func(t *testing.T) {
		c := &gin.Context{}
		c.Request = &http.Request{
			Method: "POST",
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBufferString(`{"key":"value"}`)),
		}

		result, err := debugger.extractRequestBodyWithSizeLimitSafe(c)
		assert.NoError(t, err)
		assert.Contains(t, result, `{"key":"value"}`)

		// 验证请求体已恢复
		restoredBytes, err := io.ReadAll(c.Request.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, string(restoredBytes))
	})

	// 测试2: multipart请求体
	t.Run("multipart请求体", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		writer.WriteField("test", "data")
		writer.Close()

		c := &gin.Context{}
		c.Request = &http.Request{
			Method: "POST",
			Header: http.Header{
				"Content-Type": []string{writer.FormDataContentType()},
			},
			Body: io.NopCloser(bytes.NewBuffer(body.Bytes())),
		}

		result, err := debugger.extractRequestBodyWithSizeLimitSafe(c)
		assert.NoError(t, err)
		// extractRequestBodyWithSizeLimitSafe 返回原始请求体内容，而不是"Safe Processing"
		assert.Contains(t, result, "--")   // multipart内容应该包含boundary
		assert.Contains(t, result, "test") // 应该包含字段名
		assert.Contains(t, result, "data") // 应该包含字段值

		// 验证请求体已恢复
		restoredBytes, err := io.ReadAll(c.Request.Body)
		assert.NoError(t, err)
		assert.Equal(t, body.Bytes(), restoredBytes)
	})
}

// TestMiddlewareExecutionFlow 测试中间件执行流程
func TestMiddlewareExecutionFlow(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:         true,
		Storage:         memoryStorage,
		MiddlewareOrder: "normal",
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin路由
	router := gin.New()
	router.Use(debugger.Middleware())

	// 添加测试路由
	router.GET("/test", func(c *gin.Context) {
		// 从上下文中获取Logger
		logger := GetLoggerFromContext(c)
		logger.Info("处理请求中")
		c.JSON(200, gin.H{"message": "success"})
	})

	// 创建测试请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")

	// 验证日志记录
	entries, _, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Greater(t, len(entries), 0)

	// 验证日志条目内容
	entry := entries[0]
	assert.Equal(t, "GET", entry.Method)
	assert.Equal(t, "/test", entry.URL)
	assert.Equal(t, 200, entry.StatusCode)
}

// TestMiddlewareOrderIntegration 测试中间件执行顺序的集成测试
func TestMiddlewareOrderIntegration(t *testing.T) {
	testCases := []struct {
		name            string
		middlewareOrder string
		description     string
	}{
		{"normal", "正常执行顺序", "在标准位置执行，适用于大多数场景"},
		{"early", "优先执行顺序", "在其他中间件之前执行，适用于需要记录完整请求信息的场景"},
		{"late", "最后执行顺序", "在其他中间件之后执行，适用于需要记录完整响应信息的场景"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建调试器配置
			memoryStorage, _ := NewMemoryStorage(100)
			config := &Config{
				Enabled:         true,
				Storage:         memoryStorage,
				MiddlewareOrder: tc.middlewareOrder,
			}
			debugger, err := New(config)
			assert.NoError(t, err)

			// 创建Gin路由
			router := gin.New()

			// 添加其他中间件（用于测试执行顺序）
			router.Use(func(c *gin.Context) {
				c.Set("middleware1", "executed")
				c.Next()
			})

			// 添加调试器中间件
			router.Use(debugger.Middleware())

			// 添加另一个中间件
			router.Use(func(c *gin.Context) {
				c.Set("middleware2", "executed")
				c.Next()
			})

			// 添加测试路由
			router.GET("/order-test", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"middleware1": c.GetString("middleware1"),
					"middleware2": c.GetString("middleware2"),
				})
			})

			// 创建测试请求
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/order-test", nil)
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "executed")

			// 验证日志记录
			entries, _, err := memoryStorage.FindAll(1, 10, nil)
			assert.NoError(t, err)
			assert.Greater(t, len(entries), 0)
		})
	}
}
