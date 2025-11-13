# ORM 组件

基于 GORM 的数据库 ORM 封装组件，支持 MySQL 和 SQLite 数据库，提供统一的数据库操作接口和便捷的 CRUD 功能。

## 概述

ORM 组件为 Go 应用程序提供了强大的数据库操作能力，支持多种数据库类型，具有以下特点：

- **多数据库支持**：支持 MySQL 和 SQLite 数据库
- **统一接口**：提供一致的数据库操作接口
- **自动表名处理**：支持表前缀、单复数表名配置
- **软删除支持**：内置软删除功能，支持自定义软删除字段
- **分页查询**：提供灵活的分页查询功能
- **连接池管理**：自动配置数据库连接池参数

## 功能特性

### 数据库连接管理
- 支持 MySQL 和 SQLite 数据库连接
- 自动配置连接池参数
- 支持调试模式
- 配置信息环境变量存储

### 表名处理
- 自动表前缀处理
- 单复数表名配置
- 自定义表名支持
- 表别名处理

### 模型解析
- 自动解析模型结构体
- 字段映射处理
- 软删除字段识别
- GORM 标签解析

### 分页查询
- 灵活的分页选项配置
- 自定义查询回调
- 结果集处理
- 总数统计

### 错误处理
- 错误收集机制
- 错误过滤处理
- 链式操作错误处理

## 安装指南

### 依赖要求

- Go 1.16+
- GORM v2.0+
- MySQL 驱动或 SQLite 驱动

### 安装依赖

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u gorm.io/driver/sqlite
```

## 使用示例

### 基本使用

#### MySQL 数据库连接

```go
import "github.com/jcbowen/jcbaseGo/component/orm/mysql"

// 配置数据库连接
dbConfig := jcbaseGo.DbStruct{
    Username:  "root",
    Password:  "password",
    Host:      "localhost",
    Port:      "3306",
    Dbname:    "testdb",
    Charset:   "utf8mb4",
    TablePrefix: "t_",
}

// 创建数据库实例
db := mysql.New(dbConfig)

// 获取数据库连接
gormDB := db.GetDb()
```

#### SQLite 数据库连接

```go
import "github.com/jcbowen/jcbaseGo/component/orm/sqlite"

// 配置 SQLite 数据库
sqliteConfig := jcbaseGo.SqlLiteStruct{
    DbFile:    "./data.db",
    TablePrefix: "t_",
}

// 创建 SQLite 实例
db := sqlite.New(sqliteConfig)

// 获取数据库连接
gormDB := db.GetDb()
```

### 模型定义

#### 基础模型定义

```go
import "github.com/jcbowen/jcbaseGo/component/orm/base"

// User 模型
type User struct {
    base.MysqlBaseModel  // 继承基础模型
    
    ID        uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    Username string `gorm:"column:username;type:varchar(50);not null" json:"username"`
    Email    string `gorm:"column:email;type:varchar(100);not null" json:"email"`
    
    // 软删除字段
    DeletedAt *string `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"`
}

// 自定义表名
func (User) TableName() string {
    return "users"
}

// 自定义配置别名
func (User) ConfigAlias() string {
    return "user_db"
}
```

### 分页查询

#### 基本分页查询

```go
// 定义查询选项
options := &mysql.FindPageOptions{
    Page:     1,
    PageSize: 10,
    PkId:     "id",
}

// 执行分页查询
result, err := db.FindForPage(&User{}, options)
if err != nil {
    // 处理错误
    return
}

// 使用查询结果
fmt.Printf("总数: %d, 当前页: %d\n", result.Total, result.Page)
for _, item := range result.List {
    user := item.(User)
    fmt.Printf("用户: %s, 邮箱: %s\n", user.Username, user.Email)
}
```

#### 自定义查询回调

```go
options := &mysql.FindPageOptions{
    Page:     1,
    PageSize: 10,
    
    // 自定义查询条件
    ListQuery: func(query *gorm.DB) *gorm.DB {
        return query.Where("status = ?", "active")
    },
    
    // 自定义排序
    ListOrder: func() interface{} {
        return "created_at DESC"
    },
    
    // 自定义结果处理
    ListEach: func(item interface{}) interface{} {
        user := item.(*User)
        // 处理用户数据
        return map[string]interface{}{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
        }
    },
}

result, err := db.FindForPage(&User{}, options)
```

### 表名处理

#### 自动表名处理

```go
// 自动添加表前缀和引号
tableName := "users"
db.TableName(&tableName, true)
// tableName 变为 "`t_users`"

