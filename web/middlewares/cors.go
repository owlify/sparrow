package middlewares

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func CORS(next httprouter.Handle) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		writer.Header().Set("Access-Control-Max-Age", "3600")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		next(writer, request, params)
	}
}
