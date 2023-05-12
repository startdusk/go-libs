package session

import (
	"context"
	"net/http"
)

// Store 管理 Session
type Store interface {
	// session 对应的 ID
	Generate(ctx context.Context, id string) (Session, error)

	Refresh(ctx context.Context, id string) error

	Remove(ctx context.Context, id string) error

	Get(ctx context.Context, id string) (Session, error)
}

type Session interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any) error
	ID() string
}

type Propagator interface {
	Inject(id string, writer http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	Remove(writer http.ResponseWriter) error
}
