package cache

import (
	"context"
	"fmt"
	"github.com/startdusk/go-libs/cache/mocks"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func Test_RedisCache_Set(t *testing.T) {
	cases := []struct {
		name string
		key  string
		// 单元测试不连redis
		mock       func(ctrl *gomock.Controller) redis.Cmdable
		value      string
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "set value",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetVal("OK")
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
		},
		{
			name: "timeout",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "unexpected msg",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				status := redis.NewStatusCmd(context.Background())
				status.SetVal("NO OK")
				cmd.EXPECT().Set(context.Background(), "key1", "value1", time.Second).Return(status)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Second,
			wantErr:    fmt.Errorf("%w, 返回信息 %s", errFailedToSetCache, "NO OK"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rc := NewRedisCache(c.mock(ctrl))
			err := rc.Set(context.Background(), c.key, c.value, c.expiration)
			assert.Equal(t, c.wantErr, err)
		})
	}
}

func Test_RedisCache_Get(t *testing.T) {
	cases := []struct {
		name string
		key  string
		// 单元测试不连redis
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		wantErr error
		wantVal string
	}{
		{
			name: "get value",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				stringCmd := redis.NewStringCmd(context.Background())
				stringCmd.SetVal("value1")
				cmd.EXPECT().Get(context.Background(), "key1").Return(stringCmd)
				return cmd
			},
			key:     "key1",
			wantVal: "value1",
		},
		{
			name: "timeout",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				stringCmd := redis.NewStringCmd(context.Background())
				stringCmd.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Get(context.Background(), "key1").Return(stringCmd)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			rc := NewRedisCache(c.mock(ctrl))
			val, err := rc.Get(context.Background(), c.key)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantVal, val)
		})
	}
}
