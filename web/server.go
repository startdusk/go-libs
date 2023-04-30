package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	Start(addr string) error
}

type HTTPServer struct {
	router

	middlewares []Middleware
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

	// 最后一个是这个root
	root := h.serve

	// 然后这里就是利用最后一个不断往前回溯组装链条
	// 从后往前
	// 把后一个作为前一个的next构造好链条
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		root = h.middlewares[i](root)
	}
	// 这里执行的时候就是从前往后了
	root(ctx)
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

func (h *HTTPServer) serve(ctx *Context) {
	// 查找路由, 并且执行命中的业务逻辑
	route, ok := h.router.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || route.n.handler == nil {
		ctx.Resp.WriteHeader(http.StatusNotFound)
		ctx.Resp.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}
	ctx.PathParams = route.pathParams
	route.n.handler(ctx)
}
