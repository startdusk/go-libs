package reflect

import (
	"reflect"
)

func IterateArray(entity any) ([]any, error) {
	val := reflect.ValueOf(entity)
	res := make([]any, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		ele := val.Index(i)
		res = append(res, ele.Interface())
	}
	return res, nil
}

// 返回值是 keys, values, error
func IterateMap(entity any) ([]any, []any, error) {
	// TODO: 实际情况要对输入的值进行类型检测
	val := reflect.ValueOf(entity)
	resKeys := make([]any, 0, val.Len())
	resValues := make([]any, 0, val.Len())

	it := val.MapRange()
	for it.Next() {
		resKeys = append(resKeys, it.Key().Interface())
		resValues = append(resValues, it.Value().Interface())
	}
	return resKeys, resValues, nil
}
