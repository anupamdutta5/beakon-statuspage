// Package services provides business logic for the Monitoring Service.
package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/models"
)

// SSLScannerService handles SSL certificate discovery, validation, and expiration monitoring.
type SSLScannerService struct {
	db *gorm.DB
}

// NewSSLScannerService creates a new SSL scanner service instance.
func NewSSLScannerService(db *gorm.DB) *SSLScannerService {
	return &SSLScannerService{
		db: db,
	}
}

// ScanDomain scans a domain for its SSL certificate and stores/updates the information.
func (s *SSLScannerService) ScanDomain(tenantID uuid.UUID, domain string) (*models.SSLCertificate, error) {
	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Split(domain, "/")[0] // Remove path if present

	// Fetch the certificate
	cert, err := s.fetchCertificate(domain)
	if err != nil {
		// Store error in database
		return s.saveErrorCertificate(tenantID, domain, err.Error())
	}

	// Calculate days until expiry
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

	// Extract certificate information
	sslCert := &models.SSLCertificate{
		TenantID:        tenantID,
		Domain:          domain,
		Issuer:          cert.Issuer.CommonName,
		Subject:         cert.Subject.CommonName,
		SerialNumber:    cert.SerialNumber.String(),
		ValidFrom:       &cert.NotBefore,
		ValidUntil:      &cert.NotAfter,
		DaysUntilExpiry: &daysUntilExpiry,
		LastChecked:     timePtr(time.Now()),
		IsValid:         time.Now().After(cert.NotBefore) && time.Now().Before(cert.NotAfter),
		IsSelfSigned:    s.isSelfSigned(cert),
	}

	// Save or update in database
	if err := s.upsertCertificate(sslCert); err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	return sslCert, nil
}

// fetchCertificate connects to the domain and retrieves its SSL certificate.
func (s *SSLScannerService) fetchCertificate(domain string) (*x509.Certificate, error) {
	// Add default HTTPS port if not specified
	if !strings.Contains(domain, ":") {
		domain = domain + ":443"
	}

	// Set connection timeout
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	// Connect with TLS
	conn, err := tls.DialWithDialer(dialer, "tcp", domain, &tls.Config{
		InsecureSkipVerify: false, // Verify certificates
	})
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	// Get peer certificates
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	// Return the first certificate (server certificate)
	return certs[0], nil
}

// isSelfSigned checks if a certificate is self-signed.
func (s *SSLScannerService) isSelfSigned(cert *x509.Certificate) bool {
	return cert.Issuer.CommonName == cert.Subject.CommonName
}

// upsertCertificate inserts or updates an SSL certificate in the database.
func (s *SSLScannerService) upsertCertificate(cert *models.SSLCertificate) error {
	var existing models.SSLCertificate

	err := s.db.Where("tenant_id = ? AND domain = ?", cert.TenantID, cert.Domain).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Insert new certificate
		return s.db.Create(cert).Error
	} else if err != nil {
		return err
	}

	// Update existing certificate
	cert.ID = existing.ID
	cert.CreatedAt = existing.CreatedAt

	// Preserve warning flags if certificate hasn't changed
	if existing.SerialNumber == cert.SerialNumber {
		cert.WarningSent30d = existing.WarningSent30d
		cert.WarningSent14d = existing.WarningSent14d
		cert.WarningSent7d = existing.WarningSent7d
	} else {
		// Reset warnings if certificate was renewed
		cert.WarningSent30d = false
		cert.WarningSent14d = false
		cert.WarningSent7d = false
	}

	return s.db.Save(cert).Error
}

// saveErrorCertificate stores a certificate record with error information.
func (s *SSLScannerService) saveErrorCertificate(tenantID uuid.UUID, domain string, errorMsg string) (*models.SSLCertificate, error) {
	var existing models.SSLCertificate

	err := s.db.Where("tenant_id = ? AND domain = ?", tenantID, domain).
		First(&existing).Error

	cert := &models.SSLCertificate{
		TenantID:     tenantID,
		Domain:       domain,
		LastChecked:  timePtr(time.Now()),
		IsValid:      false,
		ErrorMessage: errorMsg,
	}

	if err == gorm.ErrRecordNotFound {
		// Create new error record
		if err := s.db.Create(cert).Error; err != nil {
			return nil, err
		}
		return cert, nil
	}

	// Update existing with error
	cert.ID = existing.ID
	cert.CreatedAt = existing.CreatedAt
	if err := s.db.Save(cert).Error; err != nil {
		return nil, err
	}

	return cert, nil
}

