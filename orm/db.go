package orm

import (
	"context"
	"database/sql"
	"github.com/startdusk/go-libs/orm/internal/valuer"

	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/startdusk/go-libs/orm/model"
)

var (
	_ Session = &DB{}
)

type DBOption func(db *DB)

type DB struct {
	core
	db *sql.DB
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error, opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			rollbackErr := tx.Rollback()
			err = errs.NewErrFailedToRollbackTx(err, rollbackErr, panicked)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	// 执行过程中没有发生panic, 则标志位置为false
	panicked = false
	return err
}

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}

	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	newDB := &DB{
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySQL,
		},
		db: db,
	}

	for _, opt := range opts {
		opt(newDB)
	}

	return newDB, nil
}

func MustOpenDB(db *sql.DB, opts ...DBOption) *DB {
	newDB, err := OpenDB(db, opts...)
	if err != nil {
		panic(err)
	}
	return newDB
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	newDB, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return newDB
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}
