// Package handlers provides HTTP handlers for the Tenant Admin Service.
//
// This file contains the base handler struct and shared functionality used across
// all domain-specific handlers. It defines the core TenantAdminHandler type and
// common operations like health checks and tenant ID extraction.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantAdminHandler handles tenant admin-related HTTP requests.
//
// This is the main handler struct that provides access to services and utilities
// needed by all handler functions. It follows the dependency injection pattern,
// receiving all dependencies through its constructor.
//
// Fields:
//   - service: Core tenant admin service for database operations
//   - statusPageService: Service for managing status pages
//   - logger: Structured logger for request/error logging
type TenantAdminHandler struct {
	service           *services.TenantAdminService
	statusPageService *services.StatusPageManagementService
	logger            *zap.Logger
}

// NewTenantAdminHandler creates a new tenant admin handler with injected dependencies.
//
// This constructor follows the dependency injection pattern, making the handler
// easier to test and more maintainable.
//
// Parameters:
//   - service: The tenant admin service for database operations
//   - statusPageService: The status page management service
//   - logger: A structured logger instance
//
// Returns:
//   - *TenantAdminHandler: A fully initialized handler ready to process requests
//
// Example usage:
//
//	handler := NewTenantAdminHandler(tenantService, statusPageService, logger)
//	router.GET("/health", handler.HealthCheck)
func NewTenantAdminHandler(service *services.TenantAdminService, statusPageService *services.StatusPageManagementService, logger *zap.Logger) *TenantAdminHandler {
	return &TenantAdminHandler{
		service:           service,
		statusPageService: statusPageService,
		logger:            logger,
	}
}

// HealthCheck handles health check requests for monitoring and load balancers.
//
// This endpoint is used by infrastructure tools (Kubernetes, Docker, load balancers)
// to verify the service is running and healthy. It checks the database connection
// and returns appropriate status codes.
//
// Endpoint: GET /health
//
// Response (healthy):
//
//	Status: 200 OK
//	{
//	    "status": "healthy",
//	    "service": "tenant-admin-service",
//	    "version": "1.0.0"
//	}
//
// Response (unhealthy):
//
//	Status: 503 Service Unavailable
//	{
//	    "status": "unhealthy",
//	    "error": "database connection failed",
//	    "service": "tenant-admin-service"
//	}
func (h *TenantAdminHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "tenant-admin-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "tenant-admin-service",
		"version": "1.0.0",
	})
}

// getTenantID retrieves and validates the tenant ID from the Gin context.
//
// This helper is used by handlers that need to extract the tenant ID that was
// set by the TenantContextMiddleware. It performs validation and returns
// appropriate errors if the tenant ID is missing or malformed.
//
// The tenant ID is expected to be set by middleware earlier in the request chain
// and should be a valid UUID string.
//
// Parameters:
//   - c: The Gin context containing the tenant_id value
//
// Returns:
//   - uuid.UUID: The parsed tenant UUID
//   - error: An error if the tenant ID is missing or invalid
//
// Example usage:
//
//	tenantID, err := h.getTenantID(c)
//	if err != nil {
//	    h.logger.Error("Failed to get tenant ID", zap.Error(err))
//	    RespondWithError(c, http.StatusBadRequest, "Invalid tenant context")
//	    return
//	}
func (h *TenantAdminHandler) getTenantID(c *gin.Context) (uuid.UUID, error) {
	tenantIDStr, exists := middleware.GetTenantID(c)
	if !exists {
		h.logger.Error("Tenant ID not found in context")
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.logger.Error("Invalid tenant ID format", zap.String("tenant_id", tenantIDStr), zap.Error(err))
		return uuid.Nil, fmt.Errorf("invalid tenant ID format: %w", err)
	}

	return tenantID, nil
}
