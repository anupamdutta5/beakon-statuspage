// Package models provides Docker monitoring data models.
package models

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/monitoring-service/internal/utils"
	"gorm.io/gorm"
)

// DockerContainer represents a Docker container being monitored.
type DockerContainer struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Image       string         `gorm:"size:500" json:"image"`
	Status      string         `gorm:"default:unknown;size:50" json:"status"` // running, stopped, paused, restarting, dead, created
	State       string         `gorm:"size:50" json:"state"`                  // running, exited, paused, restarting, dead, created, removing
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	LastChecked *time.Time     `json:"last_checked"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	HealthChecks []DockerContainerHealthCheck `gorm:"foreignKey:ContainerID;constraint:OnDelete:CASCADE" json:"health_checks,omitempty"`
}

// DockerContainerHealthCheck represents a health check for a Docker container.
type DockerContainerHealthCheck struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
	ContainerID uint            `gorm:"not null;index" json:"container_id"`
	Container   DockerContainer `gorm:"foreignKey:ContainerID" json:"container"`
	Type        string          `gorm:"not null;size:50" json:"type"`          // health, readiness, liveness
	Status      string          `gorm:"default:unknown;size:50" json:"status"` // healthy, unhealthy, starting, none
	LastChecked *time.Time      `json:"last_checked"`
	Message     string          `gorm:"type:text" json:"message"`
	Metadata    string          `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for DockerContainer.
func (DockerContainer) TableName() string {
	return "docker_containers"
}

// TableName returns the table name for DockerContainerHealthCheck.
func (DockerContainerHealthCheck) TableName() string {
	return "docker_container_health_checks"
}

// Validate performs validation on DockerContainer.
func (d *DockerContainer) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("container name is required")
	}
	if d.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate container status
	validStatuses := []string{
		"running", "stopped", "paused", "restarting",
		"dead", "created", "removing",
	}
	if d.Status != "" && !utils.ContainsString(validStatuses, d.Status) {
		return fmt.Errorf("invalid container status: %s", d.Status)
	}

	return nil
}

// Validate performs validation on DockerContainerHealthCheck.
func (d *DockerContainerHealthCheck) Validate() error {
	if d.ContainerID == 0 {
		return fmt.Errorf("container ID is required")
	}
	if d.Type == "" {
		return fmt.Errorf("health check type is required")
	}

	// Validate health check type
	validTypes := []string{"health", "readiness", "liveness"}
	if !utils.ContainsString(validTypes, d.Type) {
		return fmt.Errorf("invalid health check type: %s", d.Type)
	}

	// Validate health check status
	validStatuses := []string{"healthy", "unhealthy", "starting", "none", "unknown"}
	if d.Status != "" && !utils.ContainsString(validStatuses, d.Status) {
		return fmt.Errorf("invalid health check status: %s", d.Status)
	}

	return nil
}
