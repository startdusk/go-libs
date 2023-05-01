package web

import (
	"regexp"
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
		root.fullPath = path
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
	root.fullPath = path
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

			if child.regexps != nil {
				if !child.regexps.MatchString(seg) {
					return nil, false
				}

				// path 是 :id 这种形式
				pathParams[child.path[1:]] = seg
			} else {
				// path 是 :id 这种形式
				pathParams[child.path[1:]] = seg
			}
		}
		root = child
	}
	return &matchInfo{
		n:          root,
		pathParams: pathParams,
	}, true
}

type node struct {
	// 命中路由的完整路径
	fullPath string
	// 命中路由的那段路径
	path string

	// 静态匹配的节点
	// 子path到子节点的映射
	children map[string]*node

	// 通配符节点
	starChild *node

	// 路径参数节点
	paramChild *node

	// 正则表达式节点(如 `/req/:id(.*)` )
	regexps *regexp.Regexp

	handler HandleFunc
}

func (n *node) childOrCreate(seg string) *node {

	if seg[0] == ':' {
		// 不允许同时注册路径参数和通配符匹配
		if n.starChild != nil {
			panic("不允许同时注册路径参数和通配符匹配, 已有通配符匹配")
		}

		// 1.正则匹配
		ok, param, reg := splitSegment(seg)
		if ok {
			if n.paramChild != nil {
				if n.paramChild.path != param {
					panic("不允许同时注册相同路径的不同的正则匹配")
				}
				if n.paramChild.regexps == nil {
					panic("不允许同时注册路径参数和正则匹配, 已有路径参数匹配")
				}
				if n.paramChild.regexps.String() != reg {
					panic("不允许同时注册相同路径的不同的正则匹配")
				}
				return n.paramChild
			}
			n.paramChild = &node{
				path: param,
			}

			r, err := regexp.Compile(reg)
			if err != nil {
				panic("正则表达式路由错误")
			}
			n.paramChild.regexps = r
			return n.paramChild
		}

		if n.paramChild != nil {
			return n.paramChild
		}

		// 2.匹配参数路径
		n.paramChild = &node{
			path: seg,
		}
		return n.paramChild
	}

	if seg == "*" {
		// 3.匹配通配符
		if n.paramChild != nil {
			panic("不允许同时注册路径参数和通配符匹配, 已有路径参数匹配")
		}

		// 已注册，就直接返回
		if n.starChild != nil {
			return n.starChild
		}

		n.starChild = &node{
			path: seg,
		}
		return n.starChild
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

func splitSegment(key string) (bool, string, string) {
	var param string
	var reg string
	var startReg bool
	for _, c := range key {
		switch {
		case c == '(':
			startReg = true
			reg += string(c)
		case c == ')' && startReg:
			startReg = false
			reg += string(c)
			return true, param, reg
		case !startReg:
			param += string(c)
		case startReg:
			reg += string(c)
		}
	}

	return len(reg) != 0, param, reg
}
