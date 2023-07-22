package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/startdusk/go-libs/micro/proto/gen"
	"github.com/startdusk/go-libs/micro/rpc/message"
	"github.com/startdusk/go-libs/micro/rpc/serialize/json"
)

func Test_setFuncField(t *testing.T) {

	s := &json.Serializer{}
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
				req := &message.Request{
					ServiceName: "user-service",
					MethodName:  "GetByID",
					Serializer:  s.Code(),
					Data:        []byte(`{"ID":123}`),
				}
				req.CalculateHeaderLength()
				req.CalculateBodyLength()
				p.EXPECT().Invoke(gomock.Any(), req).Return(&message.Response{
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

			err := setFuncField(c.service, c.mock(ctrl), s)
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

	// 测试proto协议, 由protoc生成go代码
	GetByIDProto func(ctx context.Context, req *gen.GetByIDReq) (*gen.GetByIDResp, error)
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

type UserServiceServer struct {
	Msg string
	Err error
}

func (u UserServiceServer) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	return &GetByIDResp{
		Msg: u.Msg,
	}, u.Err
}

func (u UserServiceServer) GetByIDProto(ctx context.Context, req *gen.GetByIDReq) (*gen.GetByIDResp, error) {
	return &gen.GetByIDResp{
		User: &gen.User{
			Id:   req.Id,
			Name: u.Msg,
		},
	}, u.Err
}

func (u UserServiceServer) Name() string {
	return "user-service"
}

type UserServiceServerTimeout struct {
	t     *testing.T
	sleep time.Duration
	Msg   string
	Err   error
}

func (u UserServiceServerTimeout) Name() string {
	return "user-service"
}

func (u UserServiceServerTimeout) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	if _, ok := ctx.Deadline(); !ok {
		u.t.Fatal("没有设置超时")
	}
	time.Sleep(u.sleep)
	return &GetByIDResp{
		Msg: u.Msg,
	}, u.Err
}
