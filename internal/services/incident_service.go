package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
)

// IncidentService provides methods for incident management
type IncidentService struct{}

// NewIncidentService creates a new IncidentService
func NewIncidentService() *IncidentService {
	return &IncidentService{}
}

// CreateIncident creates a new incident
func (s *IncidentService) CreateIncident(incident *models.Incident) (*models.Incident, error) {
	if err := database.DB.Create(incident).Error; err != nil {
		return nil, err
	}
	return incident, nil
}

// UpdateIncident updates an existing incident
func (s *IncidentService) UpdateIncident(incident *models.Incident) (*models.Incident, error) {
	if err := database.DB.Save(incident).Error; err != nil {
		return nil, err
	}
	return incident, nil
}

// DeleteIncident deletes an incident by ID
func (s *IncidentService) DeleteIncident(id uint) error {
	return database.DB.Delete(&models.Incident{}, id).Error
}

// GetAllIncidents retrieves all incidents
func (s *IncidentService) GetAllIncidents() ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetIncidentByID retrieves an incident by ID
func (s *IncidentService) GetIncidentByID(id uint) (*models.Incident, error) {
	var incident models.Incident
	if err := database.DB.Preload("Services").Preload("Updates").First(&incident, id).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

// GetIncidentsByTenantID retrieves all incidents for a specific tenant
func (s *IncidentService) GetIncidentsByTenantID(tenantID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ?", tenantID).Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetActiveIncidentsByTenantID retrieves active incidents for a specific tenant
func (s *IncidentService) GetActiveIncidentsByTenantID(tenantID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ? AND status IN ?", tenantID, []string{"investigating", "identified", "monitoring"}).Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetResolvedIncidentsByTenantID retrieves resolved incidents for a specific tenant
func (s *IncidentService) GetResolvedIncidentsByTenantID(tenantID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ? AND status = ?", tenantID, "resolved").Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetIncidentsByStatus retrieves incidents by status for a tenant
func (s *IncidentService) GetIncidentsByStatus(tenantID uint, status string) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ? AND status = ?", tenantID, status).Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetIncidentsByImpact retrieves incidents by impact level for a tenant
func (s *IncidentService) GetIncidentsByImpact(tenantID uint, impact string) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ? AND impact = ?", tenantID, impact).Preload("Services").Order("created_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// UpdateIncidentStatus updates the status of an incident
func (s *IncidentService) UpdateIncidentStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// If resolving the incident, set resolved_at timestamp
	if status == models.IncidentStatusResolved {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	return database.DB.Model(&models.Incident{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// AddIncidentUpdate adds a status update to an incident
func (s *IncidentService) AddIncidentUpdate(incidentID uint, description, status string) (*models.StatusUpdate, error) {
	update := &models.StatusUpdate{
		IncidentID:  incidentID,
		Description: description,
		Status:      status,
	}

	if err := database.DB.Create(update).Error; err != nil {
		return nil, err
	}

	return update, nil
}

// GetIncidentUpdates retrieves all updates for an incident
func (s *IncidentService) GetIncidentUpdates(incidentID uint) ([]models.StatusUpdate, error) {
	var updates []models.StatusUpdate
	if err := database.DB.Where("incident_id = ?", incidentID).Order("created_at ASC").Find(&updates).Error; err != nil {
		return nil, err
	}
	return updates, nil
}

// GetIncidentStatistics returns incident statistics for a tenant
func (s *IncidentService) GetIncidentStatistics(tenantID uint, days int) (*IncidentStatistics, error) {
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	var stats IncidentStatistics

	// Total incidents
	if err := database.DB.Model(&models.Incident{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&stats.TotalIncidents).Error; err != nil {
		return nil, err
	}

	// Active incidents
	if err := database.DB.Model(&models.Incident{}).
		Where("tenant_id = ? AND status IN ? AND created_at >= ?", tenantID, []string{"investigating", "identified", "monitoring"}, since).
		Count(&stats.ActiveIncidents).Error; err != nil {
		return nil, err
	}

	// Resolved incidents
	if err := database.DB.Model(&models.Incident{}).
		Where("tenant_id = ? AND status = ? AND created_at >= ?", tenantID, "resolved", since).
		Count(&stats.ResolvedIncidents).Error; err != nil {
		return nil, err
	}

	// Average resolution time
	var avgResolutionTime float64
	if err := database.DB.Model(&models.Incident{}).
		Where("tenant_id = ? AND status = ? AND resolved_at IS NOT NULL AND created_at >= ?", tenantID, "resolved", since).
		Select("AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600)").
		Scan(&avgResolutionTime).Error; err != nil {
		return nil, err
	}
	stats.AverageResolutionTimeHours = avgResolutionTime

	// Incidents by impact
	var impactStats []struct {
		Impact string
		Count  int
	}
	if err := database.DB.Model(&models.Incident{}).
		Select("impact, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Group("impact").
		Scan(&impactStats).Error; err != nil {
		return nil, err
	}

	stats.IncidentsByImpact = make(map[string]int)
	for _, stat := range impactStats {
		stats.IncidentsByImpact[stat.Impact] = stat.Count
	}

	return &stats, nil
}

// CreateIncidentFromTemplate creates an incident from a template
func (s *IncidentService) CreateIncidentFromTemplate(templateID uint, tenantID uint, customTitle, customDescription string) (*models.Incident, error) {
	var template models.IncidentTemplate
	if err := database.DB.Preload("Services").First(&template, templateID).Error; err != nil {
		return nil, err
	}

	incident := &models.Incident{
		Title:       customTitle,
		Description: customDescription,
		Status:      models.IncidentStatusInvestigating,
		Impact:      template.Impact,
		TenantID:    tenantID,
		TemplateID:  &templateID,
	}

	// If no custom title/description provided, use template defaults
	if incident.Title == "" {
		incident.Title = template.Title
	}
	if incident.Description == "" {
		incident.Description = template.Description
	}

	if err := database.DB.Create(incident).Error; err != nil {
		return nil, err
	}

	// Associate services from template
	if len(template.Services) > 0 {
		if err := database.DB.Model(incident).Association("Services").Append(template.Services); err != nil {
			return nil, err
		}
	}

	return incident, nil
}

// GetIncidentsByService retrieves incidents affecting a specific service
func (s *IncidentService) GetIncidentsByService(serviceID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Joins("JOIN incident_services ON incidents.id = incident_services.incident_id").
		Where("incident_services.service_id = ?", serviceID).
		Preload("Services").
		Order("created_at DESC").
		Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetRecentIncidents retrieves recent incidents for a tenant
func (s *IncidentService) GetRecentIncidents(tenantID uint, limit int) ([]models.Incident, error) {
	var incidents []models.Incident
	if err := database.DB.Where("tenant_id = ?", tenantID).
		Preload("Services").
		Order("created_at DESC").
		Limit(limit).
		Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// SearchIncidents searches incidents by title or description
func (s *IncidentService) SearchIncidents(tenantID uint, query string) ([]models.Incident, error) {
	var incidents []models.Incident
	searchQuery := "%" + query + "%"
	if err := database.DB.Where("tenant_id = ? AND (title ILIKE ? OR description ILIKE ?)", tenantID, searchQuery, searchQuery).
		Preload("Services").
		Order("created_at DESC").
		Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

// GetIncidentTimeline retrieves a timeline of incident events
func (s *IncidentService) GetIncidentTimeline(incidentID uint) ([]IncidentTimelineEvent, error) {
	var timeline []IncidentTimelineEvent

	// Get incident creation
	var incident models.Incident
	if err := database.DB.First(&incident, incidentID).Error; err != nil {
		return nil, err
	}

	timeline = append(timeline, IncidentTimelineEvent{
		Type:        "incident_created",
		Description: fmt.Sprintf("Incident created: %s", incident.Title),
		Timestamp:   incident.CreatedAt,
		Status:      incident.Status,
	})

	// Get status updates
	updates, err := s.GetIncidentUpdates(incidentID)
	if err != nil {
		return nil, err
	}

	for _, update := range updates {
		timeline = append(timeline, IncidentTimelineEvent{
			Type:        "status_update",
			Description: update.Description,
			Timestamp:   update.CreatedAt,
			Status:      update.Status,
		})
	}

	// Get resolution
	if incident.ResolvedAt != nil {
		timeline = append(timeline, IncidentTimelineEvent{
			Type:        "incident_resolved",
			Description: "Incident resolved",
			Timestamp:   *incident.ResolvedAt,
			Status:      "resolved",
		})
	}

	// Sort by timestamp
	for i := 0; i < len(timeline)-1; i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[i].Timestamp.After(timeline[j].Timestamp) {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}

	return timeline, nil
}

// IncidentStatistics represents incident statistics
type IncidentStatistics struct {
	TotalIncidents             int64          `json:"total_incidents"`
	ActiveIncidents            int64          `json:"active_incidents"`
	ResolvedIncidents          int64          `json:"resolved_incidents"`
	AverageResolutionTimeHours float64        `json:"average_resolution_time_hours"`
	IncidentsByImpact          map[string]int `json:"incidents_by_impact"`
}

// IncidentTimelineEvent represents an event in the incident timeline
type IncidentTimelineEvent struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}
