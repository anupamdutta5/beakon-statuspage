// Package handlers provides HTTP handlers for the SaaS Admin Service.
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/models"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	// Validate input
	if strings.TrimSpace(plan.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Plan name is required",
		})
		return
	}
	if strings.TrimSpace(plan.Slug) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Plan slug is required",
		})
		return
	}
	if plan.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Plan price cannot be negative",
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
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", planIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID format, expected UUID",
		})
		return
	}

	h.logger.Info("Getting SaaS plan", zap.String("plan_id", planID.String()))

	plan, err := h.service.GetPlan(c.Request.Context(), planID)
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
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", planIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID format, expected UUID",
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

	h.logger.Info("Updating SaaS plan", zap.String("plan_id", planID.String()))

	updatedPlan, err := h.service.UpdatePlan(c.Request.Context(), planID, &updates)
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
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", planIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID format, expected UUID",
		})
		return
	}

	h.logger.Info("Deleting SaaS plan", zap.String("plan_id", planID.String()))

	if err := h.service.DeletePlan(c.Request.Context(), planID); err != nil {
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

// Pricing Management Handlers

// CreatePricingFeature creates a new pricing feature.
func (h *SaaSAdminHandler) CreatePricingFeature(c *gin.Context) {
	h.logger.Info("Creating pricing feature")

	var feature models.PricingFeature
	if err := c.ShouldBindJSON(&feature); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.service.CreatePricingFeature(c.Request.Context(), &feature); err != nil {
		h.logger.Error("Failed to create pricing feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pricing feature", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing feature created successfully", zap.Uint("feature_id", feature.ID))
	c.JSON(http.StatusCreated, gin.H{
		"message": "Pricing feature created successfully",
		"feature": feature,
	})
}

// GetPricingFeatures retrieves all pricing features.
func (h *SaaSAdminHandler) GetPricingFeatures(c *gin.Context) {
	h.logger.Info("Getting pricing features")

	category := c.Query("category")
	features, err := h.service.GetPricingFeatures(c.Request.Context(), category)
	if err != nil {
		h.logger.Error("Failed to get pricing features", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pricing features", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing features retrieved successfully", zap.Int("count", len(features)))
	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"count":    len(features),
	})
}

// UpdatePricingFeature updates a pricing feature.
func (h *SaaSAdminHandler) UpdatePricingFeature(c *gin.Context) {
	h.logger.Info("Updating pricing feature")

	featureIDStr := c.Param("id")
	featureID, err := strconv.ParseUint(featureIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid feature ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	var updates models.PricingFeature
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	updatedFeature, err := h.service.UpdatePricingFeature(c.Request.Context(), uint(featureID), &updates)
	if err != nil {
		h.logger.Error("Failed to update pricing feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pricing feature", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing feature updated successfully", zap.Uint("feature_id", uint(featureID)))
	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing feature updated successfully",
		"feature": updatedFeature,
	})
}

// DeletePricingFeature deletes a pricing feature.
func (h *SaaSAdminHandler) DeletePricingFeature(c *gin.Context) {
	h.logger.Info("Deleting pricing feature")

	featureIDStr := c.Param("id")
	featureID, err := strconv.ParseUint(featureIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid feature ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	if err := h.service.DeletePricingFeature(c.Request.Context(), uint(featureID)); err != nil {
		h.logger.Error("Failed to delete pricing feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pricing feature", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing feature deleted successfully", zap.Uint("feature_id", uint(featureID)))
	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing feature deleted successfully",
	})
}

// CreatePricingTier creates a new pricing tier.
func (h *SaaSAdminHandler) CreatePricingTier(c *gin.Context) {
	h.logger.Info("Creating pricing tier")

	var request struct {
		models.PricingTier
		PlanID string `json:"plan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Parse plan ID as UUID
	planID, err := uuid.Parse(request.PlanID)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", request.PlanID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID format, expected UUID",
		})
		return
	}

	tier := request.PricingTier
	tier.PlanID = planID

	if err := h.service.CreatePricingTier(c.Request.Context(), &tier); err != nil {
		h.logger.Error("Failed to create pricing tier", 
			zap.String("plan_id", planID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create pricing tier", 
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Pricing tier created successfully", zap.Uint("tier_id", tier.ID))
	c.JSON(http.StatusCreated, gin.H{
		"message": "Pricing tier created successfully",
		"tier":    tier,
	})
}

// GetPricingTiers retrieves pricing tiers for a plan.
func (h *SaaSAdminHandler) GetPricingTiers(c *gin.Context) {
	h.logger.Info("Getting pricing tiers")

	planIDStr := c.Param("planId")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", planIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID format, expected UUID",
		})
		return
	}

	tiers, err := h.service.GetPricingTiers(c.Request.Context(), planID)
	if err != nil {
		h.logger.Error("Failed to get pricing tiers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pricing tiers", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing tiers retrieved successfully", zap.Int("count", len(tiers)))
	c.JSON(http.StatusOK, gin.H{
		"tiers": tiers,
		"count": len(tiers),
	})
}

// UpdatePricingTier updates a pricing tier.
func (h *SaaSAdminHandler) UpdatePricingTier(c *gin.Context) {
	h.logger.Info("Updating pricing tier")

	tierIDStr := c.Param("id")
	tierID, err := strconv.ParseUint(tierIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid tier ID format", 
			zap.String("tier_id", tierIDStr), 
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tier ID, expected numeric value",
		})
		return
	}

	var request struct {
		models.PricingTier
		PlanID *string `json:"plan_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body", 
			"details": err.Error(),
		})
		return
	}

	// If plan_id is provided in the request, validate it's a valid UUID
	if request.PlanID != nil {
		if _, err := uuid.Parse(*request.PlanID); err != nil {
			h.logger.Error("Invalid plan ID format", 
				zap.String("plan_id", *request.PlanID), 
				zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid plan ID format, expected UUID",
			})
			return
		}
	}

	updatedTier, err := h.service.UpdatePricingTier(c.Request.Context(), uint(tierID), &request.PricingTier)
	if err != nil {
		h.logger.Error("Failed to update pricing tier", 
			zap.Uint("tier_id", uint(tierID)),
			zap.Error(err))
		status := http.StatusInternalServerError
		errMsg := "Failed to update pricing tier"
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
			errMsg = "Pricing tier not found"
		}
		c.JSON(status, gin.H{
			"error": errMsg, 
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Pricing tier updated successfully", 
		zap.Uint("tier_id", updatedTier.ID),
		zap.String("plan_id", updatedTier.PlanID.String()))
	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing tier updated successfully",
		"tier":    updatedTier,
	})
}

