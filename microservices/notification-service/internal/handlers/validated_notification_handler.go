// Package handlers provides HTTP handlers for the Notification Service with comprehensive validation.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anupamdutta5/notification-service/internal/models"
	"github.com/anupamdutta5/notification-service/internal/services"
	"github.com/anupamdutta5/notification-service/internal/validation"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ValidatedNotificationHandler handles notification-related HTTP requests with comprehensive validation.
type ValidatedNotificationHandler struct {
	notificationService *services.NotificationService
	validator           *validation.ServiceValidator
	logger              *zap.Logger
}

// NewValidatedNotificationHandler creates a new validated notification handler.
func NewValidatedNotificationHandler(notificationService *services.NotificationService, logger *zap.Logger) *ValidatedNotificationHandler {
	return &ValidatedNotificationHandler{
		notificationService: notificationService,
		validator:           validation.NewServiceValidator(logger),
		logger:              logger,
	}
}

// SendNotificationRequest represents a validated notification sending request
type SendNotificationRequest struct {
	Type        string                 `json:"type" binding:"required"`
	Recipients  []string              `json:"recipients" binding:"required"`
	Subject     string                `json:"subject" binding:"required"`
	Message     string                `json:"message" binding:"required"`
	Priority    string                `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
	ScheduledAt string                `json:"scheduled_at"`
}

// Validate validates and sanitizes the notification request
func (r *SendNotificationRequest) Validate(validator *validation.ServiceValidator) error {
	var err error

	// Validate notification type
	allowedTypes := []string{"email", "sms", "push", "webhook", "slack"}
	r.Type, err = validator.ValidateEnum(r.Type, allowedTypes, "type")
	if err != nil {
		return err
	}

	// Validate recipients
	if len(r.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if len(r.Recipients) > 100 {
		return fmt.Errorf("cannot send to more than 100 recipients at once")
	}

	// Validate each recipient based on notification type
	for i, recipient := range r.Recipients {
		switch r.Type {
		case "email":
			r.Recipients[i], err = validator.ValidateEmail(recipient)
			if err != nil {
				return fmt.Errorf("invalid email recipient at position %d: %w", i, err)
			}
		case "webhook":
			r.Recipients[i], err = validator.ValidateURL(recipient)
			if err != nil {
				return fmt.Errorf("invalid webhook URL at position %d: %w", i, err)
			}
		default:
			// For other types (SMS, push, slack), sanitize as string
			r.Recipients[i], err = validator.ValidateAndSanitizeString(recipient, "recipient", true)
			if err != nil {
				return fmt.Errorf("invalid recipient at position %d: %w", i, err)
			}
		}
	}

	// Validate and sanitize subject
	r.Subject, err = validator.ValidateAndSanitizeString(r.Subject, "subject", true)
	if err != nil {
		return err
	}

	if err := validator.ValidateStringLength(r.Subject, 1, 255, "subject"); err != nil {
		return err
	}

	// Validate and sanitize message
	r.Message, err = validator.ValidateAndSanitizeString(r.Message, "message", true)
	if err != nil {
		return err
	}

	if err := validator.ValidateStringLength(r.Message, 1, 10000, "message"); err != nil {
		return err
	}

	// Validate priority
	if r.Priority != "" {
		allowedPriorities := []string{"low", "normal", "high", "critical"}
		r.Priority, err = validator.ValidateEnum(r.Priority, allowedPriorities, "priority")
		if err != nil {
			return err
		}
	} else {
		r.Priority = "normal" // Default priority
	}

	// Sanitize metadata
	if r.Metadata != nil {
		r.Metadata = validator.SanitizeMap(r.Metadata)
	}

	// Validate scheduled_at if provided
	if r.ScheduledAt != "" {
		r.ScheduledAt, err = validator.ValidateAndSanitizeString(r.ScheduledAt, "scheduled_at", false)
		if err != nil {
			return err
		}
	}

	return nil
}

// SendNotification handles sending notifications with comprehensive validation.
func (h *ValidatedNotificationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid JSON in send notification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate and sanitize the request
	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn("Notification send validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.Warn("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	validatedTenantID, err := h.validator.ValidateID(tenantID, "tenant_id")
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Convert recipients array to JSON string
	recipientsJSON, err := json.Marshal(req.Recipients)
	if err != nil {
		h.logger.Error("Failed to marshal recipients", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process recipients"})
		return
	}

	// Convert metadata map to JSON string
	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		h.logger.Error("Failed to marshal metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process metadata"})
		return
	}

	// Parse scheduled time if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			h.logger.Error("Failed to parse scheduled_at", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at format, use RFC3339"})
			return
		}
		scheduledAt = &parsedTime
	}

	// Create notification object
	notification := &models.Notification{
		TenantID:    validatedTenantID,
		Type:        req.Type,
		Recipients:  string(recipientsJSON),
		Subject:     req.Subject,
		Content:     req.Message,
		Priority:    req.Priority,
		Metadata:    string(metadataJSON),
		ScheduledAt: scheduledAt,
		Status:      "pending",
	}

	// First save the notification
	if err := h.notificationService.CreateNotification(notification); err != nil {
		h.logger.Error("Failed to create notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// Send notification through service
	err = h.notificationService.SendNotification(notification.ID)
	if err != nil {
		h.logger.Error("Failed to send notification",
			zap.Error(err),
			zap.String("type", req.Type),
			zap.Int("recipient_count", len(req.Recipients)),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	h.logger.Info("Notification sent successfully",
		zap.Uint("notification_id", notification.ID),
		zap.String("type", req.Type),
		zap.Int("recipient_count", len(req.Recipients)),
		zap.Uint("tenant_id", validatedTenantID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification sent successfully",
		"notification_id": notification.ID,
		"status": notification.Status,
	})
}

// GetNotificationStatus handles getting notification status with validation.
func (h *ValidatedNotificationHandler) GetNotificationStatus(c *gin.Context) {
	// Validate notification ID from URL parameter
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	// Sanitize notification ID
	sanitizedID, err := h.validator.ValidateAndSanitizeString(notificationID, "notification_id", true)
	if err != nil {
		h.logger.Warn("Invalid notification ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.Warn("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	validatedTenantID, err := h.validator.ValidateID(tenantID, "tenant_id")
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	status, err := h.notificationService.GetNotificationStatus(validatedTenantID)
	if err != nil {
		h.logger.Error("Failed to get notification status",
			zap.Error(err),
			zap.String("notification_id", sanitizedID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// Health returns the health status of the Notification Service.
func (h *ValidatedNotificationHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "notification-service",
		"version":   "1.0.0",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}