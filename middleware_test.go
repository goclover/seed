package seed

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestMiddleware(t *testing.T) {
	var f1 = func(ctx context.Context, w http.ResponseWriter, req *http.Request, next MiddleWareQueue) bool {
		fmt.Println("run f1")
		return next.Next(ctx, w, req)
	}
	var f2 = func(ctx context.Context, w http.ResponseWriter, req *http.Request, next MiddleWareQueue) bool {
		fmt.Println("run f2")
		return next.Next(ctx, w, req)
	}
	var f3 = func(ctx context.Context, w http.ResponseWriter, req *http.Request, next MiddleWareQueue) bool {
		fmt.Println("run f3")
		return false
	}
	var ms = MiddlewareFuncs{}
	ms = append(ms, f1, f2, f3)
	fmt.Printf("%+v\n", ms)
	ms.Next(context.Background(), nil, nil)
}
