package debugger

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestSimpleMiddleware 简单测试中间件是否被调用
func TestSimpleMiddleware(t *testing.T) {
	t.Run("简单中间件测试", func(t *testing.T) {
		fmt.Println("=== 开始简单中间件测试 ===")

		// 创建Gin引擎
		router := gin.New()

		// 添加一个简单的中间件来验证调用
		router.Use(func(c *gin.Context) {
			fmt.Println("=== 简单中间件被调用 ===")
			fmt.Printf("请求路径: %s\n", c.Request.URL.Path)
			c.Next()
		})

		// 添加业务路由
		router.GET("/api/test", func(c *gin.Context) {
			fmt.Println("业务路由处理函数被调用")
			c.JSON(200, gin.H{"message": "test successful"})
		})

		// 创建测试请求
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		// 执行请求
		fmt.Println("开始执行请求...")
		router.ServeHTTP(w, req)
		fmt.Println("请求执行完成")

		// 验证响应
		assert.Equal(t, 200, w.Code)

		fmt.Println("=== 简单中间件测试结束 ===")
	})
}

// TestDebuggerMiddlewareSimple 简单测试调试器中间件
func TestDebuggerMiddlewareSimple(t *testing.T) {
	t.Run("简单调试器中间件测试", func(t *testing.T) {
		fmt.Println("=== 开始简单调试器中间件测试 ===")

		// 创建内存存储
		storage, _ := NewMemoryStorage()

		// 创建调试器实例，确保所有配置都正确
		config := &Config{
			Enabled:     true,
			SkipPaths:   []string{}, // 空数组确保不跳过任何路径
			SkipMethods: []string{}, // 空数组确保不跳过任何方法
			SampleRate:  1.0,        // 100%采样率确保记录所有请求
			MaxRecords:  1000,       // 设置足够的记录数量
		}
		config.Storage = storage
		dbg, _ := New(config)

		// 创建Gin引擎
		router := gin.New()

		// 正确注册调试器中间件
		router.Use(dbg.Middleware())

		// 添加业务路由
		router.GET("/api/test", func(c *gin.Context) {
			fmt.Println("业务路由处理函数被调用")
			c.JSON(200, gin.H{"message": "test successful"})
		})

		// 创建测试请求
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		// 执行请求
		fmt.Println("开始执行请求...")
		router.ServeHTTP(w, req)
		fmt.Println("请求执行完成")

		// 验证响应
		assert.Equal(t, 200, w.Code)

		// 检查存储中是否有日志记录
		allEntries, total, err := storage.FindAll(1, 1000, nil)
		if err != nil {
			t.Errorf("查询存储失败: %v", err)
		}

		fmt.Printf("存储查询结果: total=%d, 条目数量=%d\n", total, len(allEntries))
		if total > 0 {
			fmt.Printf("第一条日志条目: %+v\n", allEntries[0])
		}

		fmt.Println("=== 简单调试器中间件测试结束 ===")
	})
}
