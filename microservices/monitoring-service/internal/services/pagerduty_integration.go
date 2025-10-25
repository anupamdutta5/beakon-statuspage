package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PagerDutyIntegration represents a PagerDuty integration configuration
type PagerDutyIntegration struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	IntegrationName     string    `gorm:"size:255;not null" json:"integration_name"`
	IntegrationKey      string    `gorm:"size:255;not null" json:"integration_key"` // Events API v2 integration key
	APIKey              string    `gorm:"size:255" json:"api_key"` // REST API key (optional)
	ServiceID           string    `gorm:"size:255" json:"service_id"`
	ServiceName         string    `gorm:"size:255" json:"service_name"`
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	AutoResolve         bool      `gorm:"default:true" json:"auto_resolve"` // Auto-resolve incidents when monitor recovers
	Severity            string    `gorm:"size:50;default:'error'" json:"severity"` // critical, error, warning, info
	NotifyOnDown        bool      `gorm:"default:true" json:"notify_on_down"`
	NotifyOnDegraded    bool      `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool      `gorm:"default:false" json:"notify_on_maintenance"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TableName specifies the table name for PagerDutyIntegration
func (PagerDutyIntegration) TableName() string {
	return "pagerduty_integrations"
}

// PagerDutyMonitorMapping maps monitors to PagerDuty services
type PagerDutyMonitorMapping struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	IntegrationID uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID     uint      `gorm:"not null;index" json:"monitor_id"`
	Severity      string    `gorm:"size:50" json:"severity"` // Override default severity for this monitor
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the table name for PagerDutyMonitorMapping
func (PagerDutyMonitorMapping) TableName() string {
	return "pagerduty_monitor_mappings"
}

// PagerDutyIncident represents a PagerDuty incident
type PagerDutyIncident struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	IntegrationID  uint       `gorm:"not null;index" json:"integration_id"`
	MonitorID      uint       `gorm:"not null;index" json:"monitor_id"`
	IncidentKey    string     `gorm:"size:255;not null;uniqueIndex" json:"incident_key"` // Unique deduplication key
	PDIncidentID   string     `gorm:"size:255" json:"pd_incident_id"` // PagerDuty's incident ID
	EventType      string     `gorm:"size:50;not null" json:"event_type"` // trigger, acknowledge, resolve
	Status         string     `gorm:"size:50;not null" json:"status"` // triggered, acknowledged, resolved
	Severity       string     `gorm:"size:50" json:"severity"`
	Summary        string     `gorm:"type:text" json:"summary"`
	Source         string     `gorm:"size:255" json:"source"`
	Component      string     `gorm:"size:255" json:"component"`
	Details        string     `gorm:"type:jsonb" json:"details"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	ResponseCode   int        `json:"response_code"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName specifies the table name for PagerDutyIncident
func (PagerDutyIncident) TableName() string {
	return "pagerduty_incidents"
}

// PagerDutyEvent represents a PagerDuty Events API v2 payload
type PagerDutyEvent struct {
	RoutingKey  string                 `json:"routing_key"`
	EventAction string                 `json:"event_action"` // trigger, acknowledge, resolve
	DedupKey    string                 `json:"dedup_key,omitempty"`
	Payload     PagerDutyEventPayload  `json:"payload"`
	Links       []PagerDutyLink        `json:"links,omitempty"`
	Images      []PagerDutyImage       `json:"images,omitempty"`
}

// PagerDutyEventPayload represents the payload section of a PagerDuty event
type PagerDutyEventPayload struct {
	Summary       string                 `json:"summary"`
	Source        string                 `json:"source"`
	Severity      string                 `json:"severity"` // critical, error, warning, info
	Timestamp     string                 `json:"timestamp,omitempty"`
	Component     string                 `json:"component,omitempty"`
	Group         string                 `json:"group,omitempty"`
	Class         string                 `json:"class,omitempty"`
	CustomDetails map[string]interface{} `json:"custom_details,omitempty"`
}

// PagerDutyLink represents a link in a PagerDuty event
type PagerDutyLink struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

// PagerDutyImage represents an image in a PagerDuty event
type PagerDutyImage struct {
	Src  string `json:"src"`
	Href string `json:"href,omitempty"`
	Alt  string `json:"alt,omitempty"`
}

// PagerDutyEventResponse represents the API response from PagerDuty
type PagerDutyEventResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	DedupKey string `json:"dedup_key"`
}

// PagerDutyIntegrationService manages PagerDuty integrations
type PagerDutyIntegrationService struct {
	db         *gorm.DB
	logger     *zap.Logger
	httpClient *http.Client
	eventsAPIURL string
}

