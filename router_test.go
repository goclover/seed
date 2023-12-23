package seed

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestRouter(t *testing.T) {
	var notFound http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("404 NOT FOUND"))
	}
	var rs = &router{
		mapper: &routeMapper{
			tree: map[string]*routeNode{},
		},
		notFound: notFound,
	}
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("index hello world"))
	}
	rs.Use(func(ctx context.Context, w http.ResponseWriter, req *http.Request, next MiddleWareQueue) bool {
		return next.Next(ctx, w, req)
	})
	rs.HandleStd(http.MethodPost, "/", h)
	rs.HandleStd("POST", "/", h)
	rs.HandleStd(http.MethodPost, "/a/b/c/d", h)
	rs.HandleStd(http.MethodPost, "/a/b/*/d", h)
	rs.HandleStd(http.MethodPost, "/a/b", h)
	rs.HandleStd(http.MethodGet, "/", h)

	var node = rs.mapper.(*routeMapper)
	var mode = node.tree["GET"]

	var req = &http.Request{
		URL: &url.URL{
			Path: "",
		},
		Method: "GET",
	}
	var r = rs.mapper.Find(req)
	_ = http.ListenAndServe(":8080", rs)
	fmt.Println(r)
	fmt.Println(mode)
}
