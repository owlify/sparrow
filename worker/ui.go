package worker

import (
	"fmt"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/julienschmidt/httprouter"
	"github.com/newrelic/go-agent/v3/integrations/nrhttprouter"
)

type UIOpts struct {
	Username string
	Password string
	Endpoint string
	RedisUrl string
}

func SetupUI(router *nrhttprouter.Router, opts *UIOpts) {
	handler := asynqmon.New(asynqmon.Options{
		RootPath:     fmt.Sprintf("/%s", opts.Endpoint),
		RedisConnOpt: asynq.RedisClientOpt{Addr: opts.RedisUrl},
	})

	router.GET(fmt.Sprintf("/%s/*a", opts.Endpoint), asynqmonBasicAuth(handler, opts))
}

func asynqmonBasicAuth(next http.Handler, opts *UIOpts) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != opts.Username || pass != opts.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Authorization required"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