// NewPagerDutyIntegrationService creates a new PagerDuty integration service
func NewPagerDutyIntegrationService(db *gorm.DB, logger *zap.Logger) *PagerDutyIntegrationService {
	return &PagerDutyIntegrationService{
		db:     db,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		eventsAPIURL: "https://events.pagerduty.com/v2/enqueue",
	}
}

// CreateIntegration creates a new PagerDuty integration
func (s *PagerDutyIntegrationService) CreateIntegration(integration *PagerDutyIntegration) error {
	// Validate integration key by sending a test event
	err := s.TestIntegration(integration.IntegrationKey)
	if err != nil {
		return fmt.Errorf("invalid integration key: %w", err)
	}

	err = s.db.Create(integration).Error
	if err != nil {
		s.logger.Error("Failed to create PagerDuty integration",
			zap.Error(err),
			zap.String("name", integration.IntegrationName))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("PagerDuty integration created",
		zap.Uint("id", integration.ID),
		zap.String("name", integration.IntegrationName))

	return nil
}

// UpdateIntegration updates an existing PagerDuty integration
func (s *PagerDutyIntegrationService) UpdateIntegration(integration *PagerDutyIntegration) error {
	err := s.db.Save(integration).Error
	if err != nil {
		s.logger.Error("Failed to update PagerDuty integration", zap.Error(err))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("PagerDuty integration updated", zap.Uint("id", integration.ID))
	return nil
}

// GetIntegration retrieves a PagerDuty integration by ID
func (s *PagerDutyIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*PagerDutyIntegration, error) {
	var integration PagerDutyIntegration
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&integration).Error
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all integrations for a tenant
func (s *PagerDutyIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]PagerDutyIntegration, error) {
	var integrations []PagerDutyIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	return integrations, nil
}

// DeleteIntegration deletes a PagerDuty integration
func (s *PagerDutyIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&PagerDutyIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("PagerDuty integration deleted", zap.Uint("id", id))
	return nil
}

// MapMonitor maps a monitor to a PagerDuty integration
func (s *PagerDutyIntegrationService) MapMonitor(mapping *PagerDutyMonitorMapping) error {
	err := s.db.Create(mapping).Error
	if err != nil {
		s.logger.Error("Failed to create monitor mapping", zap.Error(err))
		return fmt.Errorf("failed to map monitor: %w", err)
	}

	s.logger.Info("Monitor mapped to PagerDuty",
		zap.Uint("integration_id", mapping.IntegrationID),
		zap.Uint("monitor_id", mapping.MonitorID))

	return nil
}

// UnmapMonitor removes a monitor mapping
func (s *PagerDutyIntegrationService) UnmapMonitor(id uint) error {
	result := s.db.Delete(&PagerDutyMonitorMapping{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to unmap monitor: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mapping not found")
	}

	s.logger.Info("Monitor unmapped", zap.Uint("id", id))
	return nil
}

// GetMonitorMappings retrieves all mappings for a monitor
func (s *PagerDutyIntegrationService) GetMonitorMappings(monitorID uint) ([]PagerDutyMonitorMapping, error) {
	var mappings []PagerDutyMonitorMapping
	err := s.db.Where("monitor_id = ? AND is_active = ?", monitorID, true).Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get mappings: %w", err)
	}
	return mappings, nil
}

// TriggerIncident triggers a PagerDuty incident for a monitor
func (s *PagerDutyIntegrationService) TriggerIncident(ctx context.Context, monitorID uint, tenantID uuid.UUID, monitorName string, details map[string]interface{}) error {
	// Get active integrations
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Get monitor mappings
	mappings, err := s.GetMonitorMappings(monitorID)
	if err != nil {
		return fmt.Errorf("failed to get mappings: %w", err)
	}

	// Determine which integrations to use
	var activeIntegrations []PagerDutyIntegration
	for _, integration := range integrations {
		if !integration.IsActive {
			continue
		}

		// Check if monitor is mapped to this integration
		mapped := false
		for _, mapping := range mappings {
			if mapping.IntegrationID == integration.ID && mapping.IsActive {
				mapped = true
				break
			}
		}

		// If no specific mappings, use all active integrations
		if len(mappings) == 0 || mapped {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}

	if len(activeIntegrations) == 0 {
		s.logger.Info("No active PagerDuty integrations for monitor",
			zap.Uint("monitor_id", monitorID))
		return nil
	}

	// Trigger incident for each integration
	for _, integration := range activeIntegrations {
		if !integration.NotifyOnDown {
			continue
		}

		err := s.sendEvent(ctx, &integration, monitorID, "trigger", monitorName, details)
		if err != nil {
			s.logger.Error("Failed to trigger PagerDuty incident",
				zap.Error(err),
				zap.Uint("integration_id", integration.ID))
		}
	}

	return nil
}

// AcknowledgeIncident acknowledges a PagerDuty incident
func (s *PagerDutyIntegrationService) AcknowledgeIncident(ctx context.Context, monitorID uint, tenantID uuid.UUID) error {
	// Get active integrations
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Find existing incident
	var incidents []PagerDutyIncident
	err = s.db.Where("monitor_id = ? AND status = ?", monitorID, "triggered").Find(&incidents).Error
	if err != nil {
		return fmt.Errorf("failed to find incidents: %w", err)
	}

	// Acknowledge each incident
	for _, incident := range incidents {
		for _, integration := range integrations {
			if integration.ID == incident.IntegrationID {
				err := s.sendEvent(ctx, &integration, monitorID, "acknowledge", "", nil)
				if err != nil {
					s.logger.Error("Failed to acknowledge incident", zap.Error(err))
				} else {
					// Update incident status
					now := time.Now()
					incident.Status = "acknowledged"
					incident.AcknowledgedAt = &now
					s.db.Save(&incident)
				}
				break
			}
		}
	}

	return nil
}

// ResolveIncident resolves a PagerDuty incident
func (s *PagerDutyIntegrationService) ResolveIncident(ctx context.Context, monitorID uint, tenantID uuid.UUID) error {
	// Get active integrations
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Find existing incidents
	var incidents []PagerDutyIncident
	err = s.db.Where("monitor_id = ? AND status IN ?", monitorID, []string{"triggered", "acknowledged"}).Find(&incidents).Error
	if err != nil {
		return fmt.Errorf("failed to find incidents: %w", err)
	}

	// Resolve each incident
	for _, incident := range incidents {
		for _, integration := range integrations {
			if integration.ID == incident.IntegrationID && integration.AutoResolve {
				err := s.sendEvent(ctx, &integration, monitorID, "resolve", "", nil)
				if err != nil {
					s.logger.Error("Failed to resolve incident", zap.Error(err))
				} else {
					// Update incident status
					now := time.Now()
					incident.Status = "resolved"
					incident.ResolvedAt = &now
					s.db.Save(&incident)
				}
				break
			}
		}
	}

	return nil
}

// sendEvent sends an event to PagerDuty Events API v2
func (s *PagerDutyIntegrationService) sendEvent(ctx context.Context, integration *PagerDutyIntegration, monitorID uint, eventAction string, monitorName string, details map[string]interface{}) error {
	// Create deduplication key
	dedupKey := fmt.Sprintf("monitor-%d", monitorID)

	// Determine severity
	severity := integration.Severity
	if mappings, err := s.GetMonitorMappings(monitorID); err == nil {
		for _, mapping := range mappings {
			if mapping.IntegrationID == integration.ID && mapping.Severity != "" {
				severity = mapping.Severity
				break
			}
		}
	}

	// Build event payload
	event := PagerDutyEvent{
		RoutingKey:  integration.IntegrationKey,
		EventAction: eventAction,
		DedupKey:    dedupKey,
	}

	// Only include payload for trigger events
	if eventAction == "trigger" {
		summary := fmt.Sprintf("Monitor '%s' is DOWN", monitorName)
		if details != nil {
			if errorMsg, ok := details["error"].(string); ok && errorMsg != "" {
				summary = fmt.Sprintf("Monitor '%s' is DOWN: %s", monitorName, errorMsg)
			}
		}

		event.Payload = PagerDutyEventPayload{
			Summary:       summary,
			Source:        "Beakon Status Page",
			Severity:      severity,
			Timestamp:     time.Now().Format(time.RFC3339),
			Component:     monitorName,
			CustomDetails: details,
		}

		// Add links if monitor URL is provided
		if monitorURL, ok := details["monitor_url"].(string); ok {
			event.Links = []PagerDutyLink{
				{
					Href: monitorURL,
					Text: "View Monitor Details",
				},
			}
		}
	}

	// Marshal event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.eventsAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)

	// Parse response
	var eventResp PagerDutyEventResponse
	json.Unmarshal(body, &eventResp)

	// Create or update incident record
	incident := PagerDutyIncident{
		IntegrationID: integration.ID,
		MonitorID:     monitorID,
		IncidentKey:   dedupKey,
		EventType:     eventAction,
		Status:        eventAction + "ed",
		Severity:      severity,
		Summary:       event.Payload.Summary,
		Source:        event.Payload.Source,
		Component:     event.Payload.Component,
		ResponseCode:  resp.StatusCode,
	}

	if eventAction == "trigger" {
		incident.TriggeredAt = time.Now()
	}

	if resp.StatusCode == http.StatusAccepted {
		incident.PDIncidentID = eventResp.DedupKey

		s.logger.Info("PagerDuty event sent successfully",
			zap.Uint("integration_id", integration.ID),
			zap.Uint("monitor_id", monitorID),
			zap.String("event_action", eventAction),
			zap.String("dedup_key", eventResp.DedupKey))
	} else {
		incident.ErrorMessage = string(body)
		s.logger.Error("PagerDuty event failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
	}

	// Save incident
	if eventAction == "trigger" {
		s.db.Create(&incident)
	} else {
		// Update existing incident
		s.db.Where("incident_key = ?", dedupKey).Updates(&incident)
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("event failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TestIntegration tests a PagerDuty integration key
func (s *PagerDutyIntegrationService) TestIntegration(integrationKey string) error {
	event := PagerDutyEvent{
		RoutingKey:  integrationKey,
		EventAction: "trigger",
		DedupKey:    fmt.Sprintf("test-%d", time.Now().Unix()),
		Payload: PagerDutyEventPayload{
			Summary:   "Beakon Status Page integration test",
			Source:    "Beakon Status Page",
			Severity:  "info",
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal test event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.eventsAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("test event failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Auto-resolve the test incident
	resolveEvent := PagerDutyEvent{
		RoutingKey:  integrationKey,
		EventAction: "resolve",
		DedupKey:    event.DedupKey,
	}

	resolvePayload, _ := json.Marshal(resolveEvent)
	resolveReq, _ := http.NewRequestWithContext(ctx, "POST", s.eventsAPIURL, bytes.NewBuffer(resolvePayload))
	resolveReq.Header.Set("Content-Type", "application/json")
	s.httpClient.Do(resolveReq)

	return nil
}

// GetIncidentHistory retrieves incident history for a monitor
func (s *PagerDutyIntegrationService) GetIncidentHistory(monitorID uint, limit int) ([]PagerDutyIncident, error) {
	var incidents []PagerDutyIncident
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&incidents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get incident history: %w", err)
	}

	return incidents, nil
}

// GetIncidentStats retrieves incident statistics
func (s *PagerDutyIntegrationService) GetIncidentStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var integrations []PagerDutyIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		return map[string]interface{}{
			"total_triggered":   0,
			"total_acknowledged": 0,
			"total_resolved":    0,
			"avg_resolution_time": 0.0,
		}, nil
	}

	integrationIDs := make([]uint, len(integrations))
	for i, integration := range integrations {
		integrationIDs[i] = integration.ID
	}

	// Count incidents by status
	var totalTriggered, totalAcknowledged, totalResolved int64
	s.db.Model(&PagerDutyIncident{}).
		Where("integration_id IN ? AND triggered_at BETWEEN ? AND ?", integrationIDs, startDate, endDate).
		Count(&totalTriggered)

	s.db.Model(&PagerDutyIncident{}).
		Where("integration_id IN ? AND acknowledged_at BETWEEN ? AND ?", integrationIDs, startDate, endDate).
		Count(&totalAcknowledged)

	s.db.Model(&PagerDutyIncident{}).
		Where("integration_id IN ? AND resolved_at BETWEEN ? AND ?", integrationIDs, startDate, endDate).
		Count(&totalResolved)

	// Calculate average resolution time
	var resolvedIncidents []PagerDutyIncident
	s.db.Where("integration_id IN ? AND resolved_at IS NOT NULL AND triggered_at BETWEEN ? AND ?",
		integrationIDs, startDate, endDate).
		Find(&resolvedIncidents)

	avgResolutionTime := 0.0
	if len(resolvedIncidents) > 0 {
		totalResolutionTime := 0.0
		for _, incident := range resolvedIncidents {
			if incident.ResolvedAt != nil {
				resolutionTime := incident.ResolvedAt.Sub(incident.TriggeredAt).Minutes()
				totalResolutionTime += resolutionTime
			}
		}
		avgResolutionTime = totalResolutionTime / float64(len(resolvedIncidents))
	}

	return map[string]interface{}{
		"total_triggered":      totalTriggered,
		"total_acknowledged":   totalAcknowledged,
		"total_resolved":       totalResolved,
		"avg_resolution_time":  avgResolutionTime,
	}, nil
}
