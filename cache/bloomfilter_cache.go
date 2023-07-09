package cache

import (
	"context"
)

type BloomFilterCache struct {
	ReadThroughCache
}

func NewBloomFilterCache(cache Cache, bf BloomFilter, loadFunc func(ctx context.Context, key string) (any, error)) *BloomFilterCache {
	return &BloomFilterCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HashKey(ctx, key) {
					return nil, errKeyNotFound
				}
				return loadFunc(ctx, key)
			},
		},
	}
}

type BloomFilter interface {
	HashKey(ctx context.Context, key string) bool
}

type BloomFilterCacheV1 struct {
	ReadThroughCache
	BF BloomFilter
}

func (r *BloomFilterCacheV1) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound && r.BF.HashKey(ctx, key) {
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
