package cache

import (
	"context"
	"github.com/startdusk/go-libs/cache/mocks"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"

	gomock "go.uber.org/mock/gomock"
)

func Test_Client_TryLock(t *testing.T) {
	cases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) redis.Cmdable
		key      string
		wantErr  error
		wantLock *Lock
	}{
		{
			name: "set nx error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "failed to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key:        "key1",
				expiration: time.Minute,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := NewClient(c.mock(ctrl))
			l, err := client.TryLock(context.Background(), c.key, time.Minute)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantLock.key, l.key)
			assert.Equal(t, c.wantLock.expiration, l.expiration)
			assert.NotEmpty(t, l.val)
		})
	}
}

func Test_Client_Unlock(t *testing.T) {
	cases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		key     string
		value   string
		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlocked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).Return(res)
				return cmd
			},
			key:   "key1",
			value: "value1",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:    c.key,
				val:    c.value,
				client: c.mock(ctrl),
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, c.wantErr, err)
		})
	}
}

func Test_Client_Refresh(t *testing.T) {
	cases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) redis.Cmdable
		key        string
		value      string
		wantErr    error
		expiration time.Duration
	}{
		{
			name: "refresh eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"refresh_key1"}, []any{"value1", float64(60)}).Return(res)
				return cmd
			},
			key:        "refresh_key1",
			value:      "value1",
			expiration: time.Minute,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "refresh lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"refresh_key2"}, []any{"value2", float64(60)}).Return(res)
				return cmd
			},
			key:        "refresh_key2",
			value:      "value2",
			expiration: time.Minute,
			wantErr:    ErrLockNotHold,
		},
		{
			name: "refreshed",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaRefresh, []string{"refresh_key3"}, []any{"value3", float64(60)}).Return(res)
				return cmd
			},
			key:        "refresh_key3",
			value:      "value3",
			expiration: time.Minute,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:        c.key,
				val:        c.value,
				client:     c.mock(ctrl),
				expiration: c.expiration,
			}
			err := lock.Refresh(context.Background())
			assert.Equal(t, c.wantErr, err)
		})
	}
}
