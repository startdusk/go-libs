package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/startdusk/go-libs/web/session"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var (
	// errKeyNotFound 推荐使用小写, 当有人问你为什么不提供大写的时候, 你再改成大写(避免 增大错误的表面积)
	errKeyNotFound = errors.New("session: key not found")
)

// 使用hset
//
//	sess id    key     value
//
// map[string]map[string]string
type Store struct {
	client     redis.Cmdable
	expiration time.Duration
	prefix     string
}

func NewStore(client redis.Cmdable, opts ...StoreOption) *Store {
	s := &Store{
		expiration: 15 * time.Minute,
		client:     client,
		prefix:     "session-id",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type StoreOption func(s *Store)

func StoreWithPrefix(prefix string) StoreOption {
	return func(s *Store) {
		s.prefix = prefix
	}
}

func StoreWithExpiration(expiration time.Duration) StoreOption {
	return func(s *Store) {
		s.expiration = expiration
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	key := redisKey(s.prefix, id)
	_, err := s.client.HSet(ctx, key, id, id).Result()
	if err != nil {
		return nil, err
	}
	if _, err := s.client.Expire(ctx, key, s.expiration).Result(); err != nil {
		return nil, err
	}

	return &Session{
		id:     id,
		key:    key,
		client: s.client,
	}, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	key := redisKey(s.prefix, id)
	ok, err := s.client.Expire(ctx, key, s.expiration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("session [%s] not found", id)
	}
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	key := redisKey(s.prefix, id)
	_, err := s.client.Del(ctx, key).Result()
	return err
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	// 自由决策要不要提前把 session 存储的用户数据一并捞过来
	// 1.都不拿
	// 2.只拿高频数据(热点数据)
	// 3.都拿

	// 这里选第一种, 都不拿
	key := redisKey(s.prefix, id)
	return &Session{
		id:     id,
		key:    key,
		client: s.client,
	}, nil
}

type Session struct {
	id     string
	key    string
	client redis.Cmdable
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	val, err := s.client.HGet(ctx, s.key, key).Result()
	return val, err
}

func (s *Session) Set(ctx context.Context, key string, val any) error {
	const lua = `
		if redis.call("exists", KEYS[1]) 
		then
			return redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
		else
			return -1
		end
	`
	res, err := s.client.Eval(ctx, lua, []string{s.key}, key, val).Int()
	if err != nil {
		return err
	}
	if res < 0 {
		return errKeyNotFound
	}

	return nil
}

func (s *Session) ID() string {
	return s.id
}

func redisKey(prefix string, id string) string {
	return prefix + "-" + id
}
