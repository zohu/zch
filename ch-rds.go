package zch

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"reflect"
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

// FlushCatchBatch
// @Description: 批量清除缓存
// @receiver r
// @param ctx
// @param keys
// @return error
func (r *Rds) FlushCatchBatch(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		km, err := r.Do(ctx, "KEYS", fmt.Sprintf("%s*", key)).Result()
		if err != nil {
			return err
		}
		if reflect.TypeOf(km).Kind() == reflect.Slice {
			val := reflect.ValueOf(km)
			if val.Len() == 0 {
				continue
			}
			for i := 0; i < val.Len(); i++ {
				r.Del(ctx, val.Index(i).Interface().(string))
			}
		}
	}
	return nil
}
