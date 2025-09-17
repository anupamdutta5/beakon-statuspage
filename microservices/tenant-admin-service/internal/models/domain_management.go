package models

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// DomainVerificationMethod represents the method used to verify domain ownership
type DomainVerificationMethod string

const (
	// DNSVerification verifies domain ownership via DNS TXT record
	DNSVerification DomainVerificationMethod = "dns_txt"
	// HTTPVerification verifies domain ownership via HTTP request
	HTTPVerification DomainVerificationMethod = "http"
	// ManualVerification requires manual verification
	ManualVerification DomainVerificationMethod = "manual"
)

// DomainStatus represents the status of a domain
//
//go:generate stringer -type=DomainStatus
type DomainStatus int

const (
	// DomainPendingVerification means the domain is pending verification
	DomainPendingVerification DomainStatus = iota + 1
	// DomainVerified means the domain has been verified
	DomainVerified
	// DomainFailedVerification means domain verification failed
	DomainFailedVerification
	// DomainDisabled means the domain has been disabled
	DomainDisabled
)

// Domain represents a custom domain for a status page
type Domain struct {
	ID                 uint                   `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          gorm.DeletedAt         `gorm:"index" json:"deleted_at,omitempty"`
	TenantID           uint                   `gorm:"not null;index" json:"tenant_id"`
	StatusPageID       uint                   `gorm:"not null;index" json:"status_page_id"`
	Domain             string                 `gorm:"not null;uniqueIndex" json:"domain"`
	Status             DomainStatus           `gorm:"not null;default:1" json:"status"` // 1 = pending, 2 = verified, 3 = failed, 4 = disabled
	VerificationMethod DomainVerificationMethod `gorm:"not null" json:"verification_method"`
	VerificationToken  string                 `gorm:"not null" json:"-"` // Don't expose this in API responses
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	IsPrimary          bool                   `gorm:"default:false" json:"is_primary"`
	SSLEnabled         bool                   `gorm:"default:false" json:"ssl_enabled"`
	SSLCertIssuedAt    *time.Time             `json:"ssl_cert_issued_at,omitempty"`
	SSLCertExpiresAt   *time.Time             `json:"ssl_cert_expires_at,omitempty"`
	SSLCertIssuer      string                 `json:"ssl_cert_issuer,omitempty"`
	SSLCertCommonName  string                 `json:"ssl_cert_common_name,omitempty"`
	SSLCertSANs        []string               `gorm:"-" json:"ssl_cert_sans,omitempty"`
	SSLCertRaw         string                 `gorm:"type:text" json:"-"` // Raw certificate data
	Metadata           string                 `gorm:"type:text" json:"metadata"`
}

// TableName returns the table name for Domain
func (Domain) TableName() string {
	return "status_page_domains"
}

// BeforeCreate is a hook that runs before creating a domain
func (d *Domain) BeforeCreate(tx *gorm.DB) error {
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// BeforeUpdate is a hook that runs before updating a domain
func (d *Domain) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate performs validation on Domain
func (d *Domain) Validate() error {
	if d.TenantID == 0 {
		return errors.New("tenant ID is required")
	}

	if d.StatusPageID == 0 {
		return errors.New("status page ID is required")
	}

	if d.Domain == "" {
		return errors.New("domain is required")
	}

	if d.VerificationToken == "" && d.Status == DomainPendingVerification {
		return errors.New("verification token is required for pending domains")
	}

	return nil
}

// GenerateVerificationToken generates a unique verification token for the domain
func (d *Domain) GenerateVerificationToken() {
	d.VerificationToken = "beakon_verify_" + time.Now().Format("20060102150405")
}

// Verify checks if the domain has been verified
func (d *Domain) Verify(token string) error {
	if d.Status == DomainVerified {
		return nil // Already verified
	}

	if d.VerificationToken != token {
		return errors.New("invalid verification token")
	}

	now := time.Now().UTC()
	d.Status = DomainVerified
	d.VerifiedAt = &now
	d.VerificationToken = "" // Clear the token after verification

	return nil
}

// UpdateSSLInfo updates the SSL certificate information
func (d *Domain) UpdateSSLInfo(issuer, commonName string, sans []string, issuedAt, expiresAt time.Time, certData string) {
	d.SSLEnabled = true
	d.SSLCertIssuer = issuer
	d.SSLCertCommonName = commonName
	d.SSLCertSANs = sans
	d.SSLCertIssuedAt = &issuedAt
	d.SSLCertExpiresAt = &expiresAt
	d.SSLCertRaw = certData
}

// DomainVerificationRecord represents a domain verification record for DNS or HTTP verification
type DomainVerificationRecord struct {
	Type  string `json:"type"`  // "TXT" for DNS, "file" for HTTP
	Name  string `json:"name"`  // Record name (e.g., "_beakon-verification.example.com")
	Value string `json:"value"` // Record value (e.g., "beakon-verification=abc123")
}

// GetVerificationRecord returns the verification record for the domain
func (d *Domain) GetVerificationRecord() DomainVerificationRecord {
	switch d.VerificationMethod {
	case DNSVerification:
		return DomainVerificationRecord{
			Type:  "TXT",
			Name:  "_beakon-verification.",
			Value: d.VerificationToken,
		}
	case HTTPVerification:
		return DomainVerificationRecord{
			Type:  "file",
			Name:  "/.well-known/beakon-verification/" + d.VerificationToken,
			Value: d.VerificationToken,
		}
	default:
		return DomainVerificationRecord{}
	}
}

// DomainListOptions represents options for listing domains
type DomainListOptions struct {
	TenantID     uint
	StatusPageID uint
	Status       DomainStatus
	Search       string
	Page         int
	PageSize     int
}

// DomainRepository defines the interface for domain data access
type DomainRepository interface {
	Create(ctx context.Context, domain *Domain) error
	GetByID(ctx context.Context, id uint) (*Domain, error)
	GetByDomain(ctx context.Context, domain string) (*Domain, error)
	List(ctx context.Context, opts DomainListOptions) ([]*Domain, int64, error)
	Update(ctx context.Context, domain *Domain) error
	Delete(ctx context.Context, id uint) error
	VerifyDomain(ctx context.Context, domain, token string) error
	SetPrimaryDomain(ctx context.Context, tenantID, statusPageID, domainID uint) error
	GetPrimaryDomain(ctx context.Context, tenantID, statusPageID uint) (*Domain, error)
	GetByVerificationToken(ctx context.Context, token string) (*Domain, error)
}
