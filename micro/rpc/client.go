package rpc

import (
	"context"
	"errors"
	"net"
	"reflect"
	"time"

	"github.com/silenceper/pool"

	"github.com/startdusk/go-libs/micro/rpc/message"
	"github.com/startdusk/go-libs/micro/rpc/serialize"
	"github.com/startdusk/go-libs/micro/rpc/serialize/json"
)

// InitService
func (c *Client) InitService(service Service) error {
	return setFuncField(service, c, c.serializer)
}

func setFuncField(service Service, p Proxy, s serialize.Serializer) error {
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
				reqData, err := s.Encode(args[1].Interface())
				if err != nil {
					// 这里相当于返回 (类型的零值, error)
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				var meta map[string]string
				if isOneway(ctx) {
					meta = map[string]string{"one-way": "true"}
				}
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Serializer:  s.Code(),
					Data:        reqData,
					Meta:        meta,
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
					if err := s.Decode(resp.Data, retVal.Interface()); err != nil {
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
	pool       pool.Pool
	serializer serialize.Serializer
}

type ClientOption func(c *Client)

func ClientWithSerializer(serializer serialize.Serializer) ClientOption {
	return func(c *Client) {
		c.serializer = serializer
	}
}

func NewClient(addr string, opts ...ClientOption) (*Client, error) {
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
	c := &Client{
		pool:       p,
		serializer: &json.Serializer{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	resp, err := c.send(ctx, data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) send(ctx context.Context, data []byte) ([]byte, error) {
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

	if isOneway(ctx) {
		return nil, errors.New("micro: 这是一个 oneway 调用, 你不应该处理任何结果")
	}

	return ReadMsg(conn)
}
