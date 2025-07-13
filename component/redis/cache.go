package redis

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

// Cache 缓存管理器结构体
// 提供基于Redis的缓存操作，支持键前缀、过期时间等配置
type Cache struct {
	redis  *Instance
	Prefix string        `json:"prefix" default:"jcbase"`
	Expire time.Duration `json:"expire" default:"0"`
}

// CacheOption 缓存配置选项函数
type CacheOption func(*Cache)

// WithPrefix 设置缓存键前缀
func WithPrefix(prefix string) CacheOption {
	return func(c *Cache) {
		c.Prefix = prefix
	}
}

// WithExpire 设置默认过期时间
func WithExpire(expire time.Duration) CacheOption {
	return func(c *Cache) {
		c.Expire = expire
	}
}

// NewCache 创建一个新的缓存管理器实例
// 参数：
//   - redis: Redis实例，不能为空
//   - opts: 可选的配置选项
//
// 返回：
//   - *Cache: 缓存管理器实例
//   - error: 创建失败时的错误信息
func NewCache(redis *Instance, opts ...CacheOption) (*Cache, error) {
	if redis == nil {
		return nil, errors.New("Redis实例不能为空")
	}

	c := &Cache{
		redis: redis,
	}

	// 设置默认值
	if err := helper.CheckAndSetDefault(c); err != nil {
		return nil, err
	}

	// 应用选项
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Set 设置缓存键值
// 参数：
//   - key: 缓存键
//   - value: 缓存值，会被序列化为JSON
//   - expire: 过期时间，如果为0则使用默认过期时间
//
// 返回：
//   - error: 设置失败时的错误信息
func (c *Cache) Set(key string, value interface{}, expire ...time.Duration) error {
	if c.redis == nil {
		return errors.New("Redis实例未初始化")
	}

	if key == "" {
		return errors.New("缓存键不能为空")
	}

	// 确定过期时间
	var expiration time.Duration
	if len(expire) > 0 {
		expiration = expire[0]
	} else {
		expiration = c.Expire
	}

	return c.redis.Set(c.keygen(key), value, expiration)
}

// GetString 获取字符串类型的缓存值
// 参数：
//   - key: 缓存键
//   - defaultValue: 可选的默认值
//
// 返回：
//   - string: 缓存值或默认值
//   - error: 获取失败时的错误信息
func (c *Cache) GetString(key string, defaultValue ...string) (string, error) {
	if c.redis == nil {
		return "", errors.New("Redis实例未初始化")
	}

	if key == "" {
		return "", errors.New("缓存键不能为空")
	}

	var def string
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	return c.redis.GetString(c.keygen(key), def)
}

// GetStruct 获取结构体类型的缓存值
// 参数：
//   - key: 缓存键
//   - value: 指向目标结构体的指针
//   - defaultValue: 可选的默认值
//
// 返回：
//   - error: 获取失败时的错误信息
func (c *Cache) GetStruct(key string, value interface{}, defaultValue ...interface{}) error {
	if c.redis == nil {
		return errors.New("Redis实例未初始化")
	}

	if key == "" {
		return errors.New("缓存键不能为空")
	}

	var args []interface{}
	if len(defaultValue) > 0 {
		args = defaultValue
	}

	return c.redis.GetStruct(c.keygen(key), value, args...)
}

// Del 删除缓存键
// 参数：
//   - key: 要删除的缓存键
//
// 返回：
//   - error: 删除失败时的错误信息
func (c *Cache) Del(key string) error {
	if c.redis == nil {
		return errors.New("Redis实例未初始化")
	}

	if key == "" {
		return errors.New("缓存键不能为空")
	}

	return c.redis.Del(c.keygen(key))
}

// Exists 检查缓存键是否存在
// 参数：
//   - key: 要检查的缓存键
//
// 返回：
//   - bool: 键存在返回true，否则返回false
//   - error: 检查失败时的错误信息
func (c *Cache) Exists(key string) (bool, error) {
	if c.redis == nil {
		return false, errors.New("Redis实例未初始化")
	}

	if key == "" {
		return false, errors.New("缓存键不能为空")
	}

	return c.redis.Exists(c.keygen(key))
}

// ClearAll 清除所有当前前缀的缓存
// 返回：
//   - error: 清除失败时的错误信息
func (c *Cache) ClearAll() error {
	if c.redis == nil {
		return errors.New("Redis实例未初始化")
	}

	// 使用keygen方法生成模式匹配的键
	pattern := c.keygen("*")
	keys, err := c.redis.Client.Keys(c.redis.Context, pattern).Result()
	if err != nil && !errors.Is(err, Nil) {
		return err
	}

	// 批量删除键
	if len(keys) > 0 {
		err = c.redis.Client.Del(c.redis.Context, keys...).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetKeys 获取所有匹配模式的缓存键
// 参数：
//   - pattern: 键模式，支持通配符
//
// 返回：
//   - []string: 匹配的键列表
//   - error: 获取失败时的错误信息
func (c *Cache) GetKeys(pattern string) ([]string, error) {
	if c.redis == nil {
		return nil, errors.New("Redis实例未初始化")
	}

	// 如果pattern不包含通配符，则添加前缀
	if pattern == "" || pattern == "*" {
		pattern = "*"
	} else if pattern[0] != '*' && pattern[len(pattern)-1] != '*' {
		pattern = pattern + "*"
	}

	fullPattern := c.keygen(pattern)
	keys, err := c.redis.Client.Keys(c.redis.Context, fullPattern).Result()
	if err != nil && !errors.Is(err, Nil) {
		return nil, err
	}

	// 移除前缀，返回原始键名
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		if originalKey := c.removePrefix(key); originalKey != "" {
			result = append(result, originalKey)
		}
	}

	return result, nil
}

// GetStats 获取缓存统计信息
// 返回：
//   - map[string]interface{}: 统计信息
//   - error: 获取失败时的错误信息
func (c *Cache) GetStats() (map[string]interface{}, error) {
	if c.redis == nil {
		return nil, errors.New("Redis实例未初始化")
	}

	pattern := c.keygen("*")
	keys, err := c.redis.Client.Keys(c.redis.Context, pattern).Result()
	if err != nil && !errors.Is(err, Nil) {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_keys":     len(keys),
		"prefix":         c.Prefix,
		"default_expire": c.Expire.String(),
	}

	return stats, nil
}

// keygen 生成带前缀的缓存键
// 参数：
//   - key: 原始键名
//
// 返回：
//   - string: 带前缀的完整键名
func (c *Cache) keygen(key string) string {
	if key == "*" {
		return "jcbase_" + c.Prefix + "_*"
	}

	hashed := md5.New()
	hashed.Write([]byte(key))
	return "jcbase_" + c.Prefix + "_" + hex.EncodeToString(hashed.Sum(nil))
}

// removePrefix 从完整键名中移除前缀，返回原始键名
// 参数：
//   - fullKey: 完整的键名
//
// 返回：
//   - string: 原始键名，如果无法解析则返回空字符串
func (c *Cache) removePrefix(fullKey string) string {
	prefix := "jcbase_" + c.Prefix + "_"
	if len(fullKey) <= len(prefix) {
		return ""
	}

	if fullKey[:len(prefix)] != prefix {
		return ""
	}

	// 由于使用了MD5哈希，无法直接还原原始键名
	// 这里返回哈希部分，实际使用中可能需要维护一个反向映射
	return fullKey[len(prefix):]
}
