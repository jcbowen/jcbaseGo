# Timezone 组件

「数据库存 UTC、前端展示北京时间」约定下的通用时区工具，提供时间格式化与解析能力，并通过 `Formatter` 支持调用方注入业务专属的时间字段集合，避免全局可变状态。

## 概述

多个中国区项目共用同一约定：时间以 UTC 写入数据库，对外展示时转换为北京时间（Asia/Shanghai）。Timezone 组件封装该约定下的全套操作：

- **约定统一**：写库用 `NowUTCString`，展示用 `*ToShanghai` / `FormatShanghai`，全项目一套口径
- **解析容错**：优先按 UTC 解析，失败时回退服务器本地时区，兼容历史遗留的本地时间数据
- **无全局可变状态**：时间字段集合由 `Formatter` 实例持有，按实例隔离，多业务线互不干扰
- **零依赖降级**：系统缺失 tzdata 时自动回退为固定东八区偏移，行为不变

## 功能特性

### 常量与变量

- `DateTimeFormat`：标准格式 `2006-01-02 15:04:05`
- `Shanghai`：北京时间时区，`time.LoadLocation` 失败时回退 `time.FixedZone("Asia/Shanghai", 8*3600)`

### 写库（UTC）

- `NowUTCString()`：当前 UTC 时间字符串，直接写入数据库时间字段

### 展示（北京时间）

- `UTCStringToShanghai(s)`：UTC 字符串转北京时间字符串；空值或解析失败原样返回
- `FormatShanghai(t)`：`time.Time` 格式化为北京时间字符串；零值返回空字符串
- `Formatter.ToShanghaiString(s)` / `Formatter.FormatShanghai(t)`：同上，时区可实例级覆盖

### 解析

- `ParseUTCString(s)`：按内置 layout 列表依次尝试，全部失败返回错误
- 内置 layout 覆盖 RFC3339（含 Nano）、标准日期时间（含毫秒/微秒）、`2006-01-02` 等常见格式

### Formatter（实例级配置）

- 时间字段集合：`AddTimeFields` / `IsTimeField`，判定某字段是否需时区格式化
- 时区覆盖：`Location` 字段，未设置时回退 `Shanghai`
- 解析 layout 覆盖：`ParseLayouts` 字段，未设置时用内置默认值
- `FormatValue(val)`：统一处理 `time.Time` 或时间字符串，供响应序列化的 `FormatTime` 回调注入

## 安装指南

### 依赖要求

- Go 1.26.0+

### 安装依赖

组件已包含在 jcbaseGo 模块中，无需额外安装：

```bash
go get github.com/jcbowen/jcbaseGo
```

## 使用示例

### 写库与展示的基本流

```go
package main

import (
    "fmt"

    "github.com/jcbowen/jcbaseGo/component/timezone"
)

func main() {
    // 写库：UTC 字符串
    utcStr := timezone.NowUTCString()
    fmt.Println(utcStr) // 例如: 2026-10-21 00:30:00（UTC）

    // 展示：转北京时间
    fmt.Println(timezone.UTCStringToShanghai(utcStr)) // 例如: 2026-10-21 08:30:00

    // time.Time 直接格式化
    fmt.Println(timezone.FormatShanghai(time.Now())) // 例如: 2026-10-21 08:30:00
}
```

### 解析时间字符串

```go
t, err := timezone.ParseUTCString("2026-10-21 08:30:00")
if err != nil {
    // 所有 layout 均解析失败
    return
}
// 解析结果按 UTC 语义使用；对历史本地时间数据会自动回退本地时区解析
```

### Formatter：注入业务时间字段

```go
formatter := timezone.NewFormatter("created_at", "updated_at", "pay_time")

// 判定字段是否走时区格式化
if formatter.IsTimeField("created_at") {
    // ...
}

// 单个值处理：time.Time 或字符串均可
if s, ok := formatter.FormatValue("2026-10-21 00:30:00"); ok {
    fmt.Println(s) // 2026-10-21 08:30:00
}
```

### 与 serializer 组件联动

`Formatter.FormatValue` 可直接作为 `serializer.Options.FormatTime` 注入，实现响应数据的时间字段统一转北京时间：

```go
import (
    "github.com/jcbowen/jcbaseGo/component/serializer"
    "github.com/jcbowen/jcbaseGo/component/timezone"
)

formatter := timezone.NewFormatter("created_at", "updated_at")

data := serializer.Process(respData,
    serializer.WithTimeField(formatter.IsTimeField),
    serializer.WithFormatTime(formatter.FormatValue),
)
```

## 详细功能说明

### 解析优先级

`ParseUTCString` 与 `Formatter.parseString` 均采用两轮策略：

1. 第一轮：所有 layout 按 **UTC** 解析（符合 UTC 存库约定）
2. 第二轮：所有 layout 按 **服务器本地时区** 解析（兼容历史遗留本地时间数据）

### 零值与空值语义

| 输入 | 处理 |
|---|---|
| 空字符串 | 原样返回空串 |
| `0000-00-00 00:00:00` | `FormatValue` 视为无效，返回空串且 `ok=false` |
| `time.Time` 零值 | 返回空串 |
| 解析失败的字符串 | 原样返回（不覆盖非标准字段） |

### Formatter 的隔离性

每个 `Formatter` 实例独立持有时间字段集合，`AddTimeFields` 只影响当前实例。不同业务线（如订单、IM）可各自创建 Formatter，避免全局注册表带来的耦合。

## API 参考

### 包级函数

- `NowUTCString() string` - 当前 UTC 时间字符串（写库用）
- `UTCStringToShanghai(s string) string` - UTC 字符串转北京时间；空/解析失败原样返回
- `FormatShanghai(t time.Time) string` - time.Time 格式化为北京时间；零值返回空串
- `ParseUTCString(s string) (time.Time, error)` - 解析时间字符串，优先 UTC 后本地时区

### Formatter

- `NewFormatter(extraFields ...string) *Formatter` - 创建格式化器，可追加业务时间字段
- `AddTimeFields(keys ...string)` - 追加时间字段名（仅当前实例）
- `IsTimeField(key string) bool` - 判断字段是否需时区格式化
- `ToShanghaiString(s string) string` - UTC 字符串转当前实例时区字符串
- `FormatShanghai(t time.Time) string` - time.Time 格式化为当前实例时区字符串
- `FormatValue(val interface{}) (string, bool)` - 统一处理 time.Time/字符串，返回 (北京时间字符串, 是否成功)

### 可配置字段

```go
type Formatter struct {
    Location     *time.Location // 时区，未设置回退 Shanghai
    ParseLayouts []string       // 解析 layout，未设置用内置默认值
    // timeFields 为私有字段，经 AddTimeFields / IsTimeField 操作
}
```

## 常见问题

### 为什么展示时间比库里时间多 8 小时？

库中存的是 UTC，展示按约定转北京时间（UTC+8），这是预期行为。若前端直接展示库中原始值出现偏差，检查是否漏走了 `ToShanghaiString` / `FormatValue`。

### 历史数据是本地时间怎么办？

解析函数第一轮按 UTC、第二轮按本地时区，新旧数据都能正确解析；展示侧统一输出北京时间，用户无感知。

### 容器里没有 tzdata 会影响吗？

不会。`Shanghai` 初始化失败时自动回退为固定东八区偏移（`time.FixedZone`），格式化与解析行为一致。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个组件。

## 许可证

MIT License
