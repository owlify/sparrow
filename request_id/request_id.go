package request_id

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

type ctxKey struct{}

const (
	RequestIDHeader = "X-Request-Id"
	RequestIDLogKey = "request_id"
)

func SetRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, reqID)
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(ctxKey{}).(string); ok {
		return reqID
	}
	return ""
}

func GetRequestIDFromRequestHeader(r *http.Request) string {
	requestID := r.Header.Get(RequestIDHeader)
	if requestID == "" {
		requestID = uuid.NewString()
	}
	return requestID
}
