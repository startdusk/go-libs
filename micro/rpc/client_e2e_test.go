package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/startdusk/go-libs/micro/proto/gen"
	"github.com/startdusk/go-libs/micro/rpc/serialize/proto"
)

// 测试proto序列化
func Test_InitServiceProto(t *testing.T) {
	addr := ":8082"
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerializer(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", addr)
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(3 * time.Second)
	usClient := &UserService{}
	client, err := NewClient(addr, ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	cases := []struct {
		name     string
		mock     func()
		wantResp *GetByIDResp
		wantErr  error
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello world"
			},
			wantResp: &GetByIDResp{
				Msg: "hello world",
			},
		},
		{
			name: "no msg",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("this is a error")
			},
			wantErr:  errors.New("this is a error"),
			wantResp: &GetByIDResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Err = errors.New("this is a error")
				service.Msg = "hello world"
			},
			wantResp: &GetByIDResp{
				Msg: "hello world",
			},
			wantErr: errors.New("this is a error"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.mock()
			resp, err := usClient.GetByIDProto(context.Background(), &gen.GetByIDReq{Id: 123})
			assert.Equal(t, c.wantErr, err)
			if resp != nil && resp.User != nil {
				assert.Equal(t, c.wantResp.Msg, resp.User.Name)
			}
		})
	}
}

// 测试json序列化
func Test_InitClientProxy(t *testing.T) {
	addr := ":8081"
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", addr)
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(3 * time.Second)
	usClient := &UserService{}
	client, err := NewClient(addr)
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	cases := []struct {
		name     string
		mock     func()
		wantResp *GetByIDResp
		wantErr  error
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello world"
			},
			wantResp: &GetByIDResp{
				Msg: "hello world",
			},
		},
		{
			name: "no msg",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("this is a error")
			},
			wantErr:  errors.New("this is a error"),
			wantResp: &GetByIDResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Err = errors.New("this is a error")
				service.Msg = "hello world"
			},
			wantResp: &GetByIDResp{
				Msg: "hello world",
			},
			wantErr: errors.New("this is a error"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.mock()
			resp, err := usClient.GetByID(context.Background(), &GetByIDReq{ID: 123})
			assert.Equal(t, c.wantErr, err)
			assert.Equal(t, c.wantResp, resp)
		})
	}
}

// 测试一次调用(服务端没有返回值)
func Test_oneway(t *testing.T) {
	addr := ":8083"
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", addr)
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(3 * time.Second)
	usClient := &UserService{}
	client, err := NewClient(addr)
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	cases := []struct {
		name     string
		mock     func()
		wantResp *GetByIDResp
		wantErr  error
	}{
		{
			name: "oneway",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello world"
			},
			wantResp: &GetByIDResp{},
			wantErr:  errors.New("micro: 这是一个 oneway 调用, 你不应该处理任何结果"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.mock()
			ctx := CtxWithOneway(context.Background())
			resp, err := usClient.GetByID(ctx, &GetByIDReq{ID: 123})
			assert.Equal(t, c.wantErr, err)
			assert.Equal(t, c.wantResp, resp)
			time.Sleep(2 * time.Second)
			assert.Equal(t, "hello world", service.Msg)
		})
	}
}

func Test_timeout(t *testing.T) {
	addr := ":8084"
	server := NewServer()
	service := &UserServiceServerTimeout{t: t}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", addr)
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(3 * time.Second)
	usClient := &UserService{}
	client, err := NewClient(addr)
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	cases := []struct {
		name     string
		mock     func() (context.Context, func())
		wantResp *GetByIDResp
		wantErr  error
	}{
		{
			name: "timeout",
			mock: func() (context.Context, func()) {
				service.Err = errors.New("mock error")
				service.Msg = "hello world"
				// 服务端sleep 2s
				// 但是超时设置了1s，所以客户端预期拿到一个超时响应
				service.sleep = 2 * time.Second
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				return ctx, cancel
			},
			wantResp: &GetByIDResp{},
			wantErr:  context.DeadlineExceeded,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := c.mock()
			defer cancel()
			resp, err := usClient.GetByID(ctx, &GetByIDReq{ID: 123})
			assert.Equal(t, c.wantErr, err)
			assert.Equal(t, c.wantResp, resp)
			time.Sleep(2 * time.Second)
			assert.Equal(t, "hello world", service.Msg)
		})
	}
}
