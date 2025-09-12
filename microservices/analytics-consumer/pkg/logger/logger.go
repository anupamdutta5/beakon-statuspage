// Package logger provides structured logging for the Analytics Consumer.
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

// WithConsumer adds consumer information to the logger.
func (l *Logger) WithConsumer(consumerType, queueName string) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			zap.String("consumer_type", consumerType),
			zap.String("queue_name", queueName),
		),
	}
}

// WithMessage adds message information to the logger.
func (l *Logger) WithMessage(messageID, eventType string) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			zap.String("message_id", messageID),
			zap.String("event_type", eventType),
		),
	}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

