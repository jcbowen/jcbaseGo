package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// BasicUsage 演示debugger组件的基本用法
// 这个示例展示了如何使用内存存储器的简单配置
func main() {
	fmt.Println("=== Debugger 基本使用示例 ===")

	// 创建Gin引擎
	r := gin.Default()

	// 创建调试器配置（使用内存存储器，最多存储100条记录）
	debuggerInstance, err := debugger.NewWithMemoryStorage(100)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 添加调试器中间件
	r.Use(debuggerInstance.Middleware())

	// 添加测试路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "欢迎使用Debugger示例",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	r.GET("/api/users", func(c *gin.Context) {
		// 模拟一些业务逻辑
		time.Sleep(100 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "张三", "email": "zhangsan@example.com"},
				{"id": 2, "name": "李四", "email": "lisi@example.com"},
				{"id": 3, "name": "王五", "email": "wangwu@example.com"},
			},
		})
	})

	r.POST("/api/users", func(c *gin.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的请求数据",
			})
			return
		}

		// 模拟创建用户
		time.Sleep(200 * time.Millisecond)

		c.JSON(http.StatusCreated, gin.H{
			"id":         4,
			"name":       user.Name,
			"email":      user.Email,
			"created_at": time.Now().Format(time.RFC3339),
		})
	})

	// 启动服务器
	fmt.Println("启动调试器示例服务器...")
	fmt.Println("访问 http://localhost:8080/ 测试API")
	fmt.Println("访问 http://localhost:8080/api/users 获取用户列表")
	fmt.Println("使用POST方法访问 http://localhost:8080/api/users 创建用户")

	r.Run(":8080")
}
