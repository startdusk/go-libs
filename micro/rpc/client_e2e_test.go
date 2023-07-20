package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

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
