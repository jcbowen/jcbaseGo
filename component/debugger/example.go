package debugger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Example 使用示例
// 展示如何在Gin项目中使用调试器组件

// ExampleNewWay 新的配置方式示例
// 展示如何使用便捷的构造函数来创建调试器
func ExampleNewWay() {
	fmt.Println("=== 新的配置方式示例 ===")

	// 方式1：使用便捷构造函数（推荐）
	fmt.Println("方式1：使用便捷构造函数")
	debugger1, _ := NewSimpleDebugger()
	_ = debugger1

	// 方式2：使用内存存储器
	fmt.Println("方式2：使用内存存储器")
	debugger2, _ := NewWithMemoryStorage(200)
	_ = debugger2

	// 方式3：使用文件存储器
	fmt.Println("方式3：使用文件存储器")
	debugger3, _ := NewWithFileStorage("/tmp/debug_logs", 1000)
	_ = debugger3

	// 方式4：使用自定义存储器
	fmt.Println("方式4：使用自定义存储器")
	customStorage, _ := NewMemoryStorage(150)
	debugger4, _ := NewWithCustomStorage(customStorage)
	_ = debugger4

	// 方式5：使用生产环境配置
	fmt.Println("方式5：使用生产环境配置")
	debugger5, _ := NewProductionDebugger("/var/log/debug_logs")
	_ = debugger5

	fmt.Println("新的配置方式示例完成")
}

// ExampleMemoryStorage 内存存储器使用示例（向后兼容）
func ExampleMemoryStorage() {
	fmt.Println("=== 内存存储器使用示例 ===")

	// 新的推荐方式
	debugger, _ := NewWithMemoryStorage(100)

	// 创建Gin中间件
	middleware := NewMiddleware(debugger)

	// 创建Gin路由
	router := gin.New()
	router.Use(middleware.Handler())

	// 添加测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Debugger!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	fmt.Println("内存存储器示例创建完成")
	fmt.Println("访问 http://localhost:8080/test 测试调试器功能")
}

// ExampleFileStorage 文件存储器使用示例
func ExampleFileStorage() {
	fmt.Println("=== 文件存储器使用示例 ===")

	// 创建文件存储器
	storage, _ := NewFileStorage("/tmp/debug_logs", 1000)

	// 创建调试器
	config := &Config{
		Enabled: true,
	}
	config.Storage = storage
	debugger, _ := New(config)

	// 使用快捷方式创建中间件
	router := gin.New()
	router.Use(EnableMiddleware(storage))

	// 使用debugger变量避免编译警告
	_ = debugger

	// 添加测试路由
	router.POST("/api/users", func(c *gin.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的请求数据",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":    1,
			"name":  user.Name,
			"email": user.Email,
		})
	})

	fmt.Println("文件存储器示例创建完成")
	fmt.Println("日志将保存到 /tmp/debug_logs 目录")
}

// ExampleDatabaseStorage 数据库存储器使用示例
func ExampleDatabaseStorage(db *gorm.DB) {
	fmt.Println("=== 数据库存储器使用示例 ===")

	// 创建数据库存储器
	storage, err := NewDatabaseStorage(db, 1000, "debug_logs")
	if err != nil {
		log.Fatalf("创建数据库存储器失败: %v", err)
	}

	// 创建调试器
	config := &Config{
		Enabled: true,
	}
	config.Storage = storage
	debugger, _ := New(config)

	// 创建查询管理器
	queryManager := NewQueryManager(storage)

	// 使用debugger变量避免编译警告
	_ = debugger

	// 创建详情查看器
	detailViewer := NewDetailViewer(queryManager)

	// 创建Gin路由
	router := gin.New()
	router.Use(EnableMiddleware(storage))

	// 添加调试管理接口
	router.GET("/debug/logs", func(c *gin.Context) {
		// 获取查询参数
		page := getIntParam(c, "page", 1)
		pageSize := getIntParam(c, "page_size", 20)

		// 构建查询选项
		options := QueryOptions{
			Page:      page,
			PageSize:  pageSize,
			Filters:   buildFilters(c),
			SortBy:    c.DefaultQuery("sort_by", "timestamp"),
			SortOrder: c.DefaultQuery("sort_order", "desc"),
		}

		// 执行查询
		result, err := queryManager.Query(options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查询日志失败",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	router.GET("/debug/logs/:id", func(c *gin.Context) {
		id := c.Param("id")

		// 获取日志详情
		detail, err := detailViewer.GetDetail(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "日志条目不存在",
			})
			return
		}

		c.JSON(http.StatusOK, detail)
	})

	router.GET("/debug/stats", func(c *gin.Context) {
		// 获取统计信息
		stats, err := queryManager.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取统计信息失败",
			})
			return
		}

		c.JSON(http.StatusOK, stats)
	})

	fmt.Println("数据库存储器示例创建完成")
	fmt.Println("访问 /debug/logs 查看日志列表")
	fmt.Println("访问 /debug/logs/:id 查看日志详情")
	fmt.Println("访问 /debug/stats 查看统计信息")
}

