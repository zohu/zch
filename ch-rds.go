package zch

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Rds struct {
	redis.UniversalClient
}

func NewRds(ops *redis.UniversalOptions) *Rds {
	cl := redis.NewUniversalClient(ops)
	if err := cl.Ping(context.TODO()).Err(); err != nil {
		panic(err)
	}
	return &Rds{
		cl,
	}
}
