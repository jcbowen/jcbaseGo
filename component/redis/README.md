# Redis 组件

Redis 组件提供完整的 Redis 客户端功能，包括连接管理、数据操作、缓存管理和发布订阅等功能。

## 概述

Redis 组件是基于 go-redis 库封装的 Redis 客户端组件，提供简单易用的 API 接口，支持字符串、列表、集合、哈希表等多种数据类型的操作，同时提供缓存管理功能，支持键前缀、过期时间等配置。

## 功能特性

- **连接管理**：自动连接池管理，支持连接测试和健康检查
- **数据类型支持**：字符串、列表、集合、哈希表、有序集合等
- **缓存管理**：提供缓存管理器，支持键前缀、过期时间配置
- **发布订阅**：支持消息发布和订阅功能
- **事务支持**：支持 Redis 事务操作
- **Lua 脚本**：支持执行 Lua 脚本
- **错误处理**：完善的错误处理机制
- **配置管理**：支持默认值检查和配置验证

## 快速开始

### 安装

```go
go get github.com/jcbowen/jcbaseGo/component/redis
```

### 基本使用示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
    "time"
)

func main() {
    // Redis 配置
    config := jcbaseGo.RedisStruct{
        Host:     "localhost",
        Port:     "6379",
        Password: "",
        Db:       "0",
    }

    // 创建 Redis 实例
    redisInstance := redis.New(config)

    // 测试连接
    pong, err := redisInstance.Ping()
    if err != nil {
        panic(fmt.Sprintf("Redis 连接失败: %v", err))
    }
    fmt.Printf("Redis 连接成功: %s\n", pong)

    // 设置键值
    err = redisInstance.Set("test_key", "hello world", time.Hour)
    if err != nil {
        panic(fmt.Sprintf("设置键值失败: %v", err))
    }

    // 获取键值
    value, err := redisInstance.GetString("test_key")
    if err != nil {
        panic(fmt.Sprintf("获取键值失败: %v", err))
    }
    fmt.Printf("获取的值: %s\n", value)

    // 删除键值
    err = redisInstance.Del("test_key")
    if err != nil {
        panic(fmt.Sprintf("删除键值失败: %v", err))
    }
    fmt.Println("键值删除成功")
}
```

### 使用缓存管理器示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
    "time"
)

func main() {
    // Redis 配置
    config := jcbaseGo.RedisStruct{
        Host:     "localhost",
        Port:     "6379",
        Password: "",
        Db:       "0",
    }

    // 创建 Redis 实例
    redisInstance := redis.New(config)

    // 创建缓存管理器
    cache, err := redis.NewCache(redisInstance, 
        redis.WithPrefix("myapp"),
        redis.WithExpire(time.Hour),
    )
    if err != nil {
        panic(fmt.Sprintf("创建缓存管理器失败: %v", err))
    }

    // 设置缓存
    user := map[string]interface{}{
        "id":   1,
        "name": "张三",
        "age":  30,
    }
    
    err = cache.Set("user:1", user, 2*time.Hour)
    if err != nil {
        panic(fmt.Sprintf("设置缓存失败: %v", err))
    }

    // 获取缓存
    var cachedUser map[string]interface{}
    err = cache.GetStruct("user:1", &cachedUser)
    if err != nil {
        panic(fmt.Sprintf("获取缓存失败: %v", err))
    }
    fmt.Printf("缓存用户: %+v\n", cachedUser)

    // 检查缓存是否存在
    exists, err := cache.Exists("user:1")
    if err != nil {
        panic(fmt.Sprintf("检查缓存失败: %v", err))
    }
    fmt.Printf("缓存存在: %v\n", exists)

    // 获取缓存统计信息
    stats, err := cache.GetStats()
    if err != nil {
        panic(fmt.Sprintf("获取统计信息失败: %v", err))
    }
    fmt.Printf("缓存统计: %+v\n", stats)
}
```

### 列表操作示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
)

