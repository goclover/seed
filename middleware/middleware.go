package middleware

import (
	"context"
	"github.com/goclover/seed"
	"net/http"
)

// New will create a new middleware from a http.Handler.
func New(h http.Handler) seed.MiddlewareFunc {
	return func(ctx context.Context, w http.ResponseWriter, req *http.Request, next seed.MiddleWareQueue) bool {
		h.ServeHTTP(w, req)
		return next.Next(ctx, w, req)
	}
}
