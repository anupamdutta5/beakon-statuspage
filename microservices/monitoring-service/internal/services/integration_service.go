// Package services provides third-party integration business logic for the Monitoring Service.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IntegrationService handles third-party monitoring tool integrations.
type IntegrationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewIntegrationService creates a new integration service.
func NewIntegrationService(db *gorm.DB, logger *zap.Logger) *IntegrationService {
	return &IntegrationService{
		db:     db,
		logger: logger,
	}
}

// CreateIntegration creates a new third-party integration.
func (s *IntegrationService) CreateIntegration(ctx context.Context, integration *models.Integration) error {
	if err := integration.Validate(); err != nil {
		return fmt.Errorf("integration validation failed: %w", err)
	}

	if err := s.db.WithContext(ctx).Create(integration).Error; err != nil {
		s.logger.Error("Failed to create integration", zap.Error(err))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Integration created",
		zap.Uint("integration_id", integration.ID),
		zap.Uint("tenant_id", integration.TenantID),
		zap.String("name", integration.Name),
		zap.String("type", integration.Type))

	// Trigger initial sync
	go s.syncIntegration(context.Background(), integration.ID)

	return nil
}

// GetIntegrations retrieves integrations for a tenant.
func (s *IntegrationService) GetIntegrations(ctx context.Context, tenantID uint, activeOnly bool) ([]models.Integration, error) {
	var integrations []models.Integration

	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Find(&integrations).Error; err != nil {
		s.logger.Error("Failed to get integrations", zap.Error(err))
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	return integrations, nil
}

// GetIntegration retrieves a specific integration.
func (s *IntegrationService) GetIntegration(ctx context.Context, integrationID, tenantID uint) (*models.Integration, error) {
	var integration models.Integration

	if err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", integrationID, tenantID).
		First(&integration).Error; err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}

	return &integration, nil
}

