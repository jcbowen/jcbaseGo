# Pushplus 组件

基于标准库 `net/http` 的 PushPlus 推送薄封装。不引入第三方 SDK，只负责三件事：合并默认配置、发送 `/send` 请求、把错误分成可判定的几类。

## 概述

PushPlus 的发送能力只有一个 `POST /send` 接口（token、title、content、template、channel、topic、webhook、callbackUrl），因此本组件刻意保持轻量：

- **默认值回落**：客户端持有 token、群组、渠道、模板、超时，单条消息可按需覆盖
- **错误分层**：HTTP 异常、响应解析失败、业务 `code` 非 200 三类错误可区分
- **上下文友好**：提供 `SendContext`，可直接传入 `*gin.Context.Request.Context()`
- **资源可控**：响应体最多读取 1MB，错误信息只携带前 512 字节响应片段
- **并发安全**：客户端创建后字段只读，可被多个 goroutine 复用
- **零新增依赖**：仅使用标准库

## 快速开始

```go
package main

import (
    "errors"
    "log"

    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/pushplus"
)

func main() {
    client := pushplus.New(jcbaseGo.PushPlusStruct{
        Token: "你的用户令牌",
        Topic: "ops", // 可选：群组编码
    })

    resp, err := client.Send(pushplus.Message{
        Title:   "部署完成",
        Content: "v1.0.0 已上线",
    })
    if err != nil {
        var apiErr *pushplus.APIError
        if errors.As(err, &apiErr) {
            log.Printf("推送被拒绝: code=%d msg=%s", apiErr.Code, apiErr.Msg)
            return
        }
        log.Printf("推送失败: %v", err)
        return
    }

    // data 是消息流水号，可用于后续查询
    log.Printf("已受理，流水号=%s", resp.Data)
}
```

在 HTTP 请求中使用调用方的上下文：

```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
defer cancel()

_, err := client.SendContext(ctx, pushplus.Message{
    Title:   "任务异常",
    Content: "队列积压超过阈值",
})
```

## 配置

`jcbaseGo.PushPlusStruct` 字段说明：

| 字段 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `Token` | string | 无（必填） | PushPlus 用户令牌 |
| `Topic` | string | 空 | 默认群组编码，为空表示一对一推送 |
| `Channel` | string | `wechat` | 默认发送渠道 |
| `Template` | string | `html` | 默认消息模板 |
| `Timeout` | time.Duration | `10s` | HTTP 请求超时 |
| `BaseURL` | string | `https://www.pushplus.plus/send` | 发送接口地址 |

> `Timeout` 在 JSON 配置中按纳秒整数填写（如 `10000000000` 表示 10s），INI 中可直接写 `10s`。
> 不配置时由 `New` 使用默认值 10s。

JSON 配置示例：

```json
{
  "pushplus": {
    "token": "你的用户令牌",
    "topic": "ops",
    "channel": "wechat",
    "template": "html"
  }
}
```

INI 配置示例：

```ini
[pushplus]
token = 你的用户令牌
topic = ops
channel = wechat
template = html
```

也可以通过 `DefaultConfigStruct` 使用：

```go
cfg := jcbaseGo.DefaultConfigStruct{
    PushPlus: jcbaseGo.PushPlusStruct{Token: "你的用户令牌"},
}
client := pushplus.New(cfg.PushPlus)
```

## 消息字段

`pushplus.Message` 中为空的值会回落到客户端默认值：

| 字段 | 说明 |
| --- | --- |
| `Title` | 消息标题 |
| `Content` | 消息正文，格式由 `Template` 决定 |
| `Topic` | 群组编码，覆盖客户端默认群组 |
| `Channel` | 发送渠道，覆盖客户端默认渠道 |
| `Template` | 消息模板，覆盖客户端默认模板 |
| `Webhook` | webhook 渠道的接收地址 |
| `CallbackURL` | 推送结果回调地址 |

### 渠道常量

`ChannelWeChat`、`ChannelWebhook`、`ChannelCP`、`ChannelMail`、`ChannelSMS`、`ChannelVoice`。

### 模板常量

`TemplateHTML`、`TemplateJSON`、`TemplateMarkdown`、`TemplateTxt`。

## 返回结构与错误处理

```go
type Response struct {
    Code int    // 业务状态码，200 表示请求已受理
    Msg  string // 业务提示信息
    Data string // 消息流水号
}
```

错误分三类，均可用 `errors` 包判断：

1. **请求层错误**：网络失败、超时、上下文取消，错误链中包含原始原因（`errors.Is(err, context.Canceled)` 可用）
2. **响应层错误**：HTTP 状态码非 2xx、响应不是合法 JSON
3. **业务层错误**：HTTP 成功但 `code` 非 200，返回 `*pushplus.APIError`，同时返回 `*Response` 便于读取 `Msg`

```go
resp, err := client.Send(msg)

var apiErr *pushplus.APIError
switch {
case errors.As(err, &apiErr):
    // PushPlus 明确拒绝：token 无效、内容超长、频率超限等
case err != nil:
    // 网络/解析/HTTP 异常
default:
    // 已受理，resp.Data 为流水号
}
```

## 重要语义：受理不等于送达

- `code = 200` 只代表 PushPlus **已受理**请求，实际投递是异步的，不保证已送达用户
- `Data` 是消息流水号，需要确认送达时应结合 PushPlus 的查询接口或 `CallbackURL` 回调
- 组件**不做自动重试**：推送属于非关键路径，自动重试容易造成重复推送；需要重试时请由调用方自行包装并控制幂等

## 安全说明

- `Token` 属于敏感配置，不要硬编码进代码或提交到仓库
- debugger 组件的默认请求体掩码字段集合已包含 `token`，调试面板中会自动脱敏
- 默认使用 HTTPS 接口地址；如确需私有代理，可通过 `BaseURL` 覆盖

## 测试

```bash
go test ./component/pushplus/
```

测试全部基于 `httptest`，不会请求真实 PushPlus 接口，也不依赖网络。
