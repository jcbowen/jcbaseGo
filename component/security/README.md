# Security 组件

安全组件提供全面的安全功能，包括数据加密、密码处理、输入过滤和国密算法支持。

## 概述

Security 组件是一个功能丰富的安全工具库，为 Go 应用程序提供企业级的安全保障。组件包含 AES 加密、SM4 国密算法、密码哈希、输入过滤等多种安全功能。

## 功能特性

- **AES 加密/解密**：支持 CBC 模式的 AES 加密，自动处理密钥长度验证和 PKCS7 填充
- **SM4 国密算法**：支持 SM4 对称加密，提供 CBC 和 GCM 两种加密模式
- **密码处理**：基于 bcrypt 的密码哈希和验证功能
- **输入过滤**：智能输入清理，防止 SQL 注入和 XSS 攻击
- **安全过滤**：多层次的字符串清理和类型安全处理
- **基础加密**：提供 PBKDF2 和 HKDF 密钥派生功能

## 快速开始

### 安装

```go
go get github.com/jcbowen/jcbaseGo/component/security
```

### AES 加密示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    // 创建 AES 实例
    aes := security.AES{
        Text: "Hello, World!",
        Key:  "16bytekey12345678", // 16/24/32 字节密钥
        Iv:   "16byteiv12345678",  // 16 字节 IV
    }

    // 加密
    var encrypted string
    err := aes.Encrypt(&encrypted)
    if err != nil {
        panic(err)
    }
    fmt.Println("加密结果:", encrypted)

    // 解密
    var decrypted string
    aes.Text = encrypted
    err = aes.Decrypt(&decrypted)
    if err != nil {
        panic(err)
    }
    fmt.Println("解密结果:", decrypted)
}
```

### SM4 加密示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    // 创建 SM4 实例
    sm4 := security.SM4{
        Text:     "Hello, SM4!",
        Key:      "16bytekey12345678", // 16 字节密钥
        Iv:       "16byteiv12345678",  // 16 字节 IV
        Mode:     "CBC",               // CBC 或 GCM 模式
        Encoding: "Std",               // 输出编码格式
    }

    // 加密
    var encrypted string
    err := sm4.Encrypt(&encrypted)
    if err != nil {
        panic(err)
    }
    fmt.Println("SM4 加密结果:", encrypted)

    // 解密
    var decrypted string
    sm4.Text = encrypted
    err = sm4.Decrypt(&decrypted)
    if err != nil {
        panic(err)
    }
    fmt.Println("SM4 解密结果:", decrypted)
}
```

### 密码处理示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    password := "mySecurePassword123"

    // 生成密码哈希
    hash, err := security.PasswordHash(password, 12) // cost 参数建议 10-14
    if err != nil {
        panic(err)
    }
    fmt.Println("密码哈希:", hash)

    // 验证密码
    isValid := security.PasswordVerify(password, hash)
    fmt.Println("密码验证结果:", isValid)
}
```

### 输入过滤示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    // 清理字符串输入
    input := security.Input{
        Value:        "<script>alert('xss')</script>SELECT * FROM users",
        DefaultValue: "default value",
    }

    // 通用清理
    sanitized := input.Sanitize()
    fmt.Println("通用清理结果:", sanitized)

    // 特定类型清理
    sanitizedStr := input.SanitizeString(input.Value.(string), []string{"sql", "xss"})
    fmt.Println("SQL+XSS 清理结果:", sanitizedStr)

    // HTML 内容清理
    htmlContent := input.Html()
    fmt.Println("HTML 清理结果:", htmlContent)
}
```

## 详细配置

### AES 配置

AES 结构体支持以下配置：

```go
type AES struct {
    Text string `json:"text" default:""`                // 待加密/解密的文本
    Key  string `json:"key" default:"jcbase.aes_key__"` // 加密密钥（16/24/32 字节）
    Iv   string `json:"iv" default:"jcbase.aes_iv___"`   // 初始化向量（16 字节）
}
```

### SM4 配置

SM4 结构体支持以下配置：

```go
type SM4 struct {
    Text     string `json:"text" default:""`                // 待加密/解密的文本
    Key      string `json:"key" default:"jcbase.sm4_key__"` // 加密密钥（16 字节）
    Iv       string `json:"iv" default:"jcbase.sm4_iv___"`  // 初始化向量（16 字节）
    Mode     string `json:"mode" default:"CBC"`             // 加密模式：CBC/GCM
    Encoding string `json:"encoding" default:"Std"`         // 编码格式：Std/Raw/RawURL/Hex
}
```

### 输入过滤配置

Input 结构体提供灵活的输入清理：

```go
type Input struct {
    Value        interface{} // 输入值
    DefaultValue interface{} // 默认值
}
```

支持以下清理类型：
- `"badStr"`：清理潜在有害字符串
- `"htmlEntity"`：处理 HTML 实体
- `"sql"`：防止 SQL 注入
- `"xss"`：防止 XSS 攻击

