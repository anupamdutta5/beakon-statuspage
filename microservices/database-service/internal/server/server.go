// Package server provides the Database Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/database-service/internal/config"
	"github.com/anupamdutta5/database-service/internal/handlers"
	"github.com/anupamdutta5/database-service/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the Database Service server.
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	router  *gin.Engine
	server  *http.Server
	service *services.DatabaseService
}

// New creates a new Database Service server.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize database service
	service, err := services.NewDatabaseService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database service: %w", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())
	// TODO: Add custom request ID and security middleware

	// Initialize handlers
	handler := handlers.NewDatabaseHandler(service, logger)

	// Setup routes
	setupRoutes(router, handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return &Server{
		config:  cfg,
		logger:  logger,
		router:  router,
		server:  server,
		service: service,
	}, nil
}

// Start starts the Database Service server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting Database Service server",
		zap.String("address", s.server.Addr))

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("Shutting down Database Service server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("Database Service server stopped")
	return nil
}

// setupRoutes sets up the API routes.
func setupRoutes(router *gin.Engine, handler *handlers.DatabaseHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Database operations
		v1.GET("/databases", handler.ListDatabases)
		v1.POST("/databases", handler.CreateDatabase)
		v1.GET("/databases/:id", handler.GetDatabase)
		v1.PUT("/databases/:id", handler.UpdateDatabase)
		v1.DELETE("/databases/:id", handler.DeleteDatabase)

		// Table operations
		v1.GET("/databases/:id/tables", handler.ListTables)
		v1.POST("/databases/:id/tables", handler.CreateTable)
		v1.GET("/databases/:id/tables/:table", handler.GetTable)
		v1.PUT("/databases/:id/tables/:table", handler.UpdateTable)
		v1.DELETE("/databases/:id/tables/:table", handler.DeleteTable)

		// Data operations
		v1.GET("/databases/:id/tables/:table/data", handler.GetData)
		v1.POST("/databases/:id/tables/:table/data", handler.CreateData)
		v1.PUT("/databases/:id/tables/:table/data/:id", handler.UpdateData)
		v1.DELETE("/databases/:id/tables/:table/data/:id", handler.DeleteData)

		// Query operations
		v1.POST("/databases/:id/query", handler.ExecuteQuery)
		v1.POST("/databases/:id/transaction", handler.ExecuteTransaction)

		// Backup operations
		v1.POST("/databases/:id/backup", handler.CreateBackup)
		v1.GET("/databases/:id/backups", handler.ListBackups)
		v1.GET("/databases/:id/backups/:backup", handler.GetBackup)
		v1.POST("/databases/:id/restore", handler.RestoreBackup)

		// Migration operations
		v1.GET("/databases/:id/migrations", handler.ListMigrations)
		v1.POST("/databases/:id/migrations", handler.RunMigration)
		v1.POST("/databases/:id/migrations/rollback", handler.RollbackMigration)

		// Statistics
		v1.GET("/databases/:id/stats", handler.GetDatabaseStats)
		v1.GET("/databases/:id/tables/:table/stats", handler.GetTableStats)
	}
}

