package orm

import (
	"database/sql"
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInserter_Build(t *testing.T) {
	db, err := OpenDB(memoryDB(t))
	assert.NoError(t, err)

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
