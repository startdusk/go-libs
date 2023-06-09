package orm

import (
	"context"
	"database/sql"
)

// Querier 用于 `SELECT` 语句
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)

	// 返回值的形式也可以, 但一般返回指针
	// 返回指针 是允许在 AOP 的场景下修改返回值, 从而不引起数据拷贝
	// 虽然返回指针在一些场景下会有内存逃逸的问题, 但这是中间件开发, 没到那种极致性能的优化, 开发速度是第一优先
	// Get(ctx context.Context) (*T, error)
	// GetMulti(ctx context.Context) (*T, error)
}

// Executor 用于 `INSERT`, `UPDATE`, `DELETE` 语句
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

type QueryBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}
