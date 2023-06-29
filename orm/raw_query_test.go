package orm

import (
	"database/sql"
	"testing"

	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"errors"
	_ "github.com/mattn/go-sqlite3"
)

func Test_RawQuerier_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	db, err := OpenDB(mockDB, DBWithDialect(DialectMySQL))
	assert.NoError(t, err)

	// 对应 query error
	queryError := errors.New("query error")
	mock.ExpectQuery("SELECT .*").WillReturnError(queryError)

	// 对应 no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// // 对应 scan error
	// rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	// // 本来ID应该应该是数字类型, 但故意给了个abc, mock scan error
	// rows.AddRow("abc", "Tom", "18", "Jerry")
	// mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// 对应 query row success
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	// 数据库查询出来的数据返回的都是文本类型, 所以这里可以用字符串
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	cases := []struct {
		name    string
		r       *RawQuerier[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "query error",
			r:       RawQuery[TestModel](db, "SELECT * FROM `test_model`"),
			wantErr: queryError,
		},
		{
			name:    "no rows",
			r:       RawQuery[TestModel](db, "SELECT * FROM `test_model` WHERE `id` = ?", -1),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			r:    RawQuery[TestModel](db, "SELECT * FROM `test_model` WHERE `id` = ?", 1),
			wantRes: &TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := c.r.Get(context.Background())
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantRes, res)
		})
	}
}
