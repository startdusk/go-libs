package orm

import (
	"context"
	"database/sql"

	"github.com/startdusk/go-libs/orm/internal/errs"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.upsert = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}

	return o.i
}

type Inserter[T any] struct {
	builder

	// INSERT 语句要插入的值的结构体的列表
	values []*T

	// INSERT 语句要插入的指定的列
	columns []string

	upsert *Upsert

	sess Session
}

func NewInserter[T any](sess Session) *Inserter[T] {
	core := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			core:   core,
			quoter: core.dialect.quoter(),
		},
		sess: sess,
	}
}

func (i *Inserter[T]) Upsert() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
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

	i.sb.WriteString("INSERT INTO ")
	m := i.model
	if m == nil {
		var err error
		m, err = i.r.Get(i.values[0])
		if err != nil {
			return nil, err
		}
		i.model = m
	}
	// 拿到元数据, 拼接表名
	i.quote(m.TableName)

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

	i.sb.WriteByte('(')
	// Golang中的map每一次遍历都是无序的
	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.ColName)
	}
	i.sb.WriteByte(')')
	i.sb.WriteString(" VALUES ")
	argsCapcity := len(i.values) * len(fields)
	i.args = make([]any, 0, argsCapcity)
	for valIdx := range i.values {
		if valIdx > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		val := i.creator(i.model, i.values[valIdx])
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 读取结构体的参数
			arg, err := val.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(arg)
		}
		i.sb.WriteByte(')')
	}

	if i.upsert != nil {
		if err := i.dialect.buildUpsert(&i.builder, i.upsert); err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')

	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	var err error
	var result Result
	i.model, err = i.r.Get(new(T))
	if err != nil {
		result.err = err
		return result
	}

	res := exec(ctx, i.sess, i.core, &QueryContext{
		Type:    "INSERT",
		Builder: i,
		Model:   i.model,
	})
	var sqlRes sql.Result
	if val, ok := res.Result.(sql.Result); ok {
		sqlRes = val
	}
	result.res = sqlRes
	result.err = res.Err
	return result
}

// var _ Handler = (&Inserter[any]{}).execHandler

// func (i *Inserter[T]) execHandler(ctx context.Context, qc *QueryContext) *QueryResult {
// 	qr := &QueryResult{}
// 	q, err := i.Build()
// 	if err != nil {
// 		qr.Err = err
// 		qr.Result = Result{err: err}
// 		return qr
// 	}
// 	res, err := i.sess.execContext(ctx, q.SQL, q.Args...)
// 	qr.Err = err
// 	qr.Result = Result{res: res, err: err}
// 	return qr
// }
