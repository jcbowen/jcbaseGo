# jcbaseGo

一个功能丰富的 Go 语言基础工具包，提供多种可复用的组件和工具函数，帮助开发者快速构建高质量的 Go 应用程序。

## 🚀 功能特性

- **🔐 安全组件**: SM4/AES 加密、密码处理、数据验证
- **🗄️ 数据库 ORM**: MySQL/SQLite 支持，基于 GORM 的抽象层，包含基础模型和 CRUD trait
- **📧 邮件服务**: 支持 SMTP 邮件发送，HTML/文本邮件，附件支持
- **📁 附件管理**: 本地、FTP、SFTP、OSS、COS 等多种存储方式
- **🛠️ 工具函数**: 类型转换、字符串处理、JSON 操作、文件处理、HTTP工具、IP地址处理、SSH密钥管理、单位转换等
- **💾 缓存支持**: Redis 缓存组件，连接池优化
- **✅ 数据验证**: 邮箱、手机号、身份证、URL、IP 等验证
- **🎛️ 配置管理**: 支持 JSON、INI、命令行等多种配置源
- **🔄 升级工具**: Git 代码自动升级，支持回滚和备份
- **🔗 TLS 配置**: 完整的 TLS/SSL 证书管理
- **🐘 PHP 集成**: 内置 PHP 解释器，支持混合开发
- **🔗 中间件支持**: 跨域(CORS)、真实IP获取、请求参数解析(支持 JSON、表单、multipart、XML 格式)
- **🔍 HTTP调试工具**: Gin框架HTTP请求调试，支持多存储方式、日志查询、全文搜索、流式请求记录、进程级调试
- **💬 消息提示组件**: 消息渲染和快捷函数，支持配置管理和模板应用
- **🌐 HTTP工具**: HTTP请求处理、URL构建等
- **📍 IP地址处理**: IP验证、地理位置查询
- **🔑 SSH密钥管理**: SSH密钥生成、获取和管理
- **📏 单位转换**: 长度、重量、时间等单位转换

## 📦 安装

### 基础安装

```bash
go get github.com/jcbowen/jcbaseGo
```

### 依赖要求

- **Go**: 1.23.0+ (推荐 1.23.4+)
- **MySQL**: 5.7+ 或 8.0+ (可选，用于数据库功能)
- **SQLite**: 3.x (可选，用于轻量级数据库)
- **Redis**: 6.0+ (可选，用于缓存功能)
- **PHP**: 7.4+ 或 8.x (可选，用于 PHP 集成功能)

### 完整安装 (包含可选依赖)

```bash
# 安装核心包
go get github.com/jcbowen/jcbaseGo

# 安装 GORM 相关驱动
go get gorm.io/driver/mysql
go get gorm.io/driver/sqlite
go get gorm.io/gorm

# 安装 Redis 客户端
go get github.com/go-redis/redis/v8

# 安装其他常用依赖
go get github.com/gin-gonic/gin
go get github.com/go-playground/validator/v10
```

## 🏗️ 项目结构

