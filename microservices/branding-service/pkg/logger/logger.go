// Package logger provides structured logging for the Branding Service.
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

// WithBrand adds brand information to the logger.
func (l *Logger) WithBrand(brandID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("brand_id", brandID))
}

// WithTheme adds theme information to the logger.
func (l *Logger) WithTheme(themeID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("theme_id", themeID))
}

// WithAsset adds asset information to the logger.
func (l *Logger) WithAsset(assetID uint) *zap.Logger {
	return l.Logger.With(zap.Uint("asset_id", assetID))
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

