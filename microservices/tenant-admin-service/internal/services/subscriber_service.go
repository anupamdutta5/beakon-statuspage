// Package services provides business logic for subscriber management in Tenant Admin Service.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SubscriberService handles subscriber management business logic.
type SubscriberService struct {
	db     *gorm.DB
	cache  resilience.Cache
	logger *zap.Logger
}

// NewSubscriberService creates a new subscriber service.
func NewSubscriberService(db *gorm.DB, redisCache resilience.Cache, logger *zap.Logger) *SubscriberService {
	return &SubscriberService{
		db:     db,
		cache:  redisCache,
		logger: logger,
	}
}

// GetSubscribers retrieves all subscribers for a tenant with pagination.
// Uses cache-aside pattern: check cache first, fallback to database, then update cache.
func (s *SubscriberService) GetSubscribers(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Subscriber, int64, error) {
	tenantIDStr := tenantID.String()

	// Create cache key with tenant_id, limit, and offset
	cacheKey := fmt.Sprintf("subscribers:%s:%d:%d", tenantIDStr, limit, offset)

	// Try to get from cache first (cache-aside pattern)
	type CachedData struct {
		Subscribers []*models.Subscriber `json:"subscribers"`
		Total       int64                `json:"total"`
	}

	if s.cache != nil {
		if cachedData, found := s.cache.Get(ctx, cacheKey); found {
			s.logger.Debug("Cache hit for subscribers",
				zap.String("tenant_id", tenantIDStr),
				zap.Int("limit", limit),
				zap.Int("offset", offset))

			// Type assert the cached data
			if cached, ok := cachedData.(CachedData); ok {
				return cached.Subscribers, cached.Total, nil
			}
			// Try JSON unmarshal if type assertion fails
			if jsonBytes, ok := cachedData.([]byte); ok {
				var cached CachedData
				if err := json.Unmarshal(jsonBytes, &cached); err == nil {
					return cached.Subscribers, cached.Total, nil
				}
			}
		}
		// Cache miss - continue to database
		s.logger.Debug("Cache miss for subscribers", zap.String("tenant_id", tenantIDStr))
	}

	var subscribers []*models.Subscriber
	var total int64

	// Count total subscribers for this tenant
	if err := s.db.WithContext(ctx).Model(&models.Subscriber{}).Where("tenant_id = ?", tenantIDStr).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscribers: %w", err)
	}

	// Fetch subscribers with pagination
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantIDStr).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&subscribers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get subscribers: %w", err)
	}

	// Update cache in background (non-blocking)
	if s.cache != nil {
		go func() {
			bgCtx := context.Background()
			cached := CachedData{
				Subscribers: subscribers,
				Total:       total,
			}
			// Cache for 10 minutes (subscribers change less frequently than incidents)
			if err := s.cache.Set(bgCtx, cacheKey, cached, 10*time.Minute); err != nil {
				s.logger.Warn("Failed to cache subscribers", zap.Error(err))
			}
		}()
	}

	return subscribers, total, nil
}

// GetSubscriberByID retrieves a single subscriber by ID.
func (s *SubscriberService) GetSubscriberByID(ctx context.Context, tenantID uuid.UUID, subscriberID uuid.UUID) (*models.Subscriber, error) {
	var subscriber models.Subscriber

	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", subscriberID, tenantID.String()).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}

	return &subscriber, nil
}

// GetSubscriberByEmail retrieves a subscriber by email.
func (s *SubscriberService) GetSubscriberByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.Subscriber, error) {
	var subscriber models.Subscriber

	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND email = ?", tenantID.String(), email).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subscriber by email: %w", err)
	}

	return &subscriber, nil
}