```
jcbaseGo/
├── component/                   # 核心组件目录
│   ├── attachment/             # 📁 附件管理组件
│   │   ├── attachment.go       # 主附件管理器
│   │   ├── method.go           # 附件操作方法
│   │   └── remote/             # 远程存储实现
│   │       ├── cos.go          # 腾讯云 COS 存储
│   │       ├── ftp.go          # FTP 文件传输
│   │       ├── oss.go          # 阿里云 OSS 存储
│   │       ├── sftp.go         # SFTP 安全传输
│   │       └── remote.go       # 远程存储接口定义
│   ├── command/                # 💻 命令行工具
│   │   └── main.go             # 命令执行封装
│   ├── debugger/               # 🔍 调试工具（请求、进程）
│   │   ├── controller.go       # 调试控制器
│   │   ├── storage.go          # 存储实现（支持内存、文件、数据库）
│   │   ├── debugger.go         # 调试器主文件
│   │   ├── query_manager.go    # 查询管理器
│   │   ├── detail_viewer.go    # 详情查看器
│   │   ├── README.md           # 调试器文档
│   │   └── 18个相关文件        # 完整的调试功能实现（流式请求、进程级调试等）
│   ├── helper/                 # 🛠️ 工具函数集合
│   │   ├── convert.go          # 类型转换工具
│   │   ├── file.go             # 文件操作工具
│   │   ├── http.go             # HTTP相关工具
│   │   ├── ip.go               # IP地址处理
│   │   ├── json.go             # JSON 处理工具
│   │   ├── mask.go             # 敏感信息脱敏
│   │   ├── money.go            # 金额处理工具
│   │   ├── ssh.go              # SSH密钥管理
│   │   ├── string.go           # 字符串处理工具
│   │   ├── unit.go             # 单位转换工具
│   │   └── util.go             # 通用工具函数
│   ├── mailer/                 # 📧 邮件发送组件
│   │   └── mailer.go           # SMTP 邮件服务
│   ├── message/                # 💬 消息提示组件
│   │   ├── config.go           # 配置管理器
│   │   ├── message.go          # 消息提示组件
│   │   └── renderer.go         # 消息渲染
│   ├── orm/                    # 🗄️ 数据库 ORM 抽象层
│   │   ├── instance.go         # 数据库实例接口
│   │   ├── base/               # 基础模型定义
│   │   │   ├── base_mysql.go   # MySQL 基础模型
│   │   │   ├── base_sqlite.go  # SQLite 基础模型
│   │   │   └── model_utils.go  # 模型工具函数
│   │   ├── mysql/              # MySQL 数据库实现
│   │   │   └── main.go
│   │   └── sqlite/             # SQLite 数据库实现
│   │       └── main.go
│   ├── php/                    # 🐘 PHP 解释器集成
│   │   ├── jcbasePHP.go        # PHP 解释器接口
│   │   └── main.go             # PHP 集成主文件
│   ├── ratelimit/              # 🚦 滑动窗口限流（Redis + 本地降级）
│   │   └── ratelimit.go        # 限流器实现
│   ├── redis/                  # 💾 Redis 缓存组件
│   │   ├── cache.go            # 缓存操作实现
│   │   └── main.go             # Redis 连接管理
│   ├── security/               # 🔐 安全相关功能
│   │   ├── aes.go              # AES 加密算法
│   │   ├── base.go             # 安全基础功能
│   │   ├── id.go               # uint64 ID 加解密
│   │   ├── password.go         # 密码哈希处理
│   │   ├── safe.go             # 安全验证工具
│   │   └── sm4.go              # SM4 国密算法
│   ├── serializer/             # 🔄 响应数据递归序列化
│   │   └── serializer.go       # 统一序列化器（结构体转 map、时间、ID 加密）
│   ├── snowflake/              # ❄️ 雪花算法分布式 ID
│   │   └── snowflake.go        # Snowflake ID 生成器
│   ├── timezone/               # 🌏 时区格式化（UTC ↔ 北京时间）
│   │   └── timezone.go         # 时区工具与 Formatter
│   ├── tlsconfig/              # 🔒 TLS 配置管理读取
│   │   └── tlsconfig.go        # TLS 配置实现
│   ├── trait/                  # 🎭 Trait 模式实现
│   │   └── crud/               # CRUD 操作模板
│   │       ├── all.go          # 获取所有数据
│   │       ├── base.go         # CRUD 基础功能
│   │       ├── create.go       # 创建操作
│   │       ├── delete.go       # 删除操作
│   │       ├── detail.go       # 详情查询
│   │       ├── list.go         # 列表查询
│   │       ├── save.go         # 智能保存
│   │       ├── set-value.go    # 字段值设置
│   │       ├── update.go       # 更新操作
│   │       └── ReadMe.md       # CRUD 使用文档
│   ├── upgrade/                # 🔄 代码升级工具
│   │   └── main.go             # Git 自动升级
│   └── validator/              # ✅ 数据验证组件
│       └── main.go             # 验证器实现
├── config.go                   # 📋 全局配置管理
├── type.go                     # 📐 全局类型定义
├── errcode/                    # ❌ 错误码定义
│   └── errcode.go              # 标准错误码
├── example/                    # 📖 使用示例
│   ├── README.md               # 示例总览文档
│   ├── security/               # 安全组件示例
│   ├── helper/                 # 工具函数示例
│   ├── orm/                    # 数据库操作示例
│   ├── mailer/                 # 邮件发送示例
│   ├── redis/                  # Redis 缓存示例
│   ├── validator/              # 数据验证示例
│   ├── attachment/             # 附件管理示例
│   ├── debugger/               # 调试器组件示例
│   └── php/                    # PHP 集成示例
├── middleware/                 # 🔗 中间件集合
│   ├── main.go                 # 通用中间件
│   └── response.go             # 统一 JSON 响应与数据序列化
├── go.mod                      # 📦 Go 模块定义
├── go.sum                      # 🔐 依赖锁定文件
└── LICENSE                     # 📄 MIT 许可证
```

## 🎯 快速开始

### 1. 基础 CRUD 操作

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/orm/base"
    "github.com/jcbowen/jcbaseGo/component/orm/mysql"
    "github.com/jcbowen/jcbaseGo/component/trait/crud"
    "officeAutomation/library"
)

// User 用户模型
type User struct {
    base.MysqlBaseModel                           // 继承基础模型 (ID, CreatedAt, UpdatedAt, DeletedAt)
    Username string `gorm:"uniqueIndex;size:50" json:"username"` // 用户名
    Email    string `gorm:"index;size:100" json:"email"`         // 邮箱
    Status   int    `gorm:"default:1" json:"status"`             // 状态
}

func (User) TableName() string {
    return "users"
}

// UserController 用户控制器
type UserController struct {
    *crud.Trait
}

func NewUserController() *UserController {
    // 配置数据库连接
    dbConfig := jcbaseGo.DbStruct{
        Host:         "localhost",
        Port:         "3306",
        Username:     "root",
        Password:     "password",
        Dbname:       "test_db",
        Charset:      "utf8mb4",
        TablePrefix:  "tb_",
        SingularTable: false,
    }
    
    // 创建数据库实例
    db := mysql.New(dbConfig)
    
    // 初始化控制器
    controller := &UserController{
        Trait: &crud.Trait{
            Model: &User{},
            DBI:   db,
        },
    }
    
    controller.Trait.Controller = controller
    return controller
}

// CheckInit 控制器初始化确认方法
// 参数：
//   - ctx any: crud上下文对象(*crud.Context或者*gin.Context,不同的传入，处理不同的逻辑)
//
// 返回：
//   - *crud.Context: 确认后的crud上下文对象
//
// 功能：在CRUD初始化时调用，用于控制器级别的初始化确认
func (uc *UserController) CheckInit(ctx any) *crud.Context {
    var (
        crudCtx *crud.Context
        ok      bool
    )
    // 确认crud上下文对象是否为*crud.Context类型
    if crudCtx, ok = ctx.(*crud.Context); ok {
        // 根据运行模式设置调试标识（示例用 Gin 模式）
        crudCtx.Debug = library.Mode == "debug"
    } else if ginCtx, ok := ctx.(*gin.Context); ok {
        // 传入的是 *gin.Context 时，构建一个新的 CRUD 上下文
        crudCtx = crud.NewContext(&crud.NewContextOpt{
            ActionName: "custom",
            GinContext: ginCtx,
            Debug:      library.Mode == "debug",
        })
    }

    return crudCtx
}

// CreateBefore 创建前的数据验证
func (uc *UserController) CreateBefore(ctx *crud.Context, modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
    user := modelValue.(*User)
    
    // 检查用户名是否已存在
    var count int64
    uc.DBI.GetDb().Model(&User{}).Where("username = ?", user.Username).Count(&count)
    if count > 0 {
        return nil, nil, errors.New("用户名已存在")
    }
    
    return user, mapData, nil
}

