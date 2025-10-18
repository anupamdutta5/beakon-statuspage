package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Authentication errors (1000-1099)
	ErrCodeUnauthorized        ErrorCode = "AUTH_001"
	ErrCodeInvalidCredentials  ErrorCode = "AUTH_002"
	ErrCodeTokenExpired        ErrorCode = "AUTH_003"
	ErrCodeTokenInvalid        ErrorCode = "AUTH_004"
	ErrCodePermissionDenied    ErrorCode = "AUTH_005"
	ErrCodeSessionExpired      ErrorCode = "AUTH_006"

	// Validation errors (1100-1199)
	ErrCodeValidationFailed    ErrorCode = "VAL_001"
	ErrCodeInvalidInput        ErrorCode = "VAL_002"
	ErrCodeMissingField        ErrorCode = "VAL_003"
	ErrCodeInvalidFormat       ErrorCode = "VAL_004"
	ErrCodeOutOfRange          ErrorCode = "VAL_005"

	// Resource errors (1200-1299)
	ErrCodeNotFound            ErrorCode = "RES_001"
	ErrCodeAlreadyExists       ErrorCode = "RES_002"
	ErrCodeConflict            ErrorCode = "RES_003"
	ErrCodeResourceLocked      ErrorCode = "RES_004"
	ErrCodeResourceUnavailable ErrorCode = "RES_005"

	// Database errors (1300-1399)
	ErrCodeDatabaseConnection  ErrorCode = "DB_001"
	ErrCodeDatabaseQuery       ErrorCode = "DB_002"
	ErrCodeDatabaseTransaction ErrorCode = "DB_003"
	ErrCodeDatabaseTimeout     ErrorCode = "DB_004"
	ErrCodeDatabaseConstraint  ErrorCode = "DB_005"

	// Business logic errors (1400-1499)
	ErrCodeBusinessRule        ErrorCode = "BIZ_001"
	ErrCodeQuotaExceeded       ErrorCode = "BIZ_002"
	ErrCodeInvalidState        ErrorCode = "BIZ_003"
	ErrCodeOperationNotAllowed ErrorCode = "BIZ_004"
	ErrCodeDependencyFailed    ErrorCode = "BIZ_005"

	// External service errors (1500-1599)
	ErrCodeExternalService     ErrorCode = "EXT_001"
	ErrCodeServiceTimeout      ErrorCode = "EXT_002"
	ErrCodeServiceUnavailable  ErrorCode = "EXT_003"
	ErrCodeRateLimitExceeded   ErrorCode = "EXT_004"
	ErrCodeCircuitBreakerOpen  ErrorCode = "EXT_005"

	// System errors (1600-1699)
	ErrCodeInternal            ErrorCode = "SYS_001"
	ErrCodeConfiguration       ErrorCode = "SYS_002"
	ErrCodePanic               ErrorCode = "SYS_003"
	ErrCodeTimeout             ErrorCode = "SYS_004"
	ErrCodeUnknown             ErrorCode = "SYS_999"
)

// ErrorSeverity indicates the severity of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// AppError represents a standardized application error
type AppError struct {
	ID            string                 `json:"id"`
	Code          ErrorCode              `json:"code"`
	Message       string                 `json:"message"`
	Details       string                 `json:"details,omitempty"`
	Severity      ErrorSeverity          `json:"severity"`
	HTTPStatus    int                    `json:"-"`
	Internal      error                  `json:"-"`
	Context       map[string]interface{} `json:"context,omitempty"`
	StackTrace    string                 `json:"stack_trace,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Retryable     bool                   `json:"retryable"`
	RetryAfter    *time.Duration         `json:"retry_after,omitempty"`
	Documentation string                 `json:"documentation,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithInternal wraps an internal error
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// WithStackTrace adds stack trace to the error
func (e *AppError) WithStackTrace() *AppError {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	e.StackTrace = string(buf[:n])
	return e
}

// ToJSON converts error to JSON
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		ID:         uuid.New().String(),
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Severity:   getSeverityForCode(code),
		Timestamp:  time.Now(),
		Retryable:  isRetryable(code),
	}
}

// Common error constructors

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

func NewInternalError(message string) *AppError {
	return NewAppError(ErrCodeInternal, message, http.StatusInternalServerError).
		WithSeverity(SeverityHigh)
}

func NewDatabaseError(err error) *AppError {
	return NewAppError(ErrCodeDatabaseQuery, "database operation failed", http.StatusInternalServerError).
		WithInternal(err).
		WithSeverity(SeverityHigh)
}

func NewRateLimitError(retryAfter time.Duration) *AppError {
	return NewAppError(ErrCodeRateLimitExceeded, "rate limit exceeded", http.StatusTooManyRequests).
		WithRetryAfter(retryAfter)
}

