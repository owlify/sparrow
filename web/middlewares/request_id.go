package middlewares

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/owlify/sparrow/request_id"
)

func RequestID(next httprouter.Handle) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		reqID := request_id.GetRequestIDFromRequestHeader(request)
		ctx := request_id.SetRequestID(request.Context(), reqID)
		writer.Header().Set(request_id.RequestIDHeader, reqID)
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		writer.Header().Set("Access-Control-Max-Age", "3600")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		next(writer, request.WithContext(ctx), params)
	}
}
