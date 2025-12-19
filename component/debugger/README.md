# 调试器组件 (Debugger Component)

## 概述

调试器组件是一个专为Gin框架设计的HTTP请求调试工具，用于解决Gin框架默认日志输出过多、并发请求下日志混杂的问题。该组件支持多种存储方式（文件、内存、数据库），提供完整的请求和响应信息记录、查询、搜索和详情查看功能。

## 功能特性

### 核心功能
- ✅ **请求拦截**: 自动拦截HTTP请求并记录完整的调试信息，支持请求体大小限制和二进制数据检测
- ✅ **多存储支持**: 支持内存、文件、数据库等多种存储方式，提供统一的Storage接口
- ✅ **日志查询**: 支持按时间、请求路径、状态码、记录类型等条件过滤，实现真正的分页查询
- ✅ **全文搜索**: 支持关键词搜索进程名称、URL、请求体、响应体、错误信息等内容
- ✅ **详情查看**: 显示完整的请求和响应信息，支持HTML和JSON格式，自动计算存储大小
- ✅ **统计信息**: 提供请求统计、错误率、响应时间、HTTP方法、状态码等分析数据
- ✅ **流式请求支持**: 自动检测和记录流式响应（SSE、分块传输等），支持分块记录、元数据统计和内存管理优化
- ✅ **进程级调试**: 支持非HTTP进程的调试记录，适用于后台任务、批处理作业等场景
- ✅ **内置Logger**: 提供多级别日志记录器，支持字段附加和日志收集

### 记录内容
- 请求方法、URL、查询参数
- 请求头和响应头
- 请求体和响应体（支持JSON格式化、二进制数据检测）
- 会话数据和错误信息
- 客户端IP、User Agent、请求ID
- 请求处理时间和状态码
- 流式响应元数据（分块数量、分块大小、流式状态、最大分块限制等）
- 进程记录信息（进程ID、进程名称、进程类型、开始时间、结束时间、状态）
- Logger日志信息（时间戳、级别、消息、附加字段、位置信息）
- 存储大小信息（自动计算和格式化显示）

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

// 方式5：生产环境调试器
	dbg, err := debugger.NewProductionDebugger("/var/log/debug_logs")
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
	Level:           debugger.LevelWarn,     // 日志级别：LevelInfo/LevelWarn/LevelError/LevelSilent（默认：LevelWarn）
	MaxRecords:      150,                     // 最大记录数量，默认150

	// 过滤配置
	SkipPaths:   []string{"/static", "/health"}, // 跳过的路径
	SkipMethods: []string{"OPTIONS"},              // 跳过的HTTP方法

	// 采样配置
	SampleRate: 1.0,                            // 采样率（0-1之间），默认1.0（记录所有请求）

	// IP访问控制配置
	AllowedIPs: []string{},                     // 允许访问的IP白名单，空数组表示不限制

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

- **LevelSilent**: 静默模式，不记录任何调试日志（生产环境/关闭调试）
- **LevelInfo**: 记录所有详细信息，包括请求体、响应体等（开发环境/最高调试级别）
- **LevelWarn**: 只记录警告级别的信息
- **LevelError**: 只记录错误级别的信息

### IP访问控制

调试器组件支持IP访问控制功能，可以限制只有特定IP地址或IP段的客户端才能访问调试器界面和API接口。

#### 配置说明

```go
config := &debugger.Config{
    // ... 其他配置
    
    // IP访问控制配置
    AllowedIPs: []string{
        "192.168.1.100",      // 允许单个IP
        "10.0.0.0/8",         // 允许10.0.0.0-10.255.255.255整个B类网络
        "172.16.0.0/12",      // 允许172.16.0.0-172.31.255.255
        "192.168.1.0/24",     // 允许192.168.1.0-192.168.1.255
    },
}
```

#### 使用规则

1. **无限制访问**：当`AllowedIPs`为空数组或未配置时，允许所有IP地址访问
2. **白名单限制**：当配置了IP白名单时，只有白名单中的IP地址可以访问
3. **CIDR支持**：支持CIDR格式的IP段配置（如`192.168.1.0/24`）
4. **优先级检查**：支持`X-Forwarded-For`头，优先检查代理链中的第一个客户端IP

#### 示例配置