// CreateSubscriber creates a new subscriber.
func (s *SubscriberService) CreateSubscriber(ctx context.Context, tenantID uuid.UUID, subscriber *models.Subscriber) error {
	// Set tenant ID
	subscriber.TenantID = tenantID

	// Validate subscriber
	if err := subscriber.Validate(); err != nil {
		return err
	}

	// Check if subscriber already exists
	var existing models.Subscriber
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND email = ?", subscriber.TenantID, subscriber.Email).First(&existing).Error
	if err == nil {
		return fmt.Errorf("subscriber with email %s already exists", subscriber.Email)
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing subscriber: %w", err)
	}

	// Generate unsubscribe token
	token, err := generateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate unsubscribe token: %w", err)
	}
	subscriber.UnsubscribeToken = token

	// Set default status
	if subscriber.Status == "" {
		subscriber.Status = "active"
	}

	// Create subscriber
	if err := s.db.WithContext(ctx).Create(subscriber).Error; err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	// Invalidate cache after successful creation
	s.invalidateSubscriberCache(tenantID)

	s.logger.Info("Created subscriber",
		zap.String("tenant_id", tenantID.String()),
		zap.String("subscriber_id", subscriber.ID.String()),
		zap.String("email", subscriber.Email))

	return nil
}

// UpdateSubscriber updates an existing subscriber.
func (s *SubscriberService) UpdateSubscriber(ctx context.Context, tenantID uuid.UUID, subscriberID uuid.UUID, updates map[string]interface{}) error {
	// Verify subscriber exists and belongs to tenant
	var subscriber models.Subscriber
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", subscriberID, tenantID.String()).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Validate status if being updated
	if status, ok := updates["status"].(string); ok {
		validStatuses := map[string]bool{
			"active":       true,
			"unsubscribed": true,
			"bounced":      true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid subscriber status: %s", status)
		}
	}

	// Update subscriber
	if err := s.db.WithContext(ctx).Model(&subscriber).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscriber: %w", err)
	}

	// Invalidate cache after successful update
	s.invalidateSubscriberCache(tenantID)

	s.logger.Info("Updated subscriber",
		zap.String("tenant_id", tenantID.String()),
		zap.String("subscriber_id", subscriberID.String()),
		zap.Any("updates", updates))

	return nil
}

// VerifySubscriber marks a subscriber's email as verified.
func (s *SubscriberService) VerifySubscriber(ctx context.Context, tenantID uuid.UUID, subscriberID uuid.UUID) error {
	// Verify subscriber exists and belongs to tenant
	var subscriber models.Subscriber
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", subscriberID, tenantID.String()).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Check if already verified
	if subscriber.IsVerified() {
		return fmt.Errorf("subscriber already verified")
	}

	// Set verified timestamp
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&subscriber).Update("verified_at", now).Error; err != nil {
		return fmt.Errorf("failed to verify subscriber: %w", err)
	}

	s.logger.Info("Verified subscriber",
		zap.String("tenant_id", tenantID.String()),
		zap.String("subscriber_id", subscriberID.String()),
		zap.String("email", subscriber.Email))

	return nil
}

// UnsubscribeByToken unsubscribes a subscriber using their unique token.
func (s *SubscriberService) UnsubscribeByToken(ctx context.Context, token string) error {
	var subscriber models.Subscriber

	if err := s.db.WithContext(ctx).Where("unsubscribe_token = ?", token).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get subscriber by token: %w", err)
	}

	// Update status to unsubscribed
	if err := s.db.WithContext(ctx).Model(&subscriber).Update("status", "unsubscribed").Error; err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	s.logger.Info("Unsubscribed subscriber",
		zap.String("tenant_id", subscriber.TenantID.String()),
		zap.String("subscriber_id", subscriber.ID.String()),
		zap.String("email", subscriber.Email))

	return nil
}

// DeleteSubscriber soft-deletes a subscriber.
func (s *SubscriberService) DeleteSubscriber(ctx context.Context, tenantID uuid.UUID, subscriberID uuid.UUID) error {
	// Verify subscriber exists and belongs to tenant
	var subscriber models.Subscriber
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", subscriberID, tenantID.String()).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get subscriber: %w", err)
	}

	// Soft delete subscriber
	if err := s.db.WithContext(ctx).Delete(&subscriber).Error; err != nil {
		return fmt.Errorf("failed to delete subscriber: %w", err)
	}

	// Invalidate cache after successful deletion
	s.invalidateSubscriberCache(tenantID)

	s.logger.Info("Deleted subscriber",
		zap.String("tenant_id", tenantID.String()),
		zap.String("subscriber_id", subscriberID.String()),
		zap.String("email", subscriber.Email))

	return nil
}

