package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	redis "github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 你没有持有锁")
)

//go:embed lua/unlock.lua
var luaUnlock string

//go:embed lua/refresh.lua
var luaRefresh string

//go:embed lua/lock.lua
var luaLock string

// Client 是对redis.Cmdable的封装
type Client struct {
	client redis.Cmdable
	g      *singleflight.Group
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
		g:      &singleflight.Group{},
	}
}

// SingleflightLock 在并发非常高的情况下，可以考虑结合singleflight来进行优化。也就是说
// 本地所有的goroutine自己先竞争一把，胜利者再去抢全局的分布式锁
// 除了装逼 面试，实际情况我们不会用
func (c *Client) SingleflightLock(ctx context.Context,
	key string,
	expiration time.Duration,
	timeout time.Duration,
	retry RetryStrategy,
) (*Lock, error) {
	for {
		var flag bool = false // 标记是不是自己拿到了锁
		resCh := c.g.DoChan(key, func() (any, error) {
			flag = true
			return c.Lock(ctx, key, expiration, timeout, retry)
		})
		select {
		case res := <-resCh:
			if flag { // 确实是自己拿到了锁
				c.g.Forget(key) // 为了确保下一次还触发 DoChan, 让另一个goroutine执行
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <-ctx.Done(): // 监听超时
			return nil, ctx.Err()
		}
	}
}

// Lock 有重试
func (c *Client) Lock(ctx context.Context,
	key string,
	expiration time.Duration,
	timeout time.Duration,
	retry RetryStrategy,
) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {
		// 在这里重试
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if res == "OK" {
			// 说明加锁成功
			return &Lock{
				key:        key,
				val:        val,
				client:     c.client,
				expiration: expiration,
				unlockChan: make(chan struct{}),
			}, nil
		}
		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("redis-lock: 超过重试限制, %w", ErrFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}
		select {
		case <-timer.C:
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// TryLock 没有重试
func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	// 使用uuid用来表示给锁一个标识
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		// 代表的是别人抢到了锁
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		key:        key,
		val:        val,
		client:     c.client,
		expiration: expiration,
		unlockChan: make(chan struct{}),
	}, nil
}

type Lock struct {
	client     redis.Cmdable
	key        string
	val        string
	expiration time.Duration
	unlockChan chan struct{}
}

func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.val).Int64()
	defer func() {
		if l.unlockChan == nil {
			return
		}
		close(l.unlockChan)
		l.unlockChan = nil
	}()
	// if err == redis.Nil {
	// 	return ErrLockNotHold
	// }
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.val, l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

// 自动续约
// interval 间隔多久续约一次
// timeout 超时
func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutChan := make(chan struct{}, 1)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-timeoutChan:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-l.unlockChan:
			return nil
		}
	}
}
