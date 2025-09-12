// Package logger provides structured logging for the Incident Service.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger represents a structured logger.
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance.
func New(environment string) (*Logger, error) {
	var config zap.Config

	if environment == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// Configure encoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: logger}, nil
}

// WithService adds service information to the logger.
func (l *Logger) WithService(name, version string) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			zap.String("service", name),
			zap.String("version", version),
		),
	}
}

// WithRequest adds request information to the logger.
func (l *Logger) WithRequest(method, path, userID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			zap.String("method", method),
			zap.String("path", path),
			zap.String("user_id", userID),
		),
	}
}

// WithTenant adds tenant information to the logger.
func (l *Logger) WithTenant(tenantID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("tenant_id", tenantID)),
	}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
