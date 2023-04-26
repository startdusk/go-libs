package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	var mockHandler HandleFunc = func(ctx Context) {}
	r := newRouter()
	for _, route := range cases {
		r.addRoute(route.method, route.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},

			http.MethodPost: &node{
				path: "/",
				children: map[string]*node{
					"order": &node{
						path: "order",
						children: map[string]*node{
							"create": &node{
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": &node{
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}

	msg, ok := wantRouter.equal(r)

	assert.True(t, ok, msg)

	// 路由错误
	panicRouter := newRouter()
	assert.Panics(t, func() {
		panicRouter.addRoute(http.MethodGet, "", mockHandler)
	})
	assert.Panics(t, func() {
		panicRouter.addRoute(http.MethodGet, "panic", mockHandler)
	})
	assert.Panics(t, func() {
		panicRouter.addRoute(http.MethodGet, "/panic/", mockHandler)
	})
	assert.Panics(t, func() {
		panicRouter.addRoute(http.MethodGet, "/panic//123", mockHandler)
	})

	// 路由重复注册
	panicReRegisterRouter := newRouter()
	assert.Panicsf(t, func() {
		panicReRegisterRouter.addRoute(http.MethodGet, "/", mockHandler)
		panicReRegisterRouter.addRoute(http.MethodGet, "/", mockHandler)
	}, "路由[/]重复注册")

	assert.Panicsf(t, func() {
		panicReRegisterRouter.addRoute(http.MethodGet, "/a/b/c", mockHandler)
		panicReRegisterRouter.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	}, "路由[/a/b/c]重复注册")
}

// 比较两个router是否相等
// 返回一个错误信息帮助排查问题和bool判断是否相等
func (r *router) equal(otherRouter *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := otherRouter.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的http method"), false
		}
		msg, equal := v.equal(dst)
		if !equal {
			return msg, false
		}
	}
	return "", true
}

// 比较两个node是否相等
// 返回一个错误信息帮助排查问题和bool判断是否相等
func (n *node) equal(otherNode *node) (string, bool) {
	if n.path != otherNode.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	if len(n.children) != len(otherNode.children) {
		return fmt.Sprintf("节点children数量不相等"), false
	}

	if reflect.ValueOf(n.handler) != reflect.ValueOf(otherNode.handler) {
		return fmt.Sprintf("Handler 不相等"), false
	}

	for path, c := range n.children {
		dst, ok := otherNode.children[path]
		if !ok {
			return fmt.Sprintf("子节点 [%s] 找不到", path), false
		}

		msg, equal := c.equal(dst)
		if !equal {
			return msg, false
		}
	}

	return "", true
}
