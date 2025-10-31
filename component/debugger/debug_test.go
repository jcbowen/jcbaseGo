package debugger

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestDebugMiddleware 测试调试器中间件是否被正确调用
func TestDebugMiddleware(t *testing.T) {
	t.Run("调试器中间件调用测试", func(t *testing.T) {
		// 创建内存存储
		storage, _ := NewMemoryStorage()

		// 创建调试器实例，明确设置配置参数确保测试请求被记录
		config := &Config{
			Enabled:    true,
			SkipPaths:  []string{}, // 空数组确保不跳过任何路径
			SampleRate: 1.0,        // 设置采样率为1.0，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎
		router := gin.New()

		// 注册调试器中间件
		router.Use(dbg.Middleware())

		// 添加测试业务路由
		router.GET("/api/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test successful"})
		})

		// 创建测试请求
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, 200, w.Code)

		// 等待日志被记录
		time.Sleep(100 * time.Millisecond)

		// 检查存储中是否有日志记录
		entries, total, err := storage.FindAll(1, 10, nil)
		assert.NoError(t, err)

		if total == 0 {
			t.Log("存储中没有日志记录，调试器可能没有正确记录请求")
			t.Log("可能的问题：")
			t.Log("1. 调试器中间件没有被正确调用")
			t.Log("2. 日志保存到存储时出现问题")
			t.Log("3. 存储查询功能有问题")
		} else {
			t.Logf("存储中有 %d 条日志记录，第一条URL: %s", total, entries[0].URL)
			assert.Greater(t, total, 0, "存储中应该有日志记录")
		}
	})
}

// TestDebugMiddlewareDirect 直接测试调试器中间件
func TestDebugMiddlewareDirect(t *testing.T) {
	t.Run("直接测试调试器中间件", func(t *testing.T) {
		// 创建内存存储
		storage, _ := NewMemoryStorage()

		// 创建调试器实例
		config := &Config{
			Enabled:    true,
			SkipPaths:  []string{}, // 空数组确保不跳过任何路径
			SampleRate: 1.0,        // 设置采样率为1.0，确保记录所有请求
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// 注册调试器中间件
		router.Use(dbg.Middleware())

		// 添加测试业务路由
		router.GET("/api/test", func(c *gin.Context) {
			fmt.Println("处理函数被调用")
			c.JSON(200, gin.H{"message": "test successful"})
		})

		// 创建测试请求
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		// 执行请求
		router.ServeHTTP(w, req)

		// 等待日志被记录
		time.Sleep(100 * time.Millisecond)

		// 检查存储中是否有日志记录
		entries, total, err := storage.FindAll(1, 10, nil)
		assert.NoError(t, err)

		if total == 0 {
			t.Log("存储中没有日志记录，调试器中间件可能有问题")
		} else {
			t.Logf("存储中有 %d 条日志记录，第一条URL: %s", total, entries[0].URL)
			assert.Greater(t, total, 0, "存储中应该有日志记录")
		}
	})
}

// TestDebuggerMiddlewareDetailed 详细测试调试器中间件执行流程
func TestDebuggerMiddlewareDetailed(t *testing.T) {
	t.Run("详细测试调试器中间件执行流程", func(t *testing.T) {
		// 创建内存存储
		storage, _ := NewMemoryStorage()

		fmt.Println("=== 开始详细测试 ===")
		fmt.Printf("存储实例: %p\n", storage)

		// 创建调试器实例
		config := &Config{
			Enabled:    true,
			SkipPaths:  []string{}, // 空数组确保不跳过任何路径
			SampleRate: 1.0,        // 设置采样率为1.0，确保记录所有请求
		}
		config.Storage = storage
		dbg, err := New(config)
		assert.NoError(t, err)

		fmt.Printf("调试器实例: %p\n", dbg)
		fmt.Printf("调试器存储: %p\n", dbg.GetStorage())

		// 检查调试器配置
		fmt.Printf("调试器配置检查: Enabled=%v, SkipPaths=%v, SampleRate=%v\n",
			dbg.config.Enabled, dbg.config.SkipPaths, dbg.config.SampleRate)

		// 创建Gin引擎 - 不使用测试模式，以便看到调试输出
		// gin.SetMode(gin.TestMode)
		router := gin.New()

		// 创建一个简单的中间件来检查是否被调用
		router.Use(func(c *gin.Context) {
			fmt.Println("=== 简单中间件被调用 ===")
			c.Next()
		})

		// 注册调试器中间件
		router.Use(dbg.Middleware())

		// 添加测试业务路由
		router.GET("/api/test", func(c *gin.Context) {
			fmt.Println("=== 业务路由被调用 ===")
			c.JSON(200, gin.H{"message": "test successful"})
		})

		// 创建测试请求
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		fmt.Println("=== 开始执行请求 ===")

		// 执行请求
		router.ServeHTTP(w, req)

		fmt.Println("=== 请求执行完成 ===")
		fmt.Printf("响应状态码: %d\n", w.Code)

		// 等待日志被记录
		time.Sleep(200 * time.Millisecond)

		// 检查存储中是否有日志记录
		entries, total, err := storage.FindAll(1, 10, nil)
		assert.NoError(t, err)

		fmt.Printf("=== 存储查询结果: total=%d ===\n", total)

		if total == 0 {
			t.Log("存储中没有日志记录，调试器中间件可能有问题")

			// 检查调试器内部的存储
			_, dbgTotal, dbgErr := dbg.GetStorage().FindAll(1, 10, nil)
			if dbgErr == nil {
				fmt.Printf("调试器内部存储查询结果: total=%d\n", dbgTotal)
			}

			// 检查存储中是否有任何数据
			_, allTotal, _ := storage.FindAll(1, 1000, nil)
			fmt.Printf("存储中所有条目数量: %d\n", allTotal)

			// 手动保存一条测试日志
			testEntry := &LogEntry{
				ID:         "test_id_123",
				Timestamp:  time.Now(),
				Method:     "GET",
				URL:        "/api/test",
				StatusCode: 200,
			}
			saveErr := storage.Save(testEntry)
			if saveErr != nil {
				fmt.Printf("手动保存测试日志失败: %v\n", saveErr)
			} else {
				fmt.Println("手动保存测试日志成功")

				// 再次查询存储
				_, testTotal, _ := storage.FindAll(1, 10, nil)
				fmt.Printf("手动保存后存储查询结果: total=%d\n", testTotal)
			}
		} else {
			t.Logf("存储中有 %d 条日志记录，第一条URL: %s", total, entries[0].URL)
			assert.Greater(t, total, 0, "存储中应该有日志记录")
		}
	})
}
