package integration

import (
	"context"
	"github.com/startdusk/go-libs/orm"
	"github.com/startdusk/go-libs/orm/internal/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"
)

func TestInsert(t *testing.T) {
	db, err := orm.Open("mysql", "root:root@tcp(localhost:13306)/integration_test")
	require.NoError(t, err)

	cases := []struct {
		name         string
		i            *orm.Inserter[test.SimpleStruct]
		wantAffected int64 // 插入行数
	}{
		{
			name:         "insert one",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(15)),
			wantAffected: 1,
		},
		{
			name: "insert multiple",
			i: orm.NewInserter[test.SimpleStruct](db).Values(
				test.NewSimpleStruct(16),
				test.NewSimpleStruct(17),
			),
			wantAffected: 2,
		},
		{
			name:         "insert id",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(&test.SimpleStruct{Id: 18}),
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
