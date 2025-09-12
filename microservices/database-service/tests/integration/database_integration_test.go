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

	"github.com/enterprise-status/statuspage-database-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8089"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start database service for integration tests
	cmd := startDatabaseService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Database service not ready, skipping integration tests")
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

func TestDatabaseServiceIntegration_HealthCheck(t *testing.T) {
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
	assert.Equal(t, "database-service", response["service"])
}

func TestDatabaseServiceIntegration_CreateAndGetDatabase(t *testing.T) {
	// Create database
	database := models.Database{
		Name:     "integration_test_db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "integration_test_db",
	}

	jsonData, err := json.Marshal(database)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/databases", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdDatabase models.Database
	err = json.NewDecoder(resp.Body).Decode(&createdDatabase)
	require.NoError(t, err)

	assert.Equal(t, database.Name, createdDatabase.Name)
	assert.Equal(t, database.Type, createdDatabase.Type)
	assert.NotEmpty(t, createdDatabase.ID)

	// Get database by ID
	req, err = http.NewRequest("GET", baseURL+"/databases/"+fmt.Sprintf("%d", createdDatabase.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedDatabase models.Database
	err = json.NewDecoder(resp.Body).Decode(&retrievedDatabase)
	require.NoError(t, err)

	assert.Equal(t, createdDatabase.ID, retrievedDatabase.ID)
	assert.Equal(t, createdDatabase.Name, retrievedDatabase.Name)
}

func TestDatabaseServiceIntegration_UpdateDatabase(t *testing.T) {
	// Create database first
	database := models.Database{
		Name:     "update_test_db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "update_test_db",
	}

	jsonData, err := json.Marshal(database)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/databases", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdDatabase models.Database
	err = json.NewDecoder(resp.Body).Decode(&createdDatabase)
	require.NoError(t, err)

	// Update database
	updateData := models.Database{
		Name:     "updated_test_db",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "updated_user",
		Password: "updated_password",
		Database: "updated_test_db",
	}

	jsonData, err = json.Marshal(updateData)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", baseURL+"/databases/"+fmt.Sprintf("%d", createdDatabase.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedDatabase models.Database
	err = json.NewDecoder(resp.Body).Decode(&updatedDatabase)
	require.NoError(t, err)

	assert.Equal(t, "updated_test_db", updatedDatabase.Name)
	assert.Equal(t, "mysql", updatedDatabase.Type)
	assert.Equal(t, 3306, updatedDatabase.Port)
}

func TestDatabaseServiceIntegration_DeleteDatabase(t *testing.T) {
	// Create database first
	database := models.Database{
		Name:     "delete_test_db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "delete_test_db",
	}

	jsonData, err := json.Marshal(database)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/databases", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdDatabase models.Database
	err = json.NewDecoder(resp.Body).Decode(&createdDatabase)
	require.NoError(t, err)

	// Delete database
	req, err = http.NewRequest("DELETE", baseURL+"/databases/"+fmt.Sprintf("%d", createdDatabase.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Database deleted successfully", response["message"])

	// Verify deletion
	req, err = http.NewRequest("GET", baseURL+"/databases/"+fmt.Sprintf("%d", createdDatabase.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDatabaseServiceIntegration_ListDatabases(t *testing.T) {
	// Create multiple databases
	databases := []models.Database{
		{Name: "list_test_db_1", Type: "postgresql", Host: "localhost", Port: 5432, User: "user1", Password: "pass1", Database: "db1"},
		{Name: "list_test_db_2", Type: "mysql", Host: "localhost", Port: 3306, User: "user2", Password: "pass2", Database: "db2"},
		{Name: "list_test_db_3", Type: "sqlite", Host: "localhost", Port: 0, User: "", Password: "", Database: "db3"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, database := range databases {
		jsonData, err := json.Marshal(database)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/databases", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List databases
	req, err := http.NewRequest("GET", baseURL+"/databases", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Databases []models.Database `json:"databases"`
		Total     int               `json:"total"`
		Limit     int               `json:"limit"`
		Offset    int               `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Databases), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestDatabaseServiceIntegration_ExecuteQuery(t *testing.T) {
	// Create a database first
	database := models.Database{
		Name:     "query_test_db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "query_test_db",
	}

	jsonData, err := json.Marshal(database)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("POST", baseURL+"/databases", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Execute a simple query
	queryRequest := map[string]interface{}{
		"database_id": 1,
		"query":       "SELECT 1 as test_column",
	}

	jsonData, err = json.Marshal(queryRequest)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/query", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Rows     []map[string]interface{} `json:"rows"`
		RowCount int                      `json:"row_count"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.RowCount)
	assert.Equal(t, 1, len(response.Rows))

	// Verify the query result
	assert.Equal(t, float64(1), response.Rows[0]["test_column"])
}

func TestDatabaseServiceIntegration_GetDatabaseStats(t *testing.T) {
	// Create a database first
	database := models.Database{
		Name:     "stats_test_db",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "stats_test_db",
	}

	jsonData, err := json.Marshal(database)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("POST", baseURL+"/databases", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Get database stats
	req, err = http.NewRequest("GET", baseURL+"/databases/1/stats", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		TableCount int   `json:"table_count"`
		RowCount   int64 `json:"row_count"`
		SizeBytes  int64 `json:"size_bytes"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, response.TableCount, 0)
	assert.GreaterOrEqual(t, response.RowCount, int64(0))
	assert.GreaterOrEqual(t, response.SizeBytes, int64(0))
}

// Helper functions
func startDatabaseService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "/Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/database-service"
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
