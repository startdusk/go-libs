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
	// 可以看到该函数不支持多个HandleFunc(handleFunc ...HandleFunc)
	// 因为用户可以传nil, 而且多个HandleFunc之间如果要中断, 必须提供像gin类似的Abort()方法
	// 比较复杂, 且容易忘记添加
	addRoute(method string, path string, handleFunc HandleFunc)
}

type HTTPServer struct {
}

// ServeHTTP HTTPServer 处理请求入口
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context {
		Req: r,
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

func (h *HTTPServer) addRoute(method string, path string, handleFunc HandleFunc) {

}

func (h *HTTPServer) serve(ctx *Context) error { 
	// 查找路由, 并且执行命中的业务逻辑
	
}