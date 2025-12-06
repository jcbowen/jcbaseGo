package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(100)" json:"name"`
	Email     string    `gorm:"column:email;type:varchar(100);unique" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

func main() {
	// 创建Gin引擎
	r := gin.Default()

	// 初始化debugger组件
	debuggerConfig := &debugger.Config{
		Enabled:    true,
		Level:      debugger.LevelInfo,
		MaxRecords: 1000,
		SkipPaths:  []string{"/static/", "/favicon.ico"},
		AllowedIPs: []string{}, // 允许所有IP访问
	}

	// 创建debugger实例（使用默认内存存储）
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

	// 方式1：使用NewWithDebugger创建数据库实例（推荐方式）
	// 这种方式会自动配置SQL日志记录，使用debugger的日志级别
	db, err := mysql.NewWithDebugger(dbConfig, debug.GetLogger())
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 方式2：先创建实例，后启用SQL日志记录（如果需要更精细的控制）
	// db, _ := mysql.New(dbConfig)
	// db.EnableSQLLogging(debug.GetLogger(), debugger.LevelInfo, 200*time.Millisecond)

	// 自动迁移表结构
	if err := db.GetDb().AutoMigrate(&User{}); err != nil {
		panic(fmt.Sprintf("表结构迁移失败: %v", err))
	}

	// 设置路由
	r.GET("/users", func(c *gin.Context) {
		// 从上下文中获取debugger logger
		loggerInterface, exists := c.Get("debugger_logger")
		if !exists {
			c.JSON(500, gin.H{"error": "无法获取调试日志记录器"})
			return
		}

		logger := loggerInterface.(debugger.LoggerInterface)
		logger.Info("开始查询用户列表")

		// 查询用户
		var users []User
		result := db.GetDb().Find(&users)
		if result.Error != nil {
			logger.Error("查询用户失败", map[string]interface{}{
				"error": result.Error.Error(),
			})
			c.JSON(500, gin.H{"error": "查询失败"})
			return
		}

		logger.Info("查询用户成功", map[string]interface{}{
			"count": len(users),
		})

		c.JSON(200, gin.H{
			"data":  users,
			"count": len(users),
		})
	})

	r.POST("/users", func(c *gin.Context) {
		// 从上下文中获取debugger logger
		loggerInterface, exists := c.Get("debugger_logger")
		if !exists {
			c.JSON(500, gin.H{"error": "无法获取调试日志记录器"})
			return
		}

		logger := loggerInterface.(debugger.LoggerInterface)

		// 解析请求数据
		var userData struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.ShouldBindJSON(&userData); err != nil {
			logger.Error("参数解析失败", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		logger.Info("开始创建用户", map[string]interface{}{
			"name":  userData.Name,
			"email": userData.Email,
		})

		// 创建用户
		user := User{
			Name:      userData.Name,
			Email:     userData.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := db.GetDb().Create(&user)
		if result.Error != nil {
			logger.Error("创建用户失败", map[string]interface{}{
				"error": result.Error.Error(),
				"user":  userData,
			})
			c.JSON(500, gin.H{"error": "创建失败"})
			return
		}

		logger.Info("创建用户成功", map[string]interface{}{
			"user_id": user.ID,
			"name":    user.Name,
		})

		c.JSON(201, gin.H{
			"message": "创建成功",
			"data":    user,
		})
	})

	r.GET("/slow-query", func(c *gin.Context) {
		// 演示慢查询检测
		loggerInterface, exists := c.Get("debugger_logger")
		if !exists {
			c.JSON(500, gin.H{"error": "无法获取调试日志记录器"})
			return
		}

		logger := loggerInterface.(debugger.LoggerInterface)
		logger.Info("执行慢查询演示")

		// 执行一个复杂的查询来模拟慢查询
		var result []map[string]interface{}
		db.GetDb().Raw("SELECT SLEEP(0.3) as sleep_time").Scan(&result)

		c.JSON(200, gin.H{
			"message": "慢查询执行完成",
			"result":  result,
		})
	})

	// 启动HTTP服务器
	fmt.Println("服务器启动在 :8080")
	fmt.Println("调试面板访问: http://localhost:8080/debugger")
	fmt.Println("API 端点:")
	fmt.Println("  GET  /users      - 查询用户列表")
	fmt.Println("  POST /users      - 创建用户")
	fmt.Println("  GET  /slow-query - 慢查询演示")

	if err := r.Run(":8080"); err != nil {
		panic(fmt.Sprintf("服务器启动失败: %v", err))
	}
}