// GetActiveSubscribers retrieves all active and verified subscribers.
func (s *SubscriberService) GetActiveSubscribers(ctx context.Context, tenantID uuid.UUID) ([]*models.Subscriber, error) {
	var subscribers []*models.Subscriber

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND verified_at IS NOT NULL", tenantID.String(), "active").
		Order("created_at DESC").
		Find(&subscribers).Error; err != nil {
		return nil, fmt.Errorf("failed to get active subscribers: %w", err)
	}

	return subscribers, nil
}

// GetSubscribersByStatus retrieves subscribers filtered by status.
func (s *SubscriberService) GetSubscribersByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*models.Subscriber, error) {
	var subscribers []*models.Subscriber

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID.String(), status).
		Order("created_at DESC").
		Find(&subscribers).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscribers by status: %w", err)
	}

	return subscribers, nil
}

// GetSubscriberStats retrieves statistics about subscribers for the tenant.
func (s *SubscriberService) GetSubscriberStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	var total, active, verified, unsubscribed, bounced int64

	tenantIDStr := tenantID.String()

	// Total subscribers
	if err := s.db.WithContext(ctx).Model(&models.Subscriber{}).Where("tenant_id = ?", tenantIDStr).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total subscribers: %w", err)
	}

	// Active subscribers
	if err := s.db.WithContext(ctx).Model(&models.Subscriber{}).Where("tenant_id = ? AND status = ?", tenantIDStr, "active").Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active subscribers: %w", err)
	}

	// Verified subscribers
	if err := s.db.WithContext(ctx).Model(&models.Subscriber{}).
		Where("tenant_id = ? AND verified_at IS NOT NULL", tenantIDStr).
		Count(&verified).Error; err != nil {
		return nil, fmt.Errorf("failed to count verified subscribers: %w", err)
	}

	// Unsubscribed
	if err := s.db.WithContext(ctx).Model(&models.Subscriber{}).
		Where("tenant_id = ? AND status = ?", tenantIDStr, "unsubscribed").
		Count(&unsubscribed).Error; err != nil {
		return nil, fmt.Errorf("failed to count unsubscribed: %w", err)
	}

	// Bounced
	if err := s.db.WithContext(ctx).Model(&models.Subscriber{}).
		Where("tenant_id = ? AND status = ?", tenantIDStr, "bounced").
		Count(&bounced).Error; err != nil {
		return nil, fmt.Errorf("failed to count bounced: %w", err)
	}

	stats := map[string]interface{}{
		"total":        total,
		"active":       active,
		"verified":     verified,
		"unsubscribed": unsubscribed,
		"bounced":      bounced,
	}

	return stats, nil
}

// generateRandomToken generates a cryptographically secure random token.
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// invalidateSubscriberCache invalidates all cached subscriber data for a tenant.
// This is called after create, update, verify, or delete operations to ensure cache consistency.
func (s *SubscriberService) invalidateSubscriberCache(tenantID uuid.UUID) {
	if s.cache == nil {
		return
	}

	// Invalidate common cache keys (all pagination combinations)
	keysToInvalidate := []string{
		fmt.Sprintf("subscribers:%s:50:0", tenantID.String()),  // Default pagination
		fmt.Sprintf("subscribers:%s:100:0", tenantID.String()), // Large page
		fmt.Sprintf("subscribers:%s:10:0", tenantID.String()),  // Small page
		fmt.Sprintf("subscribers:%s:20:0", tenantID.String()),  // Medium page
	}

	go func() {
		ctx := context.Background()
		for _, key := range keysToInvalidate {
			if err := s.cache.Delete(ctx, key); err != nil {
				s.logger.Warn("Failed to invalidate subscriber cache key",
					zap.String("tenant_id", tenantID.String()),
					zap.String("key", key),
					zap.Error(err))
			}
		}
		s.logger.Debug("Invalidated subscriber cache",
			zap.String("tenant_id", tenantID.String()),
			zap.Int("keys", len(keysToInvalidate)))
	}()
}
