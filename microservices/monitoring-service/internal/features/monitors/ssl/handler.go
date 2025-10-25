// Package handlers provides HTTP request handlers for the Monitoring Service.
package ssl

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/services"
)

// SSLHandler handles SSL certificate monitoring API endpoints.
type SSLHandler struct {
	sslService *services.SSLScannerService
}

// NewSSLHandler creates a new SSL handler instance.
func NewSSLHandler(db *gorm.DB) *SSLHandler {
	return &SSLHandler{
		sslService: services.NewSSLScannerService(db),
	}
}

// ScanDomainRequest represents a request to scan a domain for its SSL certificate.
type ScanDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

// ScanDomain scans a domain and stores its SSL certificate information.
// POST /api/v1/ssl/scan
func (h *SSLHandler) ScanDomain(c *gin.Context) {
	var req ScanDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Get tenant ID from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Tenant ID not found in context",
		})
		return
	}

	tenantUUID, err := uuid.Parse(tenantID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid tenant ID",
		})
		return
	}

	// Scan the domain
	cert, err := h.sslService.ScanDomain(tenantUUID, req.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to scan domain",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Domain scanned successfully",
		"data":    cert,
	})
}

// GetCertificates retrieves all SSL certificates for the authenticated tenant.
// GET /api/v1/ssl/certificates
func (h *SSLHandler) GetCertificates(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Tenant ID not found in context",
		})
		return
	}

	tenantUUID, err := uuid.Parse(tenantID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid tenant ID",
		})
		return
	}

	certs, err := h.sslService.GetTenantCertificates(tenantUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch certificates",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(certs),
		"data":   certs,
	})
}

// GetCertificate retrieves a specific SSL certificate by ID.
// GET /api/v1/ssl/certificates/:id
func (h *SSLHandler) GetCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid certificate ID",
		})
		return
	}

	cert, err := h.sslService.GetCertificate(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Certificate not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch certificate",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   cert,
	})
}

// GetExpiringCertificates retrieves certificates expiring within a specified number of days.
// GET /api/v1/ssl/expiring?days=30
func (h *SSLHandler) GetExpiringCertificates(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Tenant ID not found in context",
		})
		return
	}

	tenantUUID, err := uuid.Parse(tenantID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid tenant ID",
		})
		return
	}

	// Get days parameter (default: 30)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid days parameter (must be positive integer)",
		})
		return
	}

	certs, err := h.sslService.GetExpiringCertificates(tenantUUID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch expiring certificates",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(certs),
		"data":   certs,
	})
}

// DeleteCertificate deletes an SSL certificate by ID.
// DELETE /api/v1/ssl/certificates/:id
func (h *SSLHandler) DeleteCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid certificate ID",
		})
		return
	}

	if err := h.sslService.DeleteCertificate(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete certificate",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Certificate deleted successfully",
	})
}

// RescanCertificate rescans a specific SSL certificate.
// POST /api/v1/ssl/certificates/:id/rescan
func (h *SSLHandler) RescanCertificate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid certificate ID",
		})
		return
	}

	// Get existing certificate
	cert, err := h.sslService.GetCertificate(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Certificate not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch certificate",
			"error":   err.Error(),
		})
		return
	}

	// Rescan the domain
	updatedCert, err := h.sslService.ScanDomain(cert.TenantID, cert.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to rescan certificate",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Certificate rescanned successfully",
		"data":    updatedCert,
	})
}
