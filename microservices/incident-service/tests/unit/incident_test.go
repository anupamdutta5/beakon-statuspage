package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enterprise-status/statuspage-incident-service/internal/handlers"
	"github.com/enterprise-status/statuspage-incident-service/internal/models"
	"github.com/enterprise-status/statuspage-incident-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIncidentHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	router := gin.New()
	router.GET("/health", handler.Health)

	// Test
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "incident-service", response["service"])
}

func TestIncidentHandler_CreateIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	router := gin.New()
	router.POST("/incidents", handler.CreateIncident)

	incident := models.Incident{
		Title:       "API Server Outage",
		Description: "API server is experiencing downtime",
		Status:      "investigating",
		Severity:    "major",
		TenantID:    1,
		Metadata:    `{"affected_users":1000,"region":"us-east-1"}`,
	}

	jsonData, _ := json.Marshal(incident)
	req, _ := http.NewRequest("POST", "/incidents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Incident
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, incident.Title, response.Title)
	assert.Equal(t, incident.Description, response.Description)
	assert.Equal(t, incident.Status, response.Status)
	assert.Equal(t, incident.Severity, response.Severity)
	assert.NotEmpty(t, response.ID)
}

func TestIncidentHandler_GetIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create an incident first
	incident := models.Incident{
		Title:       "Test Incident",
		Description: "Test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}
	err := incidentService.CreateIncident(&incident)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/incidents/:id", handler.GetIncident)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/incidents/%d", incident.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Incident
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, incident.ID, response.ID)
	assert.Equal(t, incident.Title, response.Title)
}

func TestIncidentHandler_UpdateIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create an incident first
	incident := models.Incident{
		Title:       "Test Incident",
		Description: "Test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}
	err := incidentService.CreateIncident(&incident)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/incidents/:id", handler.UpdateIncident)

	// Update data
	updateData := models.Incident{
		Title:       "Updated Incident",
		Description: "Updated incident description",
		Status:      "identified",
		Severity:    "major",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/incidents/%d", incident.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Incident
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Incident", response.Title)
	assert.Equal(t, "identified", response.Status)

	// Verify in database
	updatedIncident, err := incidentService.GetIncident(incident.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Incident", updatedIncident.Title)
	assert.Equal(t, "identified", updatedIncident.Status)
}

func TestIncidentHandler_ListIncidents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create multiple incidents
	incidents := []models.Incident{
		{Title: "Incident 1", Description: "Description 1", Status: "investigating", Severity: "minor", TenantID: 1},
		{Title: "Incident 2", Description: "Description 2", Status: "identified", Severity: "major", TenantID: 1},
		{Title: "Incident 3", Description: "Description 3", Status: "monitoring", Severity: "critical", TenantID: 1},
	}

	for _, incident := range incidents {
		err := incidentService.CreateIncident(&incident)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/incidents", handler.GetIncidents)

	// Test
	req, _ := http.NewRequest("GET", "/incidents?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Incidents []models.Incident `json:"incidents"`
		Total     int               `json:"total"`
		Limit     int               `json:"limit"`
		Offset    int               `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Incidents), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestIncidentHandler_DeleteIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create an incident first
	incident := models.Incident{
		Title:       "Test Incident",
		Description: "Test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}
	err := incidentService.CreateIncident(&incident)
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/incidents/:id", handler.DeleteIncident)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/incidents/%d", incident.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Incident deleted successfully", response["message"])

	// Verify deletion
	_, err = incidentService.GetIncident(incident.ID)
	assert.Error(t, err)
}

func TestIncidentHandler_UpdateIncidentStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create an incident first
	incident := models.Incident{
		Title:       "Test Incident",
		Description: "Test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}
	err := incidentService.CreateIncident(&incident)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/incidents/:id/status", handler.UpdateIncidentStatus)

	// Update status
	statusUpdate := map[string]interface{}{
		"status":  "resolved",
		"message": "Issue has been resolved",
	}

	jsonData, _ := json.Marshal(statusUpdate)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/incidents/%d", incident.ID)+"/status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Incident
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "resolved", response.Status)
}

func TestIncidentHandler_AddIncidentUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create an incident first
	incident := models.Incident{
		Title:       "Test Incident",
		Description: "Test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}
	err := incidentService.CreateIncident(&incident)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/incidents/:id/updates", handler.AddIncidentUpdate)

	// Add update
	update := models.IncidentUpdate{
		Message: "We are investigating the issue",
		Status:  "investigating",
	}

	jsonData, _ := json.Marshal(update)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/incidents/%d", incident.ID)+"/updates", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.IncidentUpdate
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, update.Message, response.Message)
	assert.Equal(t, update.Status, response.Status)
	assert.NotEmpty(t, response.ID)
}

func TestIncidentHandler_GetIncidentUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create an incident first
	incident := models.Incident{
		Title:       "Test Incident",
		Description: "Test incident description",
		Status:      "investigating",
		Severity:    "minor",
		TenantID:    1,
	}
	err := incidentService.CreateIncident(&incident)
	require.NoError(t, err)

	// Add some updates
	updates := []models.IncidentUpdate{
		{Message: "Update 1", Status: "investigating"},
		{Message: "Update 2", Status: "identified"},
		{Message: "Update 3", Status: "resolved"},
	}

	for _, update := range updates {
		err := incidentService.AddIncidentUpdate(&update)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/incidents/:id/updates", handler.GetIncident)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/incidents/%d", incident.ID)+"/updates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Updates []models.IncidentUpdate `json:"updates"`
		Total   int                     `json:"total"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 3, len(response.Updates))
	assert.Equal(t, 3, response.Total)
}

func TestIncidentHandler_GetIncidentsByStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create incidents with different statuses
	incidents := []models.Incident{
		{Title: "Incident 1", Status: "investigating", Severity: "minor", TenantID: 1},
		{Title: "Incident 2", Status: "investigating", Severity: "major", TenantID: 1},
		{Title: "Incident 3", Status: "resolved", Severity: "critical", TenantID: 1},
	}

	for _, incident := range incidents {
		err := incidentService.CreateIncident(&incident)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/incidents/status/:status", handler.GetIncidents)

	// Test
	req, _ := http.NewRequest("GET", "/incidents/status/investigating?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Incidents []models.Incident `json:"incidents"`
		Total     int               `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Incidents))
	assert.Equal(t, 2, response.Total)

	// Verify all incidents have investigating status
	for _, incident := range response.Incidents {
		assert.Equal(t, "investigating", incident.Status)
	}
}

func TestIncidentHandler_GetIncidentsBySeverity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, nil)
	handler := handlers.NewIncidentHandler(incidentService, nil)

	// Create incidents with different severities
	incidents := []models.Incident{
		{Title: "Incident 1", Status: "investigating", Severity: "minor", TenantID: 1},
		{Title: "Incident 2", Status: "investigating", Severity: "major", TenantID: 1},
		{Title: "Incident 3", Status: "investigating", Severity: "major", TenantID: 1},
	}

	for _, incident := range incidents {
		err := incidentService.CreateIncident(&incident)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/incidents/severity/:severity", handler.GetIncidents)

	// Test
	req, _ := http.NewRequest("GET", "/incidents/severity/major?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Incidents []models.Incident `json:"incidents"`
		Total     int               `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Incidents))
	assert.Equal(t, 2, response.Total)

	// Verify all incidents have major severity
	for _, incident := range response.Incidents {
		assert.Equal(t, "major", incident.Severity)
	}
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.Incident{}, &models.IncidentUpdate{})
	require.NoError(t, err)

	return db
}
