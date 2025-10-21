// Package services provides business logic for component management in Tenant Admin Service.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ComponentService handles component management business logic.
type ComponentService struct {
	db     *gorm.DB
	cache  resilience.Cache
	logger *zap.Logger
}

// NewComponentService creates a new component service.
func NewComponentService(db *gorm.DB, redisCache resilience.Cache, logger *zap.Logger) *ComponentService {
	return &ComponentService{
		db:     db,
		cache:  redisCache,
		logger: logger,
	}
}

// GetComponents retrieves all components for a tenant with pagination.
// Uses cache-aside pattern: check cache first, fallback to database, then update cache.
func (s *ComponentService) GetComponents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Component, int64, error) {
	// Create cache key with tenant_id, limit, and offset
	cacheKey := fmt.Sprintf("components:%s:%d:%d", tenantID, limit, offset)

	// Try to get from cache first (cache-aside pattern)
	type CachedData struct {
		Components []*models.Component `json:"components"`
		Total      int64               `json:"total"`
	}

	if s.cache != nil {
		if cachedData, found := s.cache.Get(ctx, cacheKey); found {
			s.logger.Debug("Cache hit for components",
				zap.String("tenant_id", tenantID.String()),
				zap.Int("limit", limit),
				zap.Int("offset", offset))

			// Type assert the cached data
			if cached, ok := cachedData.(CachedData); ok {
				return cached.Components, cached.Total, nil
			}
			// Try JSON unmarshal if type assertion fails
			if jsonBytes, ok := cachedData.([]byte); ok {
				var cached CachedData
				if err := json.Unmarshal(jsonBytes, &cached); err == nil {
					return cached.Components, cached.Total, nil
				}
			}
		}
		// Cache miss - continue to database
		s.logger.Debug("Cache miss for components", zap.String("tenant_id", tenantID.String()))
	}

	var components []*models.Component
	var total int64

	// Count total components for this tenant
	if err := s.db.WithContext(ctx).Model(&models.Component{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count components: %w", err)
	}

	// Fetch components with pagination, ordered by position
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&components).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get components: %w", err)
	}

	// Update cache in background (non-blocking)
	if s.cache != nil {
		go func() {
			bgCtx := context.Background()
			cached := CachedData{
				Components: components,
				Total:      total,
			}
			// Cache for 5 minutes (components don't change frequently)
			if err := s.cache.Set(bgCtx, cacheKey, cached, 5*time.Minute); err != nil {
				s.logger.Warn("Failed to cache components", zap.Error(err))
			}
		}()
	}

	return components, total, nil
}

// GetComponentByID retrieves a single component by ID.
func (s *ComponentService) GetComponentByID(ctx context.Context, tenantID uuid.UUID, componentID uuid.UUID) (*models.Component, error) {
	var component models.Component

	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	return &component, nil
}

// CreateComponent creates a new component.
func (s *ComponentService) CreateComponent(ctx context.Context, tenantID uuid.UUID, component *models.Component) error {
	// Set tenant ID
	component.TenantID = tenantID

	// Validate component
	if err := component.Validate(); err != nil {
		return err
	}

	// Set default status if not provided
	if component.Status == "" {
		component.Status = "operational"
	}

	// Create component
	if err := s.db.WithContext(ctx).Create(component).Error; err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	// Invalidate cache after successful creation
	s.invalidateComponentCache(tenantID)

	s.logger.Info("Created component",
		zap.String("tenant_id", tenantID.String()),
		zap.String("component_id", component.ID.String()),
		zap.String("name", component.Name))

	return nil
}

// UpdateComponent updates an existing component.
func (s *ComponentService) UpdateComponent(ctx context.Context, tenantID uuid.UUID, componentID uuid.UUID, updates map[string]interface{}) error {
	// Verify component exists and belongs to tenant
	var component models.Component
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get component: %w", err)
	}

	// Validate status if being updated
	if status, ok := updates["status"].(string); ok {
		validStatuses := map[string]bool{
			"operational":          true,
			"degraded_performance": true,
			"partial_outage":       true,
			"major_outage":         true,
			"maintenance":          true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid component status: %s", status)
		}
	}

	// Update component
	if err := s.db.WithContext(ctx).Model(&component).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	// Invalidate cache after successful update
	s.invalidateComponentCache(tenantID)

	s.logger.Info("Updated component",
		zap.String("tenant_id", tenantID.String()),
		zap.String("component_id", componentID.String()),
		zap.Any("updates", updates))

	return nil
}

// UpdateComponentStatus updates the status of a component.
// This is a specialized method for status updates which are common operations.
func (s *ComponentService) UpdateComponentStatus(ctx context.Context, tenantID uuid.UUID, componentID uuid.UUID, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"operational":          true,
		"degraded_performance": true,
		"partial_outage":       true,
		"major_outage":         true,
		"maintenance":          true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid component status: %s", status)
	}

	// Verify component exists and belongs to tenant
	var component models.Component
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get component: %w", err)
	}

	// Update status
	if err := s.db.WithContext(ctx).Model(&component).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update component status: %w", err)
	}

	// Invalidate cache after successful status update
	s.invalidateComponentCache(tenantID)

	s.logger.Info("Updated component status",
		zap.String("tenant_id", tenantID.String()),
		zap.String("component_id", componentID.String()),
		zap.String("old_status", component.Status),
		zap.String("new_status", status))

	return nil
}

