package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: 键不存在")
	errKeyExpired  = errors.New("cache: 键过期")
)

var _ Cache = new(BuildInMapCache)

type BuildInMapCacheOption func(cache *BuildInMapCache)

func BuildInMapCacheWithEvictedCallback(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = fn
	}
}

type BuildInMapCache struct {
	data  map[string]*Item
	mutex sync.RWMutex
	close chan struct{}

	// 变更通知（回调函数)
	onEvicted func(key string, val any)
}

func NewBuildInMapCache(interval time.Duration, opts ...BuildInMapCacheOption) *BuildInMapCache {
	b := &BuildInMapCache{
		data:  make(map[string]*Item, 100),
		close: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(b)
	}

	// 轮询删除过期的key
	// 定时轮询的缺陷, 不保证每个过期的key都能及时被删除
	// 所以需要用户获取该key的时候再检查是否过期
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case now := <-ticker.C:
				b.mutex.Lock()
				var i int
				for key, val := range b.data {
					if i > 1000 {
						// 1000 是需要实际压测的
						// 这里的 i>1000 是控制每次定时删除过期的key遍历的数据(减少耗时)
						// 又因为map的遍历是无序的, 所以能保证每次都是随机轮询过期的key
						break
					}
					// 设置了过期时间, 且已经过期
					// 频繁的创建 time.Now() 对象对性能影响很大
					if !val.deadline.IsZero() && val.deadline.Before(now) {
						b.delete(key)
					}
					i++
				}
				b.mutex.Unlock()

			case <-b.close:
				return
			}
		}
	}()

	return b
}

func (b *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.set(key, val, expiration)
	// if expiration > 0 {
	// 	// expiration == 0 是永不过期
	// 	// 过期后删除
	// 	// time.AfterFunc 并不是每一个函数都开一个goroutine去盯着过期时间
	// 	// 但key多了, goroutine也多
	// 	// 这些goroutine大部分时候都被阻塞, 浪费资源
	// 	time.AfterFunc(expiration, func() {
	// 		b.mutex.Lock()
	// 		defer b.mutex.Unlock()
	// 		val, ok := b.data[key]
	// 		if ok && !val.deadline.IsZero() && val.deadline.Before(time.Now()) {
	// 			delete(b.data, key)
	// 		}
	// 	})
	// }
}

func (b *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	b.mutex.RLock()
	val, ok := b.data[key]
	b.mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
	}

	now := time.Now()
	if !val.deadline.IsZero() && val.deadline.Before(now) {
		// double check
		b.mutex.Lock()
		defer b.mutex.Unlock()
		val, ok = b.data[key]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
		if !val.deadline.IsZero() && val.deadline.Before(now) {
			b.delete(key)
			// 过期和找不到 用户不应该区分这个
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
	}
	return val.val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.delete(key)
	return nil
}

func (b *BuildInMapCache) Close() error {
	close(b.close)
	return nil
}

func (b *BuildInMapCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	val, ok := b.data[key]
	if !ok {
		return nil, errKeyNotFound
	}
	b.delete(key)
	return val.val, nil
}

func (b *BuildInMapCache) delete(key string) {
	item, ok := b.data[key]
	if !ok {
		return
	}
	delete(b.data, key)
	if b.onEvicted != nil {
		b.onEvicted(key, item.val)
	}
}

func (b *BuildInMapCache) set(key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	b.data[key] = &Item{
		val:      val,
		deadline: dl,
	}
	return nil
}

type Item struct {
	val      any
	deadline time.Time // 过期的时间点
}
