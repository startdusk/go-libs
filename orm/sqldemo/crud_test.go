package sqldemo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test_crud.db?cache=shared&mode=memory")
	require.NoError(t, err)

	db.Ping()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test1_model(
			id INTEGER PRIMARY KEY,
			first_name TEXT NOT NULL,
			age INTEGER,
			last_name TEXT NOT NULL
		)
	`); err != nil {
		require.NoError(t, err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO test1_model(`id`, `first_name`, `age`, `last_name`) VALUES(?, ?, ?, ?)",
		1,
		"tom",
		18,
		"ben",
	); err != nil {
		require.NoError(t, err)
	}
}
