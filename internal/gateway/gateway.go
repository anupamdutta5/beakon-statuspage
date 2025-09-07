package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/gateway/discovery"
	"github.com/enterprise-status/statuspage/internal/gateway/middleware"
	"github.com/enterprise-status/statuspage/internal/gateway/proxy"
	"github.com/enterprise-status/statuspage/internal/gateway/routing"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Gateway represents the API Gateway
type Gateway struct {
	config     *config.Config
	router     *routing.Router
	discovery  *discovery.ServiceDiscovery
	proxy      *proxy.ReverseProxy
	middleware *middleware.Middleware
}

// NewGateway creates a new API Gateway instance
func NewGateway(cfg *config.Config) *Gateway {
	// Initialize service discovery
	serviceDiscovery := discovery.NewServiceDiscovery(cfg)

	// Initialize reverse proxy
	reverseProxy := proxy.NewReverseProxy(cfg, serviceDiscovery)

	// Initialize middleware
	middleware := middleware.NewMiddleware(cfg)

	// Initialize router
	router := routing.NewRouter(cfg, reverseProxy, middleware)

	gateway := &Gateway{
		config:     cfg,
		router:     router,
		discovery:  serviceDiscovery,
		proxy:      reverseProxy,
		middleware: middleware,
	}

	// Setup routes
	gateway.setupRoutes()

	// Start service discovery
	go gateway.discovery.Start()

	return gateway
}

// ServeHTTP implements the http.Handler interface
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add request ID for tracing
	requestID := generateRequestID()
	ctx := context.WithValue(r.Context(), "request_id", requestID)
	r = r.WithContext(ctx)

	// Add request ID to response headers
	w.Header().Set("X-Request-ID", requestID)

	// Add CORS headers
	g.addCORSHeaders(w)

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Start request timing
	start := time.Now()

	// Process request through middleware and routing
	g.router.ServeHTTP(w, r)

	// Log request
	duration := time.Since(start)
	logger.Log.Info("Request processed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.Duration("duration", duration),
		zap.String("request_id", requestID),
	)
}