```go
// 开发环境：允许本地和内部网络访问
config := &debugger.Config{
    Enabled: true,
    Storage: memoryStorage,
    AllowedIPs: []string{
        "127.0.0.1",           // 本地回环
        "::1",                 // IPv6本地回环
        "192.168.1.0/24",      // 内部网络
        "10.0.0.0/8",          // 私有网络A类
    },
}

// 生产环境：只允许管理员IP访问
config := &debugger.Config{
    Enabled: true,
    Storage: fileStorage,
    AllowedIPs: []string{
        "203.0.113.100",       // 管理员IP
        "203.0.113.101",       // 备用管理员IP
    },
}

// 完全开放：不限制IP访问
config := &debugger.Config{
    Enabled: true,
    Storage: memoryStorage,
    AllowedIPs: []string{},     // 空数组表示不限制
}
```

#### 错误响应

当IP不在白名单中时，调试器会返回HTTP 403状态码和JSON格式的错误信息：

```json
{
    "error": "禁止访问：IP地址 172.16.1.100 不在允许列表中",
    "client_ip": "172.16.1.100",
    "allowed_ips": ["192.168.1.100", "10.0.0.0/8"]
}
```

## 流式请求支持

### 概述

debugger组件现在支持流式请求的自动检测和记录功能，可以识别和记录Server-Sent Events (SSE)、分块传输编码等流式响应。流式请求支持允许您监控和分析实时数据流，同时保持对传统HTTP请求的完整兼容性。

### 功能特性

- ✅ **自动检测**: 自动识别流式响应类型（SSE、分块传输、二进制流等）
- ✅ **分块记录**: 支持流式响应的分块记录和元数据统计
- ✅ **配置控制**: 可配置分块大小限制和最大分块数量
- ✅ **统计信息**: 提供流式请求数量、平均分块数、最大分块数等统计
- ✅ **筛选查询**: 支持按流式请求状态和活跃状态进行筛选
- ✅ **Web界面支持**: 在调试器Web界面中可查看流式请求详情

### 配置说明

流式请求支持需要显式启用，相关配置选项如下：

```go
config := &debugger.Config{
    // ... 其他配置
    
    // 流式请求支持配置
    EnableStreamingSupport: true,                    // 启用流式请求支持
    StreamingChunkSize:     1024,                    // 流式响应分块大小（KB），默认1MB，0表示无限制
    MaxStreamingChunks:     10,                      // 最大流式响应分块数量，默认10个，0表示无限制
    MaxStreamingMemory:     10485760,                // 流式响应总内存限制（字节），默认10MB，0表示无限制
}
```

#### 配置选项说明

- **EnableStreamingSupport**: 是否启用流式请求支持，默认为false
- **StreamingChunkSize**: 单个流式响应分块的大小限制（KB），默认1024KB（1MB），设置为0表示无限制
- **MaxStreamingChunks**: 最大流式响应分块数量，默认10个，设置为0表示无限制
- **MaxStreamingMemory**: 流式响应总内存使用限制（字节），默认10485760字节（10MB），设置为0表示无限制

#### 无限制配置示例

如果您希望完全禁用流式请求的限制，可以这样配置：

```go
config := &debugger.Config{
    EnableStreamingSupport: true,
    StreamingChunkSize:     0,      // 无分块大小限制
    MaxStreamingChunks:     0,      // 无分块数量限制
    MaxStreamingMemory:     0,      // 无内存限制
}
```

### 支持的流式响应类型

- **Server-Sent Events (SSE)**: `Content-Type: text/event-stream`
- **分块传输编码**: `Transfer-Encoding: chunked`
- **流式JSON响应**: `Content-Type: application/x-ndjson` 或 `application/json-seq`
- **二进制流**: `Content-Type: application/octet-stream`
- **WebSocket升级响应**: `Upgrade: websocket`

