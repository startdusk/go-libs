package orm

import (
	"database/sql"
	"testing"

	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
)

func Test_Selector_Build(t *testing.T) {
	db, err := NewDB()
	assert.NoError(t, err)
	cases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "select_from",
			builder: NewSelector[TestModel](db).From("`TEST_MODEL`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TEST_MODEL`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_from_db",
			builder: NewSelector[TestModel](db).From("`my_db`.`TEST_MODEL`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `my_db`.`TEST_MODEL`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_empty_from",
			builder: NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_empty_where",
			builder: NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_no_from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_from_test_model_where_age=18",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_test_model_where_not(age=18)",
			builder: NewSelector[TestModel](db).Where(Not(C(`Age`).Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_test_model_where_age=18_and_first_name=tom",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18).And(C("FirstName").Eq("tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "tom"},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_test_model_where_age=18_or_first_name=tom",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18).Or(C("FirstName").Eq("tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "tom"},
			},
			wantErr: nil,
		},
		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18).Or(C("XXX").Eq("tom"))),

			wantErr: errs.NewErrUnknownField("XXX"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q, err := c.builder.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantQuery, q)
		})
	}
}

type TestModel struct {
	ID        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
