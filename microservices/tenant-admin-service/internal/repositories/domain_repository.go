package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/anupamdutta5/statuspage-tenant-admin-service/internal/models"
)

type domainRepository struct {
	db *gorm.DB
}

// NewDomainRepository creates a new domain repository
func NewDomainRepository(db *gorm.DB) models.DomainRepository {
	return &domainRepository{db: db}
}

// Create creates a new domain
func (r *domainRepository) Create(ctx context.Context, domain *models.Domain) error {
	if err := domain.Validate(); err != nil {
		return fmt.Errorf("domain validation failed: %w", err)
	}

	// Generate verification token if not set
	if domain.VerificationToken == "" {
		domain.GenerateVerificationToken()
	}

	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()

	// Check if domain already exists for any tenant
	var existingDomain models.Domain
	if err := tx.Where("domain = ?", domain.Domain).First(&existingDomain).Error; err == nil {
		tx.Rollback()
		return fmt.Errorf("domain %s is already in use", domain.Domain)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("failed to check domain existence: %w", err)
	}

	// If this is set as primary, unset any existing primary domain
	if domain.IsPrimary {
		if err := tx.Model(&models.Domain{}).
			Where("tenant_id = ? AND status_page_id = ? AND is_primary = ?", domain.TenantID, domain.StatusPageID, true).
			Update("is_primary", false).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to unset existing primary domain: %w", err)
		}
	}

	// Create the domain
	if err := tx.Create(domain).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create domain: %w", err)
	}

	// If this is the first domain for the status page, set it as primary
	var count int64
	if err := tx.Model(&models.Domain{}).
		Where("tenant_id = ? AND status_page_id = ?", domain.TenantID, domain.StatusPageID).
		Count(&count).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to count domains: %w", err)
	}

	if count == 1 {
		if err := tx.Model(domain).Update("is_primary", true).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set domain as primary: %w", err)
		}
		domain.IsPrimary = true
	}

	return tx.Commit().Error
}

// GetByID gets a domain by ID
func (r *domainRepository) GetByID(ctx context.Context, id uint) (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.WithContext(ctx).First(&domain, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("domain not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return &domain, nil
}

// GetByDomain gets a domain by domain name
func (r *domainRepository) GetByDomain(ctx context.Context, domainName string) (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.WithContext(ctx).Where("domain = ?", domainName).First(&domain).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("domain not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return &domain, nil
}

// List lists domains with the given options
func (r *domainRepository) List(ctx context.Context, opts models.DomainListOptions) ([]*models.Domain, int64, error) {
	var domains []*models.Domain
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Domain{})

	if opts.TenantID > 0 {
		query = query.Where("tenant_id = ?", opts.TenantID)
	}

	if opts.StatusPageID > 0 {
		query = query.Where("status_page_id = ?", opts.StatusPageID)
	}

	if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	}

	if opts.Search != "" {
		search := "%" + opts.Search + "%"
		query = query.Where("domain LIKE ? OR ssl_cert_common_name LIKE ?", search, search)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	// Apply pagination
	if opts.Page > 0 && opts.PageSize > 0 {
	offset := (opts.Page - 1) * opts.PageSize
	query = query.Offset(offset).Limit(opts.PageSize)
	}

	// Execute query
	if err := query.Order("is_primary DESC, created_at DESC").Find(&domains).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list domains: %w", err)
	}

	return domains, total, nil
}

// Update updates a domain
func (r *domainRepository) Update(ctx context.Context, domain *models.Domain) error {
	if domain.ID == 0 {
		return errors.New("domain ID is required")
	}

	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()

	// If this is set as primary, unset any existing primary domain
	if domain.IsPrimary {
		if err := tx.Model(&models.Domain{}).
			Where("tenant_id = ? AND status_page_id = ? AND id != ? AND is_primary = ?", 
				domain.TenantID, domain.StatusPageID, domain.ID, true).
			Update("is_primary", false).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to unset existing primary domain: %w", err)
		}
	}

	// Update the domain
	result := tx.Model(domain).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"domain":               domain.Domain,
			"status":               domain.Status,
			"verification_method":  domain.VerificationMethod,
			"verification_token":   domain.VerificationToken,
			"verified_at":          domain.VerifiedAt,
			"is_primary":           domain.IsPrimary,
			"ssl_enabled":          domain.SSLEnabled,
			"ssl_cert_issued_at":   domain.SSLCertIssuedAt,
			"ssl_cert_expires_at":  domain.SSLCertExpiresAt,
			"ssl_cert_issuer":      domain.SSLCertIssuer,
			"ssl_cert_common_name": domain.SSLCertCommonName,
			"ssl_cert_raw":         domain.SSLCertRaw,
			"metadata":             domain.Metadata,
			"updated_at":           time.Now().UTC(),
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update domain: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	return tx.Commit().Error
}

