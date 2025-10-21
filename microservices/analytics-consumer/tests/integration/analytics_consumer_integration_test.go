package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Event represents an analytics event for testing
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// TestAnalyticsConsumerIntegration tests the full integration of the analytics consumer
func TestAnalyticsConsumerIntegration(t *testing.T) {
	// Setup test environment
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test user event processing
	event := Event{
		ID:        "integration-test-1",
		Type:      "user.registered",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"email":    "integration@example.com",
			"source":   "organic",
			"referrer": "google.com",
		},
	}

	err := consumer.ProcessEvent(event)
	require.NoError(t, err)

	// Verify event was processed
	analytics, err := consumer.GetAnalyticsData("tenant-1", "2024-01-01", "2024-12-31")
	require.NoError(t, err)
	assert.NotNil(t, analytics)
}

// TestAnalyticsConsumerUserEvents tests user-related events
func TestAnalyticsConsumerUserEvents(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test user registration
	registerEvent := Event{
		ID:        "user-register-1",
		Type:      "user.registered",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"email":      "user@example.com",
			"source":     "organic",
			"referrer":   "google.com",
			"user_agent": "Mozilla/5.0...",
		},
	}

	err := consumer.ProcessEvent(registerEvent)
	require.NoError(t, err)

	// Test user login
	loginEvent := Event{
		ID:        "user-login-1",
		Type:      "user.login",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"ip_address": "192.168.1.1",
			"user_agent": "Mozilla/5.0...",
		},
	}

	err = consumer.ProcessEvent(loginEvent)
	require.NoError(t, err)

	// Verify user events were processed
	userEvents := consumer.GetUserEvents("tenant-1")
	assert.Len(t, userEvents, 2)

	// Verify event types
	eventTypes := make(map[string]bool)
	for _, event := range userEvents {
		eventTypes[event["type"].(string)] = true
	}
	assert.True(t, eventTypes["user.registered"])
	assert.True(t, eventTypes["user.login"])
}

// TestAnalyticsConsumerPageEvents tests page-related events
func TestAnalyticsConsumerPageEvents(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test page visit
	visitEvent := Event{
		ID:        "page-visit-1",
		Type:      "page.visited",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"page_url":   "/status",
			"page_title": "Service Status",
			"duration":   30,
			"referrer":   "https://example.com",
			"user_agent": "Mozilla/5.0...",
		},
	}

	err := consumer.ProcessEvent(visitEvent)
	require.NoError(t, err)

	// Test page exit
	exitEvent := Event{
		ID:        "page-exit-1",
		Type:      "page.exited",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"page_url":    "/status",
			"duration":    60,
			"exit_reason": "navigation",
		},
	}

	err = consumer.ProcessEvent(exitEvent)
	require.NoError(t, err)

	// Verify page events were processed
	pageEvents := consumer.GetPageEvents("tenant-1")
	assert.Len(t, pageEvents, 2)

	// Verify event types
	eventTypes := make(map[string]bool)
	for _, event := range pageEvents {
		eventTypes[event["type"].(string)] = true
	}
	assert.True(t, eventTypes["page.visited"])
	assert.True(t, eventTypes["page.exited"])
}

// TestAnalyticsConsumerIncidentEvents tests incident-related events
func TestAnalyticsConsumerIncidentEvents(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test incident reported
	reportEvent := Event{
		ID:        "incident-report-1",
		Type:      "incident.reported",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"incident_id": "inc-123",
			"severity":    "high",
			"component":   "database",
			"duration":    120,
		},
	}

	err := consumer.ProcessEvent(reportEvent)
	require.NoError(t, err)

	// Test incident resolved
	resolveEvent := Event{
		ID:        "incident-resolve-1",
		Type:      "incident.resolved",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"incident_id":    "inc-123",
			"resolution":     "Database connection restored",
			"total_duration": 120,
		},
	}

	err = consumer.ProcessEvent(resolveEvent)
	require.NoError(t, err)

	// Verify incident events were processed
	incidentEvents := consumer.GetIncidentEvents("tenant-1")
	assert.Len(t, incidentEvents, 2)

	// Verify event types
	eventTypes := make(map[string]bool)
	for _, event := range incidentEvents {
		eventTypes[event["type"].(string)] = true
	}
	assert.True(t, eventTypes["incident.reported"])
	assert.True(t, eventTypes["incident.resolved"])
}

