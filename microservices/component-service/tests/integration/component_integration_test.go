package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-component-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8084"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start component service for integration tests
	cmd := startComponentService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Component service not ready, skipping integration tests")
		os.Exit(0)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if cmd != nil {
		cmd.Process.Kill()
	}

	os.Exit(code)
}

func TestComponentServiceIntegration_HealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/health", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "component-service", response["service"])
}

func TestComponentServiceIntegration_CreateAndGetComponent(t *testing.T) {
	// Create component
	component := models.Component{
		Name:        "Integration Test API",
		Description: "Integration test API component",
		Status:      "operational",
		TenantID:    1,
		Metadata:    `{"version":"1.0.0","region":"us-east-1"}`,
	}

	jsonData, err := json.Marshal(component)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&createdComponent)
	require.NoError(t, err)

	assert.Equal(t, component.Name, createdComponent.Name)
	assert.Equal(t, component.Description, createdComponent.Description)
	assert.Equal(t, component.Status, createdComponent.Status)
	assert.NotEmpty(t, createdComponent.ID)

	// Get component by ID
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/components/%d", baseURL, createdComponent.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&retrievedComponent)
	require.NoError(t, err)

	assert.Equal(t, createdComponent.ID, retrievedComponent.ID)
	assert.Equal(t, createdComponent.Name, retrievedComponent.Name)
}

func TestComponentServiceIntegration_UpdateComponent(t *testing.T) {
	// Create component first
	component := models.Component{
		Name:        "Update Test Component",
		Description: "Update test component description",
		Status:      "operational",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(component)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&createdComponent)
	require.NoError(t, err)

	// Update component
	updateData := models.Component{
		Name:        "Updated Component Name",
		Description: "Updated component description",
		Status:      "degraded_performance",
	}

	jsonData, err = json.Marshal(updateData)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/components/%d", baseURL, createdComponent.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&updatedComponent)
	require.NoError(t, err)

	assert.Equal(t, "Updated Component Name", updatedComponent.Name)
	assert.Equal(t, "degraded_performance", updatedComponent.Status)
	assert.Equal(t, "Updated component description", updatedComponent.Description)
}

func TestComponentServiceIntegration_ListComponents(t *testing.T) {
	// Create multiple components
	components := []models.Component{
		{Name: "List Test 1", Description: "Description 1", Status: "operational", TenantID: 1},
		{Name: "List Test 2", Description: "Description 2", Status: "degraded_performance", TenantID: 1},
		{Name: "List Test 3", Description: "Description 3", Status: "partial_outage", TenantID: 1},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, component := range components {
		jsonData, err := json.Marshal(component)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List components
	req, err := http.NewRequest("GET", baseURL+"/components?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Components []models.Component `json:"components"`
		Total      int                `json:"total"`
		Limit      int                `json:"limit"`
		Offset     int                `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Components), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestComponentServiceIntegration_UpdateComponentStatus(t *testing.T) {
	// Create component
	component := models.Component{
		Name:        "Status Update Test",
		Description: "Status update test component",
		Status:      "operational",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(component)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&createdComponent)
	require.NoError(t, err)

	// Update status
	statusUpdate := map[string]interface{}{
		"status":  "degraded_performance",
		"message": "Experiencing slow response times",
	}

	jsonData, err = json.Marshal(statusUpdate)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/components/%d", baseURL, createdComponent.ID)+"/status", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&updatedComponent)
	require.NoError(t, err)

	assert.Equal(t, "degraded_performance", updatedComponent.Status)
}

func TestComponentServiceIntegration_GetComponentsByCategory(t *testing.T) {
	// Create components with different categories
	components := []models.Component{
		{Name: "API Component 1", Status: "operational", TenantID: 1},
		{Name: "API Component 2", Status: "operational", TenantID: 1},
		{Name: "Database Component", Status: "operational", TenantID: 1},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, component := range components {
		jsonData, err := json.Marshal(component)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get components by category
	req, err := http.NewRequest("GET", baseURL+"/components/category/API?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Components []models.Component `json:"components"`
		Total      int                `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Components))
	assert.Equal(t, 2, response.Total)

	// Verify all components are returned
	for _, comp := range response.Components {
		assert.NotEmpty(t, comp.Name)
	}
}

func TestComponentServiceIntegration_GetComponentMetrics(t *testing.T) {
	// Create component
	component := models.Component{
		Name:        "Metrics Test Component",
		Description: "Metrics test component",
		Status:      "operational",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(component)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&createdComponent)
	require.NoError(t, err)

	// Get component metrics
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/components/%d", baseURL, createdComponent.ID)+"/metrics", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "uptime")
	assert.Contains(t, response, "response_time")
	assert.Contains(t, response, "error_rate")
}

func TestComponentServiceIntegration_BulkUpdateStatus(t *testing.T) {
	// Create multiple components
	components := []models.Component{
		{Name: "Bulk Test 1", Status: "operational", TenantID: 1},
		{Name: "Bulk Test 2", Status: "operational", TenantID: 1},
		{Name: "Bulk Test 3", Status: "operational", TenantID: 1},
	}

	client := &http.Client{Timeout: 5 * time.Second}
	var componentIDs []uint

	for _, component := range components {
		jsonData, err := json.Marshal(component)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)

		var createdComponent models.Component
		err = json.NewDecoder(resp.Body).Decode(&createdComponent)
		require.NoError(t, err)
		resp.Body.Close()

		componentIDs = append(componentIDs, createdComponent.ID)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Bulk update status
	bulkUpdate := map[string]interface{}{
		"component_ids": componentIDs,
		"status":        "maintenance",
		"message":       "Scheduled maintenance",
	}

	jsonData, err := json.Marshal(bulkUpdate)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", baseURL+"/components/bulk/status", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Bulk status update completed", response["message"])
	assert.Equal(t, 3, int(response["updated_count"].(float64)))

	// Verify all components are updated
	for _, id := range componentIDs {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/components/%d", baseURL, id), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		var comp models.Component
		err = json.NewDecoder(resp.Body).Decode(&comp)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, "maintenance", comp.Status)
	}
}

func TestComponentServiceIntegration_DeleteComponent(t *testing.T) {
	// Create component
	component := models.Component{
		Name:        "Delete Test Component",
		Description: "Delete test component",
		Status:      "operational",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(component)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdComponent models.Component
	err = json.NewDecoder(resp.Body).Decode(&createdComponent)
	require.NoError(t, err)

	// Delete component
	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/components/%d", baseURL, createdComponent.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Component deleted successfully", response["message"])

	// Verify deletion
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/components/%d", baseURL, createdComponent.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestComponentServiceIntegration_PerformanceTest(t *testing.T) {
	// Test creating multiple components in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Create components in bulk
	start := time.Now()

	for i := 0; i < 20; i++ {
		component := models.Component{
			Name:        fmt.Sprintf("Performance Test %d", i),
			Description: fmt.Sprintf("Performance test component %d", i),
			Status:      "operational",
			TenantID:    1,
		}

		jsonData, err := json.Marshal(component)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/components", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Created 20 components in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 10*time.Second)
}

// Helper functions
func startComponentService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8084", "DB_NAME=statuspage_component_test")

	// Start in background
	err := cmd.Start()
	if err != nil {
		return nil
	}
	return cmd
}

func waitForService(url string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	return false
}
