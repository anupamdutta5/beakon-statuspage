// Package server provides the Event Store Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/event-store-service/internal/config"
	"github.com/anupamdutta5/event-store-service/internal/handlers"
	"github.com/anupamdutta5/event-store-service/internal/middleware"
	"github.com/anupamdutta5/event-store-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the Event Store Service server.
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	router  *gin.Engine
	server  *http.Server
	service *services.EventStoreService
}

// New creates a new Event Store Service server.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize event store service
	service, err := services.NewEventStoreService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event store service: %w", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(cors.Default())
	router.Use(gin.Logger())

	// Initialize handlers
	handler := handlers.NewEventStoreHandler(service, logger)

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

// Start starts the Event Store Service server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting Event Store Service server",
		zap.String("address", s.server.Addr))

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("Shutting down Event Store Service server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("Event Store Service server stopped")
	return nil
}

// setupRoutes sets up the API routes.
func setupRoutes(router *gin.Engine, handler *handlers.EventStoreHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Stream operations
		v1.GET("/streams", handler.ListStreams)
		v1.POST("/streams", handler.CreateStream)
		v1.GET("/streams/:id", handler.GetStream)
		v1.DELETE("/streams/:id", handler.DeleteStream)

		// Event operations
		v1.POST("/streams/:id/events", handler.AppendEvents)
		v1.GET("/streams/:id/events", handler.GetEvents)
		v1.GET("/streams/:id/events/:event_id", handler.GetEvent)

		// Snapshot operations
		v1.POST("/streams/:id/snapshots", handler.CreateSnapshot)
		v1.GET("/streams/:id/snapshots", handler.GetSnapshots)
		v1.GET("/streams/:id/snapshots/:snapshot_id", handler.GetSnapshot)

		// Projection operations
		v1.GET("/projections", handler.ListProjections)
		v1.POST("/projections", handler.CreateProjection)
		v1.GET("/projections/:id", handler.GetProjection)
		v1.PUT("/projections/:id", handler.UpdateProjection)
		v1.DELETE("/projections/:id", handler.DeleteProjection)
		v1.POST("/projections/:id/start", handler.StartProjection)
		v1.POST("/projections/:id/stop", handler.StopProjection)
		v1.POST("/projections/:id/reset", handler.ResetProjection)

		// Subscription operations
		v1.GET("/subscriptions", handler.ListSubscriptions)
		v1.POST("/subscriptions", handler.CreateSubscription)
		v1.GET("/subscriptions/:id", handler.GetSubscription)
		v1.PUT("/subscriptions/:id", handler.UpdateSubscription)
		v1.DELETE("/subscriptions/:id", handler.DeleteSubscription)

		// Statistics
		v1.GET("/stats", handler.GetStats)
		v1.GET("/streams/:id/stats", handler.GetStreamStats)
	}
}