// 仅添加表前缀
tableName2 := "products"
db.TableName(&tableName2)
// tableName2 变为 "t_products"
```

### 调试模式

```go
// 开启调试模式
db.Debug().FindForPage(&User{}, options)

// 调试模式会影响所有后续操作
debugDB := db.Debug()
debugDB.GetDb().Find(&users)  // 会输出 SQL 语句
```

### 与 Debugger 集成

#### 启用 SQL 日志记录（便捷用法）

```go
import (
    "time"
    "github.com/jcbowen/jcbaseGo/component/debugger"
    "github.com/jcbowen/jcbaseGo/component/orm/mysql"
)

// 初始化 Debugger
dbg, _ := debugger.New(&debugger.Config{Enabled: true, Level: "info"})

// 创建数据库实例并自动集成 SQL 日志
db := mysql.NewWithDebugger(dbConfig, dbg.GetLogger())

// 之后的所有 SQL 执行将按 Debugger 的日志级别记录
```

#### 运行时启用/配置 SQL 日志

```go
// 已有实例场景：启用 SQL 日志记录并自定义级别与慢查询阈值
db.EnableSQLLogging(dbg.GetLogger(), "debug", 100*time.Millisecond)

// 仅设置日志记录器（沿用默认阈值与 Debugger 级别）
db.SetDebuggerLogger(dbg.GetLogger())
```

#### 在 gorm.Open 中使用配置选项

```go
import (
    "time"
    "github.com/jcbowen/jcbaseGo/component/orm"
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
)

gormDB, err := gorm.Open(
    mysql.Open(dsn),
    orm.WithSQLLogging(dbg.GetLogger(), "info", 200*time.Millisecond),
)
```

日志内容包含：SQL 语句、耗时、影响行数、错误信息与慢查询标记；慢查询阈值与日志级别可按需调整。

## 详细功能说明

### 数据库配置

#### MySQL 配置结构

```go
type DbStruct struct {
    Username                                string `default:"root"`
    Password                                string `default:""`
    Protocol                                string `default:"tcp"`
    Host                                    string `default:"127.0.0.1"`
    Port                                    string `default:"3306"`
    Dbname                                  string `default:""`
    Charset                                 string `default:"utf8mb4"`
    ParseTime                               string `default:"True"`
    TablePrefix                             string `default:""`
    SingularTable                           bool   `default:"false"`
    DisableForeignKeyConstraintWhenMigrating bool   `default:"false"`
}
```

#### SQLite 配置结构

```go
type SqlLiteStruct struct {
    DbFile                                  string `default:""`
    TablePrefix                             string `default:""`
    SingularTable                           bool   `default:"false"`
    DisableForeignKeyConstraintWhenMigrating bool   `default:"false"`
}
```

### 模型解析功能

ORM 组件提供强大的模型解析功能：

1. **表名解析**：自动根据模型名称生成表名
2. **字段映射**：解析 GORM 标签，处理字段映射
3. **软删除识别**：自动识别软删除字段和条件
4. **配置别名**：支持多数据库配置

### 分页查询选项

```go
type FindPageOptions struct {
    // 查询配置
    Page        int  `default:"1"`     // 页码，默认 1
    PageSize    int  `default:"10"`   // 分页大小，默认 10，最大 1000
    ShowDeleted bool `default:"false"` // 是否显示软删除数据
    
    // 模型配置
    PkId            string `default:"id"` // 主键字段名
    ModelTableAlias string `default:""`   // 模型表别名
    
    // 回调函数
    ListQuery  func(*gorm.DB) *gorm.DB                   // 自定义查询条件
    ListSelect func(*gorm.DB) *gorm.DB                   // 自定义查询字段
    ListOrder  func() interface{}                        // 自定义排序
    ListEach   func(interface{}) interface{}             // 自定义结果处理
    ListReturn func(jcbaseGo.ListData) jcbaseGo.ListData // 自定义返回格式
}
```

## 高级用法

### 多数据库配置

```go
// 主数据库
mainDB := mysql.New(mainConfig, "main")

// 用户数据库
userDB := mysql.New(userConfig, "user_db")

// 日志数据库
logDB := sqlite.New(logConfig, "log_db")
```

### 自定义软删除字段

```go
type Product struct {
    base.MysqlBaseModel
    
    ID        uint   `gorm:"column:id;primaryKey" json:"id"`
    Name      string `gorm:"column:name" json:"name"`
    
    // 自定义软删除字段
    IsDeleted string `gorm:"column:is_deleted;soft_delete:0" json:"is_deleted"`
}

