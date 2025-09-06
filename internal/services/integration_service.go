package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IntegrationService struct {
	db         *gorm.DB
	httpClient *http.Client
}

// IntegrationConfig represents the configuration stored in Integration.Config
type IntegrationConfig struct {
	WebhookURL string `json:"webhook_url,omitempty"`
	APIKey     string `json:"api_key,omitempty"`
	AppKey     string `json:"app_key,omitempty"`
	Channel    string `json:"channel,omitempty"`
	ServiceKey string `json:"service_key,omitempty"`
}

func NewIntegrationService() *IntegrationService {
	return &IntegrationService{
		db:         database.DB,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// parseIntegrationConfig parses the JSON config from Integration.Config field
func (s *IntegrationService) parseIntegrationConfig(integration *models.Integration) (*IntegrationConfig, error) {
	var config IntegrationConfig
	if integration.Config != "" {
		if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to parse integration config: %w", err)
		}
	}
	return &config, nil
}

// Slack Integration

type SlackWebhookPayload struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color     string `json:"color"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Timestamp int64  `json:"ts"`
}

func (s *IntegrationService) SendSlackNotification(webhookURL string, incident *models.Incident) error {
	payload := SlackWebhookPayload{
		Text: fmt.Sprintf("🚨 *%s*", incident.Title),
		Attachments: []SlackAttachment{
			{
				Color:     s.getSlackColor(incident.Impact),
				Title:     "Incident Details",
				Text:      incident.Description,
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return s.sendWebhook(webhookURL, payload)
}

// PagerDuty Integration

type PagerDutyEvent struct {
	RoutingKey  string                `json:"routing_key"`
	EventAction string                `json:"event_action"`
	DedupKey    string                `json:"dedup_key"`
	Payload     PagerDutyEventPayload `json:"payload"`
}

type PagerDutyEventPayload struct {
	Summary       string                 `json:"summary"`
	Source        string                 `json:"source"`
	Severity      string                 `json:"severity"`
	Timestamp     string                 `json:"timestamp"`
	Component     string                 `json:"component"`
	Group         string                 `json:"group"`
	Class         string                 `json:"class"`
	CustomDetails map[string]interface{} `json:"custom_details"`
}

func (s *IntegrationService) SendPagerDutyAlert(routingKey string, incident *models.Incident) error {
	event := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "trigger",
		DedupKey:    fmt.Sprintf("incident-%d", incident.ID),
		Payload: PagerDutyEventPayload{
			Summary:   incident.Title,
			Source:    "statuspage",
			Severity:  s.getPagerDutySeverity(incident.Impact),
			Timestamp: time.Now().Format(time.RFC3339),
			Component: "statuspage",
			Group:     "statuspage",
			Class:     "incident",
			CustomDetails: map[string]interface{}{
				"incident_id": incident.ID,
				"tenant_id":   incident.TenantID,
				"status":      incident.Status,
			},
		},
	}

	return s.sendPagerDutyEvent(event)
}

func (s *IntegrationService) ResolvePagerDutyAlert(routingKey string, incident *models.Incident) error {
	event := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "resolve",
		DedupKey:    fmt.Sprintf("incident-%d", incident.ID),
		Payload: PagerDutyEventPayload{
			Summary:   fmt.Sprintf("Resolved: %s", incident.Title),
			Source:    "statuspage",
			Severity:  s.getPagerDutySeverity(incident.Impact),
			Timestamp: time.Now().Format(time.RFC3339),
			Component: "statuspage",
			Group:     "statuspage",
			Class:     "incident",
		},
	}

	return s.sendPagerDutyEvent(event)
}

// Opsgenie Integration

type OpsgenieAlert struct {
	Message     string                 `json:"message"`
	Alias       string                 `json:"alias"`
	Description string                 `json:"description"`
	Entity      string                 `json:"entity"`
	Source      string                 `json:"source"`
	Priority    string                 `json:"priority"`
	User        string                 `json:"user"`
	Note        string                 `json:"note"`
	Details     map[string]interface{} `json:"details"`
}

func (s *IntegrationService) SendOpsgenieAlert(apiKey string, incident *models.Incident) error {
	alert := OpsgenieAlert{
		Message:     incident.Title,
		Alias:       fmt.Sprintf("incident-%d", incident.ID),
		Description: incident.Description,
		Entity:      "statuspage",
		Source:      "statuspage",
		Priority:    s.getOpsgeniePriority(incident.Impact),
		User:        "statuspage",
		Note:        fmt.Sprintf("Incident created at %s", time.Now().Format(time.RFC3339)),
		Details: map[string]interface{}{
			"incident_id": incident.ID,
			"tenant_id":   incident.TenantID,
			"status":      incident.Status,
		},
	}

	return s.sendOpsgenieRequest("POST", "/v2/alerts", apiKey, alert)
}

func (s *IntegrationService) ResolveOpsgenieAlert(apiKey string, incident *models.Incident) error {
	note := map[string]interface{}{
		"note": fmt.Sprintf("Incident resolved at %s", time.Now().Format(time.RFC3339)),
	}

	return s.sendOpsgenieRequest("POST", fmt.Sprintf("/v2/alerts/%s/close", fmt.Sprintf("incident-%d", incident.ID)), apiKey, note)
}

// Datadog Integration

type DatadogEvent struct {
	Title          string   `json:"title"`
	Text           string   `json:"text"`
	DateHappened   int64    `json:"date_happened"`
	Priority       string   `json:"priority"`
	Tags           []string `json:"tags"`
	AlertType      string   `json:"alert_type"`
	SourceType     string   `json:"source_type_name"`
	AggregationKey string   `json:"aggregation_key"`
}

func (s *IntegrationService) SendDatadogEvent(apiKey string, appKey string, incident *models.Incident) error {
	event := DatadogEvent{
		Title:          incident.Title,
		Text:           incident.Description,
		DateHappened:   time.Now().Unix(),
		Priority:       s.getDatadogPriority(incident.Impact),
		Tags:           []string{"statuspage", fmt.Sprintf("tenant:%d", incident.TenantID)},
		AlertType:      "error",
		SourceType:     "statuspage",
		AggregationKey: fmt.Sprintf("incident-%d", incident.ID),
	}

	return s.sendDatadogRequest(apiKey, appKey, event)
}

// New Relic Integration

type NewRelicEvent struct {
	EventType  string                 `json:"eventType"`
	Timestamp  int64                  `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
}

