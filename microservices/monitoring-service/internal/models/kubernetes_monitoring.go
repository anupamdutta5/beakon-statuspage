// Package models provides Kubernetes monitoring data models.
package models

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-monitoring-service/internal/utils"
	"gorm.io/gorm"
)

// KubernetesResource represents a Kubernetes resource being monitored.
type KubernetesResource struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Namespace   string         `gorm:"size:100" json:"namespace"`
	Type        string         `gorm:"not null;size:50" json:"type"`          // deployment, pod, service, ingress, configmap, secret
	Kind        string         `gorm:"not null;size:50" json:"kind"`          // Deployment, Pod, Service, etc.
	Status      string         `gorm:"default:unknown;size:50" json:"status"` // running, pending, failed, unknown
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	LastChecked *time.Time     `json:"last_checked"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data

	// Related entities
	Events []KubernetesEvent `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE" json:"events,omitempty"`
}

// KubernetesEvent represents a Kubernetes event.
type KubernetesEvent struct {
	ID         uint               `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	DeletedAt  gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
	ResourceID uint               `gorm:"not null;index" json:"resource_id"`
	Resource   KubernetesResource `gorm:"foreignKey:ResourceID" json:"resource"`
	Type       string             `gorm:"not null;size:50" json:"type"` // Normal, Warning, Error
	Reason     string             `gorm:"size:100" json:"reason"`
	Message    string             `gorm:"type:text" json:"message"`
	Count      int                `gorm:"default:1" json:"count"`
	FirstSeen  time.Time          `gorm:"not null" json:"first_seen"`
	LastSeen   time.Time          `gorm:"not null" json:"last_seen"`
	Metadata   string             `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for KubernetesResource.
func (KubernetesResource) TableName() string {
	return "kubernetes_resources"
}

// TableName returns the table name for KubernetesEvent.
func (KubernetesEvent) TableName() string {
	return "kubernetes_events"
}

// Validate performs validation on KubernetesResource.
func (k *KubernetesResource) Validate() error {
	if k.Name == "" {
		return fmt.Errorf("resource name is required")
	}
	if k.Type == "" {
		return fmt.Errorf("resource type is required")
	}
	if k.Kind == "" {
		return fmt.Errorf("resource kind is required")
	}
	if k.TenantID == 0 {
		return fmt.Errorf("tenant ID is required")
	}

	// Validate resource type
	validTypes := []string{
		"deployment", "pod", "service", "ingress",
		"configmap", "secret", "persistentvolume",
		"persistentvolumeclaim", "statefulset", "daemonset",
	}
	if !utils.ContainsString(validTypes, k.Type) {
		return fmt.Errorf("invalid resource type: %s", k.Type)
	}

	return nil
}

// Validate performs validation on KubernetesEvent.
func (k *KubernetesEvent) Validate() error {
	if k.ResourceID == 0 {
		return fmt.Errorf("resource ID is required")
	}
	if k.Type == "" {
		return fmt.Errorf("event type is required")
	}
	if k.FirstSeen.IsZero() {
		return fmt.Errorf("first seen time is required")
	}
	if k.LastSeen.IsZero() {
		return fmt.Errorf("last seen time is required")
	}
	if k.Count <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}

	// Validate event type
	validTypes := []string{"Normal", "Warning", "Error"}
	if !utils.ContainsString(validTypes, k.Type) {
		return fmt.Errorf("invalid event type: %s", k.Type)
	}

	return nil
}
