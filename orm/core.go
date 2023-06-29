package orm

import (
	"context"
	"github.com/startdusk/go-libs/orm/model"

	"github.com/startdusk/go-libs/orm/internal/valuer"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry

	mdls []Middleware
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	qr := &QueryResult{}
	q, err := qc.Builder.Build()
	if err != nil {
		qr.Err = err
		return qr
	}

	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		qr.Err = err
		return qr
	}

	if !rows.Next() {
		// 返回要和sql包语义一致
		qr.Err = ErrNoRows
		return qr
	}

	// 利用 columns 来解决 select 的列顺序 和 列字段类型的问题
	entity := new(T)
	// 接口定义好之后, 就两件事情, 一个是利用新接口的方法改造上层
	// 一个是提供不同的实现
	val := c.creator(c.model, entity)
	err = val.SetColumns(rows)
	qr.Result = entity
	qr.Err = err
	return qr
}

func execHandler(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	qr := &QueryResult{}
	q, err := qc.Builder.Build()
	if err != nil {
		qr.Err = err
		qr.Result = Result{err: err}
		return qr
	}
	res, err := sess.execContext(ctx, q.SQL, q.Args...)
	qr.Err = err
	qr.Result = Result{res: res, err: err}
	return qr
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}
