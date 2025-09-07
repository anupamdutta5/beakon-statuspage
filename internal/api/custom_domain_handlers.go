package api

import (
	"net/http"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/gin-gonic/gin"
)

// CustomDomainHandlers handles custom domain API endpoints
type CustomDomainHandlers struct {
	customDomainService *services.CustomDomainService
}

// NewCustomDomainHandlers creates a new custom domain handlers instance
func NewCustomDomainHandlers(customDomainService *services.CustomDomainService) *CustomDomainHandlers {
	return &CustomDomainHandlers{
		customDomainService: customDomainService,
	}
}

// SetCustomDomain sets up a custom domain for a tenant
func (h *CustomDomainHandlers) SetCustomDomain(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	var req services.DomainSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set tenant ID from context
	req.TenantID = tenant.ID

	// Set up custom domain
	status, err := h.customDomainService.SetCustomDomain(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Custom domain setup initiated",
		"status":     status,
		"next_steps": h.generateNextSteps(status),
	})
}

// GetDomainStatus returns the current status of a tenant's custom domain
func (h *CustomDomainHandlers) GetDomainStatus(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	status, err := h.customDomainService.GetDomainStatus(c.Request.Context(), tenant.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       status,
		"instructions": h.generateInstructions(status),
	})
}

// VerifyDomain manually triggers domain verification
func (h *CustomDomainHandlers) VerifyDomain(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	status, err := h.customDomainService.VerifyDomain(c.Request.Context(), tenant.Domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain verification completed",
		"status":  status,
	})
}

// RemoveCustomDomain removes a custom domain from a tenant
func (h *CustomDomainHandlers) RemoveCustomDomain(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if err := h.customDomainService.RemoveCustomDomain(c.Request.Context(), tenant.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom domain removed successfully",
	})
}

// GetDNSInstructions returns DNS setup instructions for a domain
func (h *CustomDomainHandlers) GetDNSInstructions(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if tenant.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No custom domain configured"})
		return
	}

	instructions := h.customDomainService.GenerateDNSInstructions(tenant.Domain)
	c.JSON(http.StatusOK, instructions)
}

// GetSSLInstructions returns SSL setup instructions for a domain
func (h *CustomDomainHandlers) GetSSLInstructions(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if tenant.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No custom domain configured"})
		return
	}

	instructions := h.customDomainService.GenerateSSLInstructions(tenant.Domain)
	c.JSON(http.StatusOK, instructions)
}

// UploadSSLCertificate handles SSL certificate upload
func (h *CustomDomainHandlers) UploadSSLCertificate(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get uploaded files
	_, err := c.FormFile("certificate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate file is required"})
		return
	}

	_, err = c.FormFile("private_key")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Private key file is required"})
		return
	}

	// In a real implementation, you'd:
	// 1. Validate the certificate and key
	// 2. Store them securely
	// 3. Configure the web server to use them
	// 4. Test the SSL configuration

	c.JSON(http.StatusOK, gin.H{
		"message": "SSL certificate uploaded successfully",
		"domain":  tenant.Domain,
		"status":  "processing",
	})
}

// EnableAutomaticSSL enables automatic SSL certificate management
func (h *CustomDomainHandlers) EnableAutomaticSSL(c *gin.Context) {
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if tenant.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No custom domain configured"})
		return
	}

	// In a real implementation, you'd:
	// 1. Set up Let's Encrypt certificate
	// 2. Configure automatic renewal
	// 3. Update domain status

	c.JSON(http.StatusOK, gin.H{
		"message":      "Automatic SSL enabled successfully",
		"domain":       tenant.Domain,
		"provider":     "Let's Encrypt",
		"auto_renewal": true,
	})
}

// Helper functions

// generateNextSteps generates next steps based on domain status
func (h *CustomDomainHandlers) generateNextSteps(status *services.DomainVerificationStatus) []map[string]interface{} {
	steps := []map[string]interface{}{}

	if !status.IsVerified {
		steps = append(steps, map[string]interface{}{
			"step":        1,
			"title":       "Configure DNS",
			"description": "Add the required DNS records to your domain",
			"action":      "GET /api/v1/tenant/domains/dns-instructions",
			"completed":   false,
		})
	}

	if !status.SSLStatus.IsValid {
		steps = append(steps, map[string]interface{}{
			"step":        2,
			"title":       "Set up SSL Certificate",
			"description": "Enable SSL for secure connections",
			"action":      "GET /api/v1/tenant/domains/ssl-instructions",
			"completed":   false,
		})
	}

	if status.IsVerified && status.SSLStatus.IsValid {
		steps = append(steps, map[string]interface{}{
			"step":        3,
			"title":       "Domain Ready",
			"description": "Your custom domain is fully configured and ready to use",
			"action":      "none",
			"completed":   true,
		})
	}

	return steps
}

// generateInstructions generates setup instructions based on domain status
func (h *CustomDomainHandlers) generateInstructions(status *services.DomainVerificationStatus) map[string]interface{} {
	instructions := map[string]interface{}{
		"domain": status.Domain,
		"status": map[string]interface{}{
			"verified":  status.IsVerified,
			"ssl_valid": status.SSLStatus.IsValid,
		},
	}

	if !status.IsVerified {
		instructions["dns_instructions"] = h.customDomainService.GenerateDNSInstructions(status.Domain)
	}

	if !status.SSLStatus.IsValid {
		instructions["ssl_instructions"] = h.customDomainService.GenerateSSLInstructions(status.Domain)
	}

	return instructions
}
