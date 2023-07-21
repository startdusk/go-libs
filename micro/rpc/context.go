package rpc

import (
	"context"
)

// 通过上下文标记为一次调用
// 所谓一次调用就是客户端发送请求到服务端，服务端接收后就不返回了(就去干活了，但没有返回，能有效减小开销)，尽早释放资源
type onewayKey struct{}

func CtxWithOneway(ctx context.Context) context.Context {
	return context.WithValue(ctx, onewayKey{}, true)
}

func isOneway(ctx context.Context) bool {
	val := ctx.Value(onewayKey{})
	oneway, ok := val.(bool)
	return ok && oneway
}
