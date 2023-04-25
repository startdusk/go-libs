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
	}

	var mockHandler HandleFunc = func(ctx Context) {}
	r := newRouter()
	for _, route := range cases {
		r.addRoute(route.method, route.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path: "/",
				children: map[string]*node{
					"user": &node{
						path: "user",
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
				},
			},
		},
	}

	msg, ok := wantRouter.equal(r)

	assert.True(t, ok, msg)
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
