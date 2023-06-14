package web

import (
	"regexp"
	"strings"
)

type nodeType int

const (
	// 静态路由
	nodeTypeStatic nodeType = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

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

	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式节点
	regChild *node
	// 正则表达式节点(如 `/req/:id(.*)` )
	regexps *regexp.Regexp

	handler HandleFunc

	// 路由树类型
	typ nodeType

	// 该路由带的中间件函数
	mdls []Middleware
}

func (n *node) childOrCreate(seg string) *node {

	// 以 : 开头的, 需要进一步解析, 判断是参数路由还是正则路由
	if seg[0] == ':' {
		paramName, expr, isReg := n.parseParam(seg)
		if isReg {
			return n.childOrCreateReg(seg, expr, paramName)
		}
		return n.childOrCreateParam(seg, paramName)
	}

	if seg == "*" {
		// 3.匹配通配符
		if n.paramChild != nil {
			panic("不允许同时注册路径参数和通配符匹配, 已有路径参数匹配")
		}

		if n.regChild != nil {
			panic("不允许同时注册正则匹配和通配符匹配, 已有正则匹配")
		}

		// 已注册，就直接返回
		if n.starChild != nil {
			return n.starChild
		}

		n.starChild = &node{
			path: seg,
			typ:  nodeTypeAny,
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
			typ:  nodeTypeStatic,
		}
		n.children[seg] = child
	}
	return child
}

// 匹配正则路径
func (n *node) childOrCreateReg(seg string, expr string, paramName string) *node {
	// 不允许同时注册正则匹配和通配符匹配
	if n.starChild != nil {
		panic("不允许同时注册正则匹配和通配符匹配, 已有通配符匹配")
	}

	// 不允许同时注册正则匹配和路径参数
	if n.paramChild != nil {
		panic("不允许同时注册正则匹配和路径参数, 已有路径参数")
	}

	if n.regChild != nil {
		if n.regChild.regexps.String() != expr || n.paramName != paramName {
			panic("路由冲突, 正则路由冲突")
		}
		return n.regChild
	}

	r, err := regexp.Compile(expr)
	if err != nil {
		panic("正则表达式路由错误")
	}
	n.regChild = &node{
		path:      seg,
		typ:       nodeTypeReg,
		paramName: paramName,
		regexps:   r,
	}

	return n.regChild
}

// 匹配参数路径
func (n *node) childOrCreateParam(seg string, paramName string) *node {
	// 不允许同时注册路径参数和通配符匹配
	if n.starChild != nil {
		panic("不允许同时注册路径参数和通配符匹配, 已有通配符匹配")
	}

	// 不允许同时注册路径参数和正则匹配
	if n.regChild != nil {
		panic("不允许同时注册路径参数和正则匹配, 已有正则匹配")
	}

	if n.paramChild != nil {
		if n.paramChild.path != seg {
			panic("路由冲突, 参数路由冲突")
		}
		return n.paramChild
	}

	n.paramChild = &node{
		path:      seg,
		typ:       nodeTypeParam,
		paramName: paramName,
	}
	return n.paramChild
}

// parseParam 用于解析判断是不是正则表达式
// eg: :id(.+)
// 第一个返回值是参数名字
// 第二个返回值是正则表达式
// 第三个返回值是bool, 返回true说明是正则路由
func (n *node) parseParam(seg string) (string, string, bool) {
	// 去除 :
	seg = seg[1:]
	ss := strings.SplitN(seg, "(", 2)
	if len(ss) == 2 {
		expr := ss[1]
		if strings.HasSuffix(expr, ")") {
			return ss[0], expr[:len(expr)-1], true
		}
	}
	return seg, "", false
}

// childOf 优先静态匹配, 匹配不上再通配符匹配
// 第一个返回值是子节点
// 第二个返回值是标记命中了没有
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.childOfNonStatic(path)
	}
	child, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return child, ok
}

// childOfNonStatic 从非静态匹配的子节点中查找
func (n *node) childOfNonStatic(seg string) (*node, bool) {
	if n.regChild != nil && n.regChild.regexps.Match([]byte(seg)) {
		return n.regChild, true
	}

	if n.paramChild != nil {
		return n.paramChild, true
	}

	return n.starChild, n.starChild != nil
}

// childrenOf 返回该层的所有子节点
func (n *node) childrenOf(seg string) []*node {
	var nodes []*node
	if n.children != nil {
		if child, ok := n.children[seg]; ok {
			nodes = append(nodes, child)
		}
	}

	if n.regChild != nil {
		nodes = append(nodes, n.regChild)
	}

	if n.paramChild != nil {
		nodes = append(nodes, n.paramChild)
	}

	if n.starChild != nil {
		nodes = append(nodes, n.starChild)
	}

	return nodes
}
