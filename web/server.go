package web

import (
	"log"
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

	logFunc func(msg string, args ...any)
}

type HTTPServerOption func(hs *HTTPServer)

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	hs := &HTTPServer{
		router: newRouter(),
		logFunc: func(msg string, args ...any) {
			log.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(hs)
	}

	return hs
}

func ServerWithLogFunc(logFunc func(msg string, args ...any)) HTTPServerOption {
	return func(hs *HTTPServer) {
		hs.logFunc = logFunc
	}
}

func ServerWithMiddleware(mids ...Middleware) HTTPServerOption {
	return func(hs *HTTPServer) {
		hs.middlewares = mids
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

	// 这里最后一个步骤, 就是把 RespData 和 RespStatusCode 刷新到响应里面
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			// 就设置好了 RespData 和 RespStatusCode
			h.flashResp(ctx)
		}
	}

	root = m(root)
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

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Put(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPut, path, handleFunc)
}

func (h *HTTPServer) Delete(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodDelete, path, handleFunc)
}

func (h *HTTPServer) Head(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodHead, path, handleFunc)
}

func (h *HTTPServer) Patch(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPatch, path, handleFunc)
}

func (h *HTTPServer) Connect(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodConnect, path, handleFunc)
}

func (h *HTTPServer) Trace(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodTrace, path, handleFunc)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodOptions, path, handleFunc)
}

func (h *HTTPServer) serve(ctx *Context) {
	// 查找路由, 并且执行命中的业务逻辑
	route, ok := h.router.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || route.n.handler == nil {
		ctx.RespStatusCode = http.StatusNotFound
		ctx.RespData = []byte(http.StatusText(http.StatusNotFound))
		return
	}
	ctx.PathParams = route.pathParams
	ctx.MatchedRoute = route.n.fullPath
	route.n.handler(ctx)
}

func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode > 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil {
		h.logFunc("写入响应数据错误: %v\n", err)
	}
	if n != len(ctx.RespData) {
		h.logFunc("写入响应数据错误: 写入数据长度为 %d, 实际写入为 %d\n", len(ctx.RespData), n)
	}
}