// DeletePricingTier deletes a pricing tier.
func (h *SaaSAdminHandler) DeletePricingTier(c *gin.Context) {
	h.logger.Info("Deleting pricing tier")

	tierIDStr := c.Param("id")
	tierID, err := strconv.ParseUint(tierIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid tier ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tier ID"})
		return
	}

	if err := h.service.DeletePricingTier(c.Request.Context(), uint(tierID)); err != nil {
		h.logger.Error("Failed to delete pricing tier", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pricing tier", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing tier deleted successfully", zap.Uint("tier_id", uint(tierID)))
	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing tier deleted successfully",
	})
}

// AssignFeatureToPlan assigns a feature to a plan.
func (h *SaaSAdminHandler) AssignFeatureToPlan(c *gin.Context) {
	h.logger.Info("Assigning feature to plan")

	var request struct {
		PlanID    string `json:"plan_id" binding:"required"`
		FeatureID uint   `json:"feature_id" binding:"required"`
		Order     int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	planUUID, err := uuid.Parse(request.PlanID)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID format"})
		return
	}

	if err := h.service.AssignFeatureToPlan(c.Request.Context(), planUUID, request.FeatureID, request.Order); err != nil {
		h.logger.Error("Failed to assign feature to plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign feature to plan", "details": err.Error()})
		return
	}

	h.logger.Info("Feature assigned to plan successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Feature assigned to plan successfully",
	})
}

// RemoveFeatureFromPlan removes a feature from a plan.
func (h *SaaSAdminHandler) RemoveFeatureFromPlan(c *gin.Context) {
	h.logger.Info("Removing feature from plan")

	var request struct {
		PlanID    string `json:"plan_id" binding:"required"`
		FeatureID uint   `json:"feature_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Parse plan ID as UUID
	planID, err := uuid.Parse(request.PlanID)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", request.PlanID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plan ID format, expected UUID",
		})
		return
	}

	if err := h.service.RemoveFeatureFromPlan(c.Request.Context(), planID, request.FeatureID); err != nil {
		h.logger.Error("Failed to remove feature from plan", 
			zap.String("plan_id", planID.String()),
			zap.Uint("feature_id", request.FeatureID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove feature from plan",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Feature removed from plan successfully",
		zap.String("plan_id", planID.String()),
		zap.Uint("feature_id", request.FeatureID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Feature removed from plan successfully",
	})
}

// GetPlanFeatures retrieves all features for a plan.
func (h *SaaSAdminHandler) GetPlanFeatures(c *gin.Context) {
	h.logger.Info("Getting plan features")

	planIDStr := c.Param("planId")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		h.logger.Error("Invalid plan ID format", zap.String("planId", planIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID format, expected UUID"})
		return
	}

	features, err := h.service.GetPlanFeatures(c.Request.Context(), planID)
	if err != nil {
		h.logger.Error("Failed to get plan features", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plan features", "details": err.Error()})
		return
	}

	h.logger.Info("Plan features retrieved successfully", zap.Int("count", len(features)))
	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"count":    len(features),
	})
}

// GetPublicPricingPlans retrieves all public pricing plans.
func (h *SaaSAdminHandler) GetPublicPricingPlans(c *gin.Context) {
	h.logger.Info("Getting public pricing plans")

	plans, err := h.service.GetPublicPricingPlans(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get public pricing plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get public pricing plans", "details": err.Error()})
		return
	}

	h.logger.Info("Public pricing plans retrieved successfully", zap.Int("count", len(plans)))
	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
		"count": len(plans),
	})
}

// SyncPricingToLandingPage syncs pricing plans to the landing page service.
func (h *SaaSAdminHandler) SyncPricingToLandingPage(c *gin.Context) {
	h.logger.Info("Syncing pricing plans to landing page service")

	if err := h.service.SyncPricingToLandingPage(c.Request.Context()); err != nil {
		h.logger.Error("Failed to sync pricing plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync pricing plans", "details": err.Error()})
		return
	}

	h.logger.Info("Pricing plans synced successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing plans synced to landing page service successfully",
	})
}
