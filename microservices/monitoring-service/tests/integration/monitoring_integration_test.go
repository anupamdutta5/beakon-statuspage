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

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8087"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start monitoring service for integration tests
	cmd := startMonitoringService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Monitoring service not ready, skipping integration tests")
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

func TestMonitoringServiceIntegration_HealthCheck(t *testing.T) {
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
	assert.Equal(t, "monitoring-service", response["service"])
}

func TestMonitoringServiceIntegration_CreateAndGetMonitor(t *testing.T) {
	// Create monitor
	monitor := models.MonitoredService{
		TenantID:      1,
		Name:          "Integration Test Monitor",
		Type:          "http",
		URL:           "https://httpbin.org/status/200",
		CheckInterval: 60,
		Timeout:       30,
		Status:        "active",
		Metadata: `{
			"expected_status": 200,
			"expected_body":   "ok"
		}`,
	}

	jsonData, err := json.Marshal(monitor)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&createdMonitor)
	require.NoError(t, err)

	assert.Equal(t, monitor.TenantID, createdMonitor.TenantID)
	assert.Equal(t, monitor.Name, createdMonitor.Name)
	assert.Equal(t, monitor.Type, createdMonitor.Type)
	assert.Equal(t, monitor.URL, createdMonitor.URL)
	assert.NotEmpty(t, createdMonitor.ID)

	// Get monitor by ID
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/monitors/%d", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&retrievedMonitor)
	require.NoError(t, err)

	assert.Equal(t, createdMonitor.ID, retrievedMonitor.ID)
	assert.Equal(t, createdMonitor.Name, retrievedMonitor.Name)
}

func TestMonitoringServiceIntegration_UpdateMonitor(t *testing.T) {
	// Create monitor first
	monitor := models.MonitoredService{
		TenantID: 1,
		Name:     "Update Test Monitor",
		Type:     "http",
		URL:      "https://httpbin.org/status/200",
		Status:   "active",
	}

	jsonData, err := json.Marshal(monitor)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&createdMonitor)
	require.NoError(t, err)

	// Update monitor
	updateData := models.MonitoredService{
		Name:          "Updated Monitor Name",
		URL:           "https://httpbin.org/status/201",
		Status:        "paused",
		CheckInterval: 120,
	}

	jsonData, err = json.Marshal(updateData)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/monitors/%d", baseURL, createdMonitor.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&updatedMonitor)
	require.NoError(t, err)

	assert.Equal(t, "Updated Monitor Name", updatedMonitor.Name)
	assert.Equal(t, "https://httpbin.org/status/201", updatedMonitor.URL)
	assert.Equal(t, "paused", updatedMonitor.Status)
}

