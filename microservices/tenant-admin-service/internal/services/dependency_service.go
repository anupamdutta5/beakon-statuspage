// Package services provides business logic for dependency management.
package services

import (
	"fmt"

	"github.com/anupamdutta5/tenant-admin-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DependencyService provides operations for component dependency management.
type DependencyService struct {
	db *gorm.DB
}

// NewDependencyService creates a new DependencyService instance.
func NewDependencyService(db *gorm.DB) *DependencyService {
	return &DependencyService{db: db}
}

// GetDependencyGraph retrieves the complete dependency graph for a tenant.
func (s *DependencyService) GetDependencyGraph(tenantID uuid.UUID) (*models.DependencyGraph, error) {
	// Fetch all components
	var components []models.Component
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&components).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch components: %w", err)
	}

	// Fetch all dependency edges
	var edges []models.DependencyEdge
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("FromComponent").
		Preload("ToComponent").
		Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch dependency edges: %w", err)
	}

	// Build adjacency lists
	dependenciesMap := make(map[uuid.UUID][]uuid.UUID)   // component -> what it depends on
	dependentsMap := make(map[uuid.UUID][]uuid.UUID)     // component -> what depends on it

	for _, edge := range edges {
		dependenciesMap[edge.FromComponentID] = append(dependenciesMap[edge.FromComponentID], edge.ToComponentID)
		dependentsMap[edge.ToComponentID] = append(dependentsMap[edge.ToComponentID], edge.FromComponentID)
	}

	// Build component nodes
	nodes := make([]models.ComponentNode, 0, len(components))
	var rootComponents []uuid.UUID
	var leafComponents []uuid.UUID

	for _, comp := range components {
		node := models.ComponentNode{
			ID:           comp.ID,
			Name:         comp.Name,
			Status:       comp.Status,
			Type:         comp.GroupName, // Use group as type
			HealthScore:  s.calculateComponentHealthScore(comp.ID, dependenciesMap, components),
			Dependencies: dependenciesMap[comp.ID],
			Dependents:   dependentsMap[comp.ID],
		}
		nodes = append(nodes, node)

		// Identify root components (no dependencies)
		if len(node.Dependencies) == 0 {
			rootComponents = append(rootComponents, comp.ID)
		}

		// Identify leaf components (no dependents)
		if len(node.Dependents) == 0 {
			leafComponents = append(leafComponents, comp.ID)
		}
	}

	// Detect circular dependencies
	circularPaths := s.detectCircularDependencies(dependenciesMap)

	return &models.DependencyGraph{
		Nodes:          nodes,
		Edges:          edges,
		RootComponents: rootComponents,
		LeafComponents: leafComponents,
		CircularPaths:  circularPaths,
	}, nil
}

// AddDependency creates a new dependency edge between components.
func (s *DependencyService) AddDependency(tenantID, fromComponentID, toComponentID uuid.UUID, dependencyType string) error {
	// Validate components exist and belong to tenant
	var fromComp, toComp models.Component
	if err := s.db.Where("id = ? AND tenant_id = ?", fromComponentID, tenantID).First(&fromComp).Error; err != nil {
		return fmt.Errorf("from_component not found: %w", err)
	}
	if err := s.db.Where("id = ? AND tenant_id = ?", toComponentID, tenantID).First(&toComp).Error; err != nil {
		return fmt.Errorf("to_component not found: %w", err)
	}

	// Create dependency edge
	edge := &models.DependencyEdge{
		TenantID:        tenantID,
		FromComponentID: fromComponentID,
		ToComponentID:   toComponentID,
		DependencyType:  dependencyType,
	}

	if err := edge.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check for circular dependency
	validation := s.validateNoCycle(tenantID, fromComponentID, toComponentID)
	if !validation.Valid {
		return fmt.Errorf("circular dependency detected: %s", validation.ErrorMessage)
	}

	// Create the edge
	if err := s.db.Create(edge).Error; err != nil {
		return fmt.Errorf("failed to create dependency: %w", err)
	}

	return nil
}

