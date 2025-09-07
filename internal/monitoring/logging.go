package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogConfig represents the logging configuration
type LogConfig struct {
	Level      LogLevel `json:"level"`
	Format     string   `json:"format"` // json, console
	Output     string   `json:"output"` // stdout, stderr, file
	FilePath   string   `json:"file_path,omitempty"`
	MaxSize    int      `json:"max_size,omitempty"` // MB
	MaxBackups int      `json:"max_backups,omitempty"`
	MaxAge     int      `json:"max_age,omitempty"` // days
	Compress   bool     `json:"compress,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
}

// Logger represents a structured logger
type Logger struct {
	serviceName string
	version     string
	environment string
	level       LogLevel
}

// NewLogger creates a new logger
func NewLogger(config LogConfig, serviceName, version, environment string) (*Logger, error) {
	// Set default values
	if config.Level == "" {
		config.Level = LogLevelInfo
	}
	if config.Format == "" {
		config.Format = "json"
	}
	if config.Output == "" {
		config.Output = "stdout"
	}

	return &Logger{
		serviceName: serviceName,
		version:     version,
		environment: environment,
		level:       config.Level,
	}, nil
}

// log writes a log entry
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	// Check if we should log this level
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp:   time.Now(),
		Level:       string(level),
		Service:     l.serviceName,
		Version:     l.version,
		Environment: l.environment,
		Message:     msg,
		Fields:      fields,
	}

	// Simple console output for now
	fmt.Printf("[%s] %s: %s", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
	if len(entry.Fields) > 0 {
		fmt.Printf(" %+v", entry.Fields)
	}
	fmt.Println()
}

// shouldLog checks if we should log at the given level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levels[level] >= levels[l.level]
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	fieldMap := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			fieldMap[k] = v
		}
	}
	l.log(LogLevelDebug, msg, fieldMap)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	fieldMap := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			fieldMap[k] = v
		}
	}
	l.log(LogLevelInfo, msg, fieldMap)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	fieldMap := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			fieldMap[k] = v
		}
	}
	l.log(LogLevelWarn, msg, fieldMap)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	fieldMap := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			fieldMap[k] = v
		}
	}
	l.log(LogLevelError, msg, fieldMap)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	fieldMap := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			fieldMap[k] = v
		}
	}
	l.log(LogLevelFatal, msg, fieldMap)
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{
		serviceName: l.serviceName,
		version:     l.version,
		environment: l.environment,
		level:       l.level,
	}
}

// WithContext creates a new logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := make(map[string]interface{})

	// Add trace ID if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = fmt.Sprintf("%v", traceID)
	}

	// Add span ID if available
	if spanID := ctx.Value("span_id"); spanID != nil {
		fields["span_id"] = fmt.Sprintf("%v", spanID)
	}

	// Add user ID if available
	if userID := ctx.Value("user_id"); userID != nil {
		fields["user_id"] = fmt.Sprintf("%v", userID)
	}

	// Add tenant ID if available
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		fields["tenant_id"] = fmt.Sprintf("%v", tenantID)
	}

	return l.WithFields(fields)
}

// LogHTTPRequest logs an HTTP request
func (l *Logger) LogHTTPRequest(method, path, userAgent string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	l.Info("HTTP request", map[string]interface{}{
		"method":        method,
		"path":          path,
		"user_agent":    userAgent,
		"status_code":   statusCode,
		"duration":      duration.String(),
		"request_size":  requestSize,
		"response_size": responseSize,
	})
}

// LogGRPCRequest logs a gRPC request
func (l *Logger) LogGRPCRequest(method, status string, duration time.Duration) {
	l.Info("gRPC request", map[string]interface{}{
		"method":   method,
		"status":   status,
		"duration": duration.String(),
	})
}

// LogDatabaseQuery logs a database query
func (l *Logger) LogDatabaseQuery(operation, table, status string, duration time.Duration, rowsAffected int64) {
	l.Info("Database query", map[string]interface{}{
		"operation":     operation,
		"table":         table,
		"status":        status,
		"duration":      duration.String(),
		"rows_affected": rowsAffected,
	})
}

// LogBusinessEvent logs a business event
func (l *Logger) LogBusinessEvent(eventType, entityType, entityID string, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"event_type":  eventType,
		"entity_type": entityType,
		"entity_id":   entityID,
	}
	for k, v := range fields {
		allFields[k] = v
	}

	l.Info("Business event", allFields)
}

// LogSecurityEvent logs a security event
func (l *Logger) LogSecurityEvent(eventType, severity string, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"event_type": eventType,
		"severity":   severity,
	}
	for k, v := range fields {
		allFields[k] = v
	}

	l.Warn("Security event", allFields)
}

// LogPerformanceEvent logs a performance event
func (l *Logger) LogPerformanceEvent(operation string, duration time.Duration, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"operation": operation,
		"duration":  duration.String(),
	}
	for k, v := range fields {
		allFields[k] = v
	}

	l.Info("Performance event", allFields)
}

// LogError logs an error with stack trace
func (l *Logger) LogError(err error, msg string, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"error":      err.Error(),
		"error_type": fmt.Sprintf("%T", err),
	}
	for k, v := range fields {
		allFields[k] = v
	}

	// Add stack trace for errors
	if _, file, line, ok := runtime.Caller(1); ok {
		allFields["stack"] = fmt.Sprintf("%s:%d", file, line)
	}

	l.Error(msg, allFields)
}

// LogPanic logs a panic with stack trace
func (l *Logger) LogPanic(panic interface{}, msg string, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"panic":      panic,
		"panic_type": fmt.Sprintf("%T", panic),
	}
	for k, v := range fields {
		allFields[k] = v
	}

	// Add stack trace
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	allFields["stack"] = string(stack[:length])

	l.Fatal(msg, allFields)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return nil
}

// Close closes the logger
func (l *Logger) Close() error {
	return nil
}

// LoggingMiddleware provides HTTP middleware for logging
type LoggingMiddleware struct {
	logger *Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// HTTPMiddleware returns HTTP middleware for logging
func (lm *LoggingMiddleware) HTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture response size and status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Get request size
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}

			// Create context with request ID
			ctx := context.WithValue(r.Context(), "request_id", generateRequestID())
			r = r.WithContext(ctx)

			// Log request start
			lm.logger.Info("HTTP request started", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"query":       r.URL.RawQuery,
				"user_agent":  r.UserAgent(),
				"remote_addr": r.RemoteAddr,
				"request_id":  getRequestID(ctx),
			})

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			lm.logger.LogHTTPRequest(
				r.Method,
				r.URL.Path,
				r.UserAgent(),
				wrapped.statusCode,
				duration,
				requestSize,
				wrapped.responseSize,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture response size and status code
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(n)
	return n, err
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getRequestID gets the request ID from context
func getRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		return fmt.Sprintf("%v", requestID)
	}
	return ""
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config LogConfig, serviceName, version, environment string) error {
	logger, err := NewLogger(config, serviceName, version, environment)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Debug is a convenience function for debug logging
func Debug(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

// Info is a convenience function for info logging
func Info(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

// Warn is a convenience function for warning logging
func Warn(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

// Error is a convenience function for error logging
func Error(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

// Fatal is a convenience function for fatal logging
func Fatal(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, fields...)
	}
}

// LogError is a convenience function for error logging with stack trace
func LogError(err error, msg string, fields map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogError(err, msg, fields)
	}
}

// LogPanic is a convenience function for panic logging
func LogPanic(panic interface{}, msg string, fields map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogPanic(panic, msg, fields)
	}
}
