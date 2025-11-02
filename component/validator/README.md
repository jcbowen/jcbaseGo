# Validator 组件

Validator 组件提供常用的数据验证功能，包括手机号、邮箱、URL、IP地址、端口、身份证号码等格式验证。

## 概述

Validator 组件是一个轻量级的数据验证工具库，提供多种常见数据格式的验证功能。组件采用简单易用的 API 设计，支持中国大陆手机号、电子邮件地址、URL、IP地址、端口号、身份证号码等格式的验证。

## 功能特性

- **手机号验证**：支持中国大陆手机号格式验证
- **邮箱验证**：支持标准电子邮件地址格式验证
- **URL验证**：支持 HTTP/HTTPS URL 格式验证
- **IP地址验证**：支持 IPv4 和 IPv6 地址验证
- **端口验证**：支持端口号范围验证
- **身份证验证**：支持中国大陆15位和18位身份证号码验证
- **轻量级**：无外部依赖，纯 Go 实现
- **高性能**：使用正则表达式和标准库函数优化性能

## 快速开始

### 安装

```go
go get github.com/jcbowen/jcbaseGo/component/validator
```

### 基本使用示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/validator"
)

func main() {
    // 手机号验证
    mobile := "13800138000"
    if validator.IsMobile(mobile) {
        fmt.Printf("%s 是有效的手机号\n", mobile)
    } else {
        fmt.Printf("%s 不是有效的手机号\n", mobile)
    }

    // 邮箱验证
    email := "test@example.com"
    if validator.IsEmail(email) {
        fmt.Printf("%s 是有效的邮箱地址\n", email)
    } else {
        fmt.Printf("%s 不是有效的邮箱地址\n", email)
    }

    // URL验证
    url := "https://www.example.com"
    if validator.IsURL(url) {
        fmt.Printf("%s 是有效的URL\n", url)
    } else {
        fmt.Printf("%s 不是有效的URL\n", url)
    }

    // IP地址验证
    ip := "192.168.1.1"
    isValid, ipType := validator.IsIP(ip)
    if isValid {
        switch ipType {
        case validator.IPv4:
            fmt.Printf("%s 是有效的IPv4地址\n", ip)
        case validator.IPv6:
            fmt.Printf("%s 是有效的IPv6地址\n", ip)
        }
    } else {
        fmt.Printf("%s 不是有效的IP地址\n", ip)
    }

    // 端口验证
    port := "8080"
    if validator.IsPort(port) {
        fmt.Printf("%s 是有效的端口号\n", port)
    } else {
        fmt.Printf("%s 不是有效的端口号\n", port)
    }

    // 身份证验证
    idCard := "110101199001011234"
    if validator.IsChineseIDCard(idCard) {
        fmt.Printf("%s 是有效的身份证号码\n", idCard)
    } else {
        fmt.Printf("%s 不是有效的身份证号码\n", idCard)
    }
}
```

### Web 表单验证示例

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo/component/validator"
    "net/http"
)

func main() {
    r := gin.Default()

    r.POST("/register", func(c *gin.Context) {
        var req struct {
            Mobile   string `json:"mobile" binding:"required"`
            Email    string `json:"email" binding:"required"`
            Password string `json:"password" binding:"required"`
        }

        if err := c.BindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
            return
        }

        // 验证手机号
        if !validator.IsMobile(req.Mobile) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确"})
            return
        }

        // 验证邮箱
        if !validator.IsEmail(req.Email) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
            return
        }

        // 验证密码长度
        if len(req.Password) < 6 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "密码长度不能少于6位"})
            return
        }

        // 注册逻辑...
        c.JSON(http.StatusOK, gin.H{
            "message": "注册成功",
            "data": gin.H{
                "mobile": req.Mobile,
                "email":  req.Email,
            },
        })
    })

    r.Run(":8080")
}
```

### 用户信息验证示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/validator"
)

type UserInfo struct {
    Name    string `json:"name"`
    Mobile  string `json:"mobile"`
    Email   string `json:"email"`
    IDCard  string `json:"id_card"`
    Website string `json:"website"`
}

