package cache

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

var (
	ErrFailedToRefreshCache = errors.New("刷新缓存失败")
)

// 缓存模式 read-through 模式
// 缓存中读不到数据就去数据库拿, 拿到后设置到缓存里面

// ReadThroughCache 一定要赋值 LoadFunc 和 Expiration
// Expiration 是你的过期时间
type ReadThroughCache struct {
	Cache
	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration

	// g singleflight.Group
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			if err := r.Cache.Set(ctx, key, val, r.Expiration); err != nil {
				return val, fmt.Errorf("%w, 原因: %s", ErrFailedToRefreshCache, err)
			}
		}
		return val, err
	}
	return val, err
}

// 全异步(找不到数据, 就异步去数据库获取数据, 当前请求不会返回数据, 需要用户重刷)
func (r *ReadThroughCache) GetV1(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		go func() {
			val, err = r.LoadFunc(ctx, key)
			if err == nil {
				if err := r.Cache.Set(ctx, key, val, r.Expiration); err != nil {
					log.Println(err)
				}
			}
		}()
	}
	return val, err
}

// 半异步(从数据库获取到数据后再异步)
func (r *ReadThroughCache) GetV2(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		go func() {
			if err == nil {
				if err := r.Cache.Set(ctx, key, val, r.Expiration); err != nil {
					log.Println(err)
				}
			}
		}()
		return val, err
	}
	return val, err
}

// 侵入式的写法，不推荐
// func (r *ReadThroughCache) GetV3(ctx context.Context, key string) (any, error) {
// 	val, err := r.Cache.Get(ctx, key)
// 	if err == errKeyNotFound {
// 		val, err, _ = r.g.Do(key, func() (any, error) {
// 			v, err := r.LoadFunc(ctx, key)
// 			if err == nil {
// 				if err := r.Cache.Set(ctx, key, val, r.Expiration); err != nil {
// 					return v, fmt.Errorf("%w, 原因: %s", ErrFailedToRefreshCache, err)
// 				}
// 			}
// 			return v, err
// 		})
// 		return val, err
// 	}
// 	return val, err
// }
