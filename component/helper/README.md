# Helper 组件

Helper 组件提供丰富的工具函数，包括数据类型转换、字符串处理、文件操作、JSON处理、货币计算、SSH密钥管理等常用功能。

## 概述

Helper 组件是一个功能丰富的工具库，为 Go 开发提供各种常用功能的封装。组件采用模块化设计，包含多个专门的功能模块，每个模块都提供简单易用的 API 接口。

## 功能特性

- **数据类型转换**：支持各种数据类型之间的安全转换
- **字符串处理**：提供丰富的字符串操作功能，支持多字节字符
- **JSON处理**：强大的 JSON 序列化和反序列化功能
- **文件操作**：文件路径处理、文件状态检查等实用功能
- **货币计算**：精确的货币计算，支持加减乘除等操作
- **SSH密钥管理**：SSH 密钥的生成和获取
- **数组和Map操作**：提供数组和 Map 的常用操作方法
- **工具函数**：各种实用工具函数

## 快速开始

### 安装

```go
go get github.com/jcbowen/jcbaseGo/component/helper
```

### 基本使用示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 数据类型转换示例
    convert := helper.Convert{Value: "123"}
    number := convert.ToInt()
    fmt.Printf("字符串转数字: %d\n", number)

    // 字符串处理示例
    str := helper.NewStr("hello world")
    upper := str.ToUpper()
    fmt.Printf("字符串转大写: %s\n", upper)

    // JSON处理示例
    jsonStr := `{"name":"张三","age":25}`
    jsonHelper := helper.Json(jsonStr)
    var data map[string]interface{}
    jsonHelper.ToStruct(&data)
    fmt.Printf("JSON解析结果: %+v\n", data)

    // 货币计算示例
    money := helper.Money("100.50")
    fmt.Printf("货币金额: %s\n", money.FloatString("¥", "元"))
}
```

### 数据类型转换示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 字符串转数字
    convert := helper.Convert{Value: "123.45"}
    fmt.Printf("字符串转int: %d\n", convert.ToInt())
    fmt.Printf("字符串转float64: %.2f\n", convert.ToFloat64())

    // 数字转字符串
    convert = helper.Convert{Value: 123.45}
    fmt.Printf("数字转字符串: %s\n", convert.ToString())

    // 布尔值转换
    convert = helper.Convert{Value: "true"}
    fmt.Printf("字符串转bool: %t\n", convert.ToBool())

    // 文件权限转换
    convert = helper.Convert{Value: "0644"}
    fmt.Printf("字符串转FileMode: %o\n", convert.ToFileMode())
}
```

