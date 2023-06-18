package orm

type Assignable interface {
	assign()
}

type Assignment struct {
	col string
	val any
}

func (a Assignment) assign() {}

func Assign(col string, val any) Assignment {
	return Assignment{
		col: col,
		val: val,
	}
}
