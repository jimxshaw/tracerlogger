package log

import (
	"context"

	log "github.com/jimxshaw/tracerlogger/logger"
	"github.com/jimxshaw/tracerlogger/tracer"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Info logs a message at level Info on the standard logger with trace details.
func Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	log.Info(msg, appendFieldsWithTrace(ctx, fields...)...)
}

// Debug logs a message at level Debug on the standard logger with trace details.
func Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	log.Debug(msg, appendFieldsWithTrace(ctx, fields...)...)
}

// Error logs a message at level Error on the standard logger with trace details.
func Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	log.Error(msg, appendFieldsWithTrace(ctx, fields...)...)
}

// Fatal logs a message at level Fatal on the standard logger with trace details.
func Fatal(ctx context.Context, msg string, fields ...zapcore.Field) {
	log.Fatal(msg, appendFieldsWithTrace(ctx, fields...)...)
}

// Warn logs a message at level Warn on the standard logger with trace details.
func Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	log.Warn(msg, appendFieldsWithTrace(ctx, fields...)...)
}

// traceField extracts the trace details from the context and returns it as a zap field.
func traceField(ctx context.Context) zapcore.Field {
	propagator := tracer.ExtractFromCtx(ctx)
	field := propagator.Sanitize()
	return zap.Any("Trace", field)
}

// appendFieldsWithTrace combines trace details from context with other fields.
func appendFieldsWithTrace(ctx context.Context, fields ...zapcore.Field) []zapcore.Field {
	return append([]zapcore.Field{traceField(ctx)}, fields...)
}
