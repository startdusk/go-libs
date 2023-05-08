package web

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 匹配优先级 静态匹配 > 正则匹配 > 参数匹配(路径参数匹配可以看做是正则匹配的一种特殊形态，例如 :id(.+)。比路径参数更精准) > 通配符匹配
func TestRouter_addRoute(t *testing.T) {
	routes := []struct {
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
			method: http.MethodGet,
			path:   "/order/detail/:id",
		},
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		// 正则路由(已测试通过)
		{
			method: http.MethodDelete,
			path:   "/req/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}
	r := newRouter()
	for _, route := range routes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	idReg := regexp.MustCompile("(.*)")
	nameReg := regexp.MustCompile("(^.+$)")
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				fullPath: "/",
				path:     "/",
				handler:  mockHandler,
				starChild: &node{
					path:     "*",
					fullPath: "/*",
					handler:  mockHandler,
					starChild: &node{
						path:     "*",
						fullPath: "/*/*",
						handler:  mockHandler,
					},
					children: map[string]*node{
						"abc": {
							path:     "abc",
							fullPath: "/*/abc",
							handler:  mockHandler,
							starChild: &node{
								path:     "*",
								fullPath: "/*/abc/*",
								handler:  mockHandler,
							},
						},
					},
				},
				children: map[string]*node{
					"user": {
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": {
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:    "detail",
								handler: mockHandler,
								paramChild: &node{
									path:    ":id",
									handler: mockHandler,
								},
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
						},
					},
				},
			},

			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {
						path: "order",
						children: map[string]*node{
							"create": {
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": {
						path:    "login",
						handler: mockHandler,
					},
				},
			},

			http.MethodDelete: {
				path: "/",
				paramChild: &node{
					path:    ":name",
					regexps: nameReg,
					children: map[string]*node{
						"abc": {
							path:    "abc",
							handler: mockHandler,
						},
					},
				},
				children: map[string]*node{
					"req": {
						path: "req",
						paramChild: &node{
							path:    ":id",
							handler: mockHandler,
							regexps: idReg,
						},
					},
				},
			},
		},
	}

	msg, ok := r.equal(wantRouter)

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

	// 同时存在路径参数匹配和通配符匹配
	panicBothRouter := newRouter()
	assert.Panicsf(t, func() {
		panicBothRouter.addRoute(http.MethodGet, "/a/*", mockHandler)
		panicBothRouter.addRoute(http.MethodGet, "/a/:id", mockHandler)
	}, "不允许同时注册路径参数和通配符匹配, 已有通配符匹配")

	assert.Panicsf(t, func() {
		panicBothRouter.addRoute(http.MethodGet, "/a/:id", mockHandler)
		panicBothRouter.addRoute(http.MethodGet, "/a/*", mockHandler)
	}, "不允许同时注册路径参数和通配符匹配, 已有路径参数匹配")

	// TODO: 同时存在 正则匹配
}