func (s *IntegrationService) SendNewRelicEvent(apiKey string, incident *models.Incident) error {
	event := NewRelicEvent{
		EventType: "StatusPageIncident",
		Timestamp: time.Now().Unix(),
		Attributes: map[string]interface{}{
			"incidentId":  incident.ID,
			"tenantId":    incident.TenantID,
			"title":       incident.Title,
			"description": incident.Description,
			"status":      incident.Status,
			"impact":      incident.Impact,
			"startedAt":   incident.CreatedAt.Format(time.RFC3339),
		},
	}

	return s.sendNewRelicRequest(apiKey, event)
}

// Webhook Integration

type WebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func (s *IntegrationService) SendWebhookNotification(webhookURL string, eventType string, data map[string]interface{}) error {
	payload := WebhookPayload{
		Event:     eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}

	return s.sendWebhook(webhookURL, payload)
}

// Generic notification method that routes to appropriate integrations

func (s *IntegrationService) SendIncidentNotification(tenantID uint, incident *models.Incident) error {
	// Get tenant integrations
	var integrations []models.Integration
	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&integrations).Error; err != nil {
		return err
	}

	for _, integration := range integrations {
		config, err := s.parseIntegrationConfig(&integration)
		if err != nil {
			logger.Error("Failed to parse integration config", zap.Error(err))
			continue
		}

		switch integration.Type {
		case "slack":
			if config.WebhookURL != "" {
				if err := s.SendSlackNotification(config.WebhookURL, incident); err != nil {
					logger.Error("Failed to send Slack notification", zap.Error(err))
				}
			}
		case "pagerduty":
			if config.APIKey != "" {
				if err := s.SendPagerDutyAlert(config.APIKey, incident); err != nil {
					logger.Error("Failed to send PagerDuty alert", zap.Error(err))
				}
			}
		case "opsgenie":
			if config.APIKey != "" {
				if err := s.SendOpsgenieAlert(config.APIKey, incident); err != nil {
					logger.Error("Failed to send Opsgenie alert", zap.Error(err))
				}
			}
		case "datadog":
			if config.APIKey != "" && config.AppKey != "" {
				if err := s.SendDatadogEvent(config.APIKey, config.AppKey, incident); err != nil {
					logger.Error("Failed to send Datadog event", zap.Error(err))
				}
			}
		case "newrelic":
			if config.APIKey != "" {
				if err := s.SendNewRelicEvent(config.APIKey, incident); err != nil {
					logger.Error("Failed to send New Relic event", zap.Error(err))
				}
			}
		case "webhook":
			if config.WebhookURL != "" {
				data := map[string]interface{}{
					"incident": incident,
				}
				if err := s.SendWebhookNotification(config.WebhookURL, "incident.created", data); err != nil {
					logger.Error("Failed to send webhook notification", zap.Error(err))
				}
			}
		}
	}

	return nil
}

// Private helper methods

func (s *IntegrationService) sendWebhook(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *IntegrationService) sendPagerDutyEvent(event PagerDutyEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Post("https://events.pagerduty.com/v2/enqueue", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("PagerDuty returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *IntegrationService) sendOpsgenieRequest(method, path, apiKey string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.opsgenie.com%s", path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("GenieKey %s", apiKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Opsgenie returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *IntegrationService) sendDatadogRequest(apiKey, appKey string, event DatadogEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.datadoghq.com/api/v1/events?api_key=%s&application_key=%s", apiKey, appKey)
	resp, err := s.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Datadog returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *IntegrationService) sendNewRelicRequest(apiKey string, event NewRelicEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%s/events", "YOUR_ACCOUNT_ID") // This should be configurable
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Insert-Key", apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("New Relic returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *IntegrationService) getSlackColor(impact string) string {
	switch impact {
	case "critical":
		return "danger"
	case "major":
		return "warning"
	case "minor":
		return "good"
	default:
		return "good"
	}
}

func (s *IntegrationService) getPagerDutySeverity(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "info"
	default:
		return "warning"
	}
}

func (s *IntegrationService) getOpsgeniePriority(severity string) string {
	switch severity {
	case "critical":
		return "P1"
	case "high":
		return "P2"
	case "medium":
		return "P3"
	case "low":
		return "P4"
	default:
		return "P3"
	}
}

func (s *IntegrationService) getDatadogPriority(severity string) string {
	switch severity {
	case "critical":
		return "high"
	case "high":
		return "normal"
	case "medium":
		return "normal"
	case "low":
		return "low"
	default:
		return "normal"
	}
}
