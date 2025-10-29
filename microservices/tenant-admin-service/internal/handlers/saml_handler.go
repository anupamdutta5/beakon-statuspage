// Package handlers provides HTTP handlers for SAML/SSO authentication
package handlers

import (
	"encoding/xml"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/anupamdutta5/tenant-admin-service/internal/middleware"
	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/anupamdutta5/tenant-admin-service/internal/services"
)

// SAMLHandler handles SAML/SSO HTTP requests with multi-tenant support
type SAMLHandler struct {
	samlService   *services.SAMLService
	tenantService *services.TenantAdminService
	logger        *zap.Logger
}

// NewSAMLHandler creates a new SAML handler
func NewSAMLHandler(samlService *services.SAMLService, tenantService *services.TenantAdminService, logger *zap.Logger) *SAMLHandler {
	return &SAMLHandler{
		samlService:   samlService,
		tenantService: tenantService,
		logger:        logger,
	}
}

// InitiateLogin godoc
// @Summary      Initiate SAML login
// @Description  Start SAML authentication flow for a given organization domain
// @Tags         SAML
// @Accept       json
// @Produce      json
// @Param        request body InitiateLoginRequest true "Login request"
// @Success      200 {object} InitiateLoginResponse
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /saml/login [post]
func (h *SAMLHandler) InitiateLogin(c *gin.Context) {
	var req InitiateLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get tenant from middleware context
	tenantIDStr, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_tenant",
			Message: "Tenant context is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_tenant",
			Message: "Invalid tenant ID",
		})
		return
	}

	// Extract subdomain for tenant-specific URLs
	subdomain := extractSubdomain(c.Request.Host)

	// Validate organization domain
	if req.OrganizationDomain == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_domain",
			Message: "Organization domain is required",
		})
		return
	}

	// Initiate SAML login with tenant context
	loginReq, err := h.samlService.InitiateLogin(c.Request.Context(), tenantID, req.OrganizationDomain, req.RelayState, c.Request, subdomain)
	if err != nil {
		h.logger.Error("Failed to initiate SAML login",
			zap.String("domain", req.OrganizationDomain),
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "login_failed",
			Message: "Failed to initiate SAML login",
		})
		return
	}

	c.JSON(http.StatusOK, InitiateLoginResponse{
		RequestID:   loginReq.RequestID,
		RedirectURL: loginReq.RedirectURL,
	})
}

