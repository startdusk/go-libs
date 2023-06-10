package orm

// Expression 是一个标记接口, 代表表达式
type Expression interface {
	expr()
}

// RawExpr 代表的是原生表达式
// 是一种兜底方式, 由于用户的输入SQL过于复杂, 就交给用户自己手写SQL, 我们就不能帮忙构建了
type RawExpr struct {
	raw  string
	args []any
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

func (r RawExpr) selectable() {}
func (r RawExpr) expr()       {}
