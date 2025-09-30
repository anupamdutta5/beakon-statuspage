// Package services provides business logic for custom domain management.
package services

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
)

// CustomDomainService handles custom domain operations.
type CustomDomainService struct {
	db                 *gorm.DB
	defaultDomain      string
	defaultPort        string
	acmeClient         ACMEClient
	dnsProvider        DNSProvider
	enableSSL          bool
	enableVerification bool
}

// ACMEClient interface for SSL certificate management.
type ACMEClient interface {
	ObtainCertificate(domain string, challenge string) (*models.SSLCertificate, error)
	RenewCertificate(cert *models.SSLCertificate) (*models.SSLCertificate, error)
	RevokeCertificate(cert *models.SSLCertificate) error
}

// DNSProvider interface for DNS operations.
type DNSProvider interface {
	LookupCNAME(domain string) (string, error)
	LookupTXT(domain string) ([]string, error)
	LookupA(domain string) ([]net.IP, error)
	VerifyDNSRecord(domain, recordType, expectedValue string) (bool, error)
}

// NewCustomDomainService creates a new custom domain service.
func NewCustomDomainService(db *gorm.DB, defaultDomain, defaultPort string) *CustomDomainService {
	return &CustomDomainService{
		db:                 db,
		defaultDomain:      defaultDomain,
		defaultPort:        defaultPort,
		enableSSL:          true,
		enableVerification: true,
	}
}

// CreateCustomDomain creates a new custom domain configuration.
func (s *CustomDomainService) CreateCustomDomain(tenantID, domain string) (*models.CustomDomain, error) {
	// Validate domain format
	if err := s.ValidateDomainFormat(domain); err != nil {
		return nil, fmt.Errorf("invalid domain format: %w", err)
	}

	// Check if domain already exists
	var existing models.CustomDomain
	if err := s.db.Where("domain = ?", domain).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("domain already exists: %s", domain)
	}

	// Generate verification key and value
	verificationKey, verificationValue, err := s.generateVerificationChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification challenge: %w", err)
	}

	// Create custom domain record
	customDomain := &models.CustomDomain{
		ID:                uuid.New(),
		TenantID:          tenantID,
		Domain:            domain,
		Status:            "pending",
		VerificationKey:   verificationKey,
		VerificationValue: verificationValue,
		VerificationType:  "dns",
		DNSRecordType:     "CNAME",
		DNSRecordValue:    fmt.Sprintf("%s.%s", tenantID, s.defaultDomain),
		SSLStatus:         "pending",
		SSLProvider:       "lets_encrypt",
		SSLAutoRenew:      true,
		IsActive:          false,
		IsWildcard:        strings.HasPrefix(domain, "*."),
		MaxTries:          5,
	}

	if err := s.db.Create(customDomain).Error; err != nil {
		return nil, fmt.Errorf("failed to create custom domain: %w", err)
	}

	// Start verification process asynchronously
	go s.startDomainVerification(customDomain)

	return customDomain, nil
}

// ValidateDomainFormat validates the domain name format.
func (s *CustomDomainService) ValidateDomainFormat(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// Basic domain regex validation
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format")
	}

	// Check for reserved domains
	reservedDomains := []string{"localhost", "example.com", "test.com", "invalid"}
	for _, reserved := range reservedDomains {
		if strings.Contains(domain, reserved) {
			return fmt.Errorf("domain contains reserved word: %s", reserved)
		}
	}

	// Check length constraints
	if len(domain) > 253 {
		return fmt.Errorf("domain too long (max 253 characters)")
	}

	return nil
}

// VerifyCustomDomain verifies domain ownership.
func (s *CustomDomainService) VerifyCustomDomain(domainID uuid.UUID) error {
	var customDomain models.CustomDomain
	if err := s.db.First(&customDomain, "id = ?", domainID).Error; err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	// Check if already verified
	if customDomain.Status == "verified" {
		return fmt.Errorf("domain already verified")
	}

	// Check retry limits
	if !customDomain.CanRetryVerification() {
		return fmt.Errorf("maximum verification attempts exceeded")
	}

	// Perform DNS verification
	verified, err := s.verifyDNSRecord(&customDomain)
	if err != nil {
		customDomain.VerificationTries++
		s.db.Save(&customDomain)
		return fmt.Errorf("verification failed: %w", err)
	}

	if verified {
		customDomain.Status = "verified"
		customDomain.LastVerified = &time.Time{}
		*customDomain.LastVerified = time.Now()

		if err := s.db.Save(&customDomain).Error; err != nil {
			return fmt.Errorf("failed to update domain status: %w", err)
		}

		// Start SSL provisioning if enabled
		if s.enableSSL {
			go s.provisionSSLCertificate(&customDomain)
		}

		return nil
	}

	customDomain.VerificationTries++
	s.db.Save(&customDomain)
	return fmt.Errorf("domain verification failed")
}