// DeleteComponent soft-deletes a component.
func (s *ComponentService) DeleteComponent(ctx context.Context, tenantID uuid.UUID, componentID uuid.UUID) error {
	// Verify component exists and belongs to tenant
	var component models.Component
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get component: %w", err)
	}

	// Soft delete component
	if err := s.db.WithContext(ctx).Delete(&component).Error; err != nil {
		return fmt.Errorf("failed to delete component: %w", err)
	}

	// Invalidate cache after successful deletion
	s.invalidateComponentCache(tenantID)

	s.logger.Info("Deleted component",
		zap.String("tenant_id", tenantID.String()),
		zap.String("component_id", componentID.String()),
		zap.String("name", component.Name))

	return nil
}

// GetComponentsByStatus retrieves components filtered by status.
func (s *ComponentService) GetComponentsByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*models.Component, error) {
	var components []*models.Component

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Order("name ASC").
		Find(&components).Error; err != nil {
		return nil, fmt.Errorf("failed to get components by status: %w", err)
	}

	return components, nil
}

// GetVisibleComponents retrieves only visible components (for public status page).
func (s *ComponentService) GetVisibleComponents(ctx context.Context, tenantID uuid.UUID) ([]*models.Component, error) {
	var components []*models.Component

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND visible = ?", tenantID, true).
		Order("name ASC").
		Find(&components).Error; err != nil {
		return nil, fmt.Errorf("failed to get visible components: %w", err)
	}

	return components, nil
}

// ReorderComponents updates the position of components for custom ordering.
func (s *ComponentService) ReorderComponents(ctx context.Context, tenantID uuid.UUID, componentPositions map[string]int) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update each component position
	for componentID, position := range componentPositions {
		// Verify component belongs to tenant before updating
		var component models.Component
		if err := tx.Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("component %d not found", componentID)
			}
			return fmt.Errorf("failed to get component %d: %w", componentID, err)
		}

		if err := tx.Model(&component).Update("position", position).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update component position: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Reordered components",
		zap.String("tenant_id", tenantID.String()),
		zap.Int("count", len(componentPositions)))

	return nil
}

// GetComponentStats retrieves statistics about components for the tenant.
func (s *ComponentService) GetComponentStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	var total, visible, operational, degraded, outage int64

	// Total components
	if err := s.db.WithContext(ctx).Model(&models.Component{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total components: %w", err)
	}

	// Visible components
	if err := s.db.WithContext(ctx).Model(&models.Component{}).Where("tenant_id = ? AND visible = ?", tenantID, true).Count(&visible).Error; err != nil {
		return nil, fmt.Errorf("failed to count visible components: %w", err)
	}

	// Operational components
	if err := s.db.WithContext(ctx).Model(&models.Component{}).Where("tenant_id = ? AND status = ?", tenantID, "operational").Count(&operational).Error; err != nil {
		return nil, fmt.Errorf("failed to count operational components: %w", err)
	}

	// Degraded components
	if err := s.db.WithContext(ctx).Model(&models.Component{}).Where("tenant_id = ? AND status = ?", tenantID, "degraded_performance").Count(&degraded).Error; err != nil {
		return nil, fmt.Errorf("failed to count degraded components: %w", err)
	}

	// Outage components (partial + major)
	if err := s.db.WithContext(ctx).Model(&models.Component{}).
		Where("tenant_id = ? AND status IN ?", tenantID, []string{"partial_outage", "major_outage"}).
		Count(&outage).Error; err != nil {
		return nil, fmt.Errorf("failed to count outage components: %w", err)
	}

	stats := map[string]interface{}{
		"total":       total,
		"visible":     visible,
		"operational": operational,
		"degraded":    degraded,
		"outage":      outage,
	}

	return stats, nil
}

// invalidateComponentCache invalidates all cached component data for a tenant.
// This is called after create, update, or delete operations to ensure cache consistency.
func (s *ComponentService) invalidateComponentCache(tenantID uuid.UUID) {
	if s.cache == nil {
		return
	}

	// Invalidate common cache keys (all pagination combinations)
	// In a production system, you might track which keys exist or use a cache group/tag system
	keysToInvalidate := []string{
		fmt.Sprintf("components:%s:50:0", tenantID),   // Default pagination
		fmt.Sprintf("components:%s:100:0", tenantID),  // Common limit
		fmt.Sprintf("components:%s:10:0", tenantID),   // Small page
		fmt.Sprintf("components:%s:20:0", tenantID),   // Medium page
	}

	go func() {
		ctx := context.Background()
		for _, key := range keysToInvalidate {
			if err := s.cache.Delete(ctx, key); err != nil {
				s.logger.Warn("Failed to invalidate component cache key",
					zap.String("tenant_id", tenantID.String()),
					zap.String("key", key),
					zap.Error(err))
			}
		}
		s.logger.Debug("Invalidated component cache",
			zap.String("tenant_id", tenantID.String()),
			zap.Int("keys", len(keysToInvalidate)))
	}()
}
