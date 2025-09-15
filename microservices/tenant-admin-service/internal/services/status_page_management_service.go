// Package services provides business logic for status page management in Tenant Admin Service.
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage-tenant-admin-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StatusPageManagementService handles status page management business logic.
type StatusPageManagementService struct {
	db     *gorm.DB
	logger *zap.Logger
	config *StatusPageConfig
}

// StatusPageConfig represents the service configuration for external services.
type StatusPageConfig struct {
	ComponentServiceURL    string
	IncidentServiceURL     string
	MonitoringServiceURL   string
	NotificationServiceURL string
	BrandingServiceURL     string
}

// NewStatusPageManagementService creates a new status page management service.
func NewStatusPageManagementService(db *gorm.DB, logger *zap.Logger, config *StatusPageConfig) *StatusPageManagementService {
	return &StatusPageManagementService{
		db:     db,
		logger: logger,
		config: config,
	}
}

// GetStatusPageData retrieves all data needed for rendering the status page.
func (s *StatusPageManagementService) GetStatusPageData(tenantID uint, slug string) (*models.StatusPageData, error) {
	// Get status page configuration
	statusPage, err := s.getStatusPage(tenantID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get status page: %w", err)
	}

	config, err := s.getStatusPageConfig(statusPage.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status page config: %w", err)
	}

	// Get components status
	components, err := s.getComponentsStatus(tenantID)
	if err != nil {
		s.logger.Warn("Failed to get components status", zap.Error(err))
		components = []models.ComponentStatus{}
	}

	// Get recent incidents
	incidents, err := s.getRecentIncidents(tenantID)
	if err != nil {
		s.logger.Warn("Failed to get recent incidents", zap.Error(err))
		incidents = []models.IncidentStatus{}
	}

	// Get scheduled maintenance
	maintenance, err := s.getScheduledMaintenance(tenantID)
	if err != nil {
		s.logger.Warn("Failed to get scheduled maintenance", zap.Error(err))
		maintenance = []models.MaintenanceStatus{}
	}

	// Calculate overall status
	overallStatus := s.calculateOverallStatus(components, incidents)

	// Build status page data
	data := &models.StatusPageData{
		SiteName:             config.SiteName,
		SiteDescription:      config.SiteDescription,
		LogoURL:              config.LogoURL,
		Favicon:              config.Favicon,
		OGImage:              config.OGImage,
		SiteURL:              config.SiteURL,
		SubscribeURL:         config.SubscribeURL,
		HistoryURL:           config.HistoryURL,
		APIURL:               config.APIURL,
		WebSocketURL:         config.WebSocketURL,
		PoweredBy:            config.PoweredBy,
		PoweredByURL:         config.PoweredByURL,
		LastUpdated:          time.Now().Format(time.RFC3339),
		OverallStatus:        overallStatus.Status,
		StatusTitle:          overallStatus.Title,
		StatusDescription:    overallStatus.Description,
		Components:           components,
		RecentIncidents:      incidents,
		ScheduledMaintenance: maintenance,
	}

	return data, nil
}

// CreateStatusPage creates a new status page.
func (s *StatusPageManagementService) CreateStatusPage(tenantID uint, statusPage *models.StatusPage) error {
	if err := statusPage.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	statusPage.TenantID = tenantID
	if err := s.db.Create(statusPage).Error; err != nil {
		return fmt.Errorf("failed to create status page: %w", err)
	}

	// Create default configuration
	config := &models.StatusPageConfig{
		StatusPageID:    statusPage.ID,
		SiteName:        statusPage.Name,
		SiteDescription: statusPage.Description,
		LogoURL:         "/static/images/logo.png",
		Favicon:         "/static/images/favicon.ico",
		OGImage:         "/static/images/og-image.png",
		SiteURL:         fmt.Sprintf("https://%s", statusPage.Domain),
		SubscribeURL:    "/subscribe",
		HistoryURL:      "/history",
		APIURL:          "/api",
		WebSocketURL:    "ws://localhost:8080",
		PoweredBy:       "Status Page",
		PoweredByURL:    "https://status.example.com",
		Theme:           "default",
	}

	if err := s.db.Create(config).Error; err != nil {
		return fmt.Errorf("failed to create status page config: %w", err)
	}

	s.logger.Info("Created status page",
		zap.Uint("tenant_id", tenantID),
		zap.Uint("status_page_id", statusPage.ID),
		zap.String("slug", statusPage.Slug))

	return nil
}

