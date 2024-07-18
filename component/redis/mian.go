package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"reflect"
	"time"
)

const Nil = redis.Nil

type Client = redis.Client

type StatusCmd = redis.StatusCmd
type StringCmd = redis.StringCmd

type Instance struct {
	Context context.Context
	Conf    jcbaseGo.RedisStruct
	Client  *redis.Client
}

// New 获取新的redis连接
func New(conf jcbaseGo.RedisStruct) *Instance {
	instance := &Instance{}

	err := helper.CheckAndSetDefault(&conf)
	jcbaseGo.PanicIfError(err)

	newClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       helper.ToInt(conf.Db),
	})
	ctx := context.Background()
	_, err = newClient.Ping(ctx).Result()

	jcbaseGo.PanicIfError(err)

	instance.Context = ctx
	instance.Client = newClient
	instance.Conf = conf

	return instance
}

// ------ 基础方法 ------ /

// GetClient 获取redis client
func (i *Instance) GetClient() *redis.Client {
	return i.Client
}

// Ping 检查与 Redis 服务器的连接是否正常。
func (i *Instance) Ping() (string, error) {
	return i.Client.Ping(i.Context).Result()
}

// Info 获取 Redis 服务器的信息。
func (i *Instance) Info() (string, error) {
	return i.Client.Info(i.Context).Result()
}

// Eval 执行 Lua 脚本。
func (i *Instance) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return i.Client.Eval(i.Context, script, keys, args...).Result()
}

// ------ 常用操作 ------ /

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
func (i *Instance) Set(key string, value interface{}, args ...time.Duration) error {
	var expire time.Duration
	if len(args) > 0 {
		expire = args[0]
	}

	jsonString, _ := json.Marshal(value)
	err := i.Client.Set(i.Context, key, string(jsonString), expire).Err()
	return err
}

