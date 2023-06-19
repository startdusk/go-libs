package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
)

var (
	DialectMySQL       Dialect = &mysqlDialect{}
	DialectPostgresSQL Dialect = &postgreDialect{}
	DialectSQLite      Dialect = &sqliteDialect{}
)

type Dialect interface {
	// quoter 就是为了解决引号问题
	// MySQL 反引号 `
	// Oracle 是双引号
	quoter() byte

	buildUpsert(b *builder, odk *Upsert) error
}

type standardSQL struct{}

func (d standardSQL) quoter() byte {
	return ' '
}

func (d standardSQL) buildUpsert(b *builder, odk *Upsert) error {
	return nil
}

type mysqlDialect struct {
	standardSQL
}

func (d mysqlDialect) buildUpsert(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

func (d mysqlDialect) quoter() byte {
	return '`'
}

type sqliteDialect struct {
	standardSQL
}

func (d sqliteDialect) quoter() byte {
	return '`'
}

func (d sqliteDialect) buildUpsert(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON CONFLICT(")
	for i, col := range odk.conflictColumns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		err := b.buildColumn(col)
		if err != nil {
			return err
		}

	}
	b.sb.WriteString(") DO UPDATE SET ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type postgreDialect struct {
	standardSQL
}
