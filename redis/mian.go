package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
)

var Ctx = context.Background()
var Rdb *redis.Client
var Config jcbaseGo.RedisStruct

func init() {
	Config = jcbaseGo.Config.Get().Redis
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", Config.Host, Config.Port),
		Password: Config.Password,
		DB:       helper.ToInt(Config.Db),
	})
	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		panic(err)
	}
}

// GetRedis 获取指定db的redis实例
func GetRedis(db any) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", Config.Host, Config.Port),
		Password: Config.Password,
		DB:       helper.ToInt(db),
	})
	_, err := rdb.Ping(Ctx).Result()
	if err != nil {
		panic(err)
	}
	return rdb
}
