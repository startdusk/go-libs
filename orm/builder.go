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
func (b *builder) buildColumn(fd string) error {
	meta, ok := b.model.FieldMap[fd]
	if !ok {
		return errs.NewErrUnknownField(fd)
	}
	b.quote(meta.ColName)
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
