package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx Context)

var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	Start(addr string) error
}

type HTTPServer struct {
	*router
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
	}
}

// ServeHTTP HTTPServer 处理请求入口
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req:  r,
		Resp: w,
	}

	h.serve(ctx)
}

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 这里不直接使用 http.ListenAndSerer()
	// 目的是为了注册各种 生命周期函数

	return http.Serve(l, h)
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) serve(ctx *Context) error {
	// 查找路由, 并且执行命中的业务逻辑
	return nil
}