// ListEach 列表数据处理
func (uc *UserController) ListEach(ctx *crud.Context, item interface{}) interface{} {
    user := item.(*User)
    // 可以在这里添加计算字段或隐藏敏感信息
    return user
}

func main() {
    r := gin.Default()
    
    userController := NewUserController()
    
    // 注册 CRUD 路由
    api := r.Group("/api/users")
    {
        api.GET("/list", userController.ActionList)        // 获取用户列表
        api.GET("/detail", userController.ActionDetail)    // 获取用户详情
        api.POST("/create", userController.ActionCreate)   // 创建用户
        api.POST("/update", userController.ActionUpdate)   // 更新用户
        api.POST("/save", userController.ActionSave)       // 智能保存
        api.POST("/delete", userController.ActionDelete)   // 删除用户
        api.GET("/all", userController.ActionAll)          // 获取所有用户
        api.POST("/set-value", userController.ActionSetValue) // 设置字段值
    }
    
    r.Run(":8080")
}
```

### 2. 安全加密

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    // SM4 国密加密 (推荐使用 GCM 模式)
    sm4 := security.SM4{
        Text: "敏感数据需要加密",
        Key:  "1234567890123456", // 16字节密钥
        Iv:   "abcdefghijklmnop", // 16字节初始向量
        Mode: "GCM",               // 推荐使用 GCM 模式
    }

    var cipherText string
    err := sm4.Encrypt(&cipherText)
    if err != nil {
        panic(err)
    }
    fmt.Printf("SM4 加密结果: %s\n", cipherText)

    // 解密
    sm4Decrypt := security.SM4{
        Text: cipherText,
        Key:  "1234567890123456",
        Iv:   "abcdefghijklmnop",
        Mode: "GCM",
    }
    
    var plainText string
    err = sm4Decrypt.Decrypt(&plainText)
    if err != nil {
        panic(err)
    }
    fmt.Printf("SM4 解密结果: %s\n", plainText)

    // AES 加密
    aes := security.AES{
        Text: "Hello, AES Encryption!",
        Key:  "1234567890123456", // 16字节密钥 (AES-128)
        Iv:   "abcdefghijklmnop",
    }

    err = aes.Encrypt(&cipherText)
    if err != nil {
        panic(err)
    }
    fmt.Printf("AES 加密结果: %s\n", cipherText)

    // 密码安全处理
    password := "user_password_123"
    hashedPassword := security.PasswordHash(password)
    fmt.Printf("密码哈希: %s\n", hashedPassword)
    
    // 验证密码
    isValid := security.PasswordVerify(password, hashedPassword)
    fmt.Printf("密码验证结果: %v\n", isValid)
}
```

### 3. 邮件发送

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/mailer"
)

func main() {
    // 配置邮件服务
    mailConfig := jcbaseGo.MailerStruct{
        Host:     "smtp.qq.com",
        Port:     "587",
        Username: "your-email@qq.com",
        Password: "your-smtp-password", // QQ邮箱需要使用授权码
        From:     "your-email@qq.com",
    }

    // 创建邮件实例
    email := mailer.New(mailConfig)

    // 发送文本邮件
    email.AddRecipient("recipient@example.com")
    email.SetSubject("测试邮件")
    email.SetBody("这是一封测试邮件", false)
    err := email.Send()
    if err != nil {
        panic(err)
    }

    // 发送 HTML 邮件
    htmlContent := `
    <h1>欢迎注册我们的服务</h1>
    <p>感谢您的注册，请点击下面的链接激活账户：</p>
    <a href="https://example.com/activate?token=abc123">激活账户</a>
    `

    emailHTML := mailer.New(mailConfig)
    emailHTML.AddRecipient("recipient@example.com")
    emailHTML.SetSubject("账户激活")
    emailHTML.SetBody(htmlContent, true)
    err = emailHTML.Send()
    if err != nil {
        panic(err)
    }

    // 批量发送邮件
    recipients := []string{
        "user1@example.com",
        "user2@example.com",
        "user3@example.com",
    }

    for _, recipient := range recipients {
        emailBatch := mailer.New(mailConfig)
        emailBatch.AddRecipient(recipient)
        emailBatch.SetSubject("批量通知")
        emailBatch.SetBody("这是一封批量发送的邮件", false)
        err := emailBatch.Send()
        if err != nil {
            fmt.Printf("发送到 %s 失败: %v\n", recipient, err)
        }
    }
}
```

### 4. 附件管理

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    // 创建 gin 上下文（实际使用时应从请求中获取）
    r := gin.Default()
    r.POST("/upload", func(c *gin.Context) {
        // 本地文件存储配置
        baseConfig := &jcbaseGo.AttachmentStruct{
            StorageType: "local",
            LocalDir:    "./uploads",
        }

        // 创建附件实例
        attach := attachment.New(c, baseConfig)

        // 获取上传的文件
        fileHeader, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // 上传文件
        attach.Upload(&attachment.Options{
            FileData: fileHeader,
            FileType: "image",
        })
        attach.Save()

        if attach.HasError() {
            c.JSON(400, gin.H{"error": attach.Error().Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "文件上传成功",
            "file":    attach.FileAttachment,
        })
    })

    r.Run(":8080")
}
```

### 5. Redis 缓存

