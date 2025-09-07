package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/gateway"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// Initialize temporary logger for configuration errors
		logger.InitLogger("development")
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize logger with configuration
	logger.InitLogger(cfg.Environment)
	defer logger.Sync()

	logger.Log.Info("Starting API Gateway",
		zap.String("environment", cfg.Environment),
		zap.String("version", getVersion()),
		zap.String("config_path", configPath))

	// Create and start the API Gateway
	gateway := gateway.NewGateway(cfg)

	// Start the server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      gateway,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Log.Info("API Gateway starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start API Gateway", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down API Gateway...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("API Gateway stopped")
}

// getVersion returns the application version
func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}
	return version
}
