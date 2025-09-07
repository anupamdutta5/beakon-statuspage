package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/gin-gonic/gin"
)

// SecurityHandlers handles security-related API endpoints
type SecurityHandlers struct {
	securityService *services.TenantSecurityService
}

// NewSecurityHandlers creates a new security handlers instance
func NewSecurityHandlers(securityService *services.TenantSecurityService) *SecurityHandlers {
	return &SecurityHandlers{
		securityService: securityService,
	}
}

// GetSecurityConfig returns security configuration for a tenant
func (h *SecurityHandlers) GetSecurityConfig(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	config, err := h.securityService.GetSecurityConfig(c.Request.Context(), tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
	})
}

// UpdateSecurityConfig updates security configuration for a tenant
func (h *SecurityHandlers) UpdateSecurityConfig(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var config services.SecurityConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.securityService.UpdateSecurityConfig(c.Request.Context(), tenant.ID, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Security configuration updated successfully",
		"config":  config,
	})
}

// GetSecurityMetrics returns security metrics for a tenant
func (h *SecurityHandlers) GetSecurityMetrics(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get days parameter
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	metrics, err := h.securityService.GetSecurityMetrics(c.Request.Context(), tenant.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetAuditLogs returns audit logs for a tenant
func (h *SecurityHandlers) GetAuditLogs(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	// Get filter parameters
	action := c.Query("action")
	success := c.Query("success")

	// In a real implementation, you'd query the database
	// For now, we'll return mock data
	auditLogs := []services.SecurityAuditLog{
		{
			ID:        1,
			TenantID:  tenant.ID,
			UserID:    1,
			Action:    "GET",
			Resource:  "/api/v1/tenant/incidents",
			Details:   "Status: 200, Duration: 45ms",
			IPAddress: "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Success:   true,
		},
		{
			ID:        2,
			TenantID:  tenant.ID,
			UserID:    1,
			Action:    "POST",
			Resource:  "/api/v1/tenant/incidents",
			Details:   "Status: 201, Duration: 120ms",
			IPAddress: "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timestamp: time.Date(2024, 1, 15, 10, 25, 0, 0, time.UTC),
			Success:   true,
		},
	}

	// Apply filters
	filteredLogs := []services.SecurityAuditLog{}
	for _, log := range auditLogs {
		if action != "" && log.Action != action {
			continue
		}
		if success != "" {
			successBool := success == "true"
			if log.Success != successBool {
				continue
			}
		}
		filteredLogs = append(filteredLogs, log)
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= len(filteredLogs) {
		filteredLogs = []services.SecurityAuditLog{}
	} else if end > len(filteredLogs) {
		filteredLogs = filteredLogs[start:]
	} else {
		filteredLogs = filteredLogs[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": filteredLogs,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       len(auditLogs),
			"total_pages": (len(auditLogs) + limit - 1) / limit,
		},
	})
}

// GetSecurityViolations returns security violations for a tenant
func (h *SecurityHandlers) GetSecurityViolations(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	// Get filter parameters
	severity := c.Query("severity")
	resolved := c.Query("resolved")

	// In a real implementation, you'd query the database
	// For now, we'll return mock data
	violations := []services.SecurityViolation{
		{
			ID:          1,
			TenantID:    tenant.ID,
			Type:        "unauthorized_access",
			Severity:    "high",
			Description: "Unauthorized access attempt to admin panel",
			IPAddress:   "192.168.1.200",
			UserAgent:   "curl/7.68.0",
			Details:     "Multiple failed login attempts",
			Timestamp:   time.Date(2024, 1, 15, 9, 15, 0, 0, time.UTC),
			Resolved:    false,
		},
		{
			ID:          2,
			TenantID:    tenant.ID,
			Type:        "suspicious_activity",
			Severity:    "medium",
			Description: "Rate limit exceeded",
			IPAddress:   "192.168.1.150",
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Details:     "100+ requests in 1 minute",
			Timestamp:   time.Date(2024, 1, 15, 8, 45, 0, 0, time.UTC),
			Resolved:    true,
		},
	}

	// Apply filters
	filteredViolations := []services.SecurityViolation{}
	for _, violation := range violations {
		if severity != "" && violation.Severity != severity {
			continue
		}
		if resolved != "" {
			resolvedBool := resolved == "true"
			if violation.Resolved != resolvedBool {
				continue
			}
		}
		filteredViolations = append(filteredViolations, violation)
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= len(filteredViolations) {
		filteredViolations = []services.SecurityViolation{}
	} else if end > len(filteredViolations) {
		filteredViolations = filteredViolations[start:]
	} else {
		filteredViolations = filteredViolations[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"violations": filteredViolations,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       len(violations),
			"total_pages": (len(violations) + limit - 1) / limit,
		},
	})
}

// ResolveSecurityViolation marks a security violation as resolved
func (h *SecurityHandlers) ResolveSecurityViolation(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	violationIDStr := c.Param("id")
	violationID, err := strconv.ParseUint(violationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid violation ID"})
		return
	}

	// In a real implementation, you'd update the database
	// For now, we'll just return success
	_ = violationID
	_ = tenant.ID

	c.JSON(http.StatusOK, gin.H{
		"message": "Security violation resolved successfully",
	})
}

// GenerateAPIKey generates a new API key for a tenant
func (h *SecurityHandlers) GenerateAPIKey(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var request struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.securityService.GenerateAPIKey(c.Request.Context(), tenant.ID, request.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key generated successfully",
		"api_key": apiKey,
		"name":    request.Name,
		"warning": "Store this API key securely. It will not be shown again.",
	})
}

// ValidateAPIKey validates an API key
func (h *SecurityHandlers) ValidateAPIKey(c *gin.Context) {
	var request struct {
		APIKey string `json:"api_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := h.securityService.ValidateAPIKey(c.Request.Context(), request.APIKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"tenant_id": tenantID,
	})
}

// EncryptData encrypts sensitive data
func (h *SecurityHandlers) EncryptData(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var request struct {
		Data string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedData, err := h.securityService.EncryptData(c.Request.Context(), tenant.ID, request.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"encrypted_data": encryptedData,
	})
}

// DecryptData decrypts sensitive data
func (h *SecurityHandlers) DecryptData(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var request struct {
		EncryptedData string `json:"encrypted_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	decryptedData, err := h.securityService.DecryptData(c.Request.Context(), tenant.ID, request.EncryptedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": decryptedData,
	})
}
