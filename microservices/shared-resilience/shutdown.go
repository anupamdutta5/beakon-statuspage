package resilience

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

// ShutdownManager handles graceful shutdown across all services
type ShutdownManager struct {
	logger      *zap.Logger
	shutdownFns []ShutdownFunc
	timeout     time.Duration
	mu          sync.Mutex
	shutdown    chan os.Signal
	done        chan struct{}
}

// ShutdownFunc represents a function that should be called during shutdown
type ShutdownFunc func(ctx context.Context) error

// ShutdownConfig represents shutdown configuration
type ShutdownConfig struct {
	Timeout time.Duration `yaml:"timeout" env:"SHUTDOWN_TIMEOUT" default:"30s"`
}

// NewShutdownManager creates a new graceful shutdown manager
func NewShutdownManager(logger *zap.Logger, config ShutdownConfig) *ShutdownManager {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return &ShutdownManager{
		logger:      logger,
		shutdownFns: make([]ShutdownFunc, 0),
		timeout:     config.Timeout,
		shutdown:    shutdown,
		done:        make(chan struct{}),
	}
}

// RegisterShutdownHandler registers a function to be called during shutdown
func (sm *ShutdownManager) RegisterShutdownHandler(name string, fn ShutdownFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	wrappedFn := func(ctx context.Context) error {
		sm.logger.Info("Executing shutdown handler", zap.String("handler", name))

		start := time.Now()
		err := fn(ctx)
		duration := time.Since(start)

		if err != nil {
			sm.logger.Error("Shutdown handler failed",
				zap.String("handler", name),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			sm.logger.Info("Shutdown handler completed successfully",
				zap.String("handler", name),
				zap.Duration("duration", duration))
		}

		return err
	}

	sm.shutdownFns = append(sm.shutdownFns, wrappedFn)
}

// WaitForShutdown waits for a shutdown signal and executes all registered handlers
func (sm *ShutdownManager) WaitForShutdown() error {
	sig := <-sm.shutdown
	sm.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	return sm.executeShutdown()
}

// executeShutdown executes all registered shutdown handlers
func (sm *ShutdownManager) executeShutdown() error {
	sm.logger.Info("Starting graceful shutdown",
		zap.Int("handlers", len(sm.shutdownFns)),
		zap.Duration("timeout", sm.timeout))

	ctx, cancel := context.WithTimeout(context.Background(), sm.timeout)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, len(sm.shutdownFns))

	for i, fn := range sm.shutdownFns {
		wg.Add(1)
		go func(index int, shutdownFn ShutdownFunc) {
			defer wg.Done()
			if err := shutdownFn(ctx); err != nil {
				errCh <- fmt.Errorf("shutdown handler %d failed: %w", index, err)
			}
		}(i, fn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		sm.logger.Info("All shutdown handlers completed successfully")
		close(sm.done)
		return nil
	case <-ctx.Done():
		sm.logger.Error("Shutdown timeout exceeded", zap.Duration("timeout", sm.timeout))

		close(errCh)
		var errors []error
		for err := range errCh {
			errors = append(errors, err)
		}

		close(sm.done)

		if len(errors) > 0 {
			return fmt.Errorf("shutdown timeout with errors: %v", errors)
		}

		return fmt.Errorf("shutdown timeout")
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