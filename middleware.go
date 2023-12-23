package seed

import (
	"context"
	"net/http"
)

// MiddleWareQueue 中间件执行队列
type MiddleWareQueue interface {
	Next(ctx context.Context, w http.ResponseWriter, req *http.Request) bool
}

// MiddlewareFunc 中间件执行器
type MiddlewareFunc func(ctx context.Context, w http.ResponseWriter, req *http.Request, next MiddleWareQueue) bool

// MiddlewareFuncs 中间件执行器队列示例
type MiddlewareFuncs []MiddlewareFunc

// Next 触发下一个 MiddleWareFunc 或者是业务 handler
//
// 执行顺序： filter1 -> filter2 -> filter3
// 若 filter2 返回 false，调用将终止，即 filter3 不会被执行
func (ms MiddlewareFuncs) Next(ctx context.Context, w http.ResponseWriter, req *http.Request) bool {
	if len(ms) <= 0 {
		return false
	}
	var f = ms[0]
	return f(ctx, w, req, ms[1:])
}

// NewMiddleWareQueue 创建一个中间件执行队列
func NewMiddleWareQueue(funcs ...MiddlewareFunc) MiddleWareQueue {
	return MiddlewareFuncs(funcs)
}