func ValidateUserInfo(user UserInfo) (bool, []string) {
    var errors []string

    // 验证姓名
    if len(user.Name) == 0 {
        errors = append(errors, "姓名不能为空")
    }

    // 验证手机号
    if !validator.IsMobile(user.Mobile) {
        errors = append(errors, "手机号格式不正确")
    }

    // 验证邮箱
    if !validator.IsEmail(user.Email) {
        errors = append(errors, "邮箱格式不正确")
    }

    // 验证身份证
    if user.IDCard != "" && !validator.IsChineseIDCard(user.IDCard) {
        errors = append(errors, "身份证号码格式不正确")
    }

    // 验证网站
    if user.Website != "" && !validator.IsURL(user.Website) {
        errors = append(errors, "网站地址格式不正确")
    }

    return len(errors) == 0, errors
}

func main() {
    user := UserInfo{
        Name:    "张三",
        Mobile:  "13800138000",
        Email:   "zhangsan@example.com",
        IDCard:  "110101199001011234",
        Website: "https://www.example.com",
    }

    isValid, errors := ValidateUserInfo(user)
    if isValid {
        fmt.Println("用户信息验证通过")
    } else {
        fmt.Printf("用户信息验证失败，错误：%v\n", errors)
    }
}
```

### 网络配置验证示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/validator"
)

type NetworkConfig struct {
    Host string `json:"host"`
    Port string `json:"port"`
}

func ValidateNetworkConfig(config NetworkConfig) (bool, []string) {
    var errors []string

    // 验证主机地址（IP或域名）
    if config.Host == "" {
        errors = append(errors, "主机地址不能为空")
    } else {
        // 如果是IP地址，验证格式
        if isValid, ipType := validator.IsIP(config.Host); isValid {
            fmt.Printf("主机地址是有效的%s地址\n", func() string {
                switch ipType {
                case validator.IPv4:
                    return "IPv4"
                case validator.IPv6:
                    return "IPv6"
                default:
                    return "IP"
                }
            }())
        }
        // 如果不是IP地址，假设是域名（这里不进行域名验证）
    }

    // 验证端口
    if !validator.IsPort(config.Port) {
        errors = append(errors, "端口号格式不正确")
    }

    return len(errors) == 0, errors
}

func main() {
    configs := []NetworkConfig{
        {
            Host: "192.168.1.1",
            Port: "8080",
        },
        {
            Host: "localhost",
            Port: "3000",
        },
        {
            Host: "invalid-ip",
            Port: "99999", // 无效端口
        },
    }

    for i, config := range configs {
        fmt.Printf("配置 %d:\n", i+1)
        isValid, errors := ValidateNetworkConfig(config)
        if isValid {
            fmt.Println("  网络配置验证通过")
        } else {
            fmt.Printf("  网络配置验证失败，错误：%v\n", errors)
        }
        fmt.Println()
    }
}
```

## 详细功能说明

### 手机号验证 (IsMobile)

验证中国大陆手机号格式：
- 以1开头
- 第二位为3-9
- 总长度为11位
- 全部为数字

**示例：**
```go
validator.IsMobile("13800138000") // true
validator.IsMobile("12345678901") // false（第二位不是3-9）
validator.IsMobile("1380013800")  // false（长度不足）
```

### 邮箱验证 (IsEmail)

验证标准电子邮件地址格式：
- 长度在3-254个字符之间
- 包含@符号
- 域名部分至少3个字符
- 使用正则表达式验证格式

**示例：**
```go
validator.IsEmail("test@example.com")    // true
validator.IsEmail("test@example")       // false（域名不完整）
validator.IsEmail("test@.com")          // false（域名无效）
```

### URL验证 (IsURL)

验证HTTP/HTTPS URL格式：
- 包含协议头（http:// 或 https://）
- 包含主机名
- 使用标准库url.Parse验证

