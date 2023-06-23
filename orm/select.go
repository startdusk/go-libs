package orm

import (
	"context"
	"github.com/startdusk/go-libs/orm/internal/errs"
)

// Selectable select 指定列
// 避免用户使用数据库列 存在耦合问题(用户应该使用Go结构体的字段名, 就能与数据库表字段解耦)
// 使用Go的结构体字段名同时也可以避免传入的SQL列名存在SQL注入问题
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	tableName string
	where     []Predicate

	// 指定 select 的列
	columns []Selectable

	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	core := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   core,
			quoter: core.dialect.quoter(),
		},
		sess: sess,
	}
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
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
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT ")
	if err := s.buildSelectColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")

	if s.tableName == "" {

		// 这里给表名加 ``
		s.quote(s.model.TableName)
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
	root := s.getHandler
	for i := len(s.mdls) - 1; i >= 0; i-- {
		root = s.mdls[i](root)
	}

	res := root(ctx, &QueryContext{
		Type:    "SELECT",
		Builder: s,
	})
	var t *T
	if val, ok := res.Result.(*T); ok {
		t = val
	}
	return t, res.Err
}

var _ Handler = (&Selector[any]{}).getHandler

func (s *Selector[T]) getHandler(ctx context.Context, qc *QueryContext) *QueryResult {
	qr := &QueryResult{}
	q, err := s.Build()
	if err != nil {
		qr.Err = err
		return qr
	}

	rows, err := s.sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		qr.Err = err
		return qr
	}

	if !rows.Next() {
		// 返回要和sql包语义一致
		qr.Err = ErrNoRows
		return qr
	}

	// 利用 columns 来解决 select 的列顺序 和 列字段类型的问题
	entity := new(T)
	// 接口定义好之后, 就两件事情, 一个是利用新接口的方法改造上层
	// 一个是提供不同的实现
	val := s.creator(s.model, entity)
	err = val.SetColumns(rows)
	qr.Result = entity
	qr.Err = err
	return qr
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		// 返回要和sql包语义一致
		return nil, ErrNoRows
	}

	var res []*T
	for rows.Next() {
		entity := new(T)
		val := s.creator(s.model, entity)
		if err := val.SetColumns(rows); err != nil {
			return nil, err
		}
		res = append(res, entity)
	}

	return res, nil
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

		if exp.op != "" {
			s.sb.WriteByte(' ')
			s.sb.WriteString(exp.op.String())
			s.sb.WriteByte(' ')
		}

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
		exp.alias = "" // where 部分不允许使用 AS(但这行代码写得很隐晦, 另一种写法就是标志位的写法, 也不是很好)
		return s.buildColumn(exp)
	case RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(exp.raw)
		s.addArgs(exp.args...)
		s.sb.WriteByte(')')
	case value: // 代表参数, 加入参数列表
		s.sb.WriteString("?")
		s.addArgs(exp.val)
	case nil:
		return nil
	default:
		return errs.NewErrUnsupportedExpressionType(expr)
	}
	return nil
}

// buildSelectColumns 构建 SELECT 的列
func (s *Selector[T]) buildSelectColumns() error {
	if len(s.columns) == 0 {
		// 没有指定列
		s.sb.WriteByte('*')
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteString(", ")
		}
		switch c := col.(type) {
		case Column:
			if err := s.buildColumn(c); err != nil {
				return err
			}
		case Aggregate:
			// 聚合函数名
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			// 聚合字段名
			if err := s.buildColumn(Column{name: c.arg}); err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数使用别名
			if c.alias != "" {
				s.sb.WriteString(" AS ")
				s.quote(c.alias)
			}
		case RawExpr:
			// 用户输入SQL
			s.sb.WriteString(c.raw)
			s.addArgs(c.args...)
		default:
			return errs.NewErrUnsupportedExpressionType(col)
		}
	}
	return nil
}

func (s *Selector[T]) buildColumn(col Column) error {
	fd, ok := s.model.FieldMap[col.name]
	if !ok {
		return errs.NewErrUnknownField(col.name)
	}

	s.quote(fd.ColName)
	// 字段使用别名
	if col.alias != "" {
		s.sb.WriteString(" AS ")
		s.quote(col.alias)
	}
	return nil
}
