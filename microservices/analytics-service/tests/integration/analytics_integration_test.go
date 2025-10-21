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

	"github.com/anupamdutta5/analytics-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create uint pointer
func uintPtr(u uint) *uint {
	return &u
}

const (
	baseURL = "http://localhost:8088"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start analytics service for integration tests
	cmd := startAnalyticsService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Analytics service not ready, skipping integration tests")
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

func TestAnalyticsServiceIntegration_HealthCheck(t *testing.T) {
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
	assert.Equal(t, "analytics-service", response["service"])
}

func TestAnalyticsServiceIntegration_TrackAndGetEvents(t *testing.T) {
	// Track multiple events
	events := []models.AnalyticsEvent{
		{
			TenantID:  1,
			UserID:    uintPtr(1),
			EventType: "page_view",
			EventName: "status_page_viewed",
			Properties: `{
				"page_url": "/status",
				"referrer": "https://google.com"
			}`,
		},
		{
			TenantID:  1,
			UserID:    uintPtr(2),
			EventType: "click",
			EventName: "component_clicked",
			Properties: `{
				"component_id":     "api-server",
				"component_status": "operational"
			}`,
		},
		{
			TenantID:  1,
			UserID:    uintPtr(1),
			EventType: "form_submit",
			EventName: "contact_form_submitted",
			Properties: `{
				"form_type": "contact",
				"success":   true
			}`,
		},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// Track events
	for _, event := range events {
		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdEvent models.AnalyticsEvent
		err = json.NewDecoder(resp.Body).Decode(&createdEvent)
		require.NoError(t, err)

		assert.Equal(t, event.TenantID, createdEvent.TenantID)
		assert.Equal(t, event.UserID, createdEvent.UserID)
		assert.Equal(t, event.EventType, createdEvent.EventType)
		assert.Equal(t, event.EventName, createdEvent.EventName)
		assert.NotEmpty(t, createdEvent.ID)
	}

	// Get events
	req, err := http.NewRequest("GET", baseURL+"/events?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Events []models.AnalyticsEvent `json:"events"`
		Total  int                     `json:"total"`
		Limit  int                     `json:"limit"`
		Offset int                     `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Events), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestAnalyticsServiceIntegration_GetAnalytics(t *testing.T) {
	// Create some events first
	events := []models.AnalyticsEvent{
		{TenantID: 1, UserID: uintPtr(1), EventType: "page_view", EventName: "status_page_viewed"},
		{TenantID: 1, UserID: uintPtr(2), EventType: "page_view", EventName: "status_page_viewed"},
		{TenantID: 1, UserID: uintPtr(1), EventType: "click", EventName: "component_clicked"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get analytics
	req, err := http.NewRequest("GET", baseURL+"/analytics?tenant_id=integration-test-tenant&start_date=2024-01-01&end_date=2024-12-31", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_events")
	assert.Contains(t, response, "unique_users")
	assert.Contains(t, response, "event_breakdown")
}

func TestAnalyticsServiceIntegration_GetDashboard(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/dashboard?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "overview")
	assert.Contains(t, response, "charts")
	assert.Contains(t, response, "metrics")
}

func TestAnalyticsServiceIntegration_GetUserAnalytics(t *testing.T) {
	// Create events for a specific user
	events := []models.AnalyticsEvent{
		{TenantID: 1, UserID: uintPtr(1), EventType: "page_view", EventName: "status_page_viewed"},
		{TenantID: 1, UserID: uintPtr(1), EventType: "click", EventName: "component_clicked"},
		{TenantID: 1, UserID: uintPtr(1), EventType: "page_view", EventName: "incident_page_viewed"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get user analytics
	req, err := http.NewRequest("GET", baseURL+"/analytics/user/integration-test-user?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "user_id")
	assert.Contains(t, response, "total_events")
	assert.Contains(t, response, "event_breakdown")
	assert.Contains(t, response, "activity_timeline")
}

func TestAnalyticsServiceIntegration_GetEventTypes(t *testing.T) {
	// Create events with different types
	events := []models.AnalyticsEvent{
		{TenantID: 1, UserID: uintPtr(1), EventType: "page_view", EventName: "status_page_viewed"},
		{TenantID: 1, UserID: uintPtr(2), EventType: "click", EventName: "component_clicked"},
		{TenantID: 1, UserID: uintPtr(3), EventType: "form_submit", EventName: "contact_form_submitted"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get event types
	req, err := http.NewRequest("GET", baseURL+"/analytics/event-types?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		EventTypes []string `json:"event_types"`
		Total      int      `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 3, len(response.EventTypes))
	assert.Equal(t, 3, response.Total)

	// Verify all event types are present
	eventTypes := make(map[string]bool)
	for _, eventType := range response.EventTypes {
		eventTypes[eventType] = true
	}

	assert.True(t, eventTypes["page_view"])
	assert.True(t, eventTypes["click"])
	assert.True(t, eventTypes["form_submit"])
}

func TestAnalyticsServiceIntegration_GetTopPages(t *testing.T) {
	// Create events for different pages
	events := []models.AnalyticsEvent{
		{TenantID: 1, UserID: uintPtr(1), EventType: "page_view", EventName: "status_page_viewed", Properties: `{"page_url": "/status"}`},
		{TenantID: 1, UserID: uintPtr(2), EventType: "page_view", EventName: "status_page_viewed", Properties: `{"page_url": "/status"}`},
		{TenantID: 1, UserID: uintPtr(3), EventType: "page_view", EventName: "incident_page_viewed", Properties: `{"page_url": "/incidents"}`},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get top pages
	req, err := http.NewRequest("GET", baseURL+"/analytics/top-pages?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Pages []map[string]interface{} `json:"pages"`
		Total int                      `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Pages), 2)
	assert.GreaterOrEqual(t, response.Total, 2)
}

func TestAnalyticsServiceIntegration_GetRealTimeAnalytics(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/analytics/realtime?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "active_users")
	assert.Contains(t, response, "events_per_minute")
	assert.Contains(t, response, "top_events")
}

func TestAnalyticsServiceIntegration_ExportAnalytics(t *testing.T) {
	// Create some events first
	events := []models.AnalyticsEvent{
		{TenantID: 1, UserID: uintPtr(1), EventType: "page_view", EventName: "status_page_viewed"},
		{TenantID: 1, UserID: uintPtr(2), EventType: "click", EventName: "component_clicked"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Export analytics
	req, err := http.NewRequest("GET", baseURL+"/analytics/export?tenant_id=integration-test-tenant&format=csv", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
}

func TestAnalyticsServiceIntegration_PerformanceTest(t *testing.T) {
	// Test tracking multiple events in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Track events in bulk
	start := time.Now()

	for i := 0; i < 50; i++ {
		event := models.AnalyticsEvent{
			TenantID:   1,
			UserID:     uintPtr(uint(i + 1)),
			EventType:  "page_view",
			EventName:  "status_page_viewed",
			Properties: fmt.Sprintf(`{"page_url": "/page-%d"}`, i),
		}

		jsonData, err := json.Marshal(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/events", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Tracked 50 events in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 15*time.Second)
}

// Helper functions
func startAnalyticsService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8088", "DB_NAME=statuspage_analytics_test")

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
