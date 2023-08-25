package http

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/imroc/req/v3"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/owlify/sparrow/request_id"
)

type clientMiddleware func(client *req.Client) *req.Client

// Headers defines the headers for the HTTP request
type Headers map[string]string

type HttpClientOpts struct {
	Timeout time.Duration
	Retries int
}

type InternalAuthHttpClientOpts struct {
	ServiceID  string
	ServiceKey string
	*HttpClientOpts
}

func InternalAuthHTTPClient(opts InternalAuthHttpClientOpts) *req.Client {
	client := req.C().SetCommonRetryCondition(func(resp *req.Response, err error) bool {
		//Retrying only 5XX errors
		if resp != nil && resp.Response != nil {
			return resp.StatusCode >= http.StatusInternalServerError
		}
		return false
	}).SetCommonRetryCount(opts.Retries).SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).SetTimeout(opts.Timeout)

	middlewares := defaultClientMiddlewares()
	middlewares = append(middlewares, internalAuthMiddleware(opts))
	return withMiddlewares(client, middlewares)
}

func RetryableHTTPClient(opts HttpClientOpts) *req.Client {
	client := req.C().SetCommonRetryCondition(func(resp *req.Response, err error) bool {
		//Retrying only 5XX errors
		if resp != nil && resp.Response != nil {
			return resp.StatusCode >= http.StatusInternalServerError
		}
		return false
	}).SetCommonRetryCount(opts.Retries).SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).SetTimeout(opts.Timeout)

	return withMiddlewares(client, defaultClientMiddlewares())
}

func withMiddlewares(client *req.Client, middlewares []clientMiddleware) *req.Client {
	for _, middleware := range middlewares {
		client = middleware(client)
	}
	return client
}

func internalAuthMiddleware(opts InternalAuthHttpClientOpts) clientMiddleware {
	return func(client *req.Client) *req.Client {
		client.OnBeforeRequest(func(client *req.Client, req *req.Request) error {
			req.SetHeaders(headersForInternalRequest(opts.ServiceID, opts.ServiceKey))
			return nil
		})
		return client
	}
}

func defaultClientMiddlewares() []clientMiddleware {
	return []clientMiddleware{
		configureNewrelicTracing,
		configureSentryTracingID,
		configureSetRequestID,
		noticeTimeouts,
	}
}

func noticeTimeouts(client *req.Client) *req.Client {
	client.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			// txn := newrelic.FromContext(req.Context())
			resp, err = rt.RoundTrip(req)
			if err != nil {
				// txn.NoticeError(noticeNewReliClientTimeoutError(req.Context(), err, req.URL.RequestURI()))
				sentry.WithScope(func(scope *sentry.Scope) {
					scope.SetContext("context", map[string]interface{}{
						"url": req.URL.RequestURI(),
					})
				})
				sentry.CaptureException(err)
			}
			return
		}
	})
	return client
}

func configureSetRequestID(client *req.Client) *req.Client {
	client.OnBeforeRequest(func(client *req.Client, req *req.Request) error {
		if requestID := request_id.GetRequestID(req.Context()); requestID != "" {
			req.SetHeaders(map[string]string{request_id.RequestIDHeader: requestID})
		}
		return nil
	})
	return client
}

func configureNewrelicTracing(client *req.Client) *req.Client {
	client.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			txn := newrelic.FromContext(req.Context())
			es := newrelic.StartExternalSegment(txn, &http.Request{
				Method: req.Method,
				URL:    req.URL,
				Header: req.Headers,
			})
			if requestID := request_id.GetRequestID(req.Context()); requestID != "" {
				es.AddAttribute(request_id.RequestIDHeader, requestID)
			}
			defer es.End()

			return rt.RoundTrip(req)
		}
	})
	return client
}

func configureSentryTracingID(client *req.Client) *req.Client {
	client.OnBeforeRequest(func(client *req.Client, req *req.Request) error {
		if txn := sentry.TransactionFromContext(req.Context()); txn != nil {
			req.SetHeader(sentry.SentryTraceHeader, txn.TraceID.String())
		}
		return nil
	})
	return client
}