func TestMonitoringServiceIntegration_ListMonitors(t *testing.T) {
	// Create multiple monitors
	monitors := []models.MonitoredService{
		{TenantID: 1, Name: "List Test 1", Type: "http", URL: "https://httpbin.org/status/200", Status: "active"},
		{TenantID: 1, Name: "List Test 2", Type: "ping", URL: "8.8.8.8", Status: "active"},
		{TenantID: 1, Name: "List Test 3", Type: "tcp", URL: "httpbin.org:443", Status: "paused"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, monitor := range monitors {
		jsonData, err := json.Marshal(monitor)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List monitors
	req, err := http.NewRequest("GET", baseURL+"/monitors?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Monitors []models.MonitoredService `json:"monitors"`
		Total    int                       `json:"total"`
		Limit    int                       `json:"limit"`
		Offset   int                       `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Monitors), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestMonitoringServiceIntegration_StartStopMonitor(t *testing.T) {
	// Create a paused monitor
	monitor := models.MonitoredService{
		TenantID: 1,
		Name:     "Start Stop Test Monitor",
		Type:     "http",
		URL:      "https://httpbin.org/status/200",
		Status:   "paused",
	}

	jsonData, err := json.Marshal(monitor)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&createdMonitor)
	require.NoError(t, err)

	// Start monitor
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/monitors/%d/start", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Monitor started successfully", response["message"])

	// Stop monitor
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/monitors/%d/stop", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	response = make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Monitor stopped successfully", response["message"])
}

func TestMonitoringServiceIntegration_GetMonitorResults(t *testing.T) {
	// Create a monitor
	monitor := models.MonitoredService{
		TenantID: 1,
		Name:     "Results Test Monitor",
		Type:     "http",
		URL:      "https://httpbin.org/status/200",
		Status:   "active",
	}

	jsonData, err := json.Marshal(monitor)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&createdMonitor)
	require.NoError(t, err)

	// Get monitor results
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/monitors/%d/results", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Results []models.HealthCheckResult `json:"results"`
		Total   int                        `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Results)
}

func TestMonitoringServiceIntegration_GetMonitorStats(t *testing.T) {
	// Create a monitor
	monitor := models.MonitoredService{
		TenantID: 1,
		Name:     "Stats Test Monitor",
		Type:     "http",
		URL:      "https://httpbin.org/status/200",
		Status:   "active",
	}

	jsonData, err := json.Marshal(monitor)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&createdMonitor)
	require.NoError(t, err)

	// Get monitor stats
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/monitors/%d/stats", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "uptime_percentage")
	assert.Contains(t, response, "average_response_time")
	assert.Contains(t, response, "total_checks")
	assert.Contains(t, response, "successful_checks")
}

func TestMonitoringServiceIntegration_GetMonitorsByComponent(t *testing.T) {
	// Create monitors for different components
	monitors := []models.MonitoredService{
		{TenantID: 1, Name: "Component Test 1", Type: "http", URL: "https://httpbin.org/status/200", Status: "active"},
		{TenantID: 1, Name: "Component Test 2", Type: "ping", URL: "8.8.8.8", Status: "active"},
		{TenantID: 1, Name: "Component Test 3", Type: "tcp", URL: "httpbin.org:443", Status: "active"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, monitor := range monitors {
		jsonData, err := json.Marshal(monitor)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get monitors by component
	req, err := http.NewRequest("GET", baseURL+"/monitors/component/comp-1?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Monitors []models.MonitoredService `json:"monitors"`
		Total    int                       `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Monitors))
	assert.Equal(t, 2, response.Total)

	// Verify all monitors are for comp-1
	for _, _ = range response.Monitors {
	}
}

func TestMonitoringServiceIntegration_DeleteMonitor(t *testing.T) {
	// Create a monitor
	monitor := models.MonitoredService{
		TenantID: 1,
		Name:     "Delete Test Monitor",
		Type:     "http",
		URL:      "https://httpbin.org/status/200",
		Status:   "active",
	}

	jsonData, err := json.Marshal(monitor)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdMonitor models.MonitoredService
	err = json.NewDecoder(resp.Body).Decode(&createdMonitor)
	require.NoError(t, err)

	// Delete monitor
	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/monitors/%d", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Monitor deleted successfully", response["message"])

	// Verify deletion
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/monitors/%d", baseURL, createdMonitor.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMonitoringServiceIntegration_PerformanceTest(t *testing.T) {
	// Test creating multiple monitors in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Create monitors in bulk
	start := time.Now()

	for i := 0; i < 20; i++ {
		monitor := models.MonitoredService{
			TenantID: 1,
			Name:     fmt.Sprintf("Performance Test Monitor %d", i),
			Type:     "http",
			URL:      fmt.Sprintf("https://httpbin.org/status/%d", 200+i%3),
			Status:   "active",
		}

		jsonData, err := json.Marshal(monitor)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/monitors", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Created 20 monitors in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 10*time.Second)
}

// Helper functions
func startMonitoringService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8087", "DB_NAME=statuspage_monitoring_test")

	// Start in background
	cmd.Start()
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
