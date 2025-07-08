# jcbaseGo

一个功能丰富的 Go 语言基础工具包，提供多种可复用的组件和工具函数，帮助开发者快速构建高质量的 Go 应用程序。

## 🚀 功能特性

- **🔐 安全组件**: SM4/AES 加密、密码处理、数据验证
- **🗄️ 数据库 ORM**: MySQL/SQLite 支持，基于 GORM 的抽象层
- **📧 邮件服务**: 支持 SMTP 邮件发送
- **📁 附件管理**: 本地、FTP、SFTP、OSS、COS 等多种存储方式
- **🛠️ 工具函数**: 类型转换、字符串处理、JSON 操作等
- **💾 缓存支持**: Redis 缓存组件
- **✅ 数据验证**: 邮箱、手机号、身份证、URL 等验证
- **🎛️ 配置管理**: 支持 JSON、INI、命令行等多种配置源

## 📦 安装

```bash
go get github.com/jcbowen/jcbaseGo
```

## 🏗️ 项目结构

```
jcbaseGo/
├── component/                   # 核心组件目录
│   ├── attachment/             # 附件管理组件
│   │   ├── attachment.go       # 附件管理主文件
│   │   ├── method.go           # 附件操作方法
│   │   └── remote/             # 远程存储实现
│   │       ├── cos.go          # 腾讯云 COS
│   │       ├── ftp.go          # FTP 存储
│   │       ├── oss.go          # 阿里云 OSS
│   │       ├── sftp.go         # SFTP 存储
│   │       └── remote.go       # 远程存储接口
│   ├── helper/                 # 工具函数集合
│   │   ├── convert.go          # 类型转换工具
│   │   ├── file.go             # 文件操作工具
│   │   ├── json.go             # JSON 处理工具
│   │   ├── money.go            # 金额处理工具
│   │   ├── ssh.go              # SSH 工具
│   │   ├── string.go           # 字符串处理工具
│   │   └── util.go             # 通用工具函数
│   ├── mailer/                 # 邮件发送组件
│   │   └── mailer.go
│   ├── orm/                    # 数据库 ORM 抽象层
│   │   ├── instance.go         # 数据库实例接口
│   │   ├── mysql/              # MySQL 实现
│   │   └── sqlite/             # SQLite 实现
│   ├── redis/                  # Redis 缓存组件
│   │   ├── cache.go
│   │   └── main.go
│   ├── security/               # 安全相关功能
│   │   ├── aes.go              # AES 加密
│   │   ├── base.go             # 安全基础功能
│   │   ├── password.go         # 密码处理
│   │   ├── safe.go             # 安全工具
│   │   └── sm4.go              # SM4 加密
│   ├── trait/                  # Trait 模式实现
│   │   ├── controller/         # 控制器基础功能
│   │   └── crud/               # CRUD 操作模板
│   ├── tlsconfig/              # TLS 配置
│   │   └── tlsconfig.go
│   ├── upgrade/                # 升级工具
│   │   └── main.go
│   └── validator/              # 数据验证组件
│       └── main.go
├── config.go                   # 配置管理
├── type.go                     # 全局类型定义
├── errcode/                    # 错误码定义
│   └── errcode.go
├── example/                    # 使用示例
│   ├── README.md               # 示例说明文档
│   ├── security/               # 安全组件示例
│   ├── helper/                 # 工具函数示例
│   ├── orm/                    # 数据库示例
│   ├── mailer/                 # 邮件示例
│   ├── redis/                  # Redis 示例
│   ├── validator/              # 验证器示例
│   └── attachment/             # 附件管理示例
├── middleware/                 # 中间件
│   └── main.go
├── go.mod                      # Go 模块文件
└── LICENSE                     # 许可证文件
```

## 🎯 快速开始

### 1. 安全加密

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // SM4 加密
    sm4 := security.SM4{
        Text: "Hello, SM4!",
        Key:  "1234567890123456",
        Iv:   "abcdefghijklmnop",
        Mode: "CBC",
    }

    var cipherText string
    err := sm4.Encrypt(&cipherText)
    if err != nil {
        panic(err)
    }
    fmt.Printf("加密结果: %s\n", cipherText)

    // AES 加密
    aes := security.AES{
        Text: "Hello, AES!",
        Key:  "1234567890123456",
        Iv:   "abcdefghijklmnop",
    }

    err = aes.Encrypt(&cipherText)
    if err != nil {
        panic(err)
    }
    fmt.Printf("AES 加密结果: %s\n", cipherText)
}
```

### 2. 数据库操作

```go
package main

