// Package handlers provides HTTP handlers for third-party integrations in the Monitoring Service.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"github.com/anupamdutta5/statuspage-monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IntegrationHandler handles integration-related HTTP requests.
type IntegrationHandler struct {
	integrationService *services.IntegrationService
	logger             *zap.Logger
}

// NewIntegrationHandler creates a new integration handler.
func NewIntegrationHandler(integrationService *services.IntegrationService, logger *zap.Logger) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
		logger:             logger,
	}
}

// CreateIntegration handles creating a new integration.
func (h *IntegrationHandler) CreateIntegration(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var req struct {
		Name          string                 `json:"name" binding:"required"`
		Type          string                 `json:"type" binding:"required"`
		Description   string                 `json:"description"`
		Configuration map[string]interface{} `json:"configuration" binding:"required"`
		SyncInterval  int                    `json:"sync_interval"`
		IsActive      *bool                  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create integration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Convert configuration to JSON
	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration format"})
		return
	}

	// Set defaults
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.SyncInterval == 0 {
		req.SyncInterval = 300 // 5 minutes default
	}

	integration := &models.Integration{
		TenantID:      tenantID.(uint),
		Name:          req.Name,
		Type:          req.Type,
		Description:   req.Description,
		Configuration: string(configJSON),
		SyncInterval:  req.SyncInterval,
		IsActive:      isActive,
		SyncStatus:    models.SyncStatusPending,
	}

	if err := h.integrationService.CreateIntegration(c.Request.Context(), integration); err != nil {
		h.logger.Error("Failed to create integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create integration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Integration created successfully",
		"integration": integration,
	})
}

// GetIntegrations handles retrieving integrations.
func (h *IntegrationHandler) GetIntegrations(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	activeOnly := c.Query("active_only") == "true"

	integrations, err := h.integrationService.GetIntegrations(c.Request.Context(), tenantID.(uint), activeOnly)
	if err != nil {
		h.logger.Error("Failed to get integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve integrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"integrations": integrations,
		"count":        len(integrations),
	})
}

// GetIntegration handles retrieving a specific integration.
func (h *IntegrationHandler) GetIntegration(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	integration, err := h.integrationService.GetIntegration(c.Request.Context(), uint(integrationID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get integration", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"integration": integration,
	})
}

// UpdateIntegration handles updating an integration.
func (h *IntegrationHandler) UpdateIntegration(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var req struct {
		Name          *string                 `json:"name"`
		Description   *string                 `json:"description"`
		Configuration *map[string]interface{} `json:"configuration"`
		SyncInterval  *int                    `json:"sync_interval"`
		IsActive      *bool                   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update integration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Configuration != nil {
		configJSON, err := json.Marshal(*req.Configuration)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration format"})
			return
		}
		updates["configuration"] = string(configJSON)
	}
	if req.SyncInterval != nil {
		updates["sync_interval"] = *req.SyncInterval
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.integrationService.UpdateIntegration(c.Request.Context(), uint(integrationID), tenantID.(uint), updates); err != nil {
		h.logger.Error("Failed to update integration", zap.Error(err))
		if err.Error() == "integration not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update integration"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration updated successfully"})
}

// DeleteIntegration handles deleting an integration.
func (h *IntegrationHandler) DeleteIntegration(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := h.integrationService.DeleteIntegration(c.Request.Context(), uint(integrationID), tenantID.(uint)); err != nil {
		h.logger.Error("Failed to delete integration", zap.Error(err))
		if err.Error() == "integration not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete integration"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration deleted successfully"})
}

// SyncIntegration handles manually triggering integration sync.
func (h *IntegrationHandler) SyncIntegration(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := h.integrationService.SyncIntegration(c.Request.Context(), uint(integrationID), tenantID.(uint)); err != nil {
		h.logger.Error("Failed to sync integration", zap.Error(err))
		if err.Error() == "integration not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		} else if err.Error() == "integration is not active" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Integration is not active"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync integration"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration sync initiated"})
}

// TestIntegration handles testing integration connectivity.
func (h *IntegrationHandler) TestIntegration(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	result, err := h.integrationService.TestIntegration(c.Request.Context(), uint(integrationID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to test integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Integration test completed",
		"result":  result,
	})
}

// CreateComponentMapping handles creating component mappings.
func (h *IntegrationHandler) CreateComponentMapping(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var req struct {
		ComponentID         uint                   `json:"component_id" binding:"required"`
		ExternalServiceID   string                 `json:"external_service_id" binding:"required"`
		ExternalServiceName string                 `json:"external_service_name" binding:"required"`
		MappingConfig       map[string]interface{} `json:"mapping_config"`
		IsActive            *bool                  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create component mapping request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Verify integration exists and belongs to tenant
	_, err = h.integrationService.GetIntegration(c.Request.Context(), uint(integrationID), tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Convert mapping config to JSON
	var mappingConfigJSON string
	if req.MappingConfig != nil {
		configBytes, err := json.Marshal(req.MappingConfig)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping config format"})
			return
		}
		mappingConfigJSON = string(configBytes)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	mapping := &models.ComponentMapping{
		IntegrationID:       uint(integrationID),
		ComponentID:         req.ComponentID,
		ExternalServiceID:   req.ExternalServiceID,
		ExternalServiceName: req.ExternalServiceName,
		MappingConfig:       mappingConfigJSON,
		IsActive:            isActive,
	}

	if err := h.integrationService.CreateComponentMapping(c.Request.Context(), mapping); err != nil {
		h.logger.Error("Failed to create component mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create component mapping"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Component mapping created successfully",
		"mapping": mapping,
	})
}

// GetComponentMappings handles retrieving component mappings.
func (h *IntegrationHandler) GetComponentMappings(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	mappings, err := h.integrationService.GetComponentMappings(c.Request.Context(), uint(integrationID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get component mappings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve component mappings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mappings": mappings,
		"count":    len(mappings),
	})
}

// GetIntegrationSyncLogs handles retrieving sync logs.
func (h *IntegrationHandler) GetIntegrationSyncLogs(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	integrationIDStr := c.Param("id")
	integrationID, err := strconv.ParseUint(integrationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	logs, total, err := h.integrationService.GetIntegrationSyncLogs(c.Request.Context(), uint(integrationID), tenantID.(uint), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get integration sync logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sync logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetSupportedIntegrations returns supported integration types.
func (h *IntegrationHandler) GetSupportedIntegrations(c *gin.Context) {
	integrations := h.integrationService.GetSupportedIntegrations()

	c.JSON(http.StatusOK, gin.H{
		"integrations": integrations,
		"count":        len(integrations),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}