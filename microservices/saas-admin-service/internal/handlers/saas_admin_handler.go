// Package handlers provides HTTP handlers for the SaaS Admin Service.
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/auth"
	"github.com/anupamdutta5/saas-admin-service/internal/config"
	"github.com/anupamdutta5/saas-admin-service/internal/events"
	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SaaSAdminHandler handles SaaS admin-related HTTP requests.
type SaaSAdminHandler struct {
	httpClient           *http.Client
	service              *services.SaaSAdminService
	serviceURLs          config.ServiceURLs
	tenantAdminServiceURL string
	tenantAdminDB        *gorm.DB // Direct connection to tenant_admin_db for credential sync (deprecated, use events)
	eventPublisher       *events.Publisher // RabbitMQ event publisher for tenant lifecycle events
	logger               *zap.Logger
	jwtManager           *auth.JWTManager
}

// NewSaaSAdminHandler creates a new SaaS admin handler.
func NewSaaSAdminHandler(service *services.SaaSAdminService, tenantAdminDB *gorm.DB, eventPublisher *events.Publisher, serviceURLs config.ServiceURLs, logger *zap.Logger) *SaaSAdminHandler {
	tenantAdminURL := os.Getenv("TENANT_ADMIN_SERVICE_URL")
	if tenantAdminURL == "" {
		tenantAdminURL = "http://localhost:8099"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production"
	}

	// Create JWT manager with 24-hour token expiration
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	return &SaaSAdminHandler{
		httpClient:            &http.Client{},
		service:               service,
		serviceURLs:           serviceURLs,
		tenantAdminServiceURL: tenantAdminURL,
		tenantAdminDB:         tenantAdminDB,
		eventPublisher:        eventPublisher,
		logger:                logger,
		jwtManager:            jwtManager,
	}
}

// Core SaaS Admin Methods - These methods handle platform-wide administration

// HealthCheck handles health check requests.
func (h *SaaSAdminHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "saas-admin-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "saas-admin-service",
		"version": "1.0.0",
	})
}

// GetPlatform handles retrieving platform configuration.
func (h *SaaSAdminHandler) GetPlatform(c *gin.Context) {
	h.logger.Info("Getting platform configuration")

	platform, err := h.service.GetPlatform(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get platform", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get platform configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"platform": platform,
	})
}

// UpdatePlatform handles updating platform configuration.
func (h *SaaSAdminHandler) UpdatePlatform(c *gin.Context) {
	var updates models.Platform
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating platform configuration")

	if err := h.service.UpdatePlatform(c.Request.Context(), &updates); err != nil {
		h.logger.Error("Failed to update platform", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update platform configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Platform configuration updated successfully",
	})
}

// ListPlans handles listing SaaS plans.
func (h *SaaSAdminHandler) ListPlans(c *gin.Context) {
	h.logger.Info("Listing SaaS plans")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	plans, err := h.service.ListPlans(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list plans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans":  plans,
		"count":  len(plans),
		"limit":  limit,
		"offset": offset,
	})
}

// CreatePlan handles creating a new SaaS plan.
func (h *SaaSAdminHandler) CreatePlan(c *gin.Context) {
	var plan models.SaaSPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate input
	if strings.TrimSpace(plan.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Plan name is required",
		})
		return
	}

	h.logger.Info("Creating SaaS plan", zap.String("name", plan.Name))

	if err := h.service.CreatePlan(c.Request.Context(), &plan); err != nil {
		h.logger.Error("Failed to create plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create plan",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Plan created successfully",
		"plan":    plan,
	})
}

// GetPlan handles retrieving a specific SaaS plan.
func (h *SaaSAdminHandler) GetPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	h.logger.Info("Getting SaaS plan", zap.String("id", idStr))

	plan, err := h.service.GetPlan(c.Request.Context(), id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Plan not found",
			})
			return
		}
		h.logger.Error("Failed to get plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
	})
}

// UpdatePlan handles updating a SaaS plan.
func (h *SaaSAdminHandler) UpdatePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	var updates models.SaaSPlan
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating SaaS plan", zap.String("id", idStr))

	updatedPlan, err := h.service.UpdatePlan(c.Request.Context(), id, &updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Plan not found",
			})
			return
		}
		h.logger.Error("Failed to update plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan updated successfully",
		"plan":    updatedPlan,
	})
}

// DeletePlan handles deleting a SaaS plan.
func (h *SaaSAdminHandler) DeletePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	h.logger.Info("Deleting SaaS plan", zap.String("id", idStr))

	if err := h.service.DeletePlan(c.Request.Context(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Plan not found",
			})
			return
		}
		h.logger.Error("Failed to delete plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan deleted successfully",
	})
}

// GetPlanBySlug handles retrieving a plan by slug.
func (h *SaaSAdminHandler) GetPlanBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if strings.TrimSpace(slug) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Plan slug is required",
		})
		return
	}

	h.logger.Info("Getting plan by slug", zap.String("slug", slug))

	plan, err := h.service.GetPlanBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Plan not found",
			})
			return
		}
		h.logger.Error("Failed to get plan by slug", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
	})
}

// ListFeatures handles listing features.
func (h *SaaSAdminHandler) ListFeatures(c *gin.Context) {
	h.logger.Info("Listing features")

	// Parse query parameters
	category := c.DefaultQuery("category", "")
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	features, err := h.service.ListFeatures(c.Request.Context(), category, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list features", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list features",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"count":    len(features),
		"limit":    limit,
		"offset":   offset,
	})
}

// CreateFeature handles creating a new feature.
func (h *SaaSAdminHandler) CreateFeature(c *gin.Context) {
	var feature models.SaaSFeature
	if err := c.ShouldBindJSON(&feature); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating feature", zap.String("name", feature.Name))

	if err := h.service.CreateFeature(c.Request.Context(), &feature); err != nil {
		h.logger.Error("Failed to create feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feature",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feature created successfully",
		"feature": feature,
	})
}

// GetFeature handles retrieving a specific feature.
func (h *SaaSAdminHandler) GetFeature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature ID",
		})
		return
	}

	feature, err := h.service.GetFeature(c.Request.Context(), uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Feature not found",
			})
			return
		}
		h.logger.Error("Failed to get feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature": feature,
	})
}

