# jcbaseGo

一个功能丰富的 Go 语言基础工具包，提供多种可复用的组件和工具函数，帮助开发者快速构建高质量的 Go 应用程序。

## 🚀 功能特性

- **🔐 安全组件**: SM4/AES 加密、密码处理、数据验证
- **🗄️ 数据库 ORM**: MySQL/SQLite 支持，基于 GORM 的抽象层，包含基础模型和 CRUD trait
- **📧 邮件服务**: 支持 SMTP 邮件发送，HTML/文本邮件，附件支持
- **📁 附件管理**: 本地、FTP、SFTP、OSS、COS 等多种存储方式
- **🛠️ 工具函数**: 类型转换、字符串处理、JSON 操作、文件处理等
- **💾 缓存支持**: Redis 缓存组件，连接池优化
- **✅ 数据验证**: 邮箱、手机号、身份证、URL、IP 等验证
- **🎛️ 配置管理**: 支持 JSON、INI、命令行等多种配置源
- **🔄 升级工具**: Git 代码自动升级，支持回滚和备份
- **🔗 TLS 配置**: 完整的 TLS/SSL 证书管理
- **🐘 PHP 集成**: 内置 PHP 解释器，支持混合开发

## 📦 安装

### 基础安装

```bash
go get github.com/jcbowen/jcbaseGo
```

### 依赖要求

- **Go**: 1.21.0+ (推荐 1.23.0+)
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
│   ├── helper/                 # 🛠️ 工具函数集合
│   │   ├── convert.go          # 类型转换工具
│   │   ├── file.go             # 文件操作工具
│   │   ├── json.go             # JSON 处理工具
│   │   ├── money.go            # 金额处理工具
│   │   ├── ssh.go              # SSH 连接工具
│   │   ├── string.go           # 字符串处理工具
│   │   └── util.go             # 通用工具函数
│   ├── mailer/                 # 📧 邮件发送组件
│   │   └── mailer.go           # SMTP 邮件服务
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
│   ├── redis/                  # 💾 Redis 缓存组件
│   │   ├── cache.go            # 缓存操作实现
│   │   └── main.go             # Redis 连接管理
│   ├── security/               # 🔐 安全相关功能
│   │   ├── aes.go              # AES 加密算法
│   │   ├── base.go             # 安全基础功能
│   │   ├── password.go         # 密码哈希处理
│   │   ├── safe.go             # 安全验证工具
│   │   └── sm4.go              # SM4 国密算法
│   ├── trait/                  # 🎭 Trait 模式实现
│   │   ├── controller/         # 控制器基础功能
│   │   │   └── controller.go   # 控制器基类
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
│   ├── tlsconfig/              # 🔒 TLS 配置管理
│   │   └── tlsconfig.go        # TLS 证书配置
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
│   └── php/                    # PHP 集成示例
├── middleware/                 # 🔗 中间件集合
│   └── main.go                 # 通用中间件
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

// CreateBefore 创建前的数据验证
func (uc *UserController) CreateBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
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
func (uc *UserController) ListEach(item interface{}) interface{} {
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
    "github.com/jcbowen/jcbaseGo/component/mailer"
)

func main() {
    // 配置邮件服务
    mailConfig := mailer.Mailer{
        Host:     "smtp.qq.com",
        Port:     587,
        Username: "your-email@qq.com",
        Password: "your-smtp-password", // QQ邮箱需要使用授权码
        From:     "your-email@qq.com",
        FromName: "系统通知",
    }

    // 发送文本邮件
    err := mailConfig.Send("recipient@example.com", "测试邮件", "这是一封测试邮件")
    if err != nil {
        panic(err)
    }

    // 发送 HTML 邮件
    htmlContent := `
    <h1>欢迎注册我们的服务</h1>
    <p>感谢您的注册，请点击下面的链接激活账户：</p>
    <a href="https://example.com/activate?token=abc123">激活账户</a>
    `
    
    err = mailConfig.SendHTML("recipient@example.com", "账户激活", htmlContent)
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
        err := mailConfig.Send(recipient, "批量通知", "这是一封批量发送的邮件")
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
    "github.com/jcbowen/jcbaseGo/component/attachment"
    "github.com/jcbowen/jcbaseGo/component/attachment/remote"
)

