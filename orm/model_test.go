package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseModel(t *testing.T) {
	cases := []struct {
		name      string
		entity    any
		wantModel *model
		wantErr   error
	}{
		{
			name:   "test pointer model",
			entity: &TestModel{},
			wantModel: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"ID": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
					"Age": {
						colName: "age",
					},
				},
			},
		},
		{
			name:    "test struct model",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "primitive type",
			entity:  0,
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "map",
			entity:  map[string]string{"1": "1"},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{1, 2, 3},
			wantErr: errs.ErrPointerOnly,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := parseModel(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.wantModel, m)
		})
	}
}
