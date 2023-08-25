package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/owlify/sparrow/environment"
	"github.com/owlify/sparrow/request_id"
	"github.com/owlify/sparrow/sentry"
)

var logger *zap.Logger
var jsonEncoder zapcore.Encoder

type LoggingFunc func(message string, fields ...zapcore.Field)

func Init(mode int, env environment.Environment) {
	var logLevel zapcore.Level
	switch mode {
	case DEBUG:
		logLevel = zapcore.DebugLevel
	case INFO:
		logLevel = zapcore.InfoLevel
	case WARNING:
		logLevel = zapcore.WarnLevel
	case ERROR:
		logLevel = zapcore.ErrorLevel
	case FATAL:
		logLevel = zapcore.FatalLevel
	}

	cfg := zap.Config{
		Encoding: "json",
		Level:    zap.NewAtomicLevelAt(logLevel),
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		},
	}

	logger, _ = cfg.Build()
	jsonEncoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)

	if env == environment.DevEnv {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger = logger.WithOptions(
			zap.WrapCore(
				func(zapcore.Core) zapcore.Core {
					return zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.AddSync(os.Stderr), zapcore.DebugLevel)
				}))
	} else {
		logger = logger.WithOptions(
			zap.WrapCore(
				func(zapcore.Core) zapcore.Core {
					return zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stderr), logLevel)
				}))
	}
}

func Get() *zap.Logger {
	return logger
}

func Field(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}

func addKeysFromContext(ctx context.Context, fields ...zapcore.Field) []zapcore.Field {
	if requestID := request_id.GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String(request_id.RequestIDLogKey, requestID))
	}

	return fields
}

func I(ctx context.Context, message string, fields ...zapcore.Field) {
	logger.Info(message, addKeysFromContext(ctx, fields...)...)
}

func D(ctx context.Context, message string, fields ...zapcore.Field) {
	logger.Debug(message, addKeysFromContext(ctx, fields...)...)
}

func W(ctx context.Context, message string, fields ...zapcore.Field) {
	logger.Warn(message, addKeysFromContext(ctx, fields...)...)
}

func E(ctx context.Context, err error, message string, fields ...zapcore.Field) {
	fields = append(fields, Field("error", err))
	fieldsWithRequestID := addKeysFromContext(ctx, fields...)
	logger.Error(message, fieldsWithRequestID...)
	sentry.NotifyZap(jsonEncoder.Clone(), message, fieldsWithRequestID...)
}

func Sync() {
	logger.Info("SYNCING LOGGER....")
	err := logger.Sync()
	if err != nil {
		fmt.Println("FAILED TO SYNC LOGGER...")
	}
}
