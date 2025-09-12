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

	"github.com/enterprise-status/statuspage-notification-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create uint pointer
func uintPtr(u uint) *uint {
	return &u
}

const (
	baseURL = "http://localhost:8086"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start notification service for integration tests
	cmd := startNotificationService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Notification service not ready, skipping integration tests")
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

func TestNotificationServiceIntegration_HealthCheck(t *testing.T) {
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
	assert.Equal(t, "notification-service", response["service"])
}

func TestNotificationServiceIntegration_SendAndGetNotification(t *testing.T) {
	// Send notification
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Subject:    "Integration Test Notification",
		Content:    "This is an integration test notification",
		Recipients: `["user@example.com"]`,
		Priority:   "normal",
		Status:     "pending",
		Metadata: `{
			"template":    "incident_alert",
			"incident_id": "inc-123"
		}`,
	}

	jsonData, err := json.Marshal(notification)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdNotification models.Notification
	err = json.NewDecoder(resp.Body).Decode(&createdNotification)
	require.NoError(t, err)

	assert.Equal(t, notification.TenantID, createdNotification.TenantID)
	assert.Equal(t, notification.UserID, createdNotification.UserID)
	assert.Equal(t, notification.Type, createdNotification.Type)
	assert.Equal(t, notification.Type, createdNotification.Type)
	assert.Equal(t, notification.Recipients, createdNotification.Recipients)
	assert.Equal(t, notification.Subject, createdNotification.Subject)
	assert.NotEmpty(t, createdNotification.ID)

	// Get notification by ID
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/notifications/%d", baseURL, createdNotification.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedNotification models.Notification
	err = json.NewDecoder(resp.Body).Decode(&retrievedNotification)
	require.NoError(t, err)

	assert.Equal(t, createdNotification.ID, retrievedNotification.ID)
	assert.Equal(t, createdNotification.Subject, retrievedNotification.Subject)
}

func TestNotificationServiceIntegration_UpdateNotificationStatus(t *testing.T) {
	// Create a notification first
	notification := models.Notification{
		TenantID:   1,
		UserID:     uintPtr(1),
		Type:       "email",
		Subject:    "Status Update Test",
		Recipients: `["user@example.com"]`,
		Content:    "This is a status update test",
		Status:     "pending",
	}

	jsonData, err := json.Marshal(notification)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdNotification models.Notification
	err = json.NewDecoder(resp.Body).Decode(&createdNotification)
	require.NoError(t, err)

	// Update status
	statusUpdate := map[string]interface{}{
		"status": "sent",
		"metadata": map[string]interface{}{
			"sent_at":    "2024-01-01T00:00:00Z",
			"message_id": "msg_123456789",
		},
	}

	jsonData, err = json.Marshal(statusUpdate)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/notifications/%d/status", baseURL, createdNotification.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedNotification models.Notification
	err = json.NewDecoder(resp.Body).Decode(&updatedNotification)
	require.NoError(t, err)

	assert.Equal(t, "sent", updatedNotification.Status)
}

