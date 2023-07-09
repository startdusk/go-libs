package cache

import (
	"context"
	"time"
	
	"golang.org/x/sync/singleflight"
)

// singleflight 多数用于读，写很少
// 能缓解缓存穿透问题，但如果是黑客伪造不存在的key，就没办法了
type SingleflightCacheV1 struct {
	ReadThroughCache
}

func NewSingleflightCacheV1(cache Cache, loadFunc func(ctx context.Context, key string) (any, error), expiration time.Duration) *SingleflightCacheV1 {
	g := &singleflight.Group{}
	return &SingleflightCacheV1{
		ReadThroughCache: ReadThroughCache{
			cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				val, err, _ := g.Do(key, func() (any, error) {
					return loadFunc(ctx, key)
				})
				return val, err
			},
			Expiration: expiration
		},
	}
}

// 侵入式写法，不推荐
type SingleflightCacheV2 struct {
	ReadThroughCache
	g singleflight.Group
}

func (r *SingleflightCacheV2) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err, _ = r.g.Do(key, func() (any, error) {
			v, err := r.LoadFunc(ctx, key)
			if err == nil {
				if err := r.Cache.Set(ctx, key, val, r.Expiration); err != nil {
					return v, fmt.Errorf("%w, 原因: %s", ErrFailedToRefreshCache, err)
				}
			}
			return v, err
		})
		return val, err
	}
	return val, err
}

