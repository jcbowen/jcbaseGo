# Serializer 组件

响应数据的递归序列化组件：将结构体、切片、map、指针等嵌套数据转换为统一的 map/slice 结构，并在转换过程中按需格式化时间字段、加密 uint64 类型的 ID 字段。内置最大深度限制与循环引用检测，避免自引用结构体或极深嵌套导致栈溢出。

## 概述

Serializer 组件解决 API 响应数据的统一后处理问题，业务策略全部通过回调注入，组件本身不含任何业务假设：

- **结构统一**：任意嵌套结构递归转换为 `map[string]interface{}` / `[]interface{}`，输出形态稳定
- **策略注入**：时间字段判定与格式化、ID 字段判定与加密均由调用方以函数选项注入
- **安全递归**：最大深度（默认 32）+ 指针级循环引用检测，DAG 共享节点直接复用，不会栈溢出
- **向后兼容**：`LegacyMode` 保留重构前的 `json.Marshal/Unmarshal` 行为，已上线项目可零差异切换

## 功能特性

### 处理入口

- `Process(data, opts...)`：整体递归处理，支持 map、slice、struct、指针及任意嵌套
- `ProcessValue(key, val, opts...)`：按字段名处理单个值，先判时间字段再判 ID 字段

### 函数式选项

| 选项 | 说明 |
|---|---|
| `WithEncryptID(fn)` | 注入 ID 加密函数（如把 uint64 ID 转为密文字符串） |
| `WithIDField(fn)` | 注入 ID 字段判定函数，默认不加密 |
| `WithTimeField(fn)` | 注入时间字段判定函数，默认不处理时间字段 |
| `WithFormatTime(fn)` | 注入时间值格式化函数，返回 (字符串, 是否成功) |
| `WithSkipField(fn)` | 注入字段跳过判定（如跳过匿名嵌入的 orm base model） |
| `WithMaxDepth(depth)` | 递归最大深度，<= 0 时用默认值 32 |
| `WithLegacyMode(enabled)` | 开启旧版序列化行为，用于向后兼容 |

### 转换规则

- 按 `json` tag 取字段名；`json:"-"` 跳过；无 tag 用字段名；未导出字段跳过
- 匿名嵌入字段经 `SkipField` 判定后可整体跳过（典型场景：跳过 `base.MysqlBaseModel`）
- nil 指针输出 `null`；非 nil 指针保留指针参与递归（保证循环引用检测命中同一地址）
- uint64 及 uint64 命名类型（如 `utils.EncryptedID`）在 `IsIDField` 命中时统一走 `EncryptID`

### 安全机制

- **深度限制**：超过 `MaxDepth` 的子树直接返回原始数据，避免子树数据缺失以外的异常
- **循环引用检测**：以指针地址做 DFS 状态标记，检测到环时对应位置返回 `nil` 打破循环；已完成的共享节点（DAG）直接返回原值，不重复展开

## 安装指南

### 依赖要求

- Go 1.26.0+

### 安装依赖

组件已包含在 jcbaseGo 模块中，无需额外安装：

```bash
go get github.com/jcbowen/jcbaseGo
```

## 使用示例

### 基本序列化

```go
package main

import (
    "fmt"

    "github.com/jcbowen/jcbaseGo/component/serializer"
)

type User struct {
    ID       uint64 `json:"id"`
    Username string `json:"username"`
}

func main() {
    user := &User{ID: 123, Username: "张三"}
    result := serializer.Process(user)
    fmt.Println(result)
    // map[id:123 username:张三]
}
```

### 时间格式化 + ID 加密

```go
import (
    "github.com/jcbowen/jcbaseGo/component/serializer"
    "github.com/jcbowen/jcbaseGo/component/timezone"
)

formatter := timezone.NewFormatter("created_at", "updated_at")

data := serializer.Process(resp,
    serializer.WithTimeField(formatter.IsTimeField),
    serializer.WithFormatTime(formatter.FormatValue),
    serializer.WithIDField(func(key string) bool { return key == "id" }),
    serializer.WithEncryptID(func(id uint64) string { return utils.EncryptID(id) }),
)
// created_at/updated_at 转为北京时间字符串，id 加密为字符串
```

