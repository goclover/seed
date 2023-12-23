package seed

import (
	"encoding/json"
	"net/http"
)

type Response interface {
	WriteTo(w http.ResponseWriter) error
}

type jsonResponse struct {
	StatusCode int
	Data       interface{}
}

// WriteTo 实现Response interface
func (j *jsonResponse) WriteTo(w http.ResponseWriter) error {
	var bs, err = json.Marshal(j.Data)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(j.StatusCode)
	_, err = w.Write(bs)
	return err
}

// JsonResponse 返回JsonResponse
func JsonResponse(statusCode int, data interface{}) Response {
	return &jsonResponse{StatusCode: statusCode, Data: data}
}
