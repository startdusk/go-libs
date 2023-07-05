package sync

import (
	"sync"
)

type Map[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

func (m *Map[K, V]) Put(key K, val V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = val
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, ok := m.data[key]
	return val, ok
}

// 使用RWMutex实现double check
// 加读锁先检查一遍
// 释放读锁
// 加写锁
// 再检查一遍
func (m *Map[K, V]) LoadAndStore(key K, newVal V) (V, bool) {
	m.mutex.RLock()
	val, ok := m.data[key]
	if ok {
		return val, true
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	defer m.mutex.Unlock()
	// double check 避免线程覆盖问题
	val, ok = m.data[key]
	if ok {
		return val, true
	}

	m.data[key] = newVal
	return newVal, false
}
