package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	errOverCapacity = errors.New("cache: 超过容量")
)

// MaxCntCache 控制住缓存住的键值对数量
type MaxCntCache struct {
	*BuildInMapCache
	cnt    int32
	maxCnt int32
}

var _ Cache = new(MaxCntCache)

func NewMaxCntCache(c *BuildInMapCache, maxCnt int32) *MaxCntCache {
	cache := &MaxCntCache{
		BuildInMapCache: c,
		maxCnt:          maxCnt,
	}

	origin := c.onEvicted
	cache.onEvicted = func(key string, val any) {
		atomic.AddInt32(&cache.cnt, -1)
		if origin != nil {
			origin(key, val)
		}
	}
	return cache
}

func (c *MaxCntCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 写法1, 如果key已经存在, 那么计数就不准了
	// cnt := atomic.AddInt32(&c.cnt, 1)
	// if cnt > c.maxCnt {
	// 	atomic.AddInt32(&c.cnt, -1)
	// 	return errOverCapacity
	// }

	// return  c.BuildInMapCache.Set(ctx, key, val, expiration)

	// 写法2, 存在并发问题, 在return前的解锁会导致c.cnt加多次(加锁存在间隙)
	// c.mutex.Lock()
	// _, ok := c.data[key]
	// if !ok {
	// 	c.cnt++
	// }
	// if c.cnt > c.maxCnt {
	// 	c.mutex.Unlock()
	// 	return errOverCapacity
	// }
	// c.mutex.Unlock()
	// return  c.BuildInMapCache.Set(ctx, key, val, expiration)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.data[key]
	if !ok {
		if c.cnt+1 > c.maxCnt {
			return errOverCapacity
		}
		c.cnt++
	}
	return c.BuildInMapCache.set(key, val, expiration)
}