```go
package main

import (
    "fmt"
    "time"

    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
)

func main() {
    // 配置 Redis 连接
    redisConfig := jcbaseGo.RedisStruct{
        Host:     "localhost",
        Port:     "6379",
        Password: "", // Redis 密码
        Db:       "0", // 数据库编号
    }

    // 创建 Redis 实例
    rdb := redis.New(redisConfig)

    // 设置缓存
    err := rdb.Set("user:1001", "用户数据", 30*time.Minute)
    if err != nil {
        panic(err)
    }

    // 获取缓存
    value, err := rdb.GetString("user:1001")
    if err != nil {
        panic(err)
    }
    fmt.Printf("缓存值: %s\n", value)

    // 设置哈希缓存
    err = rdb.HSet("user:profile:1001", "name", "张三")
    if err != nil {
        panic(err)
    }
    err = rdb.HSet("user:profile:1001", "email", "zhangsan@example.com")
    if err != nil {
        panic(err)
    }
    err = rdb.HSet("user:profile:1001", "age", 25)
    if err != nil {
        panic(err)
    }

    // 获取哈希缓存
    profile, err := rdb.HGetAll("user:profile:1001")
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户资料: %+v\n", profile)

    // 列表操作
    err = rdb.LPush("message_queue", "消息1", "消息2", "消息3")
    if err != nil {
        panic(err)
    }

    // 消费队列消息
    message, err := rdb.RPop("message_queue")
    if err != nil {
        panic(err)
    }
    fmt.Printf("队列消息: %s\n", message)
}
```

### 6. 数据验证

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/validator"
)

func main() {
    // 邮箱验证
    email := "user@example.com"
    if validator.IsEmail(email) {
        fmt.Printf("%s 是有效的邮箱地址\n", email)
    }

    // 中国大陆手机号验证
    mobile := "13812345678"
    if validator.IsMobile(mobile) {
        fmt.Printf("%s 是有效的手机号\n", mobile)
    }

    // 身份证号验证
    idCard := "110101199001011234"
    if validator.IsChineseIDCard(idCard) {
        fmt.Printf("%s 是有效的身份证号\n", idCard)
    }

    // URL 验证
    urlStr := "https://www.example.com"
    if validator.IsURL(urlStr) {
        fmt.Printf("%s 是有效的URL\n", urlStr)
    }

    // IP 地址验证
    ip := "192.168.1.1"
    if valid, ipType := validator.IsIP(ip); valid {
        if ipType == validator.IPv4 {
            fmt.Printf("%s 是有效的IPv4地址\n", ip)
        } else if ipType == validator.IPv6 {
            fmt.Printf("%s 是有效的IPv6地址\n", ip)
        }
    }

    // 端口验证
    port := "8080"
    if validator.IsPort(port) {
        fmt.Printf("%s 是有效的端口号\n", port)
    }
}
```

### 7. 工具函数

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 类型转换
    converter := helper.Convert{Value: "123"}
    intValue := converter.ToInt()
    floatValue := converter.ToFloat64()
    boolValue := helper.Convert{Value: "true"}.ToBool()

    fmt.Printf("转换结果: int=%d, float=%.2f, bool=%v\n", intValue, floatValue, boolValue)

    // 字符串处理
    str := helper.NewStr("Hello World")
    snakeCase := str.ConvertCamelToSnake()    // hello_world
    substr := str.ByteSubstr(0, 5)           // Hello
    trimmed := str.TrimSpace()               // "Hello World"

    fmt.Printf("字符串处理: snake=%s, substr=%s, trimmed=%s\n", snakeCase, substr, trimmed)

    // JSON 处理
    data := map[string]interface{}{
        "name": "张三",
        "age":  25,
        "city": "北京",
    }

    var jsonStr string
    helper.Json(data).ToString(&jsonStr)
    fmt.Printf("JSON字符串: %s\n", jsonStr)

    // 从JSON字符串解析
    var parsedData map[string]interface{}
    helper.Json(jsonStr).ToStruct(&parsedData)
    fmt.Printf("解析后的数据: %+v\n", parsedData)

    // 金额处理 (以厘为单位，1000厘=1元)
    money := helper.Money("123.45")
    yuanStr := money.FloatString()           // "123.45"
    formattedStr := money.FloatString("¥")   // "¥123.45"

    fmt.Printf("金额处理: 元=%s, 格式化=%s\n", yuanStr, formattedStr)

    // 文件操作
    file := helper.NewFile(&helper.File{Path: "./test.txt"})

    // 创建文件
    err := file.CreateFile([]byte("Hello, jcbaseGo!"), true)
    if err != nil {
        fmt.Printf("创建文件失败: %v\n", err)
    }

    // 检查文件是否存在
    if file.Exists() {
        fmt.Println("文件存在")

        // 获取文件名
        basename := file.Basename("")
        fmt.Printf("文件名: %s\n", basename)
    }

    // 读取JSON文件
    var jsonData interface{}
    err = file.JsonToData(&jsonData)
    if err != nil {
        fmt.Printf("读取JSON文件失败: %v\n", err)
    }
}
```

### 8. PHP 集成

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
    // 初始化 PHP 解释器
    opt := jcbaseGo.Option{
        ConfigSource: "./config.json", // 配置文件路径
    }
    phpEngine := php.New(opt)

    // 调用 PHP 函数
    result, err := phpEngine.RunFunc("calculate", "10", "5", "add")
    if err != nil {
        panic(err)
    }
    fmt.Printf("PHP 计算结果: %s\n", result)

    // 调用自定义 PHP 函数处理数据
    result, err = phpEngine.RunFunc("processUsers")
    if err != nil {
        panic(err)
    }
    fmt.Printf("PHP 数组处理结果: %s\n", result)
}
```

### 9. 命令行工具

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/command"
)

func main() {
    // 执行简单的系统命令
    output, err := command.Run("ls", "-la")
    if err != nil {
        panic(err)
    }
    fmt.Printf("目录列表:\n%s\n", output)

    // 切换工作目录后执行命令
    command.CmdPath = "/tmp"
    output, err = command.Run("pwd")
    if err != nil {
        panic(err)
    }
    fmt.Printf("当前路径: %s\n", output)

    // 执行 Git 命令
    output, err = command.Run("git", "status")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Git 状态:\n%s\n", output)

    // 执行带参数的命令
    output, err = command.Run("echo", "Hello, jcbaseGo!")
    if err != nil {
        panic(err)
    }
    fmt.Printf("命令输出: %s\n", output)

    // 批量执行命令
    commands := [][]string{
        {"pwd"},
        {"whoami"},
        {"date"},
    }

    for _, cmd := range commands {
        args := []string{}
        if len(cmd) > 1 {
            args = cmd[1:]
        }
        output, err := command.Run(cmd[0], args...)
        if err != nil {
            fmt.Printf("命令 %s 执行失败: %v\n", cmd[0], err)
            continue
        }
        fmt.Printf("%s 输出: %s\n", cmd[0], output)
    }
}
```

