package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CustomDomainService handles custom domain management
type CustomDomainService struct {
	db *gorm.DB
}

// NewCustomDomainService creates a new custom domain service
func NewCustomDomainService(db *gorm.DB) *CustomDomainService {
	return &CustomDomainService{
		db: db,
	}
}

// DomainVerificationStatus represents the verification status of a domain
type DomainVerificationStatus struct {
	Domain           string    `json:"domain"`
	IsVerified       bool      `json:"is_verified"`
	VerificationType string    `json:"verification_type"` // dns, file, ssl
	LastChecked      time.Time `json:"last_checked"`
	Error            string    `json:"error,omitempty"`
	SSLStatus        SSLStatus `json:"ssl_status"`
}

// SSLStatus represents SSL certificate status
type SSLStatus struct {
	IsValid     bool      `json:"is_valid"`
	ExpiresAt   time.Time `json:"expires_at"`
	Issuer      string    `json:"issuer"`
	CommonName  string    `json:"common_name"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
}

// DomainSetupRequest represents a request to set up a custom domain
type DomainSetupRequest struct {
	TenantID         uint   `json:"tenant_id" binding:"required"`
	Domain           string `json:"domain" binding:"required"`
	VerificationType string `json:"verification_type"`      // dns, file
	AdminDomain      string `json:"admin_domain,omitempty"` // Optional separate admin domain
}

// SetCustomDomain sets up a custom domain for a tenant
func (s *CustomDomainService) SetCustomDomain(ctx context.Context, req *DomainSetupRequest) (*DomainVerificationStatus, error) {
	// Validate domain format
	if !s.isValidDomain(req.Domain) {
		return nil, fmt.Errorf("invalid domain format: %s", req.Domain)
	}

	// Check if domain is already in use
	var existingTenant models.Tenant
	if err := s.db.Where("domain = ? OR subdomain = ?", req.Domain, req.Domain).First(&existingTenant).Error; err == nil {
		if existingTenant.ID != req.TenantID {
			return nil, fmt.Errorf("domain %s is already in use by another tenant", req.Domain)
		}
	}

	// Get tenant
	var tenant models.Tenant
	if err := s.db.First(&tenant, req.TenantID).Error; err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Update tenant with custom domain
	tenant.Domain = req.Domain
	if req.AdminDomain != "" {
		// For now, we'll use the same domain for both status page and admin
		// In a more advanced setup, we could have separate domains
		// TODO: Implement separate admin domain functionality
		_ = req.AdminDomain // Acknowledge the parameter to avoid linter warning
	}

	if err := s.db.Save(&tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to update tenant domain: %w", err)
	}

	// Generate verification record
	verificationStatus := &DomainVerificationStatus{
		Domain:           req.Domain,
		IsVerified:       false,
		VerificationType: req.VerificationType,
		LastChecked:      time.Now(),
		SSLStatus:        SSLStatus{},
	}

	// Perform initial verification
	if err := s.verifyDomain(ctx, verificationStatus); err != nil {
		logger.Warn("Domain verification failed", zap.String("domain", req.Domain), zap.Error(err))
		verificationStatus.Error = err.Error()
	}

	return verificationStatus, nil
}

// VerifyDomain verifies domain ownership and SSL status
func (s *CustomDomainService) VerifyDomain(ctx context.Context, domain string) (*DomainVerificationStatus, error) {
	status := &DomainVerificationStatus{
		Domain:           domain,
		IsVerified:       false,
		VerificationType: "dns", // Default to DNS verification
		LastChecked:      time.Now(),
		SSLStatus:        SSLStatus{},
	}

	if err := s.verifyDomain(ctx, status); err != nil {
		status.Error = err.Error()
		return status, err
	}

	return status, nil
}

// verifyDomain performs the actual domain verification
func (s *CustomDomainService) verifyDomain(ctx context.Context, status *DomainVerificationStatus) error {
	// Check DNS resolution
	if err := s.checkDNSResolution(status.Domain); err != nil {
		return fmt.Errorf("DNS resolution failed: %w", err)
	}

	// Check SSL certificate
	sslStatus, err := s.checkSSLStatus(status.Domain)
	if err != nil {
		logger.Warn("SSL check failed", zap.String("domain", status.Domain), zap.Error(err))
		status.SSLStatus = SSLStatus{
			IsValid:     false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}
	} else {
		status.SSLStatus = *sslStatus
	}

	// For now, we'll consider the domain verified if DNS resolves
	// In a production system, you'd want more robust verification
	status.IsVerified = true

	return nil
}

// checkDNSResolution checks if the domain resolves to our servers
func (s *CustomDomainService) checkDNSResolution(domain string) error {
	// Check A record
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("failed to resolve domain: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for domain")
	}

	// In a real implementation, you'd check if the IPs point to your servers
	// For now, we'll just verify that the domain resolves
	logger.Info("Domain DNS resolution successful", zap.String("domain", domain), zap.Strings("ips", s.ipsToStrings(ips)))

	return nil
}

// checkSSLStatus checks the SSL certificate status
func (s *CustomDomainService) checkSSLStatus(domain string) (*SSLStatus, error) {
	// Create a custom transport with timeout
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	// Try HTTPS first
	url := fmt.Sprintf("https://%s", domain)
	resp, err := client.Get(url)
	if err != nil {
		// If HTTPS fails, try HTTP
		url = fmt.Sprintf("http://%s", domain)
		resp, err = client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to domain: %w", err)
		}
	}
	defer resp.Body.Close()

	// Get TLS connection state if available
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		return &SSLStatus{
			IsValid:     time.Now().Before(cert.NotAfter),
			ExpiresAt:   cert.NotAfter,
			Issuer:      cert.Issuer.String(),
			CommonName:  cert.Subject.CommonName,
			LastChecked: time.Now(),
		}, nil
	}

	// No SSL certificate found
	return &SSLStatus{
		IsValid:     false,
		LastChecked: time.Now(),
		Error:       "No SSL certificate found",
	}, nil
}

// GetDomainStatus returns the current status of a domain
func (s *CustomDomainService) GetDomainStatus(ctx context.Context, tenantID uint) (*DomainVerificationStatus, error) {
	var tenant models.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.Domain == "" {
		return nil, fmt.Errorf("no custom domain configured for tenant")
	}

	return s.VerifyDomain(ctx, tenant.Domain)
}

// RemoveCustomDomain removes a custom domain from a tenant
func (s *CustomDomainService) RemoveCustomDomain(ctx context.Context, tenantID uint) error {
	var tenant models.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	tenant.Domain = ""
	if err := s.db.Save(&tenant).Error; err != nil {
		return fmt.Errorf("failed to remove custom domain: %w", err)
	}

	return nil
}

// GenerateDNSInstructions generates DNS setup instructions for a domain
func (s *CustomDomainService) GenerateDNSInstructions(domain string) map[string]interface{} {
	// In a real implementation, you'd get the actual server IPs
	// For now, we'll provide example instructions
	return map[string]interface{}{
		"domain": domain,
		"instructions": []map[string]string{
			{
				"type":  "A",
				"name":  "@",
				"value": "YOUR_SERVER_IP",
				"ttl":   "300",
			},
			{
				"type":  "CNAME",
				"name":  "www",
				"value": domain,
				"ttl":   "300",
			},
		},
		"verification": map[string]string{
			"type":  "TXT",
			"name":  "_statuspage-verification",
			"value": fmt.Sprintf("statuspage-verification=%s", generateVerificationToken()),
		},
	}
}

// GenerateSSLInstructions generates SSL setup instructions
func (s *CustomDomainService) GenerateSSLInstructions(domain string) map[string]interface{} {
	return map[string]interface{}{
		"domain": domain,
		"options": []map[string]string{
			{
				"name":        "Let's Encrypt (Recommended)",
				"description": "Free SSL certificate with automatic renewal",
				"setup":       "Automatic setup via our platform",
			},
			{
				"name":        "Custom Certificate",
				"description": "Upload your own SSL certificate",
				"setup":       "Manual upload via admin panel",
			},
		},
		"automatic_setup": true,
		"certificate_upload": map[string]string{
			"endpoint": "/api/v1/tenant/ssl/upload",
			"method":   "POST",
			"format":   "PEM",
		},
	}
}

// Helper functions

// isValidDomain validates domain format
func (s *CustomDomainService) isValidDomain(domain string) bool {
	// Basic domain validation
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Check for valid characters
	if strings.ContainsAny(domain, " \t\n\r") {
		return false
	}

	// Must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	// Check each label
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}

	return true
}

// ipsToStrings converts IP addresses to strings
func (s *CustomDomainService) ipsToStrings(ips []net.IP) []string {
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result
}

// generateVerificationToken generates a verification token
func generateVerificationToken() string {
	// In a real implementation, you'd generate a secure random token
	return fmt.Sprintf("verify-%d", time.Now().Unix())
}
