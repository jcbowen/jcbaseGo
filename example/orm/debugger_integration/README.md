# ORM与Debugger集成示例

本示例展示了如何在jcbaseGo项目中集成ORM组件与Debugger组件，实现自动记录SQL执行情况。

## 功能特点

1. **自动SQL日志记录**：所有数据库操作都会自动记录到debugger中
2. **慢查询检测**：可配置慢查询阈值，超过阈值的查询会被标记为慢查询
3. **错误记录**：SQL执行错误会被详细记录，包括错误类型和信息
4. **性能统计**：记录SQL执行时间、影响行数等性能指标
5. **多种使用方式**：支持多种集成方式，灵活配置

## 使用方式

### 方式1：使用NewWithDebugger创建实例（推荐）

```go
// 初始化debugger
debug, _ := debugger.New(&debugger.Config{
    Enabled: true,
    Level:   "debug",
})

// 数据库配置
dbConfig := jcbaseGo.DbStruct{
    Host:     "localhost",
    Port:     "3306",
    Username: "root",
    Password: "password",
    Dbname:   "test_db",
    Charset:  "utf8mb4",
}

// 创建数据库实例并自动启用SQL日志记录
db := mysql.NewWithDebugger(dbConfig, debug.GetLogger())
```

### 方式2：先创建实例，后启用SQL日志记录

```go
// 创建数据库实例
db := mysql.New(dbConfig)

// 启用SQL日志记录
db.EnableSQLLogging(debug.GetLogger(), logger.Info, 200*time.Millisecond)
```

### 方式3：运行时动态设置日志记录器

```go
// 创建数据库实例
db := mysql.New(dbConfig)

// 在运行时设置debugger日志记录器
db.SetDebuggerLogger(debug.GetLogger())
```

## 配置选项

### SQL日志级别
- `logger.Silent`：不记录任何SQL日志
- `logger.Error`：只记录错误SQL
- `logger.Warn`：记录错误和慢查询
- `logger.Info`：记录所有SQL（默认）

### 慢查询阈值
可以设置慢查询的时间阈值，超过该时间的查询会被标记为慢查询：

```go
// 设置100ms为慢查询阈值
db.EnableSQLLogging(debug.GetLogger(), logger.Info, 100*time.Millisecond)
```

## 日志内容

每条SQL日志记录包含以下信息：

- **SQL语句**：执行的完整SQL语句
- **执行时间**：SQL执行耗时（毫秒）
- **影响行数**：SQL影响的行数
- **错误信息**：如果执行失败，包含错误详情
- **慢查询标记**：如果超过阈值，标记为慢查询

## 运行示例

### 运行Web服务示例
```bash
cd /Users/bowen/projects/mine/jcbaseGo/example/orm/debugger_integration
go run main.go
```

访问：
- API服务：http://localhost:8080
- 调试面板：http://localhost:8080/debugger

### 运行演示程序
```bash
cd /Users/bowen/projects/mine/jcbaseGo/example/orm/debugger_integration
go run demo.go
```

## API端点

Web服务示例提供以下API端点：

- `GET /users` - 查询用户列表
- `POST /users` - 创建新用户  
- `GET /slow-query` - 慢查询演示

## 注意事项

1. **性能影响**：启用SQL日志记录会对性能产生轻微影响，建议在生产环境中根据需要启用
2. **日志级别**：选择合适的日志级别，避免记录过多不必要的日志
3. **慢查询阈值**：根据实际业务需求设置合理的慢查询阈值
4. **存储容量**：注意debugger的存储容量限制，定期清理过期日志

## 扩展功能

可以通过实现自定义的`debugger.LoggerInterface`接口来扩展日志记录功能，例如：
- 将SQL日志发送到远程日志服务
- 添加自定义的日志字段
- 实现日志聚合和分析