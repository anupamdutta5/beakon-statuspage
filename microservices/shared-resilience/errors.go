package resilience

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrorCodeConflict            ErrorCode = "CONFLICT"
	ErrorCodeValidation          ErrorCode = "VALIDATION_ERROR"
	ErrorCodeRateLimit           ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrorCodePayloadTooLarge     ErrorCode = "PAYLOAD_TOO_LARGE"
	ErrorCodeUnsupportedMedia    ErrorCode = "UNSUPPORTED_MEDIA_TYPE"

	// Server errors (5xx)
	ErrorCodeInternalServer      ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeTimeout             ErrorCode = "TIMEOUT"
	ErrorCodeDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrorCodeExternalService     ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeCircuitBreakerOpen  ErrorCode = "CIRCUIT_BREAKER_OPEN"

	// Business logic errors
	ErrorCodeBusinessLogic       ErrorCode = "BUSINESS_LOGIC_ERROR"
	ErrorCodeInsufficientFunds   ErrorCode = "INSUFFICIENT_FUNDS"
	ErrorCodeResourceLocked      ErrorCode = "RESOURCE_LOCKED"
	ErrorCodeQuotaExceeded       ErrorCode = "QUOTA_EXCEEDED"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	TenantID   string                 `json:"tenant_id,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Cause
}

// MarshalJSON implements json.Marshaler interface
func (e *AppError) MarshalJSON() ([]byte, error) {
	type Alias AppError
	return json.Marshal(&struct {
		*Alias
		CauseMessage string `json:"cause,omitempty"`
	}{
		Alias:        (*Alias)(e),
		CauseMessage: func() string {
			if e.Cause != nil {
				return e.Cause.Error()
			}
			return ""
		}(),
	})
}

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
	Error       *AppError `json:"error"`
	Success     bool      `json:"success"`
	Timestamp   time.Time `json:"timestamp"`
	RequestID   string    `json:"request_id,omitempty"`
	TraceID     string    `json:"trace_id,omitempty"`
	Path        string    `json:"path,omitempty"`
	Method      string    `json:"method,omitempty"`
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    make(map[string]interface{}),
		Timestamp:  time.Now().UTC(),
	}
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithRequestID adds a request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithUserID adds a user ID to the error
func (e *AppError) WithUserID(userID string) *AppError {
	e.UserID = userID
	return e
}

// WithTenantID adds a tenant ID to the error
func (e *AppError) WithTenantID(tenantID string) *AppError {
	e.TenantID = tenantID
	return e
}

// Predefined errors for common scenarios
var (
	ErrBadRequest = NewAppError(
		ErrorCodeBadRequest,
		"Bad request",
		http.StatusBadRequest,
	)

	ErrUnauthorized = NewAppError(
		ErrorCodeUnauthorized,
		"Unauthorized",
		http.StatusUnauthorized,
	)

	ErrForbidden = NewAppError(
		ErrorCodeForbidden,
		"Forbidden",
		http.StatusForbidden,
	)

	ErrNotFound = NewAppError(
		ErrorCodeNotFound,
		"Resource not found",
		http.StatusNotFound,
	)

	ErrConflict = NewAppError(
		ErrorCodeConflict,
		"Resource conflict",
		http.StatusConflict,
	)

	ErrValidation = NewAppError(
		ErrorCodeValidation,
		"Validation failed",
		http.StatusBadRequest,
	)

	ErrRateLimit = NewAppError(
		ErrorCodeRateLimit,
		"Rate limit exceeded",
		http.StatusTooManyRequests,
	)

	ErrInternalServer = NewAppError(
		ErrorCodeInternalServer,
		"Internal server error",
		http.StatusInternalServerError,
	)

	ErrServiceUnavailable = NewAppError(
		ErrorCodeServiceUnavailable,
		"Service unavailable",
		http.StatusServiceUnavailable,
	)

	ErrTimeout = NewAppError(
		ErrorCodeTimeout,
		"Request timeout",
		http.StatusGatewayTimeout,
	)

	ErrDatabaseError = NewAppError(
		ErrorCodeDatabaseError,
		"Database error",
		http.StatusInternalServerError,
	)

	ErrExternalService = NewAppError(
		ErrorCodeExternalService,
		"External service error",
		http.StatusBadGateway,
	)

	ErrCircuitBreakerOpen = NewAppError(
		ErrorCodeCircuitBreakerOpen,
		"Service temporarily unavailable",
		http.StatusServiceUnavailable,
	)
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger *zap.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle processes an error and returns appropriate HTTP response
func (eh *ErrorHandler) Handle(c *gin.Context, err error) {
	var appErr *AppError

	// Convert to AppError if not already
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		appErr = eh.convertToAppError(err)
	}

	// Enrich error with context information
	eh.enrichError(c, appErr)

	// Log the error
	eh.logError(appErr)

	// Send error response
	eh.sendErrorResponse(c, appErr)
}

// convertToAppError converts a generic error to AppError
func (eh *ErrorHandler) convertToAppError(err error) *AppError {
	// Try to determine error type based on error message or type
	switch {
	case err.Error() == "record not found":
		return ErrNotFound.WithCause(err)
	case err.Error() == "context deadline exceeded":
		return ErrTimeout.WithCause(err)
	default:
		return ErrInternalServer.WithCause(err)
	}
}

// enrichError adds context information to the error
func (eh *ErrorHandler) enrichError(c *gin.Context, appErr *AppError) {
	// Add request ID if available
	if requestID, exists := c.Get("correlation_id"); exists {
		appErr.RequestID = requestID.(string)
	}

	// Add user ID if available
	if userID, exists := c.Get("user_id"); exists {
		appErr.UserID = userID.(string)
	}

	// Add tenant ID if available
	if tenantID, exists := c.Get("tenant_id"); exists {
		appErr.TenantID = tenantID.(string)
	}
}

// logError logs the error with appropriate level
func (eh *ErrorHandler) logError(appErr *AppError) {
	fields := []zap.Field{
		zap.String("error_code", string(appErr.Code)),
		zap.String("message", appErr.Message),
		zap.Int("http_status", appErr.HTTPStatus),
		zap.String("request_id", appErr.RequestID),
		zap.String("user_id", appErr.UserID),
		zap.String("tenant_id", appErr.TenantID),
		zap.Any("details", appErr.Details),
	}

	if appErr.Cause != nil {
		fields = append(fields, zap.Error(appErr.Cause))
	}

	// Log with appropriate level based on HTTP status
	switch {
	case appErr.HTTPStatus >= 500:
		eh.logger.Error("Server error", fields...)
	case appErr.HTTPStatus >= 400:
		eh.logger.Warn("Client error", fields...)
	default:
		eh.logger.Info("Error handled", fields...)
	}
}

// sendErrorResponse sends the error response to the client
func (eh *ErrorHandler) sendErrorResponse(c *gin.Context, appErr *AppError) {
	response := ErrorResponse{
		Error:     appErr,
		Success:   false,
		Timestamp: time.Now().UTC(),
		RequestID: appErr.RequestID,
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}

	// Add trace ID if available (from distributed tracing)
	if traceID, exists := c.Get("trace_id"); exists {
		response.TraceID = traceID.(string)
	}

	c.JSON(appErr.HTTPStatus, response)
}

// ErrorMiddleware is a Gin middleware for centralized error handling
func ErrorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	handler := NewErrorHandler(logger)

	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		var err error
		if e, ok := recovered.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", recovered)
		}

		handler.Handle(c, err)
	})
}

// ValidationError creates a validation error with field details
func ValidationError(field, message string) *AppError {
	return ErrValidation.
		WithDetail("field", field).
		WithDetail("validation_message", message)
}

// ValidationErrors creates a validation error with multiple field errors
func ValidationErrors(fieldErrors map[string]string) *AppError {
	err := ErrValidation.WithDetail("field_errors", fieldErrors)
	err.Message = fmt.Sprintf("Validation failed for %d fields", len(fieldErrors))
	return err
}

// DatabaseError creates a database error
func DatabaseError(operation string, cause error) *AppError {
	return ErrDatabaseError.
		WithCause(cause).
		WithDetail("operation", operation)
}

// ExternalServiceError creates an external service error
func ExternalServiceError(service string, cause error) *AppError {
	return ErrExternalService.
		WithCause(cause).
		WithDetail("service", service)
}

// BusinessLogicError creates a business logic error
func BusinessLogicError(message string) *AppError {
	return NewAppError(
		ErrorCodeBusinessLogic,
		message,
		http.StatusBadRequest,
	)
}

// NotFoundError creates a not found error for a specific resource
func NotFoundError(resource string, id interface{}) *AppError {
	return ErrNotFound.
		WithDetail("resource", resource).
		WithDetail("id", id)
}

// ConflictError creates a conflict error for a specific resource
func ConflictError(resource string, field string, value interface{}) *AppError {
	return ErrConflict.
		WithDetail("resource", resource).
		WithDetail("field", field).
		WithDetail("value", value)
}

// RateLimitError creates a rate limit error
func RateLimitError(limit int, window time.Duration) *AppError {
	return ErrRateLimit.
		WithDetail("limit", limit).
		WithDetail("window", window.String())
}

// UnauthorizedError creates an unauthorized error with reason
func UnauthorizedError(reason string) *AppError {
	return ErrUnauthorized.WithDetail("reason", reason)
}

// ForbiddenError creates a forbidden error with reason
func ForbiddenError(reason string) *AppError {
	return ErrForbidden.WithDetail("reason", reason)
}

// TimeoutError creates a timeout error with operation details
func TimeoutError(operation string, timeout time.Duration) *AppError {
	return ErrTimeout.
		WithDetail("operation", operation).
		WithDetail("timeout", timeout.String())
}

// CircuitBreakerError creates a circuit breaker error
func CircuitBreakerError(service string) *AppError {
	return ErrCircuitBreakerOpen.WithDetail("service", service)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error chain
func GetAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a generic error as an AppError
func WrapError(err error, code ErrorCode, message string, httpStatus int) *AppError {
	return NewAppError(code, message, httpStatus).WithCause(err)
}