func main() {
    config := jcbaseGo.RedisStruct{
        Host: "localhost",
        Port: "6379",
    }

    redisInstance := redis.New(config)

    // 向列表左端插入元素
    err := redisInstance.LPush("mylist", "item1", "item2")
    if err != nil {
        panic(fmt.Sprintf("LPush 失败: %v", err))
    }

    // 向列表右端插入元素
    err = redisInstance.RPush("mylist", "item3", "item4")
    if err != nil {
        panic(fmt.Sprintf("RPush 失败: %v", err))
    }

    // 获取列表范围
    items, err := redisInstance.LRange("mylist", 0, -1)
    if err != nil {
        panic(fmt.Sprintf("LRange 失败: %v", err))
    }
    fmt.Printf("列表内容: %v\n", items)

    // 获取列表长度
    length, err := redisInstance.LLen("mylist")
    if err != nil {
        panic(fmt.Sprintf("LLen 失败: %v", err))
    }
    fmt.Printf("列表长度: %d\n", length)

    // 从列表左端弹出元素
    item, err := redisInstance.LPop("mylist")
    if err != nil {
        panic(fmt.Sprintf("LPop 失败: %v", err))
    }
    fmt.Printf("弹出的元素: %s\n", item)
}
```

### 哈希表操作示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
)

func main() {
    config := jcbaseGo.RedisStruct{
        Host: "localhost",
        Port: "6379",
    }

    redisInstance := redis.New(config)

    // 设置哈希表字段
    err := redisInstance.HSet("user:profile", "name", "李四")
    if err != nil {
        panic(fmt.Sprintf("HSet 失败: %v", err))
    }

    err = redisInstance.HSet("user:profile", "age", "25")
    if err != nil {
        panic(fmt.Sprintf("HSet 失败: %v", err))
    }

    // 获取哈希表字段
    name, err := redisInstance.HGet("user:profile", "name")
    if err != nil {
        panic(fmt.Sprintf("HGet 失败: %v", err))
    }
    fmt.Printf("用户名: %s\n", name)

    // 获取所有哈希表字段
    profile, err := redisInstance.HGetAll("user:profile")
    if err != nil {
        panic(fmt.Sprintf("HGetAll 失败: %v", err))
    }
    fmt.Printf("用户资料: %+v\n", profile)

    // 检查字段是否存在
    exists, err := redisInstance.HExists("user:profile", "name")
    if err != nil {
        panic(fmt.Sprintf("HExists 失败: %v", err))
    }
    fmt.Printf("字段存在: %v\n", exists)
}
```

### 发布订阅示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
    "time"
)

func main() {
    config := jcbaseGo.RedisStruct{
        Host: "localhost",
        Port: "6379",
    }

    redisInstance := redis.New(config)

    // 启动订阅者协程
    go func() {
        pubsub, err := redisInstance.Subscribe("mychannel")
        if err != nil {
            panic(fmt.Sprintf("订阅失败: %v", err))
        }
        defer pubsub.Close()

        // 接收消息
        for {
            msg, err := pubsub.ReceiveMessage(redisInstance.Context)
            if err != nil {
                fmt.Printf("接收消息失败: %v\n", err)
                break
            }
            fmt.Printf("收到消息: 频道=%s, 内容=%s\n", msg.Channel, msg.Payload)
        }
    }()

    // 等待订阅者准备
    time.Sleep(100 * time.Millisecond)

    // 发布消息
    for i := 1; i <= 3; i++ {
        message := fmt.Sprintf("消息 %d", i)
        err := redisInstance.Publish("mychannel", message)
        if err != nil {
            panic(fmt.Sprintf("发布消息失败: %v", err))
        }
        fmt.Printf("发布消息: %s\n", message)
        time.Sleep(500 * time.Millisecond)
    }

    time.Sleep(1 * time.Second)
}
```

## 详细配置

### RedisStruct 配置

RedisStruct 是 Redis 连接的基础配置结构体：

```go
type RedisStruct struct {
    Host     string `json:"host" default:"127.0.0.1"`     // Redis 主机地址
    Port     string `json:"port" default:"6379"`         // Redis 端口
    Password string `json:"password" default:""`         // Redis 密码
    Db       string `json:"db" default:"0"`             // Redis 数据库编号
}
```

### Cache 配置选项

Cache 管理器支持以下配置选项：

```go
// 设置缓存键前缀
cache, err := redis.NewCache(redisInstance, 
    redis.WithPrefix("myapp"),
)

