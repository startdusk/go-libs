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

func newRouter() router {
	return router{
		trees: make(map[string]*node),
	}
}

// 可以看到该函数不支持多个HandleFunc(handleFunc ...HandleFunc)
// 因为用户可以传nil, 而且多个HandleFunc之间如果要中断, 必须提供像gin类似的Abort()方法
// 比较复杂, 且容易忘记添加
// method不检验的原因: 我们不暴露addRoute方法
// handleFunc不校验的原因: 如果用户传了nil, 那就相当于没有注册
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("路由[" + path + "]格式错误, 路由为空!")
	}

	if path[0] != '/' {
		panic("路由[" + path + "]格式错误, 不以 `/` 开头!")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("路由[" + path + "]格式错误, 不能以 `/` 结尾!")
	}

	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	if path == "/" {
		if root.handler != nil {
			panic("路由[" + path + "]重复注册")
		}
		root.handler = handleFunc
		return
	}

	trimPath := path[1:]
	segs := strings.Split(trimPath, "/")
	for _, seg := range segs {
		if seg == "" {
			panic("路由[" + path + "]格式错误!")
		}
		child := root.childOrCreate(seg)
		root = child
	}
	if root.handler != nil {
		panic("路由[" + path + "]重复注册")
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return &matchInfo{
			n: root,
		}, true
	}
	path = strings.Trim(path, "/")
	segs := strings.Split(path, "/")
	var pathParams map[string]string
	for _, seg := range segs {
		child, paramChild, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			// path 是 :id 这种形式
			pathParams[child.path[1:]] = seg
		}
		root = child
	}
	return &matchInfo{
		n:          root,
		pathParams: pathParams,
	}, true
}

type node struct {
	path string

	// 静态匹配的节点
	// 子path到子节点的映射
	children map[string]*node

	// 通配符节点
	starChild *node

	// 路径参数节点
	paramChild *node

	handler HandleFunc
}

func (n *node) childOrCreate(seg string) *node {
	if seg[0] == ':' {
		// 不允许同时注册路径参数和通配符匹配
		if n.starChild != nil {
			panic("不允许同时注册路径参数和通配符匹配, 已有通配符匹配")
		}
		child := &node{
			path: seg,
		}
		n.paramChild = child
		return child
	}
	if seg == "*" {
		if n.paramChild != nil {
			panic("不允许同时注册路径参数和通配符匹配, 已有路径参数匹配")
		}
		child := &node{
			path: seg,
		}
		n.starChild = child
		return child
	}
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

// childOf 优先静态匹配, 匹配不上再通配符匹配
// 第一个返回值是子节点
// 第二个返回值是标记是否是路径参数
// 第三个返回值是标记命中了没有
func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			// 参数路径是一个更具体的东西, 所以优先级要比通配符高
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	child, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			// 参数路径是一个更具体的东西, 所以优先级要比通配符高
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return child, false, ok
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}
