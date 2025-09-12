// Package logger provides structured logging for the SaaS Admin Service.
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

// WithPlan adds plan information to the logger.
func (l *Logger) WithPlan(planID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("plan_id", planID))
}

// WithFeature adds feature information to the logger.
func (l *Logger) WithFeature(featureID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("feature_id", featureID))
}

// WithFeatureFlag adds feature flag information to the logger.
func (l *Logger) WithFeatureFlag(flagID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("feature_flag_id", flagID))
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

