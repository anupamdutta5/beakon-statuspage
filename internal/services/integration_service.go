package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IntegrationService struct {
	db *gorm.DB
}

func NewIntegrationService() *IntegrationService {
	return &IntegrationService{
		db: database.GetDB(),
	}
}

// Integration types
const (
	IntegrationTypeSlack      = "slack"
	IntegrationTypePagerDuty   = "pagerduty"
	IntegrationTypeOpsgenie    = "opsgenie"
	IntegrationTypePrometheus  = "prometheus"
	IntegrationTypeDatadog     = "datadog"
	IntegrationTypeNewRelic   = "newrelic"
)

// CreateIntegration creates a new integration
func (s *IntegrationService) CreateIntegration(integration *models.Integration) error {
	if err := s.db.Create(integration).Error; err != nil {
		logger.Error("Failed to create integration", zap.Error(err))
		return err
	}

	logger.Info("Integration created successfully", 
		zap.String("name", integration.Name),
		zap.String("type", integration.Type))
	return nil
}

// GetAllIntegrations retrieves all integrations
func (s *IntegrationService) GetAllIntegrations() ([]models.Integration, error) {
	var integrations []models.Integration
	if err := s.db.Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

// GetIntegrationByID retrieves an integration by ID
func (s *IntegrationService) GetIntegrationByID(id uint) (*models.Integration, error) {
	var integration models.Integration
	if err := s.db.First(&integration, id).Error; err != nil {
		return nil, err
	}
	return &integration, nil
}

// UpdateIntegration updates an integration
func (s *IntegrationService) UpdateIntegration(integration *models.Integration) error {
	if err := s.db.Save(integration).Error; err != nil {
		logger.Error("Failed to update integration", zap.Error(err))
		return err
	}

	logger.Info("Integration updated successfully", zap.String("name", integration.Name))
	return nil
}

// DeleteIntegration deletes an integration
func (s *IntegrationService) DeleteIntegration(id uint) error {
	if err := s.db.Delete(&models.Integration{}, id).Error; err != nil {
		logger.Error("Failed to delete integration", zap.Error(err))
		return err
	}

	logger.Info("Integration deleted successfully", zap.Uint("id", id))
	return nil
}

// TestIntegration tests an integration configuration
func (s *IntegrationService) TestIntegration(integration *models.Integration) error {
	switch integration.Type {
	case IntegrationTypeSlack:
		return s.testSlackIntegration(integration)
	case IntegrationTypePagerDuty:
		return s.testPagerDutyIntegration(integration)
	case IntegrationTypeOpsgenie:
		return s.testOpsgenieIntegration(integration)
	case IntegrationTypePrometheus:
		return s.testPrometheusIntegration(integration)
	default:
		return fmt.Errorf("unsupported integration type: %s", integration.Type)
	}
}

// Slack Integration Methods

type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
}

func (s *IntegrationService) testSlackIntegration(integration *models.Integration) error {
	var config SlackConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Slack configuration: %w", err)
	}

	// Test webhook URL
	payload := map[string]interface{}{
		"text": "Status page integration test",
		"channel": config.Channel,
		"username": config.Username,
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(config.WebhookURL, "application/json", 
		strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to test Slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// PagerDuty Integration Methods

type PagerDutyConfig struct {
	APIKey     string `json:"api_key"`
	ServiceKey string `json:"service_key"`
}

func (s *IntegrationService) testPagerDutyIntegration(integration *models.Integration) error {
	var config PagerDutyConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid PagerDuty configuration: %w", err)
	}

	// Test API key by making a simple request
	req, err := http.NewRequest("GET", "https://api.pagerduty.com/services", nil)
	if err != nil {
		return fmt.Errorf("failed to create PagerDuty request: %w", err)
	}

	req.Header.Set("Authorization", "Token token="+config.APIKey)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to test PagerDuty API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PagerDuty API returned status %d", resp.StatusCode)
	}

	return nil
}

// Opsgenie Integration Methods

type OpsgenieConfig struct {
	APIKey string `json:"api_key"`
	TeamID string `json:"team_id"`
}

func (s *IntegrationService) testOpsgenieIntegration(integration *models.Integration) error {
	var config OpsgenieConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Opsgenie configuration: %w", err)
	}

	// Test API key by making a simple request
	req, err := http.NewRequest("GET", "https://api.opsgenie.com/v2/teams", nil)
	if err != nil {
		return fmt.Errorf("failed to create Opsgenie request: %w", err)
	}

	req.Header.Set("Authorization", "GenieKey "+config.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to test Opsgenie API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Opsgenie API returned status %d", resp.StatusCode)
	}

	return nil
}

// Prometheus Integration Methods

func (s *IntegrationService) testPrometheusIntegration(integration *models.Integration) error {
	var config struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Prometheus configuration: %w", err)
	}

	// Test Prometheus endpoint
	req, err := http.NewRequest("GET", config.URL+"/api/v1/query?query=up", nil)
	if err != nil {
		return fmt.Errorf("failed to create Prometheus request: %w", err)
	}

	if config.Username != "" && config.Password != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to test Prometheus API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Prometheus API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendSlackNotification sends a notification to Slack
func (s *IntegrationService) SendSlackNotification(integration *models.Integration, message string) error {
	var config SlackConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Slack configuration: %w", err)
	}

	payload := map[string]interface{}{
		"text": message,
		"channel": config.Channel,
		"username": config.Username,
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(config.WebhookURL, "application/json", 
		strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