**示例：**
```go
validator.IsURL("https://www.example.com") // true
validator.IsURL("www.example.com")         // false（缺少协议）
validator.IsURL("ftp://example.com")       // false（不支持FTP协议）
```

### IP地址验证 (IsIP)

验证IP地址格式并返回类型：
- 支持IPv4和IPv6地址
- 返回验证结果和IP类型
- 使用标准库net.ParseIP验证

**返回值：**
- `bool`：是否有效IP地址
- `int`：IP类型（IPv4、IPv6或InvalidIP）

**示例：**
```go
isValid, ipType := validator.IsIP("192.168.1.1") // true, validator.IPv4
isValid, ipType := validator.IsIP("::1")           // true, validator.IPv6
isValid, ipType := validator.IsIP("invalid")     // false, validator.InvalidIP
```

### 端口验证 (IsPort)

验证端口号格式：
- 必须是数字
- 范围在0-65535之间

**示例：**
```go
validator.IsPort("8080")   // true
validator.IsPort("0")      // true
validator.IsPort("65535")  // true
validator.IsPort("65536")  // false（超出范围）
validator.IsPort("abc")    // false（不是数字）
```

### 身份证验证 (IsChineseIDCard)

验证中国大陆身份证号码：
- 支持15位和18位格式
- 验证区域码（省份代码）
- 验证生日格式
- 验证18位身份证的校验码

**验证规则：**
1. **长度验证**：必须是15位或18位
2. **格式验证**：15位全数字，18位前17位数字+最后1位数字或X
3. **区域码验证**：前2位必须在11-91之间
4. **生日验证**：生日部分必须是有效日期
5. **校验码验证**：18位身份证需要验证校验码

**示例：**
```go
validator.IsChineseIDCard("110101199001011234") // true（18位）
validator.IsChineseIDCard("110101900101123")   // true（15位）
validator.IsChineseIDCard("123456789012345")   // false（区域码无效）
```

## 高级用法

### 自定义验证组合

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/validator"
    "strings"
)

// 验证用户名格式
func IsValidUsername(username string) bool {
    if len(username) < 3 || len(username) > 20 {
        return false
    }
    
    // 只能包含字母、数字、下划线
    for _, char := range username {
        if !(char >= 'a' && char <= 'z') && 
           !(char >= 'A' && char <= 'Z') && 
           !(char >= '0' && char <= '9') && 
           char != '_' {
            return false
        }
    }
    
    return true
}

// 验证密码强度
func IsStrongPassword(password string) bool {
    if len(password) < 8 {
        return false
    }
    
    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false
    
    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasDigit = true
        case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
            hasSpecial = true
        }
    }
    
    return hasUpper && hasLower && hasDigit && hasSpecial
}

func main() {
    // 组合验证
    tests := []struct {
        username string
        password string
        mobile   string
        email    string
    }{
        {"user123", "Password123!", "13800138000", "user@example.com"},
        {"ab", "weak", "12345678901", "invalid-email"},
    }
    
    for i, test := range tests {
        fmt.Printf("测试 %d:\n", i+1)
        
        validations := []struct {
            name string
            valid bool
        }{
            {"用户名", IsValidUsername(test.username)},
            {"密码", IsStrongPassword(test.password)},
            {"手机号", validator.IsMobile(test.mobile)},
            {"邮箱", validator.IsEmail(test.email)},
        }
        
        allValid := true
        for _, v := range validations {
            status := "通过"
            if !v.valid {
                status = "失败"
                allValid = false
            }
            fmt.Printf("  %s验证: %s\n", v.name, status)
        }
        
        if allValid {
            fmt.Println("  所有验证通过")
        } else {
            fmt.Println("  存在验证失败项")
        }
        fmt.Println()
    }
}
```

### 批量验证工具

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/validator"
)

type ValidationResult struct {
    Field   string `json:"field"`
    Value   string `json:"value"`
    IsValid bool   `json:"is_valid"`
    Message string `json:"message"`
}

func BatchValidate(data map[string]string) []ValidationResult {
    var results []ValidationResult
    
    for field, value := range data {
        var isValid bool
        var message string
        
        switch field {
        case "mobile":
            isValid = validator.IsMobile(value)
            message = "手机号格式"
        case "email":
            isValid = validator.IsEmail(value)
            message = "邮箱格式"
        case "url":
            isValid = validator.IsURL(value)
            message = "URL格式"
        case "id_card":
            isValid = validator.IsChineseIDCard(value)
            message = "身份证格式"
        default:
            isValid = true
            message = "未知字段"
        }
        
        results = append(results, ValidationResult{
            Field:   field,
            Value:   value,
            IsValid: isValid,
            Message: message,
        })
    }
    
    return results
}

func main() {
    testData := map[string]string{
        "mobile":  "13800138000",
        "email":   "test@example.com",
        "url":     "https://www.example.com",
        "id_card": "110101199001011234",
        "unknown": "some_value",
    }
    
    results := BatchValidate(testData)
    
    fmt.Println("批量验证结果:")
    for _, result := range results {
        status := "✓ 通过"
        if !result.IsValid {
            status = "✗ 失败"
        }
        fmt.Printf("  %s (%s): %s - %s\n", 
            result.Field, result.Value, status, result.Message)
    }
}
```

