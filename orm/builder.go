package orm

import (
	"strings"

	"github.com/startdusk/go-libs/orm/internal/errs"
)

type builder struct {
	core
	sb     strings.Builder
	args   []any
	quoter byte
}

// buildColumn 构造列
func (b *builder) buildColumn(col Column) error {
	// meta, ok := b.model.FieldMap[fd]
	// if !ok {
	// 	return errs.NewErrUnknownField(fd)
	// }
	// b.quote(meta.ColName)
	// return nil

	switch table := col.table.(type) {
	case nil:
		fd, ok := b.model.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.name)
		}

		b.quote(fd.ColName)
		// 字段使用别名
		if col.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}

		fd, ok := m.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.name)
		}

		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}

		b.quote(fd.ColName)
		// 字段使用别名
		if col.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) addArgs(args ...any) {
	if len(args) == 0 {
		return
	}
	if b.args == nil {
		// 很少有查询能够超过8个参数
		// INSERT除外
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}
