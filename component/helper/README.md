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

### HTTP处理功能

提供HTTP请求和响应处理的实用功能：

- **头信息提取**: 从HTTP请求中提取和解析头信息
- **主机信息获取**: 获取请求的主机名、端口和协议信息
- **URL编码**: 对URL参数进行编码和解码
- **头信息格式化**: 将头信息格式化为标准格式

#### 核心函数说明

**ExtractHeaders(headers map[string][]string) map[string]string**
- 从HTTP头信息中提取单个值，处理多值情况
- 返回简化的键值对映射

**GetHostInfo(host string) (string, string, string)**
- 解析主机字符串，返回主机名、端口和协议
- 支持带端口和协议的主机格式

**FormatHeaders(headers map[string]string) string**
- 将头信息格式化为标准HTTP头格式
- 每个头信息单独一行，格式为"Key: Value"

**URLEncodeParams(params map[string]string) string**
- 对URL参数进行编码，生成查询字符串
- 自动处理特殊字符编码

**URLDecodeParams(query string) (map[string]string, error)**
- 解码URL查询字符串，返回参数映射
- 处理URL编码的特殊字符

### IP地址处理功能

提供IP地址验证和处理的实用功能：

- **IP地址验证**: 验证IP地址格式的有效性
- **地址类型判断**: 判断IP地址是否为回环地址、私有地址等
- **CIDR范围检查**: 检查IP地址是否在指定CIDR范围内

#### 核心函数说明

**IsValid(ip string) bool**
- 验证IP地址格式是否有效
- 支持IPv4和IPv6地址格式

**IsLoopback(ip string) bool**
- 判断IP地址是否为回环地址
- IPv4: 127.0.0.0/8, IPv6: ::1

**IsPrivate(ip string) bool**
- 判断IP地址是否为私有地址
- 包括局域网、私有网络等地址范围

**IsInCIDR(ip, cidr string) bool**
- 检查IP地址是否在指定CIDR范围内
- 支持IPv4和IPv6的CIDR表示法

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

#### 核心函数说明

**NewStr(s string) *Str**
- 创建新的字符串对象
- 初始化字符串内容和状态

**ByteLength() int**
- 获取字符串的字节长度
- 与字符长度不同，考虑多字节字符

**ByteSubstr(start, length int) string**
- 按字节位置截取子字符串
- 支持负数的起始位置和长度

**Truncate(maxLength int, suffix string) string**
- 截断字符串到指定长度
- 可选添加后缀表示截断

**TruncateWords(maxWords int, suffix string) string**
- 按单词数量截断字符串
- 保留完整的单词边界

**StartsWith(prefix string, caseSensitive bool) bool**
- 检查字符串是否以指定前缀开头
- `caseSensitive`: 是否区分大小写

**EndsWith(suffix string, caseSensitive bool) bool**
- 检查字符串是否以指定后缀结尾
- `caseSensitive`: 是否区分大小写

**Explode(delimiter string, trimSpace, removeEmpty bool) []string**
- 按分隔符分割字符串
- `trimSpace`: 是否去除空白字符
- `removeEmpty`: 是否移除空元素

**CountWords() int**
- 统计字符串中的单词数量
- 基于空白字符分割单词

**Base64UrlEncode() string**
- 对字符串进行Base64 URL安全编码
- 替换特殊字符，适合URL传输

**Base64UrlDecode() (string, error)**
- 对Base64 URL编码的字符串进行解码
- 处理URL安全的Base64编码

**EscapeHTML() string**
- 对HTML特殊字符进行转义
- 防止XSS攻击

**TrimSpace() string**
- 去除字符串两端的空白字符
- 包括空格、制表符、换行符等

**ConvertCamelToSnake() string**
- 将驼峰命名转换为下划线命名
- 处理大小写转换和分隔符添加

### JsonHelper JSON处理模块

强大的 JSON 序列化和反序列化功能：

- **多数据源支持**: 字符串、结构体、Map、文件
- **链式操作**: 支持方法链式调用
- **文件操作**: 支持 JSON 文件的读取和生成
- **错误处理**: 完善的错误处理机制

#### 核心函数说明

**Json(data interface{}) *JsonHelper**
- 创建JSON处理对象，支持多种数据源
- `data`: 可以是字符串、结构体、Map等

**JsonFile(filename string) *JsonHelper**
- 从JSON文件创建处理对象
- 自动读取文件内容并解析

**ToStruct(target interface{}) error**
- 将JSON数据解析到目标结构体
- 支持结构体标签映射

