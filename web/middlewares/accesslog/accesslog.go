package accesslog

import (
	"encoding/json"
	"github.com/startdusk/go-libs/web"
)

type MiddlewareBuilder struct {
	logFunc func(log string)
}

func (m *MiddlewareBuilder) LogFunc(fn func(log string)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				// next 可能会panic, 保险起见在defer里面记录日志
				al := accessLog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchedRoute,
					HTTPMethod: ctx.Req.Method,
					Path:       ctx.Req.URL.Path,
				}
				// Ignore error here, we are sure our data is good.
				data, _ := json.Marshal(al)
				m.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

type accessLog struct {
	Host       string `json:"host"`
	Route      string `json:"route"`
	HTTPMethod string `json:"http_method"`
	Path       string `json:"path"`
}
