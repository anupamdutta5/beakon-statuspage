// Package handlers provides HTTP request handlers for the Tenant Admin Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
)

// SubscriberHandler handles subscriber management HTTP requests.
type SubscriberHandler struct {
	service *services.SubscriberService
	logger  *zap.Logger
}

// NewSubscriberHandler creates a new subscriber handler.
func NewSubscriberHandler(service *services.SubscriberService, logger *zap.Logger) *SubscriberHandler {
	return &SubscriberHandler{
		service: service,
		logger:  logger,
	}
}

// getTenantIDFromContext extracts and validates tenant ID from context.
func (h *SubscriberHandler) getTenantIDFromContext(c *gin.Context) (uuid.UUID, error) {
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, http.ErrNoCookie // Using as a sentinel error
	}

	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		return uuid.Nil, err
	}

	return tenantID, nil
}

// CreateSubscriberRequest represents the request body for creating a subscriber.
type CreateSubscriberRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone"`
	EventTypes  string `json:"event_types"`  // JSON array
	Components  string `json:"components"`   // JSON array
	Preferences string `json:"preferences"`  // JSON object
}

// UpdateSubscriberRequest represents the request body for updating a subscriber.
type UpdateSubscriberRequest struct {
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	Status      *string `json:"status"`
	EventTypes  *string `json:"event_types"`
	Components  *string `json:"components"`
	Preferences *string `json:"preferences"`
}

// GetSubscribers retrieves all subscribers for a tenant.
// GET /api/v1/subscribers
func (h *SubscriberHandler) GetSubscribers(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Check for status filter
	statusFilter := c.Query("status")
	if statusFilter != "" {
		subscribers, err := h.service.GetSubscribersByStatus(c.Request.Context(), tenantID, statusFilter)
		if err != nil {
			h.logger.Error("Failed to get subscribers by status",
				zap.String("tenant_id", tenantID.String()),
				zap.String("status", statusFilter),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": subscribers, "total": len(subscribers)})
		return
	}

	// Check for active-only filter
	activeOnly := c.Query("active_only") == "true"
	if activeOnly {
		subscribers, err := h.service.GetActiveSubscribers(c.Request.Context(), tenantID)
		if err != nil {
			h.logger.Error("Failed to get active subscribers",
				zap.String("tenant_id", tenantID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": subscribers, "total": len(subscribers)})
		return
	}

	// Fetch all subscribers with pagination
	subscribers, total, err := h.service.GetSubscribers(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get subscribers",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   subscribers,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetSubscriber retrieves a single subscriber by ID.
// GET /api/v1/subscribers/:id
func (h *SubscriberHandler) GetSubscriber(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	subscriberIDStr := c.Param("id")
	subscriberID, err := uuid.Parse(subscriberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	// Fetch subscriber
	subscriber, err := h.service.GetSubscriberByID(c.Request.Context(), tenantID, subscriberID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
			return
		}
		h.logger.Error("Failed to get subscriber",
			zap.String("tenant_id", tenantID.String()),
			zap.String("subscriber_id", subscriberIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriber})
}

// CreateSubscriber creates a new subscriber for a tenant.
// POST /api/v1/subscribers
func (h *SubscriberHandler) CreateSubscriber(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	var req CreateSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create subscriber model
	subscriber := &models.Subscriber{
		TenantID:    tenantID,
		Email:       req.Email,
		Phone:       req.Phone,
		EventTypes:  req.EventTypes,
		Components:  req.Components,
		Preferences: req.Preferences,
		Status:      "active",
	}

	// Create subscriber
	if err := h.service.CreateSubscriber(c.Request.Context(), tenantID, subscriber); err != nil {
		h.logger.Error("Failed to create subscriber",
			zap.String("tenant_id", tenantID.String()),
			zap.String("email", req.Email),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Subscriber created successfully",
		"data":    subscriber,
	})
}

// UpdateSubscriber updates subscriber information.
// PUT /api/v1/subscribers/:id
func (h *SubscriberHandler) UpdateSubscriber(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	subscriberIDStr := c.Param("id")
	subscriberID, err := uuid.Parse(subscriberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	var req UpdateSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.EventTypes != nil {
		updates["event_types"] = *req.EventTypes
	}
	if req.Components != nil {
		updates["components"] = *req.Components
	}
	if req.Preferences != nil {
		updates["preferences"] = *req.Preferences
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update subscriber
	if err := h.service.UpdateSubscriber(c.Request.Context(), tenantID, subscriberID, updates); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
			return
		}
		h.logger.Error("Failed to update subscriber",
			zap.String("tenant_id", tenantID.String()),
			zap.String("subscriber_id", subscriberIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscriber updated successfully"})
}

// VerifySubscriber marks a subscriber's email as verified.
// POST /api/v1/subscribers/:id/verify
func (h *SubscriberHandler) VerifySubscriber(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	subscriberIDStr := c.Param("id")
	subscriberID, err := uuid.Parse(subscriberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	// Verify subscriber
	if err := h.service.VerifySubscriber(c.Request.Context(), tenantID, subscriberID); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
			return
		}
		h.logger.Error("Failed to verify subscriber",
			zap.String("tenant_id", tenantID.String()),
			zap.String("subscriber_id", subscriberIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscriber verified successfully"})
}

// UnsubscribeByToken unsubscribes a subscriber using their unique token.
// POST /api/v1/subscribers/unsubscribe/:token
func (h *SubscriberHandler) UnsubscribeByToken(c *gin.Context) {
	token := c.Param("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsubscribe token is required"})
		return
	}

	// Unsubscribe by token (no tenant context needed - token is globally unique)
	if err := h.service.UnsubscribeByToken(c.Request.Context(), token); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired unsubscribe token"})
			return
		}
		h.logger.Error("Failed to unsubscribe by token",
			zap.String("token", token),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unsubscribed"})
}

// DeleteSubscriber soft-deletes a subscriber.
// DELETE /api/v1/subscribers/:id
func (h *SubscriberHandler) DeleteSubscriber(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	subscriberIDStr := c.Param("id")
	subscriberID, err := uuid.Parse(subscriberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	// Delete subscriber
	if err := h.service.DeleteSubscriber(c.Request.Context(), tenantID, subscriberID); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
			return
		}
		h.logger.Error("Failed to delete subscriber",
			zap.String("tenant_id", tenantID.String()),
			zap.String("subscriber_id", subscriberIDStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscriber deleted successfully"})
}

// GetSubscriberStats retrieves subscriber statistics for a tenant.
// GET /api/v1/subscribers/stats
func (h *SubscriberHandler) GetSubscriberStats(c *gin.Context) {
	// Extract tenant ID from context (set by TenantMiddleware)
	tenantID, err := h.getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	// Get subscriber stats
	stats, err := h.service.GetSubscriberStats(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get subscriber stats",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriber statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
