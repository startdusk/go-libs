//go:build integration

package cache

import (
	"context"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Client_e2e_TryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
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

func Test_Client_e2e_Unlock(t *testing.T) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
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
				key: "unlock_key1",
				val: "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name:   "lock hold by other",
			before: func(t *testing.T) {
				// 模拟你自己加的锁
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key3", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after:  func(t *testing.T) {
				// 锁被释放 key 不存在
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Exists(ctx, "unlock_key3").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), res)
			},
			lock: &Lock{
				key: "unlock_key3",
				val: "123",
				client: rdb,
			},
		},
		{
			name:   "unlocked",
			before: func(t *testing.T) {
				// 模拟别人的锁, 值不相同, 说明锁不是你的
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key2", "value2", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after:  func(t *testing.T) {
				// 没释放锁，键值对不变
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := rdb.GetDel(ctx, "unlock_key2").Result()
				require.NoError(t, err)
				assert.Equal(t, "value2", res)
			},
			lock: &Lock{
				key: "unlock_key2",
				val: "123",
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
