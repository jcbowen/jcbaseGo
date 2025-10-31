package debugger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ExampleControllerUsage 演示如何使用调试器控制器
func ExampleControllerUsage() {
	// 创建Gin引擎
	r := gin.Default()

	// 创建调试器配置
	memoryStorage, _ := NewMemoryStorage(1000)
	config := &Config{
		Enabled:         true,
		Storage:         memoryStorage,
		MaxRecords:      1000,
		RetentionPeriod: 24 * time.Hour,
		SampleRate:      1.0,
	}

	// 创建调试器实例
	debugger, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 创建控制器配置
	controllerConfig := &ControllerConfig{
		BasePath: "/jcbase/debugger",
		Title:    "调试器管理界面",
	}

	// 创建路由组并注册调试器控制器
	debuggerGroup := r.Group(controllerConfig.BasePath)

	// 方法1: 在创建调试器时直接传入路由组
	debugger.WithController(debuggerGroup, controllerConfig)

	// 方法2: 或者先创建调试器，再单独注册路由
	// debugger.RegisterRoutes(debuggerGroup)

	// 添加调试器中间件到主路由
	r.Use(debugger.Middleware())

	// 添加一些测试路由来生成调试日志
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "欢迎使用调试器示例",
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

	// 启动服务器
	fmt.Println("启动调试器示例服务器...")
	fmt.Printf("调试器界面地址: http://localhost:8080%s\n", controllerConfig.BasePath)
	fmt.Println("测试API地址: http://localhost:8080/")

	r.Run(":8080")
}

// ExampleAdvancedControllerUsage 演示高级控制器用法
func ExampleAdvancedControllerUsage() {
	// 创建Gin引擎
	r := gin.Default()

	// 创建文件存储的调试器
	fileStorage, _ := NewFileStorage("./debug_logs", 5000)
	config := &Config{
		Enabled:         true,
		Storage:         fileStorage,
		MaxRecords:      5000,
		RetentionPeriod: 7 * 24 * time.Hour, // 保留7天
		SampleRate:      0.5,                // 50%采样率
		SkipPaths:       []string{"/static", "/favicon.ico"},
		SkipMethods:     []string{"OPTIONS"},
	}

	// 创建调试器实例
	debugger, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 自定义控制器配置
	controllerConfig := &ControllerConfig{
		BasePath: "/admin/debug",
		Title:    "系统调试监控",
	}

	// 创建管理路由组（可以添加认证中间件）
	adminGroup := r.Group("/admin")
	// 这里可以添加认证中间件
	// adminGroup.Use(authMiddleware())

	// 注册调试器路由到管理组
	debugGroup := adminGroup.Group("/debug")
	debugger.WithController(debugGroup, controllerConfig)

	// 添加调试器中间件
	r.Use(debugger.Middleware())

	// 添加业务路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     "示例应用",
			"version": "1.0.0",
			"debug":   "访问 /admin/debug 查看调试信息",
		})
	})

	// 启动服务器
	fmt.Println("启动高级调试器示例服务器...")
	fmt.Printf("调试器界面地址: http://localhost:8080%s\n", controllerConfig.BasePath)
	fmt.Println("注意: 此示例使用文件存储，日志将保存在 ./debug_logs 目录")

	r.Run(":8080")
}

// ExampleControllerWithDatabase 演示使用数据库存储的控制器
func ExampleControllerWithDatabase() {
	// 注意: 这个示例需要实际的数据库连接
	// 这里只是展示如何配置

	r := gin.Default()

	// 创建数据库存储（需要实际的数据库连接）
	// db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	// if err != nil {
	//     panic("数据库连接失败")
	// }

	// storage, err := NewDatabaseStorage(db, 10000, "debug_logs")
	// if err != nil {
	//     panic(fmt.Sprintf("创建数据库存储失败: %v", err))
	// }

	memoryStorage, _ := NewMemoryStorage(10000)
	config := &Config{
		Enabled:         true,
		Storage:         memoryStorage,
		MaxRecords:      10000,
		RetentionPeriod: 30 * 24 * time.Hour,
		SampleRate:      1.0,
	}

	debugger, err := New(config)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 注册控制器
	debugger.WithController(r.Group("/debug"), &ControllerConfig{
		BasePath: "/debug",
		Title:    "数据库调试器",
	})

	r.Use(debugger.Middleware())

	fmt.Println("启动数据库调试器示例服务器...")
	fmt.Println("注意: 此示例需要配置实际的数据库连接")

	r.Run(":8080")
}
