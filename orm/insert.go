package orm

import (
	"reflect"
	"strings"

	"github.com/startdusk/go-libs/orm/internal/errs"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns: assigns,
	}

	return o.i
}

type Inserter[T any] struct {
	// INSERT 语句要插入的值的结构体的列表
	values []*T

	// INSERT 语句要插入的指定的列
	columns []string
	db      *DB

	onDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
}

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

// Columns 指定插入的列
func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

// Values 指定插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRows
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	// 拿到元数据, 拼接表名
	sb.WriteByte('`')
	sb.WriteString(m.TableName)
	sb.WriteByte('`')
	// 一定要显式指定列的顺序, 不然我们不知道数据库中默认的顺序
	// 我们要构造 `table_name`(col1, col2)

	// 用户指定了列
	fields := m.Fields
	if len(i.columns) > 0 {
		fields = fields[:0]
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	sb.WriteByte('(')
	// Golang中的map每一次遍历都是无序的
	for idx, field := range fields {
		if idx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(field.ColName)
		sb.WriteByte('`')
	}
	sb.WriteByte(')')
	sb.WriteString(" VALUES ")
	argsCapcity := len(i.values) * len(fields)
	args := make([]any, 0, argsCapcity)
	for valIdx := range i.values {
		if valIdx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('(')
		for idx, field := range fields {
			if idx > 0 {
				sb.WriteByte(',')
			}
			sb.WriteByte('?')
			// 读取结构体的参数
			val := reflect.ValueOf(i.values[valIdx]).Elem().FieldByName(field.GoName).Interface()
			args = append(args, val)
		}
		sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for idx, assign := range i.onDuplicateKey.assigns {
			if idx > 0 {
				sb.WriteByte(',')
			}
			switch a := assign.(type) {
			case Assignment:
				fd, ok := m.FieldMap[a.col]
				if !ok {
					return nil, errs.NewErrUnknownField(a.col)
				}
				sb.WriteByte('`')
				sb.WriteString(fd.ColName)
				sb.WriteByte('`')
				sb.WriteString("=?")
				args = append(args, a.val)
			case Column:
				fd, ok := m.FieldMap[a.name]
				if !ok {
					return nil, errs.NewErrUnknownField(a.name)
				}
				sb.WriteByte('`')
				sb.WriteString(fd.ColName)
				sb.WriteByte('`')
				sb.WriteString("=VALUES(")
				sb.WriteByte('`')
				sb.WriteString(fd.ColName)
				sb.WriteByte('`')
				sb.WriteByte(')')
			default:
				return nil, errs.NewErrUnsupportedAssignable(assign)
			}
		}
	}

	sb.WriteByte(';')

	return &Query{
		SQL:  sb.String(),
		Args: args,
	}, nil
}
