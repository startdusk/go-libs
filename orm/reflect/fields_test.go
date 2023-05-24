package reflect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IterateFields(t *testing.T) {
	// TODO: 把结构体挪到林外包, 在导入这里测试
	type User struct {
		Name string
		age  int
	}
	cases := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]any
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			},
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			},
		},
		{
			// 多级指针
			name: "multiple_pointer",
			entity: func() **User {
				res := &User{
					Name: "Tom",
					age:  18,
				}
				return &res
			}(),
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			},
		},
		{
			name:    "primitive_type",
			entity:  18,
			wantErr: errNotSupportType,
		},
		{
			name:    "nil",
			entity:  nil,
			wantErr: errNotSupportNil,
		},
		{
			name:    "user nil",
			entity:  (*User)(nil), // 相当于给interface赋值了User类型，但没赋值数据, 导致的现象就是 反射type有数据, 反射value是零值
			wantErr: errNotSupportZeroValue,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := IterateFields(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantRes, res)
		})
	}
}

func Test_SetField(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	cases := []struct {
		name       string
		entity     any
		field      string
		newVal     any
		wantErr    error
		wantEntity any
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
			},
			field:   "Name",
			newVal:  "Jerry",
			wantErr: errNotSetField,
		},
		{
			name: "pointer unexported field",
			entity: &User{
				age: 18,
			},
			field:   "age",
			newVal:  19,
			wantErr: errNotSetField,
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
			},
			field:  "Name",
			newVal: "Jerry",
			wantEntity: &User{
				Name: "Jerry",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := SetField(c.entity, c.field, c.newVal)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantEntity, c.entity)
		})
	}
}
