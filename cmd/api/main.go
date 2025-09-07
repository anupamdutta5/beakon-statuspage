package main

import (
	"flag"
	"os"

	"github.com/enterprise-status/statuspage/internal/api"
	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/enterprise-status/statuspage/pkg/database"
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

	logger.Log.Info("Starting Status Page application",
		zap.String("environment", cfg.Environment),
		zap.String("version", getVersion()),
		zap.String("config_path", configPath))

	// Connect to the database
	if err := database.Connect(&cfg.Database); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize the monitoring service
	monitoringService := services.NewMonitoringService()
	_ = monitoringService // Use the service to avoid unused variable warning

	// Create and start the API server
	server := api.NewServer(cfg)
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// getVersion returns the application version
func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}
	return version
}