## 高级功能

### 自定义加密配置

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    // 自定义加密配置
    config := security.CipherConfig{
        Cipher: "AES-256-CBC",
        AllowedCiphers: map[string][]int{
            "AES-256-CBC": {16, 32},
        },
        KdfHash:              "sha256",
        MacHash:              "sha256",
        AuthKeyInfo:          "MyAppAuth",
        DerivationIterations: 100000,
    }

    data := "敏感数据"
    secret := "mySecretKey"

    // 基于密码的加密
    encrypted, err := security.Encrypt(data, secret, true, config)
    if err != nil {
        panic(err)
    }
    fmt.Println("加密结果:", encrypted)

    // 解密
    decrypted, err := security.Decrypt(encrypted, secret, true, config)
    if err != nil {
        panic(err)
    }
    fmt.Println("解密结果:", decrypted)
}
```

### 安全输入验证

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
    input := security.Input{
        Value: "admin",
        DefaultValue: "guest",
    }

    // 检查值是否在允许列表中
    allowedValues := []string{"admin", "user", "guest"}
    result := input.Belong(allowedValues, false) // strict=false 进行宽松匹配
    fmt.Println("验证结果:", result)

    // 严格匹配
    resultStrict := input.Belong(allowedValues, true) // strict=true 进行严格匹配
    fmt.Println("严格验证结果:", resultStrict)
}
```

## 性能优化建议

1. **密钥管理**：对于生产环境，建议使用密钥管理系统而非硬编码密钥
2. **密码强度**：bcrypt 的 cost 参数建议设置为 10-14，平衡安全性和性能
3. **加密模式**：GCM 模式提供认证加密，比 CBC 模式更安全
4. **输入过滤**：根据实际需求选择清理类型，避免不必要的性能开销

## 安全考虑

- 组件自动验证密钥和 IV 的长度要求
- 提供默认的安全配置，防止常见安全漏洞
- 支持国密算法 SM4，满足国内安全标准
- 输入过滤功能可有效防止 SQL 注入和 XSS 攻击

## 故障排除

### 常见错误

1. **密钥长度错误**：确保 AES 密钥为 16/24/32 字节，SM4 密钥为 16 字节
2. **IV 长度错误**：确保 IV 为 16 字节
3. **填充错误**：组件自动处理 PKCS7 填充，无需手动干预

### 调试建议

启用详细日志记录以排查加密/解密问题：

```go
// 在调试模式下启用详细输出
aes := security.AES{
    Text: "test",
    Key:  "16bytekey12345678",
    Iv:   "16byteiv12345678",
}
// 检查密钥和 IV 长度
fmt.Printf("Key length: %d\n", len(aes.Key))
fmt.Printf("IV length: %d\n", len(aes.Iv))
```

## API 参考

### 主要结构体

#### AES
提供 AES 加密/解密功能

**字段：**
- `Text string` - 待加密/解密的文本
- `Key string` - 加密密钥（16/24/32 字节）
- `Iv string` - 初始化向量（16 字节）

**方法：**
- `Encrypt(cipherText *string) error` - 加密文本
- `Decrypt(plaintext *string) error` - 解密文本

#### SM4
提供 SM4 国密算法加密/解密功能

**字段：**
- `Text string` - 待加密/解密的文本
- `Key string` - 加密密钥（16 字节）
- `Iv string` - 初始化向量（16 字节）
- `Mode string` - 加密模式（CBC/GCM）
- `Encoding string` - 编码格式（Std/Raw/RawURL/Hex）

**方法：**
- `Encrypt(cipherText *string) error` - 加密文本
- `Decrypt(plaintext *string) error` - 解密文本
- `EncryptCBC(cipherText *string) error` - CBC 模式加密
- `DecryptCBC(plaintext *string) error` - CBC 模式解密
- `EncryptGCM(cipherText *string) error` - GCM 模式加密
- `DecryptGCM(plaintext *string) error` - GCM 模式解密

#### Input
提供输入过滤和验证功能

**字段：**
- `Value interface{}` - 输入值
- `DefaultValue interface{}` - 默认值

**方法：**
- `Sanitize() interface{}` - 通用清理
- `SanitizeString(str string, args ...any) string` - 字符串清理
- `Belong(allow interface{}, strict bool) interface{}` - 值验证
- `Html() string` - HTML 内容清理

### 工具函数

- `PasswordHash(password string, cost int) (string, error)` - 密码哈希
- `PasswordVerify(password, hash string) bool` - 密码验证
- `Encrypt(data, secret string, passwordBased bool, config CipherConfig) (string, error)` - 通用加密
- `Decrypt(data, secret string, passwordBased bool, config CipherConfig) (string, error)` - 通用解密

## 版本历史

- v1.0.0：初始版本，包含 AES、SM4、密码处理和输入过滤功能

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Security 组件。

## 许可证

Security 组件遵循 MIT 许可证。