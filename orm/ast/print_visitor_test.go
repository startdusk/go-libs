package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_PrintVisitor(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", `
package ast

import (
	"fmt"
	"go/ast"
	"reflect"
)

type PrintVisitor struct{}

var _ ast.Visitor = PrintVisitor{}

func (p PrintVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return p
	}
	typ := reflect.TypeOf(node)
	val := reflect.ValueOf(node)
	// 解引用, 拿到指针指向的数据
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	fmt.Printf("type: %s, val: %v\n", typ.Name(), val.Interface())
	return p
}

	`, parser.ParseComments)
	require.NoError(t, err)
	v := &PrintVisitor{}
	ast.Walk(v, f)
}
