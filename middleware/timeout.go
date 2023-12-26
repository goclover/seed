package middleware

import (
	"context"
	"errors"
	"github.com/goclover/seed"
	"net/http"
	"time"
)

// Timeout is a middleware that cancels ctx after a given timeout and return
// a 504 Gateway Timeout error to the client.
//
// It's required that you select the ctx.Done() channel to check for the signal
// if the context has reached its deadline and return, otherwise the timeout
// signal will be just ignored.
func Timeout(timeout time.Duration) seed.MiddlewareFunc {
	return func(ctx context.Context, w http.ResponseWriter, req *http.Request, next seed.MiddleWareQueue) bool {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer func() {
			cancel()
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				w.WriteHeader(http.StatusGatewayTimeout)
			}
		}()
		return next.Next(ctx, w, req.WithContext(ctx))
	}
}
