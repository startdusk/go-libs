package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_Register(t *testing.T) {
	t.Parallel()

	var testModel = &TestModel{}
	cases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error

		opts []ModelOption
	}{
		{
			name:   "test pointer model",
			entity: testModel,
			wantModel: &Model{
				tableName: "test_model",
				fields: map[string]*Field{
					"ID": {
						colName: "id",
						goName:  "ID",
						typ:     reflect.TypeOf(testModel.ID),
					},
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ:     reflect.TypeOf(testModel.FirstName),
					},
					"LastName": {
						colName: "last_name",
						goName:  "LastName",
						typ:     reflect.TypeOf(testModel.LastName),
					},
					"Age": {
						colName: "age",
						goName:  "Age",
						typ:     reflect.TypeOf(testModel.Age),
					},
				},
			},
		},

		{
			name:   "test pointer model with opts",
			entity: testModel,
			wantModel: &Model{
				tableName: "TEST_MODEL",
				fields: map[string]*Field{
					"ID": {
						colName: "id",
						goName:  "ID",
						typ:     reflect.TypeOf(testModel.ID),
					},
					"FirstName": {
						colName: "firstname",
						goName:  "FirstName",
						typ:     reflect.TypeOf(testModel.FirstName),
					},
					"LastName": {
						colName: "last_name",
						goName:  "LastName",
						typ:     reflect.TypeOf(testModel.LastName),
					},
					"Age": {
						colName: "age",
						goName:  "Age",
						typ:     reflect.TypeOf(testModel.Age),
					},
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

			assert.Equal(t, c.wantModel, m)
		})
	}
}

func Test_RegistryGet(t *testing.T) {
	t.Parallel()

	var testModel = &TestModel{}
	cases := []struct {
		name string

		entity    any
		wantModel *Model
		wantErr   error
	}{
		{
			name:   "test pointer model",
			entity: testModel,
			wantModel: &Model{
				tableName: "test_model",
				fields: map[string]*Field{
					"ID": {
						colName: "id",
						goName:  "ID",
						typ:     reflect.TypeOf(testModel.ID),
					},
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ:     reflect.TypeOf(testModel.FirstName),
					},
					"LastName": {
						colName: "last_name",
						goName:  "LastName",
						typ:     reflect.TypeOf(testModel.LastName),
					},
					"Age": {
						colName: "age",
						goName:  "Age",
						typ:     reflect.TypeOf(testModel.Age),
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
			wantModel: &Model{
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name_t",
						goName:  "FirstName",
						typ: func() reflect.Type {
							// 因为结构体只有一个字段, 没有对齐的问题, 所以可以直接使用原生类型的元数据
							var a string
							return reflect.TypeOf(a)
						}(),
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
			wantModel: &Model{
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ: func() reflect.Type {
							var a string
							return reflect.TypeOf(a)
						}(),
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
			wantModel: &Model{
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ: func() reflect.Type {
							var a string
							return reflect.TypeOf(a)
						}(),
					},
				},
			},
		},

		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				tableName: "empty_table_name",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ: func() reflect.Type {
							var a string
							return reflect.TypeOf(a)
						}(),
					},
				},
			},
		},

		{
			name:   "custom table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				tableName: "custom_table_name_t",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ: func() reflect.Type {
							var a string
							return reflect.TypeOf(a)
						}(),
					},
				},
			},
		},

		{
			name:   "custom table name for ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				tableName: "custom_table_name_ptr_t",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
						goName:  "FirstName",
						typ: func() reflect.Type {
							var a string
							return reflect.TypeOf(a)
						}(),
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
			m, err := r.Get(c.entity)
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
