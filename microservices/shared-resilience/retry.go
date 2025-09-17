package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithContext represents a function that can be retried with context
type RetryableFuncWithContext func(ctx context.Context) error

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries      int           `yaml:"max_retries" env:"RETRY_MAX_RETRIES" default:"3"`
	InitialDelay    time.Duration `yaml:"initial_delay" env:"RETRY_INITIAL_DELAY" default:"100ms"`
	MaxDelay        time.Duration `yaml:"max_delay" env:"RETRY_MAX_DELAY" default:"30s"`
	BackoffFactor   float64       `yaml:"backoff_factor" env:"RETRY_BACKOFF_FACTOR" default:"2.0"`
	Jitter          bool          `yaml:"jitter" env:"RETRY_JITTER" default:"true"`
	RetryableErrors []string      `yaml:"retryable_errors"`
}

// RetryManager manages retry operations
type RetryManager struct {
	config RetryConfig
	logger *zap.Logger
}

// RetryResult represents the result of a retry operation
type RetryResult struct {
	Success     bool          `json:"success"`
	Attempts    int           `json:"attempts"`
	TotalTime   time.Duration `json:"total_time"`
	LastError   string        `json:"last_error,omitempty"`
	FirstError  string        `json:"first_error,omitempty"`
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config RetryConfig, logger *zap.Logger) *RetryManager {
	// Set defaults
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.InitialDelay == 0 {
		config.InitialDelay = 100 * time.Millisecond
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.BackoffFactor == 0 {
		config.BackoffFactor = 2.0
	}

	return &RetryManager{
		config: config,
		logger: logger,
	}
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RetryManager) ExecuteWithRetry(fn RetryableFunc) error {
	return rm.ExecuteWithRetryContext(context.Background(), func(ctx context.Context) error {
		return fn()
	})
}

// ExecuteWithRetryContext executes a function with retry logic and context
func (rm *RetryManager) ExecuteWithRetryContext(ctx context.Context, fn RetryableFuncWithContext) error {
	var firstError error
	var lastError error

	for attempt := 1; attempt <= rm.config.MaxRetries+1; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		err := fn(ctx)
		if err == nil {
			if attempt > 1 {
				rm.logger.Info("Operation succeeded after retries",
					zap.Int("attempts", attempt),
					zap.String("first_error", firstError.Error()))
			}
			return nil
		}

		if firstError == nil {
			firstError = err
		}
		lastError = err

		if attempt > rm.config.MaxRetries {
			rm.logger.Error("Operation failed after all retries",
				zap.Int("attempts", attempt),
				zap.String("first_error", firstError.Error()),
				zap.String("last_error", lastError.Error()))
			return fmt.Errorf("operation failed after %d attempts, first error: %w, last error: %v",
				attempt, firstError, lastError)
		}

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			rm.logger.Warn("Non-retryable error encountered",
				zap.Int("attempt", attempt),
				zap.Error(err))
			return fmt.Errorf("non-retryable error: %w", err)
		}

		delay := rm.calculateDelay(attempt - 1) // attempt-1 because we start from 1
		rm.logger.Warn("Operation failed, retrying",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", rm.config.MaxRetries+1),
			zap.Duration("delay", delay),
			zap.Error(err))

		// Wait for the calculated delay
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", rm.config.MaxRetries+1, lastError)
}