### 使用示例

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	// 创建调试器配置，启用流式请求支持
	memoryStorage, _ := debugger.NewMemoryStorage(100)
	config := &debugger.Config{
		Enabled:                true,
		Storage:                memoryStorage,
		EnableStreamingSupport: true,    // 启用流式请求支持
		StreamingChunkSize:     512,      // 分块大小限制为512KB
		MaxStreamingChunks:     5,        // 最多记录5个分块
	}
	
	dbg, err := debugger.New(config)
	if err != nil {
		panic(err)
	}

	router := gin.New()
	router.Use(dbg.Middleware())

	// SSE流式响应示例
	router.GET("/sse", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		for i := 1; i <= 3; i++ {
			c.Writer.Write([]byte(fmt.Sprintf("data: 消息 %d\\n\\n", i)))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(1 * time.Second)
		}
	})

	// 分块传输编码示例
	router.GET("/chunked", func(c *gin.Context) {
		c.Header("Transfer-Encoding", "chunked")
		
		for i := 1; i <= 4; i++ {
			chunk := fmt.Sprintf("分块 %d 内容\\n", i)
			c.Writer.Write([]byte(chunk))
			c.Writer.(http.Flusher).Flush()
			time.Sleep(500 * time.Millisecond)
		}
	})

	router.Run(":8080")
}
```

### 流式请求记录内容

流式请求记录包含以下额外信息：
- **流式响应标记**: `IsStreamingResponse: true`
- **分块数量**: `StreamingChunks: 3`
- **分块大小限制**: `StreamingChunkSize: 512`
- **最大分块数量**: `MaxStreamingChunks: 5`
- **内存使用限制**: `MaxStreamingMemory: 10485760`
- **流式数据摘要**: `StreamingData: "Streaming Response: 3 chunks, total size: 1024 bytes"`

### 内存管理机制

调试器组件实现了智能的内存管理机制，确保流式请求记录不会导致内存溢出：

1. **分块数量限制**: 当分块数量超过`MaxStreamingChunks`限制时，采用LRU策略移除最旧的分块
2. **分块大小限制**: 单个分块大小超过`StreamingChunkSize`限制时，数据会被截断并添加"[truncated]"标记
3. **总内存限制**: 当所有分块的总内存使用量超过`MaxStreamingMemory`限制时，持续移除最旧的分块直到满足限制
4. **无限制支持**: 所有限制都支持设置为0，表示无限制记录

### Web界面筛选

在调试器Web界面中，您可以使用以下筛选选项：
- **流式状态筛选**: 所有流式状态 / 流式请求 / 非流式请求
- **流式活跃状态**: 活跃流式请求 / 非流式请求

## Logger功能

### 概述

debugger组件提供了内置的Logger功能，支持多级别日志记录和结构化字段附加。Logger可以在HTTP请求处理过程中记录调试信息，也可以在进程级Debugger中使用，为应用程序提供统一的日志记录解决方案。

### 功能特性

- ✅ **多级别日志**: 支持Debug、Info、Warn、Error四种日志级别
- ✅ **结构化字段**: 支持添加自定义字段到日志记录
- ✅ **日志收集**: 自动收集Logger日志并与HTTP请求或进程记录关联
- ✅ **级别控制**: 支持根据配置的日志级别过滤日志记录

### 基本使用示例

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	dbg, err := debugger.NewSimpleDebugger()
	if err != nil {
		panic(err)
	}

	router := gin.New()
	router.Use(dbg.Middleware())

	router.GET("/api/users", func(c *gin.Context) {
		// 获取当前请求的Logger实例
		logger := debugger.GetLoggerFromContext(c)
		
		// 记录不同级别的日志
		logger.Info("开始处理用户请求")
		logger.Info("用户ID", map[string]interface{}{"user_id": 123})
		logger.Warn("用户权限检查", map[string]interface{}{"permission": "read"})
		
		// 使用WithFields创建带字段的Logger
		userLogger := logger.WithFields(map[string]interface{}{
			"user_id": 123,
			"action": "get_user",
		})
		
		userLogger.Info("获取用户信息")
		
		// 记录错误信息
		if err := someBusinessLogic(); err != nil {
			logger.Error("业务逻辑错误", map[string]interface{}{
				"error": err.Error(),
				"stack": "调用堆栈信息",
			})
		}
		
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	router.Run(":8080")
}
```

### Logger接口方法

```go
// LoggerInterface 日志记录器接口
type LoggerInterface interface {
	// Info 记录信息级别日志
	Info(msg any, fields ...map[string]interface{})

	// Warn 记录警告级别日志
	Warn(msg any, fields ...map[string]interface{})

	// Error 记录错误级别日志
	Error(msg any, fields ...map[string]interface{})

	// WithFields 创建带有字段的日志记录器
	WithFields(fields map[string]interface{}) LoggerInterface

	// GetLevel 获取当前日志记录器的日志级别
	GetLevel() LogLevel
}
```

### 位置信息记录功能

调试器组件现在支持自动记录日志打印的位置信息，包括文件名、行号和函数名。这个功能可以帮助开发者快速定位日志输出的具体位置，提高调试效率。

#### 功能特性