**ToString(target *string) error**
- 将JSON数据转换为字符串
- 支持格式化输出

**MakeFile(filename string) *JsonHelper**
- 设置输出文件名，用于生成JSON文件
- 配合ToString方法使用

**GetError() error**
- 获取处理过程中的错误信息
- 返回最后一次操作的错误

### File 文件操作模块

文件路径和状态操作：

- **文件状态**: Exists(), IsDir(), IsFile(), IsEmpty()
- **权限检查**: IsReadable(), IsWritable(), IsExecutable()
- **路径操作**: GetAbsPath(), Basename(), DirName()
- **目录管理**: DirExists()

#### 核心函数说明

**NewFile(file *File) *File**
- 创建新的文件对象实例
- 初始化文件路径和状态信息

**Exists() bool**
- 检查文件或目录是否存在
- 返回布尔值表示存在状态

**IsDir() bool**
- 判断路径是否为目录
- 如果路径不存在或不是目录返回false

**IsFile() bool**
- 判断路径是否为普通文件
- 排除目录、符号链接等特殊文件类型

**IsEmpty() bool**
- 检查文件是否为空
- 对于目录，检查是否包含文件或子目录

**IsReadable() bool**
- 检查文件是否可读
- 验证当前用户对文件的读取权限

**IsWritable() bool**
- 检查文件是否可写
- 验证当前用户对文件的写入权限

**IsExecutable() bool**
- 检查文件是否可执行
- 验证当前用户对文件的执行权限

**GetAbsPath() (string, error)**
- 获取文件的绝对路径
- 解析相对路径和符号链接

**Basename(suffix string) string**
- 获取文件的基本名称
- 可选去除指定后缀

**DirName() string**
- 获取文件所在目录的路径
- 返回父目录的路径字符串

**DirExists(createIfNotExist bool) (bool, error)**
- 检查目录是否存在，可选自动创建
- `createIfNotExist`: 如果目录不存在是否自动创建

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

### Map字段提取功能

提供多级Map字段提取功能，支持点号分隔的路径访问：

- **ExtractString()**: 从多级map中提取字符串字段值
- **Extract()**: 从多级map中提取任意类型的字段值
- **ExtractWithDefault()**: 从多级map中提取字段值，支持默认值

#### 核心函数说明

**NewMap(data map[string]interface{}) *MapHelper**
- 创建MapHelper对象
- 初始化Map数据

**DoSort() *MapHelper**
- 对Map数据进行排序
- 返回排序后的MapHelper对象

**GetData() map[string]interface{}**
- 获取原始的Map数据
- 返回当前存储的数据

**ArrayKeys() []string**
- 获取Map的所有键
- 返回键的字符串切片

**ArrayValues() []interface{}**
- 获取Map的所有值
- 返回值的接口切片

**ExtractString(path string) string**
- 从多级Map中提取字符串值
- 支持点号分隔的路径访问
- 如果路径不存在返回空字符串

**Extract(path string) interface{}**
- 从多级Map中提取任意类型的值
- 支持点号分隔的路径访问
- 返回原始值或nil

**ExtractWithDefault(path string, defaultValue interface{}) interface{}**
- 从多级Map中提取值，支持默认值
- 如果路径不存在返回指定的默认值

#### 使用示例

```go
package main

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
	// 创建多层嵌套的Map数据
	data := map[string]interface{}{
		"api": map[string]interface{}{
			"response": map[string]interface{}{
				"status": "success",
				"code":   200,
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"id":       123,
						"username": "john_doe",
						"profile": map[string]interface{}{
							"name":  "John Doe",
							"email": "john@example.com",
						},
					},
				},
			},
		},
	}

	mapHelper := helper.NewMap(data)

	// 使用 ExtractString 提取字符串值
	username := mapHelper.ExtractString("api.response.data.user.username")
	fmt.Printf("用户名: %s\n", username) // 输出: john_doe

	// 使用 Extract 提取原始值
	userID := mapHelper.Extract("api.response.data.user.id")
	fmt.Printf("用户ID: %v\n", userID) // 输出: 123

	// 使用 ExtractWithDefault 处理可能不存在的字段
	avatar := mapHelper.ExtractWithDefault("api.response.data.user.avatar", "default.png")
	fmt.Printf("头像: %v\n", avatar) // 输出: default.png

	// 处理多层嵌套
	email := mapHelper.ExtractString("api.response.data.user.profile.email")
	fmt.Printf("邮箱: %s\n", email) // 输出: john@example.com

	// 处理不存在的路径
	nonexistent := mapHelper.ExtractString("api.response.data.user.nonexistent")
	fmt.Printf("不存在的字段: '%s'\n", nonexistent) // 输出: ''
}
```

