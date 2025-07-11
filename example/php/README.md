# PHP 组件使用文档

## 功能介绍

PHP 组件允许在 Go 程序中调用 PHP 函数，通过生成一个 PHP 脚本文件来执行 PHP 代码。这个组件特别适用于需要在 Go 项目中集成现有 PHP 功能或调用 PHP 库的场景。

## 主要特性

- **函数调用**：支持调用任意 PHP 函数
- **参数传递**：支持多种数据类型（字符串、数字、布尔值、JSON 对象）
- **自动类型转换**：自动将参数转换为合适的 PHP 类型
- **错误处理**：完善的错误处理和返回值处理
- **命令行接口**：提供完整的命令行参数解析

## 安装要求

### 系统要求
- Go 1.16 或更高版本
- PHP 7.0 或更高版本（需要支持命令行执行）
- 确保 `php` 命令在系统 PATH 中可用

### 依赖安装
```bash
# 检查 PHP 是否可用
php --version

# 如果 PHP 未安装，请根据系统安装
# macOS
brew install php

# Ubuntu/Debian
sudo apt-get install php-cli

# CentOS/RHEL
sudo yum install php-cli
```

## 基本使用

### 1. 导入组件
```go
import (
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/php"
)
```

### 2. 创建 PHP 组件实例
```go
// 使用默认配置
phpComponent := php.New(jcbaseGo.Option{})

// 或指定配置路径
phpComponent := php.New(jcbaseGo.Option{
    RuntimePath: "/path/to/runtime",
    ConfigSource: "/path/to/config",
})
```

### 3. 调用 PHP 函数
```go
// 调用无参数的 PHP 函数
result, err := phpComponent.RunFunc("phpinfo")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result)

// 调用带参数的 PHP 函数
result, err := phpComponent.RunFunc("strtoupper", "hello world")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // 输出: HELLO WORLD
```

## 参数类型支持

### 字符串参数
```go
result, err := phpComponent.RunFunc("strlen", "hello")
// 返回: 5
```

### 数字参数
```go
result, err := phpComponent.RunFunc("pow", "2", "3")
// 返回: 8
```

### 布尔值参数
```go
result, err := phpComponent.RunFunc("var_dump", "true")
// 返回: bool(true)
```

### JSON 对象参数
```go
result, err := phpComponent.RunFunc("json_encode", `{"name":"张三","age":25}`)
// 返回: {"name":"张三","age":25}
```

### PHP 序列化数据参数
```go
// 序列化数据
result, err := phpComponent.RunFunc("serialize", `{"name":"张三","age":25}`)
// 返回: a:2:{s:4:"name";s:6:"张三";s:3:"age";i:25;}

// 反序列化数据
result, err := phpComponent.RunFunc("unserialize", "a:2:{s:4:\"name\";s:6:\"张三\";s:3:\"age\";i:25;}")
// 返回: {"name":"张三","age":25}
```

## 高级用法

### 调用自定义 PHP 函数
```go
// 首先需要定义 PHP 函数
phpCode := `
function calculateSum($a, $b) {
    return $a + $b;
}

function formatName($firstName, $lastName) {
    return ucfirst($firstName) . ' ' . ucfirst($lastName);
}
`

// 将 PHP 代码写入文件或通过其他方式加载
// 然后调用函数
result, err := phpComponent.RunFunc("calculateSum", "10", "20")
// 返回: 30

result, err := phpComponent.RunFunc("formatName", "john", "doe")
// 返回: John Doe
```

### 处理复杂数据结构
```go
// 传递 JSON 数组
result, err := phpComponent.RunFunc("json_encode", `[1,2,3,4,5]`)
// 返回: [1,2,3,4,5]

// 传递嵌套对象
nestedObj := `{"user":{"name":"张三","profile":{"age":25,"city":"北京"}}}`
result, err := phpComponent.RunFunc("json_decode", nestedObj, "true")

// 使用 PHP 序列化处理复杂数据
complexData := `{"users":[{"id":1,"name":"用户1","active":true},{"id":2,"name":"用户2","active":false}]}`
serialized, err := phpComponent.RunFunc("serialize", complexData)
// 返回: a:1:{s:5:"users";a:2:{i:0;a:3:{s:2:"id";i:1;s:4:"name";s:7:"用户1";s:6:"active";b:1;}i:1;a:3:{s:2:"id";i:2;s:4:"name";s:7:"用户2";s:6:"active";b:0;}}}

// 反序列化 PHP 数据
unserialized, err := phpComponent.RunFunc("unserialize", serialized)
// 返回: {"users":[{"id":1,"name":"用户1","active":true},{"id":2,"name":"用户2","active":false}]}
```