### 字符串处理示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 创建字符串对象
    str := helper.NewStr("Hello World 你好世界")

    // 基本操作
    fmt.Printf("原始字符串: %s\n", str.String)
    fmt.Printf("转大写: %s\n", str.ToUpper())
    fmt.Printf("转小写: %s\n", str.ToLower())
    fmt.Printf("字节长度: %d\n", str.ByteLength())

    // 子字符串操作
    fmt.Printf("截取前5个字符: %s\n", str.ByteSubstr(0, 5))
    fmt.Printf("截断为10个字符: %s\n", str.Truncate(10, "..."))

    // 字符串检查
    fmt.Printf("是否以Hello开头: %t\n", str.StartsWith("Hello", true))
    fmt.Printf("是否以世界结尾: %t\n", str.EndsWith("世界", true))

    // 分割字符串
    parts := str.Explode(" ", true, false)
    fmt.Printf("按空格分割: %v\n", parts)

    // 单词统计
    fmt.Printf("单词数量: %d\n", str.CountWords())

    // Base64编码
    encoded := str.Base64UrlEncode()
    fmt.Printf("Base64编码: %s\n", encoded)
    
    decoded, _ := helper.NewStr(encoded).Base64UrlDecode()
    fmt.Printf("Base64解码: %s\n", decoded)

    // HTML转义
    htmlStr := helper.NewStr(`<script>alert("xss")</script>`)
    fmt.Printf("HTML转义: %s\n", htmlStr.EscapeHTML())

    // 驼峰转下划线
    camelStr := helper.NewStr("helloWorldTest")
    fmt.Printf("驼峰转下划线: %s\n", camelStr.ConvertCamelToSnake())
}
```

### JSON处理示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // 从字符串解析JSON
    jsonStr := `{"name":"张三","age":25}`
    jsonHelper := helper.Json(jsonStr)
    
    var user User
    jsonHelper.ToStruct(&user)
    fmt.Printf("解析结果: %+v\n", user)

    // 从结构体生成JSON
    newUser := User{Name: "李四", Age: 30}
    jsonHelper = helper.Json(newUser)
    var jsonString string
    jsonHelper.ToString(&jsonString)
    fmt.Printf("生成的JSON: %s\n", jsonString)

    // 从Map生成JSON
    data := map[string]interface{}{
        "name": "王五",
        "age":  35,
        "tags": []string{"go", "programming"},
    }
    jsonHelper = helper.Json(data)
    jsonHelper.ToString(&jsonString)
    fmt.Printf("Map生成的JSON: %s\n", jsonString)

    // 从文件读取JSON
    // jsonHelper = helper.JsonFile("config.json")
    // var config map[string]interface{}
    // jsonHelper.ToStruct(&config)
    // fmt.Printf("配置文件: %+v\n", config)

    // 生成JSON文件
    // jsonHelper.MakeFile("output.json").ToString(&jsonString)
}
```

### 文件操作示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 创建文件对象
    file := helper.NewFile(&helper.File{Path: "/path/to/file.txt"})

    // 文件状态检查
    fmt.Printf("文件是否存在: %t\n", file.Exists())
    fmt.Printf("是否是目录: %t\n", file.IsDir())
    fmt.Printf("是否是文件: %t\n", file.IsFile())
    fmt.Printf("是否为空: %t\n", file.IsEmpty())
    fmt.Printf("是否可读: %t\n", file.IsReadable())
    fmt.Printf("是否可写: %t\n", file.IsWritable())

    // 路径操作
    absPath, _ := file.GetAbsPath()
    fmt.Printf("绝对路径: %s\n", absPath)
    fmt.Printf("文件名: %s\n", file.Basename(".txt"))
    fmt.Printf("目录名: %s\n", file.DirName())

    // 目录操作
    dir := helper.NewFile(&helper.File{Path: "/path/to/directory"})
    exists, err := dir.DirExists(true)
    if err != nil {
        fmt.Printf("目录操作错误: %v\n", err)
    } else {
        fmt.Printf("目录存在或已创建: %t\n", exists)
    }
}
```

### 货币计算示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 创建货币对象
    money1 := helper.Money("100.50")  // 100.50元
    money2 := helper.Money(50.25)     // 50.25元

    // 基本操作
    fmt.Printf("金额1: %s\n", money1.FloatString("¥", ""))
    fmt.Printf("金额2: %s\n", money2.FloatString("¥", ""))

    // 加法
    result := money1.Add(money2)
    fmt.Printf("加法结果: %s\n", result.FloatString("¥", ""))

    // 减法
    result = money1.Subtract(helper.Money("20.00"))
    fmt.Printf("减法结果: %s\n", result.FloatString("¥", ""))

    // 乘法
    result = money1.Multiply(1.1) // 增加10%
    fmt.Printf("乘法结果: %s\n", result.FloatString("¥", ""))

    // 除法
    result = money1.Divide(2) // 除以2
    fmt.Printf("除法结果: %s\n", result.FloatString("¥", ""))

    // 比较操作
    fmt.Printf("金额1 > 金额2: %t\n", money1.GreaterThan(money2))
    fmt.Printf("金额1 < 金额2: %t\n", money1.LessThan(money2))
    fmt.Printf("金额1 == 金额2: %t\n", money1.Equals(money2))

    // 错误处理
    money3 := helper.Money("invalid")
    if err := money3.GetError(); err != nil {
        fmt.Printf("错误信息: %v\n", err)
    }
}
```

