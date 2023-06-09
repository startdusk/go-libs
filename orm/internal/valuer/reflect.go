package valuer

import (
	"database/sql"
	"reflect"

	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/startdusk/go-libs/orm/model"
)

type reflectValue struct {
	model *model.Model

	// val 对应 泛型 T 的指针
	val any
}

// 确保类型变更 我们能得到通知
var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return &reflectValue{
		model: model,
		val:   val,
	}
}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	// 获取 查询的 columns
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 利用 columns 来解决 select 的列顺序 和 列字段类型的问题
	vals := make([]any, 0, len(columns))
	valElems := make([]reflect.Value, 0, len(columns))
	for _, colName := range columns {
		// colName 是列名
		fd, ok := r.model.ColumnMap[colName]
		if !ok {
			return errs.NewErrUnknownColumn(colName)
		}
		// 反射创建一个实例
		// 这里创建的实例是原本类型的指针类型
		// 例如: fd.Type = int类型, 那么 val 就是 *int类型, 所以需要 取Elem() 获取它的实例, 而不是指针
		val := reflect.New(fd.Type)
		vals = append(vals, val.Interface())
		// val.Elem() 就是 val 指向的数据
		valElems = append(valElems, val.Elem())
	}

	if err := rows.Scan(vals...); err != nil {
		return err
	}

	// 把 scan 后的数据放到构造的entity中
	valueElem := reflect.ValueOf(r.val).Elem()
	for i, colName := range columns {
		// colName 是列名
		fd, ok := r.model.ColumnMap[colName]
		if !ok {
			return errs.NewErrUnknownColumn(colName)
		}
		valueElem.FieldByName(fd.GoName).Set(valElems[i])
	}
	return rows.Err()
}
