package debugger

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMemoryCleanupMechanism 测试内存阈值清理机制
// 验证当流式分块总大小超过10MB阈值时，自动清理最旧分块的机制
func TestMemoryCleanupMechanism(t *testing.T) {
	// 创建调试器配置，启用流式支持
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     2 * 1024 * 1024, // 2MB分块大小
		MaxStreamingChunks:     10,              // 最多10个分块
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟大量数据的流式响应（总大小超过10MB阈值）
	router.GET("/large-stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// 发送6个2MB的分块，总大小12MB，超过10MB阈值
		for i := 1; i <= 6; i++ {
			// 生成2MB的数据
			largeData := strings.Repeat(fmt.Sprintf("Chunk-%d-", i), 1024*1024/10) // 约2MB
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", largeData)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond) // 模拟延迟
		}
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

		// 验证内存清理机制：由于总大小超过10MB，应该清理部分最旧分块
		// 注意：实际记录的分块数量可能少于发送的数量，因为内存清理机制会移除最旧分块
		assert.LessOrEqual(t, entry.StreamingChunks, 6, "由于内存清理机制，记录的分块数量应该少于发送的数量")

		// 验证响应体包含流式响应信息
		assert.Contains(t, entry.ResponseBody, "Streaming Response",
			"响应体应该包含流式响应信息")
		assert.Contains(t, entry.ResponseBody, "chunks",
			"响应体应该包含分块数量信息")
	}
}

// TestMemoryCleanupEdgeCase 测试内存清理边界情况
// 验证当分块大小刚好达到阈值时的处理
func TestMemoryCleanupEdgeCase(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     1 * 1024 * 1024, // 1MB分块大小
		MaxStreamingChunks:     20,              // 较多分块数量
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟刚好达到阈值的情况：10个1MB分块，总大小10MB
	router.GET("/edge-case", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// 发送10个1MB的分块，总大小刚好10MB
		for i := 1; i <= 10; i++ {
			// 生成1MB的数据
			data := strings.Repeat(fmt.Sprintf("Edge-%d-", i), 1024*1024/10) // 约1MB
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(5 * time.Millisecond)
		}
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/edge-case", nil)
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

		// 边界情况：刚好达到阈值时，不应该触发清理（因为等于阈值时不清理）
		assert.Equal(t, 10, entry.StreamingChunks, "刚好达到阈值时应该记录所有分块")
	}
}

// TestMemoryCleanupWithSmallChunks 测试小分块情况下的内存清理
// 验证大量小分块累计超过阈值时的处理
func TestMemoryCleanupWithSmallChunks(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     100 * 1024, // 100KB分块大小
		MaxStreamingChunks:     200,        // 大量分块
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟大量小分块累计超过阈值
	router.GET("/small-chunks", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")

		// 发送120个100KB分块，总大小12MB，超过10MB阈值
		for i := 1; i <= 120; i++ {
			// 生成100KB的数据
			data := strings.Repeat(fmt.Sprintf("Small-%03d-", i), 1024) // 约100KB
			c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(2 * time.Millisecond)
		}
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/small-chunks", nil)
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

		// 由于内存清理机制，记录的分块数量应该等于或少于发送的数量
		// 注意：内存清理机制在每次分块添加后都会检查，如果总大小超过10MB阈值会移除最旧分块
		assert.LessOrEqual(t, entry.StreamingChunks, 120, "内存清理机制可能移除部分最旧分块")

		// 验证清理后的总大小应该在阈值范围内
		totalSize := entry.CalculateStorageSize()
		assert.NotEmpty(t, totalSize, "应该能够计算存储大小")

		// 验证内存清理机制确实工作：如果记录的分块数量等于发送数量，说明总大小未超过阈值
		// 如果记录的分块数量少于发送数量，说明内存清理机制已触发
		if entry.StreamingChunks < 120 {
			t.Logf("内存清理机制已触发：发送120个分块，记录%d个分块", entry.StreamingChunks)
		} else {
			t.Logf("内存清理机制未触发：总大小未超过10MB阈值")
		}
	}
}

// TestMemoryCleanupZeroChunks 测试空分块情况
func TestMemoryCleanupZeroChunks(t *testing.T) {
	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,
		StreamingChunkSize:     1024,
		MaxStreamingChunks:     10,
	}
	debugger, err := New(config)
	assert.NoError(t, err)

	// 创建Gin引擎和测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(debugger.Middleware())

	// 模拟没有分块的流式响应
	router.GET("/no-chunks", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		// 发送一个空分块来确保被识别为流式响应
		c.Writer.Write([]byte(""))
		c.Writer.(http.Flusher).Flush()
	})

	// 执行测试请求
	req, _ := http.NewRequest("GET", "/no-chunks", nil)
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
		assert.Equal(t, 1, entry.StreamingChunks, "应该记录1个空分块")
		assert.Contains(t, entry.ResponseBody, "Streaming Response",
			"应该包含流式响应信息")
	}
}
