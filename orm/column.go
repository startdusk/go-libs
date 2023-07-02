package orm

func (c Column) selectable() {}

func (c Column) expr() {}

type Column struct {
	table TableReference // 代表的是哪一个table(用于join查询, 需要知道是哪个表的字段)
	name  string
	alias string
}

// As 给字段设置别名
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
		table: c.table,
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: valueOf(arg),
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: valueOf(arg),
	}
}

func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
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