- ✅ **自动位置记录**: 每次日志调用自动记录调用位置
- ✅ **完整位置信息**: 包含文件名、行号、函数名
- ✅ **结构化存储**: 位置信息存储在LoggerLog结构体中
- ✅ **格式化输出**: 日志输出包含位置信息格式

#### 位置信息字段

```go
type LoggerLog struct {
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Level     string                 `json:"level"`     // 日志级别
	Message   string                 `json:"message"`   // 日志消息
	Fields    map[string]interface{} `json:"fields"`    // 附加字段
	FileName  string                 `json:"fileName"`  // 文件名（新增）
	Line      int                    `json:"line"`      // 行号（新增）
	Function  string                 `json:"function"`  // 函数名（新增）
}
```

#### 日志输出格式

日志输出现在包含位置信息，格式为：
```
[级别] 文件名:行号 函数名: 消息内容
```

**示例输出：**
```
[INFO] debugger_test.go:209 TestLoggerLocationInfo.func1: 测试位置信息记录
```

#### 使用示例

位置信息记录功能是自动启用的，无需额外配置。所有通过Logger接口记录的日志都会自动包含位置信息。

```go
// 获取调试器的Logger实例
logger := dbg.GetLogger()

// 记录日志（自动包含位置信息）
logger.Info("用户登录成功", map[string]interface{}{
	"user_id": 123,
	"username": "张三",
})

// 输出示例：
// [INFO] user_controller.go:45 handleUserLogin: 用户登录成功
```

#### 测试验证

可以通过测试用例验证位置信息记录功能：

```go
func TestLoggerLocationInfo(t *testing.T) {
	dbg, _ := debugger.NewSimpleDebugger()
	logger := dbg.GetLogger()
	
	// 记录测试日志
	logger.Info("测试位置信息记录")
	
	// 验证位置信息字段
	// FileName: 包含当前测试文件名
	// Line: 包含调用日志的行号
	// Function: 包含调用日志的函数名
}
```

### 日志级别说明

- **LevelSilent**: 静默模式，不记录任何调试日志（生产环境/关闭调试）
- **LevelError**: 只记录错误级别的信息
- **LevelWarn**: 记录错误和警告级别的信息（默认级别）
- **LevelInfo**: 记录所有详细信息，包括请求体、响应体等（开发环境/最高调试级别）

### 消息格式支持

Logger支持多种消息格式：
- **字符串**: `logger.Info("用户登录成功")`
- **结构体**: `logger.Info(User{ID: 123, Name: "张三"})`
- **Map**: `logger.Info(map[string]interface{}{"user_id": 123, "action": "login"})`
- **实现Stringer接口的类型**: `logger.Info(customStringer)`

## 进程级Debugger功能

### 概述

debugger组件现在支持进程级日志记录功能，可以在非HTTP请求场景下记录进程执行日志。进程级debugger允许您为后台任务、批处理作业、定时任务等非HTTP进程创建独立的日志记录器，并支持与HTTP请求日志共存于同一存储中。

### 功能特性

- ✅ **进程记录管理**: 支持创建、获取、结束进程记录
- ✅ **多级日志记录**: 支持Debug、Info、Warn、Error四种日志级别
- ✅ **结构化字段**: 支持添加自定义字段到日志记录
- ✅ **存储共存**: 进程记录与HTTP请求记录共存于同一存储
- ✅ **查询过滤**: 支持按进程名称、进程ID、记录类型等条件过滤
- ✅ **Web界面支持**: 在调试器Web界面中可查看进程记录

> 注：进程级Debugger继承了Logger的所有功能特性。

