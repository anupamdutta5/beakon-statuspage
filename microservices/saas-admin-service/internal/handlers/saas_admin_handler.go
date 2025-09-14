// Package handlers provides HTTP handlers for the SaaS Admin Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/enterprise-status/statuspage-saas-admin-service/internal/models"
	"github.com/enterprise-status/statuspage-saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SaaSAdminHandler handles SaaS admin-related HTTP requests.
type SaaSAdminHandler struct {
	service *services.SaaSAdminService
	logger  *zap.Logger
}

// NewSaaSAdminHandler creates a new SaaS admin handler.
func NewSaaSAdminHandler(service *services.SaaSAdminService, logger *zap.Logger) *SaaSAdminHandler {
	return &SaaSAdminHandler{
		service: service,
		logger:  logger,
	}
}

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

// Platform Management Handlers

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

// Plan Management Handlers

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

	h.logger.Info("Creating SaaS plan", zap.String("plan_name", plan.Name))

	if err := h.service.CreatePlan(c.Request.Context(), &plan); err != nil {
		h.logger.Error("Failed to create plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create plan",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"plan":    plan,
		"message": "SaaS plan created successfully",
	})
}

// GetPlan handles retrieving a plan by ID.
func (h *SaaSAdminHandler) GetPlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	h.logger.Info("Getting SaaS plan", zap.Uint64("plan_id", planID))

	plan, err := h.service.GetPlan(c.Request.Context(), uint(planID))
	if err != nil {
		h.logger.Error("Failed to get plan", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Plan not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
	})
}

// GetPlanBySlug handles retrieving a plan by slug.
func (h *SaaSAdminHandler) GetPlanBySlug(c *gin.Context) {
	slug := c.Param("slug")

	h.logger.Info("Getting SaaS plan by slug", zap.String("slug", slug))

	plan, err := h.service.GetPlanBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error("Failed to get plan by slug", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Plan not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan": plan,
	})
}

// UpdatePlan handles updating a plan.
func (h *SaaSAdminHandler) UpdatePlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
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

	h.logger.Info("Updating SaaS plan", zap.Uint64("plan_id", planID))

	updatedPlan, err := h.service.UpdatePlan(c.Request.Context(), uint(planID), &updates)
	if err != nil {
		h.logger.Error("Failed to update plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan":    updatedPlan,
		"message": "SaaS plan updated successfully",
	})
}

// DeletePlan handles deleting a plan.
func (h *SaaSAdminHandler) DeletePlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID",
		})
		return
	}

	h.logger.Info("Deleting SaaS plan", zap.Uint64("plan_id", planID))

	if err := h.service.DeletePlan(c.Request.Context(), uint(planID)); err != nil {
		h.logger.Error("Failed to delete plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete plan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SaaS plan deleted successfully",
	})
}

// Feature Management Handlers

// ListFeatures handles listing SaaS features.
func (h *SaaSAdminHandler) ListFeatures(c *gin.Context) {
	h.logger.Info("Listing SaaS features")

	category := c.Query("category")

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

// CreateFeature handles creating a new SaaS feature.
func (h *SaaSAdminHandler) CreateFeature(c *gin.Context) {
	var feature models.SaaSFeature
	if err := c.ShouldBindJSON(&feature); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating SaaS feature", zap.String("feature_name", feature.Name))

	if err := h.service.CreateFeature(c.Request.Context(), &feature); err != nil {
		h.logger.Error("Failed to create feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feature",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"feature": feature,
		"message": "SaaS feature created successfully",
	})
}

// GetFeature handles retrieving a feature by ID.
func (h *SaaSAdminHandler) GetFeature(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := strconv.ParseUint(featureIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature ID",
		})
		return
	}

	h.logger.Info("Getting SaaS feature", zap.Uint64("feature_id", featureID))

	feature, err := h.service.GetFeature(c.Request.Context(), uint(featureID))
	if err != nil {
		h.logger.Error("Failed to get feature", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Feature not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature": feature,
	})
}

