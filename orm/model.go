package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
	"reflect"
	"unicode"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	// 列名
	colName string
}

// 只支持输入指针类型的结构体
func parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)

	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fieldMap := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fieldMap[fd.Name] = &field{
			colName: underscoreName(fd.Name),
		}
	}

	return &model{
		tableName: underscoreName(typ.Name()),
		fields:    fieldMap,
	}, nil
}

// 驼峰名字符串转下划线命名
func underscoreName(name string) string {
	var buf []byte
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 && i < len(name)-1 && !unicode.IsUpper([]rune(name)[i+1]) {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