### SSH密钥管理示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 获取SSH公钥
    pubKey, err := helper.GetSSHKey()
    if err != nil {
        fmt.Printf("获取SSH密钥错误: %v\n", err)
        
        // 生成新的SSH密钥
        if err := helper.GenerateSSHKey(); err != nil {
            fmt.Printf("生成SSH密钥错误: %v\n", err)
        } else {
            fmt.Println("SSH密钥生成成功")
            
            // 重新获取公钥
            pubKey, err = helper.GetSSHKey()
            if err != nil {
                fmt.Printf("重新获取SSH密钥错误: %v\n", err)
            }
        }
    }

    if pubKey != "" {
        fmt.Printf("SSH公钥: %s\n", pubKey)
    }
}
```

### 数组和Map操作示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 数组操作
    arr := []string{"apple", "banana", "orange", "grape"}
    arrHelper := helper.SetArrStr(arr).DoSort()
    
    fmt.Printf("原始数组: %v\n", arr)
    fmt.Printf("排序后数组: %v\n", arrHelper.ArrayValue())
    
    // 数组差异
    otherArr := []string{"banana", "grape", "pear"}
    diff := arrHelper.ArrayDiff(otherArr)
    fmt.Printf("数组差异: %v\n", diff)
    
    // 数组交集
    intersect := arrHelper.ArrayIntersect(otherArr)
    fmt.Printf("数组交集: %v\n", intersect)

    // Map操作
    data := map[string]interface{}{
        "name": "张三",
        "age":  25,
        "city": "北京",
    }
    mapHelper := helper.NewMap(data).DoSort()
    
    fmt.Printf("原始Map: %+v\n", data)
    fmt.Printf("排序后Map: %+v\n", mapHelper.GetData())
    fmt.Printf("Map键: %v\n", mapHelper.ArrayKeys())
    fmt.Printf("Map值: %v\n", mapHelper.ArrayValues())
}
```

