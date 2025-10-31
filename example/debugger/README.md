# Debugger 组件示例

本目录包含 debugger 组件的使用示例，展示了如何在不同的场景下配置和使用调试器功能。

## 示例文件说明

### 1. basic_usage.go
**基本使用示例**
- 展示如何使用内存存储器的简单配置
- 包含基本的API路由和调试器中间件
- 适合快速入门和简单项目

**运行方式：**
```bash
go run basic_usage.go
```

### 2. file_storage.go
**文件存储示例**
- 展示如何将调试日志保存到文件系统中
- 包含采样率、保留时间等高级配置
- 适合生产环境和需要持久化存储的场景

**运行方式：**
```bash
go run file_storage.go
```

### 3. controller_usage.go
**控制器使用示例**
- 展示如何通过Web界面查看和管理调试日志
- 包含完整的REST API和调试器管理界面
- 适合需要可视化监控和管理的场景

**运行方式：**
```bash
go run controller_usage.go
```

### 4. config_examples.go
**配置示例**
- 展示各种构造函数和配置选项
- 包含便捷构造函数、自定义配置等不同方式
- 适合了解所有可用配置选项

**运行方式：**
```bash
go run config_examples.go
```

## 快速开始

### 1. 基本使用（推荐）
```go
import "github.com/jcbasego/component/debugger"

// 使用便捷构造函数
debuggerInstance, err := debugger.NewSimpleDebugger()
if err != nil {
    panic(err)
}

// 添加中间件到Gin路由
router.Use(debuggerInstance.Middleware())
```

### 2. 文件存储配置
```go
// 创建文件存储器
fileStorage, err := debugger.NewFileStorage("./logs", 1000)
if err != nil {
    panic(err)
}

// 创建调试器配置
config := &debugger.Config{
    Enabled:         true,
    Storage:         fileStorage,
    MaxRecords:      1000,
    RetentionPeriod: 7 * 24 * time.Hour, // 保留7天
    SampleRate:      0.5,                // 50%采样率
}

debuggerInstance, err := debugger.New(config)
```

### 3. 控制器配置
```go
// 创建控制器配置
controllerConfig := &debugger.ControllerConfig{
    BasePath: "/debug",
    Title:    "调试器管理界面",
}

// 注册控制器
debuggerGroup := router.Group(controllerConfig.BasePath)
debuggerInstance.WithController(debuggerGroup, controllerConfig)
```

## 配置选项说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Enabled | bool | true | 是否启用调试器 |
| Storage | Storage | 内存存储 | 日志存储方式 |
| MaxRecords | int | 1000 | 最大记录数 |
| RetentionPeriod | time.Duration | 168h | 日志保留时间 |
| SampleRate | float64 | 1.0 | 采样率（0.0-1.0） |
| SkipPaths | []string | [] | 跳过的路径 |
| SkipMethods | []string | [] | 跳过的HTTP方法 |
| MaxBodySize | int64 | 1024 | 最大请求体大小（KB） |

## 存储方式

### 1. 内存存储（默认）
- 速度快，适合开发和测试环境
- 重启后数据丢失
- 使用 `NewWithMemoryStorage(maxRecords)`

### 2. 文件存储
- 数据持久化，适合生产环境
- 日志保存在指定目录
- 使用 `NewWithFileStorage(path, maxRecords)`

### 3. 自定义存储
- 支持自定义存储实现
- 可集成数据库等外部存储
- 使用 `NewWithCustomStorage(storage)`

## 注意事项

1. **性能考虑**：在生产环境中建议设置合适的采样率
2. **存储空间**：文件存储需注意磁盘空间使用
3. **安全性**：调试器界面应限制访问权限
4. **数据清理**：定期清理过期日志数据

## 故障排除

### 常见问题

1. **端口冲突**：确保8080端口未被占用
2. **权限问题**：文件存储需要写权限
3. **内存不足**：调整MaxRecords限制

### 调试技巧

1. 检查中间件是否正确注册
2. 验证存储配置是否正确
3. 查看控制台输出获取详细错误信息

## 更多资源

- [组件文档](../component/debugger/README.md)
- [API参考](../component/debugger/)
- [测试用例](../component/debugger/)

## 贡献

欢迎提交问题和改进建议！