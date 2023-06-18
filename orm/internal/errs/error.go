package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly    = errors.New("orm: 只支持指向结构体的一级指针")
	ErrNoRows         = errors.New("orm: 没有数据")
	ErrInsertZeroRows = errors.New("orm: 插入0行数据")
)

func NewErrUnsupportedExpressionType(expr any) error {
	return fmt.Errorf("orm: 不支持的表达式 %v", expr)
}

func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm: 未知字段 %s", name)
}

func NewErrUnknownColumn(name string) error {
	return fmt.Errorf("orm: 未知数据库列名 %s", name)
}

func NewErrIinvalidTagContent(pair string) error {
	return fmt.Errorf("orm: 非法标签值 %s", pair)
}

func NewErrUnsupportedAssignable(expr any) error {
	return fmt.Errorf("orm: 不支持的赋值表达式类型 %v", expr)
}