func main() {
    // 本地文件存储
    localAttachment := attachment.Attachment{
        StorageType: "local",
        LocalPath:   "./uploads",
    }

    // 上传文件
    fileInfo, err := localAttachment.Upload("avatar.jpg", fileBytes)
    if err != nil {
        panic(err)
    }
    fmt.Printf("文件上传成功: %+v\n", fileInfo)

    // 阿里云 OSS 存储
    ossConfig := remote.OSSConfig{
        AccessKeyID:     "your-access-key-id",
        AccessKeySecret: "your-access-key-secret",
        Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
        BucketName:      "your-bucket-name",
    }

    ossAttachment := attachment.Attachment{
        StorageType:   "oss",
        RemoteConfig:  ossConfig,
    }

    fileInfo, err = ossAttachment.Upload("documents/report.pdf", fileBytes)
    if err != nil {
        panic(err)
    }

    // 腾讯云 COS 存储
    cosConfig := remote.COSConfig{
        SecretID:  "your-secret-id",
        SecretKey: "your-secret-key",
        Region:    "ap-guangzhou",
        Bucket:    "your-bucket-name",
    }

    cosAttachment := attachment.Attachment{
        StorageType:  "cos",
        RemoteConfig: cosConfig,
    }

    // 支持的文件类型检查
    allowedTypes := []string{"jpg", "jpeg", "png", "gif", "pdf", "doc", "docx"}
    if !attachment.IsAllowedFileType("image.jpg", allowedTypes) {
        fmt.Println("不支持的文件类型")
        return
    }
}
```

### 5. Redis 缓存

```go
package main

import (
    "context"
    "time"
    "github.com/jcbowen/jcbaseGo/component/redis"
)