### 10. HTTP 调试工具

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo/component/debugger"
)

func main() {
    r := gin.Default()

    // 使用简单调试器（内存存储，默认配置）
    dbg, err := debugger.NewSimpleDebugger()
    if err != nil {
        panic(err)
    }
    r.Use(dbg.Middleware())

    // 或者使用自定义配置
    config := &debugger.Config{
        Enabled: true,
    }

    customDbg, err := debugger.New(config)
    if err != nil {
        panic(err)
    }
    // 注册调试器路由（可选，用于查看调试日志）
    customDbg.WithController(r, nil)
    r.Use(customDbg.Middleware())

    // 添加测试路由
    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "users": []map[string]interface{}{
                {"id": 1, "name": "张三"},
                {"id": 2, "name": "李四"},
            },
        })
    })

    r.POST("/api/users", func(c *gin.Context) {
        var user struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }

        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        c.JSON(201, gin.H{
            "message": "用户创建成功",
            "user": user,
        })
    })

    // 启动服务器
    r.Run(":8080")
}
```

### 11. 消息提示组件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo/component/message"
)

func main() {
    r := gin.Default()

    r.GET("/success", func(c *gin.Context) {
        // 成功消息
        message.Success(c, "操作成功", "您的请求已成功处理")
    })

    r.GET("/error", func(c *gin.Context) {
        // 错误消息
        message.Error(c, "操作失败", "请检查输入参数")
    })

    r.GET("/info", func(c *gin.Context) {
        // 信息消息
        message.Info(c, "系统提示", "新版本即将发布")
    })

    r.GET("/warning", func(c *gin.Context) {
        // 警告消息
        message.Warning(c, "注意安全", "请及时修改密码")
    })

    r.GET("/custom", func(c *gin.Context) {
        // 自定义消息（带跳转）
        message.Render(c, "自定义标题", "自定义内容", "custom",
            message.WithRedirect("/home"),
            message.WithAutoRedirect(true),
        )
    })

    // API 响应格式
    r.GET("/api/response", func(c *gin.Context) {
        // JSON 格式的响应
        message.APIResponseWithMessage(c, "200", "操作成功", map[string]interface{}{
            "user": map[string]string{
                "name":  "张三",
                "email": "zhangsan@example.com",
            },
        })
    })

    // 简化版消息
    r.GET("/simple", func(c *gin.Context) {
        message.Simple(c, "操作已完成", true)
    })

    // 启动服务器
    r.Run(":8080")
}
```

## 📚 组件详细说明

### 🔐 安全组件 (security/)

#### SM4 国密算法
- **支持模式**: CBC、GCM (推荐 GCM)
- **密钥长度**: 128位 (16字节)
- **特点**: 符合国密标准，适用于敏感数据加密

```go
// GCM 模式 (推荐 - 提供认证加密)
sm4 := security.SM4{
    Text: "敏感数据",
    Key:  "1234567890123456", // 16字节
    Mode: "GCM",
}

// CBC 模式 (需要 IV)
sm4 := security.SM4{
    Text: "敏感数据",
    Key:  "1234567890123456", // 16字节
    Iv:   "abcdefghijklmnop", // 16字节
    Mode: "CBC",
}
```

#### AES 标准算法
- **支持密钥**: 128/192/256位
- **模式**: CBC
- **应用**: 通用数据加密

#### 密码安全
- **哈希算法**: bcrypt (推荐)
- **盐值**: 自动生成
- **成本因子**: 可配置

### 🗄️ 数据库 ORM (orm/)

#### 基础模型
```go
// MySQL 基础模型
type User struct {
    base.MysqlBaseModel                    // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
    Username string `gorm:"uniqueIndex"`   // 业务字段
}

// SQLite 基础模型
type Product struct {
    base.SqliteBaseModel                   // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
    Name string `gorm:"size:100"`          // 业务字段
}
```

#### CRUD Trait 特性
- **8个标准接口**: Create, Update, Delete, List, Detail, All, Save, SetValue
- **钩子方法**: Before/After 钩子支持自定义逻辑
- **软删除**: 灵活的软删除配置
- **事务安全**: 自动事务管理
- **分页支持**: 内置分页功能

#### 软删除配置
```go
type User struct {
    base.MysqlBaseModel
    // 方式1: 使用默认 deleted_at 字段
    
    // 方式2: 自定义字段名和条件
    IsDeleted string `gorm:"soft_delete:IS NULL"`
    
    // 方式3: 使用状态字段
    Status int `gorm:"soft_delete:= 1"`
}
```

### 📧 邮件服务 (mailer/)

#### 功能特性
- **SMTP 协议**: 标准 SMTP 支持
- **安全连接**: TLS/SSL 支持
- **多种格式**: 文本、HTML 邮件
- **附件支持**: 文件附件功能
- **批量发送**: 支持批量邮件发送

#### 常用邮箱配置
```go
// QQ邮箱
mailer.Mailer{
    Host: "smtp.qq.com",
    Port: 587, // 或 465 (SSL)
    Username: "your-email@qq.com",
    Password: "授权码", // 不是登录密码
}

// 网易邮箱
mailer.Mailer{
    Host: "smtp.163.com",
    Port: 587,
    Username: "your-email@163.com",
    Password: "授权码",
}

// Gmail
mailer.Mailer{
    Host: "smtp.gmail.com",
    Port: 587,
    Username: "your-email@gmail.com",
    Password: "应用专用密码",
}
```

### 📁 附件管理 (attachment/)

#### 支持的存储类型
- **本地存储**: 本地文件系统
- **FTP**: 标准 FTP 协议
- **SFTP**: SSH 文件传输协议
- **阿里云 OSS**: 对象存储服务
- **腾讯云 COS**: 云对象存储

