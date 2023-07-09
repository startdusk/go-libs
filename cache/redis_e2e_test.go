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

func Test_RedisCache_e2e_Set(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	c := NewRedisCache(rdb)
	err := c.Set(context.Background(), "key1", "abc", time.Minute)
	require.NoError(t, err)
	val, err := c.Get(context.Background(), "key1")
	require.NoError(t, err)
	assert.Equal(t, "abc", val)
}
