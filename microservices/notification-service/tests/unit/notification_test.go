// Package unit provides unit tests for the Notification Service.
package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anupamdutta5/statuspage-notification-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-notification-service/internal/models"
	"github.com/anupamdutta5/statuspage-notification-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotificationHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

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
	assert.Equal(t, "notification-service", response["service"])
}

func TestNotificationHandler_CreateNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/notifications", handler.CreateNotification)

	// Test data
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Status:     "pending",
		Priority:   "normal",
		Subject:    "Test Notification",
		Content:    "This is a test notification",
		Recipients: `["user@example.com"]`,
		Metadata:   `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(notification)
	req, _ := http.NewRequest("POST", "/notifications", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Message      string              `json:"message"`
		Notification models.Notification `json:"notification"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Notification created successfully", responseWrapper.Message)
	assert.Equal(t, notification.Type, responseWrapper.Notification.Type)
	assert.Equal(t, notification.Status, responseWrapper.Notification.Status)
	assert.Equal(t, notification.Priority, responseWrapper.Notification.Priority)
	assert.Equal(t, notification.Subject, responseWrapper.Notification.Subject)
	assert.Equal(t, notification.Content, responseWrapper.Notification.Content)
	assert.Equal(t, notification.Recipients, responseWrapper.Notification.Recipients)
	assert.Equal(t, notification.TenantID, responseWrapper.Notification.TenantID)
}

func TestNotificationHandler_GetNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create a notification first
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Status:     "pending",
		Priority:   "normal",
		Subject:    "Test Notification",
		Content:    "This is a test notification",
		Recipients: `["user@example.com"]`,
		Metadata:   `{"source": "test"}`,
	}
	err := notificationService.CreateNotification(&notification)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/notifications/:id", handler.GetNotification)

	// Test
	req, _ := http.NewRequest("GET", fmt.Sprintf("/notifications/%d", notification.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Notification models.Notification `json:"notification"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, notification.Type, responseWrapper.Notification.Type)
	assert.Equal(t, notification.Status, responseWrapper.Notification.Status)
	assert.Equal(t, notification.Priority, responseWrapper.Notification.Priority)
	assert.Equal(t, notification.Subject, responseWrapper.Notification.Subject)
	assert.Equal(t, notification.Content, responseWrapper.Notification.Content)
}

func TestNotificationHandler_UpdateNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create a notification first
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Status:     "pending",
		Priority:   "normal",
		Subject:    "Original Subject",
		Content:    "Original content",
		Recipients: `["user@example.com"]`,
		Metadata:   `{"source": "test"}`,
	}
	err := notificationService.CreateNotification(&notification)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/notifications/:id", handler.UpdateNotification)

	// Update data
	updateData := models.Notification{
		Status:   "sent",
		Priority: "high",
		Subject:  "Updated Subject",
		Content:  "Updated content",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/notifications/%d", notification.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Notification models.Notification `json:"notification"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "sent", responseWrapper.Notification.Status)
	assert.Equal(t, "high", responseWrapper.Notification.Priority)
	assert.Equal(t, "Updated Subject", responseWrapper.Notification.Subject)
	assert.Equal(t, "Updated content", responseWrapper.Notification.Content)
}

func TestNotificationHandler_DeleteNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create a notification first
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Status:     "pending",
		Priority:   "normal",
		Subject:    "Test Notification",
		Content:    "This is a test notification",
		Recipients: `["user@example.com"]`,
		Metadata:   `{"source": "test"}`,
	}
	err := notificationService.CreateNotification(&notification)
	require.NoError(t, err)

	router := gin.New()
	router.DELETE("/notifications/:id", handler.DeleteNotification)

	// Test
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/notifications/%d", notification.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	_, err = notificationService.GetNotification(notification.ID)
	assert.Error(t, err)
}