#### 功能特点

- **路径支持**: 使用点号分隔的多级路径访问
- **类型安全**: 自动处理类型转换和边界情况
- **默认值**: 支持自定义默认值处理
- **链式调用**: 与现有 MapHelper 方法兼容

### 单位转换功能

提供强大的单位转换功能，支持存储单位、时间单位和数据单位之间的转换：

- **存储单位转换**: 字节(Byte)、千字节(KB)、兆字节(MB)、吉字节(GB)、太字节(TB)、拍字节(PB)
- **时间单位转换**: 纳秒(ns)、微秒(μs)、毫秒(ms)、秒(s)、分钟(min)、小时(h)、天(d)
- **数据单位转换**: 比特(b)、字节(byte)、千比特(Kb)、千字节(Kbyte)、兆比特(Mb)、兆字节(Mbyte)、吉比特(Gb)、吉字节(Gbyte)
- **智能解析**: 支持多种单位符号格式，自动识别单位类型
- **精确计算**: 基于单位换算因子进行精确转换

#### 核心函数说明

**ParseUnitString(value string, unitType int) (float64, Unit, error)**
- 解析包含数值和单位的字符串，返回数值和单位信息
- `unitType`: 指定解析的单位类型（UnitTypeStorage/UnitTypeTime/UnitTypeData）

**ConvertUnit(value float64, fromUnit, toUnit string, unitType int) (float64, error)**
- 在不同单位之间进行数值转换
- 支持存储、时间、数据三种单位类型的转换

**FormatUnit(value interface{}, unitType UnitType, precision int, toUnit ...string) (string, error)**
- 将数值格式化为人类可读的单位字符串
- `value`: 支持数值或带单位的字符串
- `unitType`: 指定单位类型（UnitTypeStorage/UnitTypeTime/UnitTypeData）
- `precision`: 指定小数位数精度
- `toUnit`: 输出格式（"auto"自动选择单位或具体单位符号）



**GetAvailableUnits(unitType int) []string**
- 获取指定单位类型下所有可用的单位符号列表

**IsValidUnit(unit string, unitType int) bool**
- 检查指定单位符号在给定单位类型下是否有效

#### 使用示例

```go
package main

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
	// 存储单位转换示例
	fmt.Println("=== 存储单位转换 ===")
	result, err := helper.ConvertUnit(1024, "B", "KB", helper.UnitTypeStorage)
	if err != nil {
		fmt.Printf("转换错误: %v\n", err)
	} else {
		fmt.Printf("1024 B = %.2f KB\n", result) // 输出: 1024 B = 1.00 KB
	}

	// 时间单位转换示例
	fmt.Println("\n=== 时间单位转换 ===")
	result, err = helper.ConvertUnit(3600, "s", "h", helper.UnitTypeTime)
	if err != nil {
		fmt.Printf("转换错误: %v\n", err)
	} else {
		fmt.Printf("3600 s = %.2f h\n", result) // 输出: 3600 s = 1.00 h
	}

	// 数据单位转换示例
	fmt.Println("\n=== 数据单位转换 ===")
	result, err = helper.ConvertUnit(8, "b", "byte", helper.UnitTypeData)
	if err != nil {
		fmt.Printf("转换错误: %v\n", err)
	} else {
		fmt.Printf("8 b = %.2f byte\n", result) // 输出: 8 b = 1.00 byte
	}

	// 单位字符串解析示例
	fmt.Println("\n=== 单位字符串解析 ===")
	value, unit, err := helper.ParseUnitString("1MB", helper.UnitTypeStorage)
	if err != nil {
		fmt.Printf("解析错误: %v\n", err)
	} else {
		fmt.Printf("解析结果: 值=%.0f, 单位=%s (类型: %d)\n", value, unit.Name, unit.Type)
	}

	// 格式化单位值示例
	fmt.Println("\n=== 格式化单位值 ===")
	formatted := helper.FormatUnit(1500000, helper.UnitTypeStorage, 2)
	fmt.Printf("格式化结果: %s\n", formatted) // 输出: 1.43 MB

	// 获取可用单位列表
	fmt.Println("\n=== 可用单位列表 ===")
	storageUnits := helper.GetAvailableUnits(helper.UnitTypeStorage)
	timeUnits := helper.GetAvailableUnits(helper.UnitTypeTime)
	dataUnits := helper.GetAvailableUnits(helper.UnitTypeData)
	
	fmt.Printf("存储单位: %v\n", storageUnits)
	fmt.Printf("时间单位: %v\n", timeUnits)
	fmt.Printf("数据单位: %v\n", dataUnits)

	// 人类可读格式示例
	fmt.Println("\n=== 人类可读格式 ===")
	
	

}
```

