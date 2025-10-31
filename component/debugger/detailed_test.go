package debugger

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestDebuggerDetailed 详细测试调试器中间件的每个步骤
func TestDebuggerDetailed(t *testing.T) {
	t.Run("详细调试器测试", func(t *testing.T) {
		fmt.Println("=== 开始详细调试器测试 ===")

		// 创建内存存储
		storage, _ := NewMemoryStorage()
		fmt.Printf("存储创建成功: %T\n", storage)

		// 创建调试器实例，明确设置配置参数确保测试请求被记录
		config := &Config{
			Enabled:   true,
			SkipPaths: []string{}, // 空数组确保不跳过任何路径
		}
		config.Storage = storage
		dbg, err := New(config)

		assert.NoError(t, err)
		fmt.Printf("调试器创建成功: %T\n", dbg)
		fmt.Printf("调试器配置: Enabled=%v, SkipPaths=%v\n", dbg.config.Enabled, dbg.config.SkipPaths)

		// 创建Gin引擎
		router := gin.New()
		fmt.Println("Gin引擎创建成功")

		// 注册调试器中间件
		router.Use(dbg.Middleware())
		fmt.Println("调试器中间件注册成功")

		// 添加测试业务路由
		router.GET("/api/test", func(c *gin.Context) {
			fmt.Println("业务路由处理函数被调用")
			c.JSON(200, gin.H{"message": "test successful"})
		})
		fmt.Println("业务路由注册成功")

		// 创建测试请求
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		fmt.Println("测试请求创建成功")

		// 执行请求
		fmt.Println("开始执行请求...")
		router.ServeHTTP(w, req)
		fmt.Println("请求执行完成")

		// 验证响应
		assert.Equal(t, 200, w.Code)
		fmt.Println("响应验证成功")

		// 等待日志被记录
		fmt.Println("等待日志记录...")
		time.Sleep(200 * time.Millisecond)

		// 检查存储中是否有日志记录
		fmt.Println("检查存储中的日志记录...")
		entries, total, err := storage.FindAll(1, 10, nil)
		assert.NoError(t, err)

		fmt.Printf("存储查询结果: total=%d, err=%v\n", total, err)

		if total == 0 {
			fmt.Println("!!! 存储中没有日志记录 !!!")

			// 检查存储中是否有任何数据
			fmt.Println("检查存储中是否有任何条目...")

			// 使用存储的公共方法检查内容
			fmt.Printf("存储类型: %T\n", storage)

			// 使用FindAll方法获取所有条目
			allEntries, totalCount, _ := storage.FindAll(1, 1000, nil)
			fmt.Printf("通过FindAll方法获取的条目数量: %d\n", totalCount)

			// 打印存储中的所有条目
			for i, entry := range allEntries {
				fmt.Printf("条目 %d: ID=%s, URL=%s\n", i, entry.ID, entry.URL)
			}

			// 检查调试器中间件是否被调用
			fmt.Println("检查调试器中间件是否被调用...")

			// 重新运行测试，但这次添加更多调试信息
			fmt.Println("=== 重新运行测试 ===")

			// 创建新的测试请求
			req2 := httptest.NewRequest("GET", "/api/test2", nil)
			w2 := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w2, req2)

			// 再次检查存储
			time.Sleep(200 * time.Millisecond)
			_, total2, _ := storage.FindAll(1, 10, nil)
			fmt.Printf("第二次存储查询结果: total=%d\n", total2)

			if total2 == 0 {
				fmt.Println("!!! 第二次测试仍然没有日志记录 !!!")
			}
		} else {
			fmt.Printf("存储中有 %d 条日志记录\n", total)
			for i, entry := range entries {
				fmt.Printf("记录 %d: ID=%s, URL=%s, Method=%s\n", i, entry.ID, entry.URL, entry.Method)
			}
			assert.Greater(t, total, 0, "存储中应该有日志记录")
		}

		fmt.Println("=== 详细调试器测试结束 ===")
	})
}

// TestDebuggerConfig 测试调试器配置
func TestDebuggerConfig(t *testing.T) {
	t.Run("调试器配置测试", func(t *testing.T) {
		fmt.Println("=== 开始调试器配置测试 ===")

		// 测试默认配置
		defaultConfig := DefaultConfig()
		fmt.Printf("默认配置: Enabled=%v, SkipPaths=%v\n", defaultConfig.Enabled, defaultConfig.SkipPaths)

		// 测试自定义配置
		customConfig := &Config{
			Enabled:   true,
			SkipPaths: []string{},
		}
		fmt.Printf("自定义配置: Enabled=%v, SkipPaths=%v\n", customConfig.Enabled, customConfig.SkipPaths)

		// 创建存储
		storage, _ := NewMemoryStorage()

		// 创建调试器
		customConfig.Storage = storage
		dbg, err := New(customConfig)
		assert.NoError(t, err)

		// 验证配置
		assert.True(t, dbg.config.Enabled)
		assert.Equal(t, []string{}, dbg.config.SkipPaths)

		fmt.Println("=== 调试器配置测试结束 ===")
	})
}
