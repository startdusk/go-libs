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

	tplEngine TemplateEngine
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

func ServerWithTemplateEngine(engine TemplateEngine) HTTPServerOption {
	return func(hs *HTTPServer) {
		hs.tplEngine = engine
	}
}

// ServeHTTP HTTPServer 处理请求入口
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req:       r,
		Resp:      w,
		tplEngine: h.tplEngine,
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

// 可路由的 Middleware
// 功能需求: 允许用户在特定路由上注册Middleware, Middleware选取所有能够匹配上的路由的Middleware作为结果
// 支持原则: 能匹配则匹配, 越具体的匹配越往后调度
//
// 支持以下场景
// 1.Use("GET", "/a/b", ms) 当输入路径/a/b的时候, 会调对应的ms, 输入路径/a/b/c, 也会调用对应的ms
//
// 2.Use("GET", "/a/*", ms1) 当输入路径/a/b的时候, 会调对应的ms
//
//	Use("GET", "/a/b/*", ms2) 当输入路径/a/b/c的时候, 会调对应的ms1和ms2
//
// 3.Use("GET", "/a/*/c", ms1) 当输入路径/a/d/c的时候, 会调对应的ms1
//
//	Use("GET", "/a/b/c", ms2) 当输入路径/a/b/c的时候, 会调对应的ms1和ms2
//
// 4.Use("GET", "/a/:id", ms1) 当输入路径/a/123的时候, 会调对应的ms1
//
//	Use("GET", "/a/123/c", ms1) 当输入路径/a/123/c的时候, 会调对应的ms1和ms2
//
// 不支持的场景
// 1.Use("GET", "/a/*/c", ms1) 当输入路径/a/b/c的时候, 不会调对应的ms1
//
//	Use("GET", "/a/b/c", ms2) 当输入路径/a/b/c的时候, 会调对应ms2
//
// 如果用户同时注册了Middleware
// 1.Use("GET", "/a/b", ms1)
// 2.Use("GET", "/a/*", ms2)
// 3.Use("GET", "/a", ms3)
// 那么调用顺序为 ms3, ms2, ms1
func (h *HTTPServer) Use(method string, path string, ms ...Middleware) {
	h.addRoute(method, path, nil, ms...) // 依托于原有的路由树来完成这个功能
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
