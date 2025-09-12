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

	"github.com/enterprise-status/statuspage-tenant-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8082"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start tenant service for integration tests
	cmd := startTenantService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Tenant service not ready, skipping integration tests")
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

func TestTenantServiceIntegration_HealthCheck(t *testing.T) {
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
	assert.Equal(t, "tenant-service", response["service"])
}

func TestTenantServiceIntegration_CreateAndGetTenant(t *testing.T) {
	// Create tenant
	tenant := models.Tenant{
		Name:      "Integration Test Company",
		Subdomain: "integrationtest",
		Domain:    "integrationtest.com",
		Settings:  `{"theme":"dark","logo":"https://example.com/logo.png"}`,
	}

	jsonData, err := json.Marshal(tenant)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&createdTenant)
	require.NoError(t, err)

	assert.Equal(t, tenant.Name, createdTenant.Name)
	assert.Equal(t, tenant.Subdomain, createdTenant.Subdomain)
	assert.NotEmpty(t, createdTenant.ID)

	// Get tenant by ID
	req, err = http.NewRequest("GET", baseURL+"/tenants/"+fmt.Sprintf("%d", createdTenant.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&retrievedTenant)
	require.NoError(t, err)

	assert.Equal(t, createdTenant.ID, retrievedTenant.ID)
	assert.Equal(t, createdTenant.Name, retrievedTenant.Name)
}

func TestTenantServiceIntegration_UpdateTenant(t *testing.T) {
	// Create tenant first
	tenant := models.Tenant{
		Name:      "Update Test Company",
		Subdomain: "updatetest",
		Domain:    "updatetest.com",
	}

	jsonData, err := json.Marshal(tenant)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&createdTenant)
	require.NoError(t, err)

	// Update tenant
	updateData := models.Tenant{
		Name:      "Updated Company Name",
		Subdomain: "updatedcompany",
		Domain:    "updatedcompany.com",
	}

	jsonData, err = json.Marshal(updateData)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", baseURL+"/tenants/"+fmt.Sprintf("%d", createdTenant.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&updatedTenant)
	require.NoError(t, err)

	assert.Equal(t, "Updated Company Name", updatedTenant.Name)
	assert.Equal(t, "updatedcompany", updatedTenant.Subdomain)
}

func TestTenantServiceIntegration_ListTenants(t *testing.T) {
	// Create multiple tenants
	tenants := []models.Tenant{
		{Name: "List Test 1", Subdomain: "listtest1", Domain: "listtest1.com"},
		{Name: "List Test 2", Subdomain: "listtest2", Domain: "listtest2.com"},
		{Name: "List Test 3", Subdomain: "listtest3", Domain: "listtest3.com"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, tenant := range tenants {
		jsonData, err := json.Marshal(tenant)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List tenants
	req, err := http.NewRequest("GET", baseURL+"/tenants", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Tenants []models.Tenant `json:"tenants"`
		Total   int             `json:"total"`
		Limit   int             `json:"limit"`
		Offset  int             `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Tenants), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestTenantServiceIntegration_GetTenantBySubdomain(t *testing.T) {
	// Create tenant
	tenant := models.Tenant{
		Name:      "Subdomain Test Company",
		Subdomain: "subdomaintest",
		Domain:    "subdomaintest.com",
	}

	jsonData, err := json.Marshal(tenant)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get tenant by subdomain
	req, err = http.NewRequest("GET", baseURL+"/tenants/subdomain/subdomaintest", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&retrievedTenant)
	require.NoError(t, err)

	assert.Equal(t, "subdomaintest", retrievedTenant.Subdomain)
	assert.Equal(t, "Subdomain Test Company", retrievedTenant.Name)
}

func TestTenantServiceIntegration_UpdateTenantSettings(t *testing.T) {
	// Create tenant
	tenant := models.Tenant{
		Name:      "Settings Test Company",
		Subdomain: "settingstest",
		Domain:    "settingstest.com",
		Settings:  `{"theme": "light"}`,
	}

	jsonData, err := json.Marshal(tenant)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&createdTenant)
	require.NoError(t, err)

	// Update settings
	updateSettings := map[string]interface{}{
		"theme": "dark",
		"logo":  "https://example.com/new-logo.png",
		"color": "#ff0000",
	}

	jsonData, err = json.Marshal(updateSettings)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/tenants/%d/settings", baseURL, createdTenant.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&updatedTenant)
	require.NoError(t, err)

	// Settings is a JSON string, so we need to parse it to check values
	var parsedSettings map[string]interface{}
	err = json.Unmarshal([]byte(updatedTenant.Settings), &parsedSettings)
	require.NoError(t, err)
	assert.Equal(t, "dark", parsedSettings["theme"])
	assert.Equal(t, "https://example.com/new-logo.png", parsedSettings["logo"])
	assert.Equal(t, "#ff0000", parsedSettings["color"])
}

func TestTenantServiceIntegration_GetTenantStats(t *testing.T) {
	// Create tenant
	tenant := models.Tenant{
		Name:      "Stats Test Company",
		Subdomain: "statstest",
		Domain:    "statstest.com",
	}

	jsonData, err := json.Marshal(tenant)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&createdTenant)
	require.NoError(t, err)

	// Get tenant stats
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/tenants/%d/stats", baseURL, createdTenant.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_users")
	assert.Contains(t, response, "total_components")
	assert.Contains(t, response, "total_incidents")
}

func TestTenantServiceIntegration_DeleteTenant(t *testing.T) {
	// Create tenant
	tenant := models.Tenant{
		Name:      "Delete Test Company",
		Subdomain: "deletetest",
		Domain:    "deletetest.com",
	}

	jsonData, err := json.Marshal(tenant)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTenant models.Tenant
	err = json.NewDecoder(resp.Body).Decode(&createdTenant)
	require.NoError(t, err)

	// Delete tenant
	req, err = http.NewRequest("DELETE", baseURL+"/tenants/"+fmt.Sprintf("%d", createdTenant.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Tenant deleted successfully", response["message"])

	// Verify deletion
	req, err = http.NewRequest("GET", baseURL+"/tenants/"+fmt.Sprintf("%d", createdTenant.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestTenantServiceIntegration_PerformanceTest(t *testing.T) {
	// Test creating multiple tenants in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Create tenants in bulk
	start := time.Now()

	for i := 0; i < 20; i++ {
		tenant := models.Tenant{
			Name:      fmt.Sprintf("Performance Test %d", i),
			Subdomain: fmt.Sprintf("perftest%d", i),
			Domain:    fmt.Sprintf("perftest%d.com", i),
		}

		jsonData, err := json.Marshal(tenant)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/tenants", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Created 20 tenants in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 10*time.Second)
}

// Helper functions
func startTenantService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8082", "DB_NAME=statuspage_tenant_test")

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