// AssertionConsumerService godoc
// @Summary      SAML Assertion Consumer Service (ACS)
// @Description  Handle SAML assertion response from IdP
// @Tags         SAML
// @Accept       form
// @Produce      html
// @Param        SAMLResponse formData string true "SAML Response"
// @Param        RelayState   formData string false "Relay State"
// @Success      302 "Redirect to application"
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /saml/acs [post]
func (h *SAMLHandler) AssertionConsumerService(c *gin.Context) {
	// Extract subdomain for tenant context
	subdomain := extractSubdomain(c.Request.Host)

	// Handle SAML assertion with tenant context
	result, err := h.samlService.HandleACS(c.Request.Context(), c.Request, subdomain)
	if err != nil {
		h.logger.Error("Failed to handle SAML assertion",
			zap.String("subdomain", subdomain),
			zap.Error(err))

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "authentication_failed",
			Message: "Failed to process SAML assertion",
		})
		return
	}

	// Generate JWT access token (15 minutes)
	accessTokenExpiry := time.Minute * 15

	claims := jwt.MapClaims{
		"user_id":   result.User.ID.String(),
		"email":     result.User.Email,
		"tenant_id": result.TenantID.String(),
		"exp":       time.Now().Add(accessTokenExpiry).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		h.logger.Error("JWT_SECRET not configured")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Authentication service configuration error",
		})
		return
	}

	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate authentication token",
		})
		return
	}

	// Set JWT cookie with tenant-specific domain
	cookieDomain := ""
	if subdomain != "" {
		// For subdomain: set cookie domain to parent domain (e.g., .example.com)
		parts := strings.Split(c.Request.Host, ":")
		hostname := parts[0]
		if strings.Contains(hostname, ".") {
			hostParts := strings.Split(hostname, ".")
			if len(hostParts) > 1 {
				cookieDomain = "." + strings.Join(hostParts[1:], ".")
			}
		}
	}

	c.SetCookie(
		"access_token",
		accessToken,
		int(accessTokenExpiry.Seconds()),
		"/",
		cookieDomain,
		true,  // Secure (HTTPS only)
		true,  // HttpOnly
	)

	h.logger.Info("SAML authentication successful",
		zap.String("user_id", result.User.ID.String()),
		zap.String("email", result.User.Email),
		zap.String("tenant_id", result.TenantID.String()),
		zap.Bool("new_user", result.IsNewUser))

	// Redirect to relay state or default URL
	redirectURL := result.RelayState
	if redirectURL == "" {
		redirectURL = "/"
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// GetMetadata godoc
// @Summary      Get Service Provider metadata
// @Description  Return SP metadata XML for IdP configuration
// @Tags         SAML
// @Produce      xml
// @Param        organization_domain query string true "Organization domain"
// @Success      200 {string} string "SP Metadata XML"
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /saml/metadata [get]
func (h *SAMLHandler) GetMetadata(c *gin.Context) {
	organizationDomain := c.Query("organization_domain")
	if organizationDomain == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_domain",
			Message: "Organization domain is required",
		})
		return
	}

	// Get tenant from middleware context
	tenantIDStr, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_tenant",
			Message: "Tenant context is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_tenant",
			Message: "Invalid tenant ID",
		})
		return
	}

	subdomain := extractSubdomain(c.Request.Host)

	// Get SSO provider with tenant scoping
	var ssoProvider models.SSOProvider
	if err := h.samlService.DB().Where("tenant_id = ? AND organization_domain = ?",
		tenantID, organizationDomain).First(&ssoProvider).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "provider_not_found",
			Message: "SSO provider not found for organization",
		})
		return
	}

	// Get service provider
	sp, err := h.samlService.GetServiceProvider(c.Request.Context(), &ssoProvider, subdomain)
	if err != nil {
		h.logger.Error("Failed to get service provider",
			zap.String("organization_domain", organizationDomain),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "metadata_failed",
			Message: "Failed to generate metadata",
		})
		return
	}

	// Generate metadata XML
	metadata := sp.Metadata()
	metadataXML, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		h.logger.Error("Failed to marshal metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "metadata_failed",
			Message: "Failed to generate metadata XML",
		})
		return
	}

	c.Data(http.StatusOK, "application/xml", metadataXML)
}

// SingleLogout godoc
// @Summary      SAML Single Logout
// @Description  Handle SAML logout request from IdP
// @Tags         SAML
// @Accept       form
// @Produce      html
// @Success      302 "Redirect after logout"
// @Failure      500 {object} ErrorResponse
// @Router       /saml/logout [post]
func (h *SAMLHandler) SingleLogout(c *gin.Context) {
	subdomain := extractSubdomain(c.Request.Host)

	// TODO: Implement SAML Single Logout (SLO) with IdP
	// For now, just clear the local session

	// Clear JWT cookie
	c.SetCookie("access_token", "", -1, "/", "", true, true)

	h.logger.Info("SAML logout successful", zap.String("subdomain", subdomain))

	// Get relay state or default redirect
	relayState := c.PostForm("RelayState")
	if relayState == "" {
		relayState = "/"
	}

	c.Redirect(http.StatusFound, relayState)
}

// ========== Admin Management Endpoints (Protected) ==========