// 设置默认过期时间
cache, err := redis.NewCache(redisInstance,
    redis.WithExpire(time.Hour),
)

// 同时设置前缀和过期时间
cache, err := redis.NewCache(redisInstance,
    redis.WithPrefix("myapp"),
    redis.WithExpire(30*time.Minute),
)
```

## 高级功能

### 使用 Lua 脚本

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
)

func main() {
    config := jcbaseGo.RedisStruct{
        Host: "localhost",
        Port: "6379",
    }

    redisInstance := redis.New(config)

    // 定义 Lua 脚本
    script := `
        local key = KEYS[1]
        local increment = ARGV[1]
        local current = redis.call('GET', key) or 0
        local new_value = current + increment
        redis.call('SET', key, new_value)
        return new_value
    `

    // 执行 Lua 脚本
    result, err := redisInstance.Eval(script, []string{"counter"}, 5)
    if err != nil {
        panic(fmt.Sprintf("执行 Lua 脚本失败: %v", err))
    }

    fmt.Printf("Lua 脚本执行结果: %v\n", result)
}
```

### 批量操作示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
    "time"
)

func main() {
    config := jcbaseGo.RedisStruct{
        Host: "localhost",
        Port: "6379",
    }

    redisInstance := redis.New(config)

    // 批量设置多个键值
    keys := []string{"key1", "key2", "key3"}
    values := []string{"value1", "value2", "value3"}

    for i, key := range keys {
        err := redisInstance.Set(key, values[i], time.Hour)
        if err != nil {
            fmt.Printf("设置 %s 失败: %v\n", key, err)
        }
    }

    // 批量删除多个键
    err := redisInstance.Del(keys...)
    if err != nil {
        panic(fmt.Sprintf("批量删除失败: %v", err))
    }

    fmt.Println("批量操作完成")
}
```

### 缓存管理器高级用法

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/redis"
    "time"
)

func main() {
    config := jcbaseGo.RedisStruct{
        Host: "localhost",
        Port: "6379",
    }

    redisInstance := redis.New(config)

    cache, err := redis.NewCache(redisInstance,
        redis.WithPrefix("myapp"),
        redis.WithExpire(time.Hour),
    )
    if err != nil {
        panic(fmt.Sprintf("创建缓存管理器失败: %v", err))
    }

    // 获取所有匹配模式的键
    keys, err := cache.GetKeys("user:*")
    if err != nil {
        panic(fmt.Sprintf("获取键列表失败: %v", err))
    }
    fmt.Printf("匹配的键: %v\n", keys)

    // 清除所有缓存
    err = cache.ClearAll()
    if err != nil {
        panic(fmt.Sprintf("清除缓存失败: %v", err))
    }
    fmt.Println("缓存清除完成")

    // 获取缓存统计信息
    stats, err := cache.GetStats()
    if err != nil {
        panic(fmt.Sprintf("获取统计信息失败: %v", err))
    }
    fmt.Printf("缓存统计: %+v\n", stats)
}
```

## 错误处理

### 错误检查方法

```go
// 检查键是否存在
exists, err := redisInstance.Exists("mykey")
if err != nil {
    // 处理错误
    fmt.Printf("检查键存在失败: %v\n", err)
} else {
    fmt.Printf("键存在: %v\n", exists)
}

// 获取键值，提供默认值
value, err := redisInstance.GetString("mykey", "default_value")
if err != nil {
    fmt.Printf("获取键值失败: %v\n", err)
} else {
    fmt.Printf("键值: %s\n", value)
}
```

### 常见错误类型

- **连接错误**：无法连接到 Redis 服务器
- **认证错误**：密码认证失败
- **键不存在**：尝试获取不存在的键
- **类型错误**：操作类型不匹配
- **超时错误**：操作超时

## 性能优化建议

1. **连接池配置**：合理配置连接池大小
2. **批量操作**：使用批量操作减少网络往返
3. **管道技术**：对于多个操作使用管道
4. **键设计**：合理设计键名，避免过长的键名
5. **过期时间**：设置合理的过期时间，避免内存泄漏