#### 文件类型安全
```go
// 允许的文件类型
allowedTypes := []string{
    "jpg", "jpeg", "png", "gif", "webp",    // 图片
    "pdf", "doc", "docx", "xls", "xlsx",    // 文档
    "zip", "rar", "7z",                     // 压缩包
    "mp4", "avi", "mov",                    // 视频
}

// 文件大小限制
maxSize := 10 * 1024 * 1024 // 10MB
```

### 💾 缓存支持 (redis/)

#### Redis 操作
- **基础操作**: GET, SET, DEL, EXISTS
- **哈希操作**: HGET, HSET, HGETALL, HDEL
- **列表操作**: LPUSH, RPUSH, LPOP, RPOP
- **集合操作**: SADD, SREM, SMEMBERS
- **有序集合**: ZADD, ZREM, ZRANGE

#### 连接池配置
```go
config := redis.Config{
    Host:     "localhost",
    Port:     "6379",
    Password: "",
    DB:       0,
    PoolSize: 10,                    // 连接池大小
    MinIdleConns: 5,                 // 最小空闲连接
    MaxConnAge: 30 * time.Minute,    // 连接最大生命周期
    IdleTimeout: 5 * time.Minute,    // 空闲连接超时
}
```

### ✅ 数据验证 (validator/)

#### 内置验证规则
- **邮箱**: RFC 5322 标准
- **手机号**: 中国大陆 11 位手机号
- **身份证**: 15位/18位身份证号
- **URL**: HTTP/HTTPS URL 格式
- **IP 地址**: IPv4/IPv6 地址格式

#### 自定义验证
```go
// 自定义验证器
func CustomValidator(value interface{}) bool {
    str, ok := value.(string)
    if !ok {
        return false
    }
    // 自定义验证逻辑
    return len(str) >= 6 && len(str) <= 20
}

// 注册自定义验证器
validator.RegisterValidator("custom", CustomValidator)
```

### 🛠️ 工具函数 (helper/)

#### 类型转换
```go
converter := helper.Convert{Value: "123.45"}

intVal := converter.ToInt()           // 123
floatVal := converter.ToFloat64()     // 123.45
boolVal := converter.ToBool()         // true (非空字符串)
stringVal := converter.ToString()     // "123.45"
```

#### 字符串处理
```go
str := helper.NewStr("UserProfile")

snake := str.ConvertCamelToSnake()    // "user_profile"
camel := str.ConvertSnakeToCamel()    // "UserProfile"
substr := str.Substr(0, 4)           // "User"
contains := str.Contains("Profile")   // true
```

#### JSON 操作
```go
// 结构体转 JSON
data := map[string]interface{}{"name": "张三", "age": 25}
jsonStr := helper.Json(data).ToString()

// JSON 转结构体
var result map[string]interface{}
helper.Json(jsonStr).ToStruct(&result)
```

#### HTTP 工具
```go
// URL 构建和解析
url := helper.BuildURL("https", "example.com", "/api/users", map[string]string{"page": "1", "limit": "10"})
// 结果: https://example.com/api/users?page=1&limit=10

// HTTP 请求头处理
headers := helper.ParseHTTPHeaders("Content-Type: application/json\nAuthorization: Bearer token123")
```

#### IP 地址处理
```go
// IP 地址验证
ip := "192.168.1.1"
if helper.IsIPv4(ip) {
    fmt.Printf("%s 是有效的IPv4地址\n", ip)
}

// IP 地理位置查询（需要配置）
location, err := helper.GetIPLocation("8.8.8.8")
if err == nil {
    fmt.Printf("IP位置: %+v\n", location)
}
```

#### SSH 密钥管理
```go
// 获取 SSH 公钥
publicKey, err := helper.GetSSHPublicKey()
if err != nil {
    fmt.Printf("获取SSH公钥失败: %v\n", err)
} else {
    fmt.Printf("SSH公钥: %s\n", publicKey)
}

// 生成 SSH 密钥对
privateKey, publicKey, err := helper.GenerateSSHKeyPair()
if err != nil {
    fmt.Printf("生成SSH密钥对失败: %v\n", err)
} else {
    fmt.Printf("私钥: %s\n公钥: %s\n", privateKey, publicKey)
}
```

#### 单位转换
```go
// 长度单位转换
meters := 1000.0
kilometers := helper.MeterToKilometer(meters)
fmt.Printf("%.2f 米 = %.2f 千米\n", meters, kilometers)

// 重量单位转换
kilograms := 1.5
pounds := helper.KilogramToPound(kilograms)
fmt.Printf("%.2f 千克 = %.2f 磅\n", kilograms, pounds)

// 时间单位转换
seconds := 3661.0
hours := helper.SecondToHour(seconds)
fmt.Printf("%.2f 秒 = %.2f 小时\n", seconds, hours)
```

### 🔄 升级工具 (upgrade/)

#### Git 代码升级
- **默认模式**: 安全升级，保留本地修改
- **强制模式**: 强制覆盖本地修改
- **回滚支持**: 支持版本回滚
- **备份功能**: 自动备份当前版本

```go
upgrade := upgrade.Upgrade{
    RepoURL: "https://github.com/user/repo.git",
    Branch:  "main",
    Mode:    "default", // 或 "hard"
}

err := upgrade.Execute()
if err != nil {
    // 升级失败，尝试回滚
    upgrade.Rollback()
}
```

### 🔒 TLS 配置 (tlsconfig/)

#### TLS 功能
- **证书生成**: 自签名证书生成
- **证书验证**: 证书有效性验证
- **mTLS 支持**: 双向 TLS 认证
- **动态加载**: 热加载证书更新

```go
config := tlsconfig.Config{
    CertFile: "/path/to/cert.pem",
    KeyFile:  "/path/to/key.pem",
    CAFile:   "/path/to/ca.pem", // 可选，用于 mTLS
}

tlsConfig, err := config.LoadTLSConfig()
if err != nil {
    panic(err)
}
```

