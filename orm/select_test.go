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
			name:    "select_no_from",
			builder: &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
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
