package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a structured application error with HTTP context
type AppError struct {
	Code       string      `json:"code"`               // Error code (e.g., "VALIDATION_ERROR")
	Message    string      `json:"message"`            // User-friendly message
	Details    interface{} `json:"details,omitempty"`  // Additional error details
	StatusCode int         `json:"-"`                  // HTTP status code
	Internal   error       `json:"-"`                  // Internal error (not exposed to client)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorResponse represents the JSON error response sent to clients
type ErrorResponse struct {
	Status       string      `json:"status"`             // Always "error"
	Code         string      `json:"code"`               // Error code
	Message      string      `json:"message"`            // User-friendly message
	Details      interface{} `json:"details,omitempty"`  // Additional details
	CorrelationID string     `json:"correlation_id,omitempty"` // Request correlation ID
}

// Domain-specific error codes
const (
	// Validation errors
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_FIELD"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"

	// Resource errors
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeAlreadyExists    = "ALREADY_EXISTS"
	ErrCodeConflict         = "CONFLICT"

	// Authentication/Authorization errors
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeInvalidToken     = "INVALID_TOKEN"
	ErrCodeExpiredToken     = "EXPIRED_TOKEN"
	ErrCodeMissingTenant    = "MISSING_TENANT_CONTEXT"

	// Business logic errors
	ErrCodeMaxUsersExceeded = "MAX_USERS_EXCEEDED"
	ErrCodeInvalidStatus    = "INVALID_STATUS"
	ErrCodeInvalidRole      = "INVALID_ROLE"
	ErrCodeOperationFailed  = "OPERATION_FAILED"

	// Database errors
	ErrCodeDatabaseError    = "DATABASE_ERROR"
	ErrCodeQueryFailed      = "QUERY_FAILED"
	ErrCodeTransactionFailed = "TRANSACTION_FAILED"

	// External service errors
	ErrCodeCacheError       = "CACHE_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"

	// Internal errors
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeUnknownError     = "UNKNOWN_ERROR"
)

// ====================================
// Validation Errors
// ====================================

// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(field string, reason string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidInput,
		Message:    fmt.Sprintf("Invalid input for field '%s'", field),
		Details:    map[string]string{"field": field, "reason": reason},
		StatusCode: http.StatusBadRequest,
	}
}

// NewMissingFieldError creates a missing field error
func NewMissingFieldError(field string) *AppError {
	return &AppError{
		Code:       ErrCodeMissingField,
		Message:    fmt.Sprintf("Required field '%s' is missing", field),
		Details:    map[string]string{"field": field},
		StatusCode: http.StatusBadRequest,
	}
}

// NewInvalidFormatError creates an invalid format error
func NewInvalidFormatError(field string, expectedFormat string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidFormat,
		Message:    fmt.Sprintf("Field '%s' has invalid format", field),
		Details:    map[string]string{"field": field, "expected_format": expectedFormat},
		StatusCode: http.StatusBadRequest,
	}
}

// ====================================
// Resource Errors
// ====================================

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string, identifier string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Details:    map[string]string{"resource": resource, "identifier": identifier},
		StatusCode: http.StatusNotFound,
	}
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource string, identifier string) *AppError {
	return &AppError{
		Code:       ErrCodeAlreadyExists,
		Message:    fmt.Sprintf("%s already exists", resource),
		Details:    map[string]string{"resource": resource, "identifier": identifier},
		StatusCode: http.StatusConflict,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string, details interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusConflict,
	}
}

// ====================================
// Authentication/Authorization Errors
// ====================================

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string, requiredPermission string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		Details:    map[string]string{"required_permission": requiredPermission},
		StatusCode: http.StatusForbidden,
	}
}

// NewInvalidTokenError creates an invalid token error
func NewInvalidTokenError(reason string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidToken,
		Message:    "Invalid authentication token",
		Details:    map[string]string{"reason": reason},
		StatusCode: http.StatusUnauthorized,
	}
}