func main() {
    // 配置 Redis 连接
    redisConfig := redis.Config{
        Host:     "localhost",
        Port:     "6379",
        Password: "", // Redis 密码
        DB:       0,  // 数据库编号
        PoolSize: 10, // 连接池大小
    }

    // 创建 Redis 客户端
    redisClient := redis.NewClient(redisConfig)
    ctx := context.Background()

    // 设置缓存
    err := redisClient.Set(ctx, "user:1001", "用户数据", 30*time.Minute).Err()
    if err != nil {
        panic(err)
    }

    // 获取缓存
    value, err := redisClient.Get(ctx, "user:1001").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("缓存值: %s\n", value)

    // 设置哈希缓存
    err = redisClient.HSet(ctx, "user:profile:1001", map[string]interface{}{
        "name":  "张三",
        "email": "zhangsan@example.com",
        "age":   25,
    }).Err()
    if err != nil {
        panic(err)
    }

    // 获取哈希缓存
    profile, err := redisClient.HGetAll(ctx, "user:profile:1001").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户资料: %+v\n", profile)

    // 列表操作
    err = redisClient.LPush(ctx, "message_queue", "消息1", "消息2", "消息3").Err()
    if err != nil {
        panic(err)
    }

    // 消费队列消息
    message, err := redisClient.RPop(ctx, "message_queue").Result()
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
    if validator.IsIDCard(idCard) {
        fmt.Printf("%s 是有效的身份证号\n", idCard)
    }

    // URL 验证
    url := "https://www.example.com"
    if validator.IsURL(url) {
        fmt.Printf("%s 是有效的URL\n", url)
    }

    // IP 地址验证
    ipv4 := "192.168.1.1"
    if validator.IsIPv4(ipv4) {
        fmt.Printf("%s 是有效的IPv4地址\n", ipv4)
    }

    ipv6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
    if validator.IsIPv6(ipv6) {
        fmt.Printf("%s 是有效的IPv6地址\n", ipv6)
    }

    // 批量验证
    data := map[string]interface{}{
        "email":  "test@example.com",
        "mobile": "13800138000",
        "age":    25,
    }

    rules := map[string][]string{
        "email":  {"required", "email"},
        "mobile": {"required", "mobile"},
        "age":    {"required", "integer", "min:18", "max:100"},
    }

    errors := validator.Validate(data, rules)
    if len(errors) > 0 {
        fmt.Printf("验证失败: %+v\n", errors)
    } else {
        fmt.Println("所有数据验证通过")
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
    camelCase := str.ConvertSnakeToCamel()    // HelloWorld
    substr := str.Substr(0, 5)               // Hello
    
    fmt.Printf("字符串处理: snake=%s, camel=%s, substr=%s\n", snakeCase, camelCase, substr)

    // JSON 处理
    data := map[string]interface{}{
        "name": "张三",
        "age":  25,
        "city": "北京",
    }
    
    jsonStr := helper.Json(data).ToString()
    fmt.Printf("JSON字符串: %s\n", jsonStr)

    // 从JSON字符串解析
    var parsedData map[string]interface{}
    helper.Json(jsonStr).ToStruct(&parsedData)
    fmt.Printf("解析后的数据: %+v\n", parsedData)

    // 金额处理 (以分为单位)
    amount := int64(12345) // 123.45 元
    money := helper.Money{Amount: amount}
    yuanStr := money.ToYuan()        // "123.45"
    formattedStr := money.Format()   // "¥123.45"
    
    fmt.Printf("金额处理: 元=%s, 格式化=%s\n", yuanStr, formattedStr)

    // 文件操作
    file := &helper.File{Path: "./test.txt"}
    
    // 写入文件
    err := file.Write("Hello, jcbaseGo!")
    if err != nil {
        fmt.Printf("写入文件失败: %v\n", err)
    }
    
    // 读取文件
    content, err := file.Read()
    if err != nil {
        fmt.Printf("读取文件失败: %v\n", err)
    } else {
        fmt.Printf("文件内容: %s\n", content)
    }

    // 检查文件是否存在
    exists, err := file.Exists()
    if err == nil && exists {
        fmt.Println("文件存在")
        
        // 获取文件信息
        size, _ := file.Size()
        fmt.Printf("文件大小: %d 字节\n", size)
    }
}
```

### 8. PHP 集成

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
    // 初始化 PHP 解释器
    phpEngine := php.NewPHP()
    defer phpEngine.Close()

    // 执行 PHP 代码
    code := `
    <?php
    $name = "jcbaseGo";
    $version = "1.0.0";
    echo "欢迎使用 " . $name . " 版本 " . $version;
    return ["name" => $name, "version" => $version];
    `
    
    result, err := phpEngine.Exec(code)
    if err != nil {
        panic(err)
    }
    fmt.Printf("PHP 执行结果: %s\n", result)

    // 调用 PHP 函数
    mathCode := `
    <?php
    function calculate($a, $b, $operation) {
        switch($operation) {
            case 'add': return $a + $b;
            case 'subtract': return $a - $b;
            case 'multiply': return $a * $b;
            case 'divide': return $b != 0 ? $a / $b : 0;
            default: return 0;
        }
    }
    
    return calculate(10, 5, 'add');
    `
    
    result, err = phpEngine.Exec(mathCode)
    if err != nil {
        panic(err)
    }
    fmt.Printf("PHP 计算结果: %s\n", result)

    // 使用 PHP 处理数组和对象
    arrayCode := `
    <?php
    $users = [
        ["id" => 1, "name" => "张三", "email" => "zhangsan@example.com"],
        ["id" => 2, "name" => "李四", "email" => "lisi@example.com"],
        ["id" => 3, "name" => "王五", "email" => "wangwu@example.com"]
    ];
    
    // 过滤和转换数据
    $activeUsers = array_filter($users, function($user) {
        return $user['id'] > 1;
    });
    
    return json_encode($activeUsers);
    `
    
    result, err = phpEngine.Exec(arrayCode)
    if err != nil {
        panic(err)
    }
    fmt.Printf("PHP 数组处理结果: %s\n", result)
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

# 3. 运行测试
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

## 📋 更新日志

### v1.0.0 (2024-01-01)
- ✨ 初始版本发布
- 🔐 安全组件 (SM4/AES 加密)
- 🗄️ 数据库 ORM (MySQL/SQLite)
- 📧 邮件服务组件
- 📁 附件管理组件
- 🛠️ 工具函数集合
- 💾 Redis 缓存组件
- ✅ 数据验证组件

### 即将发布
- 🔄 分布式锁支持
- 📊 性能监控组件
- 🔍 全文搜索支持
- 🌐 国际化支持
- 📱 移动端 API 适配

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

感谢以下开源项目和贡献者：

- [GORM](https://gorm.io) - 优秀的 Go ORM 库
- [Gin](https://gin-gonic.com) - 高性能的 Go Web 框架
- [Redis](https://redis.io) - 内存数据结构存储
- 所有为本项目做出贡献的开发者

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