// UpdateFeature handles updating a feature.
func (h *SaaSAdminHandler) UpdateFeature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature ID",
		})
		return
	}

	var updates models.SaaSFeature
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.service.UpdateFeature(c.Request.Context(), uint(id), &updates); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Feature not found",
			})
			return
		}
		h.logger.Error("Failed to update feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature updated successfully",
	})
}

// DeleteFeature handles deleting a feature.
func (h *SaaSAdminHandler) DeleteFeature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature ID",
		})
		return
	}

	if err := h.service.DeleteFeature(c.Request.Context(), uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Feature not found",
			})
			return
		}
		h.logger.Error("Failed to delete feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature deleted successfully",
	})
}

// ListFeatureFlags handles listing feature flags.
func (h *SaaSAdminHandler) ListFeatureFlags(c *gin.Context) {
	h.logger.Info("Listing feature flags")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	flags, err := h.service.ListFeatureFlags(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list feature flags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list feature flags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flags": flags,
		"count":         len(flags),
		"limit":         limit,
		"offset":        offset,
	})
}

// CreateFeatureFlag handles creating a new feature flag.
func (h *SaaSAdminHandler) CreateFeatureFlag(c *gin.Context) {
	var flag models.SaaSFeatureFlag
	if err := c.ShouldBindJSON(&flag); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating feature flag", zap.String("name", flag.Name))

	if err := h.service.CreateFeatureFlag(c.Request.Context(), &flag); err != nil {
		h.logger.Error("Failed to create feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feature flag",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Feature flag created successfully",
		"feature_flag": flag,
	})
}

// GetFeatureFlag handles retrieving a specific feature flag.
func (h *SaaSAdminHandler) GetFeatureFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	flag, err := h.service.GetFeatureFlag(c.Request.Context(), uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Feature flag not found",
			})
			return
		}
		h.logger.Error("Failed to get feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flag": flag,
	})
}

// UpdateFeatureFlag handles updating a feature flag.
func (h *SaaSAdminHandler) UpdateFeatureFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	var updates models.SaaSFeatureFlag
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	updatedFlag, err := h.service.UpdateFeatureFlag(c.Request.Context(), uint(id), &updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Feature flag not found",
			})
			return
		}
		h.logger.Error("Failed to update feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Feature flag updated successfully",
		"feature_flag": updatedFlag,
	})
}

// DeleteFeatureFlag handles deleting a feature flag.
func (h *SaaSAdminHandler) DeleteFeatureFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	if err := h.service.DeleteFeatureFlag(c.Request.Context(), uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Feature flag not found",
			})
			return
		}
		h.logger.Error("Failed to delete feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature flag deleted successfully",
	})
}

// GetStats handles retrieving platform statistics.
func (h *SaaSAdminHandler) GetStats(c *gin.Context) {
	h.logger.Info("Getting platform statistics")

	// This would typically aggregate stats from various services
	// For now, return basic platform stats
	c.JSON(http.StatusOK, gin.H{
		"platform_stats": gin.H{
			"status":      "operational",
			"total_plans": 0, // This would be retrieved from service
			"total_features": 0, // This would be retrieved from service
			"total_feature_flags": 0, // This would be retrieved from service
		},
	})
}

// ListAdminUsers handles listing admin users.
func (h *SaaSAdminHandler) ListAdminUsers(c *gin.Context) {
	h.logger.Info("Listing admin users")
	// This would typically proxy to user service or handle internally
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Admin user management not yet implemented",
	})
}

// CreateAdminUser handles creating admin users.
func (h *SaaSAdminHandler) CreateAdminUser(c *gin.Context) {
	h.logger.Info("Creating admin user")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Admin user management not yet implemented",
	})
}

// GetAdminUser handles retrieving admin users.
func (h *SaaSAdminHandler) GetAdminUser(c *gin.Context) {
	h.logger.Info("Getting admin user")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Admin user management not yet implemented",
	})
}

// UpdateAdminUser handles updating admin users.
func (h *SaaSAdminHandler) UpdateAdminUser(c *gin.Context) {
	h.logger.Info("Updating admin user")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Admin user management not yet implemented",
	})
}

// DeleteAdminUser handles deleting admin users.
func (h *SaaSAdminHandler) DeleteAdminUser(c *gin.Context) {
	h.logger.Info("Deleting admin user")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Admin user management not yet implemented",
	})
}

// ResetAdminPassword handles admin password reset requests
func (h *SaaSAdminHandler) ResetAdminPassword(c *gin.Context) {
	adminIDStr := c.Param("id")
	if adminIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin ID is required"})
		return
	}

	adminID, err := strconv.ParseUint(adminIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin ID"})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or password too short (min 8 chars)"})
		return
	}

	if err := h.service.ResetAdminPassword(c.Request.Context(), uint(adminID), req.NewPassword); err != nil {
		h.logger.Error("Failed to reset password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password reset successfully",
	})
}

// ListNotifications handles listing notifications.
func (h *SaaSAdminHandler) ListNotifications(c *gin.Context) {
	h.logger.Info("Listing notifications")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Notification management not yet implemented",
	})
}

// CreateNotification handles creating notifications.
func (h *SaaSAdminHandler) CreateNotification(c *gin.Context) {
	h.logger.Info("Creating notification")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Notification management not yet implemented",
	})
}

// GetNotification handles retrieving a notification.
func (h *SaaSAdminHandler) GetNotification(c *gin.Context) {
	h.logger.Info("Getting notification")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Notification management not yet implemented",
	})
}

// UpdateNotification handles updating notifications.
func (h *SaaSAdminHandler) UpdateNotification(c *gin.Context) {
	h.logger.Info("Updating notification")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Notification management not yet implemented",
	})
}

// DeleteNotification handles deleting notifications.
func (h *SaaSAdminHandler) DeleteNotification(c *gin.Context) {
	h.logger.Info("Deleting notification")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Notification management not yet implemented",
	})
}

// ListActivities handles listing activities.
func (h *SaaSAdminHandler) ListActivities(c *gin.Context) {
	h.logger.Info("Listing activities")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Activity management not yet implemented",
	})
}

