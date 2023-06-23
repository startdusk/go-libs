package opentelemetry

import (
	"context"
	"fmt"

	"github.com/startdusk/go-libs/orm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/startdusk/go-libs/orm/middlewares/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			// span name: SELECT-TABLE_NAME
			tableName := qc.Model.TableName
			spanCtx, span := m.Tracer.Start(ctx, fmt.Sprintf("%s-%s", qc.Type, tableName))
			defer span.End()

			q, _ := qc.Builder.Build()
			if q != nil {
				span.SetAttributes(attribute.String("sql", q.SQL))
				// tracing这里没必要记录参数, 防止数据过大(如 blob), 防止敏感数据被记录到tracing(如 用户密码)
			}
			span.SetAttributes(attribute.String("table", tableName))
			span.SetAttributes(attribute.String("component", "orm"))

			res := next(spanCtx, qc)
			if res.Err != nil {
				span.RecordError(res.Err)
			}
			return res
		}
	}
}
