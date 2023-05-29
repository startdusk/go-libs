package orm

import (
	"context"
	"github.com/startdusk/go-libs/orm/internal/errs"
	"strings"
)

type Selector[T any] struct {
	tableName string
	where     []Predicate
	sb        strings.Builder
	args      []any
	model     *model

	db *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) From(tableName string) *Selector[T] {
	s.tableName = tableName
	return s
}

func (s *Selector[T]) Where(where ...Predicate) *Selector[T] {
	s.where = where
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.db.r.get(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT * FROM ")
	if s.tableName == "" {

		// 这里给表名加 ``
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	} else {
		// 这里是用户传进来的表名, 用户应该保证它的正确性
		// 如 `tableName`
		// 如 `db`.`tableName`
		// 我们不处理反引号的问题
		s.sb.WriteString(s.tableName)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	panic("impl")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("impl")
}

func (s *Selector[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case Predicate: // 代表一个查询条件
		// 处理p
		// p.left 构建好
		// p.op 构建好
		// p.right 构建好

		// 注意: 生成的SQL中, 处理加空格, 加标点符号的问题会让代码很难看, 但这是必须的
		// 空格不一定处理的完美, 宁多勿少, 反正数据库能解析
		_, lok := exp.left.(Predicate)
		if lok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if lok {
			s.sb.WriteByte(')')
		}

		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')

		_, rok := exp.right.(Predicate)
		if rok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if rok {
			s.sb.WriteByte(')')
		}
	case Column: // 代表列名, 直接拼接列名
		fd, ok := s.model.fields[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')
	case value: // 代表参数, 加入参数列表
		s.sb.WriteString("?")
		s.addArg(exp.val)
	case nil:
		return nil
	default:
		return errs.NewErrUnsupportedExpressionType(expr)
	}
	return nil
}

func (s *Selector[T]) addArg(val any) {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
}
