package cache

import (
	"context"
	"time"
	"math/rand"
)

// 缓存雪崩解决方案
// 缓存雪崩: 同一个时刻，大量key过期，查询都要打到数据库
// 解决方案: 在设置key过期时间的时候，加上一个随机的偏移量，保证不在同一个时刻过期
type RandomExpirationCache struct {
	Cache
}

func (r *RandomExpirationCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	if expiration > 0 {
		// 加上一个 [0, 300s) 的偏移量
		offset := time.Duration(rand.Intn(300)) * time.Second
		expiration = expiration + offset
	}
	return r.Cache.Set(ctx, key, val, expiration)
}
