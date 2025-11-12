package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookIntegration represents a custom webhook configuration
type WebhookIntegration struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	TenantID            uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name                string    `gorm:"size:255;not null" json:"name"`
	URL                 string    `gorm:"type:text;not null" json:"url"`
	Method              string    `gorm:"size:10;default:'POST'" json:"method"` // POST, PUT, PATCH
	ContentType         string    `gorm:"size:100;default:'application/json'" json:"content_type"`
	SecretKey           string    `gorm:"size:255" json:"secret_key"` // For HMAC signature
	CustomHeaders       string    `gorm:"type:jsonb" json:"custom_headers"` // JSON object of headers
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	NotifyOnDown        bool      `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool      `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool      `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool      `gorm:"default:false" json:"notify_on_maintenance"`
	TimeoutSeconds      int       `gorm:"default:10" json:"timeout_seconds"`
	RetryCount          int       `gorm:"default:3" json:"retry_count"`
	RetryDelaySeconds   int       `gorm:"default:5" json:"retry_delay_seconds"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// WebhookMonitorMapping maps monitors to specific webhooks
type WebhookMonitorMapping struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	IntegrationID uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID     uint      `gorm:"not null;index" json:"monitor_id"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IntegrationID  uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID      uint      `gorm:"not null;index" json:"monitor_id"`
	EventType      string    `gorm:"size:50;not null" json:"event_type"`
	URL            string    `gorm:"type:text" json:"url"`
	Method         string    `gorm:"size:10" json:"method"`
	RequestBody    string    `gorm:"type:text" json:"request_body"`
	RequestHeaders string    `gorm:"type:jsonb" json:"request_headers"`
	ResponseCode   int       `json:"response_code"`
	ResponseBody   string    `gorm:"type:text" json:"response_body"`
	Status         string    `gorm:"size:50" json:"status"` // pending, sent, failed, retrying
	AttemptCount   int       `gorm:"default:1" json:"attempt_count"`
	ErrorMessage   string    `gorm:"type:text" json:"error_message"`
	SentAt         *time.Time `json:"sent_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WebhookPayload represents the payload sent to custom webhooks