#### 功能特点

- **多类型支持**: 支持存储、时间、数据三种单位类型
- **符号兼容**: 支持多种单位符号格式（如B、byte、bytes等）
- **精确转换**: 基于精确的换算因子进行计算
- **错误处理**: 完善的错误处理机制
- **格式化输出**: 支持自定义精度和格式化
- **智能解析**: 自动识别和解析单位字符串
- **单位类型参数**: 支持指定解析的单位类型，解决符号冲突问题

#### 单位类型参数使用示例

当单位符号存在冲突时（如"byte"同时存在于存储单位和数据单位），可以使用单位类型参数明确指定解析类型：

```go
package main

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
	// 默认情况下，"byte"被解析为存储单位
	fmt.Println("=== 默认解析 ===")
	_, unitInfo, err := helper.ParseUnitString("8byte", helper.UnitTypeStorage)
	if err != nil {
		fmt.Printf("解析错误: %v\n", err)
	} else {
		fmt.Printf("默认解析: 8byte -> 类型: %v\n", unitInfo.UnitType)
	}

	// 使用单位类型参数指定解析为数据单位
	fmt.Println("\n=== 指定单位类型 ===")
	_, unitInfo, err = helper.ParseUnitString("8byte", helper.UnitTypeData)
	if err != nil {
		fmt.Printf("解析错误: %v\n", err)
	} else {
		fmt.Printf("指定类型解析: 8byte -> 类型: %v\n", unitInfo.UnitType)
	}

	// 单位转换时使用单位类型参数
	fmt.Println("\n=== 单位转换类型参数 ===")
	
	// 默认情况下，b和byte类型不同，转换会失败
	result, err := helper.ConvertUnit(8, "b", "byte", helper.UnitTypeStorage)
	if err != nil {
		fmt.Printf("默认转换失败: %v\n", err)
	} else {
		fmt.Printf("默认转换成功: 8b = %.2f byte\n", result)
	}

	// 指定单位类型为数据单位，转换成功
	result, err = helper.ConvertUnit(8, "b", "byte", helper.UnitTypeData)
	if err != nil {
		fmt.Printf("指定类型转换失败: %v\n", err)
	} else {
		fmt.Printf("指定类型转换成功: 8b = %.2f byte\n", result)
	}

	// 其他函数也支持单位类型参数
	fmt.Println("\n=== 其他函数类型参数 ===")
	
	// 单位有效性检查
	isValid := helper.IsValidUnit("byte", helper.UnitTypeData)
	fmt.Printf("byte作为数据单位是否有效: %t\n", isValid)

	// 获取单位类型
	unitType, err := helper.GetUnitType("byte", helper.UnitTypeData)
	if err != nil {
		fmt.Printf("获取类型错误: %v\n", err)
	} else {
		fmt.Printf("byte作为数据单位的类型: %v\n", unitType)
	}
}
```

#### 单位类型参数说明

- **UnitTypeStorage**: 存储单位类型（默认）
- **UnitTypeTime**: 时间单位类型  
- **UnitTypeData**: 数据单位类型

当单位符号存在歧义时，系统会匹配指定类型的单位。如果没有指定单位类型，则按照存储单位→时间单位→数据单位的顺序进行匹配。

#### 支持的单位符号

**存储单位**:
- Byte: `B`, `bytes`
- Kilobyte: `KB`, `K`, `kilobyte`, `kilobytes`
- Megabyte: `MB`, `M`, `megabyte`, `megabytes`
- Gigabyte: `GB`, `G`, `gigabyte`, `gigabytes`
- Terabyte: `TB`, `T`, `terabyte`, `terabytes`
- Petabyte: `PB`, `P`, `petabyte`, `petabytes`

**时间单位**:
- Nanosecond: `ns`, `nanosecond`, `nanoseconds`
- Microsecond: `μs`, `us`, `microsecond`, `microseconds`
- Millisecond: `ms`, `millisecond`, `milliseconds`
- Second: `s`, `sec`, `second`, `seconds`
- Minute: `m`, `min`, `minute`, `minutes`
- Hour: `h`, `hr`, `hour`, `hours`
- Day: `d`, `day`, `days`

