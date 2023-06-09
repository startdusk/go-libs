package orm

import (
	"database/sql"
	"github.com/startdusk/go-libs/orm/model"
)

type DBOption func(db *DB)

type DB struct {
	r  model.Registry
	db *sql.DB
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
		r:  model.NewRegistry(),
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
