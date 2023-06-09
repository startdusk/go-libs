package model

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
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "ID",
					Type:    reflect.TypeOf(int64(0)),
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Type:    reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Type:    reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
		},

		{
			name:   "test pointer model with opts",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "TEST_MODEL",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "ID",
					Type:    reflect.TypeOf(int64(0)),
				},
				{
					ColName: "firstname",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Type:    reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Type:    reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
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

	r := NewRegistry()
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
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			c.wantModel.FieldMap = fieldMap
			c.wantModel.ColumnMap = columnMap
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
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "ID",
					Type:    reflect.TypeOf(int64(0)),
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Type:    reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Type:    reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
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
				TableName: "tag_table",
			},

			fields: []*Field{
				{
					ColName: "first_name_t",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
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
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
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
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
				},
			},
		},

		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
				},
			},
		},

		{
			name:   "custom table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
				},
			},
		},

		{
			name:   "custom table name for ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
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

	r := NewRegistry()
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
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			c.wantModel.FieldMap = fieldMap
			c.wantModel.ColumnMap = columnMap

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

type TestModel struct {
	ID        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
