package model

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
)

const (
	tagKeyColumn = "column"
)

type TableName interface {
	TableName() string
}

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOption) (*Model, error)
}

type ModelOption func(m *Model) error

// ModelOption 里面定义的函数没有对输入进行严格的校验, 这些检验应该交个用户
func ModelWithTableName(tableName string) ModelOption {
	return func(m *Model) error {
		m.TableName = tableName
		return nil
	}
}

// ModelOption 里面定义的函数没有对输入进行严格的校验, 这些检验应该交个用户
func ModelWithColumnName(field string, colName string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}

		// 同步更新 columnMap 的 colName(key)
		delete(m.ColumnMap, fd.ColName)
		fd.ColName = colName
		m.ColumnMap[fd.ColName] = fd

		for i := range m.Fields {
			if m.Fields[i].ColName == colName {
				m.Fields[i] = fd
			}
		}

		return nil
	}
}

type Model struct {
	TableName string
	// 字段名(Golang结构体的字段名)到字段定义的映射
	FieldMap map[string]*Field
	// 数据库列名到字段定义的映射
	ColumnMap map[string]*Field

	// 记录 field 的顺序(map是无序的)
	Fields []*Field
}

type Field struct {
	// 字段名(Golang结构体的字段名)
	GoName string

	// 列名
	ColName string

	// 代表的是字段的类型
	Type reflect.Type

	// 字段相对于结构体本身的偏移量
	Offset uintptr
}

// registry 代表元数据的注册中心
type registry struct {
	// 为什么要用reflect.Type作为key
	// 因为有同名结构体但表名不一样的需求
	// 如: buyer下的User 和 seller下的User
	// 那么reflect.Type就能很好的记录和区分这两个同名结构体
	models map[reflect.Type]*Model

	// 保护map
	// 也可以使用sync.Map, 但sync.Map有线程覆盖的问题
	// 使用严格的读写锁, 采用double check的读写锁写法就没有线程覆盖的问题
	lock sync.RWMutex
}

func NewRegistry() *registry {
	return &registry{
		// 一个项目如果超过64张表, 说明需要拆分了
		models: make(map[reflect.Type]*Model, 64),
	}
}

// func (r *registry) Get(val any) (*Model, error) {
// 	typ := reflect.TypeOf(val)
// 	m, ok := r.models.Load(typ)
// 	if !ok {
// 		var err error
// 		if m, err = r.parseModel(typ); err != nil {
// 			return nil, err
// 		}
// 	}
// 	r.models.Store(typ, m) // 多线程同时执行到这里, 会出现线程覆盖的问题
// 	return m.(*Model), nil
// }

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	r.lock.RLock()
	m, ok := r.models[typ]
	r.lock.RUnlock()
	if ok {
		return m, nil
	}

	r.lock.Lock()
	// double check 写法, 保证不重复创建对象
	m, ok = r.models[typ]
	r.lock.Unlock()
	if ok {
		return m, nil
	}

	m, err := r.Register(val)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// 只支持输入指针类型的结构体
func (r *registry) Register(entity any, opts ...ModelOption) (*Model, error) {
	typ := reflect.TypeOf(entity)

	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemTyp := typ.Elem()
	numField := elemTyp.NumField()
	fields := make([]*Field, numField)
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := elemTyp.Field(i)
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fd.Name)
		}
		field := &Field{
			// 数据库字段的列名
			ColName: colName,
			// 字段类型
			Type: fd.Type,
			// 字段名(结构体的字段名)
			GoName: fd.Name,
			Offset: fd.Offset,
		}

		fieldMap[fd.Name] = field
		columnMap[colName] = field
		fields[i] = field
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}

	m := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
		Fields:    fields,
	}
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	r.lock.Lock()
	r.models[typ] = m
	r.lock.Unlock()

	return m, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return nil, nil
	}
	pairs := strings.Split(ormTag, ",")
	tags := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segs := strings.Split(pair, "=")
		if len(segs) != 2 {
			return nil, errs.NewErrIinvalidTagContent(pair)
		}
		tags[segs[0]] = segs[1]
	}
	return tags, nil
}

// 驼峰名字符串转下划线命名
func underscoreName(str string) string {
	var builder strings.Builder
	// Normally, most underscore named strings have 1 to 2 separators, so 2 is added here
	builder.Grow(len(str) + 2)

	var prev byte
	var prevUpper bool
	for index, sl := 0, len(str); index < sl; index++ {
		cur := str[index]
		curUpper, curLower, curNum := isUpper(cur), isLower(cur), isNumber(cur)
		if curUpper {
			cur = toLower(cur)
		}

		if next, ok := nextVal(index, str); ok {
			nextUpper, nextLower, nextNum := isUpper(next), isLower(next), isNumber(next)
			if (curUpper && (nextLower || nextNum)) || (curLower && (nextUpper || nextNum)) || (curNum && (nextUpper || nextLower)) {
				if prevUpper && curUpper && nextLower {
					builder.WriteByte('_')
				}
				builder.WriteByte(cur)
				if curLower || curNum || nextNum {
					builder.WriteByte('_')
				}

				prev, prevUpper = cur, curUpper
				continue
			}
		}
		if !curUpper && !curLower && !curNum {
			builder.WriteByte('_')
		} else {
			builder.WriteByte(cur)
		}
		prev, prevUpper = cur, curUpper
	}

	_ = prev
	_ = prevUpper

	res := builder.String()

	return res
}

func nextVal(index int, str string) (byte, bool) {
	var b byte
	next := index + 1
	if next < len(str) {
		b = str[next]
		return b, true
	}
	return b, false
}

// ascii A -> a
const transNum = 'a' - 'A'

// snakeDelimiter for snake
const snakeDelimiter = '_'

func toUpper(b byte) byte {
	return b - transNum
}

func toLower(b byte) byte {
	return b + transNum
}

func isNumber(b byte) bool {
	return '0' <= b && b <= '9'
}

func isUpper(b byte) bool {
	return 'A' <= b && b <= 'Z'
}

func isLower(b byte) bool {
	return 'a' <= b && b <= 'z'
}
