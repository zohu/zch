package zch

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type L2 struct {
	c      *Memory
	r      *Rds
	prefix string
}

var l2 *L2

type Options struct {
	Expiration    time.Duration
	CleanInterval time.Duration
	PrefixL2      string
	RedisOptions  *redis.UniversalOptions
}

// NewL2
// @Description: 创建l2缓存
// @param expiration
// @param cleanInterval
// @param ops
// @return *L2
func NewL2(ops *Options) *L2 {
	if ops.Expiration == 0 {
		ops.Expiration = time.Hour
	}
	if ops.CleanInterval == 0 {
		ops.CleanInterval = time.Minute * 5
	}
	if ops.PrefixL2 == "" {
		ops.PrefixL2 = "l2"
	}
	if l2 == nil {
		l2 = &L2{
			c:      NewMemory(ops.Expiration, ops.CleanInterval),
			r:      NewRds(ops.RedisOptions),
			prefix: ops.PrefixL2,
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
	key = l.prefix + ":" + key
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
	key = l.prefix + ":" + key
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
	key = l.prefix + ":" + key
	l.c.Delete(key)
	return l.r.Del(ctx, key).Err()
}

// Flush
// @Description: 释放二级缓存
// @receiver l
// @param ctx
// @return error
func (l *L2) Flush(ctx context.Context) error {
	l.c.Flush()
	return l.r.FlushCatchBatch(ctx, l.prefix+":*")
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