// ListBackups handles listing backups.
func (h *SaaSAdminHandler) ListBackups(c *gin.Context) {
	h.logger.Info("Listing backups")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup management not yet implemented",
	})
}

// CreateBackup handles creating backups.
func (h *SaaSAdminHandler) CreateBackup(c *gin.Context) {
	h.logger.Info("Creating backup")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup management not yet implemented",
	})
}

// GetBackup handles retrieving a backup.
func (h *SaaSAdminHandler) GetBackup(c *gin.Context) {
	h.logger.Info("Getting backup")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup management not yet implemented",
	})
}

// DeleteBackup handles deleting backups.
func (h *SaaSAdminHandler) DeleteBackup(c *gin.Context) {
	h.logger.Info("Deleting backup")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup management not yet implemented",
	})
}

// GetPricingFeatures handles retrieving pricing features.
func (h *SaaSAdminHandler) GetPricingFeatures(c *gin.Context) {
	h.logger.Info("Getting pricing features")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing feature management not yet implemented",
	})
}

// CreatePricingFeature handles creating pricing features.
func (h *SaaSAdminHandler) CreatePricingFeature(c *gin.Context) {
	h.logger.Info("Creating pricing feature")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing feature management not yet implemented",
	})
}

// UpdatePricingFeature handles updating pricing features.
func (h *SaaSAdminHandler) UpdatePricingFeature(c *gin.Context) {
	h.logger.Info("Updating pricing feature")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing feature management not yet implemented",
	})
}

// DeletePricingFeature handles deleting pricing features.
func (h *SaaSAdminHandler) DeletePricingFeature(c *gin.Context) {
	h.logger.Info("Deleting pricing feature")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing feature management not yet implemented",
	})
}

// GetPricingTiers handles retrieving pricing tiers.
func (h *SaaSAdminHandler) GetPricingTiers(c *gin.Context) {
	h.logger.Info("Getting pricing tiers")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing tier management not yet implemented",
	})
}

// CreatePricingTier handles creating pricing tiers.
func (h *SaaSAdminHandler) CreatePricingTier(c *gin.Context) {
	h.logger.Info("Creating pricing tier")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing tier management not yet implemented",
	})
}

// UpdatePricingTier handles updating pricing tiers.
func (h *SaaSAdminHandler) UpdatePricingTier(c *gin.Context) {
	h.logger.Info("Updating pricing tier")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing tier management not yet implemented",
	})
}

// DeletePricingTier handles deleting pricing tiers.
func (h *SaaSAdminHandler) DeletePricingTier(c *gin.Context) {
	h.logger.Info("Deleting pricing tier")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing tier management not yet implemented",
	})
}

// GetPlanFeatures handles retrieving plan features.
func (h *SaaSAdminHandler) GetPlanFeatures(c *gin.Context) {
	h.logger.Info("Getting plan features")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Plan feature management not yet implemented",
	})
}

// AssignFeatureToPlan handles assigning features to plans.
func (h *SaaSAdminHandler) AssignFeatureToPlan(c *gin.Context) {
	h.logger.Info("Assigning feature to plan")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Plan feature management not yet implemented",
	})
}

// RemoveFeatureFromPlan handles removing features from plans.
func (h *SaaSAdminHandler) RemoveFeatureFromPlan(c *gin.Context) {
	h.logger.Info("Removing feature from plan")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Plan feature management not yet implemented",
	})
}

// GetPublicPricingPlans handles retrieving public pricing plans.
func (h *SaaSAdminHandler) GetPublicPricingPlans(c *gin.Context) {
	h.logger.Info("Getting public pricing plans")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Public pricing plans not yet implemented",
	})
}

// SyncPricingToLandingPage handles syncing pricing to landing page.
func (h *SaaSAdminHandler) SyncPricingToLandingPage(c *gin.Context) {
	h.logger.Info("Syncing pricing to landing page")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Pricing sync not yet implemented",
	})
}

// GetAnalyticsOverview handles retrieving analytics overview.
func (h *SaaSAdminHandler) GetAnalyticsOverview(c *gin.Context) {
	h.logger.Info("Getting analytics overview")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Analytics management not yet implemented",
	})
}

// GetAnalyticsMetrics handles retrieving analytics metrics.
func (h *SaaSAdminHandler) GetAnalyticsMetrics(c *gin.Context) {
	h.logger.Info("Getting analytics metrics")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Analytics management not yet implemented",
	})
}

// CreateAnalyticsMetric handles creating analytics metrics.
func (h *SaaSAdminHandler) CreateAnalyticsMetric(c *gin.Context) {
	h.logger.Info("Creating analytics metric")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Analytics management not yet implemented",
	})
}

// CheckAnalyticsHealth handles checking analytics health.
func (h *SaaSAdminHandler) CheckAnalyticsHealth(c *gin.Context) {
	h.logger.Info("Checking analytics health")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Analytics management not yet implemented",
	})
}

// GetMonitors handles retrieving monitors.
func (h *SaaSAdminHandler) GetMonitors(c *gin.Context) {
	h.logger.Info("Getting monitors")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// CreateMonitor handles creating monitors.
func (h *SaaSAdminHandler) CreateMonitor(c *gin.Context) {
	h.logger.Info("Creating monitor")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// UpdateMonitor handles updating monitors.
func (h *SaaSAdminHandler) UpdateMonitor(c *gin.Context) {
	h.logger.Info("Updating monitor")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// DeleteMonitor handles deleting monitors.
func (h *SaaSAdminHandler) DeleteMonitor(c *gin.Context) {
	h.logger.Info("Deleting monitor")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// PauseMonitor handles pausing monitors.
func (h *SaaSAdminHandler) PauseMonitor(c *gin.Context) {
	h.logger.Info("Pausing monitor")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// ResumeMonitor handles resuming monitors.
func (h *SaaSAdminHandler) ResumeMonitor(c *gin.Context) {
	h.logger.Info("Resuming monitor")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// PauseAllMonitors handles pausing all monitors.
func (h *SaaSAdminHandler) PauseAllMonitors(c *gin.Context) {
	h.logger.Info("Pausing all monitors")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// ResumeAllMonitors handles resuming all monitors.
func (h *SaaSAdminHandler) ResumeAllMonitors(c *gin.Context) {
	h.logger.Info("Resuming all monitors")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Monitor management not yet implemented",
	})
}

// GetIntegrations handles retrieving integrations.
func (h *SaaSAdminHandler) GetIntegrations(c *gin.Context) {
	h.logger.Info("Getting integrations")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Integration management not yet implemented",
	})
}

// CreateIntegration handles creating integrations.
func (h *SaaSAdminHandler) CreateIntegration(c *gin.Context) {
	h.logger.Info("Creating integration")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Integration management not yet implemented",
	})
}

// UpdateIntegration handles updating integrations.
func (h *SaaSAdminHandler) UpdateIntegration(c *gin.Context) {
	h.logger.Info("Updating integration")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Integration management not yet implemented",
	})
}