### 工具函数示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 字符串替换
    result, err := helper.StrReplace("world", "Go", "hello world", -1)
    if err != nil {
        fmt.Printf("字符串替换错误: %v\n", err)
    } else {
        fmt.Printf("字符串替换结果: %s\n", result)
    }

    // 检查字符串开头和结尾
    str := "hello world"
    fmt.Printf("是否以hello开头: %t\n", helper.StringStartWith(str, "hello"))
    fmt.Printf("是否以world结尾: %t\n", helper.StringEndWith(str, "world"))

    // 检查元素是否在数组中
    arr := []string{"apple", "banana", "orange"}
    fmt.Printf("apple是否在数组中: %t\n", helper.InArray("apple", arr))
    fmt.Printf("grape是否在数组中: %t\n", helper.InArray("grape", arr))

    // 通配符匹配
    pattern := "*.go"
    filename := "main.go"
    fmt.Printf("文件名 %s 是否匹配模式 %s: %t\n", 
        filename, pattern, helper.MatchWildcard(pattern, filename, false))
}
```

## 详细功能说明

### Convert 类型转换模块

提供安全的数据类型转换功能：

- **ToString()**: 将任意类型转换为字符串
- **ToInt()**: 转换为 int 类型
- **ToInt64()**: 转换为 int64 类型  
- **ToFloat64()**: 转换为 float64 类型
- **ToBool()**: 转换为 bool 类型
- **ToFileMode()**: 转换为文件权限类型
- **ToArrByte()**: 转换为字节数组

### Str 字符串处理模块

提供丰富的字符串操作功能：

- **字节操作**: ByteLength(), ByteSubstr()
- **字符串截断**: Truncate(), TruncateWords()
- **字符串检查**: StartsWith(), EndsWith()
- **字符串分割**: Explode(), CountWords()
- **编码解码**: Base64UrlEncode(), Base64UrlDecode()
- **格式转换**: ToUpper(), ToLower(), ConvertCamelToSnake()
- **安全处理**: EscapeHTML(), TrimSpace()

### JsonHelper JSON处理模块

强大的 JSON 序列化和反序列化功能：

- **多数据源支持**: 字符串、结构体、Map、文件
- **链式操作**: 支持方法链式调用
- **文件操作**: 支持 JSON 文件的读取和生成
- **错误处理**: 完善的错误处理机制

### File 文件操作模块

文件路径和状态操作：

- **文件状态**: Exists(), IsDir(), IsFile(), IsEmpty()
- **权限检查**: IsReadable(), IsWritable(), IsExecutable()
- **路径操作**: GetAbsPath(), Basename(), DirName()
- **目录管理**: DirExists()

### MoneyHelper 货币计算模块

精确的货币计算功能：

- **多种输入格式**: 字符串、数字
- **数学运算**: Add(), Subtract(), Multiply(), Divide()
- **比较操作**: GreaterThan(), LessThan(), Equals()
- **格式化输出**: FloatString()
- **错误处理**: GetError()

### SSH 密钥管理

SSH 密钥的生成和获取：

- **密钥获取**: GetSSHKey()
- **密钥生成**: GenerateSSHKey()
- **自动创建**: 密钥不存在时自动生成

### 数组和Map操作

提供数组和 Map 的常用操作方法：

- **数组操作**: ArrayValue(), ArrayDiff(), ArrayIntersect()
- **Map操作**: ArrayKeys(), ArrayValues(), GetData()
- **排序支持**: DoSort()

## 高级用法

### 组合使用示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 组合使用多个模块
    data := map[string]interface{}{
        "amount": "100.50",
        "currency": "CNY",
        "description": "商品购买",
    }

    // 使用Json模块处理数据
    jsonHelper := helper.Json(data)
    var jsonString string
    jsonHelper.ToString(&jsonString)
    
    // 使用字符串模块处理描述
    desc := helper.NewStr(data["description"].(string))
    truncatedDesc := desc.Truncate(10, "...")
    
    // 使用货币模块处理金额
    amount := helper.Money(data["amount"])
    formattedAmount := amount.FloatString("¥", "")
    
    fmt.Printf("JSON数据: %s\n", jsonString)
    fmt.Printf("描述: %s\n", truncatedDesc)
    fmt.Printf("金额: %s\n", formattedAmount)
}
```

### 错误处理最佳实践

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func ProcessPayment(amountStr string) error {
    // 货币转换和验证
    money := helper.Money(amountStr)
    if err := money.GetError(); err != nil {
        return fmt.Errorf("金额格式错误: %v", err)
    }
    
    // 金额验证
    if money.LessThan(helper.Money("0.01")) {
        return fmt.Errorf("金额不能小于0.01元")
    }
    
    // 处理支付逻辑...
    fmt.Printf("支付金额: %s\n", money.FloatString("¥", ""))
    
    return nil
}

func main() {
    amounts := []string{"100.50", "invalid", "0.00"}
    
    for _, amount := range amounts {
        fmt.Printf("处理金额: %s\n", amount)
        if err := ProcessPayment(amount); err != nil {
            fmt.Printf("处理失败: %v\n", err)
        } else {
            fmt.Println("处理成功")
        }
        fmt.Println()
    }
}
```

## 性能优化建议

1. **避免不必要的转换**: 在可能的情况下直接使用原始类型
2. **重用对象**: 对于频繁使用的对象，考虑重用而不是重复创建
3. **批量操作**: 对于数组和Map操作，使用批量方法减少循环次数
4. **错误检查**: 在使用结果前检查错误，避免无效操作
5. **内存管理**: 对于大文件操作，注意及时关闭文件句柄

## 安全考虑

- **输入验证**: 所有外部输入都应该进行验证
- **文件权限**: 文件操作时注意权限设置
- **内存安全**: 避免内存泄漏和缓冲区溢出
- **错误处理**: 完善的错误处理防止程序崩溃

## API 参考

### Convert 类型

```go
type Convert struct {
    Value interface{}
}

