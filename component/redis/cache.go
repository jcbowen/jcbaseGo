package redis

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

// CacheOpt 结构体表示缓存组件
type CacheOpt struct {
	redis  *Instance
	Prefix string        `json:"prefix" default:"jcbase"`
	Expire time.Duration `json:"expire" default:"0"`
}

// NewCache 创建一个新的 CacheOpt 实例
func NewCache(redis *Instance, args ...CacheOpt) *CacheOpt {
	c := &CacheOpt{
		redis: redis,
	}
	if len(args) > 0 {
		cacheConfig := args[0]
		c.Prefix = cacheConfig.Prefix
		c.Expire = cacheConfig.Expire
	}
	_ = helper.CheckAndSetDefault(c)
	return c
}

// Set 设置键值。
//
// 参数:
//   - key (必需): 要设置的键值。
//   - value (必需): 要设置的值，将被转换为 JSON 格式保存。
//   - expire (可选): 数据的过期时间，如果未设置则数据将永不过期。
//
// 返回值:
//   - error: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	err := Set(key, value, time.Hour)
//	if err != nil {
//	    // 处理错误
//	}
func (c *CacheOpt) Set(key string, value interface{}, args ...time.Duration) error {
	var expire time.Duration
	if len(args) > 0 {
		expire = args[0]
	}
	return c.redis.Set(c.keygen(key), value, expire)
}

// GetString 根据键值，返回字符串值。
// 如果键不存在或发生错误，则返回默认值（如果提供）。
//
// 参数:
//   - key (必需): 用于查找数据的字符串键值。
//   - args (可选): 可选的默认值（字符串类型）。
//
// 返回值:
//   - value: 获取到的字符串值，或者是默认值。
//   - err: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	value, err := GetString(key)
//	if err != nil {
//	    // 处理错误
//	} else {
//	    // 使用 value
//	}
//
//	value, err := GetString(key, "default")
//	if err != nil {
//	    // 处理错误
//	} else {
//	    // 使用 value
//	}
func (c *CacheOpt) GetString(key string) (string, error) {
	return c.redis.GetString(c.keygen(key))
}

// GetStruct 根据键值获取自定义结构体类型。
// args 参数列表，支持以下格式：
//  1. 提供 key 和 value（指向目标结构体的指针）：GetStruct(key, &value)
//  2. 提供 key, value 和 defaultValue（默认值的任意类型）：GetStruct(key, &value, defaultValue)
//
// 参数:
//   - key (必需): 用于查找数据的字符串键值。
//   - value (必需): 指向目标结构体的指针，用于存储获取到的数据。
//   - defaultValue (可选): 当获取数据失败时的默认值。
//
// 返回值:
//   - error: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	var value MyStruct
//	defaultValue := MyStruct{Field: "default"}
//	err := GetStruct(key, &value, defaultValue)
//	if err != nil {
//	    // 处理错误
//	} else {
//	    // 使用 value
//	}
func (c *CacheOpt) GetStruct(key string, value interface{}, args ...interface{}) error {
	return c.redis.GetStruct(c.keygen(key), value, args...)
}

// Del 删除键值。
//
// 参数:
//   - key (必需): 要删除的键值。
//
// 返回值:
//   - error: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	err := Del(key)
//	if err != nil {
//	    // 处理错误
//	}
func (c *CacheOpt) Del(key string) error {
	return c.redis.Del(c.keygen(key))
}

// Exists 检查键值是否存在。
//
// 参数:
//   - key (必需): 要检查的键值。
//
// 返回值:
//   - bool: 如果键值存在则返回 true，否则返回 false。
//   - error: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	exists, err := Exists(key)
//	if err != nil {
//	    // 处理错误
//	}
func (c *CacheOpt) Exists(key string) (bool, error) {
	return c.redis.Exists(c.keygen(key))
}

// ClearAll 清除所有缓存。
//
// 返回值:
//   - error: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	err := ClearAll()
//	if err != nil {
//	    // 处理错误
//	}
func (c *CacheOpt) ClearAll() error {
	keys, err := c.redis.Client.Keys(c.redis.Context, "jcbase_"+c.Prefix+"_*").Result()
	if err != nil && !errors.Is(err, Nil) {
		return err
	}
	for _, key := range keys {
		_ = c.redis.Del(key)
	}
	return nil
}

// 生成缓存键
func (c *CacheOpt) keygen(key string) string {
	hashed := md5.New()
	hashed.Write([]byte(key))
	return "jcbase_" + c.Prefix + "_" + hex.EncodeToString(hashed.Sum(nil))
}
