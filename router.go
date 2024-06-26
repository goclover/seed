package seed

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var rexp = regexp.MustCompile(`/+`)

// RouteMapper 路由匹配器
type RouteMapper interface {
	Add(route Route) error
	Find(req *http.Request) Route
}

// Router 路由器
type Router interface {
	http.Handler

	// Use 注册过滤器(中间件)
	//
	// 	可用于给 server 增加切片的功能
	// 	如可以创建注册一个用于校验是否登录的 中间件
	// 	注册的中间件会对此后的 handler 生效，在此之前注册的 handler 方法不会生效
	Use(ms ...MiddlewareFunc) Router

	// HandleStd 以http.Handler方式注册业务handler
	//
	// 	method  是http方法，如GET、POST,也可以使用逗号来连接同时传入多个，如 "GET,POST"
	// 	也可以用特殊的 ANY,会自动注册所有( ANY 的取值详见 MethodAny )
	// 	handler 是业务的逻辑
	// 	ms 是该接口特有的中间件函数
	HandleStd(methods string, path string, handler http.Handler, ms ...MiddlewareFunc)

	// HandleFunc 以HandlerFunc方式注册业务handler
	//
	// 	method  是http方法，如GET、POST,也可以使用逗号来连接同时传入多个，如 "GET,POST"
	// 	也可以用特殊的 ANY,会自动注册所有( ANY 的取值详见 MethodAny )
	// 	handler 是业务的逻辑
	// 	ms 是该接口特有的中间件函数
	HandleFunc(methods string, path string, handlerFunc HandlerFunc, ms ...MiddlewareFunc)

	// Group 路由分组
	//
	// 	如 可以将 /user/xxx 系列分成一个分组
	// 	prefix路由前缀，如 "/user"
	// 	ms 是该分组的中间件函数
	Group(prefix string, f func(r Router), ms ...MiddlewareFunc)
}

// routeNode 路由匹配器节点
type routeNode struct {
	route    Route
	children map[string]*routeNode
}

// routeMapper 路由匹配器
type routeMapper struct {
	tree map[string]*routeNode
}

// router 路由器
type router struct {
	prefix          string
	mapper          RouteMapper
	middlewareFuncs MiddlewareFuncs
	notFound        http.Handler
}

// Find  实现 RouteMapper
func (rm *routeMapper) Find(req *http.Request) Route {
	if node, ok := rm.tree[req.Method]; ok {
		var segments = rm.segment(req.URL.Path)
		var nodes = []*routeNode{node}
		var matchedNode *routeNode

		for _, segment := range segments {
			var nextNodes []*routeNode
			for _, node = range nodes {
				if matchedNode, ok = node.children[segment]; ok {
					nextNodes = append(nextNodes, matchedNode)
				}
				if matchedNode, ok = node.children["*"]; ok {
					nextNodes = append(nextNodes, matchedNode)
				}
			}
			nodes = nextNodes
		}
		for _, v := range nodes {
			if v.route != nil {
				return v.route
			}
		}
	}
	return nil
}

// Add 实现 RouteMapper
func (rm *routeMapper) Add(route Route) error {
	var segments = rm.segment(route.Path())
	var node, ok = rm.tree[route.Method()]
	if !ok {
		node = &routeNode{children: make(map[string]*routeNode)}
		rm.tree[route.Method()] = node
	}

	for _, segment := range segments {
		if child, ok := node.children[segment]; ok {
			node = child
			continue
		}
		var newNode = &routeNode{children: make(map[string]*routeNode)}
		node.children[segment] = newNode
		node = newNode
	}

	if node.route == nil {
		node.route = route
		return nil
	}
	return fmt.Errorf("conflict route method: %s path: %s ", route.Method(), route.Path())
}

// segment 对路由path进行分段
func (rm *routeMapper) segment(path string) []string {
	path = rexp.ReplaceAllString(strings.Trim(strings.ToLower(path), "/"), "/")
	var segments = strings.Split(path, "/")
	if path == "/" || path == "" {
		segments = []string{"/"}
	}
	return segments
}

// HandleStd 标准handler方式注册路由
func (r *router) HandleStd(methods string, path string, handler http.Handler, ms ...MiddlewareFunc) {
	var seps = strings.Split(methods, ",")
	var sepMethods []string
	var validMethod = false

	for _, v := range seps {
		validMethod = false
		if v == MethodAny {
			sepMethods = append(sepMethods, httpMethods...)
			continue
		}
		for _, m := range httpMethods {
			if v == m {
				validMethod = true
				break
			}
		}
		if !validMethod {
			panic(fmt.Errorf("invalid route method: %s path: %s ", v, path))
		}
		sepMethods = append(sepMethods, strings.ToUpper(v))
	}

	for _, method := range sepMethods {
		var route = &route{
			path:    r.prefix + path,
			method:  method,
			Handler: r.TransHandler(handler, ms...),
		}
		if err := r.mapper.Add(route); err != nil {
			panic(err.Error())
		}
	}
}

// HandleFunc handlerFunc方式注册路由
func (r *router) HandleFunc(methods string, path string, handlerFunc HandlerFunc, ms ...MiddlewareFunc) {
	r.HandleStd(methods, path, handlerFunc.Handler(), ms...)
}

// Group 新建路由组
func (r *router) Group(prefix string, f func(r Router), ms ...MiddlewareFunc) {
	var mws = make([]MiddlewareFunc, len(r.middlewareFuncs))

	//copy middlewares
	_ = copy(mws, r.middlewareFuncs)
	mws = append(mws, ms...)

	//keep prefix
	prefix = r.prefix + prefix
	var router = &router{mapper: r.mapper, middlewareFuncs: mws, notFound: r.notFound, prefix: prefix}
	f(router)
}

// ServeHTTP 实现 http.Handler
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if route := r.mapper.Find(req); route != nil {
		route.ServeHTTP(w, req)
		return
	}
	if r.notFound == nil {
		r.notFound = r.TransHandler(NotFoundHandler.Handler())
	}
	r.notFound.ServeHTTP(w, req)
}

// Use 添加路由中间件
func (r *router) Use(ms ...MiddlewareFunc) Router {
	if len(ms) > 0 {
		r.middlewareFuncs = append(r.middlewareFuncs, ms...)
	}
	return r
}

// TransHandler 将Handler 合并当前路由中间件成实际的route handler
func (r *router) TransHandler(h http.Handler, ms ...MiddlewareFunc) http.Handler {
	ms = append(r.middlewareFuncs, ms...)
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		var mw MiddlewareFunc = func(ctx context.Context, ww http.ResponseWriter, rr *http.Request, next MiddleWareQueue) bool {
			h.ServeHTTP(ww, rr)
			return false
		}

		var mws MiddlewareFuncs = append(ms, mw)
		mws.Next(r.Context(), w, r)
	}
	return f
}

// NewRouter 返回一个Router实例
func NewRouter() Router {
	return &router{
		prefix:          "",
		mapper:          &routeMapper{tree: map[string]*routeNode{}},
		middlewareFuncs: []MiddlewareFunc{},
		notFound:        nil,
	}
}
