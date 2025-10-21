// Package models provides third-party integration data models for the Monitoring Service.
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Integration represents a third-party monitoring tool integration.
type Integration struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      uint           `gorm:"not null;index" json:"tenant_id"`
	Name          string         `gorm:"not null" json:"name"`
	Type          string         `gorm:"not null" json:"type"`          // pagerduty, newrelic, datadog, pingdom, etc.
	Description   string         `json:"description"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	Configuration string         `gorm:"type:text" json:"configuration"` // JSON configuration for the integration
	LastSyncAt    *time.Time     `json:"last_sync_at"`
	SyncStatus    string         `gorm:"default:pending" json:"sync_status"` // pending, syncing, success, error
	SyncError     string         `json:"sync_error"`
	SyncInterval  int            `gorm:"default:300" json:"sync_interval"` // Sync interval in seconds
	Metadata      string         `gorm:"type:text" json:"metadata"`       // JSON metadata

	// Related entities
	ComponentMappings []ComponentMapping  `gorm:"foreignKey:IntegrationID" json:"component_mappings,omitempty"`
	SyncLogs          []IntegrationSyncLog `gorm:"foreignKey:IntegrationID" json:"sync_logs,omitempty"`
}

// ComponentMapping maps external service components to internal components.
type ComponentMapping struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IntegrationID       uint           `gorm:"not null;index" json:"integration_id"`
	Integration         Integration    `gorm:"foreignKey:IntegrationID" json:"integration"`
	ComponentID         uint           `gorm:"not null;index" json:"component_id"`
	ExternalServiceID   string         `gorm:"not null" json:"external_service_id"`   // ID in external system
	ExternalServiceName string         `gorm:"not null" json:"external_service_name"` // Name in external system
	MappingConfig       string         `gorm:"type:text" json:"mapping_config"`       // JSON mapping configuration
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	LastStatusSync      *time.Time     `json:"last_status_sync"`
	Metadata            string         `gorm:"type:text" json:"metadata"` // JSON metadata
}

// IntegrationSyncLog represents a log entry for integration synchronization.
type IntegrationSyncLog struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	IntegrationID uint           `gorm:"not null;index" json:"integration_id"`
	Integration   Integration    `gorm:"foreignKey:IntegrationID" json:"integration"`
	SyncType      string         `gorm:"not null" json:"sync_type"`      // full, incremental, status_update
	Status        string         `gorm:"not null" json:"status"`         // started, completed, failed
	StartedAt     time.Time      `gorm:"not null" json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	Duration      int64          `json:"duration"`                       // Duration in milliseconds
	RecordsSync   int            `json:"records_sync"`                   // Number of records synchronized
	ErrorMessage  string         `json:"error_message"`
	Details       string         `gorm:"type:text" json:"details"`       // JSON details about the sync
	Metadata      string         `gorm:"type:text" json:"metadata"`      // JSON metadata
}

// PagerDutyConfig represents PagerDuty integration configuration.
type PagerDutyConfig struct {
	APIKey           string            `json:"api_key"`
	ServiceID        string            `json:"service_id,omitempty"`        // Optional: specific service to monitor
	EscalationPolicy string            `json:"escalation_policy,omitempty"` // Optional: escalation policy
	Teams            []string          `json:"teams,omitempty"`             // Optional: teams to filter
	Services         []PagerDutyService `json:"services,omitempty"`          // Services to monitor
	IncidentFields   []string          `json:"incident_fields,omitempty"`   // Fields to include in incidents
	SyncIncidents    bool              `json:"sync_incidents"`              // Whether to sync incidents
	SyncServices     bool              `json:"sync_services"`               // Whether to sync services
	WebhookEndpoint  string            `json:"webhook_endpoint,omitempty"`  // Optional: webhook for real-time updates
}

// PagerDutyService represents a PagerDuty service configuration.
type PagerDutyService struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	ComponentID uint   `json:"component_id,omitempty"` // Mapped component ID
}

// NewRelicConfig represents New Relic integration configuration.
type NewRelicConfig struct {
	APIKey         string              `json:"api_key"`
	AccountID      string              `json:"account_id"`
	Region         string              `json:"region"`                    // US or EU
	Applications   []NewRelicApp       `json:"applications,omitempty"`   // Applications to monitor
	Servers        []NewRelicServer    `json:"servers,omitempty"`        // Servers to monitor
	Synthetics     []NewRelicSynthetic `json:"synthetics,omitempty"`     // Synthetic monitors
	IncludeAPM     bool                `json:"include_apm"`              // Include APM data
	IncludeInfra   bool                `json:"include_infrastructure"`   // Include infrastructure data
	IncludeAlerts  bool                `json:"include_alerts"`           // Include alert policies
	MetricFilters  []string            `json:"metric_filters,omitempty"` // Metric name filters
}

// NewRelicApp represents a New Relic application configuration.
type NewRelicApp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Health      string `json:"health"`
	ComponentID uint   `json:"component_id,omitempty"` // Mapped component ID
}

// NewRelicServer represents a New Relic server configuration.
type NewRelicServer struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Health      string `json:"health"`
	ComponentID uint   `json:"component_id,omitempty"` // Mapped component ID
}

// NewRelicSynthetic represents a New Relic synthetic monitor configuration.
type NewRelicSynthetic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	ComponentID uint   `json:"component_id,omitempty"` // Mapped component ID
}