// setupRoutes configures all the routes for the API Gateway
func (g *Gateway) setupRoutes() {
	// Health check endpoint
	g.router.GET("/health", g.handleHealthCheck)

	// API versioning
	apiV1 := g.router.Group("/api/v1")

	// Public routes (no authentication required)
	public := apiV1.Group("/public")
	{
		// Status page routes
		public.GET("/status/:tenant", g.handlePublicStatusPage)
		public.GET("/status/:tenant/incidents", g.handlePublicIncidents)
		public.GET("/status/:tenant/components", g.handlePublicComponents)
	}

	// Admin routes (authentication required)
	admin := apiV1.Group("/admin")
	admin.Use(g.middleware.AuthRequired())
	{
		// Tenant management
		admin.GET("/tenants", g.handleGetTenants)
		admin.POST("/tenants", g.handleCreateTenant)
		admin.GET("/tenants/:id", g.handleGetTenant)
		admin.PUT("/tenants/:id", g.handleUpdateTenant)
		admin.DELETE("/tenants/:id", g.handleDeleteTenant)

		// User management
		admin.GET("/users", g.handleGetUsers)
		admin.POST("/users", g.handleCreateUser)
		admin.GET("/users/:id", g.handleGetUser)
		admin.PUT("/users/:id", g.handleUpdateUser)
		admin.DELETE("/users/:id", g.handleDeleteUser)
	}

	// Tenant-specific routes (tenant context required)
	tenant := apiV1.Group("/tenant")
	tenant.Use(g.middleware.AuthRequired())
	tenant.Use(g.middleware.TenantRequired())
	{
		// Components
		tenant.GET("/components", g.handleGetComponents)
		tenant.POST("/components", g.handleCreateComponent)
		tenant.GET("/components/:id", g.handleGetComponent)
		tenant.PUT("/components/:id", g.handleUpdateComponent)
		tenant.DELETE("/components/:id", g.handleDeleteComponent)

		// Incidents
		tenant.GET("/incidents", g.handleGetIncidents)
		tenant.POST("/incidents", g.handleCreateIncident)
		tenant.GET("/incidents/:id", g.handleGetIncident)
		tenant.PUT("/incidents/:id", g.handleUpdateIncident)
		tenant.DELETE("/incidents/:id", g.handleDeleteIncident)

		// Notifications
		tenant.GET("/notifications", g.handleGetNotifications)
		tenant.POST("/notifications", g.handleCreateNotification)
		tenant.PUT("/notifications/:id", g.handleUpdateNotification)
		tenant.DELETE("/notifications/:id", g.handleDeleteNotification)

		// Analytics
		tenant.GET("/analytics/overview", g.handleGetAnalyticsOverview)
		tenant.GET("/analytics/incidents", g.handleGetIncidentAnalytics)
		tenant.GET("/analytics/components", g.handleGetComponentAnalytics)

		// Monitoring
		tenant.GET("/monitoring/uptime", g.handleGetUptimeData)
		tenant.GET("/monitoring/metrics", g.handleGetMonitoringMetrics)

		// Payments
		tenant.GET("/billing/overview", g.handleGetBillingOverview)
		tenant.GET("/billing/invoices", g.handleGetInvoices)
		tenant.POST("/billing/subscribe", g.handleCreateSubscription)
		tenant.PUT("/billing/subscription/:id", g.handleUpdateSubscription)

		// Settings
		tenant.GET("/settings", g.handleGetSettings)
		tenant.PUT("/settings", g.handleUpdateSettings)
	}

	// Webhook routes
	webhooks := apiV1.Group("/webhooks")
	{
		// Payment webhooks
		webhooks.POST("/stripe", g.handleStripeWebhook)
		webhooks.POST("/paypal", g.handlePayPalWebhook)
		webhooks.POST("/razorpay", g.handleRazorpayWebhook)

		// Integration webhooks
		webhooks.POST("/slack", g.handleSlackWebhook)
		webhooks.POST("/pagerduty", g.handlePagerDutyWebhook)
	}

	// Metrics endpoint
	g.router.GET("/metrics", g.handleMetrics)
}

// addCORSHeaders adds CORS headers to the response
func (g *Gateway) addCORSHeaders(w http.ResponseWriter) {
	cors := g.config.Security.CORS
	if len(cors.AllowedOrigins) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(cors.AllowedOrigins, ", "))
	}
	if len(cors.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(cors.AllowedMethods, ", "))
	}
	if len(cors.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(cors.AllowedHeaders, ", "))
	}
	if cors.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Health check handler
func (g *Gateway) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","version":"%s"}`,
		time.Now().Format(time.RFC3339), getVersion())
}

// Metrics handler
func (g *Gateway) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Proxy to Prometheus metrics endpoint
	g.proxy.ServeHTTP(w, r, "monitoring-service", "/metrics")
}

// Public status page handlers
func (g *Gateway) handlePublicStatusPage(w http.ResponseWriter, r *http.Request) {
	tenant := g.router.GetParam(r, "tenant")
	g.proxy.ServeHTTP(w, r, "tenant-service", fmt.Sprintf("/public/status/%s", tenant))
}

func (g *Gateway) handlePublicIncidents(w http.ResponseWriter, r *http.Request) {
	tenant := g.router.GetParam(r, "tenant")
	g.proxy.ServeHTTP(w, r, "incident-service", fmt.Sprintf("/public/incidents/%s", tenant))
}

func (g *Gateway) handlePublicComponents(w http.ResponseWriter, r *http.Request) {
	tenant := g.router.GetParam(r, "tenant")
	g.proxy.ServeHTTP(w, r, "component-service", fmt.Sprintf("/public/components/%s", tenant))
}

// Admin handlers
func (g *Gateway) handleGetTenants(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "tenant-service", "/admin/tenants")
}