### 基本使用示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	// 创建调试器实例
	dbg, err := debugger.NewProductionDebugger("/var/log/debug_logs")
	if err != nil {
		panic(err)
	}

	// 启动一个进程记录
	logger := dbg.StartProcess("数据同步任务", "batch")
	defer dbg.EndProcess(logger.GetProcessID(), debugger.ProcessStatusCompleted) // 确保进程结束时记录结束时间

	// 记录进程执行日志
	logger.Info("开始执行数据同步任务", map[string]interface{}{
		"source": "MySQL",
		"target": "Elasticsearch",
		"batch_size": 1000,
	})

	// 模拟数据处理
	for i := 0; i < 5; i++ {
		logger.Info("获取数据源信息", map[string]interface{}{
			"source": "MySQL",
			"table": "users",
		})
		time.Sleep(100 * time.Millisecond)

		if i == 2 {
			logger.Warn("遇到网络延迟", map[string]interface{}{
				"retry_count": 1,
				"delay_ms": 500,
			})
		}
	}

	// 记录任务完成
	logger.Info("数据同步任务完成", map[string]interface{}{
		"processed_records": 5000,
		"duration_seconds": 2.5,
		"status": "success",
	})
}
```

### 进程记录方法

#### 启动进程记录

```go
// 启动进程记录，返回进程级Logger实例
logger := dbg.StartProcess(processName string, processType string)
```

#### 获取进程Logger

```go
// 通过进程ID获取对应的Logger实例
logger, err := dbg.GetProcessLogger(processID string)
```

#### 结束进程记录

```go
// 结束进程记录，记录结束时间和状态
err := dbg.EndProcess(processID string, status string)
```

**进程状态常量**

为了更好的代码可维护性，建议使用预定义的进程状态常量：

```go
// 进程状态常量定义
const (
	ProcessStatusCompleted = "completed" // 进程正常完成
	ProcessStatusFailed    = "failed"    // 进程执行失败
	ProcessStatusCancelled = "cancelled" // 进程被取消
)
```

**使用示例：**

```go
// 推荐使用常量
dbg.EndProcess(logger.GetProcessID(), debugger.ProcessStatusCompleted)

// 不推荐使用字符串字面量
dbg.EndProcess(logger.GetProcessID(), "completed")
```

#### 获取进程记录列表

```go
// 获取进程记录列表，支持分页和过滤
records, total, err := dbg.GetProcessRecords(page int, pageSize int, filters map[string]interface{})
```

### 高级使用示例

#### 并发进程管理

```go
package main

import (
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func processWorker(dbg *debugger.Debugger, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	// 为每个工作进程创建独立的记录
	logger := dbg.StartProcess(fmt.Sprintf("工作进程-%d", workerID), "worker")
	defer dbg.EndProcess(logger.GetProcessID(), debugger.ProcessStatusCompleted)

	logger.Info("工作进程启动", map[string]interface{}{
		"worker_id": workerID,
		"start_time": time.Now().Format(time.RFC3339),
	})

	// 模拟工作负载
	for i := 0; i < 3; i++ {
		logger.Info(fmt.Sprintf("处理任务批次 %d", i+1), map[string]interface{}{
			"batch": i + 1,
			"worker": workerID,
		})
		time.Sleep(200 * time.Millisecond)
	}

	logger.Info("工作进程完成", map[string]interface{}{
		"worker_id": workerID,
		"end_time": time.Now().Format(time.RFC3339),
		"tasks_processed": 3,
	})
}

func main() {
	dbg, _ := debugger.NewWithMemoryStorage(1000)
	
	var wg sync.WaitGroup
	
	// 启动3个并发工作进程
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go processWorker(dbg, &wg, i)
	}
	
	wg.Wait()
	
	// 获取所有进程记录
	records, total, err := dbg.GetProcessRecords(1, 20, nil)
	if err != nil {
		fmt.Printf("获取进程记录失败: %v\n", err)
	} else {
		fmt.Printf("共完成 %d 个进程记录，总计 %d 条记录\n", len(records), total)
	}
}
```

#### 批处理任务监控

```go
package main