// DeleteIntegration handles deleting integrations.
func (h *SaaSAdminHandler) DeleteIntegration(c *gin.Context) {
	h.logger.Info("Deleting integration")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Integration management not yet implemented",
	})
}

// TestIntegration handles testing integrations.
func (h *SaaSAdminHandler) TestIntegration(c *gin.Context) {
	h.logger.Info("Testing integration")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Integration management not yet implemented",
	})
}

// GetWebhooks handles retrieving webhooks.
func (h *SaaSAdminHandler) GetWebhooks(c *gin.Context) {
	h.logger.Info("Getting webhooks")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Webhook management not yet implemented",
	})
}

// CreateWebhook handles creating webhooks.
func (h *SaaSAdminHandler) CreateWebhook(c *gin.Context) {
	h.logger.Info("Creating webhook")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Webhook management not yet implemented",
	})
}

// UpdateWebhook handles updating webhooks.
func (h *SaaSAdminHandler) UpdateWebhook(c *gin.Context) {
	h.logger.Info("Updating webhook")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Webhook management not yet implemented",
	})
}

// DeleteWebhook handles deleting webhooks.
func (h *SaaSAdminHandler) DeleteWebhook(c *gin.Context) {
	h.logger.Info("Deleting webhook")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Webhook management not yet implemented",
	})
}

// TestWebhook handles testing webhooks.
func (h *SaaSAdminHandler) TestWebhook(c *gin.Context) {
	h.logger.Info("Testing webhook")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Webhook management not yet implemented",
	})
}

// GetSubscribers handles retrieving subscribers.
func (h *SaaSAdminHandler) GetSubscribers(c *gin.Context) {
	h.logger.Info("Getting subscribers")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Subscriber management not yet implemented",
	})
}

// CreateSubscriber handles creating subscribers.
func (h *SaaSAdminHandler) CreateSubscriber(c *gin.Context) {
	h.logger.Info("Creating subscriber")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Subscriber management not yet implemented",
	})
}

// DeleteSubscriber handles deleting subscribers.
func (h *SaaSAdminHandler) DeleteSubscriber(c *gin.Context) {
	h.logger.Info("Deleting subscriber")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Subscriber management not yet implemented",
	})
}

// GetRealTimeStats handles retrieving real-time stats.
func (h *SaaSAdminHandler) GetRealTimeStats(c *gin.Context) {
	h.logger.Info("Getting real-time stats")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Real-time stats not yet implemented",
	})
}

// WebSocketMonitoring handles WebSocket monitoring.
func (h *SaaSAdminHandler) WebSocketMonitoring(c *gin.Context) {
	h.logger.Info("WebSocket monitoring")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WebSocket monitoring not yet implemented",
	})
}

// Proxy Methods - These methods proxy requests to appropriate services

// Tenant Management - These methods handle tenant lifecycle management

// TenantCreateRequest represents the tenant creation request
type TenantCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	ContactEmail  string `json:"contact_email" binding:"required"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=8"`
	Domain        string `json:"domain"`
	Subdomain     string `json:"subdomain"`
	PlanID        string `json:"plan_id"`
}

// TenantResponse represents the tenant response
type TenantResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GetTenants handles tenant listing
func (h *SaaSAdminHandler) GetTenants(c *gin.Context) {
	h.logger.Info("Listing tenants")

	tenants, err := h.service.ListTenants(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tenants",
		})
		return
	}

	// Convert to response format expected by UI
	tenantList := make([]gin.H, 0, len(tenants))
	for _, tenant := range tenants {
		tenantData := gin.H{
			"id":            tenant.ID,
			"name":          tenant.Name,
			"slug":          tenant.Slug,
			"domain":        tenant.Domain,
			"subdomain":     tenant.Subdomain,
			"contact_email": tenant.ContactEmail,
			"billing_email": tenant.BillingEmail,
			"plan_id":       tenant.PlanID,
			"status":        tenant.Status,
			"is_active":     tenant.IsActive,
			"created_at":    tenant.CreatedAt,
		}

		// Add max_users if it's set (pointer field - nil means unlimited)
		if tenant.MaxUsers != nil {
			tenantData["max_users"] = *tenant.MaxUsers
		} else {
			tenantData["max_users"] = nil
		}

		// Get user count from tenant-admin service (placeholder - will be 0 for now)
		// TODO: Call tenant-admin service GET /api/v1/users/{tenant_id}/stats to get actual user count
		tenantData["user_count"] = 0

		tenantList = append(tenantList, tenantData)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tenantList,
		"count":  len(tenantList),
	})
}

