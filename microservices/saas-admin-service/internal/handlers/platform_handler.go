// Package handlers provides HTTP handlers for platform management.
package handlers

import (
	"net/http"

	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/models"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PlatformHandler handles platform configuration HTTP requests.
type PlatformHandler struct {
	service *services.SaaSAdminService
	logger  *zap.Logger
}

// NewPlatformHandler creates a new platform handler.
func NewPlatformHandler(service *services.SaaSAdminService, logger *zap.Logger) *PlatformHandler {
	return &PlatformHandler{
		service: service,
		logger:  logger,
	}
}

// GetPlatform handles getting platform configuration.
func (h *PlatformHandler) GetPlatform(c *gin.Context) {
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
func (h *PlatformHandler) UpdatePlatform(c *gin.Context) {
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