func NewBusinessRuleError(message string) *AppError {
	return NewAppError(ErrCodeBusinessRule, message, http.StatusUnprocessableEntity)
}

// WithSeverity sets the severity of the error
func (e *AppError) WithSeverity(severity ErrorSeverity) *AppError {
	e.Severity = severity
	return e
}

// WithRetryAfter sets the retry after duration
func (e *AppError) WithRetryAfter(duration time.Duration) *AppError {
	e.RetryAfter = &duration
	return e
}

// ErrorHandler provides centralized error handling for Gin
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err)
		}
	}
}

// handleError processes and responds with the appropriate error
func handleError(c *gin.Context, ginErr *gin.Error) {
	var appErr *AppError

	// Check if it's already an AppError
	switch e := ginErr.Err.(type) {
	case *AppError:
		appErr = e
	case error:
		// Convert to AppError
		appErr = NewInternalError("an unexpected error occurred").
			WithInternal(e).
			WithStackTrace()
	}

	// Log error based on severity
	logError(appErr)

	// Prepare response
	response := ErrorResponse{
		Error: ErrorDetail{
			ID:            appErr.ID,
			Code:          string(appErr.Code),
			Message:       appErr.Message,
			Details:       appErr.Details,
			Timestamp:     appErr.Timestamp,
			Documentation: getDocumentationURL(appErr.Code),
		},
	}

	// Add retry information if available
	if appErr.Retryable && appErr.RetryAfter != nil {
		c.Header("Retry-After", fmt.Sprintf("%.0f", appErr.RetryAfter.Seconds()))
		response.Error.RetryAfter = appErr.RetryAfter.Seconds()
	}

	// Add context in development mode
	if gin.Mode() == gin.DebugMode {
		response.Error.Context = appErr.Context
		if appErr.StackTrace != "" {
			response.Error.StackTrace = appErr.StackTrace
		}
	}

	c.JSON(appErr.HTTPStatus, response)
}

// ErrorResponse represents the API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	ID            string                 `json:"id"`
	Code          string                 `json:"code"`
	Message       string                 `json:"message"`
	Details       string                 `json:"details,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Documentation string                 `json:"documentation,omitempty"`
	RetryAfter    float64                `json:"retry_after,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	StackTrace    string                 `json:"stack_trace,omitempty"`
}

// Helper functions

func getSeverityForCode(code ErrorCode) ErrorSeverity {
	switch code {
	case ErrCodeUnauthorized, ErrCodeValidationFailed, ErrCodeNotFound:
		return SeverityLow
	case ErrCodeConflict, ErrCodeBusinessRule, ErrCodeRateLimitExceeded:
		return SeverityMedium
	case ErrCodeDatabaseConnection, ErrCodeDatabaseTransaction, ErrCodeExternalService:
		return SeverityHigh
	case ErrCodePanic, ErrCodeInternal:
		return SeverityCritical
	default:
		return SeverityMedium
	}
}

func isRetryable(code ErrorCode) bool {
	switch code {
	case ErrCodeServiceTimeout, ErrCodeServiceUnavailable, ErrCodeRateLimitExceeded,
		ErrCodeDatabaseTimeout, ErrCodeResourceLocked:
		return true
	default:
		return false
	}
}

func getDocumentationURL(code ErrorCode) string {
	return fmt.Sprintf("https://docs.beakon.com/errors/%s", code)
}

func logError(err *AppError) {
	// This should integrate with your logging system
	// For now, just a placeholder
	switch err.Severity {
	case SeverityCritical:
		// Send alert, page on-call
		fmt.Printf("CRITICAL ERROR: %+v\n", err)
	case SeverityHigh:
		// Log error, send to monitoring
		fmt.Printf("HIGH SEVERITY ERROR: %+v\n", err)
	default:
		// Regular logging
		fmt.Printf("ERROR: %+v\n", err)
	}
}

// ErrorRecovery middleware recovers from panics
func ErrorRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Create panic error
				err := NewAppError(ErrCodePanic, "internal server error", http.StatusInternalServerError).
					WithSeverity(SeverityCritical).
					WithStackTrace().
					WithContext("panic", fmt.Sprintf("%v", r))

				// Log the panic
				logError(err)

				// Return error response
				c.JSON(err.HTTPStatus, ErrorResponse{
					Error: ErrorDetail{
						ID:        err.ID,
						Code:      string(err.Code),
						Message:   "An unexpected error occurred. Please try again later.",
						Timestamp: err.Timestamp,
					},
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}

// ValidationErrorHandler handles validation errors
func ValidationErrorHandler(validationErr error) *AppError {
	return NewValidationError("validation failed").
		WithInternal(validationErr).
		WithDetails(validationErr.Error())
}