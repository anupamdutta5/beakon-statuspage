// Package main is the entry point for the Database Service.
// This is the modernized version using the shared-resilience module.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/database-service/internal/config"
	"github.com/anupamdutta5/database-service/internal/server"
	"github.com/anupamdutta5/database-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	resilConfig := resilience.LoadConfigFromEnv()

	// Create startup manager for proper error handling
	startupMgr, err := resilience.NewStartupManager("database-service", resilConfig)
	if err != nil {
		// This is the only acceptable use of fatal - when we can't even initialize logging
		fmt.Fprintf(os.Stderr, "Failed to create startup manager: %v\n", err)
		os.Exit(1)
	}

	// Set up panic recovery
	defer startupMgr.RecoverFromPanic()

	logger := startupMgr.Logger
	defer logger.Sync()

	// Validate configuration
	if err := startupMgr.ValidateConfiguration(); err != nil {
		startupMgr.HandleStartupError(err)
		return
	}

	// Load legacy configuration
	cfg, err := config.Load()
	if err != nil {
		startupMgr.HandleStartupError(fmt.Errorf("failed to load configuration: %w", err))
		return
	}

	logger.Info("Starting Database Service",
		zap.String("service", "database-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", resilConfig.Environment))

	// Initialize server
	srv, err := server.New(cfg, logger)
	if err != nil {
		startupMgr.HandleStartupError(fmt.Errorf("failed to initialize server: %w", err))
		return
	}

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		logger.Info("Database Service starting...")
		if err := srv.Start(ctx); err != nil {
			logger.Error("Failed to start server", zap.Error(err))
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Database Service...")

	// Cancel context to stop server
	cancel()

	// Give server time to finish processing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for server to stop
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Database Service shutdown timeout")
	default:
		logger.Info("Database Service stopped gracefully")
	}

	logger.Info("Database Service exited")
}