**数据单位**:
- Bit: `b`, `bit`, `bits`
- Byte: `byte`, `bytes`
- Kilobit: `Kb`, `Kbit`, `kilobit`, `kilobits`
- Kilobyte: `Kbyte`, `kilobyte`, `kilobytes`
- Megabit: `Mb`, `Mbit`, `megabit`, `megabits`
- Megabyte: `Mbyte`, `megabyte`, `megabytes`
- Gigabit: `Gb`, `Gbit`, `gigabit`, `gigabits`
- Gigabyte: `Gbyte`, `gigabyte`, `gigabytes`

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
func (c Convert) ToTime() time.Time
func (c Convert) ToMap() map[string]interface{}
func (c Convert) ToSlice() []interface{}
func (c Convert) ToDuration() time.Duration
func (c Convert) ToInterface() interface{}
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

### MapHelper 类型

```go
type MapHelper struct {
    Data map[string]interface{}
    Keys []string
    Sort bool
}

func NewMap(mapData map[string]interface{}) *MapHelper
func (m *MapHelper) DoSort() *MapHelper
func (m *MapHelper) ArrayKeys() []string
func (m *MapHelper) ArrayValues() []interface{}
func (m *MapHelper) GetData() map[string]interface{}
func (m *MapHelper) ExtractString(path string) string
func (m *MapHelper) Extract(path string) interface{}
func (m *MapHelper) ExtractWithDefault(path string, defaultValue interface{}) interface{}
### 单位转换函数

```go
// 单位类型定义
type UnitType int
const (
    UnitTypeStorage UnitType = iota
    UnitTypeTime
    UnitTypeData
)

// 单位结构
type Unit struct {
    Name     string
    Symbols  []string
    Factor   float64
    UnitType UnitType
}

// 单位转换函数
func ParseUnitString(str string, unitType UnitType) (float64, *Unit, error)
func FormatUnit(value float64, unitType UnitType, precision int) string
func ConvertUnit(value float64, fromUnit, toUnit string, unitType UnitType) (float64, error)
func IsValidUnit(unitStr string, unitType UnitType) bool
func GetUnitType(unitStr string, unitType UnitType) (UnitType, error)
func GetAvailableUnits(unitType UnitType) []string

func RoundUnitValue(value float64, unit string, precision int, unitType UnitType) (float64, error)

// Str类型的单位相关方法
func (s *Str) ParseUnit(unitType UnitType) (float64, *Unit, error)
func (s *Str) ToUnitValue(unitType UnitType) (float64, error)
```

### 工具函数

```go
func GetSSHKey() (string, error)
func GenerateSSHKey() error
func SetArrStr(str []string) *ArrStr
func NewMap(mapData map[string]interface{}) *MapHelper
func StrReplace(search interface{}, replace interface{}, subject interface{}, count int) (interface{}, error)
func StringStartWith(str, prefix string) bool
```
func StringEndWith(str, suffix string) bool
func InArray(needle interface{}, haystack []string) bool
func MatchWildcard(pattern, s string, caseSensitive bool) bool
```

## 版本历史

- v1.0.0：初始版本，包含完整的工具函数集合
- 支持数据类型转换、字符串处理、JSON处理、文件操作、货币计算等功能

## 性能优化建议

### 字符串处理优化

1. **避免不必要的字符串转换**
   ```go
   // 不推荐：频繁创建字符串对象
   for i := 0; i < 1000; i++ {
       str := helper.NewStr("test" + strconv.Itoa(i))
       // 处理字符串
   }
   
   // 推荐：复用字符串对象
   str := helper.NewStr("")
   for i := 0; i < 1000; i++ {
       str.String = "test" + strconv.Itoa(i)
       // 处理字符串
   }
   ```

2. **使用字节操作替代字符串操作**
   ```go
   // 对于大量数据处理，优先使用字节操作
   str := helper.NewStr(largeText)
   byteLen := str.ByteLength()  // 比 len(str.String) 更高效
   ```

### JSON处理优化

1. **批量处理JSON数据**
   ```go
   // 不推荐：逐个处理JSON对象
   for _, item := range items {
       jsonHelper := helper.Json(item)
       // 处理单个JSON
   }
   
   // 推荐：批量处理JSON数组
   jsonHelper := helper.Json(items)
   // 批量处理所有JSON对象
   ```

2. **使用结构体标签优化序列化**
   ```go
   type User struct {
       Name string `json:"name,omitempty"`  // 空值不序列化
       Age  int    `json:"age"`
   }
   ```

### 文件操作优化

