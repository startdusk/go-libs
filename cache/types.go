package cache

import (
	"context"
	"time"
)

// 为什么不用泛型
// type Cache[T any] interface
// 由于Golang泛型的缺陷, 使用泛型只能用一种类型, 但缓存是会缓存多种类型, 使用any + 类型转换更合适
type Cache interface {
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error

	// 同时把删除的数据给返回
	// Delete(key string) (any, error)

}
