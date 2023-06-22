package orm

import (
	"context"
	"database/sql"
	"errors"
)

var (
	_ Session = &Tx{}
)

type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// 尝试回滚, 如果此时事务已经提交了, 或者被回滚掉了, 那么
// 就会得到sql.ErrTxDone错误, 这时候忽略这个错误就好
func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}