import (
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/orm/mysql"
)

func main() {
    // 配置数据库连接
    config := jcbaseGo.DbStruct{
        Host:     "localhost",
        Port:     "3306",
        Username: "root",
        Password: "password",
        Dbname:   "test_db",
        Charset:  "utf8mb4",
    }

    // 连接数据库
    db := mysql.New(config)
    gormDB := db.GetDb()

    // 使用 GORM 进行数据库操作
    // ... 数据库操作代码
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
    mailer := mailer.Mailer{
        Host:     "smtp.example.com",
        Port:     587,
        Username: "your-email@example.com",
        Password: "your-password",
    }

    // 发送邮件
    err := mailer.Send("recipient@example.com", "测试邮件", "这是一封测试邮件")
    if err != nil {
        panic(err)
    }
}
```

## 📚 组件说明

### 🔐 安全组件 (security/)

- **SM4**: 国密 SM4 对称加密算法，支持 CBC 和 GCM 模式
- **AES**: AES 对称加密算法，支持 16/24/32 字节密钥
- **密码处理**: 密码哈希、验证等安全功能

### 🗄️ 数据库 ORM (orm/)

- **MySQL**: 基于 GORM 的 MySQL 数据库操作
- **SQLite**: SQLite 数据库支持
- **事务支持**: 完整的事务操作支持
- **连接池**: 自动连接池管理

### 📧 邮件服务 (mailer/)

- **SMTP 支持**: 标准 SMTP 协议支持
- **TLS/SSL**: 安全连接支持
- **附件支持**: 邮件附件功能
- **模板支持**: HTML 邮件模板

### 📁 附件管理 (attachment/)

- **本地存储**: 本地文件系统存储
- **FTP**: FTP 服务器存储
- **SFTP**: SFTP 安全文件传输
- **OSS**: 阿里云对象存储
- **COS**: 腾讯云对象存储

### 🛠️ 工具函数 (helper/)

- **类型转换**: 各种数据类型之间的转换
- **字符串处理**: 字符串截取、替换、分割等
- **JSON 操作**: JSON 序列化和反序列化
- **文件操作**: 文件读写、目录操作
- **金额处理**: 货币计算和格式化
- **SSH 工具**: SSH 连接和操作

### 💾 缓存支持 (redis/)

- **Redis 连接**: Redis 数据库连接管理
- **缓存操作**: 键值对存储和检索
- **过期时间**: 自动过期管理
- **连接池**: 连接池优化

### ✅ 数据验证 (validator/)

- **邮箱验证**: 标准邮箱格式验证
- **手机号验证**: 中国大陆手机号验证
- **身份证验证**: 15位和18位身份证验证
- **URL 验证**: URL 格式验证
- **IP 地址验证**: IPv4/IPv6 地址验证

## 🎨 设计模式

### Trait 模式

项目使用 Trait 模式提供可复用的行为：

```go
// CRUD Trait 提供基础的增删改查操作
type UserController struct {
    trait.CRUD
}

// 自动获得 Create、Read、Update、Delete 方法
```

### 配置驱动

支持多种配置源：

```go
// JSON 配置
config := jcbaseGo.Config{
    Source: "config.json",
}

// INI 配置
config := jcbaseGo.Config{
    Source: "config.ini",
}

// 环境变量
config := jcbaseGo.Config{
    Source: "env",
}
```

## 📖 使用示例

详细的使用示例请查看 [example/](example/) 目录：

```bash
# 运行 SM4 加密示例
go run example/security/sm4/main.go

# 运行 AES 加密示例
go run example/security/aes/main.go

# 运行数据库示例
go run example/orm/mysql/main.go

# 运行邮件发送示例
go run example/mailer/main.go
```

## 🧪 测试

运行测试：

```bash
# 运行所有测试
go test ./...

# 运行特定组件测试
go test ./component/security/ -v
go test ./component/helper/ -v
```

## 📋 依赖要求

- Go 1.23.0+
- MySQL 5.7+ (可选)
- Redis 6.0+ (可选)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范

- 使用简体中文注释
- 遵循 Go 语言编码规范
- 添加适当的测试用例
- 更新相关文档

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

## 📞 联系方式

- 项目主页: [https://github.com/jcbowen/jcbaseGo](https://github.com/jcbowen/jcbaseGo)
- 问题反馈: [Issues](https://github.com/jcbowen/jcbaseGo/issues)

---

⭐ 如果这个项目对你有帮助，请给它一个星标！