// Delete deletes a domain by ID
func (r *domainRepository) Delete(ctx context.Context, id uint) error {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()

	// Check if this is the primary domain
	var domain models.Domain
	if err := tx.First(&domain, id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Delete the domain
	if err := tx.Delete(&models.Domain{}, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	// If this was the primary domain, set another domain as primary if available
	if domain.IsPrimary {
		var nextDomain models.Domain
		if err := tx.Where("tenant_id = ? AND status_page_id = ? AND id != ?", 
			domain.TenantID, domain.StatusPageID, id).
			First(&nextDomain).Error; err == nil {
			// Found another domain, set it as primary
			if err := tx.Model(&nextDomain).Update("is_primary", true).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to set new primary domain: %w", err)
			}
		}
	}

	return tx.Commit().Error
}

// VerifyDomain verifies a domain using the provided token
func (r *domainRepository) VerifyDomain(ctx context.Context, domainName, token string) error {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()

	// Get the domain by verification token
	var domain models.Domain
	if err := tx.Where("domain = ? AND verification_token = ?", domainName, token).
		First(&domain).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("invalid verification token for domain %s", domainName)
		}
		return fmt.Errorf("failed to verify domain: %w", err)
	}

	// Mark as verified
	now := time.Now().UTC()
	domain.Status = models.DomainVerified
	domain.VerifiedAt = &now
	domain.VerificationToken = ""

	if err := tx.Save(&domain).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update domain verification status: %w", err)
	}

	return tx.Commit().Error
}

// SetPrimaryDomain sets a domain as the primary domain for a status page
func (r *domainRepository) SetPrimaryDomain(ctx context.Context, tenantID, statusPageID, domainID uint) error {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()

	// Unset any existing primary domain for this status page
	if err := tx.Model(&models.Domain{}).
		Where("tenant_id = ? AND status_page_id = ? AND is_primary = ?", 
			tenantID, statusPageID, true).
		Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unset existing primary domain: %w", err)
	}

	// Set the new primary domain
	if err := tx.Model(&models.Domain{}).
		Where("id = ? AND tenant_id = ? AND status_page_id = ?", 
			domainID, tenantID, statusPageID).
		Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set primary domain: %w", err)
	}

	return tx.Commit().Error
}

// GetPrimaryDomain gets the primary domain for a status page
func (r *domainRepository) GetPrimaryDomain(ctx context.Context, tenantID, statusPageID uint) (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status_page_id = ? AND is_primary = ?", 
			tenantID, statusPageID, true).
		First(&domain).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get primary domain: %w", err)
	}
	return &domain, nil
}

// GetByVerificationToken gets a domain by its verification token
func (r *domainRepository) GetByVerificationToken(ctx context.Context, token string) (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.WithContext(ctx).
		Where("verification_token = ?", token).
		First(&domain).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invalid verification token")
		}
		return nil, fmt.Errorf("failed to get domain by verification token: %w", err)
	}
	return &domain, nil
}
