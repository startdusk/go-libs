package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error
	ListServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(ctx context.Context, serviceName string) (<-chan Event, error)

	io.Closer
}

type ServiceInstance struct {
	Name string
	// Address 最关键, 定位信息
	Address string

	// 这边你可以加任意字段, 完全取决于你的服务治理需要什么字段

	Weight int // 权重
}

type Event struct {
}