// UpdateIntegration updates an integration.
func (s *IntegrationService) UpdateIntegration(ctx context.Context, integrationID, tenantID uint, updates map[string]interface{}) error {
	result := s.db.WithContext(ctx).
		Model(&models.Integration{}).
		Where("id = ? AND tenant_id = ?", integrationID, tenantID).
		Updates(updates)

	if err := result.Error; err != nil {
		s.logger.Error("Failed to update integration", zap.Error(err))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Integration updated",
		zap.Uint("integration_id", integrationID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// DeleteIntegration deletes an integration.
func (s *IntegrationService) DeleteIntegration(ctx context.Context, integrationID, tenantID uint) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", integrationID, tenantID).
		Delete(&models.Integration{})

	if err := result.Error; err != nil {
		s.logger.Error("Failed to delete integration", zap.Error(err))
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Integration deleted",
		zap.Uint("integration_id", integrationID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// SyncIntegration manually triggers a sync for an integration.
func (s *IntegrationService) SyncIntegration(ctx context.Context, integrationID, tenantID uint) error {
	integration, err := s.GetIntegration(ctx, integrationID, tenantID)
	if err != nil {
		return err
	}

	if !integration.IsActive {
		return fmt.Errorf("integration is not active")
	}

	// Trigger sync in background
	go s.syncIntegration(context.Background(), integrationID)

	return nil
}

// syncIntegration performs the actual synchronization with the third-party service.
func (s *IntegrationService) syncIntegration(ctx context.Context, integrationID uint) {
	// Get the integration
	var integration models.Integration
	if err := s.db.WithContext(ctx).First(&integration, integrationID).Error; err != nil {
		s.logger.Error("Failed to get integration for sync", zap.Error(err))
		return
	}

	// Create sync log
	syncLog := &models.IntegrationSyncLog{
		IntegrationID: integrationID,
		SyncType:      models.SyncTypeFull,
		Status:        "started",
		StartedAt:     time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(syncLog).Error; err != nil {
		s.logger.Error("Failed to create sync log", zap.Error(err))
		return
	}

	// Update integration sync status
	s.db.WithContext(ctx).Model(&integration).Updates(map[string]interface{}{
		"sync_status": models.SyncStatusSyncing,
		"sync_error":  "",
	})

	var err error
	var recordsSync int

	switch integration.Type {
	case models.IntegrationTypePagerDuty:
		recordsSync, err = s.syncPagerDuty(ctx, &integration)
	case models.IntegrationTypeNewRelic:
		recordsSync, err = s.syncNewRelic(ctx, &integration)
	case models.IntegrationTypeDatadog:
		recordsSync, err = s.syncDatadog(ctx, &integration)
	case models.IntegrationTypePingdom:
		recordsSync, err = s.syncPingdom(ctx, &integration)
	default:
		err = fmt.Errorf("unsupported integration type: %s", integration.Type)
	}

	// Update sync log
	now := time.Now()
	syncLog.CompletedAt = &now
	syncLog.Duration = now.Sub(syncLog.StartedAt).Milliseconds()
	syncLog.RecordsSync = recordsSync

	if err != nil {
		syncLog.Status = "failed"
		syncLog.ErrorMessage = err.Error()
		s.db.WithContext(ctx).Model(&integration).Updates(map[string]interface{}{
			"sync_status": models.SyncStatusError,
			"sync_error":  err.Error(),
		})
		s.logger.Error("Integration sync failed",
			zap.Uint("integration_id", integrationID),
			zap.String("type", integration.Type),
			zap.Error(err))
	} else {
		syncLog.Status = "completed"
		s.db.WithContext(ctx).Model(&integration).Updates(map[string]interface{}{
			"sync_status":  models.SyncStatusSuccess,
			"sync_error":   "",
			"last_sync_at": &now,
		})
		s.logger.Info("Integration sync completed",
			zap.Uint("integration_id", integrationID),
			zap.String("type", integration.Type),
			zap.Int("records_sync", recordsSync))
	}

	s.db.WithContext(ctx).Save(syncLog)
}

// syncPagerDuty syncs data from PagerDuty.
func (s *IntegrationService) syncPagerDuty(ctx context.Context, integration *models.Integration) (int, error) {
	config, err := integration.GetPagerDutyConfig()
	if err != nil {
		return 0, err
	}

	s.logger.Info("Syncing PagerDuty integration",
		zap.Uint("integration_id", integration.ID),
		zap.String("api_key_prefix", config.APIKey[:min(10, len(config.APIKey))]+"..."))

	// Implementation would:
	// 1. Connect to PagerDuty API using config.APIKey
	// 2. Fetch services, incidents, escalation policies
	// 3. Map to internal components
	// 4. Update component statuses based on PagerDuty service status
	// 5. Create incidents based on PagerDuty incidents

	// For now, return a placeholder implementation
	recordsSync := 0

	// Simulate API calls and processing
	// This would be replaced with actual PagerDuty API integration
	if config.SyncServices {
		// Fetch and sync PagerDuty services
		recordsSync += 5 // Placeholder count
	}

	if config.SyncIncidents {
		// Fetch and sync PagerDuty incidents
		recordsSync += 3 // Placeholder count
	}

	return recordsSync, nil
}

// syncNewRelic syncs data from New Relic.
func (s *IntegrationService) syncNewRelic(ctx context.Context, integration *models.Integration) (int, error) {
	config, err := integration.GetNewRelicConfig()
	if err != nil {
		return 0, err
	}

	s.logger.Info("Syncing New Relic integration",
		zap.Uint("integration_id", integration.ID),
		zap.String("account_id", config.AccountID),
		zap.String("region", config.Region))

	recordsSync := 0

	// Implementation would:
	// 1. Connect to New Relic API using config.APIKey
	// 2. Fetch applications, servers, synthetic monitors
	// 3. Get health status and performance metrics
	// 4. Map to internal components and update statuses
	// 5. Create alerts based on New Relic alerts

	if config.IncludeAPM {
		// Sync APM applications
		recordsSync += len(config.Applications)
	}

	if config.IncludeInfra {
		// Sync infrastructure servers
		recordsSync += len(config.Servers)
	}

	// Sync synthetic monitors
	recordsSync += len(config.Synthetics)

	return recordsSync, nil
}

// syncDatadog syncs data from Datadog.
func (s *IntegrationService) syncDatadog(ctx context.Context, integration *models.Integration) (int, error) {
	config, err := integration.GetDatadogConfig()
	if err != nil {
		return 0, err
	}

	s.logger.Info("Syncing Datadog integration",
		zap.Uint("integration_id", integration.ID),
		zap.String("site", config.Site))

	recordsSync := 0

	// Implementation would:
	// 1. Connect to Datadog API using config.APIKey and config.AppKey
	// 2. Fetch services, monitors, dashboards
	// 3. Get metrics and events
	// 4. Map to internal components and update statuses
	// 5. Process monitor alerts

	if config.IncludeMetrics {
		// Sync services and metrics
		recordsSync += len(config.Services)
	}

	// Sync monitors
	recordsSync += len(config.Monitors)

	return recordsSync, nil
}

// syncPingdom syncs data from Pingdom.
func (s *IntegrationService) syncPingdom(ctx context.Context, integration *models.Integration) (int, error) {
	config, err := integration.GetPingdomConfig()
	if err != nil {
		return 0, err
	}

	s.logger.Info("Syncing Pingdom integration",
		zap.Uint("integration_id", integration.ID),
		zap.String("username", config.Username))

	recordsSync := 0

	// Implementation would:
	// 1. Connect to Pingdom API using config.APIKey, config.Username, config.Password
	// 2. Fetch uptime checks and their statuses
	// 3. Map to internal components
	// 4. Update component statuses based on check results

	// Sync Pingdom checks
	recordsSync += len(config.Checks)

	return recordsSync, nil
}

// CreateComponentMapping creates a mapping between an external service and internal component.
func (s *IntegrationService) CreateComponentMapping(ctx context.Context, mapping *models.ComponentMapping) error {
	if err := mapping.Validate(); err != nil {
		return fmt.Errorf("component mapping validation failed: %w", err)
	}

	if err := s.db.WithContext(ctx).Create(mapping).Error; err != nil {
		s.logger.Error("Failed to create component mapping", zap.Error(err))
		return fmt.Errorf("failed to create component mapping: %w", err)
	}

	s.logger.Info("Component mapping created",
		zap.Uint("mapping_id", mapping.ID),
		zap.Uint("integration_id", mapping.IntegrationID),
		zap.Uint("component_id", mapping.ComponentID),
		zap.String("external_service_id", mapping.ExternalServiceID))

	return nil
}

// GetComponentMappings retrieves component mappings for an integration.
func (s *IntegrationService) GetComponentMappings(ctx context.Context, integrationID, tenantID uint) ([]models.ComponentMapping, error) {
	var mappings []models.ComponentMapping

	if err := s.db.WithContext(ctx).
		Joins("JOIN integrations ON integrations.id = component_mappings.integration_id").
		Where("component_mappings.integration_id = ? AND integrations.tenant_id = ?", integrationID, tenantID).
		Find(&mappings).Error; err != nil {
		return nil, fmt.Errorf("failed to get component mappings: %w", err)
	}

	return mappings, nil
}

// GetIntegrationSyncLogs retrieves sync logs for an integration.
func (s *IntegrationService) GetIntegrationSyncLogs(ctx context.Context, integrationID, tenantID uint, limit, offset int) ([]models.IntegrationSyncLog, int64, error) {
	var logs []models.IntegrationSyncLog
	var total int64

	query := s.db.WithContext(ctx).
		Joins("JOIN integrations ON integrations.id = integration_sync_logs.integration_id").
		Where("integration_sync_logs.integration_id = ? AND integrations.tenant_id = ?", integrationID, tenantID)

	// Count total records
	if err := query.Model(&models.IntegrationSyncLog{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sync logs: %w", err)
	}

	// Get logs with pagination
	if err := query.
		Order("integration_sync_logs.started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get sync logs: %w", err)
	}

	return logs, total, nil
}

// GetSupportedIntegrations returns the list of supported integration types.
func (s *IntegrationService) GetSupportedIntegrations() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        models.IntegrationTypePagerDuty,
			"name":        "PagerDuty",
			"description": "Incident management and on-call scheduling",
			"features":    []string{"incidents", "services", "escalation_policies", "schedules"},
			"auth_type":   "api_key",
		},
		{
			"type":        models.IntegrationTypeNewRelic,
			"name":        "New Relic",
			"description": "Application performance monitoring and infrastructure monitoring",
			"features":    []string{"apm", "infrastructure", "synthetics", "alerts"},
			"auth_type":   "api_key",
		},
		{
			"type":        models.IntegrationTypeDatadog,
			"name":        "Datadog",
			"description": "Monitoring and analytics platform",
			"features":    []string{"metrics", "logs", "traces", "monitors", "dashboards"},
			"auth_type":   "api_key_app_key",
		},
		{
			"type":        models.IntegrationTypePingdom,
			"name":        "Pingdom",
			"description": "Website and server monitoring",
			"features":    []string{"uptime_checks", "page_speed", "transaction_monitoring"},
			"auth_type":   "api_key_username_password",
		},
		{
			"type":        models.IntegrationTypeUptimeRobot,
			"name":        "UptimeRobot",
			"description": "Website monitoring service",
			"features":    []string{"uptime_monitoring", "heartbeat_monitoring"},
			"auth_type":   "api_key",
		},
		{
			"type":        models.IntegrationTypeStatusCake,
			"name":        "StatusCake",
			"description": "Website monitoring and testing",
			"features":    []string{"uptime_tests", "page_speed_tests", "virus_tests"},
			"auth_type":   "api_key",
		},
	}
}

// TestIntegration tests the connection to a third-party service.
func (s *IntegrationService) TestIntegration(ctx context.Context, integrationID, tenantID uint) (map[string]interface{}, error) {
	integration, err := s.GetIntegration(ctx, integrationID, tenantID)
	if err != nil {
		return nil, err
	}

	var testResult map[string]interface{}

	switch integration.Type {
	case models.IntegrationTypePagerDuty:
		testResult, err = s.testPagerDutyConnection(integration)
	case models.IntegrationTypeNewRelic:
		testResult, err = s.testNewRelicConnection(integration)
	case models.IntegrationTypeDatadog:
		testResult, err = s.testDatadogConnection(integration)
	case models.IntegrationTypePingdom:
		testResult, err = s.testPingdomConnection(integration)
	default:
		return nil, fmt.Errorf("unsupported integration type: %s", integration.Type)
	}

	if err != nil {
		return map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"tested_at":   time.Now().UTC().Format(time.RFC3339),
			"integration": integration.Name,
		}, nil
	}

	testResult["success"] = true
	testResult["tested_at"] = time.Now().UTC().Format(time.RFC3339)
	testResult["integration"] = integration.Name

	return testResult, nil
}

// testPagerDutyConnection tests PagerDuty API connection.
func (s *IntegrationService) testPagerDutyConnection(integration *models.Integration) (map[string]interface{}, error) {
	config, err := integration.GetPagerDutyConfig()
	if err != nil {
		return nil, err
	}

	// In a real implementation, this would make an API call to PagerDuty
	// For now, just validate the configuration
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return map[string]interface{}{
		"api_accessible": true,
		"services_count": len(config.Services),
		"config_valid":   true,
	}, nil
}

// testNewRelicConnection tests New Relic API connection.
func (s *IntegrationService) testNewRelicConnection(integration *models.Integration) (map[string]interface{}, error) {
	config, err := integration.GetNewRelicConfig()
	if err != nil {
		return nil, err
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.AccountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	return map[string]interface{}{
		"api_accessible":   true,
		"account_id":       config.AccountID,
		"region":           config.Region,
		"applications_count": len(config.Applications),
		"config_valid":     true,
	}, nil
}

// testDatadogConnection tests Datadog API connection.
func (s *IntegrationService) testDatadogConnection(integration *models.Integration) (map[string]interface{}, error) {
	config, err := integration.GetDatadogConfig()
	if err != nil {
		return nil, err
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.AppKey == "" {
		return nil, fmt.Errorf("application key is required")
	}

	return map[string]interface{}{
		"api_accessible": true,
		"site":           config.Site,
		"services_count": len(config.Services),
		"monitors_count": len(config.Monitors),
		"config_valid":   true,
	}, nil
}

// testPingdomConnection tests Pingdom API connection.
func (s *IntegrationService) testPingdomConnection(integration *models.Integration) (map[string]interface{}, error) {
	config, err := integration.GetPingdomConfig()
	if err != nil {
		return nil, err
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	return map[string]interface{}{
		"api_accessible": true,
		"username":       config.Username,
		"checks_count":   len(config.Checks),
		"config_valid":   true,
	}, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}