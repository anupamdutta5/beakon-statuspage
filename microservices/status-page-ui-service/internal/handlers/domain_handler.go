package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/anupamdutta5/statuspage-status-ui-service/internal/services"
	"go.uber.org/zap"
)

// DomainHandler handles domain-related HTTP requests
type DomainHandler struct {
	domainService *services.DomainService
	logger        *zap.Logger
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(domainService *services.DomainService, logger *zap.Logger) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
		logger:        logger,
	}
}

// AddDomainRequest represents the request body for adding a domain
type AddDomainRequest struct {
	Domain      string `json:"domain" binding:"required,fqdn"`
	StatusPage  string `json:"status_page" binding:"required"`
}

// AddDomain adds a new custom domain for a status page
// @Summary Add a custom domain
// @Description Add a new custom domain for a status page
// @Tags Domains
// @Accept json
// @Produce json
// @Param request body AddDomainRequest true "Domain details"
// @Success 200 {object} map[string]string "Verification token"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/domains [post]
func (h *DomainHandler) AddDomain(c *gin.Context) {
	// Get tenant ID from context (set by auth middleware)
	tenantIDValue, exists := c.Get("tenantID")
	if !exists {
		h.logger.Error("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	tenantID, ok := tenantIDValue.(uint)
	if !ok {
		h.logger.Error("Invalid tenant ID in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant context"})
		return
	}

	// Parse request body
	var req AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Add the domain
	token, err := h.domainService.AddCustomDomain(c.Request.Context(), tenantID, req.StatusPage, req.Domain)
	if err != nil {
		h.logger.Error("Failed to add domain", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verification_token": token,
		"message": "Domain added. Please add a TXT record to verify ownership.",
	})
}

// VerifyDomain verifies domain ownership
// @Summary Verify domain ownership
// @Description Verify domain ownership by checking DNS records
// @Tags Domains
// @Produce json
// @Param domain path string true "Domain to verify"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Domain not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/domains/{domain}/verify [post]
func (h *DomainHandler) VerifyDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	// Get tenant ID from context (set by auth middleware)
	tenantIDValue, exists := c.Get("tenantID")
	if !exists {
		h.logger.Error("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	_, ok := tenantIDValue.(uint)
	if !ok {
		h.logger.Error("Invalid tenant ID in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant context"})
		return
	}

	// Complete domain verification
	err := h.domainService.CompleteDomainVerification(c.Request.Context(), domain)
	if err != nil {
		if err.Error() == "verification failed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain verification failed"})
			return
		}
		h.logger.Error("Failed to verify domain", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain verified successfully",
	})
}

// ListDomains lists all domains for a tenant
// @Summary List domains
// @Description List all domains for the current tenant
// @Tags Domains
// @Produce json
// @Success 200 {array} services.DomainInfo "List of domains"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/domains [get]
func (h *DomainHandler) ListDomains(c *gin.Context) {
	// Get tenant ID from context (set by auth middleware)
	tenantIDValue, exists := c.Get("tenantID")
	if !exists {
		h.logger.Error("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	_, ok := tenantIDValue.(uint)
	if !ok {
		h.logger.Error("Invalid tenant ID in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant context"})
		return
	}

	// TODO: Implement domain listing from tenant admin service
	// For now, return an empty list
	c.JSON(http.StatusOK, []services.DomainInfo{})
}

// DeleteDomain deletes a custom domain
// @Summary Delete a domain
// @Description Delete a custom domain
// @Tags Domains
// @Produce json
// @Param domain path string true "Domain to delete"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Domain not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/domains/{domain} [delete]
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	// Get tenant ID from context (set by auth middleware)
	tenantIDValue, exists := c.Get("tenantID")
	if !exists {
		h.logger.Error("Tenant ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	_, ok := tenantIDValue.(uint)
	if !ok {
		h.logger.Error("Invalid tenant ID in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant context"})
		return
	}

	// TODO: Implement domain deletion in tenant admin service
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Domain deleted successfully",
	})
}
