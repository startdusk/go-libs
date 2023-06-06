package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type UnsafeAccesser struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
}

func NewUnsafeAccessor(entity any) *UnsafeAccesser {
	typ := reflect.TypeOf(entity)
	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]FieldMeta, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			typ:    fd.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &UnsafeAccesser{
		fields: fields,
		// UnsafeAddr 是不固定的地址, 每次垃圾回收后会被移动地址
		// UnsafePointer 是 golang 层面上的指针地址, 它会维护指向
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccesser) Field(field string) (any, error) {
	// 起始地址 + 字段偏移量
	fd, ok := a.fields[field]
	if !ok {
		return nil, errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// 如果知道类型, 就这么读取类型字段的值
	// return *(*int)(fdAddress), nil

	// 如果不知道类型, 就这么读取类型字段的值
	return reflect.NewAt(fd.typ, fdAddress).Elem().Interface(), nil
}

func (a *UnsafeAccesser) SetField(field string, val any) error {
	// 起始地址 + 字段偏移量
	fd, ok := a.fields[field]
	if !ok {
		return errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// 如果知道类型, 就这么设置类型字段的值
	// *(*int)(fdAddress) = val.(int)

	// 如果不知道类型, 就这么设置类型字段的值
	reflect.NewAt(fd.typ, fdAddress).Elem().Set(reflect.ValueOf(val))
	return nil
}

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}