// GetString 根据键值，返回字符串值。
// 如果键不存在或发生错误，则返回默认值（如果提供）。
//
// 参数:
//   - key (必需): 用于查找数据的字符串键值。
//   - defaultValue (可选): 可选的默认值（字符串类型）。
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
func (i *Instance) GetString(key string, args ...string) (string, error) {
	var defaultValue string
	if len(args) > 0 {
		defaultValue = args[0]
	}

	value, err := i.Client.Get(i.Context, key).Result()
	if err != nil || value == "" {
		if errors.Is(err, Nil) {
			err = nil // Nil error means key not found, not an actual error.
		}
		return defaultValue, err
	}
	return value, err
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
func (i *Instance) GetStruct(key string, value interface{}, args ...interface{}) error {
	// 检查 value 是否为指针
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("value must be a non-nil pointer")
	}

	var defaultValue interface{}
	if len(args) >= 1 {
		defaultValue = args[0]
	}

	strValue, err := i.GetString(key, "")
	if err != nil || strValue == "" {
		// 处理默认值
		if defaultValue != nil {
			defaultV := reflect.ValueOf(defaultValue)
			if defaultV.Type() == v.Elem().Type() {
				v.Elem().Set(defaultV)
			} else {
				return errors.New("defaultValue type does not match value type")
			}
		}
		return err
	}

	// 反序列化 JSON 字符串
	err = json.Unmarshal([]byte(strValue), value)
	if err != nil && defaultValue != nil {
		// 反序列化失败时处理默认值
		defaultV := reflect.ValueOf(defaultValue)
		if defaultV.Type() == v.Elem().Type() {
			v.Elem().Set(defaultV)
		} else {
			return errors.New("defaultValue type does not match value type")
		}
	}

	return err
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
func (i *Instance) Del(key ...string) error {
	return i.Client.Del(i.Context, key...).Err()
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
func (i *Instance) Exists(key string) (bool, error) {
	exists, err := i.Client.Exists(i.Context, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// Keys 获取键值列表。
//
// 参数:
//   - pattern (必需): 匹配的键值模式。
//
// 返回值:
//   - []string: 匹配到的键值列表。
//   - error: 如果发生错误则返回相应的错误信息。
//
// 示例:
//
//	keys, err := Keys(pattern)
//	if err != nil {
//	    // 处理错误
//	}
func (i *Instance) Keys(pattern string) ([]string, error) {
	return i.Client.Keys(i.Context, pattern).Result()
}

// ----- 列表操作 ----- /

// LPush 向列表左端插入一个或多个元素。
func (i *Instance) LPush(key string, values ...interface{}) error {
	_, err := i.Client.LPush(i.Context, key, values...).Result()
	return err
}

// RPush 向列表右端插入一个或多个元素。
func (i *Instance) RPush(key string, values ...interface{}) error {
	_, err := i.Client.RPush(i.Context, key, values...).Result()
	return err
}

// LRange 获取列表中指定范围的元素。
func (i *Instance) LRange(key string, start, stop int64) ([]string, error) {
	result, err := i.Client.LRange(i.Context, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LLen 获取列表的长度。
func (i *Instance) LLen(key string) (int64, error) {
	result, err := i.Client.LLen(i.Context, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// LPop 从列表左端弹出一个元素。
func (i *Instance) LPop(key string) (string, error) {
	result, err := i.Client.LPop(i.Context, key).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

// RPop 从列表右端弹出一个元素。
func (i *Instance) RPop(key string) (string, error) {
	result, err := i.Client.RPop(i.Context, key).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

// ----- 集合操作 ----- /

// SAdd 向集合添加一个或多个成员。
func (i *Instance) SAdd(key string, members ...interface{}) (int64, error) {
	result, err := i.Client.SAdd(i.Context, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SRem 从集合中移除一个或多个成员。
func (i *Instance) SRem(key string, members ...interface{}) (int64, error) {
	result, err := i.Client.SRem(i.Context, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SMembers 获取集合中的所有成员。
func (i *Instance) SMembers(key string) ([]string, error) {
	result, err := i.Client.SMembers(i.Context, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SIsMember 检查成员是否存在于集合中。
func (i *Instance) SIsMember(key string, member interface{}) (bool, error) {
	result, err := i.Client.SIsMember(i.Context, key, member).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// ----- 哈希表操作 ----- /

// HSet 设置哈希表中的字段值。
func (i *Instance) HSet(key, field string, value interface{}) error {
	return i.Client.HSet(i.Context, key, field, value).Err()
}

// HGet 获取哈希表中指定字段的值。
func (i *Instance) HGet(key, field string) (string, error) {
	result, err := i.Client.HGet(i.Context, key, field).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

// HGetAll 获取哈希表中所有字段和值。
func (i *Instance) HGetAll(key string) (map[string]string, error) {
	result, err := i.Client.HGetAll(i.Context, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// HExists 检查哈希表中是否存在指定字段。
func (i *Instance) HExists(key, field string) (bool, error) {
	result, err := i.Client.HExists(i.Context, key, field).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// ----- 计数/计时器操作 ----- /

// Incr 增加计数器的值。
func (i *Instance) Incr(key string) (int64, error) {
	result, err := i.Client.Incr(i.Context, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Decr 减少计数器的值。
func (i *Instance) Decr(key string) (int64, error) {
	result, err := i.Client.Decr(i.Context, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Expire 设置键的过期时间。
func (i *Instance) Expire(key string, expiration time.Duration) (bool, error) {
	result, err := i.Client.Expire(i.Context, key, expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// TTL 获取键的剩余过期时间。
func (i *Instance) TTL(key string) (time.Duration, error) {
	result, err := i.Client.TTL(i.Context, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// ----- 发布与订阅 ----- /

// Publish 向指定频道发布消息。
func (i *Instance) Publish(channel string, message interface{}) error {
	return i.Client.Publish(i.Context, channel, message).Err()
}

// Subscribe 订阅指定频道接收消息。
func (i *Instance) Subscribe(channels ...string) (*redis.PubSub, error) {
	pubsub := i.Client.Subscribe(i.Context, channels...)
	_, err := pubsub.Receive(i.Context)
	if err != nil {
		return nil, err
	}
	return pubsub, nil
}
