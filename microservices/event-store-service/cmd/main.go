// Package main is the entry point for the Event Store Service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise-status/statuspage-event-store-service/internal/config"
	"github.com/enterprise-status/statuspage-event-store-service/internal/server"
	"github.com/enterprise-status/statuspage-event-store-service/pkg/logger"
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

	logger.Info("Starting Event Store Service",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Environment))

	// Initialize server
	srv, err := server.New(cfg, logger.Logger)
	if err != nil {
		logger.Fatal("Failed to initialize server", zap.Error(err))
	}

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		logger.Info("Event Store Service starting...")
		if err := srv.Start(ctx); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Event Store Service...")

	// Cancel context to stop server
	cancel()

	// Give server time to finish processing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for server to stop
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Event Store Service shutdown timeout")
	default:
		logger.Info("Event Store Service stopped gracefully")
	}

	logger.Info("Event Store Service exited")
}