// UpdateStatusPage updates an existing status page.
func (s *StatusPageManagementService) UpdateStatusPage(tenantID uint, statusPage *models.StatusPage) error {
	if err := statusPage.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if status page belongs to tenant
	var existing models.StatusPage
	if err := s.db.Where("id = ? AND tenant_id = ?", statusPage.ID, tenantID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("status page not found")
		}
		return fmt.Errorf("failed to get status page: %w", err)
	}

	statusPage.TenantID = tenantID
	if err := s.db.Save(statusPage).Error; err != nil {
		return fmt.Errorf("failed to update status page: %w", err)
	}

	s.logger.Info("Updated status page",
		zap.Uint("tenant_id", tenantID),
		zap.Uint("status_page_id", statusPage.ID))

	return nil
}

// UpdateStatusPageConfig updates status page configuration.
func (s *StatusPageManagementService) UpdateStatusPageConfig(tenantID uint, statusPageID uint, config *models.StatusPageConfig) error {
	// Check if status page belongs to tenant
	var statusPage models.StatusPage
	if err := s.db.Where("id = ? AND tenant_id = ?", statusPageID, tenantID).First(&statusPage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("status page not found")
		}
		return fmt.Errorf("failed to get status page: %w", err)
	}

	config.StatusPageID = statusPageID
	if err := s.db.Save(config).Error; err != nil {
		return fmt.Errorf("failed to update status page config: %w", err)
	}

	s.logger.Info("Updated status page config",
		zap.Uint("tenant_id", tenantID),
		zap.Uint("status_page_id", statusPageID))

	return nil
}

// GetStatusPages retrieves status pages for a tenant.
func (s *StatusPageManagementService) GetStatusPages(tenantID uint, limit, offset int) ([]*models.StatusPage, int64, error) {
	var statusPages []*models.StatusPage
	var total int64

	if err := s.db.Model(&models.StatusPage{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count status pages: %w", err)
	}

	if err := s.db.Where("tenant_id = ?", tenantID).
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&statusPages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get status pages: %w", err)
	}

	return statusPages, total, nil
}

// DeleteStatusPage soft deletes a status page.
func (s *StatusPageManagementService) DeleteStatusPage(tenantID uint, statusPageID uint) error {
	// Check if status page belongs to tenant
	var statusPage models.StatusPage
	if err := s.db.Where("id = ? AND tenant_id = ?", statusPageID, tenantID).First(&statusPage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("status page not found")
		}
		return fmt.Errorf("failed to get status page: %w", err)
	}

	if err := s.db.Delete(&statusPage).Error; err != nil {
		return fmt.Errorf("failed to delete status page: %w", err)
	}

	s.logger.Info("Deleted status page",
		zap.Uint("tenant_id", tenantID),
		zap.Uint("status_page_id", statusPageID))

	return nil
}

// getStatusPage retrieves a status page by tenant ID and slug.
func (s *StatusPageManagementService) getStatusPage(tenantID uint, slug string) (*models.StatusPage, error) {
	var statusPage models.StatusPage
	if err := s.db.Where("tenant_id = ? AND slug = ? AND is_active = ?", tenantID, slug, true).First(&statusPage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("status page not found")
		}
		return nil, fmt.Errorf("failed to get status page: %w", err)
	}
	return &statusPage, nil
}

// getStatusPageConfig retrieves status page configuration.
func (s *StatusPageManagementService) getStatusPageConfig(statusPageID uint) (*models.StatusPageConfig, error) {
	var config models.StatusPageConfig
	if err := s.db.Where("status_page_id = ?", statusPageID).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default config if not found
			return s.getDefaultConfig(statusPageID), nil
		}
		return nil, fmt.Errorf("failed to get status page config: %w", err)
	}
	return &config, nil
}

