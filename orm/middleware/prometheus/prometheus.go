package prometheus

import (
	"context"
	"github.com/startdusk/go-libs/orm"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      m.Name,
		Subsystem: m.Subsystem,
		Namespace: m.Namespace,
		Help:      m.Help,

		// 设置指标 如 0.5: 0.01 0.5是一个指标，0.01是一个误差值，表示0.5上下0.01 即误差范围为 0.49-0.51
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{
		"pattern", // 命中的路由
		"method",  // 命中的http method
		"status",  // http状态码
	})

	prometheus.MustRegister(vector)

	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				go func() {
					duration := time.Now().Sub(startTime).Milliseconds()
					// 记录执行时间
					vector.WithLabelValues(qc.Type, qc.Model.TableName).Observe(float64(duration))
				}()
			}()
			return next(ctx, qc)
		}
	}
}
