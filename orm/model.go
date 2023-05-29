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

// registry 代表元数据的注册中心
type registry struct {
	// 为什么要用reflect.Type作为key
	// 因为有同名结构体但表名不一样的需求
	// 如: buyer下的User 和 seller下的User
	// 那么reflect.Type就能很好的记录和区分这两个同名结构体
	models map[reflect.Type]*model
}

func newRegistry() *registry {
	return &registry{
		// 一个项目如果超过64张表, 说明需要拆分了
		models: make(map[reflect.Type]*model, 64),
	}
}

func (r *registry) get(val any) (*model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models[typ]
	if !ok {
		var err error
		m, err = r.parseModel(val)
		if err != nil {
			return nil, err
		}
		r.models[typ] = m
	}

	return m, nil
}

// 只支持输入指针类型的结构体
func (r *registry) parseModel(entity any) (*model, error) {
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
