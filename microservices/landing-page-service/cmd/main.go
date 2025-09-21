// Package main is the entry point for the Landing Page Service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/config"
	"github.com/anupamdutta5/landing-page-service/internal/server"
	"github.com/anupamdutta5/landing-page-service/pkg/logger"
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

	logger.Info("Starting Landing Page Service",
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
		logger.Info("Landing Page Service starting...")
		if err := srv.Start(ctx); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Landing Page Service...")

	// Cancel context to stop server
	cancel()

	// Give server time to finish processing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Wait for server to stop
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Landing Page Service shutdown timeout")
	default:
		logger.Info("Landing Page Service stopped gracefully")
	}

	logger.Info("Landing Page Service exited")
}

