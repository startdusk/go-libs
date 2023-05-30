package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
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

	r := newRegistry()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := r.parseModel(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.wantModel, m)
		})
	}
}

func Test_RegistryGet(t *testing.T) {
	cases := []struct {
		name string

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
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column=first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &model{
				tableName: "tag_table",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name_t",
					},
				},
			},
		},

		{
			name: "empty column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column="`
				}
				return &TagTable{}
			}(),
			wantModel: &model{
				tableName: "tag_table",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name: "ignore column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"abc=abc"`
				}
				return &TagTable{}
			}(),
			wantModel: &model{
				tableName: "tag_table",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &model{
				tableName: "empty_table_name",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name:   "custom table name",
			entity: &CustomTableName{},
			wantModel: &model{
				tableName: "custom_table_name_t",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name:   "custom table name for ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &model{
				tableName: "custom_table_name_ptr_t",
				fields: map[string]*field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},

		{
			name: "invalid column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrIinvalidTagContent("column"),
		},
	}

	r := newRegistry()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := r.get(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.wantModel, m)
			typ := reflect.TypeOf(c.entity)
			m, ok := r.models[typ]
			assert.True(t, ok)
			assert.Equal(t, c.wantModel, m)
		})
	}
}

type EmptyTableName struct {
	FirstName string
}

func (e EmptyTableName) TableName() string {
	return ""
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}