### 🔗 中间件支持 (middleware/)

#### 功能特性
- **跨域支持 (CORS)**: 完整的 CORS 配置，支持预检请求和凭证
- **真实 IP 获取**: 正确处理代理服务器后的真实客户端 IP
- **请求参数解析 (SetGPC)**: 自动解析请求参数到 Gin 上下文

#### 支持的 Content-Type
- **application/json**: JSON 格式数据解析
- **application/x-www-form-urlencoded**: 表单数据解析
- **multipart/form-data**: 文件上传表单解析
- **text/xml**: XML 格式数据解析（支持微信开放平台等场景）
- **application/xml**: XML 格式数据解析

#### 使用示例
```go
import "github.com/jcbowen/jcbaseGo/middleware"

func main() {
    r := gin.Default()
    
    // 使用中间件
    base := middleware.Base{}
    r.Use(base.Cors())      // 跨域支持
    r.Use(base.RealIP())    // 真实 IP 获取
    r.Use(base.SetGPC())    // 请求参数解析
    
    // 路由定义
    r.POST("/api/wechat", func(c *gin.Context) {
        // 自动解析 XML 请求体（微信开放平台格式）
        formData := c.MustGet("formData").(map[string]any)
        
        // 获取微信消息参数
        toUserName := formData["ToUserName"].(string)
        fromUserName := formData["FromUserName"].(string)
        msgType := formData["MsgType"].(string)
        
        // 处理微信消息...
        c.XML(200, gin.H{
            "ToUserName": fromUserName,
            "FromUserName": toUserName,
            "CreateTime": time.Now().Unix(),
            "MsgType": "text",
            "Content": "收到消息",
        })
    })
    
    r.Run(":8080")
}
```

### 🐘 PHP 集成 (php/)

#### PHP 解释器特性
- **内嵌解释器**: 无需外部 PHP 环境
- **混合开发**: Go 和 PHP 代码混合执行
- **性能优化**: 复用解释器实例
- **错误处理**: 完整的错误捕获机制

## 🎨 设计模式和架构

### Trait 模式
```go
// Trait 提供可复用的行为
type CRUDTrait struct {
    Model interface{}
    DB    *gorm.DB
}

// 控制器组合 Trait
type UserController struct {
    CRUDTrait
}

// 自动获得 CRUD 方法，也可以覆盖
func (uc *UserController) Create() { /* 自定义逻辑 */ }
```

### 接口抽象
```go
// 数据库接口抽象
type Instance interface {
    GetDb() *gorm.DB
}

// 存储接口抽象
type StorageInterface interface {
    Upload(filename string, data []byte) (FileInfo, error)
    Download(filename string) ([]byte, error)
    Delete(filename string) error
}
```

### 配置驱动
```go
// 支持多种配置源
type Config struct {
    Source string // "json", "ini", "env", "yaml"
    Path   string
}
```

## 🚀 性能优化和最佳实践

### 数据库优化
```go
// 连接池配置
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(100)           // 最大连接数
sqlDB.SetMaxIdleConns(10)            // 最大空闲连接
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期

// 批量操作
db.CreateInBatches(users, 1000)      // 批量插入
db.Model(&User{}).Where("status = ?", 0).Update("status", 1)  // 批量更新

// 预加载优化
db.Preload("Profile").Preload("Orders").Find(&users)  // 避免 N+1 查询
```

### 缓存策略
```go
// 分层缓存
func GetUser(id uint) (*User, error) {
    // 1. 检查内存缓存
    if user, exists := memoryCache.Get(fmt.Sprintf("user:%d", id)); exists {
        return user.(*User), nil
    }
    
    // 2. 检查 Redis 缓存
    if userData, err := redis.Get(ctx, fmt.Sprintf("user:%d", id)).Result(); err == nil {
        var user User
        json.Unmarshal([]byte(userData), &user)
        memoryCache.Set(fmt.Sprintf("user:%d", id), &user, 5*time.Minute)
        return &user, nil
    }
    
    // 3. 查询数据库
    var user User
    if err := db.First(&user, id).Error; err != nil {
        return nil, err
    }
    
    // 4. 写入缓存
    userData, _ := json.Marshal(user)
    redis.Set(ctx, fmt.Sprintf("user:%d", id), userData, 30*time.Minute)
    memoryCache.Set(fmt.Sprintf("user:%d", id), &user, 5*time.Minute)
    
    return &user, nil
}
```

### 安全最佳实践
```go
// 1. 输入验证
func ValidateUserInput(data map[string]interface{}) error {
    rules := map[string][]string{
        "username": {"required", "min:3", "max:20", "alphanum"},
        "email":    {"required", "email"},
        "password": {"required", "min:8"},
    }
    return validator.Validate(data, rules)
}

// 2. SQL 注入防护 (GORM 自动处理)
db.Where("username = ? AND status = ?", username, 1).First(&user)

// 3. XSS 防护
func SanitizeHTML(input string) string {
    return html.EscapeString(input)
}

// 4. 敏感数据加密
func EncryptSensitiveData(data string) (string, error) {
    sm4 := security.SM4{
        Text: data,
        Key:  os.Getenv("ENCRYPTION_KEY"),
        Mode: "GCM",
    }
    
    var encrypted string
    err := sm4.Encrypt(&encrypted)
    return encrypted, err
}
```

## 📖 详细示例

查看 [example/](example/) 目录获取更多示例：

### 运行示例
```bash
# 安全组件示例
go run example/security/sm4/main.go
go run example/security/aes/main.go

# 数据库示例
go run example/orm/mysql/main.go
go run example/orm/sqlite/main.go

# 邮件发送示例
go run example/mailer/main.go

# Redis 缓存示例
go run example/redis/main.go

# 附件管理示例
go run example/attachment/upload/main.go

# PHP 集成示例
go run example/php/basic/main.go

# 工具函数示例
go run example/helper/convert/main.go
go run example/helper/string/main.go

# 数据验证示例
go run example/validator/main.go
```

