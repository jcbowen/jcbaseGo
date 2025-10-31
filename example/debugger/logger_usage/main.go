package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// LoggerUsage 演示如何在控制器中使用debugger的logger功能
// 这个示例展示了如何通过GetLoggerFromContext在业务控制器中记录调试日志
func main() {
	fmt.Println("=== Debugger Logger 使用示例 ===")

	// 创建Gin引擎
	r := gin.Default()

	// 创建调试器实例（使用内存存储器，最多存储100条记录）
	debuggerInstance, err := debugger.NewWithMemoryStorage(100)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 添加调试器中间件
	r.Use(debuggerInstance.Middleware())

	// 示例1：基础路由，展示如何获取和使用logger
	r.GET("/", func(c *gin.Context) {
		// 从上下文中获取logger实例
		logger := debugger.GetLoggerFromContext(c)

		// 记录不同级别的日志
		logger.Debug("收到根路径请求", map[string]interface{}{
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"timestamp":  time.Now().Format(time.RFC3339),
		})

		logger.Info("处理根路径请求", map[string]interface{}{
			"method": c.Request.Method,
			"url":    c.Request.URL.String(),
		})

		c.JSON(http.StatusOK, gin.H{
			"message": "欢迎使用Debugger Logger示例",
			"time":    time.Now().Format(time.RFC3339),
		})

		logger.Info("根路径请求处理完成", map[string]interface{}{
			"status_code":   http.StatusOK,
			"response_time": time.Now().Format(time.RFC3339),
		})
	})

	// 示例2：用户API，展示更复杂的logger使用场景
	r.GET("/api/users", func(c *gin.Context) {
		logger := debugger.GetLoggerFromContext(c)

		logger.Info("开始处理用户列表请求", map[string]interface{}{
			"query_params": c.Request.URL.Query(),
			"page":         c.Query("page"),
			"limit":        c.Query("limit"),
		})

		// 模拟数据库查询
		logger.Debug("开始查询数据库", map[string]interface{}{
			"operation":  "query_users",
			"start_time": time.Now().Format(time.RFC3339),
		})

		time.Sleep(50 * time.Millisecond) // 模拟数据库查询时间

		logger.Debug("数据库查询完成", map[string]interface{}{
			"operation": "query_users",
			"end_time":  time.Now().Format(time.RFC3339),
			"duration":  "50ms",
		})

		// 返回用户列表
		users := []gin.H{
			{"id": 1, "name": "张三", "email": "zhangsan@example.com", "created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339)},
			{"id": 2, "name": "李四", "email": "lisi@example.com", "created_at": time.Now().Add(-48 * time.Hour).Format(time.RFC3339)},
			{"id": 3, "name": "王五", "email": "wangwu@example.com", "created_at": time.Now().Add(-72 * time.Hour).Format(time.RFC3339)},
		}

		logger.Info("用户列表查询成功", map[string]interface{}{
			"user_count":  len(users),
			"status_code": http.StatusOK,
		})

		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"total": len(users),
			"page":  1,
			"limit": 10,
		})
	})

	// 示例3：创建用户，展示错误处理和警告日志
	r.POST("/api/users", func(c *gin.Context) {
		logger := debugger.GetLoggerFromContext(c)

		logger.Info("开始处理创建用户请求", map[string]interface{}{
			"content_type":   c.ContentType(),
			"content_length": c.Request.ContentLength,
		})

		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.BindJSON(&user); err != nil {
			logger.Error("解析用户数据失败", map[string]interface{}{
				"error":        err.Error(),
				"request_body": "无法解析的JSON数据",
			})

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的请求数据",
			})
			return
		}

		// 验证用户数据
		if user.Name == "" || user.Email == "" {
			logger.Warn("用户数据验证失败", map[string]interface{}{
				"name":   user.Name,
				"email":  user.Email,
				"reason": "姓名或邮箱不能为空",
			})

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "姓名和邮箱不能为空",
			})
			return
		}

		// 模拟创建用户
		logger.Debug("开始创建用户", map[string]interface{}{
			"user_data": user,
			"operation": "create_user",
		})

		time.Sleep(100 * time.Millisecond) // 模拟创建用户时间

		logger.Info("用户创建成功", map[string]interface{}{
			"user_id":     4,
			"user_name":   user.Name,
			"user_email":  user.Email,
			"status_code": http.StatusCreated,
		})

		c.JSON(http.StatusCreated, gin.H{
			"id":         4,
			"name":       user.Name,
			"email":      user.Email,
			"created_at": time.Now().Format(time.RFC3339),
		})
	})

	// 示例4：带参数的路由，展示WithFields的使用
	r.GET("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")

		// 使用WithFields创建带有用户ID的logger
		logger := debugger.GetLoggerFromContext(c).WithFields(map[string]interface{}{
			"user_id":  userID,
			"endpoint": "/api/users/:id",
		})

		logger.Info("开始查询用户详情", map[string]interface{}{
			"operation": "get_user_by_id",
		})

		// 模拟用户不存在的情况
		if userID == "999" {
			logger.Warn("用户不存在", map[string]interface{}{
				"user_id": userID,
				"reason":  "数据库中未找到该用户",
			})

			c.JSON(http.StatusNotFound, gin.H{
				"error": "用户不存在",
			})
			return
		}

		// 模拟数据库查询
		time.Sleep(30 * time.Millisecond)

		logger.Info("用户详情查询成功", map[string]interface{}{
			"user_id":     userID,
			"status_code": http.StatusOK,
		})

		c.JSON(http.StatusOK, gin.H{
			"id":    userID,
			"name":  "示例用户",
			"email": fmt.Sprintf("user%s@example.com", userID),
		})
	})

	// 示例5：错误处理路由，展示错误日志记录
	r.GET("/api/error", func(c *gin.Context) {
		logger := debugger.GetLoggerFromContext(c)

		logger.Info("开始处理错误示例请求")

		// 模拟业务逻辑错误
		err := fmt.Errorf("数据库连接失败: 连接超时")

		logger.Error("业务处理失败", map[string]interface{}{
			"error":        err.Error(),
			"error_type":   "database_connection_timeout",
			"retry_count":  3,
			"last_attempt": time.Now().Format(time.RFC3339),
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "服务器内部错误",
			"details": "请稍后重试",
		})
	})

	// 启动服务器
	fmt.Println("启动Debugger Logger示例服务器...")
	fmt.Println("访问以下URL测试不同场景：")
	fmt.Println("1. GET http://localhost:8080/ - 基础logger使用")
	fmt.Println("2. GET http://localhost:8080/api/users - 用户列表查询")
	fmt.Println("3. POST http://localhost:8080/api/users - 创建用户（需要JSON数据）")
	fmt.Println("4. GET http://localhost:8080/api/users/123 - 查询用户详情")
	fmt.Println("5. GET http://localhost:8080/api/users/999 - 用户不存在场景")
	fmt.Println("6. GET http://localhost:8080/api/error - 错误处理场景")
	fmt.Println("")
	fmt.Println("查看控制台输出，观察logger记录的调试信息")

	r.Run(":8080")
}
