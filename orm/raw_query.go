package orm

import (
	"context"
	"database/sql"
)

type RawQuerier[T any] struct {
	core
	sess Session
	sql  string
	args []any
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	core := sess.getCore()
	return &RawQuerier[T]{
		sql:  query,
		args: args,
		sess: sess,
		core: core,
	}
}

func (r RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	var t *T
	if val, ok := res.Result.(*T); ok {
		t = val
	}
	return t, res.Err
}

func (r RawQuerier[T]) GetMulti(ctx context.Context) (*T, error) {
	panic("")
}

func (r RawQuerier[T]) Exec(ctx context.Context) Result {
	var err error
	var result Result
	r.model, err = r.r.Get(new(T))
	if err != nil {
		result.err = err
		return result
	}

	res := exec(ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	var sqlRes sql.Result
	if val, ok := res.Result.(sql.Result); ok {
		sqlRes = val
	}
	result.res = sqlRes
	result.err = res.Err
	return result
}
