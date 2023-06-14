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
// mdls 中间件
func (r *router) addRoute(method string, path string, handleFunc HandleFunc, mdls ...Middleware) {
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
	root.mdls = mdls
}

// 查找路由
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
	mi := &matchInfo{}
	cur := root
	for _, seg := range segs {
		child, found := cur.childOf(seg)
		if !found {
			// 最后一段为 * 通配符
			if cur.typ == nodeTypeAny {
				mi.n = cur
				return mi, true
			}
			return nil, false
		}
		if child.paramName != "" {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			pathParams[child.paramName] = seg
		}
		cur = child
	}

	mi.n = cur
	mi.pathParams = pathParams
	mi.ms = r.findMiddlewares(root, segs)
	return mi, true
}

// 查找路由树上的中间件
func (r *router) findMiddlewares(root *node, segs []string) []Middleware {
	// 层次遍历(广度优先)路由树, 找到middleware(目前每次查找路由都得计算一遍)
	queue := []*node{root}
	mdls := make([]Middleware, 0, 16)
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		// 保存每一段会命中的子节点
		var children []*node
		for _, cur := range queue {
			if len(cur.mdls) > 0 {
				mdls = append(mdls, cur.mdls...)
			}
			children = append(children, cur.childrenOf(seg)...)
		}
		queue = children
	}

	// 收尾
	for _, cur := range queue {
		if len(cur.mdls) > 0 {
			mdls = append(mdls, cur.mdls...)
		}
	}

	// 优化方案
	// 提前计算好路由树上的中间件, 并保存结果集, 查找的时候就找结果集就行了
	return mdls
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
	ms         []Middleware
}