// ListProviders godoc
// @Summary      List SSO providers
// @Description  List all SSO providers for the current tenant
// @Tags         SSO Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ListProvidersResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /sso/providers [get]
func (h *SAMLHandler) ListProviders(c *gin.Context) {
	// Get tenant from middleware (set by auth middleware)
	tenantIDStr, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "missing_tenant",
			Message: "Tenant context is required",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_tenant",
			Message: "Invalid tenant ID",
		})
		return
	}

	var providers []models.SSOProvider
	if err := h.samlService.DB().Where("tenant_id = ?", tenantID).Find(&providers).Error; err != nil {
		h.logger.Error("Failed to list SSO providers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve SSO providers",
		})
		return
	}

	// Convert to response format (hide sensitive fields)
	providerList := make([]ProviderSummary, len(providers))
	for i, p := range providers {
		providerList[i] = ProviderSummary{
			ID:                 p.ID.String(),
			OrganizationDomain: p.OrganizationDomain,
			OrganizationName:   p.OrganizationName,
			ProviderType:       p.ProviderType,
			ProviderName:       p.ProviderName,
			IsEnabled:          p.IsEnabled,
			IsDefault:          p.IsDefault,
			CreatedAt:          p.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, ListProvidersResponse{
		Providers: providerList,
		Count:     len(providerList),
	})
}

// GetProvider godoc
// @Summary      Get SSO provider
// @Description  Get SSO provider details by ID
// @Tags         SSO Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Provider ID"
// @Success      200 {object} models.SSOProvider
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /sso/providers/{id} [get]
func (h *SAMLHandler) GetProvider(c *gin.Context) {
	providerID := c.Param("id")
	if providerID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_id",
			Message: "Provider ID is required",
		})
		return
	}

	providerUUID, err := uuid.Parse(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid provider ID format",
		})
		return
	}

	// Get tenant from middleware
	tenantIDStr, _ := middleware.GetTenantID(c)
	tenantID, _ := uuid.Parse(tenantIDStr)

	var provider models.SSOProvider
	if err := h.samlService.DB().Where("id = ? AND tenant_id = ?", providerUUID, tenantID).First(&provider).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "provider_not_found",
			Message: "SSO provider not found",
		})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// CreateProvider godoc
// @Summary      Create SSO provider
// @Description  Create a new SSO provider for the tenant
// @Tags         SSO Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateProviderRequest true "Provider details"
// @Success      201 {object} models.SSOProvider
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /sso/providers [post]
func (h *SAMLHandler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get tenant from middleware
	tenantIDStr, _ := middleware.GetTenantID(c)
	tenantID, _ := uuid.Parse(tenantIDStr)

	// Create SSO provider with tenant context
	provider := &models.SSOProvider{
		TenantID:           tenantID,
		OrganizationDomain: req.OrganizationDomain,
		OrganizationName:   req.OrganizationName,
		ProviderType:       req.ProviderType,
		ProviderName:       req.ProviderName,
		EntityID:           req.EntityID,
		IdpEntityID:        req.IdpEntityID,
		SSOURL:             req.SSOURL,
		SLOURL:             req.SLOURL,
		IdpMetadataURL:     req.IdpMetadataURL,
		IdpCertificate:     req.IdpCertificate,
		IsEnabled:          req.IsEnabled,
		AllowIdpInitiated:  req.AllowIdpInitiated,
		AttributeMapping:   req.AttributeMapping,
	}

	if err := h.samlService.DB().Create(provider).Error; err != nil {
		h.logger.Error("Failed to create SSO provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create SSO provider",
		})
		return
	}

	h.logger.Info("SSO provider created",
		zap.String("provider_id", provider.ID.String()),
		zap.String("tenant_id", tenantID.String()),
		zap.String("organization", req.OrganizationDomain))

	c.JSON(http.StatusCreated, provider)
}

