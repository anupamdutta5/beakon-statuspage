// Package services provides business logic for incident management in Tenant Admin Service.
package services

import (
	"github.com/google/uuid"

	"context"
	"encoding/json"
	"fmt"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IncidentService handles incident management business logic.
type IncidentService struct {
	db     *gorm.DB
	cache  resilience.Cache
	logger *zap.Logger
}

// NewIncidentService creates a new incident service.
func NewIncidentService(db *gorm.DB, redisCache resilience.Cache, logger *zap.Logger) *IncidentService {
	return &IncidentService{
		db:     db,
		cache:  redisCache,
		logger: logger,
	}
}

// GetIncidents retrieves all incidents for a tenant with pagination.
// Uses cache-aside pattern: check cache first, fallback to database, then update cache.
func (s *IncidentService) GetIncidents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Incident, int64, error) {
	// Create cache key with tenant_id, limit, and offset
	cacheKey := fmt.Sprintf("incidents:%s:%d:%d", tenantID, limit, offset)

	// Try to get from cache first (cache-aside pattern)
	type CachedData struct {
		Incidents []*models.Incident `json:"incidents"`
		Total     int64              `json:"total"`
	}

	if s.cache != nil {
		if cachedData, found := s.cache.Get(ctx, cacheKey); found {
			s.logger.Debug("Cache hit for incidents",
				zap.String("tenant_id", tenantID.String()),
				zap.Int("limit", limit),
				zap.Int("offset", offset))

			// Type assert the cached data
			if cached, ok := cachedData.(CachedData); ok {
				return cached.Incidents, cached.Total, nil
			}
			// Try JSON unmarshal if type assertion fails
			if jsonBytes, ok := cachedData.([]byte); ok {
				var cached CachedData
				if err := json.Unmarshal(jsonBytes, &cached); err == nil {
					return cached.Incidents, cached.Total, nil
				}
			}
		}
		// Cache miss - continue to database
		s.logger.Debug("Cache miss for incidents", zap.String("tenant_id", tenantID.String()))
	}

	var incidents []*models.Incident
	var total int64

	// Count total incidents for this tenant
	if err := s.db.WithContext(ctx).Model(&models.Incident{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	// Fetch incidents with pagination, ordered by most recent first
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&incidents).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get incidents: %w", err)
	}

	// Update cache in background (non-blocking)
	if s.cache != nil {
		go func() {
			bgCtx := context.Background()
			cached := CachedData{
				Incidents: incidents,
				Total:     total,
			}
			// Cache for 3 minutes (incidents change more frequently than components)
			if err := s.cache.Set(bgCtx, cacheKey, cached, 3*time.Minute); err != nil {
				s.logger.Warn("Failed to cache incidents", zap.Error(err))
			}
		}()
	}

	return incidents, total, nil
}

// GetIncidentByID retrieves a single incident by ID.
func (s *IncidentService) GetIncidentByID(ctx context.Context, tenantID uuid.UUID, incidentID uuid.UUID) (*models.Incident, error) {
	var incident models.Incident

	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", incidentID, tenantID).First(&incident).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return &incident, nil
}

// CreateIncident creates a new incident.
func (s *IncidentService) CreateIncident(ctx context.Context, tenantID uuid.UUID, incident *models.Incident) error {
	// Set tenant ID
	incident.TenantID = tenantID

	// Validate incident
	if err := incident.Validate(); err != nil {
		return err
	}

	// Set default values if not provided
	if incident.Status == "" {
		incident.Status = "investigating"
	}
	if incident.Impact == "" {
		incident.Impact = "minor"
	}
	if incident.Severity == "" {
		incident.Severity = "low"
	}
	if incident.StartedAt.IsZero() {
		incident.StartedAt = time.Now()
	}

	// Create incident
	if err := s.db.WithContext(ctx).Create(incident).Error; err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	// Invalidate cache after successful creation
	s.invalidateIncidentCache(tenantID)

	s.logger.Info("Created incident",
		zap.String("tenant_id", tenantID.String()),
		zap.String("incident_id", incident.ID.String()),
		zap.String("title", incident.Title),
		zap.String("status", incident.Status),
		zap.String("impact", incident.Impact))

	return nil
}

