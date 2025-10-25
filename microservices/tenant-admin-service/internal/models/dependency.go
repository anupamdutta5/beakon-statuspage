// Package models provides data models for dependency management in Tenant Admin Service.
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DependencyEdge represents a dependency relationship between two components.
type DependencyEdge struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	FromComponentID uuid.UUID `gorm:"type:uuid;not null;index" json:"from_component_id"`
	ToComponentID   uuid.UUID `gorm:"type:uuid;not null;index" json:"to_component_id"`
	DependencyType  string    `gorm:"type:varchar(20);not null;default:hard" json:"dependency_type"` // hard or soft
	Description     string    `gorm:"type:text" json:"description,omitempty"`

	// Relationships
	FromComponent *Component `gorm:"foreignKey:FromComponentID;references:ID" json:"from_component,omitempty"`
	ToComponent   *Component `gorm:"foreignKey:ToComponentID;references:ID" json:"to_component,omitempty"`
}

// TableName returns the table name for DependencyEdge.
func (DependencyEdge) TableName() string {
	return "dependency_edges"
}

// Validate performs validation on DependencyEdge.
func (d *DependencyEdge) Validate() error {
	if d.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if d.FromComponentID == uuid.Nil {
		return fmt.Errorf("from_component_id is required")
	}
	if d.ToComponentID == uuid.Nil {
		return fmt.Errorf("to_component_id is required")
	}
	if d.FromComponentID == d.ToComponentID {
		return fmt.Errorf("component cannot depend on itself")
	}

	// Validate dependency type
	if d.DependencyType != "hard" && d.DependencyType != "soft" {
		return fmt.Errorf("dependency_type must be 'hard' or 'soft', got: %s", d.DependencyType)
	}

	return nil
}

// ComponentNode represents a component node in the dependency graph.
type ComponentNode struct {
	ID           uuid.UUID   `json:"id"`
	Name         string      `json:"name"`
	Status       string      `json:"status"`
	Type         string      `json:"type,omitempty"`
	HealthScore  float64     `json:"health_score"`
	Dependencies []uuid.UUID `json:"dependencies"` // Components this one depends on
	Dependents   []uuid.UUID `json:"dependents"`   // Components that depend on this one
}

// DependencyGraph represents the complete dependency graph for a tenant.
type DependencyGraph struct {
	Nodes           []ComponentNode  `json:"nodes"`
	Edges           []DependencyEdge `json:"edges"`
	RootComponents  []uuid.UUID      `json:"root_components"`  // Components with no dependencies
	LeafComponents  []uuid.UUID      `json:"leaf_components"`  // Components with no dependents
	CircularPaths   [][]uuid.UUID    `json:"circular_paths,omitempty"` // Detected circular dependencies
}

// ImpactAnalysis represents the impact analysis for a component failure.
type ImpactAnalysis struct {
	ComponentID           uuid.UUID                  `json:"component_id"`
	ComponentName         string                     `json:"component_name"`
	DirectImpactCount     int                        `json:"direct_impact_count"`
	TotalImpactCount      int                        `json:"total_impact_count"`
	AffectedComponents    []AffectedComponentInfo    `json:"affected_components"`
	MitigationSuggestions []string                   `json:"mitigation_suggestions"`
}

// AffectedComponentInfo represents a component affected by a failure.
type AffectedComponentInfo struct {
	ComponentID    uuid.UUID   `json:"component_id"`
	ComponentName  string      `json:"component_name"`
	ImpactLevel    string      `json:"impact_level"` // critical, high, medium, low
	DependencyPath []uuid.UUID `json:"dependency_path"`
}

// DependencyHealth represents the health status of a component's dependencies.
type DependencyHealth struct {
	ComponentID          uuid.UUID `json:"component_id"`
	ComponentName        string    `json:"component_name"`
	HealthScore          float64   `json:"health_score"` // 0-100
	HealthyDependencies  int       `json:"healthy_dependencies"`
	DegradedDependencies int       `json:"degraded_dependencies"`
	FailedDependencies   int       `json:"failed_dependencies"`
	TotalDependencies    int       `json:"total_dependencies"`
}

// DependencyValidationResult represents the result of dependency validation.
type DependencyValidationResult struct {
	Valid          bool     `json:"valid"`
	CircularPath   []string `json:"circular_path,omitempty"`
	ErrorMessage   string   `json:"error_message,omitempty"`
}

// GetImpactLevel determines the impact level based on dependency path length.
func GetImpactLevel(pathLength int) string {
	switch {
	case pathLength == 1:
		return "critical"
	case pathLength == 2:
		return "high"
	case pathLength == 3:
		return "medium"
	default:
		return "low"
	}
}

// CalculateHealthScore calculates the health score based on dependency statuses.
func CalculateHealthScore(healthyCount, degradedCount, failedCount int) float64 {
	total := healthyCount + degradedCount + failedCount
	if total == 0 {
		return 100.0 // No dependencies = perfect health
	}

	// Weight: healthy=100%, degraded=50%, failed=0%
	score := (float64(healthyCount)*100.0 + float64(degradedCount)*50.0) / float64(total)
	return score
}

// GetStatusPriority returns priority for status (higher = worse).
func GetStatusPriority(status string) int {
	statusPriority := map[string]int{
		"operational":          0,
		"maintenance":          1,
		"degraded_performance": 2,
		"partial_outage":       3,
		"major_outage":         4,
	}

	if priority, exists := statusPriority[status]; exists {
		return priority
	}
	return 0
}
