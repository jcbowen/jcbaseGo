package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// DatabaseUsage 演示如何使用数据库存储的调试器
// 这个示例展示了如何将调试日志保存到SQLite数据库中
func main() {
	fmt.Println("=== Debugger 数据库存储使用示例 ===")

	// 创建SQLite数据库连接
	db, err := gorm.Open(sqlite.Open("debug.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("连接数据库失败: %v", err))
	}

	// 创建数据库存储器
	databaseStorage, err := debugger.NewDatabaseStorage(db, 1000, "debug_logs")
	if err != nil {
		panic(fmt.Sprintf("创建数据库存储器失败: %v", err))
	}

	// 创建调试器配置
	config := &debugger.Config{
		Enabled:         true,
		Storage:         databaseStorage,
		MaxRecords:      1000,
		RetentionPeriod: 24 * time.Hour,
		SampleRate:      1.0,
	}

	// 创建调试器实例
	debuggerInstance, err := debugger.New(config)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 创建Gin引擎
	r := gin.Default()

	// 创建控制器配置
	controllerConfig := &debugger.ControllerConfig{
		BasePath: "/debug",
		Title:    "调试器数据库存储管理界面",
	}

	// 注册调试器控制器
	debuggerInstance.WithController(r, controllerConfig)

	// 添加调试器中间件到主路由
	r.Use(debuggerInstance.Middleware())

	// 添加业务路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "欢迎使用调试器数据库存储示例",
			"time":    time.Now().Format(time.RFC3339),
			"debug":   "访问 /debug 查看调试器管理界面",
			"storage": "数据库存储",
		})
	})

	// 添加进程记录测试路由
	r.GET("/api/process/test", func(c *gin.Context) {
		// 开始进程记录
		processLogger := debuggerInstance.StartProcess("测试进程", "background")
		processID := processLogger.GetProcessID()

		// 记录一些进程日志
		processLogger.Info("进程开始执行")
		processLogger.Debug("初始化配置", map[string]interface{}{"config": "test"})

		// 模拟进程执行
		time.Sleep(100 * time.Millisecond)

		// 记录更多日志
		processLogger.Info("处理数据中")
		processLogger.Warn("遇到警告信息", map[string]interface{}{"warning": "test_warning"})

		// 结束进程记录
		if err := debuggerInstance.EndProcess(processID, "completed"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("结束进程记录失败: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "进程记录已创建并完成",
			"process_id": processID,
			"status":     "completed",
		})
	})

	// 添加进程状态更新路由
	r.POST("/api/process/:id/status", func(c *gin.Context) {
		processID := c.Param("id")

		var statusUpdate struct {
			Status string `json:"status"`
		}

		if err := c.BindJSON(&statusUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的请求数据",
			})
			return
		}

		// 结束进程记录并更新状态
		if err := debuggerInstance.EndProcess(processID, statusUpdate.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("更新进程状态失败: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "进程状态已更新",
			"process_id": processID,
			"status":     statusUpdate.Status,
		})
	})

	// 添加进程列表查询路由（测试筛选功能）
	r.GET("/api/process/list", func(c *gin.Context) {
		// 获取查询参数
		status := c.Query("status")
		processName := c.Query("process_name")

		// 构建筛选条件
		filters := make(map[string]interface{})

		if status != "" {
			filters["process_status"] = status
		}

		if processName != "" {
			filters["process_name"] = processName
		}

		// 查询进程记录
		entries, total, err := debuggerInstance.GetProcessRecords(1, 100, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("查询进程记录失败: %v", err),
			})
			return
		}

		// 提取进程信息
		processes := make([]gin.H, 0)
		for _, entry := range entries {
			processInfo := gin.H{
				"id":           entry.ID,
				"process_id":   entry.ProcessID,
				"process_name": entry.ProcessName,
				"process_type": entry.ProcessType,
				"status":       entry.Status,
				"timestamp":    entry.Timestamp.Format(time.RFC3339),
			}

			if !entry.EndTime.IsZero() {
				processInfo["end_time"] = entry.EndTime.Format(time.RFC3339)
			}

			if entry.Duration > 0 {
				processInfo["duration"] = entry.Duration.String()
			}

			processes = append(processes, processInfo)
		}

		c.JSON(http.StatusOK, gin.H{
			"total":     total,
			"processes": processes,
			"filters":   filters,
		})
	})

	// 启动服务器
	fmt.Println("启动调试器数据库存储示例服务器...")
	fmt.Printf("调试器界面地址: http://localhost:8081/debug\n")
	fmt.Printf("进程测试地址: http://localhost:8081/api/process/test\n")
	fmt.Printf("进程列表地址: http://localhost:8081/api/process/list\n")
	fmt.Printf("进程状态筛选测试: http://localhost:8081/api/process/list?status=running\n")
	fmt.Printf("进程名称筛选测试: http://localhost:8081/api/process/list?process_name=测试\n")

	r.Run("0.0.0.0:8081")
}
