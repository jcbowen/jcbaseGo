package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// ControllerUsage 演示如何使用调试器控制器
// 这个示例展示了如何通过Web界面查看和管理调试日志
func main() {
	fmt.Println("=== Debugger 控制器使用示例 ===")

	// 创建Gin引擎
	r := gin.Default()

	// 创建调试器配置（使用内存存储器）
	memoryStorage, err := debugger.NewMemoryStorage(1000)
	if err != nil {
		panic(fmt.Sprintf("创建内存存储器失败: %v", err))
	}

	config := &debugger.Config{
		Enabled:         true,
		Storage:         memoryStorage,
		MaxRecords:      1000,
		RetentionPeriod: 24 * time.Hour,
		SampleRate:      1.0,
	}

	// 创建调试器实例
	debuggerInstance, err := debugger.New(config)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 创建控制器配置
	controllerConfig := &debugger.ControllerConfig{
		BasePath: "/debug",
		Title:    "调试器管理界面",
	}

	// 注册调试器控制器
	debuggerInstance.WithController(r, controllerConfig)

	// 添加调试器中间件到主路由
	r.Use(debuggerInstance.Middleware())

	// 添加业务路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "欢迎使用调试器控制器示例",
			"time":    time.Now().Format(time.RFC3339),
			"debug":   "访问 /debug 查看调试器管理界面",
		})
	})

	r.GET("/api/users", func(c *gin.Context) {
		// 模拟查询用户列表
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

	r.PUT("/api/users/:id", func(c *gin.Context) {
		id := c.Param("id")

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

		// 模拟更新用户
		time.Sleep(150 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"name":       user.Name,
			"email":      user.Email,
			"updated_at": time.Now().Format(time.RFC3339),
		})
	})

	r.DELETE("/api/users/:id", func(c *gin.Context) {
		id := c.Param("id")

		// 模拟删除用户
		time.Sleep(100 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("用户 %s 已删除", id),
		})
	})

	// 启动服务器（监听所有接口）
	fmt.Println("启动调试器控制器示例服务器...")
	fmt.Printf("调试器界面地址: http://localhost:8080%s\n", controllerConfig.BasePath)
	fmt.Printf("网络访问地址: http://%s:8080%s\n", getLocalIP(), controllerConfig.BasePath)
	fmt.Println("测试API地址: http://localhost:8080/")
	fmt.Println("访问 http://localhost:8080/api/users 测试用户API")

	r.Run("0.0.0.0:8080")
}

// getLocalIP 获取本地IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "localhost"
}
