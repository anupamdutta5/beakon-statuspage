package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Local types for testing - no shared dependencies
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// MockAnalyticsService is a mock implementation of the analytics service
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) ProcessUserEvent(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockAnalyticsService) ProcessPageEvent(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockAnalyticsService) ProcessIncidentEvent(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockAnalyticsService) ProcessSubscriptionEvent(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockAnalyticsService) ProcessCustomEvent(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockAnalyticsService) GetAnalyticsData(tenantID, startDate, endDate string) (map[string]interface{}, error) {
	args := m.Called(tenantID, startDate, endDate)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestAnalyticsConsumer_ProcessEvent tests the basic event processing functionality
func TestAnalyticsConsumer_ProcessEvent(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "test-event-1",
		Type:      "user.registered",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"email": "test@example.com"},
	}

	mockService.On("ProcessUserEvent", event).Return(nil)

	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_HandleUserRegistered tests handling of user registration events
func TestAnalyticsConsumer_HandleUserRegistered(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "user-registered-1",
		Type:      "user.registered",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"email":      "newuser@example.com",
			"source":     "organic",
			"referrer":   "google.com",
			"user_agent": "Mozilla/5.0...",
		},
	}

	mockService.On("ProcessUserEvent", mock.MatchedBy(func(e Event) bool {
		return e.Type == "user.registered" && e.TenantID == "tenant-1"
	})).Return(nil)

	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_HandlePageVisited tests handling of page visit events
func TestAnalyticsConsumer_HandlePageVisited(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "page-visited-1",
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

	mockService.On("ProcessPageEvent", mock.MatchedBy(func(e Event) bool {
		return e.Type == "page.visited" && e.TenantID == "tenant-1"
	})).Return(nil)

	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_HandleIncidentReported tests handling of incident events
func TestAnalyticsConsumer_HandleIncidentReported(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "incident-reported-1",
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

	mockService.On("ProcessIncidentEvent", mock.MatchedBy(func(e Event) bool {
		return e.Type == "incident.reported" && e.TenantID == "tenant-1"
	})).Return(nil)

	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_HandleSubscriptionCreated tests handling of subscription events
func TestAnalyticsConsumer_HandleSubscriptionCreated(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "subscription-created-1",
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

	mockService.On("ProcessSubscriptionEvent", mock.MatchedBy(func(e Event) bool {
		return e.Type == "subscription.created" && e.TenantID == "tenant-1"
	})).Return(nil)

	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_HandleCustomEvent tests handling of custom events
func TestAnalyticsConsumer_HandleCustomEvent(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "custom-event-1",
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

	mockService.On("ProcessCustomEvent", mock.MatchedBy(func(e Event) bool {
		return e.Type == "custom.button_clicked" && e.TenantID == "tenant-1"
	})).Return(nil)

	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_ProcessBulkEvents tests processing multiple events
func TestAnalyticsConsumer_ProcessBulkEvents(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

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
	}

	mockService.On("ProcessUserEvent", mock.AnythingOfType("Event")).Return(nil)
	mockService.On("ProcessPageEvent", mock.AnythingOfType("Event")).Return(nil)
	mockService.On("ProcessIncidentEvent", mock.AnythingOfType("Event")).Return(nil)

	for _, event := range testEvents {
		err := consumer.ProcessEvent(event)
		assert.NoError(t, err)
	}

	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_ProcessEventWithInvalidData tests error handling
func TestAnalyticsConsumer_ProcessEventWithInvalidData(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	event := Event{
		ID:        "invalid-event-1",
		Type:      "invalid.type",
		TenantID:  "",
		UserID:    "",
		Timestamp: time.Now(),
		Data:      nil,
	}

	// Should handle gracefully without calling any service methods
	err := consumer.ProcessEvent(event)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestAnalyticsConsumer_GetAnalyticsData tests analytics data retrieval
func TestAnalyticsConsumer_GetAnalyticsData(t *testing.T) {
	mockService := new(MockAnalyticsService)
	consumer := &AnalyticsConsumer{
		analyticsService: mockService,
	}

	expectedData := map[string]interface{}{
		"total_users":   100,
		"page_views":    1000,
		"incidents":     5,
		"subscriptions": 25,
	}

	mockService.On("GetAnalyticsData", "tenant-1", "2024-01-01", "2024-01-31").Return(expectedData, nil)

	data, err := consumer.GetAnalyticsData("tenant-1", "2024-01-01", "2024-01-31")
	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	mockService.AssertExpectations(t)
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
