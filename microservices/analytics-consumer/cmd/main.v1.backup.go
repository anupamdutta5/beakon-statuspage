// Package main is the entry point for the Analytics Consumer.
// This is the modernized version using the shared-resilience module.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/analytics-consumer/internal/config"
	"github.com/anupamdutta5/analytics-consumer/internal/consumer"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	resilConfig := resilience.LoadConfigFromEnv()

	// Create startup manager for proper error handling
	startupMgr, err := resilience.NewStartupManager("analytics-consumer", resilConfig)
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

	logger.Info("Starting Analytics Consumer",
		zap.String("service", "analytics-consumer"),
		zap.String("version", "1.0.0"),
		zap.String("environment", resilConfig.Environment))

	// Initialize consumer
	analyticsConsumer, err := consumer.NewAnalyticsConsumer(cfg, logger)
	if err != nil {
		startupMgr.HandleStartupError(fmt.Errorf("failed to initialize analytics consumer: %w", err))
		return
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in a goroutine
	go func() {
		logger.Info("Analytics Consumer starting...")
		if err := analyticsConsumer.Start(ctx); err != nil {
			logger.Error("Failed to start analytics consumer", zap.Error(err))
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Analytics Consumer...")

	// Cancel context to stop consumer
	cancel()

	// Give consumer time to finish processing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for consumer to stop
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Analytics Consumer shutdown timeout")
	default:
		logger.Info("Analytics Consumer stopped gracefully")
	}

	logger.Info("Analytics Consumer exited")
}

