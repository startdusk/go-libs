package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
	"testing"
)

func Test_setFuncField(t *testing.T) {
	cases := []struct {
		name    string
		service Service

		mock    func(ctrl *gomock.Controller) Proxy
		wantErr error
	}{
		{
			name:    "nil",
			service: nil,
			mock: func(ctrl *gomock.Controller) Proxy {
				return NewMockProxy(ctrl)
			},
			wantErr: errors.New("rpc: 不支持nil"),
		},
		{
			name:    "no pointer",
			service: UserService{},
			mock: func(ctrl *gomock.Controller) Proxy {
				return NewMockProxy(ctrl)
			},
			wantErr: errors.New("rpc: 只支持指向结构体的一级指针"),
		},
		{
			name: "user serive",
			mock: func(ctrl *gomock.Controller) Proxy {
				p := NewMockProxy(ctrl)
				p.EXPECT().Invoke(gomock.Any(), &Request{
					ServiceName: "user-service",
					MethodName:  "GetByID",
					Arg:         []byte(`{"ID":123}`),
				}).Return(&Response{
					Data: []byte(`{"Msg":"recieved 123"}`),
				}, nil)
				return p
			},
			service: &UserService{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			err := setFuncField(c.service, c.mock(ctrl))
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			resp, err := c.service.(*UserService).GetByID(context.Background(), &GetByIDReq{ID: 123})
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			t.Log(resp)
		})
	}
}

type mockProxy struct{}

type UserService struct {
	// 用反射来赋值
	// 类型是函数的字段，它不是方法(它不是定义在UserService上的方法)
	// 本质上是一个字段
	GetByID func(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIDReq struct {
	ID int
}

type GetByIDResp struct {
	Msg string
}

type UserServiceServer struct{}

func (u UserServiceServer) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	return &GetByIDResp{
		Msg: fmt.Sprintf("recieved msg: %d", req.ID),
	}, nil
}

func (u UserServiceServer) Name() string {
	return "user-service"
}
