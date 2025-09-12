// Package main is the entry point for the Payment Service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise-status/statuspage-payment-service/internal/config"
	"github.com/enterprise-status/statuspage-payment-service/internal/handlers"
	"github.com/enterprise-status/statuspage-payment-service/internal/middleware"
	"github.com/enterprise-status/statuspage-payment-service/internal/services"
	"github.com/enterprise-status/statuspage-payment-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Server.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Payment Service",
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", cfg.Server.Environment))

	// Set Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := services.InitDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize services
	paymentService := services.NewPaymentService(db, logger.Logger)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Logger))
	router.Use(middleware.Recovery(logger.Logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Initialize handlers
	paymentHandler := handlers.NewPaymentHandler(paymentService, logger.Logger)

	// Setup routes
	setupRoutes(router, paymentHandler, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Payment Service server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Payment Service server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Payment Service server exited")
}

// setupRoutes configures all the routes for the Payment Service.
func setupRoutes(router *gin.Engine, handler *handlers.PaymentHandler, cfg *config.Config) {
	// Health check endpoint
	router.GET("/health", handler.Health)

	// Webhook endpoints (no authentication required)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/stripe", handler.StripeWebhook)
		webhooks.POST("/paypal", handler.PayPalWebhook)
		webhooks.POST("/razorpay", handler.RazorpayWebhook)
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/public")
		{
			public.GET("/plans", handler.GetPublicPlans)
			public.GET("/plans/:id", handler.GetPublicPlan)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Payment management routes
			payments := protected.Group("/payments")
			{
				payments.GET("", handler.GetPayments)
				payments.POST("", handler.CreatePayment)
				payments.GET("/:id", handler.GetPayment)
				payments.PUT("/:id", handler.UpdatePayment)
				payments.POST("/:id/refund", handler.RefundPayment)
				payments.GET("/:id/transactions", handler.GetPaymentTransactions)
			}

			// Subscription management
			subscriptions := protected.Group("/subscriptions")
			{
				subscriptions.GET("", handler.GetSubscriptions)
				subscriptions.POST("", handler.CreateSubscription)
				subscriptions.GET("/:id", handler.GetSubscription)
				subscriptions.PUT("/:id", handler.UpdateSubscription)
				subscriptions.DELETE("/:id", handler.CancelSubscription)
				subscriptions.POST("/:id/upgrade", handler.UpgradeSubscription)
				subscriptions.POST("/:id/downgrade", handler.DowngradeSubscription)
			}

			// Plan management
			plans := protected.Group("/plans")
			{
				plans.GET("", handler.GetPlans)
				plans.POST("", handler.CreatePlan)
				plans.GET("/:id", handler.GetPlan)
				plans.PUT("/:id", handler.UpdatePlan)
				plans.DELETE("/:id", handler.DeletePlan)
			}

			// Billing management
			billing := protected.Group("/billing")
			{
				billing.GET("/invoices", handler.GetInvoices)
				billing.GET("/invoices/:id", handler.GetInvoice)
				billing.POST("/invoices/:id/pay", handler.PayInvoice)
				billing.GET("/usage", handler.GetUsage)
				billing.GET("/history", handler.GetBillingHistory)
			}
		}
	}
}
