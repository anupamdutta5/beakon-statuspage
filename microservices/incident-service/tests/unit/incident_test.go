package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anupamdutta5/incident-service/internal/handlers"
	"github.com/anupamdutta5/incident-service/internal/models"
	"github.com/anupamdutta5/incident-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIncidentHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/incidents", func(c *gin.Context) {
		// Set required context values for authentication
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		handler.CreateIncident(c)
	})

	incidentRequest := map[string]interface{}{
		"title":       "API Server Outage",
		"description": "API server is experiencing downtime",
		"status":      "investigating",
		"impact":      "major",
		"severity":    "major",
		"metadata":    `{"affected_users":1000,"region":"us-east-1"}`,
	}

	jsonData, _ := json.Marshal(incidentRequest)
	req, _ := http.NewRequest("POST", "/incidents", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		Message  string          `json:"message"`
		Incident models.Incident `json:"incident"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Incident created successfully", response.Message)
	assert.Equal(t, "API Server Outage", response.Incident.Title)
	assert.Equal(t, "API server is experiencing downtime", response.Incident.Description)
	assert.Equal(t, "investigating", response.Incident.Status)
	assert.Equal(t, "major", response.Incident.Severity)
	assert.NotZero(t, response.Incident.ID)
}

func TestIncidentHandler_GetIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/incidents/:id", handler.GetIncident)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/incidents/%d", incident.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Incident models.Incident `json:"incident"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, incident.ID, responseWrapper.Incident.ID)
	assert.Equal(t, incident.Title, responseWrapper.Incident.Title)
}

func TestIncidentHandler_UpdateIncident(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Message  string          `json:"message"`
		Incident models.Incident `json:"incident"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Updated Incident", responseWrapper.Incident.Title)
	assert.Equal(t, "identified", responseWrapper.Incident.Status)

	// Verify in database
	updatedIncident, err := incidentService.GetIncident(incident.ID)
	require.NoError(t, err)

	assert.Equal(t, "Updated Incident", updatedIncident.Title)
	assert.Equal(t, "identified", updatedIncident.Status)
}

func TestIncidentHandler_ListIncidents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Incident status updated successfully", responseWrapper.Message)
}

func TestIncidentHandler_AddIncidentUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Message string                `json:"message"`
		Update  models.IncidentUpdate `json:"update"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Incident update added successfully", responseWrapper.Message)
	assert.Equal(t, update.Message, responseWrapper.Update.Message)
	assert.Equal(t, update.Status, responseWrapper.Update.Status)
	assert.NotEmpty(t, responseWrapper.Update.ID)
}

func TestIncidentHandler_GetIncidentUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/incidents/:id", handler.GetIncident)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/incidents/%d", incident.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Incident models.Incident `json:"incident"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, incident.ID, responseWrapper.Incident.ID)
	assert.Equal(t, incident.Title, responseWrapper.Incident.Title)
}

func TestIncidentHandler_GetIncidentsByStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/incidents", handler.GetIncidents)

	// Test
	req, _ := http.NewRequest("GET", "/incidents", nil)
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

	assert.Equal(t, 3, len(response.Incidents))
	assert.Equal(t, 3, response.Total)

	// Verify all incidents are returned (order may vary)
	titles := make([]string, len(response.Incidents))
	for i, incident := range response.Incidents {
		titles[i] = incident.Title
	}
	assert.Contains(t, titles, "Incident 1")
	assert.Contains(t, titles, "Incident 2")
	assert.Contains(t, titles, "Incident 3")
}

func TestIncidentHandler_GetIncidentsBySeverity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	incidentService := services.NewIncidentService(db, logger)
	handler := handlers.NewIncidentHandler(incidentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/incidents", handler.GetIncidents)

	// Test
	req, _ := http.NewRequest("GET", "/incidents", nil)
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

	assert.Equal(t, 3, len(response.Incidents))
	assert.Equal(t, 3, response.Total)

	// Verify all incidents are returned (order may vary)
	titles := make([]string, len(response.Incidents))
	for i, incident := range response.Incidents {
		titles[i] = incident.Title
	}
	assert.Contains(t, titles, "Incident 1")
	assert.Contains(t, titles, "Incident 2")
	assert.Contains(t, titles, "Incident 3")
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.Incident{}, &models.IncidentUpdate{}, &models.IncidentComponent{})
	require.NoError(t, err)

	return db
}
