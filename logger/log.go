package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

var DefaultConfig = zap.Config{
	Encoding:         "console",
	Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
	EncoderConfig: zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		TimeKey:      "time",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	},
}

func InitializeLogger(cfg zap.Config) {
	var err error
	logger, err = cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

func init() {
	InitializeLogger(DefaultConfig)
}

// Debug logs a message at level Debug on the standard logger.
// log.Debug("This is a DEBUG message")
func Debug(msg string, fields ...zapcore.Field) {
	logger.Debug(msg, fields...)
}

// Info logs a message at level Info on the standard logger.
// log.Info("This is an INFO message")
// log.Info("This is an INFO message with fields", zap.String("region", "APAC"), zap.Int("id", 1))
func Info(msg string, fields ...zapcore.Field) {
	logger.Info(msg, fields...)
}

// Warn logs a message at level Warn on the standard logger.
// log.Warn("This is a Warn message")
func Warn(msg string, fields ...zapcore.Field) {
	logger.Warn(msg, fields...)
}

// Error logs a message at level Error on the standard logger.
// log.Error("This is an ERROR message")
func Error(msg string, fields ...zapcore.Field) {
	logger.Error(msg, fields...)
}

// Fatal logs a message at level Fatal on the standard logger.
// After logging, it will call os.Exit(1).
// log.Fatal("This is a FATAL message")
func Fatal(msg string, fields ...zapcore.Field) {
	logger.Fatal(msg, fields...)
}

// Cleanup flushes all log entries.
// The reason is certain loggers might not write each log message to its destination immediately upon receiving it.
// The logger could accumulate several log messages in memory and then write them out in a single batch.
// So call Cleanup before the application exits to make sure all buffered logs are properly written.
// E.g. defer log.Cleanup()
func Cleanup() {
	if err := logger.Sync(); err != nil {
		panic(fmt.Sprintf("Failed to sync logger: %v", err))
	}
}
