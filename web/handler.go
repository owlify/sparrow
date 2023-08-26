package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handle func(Endpoint, ...Middleware) httprouter.Handle

func Serve(handle Endpoint, middlewares ...Middleware) httprouter.Handle {
	handler := func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		webReq := NewRequest(req)
		for i := range ps {
			webReq.SetPathParam(ps[i].Key, ps[i].Value)
		}

		response := handle(webReq)
		WriteJsonResponse(w, response)
	}

	return execMiddlewares(handler, middlewares...)
}

func execMiddlewares(handle httprouter.Handle, middlewares ...Middleware) httprouter.Handle {
	if len(middlewares) == 0 {
		return handle
	}

	// executing middlewares in reverse order to preserve the same order
	for i := len(middlewares) - 1; i > -1; i-- {
		handle = middlewares[i](handle)
	}

	return handle
}
