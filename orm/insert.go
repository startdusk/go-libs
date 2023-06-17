package orm

import (
	"reflect"
	"strings"

	"github.com/startdusk/go-libs/orm/internal/errs"
)

type Inserter[T any] struct {
	values []*T
	db     *DB
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
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

	sb.WriteByte('(')
	// Golang中的map每一次遍历都是无序的
	for idx, field := range m.Fields {
		if idx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(field.ColName)
		sb.WriteByte('`')
	}
	sb.WriteByte(')')
	sb.WriteString(" VALUES ")
	argsCapcity := len(i.values) * len(m.Fields)
	args := make([]any, 0, argsCapcity)
	for valIdx := range i.values {
		if valIdx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('(')
		for idx, field := range m.Fields {
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
	sb.WriteByte(';')

	return &Query{
		SQL:  sb.String(),
		Args: args,
	}, nil
}
