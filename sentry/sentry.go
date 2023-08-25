package sentry

import (
	"errors"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/owlify/sparrow/environment"
)

var env environment.Environment

func Init(e environment.Environment, dsn string) error {
	env = e

	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		EnableTracing:    false,
		TracesSampleRate: DefaultSampleTraceRate(),
		Environment:      string(env),
	})
}

func Close() error {
	flushed := sentry.Flush(2 * time.Second)
	if !flushed {
		return errors.New("unable to flush sentry events")
	}
	return nil
}

func DefaultSampleTraceRate() float64 {
	switch env {
	case environment.ProductionEnv, environment.SandboxEnv:
		return 0.1
	case environment.StagingEnv:
		return 1
	case environment.DevEnv, environment.QAEnv, environment.TestingEnv, environment.UnicornEnv:
		return 0
	default:
		panic("invalid environment to setup sentry")
	}
}
