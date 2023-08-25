package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response interface {
	HttpStatus() int
	Payload() interface{}
}

type response struct {
	Data       interface{} `json:"data,omitempty"`
	Success    bool        `json:"success"`
	httpStatus int
	Version    ApiVersion `json:"version"`
}

func (d response) Payload() interface{} {
	return d
}

func (d response) HttpStatus() int {
	return d.httpStatus
}

func NewResponse(data interface{}, success bool, status int, version ApiVersion) Response {
	return &response{Data: data, httpStatus: status, Success: success, Version: version}
}

func NewSuccessResponse(data interface{}, status int, version ApiVersion) Response {
	return NewResponse(data, true, status, version)
}

func WriteJsonResponse(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.HttpStatus())
	if err := json.NewEncoder(w).Encode(resp.Payload()); err != nil {
		panic(fmt.Errorf("error encoding json resp: %w", err))
	}
}
