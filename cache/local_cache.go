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

type BuildInMapCache struct {
	data  map[string]*Item
	mutex sync.RWMutex
	close chan struct{}
}

func NewBuildInMapCache(interval time.Duration) *BuildInMapCache {
	b := &BuildInMapCache{
		data:  make(map[string]*Item, 100),
		close: make(chan struct{}),
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
						delete(b.data, key)
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
	b.data[key] = &Item{
		val:      val,
		deadline: time.Now().Add(expiration),
	}

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

	return nil
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
			delete(b.data, key)
			// 过期和找不到 用户不应该区分这个
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
	}
	return val.val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.data, key)
	return nil
}

func (b *BuildInMapCache) Close() error {
	close(b.close)
	return nil
}

type Item struct {
	val      any
	deadline time.Time // 过期的时间点
}
