package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"time"

	"github.com/silenceper/pool"

	"github.com/startdusk/go-libs/micro/rpc/message"
)

// InitClientProxy 要为 GetByID 之类的函数类型字段赋值
func InitClientProxy(addr string, service Service) error {
	client, err := NewClient(addr)
	if err != nil {
		return err
	}
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
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqData,
				}
				req.CalculateHeaderLength()
				req.CalculateBodyLength()

				// 真正发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					// 这里相当于返回 (类型的零值, error)
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				var serverErr error
				if len(resp.Error) > 0 {
					// 服务端返回的error
					serverErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					if err := json.Unmarshal(resp.Data, retVal.Interface()); err != nil {
						return []reflect.Value{
							retVal,
							reflect.ValueOf(err),
						}
					}
				}

				var retErr reflect.Value
				if serverErr != nil {
					retErr = reflect.ValueOf(serverErr)
				} else {
					// 返回 nil(写法很独特)
					retErr = reflect.Zero(reflect.TypeOf(new(error)).Elem())

				}

				return []reflect.Value{
					retVal,
					retErr,
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
	pool pool.Pool
}

func NewClient(addr string) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Minute,
		Factory: func() (any, error) {
			return net.DialTimeout("tcp", addr, 3*time.Second)
		},
		Close: func(i any) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		pool: p,
	}, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = conn.Close()
	}()

	if _, err := conn.Write(data); err != nil {
		return nil, err
	}

	return ReadMsg(conn)
}
