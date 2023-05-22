package orm

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Selector_Build(t *testing.T) {
	cases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "select_from",
			builder: (&Selector[TestModel]{}).From("`TEST_MODEL`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TEST_MODEL`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_from_db",
			builder: (&Selector[TestModel]{}).From("`my_db`.`TEST_MODEL`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `my_db`.`TEST_MODEL`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_empty_from",
			builder: (&Selector[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_empty_where",
			builder: (&Selector[TestModel]{}).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_no_from",
			builder: &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_from_TestModel_where_age=18",
			builder: (&Selector[TestModel]{}).Where(C(`age`).Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `age` = ?;",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_TestModel_where_not(age=18)",
			builder: (&Selector[TestModel]{}).Where(Not(C(`age`).Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_TestModel_where_age=18_and_name=tom",
			builder: (&Selector[TestModel]{}).Where(C(`age`).Eq(18).And(C("name").Eq("tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`age` = ?) AND (`name` = ?);",
				Args: []any{18, "tom"},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_TestModel_where_age=18_or_name=tom",
			builder: (&Selector[TestModel]{}).Where(C(`age`).Eq(18).Or(C("name").Eq("tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`age` = ?) OR (`name` = ?);",
				Args: []any{18, "tom"},
			},
			wantErr: nil,
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
