package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"time"
)

// InitClientProxy 要为 GetByID 之类的函数类型字段赋值
func InitClientProxy(addr string, service Service) error {
	client := NewClient(addr)
	return setFuncField(service, client)
}

func setFuncField(service Service, p Proxy) error {
	if service == nil {
		return errors.New("rpc: 不支持nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	// 只支持指向结构体的一级指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}
	val = val.Elem()
	typ = typ.Elem()
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)
		if fieldVal.CanSet() {
			fn := func(args []reflect.Value) []reflect.Value {
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
				// args[0] 是 context // context我们不会上传到服务端, 但context里面的数据可能会
				ctx := args[0].Interface().(context.Context)
				// args[1] 是 req
				reqData, err := json.Marshal(args[1].Interface())
				if err != nil {
					// 这里相当于返回 (类型的零值, error)
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         reqData,
				}

				// 真正发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					// 这里相当于返回 (类型的零值, error)
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}

				if err := json.Unmarshal(resp.Data, retVal.Interface()); err != nil {
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}

				return []reflect.Value{
					retVal,
					// 返回 nil(写法很独特)
					reflect.Zero(reflect.TypeOf(new(error)).Elem()),
				}
			}
			// 给结构体字段赋值
			fieldVal.Set(reflect.MakeFunc(fieldTyp.Type, fn))
		}
	}
	return nil
}

// 长度字段使用的字节数量
const numOfLengthBytes = 8

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: resp,
	}, nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", c.addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	reqLen := len(data)
	req := make([]byte, reqLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	copy(req[numOfLengthBytes:], data)

	if _, err := conn.Write(req); err != nil {
		return nil, err
	}

	lenBytes := make([]byte, numOfLengthBytes)
	if _, err := conn.Read(lenBytes); err != nil {
		return nil, err
	}

	// 响应有多长
	respLen := binary.BigEndian.Uint64(lenBytes)
	resp := make([]byte, respLen)
	if _, err := conn.Read(resp); err != nil {
		return nil, err
	}
	return resp, nil
}
