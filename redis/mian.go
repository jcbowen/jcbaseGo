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

func init() {
	redisConfig := jcbaseGo.Config.Get().Redis
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       helper.Str2Int(redisConfig.Db),
	})
	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		panic(err)
	}
}
