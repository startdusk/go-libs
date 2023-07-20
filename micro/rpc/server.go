package rpc

import (
	"context"
	"errors"
	"net"
	"reflect"

	"github.com/startdusk/go-libs/micro/rpc/message"

	"github.com/startdusk/go-libs/micro/rpc/serialize"
	"github.com/startdusk/go-libs/micro/rpc/serialize/json"
)

type Server struct {
	services    map[string]reflectionStub
	serializers map[uint8]serialize.Serializer // 客户端一般只有一种序列化协议，不同的客户端可以选不同的序列化协议
}

func NewServer() *Server {
	s := &Server{
		services:    make(map[string]reflectionStub, 16),     // 16是预估值
		serializers: make(map[uint8]serialize.Serializer, 4), // 4是预估值, 4种序列化协议顶天了
	}
	s.RegisterSerializer(&json.Serializer{})
	return s
}

func (s *Server) RegisterSerializer(serializer serialize.Serializer) {
	s.serializers[serializer.Code()] = serializer
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:           service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
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
		req := message.DecodeReq(data)

		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// 可能是你的业务error
			// 暂时不知道怎么处理的error
			resp.Error = []byte(err.Error())
		}

		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		if _, err := conn.Write(message.EncodeResp(resp)); err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	resp := &message.Response{
		RequestID:  req.RequestID,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}
	if !ok {
		return resp, errors.New("rpc: 你要调用的服务不存在")
	}

	data, err := service.invoke(ctx, req)
	resp.Data = data
	if err != nil {
		return resp, err
	}

	return resp, nil
}

type reflectionStub struct {
	s           Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer // 客户端一般只有一种序列化协议，不同的客户端可以选不同的序列化协议
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	// 通过反射找到方法, 并且执行调用
	method := s.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("micro: 不支持的序列化协议")
	}
	if err := serializer.Decode(req.Data, inReq.Interface()); err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)
	// results[0] 是返回值
	// results[1] 是error
	var err error
	if results[1].Interface() != nil {
		err = results[1].Interface().(error)
	}

	var res []byte
	if results[0].IsNil() {
		return nil, err
	} else {
		var serErr error
		res, serErr = serializer.Encode(results[0].Interface())
		if serErr != nil {
			return nil, serErr
		}
	}
	return res, err
}
