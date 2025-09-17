// Package server provides the Landing Page Service server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/statuspage-landing-service/internal/config"
	"github.com/anupamdutta5/statuspage-landing-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-landing-service/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server represents the Landing Page Service server.
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	router  *gin.Engine
	server  *http.Server
	service *services.LandingService
}

// New creates a new Landing Page Service server.
func New(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize landing service
	service, err := services.NewLandingService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize landing service: %w", err)
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
	handler := handlers.NewLandingHandler(service, logger)

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

// Start starts the Landing Page Service server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting Landing Page Service server",
		zap.String("address", s.server.Addr))

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.Info("Shutting down Landing Page Service server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("Landing Page Service server stopped")
	return nil
}

// setupRoutes sets up the API routes.
func setupRoutes(router *gin.Engine, handler *handlers.LandingHandler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// Static files
	router.Static("/static", "./web/static")

	// Public routes
	router.GET("/", handler.LandingPage)
	router.GET("/blog", handler.BlogPage)
	router.GET("/blog/:slug", handler.ArticlePage)
	router.GET("/contact", handler.ContactPage)
	router.GET("/privacy", handler.PrivacyPage)
	router.GET("/terms", handler.TermsPage)

	// API routes
	api := router.Group("/api/v1")
	{
		// Contact form
		api.POST("/contact", handler.SubmitContactForm)

		// Newsletter
		api.POST("/newsletter", handler.SubscribeNewsletter)

		// Pricing plans sync (from SaaS Admin Service)
		api.POST("/pricing/sync", handler.SyncPricingPlans)

		// Content management
		api.GET("/hero", handler.GetHeroSection)
		api.POST("/hero", handler.CreateHeroSection)
		api.PUT("/hero/:id", handler.UpdateHeroSection)

		api.GET("/features", handler.GetFeatures)
		api.POST("/features", handler.CreateFeature)
		api.PUT("/features/:id", handler.UpdateFeature)

		api.GET("/testimonials", handler.GetTestimonials)
		api.POST("/testimonials", handler.CreateTestimonial)
		api.PUT("/testimonials/:id", handler.UpdateTestimonial)

		api.GET("/faqs", handler.GetFAQs)
		api.POST("/faqs", handler.CreateFAQ)
		api.PUT("/faqs/:id", handler.UpdateFAQ)

		api.GET("/articles", handler.GetArticles)
		api.POST("/articles", handler.CreateArticle)
		api.PUT("/articles/:id", handler.UpdateArticle)
		api.DELETE("/articles/:id", handler.DeleteArticle)
	}
}
