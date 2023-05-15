package cookie

import (
	"net/http"
)

// session id 放cookie里面, 为什么不放url? url长度有限制

type Propagator struct {
	cookieName   string
	cookieOption func(c *http.Cookie)
}

func NewPropagator(opts ...PropagatorOption) *Propagator {
	return &Propagator{
		cookieName:   "seesion-id",
		cookieOption: func(c *http.Cookie) {},
	}
}

type PropagatorOption func(p *Propagator)

func PropagatorWithCookieName(name string) PropagatorOption {
	return func(p *Propagator) {
		p.cookieName = name
	}
}

func PropagatorWithCookieOption(cookieOption func(c *http.Cookie)) PropagatorOption {
	return func(p *Propagator) {
		p.cookieOption = cookieOption
	}
}

func (p *Propagator) Inject(id string, writer http.ResponseWriter) error {
	c := &http.Cookie{
		Name:  p.cookieName,
		Value: id,
	}
	p.cookieOption(c)
	http.SetCookie(writer, c)
	return nil
}

func (p *Propagator) Extract(req *http.Request) (string, error) {
	c, err := req.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func (p *Propagator) Remove(writer http.ResponseWriter) error {
	c := &http.Cookie{
		Name:   p.cookieName,
		MaxAge: -1, // 设置为过期, 浏览器自动删除
	}
	http.SetCookie(writer, c)
	return nil
}
