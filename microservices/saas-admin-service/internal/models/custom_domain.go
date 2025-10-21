// Package models provides data models for custom domain management.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomDomain represents a custom domain configuration for a tenant.
type CustomDomain struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID          string         `gorm:"not null;index" json:"tenant_id"`
	Domain            string         `gorm:"not null;uniqueIndex" json:"domain"`
	Status            string         `gorm:"default:pending" json:"status"` // pending, verified, active, failed, suspended
	VerificationKey   string         `gorm:"not null" json:"verification_key"`
	VerificationValue string         `gorm:"not null" json:"verification_value"`
	VerificationType  string         `gorm:"default:dns" json:"verification_type"` // dns, http, email
	DNSRecordType     string         `gorm:"default:CNAME" json:"dns_record_type"` // CNAME, A, AAAA
	DNSRecordValue    string         `json:"dns_record_value"`                     // Target for CNAME or IP for A record
	SSLStatus         string         `gorm:"default:pending" json:"ssl_status"`    // pending, issued, active, expired, failed
	SSLProvider       string         `gorm:"default:lets_encrypt" json:"ssl_provider"` // lets_encrypt, custom, none
	SSLIssueDate      *time.Time     `json:"ssl_issue_date"`
	SSLExpiryDate     *time.Time     `json:"ssl_expiry_date"`
	SSLAutoRenew      bool           `gorm:"default:true" json:"ssl_auto_renew"`
	LastVerified      *time.Time     `json:"last_verified"`
	VerificationTries int            `gorm:"default:0" json:"verification_tries"`
	MaxTries          int            `gorm:"default:5" json:"max_tries"`
	IsActive          bool           `gorm:"default:false;index" json:"is_active"`
	IsWildcard        bool           `gorm:"default:false" json:"is_wildcard"`
	Config            string         `gorm:"type:text" json:"config"`   // JSON configuration
	Metadata          string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// DomainVerificationChallenge represents a domain ownership verification challenge.
type DomainVerificationChallenge struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DomainID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"domain_id"`
	Domain        CustomDomain   `gorm:"foreignKey:DomainID" json:"domain"`
	Type          string         `gorm:"not null;index" json:"type"`    // dns-01, http-01, tls-alpn-01
	Token         string         `gorm:"not null" json:"token"`         // Challenge token
	KeyAuth       string         `gorm:"not null" json:"key_auth"`      // Key authorization
	Value         string         `gorm:"not null" json:"value"`         // Expected value/content
	Status        string         `gorm:"default:pending" json:"status"` // pending, processing, valid, invalid
	ValidatedAt   *time.Time     `json:"validated_at"`
	ExpiresAt     time.Time      `gorm:"not null" json:"expires_at"`
	RetryCount    int            `gorm:"default:0" json:"retry_count"`
	LastError     string         `gorm:"type:text" json:"last_error"`
	ChallengeData string         `gorm:"type:text" json:"challenge_data"` // JSON data for challenge specifics
}

// SSLCertificate represents SSL certificate information.
type SSLCertificate struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DomainID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"domain_id"`
	Domain          CustomDomain   `gorm:"foreignKey:DomainID" json:"domain"`
	Provider        string         `gorm:"not null" json:"provider"`        // lets_encrypt, custom, cloudflare
	CertificateData string         `gorm:"type:text;not null" json:"certificate_data"` // PEM encoded cert
	PrivateKeyData  string         `gorm:"type:text;not null" json:"private_key_data"`  // PEM encoded private key
	ChainData       string         `gorm:"type:text" json:"chain_data"`     // Certificate chain
	SerialNumber    string         `gorm:"uniqueIndex" json:"serial_number"`
	Fingerprint     string         `json:"fingerprint"`
	Algorithm       string         `json:"algorithm"`
	KeySize         int            `json:"key_size"`
	IssuedAt        time.Time      `gorm:"not null" json:"issued_at"`
	ExpiresAt       time.Time      `gorm:"not null;index" json:"expires_at"`
	IsActive        bool           `gorm:"default:false;index" json:"is_active"`
	AutoRenew       bool           `gorm:"default:true" json:"auto_renew"`
	RenewalDays     int            `gorm:"default:30" json:"renewal_days"` // Days before expiry to renew
	LastRenewal     *time.Time     `json:"last_renewal"`
	RenewalStatus   string         `gorm:"default:none" json:"renewal_status"` // none, scheduled, processing, completed, failed
	SubjectAltNames string         `gorm:"type:text" json:"subject_alt_names"` // JSON array of SANs
	OCSP            string         `json:"ocsp"`                               // OCSP responder URL
	Metadata        string         `gorm:"type:text" json:"metadata"`         // JSON string for additional data
}

// DomainStats represents domain usage statistics.
type DomainStats struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	TotalDomains      int       `json:"total_domains"`
	ActiveDomains     int       `json:"active_domains"`
	PendingDomains    int       `json:"pending_domains"`
	VerifiedDomains   int       `json:"verified_domains"`
	FailedDomains     int       `json:"failed_domains"`
	SSLActiveDomains  int       `json:"ssl_active_domains"`
	SSLPendingDomains int       `json:"ssl_pending_domains"`
	SSLExpiredDomains int       `json:"ssl_expired_domains"`
	LastUpdated       time.Time `json:"last_updated"`
}

// TableName returns the table name for CustomDomain.
func (CustomDomain) TableName() string {
	return "custom_domains"
}

// TableName returns the table name for DomainVerificationChallenge.
func (DomainVerificationChallenge) TableName() string {
	return "domain_verification_challenges"
}

// TableName returns the table name for SSLCertificate.
func (SSLCertificate) TableName() string {
	return "ssl_certificates"
}

// TableName returns the table name for DomainStats.
func (DomainStats) TableName() string {
	return "domain_stats"
}

// IsExpiringSoon checks if SSL certificate expires within specified days.
func (ssl *SSLCertificate) IsExpiringSoon(days int) bool {
	if days <= 0 {
		days = ssl.RenewalDays
	}
	expiry := ssl.ExpiresAt
	renewalDate := expiry.AddDate(0, 0, -days)
	return time.Now().After(renewalDate)
}

// IsExpired checks if SSL certificate is expired.
func (ssl *SSLCertificate) IsExpired() bool {
	return time.Now().After(ssl.ExpiresAt)
}

// CanRetryVerification checks if domain verification can be retried.
func (cd *CustomDomain) CanRetryVerification() bool {
	return cd.VerificationTries < cd.MaxTries && cd.Status != "verified"
}

// NeedsRenewal checks if SSL certificate needs renewal.
func (cd *CustomDomain) NeedsRenewal() bool {
	return cd.SSLExpiryDate != nil &&
		time.Now().AddDate(0, 0, 30).After(*cd.SSLExpiryDate) &&
		cd.SSLAutoRenew
}