## 安全考虑

- 使用密码认证保护 Redis 实例
- 配置防火墙限制访问来源
- 定期更新 Redis 版本
- 监控 Redis 内存使用情况
- 使用不同的数据库编号隔离不同应用

## API 参考

### Instance 结构体

Redis 实例的主要结构体

**字段：**
- `Context context.Context` - 上下文
- `Conf jcbaseGo.RedisStruct` - 配置信息
- `Client *redis.Client` - Redis 客户端

**主要方法：**
- `New(conf jcbaseGo.RedisStruct) *Instance` - 创建 Redis 实例
- `Ping() (string, error)` - 测试连接
- `Set(key string, value interface{}, args ...time.Duration) error` - 设置键值
- `GetString(key string, args ...string) (string, error)` - 获取字符串值
- `GetStruct(key string, value interface{}, args ...interface{}) error` - 获取结构体值
- `Del(key ...string) error` - 删除键
- `Exists(key string) (bool, error)` - 检查键是否存在
- `Keys(pattern string) ([]string, error)` - 获取匹配的键
- `LPush(key string, values ...interface{}) error` - 列表左端插入
- `RPush(key string, values ...interface{}) error` - 列表右端插入
- `LRange(key string, start, stop int64) ([]string, error)` - 获取列表范围
- `LLen(key string) (int64, error)` - 获取列表长度
- `LPop(key string) (string, error)` - 列表左端弹出
- `RPop(key string) (string, error)` - 列表右端弹出
- `SAdd(key string, members ...interface{}) (int64, error)` - 集合添加成员
- `SRem(key string, members ...interface{}) (int64, error)` - 集合移除成员
- `SMembers(key string) ([]string, error)` - 获取集合成员
- `SIsMember(key string, member interface{}) (bool, error)` - 检查集合成员
- `HSet(key, field string, value interface{}) error` - 哈希表设置字段
- `HGet(key, field string) (string, error)` - 哈希表获取字段
- `HGetAll(key string) (map[string]string, error)` - 获取所有哈希字段
- `HExists(key, field string) (bool, error)` - 检查哈希字段存在
- `Incr(key string) (int64, error)` - 增加计数器
- `Decr(key string) (int64, error)` - 减少计数器
- `Expire(key string, expiration time.Duration) (bool, error)` - 设置过期时间
- `TTL(key string) (time.Duration, error)` - 获取剩余过期时间
- `Publish(channel string, message interface{}) error` - 发布消息
- `Subscribe(channels ...string) (*redis.PubSub, error)` - 订阅频道
- `Eval(script string, keys []string, args ...interface{}) (interface{}, error)` - 执行 Lua 脚本

### Cache 结构体

缓存管理器结构体

**字段：**
- `redis *Instance` - Redis 实例
- `Prefix string` - 键前缀
- `Expire time.Duration` - 默认过期时间

**主要方法：**
- `NewCache(redis *Instance, opts ...CacheOption) (*Cache, error)` - 创建缓存管理器
- `Set(key string, value interface{}, expire ...time.Duration) error` - 设置缓存
- `GetString(key string, defaultValue ...string) (string, error)` - 获取字符串缓存
- `GetStruct(key string, value interface{}, defaultValue ...interface{}) error` - 获取结构体缓存
- `Del(key string) error` - 删除缓存
- `Exists(key string) (bool, error)` - 检查缓存存在
- `ClearAll() error` - 清除所有缓存
- `GetKeys(pattern string) ([]string, error)` - 获取匹配的键
- `GetStats() (map[string]interface{}, error)` - 获取统计信息

### 配置选项函数

- `WithPrefix(prefix string) CacheOption` - 设置键前缀
- `WithExpire(expire time.Duration) CacheOption` - 设置默认过期时间

## 版本历史

- v1.0.0：初始版本，包含基本的 Redis 操作和缓存管理功能
- 支持多种数据类型的操作
- 提供缓存管理器功能

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Redis 组件。

## 许可证

Redis 组件遵循 MIT 许可证。