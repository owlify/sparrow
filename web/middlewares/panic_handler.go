package middlewares

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"

	"github.com/owlify/sparrow/errors"
	"github.com/owlify/sparrow/logger"
	"github.com/owlify/sparrow/utils"
)

func PanicHandler(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {

		defer func() {
			if rv := recover(); rv != nil {
				errorMessage := utils.ConvertToString(rv)
				logger.E(req.Context(), errors.New(errorMessage), "Request Panic",
					logger.Field("panic", rv),
					logger.Field("stack", string(debug.Stack())),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				err := json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]string{
						"code":    "panic",
						"message": errorMessage,
					},
				})
				if err != nil {
					logger.E(req.Context(), err, "Unable to encode panic response")
				}
			}
		}()

		next(w, req, params)
	}
}
