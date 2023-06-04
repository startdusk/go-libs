package orm

import (
	"github.com/startdusk/go-libs/orm/internal/errs"
)

// 通过桥接的方式将内部错误导出外部
// 当然这种方式也有取舍, 就是重构的时候, 如果调用这个变量的文件被移动到另外一个包了, 那么这里就得跟着移动
var ErrNoRows = errs.ErrNoRows
