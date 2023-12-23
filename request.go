package seed

import "net/http"

type Request interface{}

type request struct {
	req *http.Request
}

func NewRequest(req *http.Request) Request {
	return request{req: req}
}
