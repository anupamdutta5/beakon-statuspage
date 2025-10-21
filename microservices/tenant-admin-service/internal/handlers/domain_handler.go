package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
	"go.uber.org/zap"
)

// DomainHandler handles domain-related HTTP requests
type DomainHandler struct {
	domainSvc services.DomainService
	logger    *zap.Logger
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(domainSvc services.DomainService, logger *zap.Logger) *DomainHandler {
	return &DomainHandler{
		domainSvc: domainSvc,
		logger:    logger,
	}
}

// AddDomainRequest represents the request body for adding a domain
type AddDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

// AddDomain adds a new domain for a status page
func (h *DomainHandler) AddDomain(c *gin.Context) {
	// Get tenant ID from path
	tenantID, err := strconv.ParseUint(c.Param("tenantID"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get status page ID from path
	statusPageID, err := strconv.ParseUint(c.Param("statusPageID"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid status page ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status page ID"})
		return
	}

	// Parse request body
	var req AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Add domain
	domain, token, err := h.domainSvc.AddDomain(c.Request.Context(), uint(tenantID), uint(statusPageID), req.Domain)
	if err != nil {
		h.logger.Error("Failed to add domain", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add domain"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"domain": domain,
		"verification_token": token,
	})
}

// VerifyDomain verifies domain ownership
func (h *DomainHandler) VerifyDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	err := h.domainSvc.VerifyDomain(c.Request.Context(), domain, token)
	if err != nil {
		h.logger.Error("Failed to verify domain", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Domain verified successfully"})
}

// GetDomain gets a domain by ID
func (h *DomainHandler) GetDomain(c *gin.Context) {
	domainID, err := strconv.ParseUint(c.Param("domainID"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid domain ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domain, err := h.domainSvc.GetDomain(c.Request.Context(), uint(domainID))
	if err != nil {
		h.logger.Error("Failed to get domain", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
		return
	}

	c.JSON(http.StatusOK, domain)
}

// DeleteDomain deletes a domain
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	domainID, err := strconv.ParseUint(c.Param("domainID"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid domain ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	err = h.domainSvc.DeleteDomain(c.Request.Context(), uint(domainID))
	if err != nil {
		h.logger.Error("Failed to delete domain", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Domain deleted successfully"})
}