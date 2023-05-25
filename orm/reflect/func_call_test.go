package reflect

import (
	"github.com/startdusk/go-libs/orm/reflect/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_IterateFunc(t *testing.T) {
	cases := []struct {
		name    string
		entity  any
		wantRes map[string]FuncInfo
		wantErr error
	}{
		{
			name:   "struct",
			entity: types.NewUser("Tom", 18),
			wantRes: map[string]FuncInfo{
				// GetAge 是值接收者, 能使用值/指针调用
				"GetAge": {
					Name: "GetAge",
					// 下标 0 的指向接收器
					InputTypes:  []reflect.Type{reflect.TypeOf(types.User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},
			},
		},
		{
			name:   "pointer",
			entity: types.NewUserPtr("Tom", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					Name: "GetAge",
					// 下标 0 的指向接收器
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},

				// ChangeName 是指针接收者, 只能使用指针调用
				"ChangeName": {
					Name: "ChangeName",
					// 下标 0 的指向接收器
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{}), reflect.TypeOf("")},
					OutputTypes: []reflect.Type{},
					Result:      []any{},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := IterateFunc(c.entity)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantRes, res)
		})
	}
}
