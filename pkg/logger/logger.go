package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger instance
	Log *zap.Logger
)

// InitLogger initializes the global logger
func InitLogger(env string) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set the log level
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	// Build the logger
	var err error
	Log, err = config.Build()
	if err != nil {
		panic(err)
	}

	// Use the global logger
	zap.ReplaceGlobals(Log)
}

// Sync flushes any buffered log entries
func Sync() {
	_ = Log.Sync()
}

// Info logs an info message with fields
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn logs a warning message with fields
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error logs an error message with fields
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal logs a fatal message with fields and exits
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
