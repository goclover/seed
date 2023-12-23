package seed

import "net/http"

const (
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch // RFC 5789
	MethodDelete  = http.MethodDelete
	MethodConnect = http.MethodConnect
	MethodOptions = http.MethodOptions
	MethodTrace   = http.MethodTrace
	MethodAny     = "ANY" //for all methods support
)

var supportMethods = append(httpMethods, MethodAny)

// httpMethods
var httpMethods = []string{
	MethodGet,
	MethodHead,
	MethodPost,
	MethodPut,
	MethodPatch,
	MethodDelete,
	MethodConnect,
	MethodOptions,
	MethodTrace,
}