type WebhookPayload struct {
	Event       string                 `json:"event"`
	EventType   string                 `json:"event_type"`
	Monitor     WebhookMonitorData     `json:"monitor"`
	Status      WebhookStatusData      `json:"status"`
	Performance WebhookPerformanceData `json:"performance,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   string                 `json:"timestamp"`
	TenantID    string                 `json:"tenant_id"`
}

// WebhookMonitorData represents monitor information in webhook payload
type WebhookMonitorData struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	Location string `json:"location,omitempty"`
}

// WebhookStatusData represents status information in webhook payload
type WebhookStatusData struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
	Changed  bool   `json:"changed"`
}

// WebhookPerformanceData represents performance metrics in webhook payload
type WebhookPerformanceData struct {
	ResponseTime   int `json:"response_time_ms"`
	StatusCode     int `json:"status_code"`
	TTFB           int `json:"ttfb_ms,omitempty"`
	DNSTime        int `json:"dns_time_ms,omitempty"`
	ConnectionTime int `json:"connection_time_ms,omitempty"`
}

// WebhookIntegrationService manages custom webhook integrations
type WebhookIntegrationService struct {
	db     *gorm.DB
	logger *zap.Logger
	httpClient *http.Client
}

// NewWebhookIntegrationService creates a new webhook integration service
func NewWebhookIntegrationService(db *gorm.DB, serviceClient *resilience.ServiceClient, logger *zap.Logger) *WebhookIntegrationService {
	return &WebhookIntegrationService{
		db:     db,
		logger: logger,
	}
}

// CreateIntegration creates a new webhook integration
func (s *WebhookIntegrationService) CreateIntegration(integration *WebhookIntegration) error {
	// Validate webhook URL
	err := s.TestWebhook(integration)
	if err != nil {
		return fmt.Errorf("webhook validation failed: %w", err)
	}

	err = s.db.Create(integration).Error
	if err != nil {
		s.logger.Error("Failed to create webhook integration",
			zap.Error(err),
			zap.String("name", integration.Name))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Webhook integration created",
		zap.Uint("id", integration.ID),
		zap.String("name", integration.Name),
		zap.String("url", integration.URL))

	return nil
}

// UpdateIntegration updates an existing webhook integration
func (s *WebhookIntegrationService) UpdateIntegration(integration *WebhookIntegration) error {
	err := s.db.Save(integration).Error
	if err != nil {
		s.logger.Error("Failed to update webhook integration", zap.Error(err))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("Webhook integration updated", zap.Uint("id", integration.ID))
	return nil
}

// GetIntegration retrieves a webhook integration by ID
func (s *WebhookIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*WebhookIntegration, error) {
	var integration WebhookIntegration
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&integration).Error
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all integrations for a tenant
func (s *WebhookIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]WebhookIntegration, error) {
	var integrations []WebhookIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	return integrations, nil
}

// DeleteIntegration deletes a webhook integration
func (s *WebhookIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&WebhookIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Webhook integration deleted", zap.Uint("id", id))
	return nil
}

// MapMonitor maps a monitor to a webhook integration
func (s *WebhookIntegrationService) MapMonitor(mapping *WebhookMonitorMapping) error {
	err := s.db.Create(mapping).Error
	if err != nil {
		s.logger.Error("Failed to create monitor mapping", zap.Error(err))
		return fmt.Errorf("failed to map monitor: %w", err)
	}

	s.logger.Info("Monitor mapped to webhook",
		zap.Uint("integration_id", mapping.IntegrationID),
		zap.Uint("monitor_id", mapping.MonitorID))

	return nil
}

// UnmapMonitor removes a monitor mapping
func (s *WebhookIntegrationService) UnmapMonitor(id uint) error {
	result := s.db.Delete(&WebhookMonitorMapping{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to unmap monitor: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("mapping not found")
	}

	s.logger.Info("Monitor unmapped", zap.Uint("id", id))
	return nil
}

// GetMonitorMappings retrieves all mappings for a monitor
func (s *WebhookIntegrationService) GetMonitorMappings(monitorID uint) ([]WebhookMonitorMapping, error) {
	var mappings []WebhookMonitorMapping
	err := s.db.Where("monitor_id = ? AND is_active = ?", monitorID, true).Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get mappings: %w", err)
	}
	return mappings, nil
}

// SendMonitorAlert sends a monitor alert to configured webhooks
func (s *WebhookIntegrationService) SendMonitorAlert(ctx context.Context, monitorID uint, tenantID uuid.UUID, eventType string, monitorData map[string]interface{}) error {
	// Get active integrations
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Get monitor mappings
	mappings, err := s.GetMonitorMappings(monitorID)
	if err != nil {
		return fmt.Errorf("failed to get mappings: %w", err)
	}

	// Determine which integrations to use
	var activeIntegrations []WebhookIntegration
	for _, integration := range integrations {
		if !integration.IsActive {
			continue
		}

		// Check event type filter
		shouldNotify := false
		switch eventType {
		case "down":
			shouldNotify = integration.NotifyOnDown
		case "up":
			shouldNotify = integration.NotifyOnUp
		case "degraded":
			shouldNotify = integration.NotifyOnDegraded
		case "maintenance":
			shouldNotify = integration.NotifyOnMaintenance
		}

		if !shouldNotify {
			continue
		}

		// Check if monitor is mapped to this integration
		if len(mappings) > 0 {
			mapped := false
			for _, mapping := range mappings {
				if mapping.IntegrationID == integration.ID && mapping.IsActive {
					mapped = true
					break
				}
			}
			if !mapped {
				continue
			}
		}

		activeIntegrations = append(activeIntegrations, integration)
	}

	if len(activeIntegrations) == 0 {
		s.logger.Info("No active webhook integrations for event",
			zap.Uint("monitor_id", monitorID),
			zap.String("event_type", eventType))
		return nil
	}

	// Build payload
	payload := s.buildPayload(monitorID, tenantID, eventType, monitorData)

	// Send to each active integration
	for _, integration := range activeIntegrations {
		err := s.sendWebhook(ctx, &integration, payload, monitorID, eventType)
		if err != nil {
			s.logger.Error("Failed to send webhook",
				zap.Error(err),
				zap.Uint("integration_id", integration.ID))
		}
	}

	return nil
}

// buildPayload builds the webhook payload
func (s *WebhookIntegrationService) buildPayload(monitorID uint, tenantID uuid.UUID, eventType string, data map[string]interface{}) *WebhookPayload {
	payload := &WebhookPayload{
		Event:     "monitor.status.changed",
		EventType: eventType,
		Monitor: WebhookMonitorData{
			ID: monitorID,
		},
		Status: WebhookStatusData{
			Current: eventType,
			Changed: true,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		TenantID:  tenantID.String(),
	}

	// Extract monitor data
	if name, ok := data["monitor_name"].(string); ok {
		payload.Monitor.Name = name
	}
	if url, ok := data["monitor_url"].(string); ok {
		payload.Monitor.URL = url
	}
	if monitorType, ok := data["monitor_type"].(string); ok {
		payload.Monitor.Type = monitorType
	}
	if location, ok := data["location"].(string); ok {
		payload.Monitor.Location = location
	}

	// Extract performance data
	if responseTime, ok := data["response_time"].(int); ok {
		payload.Performance.ResponseTime = responseTime
	}
	if statusCode, ok := data["status_code"].(int); ok {
		payload.Performance.StatusCode = statusCode
	}
	if ttfb, ok := data["ttfb"].(int); ok {
		payload.Performance.TTFB = ttfb
	}

	// Extract error
	if errorMsg, ok := data["error"].(string); ok {
		payload.Error = errorMsg
	}

	// Extract previous status
	if prevStatus, ok := data["previous_status"].(string); ok {
		payload.Status.Previous = prevStatus
	}

	return payload
}

// sendWebhook sends the webhook with retry logic
func (s *WebhookIntegrationService) sendWebhook(ctx context.Context, integration *WebhookIntegration, payload *WebhookPayload, monitorID uint, eventType string) error {
	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create delivery record
	delivery := &WebhookDelivery{
		IntegrationID: integration.ID,
		MonitorID:     monitorID,
		EventType:     eventType,
		URL:           integration.URL,
		Method:        integration.Method,
		RequestBody:   string(payloadBytes),
		Status:        "pending",
		AttemptCount:  0,
	}

	// Retry logic
	maxAttempts := integration.RetryCount + 1
	retryDelay := time.Duration(integration.RetryDelaySeconds) * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delivery.AttemptCount = attempt

		if attempt > 1 {
			delivery.Status = "retrying"
			s.db.Save(delivery)
			time.Sleep(retryDelay)
		}

		// Send request
		err := s.executeWebhook(ctx, integration, payloadBytes, delivery, payload.TenantID)
		if err == nil {
			// Success
			delivery.Status = "sent"
			now := time.Now()
			delivery.SentAt = &now
			s.db.Save(delivery)

			s.logger.Info("Webhook delivered successfully",
				zap.Uint("integration_id", integration.ID),
				zap.Uint("monitor_id", monitorID),
				zap.Int("attempt", attempt))
			return nil
		}

		// Log error
		s.logger.Warn("Webhook delivery failed",
			zap.Error(err),
			zap.Uint("integration_id", integration.ID),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", maxAttempts))
	}

	// All attempts failed
	delivery.Status = "failed"
	s.db.Save(delivery)

	return fmt.Errorf("webhook delivery failed after %d attempts", maxAttempts)
}

// executeWebhook executes a single webhook request
func (s *WebhookIntegrationService) executeWebhook(ctx context.Context, integration *WebhookIntegration, payload []byte, delivery *WebhookDelivery, tenantID string) error {
	// Create request with custom timeout
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(integration.TimeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, integration.Method, integration.URL, bytes.NewBuffer(payload))
	if err != nil {
		delivery.ErrorMessage = err.Error()
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", integration.ContentType)

	// Add custom headers
	if integration.CustomHeaders != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(integration.CustomHeaders), &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}
	}

	// Add HMAC signature if secret key is configured
	if integration.SecretKey != "" {
		signature := s.generateHMACSignature(payload, integration.SecretKey)
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Signature-256", "sha256="+signature)
	}

	// Add timestamp and tenant headers
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	req.Header.Set("X-Tenant-ID", tenantID)

	// Store request headers
	headersJSON, _ := json.Marshal(req.Header)
	delivery.RequestHeaders = string(headersJSON)

	// Send request using HTTP client
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		delivery.ErrorMessage = err.Error()
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)
	delivery.ResponseCode = resp.StatusCode
	delivery.ResponseBody = string(body)

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		delivery.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// generateHMACSignature generates HMAC SHA256 signature
func (s *WebhookIntegrationService) generateHMACSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// TestWebhook tests a webhook integration
func (s *WebhookIntegrationService) TestWebhook(integration *WebhookIntegration) error {
	testPayload := &WebhookPayload{
		Event:     "webhook.test",
		EventType: "test",
		Monitor: WebhookMonitorData{
			ID:   0,
			Name: "Test Monitor",
			Type: "http",
		},
		Status: WebhookStatusData{
			Current: "operational",
			Changed: false,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal test payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, integration.Method, integration.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Content-Type", integration.ContentType)

	if integration.SecretKey != "" {
		signature := s.generateHMACSignature(payloadBytes, integration.SecretKey)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Send request using HTTP client
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("test request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("test webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetDeliveryHistory retrieves delivery history for a monitor
func (s *WebhookIntegrationService) GetDeliveryHistory(monitorID uint, limit int) ([]WebhookDelivery, error) {
	var deliveries []WebhookDelivery
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&deliveries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get delivery history: %w", err)
	}

	return deliveries, nil
}

// GetDeliveryStats retrieves delivery statistics
func (s *WebhookIntegrationService) GetDeliveryStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var integrations []WebhookIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		return map[string]interface{}{
			"total_sent":     0,
			"total_failed":   0,
			"success_rate":   0.0,
			"avg_attempts":   0.0,
		}, nil
	}

	integrationIDs := make([]uint, len(integrations))
	for i, integration := range integrations {
		integrationIDs[i] = integration.ID
	}

	// Count sent
	var totalSent int64
	s.db.Model(&WebhookDelivery{}).
		Where("integration_id IN ? AND status = ? AND sent_at BETWEEN ? AND ?", integrationIDs, "sent", startDate, endDate).
		Count(&totalSent)

	// Count failed
	var totalFailed int64
	s.db.Model(&WebhookDelivery{}).
		Where("integration_id IN ? AND status = ? AND created_at BETWEEN ? AND ?", integrationIDs, "failed", startDate, endDate).
		Count(&totalFailed)

	// Calculate average attempts
	var deliveries []WebhookDelivery
	s.db.Where("integration_id IN ? AND created_at BETWEEN ? AND ?", integrationIDs, startDate, endDate).
		Find(&deliveries)

	avgAttempts := 0.0
	if len(deliveries) > 0 {
		totalAttempts := 0
		for _, delivery := range deliveries {
			totalAttempts += delivery.AttemptCount
		}
		avgAttempts = float64(totalAttempts) / float64(len(deliveries))
	}

	successRate := 0.0
	total := totalSent + totalFailed
	if total > 0 {
		successRate = float64(totalSent) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_sent":     totalSent,
		"total_failed":   totalFailed,
		"success_rate":   successRate,
		"avg_attempts":   avgAttempts,
	}, nil
}