// NewExpiredTokenError creates an expired token error
func NewExpiredTokenError() *AppError {
	return &AppError{
		Code:       ErrCodeExpiredToken,
		Message:    "Authentication token has expired",
		StatusCode: http.StatusUnauthorized,
	}
}

// NewMissingTenantContextError creates a missing tenant context error
func NewMissingTenantContextError() *AppError {
	return &AppError{
		Code:       ErrCodeMissingTenant,
		Message:    "Tenant context is missing from request",
		StatusCode: http.StatusBadRequest,
	}
}

// ====================================
// Business Logic Errors
// ====================================

// NewMaxUsersExceededError creates a max users exceeded error
func NewMaxUsersExceededError(maxUsers int, currentUsers int) *AppError {
	return &AppError{
		Code:       ErrCodeMaxUsersExceeded,
		Message:    fmt.Sprintf("Maximum user limit exceeded (%d/%d)", currentUsers, maxUsers),
		Details:    map[string]int{"max_users": maxUsers, "current_users": currentUsers},
		StatusCode: http.StatusBadRequest,
	}
}

// NewInvalidStatusError creates an invalid status error
func NewInvalidStatusError(status string, allowedStatuses []string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidStatus,
		Message:    fmt.Sprintf("Invalid status: %s", status),
		Details:    map[string]interface{}{"status": status, "allowed": allowedStatuses},
		StatusCode: http.StatusBadRequest,
	}
}

// NewInvalidRoleError creates an invalid role error
func NewInvalidRoleError(role string, allowedRoles []string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidRole,
		Message:    fmt.Sprintf("Invalid role: %s", role),
		Details:    map[string]interface{}{"role": role, "allowed": allowedRoles},
		StatusCode: http.StatusBadRequest,
	}
}

// NewOperationFailedError creates an operation failed error
func NewOperationFailedError(operation string, reason string) *AppError {
	return &AppError{
		Code:       ErrCodeOperationFailed,
		Message:    fmt.Sprintf("Operation '%s' failed", operation),
		Details:    map[string]string{"operation": operation, "reason": reason},
		StatusCode: http.StatusBadRequest,
	}
}

// ====================================
// Database Errors
// ====================================

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeDatabaseError,
		Message:    "Database operation failed",
		Details:    map[string]string{"operation": operation},
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// NewQueryFailedError creates a query failed error
func NewQueryFailedError(query string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeQueryFailed,
		Message:    "Database query failed",
		Details:    map[string]string{"query": query},
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// NewTransactionFailedError creates a transaction failed error
func NewTransactionFailedError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeTransactionFailed,
		Message:    "Database transaction failed",
		Details:    map[string]string{"operation": operation},
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// ====================================
// External Service Errors
// ====================================

// NewCacheError creates a cache error
func NewCacheError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeCacheError,
		Message:    "Cache operation failed",
		Details:    map[string]string{"operation": operation},
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(service string) *AppError {
	return &AppError{
		Code:       ErrCodeServiceUnavailable,
		Message:    fmt.Sprintf("Service '%s' is currently unavailable", service),
		Details:    map[string]string{"service": service},
		StatusCode: http.StatusServiceUnavailable,
	}
}

// ====================================
// Internal Errors
// ====================================

// NewInternalError creates an internal error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// NewUnknownError creates an unknown error
func NewUnknownError(err error) *AppError {
	return &AppError{
		Code:       ErrCodeUnknownError,
		Message:    "An unexpected error occurred",
		StatusCode: http.StatusInternalServerError,
		Internal:   err,
	}
}

// ====================================
// Helper Functions
// ====================================

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError converts an error to AppError, or creates a new unknown error
func AsAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewUnknownError(err)
}

// ToErrorResponse converts an AppError to an ErrorResponse
func ToErrorResponse(err *AppError, correlationID string) ErrorResponse {
	return ErrorResponse{
		Status:        "error",
		Code:          err.Code,
		Message:       err.Message,
		Details:       err.Details,
		CorrelationID: correlationID,
	}
}
