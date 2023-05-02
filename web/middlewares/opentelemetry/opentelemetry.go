package opentelemetry

import (
	"github.com/startdusk/go-libs/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/startdusk/go-libs/web/middlewares/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m MiddlewareBuilder) Build() web.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			// 尝试和客户端的 trace 结合在一起
			reqCtx := otel.GetTextMapPropagator().Extract(ctx.Req.Context(), propagation.HeaderCarrier(ctx.Req.Header))
			_, span := m.Tracer.Start(reqCtx, "unknown")
			defer func() {
				// 只有执行完next, MatchedRoute才会有值
				span.SetName(ctx.MatchedRoute)
				// 把响应码加上去
				span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))

				span.End()
			}()

			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.scheme", ctx.Req.URL.Scheme))
			span.SetAttributes(attribute.String("http.host", ctx.Req.Host))
			next(ctx)
		}
	}
}
