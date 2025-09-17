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