### 跳过匿名嵌入的 base model

```go
import (
    "reflect"
    "strings"

    "github.com/jcbowen/jcbaseGo/component/serializer"
)

data := serializer.Process(list,
    serializer.WithSkipField(func(f reflect.StructField) bool {
        return f.Anonymous && strings.HasSuffix(f.Type.String(), "MysqlBaseModel")
    }),
)
```

### 单字段处理

```go
val := serializer.ProcessValue("created_at", "2026-10-21 00:30:00",
    serializer.WithTimeField(func(key string) bool { return key == "created_at" }),
    serializer.WithFormatTime(func(v interface{}) (string, bool) {
        return formatter.FormatValue(v)
    }),
)
```

### 旧版兼容模式

```go
// 保持与重构前一致的输出行为（结构体走 json 往返，切片不递归元素，map 原样返回）
data := serializer.Process(resp, serializer.WithLegacyMode(true))
```

## 详细功能说明

### 与 Gin 的配合

`Process` 对 `gin.H`（即 `map[string]interface{}`）有快速路径，可直接处理控制器构造的响应数据；LegacyMode 下 `gin.H` 与 `map[string]interface{}` 原样返回，与历史行为一致。

### 循环引用的输出语义

| 场景 | 输出 |
|---|---|
| 自引用环（A → B → A） | 环的闭合点返回 `nil`，JSON 序列化为 `null` |
| DAG 共享节点（A → C ← B） | 后访问到的分支直接复用已处理结果 |
| 超过 MaxDepth 的子树 | 返回该子树的原始值（不再转换） |

### 深度语义

`MaxDepth` 计量的是递归处理层数，默认 32；每进入一层容器/结构体/指针解引用计数 +1。正常业务的嵌套深度远小于该值，无需调整；调整仅影响极端深数据的保护阈值。

### LegacyMode 与新版差异

| 行为 | 新版（默认） | LegacyMode |
|---|---|---|
| 结构体 | 反射逐字段转 map，可注入字段策略 | `json.Marshal/Unmarshal` 往返 |
| 切片 | 递归处理每个元素 | 仅转 `[]interface{}`，不递归元素 |
| map | 递归处理每个值 | `gin.H`/`map[string]interface{}` 原样返回，其他类型 json 往返 |
| 时间/ID 策略 | 生效 | 不生效 |

## API 参考

### 处理函数

- `Process(data interface{}, opts ...Option) interface{}` - 递归处理数据
- `ProcessValue(key string, val interface{}, opts ...Option) interface{}` - 按字段名处理单个值

### 选项类型与构造函数

```go
type Options struct {
    EncryptID   func(uint64) string
    IsIDField   func(string) bool
    IsTimeField func(string) bool
    FormatTime  func(interface{}) (string, bool)
    SkipField   func(reflect.StructField) bool
    MaxDepth    int
    LegacyMode  bool
}

type Option func(*Options)

func WithEncryptID(fn func(uint64) string) Option
func WithIDField(fn func(string) bool) Option
func WithTimeField(fn func(string) bool) Option
func WithFormatTime(fn func(interface{}) (string, bool)) Option
func WithSkipField(fn func(reflect.StructField) bool) Option
func WithMaxDepth(depth int) Option
func WithLegacyMode(enabled bool) Option
```

## 常见问题

### ID 加密后类型变了，前端怎么处理？

加密函数返回字符串，`IsIDField` 命中的字段在输出中即为字符串。前后端约定以「ID 字段一律按字符串处理」最稳妥，避免 JS 大整数精度丢失。

### 时间字段没有被转换？

确认同时注入了 `WithTimeField`（判定哪些字段是时间字段）和 `WithFormatTime`（提供格式化实现），两者缺一不生效。

### 为什么有的字段消失了？

三种可能：`json:"-"` 跳过、未导出字段跳过、匿名嵌入字段被 `SkipField` 命中。检查模型 tag 与注入的跳过策略。

### LegacyMode 还有必要开吗？

仅用于已上线且依赖旧输出形态的项目做无损切换；新项目请使用默认模式，可获得时间/ID 策略与循环引用保护。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个组件。

## 许可证

MIT License
