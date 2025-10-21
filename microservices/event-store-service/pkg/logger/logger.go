// Package logger provides structured logging for the Event Store Service.
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
func (l *Logger) WithService(service string) *zap.Logger {
	return l.Logger.With(zap.String("service", service))
}

// WithRequest adds request information to the logger.
func (l *Logger) WithRequest(requestID string) *zap.Logger {
	return l.Logger.With(zap.String("request_id", requestID))
}

// WithStream adds stream information to the logger.
func (l *Logger) WithStream(streamID string) *zap.Logger {
	return l.Logger.With(zap.String("stream_id", streamID))
}

// WithEvent adds event information to the logger.
func (l *Logger) WithEvent(eventID string) *zap.Logger {
	return l.Logger.With(zap.String("event_id", eventID))
}

// WithProjection adds projection information to the logger.
func (l *Logger) WithProjection(projectionID string) *zap.Logger {
	return l.Logger.With(zap.String("projection_id", projectionID))
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

