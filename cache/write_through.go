package cache

import (
	"context"
	"log"
	"time"
)

// write-through 先写数据库 再写缓存
type WriteThroughCache struct {
	Cache
	StoreFunc func(ctx context.Context, key string, val any) error
}

// 比较符合一般人的预期, 先操作DB
func (w *WriteThroughCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	if err != nil {
		return err
	}
	return w.Cache.Set(ctx, key, val, expiration)
}

// 先写DB还是先写缓存没啥区别, 最终都是会不一致的(五十步笑百步的区别)
func (w *WriteThroughCache) SetV1(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.Cache.Set(ctx, key, val, expiration)
	if err != nil {
		return err
	}
	return w.StoreFunc(ctx, key, val)
}

func (w *WriteThroughCache) SetV2(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	if err != nil {
		return err
	}
	go func() {
		err := w.Cache.Set(ctx, key, val, expiration)
		if err != nil {
			log.Println(err)
		}
	}()
	return nil
}
