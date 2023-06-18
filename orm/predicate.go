package orm

type op string

const (
	opEq  op = "="
	opLt  op = "<"
	opGt  op = ">"
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
)

func (o op) String() string {
	return string(o)
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

var _ Selectable = new(Predicate)

func (p Predicate) selectable() {}

var _ Expression = new(Predicate)

func (p Predicate) expr() {}

func C(name string) Column {
	return Column{name: name}
}

func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// C("id").Eq(12).And(C("name").Eq("Tom")) => id = 12 AND name = "Tom"
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

// C("id").Eq(12).Or(C("name").Eq("Tom")) => id = 12 OR name = "Tom"
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

type value struct {
	val any
}

var _ Expression = new(value)

func (v value) expr() {}