// 软删除条件：is_deleted = '0' 表示未删除
```

### 事务处理

```go
// 使用 GORM 的事务功能
err := db.GetDb().Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    
    return nil
})
```

## 性能优化建议

### 连接池配置

组件自动配置合理的连接池参数：

- **MySQL**：最大连接数 100，空闲连接数 10
- **SQLite**：最大连接数 1，空闲连接数 1（SQLite 限制）

### 查询优化

1. **使用索引**：确保常用查询字段有索引
2. **避免 N+1 查询**：使用 Preload 预加载关联数据
3. **分页优化**：合理设置分页大小
4. **字段选择**：只查询需要的字段

### 内存管理

1. **及时关闭连接**：长时间不用的连接会自动关闭
2. **结果集处理**：及时处理大结果集，避免内存泄漏

## 安全考虑

### SQL 注入防护

- 使用参数化查询
- 避免直接拼接 SQL 语句
- 使用 GORM 的安全查询方法

### 敏感信息保护

- 数据库密码存储在环境变量中
- 配置文件不包含敏感信息
- 使用安全的连接字符串

## API 参考

### Instance 接口

```go
type Instance interface {
    GetDb() *gorm.DB                    // 获取数据库连接
    GetConf() interface{}               // 获取配置信息
}
```

### MySQL Instance 方法

- `New(dbConfig DbStruct, opts ...string) *Instance` - 创建实例
- `NewWithDebugger(dbConfig DbStruct, debuggerLogger debugger.LoggerInterface, opts ...string) *Instance` - 创建并集成 Debugger
- `Debug() *Instance` - 开启调试模式
- `GetAllTableName() ([]AllTableName, error)` - 获取所有表名
- `TableName(tableName *string, quotes ...bool) *Instance` - 处理表名
- `FindForPage(model interface{}, options *FindPageOptions) (ListData, error)` - 分页查询
- `AddError(err error)` - 添加错误
- `Error() []error` - 获取错误列表
- `SetDebuggerLogger(debugger.LoggerInterface)` - 设置 Debugger 日志记录器
- `GetDebuggerLogger() debugger.LoggerInterface` - 获取 Debugger 日志记录器
- `EnableSQLLogging(debugger.LoggerInterface, opts ...interface{}) *Instance` - 启用 SQL 日志记录

### SQLite Instance 方法

- `New(conf SqlLiteStruct, opts ...string) *Instance` - 创建实例
- `NewWithDebugger(conf SqlLiteStruct, debuggerLogger debugger.LoggerInterface, opts ...string) *Instance` - 创建并集成 Debugger
- `Debug() *Instance` - 开启调试模式
- `GetAllTableName() ([]string, error)` - 获取所有表名
- `TableName(tableName *string, quotes ...bool) *Instance` - 处理表名
- `FindForPage(model interface{}, options *FindPageOptions) (ListData, error)` - 分页查询
- `AddError(err error)` - 添加错误
- `Error() []error` - 获取错误列表
- `SetDebuggerLogger(debugger.LoggerInterface)` - 设置 Debugger 日志记录器
- `GetDebuggerLogger() debugger.LoggerInterface` - 获取 Debugger 日志记录器
- `EnableSQLLogging(debugger.LoggerInterface, opts ...interface{}) *Instance` - 启用 SQL 日志记录

### 其他

- `orm.WithSQLLogging(debugger.LoggerInterface, opts ...interface{}) gorm.Option` - 在 gorm.Open 初始化阶段启用 SQL 日志记录

### 基础模型方法

- `ModelParse(modelType reflect.Type) (tableName, fields, softDeleteField, softDeleteCondition)` - 解析模型
- `BeforeCreate(tx *gorm.DB) error` - 创建前回调
- `BeforeUpdate(tx *gorm.DB) error` - 更新前回调
- `GetConfigAlias(model interface{}) string` - 获取配置别名

## 错误处理

组件提供完善的错误处理机制：

```go
// 检查错误
if len(db.Error()) > 0 {
    for _, err := range db.Error() {
        log.Printf("数据库错误: %v", err)
    }
}

// 链式操作错误处理
db.AddError(someError).TableName(&tableName)
// 如果有错误，后续操作会被跳过
```

## 常见问题

### 表名不匹配

确保模型实现了正确的 `TableName()` 方法，或检查表前缀配置。

### 连接失败

检查数据库配置信息，确保数据库服务正常运行。

### 性能问题

合理配置连接池参数，优化查询语句，添加必要的索引。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个组件。

## 许可证

MIT License