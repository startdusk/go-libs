package web

import (
	"strings"
)

// 用来支持对路由树的操作
// 代表路由树(森林)
type router struct {

	// HTTP method -> 路由树根节点
	trees map[string]*node
}

func newRouter() *router {
	return &router{
		trees: make(map[string]*node),
	}
}

// 可以看到该函数不支持多个HandleFunc(handleFunc ...HandleFunc)
// 因为用户可以传nil, 而且多个HandleFunc之间如果要中断, 必须提供像gin类似的Abort()方法
// 比较复杂, 且容易忘记添加
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	trimPath := strings.Trim(path, "/")
	segs := strings.Split(trimPath, "/")
	for _, seg := range segs {
		if seg == "" {
			panic("路由[" + path + "]格式错误!")
		}
		child := root.childOrCreate(seg)
		root = child
	}
	root.handler = handleFunc
}

type node struct {
	path string

	// 子path到子节点的映射
	children map[string]*node

	handler HandleFunc
}

func (n *node) childOrCreate(seg string) *node {
	if n.children == nil {
		n.children = make(map[string]*node)
	}

	child, ok := n.children[seg]
	if !ok {
		child = &node{
			path: seg,
		}
		n.children[seg] = child
	}
	return child
}
