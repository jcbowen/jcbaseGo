package debugger

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestStreamingResponseDetection 测试流式响应检测功能
func TestStreamingResponseDetection(t *testing.T) {
	testCases := []struct {
		name             string
		contentType      string
		transferEncoding string
		expectedResult   bool
	}{
		{
			name:             "SSE响应",
			contentType:      "text/event-stream",
			transferEncoding: "",
			expectedResult:   true,
		},
		{
			name:             "分块传输编码",
			contentType:      "application/json",
			transferEncoding: "chunked",
			expectedResult:   true,
		},
		{
			name:             "普通JSON响应",
			contentType:      "application/json",
			transferEncoding: "",
			expectedResult:   false,
		},
		{
			name:             "HTML响应",
			contentType:      "text/html",
			transferEncoding: "",
			expectedResult:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试响应头
			headers := make(http.Header)
			if tc.contentType != "" {
				headers.Set("Content-Type", tc.contentType)
			}
			if tc.transferEncoding != "" {
				headers.Set("Transfer-Encoding", tc.transferEncoding)
			}

			// 测试流式响应检测
			result := isStreamingResponseFromHeaders(headers)
			assert.Equal(t, tc.expectedResult, result,
				"流式响应检测结果不符合预期: %s", tc.name)
		})
	}
}

// TestStreamingChunkRecording 测试流式响应分块记录功能
func TestStreamingChunkRecording(t *testing.T) {
	// 创建调试器配置，启用流式支持
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     1024, // 1KB
		MaxStreamingChunks:     5,    // 最多5个分块
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟SSE流式响应
	router.GET("/sse", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// 发送多个SSE事件
		for i := 1; i <= 3; i++ {
			c.Writer.Write([]byte(fmt.Sprintf("data: 事件 %d\\n\\n", i)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond) // 模拟延迟
		}
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/sse", nil)
	w := performRequest(router, req)

	// 验证响应状态
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证流式响应记录
	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)

	if total > 0 {
		entry := entries[0]
		assert.True(t, entry.IsStreamingResponse, "应该标记为流式响应")
		assert.Equal(t, "streaming", entry.RecordType, "记录类型应该为streaming")
		assert.Greater(t, entry.StreamingChunks, 0, "应该记录流式响应分块")
		assert.Contains(t, entry.ResponseBody, "Streaming Response",
			"响应体应该包含流式响应信息")
	}
}

// TestStreamingChunkSizeLimit 测试流式响应分块大小限制
func TestStreamingChunkSizeLimit(t *testing.T) {
	// 创建调试器配置，设置较小的分块大小限制
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     100, // 100字节限制
		MaxStreamingChunks:     10,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟大数据流式响应
	router.GET("/large-stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// 发送超过限制的数据
		largeData := strings.Repeat("A", 200) // 200字节数据
		c.Writer.Write([]byte(fmt.Sprintf("data: %s\\n\\n", largeData)))
		c.Writer.(http.Flusher).Flush()
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/large-stream", nil)
	w := performRequest(router, req)

	// 验证响应状态
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证流式响应记录
	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)

	if total > 0 {
		entry := entries[0]
		assert.True(t, entry.IsStreamingResponse, "应该标记为流式响应")
		assert.Equal(t, 100, entry.StreamingChunkSize, "分块大小限制应该为100字节")
		// 验证分块数据被正确截断
		assert.Contains(t, entry.ResponseBody, "Chunk", "应该记录分块信息")
	}
}

// TestMaxStreamingChunksLimit 测试最大流式响应分块数量限制
func TestMaxStreamingChunksLimit(t *testing.T) {
	// 创建调试器配置，设置较小的分块数量限制
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     1024,
		MaxStreamingChunks:     2, // 最多2个分块
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟多个SSE事件
	router.GET("/multi-chunk", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// 发送3个事件，但只应该记录前2个
		for i := 1; i <= 3; i++ {
			c.Writer.Write([]byte(fmt.Sprintf("data: 事件 %d\\n\\n", i)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond)
		}
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/multi-chunk", nil)
	w := performRequest(router, req)

	// 验证响应状态
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证流式响应记录
	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)

	if total > 0 {
		entry := entries[0]
		assert.True(t, entry.IsStreamingResponse, "应该标记为流式响应")
		assert.Equal(t, 2, entry.MaxStreamingChunks, "最大分块数量应该为2")
		assert.LessOrEqual(t, entry.StreamingChunks, 2, "记录的分块数量不应超过限制")
	}
}

// TestStreamingDisabled 测试禁用流式支持时的行为
func TestStreamingDisabled(t *testing.T) {
	// 创建调试器配置，禁用流式支持
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: false, // 禁用流式支持
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟SSE流式响应
	router.GET("/sse-disabled", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Writer.Write([]byte("data: 测试事件\\n\\n"))
		c.Writer.(http.Flusher).Flush()
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/sse-disabled", nil)
	w := performRequest(router, req)

	// 验证响应状态
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证流式响应记录（应该作为普通HTTP请求记录）
	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)

	if total > 0 {
		entry := entries[0]
		assert.False(t, entry.IsStreamingResponse, "不应该标记为流式响应")
		assert.Equal(t, "http", entry.RecordType, "记录类型应该为http")
	}
}

// TestBinaryStreamingData 测试二进制流式数据处理
func TestBinaryStreamingData(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     1024,
		MaxStreamingChunks:     5,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟二进制数据流
	router.GET("/binary-stream", func(c *gin.Context) {
		c.Header("Content-Type", "application/octet-stream")

		// 发送二进制数据
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
		c.Writer.Write(binaryData)
		c.Writer.(http.Flusher).Flush()
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/binary-stream", nil)
	w := performRequest(router, req)

	// 验证响应状态
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证二进制流式响应记录
	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)

	if total > 0 {
		entry := entries[0]
		assert.True(t, entry.IsStreamingResponse, "应该标记为流式响应")
		assert.Contains(t, entry.ResponseBody, "Binary Data",
			"应该标记二进制数据")
	}
}

// TestStreamingMetadata 测试流式响应元数据记录
func TestStreamingMetadata(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     512,
		MaxStreamingChunks:     3,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟流式响应
	router.GET("/metadata-test", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// 发送多个事件
		for i := 1; i <= 2; i++ {
			c.Writer.Write([]byte(fmt.Sprintf("data: 元数据测试事件 %d\\n\\n", i)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond)
		}
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/metadata-test", nil)
	w := performRequest(router, req)

	// 验证响应状态
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证流式响应元数据
	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)

	if total > 0 {
		entry := entries[0]
		assert.True(t, entry.IsStreamingResponse, "应该标记为流式响应")
		assert.Equal(t, 2, entry.StreamingChunks, "应该记录2个分块")
		assert.Equal(t, 512, entry.StreamingChunkSize, "分块大小限制应该为512字节")
		assert.Equal(t, 3, entry.MaxStreamingChunks, "最大分块数量应该为3")
		assert.Contains(t, entry.StreamingData, "Streaming Response",
			"应该包含流式响应摘要信息")

		// 验证存储大小计算包含流式元数据
		storageSize := entry.CalculateStorageSize()
		assert.NotEmpty(t, storageSize, "存储大小计算应该包含流式元数据")
	}
}

// performRequest 辅助函数：执行HTTP请求并返回响应
func performRequest(r http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