// ExecuteWithRetryAndResult executes a function with retry logic and returns detailed results
func (rm *RetryManager) ExecuteWithRetryAndResult(ctx context.Context, fn RetryableFuncWithContext) (RetryResult, error) {
	start := time.Now()
	var firstError error
	var lastError error

	for attempt := 1; attempt <= rm.config.MaxRetries+1; attempt++ {
		select {
		case <-ctx.Done():
			return RetryResult{
				Success:    false,
				Attempts:   attempt - 1,
				TotalTime:  time.Since(start),
				LastError:  ctx.Err().Error(),
				FirstError: getErrorString(firstError),
			}, fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		err := fn(ctx)
		if err == nil {
			return RetryResult{
				Success:    true,
				Attempts:   attempt,
				TotalTime:  time.Since(start),
				FirstError: getErrorString(firstError),
			}, nil
		}

		if firstError == nil {
			firstError = err
		}
		lastError = err

		if attempt > rm.config.MaxRetries {
			return RetryResult{
				Success:    false,
				Attempts:   attempt,
				TotalTime:  time.Since(start),
				LastError:  lastError.Error(),
				FirstError: firstError.Error(),
			}, fmt.Errorf("operation failed after %d attempts: %w", attempt, lastError)
		}

		if !rm.isRetryableError(err) {
			return RetryResult{
				Success:    false,
				Attempts:   attempt,
				TotalTime:  time.Since(start),
				LastError:  err.Error(),
				FirstError: firstError.Error(),
			}, fmt.Errorf("non-retryable error: %w", err)
		}

		delay := rm.calculateDelay(attempt - 1)
		rm.logger.Warn("Operation failed, retrying",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err))

		select {
		case <-ctx.Done():
			return RetryResult{
				Success:    false,
				Attempts:   attempt,
				TotalTime:  time.Since(start),
				LastError:  ctx.Err().Error(),
				FirstError: firstError.Error(),
			}, fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return RetryResult{
		Success:    false,
		Attempts:   rm.config.MaxRetries + 1,
		TotalTime:  time.Since(start),
		LastError:  lastError.Error(),
		FirstError: firstError.Error(),
	}, fmt.Errorf("operation failed after all retries: %w", lastError)
}

// calculateDelay calculates the delay for the next retry with exponential backoff
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt)))

	// Apply maximum delay limit
	if delay > rm.config.MaxDelay {
		delay = rm.config.MaxDelay
	}

	// Add jitter if enabled
	if rm.config.Jitter {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1) // Up to 10% jitter
		delay += jitter
	}

	return delay
}

// isRetryableError checks if an error is retryable
func (rm *RetryManager) isRetryableError(err error) bool {
	if len(rm.config.RetryableErrors) == 0 {
		// If no specific retryable errors configured, assume all errors are retryable
		return true
	}

	errorStr := err.Error()
	for _, retryableError := range rm.config.RetryableErrors {
		if contains(errorStr, retryableError) {
			return true
		}
	}

	return false
}

// getErrorString safely gets error string
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(substr) > 0 &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     indexOf(s, substr) >= 0))
}

// indexOf finds the index of a substring in a string
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Common retry configurations

// DefaultRetryConfig provides a sensible default retry configuration
var DefaultRetryConfig = RetryConfig{
	MaxRetries:    3,
	InitialDelay:  100 * time.Millisecond,
	MaxDelay:      30 * time.Second,
	BackoffFactor: 2.0,
	Jitter:        true,
}

// QuickRetryConfig for fast operations that should retry quickly
var QuickRetryConfig = RetryConfig{
	MaxRetries:    5,
	InitialDelay:  50 * time.Millisecond,
	MaxDelay:      1 * time.Second,
	BackoffFactor: 1.5,
	Jitter:        true,
}

// SlowRetryConfig for slow operations that need longer delays
var SlowRetryConfig = RetryConfig{
	MaxRetries:    3,
	InitialDelay:  1 * time.Second,
	MaxDelay:      60 * time.Second,
	BackoffFactor: 2.0,
	Jitter:        true,
}

// NewDefaultRetryManager creates a retry manager with default configuration
func NewDefaultRetryManager(logger *zap.Logger) *RetryManager {
	return NewRetryManager(DefaultRetryConfig, logger)
}

// NewQuickRetryManager creates a retry manager optimized for quick operations
func NewQuickRetryManager(logger *zap.Logger) *RetryManager {
	return NewRetryManager(QuickRetryConfig, logger)
}

// NewSlowRetryManager creates a retry manager optimized for slow operations
func NewSlowRetryManager(logger *zap.Logger) *RetryManager {
	return NewRetryManager(SlowRetryConfig, logger)
}