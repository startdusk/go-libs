package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Server struct {
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service, 16), // 16是预估值
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
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
		// 读数据长度
		lenBytes := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBytes)
		if err != nil {
			return err
		}

		// 数据有多长
		length := binary.BigEndian.Uint64(lenBytes)
		bs := make([]byte, length)
		_, err = conn.Read(bs)
		if err != nil {
			return err
		}

		respData, err := s.handleMsg(bs)
		if err != nil {
			// 可能是你的业务error
			// 暂时不知道怎么处理的error
			return err
		}
		respLen := len(respData)

		// 构造响应数据
		// data = respLen 的 64位表示 + respData
		res := make([]byte, respLen+numOfLengthBytes)
		// 第一步:
		// 先把长度写进去前八个字节
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
		// 第二步:
		// 写入数据
		copy(res[numOfLengthBytes:], respData)
		_, err = conn.Write(res)
		return err
	}
}

func (s *Server) handleMsg(data []byte) ([]byte, error) {
	// 还原调用信息
	req := Request{}
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("rpc: 你要调用的服务不存在")
	}

	// 通过反射找到方法, 并且执行调用
	val := reflect.ValueOf(service)
	method := val.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	if err := json.Unmarshal(req.Arg, inReq.Interface()); err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)
	// results[0] 是返回值
	// results[1] 是error
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}
	resp, err := json.Marshal(results[0].Interface())
	return resp, err
}
