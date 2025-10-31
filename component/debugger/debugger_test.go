package debugger

import (
	"bytes"
	"encoding/json"
	"fmt"
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

		// 验证输出包含预期的内容
		assert.Contains(t, output, "测试Debug日志")
		assert.Contains(t, output, "debug")

		// 验证JSON格式
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(output), &logData)
		assert.NoError(t, err)

		debugLog, exists := logData["debug_log"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "debug", debugLog["level"])
		assert.Equal(t, "测试Debug日志", debugLog["message"])
	})

	// 测试Info级别日志
	t.Run("Info日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Info("测试Info日志", map[string]interface{}{
				"user_id": 123,
				"action":  "login",
			})
		})

		assert.Contains(t, output, "测试Info日志")
		assert.Contains(t, output, "info")

		var logData map[string]interface{}
		err := json.Unmarshal([]byte(output), &logData)
		assert.NoError(t, err)

		debugLog := logData["debug_log"].(map[string]interface{})
		assert.Equal(t, "info", debugLog["level"])
		assert.Equal(t, "测试Info日志", debugLog["message"])
		assert.Equal(t, float64(123), debugLog["user_id"])
		assert.Equal(t, "login", debugLog["action"])
	})

	// 测试Warn级别日志
	t.Run("Warn日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Warn("测试Warn日志", map[string]interface{}{
				"warning_type": "validation",
				"field":        "email",
			})
		})

		assert.Contains(t, output, "测试Warn日志")
		assert.Contains(t, output, "warn")

		var logData map[string]interface{}
		err := json.Unmarshal([]byte(output), &logData)
		assert.NoError(t, err)

		debugLog := logData["debug_log"].(map[string]interface{})
		assert.Equal(t, "warn", debugLog["level"])
		assert.Equal(t, "validation", debugLog["warning_type"])
		assert.Equal(t, "email", debugLog["field"])
	})

	// 测试Error级别日志
	t.Run("Error日志记录", func(t *testing.T) {
		output := captureOutput(func() {
			logger.Error("测试Error日志", map[string]interface{}{
				"error_code": "DB_CONNECTION_FAILED",
				"details":    "数据库连接超时",
			})
		})

		assert.Contains(t, output, "测试Error日志")
		assert.Contains(t, output, "error")

		var logData map[string]interface{}
		err := json.Unmarshal([]byte(output), &logData)
		assert.NoError(t, err)

		debugLog := logData["debug_log"].(map[string]interface{})
		assert.Equal(t, "error", debugLog["level"])
		assert.Equal(t, "DB_CONNECTION_FAILED", debugLog["error_code"])
		assert.Equal(t, "数据库连接超时", debugLog["details"])
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
				assert.Contains(t, output, tc.logLevel)
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

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(output), &logData)
	assert.NoError(t, err)

	debugLog := logData["debug_log"].(map[string]interface{})
	assert.Equal(t, "test_app", debugLog["app_name"])
	assert.Equal(t, "1.0.0", debugLog["version"])

	// 测试字段合并
	enhancedLogger := baseLogger.WithFields(map[string]interface{}{
		"user_id": 456,
		"action":  "purchase",
	})

	output = captureOutput(func() {
		enhancedLogger.Info("合并字段后的日志")
	})

	err = json.Unmarshal([]byte(output), &logData)
	assert.NoError(t, err)

	debugLog = logData["debug_log"].(map[string]interface{})
	assert.Equal(t, "test_app", debugLog["app_name"])
	assert.Equal(t, "1.0.0", debugLog["version"])
	assert.Equal(t, float64(456), debugLog["user_id"])
	assert.Equal(t, "purchase", debugLog["action"])
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

	var logData map[string]interface{}
	err = json.Unmarshal([]byte(output), &logData)
	assert.NoError(t, err)

	debugLog := logData["debug_log"].(map[string]interface{})
	assert.Equal(t, "test-request-123", debugLog["request_id"])
	assert.Equal(t, "GET", debugLog["method"])
	assert.Equal(t, "/api/test", debugLog["url"])
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

	// 验证输出包含错误信息
	assert.Contains(t, output, "logger_not_found_in_context")
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

	// 验证回退到简单格式输出
	assert.Contains(t, output, "测试无法序列化的字段")
	assert.Contains(t, output, "正常字段")
	// 由于JSON序列化失败，会使用fmt.Printf输出，包含时间戳和消息
	assert.True(t, strings.Contains(output, "[INFO]") || strings.Contains(output, "info"))
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

// captureOutput 捕获标准输出
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}
