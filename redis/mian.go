package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
)

const Nil = redis.Nil

type Client = redis.Client

type StatusCmd = redis.StatusCmd
type StringCmd = redis.StringCmd

type Instance struct {
	Context context.Context
	Conf    jcbaseGo.RedisStruct
	Client  *redis.Client
	Errors  []error
}

func New(conf jcbaseGo.RedisStruct) *Instance {
	instance := &Instance{}

	err := helper.CheckAndSetDefault(&conf)
	if err != nil {
		instance.AddError(err)
		return instance
	}

	newClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       helper.ToInt(conf.Db),
	})
	ctx := context.Background()
	_, err = newClient.Ping(ctx).Result()

	instance.Context = ctx
	instance.Client = newClient
	instance.Conf = conf
	instance.AddError(err)

	return instance
}

func (i *Instance) GetClient() *redis.Client {
	return i.Client
}

func (i *Instance) AddError(err error) {
	if err != nil {
		i.Errors = append(i.Errors, err)
	}
}

func (i *Instance) Error() []error {
	// 过滤掉c.Errors中的nil
	var errs []error
	for _, err := range i.Errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
