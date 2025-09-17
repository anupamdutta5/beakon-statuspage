#!/bin/bash

# Script to implement comprehensive graceful shutdown patterns

echo "Implementing graceful shutdown patterns..."

# Create graceful shutdown manager for each service
create_shutdown_manager() {
    local service_path="$1"
    local service_name="$2"

    # Create internal/shutdown directory if it doesn't exist
    mkdir -p "$service_path/internal/shutdown"

    # Create graceful shutdown manager
    cat > "$service_path/internal/shutdown/manager.go" << 'EOF'
package shutdown

import (
	"context"
	"errors"
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

		return errors.New("shutdown timeout")
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
EOF

    echo "Created graceful shutdown manager for $service_name"
}

# Create example integration for services
create_shutdown_integration_example() {
    local service_path="$1"
    local service_name="$2"

    cat > "$service_path/internal/shutdown/example_integration.go" << 'EOF'
package shutdown

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ExampleServiceIntegration shows how to integrate graceful shutdown
// This is an example - adapt for your specific service needs
func ExampleServiceIntegration(logger *zap.Logger) {
	// Create shutdown manager
	shutdownManager := NewManager(logger, Config{
		Timeout: 30 * time.Second,
	})

	// Example: HTTP Server shutdown
	var httpServer *http.Server // Your HTTP server instance
	if httpServer != nil {
		shutdownManager.RegisterShutdownHandler("http-server",
			HTTPServerShutdownHandler(httpServer))
	}

	// Example: Database shutdown
	var dbManager interface{ Close() error } // Your database manager
	if dbManager != nil {
		shutdownManager.RegisterShutdownHandler("database",
			DatabaseShutdownHandler(dbManager))
	}

	// Example: Custom cleanup
	shutdownManager.RegisterShutdownHandler("custom-cleanup", func(ctx context.Context) error {
		logger.Info("Performing custom cleanup...")
		// Add your custom cleanup logic here
		return nil
	})

	// Example: Background worker shutdown
	workerDone := make(chan struct{})
	shutdownManager.RegisterShutdownHandler("background-worker",
		ChannelShutdownHandler(workerDone))

	// Example: Cache cleanup
	shutdownManager.RegisterShutdownHandler("cache-cleanup",
		FuncShutdownHandler(func() error {
			// Clean up cache, temporary files, etc.
			return nil
		}))

	// Wait for shutdown signal and execute all handlers
	if err := shutdownManager.WaitForShutdown(); err != nil {
		logger.Error("Graceful shutdown failed", zap.Error(err))
	} else {
		logger.Info("Graceful shutdown completed successfully")
	}
}

// IntegrateWithExistingService shows how to add to existing service
func IntegrateWithExistingService() {
	// In your main.go or server initialization:

	// 1. Create shutdown manager
	// shutdownManager := NewManager(logger, Config{Timeout: 30 * time.Second})

	// 2. Register all your shutdown handlers
	// shutdownManager.RegisterShutdownHandler("server", HTTPServerShutdownHandler(server))
	// shutdownManager.RegisterShutdownHandler("database", DatabaseShutdownHandler(dbManager))

	// 3. Start your service normally
	// go server.ListenAndServe()

	// 4. Wait for shutdown
	// shutdownManager.WaitForShutdown()

	// 5. Service is now gracefully shut down
}
EOF

    echo "Created shutdown integration example for $service_name"
}

# Find all service directories and create shutdown managers
find microservices -mindepth 1 -maxdepth 1 -type d -not -name "shared-resilience" | while read -r service_dir; do
    service_name=$(basename "$service_dir")
    echo "Processing service: $service_name"

    create_shutdown_manager "$service_dir" "$service_name"
    create_shutdown_integration_example "$service_dir" "$service_name"
done

echo ""
echo "Graceful shutdown patterns implemented!"
echo ""
echo "Manual integration steps:"
echo "1. Import the shutdown package in your main.go"
echo "2. Create shutdown manager early in service initialization"
echo "3. Register shutdown handlers for all critical resources:"
echo "   - HTTP servers"
echo "   - Database connections"
echo "   - Background workers"
echo "   - Cache connections"
echo "   - File handles"
echo "   - Network connections"
echo "4. Replace existing signal handling with shutdownManager.WaitForShutdown()"
echo "5. Test graceful shutdown with SIGTERM and SIGINT signals"
echo ""
echo "Integration pattern:"
echo 'shutdownManager := shutdown.NewManager(logger, shutdown.Config{Timeout: 30 * time.Second})'
echo 'shutdownManager.RegisterShutdownHandler("server", shutdown.HTTPServerShutdownHandler(server))'
echo 'shutdownManager.RegisterShutdownHandler("database", shutdown.DatabaseShutdownHandler(dbManager))'
echo 'defer shutdownManager.WaitForShutdown()'