// Package services provides business logic for the Incident Service.
package services

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/incident-service/internal/config"
	"github.com/anupamdutta5/incident-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// IncidentService handles incident-related business logic.
type IncidentService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewIncidentService creates a new incident service.
func NewIncidentService(db *gorm.DB, logger *zap.Logger) *IncidentService {
	// NOTE: Database migrations are managed by Atlas (see migrations/ directory and atlas.hcl)
	// Run migrations before starting the service:
	//   cd microservices/incident-service
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files

	return &IncidentService{
		db:     db,
		logger: logger,
	}
}

// CreateIncident creates a new incident.
func (s *IncidentService) CreateIncident(incident *models.Incident) error {
	// Set default values
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
	if incident.IsVisible == false && incident.IsVisible != true {
		incident.IsVisible = true
	}

	// Create incident
	if err := s.db.Create(incident).Error; err != nil {
		s.logger.Error("Failed to create incident", zap.Error(err))
		return fmt.Errorf("failed to create incident: %w", err)
	}

	// Create initial update
	update := &models.IncidentUpdate{
		IncidentID: incident.ID,
		Status:     incident.Status,
		Message:    incident.Description,
		IsVisible:  true,
		CreatedBy:  incident.CreatedBy,
	}

	if err := s.db.Create(update).Error; err != nil {
		s.logger.Error("Failed to create incident update", zap.Error(err))
		// Don't fail incident creation if update creation fails
	}

	s.logger.Info("Incident created successfully", zap.Uint("incident_id", incident.ID))
	return nil
}

// GetIncident retrieves an incident by ID.
func (s *IncidentService) GetIncident(id uint) (*models.Incident, error) {
	var incident models.Incident
	if err := s.db.Preload("Components").Preload("Updates").First(&incident, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("incident not found")
		}
		s.logger.Error("Failed to get incident", zap.Error(err))
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return &incident, nil
}