func (g *Gateway) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "tenant-service", "/admin/tenants")
}

func (g *Gateway) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "tenant-service", fmt.Sprintf("/admin/tenants/%s", id))
}

func (g *Gateway) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "tenant-service", fmt.Sprintf("/admin/tenants/%s", id))
}

func (g *Gateway) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "tenant-service", fmt.Sprintf("/admin/tenants/%s", id))
}

func (g *Gateway) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "user-service", "/admin/users")
}

func (g *Gateway) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "user-service", "/admin/users")
}

func (g *Gateway) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "user-service", fmt.Sprintf("/admin/users/%s", id))
}

func (g *Gateway) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "user-service", fmt.Sprintf("/admin/users/%s", id))
}

func (g *Gateway) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "user-service", fmt.Sprintf("/admin/users/%s", id))
}

// Tenant-specific handlers
func (g *Gateway) handleGetComponents(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "component-service", "/components")
}

func (g *Gateway) handleCreateComponent(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "component-service", "/components")
}

func (g *Gateway) handleGetComponent(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "component-service", fmt.Sprintf("/components/%s", id))
}

func (g *Gateway) handleUpdateComponent(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "component-service", fmt.Sprintf("/components/%s", id))
}

func (g *Gateway) handleDeleteComponent(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "component-service", fmt.Sprintf("/components/%s", id))
}

func (g *Gateway) handleGetIncidents(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "incident-service", "/incidents")
}

func (g *Gateway) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "incident-service", "/incidents")
}

func (g *Gateway) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "incident-service", fmt.Sprintf("/incidents/%s", id))
}

func (g *Gateway) handleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "incident-service", fmt.Sprintf("/incidents/%s", id))
}

func (g *Gateway) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "incident-service", fmt.Sprintf("/incidents/%s", id))
}

func (g *Gateway) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "notification-service", "/notifications")
}

func (g *Gateway) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "notification-service", "/notifications")
}

func (g *Gateway) handleUpdateNotification(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "notification-service", fmt.Sprintf("/notifications/%s", id))
}

func (g *Gateway) handleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "notification-service", fmt.Sprintf("/notifications/%s", id))
}

func (g *Gateway) handleGetAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "analytics-service", "/analytics/overview")
}

func (g *Gateway) handleGetIncidentAnalytics(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "analytics-service", "/analytics/incidents")
}

func (g *Gateway) handleGetComponentAnalytics(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "analytics-service", "/analytics/components")
}

func (g *Gateway) handleGetUptimeData(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "monitoring-service", "/monitoring/uptime")
}

func (g *Gateway) handleGetMonitoringMetrics(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "monitoring-service", "/monitoring/metrics")
}

func (g *Gateway) handleGetBillingOverview(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "payment-service", "/billing/overview")
}

func (g *Gateway) handleGetInvoices(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "payment-service", "/billing/invoices")
}

func (g *Gateway) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "payment-service", "/billing/subscribe")
}

func (g *Gateway) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id := g.router.GetParam(r, "id")
	g.proxy.ServeHTTP(w, r, "payment-service", fmt.Sprintf("/billing/subscription/%s", id))
}

func (g *Gateway) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "tenant-service", "/settings")
}

func (g *Gateway) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "tenant-service", "/settings")
}

// Webhook handlers
func (g *Gateway) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "payment-service", "/webhooks/stripe")
}

func (g *Gateway) handlePayPalWebhook(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "payment-service", "/webhooks/paypal")
}

func (g *Gateway) handleRazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "payment-service", "/webhooks/razorpay")
}

func (g *Gateway) handleSlackWebhook(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "notification-service", "/webhooks/slack")
}

func (g *Gateway) handlePagerDutyWebhook(w http.ResponseWriter, r *http.Request) {
	g.proxy.ServeHTTP(w, r, "notification-service", "/webhooks/pagerduty")
}

// getVersion returns the application version
func getVersion() string {
	return "1.0.0"
}