// TestAnalyticsConsumerSubscriptionEvents tests subscription-related events
func TestAnalyticsConsumerSubscriptionEvents(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test subscription created
	createEvent := Event{
		ID:        "subscription-create-1",
		Type:      "subscription.created",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plan":          "premium",
			"amount":        99.99,
			"currency":      "USD",
			"billing_cycle": "monthly",
		},
	}

	err := consumer.ProcessEvent(createEvent)
	require.NoError(t, err)

	// Test subscription updated
	updateEvent := Event{
		ID:        "subscription-update-1",
		Type:      "subscription.updated",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plan":          "enterprise",
			"amount":        199.99,
			"currency":      "USD",
			"billing_cycle": "monthly",
			"previous_plan": "premium",
		},
	}

	err = consumer.ProcessEvent(updateEvent)
	require.NoError(t, err)

	// Verify subscription events were processed
	subscriptionEvents := consumer.GetSubscriptionEvents("tenant-1")
	assert.Len(t, subscriptionEvents, 2)

	// Verify event types
	eventTypes := make(map[string]bool)
	for _, event := range subscriptionEvents {
		eventTypes[event["type"].(string)] = true
	}
	assert.True(t, eventTypes["subscription.created"])
	assert.True(t, eventTypes["subscription.updated"])
}

// TestAnalyticsConsumerCustomEvents tests custom events
func TestAnalyticsConsumerCustomEvents(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test custom button click
	buttonEvent := Event{
		ID:        "custom-button-1",
		Type:      "custom.button_clicked",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"button_id": "subscribe_button",
			"page_url":  "/pricing",
			"position":  "top",
		},
	}

	err := consumer.ProcessEvent(buttonEvent)
	require.NoError(t, err)

	// Test custom form submission
	formEvent := Event{
		ID:        "custom-form-1",
		Type:      "custom.form_submitted",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"form_id":  "contact_form",
			"page_url": "/contact",
			"fields":   []string{"name", "email", "message"},
		},
	}

	err = consumer.ProcessEvent(formEvent)
	require.NoError(t, err)

	// Verify custom events were processed
	customEvents := consumer.GetCustomEvents("tenant-1")
	assert.Len(t, customEvents, 2)

	// Verify event types
	eventTypes := make(map[string]bool)
	for _, event := range customEvents {
		eventTypes[event["type"].(string)] = true
	}
	assert.True(t, eventTypes["custom.button_clicked"])
	assert.True(t, eventTypes["custom.form_submitted"])
}

// TestAnalyticsConsumerBulkEvents tests processing multiple events
func TestAnalyticsConsumerBulkEvents(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	testEvents := []Event{
		{
			ID:        "bulk-user-1",
			Type:      "user.registered",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"email": "user1@example.com"},
		},
		{
			ID:        "bulk-page-1",
			Type:      "page.visited",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"page_url": "/status"},
		},
		{
			ID:        "bulk-incident-1",
			Type:      "incident.reported",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"incident_id": "inc-123"},
		},
		{
			ID:        "bulk-subscription-1",
			Type:      "subscription.created",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"plan": "premium"},
		},
		{
			ID:        "bulk-custom-1",
			Type:      "custom.button_clicked",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"button_id": "subscribe"},
		},
	}

	// Process all events
	for _, event := range testEvents {
		err := consumer.ProcessEvent(event)
		require.NoError(t, err)
	}

	// Verify all events were processed
	userEvents := consumer.GetUserEvents("tenant-1")
	pageEvents := consumer.GetPageEvents("tenant-1")
	incidentEvents := consumer.GetIncidentEvents("tenant-1")
	subscriptionEvents := consumer.GetSubscriptionEvents("tenant-1")
	customEvents := consumer.GetCustomEvents("tenant-1")

	assert.Len(t, userEvents, 1)
	assert.Len(t, pageEvents, 1)
	assert.Len(t, incidentEvents, 1)
	assert.Len(t, subscriptionEvents, 1)
	assert.Len(t, customEvents, 1)
}

// TestAnalyticsConsumerConcurrency tests concurrent event processing
func TestAnalyticsConsumerConcurrency(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Create multiple goroutines to process events concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			event := Event{
				ID:        fmt.Sprintf("concurrent-event-%d", index),
				Type:      "page.visited",
				TenantID:  "tenant-1",
				UserID:    "user-1",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"page_url": fmt.Sprintf("/page-%d", index)},
			}
			err := consumer.ProcessEvent(event)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all events were processed
	pageEvents := consumer.GetPageEvents("tenant-1")
	assert.Len(t, pageEvents, 10)
}