// GetIncidents retrieves a list of incidents with pagination.
func (s *IncidentService) GetIncidents(tenantID uint, limit, offset int) ([]*models.Incident, int64, error) {
	var incidents []*models.Incident
	var total int64

	// Get total count
	if err := s.db.Model(&models.Incident{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count incidents", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	// Get incidents with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("Components").Preload("Updates").Limit(limit).Offset(offset).Order("created_at DESC").Find(&incidents).Error; err != nil {
		s.logger.Error("Failed to get incidents", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get incidents: %w", err)
	}

	return incidents, total, nil
}

// GetPublicIncidents retrieves public incidents for a tenant.
func (s *IncidentService) GetPublicIncidents(tenantID uint) ([]*models.Incident, error) {
	var incidents []*models.Incident
	if err := s.db.Where("tenant_id = ? AND is_visible = ?", tenantID, true).Preload("Components").Preload("Updates", "is_visible = ?", true).Order("created_at DESC").Find(&incidents).Error; err != nil {
		s.logger.Error("Failed to get public incidents", zap.Error(err))
		return nil, fmt.Errorf("failed to get public incidents: %w", err)
	}

	return incidents, nil
}

// UpdateIncident updates an incident.
func (s *IncidentService) UpdateIncident(incident *models.Incident) error {
	if err := s.db.Save(incident).Error; err != nil {
		s.logger.Error("Failed to update incident", zap.Error(err))
		return fmt.Errorf("failed to update incident: %w", err)
	}

	s.logger.Info("Incident updated successfully", zap.Uint("incident_id", incident.ID))
	return nil
}

// UpdateIncidentStatus updates an incident's status.
func (s *IncidentService) UpdateIncidentStatus(incidentID uint, status string, updatedBy uint) error {
	// Get current incident
	var incident models.Incident
	if err := s.db.First(&incident, incidentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("incident not found")
		}
		s.logger.Error("Failed to get incident", zap.Error(err))
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Update incident status
	incident.Status = status
	incident.UpdatedBy = updatedBy

	// Set resolved_at if status is resolved
	if status == "resolved" && incident.ResolvedAt == nil {
		now := time.Now()
		incident.ResolvedAt = &now
	}

	if err := s.db.Save(&incident).Error; err != nil {
		s.logger.Error("Failed to update incident status", zap.Error(err))
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	s.logger.Info("Incident status updated successfully",
		zap.Uint("incident_id", incidentID),
		zap.String("new_status", status))
	return nil
}

// DeleteIncident soft deletes an incident.
func (s *IncidentService) DeleteIncident(id uint) error {
	if err := s.db.Delete(&models.Incident{}, id).Error; err != nil {
		s.logger.Error("Failed to delete incident", zap.Error(err))
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	s.logger.Info("Incident deleted successfully", zap.Uint("incident_id", id))
	return nil
}

// AddIncidentUpdate adds an update to an incident.
func (s *IncidentService) AddIncidentUpdate(update *models.IncidentUpdate) error {
	// Set default values
	if update.IsVisible == false && update.IsVisible != true {
		update.IsVisible = true
	}

	// Create update
	if err := s.db.Create(update).Error; err != nil {
		s.logger.Error("Failed to create incident update", zap.Error(err))
		return fmt.Errorf("failed to create incident update: %w", err)
	}

	// Update incident's updated_by field
	if err := s.db.Model(&models.Incident{}).Where("id = ?", update.IncidentID).Update("updated_by", update.CreatedBy).Error; err != nil {
		s.logger.Error("Failed to update incident updated_by", zap.Error(err))
		// Don't fail the update creation if this fails
	}

	s.logger.Info("Incident update created successfully", zap.Uint("update_id", update.ID))
	return nil
}

// UpdateIncidentUpdate updates an incident update.
func (s *IncidentService) UpdateIncidentUpdate(update *models.IncidentUpdate) error {
	if err := s.db.Save(update).Error; err != nil {
		s.logger.Error("Failed to update incident update", zap.Error(err))
		return fmt.Errorf("failed to update incident update: %w", err)
	}

	s.logger.Info("Incident update updated successfully", zap.Uint("update_id", update.ID))
	return nil
}

// DeleteIncidentUpdate soft deletes an incident update.
func (s *IncidentService) DeleteIncidentUpdate(id uint) error {
	if err := s.db.Delete(&models.IncidentUpdate{}, id).Error; err != nil {
		s.logger.Error("Failed to delete incident update", zap.Error(err))
		return fmt.Errorf("failed to delete incident update: %w", err)
	}

	s.logger.Info("Incident update deleted successfully", zap.Uint("update_id", id))
	return nil
}

// AddIncidentComponent adds a component to an incident.
func (s *IncidentService) AddIncidentComponent(incidentComponent *models.IncidentComponent) error {
	if err := s.db.Create(incidentComponent).Error; err != nil {
		s.logger.Error("Failed to add incident component", zap.Error(err))
		return fmt.Errorf("failed to add incident component: %w", err)
	}

	s.logger.Info("Incident component added successfully", zap.Uint("incident_component_id", incidentComponent.ID))
	return nil
}

// RemoveIncidentComponent removes a component from an incident.
func (s *IncidentService) RemoveIncidentComponent(incidentID, componentID uint) error {
	if err := s.db.Where("incident_id = ? AND component_id = ?", incidentID, componentID).Delete(&models.IncidentComponent{}).Error; err != nil {
		s.logger.Error("Failed to remove incident component", zap.Error(err))
		return fmt.Errorf("failed to remove incident component: %w", err)
	}

	s.logger.Info("Incident component removed successfully",
		zap.Uint("incident_id", incidentID),
		zap.Uint("component_id", componentID))
	return nil
}

// Incident Template Management

// CreateIncidentTemplate creates a new incident template.
func (s *IncidentService) CreateIncidentTemplate(template *models.IncidentTemplate) error {
	// Set default values
	if template.Impact == "" {
		template.Impact = "minor"
	}
	if template.Severity == "" {
		template.Severity = "low"
	}
	if template.IsActive == false && template.IsActive != true {
		template.IsActive = true
	}

	// Create template
	if err := s.db.Create(template).Error; err != nil {
		s.logger.Error("Failed to create incident template", zap.Error(err))
		return fmt.Errorf("failed to create incident template: %w", err)
	}

	s.logger.Info("Incident template created successfully", zap.Uint("template_id", template.ID))
	return nil
}

// GetIncidentTemplate retrieves an incident template by ID.
func (s *IncidentService) GetIncidentTemplate(id uint) (*models.IncidentTemplate, error) {
	var template models.IncidentTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("incident template not found")
		}
		s.logger.Error("Failed to get incident template", zap.Error(err))
		return nil, fmt.Errorf("failed to get incident template: %w", err)
	}

	return &template, nil
}

// GetIncidentTemplates retrieves a list of incident templates.
func (s *IncidentService) GetIncidentTemplates(tenantID uint, limit, offset int) ([]*models.IncidentTemplate, int64, error) {
	var templates []*models.IncidentTemplate
	var total int64

	// Get total count
	if err := s.db.Model(&models.IncidentTemplate{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count incident templates", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count incident templates: %w", err)
	}

	// Get templates with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&templates).Error; err != nil {
		s.logger.Error("Failed to get incident templates", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get incident templates: %w", err)
	}

	return templates, total, nil
}

// UpdateIncidentTemplate updates an incident template.
func (s *IncidentService) UpdateIncidentTemplate(template *models.IncidentTemplate) error {
	if err := s.db.Save(template).Error; err != nil {
		s.logger.Error("Failed to update incident template", zap.Error(err))
		return fmt.Errorf("failed to update incident template: %w", err)
	}

	s.logger.Info("Incident template updated successfully", zap.Uint("template_id", template.ID))
	return nil
}

// DeleteIncidentTemplate soft deletes an incident template.
func (s *IncidentService) DeleteIncidentTemplate(id uint) error {
	if err := s.db.Delete(&models.IncidentTemplate{}, id).Error; err != nil {
		s.logger.Error("Failed to delete incident template", zap.Error(err))
		return fmt.Errorf("failed to delete incident template: %w", err)
	}

	s.logger.Info("Incident template deleted successfully", zap.Uint("template_id", id))
	return nil
}

// GetIncidentMetrics calculates metrics for incidents.
func (s *IncidentService) GetIncidentMetrics(tenantID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	var metrics map[string]interface{} = make(map[string]interface{})

	// Get total incidents in date range
	var totalIncidents int64
	if err := s.db.Model(&models.Incident{}).Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).Count(&totalIncidents).Error; err != nil {
		s.logger.Error("Failed to count incidents", zap.Error(err))
		return nil, fmt.Errorf("failed to count incidents: %w", err)
	}

	// Get resolved incidents
	var resolvedIncidents int64
	if err := s.db.Model(&models.Incident{}).Where("tenant_id = ? AND status = ? AND resolved_at BETWEEN ? AND ?", tenantID, "resolved", startDate, endDate).Count(&resolvedIncidents).Error; err != nil {
		s.logger.Error("Failed to count resolved incidents", zap.Error(err))
		return nil, fmt.Errorf("failed to count resolved incidents: %w", err)
	}

	// Calculate MTTR (Mean Time To Resolution)
	var avgResolutionTime float64
	if resolvedIncidents > 0 {
		rows, err := s.db.Model(&models.Incident{}).
			Select("AVG(EXTRACT(EPOCH FROM (resolved_at - started_at))/60)").
			Where("tenant_id = ? AND status = ? AND resolved_at BETWEEN ? AND ?", tenantID, "resolved", startDate, endDate).
			Rows()
		if err != nil {
			s.logger.Error("Failed to calculate MTTR", zap.Error(err))
		} else {
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&avgResolutionTime)
			}
		}
	}

	// Get incidents by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&models.Incident{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		s.logger.Error("Failed to get status counts", zap.Error(err))
	}

	// Get incidents by impact
	var impactCounts []struct {
		Impact string
		Count  int64
	}
	if err := s.db.Model(&models.Incident{}).
		Select("impact, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Group("impact").
		Scan(&impactCounts).Error; err != nil {
		s.logger.Error("Failed to get impact counts", zap.Error(err))
	}

	metrics["total_incidents"] = totalIncidents
	metrics["resolved_incidents"] = resolvedIncidents
	metrics["mttr_minutes"] = avgResolutionTime
	metrics["status_distribution"] = statusCounts
	metrics["impact_distribution"] = impactCounts
	metrics["date_range"] = map[string]interface{}{
		"start": startDate,
		"end":   endDate,
	}

	return metrics, nil
}

// InitDatabase initializes the database connection and runs migrations.
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// NOTE: Database migrations are managed by Atlas (see migrations/ directory and atlas.hcl)
	// Run migrations before starting the service:
	//   cd microservices/incident-service
	//   atlas migrate apply --env dev
	//
	// AutoMigrate is NOT used in this project as per best practices documented in CLAUDE.md
	// All schema changes must be tracked in version-controlled migration files

	return db, nil
}
