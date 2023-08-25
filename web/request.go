package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/owlify/sparrow/errors"
	"github.com/owlify/sparrow/logger"
)

type Request struct {
	*http.Request

	route      string
	pathParams map[string]string
	params     map[string]string
}

var headersToBeMasked = []string{"Authorization"}

type ValidationErrorInterface interface {
	Type() string
	Error() string
	Unwrap() error
	IsUnexpectedErr() bool
}

type ValidationError struct {
	errorType string
	message   string
	err       error
}

func (e *ValidationError) Error() string { return e.message }

func (e *ValidationError) Type() string { return e.errorType }

func (e *ValidationError) Unwrap() error { return e.err }

func (e *ValidationError) IsUnexpectedErr() bool { return e.errorType == UnexpectedErr }

var (
	UnexpectedErr  = "Unexpected"
	ErrInvalidType = func(field string, expectedType interface{}, err error) ValidationErrorInterface {
		return &ValidationError{
			errorType: "InvalidType",
			message:   fmt.Sprintf("InvalidType for field: %s. Expected: %s", field, expectedType),
			err:       err,
		}
	}
	ErrInvalidJson = func(err error) ValidationErrorInterface {
		return &ValidationError{
			errorType: "InvalidJson",
			message:   fmt.Sprintf("InvalidJson: %s", err.Error()),
			err:       err,
		}
	}
	ErrInvalidValue = func(message string, err error) ValidationErrorInterface {
		if message == "" {
			message = err.Error()
		}
		return &ValidationError{
			errorType: "InvalidValue",
			message:   fmt.Sprintf("InvalidValue: %s", message),
			err:       err,
		}
	}
)

func NewRequest(r *http.Request) *Request {
	webRequest := &Request{
		Request: r,
		route:   getPath(r),
	}
	return webRequest
}

func getPath(r *http.Request) string {
	path := "undefined"
	if r != nil && r.URL != nil && r.URL.Path != "" {
		path = r.URL.Path
	}
	return path
}

func (r *Request) GetRoute() string {
	return r.route
}

func (r *Request) GetPathParams() map[string]string {
	return r.pathParams
}

func (r *Request) GetPathParam(key string) string {
	if value, ok := r.pathParams[key]; ok {
		return value
	}
	return ""
}

func (r *Request) SetPathParam(key, value string) {
	if r.pathParams == nil {
		r.pathParams = make(map[string]string)
	}
	r.route = strings.Replace(r.route, value, fmt.Sprintf(":%s", key), 1)
	r.pathParams[key] = value
}

func (r *Request) QueryParam(key string) string {
	return r.URL.Query().Get(key)
}

func (r *Request) QueryParams() map[string]string {
	if r.params != nil {
		return r.params
	}
	r.params = map[string]string{}
	for key, val := range r.URL.Query() {
		r.params[key] = strings.Join(val, " | ")
	}
	return r.params
}

func (r *Request) QueryParamExists(keys ...string) bool {
	for _, key := range keys {
		if r.URL.Query().Get(key) == "" {
			return false
		}
	}
	return true
}

// Headers shouldn't be used for logging
func (r *Request) Headers() map[string]interface{} {
	headers := map[string]interface{}{}
	for key, value := range r.Header {
		if strings.ToLower(key) == "content-type" {
			continue
		}
		if strings.ToLower(key) == "accept" {
			continue
		}
		headers[key] = value
	}
	return headers
}

// MaskedHeaders returns a request headers with masked values read from an array
func (r *Request) MaskedHeaders() http.Header {
	headers := r.Header.Clone()
	for _, key := range headersToBeMasked {
		k := textproto.CanonicalMIMEHeaderKey(key)
		_, ok := headers[k]
		if ok {
			headers.Set(key, "*******")
		}
	}
	return headers
}

func (r *Request) ReadBody() (map[string]interface{}, error) {
	bodyMap := make(map[string]interface{})

	if r.ContentLength == 0 {
		err := errors.New("empty body")
		return bodyMap, err
	}

	bodyByte, err := io.ReadAll(r.Body)
	if err != nil {
		logger.W(r.Context(), "Error reading request body", zap.String("error", err.Error()))
		return bodyMap, err
	}

	bodyMap, err = unmarshalRequestBody(bodyByte)
	if err != nil {
		logger.W(r.Context(), "Error decoding request json",
			zap.String("error", err.Error()), zap.Any("headers", r.MaskedHeaders()),
			zap.String("url", r.URL.String()), zap.String("body", string(bodyByte)))
	}
	return bodyMap, err
}

func (r *Request) ParseAndValidateBody(s interface{}) error {
	var e error
	if err := r.Bind(s); err != nil {
		e = handleValidationErrors(err)
		return errors.NewWithErr("bad_request", e)

	}

	e = validateStruct(s)
	return errors.NewWithErr("bad_request", e)
}

func validateStruct(s interface{}, structValidations ...validator.StructLevelFunc) ValidationErrorInterface {
	var validate = validator.New()

	_ = validate.RegisterValidation("notblank", validators.NotBlank)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	for _, sValidation := range structValidations {
		validate.RegisterStructValidation(sValidation, s)
	}

	if err := validate.Struct(s); err != nil {
		return handleValidationErrors(err)
	}

	return nil
}

func handleValidationErrors(err error) ValidationErrorInterface {
	switch e := err.(type) {
	case *json.UnmarshalTypeError:
		return ErrInvalidType(e.Field, e.Type, e)
	case validator.ValidationErrors:
		var msgs []string
		for _, fe := range e {
			switch fe.Tag() {
			case "required":
				msgs = append(msgs, fmt.Sprintf("%s is a required field", fe.Field()))
			case "notblank":
				msgs = append(msgs, fmt.Sprintf("%s should not be empty", fe.Field()))
			case "max":
				msgs = append(msgs, fmt.Sprintf("%s must be a maximum of %s in length", fe.Field(), fe.Param()))
			case "url":
				msgs = append(msgs, fmt.Sprintf("%s must be a valid URL", fe.Field()))
			case "uuid":
				msgs = append(msgs, fmt.Sprintf("%s must be a valid uuid", fe.Field()))
			case "date":
				msgs = append(msgs, fmt.Sprintf("%s must be a valid date", fe.Field()))
			default:
				msgs = append(msgs, fmt.Sprintf("validation failed for %s on %s", fe.Field(), fe.Tag()))
			}
		}
		return ErrInvalidValue(strings.Join(msgs, ", "), e)
	case *json.SyntaxError:
		return ErrInvalidJson(e)
	default:
		return &ValidationError{
			errorType: UnexpectedErr,
			message:   e.Error(),
			err:       e,
		}
	}
}

func (r *Request) Bind(v interface{}) error {
	nrtxn := newrelic.FromContext(r.Context())
	defer nrtxn.StartSegment("Request::Bind").End()

	lreader := io.LimitReader(r.Body, 1048576) // 1MB
	body, err := io.ReadAll(lreader)
	if err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewReader(body)) // setup body again so it can be read by any other middleware

	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func (r *Request) GetRequestIP() string {
	fIps := r.Header["X-Forwarded-For"]
	if len(fIps) < 1 {
		if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			return ip
		}

		return net.ParseIP(r.RemoteAddr).String()
	}
	return strings.TrimSpace(strings.Split(fIps[0], ",")[0])
}

func unmarshalRequestBody(body []byte) (map[string]interface{}, error) {
	bodyMap := make(map[string]interface{})
	b := bytes.NewBuffer(body)
	decoder := json.NewDecoder(b)
	decoder.UseNumber()
	err := decoder.Decode(&bodyMap)
	return bodyMap, err
}
