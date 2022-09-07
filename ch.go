package zch

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type L2 struct {
	c *Memory
	r *Rds
}

var l2 *L2

// NewL2
// @Description: 创建l2缓存
// @param expiration
// @param cleanInterval
// @param ops
// @return *L2
func NewL2(expiration time.Duration, cleanInterval time.Duration, ops *redis.UniversalOptions) *L2 {
	if l2 == nil {
		l2 = &L2{
			c: NewMemory(expiration, cleanInterval),
			r: NewRds(ops),
		}
	}
	return l2
}

func L() *L2 {
	if l2 == nil {
		panic("Please call NewL2 before using L")
	}
	return l2
}
func C() *Memory {
	if l2 == nil {
		panic("Please call NewL2 before using C")
	}
	return l2.c
}
func R() *Rds {
	if l2 == nil {
		panic("Please call NewL2 before using R")
	}
	return l2.r
}

// Set
// @Description: 设置缓存
// @receiver l
// @param ctx
// @param key
// @param value
// @param expiration
// @return error
func (l *L2) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := l.r.Set(ctx, key, value, expiration).Err(); err == nil {
		l.c.Set(key, value, l1(expiration))
		return nil
	} else {
		return err
	}
}

// Get
// @Description: 获取缓存
// @receiver l
// @param ctx
// @param key
// @return interface{}
// @return error
func (l *L2) Get(ctx context.Context, key string) (interface{}, error) {
	if v, ok := l.c.Get(key); ok {
		return v, nil
	} else {
		if v, err := l.r.Get(ctx, key).Result(); err == nil {
			exp := l.r.TTL(ctx, key).Val()
			if exp > 0 {
				l.c.Set(key, v, l1(exp))
			}
			return v, nil
		} else {
			return nil, err
		}
	}
}

// Del
// @Description: 删除缓存
// @receiver l
// @param ctx
// @param key
// @return error
func (l *L2) Del(ctx context.Context, key string) error {
	l.c.Delete(key)
	return l.r.Del(ctx, key).Err()
}

// l1
// @Description: 计算l1缓存的过期时间, l1总是比l2短一些, 且最长是30min，减少内存占用且防止NX虚锁
// @param expiration
// @return time.Duration
func l1(expiration time.Duration) time.Duration {
	if expiration >= 35*time.Minute {
		return 30 * time.Minute
	} else if expiration >= 15*time.Minute {
		return 10 * time.Minute
	} else if expiration >= 10*time.Minute {
		return 5 * time.Minute
	} else if expiration >= 5*time.Minute {
		return time.Minute
	}
	return expiration
}