## PHP 序列化和反序列化

PHP 组件支持使用 PHP 的 `serialize()` 和 `unserialize()` 函数进行数据序列化和反序列化。这对于处理 PHP 特定的数据格式或与现有 PHP 系统集成非常有用。

### 基本序列化
```go
// 序列化简单数组
result, err := phpComponent.RunFunc("serialize", `["apple","banana","orange"]`)
// 返回: a:3:{i:0;s:5:"apple";i:1;s:6:"banana";i:2;s:6:"orange";}

// 序列化关联数组
result, err := phpComponent.RunFunc("serialize", `{"name":"张三","age":25}`)
// 返回: a:2:{s:4:"name";s:6:"张三";s:3:"age";i:25;}
```

### 基本反序列化
```go
// 反序列化数组
serializedData := "a:3:{i:0;s:5:\"apple\";i:1;s:6:\"banana\";i:2;s:6:\"orange\";}"
result, err := phpComponent.RunFunc("unserialize", serializedData)
// 返回: ["apple","banana","orange"]

// 反序列化关联数组
serializedAssoc := "a:2:{s:4:\"name\";s:6:\"张三\";s:3:\"age\";i:25;}"
result, err := phpComponent.RunFunc("unserialize", serializedAssoc)
// 返回: {"name":"张三","age":25}
```

### 支持的数据类型

#### 字符串
```go
result, err := phpComponent.RunFunc("serialize", "Hello World")
// 返回: s:11:"Hello World";
```

#### 数字
```go
result, err := phpComponent.RunFunc("serialize", "123.45")
// 返回: d:123.45;
```

#### 布尔值
```go
result, err := phpComponent.RunFunc("serialize", "true")
// 返回: b:1;
```

#### 空值
```go
result, err := phpComponent.RunFunc("serialize", "null")
// 返回: N;
```

### 复杂数据结构
```go
// 嵌套数组
nestedData := `{"user":{"name":"李四","profile":{"age":30,"hobbies":["读书","游泳"]}}}`
serialized, err := phpComponent.RunFunc("serialize", nestedData)
// 返回: a:1:{s:4:"user";a:2:{s:4:"name";s:6:"李四";s:7:"profile";a:2:{s:3:"age";i:30;s:7:"hobbies";a:2:{i:0;s:6:"读书";i:1;s:6:"游泳";}}}}

// 反序列化
unserialized, err := phpComponent.RunFunc("unserialize", serialized)
// 返回: {"user":{"name":"李四","profile":{"age":30,"hobbies":["读书","游泳"]}}}
```

### 错误处理
```go
// 尝试反序列化无效数据
invalidData := "a:1:{i:0;s:5:\"hello\""
result, err := phpComponent.RunFunc("unserialize", invalidData)
if err != nil {
    fmt.Printf("反序列化失败: %v\n", err)
}
```

### 与 JSON 的对比
```go
testData := `{"name":"王五","age":28,"hobbies":["音乐","电影"]}`

// PHP 序列化
serialized, err := phpComponent.RunFunc("serialize", testData)
// 返回: a:3:{s:4:"name";s:6:"王五";s:3:"age";i:28;s:7:"hobbies";a:2:{i:0;s:6:"音乐";i:1;s:6:"电影";}}

// JSON 序列化
jsonResult, err := phpComponent.RunFunc("json_encode", testData)
// 返回: {"name":"王五","age":28,"hobbies":["音乐","电影"]}
```

### 使用场景

1. **与现有 PHP 系统集成**: 处理 PHP 应用程序生成的序列化数据
2. **数据存储**: 将复杂数据结构序列化后存储到数据库或文件
3. **缓存**: 使用序列化数据作为缓存内容
4. **数据传输**: 在 PHP 和 Go 系统之间传输数据

### 注意事项

1. **安全性**: 反序列化用户输入的数据可能存在安全风险，请谨慎处理
2. **兼容性**: PHP 序列化格式是 PHP 特有的，不适用于其他语言
3. **性能**: 对于简单数据，JSON 可能比 PHP 序列化更高效
4. **调试**: 序列化数据不易读，建议在开发时使用 JSON 格式