// RemoveDependency deletes a dependency edge.
func (s *DependencyService) RemoveDependency(tenantID, fromComponentID, toComponentID uuid.UUID) error {
	result := s.db.Where("tenant_id = ? AND from_component_id = ? AND to_component_id = ?",
		tenantID, fromComponentID, toComponentID).
		Delete(&models.DependencyEdge{})

	if result.Error != nil {
		return fmt.Errorf("failed to remove dependency: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("dependency not found")
	}

	return nil
}

// AnalyzeImpact performs impact analysis for a component failure.
func (s *DependencyService) AnalyzeImpact(tenantID, componentID uuid.UUID) (*models.ImpactAnalysis, error) {
	// Get component
	var component models.Component
	if err := s.db.Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	// Get all edges for tenant
	var edges []models.DependencyEdge
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch edges: %w", err)
	}

	// Build reverse adjacency list (who depends on whom)
	dependentsMap := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		dependentsMap[edge.ToComponentID] = append(dependentsMap[edge.ToComponentID], edge.FromComponentID)
	}

	// BFS to find all affected components
	affected := make(map[uuid.UUID]int) // componentID -> path length
	queue := []uuid.UUID{componentID}
	affected[componentID] = 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentPathLen := affected[current]

		for _, dependent := range dependentsMap[current] {
			if _, visited := affected[dependent]; !visited {
				affected[dependent] = currentPathLen + 1
				queue = append(queue, dependent)
			}
		}
	}

	// Fetch affected component details
	var affectedComponents []models.AffectedComponentInfo
	directImpact := 0

	for compID, pathLen := range affected {
		if compID == componentID {
			continue // Skip self
		}

		var comp models.Component
		if err := s.db.Where("id = ?", compID).First(&comp).Error; err != nil {
			continue
		}

		if pathLen == 1 {
			directImpact++
		}

		// Build dependency path (simplified - just length for now)
		path := s.findPath(componentID, compID, dependentsMap)

		affectedComponents = append(affectedComponents, models.AffectedComponentInfo{
			ComponentID:    comp.ID,
			ComponentName:  comp.Name,
			ImpactLevel:    models.GetImpactLevel(pathLen),
			DependencyPath: path,
		})
	}

	// Generate mitigation suggestions
	suggestions := s.generateMitigationSuggestions(directImpact, len(affected)-1)

	return &models.ImpactAnalysis{
		ComponentID:           componentID,
		ComponentName:         component.Name,
		DirectImpactCount:     directImpact,
		TotalImpactCount:      len(affected) - 1, // Exclude self
		AffectedComponents:    affectedComponents,
		MitigationSuggestions: suggestions,
	}, nil
}

// GetDependencyHealth retrieves the health status of a component's dependencies.
func (s *DependencyService) GetDependencyHealth(tenantID, componentID uuid.UUID) (*models.DependencyHealth, error) {
	// Get component
	var component models.Component
	if err := s.db.Where("id = ? AND tenant_id = ?", componentID, tenantID).First(&component).Error; err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	// Get dependencies
	var edges []models.DependencyEdge
	if err := s.db.Where("tenant_id = ? AND from_component_id = ?", tenantID, componentID).
		Preload("ToComponent").
		Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies: %w", err)
	}

	healthyCount := 0
	degradedCount := 0
	failedCount := 0

	for _, edge := range edges {
		if edge.ToComponent == nil {
			continue
		}

		switch edge.ToComponent.Status {
		case "operational", "maintenance":
			healthyCount++
		case "degraded_performance":
			degradedCount++
		case "partial_outage", "major_outage":
			failedCount++
		}
	}

	healthScore := models.CalculateHealthScore(healthyCount, degradedCount, failedCount)

	return &models.DependencyHealth{
		ComponentID:          componentID,
		ComponentName:        component.Name,
		HealthScore:          healthScore,
		HealthyDependencies:  healthyCount,
		DegradedDependencies: degradedCount,
		FailedDependencies:   failedCount,
		TotalDependencies:    len(edges),
	}, nil
}

// validateNoCycle checks if adding a dependency would create a cycle.
func (s *DependencyService) validateNoCycle(tenantID, fromComponentID, toComponentID uuid.UUID) *models.DependencyValidationResult {
	// Get existing edges
	var edges []models.DependencyEdge
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&edges).Error; err != nil {
		return &models.DependencyValidationResult{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to fetch edges: %v", err),
		}
	}

	// Build adjacency list including the new edge
	adjList := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		adjList[edge.FromComponentID] = append(adjList[edge.FromComponentID], edge.ToComponentID)
	}
	adjList[fromComponentID] = append(adjList[fromComponentID], toComponentID)

	// DFS to detect cycle starting from toComponentID
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)
	var path []uuid.UUID

	if s.hasCycleDFS(toComponentID, fromComponentID, adjList, visited, recStack, &path) {
		// Build component names for path
		var compNames []string
		for _, id := range path {
			var comp models.Component
			if err := s.db.Where("id = ?", id).First(&comp).Error; err == nil {
				compNames = append(compNames, comp.Name)
			}
		}

		return &models.DependencyValidationResult{
			Valid:        false,
			CircularPath: compNames,
			ErrorMessage: fmt.Sprintf("adding this dependency would create a cycle: %v", compNames),
		}
	}

	return &models.DependencyValidationResult{Valid: true}
}

