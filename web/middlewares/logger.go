package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/owlify/sparrow/errors"
	"github.com/owlify/sparrow/logger"
	"github.com/owlify/sparrow/web"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Logger(next httprouter.Handle) httprouter.Handle {
	fn := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		startTime := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: web.StatusOK, body: &bytes.Buffer{}}
		next(lrw, r, params)

		if lrw.statusCode > 299 {
			reqBytes, err := httputil.DumpRequest(r, true)
			if err != nil {
				reqBytes = []byte(`error dumping body ` + err.Error())
			}

			logger.I(r.Context(), "Request details",
				logger.Field("response", lrw.body),
				logger.Field("request", string(reqBytes)),
			)
		}

		if lrw.statusCode > 499 {
			var errMsg string

			resp := make(map[string]interface{})
			err := json.Unmarshal(lrw.body.Bytes(), &resp)

			if err == nil {
				respErr := resp["error"].(map[string]interface{})
				errMsg = respErr["message"].(string)
			} else {
				errMsg = err.Error()
			}

			logger.E(r.Context(), errors.New(errMsg), "Api Request Failed",
				logger.Field("status", lrw.statusCode),
				logger.Field("url", r.URL.String()),
				logger.Field("duration_ms", float64(time.Since(startTime).Nanoseconds())/1e6),
				logger.Field("method", r.Method),
			)
		}

		logger.I(r.Context(), "Request processed",
			logger.Field("status", lrw.statusCode),
			logger.Field("url", r.URL.String()),
			logger.Field("duration_ms", float64(time.Since(startTime).Nanoseconds())/1e6),
			logger.Field("method", r.Method),
		)

	}
	return fn
}
