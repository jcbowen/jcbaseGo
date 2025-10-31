# 调试器组件 (Debugger Component)

## 概述

调试器组件是一个专为Gin框架设计的HTTP请求调试工具，用于解决Gin框架默认日志输出过多、并发请求下日志混杂的问题。该组件支持多种存储方式（文件、内存、数据库），提供完整的请求和响应信息记录、查询、搜索和详情查看功能。

## 功能特性

### 核心功能
- ✅ **请求拦截**: 自动拦截HTTP请求并记录完整的调试信息
- ✅ **多存储支持**: 支持文件、内存、数据库三种存储方式
- ✅ **日志查询**: 支持按时间、请求路径、状态码等条件过滤
- ✅ **全文搜索**: 支持关键词搜索请求头、响应头、请求体等内容
- ✅ **详情查看**: 显示完整的请求和响应信息，支持HTML和JSON格式
- ✅ **统计信息**: 提供请求统计、错误率、响应时间等分析数据

### 记录内容
- 请求方法、URL、查询参数
- 请求头和响应头
- 请求体和响应体（支持JSON格式化）
- 会话数据和错误信息
- 客户端IP、User Agent、请求ID
- 请求处理时间和状态码

## 快速开始

### 安装

确保项目已安装Gin框架：

```bash
go get -u github.com/gin-gonic/gin
```

### 基本使用

#### 便捷构造函数（推荐）

我们提供了多种便捷构造函数，使用更加简单直观：

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	// 方式1：使用简单调试器（默认配置）
	dbg, err := debugger.NewSimpleDebugger()
	if err != nil {
		panic(err)
	}

	// 方式2：使用内存存储器（开发环境）
	dbg, err := debugger.NewWithMemoryStorage(1000)
	if err != nil {
		panic(err)
	}

	// 方式3：使用文件存储器（生产环境）
	dbg, err := debugger.NewWithFileStorage("/var/log/debug_logs", 5000)
	if err != nil {
		panic(err)
	}

	// 方式4：使用自定义存储器
	customStorage, _ := debugger.NewMemoryStorage(150)
	dbg, err := debugger.NewWithCustomStorage(customStorage)
	if err != nil {
		panic(err)
	}

	// 方式5：生产环境调试器
	dbg, err := debugger.NewProductionDebugger("/var/log/debug_logs")
	if err != nil {
		panic(err)
	}

	// 创建Gin路由
	router := gin.New()
	
	// 使用调试器中间件
	router.Use(dbg.Middleware())
	
	// 添加业务路由
	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})
	
	router.Run(":8080")
}
```

### 详细使用指南

#### 配置选项

调试器支持丰富的配置选项，推荐使用便捷构造函数：

##### 便捷构造函数方式（推荐）

```go
// 方式1：使用简单调试器（默认配置）
dbg, err := debugger.NewSimpleDebugger()

// 方式2：使用内存存储器（开发环境）
dbg, err := debugger.NewWithMemoryStorage(1000)

// 方式3：使用文件存储器（生产环境）
dbg, err := debugger.NewWithFileStorage("/var/log/debug_logs", 5000)

// 方式4：使用自定义存储器
customStorage, _ := debugger.NewMemoryStorage(150)
dbg, err := debugger.NewWithCustomStorage(customStorage)

// 方式5：生产环境配置
dbg, err := debugger.NewProductionDebugger("/var/log/debug_logs", 1000)

if err != nil {
	panic(err)
}

router.Use(dbg.Middleware())
```

##### 手动配置方式（高级使用）

```go
// 创建存储器实例
memoryStorage, _ := debugger.NewMemoryStorage(150)

