package prometheus

import (
	"github.com/startdusk/go-libs/web"
)

type MiddlewareBuilder struct {
	StatusCode int
	Data       []byte
	LogFunc    func(ctx *web.Context)
}

func (m MiddlewareBuilder) Build() web.Middleware {

	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespData = m.Data
					ctx.RespStatusCode = m.StatusCode
					m.LogFunc(ctx)
				}
			}()
			next(ctx)
		}
	}
}
