package sqldemo

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONColumn[T any] struct {
	Val T

	// 处理NULL的问题
	Valid bool
}

func (j JSONColumn[T]) Value() (driver.Value, error) {
	if !j.Valid {
		// NULL
		return nil, nil
	}

	bytes, err := json.Marshal(j.Val)
	return bytes, err
}

func (j *JSONColumn[T]) Scan(src any) error {
	var bs []byte
	switch data := src.(type) {
	case string:
		// 可以考虑处理空字符串
		bs = []byte(data)
	case []byte:
		// 可以考虑处理空slice
		bs = data
	case nil:
		// 说明数据库里面存的就是 NULL
		return nil
	default:
		return errors.New("不支持类型")
	}
	err := json.Unmarshal(bs, &j.Val)
	if err == nil {
		// 代表有数据 不为 NULL
		j.Valid = true
	}
	return err
}
