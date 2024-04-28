package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
)

const Nil = redis.Nil

type StatusCmd = redis.StatusCmd
type StringCmd = redis.StringCmd

type RedisContext struct {
	Ctx    context.Context
	Rdb    *redis.Client
	Config jcbaseGo.RedisStruct
	Errors []error
}

func New(conf jcbaseGo.RedisStruct) *RedisContext {
	redisContext := &RedisContext{}

	err := helper.CheckAndSetDefault(&conf)
	if err != nil {
		redisContext.AddError(err)
		return redisContext
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       helper.ToInt(conf.Db),
	})
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()

	redisContext.Ctx = ctx
	redisContext.Rdb = rdb
	redisContext.Config = conf
	redisContext.AddError(err)

	return redisContext
}

func (cs *RedisContext) GetCtx(ctx *RedisContext) *RedisContext {
	ctx = cs
	return cs
}

func (cs *RedisContext) GetRdb() *redis.Client {
	return cs.Rdb
}

func (c *RedisContext) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

func (c *RedisContext) Error() []error {
	// 过滤掉c.Errors中的nil
	var errs []error
	for _, err := range c.Errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