### 完整应用示例
查看 [example/README.md](example/README.md) 获取完整的 Web 应用程序示例。
创建 `test.env` 文件：
```bash
# 数据库测试配置
TEST_DB_HOST=localhost
TEST_DB_PORT=3306
TEST_DB_USER=root
TEST_DB_PASSWORD=password
TEST_DB_NAME=test_jcbase

# Redis 测试配置
TEST_REDIS_HOST=localhost
TEST_REDIS_PORT=6379
TEST_REDIS_PASSWORD=

# 邮件测试配置
TEST_SMTP_HOST=smtp.qq.com
TEST_SMTP_PORT=587
TEST_SMTP_USER=test@qq.com
TEST_SMTP_PASS=test_password
```

## 🔧 故障排除

### 常见问题

#### 1. 数据库连接失败
```bash
# 检查配置
export DEBUG=true
go run your_app.go

# 常见错误和解决方案
# Error: "dial tcp: connect: connection refused"
# 解决: 检查数据库服务是否启动，端口是否正确

# Error: "Access denied for user"
# 解决: 检查用户名、密码和权限配置
```

#### 2. Redis 连接问题
```bash
# 检查 Redis 服务状态
redis-cli ping

# 检查配置
redis-cli -h localhost -p 6379 -a your_password ping
```

#### 3. 邮件发送失败
```go
// 启用调试模式
mailer := mailer.Mailer{
    Host:     "smtp.qq.com",
    Port:     587,
    Username: "your-email@qq.com",
    Password: "your-auth-code",
    Debug:    true, // 启用调试
}
```

#### 4. 加密解密失败
```go
// 检查密钥和 IV 长度
// SM4: 密钥 16 字节，IV 16 字节
// AES: 密钥 16/24/32 字节，IV 16 字节

// 检查模式匹配
// 加密和解密必须使用相同的模式和参数
```

### 性能问题诊断

#### 1. 数据库性能
```go
// 启用 SQL 日志
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

// 监控慢查询
db.Logger = db.Logger.LogMode(logger.Warn)
```

#### 2. 内存使用监控
```go
import (
    "runtime"
    "time"
)

func MonitorMemory() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        fmt.Printf("内存使用: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d\n",
            m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
    }
}
```

## 🤝 贡献指南

### 开发环境设置
```bash
# 1. 克隆仓库
git clone https://github.com/jcbowen/jcbaseGo.git
cd jcbaseGo

# 2. 安装依赖
go mod download

# 3. 运行测试（全部用例均在本地完成，不依赖外部服务器）
go test ./...

# 4. 运行示例
go run example/security/sm4/main.go
```

### 提交规范
```bash
# 提交消息格式
type(scope): description

# 类型说明
feat:     新功能
fix:      Bug 修复
docs:     文档更新
style:    代码格式化
refactor: 代码重构
test:     测试相关
chore:    构建过程或辅助工具变动

# 示例
feat(security): 添加 SM4 GCM 模式支持
fix(orm): 修复软删除查询条件问题
docs(README): 更新 CRUD 使用文档
```

### 代码规范
- **注释**: 必须使用简体中文注释
- **命名**: 遵循 Go 语言命名规范
- **格式**: 使用 `gofmt` 和 `goimports` 格式化
- **测试**: 新功能必须包含测试用例
- **文档**: 更新相关文档和示例

### Pull Request 流程
1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request




## 📄 许可证

本项目采用 **MIT 许可证** - 查看 [LICENSE](LICENSE) 文件了解详情。

### MIT 许可证要点
- ✅ 商业使用
- ✅ 修改
- ✅ 分发
- ✅ 私人使用
- ❌ 不提供担保
- ❌ 不承担责任

## 🙏 致谢

感谢以下开源项目和贡献者为本项目提供的支持和灵感：

### 核心依赖
- [GORM](https://gorm.io) - 优秀的 Go ORM 库，提供强大的数据库操作能力
- [Gin](https://gin-gonic.com) - 高性能的 Go Web 框架，支撑 HTTP 调试工具
- [Redis](https://redis.io) - 内存数据结构存储，提供缓存功能支持
- [Go-Redis](https://github.com/redis/go-redis) - Redis 客户端库
- [Golang SM4](https://github.com/tjfoc/gmsm) - 国密 SM4 算法实现

### 工具和组件
- [Cobra](https://github.com/spf13/cobra) - 命令行工具框架
- [Viper](https://github.com/spf13/viper) - 配置管理工具
- [Testify](https://github.com/stretchr/testify) - 测试工具集
- [Go-Mail](https://github.com/go-mail/mail) - 邮件发送功能

### 特别感谢
- 所有为本项目提交 Issue、PR 和提供建议的开发者
- 使用本项目的用户和社区成员
- 开源社区的持续支持和贡献

## 🌟 支持项目

如果这个项目对您有帮助，请：

1. ⭐ 给项目一个 Star
2. 🐛 报告 Bug 或提出建议
3. 📖 完善文档和示例
4. 💻 贡献代码
5. 📢 推荐给其他开发者

## 📞 联系方式

- **项目主页**: [https://github.com/jcbowen/jcbaseGo](https://github.com/jcbowen/jcbaseGo)
- **问题反馈**: [Issues](https://github.com/jcbowen/jcbaseGo/issues)
- **功能请求**: [Discussions](https://github.com/jcbowen/jcbaseGo/discussions)
- **技术交流**: 欢迎提交 Issue 或 PR

---

<div align="center">

⭐ **如果这个项目对你有帮助，请给它一个星标！** ⭐

![GitHub stars](https://img.shields.io/github/stars/jcbowen/jcbaseGo?style=social)
![GitHub forks](https://img.shields.io/github/forks/jcbowen/jcbaseGo?style=social)
![GitHub issues](https://img.shields.io/github/issues/jcbowen/jcbaseGo)
![GitHub license](https://img.shields.io/github/license/jcbowen/jcbaseGo)

**让 Go 开发更加简单高效！**

</div>