// 创建调试器配置
config := &debugger.Config{
	Enabled:         true,                    // 是否启用调试器
	Storage:         memoryStorage,           // 必须传入实例化的存储器
	MaxBodySize:     1024,                    // 最大请求/响应体大小（KB），默认1MB
	RetentionPeriod: 7 * 24 * time.Hour,      // 日志保留期限，默认7天
	Level:           debugger.LevelDebug,     // 日志级别：LevelDebug/LevelInfo/LevelWarn/LevelError
	MaxRecords:      150,                     // 最大记录数量，默认150

	// 过滤配置
	SkipPaths:   []string{"/static", "/health"}, // 跳过的路径
	SkipMethods: []string{"OPTIONS"},              // 跳过的HTTP方法

	// 采样配置
	SampleRate: 1.0,                            // 采样率（0-1之间），默认1.0（记录所有请求）

	// 日志记录器配置（可选）
	Logger:  nil,                              // 日志记录器实例
}

// 使用自定义配置
dbg, err := debugger.New(config)
if err != nil {
	panic(err)
}

router.Use(dbg.Middleware())
```

### 日志级别说明

- **LevelDebug**: 记录所有详细信息，包括请求体、响应体等（开发环境）
- **LevelInfo**: 记录基本信息，包括请求头、响应头等（生产环境）
- **LevelWarn**: 只记录警告级别的信息
- **LevelError**: 只记录错误级别的信息

## 高级功能

### 采样率配置

采样率可以控制记录日志的频率，避免在高并发场景下产生过多日志：

```go
memoryStorage, _ := debugger.NewMemoryStorage(10000)
config := &debugger.Config{
    Enabled:    true,
    Storage:    memoryStorage,
    MaxRecords: 10000,
    Level:      debugger.LevelInfo,
    
    // 采样率配置：每10个请求记录1个（10%采样率）
    SampleRate: 0.1,
}

dbg, err := debugger.New(config)
if err != nil {
    panic(err)
}
```

### 过滤配置

可以配置请求过滤规则，避免记录某些不需要的请求：

```go
memoryStorage, _ := debugger.NewMemoryStorage(10000)
config := &debugger.Config{
    Enabled:    true,
    Storage:    memoryStorage,
    MaxRecords: 10000,
    Level:      debugger.LevelInfo,
    
    // 过滤配置
    SkipPaths: []string{
        "/static/*",
        "/favicon.ico",
        "/robots.txt",
    },
    
    // 忽略特定HTTP方法
    SkipMethods: []string{"OPTIONS"},
    
    // 最大请求体大小（字节）
    MaxBodySize:     1024,        // 1MB（单位：KB）
}

dbg, err := debugger.New(config)
if err != nil {
    panic(err)
}
```

### 自定义存储

您可以实现自己的存储后端：

```go
type CustomStorage struct {
    // 实现 Storage 接口
}

func (s *CustomStorage) Save(entry *LogEntry) error {
    // 自定义保存逻辑
}

func (s *CustomStorage) FindByID(id string) (*LogEntry, error) {
    // 自定义查询逻辑
}

// 实现其他接口方法...
```

### 查询和搜索功能

#### 基本查询

```go
// 通过调试器实例获取存储
storage := dbg.GetStorage()

// 查询最近的日志
logs, total, err := storage.FindAll(1, 10, nil)

// 按时间范围查询
startTime := time.Now().AddDate(0, 0, -7) // 一周前
endTime := time.Now()
filters := map[string]interface{}{
    "start_time": startTime,
    "end_time":   endTime,
}
result, total, err := storage.FindAll(1, 20, filters)

// 按方法查询
filters = map[string]interface{}{
    "method": "POST",
}
result, total, err := storage.FindAll(1, 10, filters)

// 按状态码查询
filters = map[string]interface{}{
    "status_code": 404,
}
result, total, err := storage.FindAll(1, 10, filters)

// 查询错误日志
filters = map[string]interface{}{
    "has_error": true,
}
result, total, err := storage.FindAll(1, 10, filters)
```

#### 高级查询

```go
// 使用查询选项
options := debugger.QueryOptions{
	Page:     1,
	PageSize: 20,
	Filters: map[string]interface{}{
		"method":      "POST",
		"status_code": 200,
		"url_contains": "/api",
		"has_error":   false,
	},
	SortBy:    "timestamp",
	SortOrder: "desc",
}

