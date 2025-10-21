// Package services provides business logic for the Monitoring Service.
package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/anupamdutta5/monitoring-service/internal/models"
)

// HeartbeatService provides heartbeat monitoring functionality.
type HeartbeatService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewHeartbeatService creates a new heartbeat service instance.
func NewHeartbeatService(db *gorm.DB, logger *zap.Logger) *HeartbeatService {
	return &HeartbeatService{
		db:     db,
		logger: logger,
	}
}

// CreateMonitor creates a new heartbeat monitor.
func (s *HeartbeatService) CreateMonitor(tenantID uuid.UUID, monitor *models.HeartbeatMonitor) error {
	monitor.TenantID = tenantID
	monitor.IsAlive = false

	if err := s.db.Create(monitor).Error; err != nil {
		s.logger.Error("Failed to create heartbeat monitor",
			zap.Error(err),
			zap.String("tenant_id", tenantID.String()),
		)
		return fmt.Errorf("failed to create heartbeat monitor: %w", err)
	}

	s.logger.Info("Heartbeat monitor created",
		zap.Uint("monitor_id", monitor.ID),
		zap.String("name", monitor.Name),
		zap.String("tenant_id", tenantID.String()),
	)

	return nil
}

// RecordPing records a heartbeat ping.
func (s *HeartbeatService) RecordPing(monitorID uint, sourceIP string) error {
	var monitor models.HeartbeatMonitor
	if err := s.db.First(&monitor, monitorID).Error; err != nil {
		return fmt.Errorf("monitor not found: %w", err)
	}

	now := time.Now()
	monitor.LastPing = &now
	monitor.IsAlive = true
	monitor.ConsecutiveMisses = 0
	monitor.AlertSent = false

	if err := s.db.Save(&monitor).Error; err != nil {
		s.logger.Error("Failed to record heartbeat ping",
			zap.Error(err),
			zap.Uint("monitor_id", monitorID),
		)
		return fmt.Errorf("failed to record ping: %w", err)
	}

	s.logger.Debug("Heartbeat ping recorded",
		zap.Uint("monitor_id", monitorID),
		zap.String("source_ip", sourceIP),
	)

	return nil
}

// CheckOverdueHeartbeats checks for overdue heartbeats and sends alerts.
func (s *HeartbeatService) CheckOverdueHeartbeats() error {
	var monitors []models.HeartbeatMonitor
	if err := s.db.Find(&monitors).Error; err != nil {
		return fmt.Errorf("failed to fetch monitors: %w", err)
	}

	for _, monitor := range monitors {
		if monitor.IsOverdue() {
			s.logger.Warn("Heartbeat monitor is overdue",
				zap.Uint("monitor_id", monitor.ID),
				zap.String("name", monitor.Name),
				zap.String("tenant_id", monitor.TenantID.String()),
			)

			// Update status
			monitor.IsAlive = false
			monitor.ConsecutiveMisses++

			if err := s.db.Save(&monitor).Error; err != nil {
				s.logger.Error("Failed to update monitor status",
					zap.Error(err),
					zap.Uint("monitor_id", monitor.ID),
				)
				continue
			}

			// TODO: Send alert via notification service or RabbitMQ event
			s.logger.Info("Heartbeat alert triggered",
				zap.Uint("monitor_id", monitor.ID),
				zap.String("name", monitor.Name),
				zap.Int("consecutive_misses", monitor.ConsecutiveMisses),
			)
		}
	}

	return nil
}

// GetMonitor retrieves a specific heartbeat monitor.
func (s *HeartbeatService) GetMonitor(monitorID uint, tenantID uuid.UUID) (*models.HeartbeatMonitor, error) {
	var monitor models.HeartbeatMonitor
	if err := s.db.Where("id = ? AND tenant_id = ?", monitorID, tenantID).First(&monitor).Error; err != nil {
		return nil, fmt.Errorf("monitor not found: %w", err)
	}
	return &monitor, nil
}

// GetMonitors retrieves all heartbeat monitors for a tenant.
func (s *HeartbeatService) GetMonitors(tenantID uuid.UUID) ([]models.HeartbeatMonitor, error) {
	var monitors []models.HeartbeatMonitor
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&monitors).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch monitors: %w", err)
	}
	return monitors, nil
}

// UpdateMonitor updates a heartbeat monitor.
func (s *HeartbeatService) UpdateMonitor(monitorID uint, tenantID uuid.UUID, updates map[string]interface{}) error {
	result := s.db.Model(&models.HeartbeatMonitor{}).
		Where("id = ? AND tenant_id = ?", monitorID, tenantID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update monitor: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("monitor not found")
	}

	s.logger.Info("Heartbeat monitor updated",
		zap.Uint("monitor_id", monitorID),
		zap.String("tenant_id", tenantID.String()),
	)

	return nil
}

// DeleteMonitor deletes a heartbeat monitor.
func (s *HeartbeatService) DeleteMonitor(monitorID uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", monitorID, tenantID).Delete(&models.HeartbeatMonitor{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete monitor: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("monitor not found")
	}

	s.logger.Info("Heartbeat monitor deleted",
		zap.Uint("monitor_id", monitorID),
		zap.String("tenant_id", tenantID.String()),
	)

	return nil
}
