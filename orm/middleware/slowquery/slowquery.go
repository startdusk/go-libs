package slowquery

import (
	"context"
	"github.com/startdusk/go-libs/orm"
	"time"
)

type MiddlewareBuilder struct {
	// 存在问题, SQL参数存在敏感数据不应该被打印出来
	// 使用 debug 标记为标记是否打印出参数(不推荐做法, 会入侵大面积代码)
	logFunc func(query string, args []any)

	// 慢查询阈值, 设置需要考虑公司实际情况, 如100ms
	threshold time.Duration
}

func NewMiddlewareBuilder(threshold time.Duration, fn func(query string, args []any)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc:   fn,
		threshold: threshold,
	}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				// 不是慢查询
				if duration <= m.threshold {
					return
				}

				// 是慢查询, 记录一下, 不处理错误(如果错误了, 证明SQL都没构造出来)
				q, err := qc.Builder.Build()
				if err == nil {
					if m.logFunc != nil {
						m.logFunc(q.SQL, q.Args)
					}
				}
			}()

			// 不调用next就是dry run
			return next(ctx, qc)
		}
	}
}
