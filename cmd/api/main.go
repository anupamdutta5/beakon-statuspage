package main

import (

	"github.com/enterprise-status/statuspage/internal/api"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.InitLogger("development") // Temporary logger until config is loaded
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize logger
	logger.InitLogger(cfg.Environment)
	defer logger.Sync()

	// Connect to the database
	if err := database.Connect(&cfg.Database); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Start the monitoring service
	monitoringService := services.NewMonitoringService()
	monitoringService.Start()

	// Create and start the API server
	server := api.NewServer(cfg)
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