// CreateTenant handles tenant creation
func (h *SaaSAdminHandler) CreateTenant(c *gin.Context) {
	var req TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid tenant creation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Validate tenant name
	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tenant name is required",
		})
		return
	}

	// Validate contact email
	if strings.TrimSpace(req.ContactEmail) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Contact email is required",
		})
		return
	}

	h.logger.Info("Creating tenant", zap.String("name", req.Name))

	// Get or use default plan
	var planID uuid.UUID
	if req.PlanID != "" {
		parsedPlanID, err := uuid.Parse(req.PlanID)
		if err != nil {
			h.logger.Error("Invalid plan ID", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid plan ID",
			})
			return
		}
		planID = parsedPlanID
	} else {
		// Try to get the default plan
		plans, err := h.service.ListPlans(c.Request.Context(), 1, 0)
		if err == nil && len(plans) > 0 {
			planID = plans[0].ID
		} else {
			// If no plans exist, use a zero UUID (will be set later)
			planID = uuid.Nil
			h.logger.Warn("No plans available, tenant will need plan assignment later")
		}
	}

	// Create tenant model
	tenant := &models.SaaSTenant{
		Name:         req.Name,
		ContactEmail: req.ContactEmail,
		Domain:       req.Domain,
		Subdomain:    req.Subdomain,
		PlanID:       planID,
		Status:       "active",
		IsActive:     true,
	}

	// Create the tenant in database
	if err := h.service.CreateTenant(c.Request.Context(), tenant); err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create tenant",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("name", tenant.Name))

	// Publish tenant created event to RabbitMQ
	if h.eventPublisher != nil {
		metadata := events.EventMetadata{
			Source:        "saas-admin-service",
			CorrelationID: c.GetString("X-Correlation-ID"),
			UserID:        c.GetString("user_id"),
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		event := events.NewTenantCreatedEvent(tenant, metadata)
		if err := h.eventPublisher.PublishTenantEvent(c.Request.Context(), event); err != nil {
			h.logger.Error("Failed to publish tenant created event",
				zap.Error(err),
				zap.String("tenant_id", tenant.ID.String()))
			// Log but don't fail - tenant is already created in saas_admin
			// The tenant-admin service will need to be synced manually or via retry mechanism
		} else {
			h.logger.Info("Published tenant created event",
				zap.String("event_id", event.EventID),
				zap.String("tenant_id", tenant.ID.String()))
		}
	} else {
		h.logger.Warn("Event publisher not configured, falling back to HTTP sync")
		// Fallback to old HTTP method if event publisher is not available
		if err := h.createTenantInTenantAdminService(tenant, req.AdminEmail, req.AdminPassword); err != nil {
			h.logger.Error("Failed to create tenant in tenant-admin service",
				zap.Error(err),
				zap.String("tenant_id", tenant.ID.String()))
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"message": "Tenant created successfully",
		"data": gin.H{
			"id":            tenant.ID,
			"name":          tenant.Name,
			"domain":        tenant.Domain,
			"plan":          req.PlanID,
			"status":        tenant.Status,
		},
	})
}

// GetTenant handles single tenant retrieval
func (h *SaaSAdminHandler) GetTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tenant ID is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Getting tenant", zap.String("id", tenantIDStr))

	tenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tenant not found",
		})
		return
	}

	// Build comprehensive tenant response including all editable fields
	tenantData := gin.H{
		"id":            tenant.ID,
		"name":          tenant.Name,
		"slug":          tenant.Slug,
		"domain":        tenant.Domain,
		"subdomain":     tenant.Subdomain,
		"contact_email": tenant.ContactEmail,
		"billing_email": tenant.BillingEmail,
		"plan_id":       tenant.PlanID,
		"status":        tenant.Status,
		"is_active":     tenant.IsActive,
		"created_at":    tenant.CreatedAt,
		"settings":      tenant.Settings,
		"branding":      tenant.Branding,
		"features":      tenant.Features,
		"metadata":      tenant.Metadata,
	}

	// Add max_users (nullable pointer field)
	if tenant.MaxUsers != nil {
		tenantData["max_users"] = *tenant.MaxUsers
	} else {
		tenantData["max_users"] = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tenantData,
	})
}

// UpdateTenant handles tenant updates
func (h *SaaSAdminHandler) UpdateTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tenant ID is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	// Get current tenant data BEFORE update (needed for credential sync)
	oldTenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant for update", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tenant not found",
		})
		return
	}

	// Parse request body - includes optional password field not in model
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Extract password if provided (not stored in saas_admin database)
	newPassword, _ := requestBody["password"].(string)

	// Convert map back to SaaSTenant for service layer (exclude password)
	delete(requestBody, "password")
	updatesJSON, _ := json.Marshal(requestBody)
	var updates models.SaaSTenant
	json.Unmarshal(updatesJSON, &updates)

	h.logger.Info("Updating tenant", zap.String("id", tenantIDStr))

	// Update tenant in saas_admin database
	if err := h.service.UpdateTenant(c.Request.Context(), tenantID, &updates); err != nil {
		h.logger.Error("Failed to update tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update tenant",
		})
		return
	}

	// Get updated tenant data for event publishing
	updatedTenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get updated tenant", zap.Error(err))
		// Continue anyway, we'll use old data
		updatedTenant = oldTenant
	}

	// Publish tenant updated event to RabbitMQ
	if h.eventPublisher != nil {
		metadata := events.EventMetadata{
			Source:        "saas-admin-service",
			CorrelationID: c.GetString("X-Correlation-ID"),
			UserID:        c.GetString("user_id"),
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		event := events.NewTenantUpdatedEvent(updatedTenant, metadata)
		if err := h.eventPublisher.PublishTenantEvent(c.Request.Context(), event); err != nil {
			h.logger.Error("Failed to publish tenant updated event",
				zap.Error(err),
				zap.String("tenant_id", updatedTenant.ID.String()))
		} else {
			h.logger.Info("Published tenant updated event",
				zap.String("event_id", event.EventID),
				zap.String("tenant_id", updatedTenant.ID.String()))
		}
	} else {
		h.logger.Warn("Event publisher not configured, falling back to direct sync")

		// Fallback: Sync credentials to tenant-admin service if contact_email or password changed
		newEmail := updates.ContactEmail
		if newEmail == "" {
			newEmail = oldTenant.ContactEmail // No change
		}

		if newEmail != oldTenant.ContactEmail || newPassword != "" {
			h.logger.Info("Syncing credential changes directly to tenant_admin_db",
				zap.String("tenant_id", tenantIDStr),
				zap.Bool("email_changed", newEmail != oldTenant.ContactEmail),
				zap.Bool("password_changed", newPassword != ""))

			// Update credentials directly in tenant_admin_db (atomic operation)
			if err := h.updateTenantCredentialsDirect(
				tenantID,                // UUID (not string)
				oldTenant.ContactEmail, // old email
				newEmail,                // new email
				newPassword,             // new password (empty string if not changed)
			); err != nil {
				// Log error but don't fail the entire update
				h.logger.Error("Failed to sync credentials to tenant_admin_db",
					zap.Error(err),
					zap.String("tenant_id", tenantIDStr))
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tenant updated successfully",
	})
}

// DeleteTenant handles tenant deletion
func (h *SaaSAdminHandler) DeleteTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tenant ID is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Deleting tenant", zap.String("id", tenantIDStr))

	// Get tenant details before deletion (need slug for tenant-admin deletion)
	tenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant for deletion", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tenant not found",
		})
		return
	}

	// Delete from saas_admin database
	if err := h.service.DeleteTenant(c.Request.Context(), tenantID); err != nil {
		h.logger.Error("Failed to delete tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete tenant",
		})
		return
	}

	// Publish tenant deleted event to RabbitMQ
	if h.eventPublisher != nil {
		metadata := events.EventMetadata{
			Source:        "saas-admin-service",
			CorrelationID: c.GetString("X-Correlation-ID"),
			UserID:        c.GetString("user_id"),
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		event := events.NewTenantDeletedEvent(tenant, metadata)
		if err := h.eventPublisher.PublishTenantEvent(c.Request.Context(), event); err != nil {
			h.logger.Error("Failed to publish tenant deleted event",
				zap.Error(err),
				zap.String("tenant_id", tenant.ID.String()))
		} else {
			h.logger.Info("Published tenant deleted event",
				zap.String("event_id", event.EventID),
				zap.String("tenant_id", tenant.ID.String()))
		}
	} else {
		h.logger.Warn("Event publisher not configured, falling back to direct deletion")

		// Fallback: Delete from tenant-admin service (cascade deletion)
		if err := h.deleteTenantFromTenantAdminService(tenant.Slug); err != nil {
			h.logger.Error("Failed to delete tenant from tenant-admin service",
				zap.Error(err),
				zap.String("tenant_id", tenantID.String()),
				zap.String("slug", tenant.Slug))
			// Log but don't fail - tenant is already deleted from saas_admin
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tenant deleted successfully",
	})
}

