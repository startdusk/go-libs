//go:build integration
package integration

import (
	"context"
	"github.com/startdusk/go-libs/orm"
	"github.com/startdusk/go-libs/orm/internal/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestMySQLInsert(t *testing.T) {
	suite.Run(t, &InsertSuite{
		Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

type InsertSuite struct {
	Suite
}

// 每次执行开始前, 执行这个函数
func (i *InsertSuite) TearDownTest() {
	orm.RawQuery[test.SimpleStruct](i.db, "TRUNCATE TABLE `simple_struct`").Exec(context.Background())
}

func (i *InsertSuite) TestInsert() {
	t := i.T()
	db := i.db
	cases := []struct {
		name         string
		i            *orm.Inserter[test.SimpleStruct]
		wantAffected int64 // 插入行数
	}{
		{
			name:         "insert one",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(55)),
			wantAffected: 1,
		},
		{
			name: "insert multiple",
			i: orm.NewInserter[test.SimpleStruct](db).Values(
				test.NewSimpleStruct(56),
				test.NewSimpleStruct(57),
			),
			wantAffected: 2,
		},
		{
			name:         "insert id",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(&test.SimpleStruct{ID: 58}),
			wantAffected: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			res := c.i.Exec(ctx)
			affected, err := res.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, c.wantAffected, affected)
		})
	}
}

// type SQLite3InsertSuite struct {
// 	InsertSuite
// }

// func (i *SQLite3InsertSuite) SetupSuite() {
// 	db, err := sql.Open(i.driver, i.dsn)
// 	// 建表语句
// 	db.ExecContext(context.Background(), "")
// 	require.NoError(i.T(), err)
// 	i.db, err = orm.OpenDB(db)
// 	require.NoError(i.T(), err)
// }

// func TestSQLite3(t *testing.T) {
// 	suite.Run(t, &SQLite3InsertSuite{
// 		InsertSuite: {
// 			driver: "sqlite3",
// 			dsn:    "file:test.db?cache=shared&mode=memory",
// 		},
// 	})
// }
