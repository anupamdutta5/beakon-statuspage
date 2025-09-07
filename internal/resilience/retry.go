package resilience

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig represents the configuration for retry logic
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int
	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the multiplier for exponential backoff
	Multiplier float64
	// Jitter adds randomness to the delay
	Jitter bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryPolicy defines when to retry
type RetryPolicy interface {
	ShouldRetry(err error, attempt int) bool
}

// DefaultRetryPolicy is the default retry policy
type DefaultRetryPolicy struct{}

// ShouldRetry returns true if the error should be retried
func (p *DefaultRetryPolicy) ShouldRetry(err error, attempt int) bool {
	// Retry on any error for now
	return err != nil
}

// Retry executes a function with retry logic
func Retry(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	policy := &DefaultRetryPolicy{}
	return RetryWithPolicy(ctx, config, policy, fn)
}

// RetryWithPolicy executes a function with retry logic and custom policy
func RetryWithPolicy(ctx context.Context, config *RetryConfig, policy RetryPolicy, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !policy.ShouldRetry(err, attempt) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay
		delay := calculateDelay(config, attempt)

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithBackoff executes a function with exponential backoff
func RetryWithBackoff(ctx context.Context, config *RetryConfig, fn func() error) error {
	return Retry(ctx, config, fn)
}

// RetryWithLinearBackoff executes a function with linear backoff
func RetryWithLinearBackoff(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	// Override multiplier for linear backoff
	config.Multiplier = 1.0

	return Retry(ctx, config, fn)
}

// RetryWithFixedDelay executes a function with fixed delay between retries
func RetryWithFixedDelay(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	// Override multiplier for fixed delay
	config.Multiplier = 1.0

	return Retry(ctx, config, fn)
}

// calculateDelay calculates the delay for the next retry
func calculateDelay(config *RetryConfig, attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Apply jitter if enabled
	if config.Jitter {
		// Add random jitter (±25%)
		jitter := 0.25
		randomFactor := 1.0 - jitter + (2 * jitter * (float64(time.Now().UnixNano()%1000) / 1000.0))
		delay *= randomFactor
	}

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

// Error implements the error interface
func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// RetryAfterError creates a retryable error with a specific retry after duration
func RetryAfterError(err error, retryAfter time.Duration) *RetryableError {
	return &RetryableError{
		Err:        err,
		RetryAfter: retryAfter,
	}
}

// RetryWithCustomDelay executes a function with custom delay logic
func RetryWithCustomDelay(ctx context.Context, config *RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Check if error has custom retry after
		var delay time.Duration
		if retryableErr, ok := err.(*RetryableError); ok {
			delay = retryableErr.RetryAfter
		} else {
			delay = calculateDelay(config, attempt)
		}

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithTimeout executes a function with retry logic and timeout
func RetryWithTimeout(ctx context.Context, timeout time.Duration, config *RetryConfig, fn func() error) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return Retry(ctx, config, fn)
}

// RetryWithDeadline executes a function with retry logic and deadline
func RetryWithDeadline(ctx context.Context, deadline time.Time, config *RetryConfig, fn func() error) error {
	// Create a context with deadline
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	return Retry(ctx, config, fn)
}

// RetryForever executes a function with retry logic until it succeeds or context is cancelled
func RetryForever(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	attempt := 0
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		// Calculate delay
		delay := calculateDelay(config, attempt)

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		attempt++
	}
}

// RetryWithCondition executes a function with retry logic and custom condition
func RetryWithCondition(ctx context.Context, config *RetryConfig, condition func(error, int) bool, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check custom condition
		if !condition(err, attempt) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay
		delay := calculateDelay(config, attempt)

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
