package session

import (
	"github.com/google/uuid"
	"github.com/startdusk/go-libs/web"
)

type Manager struct {
	Propagator
	Store
	CtxSessKey string
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	// 其实也可以使用context存数据, 但是会拷贝http.Request, 效率不高
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 1)
	}
	// 尝试缓存住Session
	cacheSess, ok := ctx.UserValues[m.CtxSessKey]
	if ok {
		return cacheSess.(Session), nil
	}
	sessID, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	sess, err := m.Get(ctx.Req.Context(), sessID)
	if err != nil {
		return nil, err
	}
	ctx.UserValues[m.CtxSessKey] = sess
	return sess, nil
}

func (m *Manager) InitSession(ctx *web.Context) (Session, error) {
	id := uuid.New().String()
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	return sess, m.Inject(id, ctx.Resp)
}

func (m *Manager) RemoveSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	if err := m.Store.Remove(ctx.Req.Context(), sess.ID()); err != nil {
		return err
	}

	if err := m.Propagator.Remove(ctx.Resp); err != nil {
		return err
	}
	delete(ctx.UserValues, m.CtxSessKey)
	return nil
}

func (m *Manager) RefreshSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	return m.Refresh(ctx.Req.Context(), sess.ID())
}
