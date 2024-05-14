package ExampleRedis

import (
	"context"
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/redis"
	"log"
)

func Redis() {
	conf := jcbaseGo.RedisStruct{
		Host:     "127.0.0.1",
		Port:     "6379",
		Password: "",
		Db:       "0",
	}
	var ctx = context.Background()

	rdb := redis.New(conf).GetRdb()

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("key", val)
}