// GetExpiringCertificates returns certificates expiring within the specified number of days.
func (s *SSLScannerService) GetExpiringCertificates(tenantID uuid.UUID, days int) ([]models.SSLCertificate, error) {
	var certs []models.SSLCertificate

	err := s.db.Where("tenant_id = ? AND is_valid = ? AND days_until_expiry <= ? AND days_until_expiry >= 0",
		tenantID, true, days).
		Order("days_until_expiry ASC").
		Find(&certs).Error

	return certs, err
}

// GetCertificatesNeedingWarning returns certificates that need expiration warnings sent.
func (s *SSLScannerService) GetCertificatesNeedingWarning() ([]models.SSLCertificate, error) {
	var certs []models.SSLCertificate

	// Find certificates expiring within 30 days that haven't had all warnings sent
	err := s.db.Where(`
		is_valid = ?
		AND days_until_expiry <= 30
		AND days_until_expiry >= 0
		AND (
			(days_until_expiry <= 30 AND days_until_expiry > 14 AND warning_sent_30d = ?)
			OR (days_until_expiry <= 14 AND days_until_expiry > 7 AND warning_sent_14d = ?)
			OR (days_until_expiry <= 7 AND warning_sent_7d = ?)
		)`,
		true, false, false, false,
	).Find(&certs).Error

	return certs, err
}

// MarkWarningSent marks a warning as sent for a certificate.
func (s *SSLScannerService) MarkWarningSent(certID uint, warningType string) error {
	updates := map[string]interface{}{}

	switch warningType {
	case "30d":
		updates["warning_sent_30d"] = true
	case "14d":
		updates["warning_sent_14d"] = true
	case "7d":
		updates["warning_sent_7d"] = true
	default:
		return fmt.Errorf("invalid warning type: %s", warningType)
	}

	return s.db.Model(&models.SSLCertificate{}).
		Where("id = ?", certID).
		Updates(updates).Error
}

// ScanAllTenantDomains scans all domains for a specific tenant.
func (s *SSLScannerService) ScanAllTenantDomains(tenantID uuid.UUID, domains []string) []error {
	var errors []error

	for _, domain := range domains {
		if _, err := s.ScanDomain(tenantID, domain); err != nil {
			errors = append(errors, fmt.Errorf("domain %s: %w", domain, err))
		}
	}

	return errors
}

// GetCertificate retrieves a certificate by ID.
func (s *SSLScannerService) GetCertificate(id uint) (*models.SSLCertificate, error) {
	var cert models.SSLCertificate
	err := s.db.First(&cert, id).Error
	return &cert, err
}

// GetTenantCertificates retrieves all certificates for a tenant.
func (s *SSLScannerService) GetTenantCertificates(tenantID uuid.UUID) ([]models.SSLCertificate, error) {
	var certs []models.SSLCertificate
	err := s.db.Where("tenant_id = ?", tenantID).
		Order("days_until_expiry ASC").
		Find(&certs).Error
	return certs, err
}

// DeleteCertificate deletes a certificate by ID.
func (s *SSLScannerService) DeleteCertificate(id uint) error {
	return s.db.Delete(&models.SSLCertificate{}, id).Error
}

// RescanExpiredCertificates rescans certificates that have been checked more than 24 hours ago.
func (s *SSLScannerService) RescanExpiredCertificates() (int, []error) {
	var certs []models.SSLCertificate
	var errors []error
	scannedCount := 0

	// Find certificates not checked in last 24 hours
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	err := s.db.Where("last_checked < ? OR last_checked IS NULL", twentyFourHoursAgo).
		Find(&certs).Error

	if err != nil {
		return 0, []error{err}
	}

	for _, cert := range certs {
		if _, err := s.ScanDomain(cert.TenantID, cert.Domain); err != nil {
			errors = append(errors, fmt.Errorf("domain %s: %w", cert.Domain, err))
		} else {
			scannedCount++
		}
	}

	return scannedCount, errors
}

// Helper function to create time pointer.
func timePtr(t time.Time) *time.Time {
	return &t
}
