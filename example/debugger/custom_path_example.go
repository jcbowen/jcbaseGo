package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// DebuggerCustomPathExample 演示如何自定义debugger基础路径
// 这个示例展示了多种自定义基础路径的方法
func main() {
	fmt.Println("=== Debugger 自定义基础路径示例 ===")

	// 创建Gin引擎
	router := gin.Default()

	// 示例1：使用自定义基础路径（推荐方式）
	example1(router)

	// 示例2：使用路由组嵌套方式
	example2(router)

	// 示例3：使用不同的基础路径配置
	example3(router)

	// 添加业务路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Debugger自定义基础路径示例",
			"debug_urls": []string{
				"/api/debug/list",
				"/admin/debug/list", 
				"/custom/debug/list",
			},
		})
	})

	fmt.Println("服务器启动在 :8080 端口")
	fmt.Println("可访问以下调试器路径：")
	fmt.Println("- /api/debug/list    (API调试器)")
	fmt.Println("- /admin/debug/list  (管理后台调试器)")
	fmt.Println("- /custom/debug/list (自定义调试器)")

	router.Run(":8080")
}

// example1 演示使用自定义基础路径（推荐方式）
func example1(router *gin.Engine) {
	// 创建调试器实例
	dbg, err := debugger.NewWithMemoryStorage(500)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 创建自定义控制器配置
	controllerConfig := &debugger.ControllerConfig{
		BasePath: "/api/debug", // 自定义基础路径
		Title:    "API调试器",   // 自定义页面标题
		PageSize: 30,           // 自定义每页显示数量
	}

	// 注册调试器控制器
	dbg.WithController(router, controllerConfig)

	// 添加调试器中间件到API路由组
	apiGroup := router.Group("/api")
	apiGroup.Use(dbg.Middleware())

	// 添加示例API路由
	apiGroup.GET("/users", func(c *gin.Context) {
		// 获取当前请求的Logger实例
		logger := debugger.GetLoggerFromContext(c)
		
		logger.Info("获取用户列表", map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		})
		
		c.JSON(http.StatusOK, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "张三"},
				{"id": 2, "name": "李四"},
			},
		})
	})
}

// example2 演示使用路由组嵌套方式
func example2(router *gin.Engine) {
	// 创建调试器实例
	dbg, err := debugger.NewWithMemoryStorage(300)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 创建自定义控制器配置，使用完整的基础路径
	controllerConfig := &debugger.ControllerConfig{
		BasePath: "/admin/debug", // 直接使用完整的基础路径
		Title:    "管理后台调试器",
		PageSize: 25,
	}

	// 注册调试器控制器
	dbg.WithController(router, controllerConfig)

	// 创建管理后台路由组
	adminGroup := router.Group("/admin")
	// 添加调试器中间件到管理后台路由组
	adminGroup.Use(dbg.Middleware())

	// 添加管理后台路由
	adminGroup.GET("/dashboard", func(c *gin.Context) {
		logger := debugger.GetLoggerFromContext(c)
		logger.Info("访问管理后台仪表板")
		
		c.JSON(http.StatusOK, gin.H{
			"message": "管理后台",
			"debug_url": "/admin/debug/list", // 最终访问路径
		})
	})
}

// example3 演示使用不同的基础路径配置
func example3(router *gin.Engine) {
	// 创建调试器实例
	dbg, err := debugger.NewWithMemoryStorage(200)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 方法3：先配置控制器，后注册路由
	controllerConfig := &debugger.ControllerConfig{
		BasePath: "/custom/debug", // 完全自定义的基础路径
		Title:    "自定义调试器",
	}

	// 先创建控制器（不立即注册路由）
	dbg.WithController(nil, controllerConfig)

	// 稍后手动注册路由
	dbg.RegisterRoutes(router)

	// 添加调试器中间件
	customGroup := router.Group("/custom")
	customGroup.Use(dbg.Middleware())

	// 添加自定义路由
	customGroup.GET("/test", func(c *gin.Context) {
		logger := debugger.GetLoggerFromContext(c)
		logger.Debug("自定义路由测试")
		
		c.JSON(http.StatusOK, gin.H{
			"message": "自定义路由",
			"debug_path": "/custom/debug",
		})
	})
}

/*
使用说明：

1. 方法1（推荐）：使用WithController方法直接配置基础路径
   - 优点：简单直接，配置清晰
   - 访问路径：/api/debug/list

2. 方法2：使用路由组嵌套方式
   - 优点：可以更好地组织路由结构
   - 访问路径：/admin/debug/list

3. 方法3：先配置后注册方式
   - 优点：灵活性高，可以在不同时机注册路由
   - 访问路径：/custom/debug/list

注意事项：
- 基础路径应该以斜杠开头，不以斜杠结尾
- 避免使用与其他路由冲突的路径
- 生产环境中建议使用更复杂的路径以提高安全性
*/