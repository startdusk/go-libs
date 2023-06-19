package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
)

func TestInserter_SQLite_Upsert(t *testing.T) {
	db := memoryDB(t, DBWithDialect(DialectSQLite))
	cases := []struct {
		name      string
		i         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name: "upsert",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).Upsert().ConflictColumns("ID").Update(Assign("FirstName", "Ben"), Assign("Age", 17)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON CONFLICT(`id`) DO UPDATE SET `first_name`=?,`age`=?;",
				Args: []any{
					int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					"Ben", 17,
				},
			},
		},
		{
			name: "upsert use insert value",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).Upsert().ConflictColumns("ID").Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON CONFLICT(`id`) DO UPDATE SET `first_name`=excluded.`first_name`,`age`=excluded.`age`;",
				Args: []any{
					int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q, err := c.i.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.wantQuery, q)
		})
	}
}

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	cases := []struct {
		name      string
		i         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name:    "insert zero row",
			i:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRows,
		},
		{
			name: "insert row with unknown field",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).Columns("unknown"),
			wantErr: errs.NewErrUnknownField("unknown"),
		},
		{
			name: "upsert",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).Upsert().Update(Assign("FirstName", "Ben"), Assign("Age", 17)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=?,`age`=?;",
				Args: []any{
					int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					"Ben", 17,
				},
			},
		},
		{
			name: "insert single columns row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).Columns("ID", "FirstName"),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`) VALUES (?,?);",
				Args: []any{int64(1), "Tom"},
			},
		},
		{
			name: "insert multiple columns row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
			},
				&TestModel{
					ID:        2,
					FirstName: "Tom1",
				},
			).Columns("ID", "FirstName"),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`) VALUES (?,?),(?,?);",
				Args: []any{
					int64(1), "Tom",
					int64(2), "Tom1",
				},
			},
		},
		{
			name: "only insert single row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}},
			},
		},
		{
			name: "insert multiple row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			},
				&TestModel{
					ID:        2,
					FirstName: "Tom1",
					Age:       19,
					LastName:  &sql.NullString{String: "Jerry1", Valid: true},
				},
			),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{
					int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(2), "Tom1", int8(19), &sql.NullString{String: "Jerry1", Valid: true},
				},
			},
		},
		{
			name: "upsert-update multiple row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			},
				&TestModel{
					ID:        2,
					FirstName: "Tom1",
					Age:       19,
					LastName:  &sql.NullString{String: "Jerry1", Valid: true},
				},
			).Upsert().Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{
					int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(2), "Tom1", int8(19), &sql.NullString{String: "Jerry1", Valid: true},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q, err := c.i.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.wantQuery, q)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	db, err := OpenDB(mockDB, DBWithDialect(DialectMySQL))
	assert.NoError(t, err)

	cases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec("INSERT INTO .*").WillReturnError(errors.New("db error"))
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "query error",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Values(&TestModel{}).Columns("Invalid")
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "exec",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO .*").WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			affected: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := c.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, c.wantErr, err)
			if c.wantErr != nil {
				return
			}
			assert.Equal(t, c.affected, affected)
		})
	}
}
