# Ratelimit 组件

基于 Redis 有序集合（ZSET）的滑动窗口限流组件，内置本地内存降级能力。核心算法与业务解耦：限流 key 由调用方自行构造，Redis 客户端由调用方注入。

## 概述

Ratelimit 组件为 HTTP API、消息推送等场景提供限流保护：

- **滑动窗口**：基于 Redis ZSET 的分数窗口，窗口边界平滑过渡，无固定窗口的临界突刺问题
- **自动降级**：Redis 不可用或执行失败时，自动切换本地内存限流，保证限流不失效（退化为单实例计数）
- **资源可控**：本地降级限流有条目容量硬上限与后台过期回收，不会无限占内存
- **生命周期明确**：`Close` 停止后台清理协程，多次调用安全，便于测试与优雅退出

## 功能特性

### Redis 滑动窗口

- 单次 pipeline 完成「清理窗口外旧记录 → 写入当前请求 → 刷新 key 过期时间 → 统计窗口内计数」
- member 为 `毫秒时间戳:实例标识:自增序号`，同一毫秒内的多次请求不会被 ZSET 去重覆盖
- key 按窗口长度自动 `Expire`，无需额外清理任务

### 本地内存降级

- `client` 传 nil 或 Redis 执行失败时，自动走本地滑动窗口
- 本地限流仅在当前进程内生效，**多实例不共享计数**，适用于「有限流总比没限流好」的兜底场景
- 每条目独立互斥锁 + 映射级读写锁，并发安全
- 后台协程每分钟回收过期条目；条目数超过 10000 时淘汰最早过期的条目，作为内存增长硬上限

### 判定语义

- `Allow` 返回 `true` 表示**允许通过**（含当前请求），`false` 表示触发限流
- 窗口内计数 `<= maxRequests` 时放行，第 `maxRequests + 1` 个请求被拦截

## 安装指南

### 依赖要求

- Go 1.26.0+
- Redis（可选，不传客户端时纯本地限流）

### 安装依赖

组件已包含在 jcbaseGo 模块中，Redis 客户端使用 `github.com/go-redis/redis/v8`：

```bash
go get github.com/jcbowen/jcbaseGo
```

## 使用示例

### 基本使用（Redis 限流）

```go
package main

import (
    "context"
    "fmt"

    "github.com/go-redis/redis/v8"
    "github.com/jcbowen/jcbaseGo/component/ratelimit"
)

func main() {
    rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

    limiter := ratelimit.NewLimiter(rdb)
    defer limiter.Close() // 释放后台清理协程

    ctx := context.Background()
    // 每分钟最多 10 次
    if limiter.Allow(ctx, "api:create", 10, 60) {
        fmt.Println("允许通过")
    } else {
        fmt.Println("触发限流")
    }
}
```

### 纯本地限流

不传 Redis 客户端（或传 nil），全部走本地内存限流：

```go
limiter := ratelimit.NewLimiter(nil)
defer limiter.Close()

if limiter.AllowLocal("api:create", 10, 60) {
    // 单实例内每分钟最多 10 次
}
```

### 按业务维度构造 key

key 由调用方自行构造，常见维度：

```go
// 按用户 + 接口限流：每用户每分钟 5 次
key := fmt.Sprintf("ratelimit:create:%d", userID)
limiter.Allow(ctx, key, 5, 60)

// 按 IP + 接口限流：每 IP 每小时 100 次
key := fmt.Sprintf("ratelimit:login:%s", clientIP)
limiter.Allow(ctx, key, 100, 3600)
```

### Gin 中间件中的用法

```go
r.POST("/api/create", func(c *gin.Context) {
    key := "ratelimit:create:" + strconv.FormatInt(userID, 10)
    if !limiter.Allow(c.Request.Context(), key, 10, 60) {
        c.JSON(429, gin.H{"message": "请求过于频繁，请稍后再试"})
        c.Abort()
        return
    }
    // 正常业务逻辑
})
```

## 详细功能说明

### Redis 执行流程

一次 `Allow` 调用在 Redis 侧依次执行（同一 pipeline，一次往返）：

1. `ZRemRangeByScore(key, 0, now - window)`：清理窗口外的旧记录
2. `ZAdd(key, now, member)`：写入当前请求
3. `Expire(key, window)`：刷新 key 过期时间，与窗口一致
4. `ZCard(key)`：统计窗口内计数

第 4 步返回的计数包含当前请求，因此判定为 `count <= maxRequests` 放行。

### 降级触发条件

以下任一情况发生即自动降级到本地限流（对调用方透明，仍返回布尔值）：

- 创建时传入的 `client` 为 nil
- pipeline 执行返回错误（网络分区、Redis 超时等）
- 返回结果数量不足或类型断言失败

### 本地限流的回收机制

- **过期回收**：后台协程每 60 秒扫描一次，删除 `expiresAt` 早于当前时间的条目；持续活跃的条目每次命中都会刷新 `expiresAt`，不会被误删
- **容量淘汰**：条目数超过 10000 时，立即淘汰一个最早过期的条目，保证内存有硬上限

## API 参考

### 构造函数

- `NewLimiter(client redis.UniversalClient) *Limiter` - 创建限流器；`client` 传 nil 时纯本地限流

### 实例方法

- `Allow(ctx context.Context, key string, maxRequests int, windowSeconds int) bool` - 滑动窗口限流判定，Redis 失败自动降级本地
- `AllowLocal(key string, maxRequests, windowSeconds int) bool` - 本地内存限流判定（不需要 ctx）
- `Close()` - 停止后台清理协程，多次调用安全

### 内部常量

| 名称 | 值 | 说明 |
|---|---|---|
| `localFallbackMaxEntries` | 10000 | 本地降级限流的条目数硬上限 |
| `localFallbackSweepInterval` | 1 分钟 | 后台过期条目回收间隔 |

## 常见问题

### 降级后限流阈值会放大吗？

会。Redis 模式下多实例共享计数；降级后每个实例各自计数，总放行量最多为 `maxRequests × 实例数`。这是降级换可用性的固有代价，关键接口建议适当下调阈值。

### 限流 key 会一直留在 Redis 里吗？

不会。每次 `Allow` 都会给 key 设置与窗口等长的过期时间，窗口内无新请求后 key 自动消失。

### 为什么同一毫秒的多次请求没有被 ZSET 去重？

member 由「毫秒时间戳:实例标识:自增序号」组成，每次请求的自增序号不同，不会互相覆盖。

### 测试时如何避免后台协程泄漏？

创建后 `defer limiter.Close()`，或在测试结束时显式调用 `Close()`；协程通过 context 取消，调用后立即退出。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个组件。

## 许可证

MIT License
