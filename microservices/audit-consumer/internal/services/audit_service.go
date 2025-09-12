// Package services provides business logic for the Audit Consumer.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-audit-consumer/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditService handles audit-related business logic.
type AuditService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAuditService creates a new audit service.
func NewAuditService(db *gorm.DB, logger *zap.Logger) *AuditService {
	return &AuditService{
		db:     db,
		logger: logger,
	}
}

// ProcessAuditEvent processes an audit event.
func (s *AuditService) ProcessAuditEvent(ctx context.Context, event *models.AuditEvent) error {
	s.logger.Info("Processing audit event",
		zap.String("event_id", event.ID),
		zap.String("action", event.Action),
		zap.String("resource", event.Resource))

	// Create audit log record
	auditLog := &models.AuditLog{
		TenantID:     event.TenantID,
		UserID:       event.UserID,
		SessionID:    event.SessionID,
		Action:       event.Action,
		Resource:     event.Resource,
		ResourceID:   event.ResourceID,
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		RequestID:    event.RequestID,
		Status:       event.Status,
		ErrorMessage: event.ErrorMessage,
		Timestamp:    event.Timestamp,
	}

	// Add changes if present
	if event.Changes != nil {
		// Convert changes to JSON string
		// For now, we'll store it as empty string
		auditLog.Changes = ""
	}

	// Add metadata if present
	if event.Metadata != nil {
		// Convert metadata to JSON string
		// For now, we'll store it as empty string
		auditLog.Metadata = ""
	}

	// Save to database
	if err := s.db.Create(auditLog).Error; err != nil {
		s.logger.Error("Failed to save audit log", zap.Error(err))
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	s.logger.Info("Audit log saved successfully",
		zap.Uint("audit_log_id", auditLog.ID))

	return nil
}

// GetAuditLogs retrieves audit logs with pagination.
func (s *AuditService) GetAuditLogs(tenantID uint, limit, offset int) ([]*models.AuditLog, int64, error) {
	var auditLogs []*models.AuditLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.AuditLog{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get audit logs with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Order("timestamp DESC").Find(&auditLogs).Error; err != nil {
		s.logger.Error("Failed to get audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// GetAuditLogsByAction retrieves audit logs by action.
func (s *AuditService) GetAuditLogsByAction(tenantID uint, action string, limit, offset int) ([]*models.AuditLog, int64, error) {
	var auditLogs []*models.AuditLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.AuditLog{}).Where("tenant_id = ? AND action = ?", tenantID, action).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get audit logs with pagination
	if err := s.db.Where("tenant_id = ? AND action = ?", tenantID, action).Limit(limit).Offset(offset).Order("timestamp DESC").Find(&auditLogs).Error; err != nil {
		s.logger.Error("Failed to get audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// GetAuditLogsByResource retrieves audit logs by resource.
func (s *AuditService) GetAuditLogsByResource(tenantID uint, resource string, limit, offset int) ([]*models.AuditLog, int64, error) {
	var auditLogs []*models.AuditLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.AuditLog{}).Where("tenant_id = ? AND resource = ?", tenantID, resource).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get audit logs with pagination
	if err := s.db.Where("tenant_id = ? AND resource = ?", tenantID, resource).Limit(limit).Offset(offset).Order("timestamp DESC").Find(&auditLogs).Error; err != nil {
		s.logger.Error("Failed to get audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// GetAuditLogsByUser retrieves audit logs by user.
func (s *AuditService) GetAuditLogsByUser(tenantID uint, userID uint, limit, offset int) ([]*models.AuditLog, int64, error) {
	var auditLogs []*models.AuditLog
	var total int64

	// Get total count
	if err := s.db.Model(&models.AuditLog{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get audit logs with pagination
	if err := s.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Limit(limit).Offset(offset).Order("timestamp DESC").Find(&auditLogs).Error; err != nil {
		s.logger.Error("Failed to get audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// GetAuditStats returns audit processing statistics.
func (s *AuditService) GetAuditStats() (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Get total processed count
	var totalProcessed int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ?", "completed").Count(&totalProcessed).Error; err != nil {
		s.logger.Error("Failed to count processed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count processed logs: %w", err)
	}

	// Get total failed count
	var totalFailed int64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ?", "failed").Count(&totalFailed).Error; err != nil {
		s.logger.Error("Failed to count failed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count failed logs: %w", err)
	}

	// Get average processing time
	var avgProcessingTime float64
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ? AND processing_time > 0", "completed").Select("AVG(processing_time)").Scan(&avgProcessingTime).Error; err != nil {
		s.logger.Error("Failed to calculate average processing time", zap.Error(err))
		return nil, fmt.Errorf("failed to calculate average processing time: %w", err)
	}

	// Get recent processing logs (last 24 hours)
	var recentProcessed int64
	last24Hours := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&models.ProcessingLog{}).Where("status = ? AND created_at >= ?", "completed", last24Hours).Count(&recentProcessed).Error; err != nil {
		s.logger.Error("Failed to count recent processed logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count recent processed logs: %w", err)
	}

	stats["total_processed"] = totalProcessed
	stats["total_failed"] = totalFailed
	stats["average_processing_time_ms"] = avgProcessingTime
	stats["recent_processed_24h"] = recentProcessed
	stats["success_rate"] = float64(0)
	if totalProcessed+totalFailed > 0 {
		stats["success_rate"] = float64(totalProcessed) / float64(totalProcessed+totalFailed) * 100
	}
	stats["last_updated"] = time.Now().UTC()

	return stats, nil
}

// CleanupOldLogs cleans up old audit logs.
func (s *AuditService) CleanupOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.AuditLog{})
	if result.Error != nil {
		s.logger.Error("Failed to cleanup old audit logs", zap.Error(result.Error))
		return fmt.Errorf("failed to cleanup old audit logs: %w", result.Error)
	}

	s.logger.Info("Old audit logs cleaned up",
		zap.Int("retention_days", retentionDays),
		zap.Int64("deleted_count", result.RowsAffected))

	return nil
}

