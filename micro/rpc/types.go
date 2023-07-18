package rpc

import (
	"context"
	"github.com/startdusk/go-libs/micro/rpc/message"
)

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}
