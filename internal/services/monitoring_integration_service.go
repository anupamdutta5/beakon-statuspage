package services

import (
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

type MonitoringIntegrationService struct {
	db *gorm.DB
}

func NewMonitoringIntegrationService() *MonitoringIntegrationService {
	return &MonitoringIntegrationService{
		db: database.GetDB(),
	}
}

// Monitoring tool types
const (
	MonitoringPrometheus = "prometheus"
	MonitoringDatadog    = "datadog"
	MonitoringNewRelic   = "newrelic"
	MonitoringGrafana   = "grafana"
	MonitoringZabbix    = "zabbix"
)

// PrometheusConfig holds Prometheus-specific configuration
type PrometheusConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Query    string `json:"query"` // PromQL query for service status
}

// DatadogConfig holds Datadog-specific configuration
type DatadogConfig struct {
	APIKey    string `json:"api_key"`
	AppKey    string `json:"app_key"`
	Site      string `json:"site"` // e.g., datadoghq.com, datadoghq.eu
	MonitorID string `json:"monitor_id"`
}

// NewRelicConfig holds New Relic-specific configuration
type NewRelicConfig struct {
	APIKey     string `json:"api_key"`
	AccountID  string `json:"account_id"`
	QueryKey   string `json:"query_key"`
	ConditionID string `json:"condition_id"`
}

// GrafanaConfig holds Grafana-specific configuration
type GrafanaConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	DashboardID string `json:"dashboard_id"`
	PanelID  string `json:"panel_id"`
}

// MonitoringResult represents the result of a monitoring check
type MonitoringResult struct {
	ServiceID   uint      `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"`
	Value       float64   `json:"value"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
}

// SyncWithMonitoringTool syncs service statuses with an external monitoring tool
func (s *MonitoringIntegrationService) SyncWithMonitoringTool(integration *models.Integration) error {
	switch integration.Type {
	case MonitoringPrometheus:
		return s.syncWithPrometheus(integration)
	case MonitoringDatadog:
		return s.syncWithDatadog(integration)
	case MonitoringNewRelic:
		return s.syncWithNewRelic(integration)
	case MonitoringGrafana:
		return s.syncWithGrafana(integration)
	default:
		return fmt.Errorf("unsupported monitoring tool: %s", integration.Type)
	}
}

// syncWithPrometheus syncs with Prometheus monitoring
func (s *MonitoringIntegrationService) syncWithPrometheus(integration *models.Integration) error {
	var config PrometheusConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Prometheus configuration: %w", err)
	}

	// Query Prometheus for service statuses
	results, err := s.queryPrometheus(config)
	if err != nil {
		return fmt.Errorf("failed to query Prometheus: %w", err)
	}

	// Update service statuses based on results
	for _, result := range results {
		if err := s.updateServiceStatus(result); err != nil {
			logger.Error("Failed to update service status from Prometheus", 
				zap.Error(err),
				zap.String("service", result.ServiceName))
		}
	}

	// Update last sync time
	s.db.Model(integration).Update("last_sync_at", time.Now())
	return nil
}

// queryPrometheus queries Prometheus for service statuses
func (s *MonitoringIntegrationService) queryPrometheus(config PrometheusConfig) ([]MonitoringResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Build query URL
	queryURL := fmt.Sprintf("%s/api/v1/query", config.URL)
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication if provided
	if config.Username != "" && config.Password != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("query", config.Query)
	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Prometheus API returned status %d", resp.StatusCode)
	}

	// Parse response
	var prometheusResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&prometheusResp); err != nil {
		return nil, err
	}

	if prometheusResp.Status != "success" {
		return nil, fmt.Errorf("Prometheus query failed")
	}

	// Convert Prometheus results to our format
	var results []MonitoringResult
	for _, result := range prometheusResp.Data.Result {
		if len(result.Value) < 2 {
			continue
		}

		timestamp := time.Unix(int64(result.Value[0].(float64)), 0)
		value := result.Value[1].(float64)

		// Extract service name from metric labels
		serviceName := result.Metric["service"]
		if serviceName == "" {
			serviceName = result.Metric["instance"]
		}

		// Determine status based on value
		status := "operational"
		if value == 0 {
			status = "major_outage"
		} else if value < 0.5 {
			status = "degraded_performance"
		}

		results = append(results, MonitoringResult{
			ServiceName: serviceName,
			Status:      status,
			Value:       value,
			Timestamp:   timestamp,
			Source:      "prometheus",
		})
	}

	return results, nil
}

// syncWithDatadog syncs with Datadog monitoring
func (s *MonitoringIntegrationService) syncWithDatadog(integration *models.Integration) error {
	var config DatadogConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Datadog configuration: %w", err)
	}

	// Query Datadog for monitor statuses
	results, err := s.queryDatadog(config)
	if err != nil {
		return fmt.Errorf("failed to query Datadog: %w", err)
	}

	// Update service statuses based on results
	for _, result := range results {
		if err := s.updateServiceStatus(result); err != nil {
			logger.Error("Failed to update service status from Datadog", 
				zap.Error(err),
				zap.String("service", result.ServiceName))
		}
	}

	// Update last sync time
	s.db.Model(integration).Update("last_sync_at", time.Now())
	return nil
}

// queryDatadog queries Datadog for monitor statuses
func (s *MonitoringIntegrationService) queryDatadog(config DatadogConfig) ([]MonitoringResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Build query URL
	queryURL := fmt.Sprintf("https://api.%s/api/v1/monitor/%s", config.Site, config.MonitorID)
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	req.Header.Set("DD-API-KEY", config.APIKey)
	req.Header.Set("DD-APPLICATION-KEY", config.AppKey)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Datadog API returned status %d", resp.StatusCode)
	}

	// Parse response
	var datadogResp struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"overall_state"`
		Type   string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&datadogResp); err != nil {
		return nil, err
	}

	// Convert Datadog status to our format
	status := "operational"
	switch datadogResp.Status {
	case "Alert":
		status = "major_outage"
	case "Warn":
		status = "degraded_performance"
	case "No Data":
		status = "partial_outage"
	}

	return []MonitoringResult{
		{
			ServiceName: datadogResp.Name,
			Status:      status,
			Timestamp:   time.Now(),
			Source:      "datadog",
		},
	}, nil
}

