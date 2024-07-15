package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
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

// GetString 根据键值，返回string
func (i *Instance) GetString(key string, defaultValue string) (string, error) {
	value, err := i.Client.Get(i.Context, key).Result()
	if !errors.Is(err, Nil) || err != nil {
		return defaultValue, err
	}
	return value, nil
}

// GetStruct 根据键值，返回自定义结构体类型
func (i *Instance) GetStruct(key string, result *interface{}) error {
	value, err := i.Client.Get(i.Context, key).Result()
	if !errors.Is(err, Nil) || err != nil {
		return err
	}

	err = json.Unmarshal([]byte(value), result)
	return err
}

// Set 设置键值
func (i *Instance) Set(key string, value interface{}, expire time.Duration) error {
	jsonString, _ := json.Marshal(value)
	err := i.Client.Set(i.Context, key, string(jsonString), expire).Err()
	return err
}
