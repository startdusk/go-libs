package demo

import (
	"fmt"
	"go/ast"
)

type SingleFileVisitor struct {
	file *FileVisitor
}

func (spv *SingleFileVisitor) Get() *File {
	types := make([]Type, 0, len(spv.file.types))
	for _, typ := range spv.file.types {
		types = append(types, Type{
			Name:   typ.name,
			Fields: typ.fields,
		})
	}
	return &File{
		Package: spv.file.Package,
		Imports: spv.file.Imports,
		Types:   types,
	}
}

var _ ast.Visitor = &SingleFileVisitor{}

func (spv *SingleFileVisitor) Visit(node ast.Node) ast.Visitor {
	fn, ok := node.(*ast.File)
	if !ok {
		// 不是我们要的文件节点
		return spv
	}

	fv := &FileVisitor{
		// 在Visit中常用对象保存数据
		// 用对象保存go文件包名
		Package: fn.Name.String(),
	}
	spv.file = fv
	return fv
}

type FileVisitor struct {
	Package string
	Imports []string
	types   []*TypeVisitor
}

var _ ast.Visitor = &FileVisitor{}

func (fv *FileVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		v := &TypeVisitor{name: n.Name.String()}
		fv.types = append(fv.types, v)
		return v
	case *ast.ImportSpec:
		path := n.Path.Value
		if n.Name != nil && n.Name.String() != "" {
			// 处理导入包有别名的情况, 如 a "import/bbb"
			path = n.Name.String() + " " + path
		}
		fv.Imports = append(fv.Imports, path)
	}

	return fv
}

type TypeVisitor struct {
	name   string
	fields []Field
}

var _ ast.Visitor = &TypeVisitor{}

func (tv *TypeVisitor) Visit(node ast.Node) ast.Visitor {
	n, ok := node.(*ast.Field)
	if !ok {
		return tv
	}
	var typ string
	switch nt := n.Type.(type) {
	case *ast.Ident:
		typ = nt.String()
	case *ast.StarExpr:
		switch expr := nt.X.(type) {
		case *ast.Ident:
			typ = fmt.Sprintf("*%s", expr.String())
		case *ast.SelectorExpr:
			x := expr.X.(*ast.Ident).String()
			name := expr.Sel.String()
			typ = fmt.Sprintf("*%s.%s", x, name)
		}
	case *ast.SelectorExpr:
		x := nt.X.(*ast.Ident).String()
		name := nt.Sel.String()
		typ = fmt.Sprintf("%s.%s", x, name)
	case *ast.ArrayType:
		// 其它类型我们都不能处理处理，本来在 ORM 框架里面我们也没有支持
		switch ele := nt.Elt.(type) {
		case *ast.Ident:
			typ = fmt.Sprintf("[]%s", ele.String())
		}
		// 所以实际上我们在这里并没有处理 map, channel 之类的类型
	}
	for _, name := range n.Names {
		tv.fields = append(tv.fields, Field{
			Name: name.String(),
			Type: typ,
		})
	}
	return tv
}

type File struct {
	Package string
	Imports []string
	Types   []Type
}

type Type struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type string
}
