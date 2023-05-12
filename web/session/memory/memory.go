package memory

import (
	"context"
	"errors"
	"fmt"
	cache "github.com/patrickmn/go-cache"
	"github.com/startdusk/go-libs/web/session"
	"sync"
	"time"
)

var (
	// errKeyNotFound 推荐使用小写, 当有人问你为什么不提供大写的时候, 你再改成大写(避免 增大错误的表面积)
	errKeyNotFound = errors.New("session: key not found")
)

type Store struct {
	mu sync.RWMutex
	sessions   *cache.Cache
	expiration time.Duration
}

func NewStore(expiration time.Duration) *Store {
	return &Store{
		expiration: expiration,
		sessions:   cache.New(expiration, time.Second),
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := &Session{
		id: id,
	}
	err := s.sessions.Add(id, sess, s.expiration)
	return sess, err
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.sessions.Get(id)
	if !ok {
		return fmt.Errorf("session not found")
	}
	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions.Delete(id)
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.sessions.Get(id)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return val.(*Session), nil
}

type Session struct {
	id string

	// 想简单锁保护就用 sync.Map
	// 想更精细的锁保护就用 mutex.RWMutex
	values sync.Map
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	val, ok := s.values.Load(key)
	if !ok {
		return nil, errKeyNotFound
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, key string, val any) error {
	s.values.Store(key, val)
	return nil
}

func (s *Session) ID() string {
	return s.id
}