result, err := queryManager.Query(options)
```

#### 全文搜索

```go
// 搜索包含关键词的日志
result, total, err := storage.Search("用户", 1, 10)
```

### 详情查看功能

```go
// 创建详情查看器
detailViewer := debugger.NewDetailViewer(queryManager)

// 获取日志详情
detail, err := detailViewer.GetDetail("log_id")

// 获取HTML格式的详情
html, err := detailViewer.GetDetailHTML("log_id")

// 获取JSON格式的详情
jsonData, err := detailViewer.GetDetailJSON("log_id")

// 比较两个日志条目
comparison, err := detailViewer.CompareEntries("id1", "id2")
```

### 统计信息

```go
// 获取统计信息
stats, err := queryManager.GetStats()

// 统计信息包含：
// - 存储统计（总数、错误率等）
// - 方法统计（各HTTP方法的请求数量）
// - 状态码统计（各状态码的分布）
// - 时间统计（今天、本周、本月的请求数量）
```

### 数据导出

```go
// 导出为JSON格式
jsonData, err := queryManager.Export("json", options)

// 导出为CSV格式
csvData, err := queryManager.Export("csv", options)
```

## Gin控制器使用指南

调试器组件提供了完整的Gin控制器，可以快速集成到现有的Gin应用中，提供Web管理界面。

### 基本使用

#### 1. 使用默认路由配置

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	// 创建Gin引擎
	router := gin.Default()
	
	// 创建调试器实例（推荐使用便捷构造函数）
	dbg, err := debugger.NewWithMemoryStorage(1000)
	if err != nil {
		panic(err)
	}
	
	// 注册调试器路由（使用默认路径：/jcbase/debugger）
	dbg.RegisterRoutes(router)
	
	// 添加业务路由
	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})
	
	router.Run(":8080")
}
```

#### 2. 使用自定义路由组

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	router := gin.Default()
	
	// 创建调试器实例
	dbg, err := debugger.NewWithMemoryStorage(1000)
	if err != nil {
		panic(err)
	}
	
	// 创建自定义路由组
	adminGroup := router.Group("/admin")
	adminGroup.Use(authMiddleware()) // 添加认证中间件
	
	// 注册调试器路由到自定义路由组
	dbg.RegisterRoutes(adminGroup)
	
	router.Run(":8080")
}
```

#### 3. 高级配置

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	router := gin.Default()
	
	// 创建调试器实例
	dbg, err := debugger.NewWithMemoryStorage(1000)
	if err != nil {
		panic(err)
	}
	
	// 创建自定义控制器配置
	controllerConfig := &debugger.ControllerConfig{
		BasePath: "/custom/debug", // 自定义基础路径
		Title:    "调试管理界面",     // 自定义页面标题
		PageSize: 50,               // 自定义每页显示数量
	}
	
	// 创建自定义路由组
	adminGroup := router.Group("/admin")
	
	// 创建控制器实例
	controller := debugger.NewController(dbg, adminGroup, controllerConfig)
	
	// 控制器会自动注册路由到指定的路由组
	
	router.Run(":8080")
}
```

### 控制器功能

Gin控制器提供以下功能：

#### 1. 日志列表页面
- 路径：`GET /jcbase/debugger/`
- 功能：显示所有调试日志的列表，支持分页和搜索

#### 2. 日志详情页面
- 路径：`GET /jcbase/debugger/detail/:id`
- 功能：显示单个日志的详细信息，包括请求头、响应头、请求体等

#### 3. 搜索功能
- 路径：`GET /jcbase/debugger/search`
- 功能：支持关键词搜索，可搜索请求头、响应头、请求体等内容

#### 4. API接口
控制器还提供对应的API接口，方便前端集成：
- `GET /jcbase/debugger/api/logs` - 获取日志列表（JSON格式）
- `GET /jcbase/debugger/api/logs/:id` - 获取日志详情（JSON格式）
- `GET /jcbase/debugger/api/search` - 搜索日志（JSON格式）

### 示例代码

项目提供了完整的示例代码，位于 `example/debugger/` 目录下：

