package web

import "net/http"

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Error struct {
	Error      errorDetail `json:"error"`
	Success    bool        `json:"success"`
	httpStatus int
	Version    ApiVersion `json:"version"`
}

func (e Error) Payload() interface{} {
	return e
}

func (e Error) HttpStatus() int {
	return e.httpStatus
}

const (
	UnauthorizedRequest = "unauthorized"
	BadRequest          = "bad_request"
	Forbidden           = "forbidden"
	InternalServerError = "internal_server_error"
)

var (
	ErrUnauthenticatedRequest = func(desc string, version ApiVersion) Error {
		return NewError(UnauthorizedRequest, desc, http.StatusUnauthorized, version)
	}
	ErrUnauthorizedRequest = func(desc string, version ApiVersion) Error {
		return NewError(UnauthorizedRequest, desc, http.StatusForbidden, version)
	}
	ErrForbiddenRequest = func(desc string, version ApiVersion) Error {
		return NewError(Forbidden, desc, http.StatusForbidden, version)
	}
	ErrBadRequest = func(desc string, version ApiVersion) Error {
		return NewError(BadRequest, desc, http.StatusBadRequest, version)
	}
	ErrInternalServerError = func(desc string, version ApiVersion) Error {
		return NewError(InternalServerError, desc, http.StatusInternalServerError, version)
	}
)

func NewError(errCode string, desc string, httpCode int, version ApiVersion) Error {
	return Error{Error: errorDetail{Code: errCode, Message: desc}, httpStatus: httpCode, Version: version}
}
