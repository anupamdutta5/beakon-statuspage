// Package main is the entry point for the SaaS Admin Service with graceful shutdown.
package main

import (
	"context"
	"log"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/server"
	"github.com/anupamdutta5/saas-admin-service/internal/shutdown"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting SaaS Admin Service",
		zap.String("service", "saas-admin-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.Environment),
	)

	// Create shutdown manager early
	shutdownManager := shutdown.NewManager(logger, shutdown.Config{
		Timeout: 30 * time.Second,
	})

	// Create server
	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// Register shutdown handlers
	shutdownManager.RegisterShutdownHandler("server", func(ctx context.Context) error {
		logger.Info("Shutting down HTTP server...")
		return srv.Shutdown(ctx)
	})

	// Register database shutdown handler (if server exposes database manager)
	if dbShutdownHandler := getDBShutdownHandler(srv); dbShutdownHandler != nil {
		shutdownManager.RegisterShutdownHandler("database", dbShutdownHandler)
	}

	// Register custom cleanup handlers
	shutdownManager.RegisterShutdownHandler("cleanup", func(ctx context.Context) error {
		logger.Info("Performing final cleanup...")
		// Add any custom cleanup logic here
		// - Close connections
		// - Flush logs
		// - Save state
		// - Notify other services
		return nil
	})

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Error("Server error", zap.Error(err))
			cancel() // Signal shutdown
		}
	}()

	logger.Info("SaaS Admin Service started successfully")

	// Wait for graceful shutdown
	if err := shutdownManager.WaitForShutdown(); err != nil {
		logger.Error("Graceful shutdown failed", zap.Error(err))
	} else {
		logger.Info("SaaS Admin Service shut down gracefully")
	}
}

// getDBShutdownHandler extracts database shutdown handler from server if available
func getDBShutdownHandler(srv interface{}) func(context.Context) error {
	// Type assertion to check if server has a method to get database shutdown handler
	// This is a placeholder - adjust based on your actual server interface
	if serverWithDB, ok := srv.(interface {
		GetDatabaseShutdownHandler() func(context.Context) error
	}); ok {
		return serverWithDB.GetDatabaseShutdownHandler()
	}

	// Alternatively, if server has direct database access:
	if serverWithDBStats, ok := srv.(interface {
		CloseDatabaseConnections() error
	}); ok {
		return func(ctx context.Context) error {
			return serverWithDBStats.CloseDatabaseConnections()
		}
	}

	return nil
}