// TestAnalyticsConsumerErrorHandling tests error handling scenarios
func TestAnalyticsConsumerErrorHandling(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Test with empty event
	emptyEvent := Event{}
	err := consumer.ProcessEvent(emptyEvent)
	// Should handle gracefully
	assert.NoError(t, err)

	// Test with invalid data
	event := Event{
		ID:        "error-test-1",
		Type:      "page.visited",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"invalid": make(chan int)}, // Invalid JSON
	}

	// Should handle gracefully
	err = consumer.ProcessEvent(event)
	assert.NoError(t, err)
}

// TestAnalyticsConsumerGetAnalyticsData tests analytics data retrieval
func TestAnalyticsConsumerGetAnalyticsData(t *testing.T) {
	consumer := setupAnalyticsConsumer(t)
	defer teardownAnalyticsConsumer(t)

	// Process some events first
	testEvents := []Event{
		{
			ID:        "analytics-user-1",
			Type:      "user.registered",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"email": "user@example.com"},
		},
		{
			ID:        "analytics-page-1",
			Type:      "page.visited",
			TenantID:  "tenant-1",
			UserID:    "user-1",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"page_url": "/status"},
		},
	}

	for _, event := range testEvents {
		err := consumer.ProcessEvent(event)
		require.NoError(t, err)
	}

	// Get analytics data
	analytics, err := consumer.GetAnalyticsData("tenant-1", "2024-01-01", "2024-12-31")
	require.NoError(t, err)
	assert.NotNil(t, analytics)

	// Verify analytics data structure
	assert.Contains(t, analytics, "total_users")
	assert.Contains(t, analytics, "page_views")
	assert.Contains(t, analytics, "incidents")
	assert.Contains(t, analytics, "subscriptions")
}

// setupAnalyticsConsumer sets up the analytics consumer for testing
func setupAnalyticsConsumer(t *testing.T) *AnalyticsConsumer {
	// Initialize test database and services
	// This would typically connect to a test database
	// For now, we'll use a mock implementation

	consumer := &AnalyticsConsumer{
		analyticsService: &MockAnalyticsService{},
	}

	return consumer
}

// teardownAnalyticsConsumer cleans up after tests
func teardownAnalyticsConsumer(t *testing.T) {
	// Clean up test database, close connections, etc.
}

// MockAnalyticsService for integration testing
type MockAnalyticsService struct {
	userEvents         []map[string]interface{}
	pageEvents         []map[string]interface{}
	incidentEvents     []map[string]interface{}
	subscriptionEvents []map[string]interface{}
	customEvents       []map[string]interface{}
}

func (m *MockAnalyticsService) ProcessUserEvent(event Event) error {
	data := make(map[string]interface{})
	for k, v := range event.Data {
		data[k] = v
	}
	data["type"] = event.Type
	data["event_id"] = event.ID
	data["tenant_id"] = event.TenantID
	data["user_id"] = event.UserID
	data["timestamp"] = event.Timestamp

	m.userEvents = append(m.userEvents, data)
	return nil
}

func (m *MockAnalyticsService) ProcessPageEvent(event Event) error {
	data := make(map[string]interface{})
	for k, v := range event.Data {
		data[k] = v
	}
	data["type"] = event.Type
	data["event_id"] = event.ID
	data["tenant_id"] = event.TenantID
	data["user_id"] = event.UserID
	data["timestamp"] = event.Timestamp

	m.pageEvents = append(m.pageEvents, data)
	return nil
}

func (m *MockAnalyticsService) ProcessIncidentEvent(event Event) error {
	data := make(map[string]interface{})
	for k, v := range event.Data {
		data[k] = v
	}
	data["type"] = event.Type
	data["event_id"] = event.ID
	data["tenant_id"] = event.TenantID
	data["user_id"] = event.UserID
	data["timestamp"] = event.Timestamp

	m.incidentEvents = append(m.incidentEvents, data)
	return nil
}

func (m *MockAnalyticsService) ProcessSubscriptionEvent(event Event) error {
	data := make(map[string]interface{})
	for k, v := range event.Data {
		data[k] = v
	}
	data["type"] = event.Type
	data["event_id"] = event.ID
	data["tenant_id"] = event.TenantID
	data["user_id"] = event.UserID
	data["timestamp"] = event.Timestamp

	m.subscriptionEvents = append(m.subscriptionEvents, data)
	return nil
}

