package seed

import (
	"context"
	"net/http"
)

// HandlerFunc 标准的HandlerFunc
type HandlerFunc func(ctx context.Context, req Request) Response

// Handler HandlerFunc自身转换为http.Handler
func (h HandlerFunc) Handler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		var request = NewRequest(r)
		var response = h(r.Context(), request)
		if response != nil {
			_ = response.WriteTo(w)
		}
	}
	return f
}

// NotFoundHandler 404默认处理器
var NotFoundHandler HandlerFunc = func(ctx context.Context, req Request) Response {
	return NopResponse(http.StatusNotFound)
}
