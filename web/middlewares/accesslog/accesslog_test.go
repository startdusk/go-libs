package accesslog

import (
	"github.com/startdusk/go-libs/web"
	"net/http"
	"testing"
)

func Test_AccesslogBuilder(t *testing.T) {
	builder := MiddlewareBuilder{}
	accessLog := builder.LogFunc(func(log string) {
		t.Log(log)
	}).Build()
	hs := web.NewHTTPServer(web.ServerWithMiddleware(accessLog))
	hs.Post("/abc/*", func(ctx *web.Context) {
		t.Log("next call")
	})
	req, err := http.NewRequest(
		http.MethodPost,
		"/abc/123",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	hs.ServeHTTP(nil, req)
}
