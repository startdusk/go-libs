package micro

import (
	"context"
	"github.com/startdusk/go-libs/micro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	*grpc.Server
	name            string
	registry        registry.Registry
	registryTimeout time.Duration
}

type ServerOption func(s *Server)

func NewServer(name string, opts ...ServerOption) *Server {
	s := &Server{
		name:   name,
		Server: grpc.NewServer(),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

// 当用户调用Start的时候，就是服务已经准备好了
func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 有注册中心，要注册
	if s.registry != nil {
		// 在这里注册
		ctx, cancel := context.WithTimeout(context.Background(), s.registryTimeout)
		defer cancel()
		if err := s.registry.Register(ctx, registry.ServiceInstance{
			Name:    s.name,
			Address: lis.Addr().String(),
		}); err != nil {
			return err
		}
		// 到这里已经注册成功了
		// defer func() {
		// 	// 忽略或者log一下错误
		// 	_ = s.registry.Close()
		// }()
	}

	return s.Serve(lis)
}

// 先关registry再关listener
// 因为关registry的时候会有请求进来
func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	// 会帮我们关闭listener
	s.GracefulStop()
	return nil
}

func ServerWithRegistry(r registry.Registry) ServerOption {
	return func(s *Server) {
		s.registry = r
	}
}
