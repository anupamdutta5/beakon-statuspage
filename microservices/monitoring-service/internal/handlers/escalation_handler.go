// Package handlers provides HTTP handlers for escalation policy endpoints.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/monitoring-service/internal/models"
	"github.com/anupamdutta5/monitoring-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EscalationHandler handles escalation policy HTTP requests.
type EscalationHandler struct {
	service *services.EscalationService
	logger  *zap.Logger
}

// NewEscalationHandler creates a new escalation handler.
func NewEscalationHandler(service *services.EscalationService, logger *zap.Logger) *EscalationHandler {
	return &EscalationHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePolicy creates a new escalation policy.
// POST /api/v1/escalations/policies
func (h *EscalationHandler) CreatePolicy(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var policy models.EscalationPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreatePolicy(tenantID, &policy); err != nil {
		h.logger.Error("Failed to create escalation policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create policy"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"policy": policy,
	})
}

// GetPolicies retrieves all escalation policies for a tenant.
// GET /api/v1/escalations/policies
func (h *EscalationHandler) GetPolicies(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	policies, err := h.service.GetPolicies(tenantID)
	if err != nil {
		h.logger.Error("Failed to fetch escalation policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch policies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"count":    len(policies),
		"policies": policies,
	})
}

// GetPolicy retrieves a specific escalation policy.
// GET /api/v1/escalations/policies/:id
func (h *EscalationHandler) GetPolicy(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	policyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	policy, err := h.service.GetPolicy(uint(policyID), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"policy": policy,
	})
}

// UpdatePolicy updates an escalation policy.
// PUT /api/v1/escalations/policies/:id
func (h *EscalationHandler) UpdatePolicy(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	policyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdatePolicy(uint(policyID), tenantID, updates); err != nil {
		h.logger.Error("Failed to update policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Policy updated successfully",
	})
}

// DeletePolicy deletes an escalation policy.
// DELETE /api/v1/escalations/policies/:id
func (h *EscalationHandler) DeletePolicy(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	policyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	if err := h.service.DeletePolicy(uint(policyID), tenantID); err != nil {
		h.logger.Error("Failed to delete policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Policy deleted successfully",
	})
}

// StartEscalation starts an escalation for an incident.
// POST /api/v1/escalations/start
func (h *EscalationHandler) StartEscalation(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req struct {
		IncidentID string `json:"incident_id" binding:"required"`
		PolicyID   uint   `json:"policy_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	incidentID, err := uuid.Parse(req.IncidentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	tracker, err := h.service.StartEscalation(tenantID, incidentID, req.PolicyID)
	if err != nil {
		h.logger.Error("Failed to start escalation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start escalation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Escalation started",
		"tracker": tracker,
	})
}

// ResolveEscalation marks an escalation as resolved.
// POST /api/v1/escalations/incidents/:incident_id/resolve
func (h *EscalationHandler) ResolveEscalation(c *gin.Context) {
	incidentIDStr := c.Param("incident_id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	if err := h.service.ResolveEscalation(incidentID); err != nil {
		h.logger.Error("Failed to resolve escalation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve escalation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Escalation resolved",
	})
}

// GetActiveEscalations retrieves all active escalations for a tenant.
// GET /api/v1/escalations/active
func (h *EscalationHandler) GetActiveEscalations(c *gin.Context) {
	tenantID, err := uuid.Parse(c.GetString("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	escalations, err := h.service.GetActiveEscalations(tenantID)
	if err != nil {
		h.logger.Error("Failed to fetch active escalations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active escalations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"count":       len(escalations),
		"escalations": escalations,
	})
}

// Note: GetEscalationForIncident endpoint temporarily commented out
// Will be added when GetEscalationByIncident method is implemented in service