// syncWithNewRelic syncs with New Relic monitoring
func (s *MonitoringIntegrationService) syncWithNewRelic(integration *models.Integration) error {
	var config NewRelicConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid New Relic configuration: %w", err)
	}

	// Query New Relic for alert conditions
	results, err := s.queryNewRelic(config)
	if err != nil {
		return fmt.Errorf("failed to query New Relic: %w", err)
	}

	// Update service statuses based on results
	for _, result := range results {
		if err := s.updateServiceStatus(result); err != nil {
			logger.Error("Failed to update service status from New Relic", 
				zap.Error(err),
				zap.String("service", result.ServiceName))
		}
	}

	// Update last sync time
	s.db.Model(integration).Update("last_sync_at", time.Now())
	return nil
}

// queryNewRelic queries New Relic for alert conditions
func (s *MonitoringIntegrationService) queryNewRelic(config NewRelicConfig) ([]MonitoringResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Build query URL
	queryURL := fmt.Sprintf("https://api.newrelic.com/v2/alerts_conditions/%s.json", config.ConditionID)
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	req.Header.Set("X-Api-Key", config.APIKey)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("New Relic API returned status %d", resp.StatusCode)
	}

	// Parse response
	var newRelicResp struct {
		AlertCondition struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Enabled bool  `json:"enabled"`
		} `json:"alert_condition"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&newRelicResp); err != nil {
		return nil, err
	}

	// Determine status based on alert condition
	status := "operational"
	if !newRelicResp.AlertCondition.Enabled {
		status = "major_outage"
	}

	return []MonitoringResult{
		{
			ServiceName: newRelicResp.AlertCondition.Name,
			Status:      status,
			Timestamp:   time.Now(),
			Source:      "newrelic",
		},
	}, nil
}

// syncWithGrafana syncs with Grafana monitoring
func (s *MonitoringIntegrationService) syncWithGrafana(integration *models.Integration) error {
	var config GrafanaConfig
	if err := json.Unmarshal([]byte(integration.Config), &config); err != nil {
		return fmt.Errorf("invalid Grafana configuration: %w", err)
	}

	// Query Grafana for panel data
	results, err := s.queryGrafana(config)
	if err != nil {
		return fmt.Errorf("failed to query Grafana: %w", err)
	}

	// Update service statuses based on results
	for _, result := range results {
		if err := s.updateServiceStatus(result); err != nil {
			logger.Error("Failed to update service status from Grafana", 
				zap.Error(err),
				zap.String("service", result.ServiceName))
		}
	}

	// Update last sync time
	s.db.Model(integration).Update("last_sync_at", time.Now())
	return nil
}

// queryGrafana queries Grafana for panel data
func (s *MonitoringIntegrationService) queryGrafana(config GrafanaConfig) ([]MonitoringResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Build query URL
	queryURL := fmt.Sprintf("%s/api/panels/%s", config.URL, config.PanelID)
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Grafana API returned status %d", resp.StatusCode)
	}

	// Parse response
	var grafanaResp struct {
		Panel struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"panel"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&grafanaResp); err != nil {
		return nil, err
	}

	// For now, assume operational status
	// In a real implementation, you would parse the actual panel data
	return []MonitoringResult{
		{
			ServiceName: grafanaResp.Panel.Title,
			Status:      "operational",
			Timestamp:   time.Now(),
			Source:      "grafana",
		},
	}, nil
}

// updateServiceStatus updates a service's status based on monitoring result
func (s *MonitoringIntegrationService) updateServiceStatus(result MonitoringResult) error {
	// Find service by name
	var service models.Service
	if err := s.db.Where("name = ?", result.ServiceName).First(&service).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Service doesn't exist, create it
			service = models.Service{
				Name:        result.ServiceName,
				Description: fmt.Sprintf("Service monitored by %s", result.Source),
				Status:      result.Status,
			}
			if err := s.db.Create(&service).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Update existing service
		service.Status = result.Status
		if err := s.db.Save(&service).Error; err != nil {
			return err
		}
	}

	// Create system metric record
	metric := models.SystemMetric{
		ServiceID: service.ID,
		Name:      "status_check",
		Value:     result.Value,
		Unit:      "status",
		Timestamp: result.Timestamp,
	}

	if err := s.db.Create(&metric).Error; err != nil {
		logger.Error("Failed to create system metric", zap.Error(err))
		// Don't return error as this is not critical
	}

	return nil
}

// SyncAllMonitoringTools syncs all active monitoring integrations
func (s *MonitoringIntegrationService) SyncAllMonitoringTools() error {
	var integrations []models.Integration
	if err := s.db.Where("type IN ? AND is_active = ?", 
		[]string{MonitoringPrometheus, MonitoringDatadog, MonitoringNewRelic, MonitoringGrafana}, 
		true).Find(&integrations).Error; err != nil {
		return err
	}

	for _, integration := range integrations {
		if err := s.SyncWithMonitoringTool(&integration); err != nil {
			logger.Error("Failed to sync with monitoring tool", 
				zap.String("tool", integration.Type),
				zap.String("name", integration.Name),
				zap.Error(err))
			// Continue with other integrations even if one fails
		}
	}

	return nil
}

// GetMonitoringStatus returns the status of all monitoring integrations
func (s *MonitoringIntegrationService) GetMonitoringStatus() ([]map[string]interface{}, error) {
	var integrations []models.Integration
	if err := s.db.Where("type IN ?", 
		[]string{MonitoringPrometheus, MonitoringDatadog, MonitoringNewRelic, MonitoringGrafana}).Find(&integrations).Error; err != nil {
		return nil, err
	}

	var statuses []map[string]interface{}
	for _, integration := range integrations {
		status := map[string]interface{}{
			"id":           integration.ID,
			"name":         integration.Name,
			"type":         integration.Type,
			"is_active":    integration.IsActive,
			"last_sync_at": integration.LastSyncAt,
		}

		// Test connection
		if integration.IsActive {
			if err := s.SyncWithMonitoringTool(&integration); err != nil {
				status["connection_status"] = "error"
				status["error"] = err.Error()
			} else {
				status["connection_status"] = "ok"
			}
		} else {
			status["connection_status"] = "inactive"
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
