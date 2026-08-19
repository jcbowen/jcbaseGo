package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// FileStorageUsage 演示使用文件存储器的debugger配置
// 这个示例展示了如何将调试日志保存到文件系统中
func main() {
	fmt.Println("=== Debugger 文件存储示例 ===")

	// 创建Gin引擎
	r := gin.Default()

	// 创建文件存储器（日志将保存到 ./debug_logs 目录，最多存储5000条记录）
	fileStorage, err := debugger.NewFileStorage("./debug_logs", 5000)
	if err != nil {
		panic(fmt.Sprintf("创建文件存储器失败: %v", err))
	}

	// 创建调试器配置
	config := &debugger.Config{
		Enabled:         true,
		Storage:         fileStorage,
		MaxRecords:      5000,
		RetentionPeriod: 7 * 24 * time.Hour, // 保留7天
		SampleRate:      0.5,                // 50%采样率
		SkipPaths:       []string{"/static", "/favicon.ico"},
		SkipMethods:     []string{"OPTIONS"},
	}

	// 创建调试器实例
	debuggerInstance, err := debugger.New(config)
	if err != nil {
		panic(fmt.Sprintf("创建调试器失败: %v", err))
	}

	// 添加调试器中间件
	r.Use(debuggerInstance.Middleware())

	// 添加业务路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     "文件存储示例应用",
			"version": "1.0.0",
			"debug":   "调试日志将保存到 ./debug_logs 目录",
		})
	})

	r.GET("/api/products", func(c *gin.Context) {
		// 模拟查询产品列表
		time.Sleep(150 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"products": []gin.H{
				{"id": 1, "name": "笔记本电脑", "price": 5999.00},
				{"id": 2, "name": "智能手机", "price": 2999.00},
				{"id": 3, "name": "平板电脑", "price": 1999.00},
			},
		})
	})

	r.POST("/api/orders", func(c *gin.Context) {
		var order struct {
			ProductID int     `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Total     float64 `json:"total"`
		}

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的订单数据",
			})
			return
		}

		// 模拟创建订单
		time.Sleep(300 * time.Millisecond)

		c.JSON(http.StatusCreated, gin.H{
			"order_id":   1001,
			"product_id": order.ProductID,
			"quantity":   order.Quantity,
			"total":      order.Total,
			"status":     "pending",
			"created_at": time.Now().Format(time.RFC3339),
		})
	})

	// 启动服务器
	fmt.Println("启动文件存储调试器示例服务器...")
	fmt.Println("访问 http://localhost:8080/ 查看应用信息")
	fmt.Println("访问 http://localhost:8080/api/products 获取产品列表")
	fmt.Println("使用POST方法访问 http://localhost:8080/api/orders 创建订单")
	fmt.Println("调试日志将保存到 ./debug_logs 目录")

	r.Run(":8080")
}