1. **减少文件系统调用**
   ```go
   // 不推荐：多次检查文件状态
   if file.Exists() {
       if file.IsReadable() {
           // 读取文件
       }
   }
   
   // 推荐：缓存文件状态信息
   file := helper.NewFile(&helper.File{Path: "test.txt"})
   if exists := file.Exists(); exists {
       // 一次性获取所有状态信息
       isReadable := file.IsReadable()
       isWritable := file.IsWritable()
   }
   ```

2. **使用绝对路径避免路径解析开销**
   ```go
   // 推荐：使用绝对路径
   absPath, _ := filepath.Abs("relative/path.txt")
   file := helper.NewFile(&helper.File{Path: absPath})
   ```

### 内存使用优化

1. **及时释放大对象**
   ```go
   // 处理大文件后及时释放
   func processLargeFile() {
       file := helper.NewFile(&helper.File{Path: "large.txt"})
       // 处理文件
       // 处理完成后，让GC回收内存
       file = nil
   }
   ```

2. **避免内存泄漏**
   ```go
   // 在循环中避免创建不必要的对象引用
   var results []string
   for i := 0; i < 10000; i++ {
       str := helper.NewStr(fmt.Sprintf("item%d", i))
       // 只保留需要的数据，而不是整个对象
       results = append(results, str.String)
   }
   ```

## 使用最佳实践

### 错误处理最佳实践

1. **始终检查错误返回值**
   ```go
   // 不推荐：忽略错误
   jsonHelper.ToStruct(&data)
   
   // 推荐：正确处理错误
   if err := jsonHelper.ToStruct(&data); err != nil {
       log.Printf("JSON解析失败: %v", err)
       return
   }
   ```

2. **使用错误包装提供上下文信息**
   ```go
   func processUserData(jsonStr string) error {
       jsonHelper := helper.Json(jsonStr)
       var user User
       if err := jsonHelper.ToStruct(&user); err != nil {
           return fmt.Errorf("处理用户数据失败: %w", err)
       }
       return nil
   }
   ```

### 并发安全最佳实践

1. **避免共享可变状态**
   ```go
   // 不推荐：在goroutine中共享可变对象
   var sharedStr = helper.NewStr("shared")
   
   go func() {
       sharedStr.String = "modified"  // 竞态条件
   }()
   
   // 推荐：每个goroutine使用独立对象
   go func(str *helper.Str) {
       str.String = "modified"  // 安全
   }(helper.NewStr("local"))
   ```

2. **使用同步机制保护共享资源**
   ```go
   type SafeStringProcessor struct {
       mu sync.RWMutex
       str *helper.Str
   }
   
   func (s *SafeStringProcessor) Process() {
       s.mu.Lock()
       defer s.mu.Unlock()
       // 安全地处理字符串
   }
   ```

### 代码组织最佳实践

1. **模块化使用helper组件**
   ```go
   // 推荐：按功能模块组织代码
   type UserService struct {
       jsonHelper *helper.JsonHelper
       strHelper  *helper.Str
   }
   
   func (s *UserService) ProcessUser(jsonStr string) error {
       // 使用helper组件处理业务逻辑
       return nil
   }
   ```

2. **合理使用接口抽象**
   ```go
   // 定义接口，便于测试和替换
   type StringProcessor interface {
       Process(input string) string
   }
   
   type HelperStringProcessor struct {
       str *helper.Str
   }
   
   func (p *HelperStringProcessor) Process(input string) string {
       p.str.String = input
       return p.str.ToUpper()
   }
   ```

### 测试最佳实践

1. **编写单元测试覆盖helper功能**
   ```go
   func TestStr_ToUpper(t *testing.T) {
       str := helper.NewStr("hello")
       result := str.ToUpper()
       if result != "HELLO" {
           t.Errorf("期望 HELLO, 得到 %s", result)
       }
   }
   ```

2. **使用表格驱动测试**
   ```go
   func TestConvert_ToInt(t *testing.T) {
       tests := []struct {
           name     string
           input    interface{}
           expected int
       }{
           {"整数", 123, 123},
           {"字符串", "456", 456},
           {"浮点数", 78.9, 78},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               convert := helper.Convert{Value: tt.input}
               result := convert.ToInt()
               if result != tt.expected {
                   t.Errorf("期望 %d, 得到 %d", tt.expected, result)
               }
           })
       }
   }
   ```

通过遵循这些性能优化建议和使用最佳实践，您可以确保helper组件在各种场景下都能提供最佳的性能和可靠性。

## 错误处理和异常情况说明