## 性能优化建议

1. **预编译正则表达式**：对于频繁使用的验证规则，可以考虑预编译正则表达式
2. **缓存验证结果**：对于重复验证相同数据的情况，可以缓存验证结果
3. **批量验证**：对于多个字段的验证，使用批量验证减少函数调用开销
4. **避免不必要的验证**：根据业务需求选择必要的验证规则

## 安全考虑

- 验证规则基于公开的标准格式
- 身份证验证仅验证格式，不涉及隐私信息验证
- 所有验证都在本地进行，不涉及网络请求
- 验证结果仅供参考，重要业务场景建议结合其他验证方式

## API 参考

### 常量定义

```go
const (
    InvalidIP = iota  // 无效IP地址
    IPv4              // IPv4地址
    IPv6              // IPv6地址
)
```

### 验证函数

#### IsMobile
验证手机号格式

**函数签名：**
```go
func IsMobile(mobile string) bool
```

**参数：**
- `mobile string`：要验证的手机号字符串

**返回值：**
- `bool`：如果是有效的手机号返回true，否则返回false

#### IsEmail
验证邮箱地址格式

**函数签名：**
```go
func IsEmail(email string) bool
```

**参数：**
- `email string`：要验证的邮箱地址字符串

**返回值：**
- `bool`：如果是有效的邮箱地址返回true，否则返回false

#### IsURL
验证URL格式

**函数签名：**
```go
func IsURL(urlStr string) bool
```

**参数：**
- `urlStr string`：要验证的URL字符串

**返回值：**
- `bool`：如果是有效的URL返回true，否则返回false

#### IsIP
验证IP地址格式并返回类型

**函数签名：**
```go
func IsIP(ip string) (bool, int)
```

**参数：**
- `ip string`：要验证的IP地址字符串

**返回值：**
- `bool`：如果是有效的IP地址返回true，否则返回false
- `int`：IP地址类型（InvalidIP、IPv4、IPv6）

#### IsPort
验证端口号格式

**函数签名：**
```go
func IsPort(portStr string) bool
```

**参数：**
- `portStr string`：要验证的端口号字符串

**返回值：**
- `bool`：如果是有效的端口号返回true，否则返回false

#### IsChineseIDCard
验证中国大陆身份证号码格式

**函数签名：**
```go
func IsChineseIDCard(idCard string) bool
```

**参数：**
- `idCard string`：要验证的身份证号码字符串

**返回值：**
- `bool`：如果是有效的身份证号码返回true，否则返回false

## 版本历史

- v1.0.0：初始版本，包含基本的验证功能
- 支持手机号、邮箱、URL、IP地址、端口、身份证号码验证

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Validator 组件。

## 许可证

Validator 组件遵循 MIT 许可证。