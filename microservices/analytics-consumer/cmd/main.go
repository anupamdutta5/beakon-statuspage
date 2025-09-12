// Package main is the entry point for the Analytics Consumer.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise-status/statuspage-analytics-consumer/internal/config"
	"github.com/enterprise-status/statuspage-analytics-consumer/internal/consumer"
	"github.com/enterprise-status/statuspage-analytics-consumer/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Analytics Consumer",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Environment))

	// Initialize consumer
	analyticsConsumer, err := consumer.NewAnalyticsConsumer(cfg, logger.Logger)
	if err != nil {
		logger.Fatal("Failed to initialize analytics consumer", zap.Error(err))
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in a goroutine
	go func() {
		logger.Info("Analytics Consumer starting...")
		if err := analyticsConsumer.Start(ctx); err != nil {
			logger.Fatal("Failed to start analytics consumer", zap.Error(err))
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

