package seed

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type Request interface {
	// HTTPRequest 返回原始 *http.Request
	HTTPRequest() *http.Request

	// Query 获取GET方式传递的参数
	Query(name string) (value string, has bool)

	// QueryDefault 获取GET方式传递的参数如果没有那么返回默认值/空值
	QueryDefault(name string, defaultValue ...string) (value string)

	// PostForm 获取POST方式传递的参数
	PostForm(name string) (value string, has bool)

	// PostFormDefault 获取POST方式传递的参数如果没有那么返回默认值/空值
	PostFormDefault(name string, defaultValue ...string) (value string)

	// Param 获取GET/POST方式传递的参数
	Param(name string) (value string, has bool)

	// ParamDefault 获取GET/POST方式传递的参数如果没有那么返回默认值/空值
	ParamDefault(name string, defaultValue ...string) (value string)

	// Header 获取Header传递的参数
	Header(name string) (value string, has bool)

	// HeaderDefault 获取Header方式传递的参数如果没有那么返回默认值/空值
	HeaderDefault(name string, defaultValue ...string) (value string)

	// Cookie 获取Cookie方式传递的参数
	Cookie(name string) (value *http.Cookie, has bool)

	// RemoteAddr 获取客户端的请求地址
	RemoteAddr() string

	// JsonUnmarshal json序列化参数到目标数据
	JsonUnmarshal(dst interface{}) error
}

type request struct {
	urlQuery url.Values
	bytes    []byte

	read bool
	*http.Request
}

func (r *request) HTTPRequest() *http.Request {
	return r.Request
}

func (r *request) Query(name string) (value string, has bool) {
	if r.Request.URL == nil {
		r.urlQuery = r.URL.Query()
	}
	values := r.urlQuery[name]
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func (r *request) QueryDefault(name string, defaultValue ...string) (value string) {
	var v string
	if v, _ = r.Query(name); v != "" {
		return v
	}
	if len(defaultValue) > 0 {
		v = defaultValue[0]
	}
	return v
}

func (r *request) PostForm(name string) (value string, has bool) {
	_ = r.Request.ParseForm()
	var vs = r.Request.PostForm[name]
	if len(vs) == 0 {
		return "", false
	}
	return vs[0], true
}

func (r *request) PostFormDefault(name string, defaultValue ...string) string {
	var v string
	if v, _ = r.PostForm(name); v != "" {
		return v
	}
	if len(defaultValue) > 0 {
		v = defaultValue[0]
	}
	return v
}

func (r *request) Param(name string) (value string, has bool) {
	if v, has := r.Query(name); has {
		return v, has
	}
	return r.PostForm(name)
}

func (r *request) ParamDefault(name string, defaultValue ...string) (value string) {
	var v string
	if v, has := r.Param(name); has {
		return v
	}
	if len(defaultValue) > 0 {
		v = defaultValue[0]
	}
	return v
}

func (r *request) Header(name string) (value string, has bool) {
	var vs = r.Request.Header.Values(name)
	if len(vs) == 0 {
		return "", false
	}
	return vs[0], true
}

func (r *request) HeaderDefault(name string, defaultValue ...string) (value string) {
	var v string
	if v, has := r.Header(name); has {
		return v
	}
	if len(defaultValue) > 0 {
		v = defaultValue[0]
	}
	return v
}

func (r *request) Cookie(name string) (value *http.Cookie, has bool) {
	var cookie, err = r.Request.Cookie(name)
	if err != nil {
		return nil, false
	}
	return cookie, true
}

func (r *request) RemoteAddr() string {
	return r.Request.RemoteAddr
}

func (r *request) JsonUnmarshal(dst interface{}) error {
	var err error
	if !r.read {
		if r.bytes, err = io.ReadAll(r.Body); err == nil {
			r.read = true
		}
	}
	err = json.Unmarshal(r.bytes, dst)
	return err
}

// NewRequest 返回Request实例
func NewRequest(req *http.Request) Request {
	return &request{Request: req}
}
