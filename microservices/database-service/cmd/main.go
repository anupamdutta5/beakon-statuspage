// Package main is the entry point for the Database Service.
// This is the v2.0 version using shared-resilience v2.0 primitives.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/database-service/internal/config"
	"github.com/anupamdutta5/database-service/internal/server"
	"go.uber.org/zap"
)

// ServiceInfo contains basic service metadata
type ServiceInfo struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// DatabaseServiceConfig is the complete configuration for this service
type DatabaseServiceConfig struct {
	// Shared configuration (server, monitoring, etc.)
	SharedConfig resilience.Config `yaml:",inline"`

	// Service-specific configuration
	Service ServiceInfo `yaml:"service"`
}

func main() {
	// ========================================
	// STEP 1: Load Configuration from YAML (v2.0)
	// ========================================
	loader := resilience.NewConfigLoader("configs")
	var cfg DatabaseServiceConfig
	if err := loader.Load(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration (fail fast)
	if err := cfg.SharedConfig.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// ========================================
	// STEP 2: Initialize Logger
	// ========================================
	var logger *zap.Logger
	var err error

	if cfg.Service.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Database Service (v2.0)",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Service.Environment),
		zap.String("shared_resilience", resilience.Version))

	// ========================================
	// STEP 3: Load Legacy Configuration
	// ========================================
	legacyCfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load legacy configuration", zap.Error(err))
	}

	// ========================================
	// STEP 4: Initialize Server
	// ========================================
	srv, err := server.New(legacyCfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize server", zap.Error(err))
	}

	// ========================================
	// STEP 5: Start Server
	// ========================================
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		logger.Info("Database Service starting",
			zap.Int("port", cfg.SharedConfig.Server.Port))
		if err := srv.Start(ctx); err != nil {
			logger.Error("Failed to start server", zap.Error(err))
			cancel()
		}
	}()

	// ========================================
	// STEP 6: Graceful Shutdown
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Database Service...")

	// Cancel context to stop server
	cancel()

	// Give server time to finish processing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.SharedConfig.Server.GracefulStop)
	defer shutdownCancel()

	// Wait for server to stop
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Database Service shutdown timeout")
	default:
		logger.Info("Database Service stopped gracefully")
	}

	logger.Info("Database Service exited gracefully")
}

