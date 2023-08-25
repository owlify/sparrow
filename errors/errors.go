package errors

import (
	stderrors "errors"
	"fmt"

	pkgError "github.com/pkg/errors"
)

var (
	ErrBadRequest = NewWithCode("bad_request")
)

type baseError struct {
	code string
	msg  string
}

func (f *baseError) Error() string {
	return f.msg
}

func (f *baseError) Code() string {
	return f.code
}

func New(message string) error {
	return &baseError{
		msg: message,
	}
}

func NewWithErr(code string, err error) error {
	if err == nil {
		return nil
	}

	return &baseError{
		code: code,
		msg:  err.Error(),
	}
}

func NewWithCode(code string) error {
	return &baseError{
		code: code,
	}
}

func NewWithCodef(code string, format string, args ...interface{}) error {
	return &baseError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
}

func Wrap(err error, message string) error {
	return pkgError.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return pkgError.Wrapf(err, format, args...)
}

func Original(err error) *baseError {
	e := &baseError{}
	stderrors.As(err, &e)
	return e
}

func Is(err error, target error) bool {
	if err == nil {
		return false
	}

	e := Original(err)
	et := Original(target)

	return e.Code() == et.Code()
}