- **`basic_usage.go`** - 基础使用示例，包含GIN框架集成和基本路由
- **`file_storage.go`** - 文件存储使用示例，支持日志文件管理
- **`controller_usage.go`** - 控制器使用示例，提供调试器Web界面
- **`config_examples.go`** - 配置示例，包含6种不同的配置方式

运行示例代码：
```bash
cd example/debugger/
go run basic_usage.go
```

## Web管理界面

调试器组件可以轻松集成到Web管理界面中：

### 日志列表接口

```go
router.GET("/admin/debug/logs", func(c *gin.Context) {
	page := getIntParam(c, "page", 1)
	pageSize := getIntParam(c, "page_size", 20)
	
	options := debugger.QueryOptions{
		Page:     page,
		PageSize: pageSize,
		Filters:  buildFilters(c),
		SortBy:   c.DefaultQuery("sort_by", "timestamp"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	
	result, err := queryManager.Query(options)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, result)
})
```

### 日志详情接口

```go
router.GET("/admin/debug/logs/:id", func(c *gin.Context) {
	id := c.Param("id")
	
	detail, err := detailViewer.GetDetail(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "日志不存在"})
		return
	}
	
	c.JSON(200, detail)
})
```

### 统计信息接口

```go
router.GET("/admin/debug/stats", func(c *gin.Context) {
	stats, err := queryManager.GetStats()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, stats)
})
```

## 存储方式比较

| 存储方式 | 创建方式 | 优点 | 缺点 | 适用场景 |
|---------|----------|------|------|----------|
| 内存存储器 | `NewMemoryStorage(maxSize)` | 速度快，零配置 | 数据易丢失，重启后清空 | 开发环境，临时调试 |
| 文件存储器 | `NewFileStorage(path, maxSize)` | 数据持久化，易于备份 | 文件IO可能成为瓶颈 | 生产环境，中小规模应用 |
| 数据库存储器 | `NewDatabaseStorage(db, tableName)` | 查询性能好，支持复杂查询 | 需要数据库依赖 | 生产环境，大规模应用 |

### 存储创建示例

```go
// 内存存储器（默认10000条记录）
memoryStorage, err := debugger.NewMemoryStorage()

// 内存存储器（自定义最大记录数）
memoryStorage, err := debugger.NewMemoryStorage(5000)

// 文件存储器（默认路径，默认10000条记录）
fileStorage, err := debugger.NewFileStorage("/var/log/debug_logs")

// 文件存储器（自定义最大记录数）
fileStorage, err := debugger.NewFileStorage("/var/log/debug_logs", 20000)

// 数据库存储器（需要GORM连接）
databaseStorage, err := debugger.NewDatabaseStorage(db, "debug_logs")

// 推荐使用便捷构造函数直接创建调试器
dbg, err := debugger.NewWithMemoryStorage(1000)  // 内存存储
dbg, err := debugger.NewWithFileStorage("/var/log/debug_logs", 5000)  // 文件存储
dbg, err := debugger.NewProductionDebugger("/var/log/debug_logs", 1000)  // 生产环境配置
```

## 性能优化建议

### 生产环境配置

1. **使用LevelInfo级别**: 避免记录请求体和响应体，减少存储空间
2. **定期清理旧日志**: 设置自动清理策略
3. **使用数据库索引**: 对常用查询字段创建索引
4. **限制日志大小**: 设置单个日志文件的最大大小

### 清理策略示例

```go
// 定期清理7天前的日志
func cleanupOldLogs() {
	ticker := time.NewTicker(24 * time.Hour) // 每天执行一次
	defer ticker.Stop()
	
	for range ticker.C {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		err := queryManager.Cleanup(sevenDaysAgo)
		if err != nil {
			log.Printf("清理旧日志失败: %v", err)
		} else {
			log.Println("旧日志清理完成")
		}
	}
}
```

## 安全考虑

### 敏感数据处理

1. **避免记录敏感信息**: 在生产环境中使用LevelInfo级别
2. **加密存储**: 对敏感日志内容进行加密
3. **访问控制**: 对调试接口进行身份验证

### 生产环境安全配置