func TestNotificationHandler_GetNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create multiple notifications
	notifications := []models.Notification{
		{TenantID: 1, UserID: uintPtr(1), Type: "email", Status: "pending", Subject: "Notification 1", Content: "Content 1", Recipients: `["user1@example.com"]`},
		{TenantID: 1, UserID: uintPtr(2), Type: "sms", Status: "sent", Subject: "Notification 2", Content: "Content 2", Recipients: `["user2@example.com"]`},
		{TenantID: 1, UserID: uintPtr(3), Type: "push", Status: "failed", Subject: "Notification 3", Content: "Content 3", Recipients: `["user3@example.com"]`},
	}

	for _, notification := range notifications {
		err := notificationService.CreateNotification(&notification)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/notifications", handler.GetNotifications)

	// Test
	req, _ := http.NewRequest("GET", "/notifications?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	notificationsList, ok := response["notifications"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(notificationsList))
}

func TestNotificationHandler_SendNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create a notification first
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Status:     "pending",
		Priority:   "normal",
		Subject:    "Test Notification",
		Content:    "This is a test notification",
		Recipients: `["user@example.com"]`,
		Metadata:   `{"source": "test"}`,
	}
	err := notificationService.CreateNotification(&notification)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/notifications/:id/send", handler.SendNotification)

	// Test
	req, _ := http.NewRequest("POST", fmt.Sprintf("/notifications/%d/send", notification.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Notification sent successfully", responseWrapper.Message)
}

func TestNotificationHandler_CreateTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/templates", handler.CreateTemplate)

	// Test data
	template := models.Template{
		TenantID:    1,
		Name:        "Test Template",
		Description: "Test template description",
		Type:        "email",
		Category:    "incident",
		Subject:     "Test Subject",
		Content:     "Test content with {{.variable}}",
		Variables:   `["variable"]`,
		IsActive:    true,
		IsDefault:   false,
		Metadata:    `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(template)
	req, _ := http.NewRequest("POST", "/templates", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Template models.Template `json:"template"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, template.Name, responseWrapper.Template.Name)
	assert.Equal(t, template.Description, responseWrapper.Template.Description)
	assert.Equal(t, template.Type, responseWrapper.Template.Type)
	assert.Equal(t, template.Category, responseWrapper.Template.Category)
	assert.Equal(t, template.Subject, responseWrapper.Template.Subject)
	assert.Equal(t, template.Content, responseWrapper.Template.Content)
	assert.Equal(t, template.TenantID, responseWrapper.Template.TenantID)
}

func TestNotificationHandler_GetTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create multiple templates
	templates := []models.Template{
		{TenantID: 1, Name: "Template 1", Type: "email", Category: "incident", Subject: "Subject 1", Content: "Content 1"},
		{TenantID: 1, Name: "Template 2", Type: "sms", Category: "maintenance", Subject: "Subject 2", Content: "Content 2"},
		{TenantID: 1, Name: "Template 3", Type: "push", Category: "general", Subject: "Subject 3", Content: "Content 3"},
	}

	for _, template := range templates {
		err := notificationService.CreateTemplate(&template)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/templates", handler.GetTemplates)

	// Test
	req, _ := http.NewRequest("GET", "/templates?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	templatesList, ok := response["templates"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(templatesList))
}

func TestNotificationHandler_CreateChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/channels", handler.CreateChannel)

	// Test data
	channel := models.Channel{
		TenantID:    1,
		Name:        "Test Channel",
		Description: "Test channel description",
		Type:        "email",
		Provider:    "smtp",
		Config:      `{"host": "smtp.example.com", "port": 587}`,
		IsActive:    true,
		IsDefault:   false,
		Priority:    1,
		Metadata:    `{"source": "test"}`,
	}

	jsonData, _ := json.Marshal(channel)
	req, _ := http.NewRequest("POST", "/channels", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseWrapper struct {
		Channel models.Channel `json:"channel"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, channel.Name, responseWrapper.Channel.Name)
	assert.Equal(t, channel.Description, responseWrapper.Channel.Description)
	assert.Equal(t, channel.Type, responseWrapper.Channel.Type)
	assert.Equal(t, channel.Provider, responseWrapper.Channel.Provider)
	assert.Equal(t, channel.Config, responseWrapper.Channel.Config)
	assert.Equal(t, channel.TenantID, responseWrapper.Channel.TenantID)
}

func TestNotificationHandler_GetChannels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	notificationService := services.NewNotificationService(db, logger)
	handler := handlers.NewNotificationHandler(notificationService, logger)

	// Create multiple channels
	channels := []models.Channel{
		{TenantID: 1, Name: "Channel 1", Type: "email", Provider: "smtp", Config: `{"host": "smtp1.com"}`},
		{TenantID: 1, Name: "Channel 2", Type: "sms", Provider: "twilio", Config: `{"account_sid": "test"}`},
		{TenantID: 1, Name: "Channel 3", Type: "push", Provider: "fcm", Config: `{"api_key": "test"}`},
	}

	for _, channel := range channels {
		err := notificationService.CreateChannel(&channel)
		require.NoError(t, err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/channels", handler.GetChannels)

	// Test
	req, _ := http.NewRequest("GET", "/channels?tenant_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	channelsList, ok := response["channels"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(channelsList))
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Notification{},
		&models.Template{},
		&models.Channel{},
		&models.Subscription{},
		&models.Delivery{},
		&models.WebhookEvent{},
		&models.NotificationLog{},
	)
	require.NoError(t, err)

	return db
}

// Helper function to create uint pointer
func uintPtr(u uint) *uint {
	return &u
}
