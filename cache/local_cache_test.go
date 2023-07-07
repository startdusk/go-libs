package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_BuildInMapCache_Get(t *testing.T) {
	cases := []struct {
		name    string
		key     string
		cache   func() *BuildInMapCache
		wantVal any
		wantErr error
	}{
		{
			name: "key not found",
			key:  "not exist key",
			cache: func() *BuildInMapCache {
				return NewBuildInMapCache(10 * time.Second)
			},
			wantErr: fmt.Errorf("%w, key: %s", errKeyNotFound, "not exist key"),
		},
		{
			name: "expired key",
			key:  "expired key",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "expired key", 123, time.Second)
				require.NoError(t, err)
				time.Sleep(2 * time.Second)
				return res
			},
			wantErr: fmt.Errorf("%w, key: %s", errKeyNotFound, "expired key"),
		},
		{
			name: "get value",
			key:  "key1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "key1", 123, time.Minute)
				require.NoError(t, err)
				return res
			},
			wantVal: 123,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			val, err := c.cache().Get(context.Background(), c.key)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.wantVal, val)
		})
	}
}

func Test_BuildInMapCache_Loop(t *testing.T) {
	var cnt int
	c := NewBuildInMapCache(time.Second, BuildInMapCacheWithEvictedCallback(func(key string, val any) {
		cnt++
	}))

	err := c.Set(context.Background(), "key1", 123, time.Second)
	require.NoError(t, err)
	time.Sleep(time.Second)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.data["key1"]
	require.False(t, ok)
	require.Equal(t, 1, cnt)
}
