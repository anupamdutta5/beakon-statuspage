// Package services provides custom error types for better error handling.
package services

import (
	"errors"
	"fmt"
)

// Custom error types for service layer
var (
	// ErrNotFound indicates a resource was not found in the database
	ErrNotFound = errors.New("resource not found")
	
	// ErrDuplicateKey indicates a unique constraint violation
	ErrDuplicateKey = errors.New("duplicate key violation")
	
	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input")
	
	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = errors.New("unauthorized")
	
	// ErrForbidden indicates forbidden access
	ErrForbidden = errors.New("forbidden")
	
	// ErrConflict indicates a conflict with existing data
	ErrConflict = errors.New("conflict")
	
	// ErrDatabaseError indicates a database operation error
	ErrDatabaseError = errors.New("database error")
	
	// ErrLimitExceeded indicates a limit has been exceeded
	ErrLimitExceeded = errors.New("limit exceeded")
)

// NotFoundError represents a resource not found error with context
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %v", e.Resource, e.ID)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// DuplicateKeyError represents a duplicate key violation with context
type DuplicateKeyError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate %s.%s: %v", e.Resource, e.Field, e.Value)
}

func (e *DuplicateKeyError) Is(target error) bool {
	return target == ErrDuplicateKey
}

// LimitExceededError represents a limit exceeded error with context
type LimitExceededError struct {
	Resource string
	Limit    int
	Current  int
}

func (e *LimitExceededError) Error() string {
	return fmt.Sprintf("%s limit exceeded: %d/%d", e.Resource, e.Current, e.Limit)
}

func (e *LimitExceededError) Is(target error) bool {
	return target == ErrLimitExceeded
}

// WrapDatabaseError wraps a database error with context
func WrapDatabaseError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
