// Package logger provides structured logging for the Landing Page Service.
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

// WithArticle adds article information to the logger.
func (l *Logger) WithArticle(articleID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("article_id", articleID))
}

// WithFeature adds feature information to the logger.
func (l *Logger) WithFeature(featureID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("feature_id", featureID))
}

// WithTestimonial adds testimonial information to the logger.
func (l *Logger) WithTestimonial(testimonialID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("testimonial_id", testimonialID))
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

