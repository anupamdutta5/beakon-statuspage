// Package models provides custom metrics monitoring data models.
package models

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/utils"
	"gorm.io/gorm"
)

// CustomMetric represents a custom metric being monitored.
type CustomMetric struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID      uint           `gorm:"not null;index" json:"tenant_id"`
	Name          string         `gorm:"not null;size:255;uniqueIndex:idx_tenant_name" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	Type          string         `gorm:"not null;size:50" json:"type"` // counter, gauge, histogram, summary
	Unit          string         `gorm:"size:50" json:"unit"`          // seconds, bytes, requests, etc.
	Labels        string         `gorm:"type:text" json:"labels"`      // JSON string for metric labels
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	RetentionDays int            `gorm:"default:30" json:"retention_days"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	DataPoints []CustomMetricDataPoint `gorm:"foreignKey:MetricID;constraint:OnDelete:CASCADE" json:"data_points,omitempty"`
}

// CustomMetricDataPoint represents a data point for a custom metric.
type CustomMetricDataPoint struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	MetricID  uint           `gorm:"not null;index" json:"metric_id"`
	Metric    CustomMetric   `gorm:"foreignKey:MetricID" json:"metric"`
	Value     float64        `gorm:"not null" json:"value"`
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
	Labels    string         `gorm:"type:text" json:"labels"`   // JSON string for metric labels
	Source    string         `gorm:"size:100" json:"source"`    // api, webhook, agent, etc.
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for CustomMetric.
func (CustomMetric) TableName() string {
	return "custom_metrics"
}

// TableName returns the table name for CustomMetricDataPoint.
func (CustomMetricDataPoint) TableName() string {
	return "custom_metric_data_points"
}

// Validate performs validation on CustomMetric.
func (c *CustomMetric) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("metric name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("metric type is required")
	}
	if c.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}
	if c.RetentionDays <= 0 {
		return fmt.Errorf("retention days must be greater than 0")
	}

	// Validate metric type
	validTypes := []string{"counter", "gauge", "histogram", "summary"}
	if !utils.ContainsString(validTypes, c.Type) {
		return fmt.Errorf("invalid metric type: %s", c.Type)
	}

	return nil
}

// Validate performs validation on CustomMetricDataPoint.
func (c *CustomMetricDataPoint) Validate() error {
	if c.MetricID == 0 {
		return fmt.Errorf("metric ID is required")
	}
	if c.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}