// ExampleQueryManager 查询管理器使用示例
func ExampleQueryManager(storage Storage) {
	fmt.Println("=== 查询管理器使用示例 ===")

	// 创建查询管理器
	queryManager := NewQueryManager(storage)

	// 1. 查询最近的日志
	fmt.Println("1. 查询最近的10条日志:")
	recentLogs, err := queryManager.GetRecent(10)
	if err != nil {
		log.Printf("查询最近日志失败: %v", err)
	} else {
		fmt.Printf("找到 %d 条日志\n", len(recentLogs))
		for i, log := range recentLogs {
			fmt.Printf("  %d. %s %s (%d)\n", i+1, log.Method, log.URL, log.StatusCode)
		}
	}

	// 2. 按时间范围查询
	fmt.Println("\n2. 查询今天内的日志:")
	startTime := time.Now().Truncate(24 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)

	timeResult, err := queryManager.GetByTimeRange(startTime, endTime, 1, 10)
	if err != nil {
		log.Printf("按时间范围查询失败: %v", err)
	} else {
		fmt.Printf("找到 %d 条日志 (共 %d 条)\n", len(timeResult.Entries), timeResult.Total)
	}

	// 3. 按方法查询
	fmt.Println("\n3. 查询POST请求:")
	postResult, err := queryManager.GetByMethod("POST", 1, 10)
	if err != nil {
		log.Printf("按方法查询失败: %v", err)
	} else {
		fmt.Printf("找到 %d 条POST请求\n", postResult.Total)
	}

	// 4. 查询错误日志
	fmt.Println("\n4. 查询包含错误的日志:")
	errorResult, err := queryManager.GetErrors(1, 10)
	if err != nil {
		log.Printf("查询错误日志失败: %v", err)
	} else {
		fmt.Printf("找到 %d 条错误日志\n", errorResult.Total)
	}

	// 5. 搜索日志
	fmt.Println("\n5. 搜索包含用户关键词的日志:")
	searchResult, err := queryManager.Search("用户", 1, 10)
	if err != nil {
		log.Printf("搜索日志失败: %v", err)
	} else {
		fmt.Printf("找到 %d 条相关日志\n", searchResult.Total)
	}

	// 6. 获取统计信息
	fmt.Println("\n6. 获取统计信息:")
	stats, err := queryManager.GetStats()
	if err != nil {
		log.Printf("获取统计信息失败: %v", err)
	} else {
		fmt.Printf("统计信息: %+v\n", stats)
	}
}

// ExampleDetailViewer 详情查看器使用示例
func ExampleDetailViewer(queryManager *QueryManager) {
	fmt.Println("=== 详情查看器使用示例 ===")

	// 创建详情查看器
	detailViewer := NewDetailViewer(queryManager)

	// 获取最近的日志用于测试
	recentLogs, err := queryManager.GetRecent(1)
	if err != nil || len(recentLogs) == 0 {
		fmt.Println("没有找到日志用于测试")
		return
	}

	logID := recentLogs[0].ID

	// 1. 获取JSON格式的详情
	fmt.Println("1. 获取JSON格式的日志详情:")
	detailJSON, err := detailViewer.GetDetailJSON(logID)
	if err != nil {
		log.Printf("获取JSON详情失败: %v", err)
	} else {
		fmt.Printf("详情JSON长度: %d 字节\n", len(detailJSON))
	}

	// 2. 获取HTML格式的详情
	fmt.Println("\n2. 获取HTML格式的日志详情:")
	detailHTML, err := detailViewer.GetDetailHTML(logID)
	if err != nil {
		log.Printf("获取HTML详情失败: %v", err)
	} else {
		fmt.Printf("详情HTML长度: %d 字节\n", len(detailHTML))
	}

	// 3. 获取普通详情
	fmt.Println("\n3. 获取普通格式的日志详情:")
	detail, err := detailViewer.GetDetail(logID)
	if err != nil {
		log.Printf("获取详情失败: %v", err)
	} else {
		fmt.Printf("日志ID: %s\n", detail.LogEntry.ID)
		fmt.Printf("请求方法: %s\n", detail.LogEntry.Method)
		fmt.Printf("URL: %s\n", detail.LogEntry.URL)
		fmt.Printf("状态码: %d\n", detail.LogEntry.StatusCode)
		fmt.Printf("持续时间: %s\n", detail.FormattedDuration)
		fmt.Printf("请求大小: %d 字节\n", detail.RequestSize)
		fmt.Printf("响应大小: %d 字节\n", detail.ResponseSize)
	}
}

