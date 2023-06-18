package orm

func (c Column) selectable() {}

func (c Column) expr() {}

type Column struct {
	name  string
	alias string
}

// As 给字段设置别名
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGt,
		right: valueOf(arg),
	}
}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLt,
		right: valueOf(arg),
	}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: val}
	}
}

var _ Assignable = new(Column)

func (c Column) assign() {}