// GetTenantDomains retrieves all domains for a tenant.
func (s *CustomDomainService) GetTenantDomains(tenantID string) ([]models.CustomDomain, error) {
	var domains []models.CustomDomain
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&domains).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve domains: %w", err)
	}
	return domains, nil
}

// GetDefaultDomainForTenant generates the default domain for a tenant.
func (s *CustomDomainService) GetDefaultDomainForTenant(tenantID string) string {
	if s.defaultPort != "" && s.defaultPort != "80" && s.defaultPort != "443" {
		return fmt.Sprintf("%s.%s:%s", tenantID, s.defaultDomain, s.defaultPort)
	}
	return fmt.Sprintf("%s.%s", tenantID, s.defaultDomain)
}

// ResolveTenantFromHost extracts tenant ID from host header.
func (s *CustomDomainService) ResolveTenantFromHost(host string) (string, error) {
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
		host = host[:colonIndex]
	}

	// Check for custom domain first
	var customDomain models.CustomDomain
	if err := s.db.Where("domain = ? AND is_active = ?", host, true).First(&customDomain).Error; err == nil {
		return customDomain.TenantID, nil
	}

	// Check for default domain pattern (tenant.domain.com)
	if strings.HasSuffix(host, "."+s.defaultDomain) {
		tenantID := strings.TrimSuffix(host, "."+s.defaultDomain)
		if tenantID != "" {
			return tenantID, nil
		}
	}

	return "", fmt.Errorf("unable to resolve tenant from host: %s", host)
}

// DeleteCustomDomain removes a custom domain configuration.
func (s *CustomDomainService) DeleteCustomDomain(domainID uuid.UUID) error {
	var customDomain models.CustomDomain
	if err := s.db.First(&customDomain, "id = ?", domainID).Error; err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	// Revoke SSL certificate if active
	if customDomain.SSLStatus == "active" && s.acmeClient != nil {
		var cert models.SSLCertificate
		if err := s.db.Where("domain_id = ? AND is_active = ?", domainID, true).First(&cert).Error; err == nil {
			s.acmeClient.RevokeCertificate(&cert)
		}
	}

	// Soft delete the domain
	if err := s.db.Delete(&customDomain).Error; err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	return nil
}

// CheckSSLCertificateStatus checks SSL certificate status and schedules renewal if needed.
func (s *CustomDomainService) CheckSSLCertificateStatus(domainID uuid.UUID) (*models.SSLCertificate, error) {
	var cert models.SSLCertificate
	if err := s.db.Where("domain_id = ? AND is_active = ?", domainID, true).First(&cert).Error; err != nil {
		return nil, fmt.Errorf("SSL certificate not found: %w", err)
	}

	// Check if certificate needs renewal
	if cert.IsExpiringSoon(30) && cert.AutoRenew {
		go s.renewSSLCertificate(&cert)
	}

	return &cert, nil
}

// Internal helper methods

func (s *CustomDomainService) generateVerificationChallenge() (string, string, error) {
	// Generate random verification key
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", "", err
	}

	verificationKey := fmt.Sprintf("_statuspage-challenge-%x", key)
	verificationValue := fmt.Sprintf("statuspage-domain-verification=%x", key)

	return verificationKey, verificationValue, nil
}

func (s *CustomDomainService) startDomainVerification(domain *models.CustomDomain) {
	// Wait a bit before starting verification to allow DNS propagation
	time.Sleep(30 * time.Second)

	for i := 0; i < domain.MaxTries; i++ {
		if err := s.VerifyCustomDomain(domain.ID); err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Minute) // Exponential backoff
	}
}

func (s *CustomDomainService) verifyDNSRecord(domain *models.CustomDomain) (bool, error) {
	if s.dnsProvider == nil {
		// For development/testing, skip actual DNS verification
		return true, nil
	}

	switch domain.DNSRecordType {
	case "CNAME":
		cname, err := s.dnsProvider.LookupCNAME(domain.Domain)
		if err != nil {
			return false, err
		}
		return strings.EqualFold(cname, domain.DNSRecordValue), nil

	case "TXT":
		txtRecords, err := s.dnsProvider.LookupTXT(fmt.Sprintf("%s.%s", domain.VerificationKey, domain.Domain))
		if err != nil {
			return false, err
		}
		for _, record := range txtRecords {
			if record == domain.VerificationValue {
				return true, nil
			}
		}
		return false, nil

	default:
		return false, fmt.Errorf("unsupported DNS record type: %s", domain.DNSRecordType)
	}
}

func (s *CustomDomainService) provisionSSLCertificate(domain *models.CustomDomain) {
	if s.acmeClient == nil {
		// For development/testing, create a mock certificate
		s.createMockSSLCertificate(domain)
		return
	}

	cert, err := s.acmeClient.ObtainCertificate(domain.Domain, "http-01")
	if err != nil {
		domain.SSLStatus = "failed"
		s.db.Save(domain)
		return
	}

	// Save certificate to database
	if err := s.db.Create(cert).Error; err != nil {
		domain.SSLStatus = "failed"
		s.db.Save(domain)
		return
	}

	// Update domain SSL status
	domain.SSLStatus = "active"
	domain.SSLIssueDate = &cert.IssuedAt
	domain.SSLExpiryDate = &cert.ExpiresAt
	domain.IsActive = true
	s.db.Save(domain)
}