// GetArchivedTenants lists all soft-deleted tenants
func (h *SaaSAdminHandler) GetArchivedTenants(c *gin.Context) {
	h.logger.Info("Listing archived tenants")

	archivedTenants, err := h.service.ListArchivedTenants(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list archived tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get archived tenants",
		})
		return
	}

	// Convert to response format expected by UI
	tenantList := make([]gin.H, 0, len(archivedTenants))
	for _, tenant := range archivedTenants {
		tenantData := gin.H{
			"id":            tenant.ID,
			"name":          tenant.Name,
			"slug":          tenant.Slug,
			"domain":        tenant.Domain,
			"subdomain":     tenant.Subdomain,
			"contact_email": tenant.ContactEmail,
			"billing_email": tenant.BillingEmail,
			"plan_id":       tenant.PlanID,
			"status":        tenant.Status,
			"is_active":     tenant.IsActive,
			"created_at":    tenant.CreatedAt,
			"deleted_at":    tenant.DeletedAt,
		}

		// Add max_users if it's set (pointer field - nil means unlimited)
		if tenant.MaxUsers != nil {
			tenantData["max_users"] = *tenant.MaxUsers
		} else {
			tenantData["max_users"] = nil
		}

		tenantList = append(tenantList, tenantData)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tenantList,
		"count":  len(tenantList),
	})
}

// RestoreTenant restores a soft-deleted tenant
func (h *SaaSAdminHandler) RestoreTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tenant ID is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tenant ID",
		})
		return
	}

	h.logger.Info("Restoring tenant", zap.String("id", tenantIDStr))

	// Restore tenant in saas_admin database
	if err := h.service.RestoreTenant(c.Request.Context(), tenantID); err != nil {
		h.logger.Error("Failed to restore tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restore tenant",
		})
		return
	}

	// Get restored tenant data for event publishing
	restoredTenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get restored tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Tenant restored but failed to get data",
		})
		return
	}

	// Publish tenant restored event to RabbitMQ
	if h.eventPublisher != nil {
		metadata := events.EventMetadata{
			Source:        "saas-admin-service",
			CorrelationID: c.GetString("X-Correlation-ID"),
			UserID:        c.GetString("user_id"),
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		event := events.NewTenantRestoredEvent(restoredTenant, metadata)
		if err := h.eventPublisher.PublishTenantEvent(c.Request.Context(), event); err != nil {
			h.logger.Error("Failed to publish tenant restored event",
				zap.Error(err),
				zap.String("tenant_id", restoredTenant.ID.String()))
		} else {
			h.logger.Info("Published tenant restored event",
				zap.String("event_id", event.EventID),
				zap.String("tenant_id", restoredTenant.ID.String()))
		}
	} else {
		h.logger.Warn("Event publisher not configured, tenant restore will not sync to tenant-admin")
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tenant restored successfully",
	})
}

// GetComponents proxies component listing to component-service
func (h *SaaSAdminHandler) GetComponents(c *gin.Context) {
	h.proxyRequest(c, "GET", h.serviceURLs.ComponentService+"/api/v1/components", nil)
}

// CreateComponent proxies component creation to component-service
func (h *SaaSAdminHandler) CreateComponent(c *gin.Context) {
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", h.serviceURLs.ComponentService+"/api/v1/components", requestBody)
}

// GetComponent proxies single component retrieval to component-service
func (h *SaaSAdminHandler) GetComponent(c *gin.Context) {
	componentID := c.Param("id")
	h.proxyRequest(c, "GET", h.serviceURLs.ComponentService+"/api/v1/components/"+componentID, nil)
}

// UpdateComponent proxies component updates to component-service
func (h *SaaSAdminHandler) UpdateComponent(c *gin.Context) {
	componentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "PUT", h.serviceURLs.ComponentService+"/api/v1/components/"+componentID, requestBody)
}

// DeleteComponent proxies component deletion to component-service
func (h *SaaSAdminHandler) DeleteComponent(c *gin.Context) {
	componentID := c.Param("id")
	h.proxyRequest(c, "DELETE", h.serviceURLs.ComponentService+"/api/v1/components/"+componentID, nil)
}

// UpdateComponentStatus proxies component status updates to component-service
func (h *SaaSAdminHandler) UpdateComponentStatus(c *gin.Context) {
	componentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", h.serviceURLs.ComponentService+"/api/v1/components/"+componentID+"/status", requestBody)
}