// UpdateIncident updates an existing incident.
func (s *IncidentService) UpdateIncident(ctx context.Context, tenantID uuid.UUID, incidentID uuid.UUID, updates map[string]interface{}) error {
	// Verify incident exists and belongs to tenant
	var incident models.Incident
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", incidentID, tenantID).First(&incident).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Validate status if being updated
	if status, ok := updates["status"].(string); ok {
		validStatuses := map[string]bool{
			"investigating": true,
			"identified":    true,
			"monitoring":    true,
			"resolved":      true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid incident status: %s", status)
		}
	}

	// Validate impact if being updated
	if impact, ok := updates["impact"].(string); ok {
		validImpacts := map[string]bool{
			"none":     true,
			"minor":    true,
			"major":    true,
			"critical": true,
		}
		if !validImpacts[impact] {
			return fmt.Errorf("invalid incident impact: %s", impact)
		}
	}

	// Validate severity if being updated
	if severity, ok := updates["severity"].(string); ok {
		validSeverities := map[string]bool{
			"low":      true,
			"medium":   true,
			"high":     true,
			"critical": true,
		}
		if !validSeverities[severity] {
			return fmt.Errorf("invalid incident severity: %s", severity)
		}
	}

	// Update incident
	if err := s.db.WithContext(ctx).Model(&incident).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	// Invalidate cache after successful update
	s.invalidateIncidentCache(tenantID)

	s.logger.Info("Updated incident",
		zap.String("tenant_id", tenantID.String()),
		zap.String("incident_id", incidentID.String()),
		zap.Any("updates", updates))

	return nil
}

// ResolveIncident marks an incident as resolved.
func (s *IncidentService) ResolveIncident(ctx context.Context, tenantID uuid.UUID, incidentID uuid.UUID, resolvedBy *uuid.UUID) error {
	// Verify incident exists and belongs to tenant
	var incident models.Incident
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", incidentID, tenantID).First(&incident).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Check if already resolved
	if incident.Status == "resolved" {
		return fmt.Errorf("incident already resolved")
	}

	// Set resolution time and status
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "resolved",
		"resolved_at": now,
	}

	if resolvedBy != nil {
		updates["updated_by"] = *resolvedBy
	}

	// Update incident
	if err := s.db.WithContext(ctx).Model(&incident).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	// Invalidate cache after successful resolution
	s.invalidateIncidentCache(tenantID)

	duration := now.Sub(incident.StartedAt)
	s.logger.Info("Resolved incident",
		zap.String("tenant_id", tenantID.String()),
		zap.String("incident_id", incidentID.String()),
		zap.Duration("duration", duration))

	return nil
}

// DeleteIncident soft-deletes an incident.
func (s *IncidentService) DeleteIncident(ctx context.Context, tenantID uuid.UUID, incidentID uuid.UUID) error {
	// Verify incident exists and belongs to tenant
	var incident models.Incident
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", incidentID, tenantID).First(&incident).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Soft delete incident
	if err := s.db.WithContext(ctx).Delete(&incident).Error; err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	// Invalidate cache after successful deletion
	s.invalidateIncidentCache(tenantID)

	s.logger.Info("Deleted incident",
		zap.String("tenant_id", tenantID.String()),
		zap.String("incident_id", incidentID.String()),
		zap.String("title", incident.Title))

	return nil
}

// GetIncidentsByStatus retrieves incidents filtered by status.
func (s *IncidentService) GetIncidentsByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*models.Incident, error) {
	var incidents []*models.Incident

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Order("started_at DESC").
		Find(&incidents).Error; err != nil {
		return nil, fmt.Errorf("failed to get incidents by status: %w", err)
	}

	return incidents, nil
}

// GetActiveIncidents retrieves all non-resolved incidents.
func (s *IncidentService) GetActiveIncidents(ctx context.Context, tenantID uuid.UUID) ([]*models.Incident, error) {
	var incidents []*models.Incident

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status != ?", tenantID, "resolved").
		Order("started_at DESC").
		Find(&incidents).Error; err != nil {
		return nil, fmt.Errorf("failed to get active incidents: %w", err)
	}

	return incidents, nil
}