Helper 组件提供了完善的错误处理机制，帮助开发者优雅地处理各种异常情况。

### 错误处理原则

1. **显式错误返回**: 所有可能失败的操作都返回错误信息
2. **错误信息清晰**: 错误信息包含足够的上下文信息
3. **错误类型区分**: 区分不同类型的错误以便针对性处理
4. **错误传播**: 支持错误链式传播，便于调试

### 常见错误类型

#### 1. 输入验证错误

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 无效的JSON字符串
    jsonStr := `{"name": "张三", "age": "invalid"}`
    jsonHelper := helper.Json(jsonStr)
    var data map[string]interface{}
    
    if err := jsonHelper.ToStruct(&data); err != nil {
        fmt.Printf("JSON解析错误: %v\n", err)
        // 输出: JSON解析错误: json: cannot unmarshal string into Go struct field .age of type int
    }
    
    // 无效的货币格式
    money := helper.Money("invalid")
    if err := money.GetError(); err != nil {
        fmt.Printf("货币格式错误: %v\n", err)
        // 输出: 货币格式错误: strconv.ParseFloat: parsing "invalid": invalid syntax
    }
}
```

#### 2. 文件操作错误

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 文件不存在
    file := helper.NewFile("/path/to/nonexistent/file.txt")
    if !file.Exists() {
        fmt.Println("文件不存在")
    }
    
    // 权限不足
    file = helper.NewFile("/root/protected/file.txt")
    if !file.IsReadable() {
        fmt.Println("文件不可读，权限不足")
    }
    
    // 获取绝对路径错误
    absPath, err := file.GetAbsPath()
    if err != nil {
        fmt.Printf("获取绝对路径失败: %v\n", err)
    }
}
```

#### 3. 单位转换错误

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // 无效的单位符号
    result, err := helper.ConvertUnit(100, "invalid", "KB", helper.UnitTypeStorage)
    if err != nil {
        fmt.Printf("单位转换错误: %v\n", err)
        // 输出: 单位转换错误: 无效的单位符号: invalid
    }
    
    // 不兼容的单位类型
    result, err = helper.ConvertUnit(100, "B", "s", helper.UnitTypeStorage)
    if err != nil {
        fmt.Printf("单位类型不兼容: %v\n", err)
        // 输出: 单位类型不兼容: 单位类型不匹配
    }
    
    // 单位字符串解析错误
    _, _, err = helper.ParseUnitString("invalid", helper.UnitTypeStorage)
    if err != nil {
        fmt.Printf("单位字符串解析错误: %v\n", err)
        // 输出: 单位字符串解析错误: 无法解析单位字符串
    }
}
```

#### 4. 字符串处理错误

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
    // Base64解码错误
    str := helper.NewStr("invalid base64")
    decoded, err := str.Base64UrlDecode()
    if err != nil {
        fmt.Printf("Base64解码错误: %v\n", err)
        // 输出: Base64解码错误: illegal base64 data at input byte 7
    }
    
    // 字符串截取越界
    str = helper.NewStr("hello")
    substr := str.ByteSubstr(0, 10) // 越界访问
    fmt.Printf("截取结果: %s\n", substr) // 输出: hello (自动处理越界)
}
```

### 错误处理最佳实践

#### 1. 防御性编程

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func SafeProcessData(input string) error {
    // 输入验证
    if input == "" {
        return fmt.Errorf("输入不能为空")
    }
    
    // 使用helper组件处理
    str := helper.NewStr(input)
    if str.ByteLength() > 1000 {
        return fmt.Errorf("输入长度超过限制")
    }
    
    // 处理逻辑
    processed := str.ToUpper()
    fmt.Printf("处理结果: %s\n", processed)
    
    return nil
}

func main() {
    inputs := []string{"", "hello", "very long string..."}
    
    for _, input := range inputs {
        if err := SafeProcessData(input); err != nil {
            fmt.Printf("处理失败: %v\n", err)
        } else {
            fmt.Println("处理成功")
        }
    }
}
```

#### 2. 错误链式传播

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func ProcessUserData(jsonStr string) error {
    // 解析JSON
    jsonHelper := helper.Json(jsonStr)
    var userData map[string]interface{}
    if err := jsonHelper.ToStruct(&userData); err != nil {
        return fmt.Errorf("解析用户数据失败: %w", err)
    }
    
    // 提取用户名
    mapHelper := helper.NewMap(userData)
    name := mapHelper.ExtractString("name")
    if name == "" {
        return fmt.Errorf("用户名不能为空")
    }
    
    // 处理用户名
    str := helper.NewStr(name)
    if str.ByteLength() < 2 {
        return fmt.Errorf("用户名长度太短")
    }
    
    fmt.Printf("处理用户: %s\n", str.ToUpper())
    return nil
}

func main() {
    jsonStr := `{"name": "张"}` // 用户名太短
    
    if err := ProcessUserData(jsonStr); err != nil {
        fmt.Printf("用户数据处理失败: %v\n", err)
        // 输出: 用户数据处理失败: 用户名长度太短
    }
}
```

