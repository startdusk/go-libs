//go:build integration

package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rdb redis.Cmdable

func Test_Client_e2e_Lock(t *testing.T) {
	cases := []struct {
		name       string
		key        string
		before     func(t *testing.T)
		after      func(t *testing.T)
		expiration time.Duration
		timeout    time.Duration
		retry      RetryStrategy
		wantErr    error
		wantLock   *Lock
	}{
		{
			name:   "locked",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "lock_key1").Result()
				require.NoError(t, err)
				require.True(t, timeout >= 50*time.Second)
				_, err = rdb.Del(ctx, "lock_key1").Result()
				require.NoError(t, err)
			},
			key:        "lock_key1",
			expiration: time.Minute,
			timeout:    3 * time.Second,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
		},
		{
			name: "lock hold by other",
			before: func(t *testing.T) {
				// 模拟别人的锁
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "lock_key2", "other_lock_value2", 10*time.Second).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.GetDel(ctx, "lock_key2").Result()
				require.NoError(t, err)
				require.Equal(t, "other_lock_value2", res)
			},
			key:        "lock_key2",
			expiration: time.Minute,
			timeout:    3 * time.Second,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   3,
			},
			wantErr: fmt.Errorf("redis-lock: 超过重试限制, %w", ErrFailedToPreemptLock),
		},
		{
			name: "retry and locked",
			before: func(t *testing.T) {
				// 模拟别人的锁, 到期释放，重试拿到锁
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 这里时间 3s 设小于重试的时间
				res, err := rdb.Set(ctx, "lock_key3", "lock_value3", 3*time.Second).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "lock_key3").Result()
				require.NoError(t, err)
				require.True(t, timeout >= 50*time.Second)
				_, err = rdb.Del(ctx, "lock_key3").Result()
				require.NoError(t, err)
			},
			key:        "lock_key3",
			expiration: time.Minute,
			timeout:    3 * time.Second,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
		},
	}

	client := NewClient(rdb)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			lock, err := client.Lock(ctx, c.key, c.expiration, c.timeout, c.retry)
			assert.Equal(t, c.wantErr, err)
			if err == nil {
				assert.Equal(t, c.key, lock.key)
				assert.Equal(t, c.expiration, lock.expiration)
				assert.NotEmpty(t, lock.val)
				assert.NotNil(t, lock.client)
			}
			c.after(t)
		})
	}
}

func Test_Client_e2e_TryLock(t *testing.T) {
	cases := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		key        string
		expiration time.Duration
		wantErr    error
	}{
		{
			// 别人持有锁了
			name: "key exists",
			before: func(t *testing.T) {
				// 模拟别人有锁
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "value1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, "value1", res)

			},
			key:        "key1",
			expiration: time.Minute,
			wantErr:    ErrFailedToPreemptLock,
		},
		{
			// 加锁成功
			name: "locked",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key2").Result()
				require.NoError(t, err)
				// 加锁成功意味着你已经设置好了值
				assert.NotEmpty(t, res)

			},
			key:        "key2",
			expiration: time.Minute,
		},
	}

	client := NewClient(rdb)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			lock, err := client.TryLock(ctx, c.key, c.expiration)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.key, lock.key)
			assert.NotEmpty(t, lock.val)
			assert.NotNil(t, lock.client)
			c.after(t)
		})
	}
}

func Test_Client_e2e_Refresh(t *testing.T) {
	cases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *Lock
		wantErr error
	}{
		{
			name:   "lock not hold",
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			lock: &Lock{
				key:        "refresh_key1",
				val:        "123",
				client:     rdb,
				expiration: time.Minute,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "lock hold by other",
			before: func(t *testing.T) {
				// 模拟别人的锁, value不相同
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "refresh_key2", "other_refresh_value2", 10*time.Second).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 锁被释放 key 不存在
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "refresh_key2").Result()
				require.NoError(t, err)
				// 如果刷新成功了，过期时间是一分钟，即便考虑测试本身的时间，timeout > 10s
				// 也就是，如果 timeout <= 10s 说明没有刷新成功
				require.True(t, timeout <= 10*time.Second)
				_, err = rdb.Del(ctx, "refresh_key2").Result()
				require.NoError(t, err)
			},
			lock: &Lock{
				key:        "refresh_key2",
				val:        "refresh_value2",
				client:     rdb,
				expiration: time.Minute,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "refreshed",
			before: func(t *testing.T) {
				// 模拟你自己加的锁，value和自己的相同
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "refresh_key3", "123", 10*time.Second).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 没释放锁，键值对不变
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "refresh_key3").Result()
				require.NoError(t, err)
				// 也就是，如果 timeout > 50s 说明刷新成功
				require.True(t, timeout > 50*time.Second)
				_, err = rdb.Del(ctx, "refresh_key3").Result()
				require.NoError(t, err)
			},
			lock: &Lock{
				key:        "refresh_key3",
				val:        "123",
				client:     rdb,
				expiration: time.Minute,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err := c.lock.Refresh(ctx)
			assert.Equal(t, c.wantErr, err)
			c.after(t)
		})
	}

}

func Test_Client_e2e_Unlock(t *testing.T) {
	cases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *Lock
		wantErr error
	}{
		{
			name:   "lock not hold",
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			lock: &Lock{
				key:    "unlock_key1",
				val:    "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlocked",
			before: func(t *testing.T) {
				// 模拟你自己加的锁，value和自己的相同
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key2", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 锁被释放 key 不存在
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Exists(ctx, "unlock_key2").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), res)
			},
			lock: &Lock{
				key:    "unlock_key2",
				val:    "123",
				client: rdb,
			},
		},
		{
			name: "lock hold by other",
			before: func(t *testing.T) {
				// 模拟别人的锁, 值不相同, 说明锁不是你的
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key3", "value3", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 没释放锁，键值对不变
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.GetDel(ctx, "unlock_key3").Result()
				require.NoError(t, err)
				assert.Equal(t, "value3", res)
			},
			lock: &Lock{
				key:    "unlock_key3",
				val:    "my_value", // val 是标记是不是自己的所持有的锁
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err := c.lock.Unlock(ctx)
			assert.Equal(t, c.wantErr, err)
			c.after(t)
		})
	}

}

func TestMain(m *testing.M) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	defer func() {
		rdb.FlushAll(context.Background())
	}()
	m.Run()

}
