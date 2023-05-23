package reflect

import (
	"errors"
	"reflect"
)

var errNotSupportType = errors.New("不支持这种类型")
var errNotSupportNil = errors.New("不支持nil")
var errNotSupportZeroValue = errors.New("不支持零值")

// IterateFields 遍历对象字段
func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, errNotSupportNil
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if val.IsZero() {
		return nil, errNotSupportZeroValue
	}

	// for typ.Kind() == reflect.Pointer 反射层面上的解引用, 如 &user 直接取到 user, &&user 取到 user
	for typ.Kind() == reflect.Pointer {
		// 拿到指针指向的对象
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errNotSupportType
	}

	numField := typ.NumField()
	res := make(map[string]any, numField)
	for i := 0; i < numField; i++ {
		// 字段类型
		fieldType := typ.Field(i)
		// 字段的值
		fieldValue := val.Field(i)

		if fieldType.IsExported() {
			res[fieldType.Name] = fieldValue.Interface()
		} else {
			// 非导出字段我们是拿不到的所以赋值默认零值(是为了在断言的时候保证字段的统一性)
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}
	}

	return res, nil
}
