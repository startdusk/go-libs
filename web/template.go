package web

import (
	"bytes"
	"context"
	"text/template"
)

type TemplateEngine interface {
	// Render 渲染页面
	// tplName 模版名字, 按模版名字索引
	// data 渲染页面的数据
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}

type GoTemplateEngine struct {
	T *template.Template
}

func (t *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := t.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}
