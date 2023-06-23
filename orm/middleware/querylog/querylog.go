package querylog

import (
	"context"
	"github.com/startdusk/go-libs/orm"
)

type MiddlewareBuilder struct {
	// 存在问题, SQL参数存在敏感数据不应该被打印出来
	// 使用 debug 标记为标记是否打印出参数(不推荐做法, 会入侵大面积代码)
	logFunc func(query string, args []any)
}

func NewMiddlewareBuilder(fn func(query string, args []any)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: fn,
	}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			q, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			if m.logFunc != nil {
				m.logFunc(q.SQL, q.Args)
			}
			return next(ctx, qc)
		}
	}
}