// ExampleProductionConfig 生产环境配置示例
func ExampleProductionConfig() {
	fmt.Println("=== 生产环境配置示例 ===")

	// 生产环境推荐使用文件或数据库存储器
	storage, err := NewFileStorage("/var/log/debug_logs", 5000)
	if err != nil {
		log.Fatalf("创建文件存储器失败: %v", err)
	}

	// 生产环境配置：只记录基本信息，不记录请求体和响应体
	config := &Config{
		Enabled: true,
	}

	config.Storage = storage
	debugger, _ := New(config)
	_ = debugger // 使用debugger变量避免编译警告

	// 使用简化的请求日志中间件
	router := gin.New()
	router.Use(RequestLogger(storage))

	fmt.Println("生产环境配置完成")
	fmt.Println("只记录基本信息，不记录敏感数据")
}

// ExampleDevelopmentConfig 开发环境配置示例
func ExampleDevelopmentConfig() {
	fmt.Println("=== 开发环境配置示例 ===")

	// 开发环境可以使用内存存储器，便于调试
	storage, _ := NewMemoryStorage()

	// 开发环境配置：记录所有详细信息
	config := &Config{
		Enabled: true,
	}

	config.Storage = storage
	debugger, _ := New(config)
	_ = debugger // 使用debugger变量避免编译警告

	// 使用详细的调试日志中间件
	router := gin.New()
	router.Use(DetailedDebugLogger(storage))

	fmt.Println("开发环境配置完成")
	fmt.Println("记录所有详细信息，便于调试")
}

// getIntParam 从查询参数获取整数值
func getIntParam(c *gin.Context, key string, defaultValue int) int {
	value := c.DefaultQuery(key, fmt.Sprintf("%d", defaultValue))
	var result int
	fmt.Sscanf(value, "%d", &result)
	if result <= 0 {
		result = defaultValue
	}
	return result
}

// buildFilters 从查询参数构建过滤器
func buildFilters(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})

	// 方法过滤
	if method := c.Query("method"); method != "" {
		filters["method"] = method
	}

	// 状态码过滤
	if statusCode := c.Query("status_code"); statusCode != "" {
		var code int
		fmt.Sscanf(statusCode, "%d", &code)
		if code > 0 {
			filters["status_code"] = code
		}
	}

	// URL包含过滤
	if urlContains := c.Query("url_contains"); urlContains != "" {
		filters["url_contains"] = urlContains
	}

	// 错误过滤
	if hasError := c.Query("has_error"); hasError != "" {
		filters["has_error"] = hasError == "true"
	}

	// 时间范围过滤
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filters["start_time"] = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filters["end_time"] = t
		}
	}

	return filters
}

