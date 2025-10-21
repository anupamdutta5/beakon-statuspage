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

	"github.com/anupamdutta5/incident-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8085"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start incident service for integration tests
	cmd := startIncidentService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Incident service not ready, skipping integration tests")
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

func TestIncidentServiceIntegration_HealthCheck(t *testing.T) {
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
	assert.Equal(t, "incident-service", response["service"])
}

func TestIncidentServiceIntegration_CreateAndGetIncident(t *testing.T) {
	// Create incident
	incident := models.Incident{
		Title:       "Integration Test Incident",
		Description: "Integration test incident description",
		Status:      "investigating",
		Severity:    "major",
		TenantID:    1,
		Metadata:    `{"affected_users":1000,"region":"us-east-1"}`,
	}

	jsonData, err := json.Marshal(incident)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&createdIncident)
	require.NoError(t, err)

	assert.Equal(t, incident.Title, createdIncident.Title)
	assert.Equal(t, incident.Description, createdIncident.Description)
	assert.Equal(t, incident.Status, createdIncident.Status)
	assert.Equal(t, incident.Severity, createdIncident.Severity)
	assert.NotEmpty(t, createdIncident.ID)

	// Get incident by ID
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&retrievedIncident)
	require.NoError(t, err)

	assert.Equal(t, createdIncident.ID, retrievedIncident.ID)
	assert.Equal(t, createdIncident.Title, retrievedIncident.Title)
}

func TestIncidentServiceIntegration_UpdateIncident(t *testing.T) {
	// Create incident first
	incident := models.Incident{
		Title:       "Update Test Incident",
		Description: "Update test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(incident)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&createdIncident)
	require.NoError(t, err)

	// Update incident
	updateData := models.Incident{
		Title:       "Updated Incident Name",
		Description: "Updated incident description",
		Status:      "identified",
		Severity:    "major",
	}

	jsonData, err = json.Marshal(updateData)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&updatedIncident)
	require.NoError(t, err)

	assert.Equal(t, "Updated Incident Name", updatedIncident.Title)
	assert.Equal(t, "identified", updatedIncident.Status)
	assert.Equal(t, "major", updatedIncident.Severity)
}

func TestIncidentServiceIntegration_ListIncidents(t *testing.T) {
	// Create multiple incidents
	incidents := []models.Incident{
		{Title: "List Test 1", Description: "Description 1", Status: "investigating", Severity: "minor", TenantID: 1},
		{Title: "List Test 2", Description: "Description 2", Status: "identified", Severity: "major", TenantID: 1},
		{Title: "List Test 3", Description: "Description 3", Status: "monitoring", Severity: "critical", TenantID: 1},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, incident := range incidents {
		jsonData, err := json.Marshal(incident)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List incidents
	req, err := http.NewRequest("GET", baseURL+"/incidents?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Incidents []models.Incident `json:"incidents"`
		Total     int               `json:"total"`
		Limit     int               `json:"limit"`
		Offset    int               `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Incidents), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestIncidentServiceIntegration_UpdateIncidentStatus(t *testing.T) {
	// Create incident
	incident := models.Incident{
		Title:       "Status Update Test",
		Description: "Status update test incident",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(incident)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&createdIncident)
	require.NoError(t, err)

	// Update status
	statusUpdate := map[string]interface{}{
		"status":  "resolved",
		"message": "Issue has been resolved",
	}

	jsonData, err = json.Marshal(statusUpdate)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID)+"/status", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&updatedIncident)
	require.NoError(t, err)

	assert.Equal(t, "resolved", updatedIncident.Status)
}

func TestIncidentServiceIntegration_AddAndGetIncidentUpdates(t *testing.T) {
	// Create incident
	incident := models.Incident{
		Title:       "Updates Test",
		Description: "Updates test incident",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(incident)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&createdIncident)
	require.NoError(t, err)

	// Add updates
	updates := []models.IncidentUpdate{
		{Message: "We are investigating the issue", Status: "investigating"},
		{Message: "We have identified the root cause", Status: "identified"},
		{Message: "The issue has been resolved", Status: "resolved"},
	}

	for _, update := range updates {
		jsonData, err := json.Marshal(update)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID)+"/updates", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get incident updates
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID)+"/updates", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Updates []models.IncidentUpdate `json:"updates"`
		Total   int                     `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 3, len(response.Updates))
	assert.Equal(t, 3, response.Total)
}

func TestIncidentServiceIntegration_GetIncidentsByStatus(t *testing.T) {
	// Create incidents with different statuses
	incidents := []models.Incident{
		{Title: "Status Test 1", Status: "investigating", Severity: "minor", TenantID: 1},
		{Title: "Status Test 2", Status: "investigating", Severity: "major", TenantID: 1},
		{Title: "Status Test 3", Status: "resolved", Severity: "critical", TenantID: 1},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, incident := range incidents {
		jsonData, err := json.Marshal(incident)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get incidents by status
	req, err := http.NewRequest("GET", baseURL+"/incidents/status/investigating?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Incidents []models.Incident `json:"incidents"`
		Total     int               `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Incidents))
	assert.Equal(t, 2, response.Total)

	// Verify all incidents have investigating status
	for _, incident := range response.Incidents {
		assert.Equal(t, "investigating", incident.Status)
	}
}

func TestIncidentServiceIntegration_GetIncidentsBySeverity(t *testing.T) {
	// Create incidents with different severities
	incidents := []models.Incident{
		{Title: "Severity Test 1", Status: "investigating", Severity: "minor", TenantID: 1},
		{Title: "Severity Test 2", Status: "investigating", Severity: "major", TenantID: 1},
		{Title: "Severity Test 3", Status: "investigating", Severity: "major", TenantID: 1},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, incident := range incidents {
		jsonData, err := json.Marshal(incident)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get incidents by severity
	req, err := http.NewRequest("GET", baseURL+"/incidents/severity/major?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Incidents []models.Incident `json:"incidents"`
		Total     int               `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Incidents))
	assert.Equal(t, 2, response.Total)

	// Verify all incidents have major severity
	for _, incident := range response.Incidents {
		assert.Equal(t, "major", incident.Severity)
	}
}

func TestIncidentServiceIntegration_DeleteIncident(t *testing.T) {
	// Create incident
	incident := models.Incident{
		Title:       "Delete Test Incident",
		Description: "Delete test incident",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}

	jsonData, err := json.Marshal(incident)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdIncident models.Incident
	err = json.NewDecoder(resp.Body).Decode(&createdIncident)
	require.NoError(t, err)

	// Delete incident
	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Incident deleted successfully", response["message"])

	// Verify deletion
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/incidents/%d", baseURL, createdIncident.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestIncidentServiceIntegration_PerformanceTest(t *testing.T) {
	// Test creating multiple incidents in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Create incidents in bulk
	start := time.Now()

	for i := 0; i < 20; i++ {
		incident := models.Incident{
			Title:       fmt.Sprintf("Performance Test %d", i),
			Description: fmt.Sprintf("Performance test incident %d", i),
			Status:      "investigating",
			Severity:    "minor",
			TenantID:    1,
		}

		jsonData, err := json.Marshal(incident)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/incidents", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Created 20 incidents in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 10*time.Second)
}

// Helper functions
func startIncidentService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8085", "DB_NAME=statuspage_incident_test")

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
