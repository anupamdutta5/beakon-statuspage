package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Manager handles graceful shutdown of services
type Manager struct {
	logger      *zap.Logger
	shutdownFns []ShutdownFunc
	timeout     time.Duration
	mu          sync.Mutex
	shutdown    chan os.Signal
	done        chan struct{}
}

// ShutdownFunc represents a function that should be called during shutdown
type ShutdownFunc func(ctx context.Context) error

// Config represents shutdown configuration
type Config struct {
	Timeout time.Duration `yaml:"timeout" env:"SHUTDOWN_TIMEOUT" default:"30s"`
}

// NewManager creates a new graceful shutdown manager
func NewManager(logger *zap.Logger, config Config) *Manager {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return &Manager{
		logger:      logger,
		shutdownFns: make([]ShutdownFunc, 0),
		timeout:     config.Timeout,
		shutdown:    shutdown,
		done:        make(chan struct{}),
	}
}

// RegisterShutdownHandler registers a function to be called during shutdown
func (m *Manager) RegisterShutdownHandler(name string, fn ShutdownFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wrappedFn := func(ctx context.Context) error {
		m.logger.Info("Executing shutdown handler", zap.String("handler", name))

		start := time.Now()
		err := fn(ctx)
		duration := time.Since(start)

		if err != nil {
			m.logger.Error("Shutdown handler failed",
				zap.String("handler", name),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			m.logger.Info("Shutdown handler completed successfully",
				zap.String("handler", name),
				zap.Duration("duration", duration))
		}

		return err
	}

	m.shutdownFns = append(m.shutdownFns, wrappedFn)
}

// WaitForShutdown waits for a shutdown signal and executes all registered handlers
func (m *Manager) WaitForShutdown() error {
	// Wait for shutdown signal
	sig := <-m.shutdown
	m.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	return m.executeShutdown()
}

// WaitForShutdownWithContext waits for shutdown signal or context cancellation
func (m *Manager) WaitForShutdownWithContext(ctx context.Context) error {
	select {
	case sig := <-m.shutdown:
		m.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		return m.executeShutdown()
	case <-ctx.Done():
		m.logger.Info("Context cancelled, initiating shutdown")
		return m.executeShutdown()
	}
}

// Shutdown initiates graceful shutdown without waiting for signal
func (m *Manager) Shutdown() error {
	return m.executeShutdown()
}

// executeShutdown executes all registered shutdown handlers
func (m *Manager) executeShutdown() error {
	m.logger.Info("Starting graceful shutdown",
		zap.Int("handlers", len(m.shutdownFns)),
		zap.Duration("timeout", m.timeout))

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	// Execute all shutdown handlers
	var wg sync.WaitGroup
	errCh := make(chan error, len(m.shutdownFns))

	for i, fn := range m.shutdownFns {
		wg.Add(1)
		go func(index int, shutdownFn ShutdownFunc) {
			defer wg.Done()
			if err := shutdownFn(ctx); err != nil {
				errCh <- fmt.Errorf("shutdown handler %d failed: %w", index, err)
			}
		}(i, fn)
	}

	// Wait for all handlers to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("All shutdown handlers completed successfully")
		close(m.done)
		return nil
	case <-ctx.Done():
		m.logger.Error("Shutdown timeout exceeded", zap.Duration("timeout", m.timeout))

		// Collect any errors that occurred
		close(errCh)
		var errors []error
		for err := range errCh {
			errors = append(errors, err)
		}

		close(m.done)

		if len(errors) > 0 {
			return fmt.Errorf("shutdown timeout with errors: %v", errors)
		}

		return fmt.Errorf("shutdown timeout")
	}
}

// Done returns a channel that is closed when shutdown is complete
func (m *Manager) Done() <-chan struct{} {
	return m.done
}

// SetTimeout updates the shutdown timeout
func (m *Manager) SetTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeout = timeout
}

// GetTimeout returns the current shutdown timeout
func (m *Manager) GetTimeout() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.timeout
}

// AddShutdownHandlers is a convenience method to add multiple handlers at once
func (m *Manager) AddShutdownHandlers(handlers map[string]ShutdownFunc) {
	for name, fn := range handlers {
		m.RegisterShutdownHandler(name, fn)
	}
}

// HTTPServerShutdownHandler returns a shutdown handler for http.Server
func HTTPServerShutdownHandler(server interface {
	Shutdown(ctx context.Context) error
}) ShutdownFunc {
	return func(ctx context.Context) error {
		return server.Shutdown(ctx)
	}
}

// DatabaseShutdownHandler returns a shutdown handler for database connections
func DatabaseShutdownHandler(closer interface {
	Close() error
}) ShutdownFunc {
	return func(ctx context.Context) error {
		return closer.Close()
	}
}

// ChannelShutdownHandler returns a shutdown handler that closes a channel
func ChannelShutdownHandler(ch chan struct{}) ShutdownFunc {
	return func(ctx context.Context) error {
		close(ch)
		return nil
	}
}

// FuncShutdownHandler wraps a simple function as a shutdown handler
func FuncShutdownHandler(fn func() error) ShutdownFunc {
	return func(ctx context.Context) error {
		return fn()
	}
}

// TimeoutShutdownHandler wraps another handler with a specific timeout
func TimeoutShutdownHandler(timeout time.Duration, handler ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- handler(timeoutCtx)
		}()

		select {
		case err := <-done:
			return err
		case <-timeoutCtx.Done():
			return fmt.Errorf("handler timeout after %v", timeout)
		}
	}
}