// GetRecentIncidents retrieves the most recent incidents.
func (s *IncidentService) GetRecentIncidents(ctx context.Context, tenantID uuid.UUID, limit int) ([]*models.Incident, error) {
	var incidents []*models.Incident

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("started_at DESC").
		Limit(limit).
		Find(&incidents).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent incidents: %w", err)
	}

	return incidents, nil
}

// GetVisibleIncidents retrieves only visible incidents (for public status page).
func (s *IncidentService) GetVisibleIncidents(ctx context.Context, tenantID uuid.UUID) ([]*models.Incident, error) {
	var incidents []*models.Incident

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND is_visible = ?", tenantID, true).
		Order("started_at DESC").
		Find(&incidents).Error; err != nil {
		return nil, fmt.Errorf("failed to get visible incidents: %w", err)
	}

	return incidents, nil
}

// GetIncidentStats retrieves statistics about incidents for the tenant.
func (s *IncidentService) GetIncidentStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	var total, active, resolved, critical int64

	// Total incidents
	if err := s.db.WithContext(ctx).Model(&models.Incident{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total incidents: %w", err)
	}

	// Active incidents
	if err := s.db.WithContext(ctx).Model(&models.Incident{}).
		Where("tenant_id = ? AND status != ?", tenantID, "resolved").
		Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active incidents: %w", err)
	}

	// Resolved incidents
	if err := s.db.WithContext(ctx).Model(&models.Incident{}).
		Where("tenant_id = ? AND status = ?", tenantID, "resolved").
		Count(&resolved).Error; err != nil {
		return nil, fmt.Errorf("failed to count resolved incidents: %w", err)
	}

	// Critical impact incidents (active only)
	if err := s.db.WithContext(ctx).Model(&models.Incident{}).
		Where("tenant_id = ? AND status != ? AND impact = ?", tenantID, "resolved", "critical").
		Count(&critical).Error; err != nil {
		return nil, fmt.Errorf("failed to count critical incidents: %w", err)
	}

	// Average resolution time (last 30 days)
	var avgResolutionMinutes float64
	s.db.WithContext(ctx).Model(&models.Incident{}).
		Select("AVG(EXTRACT(EPOCH FROM (resolved_at - started_at))/60)").
		Where("tenant_id = ? AND status = ? AND resolved_at > ?", tenantID, "resolved", time.Now().AddDate(0, 0, -30)).
		Scan(&avgResolutionMinutes)

	stats := map[string]interface{}{
		"total":                   total,
		"active":                  active,
		"resolved":                resolved,
		"critical":                critical,
		"avg_resolution_minutes":  avgResolutionMinutes,
	}

	return stats, nil
}

// GetIncidentsByImpact retrieves incidents filtered by impact level.
func (s *IncidentService) GetIncidentsByImpact(ctx context.Context, tenantID uuid.UUID, impact string) ([]*models.Incident, error) {
	var incidents []*models.Incident

	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND impact = ?", tenantID, impact).
		Order("started_at DESC").
		Find(&incidents).Error; err != nil {
		return nil, fmt.Errorf("failed to get incidents by impact: %w", err)
	}

	return incidents, nil
}

// invalidateIncidentCache invalidates all cached incident data for a tenant.
// This is called after create, update, resolve, or delete operations to ensure cache consistency.
func (s *IncidentService) invalidateIncidentCache(tenantID uuid.UUID) {
	if s.cache == nil {
		return
	}

	// Invalidate common cache keys (all pagination combinations)
	// Incidents change frequently, so we invalidate more variations
	keysToInvalidate := []string{
		fmt.Sprintf("incidents:%s:50:0", tenantID),  // Default pagination
		fmt.Sprintf("incidents:%s:100:0", tenantID), // Large page
		fmt.Sprintf("incidents:%s:10:0", tenantID),  // Small page
		fmt.Sprintf("incidents:%s:20:0", tenantID),  // Medium page
	}

	go func() {
		ctx := context.Background()
		for _, key := range keysToInvalidate {
			if err := s.cache.Delete(ctx, key); err != nil {
				s.logger.Warn("Failed to invalidate incident cache key",
					zap.String("tenant_id", tenantID.String()),
					zap.String("key", key),
					zap.Error(err))
			}
		}
		s.logger.Debug("Invalidated incident cache",
			zap.String("tenant_id", tenantID.String()),
			zap.Int("keys", len(keysToInvalidate)))
	}()
}
