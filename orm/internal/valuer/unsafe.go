package valuer

import (
	"database/sql"
	"reflect"
	"unsafe"

	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/startdusk/go-libs/orm/model"
)

type unsafeValue struct {
	model *model.Model

	// val 对应 泛型 T 的指针
	val any

	// 结构体的起始地址
	address unsafe.Pointer
}

// 确保类型变更 我们能得到通知
var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	return &unsafeValue{
		model:   model,
		val:     val,
		address: reflect.ValueOf(val).UnsafePointer(),
	}
}

func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	// 获取 查询的 columns
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 利用 columns 来解决 select 的列顺序 和 列字段类型的问题
	vals := make([]any, 0, len(columns))

	for _, colName := range columns {
		// colName 是列名
		fd, ok := u.model.ColumnMap[colName]
		if !ok {
			return errs.NewErrUnknownColumn(colName)
		}
		// 字段地址 = 起始地址 + 偏移量
		fdAddress := unsafe.Pointer(uintptr(u.address) + fd.Offset)
		// 反射在特定的地址上, 创建一个特定类型的实例
		// 这里创建的实例是原本类型的指针类型
		// 例如: fd.Type = int类型, 那么 val 就是 *int类型
		val := reflect.NewAt(fd.Type, fdAddress)

		vals = append(vals, val.Interface())
	}

	// 到这里, 因为已经通过 unsafe 创建了对象, 所以 scan 就已经是对对象的字段赋值
	if err := rows.Scan(vals...); err != nil {
		return err
	}

	return rows.Err()
}
