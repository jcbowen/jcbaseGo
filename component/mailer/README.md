# Mailer 组件

邮件发送组件提供简单易用的邮件发送功能，支持 TLS 加密、内嵌图片和 HTML 邮件。

## 概述

Mailer 组件是一个功能完整的邮件发送工具，基于 Go 标准库的 SMTP 协议实现。组件支持普通连接和 TLS 加密连接，能够发送纯文本邮件、HTML 邮件以及包含内嵌图片的复杂邮件。

## 功能特性

- **SMTP 协议支持**：基于标准 SMTP 协议实现邮件发送
- **TLS 加密**：支持 TLS 加密连接，保障邮件传输安全
- **多种邮件类型**：支持纯文本邮件、HTML 邮件
- **内嵌图片**：支持在 HTML 邮件中内嵌图片
- **多收件人**：支持向多个收件人发送邮件
- **配置灵活**：提供灵活的邮件配置选项
- **错误处理**：完善的错误处理和日志记录

## 快速开始

### 安装

```go
go get github.com/jcbowen/jcbaseGo/component/mailer
```

### 基本使用示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/mailer"
)

func main() {
    // 配置邮件服务器
    config := jcbaseGo.MailerStruct{
        Host:     "smtp.example.com",
        Port:     "587",
        Username: "user@example.com",
        Password: "your_password",
        From:     "noreply@example.com",
        UseTLS:   true,
    }

    // 创建邮件实例
    email := mailer.New(config)

    // 添加收件人
    email.AddRecipient("recipient@example.com")

    // 设置邮件主题
    email.SetSubject("测试邮件主题")

    // 设置邮件正文（纯文本）
    email.SetBody("这是一封测试邮件正文。", false)

    // 发送邮件
    err := email.Send()
    if err != nil {
        fmt.Println("邮件发送失败:", err)
    } else {
        fmt.Println("邮件发送成功!")
    }
}
```

### HTML 邮件示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/mailer"
)

func main() {
    config := jcbaseGo.MailerStruct{
        Host:     "smtp.example.com",
        Port:     "587",
        Username: "user@example.com",
        Password: "your_password",
        From:     "noreply@example.com",
        UseTLS:   true,
    }

    email := mailer.New(config)
    email.AddRecipient("recipient@example.com")
    email.SetSubject("HTML 邮件测试")

    // HTML 邮件正文
    htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>HTML 邮件</title>
</head>
<body>
    <h1>欢迎使用 Mailer 组件</h1>
    <p>这是一封 HTML 格式的测试邮件。</p>
    <p>支持 <strong>粗体</strong>、<em>斜体</em> 等 HTML 标签。</p>
</body>
</html>`

    email.SetBody(htmlBody, true) // 设置为 HTML 邮件

    err := email.Send()
    if err != nil {
        fmt.Println("HTML 邮件发送失败:", err)
    } else {
        fmt.Println("HTML 邮件发送成功!")
    }
}
```

### 带内嵌图片的邮件示例

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/mailer"
    "io/ioutil"
)

func main() {
    config := jcbaseGo.MailerStruct{
        Host:     "smtp.example.com",
        Port:     "587",
        Username: "user@example.com",
        Password: "your_password",
        From:     "noreply@example.com",
        UseTLS:   true,
    }

    email := mailer.New(config)
    email.AddRecipient("recipient@example.com")
    email.SetSubject("带内嵌图片的邮件")

    // 读取图片文件并编码为 Base64
    imageData, err := ioutil.ReadFile("logo.png")
    if err != nil {
        fmt.Println("读取图片文件失败:", err)
        return
    }
    base64Image := base64.StdEncoding.EncodeToString(imageData)

    // 添加内嵌图片
    email.AddInlineImage("logo", base64Image)

    // HTML 邮件正文，引用内嵌图片
    htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>带图片的邮件</title>
</head>
<body>
    <h1>欢迎使用我们的服务</h1>
    <p>这封邮件包含内嵌图片：</p>
    <img src="cid:logo" alt="公司Logo" />
    <p>图片将直接显示在邮件中，无需下载附件。</p>
</body>
</html>`

    email.SetBody(htmlBody, true)

    err = email.Send()
    if err != nil {
        fmt.Println("带图片邮件发送失败:", err)
    } else {
        fmt.Println("带图片邮件发送成功!")
    }
}
```

## 详细配置

### MailerStruct 配置

MailerStruct 是邮件服务器的配置结构体：

```go
type MailerStruct struct {
    Host     string // SMTP 服务器地址
    Port     string // SMTP 服务器端口
    Username string // SMTP 用户名
    Password string // SMTP 密码
    From     string // 发件人地址
    UseTLS   bool   // 是否使用 TLS 加密
    CertPath string // 证书文件路径（可选）
    KeyPath  string // 私钥文件路径（可选）
    CAPath   string // CA 证书文件路径（可选）
}
```

### Email 结构体方法

Email 结构体提供以下方法：

```go
// New 创建新的邮件实例
func New(conf jcbaseGo.MailerStruct) *Email

