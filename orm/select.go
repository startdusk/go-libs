package orm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	tableName string
}

func (s *Selector[T]) From(tableName string) *Selector[T] {
	s.tableName = tableName
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	if s.tableName == "" {
		// 通过反射拿到类型的名字作为表名
		var t T
		typ := reflect.TypeOf(t)
		// 这里给表名加 ``
		sb.WriteByte('`')
		sb.WriteString(typ.Name())
		sb.WriteByte('`')
	} else {
		// 这里是用户传进来的表名, 用户应该保证它的正确性
		// 如 `tableName`
		// 如 `db`.`tableName`
		// 我们不处理反引号的问题
		sb.WriteString(s.tableName)
	}

	sb.WriteByte(';')
	return &Query{
		SQL: sb.String(),
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	panic("impl")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("impl")
}
