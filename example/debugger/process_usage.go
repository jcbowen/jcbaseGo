package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// processWorker 模拟一个后台工作进程
func processWorker(debuggerInstance *debugger.Debugger, processName string) {
	// 启动工作进程
	logger, err := debuggerInstance.StartProcess(processName, "worker")
	if err != nil {
		log.Printf("启动工作进程失败: %v", err)
		return
	}

	// 记录进程开始信息
	logger.Info("进程开始执行", debugger.Fields{
		"worker_id":  "worker-001",
		"start_time": time.Now().Format(time.RFC3339),
	})

	// 模拟工作步骤1
	logger.Debug("步骤1: 数据预处理", debugger.Fields{
		"data_size": 1024,
		"operation": "preprocessing",
	})
	time.Sleep(100 * time.Millisecond)

	// 模拟工作步骤2
	logger.Info("步骤2: 数据处理", debugger.Fields{
		"records_processed": 500,
		"status":            "in_progress",
	})
	time.Sleep(200 * time.Millisecond)

	// 模拟工作步骤3
	logger.Warn("步骤3: 发现警告", debugger.Fields{
		"warning_type":     "data_inconsistency",
		"affected_records": 5,
	})
	time.Sleep(150 * time.Millisecond)

	// 模拟工作步骤4
	logger.Error("步骤4: 处理错误", debugger.Fields{
		"error_code":    "E001",
		"error_message": "数据库连接超时",
	})
	time.Sleep(100 * time.Millisecond)

	// 记录进程结束信息
	logger.Info("进程执行完成", debugger.Fields{
		"end_time":       time.Now().Format(time.RFC3339),
		"total_duration": time.Since(logger.GetStartTime()).String(),
	})

	// 结束进程记录
	err = debuggerInstance.EndProcess(logger.GetProcessID(), "completed")
	if err != nil {
		log.Printf("结束进程记录失败: %v", err)
	}
}

// batchProcessor 模拟批量处理进程
func batchProcessor(debuggerInstance *debugger.Debugger) {
	// 启动批量处理进程
	logger, err := debuggerInstance.StartProcess("batch_processor", "batch")
	if err != nil {
		log.Printf("启动批量处理进程失败: %v", err)
		return
	}

	logger.Info("批量处理开始", debugger.Fields{
		"batch_id":    "batch-2024-01",
		"total_files": 10,
	})

	// 模拟处理多个文件
	for i := 1; i <= 10; i++ {
		logger.Debug(fmt.Sprintf("处理文件 %d", i), debugger.Fields{
			"file_name": fmt.Sprintf("data_%d.csv", i),
			"file_size": i * 1024,
		})

		// 模拟处理时间
		time.Sleep(50 * time.Millisecond)

		// 每处理3个文件记录一次进度
		if i%3 == 0 {
			logger.Info("处理进度", debugger.Fields{
				"processed": i,
				"remaining": 10 - i,
				"progress":  fmt.Sprintf("%d%%", i*10),
			})
		}

		// 模拟错误情况
		if i == 7 {
			logger.Error("文件处理失败", debugger.Fields{
				"file_name": fmt.Sprintf("data_%d.csv", i),
				"error":     "文件格式不匹配",
			})
		}
	}

	logger.Info("批量处理完成", debugger.Fields{
		"successful_files": 9,
		"failed_files":     1,
		"total_duration":   time.Since(logger.GetStartTime()).String(),
	})

	debuggerInstance.EndProcess(logger.GetProcessID(), "completed")
}

// main 主函数，演示进程级debugger的使用
func main() {
	// 创建调试器实例（使用内存存储）
	debuggerInstance, err := debugger.NewWithMemoryStorage(&debugger.Config{
		Enabled:         true,
		LogLevel:        debugger.LevelDebug,
		MaxBodySize:     1024 * 1024, // 1MB
		RetentionPeriod: 24 * time.Hour,
	})
	if err != nil {
		log.Fatal("创建调试器失败:", err)
	}

	// 创建Gin路由引擎
	router := gin.Default()

	// 注册调试器控制器
	controller := debugger.NewController(debuggerInstance, router, &debugger.ControllerConfig{
		BasePath: "/debug",
		Title:    "进程调试器",
		PageSize: 20,
	})

	// 添加HTTP请求调试中间件
	router.Use(debuggerInstance.Middleware())

	// 定义API路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "进程级debugger示例",
			"debug_url": "/debug/list",
		})
	})

	// 启动工作进程的API
	router.POST("/start-worker", func(c *gin.Context) {
		processName := c.DefaultQuery("name", "default_worker")

		// 异步启动工作进程
		go processWorker(debuggerInstance, processName)

		c.JSON(http.StatusOK, gin.H{
			"message":      "工作进程已启动",
			"process_name": processName,
		})
	})

	// 启动批量处理的API
	router.POST("/start-batch", func(c *gin.Context) {
		// 异步启动批量处理
		go batchProcessor(debuggerInstance)

		c.JSON(http.StatusOK, gin.H{
			"message": "批量处理已启动",
		})
	})

	// 获取进程记录的API
	router.GET("/process-records", func(c *gin.Context) {
		// 获取进程记录（第一页，每页20条，无过滤条件）
		records, total, err := debuggerInstance.GetProcessRecords(1, 20, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "获取进程记录失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"process_records": records,
			"total":           total,
		})
	})

	// 模拟HTTP请求和进程记录同时运行
	router.GET("/mixed-operation", func(c *gin.Context) {
		// 这个HTTP请求本身会被debugger记录

		// 同时启动一个后台进程
		go func() {
			logger, _ := debuggerInstance.StartProcess("mixed_operation_worker", "background")
			logger.Info("混合操作中的后台进程", debugger.Fields{
				"http_request_id": c.GetHeader("X-Request-ID"),
			})
			time.Sleep(100 * time.Millisecond)
			logger.Debug("后台任务完成")
			debuggerInstance.EndProcess(logger.GetProcessID(), "completed")
		}()

		c.JSON(http.StatusOK, gin.H{
			"message":   "混合操作执行完成",
			"operation": "http_request_with_background_process",
		})
	})

	// 启动HTTP服务器
	fmt.Println("启动服务器在 http://localhost:8080")
	fmt.Println("调试器界面: http://localhost:8080/debug/list")
	fmt.Println("可用端点:")
	fmt.Println("  GET  /                    - 主页")
	fmt.Println("  POST /start-worker?name=xxx - 启动工作进程")
	fmt.Println("  POST /start-batch         - 启动批量处理")
	fmt.Println("  GET  /process-records     - 获取进程记录")
	fmt.Println("  GET  /mixed-operation     - 混合操作示例")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
