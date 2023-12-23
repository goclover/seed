package seed

import "net/http"

type Route interface {
	http.Handler

	Path() string
	Method() string
}

type route struct {
	http.Handler

	path   string
	method string
}

// Path 返回当前请求的path
func (r *route) Path() string {
	return r.path
}

// Method 返回当前路由的请求方法
func (r *route) Method() string {
	return r.method
}
