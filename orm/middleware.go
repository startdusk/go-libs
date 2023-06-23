package orm

import (
	"context"
)

type QueryContext struct {
	// Type 声明查询类型 即 SELECT, UPDATE, DELETE 和 INSERT
	Type string

	// Builder 使用的时候, 大多数情况下你需要转换到具体的类型才能篡改查询
	Builder QueryBuilder
}

type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type QueryResult struct {
	// Result 在不同的查询里面, 类型是不同的
	// Selector.Get里面, 这会是单个结果
	// Selector.GetMulti, 这会是一个切片
	// 其他情况下, 它是Result类型
	Result any
	Err    error
}
