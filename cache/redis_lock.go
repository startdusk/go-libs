package cache

import (
	"context"
	_ "embed"
	"errors"
	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 你没有持有锁")
)

//go:embed lua/unlock.lua
var luaUnlock string

//go:embed lua/refresh.lua
var luaRefresh string

// Client 是对redis.Cmdable的封装
type Client struct {
	client redis.Cmdable
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

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