// GetComponentUptimeStats proxies component uptime stats to component-service
func (h *SaaSAdminHandler) GetComponentUptimeStats(c *gin.Context) {
	componentID := c.Param("id")
	h.proxyRequest(c, "GET", h.serviceURLs.ComponentService+"/api/v1/components/"+componentID+"/uptime", nil)
}

// GetComponentGroups proxies component groups to component-service
func (h *SaaSAdminHandler) GetComponentGroups(c *gin.Context) {
	h.proxyRequest(c, "GET", h.serviceURLs.ComponentService+"/api/v1/component-groups", nil)
}

// GetIncidents proxies incident listing to incident-service
func (h *SaaSAdminHandler) GetIncidents(c *gin.Context) {
	h.proxyRequest(c, "GET", h.serviceURLs.IncidentService+"/api/v1/incidents", nil)
}

// CreateIncident proxies incident creation to incident-service
func (h *SaaSAdminHandler) CreateIncident(c *gin.Context) {
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", h.serviceURLs.IncidentService+"/api/v1/incidents", requestBody)
}

// GetIncident proxies single incident retrieval to incident-service
func (h *SaaSAdminHandler) GetIncident(c *gin.Context) {
	incidentID := c.Param("id")
	h.proxyRequest(c, "GET", h.serviceURLs.IncidentService+"/api/v1/incidents/"+incidentID, nil)
}

// UpdateIncident proxies incident updates to incident-service
func (h *SaaSAdminHandler) UpdateIncident(c *gin.Context) {
	incidentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "PUT", h.serviceURLs.IncidentService+"/api/v1/incidents/"+incidentID, requestBody)
}

// DeleteIncident proxies incident deletion to incident-service
func (h *SaaSAdminHandler) DeleteIncident(c *gin.Context) {
	incidentID := c.Param("id")
	h.proxyRequest(c, "DELETE", h.serviceURLs.IncidentService+"/api/v1/incidents/"+incidentID, nil)
}

// CreateIncidentUpdate proxies incident update creation to incident-service
func (h *SaaSAdminHandler) CreateIncidentUpdate(c *gin.Context) {
	incidentID := c.Param("id")
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
	}
	h.proxyRequest(c, "POST", h.serviceURLs.IncidentService+"/api/v1/incidents/"+incidentID+"/updates", requestBody)
}

// GetIncidentUpdates proxies incident updates retrieval to incident-service
func (h *SaaSAdminHandler) GetIncidentUpdates(c *gin.Context) {
	incidentID := c.Param("id")
	h.proxyRequest(c, "GET", h.serviceURLs.IncidentService+"/api/v1/incidents/"+incidentID+"/updates", nil)
}

// GetIncidentStats proxies incident statistics to incident-service
func (h *SaaSAdminHandler) GetIncidentStats(c *gin.Context) {
	h.proxyRequest(c, "GET", h.serviceURLs.IncidentService+"/api/v1/incidents/stats", nil)
}

// Generic proxy method
func (h *SaaSAdminHandler) proxyRequest(c *gin.Context, method, targetURL string, body []byte) {
	// Create request
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, targetURL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, targetURL, nil)
	}

	if err != nil {
		h.logger.Error("Failed to create proxy request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create proxy request",
		})
		return
	}

	// Copy headers from original request
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add service-to-service authorization
	req.Header.Set("Authorization", "Bearer service-token-"+os.Getenv("JWT_SECRET"))

	// Try to get session ID from cookie first, fallback to service session
	sessionID := ""
	if cookie, err := c.Cookie("session_id"); err == nil && cookie != "" {
		sessionID = cookie
	} else {
		sessionID = "saas-admin-session-" + os.Getenv("JWT_SECRET")
	}

	// Add session ID for downstream services that require it
	req.Header.Set("Session-ID", sessionID)

	// Add query parameters
	req.URL.RawQuery = c.Request.URL.RawQuery

	// Make the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Proxy request failed",
			zap.String("target", targetURL),
			zap.String("method", method),
			zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Downstream service unavailable",
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read proxy response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	// Return response with same status code
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)
}

// Authentication Methods

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles admin user authentication
func (h *SaaSAdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	h.logger.Info("Login attempt", zap.String("username", req.Username))

	// For demo purposes - in production, validate against user database
	// Default credentials: admin/password
	if req.Username == "admin" && req.Password == "password" {
		// Generate JWT token for React frontend
		token, err := h.jwtManager.GenerateAccessToken(1, "admin", "admin@beakon.io")
		if err != nil {
			h.logger.Error("Failed to generate JWT", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
			return
		}

		h.logger.Info("Login successful", zap.String("username", req.Username))

		// Return JWT token for React frontend
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   token,
			"user": gin.H{
				"id":       1,
				"username": "admin",
				"email":    "admin@beakon.io",
				"role":     "super_admin",
			},
			"message": "Login successful",
		})
	} else {
		h.logger.Warn("Login failed - invalid credentials", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

// Logout handles admin user logout
func (h *SaaSAdminHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No session found"})
		return
	}

	// Call tenant-admin service to invalidate session
	req, err := http.NewRequest("DELETE", h.serviceURLs.TenantAdminService+"/api/v1/sessions/"+sessionID, nil)
	if err != nil {
		h.logger.Error("Failed to create logout request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	req.Header.Set("Authorization", "Bearer service-token-"+os.Getenv("JWT_SECRET"))

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to invalidate session", zap.Error(err))
	} else {
		defer resp.Body.Close()
	}

	// Clear session cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout successful"})
}

// CheckAuth checks if the current request is authenticated
func (h *SaaSAdminHandler) CheckAuth(c *gin.Context) {
	// Get Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "No authorization header"})
		return
	}

	// Extract token from "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "Invalid authorization format"})
		return
	}

	// Validate JWT token
	claims, err := h.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		h.logger.Warn("Invalid JWT token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "Invalid token"})
		return
	}

	// Return authenticated user info
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       claims.UserID,
			"username": claims.Username,
			"email":    claims.Email,
		},
	})
}

