package querylog

import (
	"context"
	"database/sql"
	"testing"

	"github.com/startdusk/go-libs/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestQueryLog(t *testing.T) {
	var query string
	var args []any
	m := NewQueryLog(func(q string, as []any) {
		query = q
		args = as
	})

	db, err := orm.Open("sqlite3", "file:test_orm_mid.db?cache=shared&mode=memory", orm.DBWithMiddlewares(m.Build()))
	require.NoError(t, err)
	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("ID").Eq(12)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{12}, args)

	orm.NewInserter[TestModel](db).Values(&TestModel{ID: 18}).Columns("ID").Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`) VALUES (?);", query)
	assert.Equal(t, []any{int64(18)}, args)
}

type TestModel struct {
	ID        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

// func (t TestModel) CreateSQL() string {
// 	return `
// 		CREATE TABLE IF NOT EXISTS test_model (
// 			id INTEGER PRIMARY KEY,
// 			first_name TEXT NOT NULL,
// 			age INTEGER,
// 			last_name TEXT NOT NULL
// 		)
// 	`
// }