// AddRecipient 添加收件人
func (e *Email) AddRecipient(to string)

// SetSubject 设置邮件主题
func (e *Email) SetSubject(subject string)

// SetBody 设置邮件正文
func (e *Email) SetBody(body string, isHTML bool)

// AddInlineImage 添加内嵌图片
func (e *Email) AddInlineImage(cid, base64Data string)

// Send 发送邮件
func (e *Email) Send() error
```

## 高级功能

### 自定义 TLS 配置

对于需要自定义 TLS 配置的场景，可以指定证书文件：

```go
config := jcbaseGo.MailerStruct{
    Host:     "smtp.example.com",
    Port:     "587",
    Username: "user@example.com",
    Password: "your_password",
    From:     "noreply@example.com",
    UseTLS:   true,
    CertPath: "/path/to/cert.pem",
    KeyPath:  "/path/to/key.pem",
    CAPath:   "/path/to/ca.pem",
}
```

### 多收件人支持

可以向多个收件人发送邮件：

```go
email := mailer.New(config)

// 添加多个收件人
recipients := []string{
    "user1@example.com",
    "user2@example.com", 
    "user3@example.com",
}

for _, recipient := range recipients {
    email.AddRecipient(recipient)
}
```

### 错误处理

Send 方法返回 error 类型，建议进行适当的错误处理：

```go
err := email.Send()
if err != nil {
    switch {
    case strings.Contains(err.Error(), "认证失败"):
        fmt.Println("SMTP 认证失败，请检查用户名和密码")
    case strings.Contains(err.Error(), "连接失败"):
        fmt.Println("无法连接到 SMTP 服务器")
    case strings.Contains(err.Error(), "TLS"):
        fmt.Println("TLS 连接失败")
    default:
        fmt.Printf("邮件发送失败: %v\n", err)
    }
}
```

## 性能优化建议

1. **连接复用**：对于频繁发送邮件的场景，建议实现连接池
2. **异步发送**：对于非关键邮件，可以使用 goroutine 异步发送
3. **批量发送**：对于大量邮件，建议使用专业的邮件发送服务
4. **错误重试**：实现简单的重试机制处理临时性错误

## 安全考虑

- 建议在生产环境中使用 TLS 加密连接
- 妥善保管 SMTP 服务器的认证信息
- 对于敏感信息，建议使用端到端加密
- 定期更新证书文件

## 故障排除

### 常见错误

1. **认证失败**：检查用户名和密码是否正确
2. **连接超时**：检查 SMTP 服务器地址和端口是否正确
3. **TLS 错误**：检查证书文件路径和格式是否正确
4. **邮件被拒**：检查发件人地址是否被服务器允许

### 调试建议

启用详细日志记录以排查邮件发送问题：

```go
// 在调试模式下记录详细日志
func debugSend(email *mailer.Email) error {
    fmt.Printf("SMTP 服务器: %s:%s\n", email.SMTPHost, email.SMTPPort)
    fmt.Printf("发件人: %s\n", email.From)
    fmt.Printf("收件人: %v\n", email.To)
    fmt.Printf("邮件主题: %s\n", email.Subject)
    
    err := email.Send()
    if err != nil {
        fmt.Printf("发送失败详情: %v\n", err)
    }
    return err
}
```

## API 参考

### 主要结构体

#### Email
邮件发送的主要结构体

**字段：**
- `SMTPHost string` - SMTP 服务器地址
- `SMTPPort string` - SMTP 服务器端口
- `SMTPUser string` - SMTP 用户名
- `SMTPPass string` - SMTP 密码
- `From string` - 发件人地址
- `To []string` - 收件人地址列表
- `Subject string` - 邮件主题
- `Body string` - 邮件正文
- `IsHTML bool` - 是否为 HTML 邮件
- `InlineImages []InlineImage` - 内嵌图片列表
- `UseTLS bool` - 是否使用 TLS 加密
- `CertFile string` - 证书文件路径
- `KeyFile string` - 私钥文件路径
- `CAFile string` - CA 证书文件路径

**方法：**
- `New(conf jcbaseGo.MailerStruct) *Email` - 创建邮件实例
- `AddRecipient(to string)` - 添加收件人
- `SetSubject(subject string)` - 设置邮件主题
- `SetBody(body string, isHTML bool)` - 设置邮件正文
- `AddInlineImage(cid, base64Data string)` - 添加内嵌图片
- `Send() error` - 发送邮件

#### InlineImage
内嵌图片结构体

**字段：**
- `CID string` - 内容 ID，用于在 HTML 中引用
- `Data string` - Base64 编码的图片数据

### 工具函数

- `splitBase64(encoded string) string` - 将 Base64 数据分行处理

## 版本历史

- v1.0.0：初始版本，包含基本的邮件发送功能
- 支持 TLS 加密、HTML 邮件和内嵌图片

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Mailer 组件。

## 许可证

Mailer 组件遵循 MIT 许可证。