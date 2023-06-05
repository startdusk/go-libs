package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagKeyColumn = "column"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOption) (*Model, error)
}

type ModelOption func(m *Model) error

// ModelOption 里面定义的函数没有对输入进行严格的校验, 这些检验应该交个用户
func ModelWithTableName(tableName string) ModelOption {
	return func(m *Model) error {
		m.tableName = tableName
		return nil
	}
}

// ModelOption 里面定义的函数没有对输入进行严格的校验, 这些检验应该交个用户
func ModelWithColumnName(field string, colName string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.fieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}

		// 同步更新 columnMap 的 colName(key)
		delete(m.columnMap, fd.colName)
		fd.colName = colName
		m.columnMap[fd.colName] = fd

		return nil
	}
}

type Model struct {
	tableName string
	// 字段名到字段定义的映射
	fieldMap map[string]*Field
	// 数据库列名到字段定义的映射
	columnMap map[string]*Field
}

type Field struct {
	// 字段名
	goName string

	// 列名
	colName string

	// 代表的是字段的类型
	typ reflect.Type
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

func newRegistry() *registry {
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
			colName: colName,
			// 字段类型
			typ: fd.Type,
			// 字段名(结构体的字段名)
			goName: fd.Name,
		}

		fieldMap[fd.Name] = field
		columnMap[colName] = field
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}

	m := &Model{
		tableName: tableName,
		fieldMap:  fieldMap,
		columnMap: columnMap,
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
