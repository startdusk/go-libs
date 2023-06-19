package valuer

import (
	"database/sql"

	"github.com/startdusk/go-libs/orm/model"
)

// Value 是对结构体实例的内部抽象
type Value interface {
	// Field 返回字段对应的值
	Field(name string) (any, error)
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

type Creator func(model *model.Model, entity any) Value
