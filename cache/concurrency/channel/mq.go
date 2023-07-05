package channel

import (
	"errors"
	"sync"
)

// 实现内存版的消息队列
type Broker struct {
	mutex sync.RWMutex
	chans []chan Msg

	// topic 实现方案
	// topic map[string][]chan Msg
}

func (b *Broker) Send(m Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for i := 0; i < len(b.chans); i++ {
		select {
		case b.chans[i] <- m:
		default:
			return errors.New("消息队列已满")
		}

	}
	return nil
}

func (b *Broker) Subscribe(cap int) (<-chan Msg, error) {
	c := make(chan Msg, cap)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.chans = append(b.chans, c)
	return c, nil
}

func (b *Broker) Close() error {
	b.mutex.Lock()
	// 减小锁的粒度
	// 同时 chans 置为nil, 避免二次close channel
	chans := b.chans
	b.chans = nil
	b.mutex.Unlock()
	for i := 0; i < len(chans); i++ {
		close(chans[i])
	}
	return nil
}

type Msg struct {
	Content string
}

// 理论上，上面一种实现的方式拓展性更好, 因为数据在channel里面, 可能控制处理时间
// 下面的实现方式发送的时候就得处理了, 不能控制超时

type BrokerV1 struct {
	mutex     sync.RWMutex
	comsumers []func(msg Msg)
}

func (b *BrokerV1) Send(m Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, c := range b.comsumers {
		c(m)
	}
	return nil
}

func (b *BrokerV1) Subscribe(cb func(msg Msg)) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.comsumers = append(b.comsumers, cb)
	return nil
}
