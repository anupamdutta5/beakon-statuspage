package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
	TraceID string            `json:"trace_id,omitempty"`
}

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Authentication and Authorization
	ErrorCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrorCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrorCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"

	// Validation
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingField     ErrorCode = "MISSING_FIELD"

	// Resource Management
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists   ErrorCode = "ALREADY_EXISTS"
	ErrorCodeConflict        ErrorCode = "CONFLICT"
	ErrorCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Server Errors
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrorCodeExternalService    ErrorCode = "EXTERNAL_SERVICE_ERROR"

	// Business Logic
	ErrorCodeBusinessRule    ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrorCodeQuotaExceeded   ErrorCode = "QUOTA_EXCEEDED"
	ErrorCodeFeatureDisabled ErrorCode = "FEATURE_DISABLED"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode
	Message    string
	Details    map[string]string
	StatusCode int
	Err        error
	Stack      string
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, statusCode int, err error) *AppError {
	stack := ""
	if _, file, line, ok := runtime.Caller(1); ok {
		stack = fmt.Sprintf("%s:%d", file, line)
	}

	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Stack:      stack,
		Details:    make(map[string]string), // Initialize Details map
	}
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	debugMode bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(debugMode bool) *ErrorHandler {
	return &ErrorHandler{
		debugMode: debugMode,
	}
}

// HandleError handles errors and returns appropriate HTTP responses
func (eh *ErrorHandler) HandleError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			eh.processError(c, err.Err)
		}
	}
}

// processError processes an error and returns appropriate response
func (eh *ErrorHandler) processError(c *gin.Context, err error) {
	var appErr *AppError
	var statusCode int
	var errorCode ErrorCode
	var message string
	var details map[string]string

	// Check if it's already an AppError
	if errors.As(err, &appErr) {
		statusCode = appErr.StatusCode
		errorCode = appErr.Code
		message = appErr.Message
		details = appErr.Details
	} else {
		// Convert generic errors to AppError
		appErr = eh.convertToAppError(err)
		statusCode = appErr.StatusCode
		errorCode = appErr.Code
		message = appErr.Message
		details = appErr.Details
	}

	// Log the error
	eh.logError(c, appErr)

	// Create error response
	response := ErrorResponse{
		Error:   message,
		Code:    string(errorCode),
		Details: details,
	}

	// Add trace ID if available
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		response.TraceID = traceID
	}

	// Add stack trace in debug mode
	if eh.debugMode && appErr.Stack != "" {
		if response.Details == nil {
			response.Details = make(map[string]string)
		}
		response.Details["stack"] = appErr.Stack
	}

	c.JSON(statusCode, response)
}

// convertToAppError converts generic errors to AppError
func (eh *ErrorHandler) convertToAppError(err error) *AppError {
	// Handle GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewAppError(ErrorCodeNotFound, "Resource not found", http.StatusNotFound, err)
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return NewAppError(ErrorCodeAlreadyExists, "Resource already exists", http.StatusConflict, err)
	}

	// Handle validation errors
	if strings.Contains(err.Error(), "validation") {
		return NewAppError(ErrorCodeValidationFailed, "Validation failed", http.StatusBadRequest, err)
	}

	// Handle authentication errors
	if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "authentication") {
		return NewAppError(ErrorCodeUnauthorized, "Authentication required", http.StatusUnauthorized, err)
	}

	// Handle authorization errors
	if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "permission") {
		return NewAppError(ErrorCodeForbidden, "Insufficient permissions", http.StatusForbidden, err)
	}

	// Handle rate limiting errors
	if strings.Contains(err.Error(), "rate limit") || strings.Contains(err.Error(), "too many requests") {
		return NewAppError(ErrorCodeTooManyRequests, "Too many requests", http.StatusTooManyRequests, err)
	}

	// Default to internal server error
	return NewAppError(ErrorCodeInternalError, "Internal server error", http.StatusInternalServerError, err)
}

// logError logs the error with appropriate level
func (eh *ErrorHandler) logError(c *gin.Context, appErr *AppError) {
	fields := []zap.Field{
		zap.String("error_code", string(appErr.Code)),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
	}

	if appErr.Stack != "" {
		fields = append(fields, zap.String("stack", appErr.Stack))
	}

	if appErr.Err != nil {
		fields = append(fields, zap.Error(appErr.Err))
	}

	// Log based on status code
	switch {
	case appErr.StatusCode >= 500:
		logger.Error("Server error occurred", fields...)
	case appErr.StatusCode >= 400:
		logger.Warn("Client error occurred", fields...)
	default:
		logger.Info("Request completed with error", fields...)
	}
}

// Recovery middleware for panic recovery
func (eh *ErrorHandler) Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		stack := ""
		if _, file, line, ok := runtime.Caller(3); ok {
			stack = fmt.Sprintf("%s:%d", file, line)
		}

		appErr := NewAppError(
			ErrorCodeInternalError,
			"Internal server error",
			http.StatusInternalServerError,
			fmt.Errorf("panic: %v", recovered),
		)
		appErr.Stack = stack

		eh.logError(c, appErr)

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  string(ErrorCodeInternalError),
		})
	})
}

// Helper functions for creating common errors
func NewValidationError(field, message string) *AppError {
	return &AppError{
		Code:       ErrorCodeValidationFailed,
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
		Details:    map[string]string{field: message},
	}
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(
		ErrorCodeNotFound,
		fmt.Sprintf("%s not found", resource),
		http.StatusNotFound,
		nil,
	)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(
		ErrorCodeUnauthorized,
		message,
		http.StatusUnauthorized,
		nil,
	)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(
		ErrorCodeForbidden,
		message,
		http.StatusForbidden,
		nil,
	)
}

func NewInternalError(message string, err error) *AppError {
	return NewAppError(
		ErrorCodeInternalError,
		message,
		http.StatusInternalServerError,
		err,
	)
}

func NewDatabaseError(operation string, err error) *AppError {
	return NewAppError(
		ErrorCodeDatabaseError,
		fmt.Sprintf("Database %s failed", operation),
		http.StatusInternalServerError,
		err,
	)
}

func NewExternalServiceError(service string, err error) *AppError {
	return NewAppError(
		ErrorCodeExternalService,
		fmt.Sprintf("External service %s error", service),
		http.StatusBadGateway,
		err,
	)
}