func (c Convert) ToString() string
func (c Convert) ToInt() int
func (c Convert) ToInt64() int64
func (c Convert) ToFloat64() float64
func (c Convert) ToBool() bool
func (c Convert) ToFileMode() os.FileMode
func (c Convert) ToArrByte() []byte
func (c Convert) ToNumber() (interface{}, bool)
```

### Str 类型

```go
type Str struct {
    String  string
    // ... 其他字段
}

func NewStr(str string) *Str
func (s *Str) ByteLength() int
func (s *Str) ByteSubstr(start int, length int) string
func (s *Str) Truncate(length int, suffix string) string
func (s *Str) StartsWith(prefix string, caseSensitive bool) bool
func (s *Str) EndsWith(suffix string, caseSensitive bool) bool
func (s *Str) Explode(delimiter string, trim bool, skipEmpty bool) []string
func (s *Str) Base64UrlEncode() string
func (s *Str) Base64UrlDecode() (string, error)
func (s *Str) EscapeHTML() string
func (s *Str) ConvertCamelToSnake() string
```

### JsonHelper 类型

```go
type JsonHelper struct {
    // ... 字段
}

func Json(input interface{}) *JsonHelper
func JsonFile(path string) *JsonHelper
func (jh *JsonHelper) ToStruct(newStruct interface{}) *JsonHelper
func (jh *JsonHelper) ToString(newStr *string) *JsonHelper
func (jh *JsonHelper) MakeFile(filepath string) *JsonHelper
```

### File 类型

```go
type File struct {
    Path string
    Perm os.FileMode
}

func NewFile(args ...any) *File
func (fh *File) Exists() bool
func (fh *File) IsDir() bool
func (fh *File) GetAbsPath() (string, error)
func (fh *File) Basename(suffix string) string
func (fh *File) DirName() string
func (fh *File) DirExists(createIfNotExists bool) (bool, error)
```

### MoneyHelper 类型

```go
type MoneyHelper struct {
    Amount int64
}

func Money(input interface{}) *MoneyHelper
func (m *MoneyHelper) SetAmount(value interface{}) *MoneyHelper
func (m *MoneyHelper) Add(other *MoneyHelper) *MoneyHelper
func (m *MoneyHelper) Subtract(other *MoneyHelper) *MoneyHelper
func (m *MoneyHelper) Multiply(factor float64) *MoneyHelper
func (m *MoneyHelper) Divide(divisor float64) *MoneyHelper
func (m *MoneyHelper) FloatString(parts ...string) string
func (m *MoneyHelper) GreaterThan(other *MoneyHelper) bool
func (m *MoneyHelper) GetError() error
```

### 工具函数

```go
func GetSSHKey() (string, error)
func GenerateSSHKey() error
func SetArrStr(str []string) *ArrStr
func NewMap(mapData map[string]interface{}) *MapHelper
func StrReplace(search interface{}, replace interface{}, subject interface{}, count int) (interface{}, error)
func StringStartWith(str, prefix string) bool
func StringEndWith(str, suffix string) bool
func InArray(needle interface{}, haystack []string) bool
func MatchWildcard(pattern, s string, caseSensitive bool) bool
```

## 版本历史

- v1.0.0：初始版本，包含完整的工具函数集合
- 支持数据类型转换、字符串处理、JSON处理、文件操作、货币计算等功能

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Helper 组件。

## 许可证

Helper 组件遵循 MIT 许可证。