// UpdateProvider godoc
// @Summary      Update SSO provider
// @Description  Update an existing SSO provider
// @Tags         SSO Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Provider ID"
// @Param        request body CreateProviderRequest true "Updated provider details"
// @Success      200 {object} models.SSOProvider
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /sso/providers/{id} [put]
func (h *SAMLHandler) UpdateProvider(c *gin.Context) {
	providerID := c.Param("id")
	providerUUID, err := uuid.Parse(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid provider ID",
		})
		return
	}

	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get tenant from middleware
	tenantIDStr, _ := middleware.GetTenantID(c)
	tenantID, _ := uuid.Parse(tenantIDStr)

	// Find existing provider
	var provider models.SSOProvider
	if err := h.samlService.DB().Where("id = ? AND tenant_id = ?", providerUUID, tenantID).First(&provider).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "provider_not_found",
			Message: "SSO provider not found",
		})
		return
	}

	// Update fields
	provider.OrganizationName = req.OrganizationName
	provider.ProviderName = req.ProviderName
	provider.IdpEntityID = req.IdpEntityID
	provider.SSOURL = req.SSOURL
	provider.SLOURL = req.SLOURL
	provider.IdpMetadataURL = req.IdpMetadataURL
	provider.IdpCertificate = req.IdpCertificate
	provider.IsEnabled = req.IsEnabled
	provider.AllowIdpInitiated = req.AllowIdpInitiated
	provider.AttributeMapping = req.AttributeMapping

	if err := h.samlService.DB().Save(&provider).Error; err != nil {
		h.logger.Error("Failed to update SSO provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update SSO provider",
		})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProvider godoc
// @Summary      Delete SSO provider
// @Description  Delete an SSO provider
// @Tags         SSO Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Provider ID"
// @Success      204 "No Content"
// @Failure      404 {object} ErrorResponse
// @Router       /sso/providers/{id} [delete]
func (h *SAMLHandler) DeleteProvider(c *gin.Context) {
	providerID := c.Param("id")
	providerUUID, err := uuid.Parse(providerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid provider ID",
		})
		return
	}

	// Get tenant from middleware
	tenantIDStr, _ := middleware.GetTenantID(c)
	tenantID, _ := uuid.Parse(tenantIDStr)

	// Soft delete provider
	if err := h.samlService.DB().Where("id = ? AND tenant_id = ?", providerUUID, tenantID).Delete(&models.SSOProvider{}).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "provider_not_found",
			Message: "SSO provider not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ========== Helper Functions ==========

// extractSubdomain extracts the subdomain from the host header
func extractSubdomain(host string) string {
	parts := strings.Split(host, ":")
	hostname := parts[0]
	hostParts := strings.Split(hostname, ".")
	if len(hostParts) > 2 {
		return hostParts[0] // e.g., "acme" from "acme.example.com"
	}
	return ""
}

// ========== Request/Response Types ==========

type InitiateLoginRequest struct {
	OrganizationDomain string `json:"organization_domain" binding:"required"`
	RelayState         string `json:"relay_state,omitempty"`
}

type InitiateLoginResponse struct {
	RequestID   string `json:"request_id"`
	RedirectURL string `json:"redirect_url"`
}

type ListProvidersResponse struct {
	Providers []ProviderSummary `json:"providers"`
	Count     int               `json:"count"`
}

type ProviderSummary struct {
	ID                 string `json:"id"`
	OrganizationDomain string `json:"organization_domain"`
	OrganizationName   string `json:"organization_name"`
	ProviderType       string `json:"provider_type"`
	ProviderName       string `json:"provider_name"`
	IsEnabled          bool   `json:"is_enabled"`
	IsDefault          bool   `json:"is_default"`
	CreatedAt          string `json:"created_at"`
}

type CreateProviderRequest struct {
	OrganizationDomain string                `json:"organization_domain" binding:"required"`
	OrganizationName   string                `json:"organization_name" binding:"required"`
	ProviderType       string                `json:"provider_type" binding:"required"` // saml, oauth, oidc
	ProviderName       string                `json:"provider_name" binding:"required"`
	EntityID           string                `json:"entity_id,omitempty"`
	IdpEntityID        string                `json:"idp_entity_id,omitempty"`
	SSOURL             string                `json:"sso_url,omitempty"`
	SLOURL             string                `json:"slo_url,omitempty"`
	IdpMetadataURL     string                `json:"idp_metadata_url,omitempty"`
	IdpCertificate     string                `json:"idp_certificate,omitempty"`
	IsEnabled          bool                  `json:"is_enabled"`
	AllowIdpInitiated  bool                  `json:"allow_idp_initiated"`
	AttributeMapping   models.AttributeMapping `json:"attribute_mapping,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
