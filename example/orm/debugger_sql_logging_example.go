package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/jcbaseGo/component/orm"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"
)

// DebuggerSQL日志记录配置示例
// 这个示例展示了如何正确配置debugger来记录GORM的SQL执行日志

func main() {
	fmt.Println("=== Debugger SQL日志记录配置示例 ===")
	fmt.Println("关键配置要点：")
	fmt.Println("1. debugger.Config.Level 必须设置为 LevelInfo 或更高")
	fmt.Println("2. GORM的日志级别也必须设置为 logger.Info 或更高")
	fmt.Println("3. 两者都满足时，SQL执行日志才会被记录")

	// 创建Gin引擎
	r := gin.Default()

	// ===========================================
	// 关键配置1：正确设置debugger的日志级别
	// ===========================================
	debuggerConfig := &debugger.Config{
		Enabled:    true,
		Level:      debugger.LevelInfo, // 必须设置为LevelInfo或更高才能记录SQL日志
		MaxRecords: 1000,
		SkipPaths:  []string{"/static/", "/favicon.ico"},
		AllowedIPs: []string{}, // 允许所有IP访问
	}

	// 创建debugger实例
	debug, err := debugger.New(debuggerConfig)
	if err != nil {
		panic(fmt.Sprintf("初始化debugger失败: %v", err))
	}

	// 将debugger中间件添加到Gin
	r.Use(debug.Middleware())

	// 数据库配置
	dbConfig := jcbaseGo.DbStruct{
		Host:      "localhost",
		Port:      "3306",
		Username:  "root",
		Password:  "password",
		Dbname:    "test_db",
		Charset:   "utf8mb4",
		ParseTime: "True",
		Protocol:  "tcp",
	}

	// ===========================================
	// 关键配置2：正确配置GORM的SQL日志记录
	// ===========================================

	// 方式1：使用NewWithDebugger创建数据库实例（推荐方式）
	// 这种方式会自动配置SQL日志记录，使用debugger的日志级别
	db := mysql.NewWithDebugger(dbConfig, debug.GetLogger())
	if db.Error() != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", db.Error()))
	}

	// 方式2：手动配置SQL日志记录（如果需要更精细的控制）
	// db := mysql.New(dbConfig)
	// db.EnableSQLLogging(debug.GetLogger(), debugger.LevelInfo, 200*time.Millisecond)

	// 方式3：使用Debug()方法启用调试模式
	// db.Debug().GetDb() // 这种方式会使用GORM的默认调试日志

	// 设置路由演示SQL日志记录
	r.GET("/test-sql-logging", func(c *gin.Context) {
		// 从上下文中获取debugger logger
		loggerInterface, exists := c.Get("debugger_logger")
		if !exists {
			c.JSON(500, gin.H{"error": "无法获取调试日志记录器"})
			return
		}

		logger := loggerInterface.(debugger.LoggerInterface)

		// 记录开始信息
		logger.Info("开始SQL日志记录测试")

		// 执行一些数据库操作来演示SQL日志记录

		// 1. 查询操作
		logger.Info("执行查询操作")
		var count int64
		db.GetDb().Table("users").Count(&count)

		// 2. 插入操作
		logger.Info("执行插入操作")
		user := map[string]interface{}{
			"name":       "测试用户",
			"email":      "test@example.com",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		db.GetDb().Table("users").Create(&user)

		// 3. 更新操作
		logger.Info("执行更新操作")
		db.GetDb().Table("users").Where("email = ?", "test@example.com").
			Update("name", "更新后的用户")

		// 4. 删除操作
		logger.Info("执行删除操作")
		db.GetDb().Table("users").Where("email = ?", "test@example.com").Delete(nil)

		logger.Info("SQL日志记录测试完成")

		c.JSON(200, gin.H{
			"message": "SQL日志记录测试完成",
			"count":   count,
		})
	})

	// 检查日志级别配置的路由
	r.GET("/check-log-level", func(c *gin.Context) {
		loggerInterface, exists := c.Get("debugger_logger")
		if !exists {
			c.JSON(500, gin.H{"error": "无法获取调试日志记录器"})
			return
		}

		logger := loggerInterface.(debugger.LoggerInterface)

		// 检查当前debugger的日志级别
		debuggerLevel := logger.GetLevel()

		// 检查GORM的日志级别配置
		gormLogger := db.GetDb().Config.Logger
		gormLevelConfigured := false
		if gormLogger != nil {
			// 检查是否配置了GORM日志记录器
			if _, ok := gormLogger.(*orm.GormDebuggerLogger); ok {
				gormLevelConfigured = true
			}
		}

		c.JSON(200, gin.H{
			"debugger_level":  debuggerLevel,
			"gorm_configured": gormLevelConfigured,
			"can_log_sql":     debuggerLevel >= debugger.LevelInfo && gormLevelConfigured,
			"message":         "如果can_log_sql为false，则SQL日志不会被记录",
		})
	})

	// 启动HTTP服务器
	fmt.Println("服务器启动在 :8080")
	fmt.Println("测试端点:")
	fmt.Println("  GET /test-sql-logging - 测试SQL日志记录功能")
	fmt.Println("  GET /check-log-level  - 检查当前日志级别配置")

	if err := r.Run(":8080"); err != nil {
		panic(fmt.Sprintf("服务器启动失败: %v", err))
	}
}

// 常见问题排查指南
/*
问题1：SQL日志没有被记录
原因：
  1. debugger的日志级别低于LevelInfo
  2. GORM的日志级别低于logger.Info
  3. 两者都低于要求级别

解决方案：
  1. 确保debugger.Config.Level设置为debugger.LevelInfo
  2. 确保GORM日志级别设置为logger.Info或更高
  3. 使用NewWithDebugger或EnableSQLLogging方法正确配置

问题2：只有错误SQL被记录，正常SQL没有被记录
原因：
  日志级别可能设置为LevelWarn或LevelError，只能记录警告和错误

解决方案：
  将日志级别提升到LevelInfo或LevelDebug

问题3：日志记录过于频繁，影响性能
解决方案：
  1. 在生产环境中降低日志级别到LevelWarn或LevelError
  2. 使用采样率配置减少日志记录频率
  3. 只记录慢查询或错误查询
*/