// createTenantInTenantAdminService creates a tenant in the tenant-admin service via API call
func (h *SaaSAdminHandler) createTenantInTenantAdminService(tenant *models.SaaSTenant, adminEmail, adminPassword string) error {
	// Prepare request payload for tenant-admin service
	payload := map[string]interface{}{
		"slug":           tenant.Slug,
		"name":           tenant.Name,
		"contact_email":  tenant.ContactEmail,
		"domain":         tenant.Domain,
		"subdomain":      tenant.Subdomain,
		"admin_email":    adminEmail,
		"admin_password": adminPassword,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Make POST request to tenant-admin service public API
	url := h.tenantAdminServiceURL + "/api/v1/public/tenants"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tenant-admin service returned status %d: %s", resp.StatusCode, string(body))
	}

	h.logger.Info("Successfully created tenant in tenant-admin service",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("slug", tenant.Slug))

	return nil
}

// deleteTenantFromTenantAdminService deletes a tenant from tenant-admin service via API call
func (h *SaaSAdminHandler) deleteTenantFromTenantAdminService(slug string) error {
	// First, get tenant by slug to get the ID
	url := h.tenantAdminServiceURL + "/api/v1/public/tenants/slug/" + slug
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Tenant doesn't exist in tenant-admin, nothing to delete
		h.logger.Info("Tenant not found in tenant-admin service, skipping deletion",
			zap.String("slug", slug))
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get tenant from tenant-admin: status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to get tenant ID
	var tenantResp struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tenantResp); err != nil {
		return fmt.Errorf("failed to decode tenant response: %w", err)
	}

	// Now delete the tenant by ID
	deleteURL := fmt.Sprintf("%s/api/v1/public/tenants/%d", h.tenantAdminServiceURL, tenantResp.Data.ID)
	deleteReq, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return err
	}

	deleteResp, err := h.httpClient.Do(deleteReq)
	if err != nil {
		return err
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(deleteResp.Body)
		return fmt.Errorf("tenant-admin delete returned status %d: %s", deleteResp.StatusCode, string(body))
	}

	h.logger.Info("Successfully deleted tenant from tenant-admin service",
		zap.String("slug", slug),
		zap.Uint("tenant_admin_id", tenantResp.Data.ID))

	return nil
}

// updateTenantCredentialsDirect updates tenant owner credentials directly in tenant_admin_db
// This ensures atomic updates without HTTP calls or sync issues.
// Note: This is an exception to the database-per-service pattern for critical credential sync.
func (h *SaaSAdminHandler) updateTenantCredentialsDirect(tenantID uuid.UUID, oldEmail, newEmail, newPassword string) error {
	if h.tenantAdminDB == nil {
		h.logger.Warn("Tenant admin DB connection not available, skipping credential sync")
		return nil
	}

	// Start transaction on tenant_admin_db
	tx := h.tenantAdminDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			h.logger.Error("Panic during credential update, rolling back",
				zap.Any("panic", r))
		}
	}()

	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Find owner user by tenant_id and old email
	var user models.TenantAdminUser
	result := tx.Where("tenant_id = ? AND email = ? AND role = ?",
		tenantID, oldEmail, "owner").First(&user)

	if result.Error != nil {
		tx.Rollback()
		if result.Error == gorm.ErrRecordNotFound {
			h.logger.Warn("Tenant owner not found for credential update",
				zap.String("tenant_id", tenantID.String()),
				zap.String("old_email", oldEmail))
			return fmt.Errorf("tenant owner not found with email %s", oldEmail)
		}
		h.logger.Error("Failed to find tenant owner",
			zap.String("tenant_id", tenantID.String()),
			zap.String("old_email", oldEmail),
			zap.Error(result.Error))
		return fmt.Errorf("database error finding owner: %w", result.Error)
	}

	// Prepare updates
	updates := make(map[string]interface{})

	// Update email if changed
	if newEmail != "" && newEmail != oldEmail {
		updates["email"] = newEmail
		h.logger.Info("Updating tenant owner email",
			zap.String("tenant_id", tenantID.String()),
			zap.String("old_email", oldEmail),
			zap.String("new_email", newEmail))
	}

	// Update password if provided
	if newPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			h.logger.Error("Failed to hash password", zap.Error(err))
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password_hash"] = string(hashedPassword)
		h.logger.Info("Updating tenant owner password",
			zap.String("tenant_id", tenantID.String()),
			zap.String("email", oldEmail))
	}

	// Apply updates if any
	if len(updates) == 0 {
		tx.Rollback()
		h.logger.Info("No credential updates to apply",
			zap.String("tenant_id", tenantID.String()))
		return nil
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to update tenant owner credentials",
			zap.String("tenant_id", tenantID.String()),
			zap.String("old_email", oldEmail),
			zap.Error(err))
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		h.logger.Error("Failed to commit credential update transaction",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	h.logger.Info("Successfully updated tenant owner credentials in database",
		zap.String("tenant_id", tenantID.String()),
		zap.String("old_email", oldEmail),
		zap.String("new_email", newEmail),
		zap.Bool("password_changed", newPassword != ""))

	return nil
}

// DEPRECATED: updateTenantAdminCredentials - HTTP-based credential sync (replaced by direct DB access)
// Keeping for reference but no longer used. Direct DB access via updateTenantCredentialsDirect is now preferred.
func (h *SaaSAdminHandler) updateTenantAdminCredentials(tenantID, oldEmail, newEmail, newPassword string) error {
	if h.tenantAdminServiceURL == "" {
		h.logger.Warn("Tenant admin service URL not configured, skipping credential sync")
		return nil
	}

	// Prepare request payload
	payload := map[string]string{
		"tenant_id":    tenantID,
		"old_email":    oldEmail,
		"new_email":    newEmail,
		"new_password": newPassword,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal credential update payload", zap.Error(err))
		return fmt.Errorf("failed to prepare credential update: %w", err)
	}

	// Call tenant-admin service UpdateCredentials endpoint
	updateURL := fmt.Sprintf("%s/api/v1/admin/update-credentials", h.tenantAdminServiceURL)
	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		h.logger.Error("Failed to create credential update request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to call tenant-admin credential update", zap.Error(err))
		return fmt.Errorf("failed to update credentials in tenant-admin service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		h.logger.Error("Tenant-admin credential update failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return fmt.Errorf("tenant-admin returned status %d: %s", resp.StatusCode, string(body))
	}

	h.logger.Info("Successfully synchronized credentials to tenant-admin service",
		zap.String("tenant_id", tenantID),
		zap.String("new_email", newEmail))

	return nil
}
