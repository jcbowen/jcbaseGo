package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
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

// GetClient 获取redis client
func (i *Instance) GetClient() *redis.Client {
	return i.Client
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
//	value, err := redis.New(config).GetClient().GetString(key)
//	if err != nil {
//	    // 处理错误
//	} else {
//	    // 使用 value
//	}
//
//	value, err := redis.New(config).GetClient().GetString(key, "default")
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
//	err := redis.New(config).GetClient().GetStruct(key, &value, defaultValue)
//	if err != nil {
//	    // 处理错误
//	} else {
//	    // 使用 value
//	}
func (i *Instance) GetStruct(key string, value interface{}, args ...interface{}) error {
	var defaultValue interface{}

	if len(args) >= 1 {
		defaultValue = args[0]
	}

	strValue, err := i.GetString(key, "")
	if err != nil || strValue == "" {
		if defaultValue != nil {
			v := reflect.ValueOf(value)
			if v.Kind() == reflect.Ptr && !v.IsNil() {
				reflect.ValueOf(value).Elem().Set(reflect.ValueOf(defaultValue))
			}
		}
		return err
	}

	err = json.Unmarshal([]byte(strValue), value)
	if err != nil && defaultValue != nil {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			reflect.ValueOf(value).Elem().Set(reflect.ValueOf(defaultValue))
		}
	}

	return err
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
//	err := redis.New(config).GetClient().Set(key, value, time.Hour)
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
