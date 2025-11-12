// Package main is the entry point for the Audit Consumer.
// This is the v2.0 version using shared-resilience v2.0 primitives.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/audit-consumer/internal/config"
	"github.com/anupamdutta5/audit-consumer/internal/consumer"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from YAML
	loader := resilience.NewConfigLoader("configs")
	var cfg config.Config
	if err := loader.Load(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logger *zap.Logger
	var err error
	if cfg.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Audit Consumer (v2.0)",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Environment),
	)

	// Initialize consumer
	auditConsumer, err := consumer.NewAuditConsumer(&cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize audit consumer", zap.Error(err))
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in a goroutine
	go func() {
		logger.Info("Audit Consumer starting...")
		if err := auditConsumer.Start(ctx); err != nil {
			logger.Error("Audit consumer stopped with error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Audit Consumer...")

	// Cancel context to stop consumer
	cancel()

	// Give consumer time to finish processing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for consumer to stop
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Audit Consumer shutdown timeout")
	case <-time.After(5 * time.Second):
		logger.Info("Audit Consumer stopped gracefully")
	}

	logger.Info("Audit Consumer exited")
}
