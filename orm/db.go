package orm

type DBOption func(db *DB)

type DB struct {
	r *registry
}

func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		r: newRegistry(),
	}

	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}

func MustNewDB(opts ...DBOption) *DB {
	db, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return db
}
