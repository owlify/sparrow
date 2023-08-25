package sentry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/owlify/sparrow/environment"
)

func NotifyZap(encoder zapcore.Encoder, message string, fields ...zapcore.Field) {

	switch env {
	case environment.DevEnv, environment.TestingEnv:
		fmt.Println("Can not add sentry hook for dev/test env")
		return
	case environment.StagingEnv, environment.QAEnv, environment.UnicornEnv, environment.SandboxEnv, environment.ProductionEnv:
		var err error
		var req *http.Request

		for _, field := range fields {
			switch field.Interface.(type) {
			case *http.Request:
				req = field.Interface.(*http.Request)
			case error:
				err = field.Interface.(error)
				continue
			default:
				field.AddTo(encoder)
			}
		}

		var data = make(map[string]interface{})

		buffer, ierr := encoder.EncodeEntry(zapcore.Entry{
			Message: message,
			Level:   zapcore.DebugLevel,
		}, []zapcore.Field{
			zap.Time("time", time.Now()),
		})
		if ierr != nil {
			data["meta_message"] = "Unable to encode data for message"
			data["meta_error"] = ierr.Error()
		} else {
			ierr := json.Unmarshal(buffer.Bytes(), &data)
			if ierr != nil {
				data["meta_message"] = "Unable to unmarshall data for message"
				data["meta_error"] = ierr.Error()
			}
		}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetContext("context", data)
			scope.SetRequest(req)
			if err != nil {
				scope.SetFingerprint([]string{err.Error()})
				sentry.CaptureException(err)
			} else {
				scope.SetFingerprint([]string{message})
				sentry.CaptureMessage(message)
			}
		})
	default:
		panic(fmt.Sprintf("unknown env %s", env))
	}
}
