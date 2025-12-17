package debugger

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// errorReader 自定义的Reader，在读取时返回错误
// 用于测试请求体读取失败时的日志记录

type errorReader struct{}

// Read 实现io.Reader接口，返回错误
func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

// TestPostRequestLogging 测试POST请求的日志记录，确保查询参数和请求体都能正确记录
func TestPostRequestLogging(t *testing.T) {
	// 设置Gin模式为测试模式
	gin.SetMode(gin.TestMode)

	// 创建内存存储实例
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// 创建调试器配置
	config := &Config{
		Enabled:                true,
		SampleRate:             1.0,
		MaxRecords:             100,
		MaxBodySize:            1024,
		EnableMultipartSupport: true,
		Storage:                storage,
	}

	// 创建调试器实例
	dbg, err := New(config)
	assert.NoError(t, err)
	assert.NotNil(t, dbg)

	// 创建Gin引擎
	router := gin.Default()

	// 添加调试器中间件
	router.Use(dbg.Middleware())

	// 添加测试路由，接收POST请求并返回JSON响应
	router.POST("/api/test", func(c *gin.Context) {
		// 解析请求体
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 返回响应
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Request processed",
			"data":    requestBody,
		})
	})

	// 创建POST请求，包含查询参数和请求体
	requestBody := bytes.NewBuffer([]byte(`{"name": "test", "value": 123}`))
	req := httptest.NewRequest("POST", "/api/test?param1=value1&param2=value2", requestBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 等待日志被记录
	time.Sleep(200 * time.Millisecond)

	// 检查存储中是否有日志记录
	entries, total, err := storage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Greater(t, total, 0, "存储中应该有日志记录")

	// 检查第一条日志记录
	entry := entries[0]
	assert.Equal(t, "POST", entry.Method, "请求方法应该是POST")
	assert.Contains(t, entry.URL, "/api/test", "请求URL应该包含/api/test")
	assert.Equal(t, http.StatusOK, entry.StatusCode, "响应状态码应该是200")

	// 检查查询参数是否被记录
	assert.NotEmpty(t, entry.QueryParams, "查询参数应该被记录")
	assert.Equal(t, "value1", entry.QueryParams["param1"], "查询参数param1应该是value1")
	assert.Equal(t, "value2", entry.QueryParams["param2"], "查询参数param2应该是value2")

	// 检查请求体是否被记录
	assert.NotEmpty(t, entry.RequestBody, "请求体应该被记录")
	assert.Contains(t, entry.RequestBody, "name", "请求体应该包含name字段")
	assert.Contains(t, entry.RequestBody, "test", "请求体应该包含test值")
	assert.Contains(t, entry.RequestBody, "value", "请求体应该包含value字段")
	assert.Contains(t, entry.RequestBody, "123", "请求体应该包含123值")

	// 检查响应体是否被记录
	assert.NotEmpty(t, entry.ResponseBody, "响应体应该被记录")
	assert.Contains(t, entry.ResponseBody, "success", "响应体应该包含success字段")
}

// TestPostRequestWithError 测试POST请求在读取请求体失败时的日志记录
func TestPostRequestWithError(t *testing.T) {
	// 设置Gin模式为测试模式
	gin.SetMode(gin.TestMode)

	// 创建内存存储实例
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// 创建调试器配置
	config := &Config{
		Enabled:                true,
		SampleRate:             1.0,
		MaxRecords:             100,
		MaxBodySize:            1024,
		EnableMultipartSupport: true,
		Storage:                storage,
	}

	// 创建调试器实例
	dbg, err := New(config)
	assert.NoError(t, err)
	assert.NotNil(t, dbg)

	// 创建Gin引擎
	router := gin.Default()

	// 添加调试器中间件
	router.Use(dbg.Middleware())

	// 创建POST请求，使用错误Reader作为请求体
	req := httptest.NewRequest("POST", "/api/test?param1=value1", errorReader{})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 等待日志被记录
	time.Sleep(200 * time.Millisecond)

	// 检查存储中是否有日志记录
	entries, total, err := storage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Greater(t, total, 0, "存储中应该有日志记录")

	// 检查第一条日志记录
	entry := entries[0]
	assert.Equal(t, "POST", entry.Method, "请求方法应该是POST")
	assert.Contains(t, entry.URL, "/api/test", "请求URL应该包含/api/test")

	// 检查查询参数是否被记录
	assert.NotEmpty(t, entry.QueryParams, "查询参数应该被记录")
	assert.Equal(t, "value1", entry.QueryParams["param1"], "查询参数param1应该是value1")

	// 检查请求体是否被记录了错误信息
	assert.NotEmpty(t, entry.RequestBody, "请求体应该被记录")
	assert.Contains(t, entry.RequestBody, "Error reading request body", "请求体应该包含错误信息")
}