// getDefaultConfig returns default status page configuration.
func (s *StatusPageManagementService) getDefaultConfig(statusPageID uint) *models.StatusPageConfig {
	return &models.StatusPageConfig{
		StatusPageID:    statusPageID,
		SiteName:        "Status Page",
		SiteDescription: "Real-time status of our services and infrastructure",
		LogoURL:         "/static/images/logo.png",
		Favicon:         "/static/images/favicon.ico",
		OGImage:         "/static/images/og-image.png",
		SiteURL:         "https://status.example.com",
		SubscribeURL:    "/subscribe",
		HistoryURL:      "/history",
		APIURL:          "/api",
		WebSocketURL:    "ws://localhost:8080",
		PoweredBy:       "Status Page",
		PoweredByURL:    "https://status.example.com",
		Theme:           "default",
	}
}

// getComponentsStatus retrieves components status from component service.
func (s *StatusPageManagementService) getComponentsStatus(tenantID uint) ([]models.ComponentStatus, error) {
	// Make HTTP request to component service
	url := fmt.Sprintf("%s/api/v1/components?tenant_id=%d", s.config.ComponentServiceURL, tenantID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("component service returned status %d", resp.StatusCode)
	}

	var response struct {
		Components []struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
		} `json:"components"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode components response: %w", err)
	}

	// Convert to ComponentStatus
	components := make([]models.ComponentStatus, len(response.Components))
	for i, comp := range response.Components {
		components[i] = models.ComponentStatus{
			ID:          comp.ID,
			Name:        comp.Name,
			Description: comp.Description,
			Status:      comp.Status,
			StatusText:  s.getStatusText(comp.Status),
			Uptime:      s.getComponentUptime(comp.ID), // Get actual uptime from monitoring service
		}
	}

	return components, nil
}

// getRecentIncidents retrieves recent incidents from incident service.
func (s *StatusPageManagementService) getRecentIncidents(tenantID uint) ([]models.IncidentStatus, error) {
	// Make HTTP request to incident service
	url := fmt.Sprintf("%s/api/v1/incidents?tenant_id=%d&limit=5", s.config.IncidentServiceURL, tenantID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("incident service returned status %d", resp.StatusCode)
	}

	var response struct {
		Incidents []struct {
			ID          uint      `json:"id"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Status      string    `json:"status"`
			CreatedAt   time.Time `json:"created_at"`
			Updates     []struct {
				ID        uint      `json:"id"`
				Message   string    `json:"message"`
				Status    string    `json:"status"`
				CreatedAt time.Time `json:"created_at"`
			} `json:"updates"`
		} `json:"incidents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode incidents response: %w", err)
	}

	// Convert to IncidentStatus
	incidents := make([]models.IncidentStatus, len(response.Incidents))
	for i, incident := range response.Incidents {
		updates := make([]models.IncidentUpdate, len(incident.Updates))
		for j, update := range incident.Updates {
			updates[j] = models.IncidentUpdate{
				ID:        update.ID,
				Message:   update.Message,
				Status:    update.Status,
				CreatedAt: update.CreatedAt.Format(time.RFC3339),
			}
		}

		incidents[i] = models.IncidentStatus{
			ID:          incident.ID,
			Title:       incident.Title,
			Description: incident.Description,
			Status:      incident.Status,
			StatusText:  s.getStatusText(incident.Status),
			CreatedAt:   incident.CreatedAt.Format(time.RFC3339),
			Updates:     updates,
		}
	}

	return incidents, nil
}

// getScheduledMaintenance retrieves scheduled maintenance from monitoring service.
func (s *StatusPageManagementService) getScheduledMaintenance(tenantID uint) ([]models.MaintenanceStatus, error) {
	// Make HTTP request to monitoring service
	url := fmt.Sprintf("%s/api/v1/maintenance?tenant_id=%d&status=scheduled", s.config.MonitoringServiceURL, tenantID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("monitoring service returned status %d", resp.StatusCode)
	}

	var response struct {
		Maintenance []struct {
			ID            uint      `json:"id"`
			Title         string    `json:"title"`
			Description   string    `json:"description"`
			Status        string    `json:"status"`
			Impact        string    `json:"impact"`
			ScheduledTime time.Time `json:"start_time"`
		} `json:"maintenance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode maintenance response: %w", err)
	}

	// Convert to MaintenanceStatus
	maintenance := make([]models.MaintenanceStatus, len(response.Maintenance))
	for i, maint := range response.Maintenance {
		maintenance[i] = models.MaintenanceStatus{
			ID:            maint.ID,
			Title:         maint.Title,
			Description:   maint.Description,
			Status:        maint.Status,
			Impact:        maint.Impact,
			ImpactText:    s.getImpactText(maint.Impact),
			ScheduledTime: maint.ScheduledTime.Format(time.RFC3339),
		}
	}

	return maintenance, nil
}

