package demo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"

	_ "embed"
	"text/template"
)

//go:embed tpl.gohtml
var genOrm string

type Data struct {
	*File
	Ops []string
}

func Gen(w io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	v := &SingleFileVisitor{}
	ast.Walk(v, f)
	file := v.Get()

	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(w, Data{
		File: file,
		Ops:  []string{"LT", "GT", "EQ"},
	})
}