import (
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func batchProcessor(dbg *debugger.Debugger) {
	// 创建批处理任务记录
	logger := dbg.StartProcess("夜间批处理任务", "batch")
	defer dbg.EndProcess(logger.GetProcessID(), debugger.ProcessStatusCompleted)

	logger.Info("开始夜间批处理", map[string]interface{}{
		"scheduled_time": "02:00",
		"estimated_duration": "2小时",
	})

	// 模拟多个处理步骤
	steps := []string{"数据备份", "统计计算", "报告生成", "清理临时文件"}
	
	for i, step := range steps {
		stepLogger := logger.WithFields(map[string]interface{}{
			"step_number": i + 1,
			"step_name": step,
		})
		
		stepLogger.Info("开始处理步骤")
		time.Sleep(500 * time.Millisecond) // 模拟处理时间
		
		if i == 1 {
			stepLogger.Warn("统计计算耗时较长", map[string]interface{}{
				"actual_duration": "45分钟",
				"expected_duration": "30分钟",
			})
		}
		
		stepLogger.Info("步骤处理完成")
	}

	logger.Info("夜间批处理任务完成", map[string]interface{}{
		"total_steps": len(steps),
		"completion_time": time.Now().Format(time.RFC3339),
		"status": "success",
	})
}

func main() {
	dbg, _ := debugger.NewWithFileStorage("/var/log/batch_logs", 5000)
	batchProcessor(dbg)
}
```

### 查询和过滤进程记录

进程记录支持与HTTP请求记录相同的查询和过滤功能：

```go
// 查询所有进程记录
filters := map[string]interface{}{
	"record_type": "process", // 按记录类型过滤
}
logs, total, err := storage.FindAll(1, 20, filters)

// 按进程名称过滤
filters = map[string]interface{}{
	"process_name": "数据同步任务",
}
logs, total, err := storage.FindAll(1, 10, filters)

// 按进程ID过滤
filters = map[string]interface{}{
	"process_id": "process-123456",
}
logs, total, err := storage.FindAll(1, 10, filters)

// 搜索进程记录
result, total, err := storage.Search("同步", 1, 10)
```

### Web界面查看

在调试器Web界面中，您可以：

1. 在日志列表页面使用"记录类型"过滤器查看进程记录
2. 通过"进程名称"或"进程ID"进行精确过滤
3. 在详情页面查看进程的完整执行日志和时间线
4. 使用关键词搜索功能查找特定进程记录

访问调试器界面：`http://localhost:8080/jcbase/debug/list`

### 查看完整示例

更多详细的使用示例，请查看：
- `example/debugger/process_usage.go` - 完整的进程级debugger使用示例

## Logger功能使用

### 在控制器中使用Logger

debugger组件提供了Logger接口，可以在业务控制器中记录调试日志。通过`GetLoggerFromContext`函数可以从Gin上下文中获取Logger实例。

#### 基本使用示例

```go
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
	r := gin.Default()

	// 创建调试器实例
	dbg, err := debugger.NewWithMemoryStorage(100)
	if err != nil {
		panic(err)
	}

	// 添加调试器中间件
	r.Use(dbg.Middleware())

	// 在控制器中使用Logger
	r.GET("/api/users", func(c *gin.Context) {
		// 从上下文中获取Logger实例
		logger := debugger.GetLoggerFromContext(c)

		// 记录不同级别的日志
		logger.Info("开始处理用户列表请求", map[string]interface{}{
			"query_params": c.Request.URL.Query(),
			"page":         c.Query("page"),
			"limit":        c.Query("limit"),
		})

		// 模拟数据库查询
		time.Sleep(50 * time.Millisecond)

		logger.Info("用户列表查询成功", map[string]interface{}{
			"user_count": 3,
			"status_code": http.StatusOK,
		})

		c.JSON(http.StatusOK, gin.H{
			"users": []gin.H{
				{"id": 1, "name": "张三"},
				{"id": 2, "name": "李四"},
				{"id": 3, "name": "王五"},
			},
		})
	})

	r.Run(":8080")
}
```

#### Logger接口方法

Logger接口支持的方法与[Logger功能](#logger功能)章节中定义的`LoggerInterface`接口一致，包括Debug、Info、Warn、Error四种日志级别记录方法，以及WithFields和GetLevel方法。

#### 使用WithFields添加结构化字段

```go
// 在控制器中使用WithFields
r.GET("/api/users/:id", func(c *gin.Context) {
	userID := c.Param("id")

	// 使用WithFields创建带有用户ID的logger
	logger := debugger.GetLoggerFromContext(c).WithFields(map[string]interface{}{
		"user_id": userID,
		"endpoint": "/api/users/:id",
	})

	logger.Info("开始查询用户详情")

	// 模拟用户不存在的情况
	if userID == "999" {
		logger.Warn("用户不存在", map[string]interface{}{
			"reason": "数据库中未找到该用户",
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	logger.Info("用户详情查询成功")
	c.JSON(http.StatusOK, gin.H{"id": userID, "name": "示例用户"})
})
```

#### 错误处理示例

```go
r.GET("/api/error", func(c *gin.Context) {
	logger := debugger.GetLoggerFromContext(c)

	logger.Info("开始处理错误示例请求")

	// 模拟业务逻辑错误
	err := fmt.Errorf("数据库连接失败: 连接超时")

	logger.Error("业务处理失败", map[string]interface{}{
		"error":         err.Error(),
		"error_type":    "database_connection_timeout",
		"retry_count":   3,
		"last_attempt":  time.Now().Format(time.RFC3339),
	})

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "服务器内部错误",
		"details": "请稍后重试",
	})
})
```

### 查看完整示例

更多详细的使用示例，请查看：
- `example/debugger/logger_usage.go` - 完整的Logger使用示例
- `example/debugger/basic_usage.go` - 基础使用示例

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

### 进程级调试器支持

调试器组件支持进程级日志记录，适用于后台任务、批处理作业、定时任务等非HTTP进程场景。

#### 基本使用

```go
// 开始进程记录
logger := dbg.StartProcess("数据同步任务", "batch")
defer dbg.EndProcess(logger.GetProcessID(), "completed")

// 记录进程日志
logger.Info("开始处理数据同步")
logger.Debug("获取数据源信息", map[string]interface{}{
    "source": "MySQL",
    "table": "users",
})

// 更新进度
logger.UpdateProgress(25.0)

// 记录警告和错误
logger.Warn("发现重复数据", map[string]interface{}{
    "duplicate_count": 5,
})

// 完成进程
logger.Info("数据同步完成", map[string]interface{}{
    "processed_count": 1000,
    "success_count": 995,
    "error_count": 5,
})
```

#### 进程记录器接口

```go
type ProcessLoggerInterface interface {
    Debug(msg any, fields ...map[string]interface{})
    Info(msg any, fields ...map[string]interface{})
    Warn(msg any, fields ...map[string]interface{})
    Error(msg any, fields ...map[string]interface{})
    GetProcessID() string
    UpdateProgress(progress float64)
    SetStatus(status string)
    AddProcessData(key string, value interface{})
}
```

### 流式请求支持

调试器组件支持流式请求的完整记录，包括流式响应的元数据和状态跟踪。

#### 启用流式支持

```go
config := &debugger.Config{
    Enabled:               true,
    EnableStreamingSupport: true,  // 启用流式请求支持
    MaxRecords:            1000,
}

dbg, err := debugger.New(config)
```

#### 流式请求记录

流式请求会自动记录以下信息：
- 流式请求标识
- 流式状态（started/processing/completed）
- 响应块数量
- 流式响应相关元数据

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

### Logger日志显示功能

debugger组件现在支持在详情页面中显示业务控制器中记录的logger日志。当您在控制器中使用`GetLoggerFromContext`记录日志时，这些日志会自动收集并在日志详情页面中显示。

#### 功能特性

- ✅ **自动收集**: 在请求处理过程中自动收集logger日志
- ✅ **级别区分**: 支持Debug、Info、Warn、Error四种日志级别
- ✅ **结构化字段**: 显示日志的附加字段信息
- ✅ **时间戳**: 记录每条日志的精确时间戳
- ✅ **响应式布局**: 适配不同屏幕尺寸的显示

#### 在详情页中查看Logger日志

1. 访问调试器列表页面：`http://localhost:8080/jcbase/debug/list`
2. 点击任意日志条目的"详情"按钮
3. 在详情页面中查看"Logger日志"区域

#### 示例效果

当您在控制器中记录如下日志：

```go
logger := debugger.GetLoggerFromContext(c)
logger.Info("开始处理请求", map[string]interface{}{
    "user_id": 123,
    "action": "create_user",
})
logger.Info("用户创建成功")
logger.Warn("密码强度较弱")
logger.Error("数据库连接失败", map[string]interface{}{
    "error": "connection timeout",
    "retry_count": 3,
})
```

在详情页面中，您将看到格式化的logger日志显示：

- **DEBUG** 开始处理请求 (user_id: 123, action: create_user)
- **INFO** 用户创建成功
- **WARN** 密码强度较弱
- **ERROR** 数据库连接失败 (error: connection timeout, retry_count: 3)

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
	
	// 注册调试器路由（使用默认路径：/jcbase/debug）
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
- 路径：`GET /jcbase/debug/`
- 功能：显示所有调试日志的列表，支持分页和搜索

#### 2. 日志详情页面
- 路径：`GET /jcbase/debug/detail/:id`
- 功能：显示单个日志的详细信息，包括请求头、响应头、请求体等

#### 3. 搜索功能
- 路径：`GET /jcbase/debug/list?q=关键词`
- 功能：在日志列表页面顶部提供搜索框，支持关键词搜索请求头、响应头、请求体等内容，搜索结果直接在列表页显示

#### 4. API接口
控制器还提供对应的API接口，方便前端集成：
- `GET /jcbase/debug/api/logs` - 获取日志列表（JSON格式）
- `GET /jcbase/debug/api/logs/:id` - 获取日志详情（JSON格式）
- `GET /jcbase/debug/api/search` - 搜索日志（JSON格式）

### 示例代码

项目提供了完整的示例代码，位于 `example/debugger/` 目录下：

- **`basic_usage.go`** - 基础使用示例，包含GIN框架集成和基本路由
- **`file_storage.go`** - 文件存储使用示例，支持日志文件管理
- **`controller_usage.go`** - 控制器使用示例，提供调试器Web界面
- **`config_examples.go`** - 配置示例，包含6种不同的配置方式
- **`process_usage.go`** - 进程级debugger使用示例，演示进程级日志记录功能

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
dbg, err := debugger.NewProductionDebugger("/var/log/debug_logs")  // 生产环境配置
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

	// 进程记录字段（用于非HTTP进程场景）
	RecordType      string            `json:"record_type,omitempty"`    // 记录类型："http" 或 "process"
	ProcessID       string            `json:"process_id,omitempty"`     // 进程ID
	ProcessName     string            `json:"process_name,omitempty"`   // 进程名称
	ProcessType     string            `json:"process_type,omitempty"`   // 进程类型：background/worker/cron/batch等
	Status          string            `json:"status,omitempty"`         // 进程状态：running/completed/failed
	Progress        float64           `json:"progress,omitempty"`       // 进度百分比（0-100）
	ProcessData     map[string]interface{} `json:"process_data,omitempty"` // 进程相关数据

	// 流式响应元数据
	IsStreaming     bool              `json:"is_streaming,omitempty"`    // 是否为流式请求
	StreamingStatus string            `json:"streaming_status,omitempty"` // 流式状态：started/processing/completed
	ChunkCount      int               `json:"chunk_count,omitempty"`     // 流式响应块数量
	StreamingData   map[string]interface{} `json:"streaming_data,omitempty"` // 流式响应相关数据
}
```

#### Config
```go
type Config struct {
	Enabled               bool             // 是否启用调试器
	Storage               Storage          // 存储器实例
	MaxBodySize           int64            // 最大请求/响应体大小（KB）
	RetentionPeriod       time.Duration    // 日志保留期限
	Level                 LogLevel         // 日志级别
	MaxRecords            int              // 最大记录数量
	SkipPaths             []string         // 跳过的路径
	SkipMethods           []string         // 跳过的HTTP方法
	SampleRate            float64          // 采样率（0-1之间）
	Logger                LoggerInterface  // 日志记录器实例
	AllowedIPs            []string         // 允许访问的IP白名单
	UseCDN                bool             // 是否使用CDN获取真实IP
	EnableStreamingSupport bool             // 启用流式请求支持
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

#### LoggerInterface接口
```go
type LoggerInterface interface {
	// Debug 记录调试级别日志
	Debug(msg any, fields ...map[string]interface{})

	// Info 记录信息级别日志
	Info(msg any, fields ...map[string]interface{})

	// Warn 记录警告级别日志
	Warn(msg any, fields ...map[string]interface{})

	// Error 记录错误级别日志
	Error(msg any, fields ...map[string]interface{})

	// WithFields 创建带有字段的日志记录器
	WithFields(fields map[string]interface{}) LoggerInterface

	// GetLevel 获取当前日志记录器的日志级别
	GetLevel() string
}
```

### 便捷构造函数

```go
// 便捷构造函数
func NewSimpleDebugger() (*Debugger, error)                    // 创建简单调试器，使用默认内存存储（150条记录）
func NewWithMemoryStorage(maxRecords int) (*Debugger, error)   // 创建使用内存存储的调试器，指定最大记录数
func NewWithFileStorage(path string, maxRecords int) (*Debugger, error) // 创建使用文件存储的调试器，指定存储路径和最大记录数
func NewWithCustomStorage(customStorage Storage) (*Debugger, error)      // 创建使用自定义存储器的调试器
func NewProductionDebugger(storagePath string) (*Debugger, error)        // 创建生产环境调试器，使用文件存储（1000条记录）

// 存储器构造函数
func NewMemoryStorage(maxRecords int) (Storage, error)                  // 创建内存存储器
func NewFileStorage(path string, maxRecords int) (Storage, error)       // 创建文件存储器
func NewDatabaseStorage(db *gorm.DB, tableName string) (Storage, error) // 创建数据库存储器
```

## 版本历史

- v1.0.0: 初始版本，支持基本调试功能
- 后续版本计划：支持更多存储后端、实时日志流、性能监控等

## 贡献指南

欢迎提交Issue和Pull Request来改进这个组件。

## 许可证

MIT License