// calculateOverallStatus calculates the overall system status based on components and incidents.
func (s *StatusPageManagementService) calculateOverallStatus(components []models.ComponentStatus, incidents []models.IncidentStatus) struct {
	Status      string
	Title       string
	Description string
} {
	// Check for active incidents
	for _, incident := range incidents {
		if incident.Status == "investigating" || incident.Status == "identified" {
			return struct {
				Status      string
				Title       string
				Description string
			}{
				Status:      "partial_outage",
				Title:       "Service Disruption",
				Description: "We are currently experiencing service disruptions. Our team is investigating.",
			}
		}
	}

	// Check component status
	hasDegraded := false
	hasOutage := false

	for _, component := range components {
		switch component.Status {
		case "degraded":
			hasDegraded = true
		case "partial_outage", "major_outage":
			hasOutage = true
		}
	}

	if hasOutage {
		return struct {
			Status      string
			Title       string
			Description string
		}{
			Status:      "partial_outage",
			Title:       "Partial Service Outage",
			Description: "Some services are experiencing issues. We are working to resolve them.",
		}
	}

	if hasDegraded {
		return struct {
			Status      string
			Title       string
			Description string
		}{
			Status:      "degraded",
			Title:       "Degraded Performance",
			Description: "Some services are experiencing degraded performance.",
		}
	}

	return struct {
		Status      string
		Title       string
		Description string
	}{
		Status:      "operational",
		Title:       "All Systems Operational",
		Description: "All systems are running smoothly.",
	}
}

// getStatusText converts status to human-readable text.
func (s *StatusPageManagementService) getStatusText(status string) string {
	statusMap := map[string]string{
		"operational":    "Operational",
		"degraded":       "Degraded Performance",
		"partial_outage": "Partial Outage",
		"major_outage":   "Major Outage",
		"maintenance":    "Under Maintenance",
		"investigating":  "Investigating",
		"identified":     "Identified",
		"monitoring":     "Monitoring",
		"resolved":       "Resolved",
	}

	if text, exists := statusMap[status]; exists {
		return text
	}
	return status
}

// getImpactText converts impact to human-readable text.
func (s *StatusPageManagementService) getImpactText(impact string) string {
	impactMap := map[string]string{
		"none":     "No Impact",
		"minor":    "Minor Impact",
		"major":    "Major Impact",
		"critical": "Critical Impact",
	}

	if text, exists := impactMap[impact]; exists {
		return text
	}
	return impact
}

// getComponentUptime retrieves actual uptime from monitoring service
func (s *StatusPageManagementService) getComponentUptime(componentID uint) float64 {
	// Make HTTP request to monitoring service to get actual uptime
	url := fmt.Sprintf("%s/api/v1/components/%d/uptime", s.config.MonitoringServiceURL, componentID)
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Warn("Failed to get component uptime",
			zap.Uint("component_id", componentID),
			zap.Error(err))
		return 99.9 // Default uptime if monitoring service is unavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("Monitoring service returned error for uptime",
			zap.Uint("component_id", componentID),
			zap.Int("status_code", resp.StatusCode))
		return 99.9 // Default uptime
	}

	var uptimeResponse struct {
		Uptime float64 `json:"uptime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&uptimeResponse); err != nil {
		s.logger.Warn("Failed to decode uptime response",
			zap.Uint("component_id", componentID),
			zap.Error(err))
		return 99.9 // Default uptime
	}

	return uptimeResponse.Uptime
}