```go
// 生产环境配置
config := debugger.Config{
	Enabled: true,
	Level:   debugger.LevelInfo, // 不记录请求体和响应体
}

// 保护调试接口
admin := router.Group("/admin")
admin.Use(authMiddleware()) // 添加认证中间件
admin.GET("/debug/logs", debugLogsHandler)
```

## 故障排除

### 常见问题

1. **内存使用过高**
   - 使用文件或数据库存储器替代内存存储器
   - 设置日志数量上限
   - 定期清理旧日志

2. **存储性能问题**
   - 对数据库字段添加索引
   - 使用更快的存储介质（SSD）
   - 考虑日志分片存储

3. **日志丢失**
   - 检查存储路径权限
   - 验证数据库连接
   - 添加错误重试机制

## API参考

### 主要结构体

#### LogEntry
```go
type LogEntry struct {
	ID              string            `json:"id"`          // 日志唯一标识
	Timestamp       time.Time         `json:"timestamp"`   // 请求时间戳
	Method          string            `json:"method"`      // HTTP方法
	URL             string            `json:"url"`         // 请求URL
	StatusCode      int               `json:"status_code"` // HTTP状态码
	Duration        time.Duration     `json:"duration"`    // 处理耗时
	ClientIP        string            `json:"client_ip"`   // 客户端IP
	UserAgent       string            `json:"user_agent"`  // 用户代理
	RequestID       string            `json:"request_id"`  // 请求ID（用于追踪）

	// 请求信息
	RequestHeaders map[string]string `json:"request_headers"` // 请求头
	QueryParams    map[string]string `json:"query_params"`    // 查询参数
	RequestBody    string            `json:"request_body"`    // 请求体内容

	// 响应信息
	ResponseHeaders map[string]string `json:"response_headers"` // 响应头
	ResponseBody    string            `json:"response_body"`    // 响应体内容

	// 会话数据（可选）
	SessionData map[string]interface{} `json:"session_data,omitempty"` // 会话数据

	// 错误信息
	Error string `json:"error,omitempty"` // 错误信息
}
```

#### Config
```go
type Config struct {
	Enabled         bool          // 是否启用调试器
	Storage         Storage       // 存储器实例
	MaxBodySize     int64         // 最大请求/响应体大小（KB）
	RetentionPeriod time.Duration // 日志保留期限
	Level           LogLevel      // 日志级别
	MaxRecords      int           // 最大记录数量
	SkipPaths       []string      // 跳过的路径
	SkipMethods     []string      // 跳过的HTTP方法
	SampleRate      float64       // 采样率（0-1之间）
	Logger          Logger        // 日志记录器实例
}
```

### 主要接口

#### Storage接口
```go
type Storage interface {
	Save(entry *LogEntry) error
	FindByID(id string) (*LogEntry, error)
	FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error)
	Search(keyword string, page, pageSize int) ([]*LogEntry, int, error)
	Cleanup(before time.Time) error
	GetStats() (map[string]interface{}, error)
	GetMethods() (map[string]int, error)
	GetStatusCodes() (map[int]int, error)
	Close() error
}
```

#### Logger接口
```go
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}
```

### 便捷构造函数

```go
// 便捷构造函数
func NewSimpleDebugger() (*Debugger, error)
func NewWithMemoryStorage(maxRecords int) (*Debugger, error)
func NewWithFileStorage(path string, maxRecords int) (*Debugger, error)
func NewWithCustomStorage(storage Storage) (*Debugger, error)
func NewProductionDebugger(path string, maxRecords int) (*Debugger, error)

// 存储器构造函数
func NewMemoryStorage(maxRecords int) (Storage, error)
func NewFileStorage(path string, maxRecords int) (Storage, error)
func NewDatabaseStorage(db *gorm.DB, tableName string) (Storage, error)
```

## 版本历史

- v1.0.0: 初始版本，支持基本调试功能
- 后续版本计划：支持更多存储后端、实时日志流、性能监控等

## 贡献指南

欢迎提交Issue和Pull Request来改进这个组件。

## 许可证

MIT License