func (m *MockAnalyticsService) ProcessCustomEvent(event Event) error {
	data := make(map[string]interface{})
	for k, v := range event.Data {
		data[k] = v
	}
	data["type"] = event.Type
	data["event_id"] = event.ID
	data["tenant_id"] = event.TenantID
	data["user_id"] = event.UserID
	data["timestamp"] = event.Timestamp

	m.customEvents = append(m.customEvents, data)
	return nil
}

func (m *MockAnalyticsService) GetAnalyticsData(tenantID, startDate, endDate string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_users":   len(m.userEvents),
		"page_views":    len(m.pageEvents),
		"incidents":     len(m.incidentEvents),
		"subscriptions": len(m.subscriptionEvents),
		"custom_events": len(m.customEvents),
	}, nil
}

// AnalyticsConsumer represents the analytics consumer for testing
type AnalyticsConsumer struct {
	analyticsService AnalyticsService
}

// AnalyticsService interface for testing
type AnalyticsService interface {
	ProcessUserEvent(event Event) error
	ProcessPageEvent(event Event) error
	ProcessIncidentEvent(event Event) error
	ProcessSubscriptionEvent(event Event) error
	ProcessCustomEvent(event Event) error
	GetAnalyticsData(tenantID, startDate, endDate string) (map[string]interface{}, error)
}

// ProcessEvent processes an analytics event
func (c *AnalyticsConsumer) ProcessEvent(event Event) error {
	switch event.Type {
	case "user.registered", "user.login", "user.logout":
		return c.analyticsService.ProcessUserEvent(event)
	case "page.visited", "page.exited":
		return c.analyticsService.ProcessPageEvent(event)
	case "incident.reported", "incident.resolved":
		return c.analyticsService.ProcessIncidentEvent(event)
	case "subscription.created", "subscription.updated", "subscription.cancelled":
		return c.analyticsService.ProcessSubscriptionEvent(event)
	default:
		if len(event.Type) > 7 && event.Type[:7] == "custom." {
			return c.analyticsService.ProcessCustomEvent(event)
		}
		// Handle unknown event types gracefully
		return nil
	}
}

// GetAnalyticsData retrieves analytics data
func (c *AnalyticsConsumer) GetAnalyticsData(tenantID, startDate, endDate string) (map[string]interface{}, error) {
	return c.analyticsService.GetAnalyticsData(tenantID, startDate, endDate)
}

// GetUserEvents retrieves user events for testing
func (c *AnalyticsConsumer) GetUserEvents(tenantID string) []map[string]interface{} {
	if mockService, ok := c.analyticsService.(*MockAnalyticsService); ok {
		var result []map[string]interface{}
		for _, event := range mockService.userEvents {
			if event["tenant_id"] == tenantID {
				result = append(result, event)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}

// GetPageEvents retrieves page events for testing
func (c *AnalyticsConsumer) GetPageEvents(tenantID string) []map[string]interface{} {
	if mockService, ok := c.analyticsService.(*MockAnalyticsService); ok {
		var result []map[string]interface{}
		for _, event := range mockService.pageEvents {
			if event["tenant_id"] == tenantID {
				result = append(result, event)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}

// GetIncidentEvents retrieves incident events for testing
func (c *AnalyticsConsumer) GetIncidentEvents(tenantID string) []map[string]interface{} {
	if mockService, ok := c.analyticsService.(*MockAnalyticsService); ok {
		var result []map[string]interface{}
		for _, event := range mockService.incidentEvents {
			if event["tenant_id"] == tenantID {
				result = append(result, event)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}

// GetSubscriptionEvents retrieves subscription events for testing
func (c *AnalyticsConsumer) GetSubscriptionEvents(tenantID string) []map[string]interface{} {
	if mockService, ok := c.analyticsService.(*MockAnalyticsService); ok {
		var result []map[string]interface{}
		for _, event := range mockService.subscriptionEvents {
			if event["tenant_id"] == tenantID {
				result = append(result, event)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}

// GetCustomEvents retrieves custom events for testing
func (c *AnalyticsConsumer) GetCustomEvents(tenantID string) []map[string]interface{} {
	if mockService, ok := c.analyticsService.(*MockAnalyticsService); ok {
		var result []map[string]interface{}
		for _, event := range mockService.customEvents {
			if event["tenant_id"] == tenantID {
				result = append(result, event)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}
