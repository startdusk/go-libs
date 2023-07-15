package cache

import (
	"time"
)

type RetryStrategy interface {
	// 第一个返回重试的间隔时间, 第二个返回要不要继续重试
	Next() (time.Duration, bool)
}

type FixedIntervalRetryStrategy struct {
	Interval time.Duration
	MaxCnt   int
	cnt      int
}

func (f *FixedIntervalRetryStrategy) Next() (time.Duration, bool) {
	if f.cnt >= f.MaxCnt {
		return 0, false
	}
	f.cnt++
	return f.Interval, true
}
