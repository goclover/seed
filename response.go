package seed

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	// HeaderContentType  HTTP Header 中 Content-Type 的 Key
	HeaderContentType = "Content-Type"

	// HeaderContentLength HTTP Header 中 Content-Length 的 Key
	HeaderContentLength = "Content-Length"
)

type Response interface {
	WriteTo(w http.ResponseWriter) error
}

type jsonResponse struct {
	statusCode int
	data       interface{}
}

func (j *jsonResponse) WriteTo(w http.ResponseWriter) error {
	var bs, err = json.Marshal(j.data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	writeHeaderIfNot(w.Header(), "application/json; charset=utf-8", strconv.Itoa(len(bs)))
	_, err = w.Write(bs)
	return err
}

var _ Response = &jsonResponse{}

// JsonResponse 返回JsonResponse
func JsonResponse(statusCode int, data interface{}) Response {
	return &jsonResponse{statusCode: statusCode, data: data}
}

type nopResponse struct {
	statusCode int
}

func (n *nopResponse) WriteTo(w http.ResponseWriter) error {
	w.WriteHeader(n.statusCode)
	return nil
}

var _ Response = &nopResponse{}

// NopResponse 返回一个nopResponse
func NopResponse(statusCode int) Response {
	return &nopResponse{statusCode: statusCode}
}

type htmlResponse struct {
	statusCode int
	html       string
}

func (h *htmlResponse) WriteTo(w http.ResponseWriter) error {
	var bs = []byte(h.html)
	writeHeaderIfNot(w.Header(), "text/plain; charset=utf-8", strconv.Itoa(len(bs)))

	var _, err = w.Write(bs)
	return err
}

var _ Response = &htmlResponse{}

func HtmlResponse(statusCode int, html string) Response {
	return &htmlResponse{statusCode: statusCode, html: html}
}

func writeHeaderIfNot(h http.Header, contentType, contentLen string) {
	if _, has := h[HeaderContentType]; !has {
		h[HeaderContentType] = []string{contentType}
	}
	h[HeaderContentLength] = []string{contentLen}
}
