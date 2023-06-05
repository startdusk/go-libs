package orm

import (
	"database/sql"
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_Register(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error
		fields    []*Field
		opts      []ModelOption
	}{
		{
			name:   "test pointer model",
			entity: &TestModel{},
			wantModel: &Model{
				tableName: "test_model",
			},
			fields: []*Field{
				{
					colName: "id",
					goName:  "ID",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
				{
					colName: "last_name",
					goName:  "LastName",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
				{
					colName: "age",
					goName:  "Age",
					typ:     reflect.TypeOf(int8(0)),
				},
			},
		},

		{
			name:   "test pointer model with opts",
			entity: &TestModel{},
			wantModel: &Model{
				tableName: "TEST_MODEL",
			},
			fields: []*Field{
				{
					colName: "id",
					goName:  "ID",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					colName: "firstname",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
				{
					colName: "last_name",
					goName:  "LastName",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
				{
					colName: "age",
					goName:  "Age",
					typ:     reflect.TypeOf(int8(0)),
				},
			},
			opts: []ModelOption{
				ModelWithTableName("TEST_MODEL"),
				ModelWithColumnName("FirstName", "firstname"),
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
			m, err := r.Register(c.entity, c.opts...)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range c.fields {
				fieldMap[field.goName] = field
				columnMap[field.colName] = field
			}
			c.wantModel.fieldMap = fieldMap
			c.wantModel.columnMap = columnMap
			assert.Equal(t, c.wantModel, m)
		})
	}
}

func Test_RegistryGet(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		entity    any
		wantModel *Model
		wantErr   error

		fields []*Field
	}{
		{
			name:   "test pointer model",
			entity: &TestModel{},
			wantModel: &Model{
				tableName: "test_model",
			},
			fields: []*Field{
				{
					colName: "id",
					goName:  "ID",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
				{
					colName: "last_name",
					goName:  "LastName",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
				{
					colName: "age",
					goName:  "Age",
					typ:     reflect.TypeOf(int8(0)),
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
			wantModel: &Model{
				tableName: "tag_table",
			},

			fields: []*Field{
				{
					colName: "first_name_t",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
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
			wantModel: &Model{
				tableName: "tag_table",
			},
			fields: []*Field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
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
			wantModel: &Model{
				tableName: "tag_table",
			},
			fields: []*Field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
		},

		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				tableName: "empty_table_name",
			},
			fields: []*Field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
		},

		{
			name:   "custom table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				tableName: "custom_table_name_t",
			},
			fields: []*Field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
		},

		{
			name:   "custom table name for ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				tableName: "custom_table_name_ptr_t",
			},
			fields: []*Field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
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
			m, err := r.Get(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range c.fields {
				fieldMap[field.goName] = field
				columnMap[field.colName] = field
			}
			c.wantModel.fieldMap = fieldMap
			c.wantModel.columnMap = columnMap

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