// ExampleLoggerUsage Logger使用示例
// 展示如何在控制器中使用Logger记录不同级别的日志
func ExampleLoggerUsage() {
	fmt.Println("=== Logger使用示例 ===")

	// 创建调试器
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelDebug, // 设置为调试级别，记录所有日志
	}
	debugger, _ := New(config)

	// 创建Gin路由
	router := gin.New()
	router.Use(debugger.Middleware())

	// 示例1：基础日志记录
	router.GET("/api/users", func(c *gin.Context) {
		// 从上下文中获取Logger
		logger := GetLoggerFromContext(c)

		logger.Info("开始处理用户列表请求")

		// 模拟业务逻辑
		logger.Debug("查询数据库获取用户列表")

		// 模拟数据处理
		logger.Info("成功获取用户数据", map[string]interface{}{
			"user_count": 10,
			"page":       1,
			"page_size":  20,
		})

		c.JSON(http.StatusOK, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "张三"},
				{"id": 2, "name": "李四"},
			},
		})

		logger.Info("用户列表请求处理完成")
	})

	// 示例2：错误处理日志记录
	router.POST("/api/users", func(c *gin.Context) {
		logger := GetLoggerFromContext(c)

		logger.Info("开始创建新用户")

		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			logger.Error("请求数据解析失败", map[string]interface{}{
				"error": err.Error(),
			})

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的请求数据",
			})
			return
		}

		// 模拟验证失败
		if user.Name == "" {
			logger.Warn("用户名为空", map[string]interface{}{
				"user_data": user,
			})

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "用户名不能为空",
			})
			return
		}

		// 模拟数据库操作失败
		logger.Debug("开始保存用户数据到数据库")

		// 模拟保存成功
		logger.Info("用户创建成功", map[string]interface{}{
			"user_id":    1001,
			"user_name":  user.Name,
			"user_email": user.Email,
		})

		c.JSON(http.StatusCreated, gin.H{
			"id":    1001,
			"name":  user.Name,
			"email": user.Email,
		})
	})

	// 示例3：使用WithFields创建带上下文的Logger
	router.GET("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")

		// 创建带有用户ID字段的Logger
		logger := GetLoggerFromContext(c).WithFields(map[string]interface{}{
			"user_id": userID,
		})

		logger.Info("开始查询用户详情")

		// 模拟用户不存在
		if userID == "999" {
			logger.Warn("用户不存在", map[string]interface{}{
				"searched_id": userID,
			})

			c.JSON(http.StatusNotFound, gin.H{
				"error": "用户不存在",
			})
			return
		}

		logger.Debug("从数据库查询用户信息")

		// 模拟查询成功
		logger.Info("用户详情查询成功", map[string]interface{}{
			"user_name": "王五",
			"user_role": "管理员",
		})

		c.JSON(http.StatusOK, gin.H{
			"id":   userID,
			"name": "王五",
			"role": "管理员",
		})
	})

	fmt.Println("Logger使用示例创建完成")
	fmt.Println("访问以下端点测试Logger功能:")
	fmt.Println("  GET  http://localhost:8080/api/users")
	fmt.Println("  POST http://localhost:8080/api/users")
	fmt.Println("  GET  http://localhost:8080/api/users/123")
}

// ExampleCustomLogger 自定义Logger使用示例
func ExampleCustomLogger() {
	fmt.Println("=== 自定义Logger使用示例 ===")

	// 创建自定义Logger实现
	customLogger := &CustomLogger{
		prefix: "[CUSTOM] ",
	}

	// 创建调试器
	memoryStorage, _ := NewMemoryStorage(100)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
	}
	config.Logger = customLogger
	debugger, _ := New(config)

	// 创建Gin路由
	router := gin.New()
	router.Use(debugger.Middleware())

	router.GET("/test", func(c *gin.Context) {
		logger := GetLoggerFromContext(c)

		logger.Info("使用自定义Logger记录日志")
		logger.Error("这是一个错误日志示例")

		c.JSON(http.StatusOK, gin.H{
			"message": "自定义Logger测试",
		})
	})

	fmt.Println("自定义Logger示例创建完成")
}

// CustomLogger 自定义Logger实现示例
type CustomLogger struct {
	prefix string
}

func (l *CustomLogger) Debug(msg string, fields ...map[string]interface{}) {
	fmt.Printf("%s[DEBUG] %s", l.prefix, msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields[0])
	}
	fmt.Println()
}

func (l *CustomLogger) Info(msg string, fields ...map[string]interface{}) {
	fmt.Printf("%s[INFO] %s", l.prefix, msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields[0])
	}
	fmt.Println()
}

func (l *CustomLogger) Warn(msg string, fields ...map[string]interface{}) {
	fmt.Printf("%s[WARN] %s", l.prefix, msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields[0])
	}
	fmt.Println()
}

func (l *CustomLogger) Error(msg string, fields ...map[string]interface{}) {
	fmt.Printf("%s[ERROR] %s", l.prefix, msg)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields[0])
	}
	fmt.Println()
}

func (l *CustomLogger) WithFields(fields map[string]interface{}) Logger {
	// 简单的实现，返回自身（实际应用中可能需要更复杂的实现）
	return l
}

// RunAllExamples 运行所有示例
func RunAllExamples() {
	fmt.Println("开始运行调试器组件示例...")

	// 运行内存存储器示例
	ExampleMemoryStorage()

	// 运行文件存储器示例
	ExampleFileStorage()

	// 运行查询管理器示例（需要先有日志数据）
	storage, _ := NewMemoryStorage()
	queryManager := NewQueryManager(storage)
	ExampleQueryManager(storage)

	// 运行详情查看器示例
	ExampleDetailViewer(queryManager)

	// 运行环境配置示例
	ExampleProductionConfig()
	ExampleDevelopmentConfig()

	// 运行Logger使用示例
	ExampleLoggerUsage()
	ExampleCustomLogger()

	fmt.Println("\n所有示例运行完成!")
	fmt.Println("请参考具体示例代码了解如何使用调试器组件")
}