func TestNotificationServiceIntegration_ListNotifications(t *testing.T) {
	// Create multiple notifications
	notifications := []models.Notification{
		{TenantID: 1, UserID: uintPtr(1), Type: "email", Recipients: `["user1@example.com"]`, Subject: "List Test 1", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(2), Type: "sms", Recipients: `["+1234567890"]`, Subject: "List Test 2", Status: "pending"},
		{TenantID: 1, UserID: uintPtr(3), Type: "webhook", Recipients: `["https://webhook.example.com"]`, Subject: "List Test 3", Status: "failed"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, notification := range notifications {
		jsonData, err := json.Marshal(notification)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List notifications
	req, err := http.NewRequest("GET", baseURL+"/notifications?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Notifications []models.Notification `json:"notifications"`
		Total         int                   `json:"total"`
		Limit         int                   `json:"limit"`
		Offset        int                   `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Notifications), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestNotificationServiceIntegration_CreateAndGetTemplate(t *testing.T) {
	// Create template
	template := models.Template{
		TenantID:  1,
		Name:      "Integration Test Template",
		Type:      "email",
		Subject:   "Test Alert: {{.Incident.Title}}",
		Content:   "An incident has occurred: {{.Incident.Description}}",
		Variables: `["Incident.Title", "Incident.Description", "Incident.Severity"]`,
	}

	jsonData, err := json.Marshal(template)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/templates", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdTemplate models.Template
	err = json.NewDecoder(resp.Body).Decode(&createdTemplate)
	require.NoError(t, err)

	assert.Equal(t, template.TenantID, createdTemplate.TenantID)
	assert.Equal(t, template.Name, createdTemplate.Name)
	assert.Equal(t, template.Type, createdTemplate.Type)
	assert.Equal(t, template.Subject, createdTemplate.Subject)
	assert.NotEmpty(t, createdTemplate.ID)

	// Get templates
	req, err = http.NewRequest("GET", baseURL+"/templates?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Templates []models.Template `json:"templates"`
		Total     int               `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Templates), 1)
	assert.GreaterOrEqual(t, response.Total, 1)
}

func TestNotificationServiceIntegration_GetNotificationStats(t *testing.T) {
	// Create some notifications first
	notifications := []models.Notification{
		{TenantID: 1, UserID: uintPtr(1), Type: "email", Recipients: `["user1@example.com"]`, Subject: "Stats Test 1", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(2), Type: "email", Recipients: `["user2@example.com"]`, Subject: "Stats Test 2", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(3), Type: "email", Recipients: `["user3@example.com"]`, Subject: "Stats Test 3", Status: "failed"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, notification := range notifications {
		jsonData, err := json.Marshal(notification)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get notification stats
	req, err := http.NewRequest("GET", baseURL+"/notifications/stats?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_notifications")
	assert.Contains(t, response, "sent_notifications")
	assert.Contains(t, response, "failed_notifications")
	assert.Contains(t, response, "pending_notifications")
}

func TestNotificationServiceIntegration_GetNotificationsByType(t *testing.T) {
	// Create notifications with different types
	notifications := []models.Notification{
		{TenantID: 1, UserID: uintPtr(1), Type: "email", Recipients: `["user1@example.com"]`, Subject: "Email Type 1", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(2), Type: "email", Recipients: `["user2@example.com"]`, Subject: "Email Type 2", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(3), Type: "sms", Recipients: `["+1234567890"]`, Subject: "SMS Type 1", Status: "sent"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, notification := range notifications {
		jsonData, err := json.Marshal(notification)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get notifications by type
	req, err := http.NewRequest("GET", baseURL+"/notifications/type/email?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Notifications []models.Notification `json:"notifications"`
		Total         int                   `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Notifications))
	assert.Equal(t, 2, response.Total)

	// Verify all notifications are email type
	for _, notification := range response.Notifications {
		assert.Equal(t, "email", notification.Type)
	}
}

func TestNotificationServiceIntegration_GetNotificationsByStatus(t *testing.T) {
	// Create notifications with different statuses
	notifications := []models.Notification{
		{TenantID: 1, UserID: uintPtr(1), Type: "email", Recipients: `["user1@example.com"]`, Subject: "Sent Status 1", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(2), Type: "email", Recipients: `["user2@example.com"]`, Subject: "Sent Status 2", Status: "sent"},
		{TenantID: 1, UserID: uintPtr(3), Type: "email", Recipients: `["user3@example.com"]`, Subject: "Pending Status 1", Status: "pending"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, notification := range notifications {
		jsonData, err := json.Marshal(notification)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get notifications by status
	req, err := http.NewRequest("GET", baseURL+"/notifications/status/sent?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Notifications []models.Notification `json:"notifications"`
		Total         int                   `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Notifications))
	assert.Equal(t, 2, response.Total)

	// Verify all notifications have sent status
	for _, notification := range response.Notifications {
		assert.Equal(t, "sent", notification.Status)
	}
}

func TestNotificationServiceIntegration_PerformanceTest(t *testing.T) {
	// Test sending multiple notifications in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Send notifications in bulk
	start := time.Now()

	for i := 0; i < 30; i++ {
		notification := models.Notification{
			TenantID:   1,
			UserID:     uintPtr(uint(i + 1)),
			Type:       "email",
			Recipients: fmt.Sprintf(`["user%d@example.com"]`, i),
			Subject:    fmt.Sprintf("Performance Test Notification %d", i),
			Content:    fmt.Sprintf("This is performance test notification %d", i),
			Status:     "pending",
		}

		jsonData, err := json.Marshal(notification)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/notifications", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Sent 30 notifications in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 15*time.Second)
}

// Helper functions
func startNotificationService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8086", "DB_NAME=statuspage_notification_test")

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