#### 3. 错误恢复机制

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func SafeUnitConversion(value float64, fromUnit, toUnit string) (result float64, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("单位转换发生panic: %v", r)
        }
    }()
    
    // 尝试单位转换
    result, err = helper.ConvertUnit(value, fromUnit, toUnit, helper.UnitTypeStorage)
    if err != nil {
        return 0, fmt.Errorf("单位转换失败: %w", err)
    }
    
    return result, nil
}

func main() {
    // 正常情况
    result, err := SafeUnitConversion(1024, "B", "KB")
    if err != nil {
        fmt.Printf("转换失败: %v\n", err)
    } else {
        fmt.Printf("转换结果: %.2f KB\n", result)
    }
    
    // 异常情况
    result, err = SafeUnitConversion(1024, "invalid", "KB")
    if err != nil {
        fmt.Printf("转换失败: %v\n", err)
        // 输出: 转换失败: 单位转换失败: 无效的单位符号: invalid
    }
}
```

### 异常情况处理

#### 1. 边界条件处理

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

func HandleEdgeCases() {
    // 空字符串处理
    emptyStr := helper.NewStr("")
    fmt.Printf("空字符串长度: %d\n", emptyStr.ByteLength()) // 输出: 0
    
    // 超大数字处理
    bigMoney := helper.Money("999999999999999.99")
    if err := bigMoney.GetError(); err != nil {
        fmt.Printf("超大金额错误: %v\n", err)
    }
    
    // 特殊字符处理
    specialStr := helper.NewStr("hello\x00world") // 包含空字符
    fmt.Printf("特殊字符处理: %s\n", specialStr.String)
}

func main() {
    HandleEdgeCases()
}
```

#### 2. 并发安全处理

```go
package main

import (
    "fmt"
    "sync"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

type SafeProcessor struct {
    mu sync.RWMutex
    data map[string]interface{}
}

func (p *SafeProcessor) SafeExtract(path string) (interface{}, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    if p.data == nil {
        return nil, fmt.Errorf("数据为空")
    }
    
    mapHelper := helper.NewMap(p.data)
    result := mapHelper.Extract(path)
    if result == nil {
        return nil, fmt.Errorf("路径不存在: %s", path)
    }
    
    return result, nil
}

func main() {
    processor := &SafeProcessor{
        data: map[string]interface{}{
            "user": map[string]interface{}{
                "name": "张三",
                "age": 25,
            },
        },
    }
    
    // 并发安全访问
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            result, err := processor.SafeExtract("user.name")
            if err != nil {
                fmt.Printf("协程%d错误: %v\n", id, err)
            } else {
                fmt.Printf("协程%d结果: %v\n", id, result)
            }
        }(i)
    }
    wg.Wait()
}
```

### 错误信息国际化

Helper 组件支持错误信息的国际化处理：

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/helper"
)

// 错误信息映射表
var errorMessages = map[string]map[string]string{
    "en": {
        "invalid_json": "Invalid JSON format",
        "file_not_found": "File not found",
        "unit_conversion_error": "Unit conversion error",
    },
    "zh": {
        "invalid_json": "JSON格式无效",
        "file_not_found": "文件不存在",
        "unit_conversion_error": "单位转换错误",
    },
}

func GetErrorMessage(key, lang string) string {
    if messages, ok := errorMessages[lang]; ok {
        if msg, ok := messages[key]; ok {
            return msg
        }
    }
    return key // 默认返回key
}

func main() {
    // 中文错误信息
    fmt.Printf("中文错误: %s\n", GetErrorMessage("file_not_found", "zh"))
    
    // 英文错误信息
    fmt.Printf("英文错误: %s\n", GetErrorMessage("unit_conversion_error", "en"))
}
```

通过遵循这些错误处理和异常情况说明，您可以确保在使用 Helper 组件时能够优雅地处理各种异常情况，提高代码的健壮性和可靠性。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Helper 组件。

## 许可证

Helper 组件遵循 MIT 许可证。