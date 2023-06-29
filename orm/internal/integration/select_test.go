//go:build integration
package integration

import (
	"context"
	"github.com/startdusk/go-libs/orm"
	"github.com/startdusk/go-libs/orm/internal/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestMySQLSelect(t *testing.T) {
	suite.Run(t, &SelectSuite{
		Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

type SelectSuite struct {
	Suite
}

func (s *SelectSuite) SetupSuite() {
	s.Suite.SetupSuite()
	res := orm.NewInserter[test.SimpleStruct](s.db).Values(test.NewSimpleStruct(103)).Exec(context.Background())
	require.NoError(s.T(), res.Err())
}

func (s *SelectSuite) TestSuite() {
	db := s.db
	cases := []struct {
		name    string
		s       *orm.Selector[test.SimpleStruct]
		wantRes *test.SimpleStruct
		wantErr error
	}{
		{
			name:    "get data",
			s:       orm.NewSelector[test.SimpleStruct](db).Where(orm.C("ID").Eq(103)),
			wantRes: test.NewSimpleStruct(103),
		},
		{
			name:    "no rows",
			s:       orm.NewSelector[test.SimpleStruct](db).Where(orm.C("ID").Eq(1002)),
			wantErr: orm.ErrNoRows,
		},
	}

	for _, c := range cases {
		s.T().Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			res, err := c.s.Get(ctx)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantRes, res)
		})
	}
}
