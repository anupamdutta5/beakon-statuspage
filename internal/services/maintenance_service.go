package services

import (
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// MaintenanceService provides methods for managing maintenance events.
type MaintenanceService struct{}

// NewMaintenanceService creates a new MaintenanceService.
func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{}
}

// GetAllMaintenanceEvents returns all maintenance events.
func (s *MaintenanceService) GetAllMaintenanceEvents() ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Preload("Services").Order("start_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetMaintenanceEventByID retrieves a single maintenance event by its ID.
func (s *MaintenanceService) GetMaintenanceEventByID(id uint) (*models.Maintenance, error) {
	var event models.Maintenance
	if err := database.DB.Preload("Services").First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// CreateMaintenanceEvent creates a new maintenance event and associates it with services.
func (s *MaintenanceService) CreateMaintenanceEvent(event *models.Maintenance, serviceIDs []uint) error {
	tx := database.DB.Begin()

	if err := tx.Create(event).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(serviceIDs) > 0 {
		var services []*models.Service
		if err := tx.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(event).Association("Services").Append(services); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// UpdateMaintenanceEvent updates an existing maintenance event.
func (s *MaintenanceService) UpdateMaintenanceEvent(event *models.Maintenance) error {
	return database.DB.Save(event).Error
}

// DeleteMaintenanceEvent deletes a maintenance event.
func (s *MaintenanceService) DeleteMaintenanceEvent(id uint) error {
	tx := database.DB.Begin()

	var event models.Maintenance
	if err := tx.First(&event, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&event).Association("Services").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&event).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetUpcomingMaintenance retrieves all scheduled maintenance events
func (s *MaintenanceService) GetUpcomingMaintenance() ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Preload("Services").Where("status = ?", models.MaintenanceStatusScheduled).Order("start_at asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetActiveMaintenance retrieves all maintenance events currently in progress
func (s *MaintenanceService) GetActiveMaintenance() ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Preload("Services").Where("status = ?", models.MaintenanceStatusInProgress).Order("start_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetMaintenanceByTenantID retrieves all maintenance events for a specific tenant
func (s *MaintenanceService) GetMaintenanceByTenantID(tenantID uint) ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Where("tenant_id = ?", tenantID).Preload("Services").Order("start_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetUpcomingMaintenanceByTenantID retrieves upcoming maintenance events for a specific tenant
func (s *MaintenanceService) GetUpcomingMaintenanceByTenantID(tenantID uint) ([]models.Maintenance, error) {
	var events []models.Maintenance
	now := time.Now()
	if err := database.DB.Where("tenant_id = ? AND status = ? AND start_at > ?", tenantID, models.MaintenanceStatusScheduled, now).
		Preload("Services").
		Order("start_at asc").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetActiveMaintenanceByTenantID retrieves active maintenance events for a specific tenant
func (s *MaintenanceService) GetActiveMaintenanceByTenantID(tenantID uint) ([]models.Maintenance, error) {
	var events []models.Maintenance
	now := time.Now()
	if err := database.DB.Where("tenant_id = ? AND status = ? AND start_at <= ? AND end_at >= ?",
		tenantID, models.MaintenanceStatusInProgress, now, now).
		Preload("Services").
		Order("start_at desc").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetMaintenanceByStatus retrieves maintenance events by status for a tenant
func (s *MaintenanceService) GetMaintenanceByStatus(tenantID uint, status string) ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Where("tenant_id = ? AND status = ?", tenantID, status).
		Preload("Services").
		Order("start_at desc").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// UpdateMaintenanceStatus updates the status of a maintenance event
func (s *MaintenanceService) UpdateMaintenanceStatus(id uint, status string) error {
	return database.DB.Model(&models.Maintenance{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// StartMaintenance marks a maintenance event as in progress
func (s *MaintenanceService) StartMaintenance(id uint) error {
	return s.UpdateMaintenanceStatus(id, models.MaintenanceStatusInProgress)
}

// CompleteMaintenance marks a maintenance event as completed
func (s *MaintenanceService) CompleteMaintenance(id uint) error {
	return s.UpdateMaintenanceStatus(id, models.MaintenanceStatusCompleted)
}

// CancelMaintenance marks a maintenance event as cancelled
func (s *MaintenanceService) CancelMaintenance(id uint) error {
	return s.UpdateMaintenanceStatus(id, models.MaintenanceStatusCancelled)
}

// GetMaintenanceStatistics returns maintenance statistics for a tenant
func (s *MaintenanceService) GetMaintenanceStatistics(tenantID uint, days int) (*MaintenanceStatistics, error) {
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	var stats MaintenanceStatistics

	// Total maintenance events
	if err := database.DB.Model(&models.Maintenance{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&stats.TotalMaintenance).Error; err != nil {
		return nil, err
	}

	// Completed maintenance
	if err := database.DB.Model(&models.Maintenance{}).
		Where("tenant_id = ? AND status = ? AND created_at >= ?", tenantID, models.MaintenanceStatusCompleted, since).
		Count(&stats.CompletedMaintenance).Error; err != nil {
		return nil, err
	}

	// Cancelled maintenance
	if err := database.DB.Model(&models.Maintenance{}).
		Where("tenant_id = ? AND status = ? AND created_at >= ?", tenantID, models.MaintenanceStatusCancelled, since).
		Count(&stats.CancelledMaintenance).Error; err != nil {
		return nil, err
	}

	// Average duration
	var avgDuration float64
	if err := database.DB.Model(&models.Maintenance{}).
		Where("tenant_id = ? AND status = ? AND created_at >= ?", tenantID, models.MaintenanceStatusCompleted, since).
		Select("AVG(EXTRACT(EPOCH FROM (end_at - start_at))/3600)").
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	stats.AverageDurationHours = avgDuration

	// Maintenance by status
	var statusStats []struct {
		Status string
		Count  int
	}
	if err := database.DB.Model(&models.Maintenance{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	stats.MaintenanceByStatus = make(map[string]int)
	for _, stat := range statusStats {
		stats.MaintenanceByStatus[stat.Status] = stat.Count
	}

	return &stats, nil
}

// CreateMaintenanceFromTemplate creates a maintenance event from a template
func (s *MaintenanceService) CreateMaintenanceFromTemplate(templateID uint, tenantID uint, startAt time.Time, customTitle, customDescription string) (*models.Maintenance, error) {
	var template models.MaintenanceTemplate
	if err := database.DB.Preload("Services").First(&template, templateID).Error; err != nil {
		return nil, err
	}

	endAt := startAt.Add(time.Duration(template.Duration) * time.Minute)

	event := &models.Maintenance{
		Title:       customTitle,
		Description: customDescription,
		Status:      models.MaintenanceStatusScheduled,
		StartAt:     startAt,
		EndAt:       endAt,
		TenantID:    tenantID,
	}

	// If no custom title/description provided, use template defaults
	if event.Title == "" {
		event.Title = template.Title
	}
	if event.Description == "" {
		event.Description = template.Description
	}

	if err := database.DB.Create(event).Error; err != nil {
		return nil, err
	}

	// Associate services from template
	if len(template.Services) > 0 {
		if err := database.DB.Model(event).Association("Services").Append(template.Services); err != nil {
			return nil, err
		}
	}

	return event, nil
}

// GetMaintenanceByService retrieves maintenance events affecting a specific service
func (s *MaintenanceService) GetMaintenanceByService(serviceID uint) ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Joins("JOIN maintenance_services ON maintenances.id = maintenance_services.maintenance_id").
		Where("maintenance_services.service_id = ?", serviceID).
		Preload("Services").
		Order("start_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetRecentMaintenance retrieves recent maintenance events for a tenant
func (s *MaintenanceService) GetRecentMaintenance(tenantID uint, limit int) ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Where("tenant_id = ?", tenantID).
		Preload("Services").
		Order("start_at DESC").
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// SearchMaintenance searches maintenance events by title or description
func (s *MaintenanceService) SearchMaintenance(tenantID uint, query string) ([]models.Maintenance, error) {
	var events []models.Maintenance
	searchQuery := "%" + query + "%"
	if err := database.DB.Where("tenant_id = ? AND (title ILIKE ? OR description ILIKE ?)", tenantID, searchQuery, searchQuery).
		Preload("Services").
		Order("start_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetMaintenanceTimeline retrieves a timeline of maintenance events
func (s *MaintenanceService) GetMaintenanceTimeline(tenantID uint, days int) ([]MaintenanceTimelineEvent, error) {
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	var events []models.Maintenance
	if err := database.DB.Where("tenant_id = ? AND start_at >= ?", tenantID, since).
		Preload("Services").
		Order("start_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}

	var timeline []MaintenanceTimelineEvent
	for _, event := range events {
		timeline = append(timeline, MaintenanceTimelineEvent{
			Type:        "maintenance_scheduled",
			Title:       event.Title,
			Description: event.Description,
			StartAt:     event.StartAt,
			EndAt:       event.EndAt,
			Status:      event.Status,
		})
	}

	return timeline, nil
}

// CheckMaintenanceConflicts checks for scheduling conflicts
func (s *MaintenanceService) CheckMaintenanceConflicts(tenantID uint, startAt, endAt time.Time, excludeID *uint) ([]models.Maintenance, error) {
	var events []models.Maintenance
	query := database.DB.Where("tenant_id = ? AND status = ? AND ((start_at <= ? AND end_at >= ?) OR (start_at <= ? AND end_at >= ?) OR (start_at >= ? AND end_at <= ?))",
		tenantID, models.MaintenanceStatusScheduled, startAt, startAt, endAt, endAt, startAt, endAt)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Preload("Services").Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

// GetMaintenanceByDateRange retrieves maintenance events within a date range
func (s *MaintenanceService) GetMaintenanceByDateRange(tenantID uint, startDate, endDate time.Time) ([]models.Maintenance, error) {
	var events []models.Maintenance
	if err := database.DB.Where("tenant_id = ? AND start_at >= ? AND end_at <= ?", tenantID, startDate, endDate).
		Preload("Services").
		Order("start_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// MaintenanceStatistics represents maintenance statistics
type MaintenanceStatistics struct {
	TotalMaintenance     int64          `json:"total_maintenance"`
	CompletedMaintenance int64          `json:"completed_maintenance"`
	CancelledMaintenance int64          `json:"cancelled_maintenance"`
	AverageDurationHours float64        `json:"average_duration_hours"`
	MaintenanceByStatus  map[string]int `json:"maintenance_by_status"`
}

// MaintenanceTimelineEvent represents an event in the maintenance timeline
type MaintenanceTimelineEvent struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Status      string    `json:"status"`
}
