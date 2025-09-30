// Package handlers provides HTTP handlers for custom domain management.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
)

// CustomDomainHandler handles custom domain HTTP requests.
type CustomDomainHandler struct {
	domainService     *services.CustomDomainService
	featureFlagService *services.FeatureFlagService
}

// CreateDomainRequest represents the request to create a custom domain.
type CreateDomainRequest struct {
	Domain   string `json:"domain" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// UpdateDomainRequest represents the request to update a custom domain.
type UpdateDomainRequest struct {
	Status    *string `json:"status,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	SSLStatus *string `json:"ssl_status,omitempty"`
}

// DomainResponse represents the response format for domain operations.
type DomainResponse struct {
	Status string                `json:"status"`
	Data   *models.CustomDomain  `json:"data,omitempty"`
	Error  string                `json:"error,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// DomainsListResponse represents the response format for domain list operations.
type DomainsListResponse struct {
	Status string                 `json:"status"`
	Data   []models.CustomDomain  `json:"data,omitempty"`
	Error  string                 `json:"error,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// DomainStatsResponse represents the response format for domain statistics.
type DomainStatsResponse struct {
	Status string               `json:"status"`
	Data   *models.DomainStats  `json:"data,omitempty"`
	Error  string               `json:"error,omitempty"`
}

// NewCustomDomainHandler creates a new custom domain handler.
func NewCustomDomainHandler(domainService *services.CustomDomainService, featureFlagService *services.FeatureFlagService) *CustomDomainHandler {
	return &CustomDomainHandler{
		domainService:     domainService,
		featureFlagService: featureFlagService,
	}
}

// CreateCustomDomain handles POST /api/v1/domains
func (h *CustomDomainHandler) CreateCustomDomain(w http.ResponseWriter, r *http.Request) {
	var req CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get tenant plan from context or query parameter
	plan := r.URL.Query().Get("plan")
	if plan == "" {
		plan = "free" // Default to free plan
	}

	// Check if custom domains are enabled for this tenant
	hasAccess, err := h.featureFlagService.HasCustomDomainAccess(req.TenantID, plan)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to check feature access")
		return
	}

	if !hasAccess {
		h.sendErrorResponse(w, http.StatusForbidden, "Custom domains are not available for your current plan. Please upgrade to Pro or Enterprise.")
		return
	}

	// Create the custom domain
	domain, err := h.domainService.CreateCustomDomain(req.TenantID, req.Domain)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response := DomainResponse{
		Status: "success",
		Data:   domain,
		Meta: map[string]interface{}{
			"verification_instructions": h.getVerificationInstructions(domain),
			"dns_instructions":          h.getDNSInstructions(domain),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetCustomDomains handles GET /api/v1/domains
func (h *CustomDomainHandler) GetCustomDomains(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "tenant_id query parameter is required")
		return
	}

	domains, err := h.domainService.GetTenantDomains(tenantID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Add default domain information
	defaultDomain := h.domainService.GetDefaultDomainForTenant(tenantID)

	response := DomainsListResponse{
		Status: "success",
		Data:   domains,
		Meta: map[string]interface{}{
			"default_domain": defaultDomain,
			"total_count":    len(domains),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCustomDomain handles GET /api/v1/domains/{id}
func (h *CustomDomainHandler) GetCustomDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	var domain models.CustomDomain
	if err := h.domainService.GetDB().First(&domain, "id = ?", domainID).Error; err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Domain not found")
		return
	}

	// Get SSL certificate status
	sslCert, _ := h.domainService.CheckSSLCertificateStatus(domainID)

	response := DomainResponse{
		Status: "success",
		Data:   &domain,
		Meta: map[string]interface{}{
			"ssl_certificate":           sslCert,
			"verification_instructions": h.getVerificationInstructions(&domain),
			"dns_instructions":          h.getDNSInstructions(&domain),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyCustomDomain handles POST /api/v1/domains/{id}/verify
func (h *CustomDomainHandler) VerifyCustomDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	if err := h.domainService.VerifyCustomDomain(domainID); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get updated domain
	var domain models.CustomDomain
	h.domainService.GetDB().First(&domain, "id = ?", domainID)

	response := DomainResponse{
		Status: "success",
		Data:   &domain,
		Meta: map[string]interface{}{
			"message": "Domain verification initiated. Check back in a few minutes.",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateCustomDomain handles PUT /api/v1/domains/{id}
func (h *CustomDomainHandler) UpdateCustomDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	var req UpdateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var domain models.CustomDomain
	if err := h.domainService.GetDB().First(&domain, "id = ?", domainID).Error; err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Domain not found")
		return
	}

	// Update fields if provided
	if req.Status != nil {
		domain.Status = *req.Status
	}
	if req.IsActive != nil {
		domain.IsActive = *req.IsActive
	}
	if req.SSLStatus != nil {
		domain.SSLStatus = *req.SSLStatus
	}

	if err := h.domainService.GetDB().Save(&domain).Error; err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to update domain")
		return
	}

	response := DomainResponse{
		Status: "success",
		Data:   &domain,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteCustomDomain handles DELETE /api/v1/domains/{id}
func (h *CustomDomainHandler) DeleteCustomDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	if err := h.domainService.DeleteCustomDomain(domainID); err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := DomainResponse{
		Status: "success",
		Meta: map[string]interface{}{
			"message": "Domain deleted successfully",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDomainStats handles GET /api/v1/domains/stats
func (h *CustomDomainHandler) GetDomainStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.domainService.GetDomainStats()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := DomainStatsResponse{
		Status: "success",
		Data:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckSSLStatus handles GET /api/v1/domains/{id}/ssl
func (h *CustomDomainHandler) CheckSSLStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	cert, err := h.domainService.CheckSSLCertificateStatus(domainID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "SSL certificate not found")
		return
	}

	response := struct {
		Status string                   `json:"status"`
		Data   *models.SSLCertificate   `json:"data"`
		Meta   map[string]interface{}   `json:"meta"`
	}{
		Status: "success",
		Data:   cert,
		Meta: map[string]interface{}{
			"is_expired":        cert.IsExpired(),
			"is_expiring_soon":  cert.IsExpiringSoon(30),
			"days_until_expiry": int(cert.ExpiresAt.Sub(time.Now()).Hours() / 24),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResolveTenant handles GET /api/v1/domains/resolve
func (h *CustomDomainHandler) ResolveTenant(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if host == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "host query parameter is required")
		return
	}

	tenantID, err := h.domainService.ResolveTenantFromHost(host)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	response := struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}{
		Status: "success",
		Data: map[string]interface{}{
			"tenant_id": tenantID,
			"host":      host,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (h *CustomDomainHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := DomainResponse{
		Status: "error",
		Error:  message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *CustomDomainHandler) getVerificationInstructions(domain *models.CustomDomain) map[string]interface{} {
	return map[string]interface{}{
		"type":        "dns",
		"record_type": domain.DNSRecordType,
		"hostname":    domain.VerificationKey + "." + domain.Domain,
		"value":       domain.VerificationValue,
		"instructions": []string{
			"1. Log in to your DNS provider (e.g., Cloudflare, GoDaddy, Namecheap)",
			"2. Navigate to DNS management for your domain",
			"3. Add a new TXT record with the following details:",
			"   - Name/Host: " + domain.VerificationKey,
			"   - Value: " + domain.VerificationValue,
			"   - TTL: 300 (or default)",
			"4. Save the record and wait for DNS propagation (up to 24 hours)",
			"5. Click 'Verify Domain' button to complete verification",
		},
		"help_links": map[string]string{
			"cloudflare": "https://developers.cloudflare.com/dns/manage-dns-records/how-to/create-dns-records/",
			"godaddy":    "https://www.godaddy.com/help/add-a-txt-record-19232",
			"namecheap":  "https://www.namecheap.com/support/knowledgebase/article.aspx/317/2237/how-do-i-add-txtspfdkimdmarc-records-for-my-domain/",
		},
	}
}

func (h *CustomDomainHandler) getDNSInstructions(domain *models.CustomDomain) map[string]interface{} {
	return map[string]interface{}{
		"record_type": domain.DNSRecordType,
		"hostname":    domain.Domain,
		"value":       domain.DNSRecordValue,
		"instructions": []string{
			"After domain verification is complete:",
			"1. Add a " + domain.DNSRecordType + " record for your domain:",
			"   - Name/Host: " + domain.Domain + " (or @ for root domain)",
			"   - Value: " + domain.DNSRecordValue,
			"   - TTL: 300 (or default)",
			"2. Save the record and wait for DNS propagation",
			"3. Your status page will be accessible at https://" + domain.Domain,
		},
		"note": "SSL certificate will be automatically issued once DNS is properly configured.",
	}
}

// RegisterRoutes registers all custom domain routes.
func (h *CustomDomainHandler) RegisterRoutes(router *mux.Router) {
	// Domain management endpoints
	router.HandleFunc("/api/v1/domains", h.CreateCustomDomain).Methods("POST")
	router.HandleFunc("/api/v1/domains", h.GetCustomDomains).Methods("GET")
	router.HandleFunc("/api/v1/domains/{id}", h.GetCustomDomain).Methods("GET")
	router.HandleFunc("/api/v1/domains/{id}", h.UpdateCustomDomain).Methods("PUT")
	router.HandleFunc("/api/v1/domains/{id}", h.DeleteCustomDomain).Methods("DELETE")

	// Domain operations
	router.HandleFunc("/api/v1/domains/{id}/verify", h.VerifyCustomDomain).Methods("POST")
	router.HandleFunc("/api/v1/domains/{id}/ssl", h.CheckSSLStatus).Methods("GET")

	// Utility endpoints
	router.HandleFunc("/api/v1/domains/stats", h.GetDomainStats).Methods("GET")
	router.HandleFunc("/api/v1/domains/resolve", h.ResolveTenant).Methods("GET")
}