// hasCycleDFS performs DFS to detect if there's a path from start to target.
func (s *DependencyService) hasCycleDFS(
	current, target uuid.UUID,
	adjList map[uuid.UUID][]uuid.UUID,
	visited, recStack map[uuid.UUID]bool,
	path *[]uuid.UUID,
) bool {
	visited[current] = true
	recStack[current] = true
	*path = append(*path, current)

	// If we reached the target, we found a cycle
	if current == target {
		return true
	}

	for _, neighbor := range adjList[current] {
		if !visited[neighbor] {
			if s.hasCycleDFS(neighbor, target, adjList, visited, recStack, path) {
				return true
			}
		} else if recStack[neighbor] {
			*path = append(*path, neighbor)
			return true
		}
	}

	recStack[current] = false
	*path = (*path)[:len(*path)-1]
	return false
}

// detectCircularDependencies finds all circular dependencies in the graph.
func (s *DependencyService) detectCircularDependencies(adjList map[uuid.UUID][]uuid.UUID) [][]uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)
	var cycles [][]uuid.UUID

	for node := range adjList {
		if !visited[node] {
			var path []uuid.UUID
			s.findCyclesDFS(node, adjList, visited, recStack, &path, &cycles)
		}
	}

	return cycles
}

// findCyclesDFS performs DFS to find all cycles.
func (s *DependencyService) findCyclesDFS(
	current uuid.UUID,
	adjList map[uuid.UUID][]uuid.UUID,
	visited, recStack map[uuid.UUID]bool,
	path *[]uuid.UUID,
	cycles *[][]uuid.UUID,
) {
	visited[current] = true
	recStack[current] = true
	*path = append(*path, current)

	for _, neighbor := range adjList[current] {
		if !visited[neighbor] {
			s.findCyclesDFS(neighbor, adjList, visited, recStack, path, cycles)
		} else if recStack[neighbor] {
			// Found a cycle
			cycleStart := -1
			for i, node := range *path {
				if node == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart != -1 {
				cycle := make([]uuid.UUID, len(*path)-cycleStart)
				copy(cycle, (*path)[cycleStart:])
				*cycles = append(*cycles, cycle)
			}
		}
	}

	recStack[current] = false
	*path = (*path)[:len(*path)-1]
}

// findPath finds a path from source to target using BFS.
func (s *DependencyService) findPath(source, target uuid.UUID, adjList map[uuid.UUID][]uuid.UUID) []uuid.UUID {
	if source == target {
		return []uuid.UUID{source}
	}

	visited := make(map[uuid.UUID]bool)
	parent := make(map[uuid.UUID]uuid.UUID)
	queue := []uuid.UUID{source}
	visited[source] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == target {
			// Reconstruct path
			var path []uuid.UUID
			for node := target; node != source; node = parent[node] {
				path = append([]uuid.UUID{node}, path...)
			}
			path = append([]uuid.UUID{source}, path...)
			return path
		}

		for _, neighbor := range adjList[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	return []uuid.UUID{source, target} // Default path if not found
}

// calculateComponentHealthScore calculates health score based on dependencies.
func (s *DependencyService) calculateComponentHealthScore(
	componentID uuid.UUID,
	dependenciesMap map[uuid.UUID][]uuid.UUID,
	components []models.Component,
) float64 {
	dependencies := dependenciesMap[componentID]
	if len(dependencies) == 0 {
		return 100.0 // No dependencies = perfect health
	}

	// Build component status map
	statusMap := make(map[uuid.UUID]string)
	for _, comp := range components {
		statusMap[comp.ID] = comp.Status
	}

	healthyCount := 0
	degradedCount := 0
	failedCount := 0

	for _, depID := range dependencies {
		status := statusMap[depID]
		switch status {
		case "operational", "maintenance":
			healthyCount++
		case "degraded_performance":
			degradedCount++
		case "partial_outage", "major_outage":
			failedCount++
		}
	}

	return models.CalculateHealthScore(healthyCount, degradedCount, failedCount)
}

// generateMitigationSuggestions generates mitigation suggestions based on impact.
func (s *DependencyService) generateMitigationSuggestions(directImpact, totalImpact int) []string {
	var suggestions []string

	if directImpact > 5 {
		suggestions = append(suggestions, "This is a critical component with many direct dependents. Consider implementing redundancy or failover mechanisms.")
	}

	if totalImpact > 10 {
		suggestions = append(suggestions, "High cascade failure risk detected. Review architecture to reduce coupling.")
	}

	if directImpact > 0 {
		suggestions = append(suggestions, "Implement circuit breakers in dependent services to prevent cascade failures.")
		suggestions = append(suggestions, "Ensure dependent services have graceful degradation strategies.")
	}

	if totalImpact > 5 {
		suggestions = append(suggestions, "Consider implementing health checks and automatic failover for this component.")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "This component has low impact. No immediate mitigation needed.")
	}

	return suggestions
}