func (s *CustomDomainService) createMockSSLCertificate(domain *models.CustomDomain) {
	now := time.Now()
	expiry := now.AddDate(0, 3, 0) // 3 months validity

	cert := &models.SSLCertificate{
		ID:              uuid.New(),
		DomainID:        domain.ID,
		Provider:        "mock",
		CertificateData: "-----BEGIN CERTIFICATE-----\nMOCK_CERTIFICATE_DATA\n-----END CERTIFICATE-----",
		PrivateKeyData:  "-----BEGIN PRIVATE KEY-----\nMOCK_PRIVATE_KEY_DATA\n-----END PRIVATE KEY-----",
		SerialNumber:    fmt.Sprintf("MOCK-%d", now.Unix()),
		Fingerprint:     "mock:fingerprint",
		Algorithm:       "RSA",
		KeySize:         2048,
		IssuedAt:        now,
		ExpiresAt:       expiry,
		IsActive:        true,
		AutoRenew:       true,
		RenewalDays:     30,
	}

	s.db.Create(cert)

	// Update domain SSL status
	domain.SSLStatus = "active"
	domain.SSLIssueDate = &now
	domain.SSLExpiryDate = &expiry
	domain.IsActive = true
	s.db.Save(domain)
}

func (s *CustomDomainService) renewSSLCertificate(cert *models.SSLCertificate) {
	if s.acmeClient == nil {
		return
	}

	newCert, err := s.acmeClient.RenewCertificate(cert)
	if err != nil {
		cert.RenewalStatus = "failed"
		s.db.Save(cert)
		return
	}

	// Deactivate old certificate
	cert.IsActive = false
	cert.RenewalStatus = "completed"
	s.db.Save(cert)

	// Save new certificate
	s.db.Create(newCert)
}

// GetDomainStats returns domain usage statistics.
func (s *CustomDomainService) GetDomainStats() (*models.DomainStats, error) {
	stats := &models.DomainStats{
		LastUpdated: time.Now(),
	}

	// Count total domains
	var totalDomains int64
	s.db.Model(&models.CustomDomain{}).Count(&totalDomains)
	stats.TotalDomains = int(totalDomains)

	// Count by status
	var activeDomains int64
	s.db.Model(&models.CustomDomain{}).Where("is_active = ?", true).Count(&activeDomains)
	stats.ActiveDomains = int(activeDomains)

	var pendingDomains int64
	s.db.Model(&models.CustomDomain{}).Where("status = ?", "pending").Count(&pendingDomains)
	stats.PendingDomains = int(pendingDomains)

	var verifiedDomains int64
	s.db.Model(&models.CustomDomain{}).Where("status = ?", "verified").Count(&verifiedDomains)
	stats.VerifiedDomains = int(verifiedDomains)

	var failedDomains int64
	s.db.Model(&models.CustomDomain{}).Where("status = ?", "failed").Count(&failedDomains)
	stats.FailedDomains = int(failedDomains)

	// Count SSL stats
	var sslActiveDomains int64
	s.db.Model(&models.CustomDomain{}).Where("ssl_status = ?", "active").Count(&sslActiveDomains)
	stats.SSLActiveDomains = int(sslActiveDomains)

	var sslPendingDomains int64
	s.db.Model(&models.CustomDomain{}).Where("ssl_status = ?", "pending").Count(&sslPendingDomains)
	stats.SSLPendingDomains = int(sslPendingDomains)

	var sslExpiredDomains int64
	s.db.Model(&models.CustomDomain{}).Where("ssl_status = ? OR ssl_expiry_date < ?", "expired", time.Now()).Count(&sslExpiredDomains)
	stats.SSLExpiredDomains = int(sslExpiredDomains)

	return stats, nil
}

// ValidateSSLCertificate validates SSL certificate for a domain.
func (s *CustomDomainService) ValidateSSLCertificate(domain string) (bool, error) {
	// Check if we can establish a TLS connection
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:443", domain), &tls.Config{
		ServerName: domain,
	})
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Check certificate validity
	cert := conn.ConnectionState().PeerCertificates[0]
	now := time.Now()

	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return false, fmt.Errorf("certificate not valid for current time")
	}

	// Verify hostname
	if err := cert.VerifyHostname(domain); err != nil {
		return false, err
	}

	return true, nil
}

// GetDB returns the database instance for direct access in handlers
func (s *CustomDomainService) GetDB() *gorm.DB {
	return s.db
}

// GetDomainByID retrieves a custom domain by its ID
func (s *CustomDomainService) GetDomainByID(id string) (*models.CustomDomain, error) {
	var domain models.CustomDomain
	if err := s.db.First(&domain, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}