// UpdateFeature handles updating a feature.
func (h *SaaSAdminHandler) UpdateFeature(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := strconv.ParseUint(featureIDStr, 10, 32)
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

	h.logger.Info("Updating SaaS feature", zap.Uint64("feature_id", featureID))

	if err := h.service.UpdateFeature(c.Request.Context(), uint(featureID), &updates); err != nil {
		h.logger.Error("Failed to update feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SaaS feature updated successfully",
	})
}

// DeleteFeature handles deleting a feature.
func (h *SaaSAdminHandler) DeleteFeature(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := strconv.ParseUint(featureIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature ID",
		})
		return
	}

	h.logger.Info("Deleting SaaS feature", zap.Uint64("feature_id", featureID))

	if err := h.service.DeleteFeature(c.Request.Context(), uint(featureID)); err != nil {
		h.logger.Error("Failed to delete feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SaaS feature deleted successfully",
	})
}

// Feature Flag Management Handlers

// ListFeatureFlags handles listing feature flags.
func (h *SaaSAdminHandler) ListFeatureFlags(c *gin.Context) {
	h.logger.Info("Listing SaaS feature flags")

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

	h.logger.Info("Creating SaaS feature flag", zap.String("flag_name", flag.Name))

	if err := h.service.CreateFeatureFlag(c.Request.Context(), &flag); err != nil {
		h.logger.Error("Failed to create feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feature flag",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"feature_flag": flag,
		"message":      "SaaS feature flag created successfully",
	})
}

// GetFeatureFlag handles retrieving a feature flag by ID.
func (h *SaaSAdminHandler) GetFeatureFlag(c *gin.Context) {
	flagIDStr := c.Param("id")
	flagID, err := strconv.ParseUint(flagIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	h.logger.Info("Getting SaaS feature flag", zap.Uint64("flag_id", flagID))

	flag, err := h.service.GetFeatureFlag(c.Request.Context(), uint(flagID))
	if err != nil {
		h.logger.Error("Failed to get feature flag", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Feature flag not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flag": flag,
	})
}

// UpdateFeatureFlag handles updating a feature flag.
func (h *SaaSAdminHandler) UpdateFeatureFlag(c *gin.Context) {
	flagIDStr := c.Param("id")
	flagID, err := strconv.ParseUint(flagIDStr, 10, 32)
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

	h.logger.Info("Updating SaaS feature flag", zap.Uint64("flag_id", flagID))

	updatedFlag, err := h.service.UpdateFeatureFlag(c.Request.Context(), uint(flagID), &updates)
	if err != nil {
		h.logger.Error("Failed to update feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature_flag": updatedFlag,
		"message":      "SaaS feature flag updated successfully",
	})
}

// DeleteFeatureFlag handles deleting a feature flag.
func (h *SaaSAdminHandler) DeleteFeatureFlag(c *gin.Context) {
	flagIDStr := c.Param("id")
	flagID, err := strconv.ParseUint(flagIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature flag ID",
		})
		return
	}

	h.logger.Info("Deleting SaaS feature flag", zap.Uint64("flag_id", flagID))

	if err := h.service.DeleteFeatureFlag(c.Request.Context(), uint(flagID)); err != nil {
		h.logger.Error("Failed to delete feature flag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete feature flag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SaaS feature flag deleted successfully",
	})
}

// Statistics Handlers

// GetStats handles getting SaaS platform statistics.
func (h *SaaSAdminHandler) GetStats(c *gin.Context) {
	h.logger.Info("Getting SaaS platform statistics")

	stats, err := h.service.GetSaaSStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Placeholder handlers for remaining endpoints
func (h *SaaSAdminHandler) ListAdminUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) CreateAdminUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) GetAdminUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) UpdateAdminUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) DeleteAdminUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) ListNotifications(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) CreateNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) GetNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) UpdateNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) DeleteNotification(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) ListActivities(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) ListBackups(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) CreateBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) GetBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *SaaSAdminHandler) DeleteBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
