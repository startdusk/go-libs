package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
)

type Context struct {
	Req  *http.Request
	Resp http.ResponseWriter

	// RespData 和 RespStatusCode 主要是为了 midleware 读写用的
	// 如果用户直接使用了 Resp
	// 那么他就直接绕开了 RespData 和 RespStatusCode 这两个
	// 那么部分 middleware 就无法运作

	// 响应数据
	RespData []byte
	// 响应状态码
	RespStatusCode int

	// 缓存URL路径参数
	PathParams map[string]string

	// 缓存URL Query参数
	queryValues url.Values

	// 命中的路由
	MatchedRoute string

	tplEngine TemplateEngine
}

func (c *Context) BindJSON(val any) error {
	if reflect.TypeOf(val).Kind() != reflect.Pointer {
		return errors.New("参数不为指针")
	}

	if c.Req.Body == nil {
		return errors.New("body为空")
	}

	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) (string, error) {
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}

	return c.Req.FormValue(key), nil
}

func (c *Context) QueryValue(key string) (string, error) {
	// Query 和 Form 表单比起来，它没有缓存
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}

	vals, ok := c.queryValues[key]
	if !ok {
		return "", errors.New("key不存在")
	}

	return vals[0], nil
}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("key不存在")
	}
	return val, nil
}

func (c *Context) RespJSON(code int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespStatusCode = code
	c.RespData = data
	return nil
}

func (c *Context) Render(tplName string, data any) error {
	renderData, err := c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	c.RespData = renderData
	c.RespStatusCode = http.StatusOK
	return nil
}
