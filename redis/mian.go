package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
)

type ContextStruct struct {
	Ctx    context.Context
	Rdb    *redis.Client
	Config jcbaseGo.RedisStruct
}

func New(conf jcbaseGo.RedisStruct) *ContextStruct {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       helper.ToInt(conf.Db),
	})
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return &ContextStruct{
		Ctx:    ctx,
		Rdb:    rdb,
		Config: conf,
	}
}

func (cs *ContextStruct) GetCtx(ctx *ContextStruct) *ContextStruct {
	ctx = cs
	return cs
}

func (cs *ContextStruct) GetRdb() *redis.Client {
	return cs.Rdb
}