// DatadogConfig represents Datadog integration configuration.
type DatadogConfig struct {
	APIKey          string             `json:"api_key"`
	AppKey          string             `json:"app_key"`
	Site            string             `json:"site"`                      // datadoghq.com, datadoghq.eu, etc.
	Services        []DatadogService   `json:"services,omitempty"`       // Services to monitor
	Monitors        []DatadogMonitor   `json:"monitors,omitempty"`       // Monitors to sync
	Dashboards      []DatadogDashboard `json:"dashboards,omitempty"`     // Dashboards to sync
	IncludeMetrics  bool               `json:"include_metrics"`          // Include metrics data
	IncludeLogs     bool               `json:"include_logs"`             // Include logs data
	IncludeEvents   bool               `json:"include_events"`           // Include events data
	MetricFilters   []string           `json:"metric_filters,omitempty"` // Metric name filters
	TagFilters      []string           `json:"tag_filters,omitempty"`    // Tag filters
}

// DatadogService represents a Datadog service configuration.
type DatadogService struct {
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	ComponentID uint     `json:"component_id,omitempty"` // Mapped component ID
}

// DatadogMonitor represents a Datadog monitor configuration.
type DatadogMonitor struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	ComponentID uint   `json:"component_id,omitempty"` // Mapped component ID
}

// DatadogDashboard represents a Datadog dashboard configuration.
type DatadogDashboard struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// PingdomConfig represents Pingdom integration configuration.
type PingdomConfig struct {
	APIKey    string          `json:"api_key"`
	Username  string          `json:"username"`
	Password  string          `json:"password"`
	Checks    []PingdomCheck  `json:"checks,omitempty"`    // Checks to monitor
	TeamID    int             `json:"team_id,omitempty"`   // Optional: team to filter
	TagFilter []string        `json:"tag_filter,omitempty"` // Optional: tag filter
}

// PingdomCheck represents a Pingdom check configuration.
type PingdomCheck struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	ComponentID uint   `json:"component_id,omitempty"` // Mapped component ID
}

// IntegrationType constants for supported integration types.
const (
	IntegrationTypePagerDuty = "pagerduty"
	IntegrationTypeNewRelic  = "newrelic"
	IntegrationTypeDatadog   = "datadog"
	IntegrationTypePingdom   = "pingdom"
	IntegrationTypeUptimeRobot = "uptimerobot"
	IntegrationTypeStatusCake = "statuscake"
)

// SyncStatus constants for integration sync status.
const (
	SyncStatusPending  = "pending"
	SyncStatusSyncing  = "syncing"
	SyncStatusSuccess  = "success"
	SyncStatusError    = "error"
)

// SyncType constants for integration sync types.
const (
	SyncTypeFull        = "full"
	SyncTypeIncremental = "incremental"
	SyncTypeStatusUpdate = "status_update"
)

// Table name functions
func (Integration) TableName() string {
	return "integrations"
}

func (ComponentMapping) TableName() string {
	return "component_mappings"
}

func (IntegrationSyncLog) TableName() string {
	return "integration_sync_logs"
}

// Validation methods
func (i *Integration) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("integration name is required")
	}
	if i.Type == "" {
		return fmt.Errorf("integration type is required")
	}
	if i.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate integration type
	validTypes := []string{
		IntegrationTypePagerDuty,
		IntegrationTypeNewRelic,
		IntegrationTypeDatadog,
		IntegrationTypePingdom,
		IntegrationTypeUptimeRobot,
		IntegrationTypeStatusCake,
	}

	isValid := false
	for _, validType := range validTypes {
		if i.Type == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid integration type: %s", i.Type)
	}

	return nil
}

func (c *ComponentMapping) Validate() error {
	if c.IntegrationID == 0 {
		return fmt.Errorf("integration ID is required")
	}
	if c.ComponentID == 0 {
		return fmt.Errorf("component ID is required")
	}
	if c.ExternalServiceID == "" {
		return fmt.Errorf("external service ID is required")
	}
	if c.ExternalServiceName == "" {
		return fmt.Errorf("external service name is required")
	}

	return nil
}

// Helper functions for configuration parsing
func (i *Integration) GetPagerDutyConfig() (*PagerDutyConfig, error) {
	var config PagerDutyConfig
	if err := json.Unmarshal([]byte(i.Configuration), &config); err != nil {
		return nil, fmt.Errorf("failed to parse PagerDuty configuration: %w", err)
	}
	return &config, nil
}

func (i *Integration) GetNewRelicConfig() (*NewRelicConfig, error) {
	var config NewRelicConfig
	if err := json.Unmarshal([]byte(i.Configuration), &config); err != nil {
		return nil, fmt.Errorf("failed to parse New Relic configuration: %w", err)
	}
	return &config, nil
}

func (i *Integration) GetDatadogConfig() (*DatadogConfig, error) {
	var config DatadogConfig
	if err := json.Unmarshal([]byte(i.Configuration), &config); err != nil {
		return nil, fmt.Errorf("failed to parse Datadog configuration: %w", err)
	}
	return &config, nil
}

func (i *Integration) GetPingdomConfig() (*PingdomConfig, error) {
	var config PingdomConfig
	if err := json.Unmarshal([]byte(i.Configuration), &config); err != nil {
		return nil, fmt.Errorf("failed to parse Pingdom configuration: %w", err)
	}
	return &config, nil
}