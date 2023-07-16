package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Server struct {
	services map[string]reflectionStub
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]reflectionStub, 16), // 16是预估值
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
}

func (s *Server) Start(network string, addr string) error {
	lis, err := net.Listen(network, addr)
	if err != nil {
		// 比较常见的错误就是端口被占用
		return err
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		go func() {
			if err := s.handleConn(conn); err != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		data, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}

		resp, err := s.Invoke(context.Background(), &req)
		if err != nil {
			// 可能是你的业务error
			// 暂时不知道怎么处理的error
			return err
		}
		res := EncodeMsg(resp.Data)

		_, err = conn.Write(res)
		return err
	}
}

func (s *Server) Invoke(ctx context.Context, req *Request) (*Response, error) {
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("rpc: 你要调用的服务不存在")
	}

	resp, err := service.invoke(ctx, req.MethodName, req.Arg)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: resp,
	}, nil
}

type reflectionStub struct {
	s     Service
	value reflect.Value
}

func (s *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// 通过反射找到方法, 并且执行调用
	method := s.value.MethodByName(methodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	if err := json.Unmarshal(data, inReq.Interface()); err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)
	// results[0] 是返回值
	// results[1] 是error
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}
	return json.Marshal(results[0].Interface())
}