## 序列化辅助工具

为了简化序列化和反序列化的使用，我们提供了一个辅助工具类 `SerializeHelper`，它封装了常用的序列化操作。

### 基本用法
```go
// 创建辅助工具实例
helper := NewSerializeHelper()

// 序列化数据
serialized, err := helper.Serialize(`{"name":"张三","age":25}`)
if err != nil {
    log.Printf("序列化失败: %v", err)
}

// 反序列化数据
unserialized, err := helper.Unserialize(serialized)
if err != nil {
    log.Printf("反序列化失败: %v", err)
}
```

### 高级功能

#### 数组序列化
```go
arrayData := []string{"apple", "banana", "orange"}
serialized, err := helper.SerializeArray(arrayData)
```

#### 映射序列化
```go
mapData := map[string]interface{}{
    "name":   "李四",
    "age":    30,
    "city":   "上海",
    "active": true,
}
serialized, err := helper.SerializeMap(mapData)
```

#### JSON 转换
```go
// JSON 转序列化
jsonData := `{"name":"王五","age":28}`
serialized, err := helper.ConvertJSONToSerialized(jsonData)

// 序列化转 JSON
jsonResult, err := helper.ConvertSerializedToJSON(serialized)
```

#### 数据验证
```go
// 检查是否为有效的序列化数据
isValid := helper.IsValidSerializedData(serializedData)
```

### 辅助工具的优势

1. **简化操作**: 提供更直观的 API 接口
2. **类型安全**: 支持 Go 原生数据类型
3. **错误处理**: 统一的错误处理机制
4. **数据验证**: 内置数据有效性检查
5. **格式转换**: 支持 JSON 和序列化格式互转

## 错误处理

### 函数不存在错误
```go
result, err := phpComponent.RunFunc("non_existent_function")
if err != nil {
    // 错误信息: fatal:Function non_existent_function not exists
    log.Printf("PHP 函数调用失败: %v", err)
}
```

### 参数错误
```go
// 传递错误类型的参数
result, err := phpComponent.RunFunc("strlen", "123", "extra_param")
if err != nil {
    log.Printf("参数错误: %v", err)
}
```

## 最佳实践

### 1. 函数命名规范
- 使用有意义的函数名
- 避免使用 PHP 内置函数名作为自定义函数名
- 函数名使用小写字母和下划线

### 2. 参数处理
- 在调用前验证参数类型和数量
- 使用 JSON 格式传递复杂数据结构
- 注意字符串转义问题

### 3. 性能优化
- 避免频繁调用相同的 PHP 函数
- 考虑使用缓存机制
- 合理使用 PHP 函数，避免过度依赖

### 4. 安全性
- 不要直接执行用户输入的 PHP 代码
- 验证所有输入参数
- 限制可调用的 PHP 函数范围

## 示例项目

查看 `example/php/` 目录下的完整示例：

- [基本用法示例](basic/main.go)
- [字符串处理示例](string/main.go)
- [数学计算示例](math/main.go)
- [JSON 处理示例](json/main.go)
- [序列化和反序列化示例](serialize/main.go)
- [自定义函数示例](custom/main.go)
- [序列化辅助工具示例](utils_example/main.go)

## 注意事项

1. **PHP 环境**：确保 PHP 命令行环境正确配置
2. **文件权限**：确保生成的 PHP 文件有执行权限
3. **路径问题**：注意文件路径的正确性，特别是在不同操作系统上
4. **内存使用**：大量调用 PHP 函数可能影响性能
5. **错误日志**：建议启用 PHP 错误日志以便调试

## 故障排除

### 常见问题

1. **PHP 命令未找到**
   ```bash
   # 检查 PHP 是否在 PATH 中
   which php
   
   # 如果未找到，添加 PHP 到 PATH
   export PATH=$PATH:/path/to/php/bin
   ```

2. **权限错误**
   ```bash
   # 给生成的 PHP 文件添加执行权限
   chmod +x /tmp/php/main.go
   ```

3. **函数调用失败**
   - 检查函数名是否正确
   - 确认函数已定义或为 PHP 内置函数
   - 验证参数类型和数量

4. **编码问题**
   - 确保 PHP 文件使用 UTF-8 编码
   - 检查中文字符的正确显示

## 更新日志

- **v1.0.0**: 初始版本，支持基本的 PHP 函数调用
- 支持多种参数类型
- 完善的错误处理机制
- 命令行参数解析 