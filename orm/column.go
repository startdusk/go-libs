package orm

func (c Column) selectable() {}

func (c Column) expr() {}

type Column struct {
	name string
}

func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGt,
		right: value{val: arg},
	}
}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLt,
		right: value{val: arg},
	}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: value{val: arg},
	}
}
