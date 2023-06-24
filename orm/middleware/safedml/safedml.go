package nodelete

import (
	"context"
	"fmt"
	"github.com/startdusk/go-libs/orm"
	"strings"
)

// 强制要执行的SQL语句
// 1.SELECT, UPDATE, DELETE必须带WHERE
// 2.或者UPDATE, DELETE必须带WHERE(SELECT要不要带自己抉择)
type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder(fn func(query string, args []any)) *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {

			if qc.Type == "SELECT" || qc.Type == "INSERT" {
				return next(ctx, qc)
			}
			q, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			if !strings.Contains(q.SQL, "WHERE") {
				return &orm.QueryResult{
					Err: fmt.Errorf("禁止执行没有WHERE的 %s 语句", qc.Type),
				}
			}
			return next(ctx, qc)
		}
	}
}