func TestRouter_findRoute(t *testing.T) {
	routes := []struct {
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
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodPost,
			path:   "/login/:username",
		},

		// 通配符路由
		// 支持 /a/b/* 匹配 /a/b/c/d/e... 目前只支持 匹配到 /a/b/c
		{
			method: http.MethodGet,
			path:   "/a/b/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},

		{
			method: http.MethodDelete,
			path:   "/req/:id([0-9]+)",
		},
		{
			method: http.MethodDelete,
			path:   "/:name(^.+$)/abc",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {}
	r := newRouter()
	for _, route := range routes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	cases := []struct {
		name          string
		method        string
		path          string
		wantFound     bool
		wantMatchInfo *matchInfo
	}{
		// ================================================================================
		// 测试通配符路由
		{
			name:      "/xxx 命中 /* 路由",
			method:    http.MethodGet,
			path:      "/xxx",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
					starChild: &node{
						path:    "*",
						handler: mockHandler,
					},
					children: map[string]*node{
						"abc": {
							path:    "abc",
							handler: mockHandler,
							starChild: &node{
								path:    "*",
								handler: mockHandler,
							},
						},
					},
				},
			},
		},
		{
			name:      "/xxx/xxx 命中 /*/* 路由",
			method:    http.MethodGet,
			path:      "/xxx/xxx",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "/xxx/abc 命中 /*/abc 路由",
			method:    http.MethodGet,
			path:      "/xxx/abc",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "abc",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "/xxx/abc/xxx 命中 /*/abc/* 路由",
			method:    http.MethodGet,
			path:      "/xxx/abc/xxx",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},

		// ================================================================================
		// 测试正则匹配路由
		{
			name:      "/req/124 命中 /req/:id([0-9]+) 路由",
			method:    http.MethodDelete,
			path:      "/req/124",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    ":id([0-9]+)",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"id": "124",
				},
			},
		},
		{
			name:      "/req/xxxx 不能命中 /req/:id([0-9]+) 路由",
			method:    http.MethodDelete,
			path:      "/req/xxxx",
			wantFound: false,
		},
		{
			name:      "/123/abc 命中 /:name(^.+$)/abc 路由",
			method:    http.MethodDelete,
			path:      "/123/abc",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "abc",
					handler: mockHandler,
				},
				pathParams: map[string]string{
					"name": "123",
				},
			},
		},
		{
			name:      "/req/abc 不能命中 /:name(^.+$)/abc 路由(req是静态路由,优先级最高)",
			method:    http.MethodDelete,
			path:      "/req/abc",
			wantFound: false,
		},

		{
			name:      "method找不到",
			method:    http.MethodOptions,
			path:      "/找不到",
			wantFound: false,
		},
		// 因为存在 /* 所以所有路径都能匹配到。。
		// {
		// 	name:      "path找不到",
		// 	method:    http.MethodGet,
		// 	path:      "/找不到/找不到/找不到",
		// 	wantFound: false,
		// },
		{
			name:      "命中, 但该路由无handler",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path: "order",
					children: map[string]*node{
						"detail": {
							handler: mockHandler,
							path:    "detail",
						},
					},
				},
			},
		},
		{
			name:      "/a/b/c 命中 /a/b/*",
			method:    http.MethodGet,
			path:      "/a/b/c",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					handler: mockHandler,
					path:    "*",
				},
			},
		},
		{
			name:      "/a/b/c/d/e 命中 /a/b/*",
			method:    http.MethodGet,
			path:      "/a/b/c/d/e",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					handler: mockHandler,
					path:    "*",
				},
			},
		},
		{
			name:      "命中",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					handler: mockHandler,
					path:    "detail",
				},
			},
		},
		{
			name:      "命中通配符*",
			method:    http.MethodGet,
			path:      "/order/abc",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					handler: mockHandler,
					path:    "*",
				},
			},
		},
		{
			name:      "命中参数路径 :username",
			method:    http.MethodPost,
			path:      "/login/startdusk",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					handler: mockHandler,
					path:    ":username",
				},
				pathParams: map[string]string{
					"username": "startdusk",
				},
			},
		},
		{
			name:      "命中根节点'/'",
			method:    http.MethodGet,
			path:      "/",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "/",
					handler: mockHandler,
					children: map[string]*node{
						"a": {
							path: "a",
							children: map[string]*node{
								"b": {
									path: "b",
									starChild: &node{
										path:    "*",
										handler: mockHandler,
									},
								},
							},
						},
						"user": {
							path:    "user",
							handler: mockHandler,
							children: map[string]*node{
								"home": {
									path:    "home",
									handler: mockHandler,
								},
							},
						},
						"order": {
							path: "order",
							children: map[string]*node{
								"detail": {
									path:    "detail",
									handler: mockHandler,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			matchInfo, found := r.findRoute(c.method, c.path)
			assert.Equal(t, c.wantFound, found)
			if !c.wantFound {
				return
			}

			assert.Equal(t, c.wantMatchInfo.pathParams, matchInfo.pathParams)
			msg, ok := c.wantMatchInfo.n.equal(matchInfo.n)
			assert.True(t, ok, msg)
		})
	}
}

// 比较两个router是否相等
// 返回一个错误信息帮助排查问题和bool判断是否相等
func (r *router) equal(otherRouter *router) (string, bool) {
	if len(r.trees) != len(otherRouter.trees) {
		return fmt.Sprintf("trees 数量不相等: expect %d actual %d", len(r.trees), len(otherRouter.trees)), false
	}
	for k, v := range r.trees {
		dst, ok := otherRouter.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的http method: %s", k), false
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
		return fmt.Sprintf("节点路径不匹配: expect %s actual %s", n.path, otherNode.path), false
	}

	if len(n.children) != len(otherNode.children) {
		return fmt.Sprintf("%s, 节点children数量不相等: expect %d actual %d", n.path, len(n.children), len(otherNode.children)), false
	}

	if n.starChild != nil {
		msg, ok := n.starChild.equal(otherNode.starChild)
		if !ok {
			return msg, ok
		}
	}

	if n.paramChild != nil {
		msg, ok := n.paramChild.equal(otherNode.paramChild)
		if !ok {
			return msg, ok
		}
	}

	if reflect.ValueOf(n.handler) != reflect.ValueOf(otherNode.handler) {
		return fmt.Sprintf("%s Handler 不相等, expect %v actual %v", n.path, reflect.ValueOf(n.handler), reflect.ValueOf(otherNode.handler)), false
	}

	if n.regexps != nil {
		if otherNode.regexps == nil {
			return fmt.Sprintf("%s 缺少正则表达式, expect %s actual nil", n.path, n.regexps.String()), false
		}
		if n.regexps.String() != otherNode.regexps.String() {
			return fmt.Sprintf("%s 正则表达式不匹配, expect %s actual %s", n.path, n.regexps.String(), otherNode.regexps.String()), false
		}
	}

	for path, c := range n.children {
		dst, ok := otherNode.children[path]
		if !ok {
			return fmt.Sprintf("%s 子节点 [%s] 找不到", n.fullPath, path), false
		}
		msg, equal := c.equal(dst)
		if !equal {
			return msg, false
		}
	}

	return "", true
}
