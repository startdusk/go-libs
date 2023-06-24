package safedml

import (
	"context"
	"errors"
	"github.com/startdusk/go-libs/orm"
)

type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder(fn func(query string, args []any)) *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			// 禁用 DELETE 语句
			if qc.Type == "DELETE" {
				return &orm.QueryResult{
					Err: errors.New("禁止使用DELETE语句"),
				}
			}
			return next(ctx, qc)
		}
	}
}
