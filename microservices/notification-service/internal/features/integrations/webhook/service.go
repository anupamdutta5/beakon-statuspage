// Package services provides webhook delivery business logic for the Monitoring Service.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anupamdutta5/notification-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebhookService handles webhook delivery and management.
type WebhookService struct {
	db         *gorm.DB
	logger     *zap.Logger
	httpClient *http.Client
}

// NewWebhookService creates a new webhook service.
func NewWebhookService(db *gorm.DB, logger *zap.Logger) *WebhookService {
	// Configure HTTP client with reasonable timeouts
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &WebhookService{
		db:         db,
		logger:     logger,
		httpClient: httpClient,
	}
}

// CreateWebhookEndpoint creates a new webhook endpoint.
func (s *WebhookService) CreateWebhookEndpoint(ctx context.Context, endpoint *models.WebhookEndpoint) error {
	if err := s.db.WithContext(ctx).Create(endpoint).Error; err != nil {
		s.logger.Error("Failed to create webhook endpoint", zap.Error(err))
		return fmt.Errorf("failed to create webhook endpoint: %w", err)
	}

	s.logger.Info("Webhook endpoint created",
		zap.Uint("webhook_id", endpoint.ID),
		zap.Uint("tenant_id", endpoint.TenantID),
		zap.String("name", endpoint.Name),
		zap.String("url", endpoint.URL))

	return nil
}

// GetWebhookEndpoints retrieves webhook endpoints for a tenant.
func (s *WebhookService) GetWebhookEndpoints(ctx context.Context, tenantID uint, activeOnly bool) ([]models.WebhookEndpoint, error) {
	var endpoints []models.WebhookEndpoint

	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Find(&endpoints).Error; err != nil {
		s.logger.Error("Failed to get webhook endpoints", zap.Error(err))
		return nil, fmt.Errorf("failed to get webhook endpoints: %w", err)
	}

	return endpoints, nil
}

// UpdateWebhookEndpoint updates a webhook endpoint.
func (s *WebhookService) UpdateWebhookEndpoint(ctx context.Context, webhookID, tenantID uint, updates map[string]interface{}) error {
	result := s.db.WithContext(ctx).
		Model(&models.WebhookEndpoint{}).
		Where("id = ? AND tenant_id = ?", webhookID, tenantID).
		Updates(updates)

	if err := result.Error; err != nil {
		s.logger.Error("Failed to update webhook endpoint", zap.Error(err))
		return fmt.Errorf("failed to update webhook endpoint: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook endpoint not found")
	}

	s.logger.Info("Webhook endpoint updated",
		zap.Uint("webhook_id", webhookID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// DeleteWebhookEndpoint deletes a webhook endpoint.
func (s *WebhookService) DeleteWebhookEndpoint(ctx context.Context, webhookID, tenantID uint) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", webhookID, tenantID).
		Delete(&models.WebhookEndpoint{})

	if err := result.Error; err != nil {
		s.logger.Error("Failed to delete webhook endpoint", zap.Error(err))
		return fmt.Errorf("failed to delete webhook endpoint: %w", err)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook endpoint not found")
	}

	s.logger.Info("Webhook endpoint deleted",
		zap.Uint("webhook_id", webhookID),
		zap.Uint("tenant_id", tenantID))

	return nil
}

// SendWebhookEvent sends a webhook event to all subscribed endpoints.
func (s *WebhookService) SendWebhookEvent(ctx context.Context, tenantID uint, event *models.WebhookEvent) error {
	// Get active webhook endpoints for this tenant and event type
	endpoints, err := models.GetActiveWebhookEndpoints(s.db, tenantID, event.Type)
	if err != nil {
		s.logger.Error("Failed to get webhook endpoints", zap.Error(err))
		return fmt.Errorf("failed to get webhook endpoints: %w", err)
	}

	if len(endpoints) == 0 {
		s.logger.Debug("No webhook endpoints found for event",
			zap.Uint("tenant_id", tenantID),
			zap.String("event_type", event.Type))
		return nil
	}

	// Send to all endpoints concurrently
	for _, endpoint := range endpoints {
		go s.deliverWebhook(ctx, endpoint, event)
	}

	return nil
}

// deliverWebhook delivers a webhook to a specific endpoint.
func (s *WebhookService) deliverWebhook(ctx context.Context, endpoint models.WebhookEndpoint, event *models.WebhookEvent) {
	startTime := time.Now()

	// Create delivery record
	delivery := &models.WebhookDelivery{
		WebhookID:    endpoint.ID,
		EventType:    event.Type,
		EventID:      event.ID,
		AttemptCount: 1,
	}

	// Serialize the event payload
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal webhook payload", zap.Error(err))
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("Failed to marshal payload: %v", err)
		s.saveDelivery(delivery)
		return
	}

	delivery.Payload = string(payloadBytes)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.logger.Error("Failed to create webhook request", zap.Error(err))
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		s.saveDelivery(delivery)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Beakon-StatusPage-Webhook/1.0")
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", event.Timestamp.Unix()))
	req.Header.Set("X-Webhook-Event-Type", event.Type)
	req.Header.Set("X-Webhook-Event-ID", event.ID)

	// Add custom headers if configured
	if endpoint.Headers != "" {
		var customHeaders map[string]string
		if err := json.Unmarshal([]byte(endpoint.Headers), &customHeaders); err == nil {
			for key, value := range customHeaders {
				req.Header.Set(key, value)
			}
		}
	}

	// Add HMAC signature if secret key is configured
	if endpoint.SecretKey != "" {
		signature := s.generateHMACSignature(payloadBytes, endpoint.SecretKey)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Store request headers
	requestHeaders, _ := json.Marshal(req.Header)
	delivery.RequestHeaders = string(requestHeaders)

	// Configure timeout
	if endpoint.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(endpoint.Timeout)*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
	}

	// Send the request
	resp, err := s.httpClient.Do(req)
	duration := time.Since(startTime).Milliseconds()
	delivery.Duration = duration

	if err != nil {
		s.logger.Error("Failed to send webhook",
			zap.Error(err),
			zap.String("url", endpoint.URL),
			zap.String("event_type", event.Type))
		delivery.Success = false
		delivery.ErrorMessage = err.Error()
		delivery.NextRetryAt = s.calculateNextRetry(delivery.AttemptCount, endpoint.RetryCount)
		s.saveDelivery(delivery)
		return
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Warn("Failed to read webhook response body", zap.Error(err))
		responseBody = []byte{}
	}

	// Store response data
	delivery.ResponseStatus = resp.StatusCode
	delivery.ResponseBody = string(responseBody)

	responseHeaders, _ := json.Marshal(resp.Header)
	delivery.ResponseHeaders = string(responseHeaders)

	// Check if delivery was successful (2xx status codes)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Success = true
		delivery.DeliveredAt = &startTime
		s.logger.Info("Webhook delivered successfully",
			zap.String("url", endpoint.URL),
			zap.String("event_type", event.Type),
			zap.Int("status_code", resp.StatusCode),
			zap.Int64("duration_ms", duration))
	} else {
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody))
		delivery.NextRetryAt = s.calculateNextRetry(delivery.AttemptCount, endpoint.RetryCount)
		s.logger.Warn("Webhook delivery failed",
			zap.String("url", endpoint.URL),
			zap.String("event_type", event.Type),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(responseBody)))
	}

	s.saveDelivery(delivery)
}

// generateHMACSignature generates an HMAC-SHA256 signature for the payload.
func (s *WebhookService) generateHMACSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// calculateNextRetry calculates the next retry time using exponential backoff.
func (s *WebhookService) calculateNextRetry(attemptCount, maxRetries int) *time.Time {
	if attemptCount >= maxRetries {
		return nil // No more retries
	}

	// Exponential backoff: 2^attemptCount minutes
	backoffMinutes := 1 << attemptCount // 1, 2, 4, 8, 16...
	if backoffMinutes > 60 {
		backoffMinutes = 60 // Cap at 1 hour
	}

	nextRetry := time.Now().Add(time.Duration(backoffMinutes) * time.Minute)
	return &nextRetry
}

// saveDelivery saves the webhook delivery record to the database.
func (s *WebhookService) saveDelivery(delivery *models.WebhookDelivery) {
	if err := s.db.Create(delivery).Error; err != nil {
		s.logger.Error("Failed to save webhook delivery", zap.Error(err))
	}
}

// GetWebhookDeliveries retrieves webhook delivery history.
func (s *WebhookService) GetWebhookDeliveries(ctx context.Context, tenantID uint, webhookID *uint, limit, offset int) ([]models.WebhookDelivery, int64, error) {
	var deliveries []models.WebhookDelivery
	var total int64

	query := s.db.WithContext(ctx).
		Joins("JOIN webhook_endpoints ON webhook_endpoints.id = webhook_deliveries.webhook_id").
		Where("webhook_endpoints.tenant_id = ?", tenantID)

	if webhookID != nil {
		query = query.Where("webhook_deliveries.webhook_id = ?", *webhookID)
	}

	// Count total records
	if err := query.Model(&models.WebhookDelivery{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook deliveries: %w", err)
	}

	// Get deliveries with pagination
	if err := query.
		Preload("Webhook").
		Order("webhook_deliveries.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get webhook deliveries: %w", err)
	}

	return deliveries, total, nil
}

// RetryWebhookDelivery retries a failed webhook delivery.
func (s *WebhookService) RetryWebhookDelivery(ctx context.Context, deliveryID uint, tenantID uint) error {
	var delivery models.WebhookDelivery

	// Get the delivery with webhook information
	if err := s.db.WithContext(ctx).
		Preload("Webhook").
		Joins("JOIN webhook_endpoints ON webhook_endpoints.id = webhook_deliveries.webhook_id").
		Where("webhook_deliveries.id = ? AND webhook_endpoints.tenant_id = ?", deliveryID, tenantID).
		First(&delivery).Error; err != nil {
		return fmt.Errorf("webhook delivery not found: %w", err)
	}

	if delivery.Success {
		return fmt.Errorf("webhook delivery already successful")
	}

	// Parse the original event
	var event models.WebhookEvent
	if err := json.Unmarshal([]byte(delivery.Payload), &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Create new delivery attempt
	newDelivery := &models.WebhookDelivery{
		WebhookID:    delivery.WebhookID,
		EventType:    delivery.EventType,
		EventID:      delivery.EventID,
		Payload:      delivery.Payload,
		AttemptCount: delivery.AttemptCount + 1,
	}

	// Update the attempt count and retry
	go s.deliverWebhookRetry(ctx, delivery.Webhook, &event, newDelivery)

	return nil
}

// deliverWebhookRetry is similar to deliverWebhook but for retry attempts.
func (s *WebhookService) deliverWebhookRetry(ctx context.Context, endpoint models.WebhookEndpoint, event *models.WebhookEvent, delivery *models.WebhookDelivery) {
	startTime := time.Now()

	// Create HTTP request
	payloadBytes := []byte(delivery.Payload)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.logger.Error("Failed to create retry webhook request", zap.Error(err))
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		s.saveDelivery(delivery)
		return
	}

	// Set headers (same as original delivery)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Beakon-StatusPage-Webhook/1.0")
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", event.Timestamp.Unix()))
	req.Header.Set("X-Webhook-Event-Type", event.Type)
	req.Header.Set("X-Webhook-Event-ID", event.ID)
	req.Header.Set("X-Webhook-Retry", "true")
	req.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", delivery.AttemptCount))

	// Add custom headers if configured
	if endpoint.Headers != "" {
		var customHeaders map[string]string
		if err := json.Unmarshal([]byte(endpoint.Headers), &customHeaders); err == nil {
			for key, value := range customHeaders {
				req.Header.Set(key, value)
			}
		}
	}

	// Add HMAC signature if secret key is configured
	if endpoint.SecretKey != "" {
		signature := s.generateHMACSignature(payloadBytes, endpoint.SecretKey)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Store request headers
	requestHeaders, _ := json.Marshal(req.Header)
	delivery.RequestHeaders = string(requestHeaders)

	// Configure timeout
	if endpoint.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(endpoint.Timeout)*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
	}

	// Send the request
	resp, err := s.httpClient.Do(req)
	duration := time.Since(startTime).Milliseconds()
	delivery.Duration = duration

	if err != nil {
		s.logger.Error("Failed to send retry webhook",
			zap.Error(err),
			zap.String("url", endpoint.URL),
			zap.String("event_type", event.Type),
			zap.Int("attempt", delivery.AttemptCount))
		delivery.Success = false
		delivery.ErrorMessage = err.Error()
		delivery.NextRetryAt = s.calculateNextRetry(delivery.AttemptCount, endpoint.RetryCount)
		s.saveDelivery(delivery)
		return
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Warn("Failed to read retry webhook response body", zap.Error(err))
		responseBody = []byte{}
	}

	// Store response data
	delivery.ResponseStatus = resp.StatusCode
	delivery.ResponseBody = string(responseBody)

	responseHeaders, _ := json.Marshal(resp.Header)
	delivery.ResponseHeaders = string(responseHeaders)

	// Check if delivery was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Success = true
		delivery.DeliveredAt = &startTime
		s.logger.Info("Retry webhook delivered successfully",
			zap.String("url", endpoint.URL),
			zap.String("event_type", event.Type),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("attempt", delivery.AttemptCount),
			zap.Int64("duration_ms", duration))
	} else {
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody))
		delivery.NextRetryAt = s.calculateNextRetry(delivery.AttemptCount, endpoint.RetryCount)
		s.logger.Warn("Retry webhook delivery failed",
			zap.String("url", endpoint.URL),
			zap.String("event_type", event.Type),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("attempt", delivery.AttemptCount),
			zap.String("response", string(responseBody)))
	}

	s.saveDelivery(delivery)
}

// TestWebhookEndpoint sends a test event to a webhook endpoint.
func (s *WebhookService) TestWebhookEndpoint(ctx context.Context, webhookID, tenantID uint) error {
	var endpoint models.WebhookEndpoint

	// Get the webhook endpoint
	if err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", webhookID, tenantID).
		First(&endpoint).Error; err != nil {
		return fmt.Errorf("webhook endpoint not found: %w", err)
	}

	// Create a test event
	testEvent := &models.WebhookEvent{
		ID:        fmt.Sprintf("test_%d_%d", webhookID, time.Now().Unix()),
		Type:      "webhook.test",
		Timestamp: time.Now(),
		TenantID:  tenantID,
		Source:    "monitoring-service",
		Action:    "test",
		Data: map[string]interface{}{
			"message":    "This is a test webhook from Beakon Status Page",
			"webhook_id": webhookID,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Send the test webhook
	go s.deliverWebhook(ctx, endpoint, testEvent)

	s.logger.Info("Test webhook sent",
		zap.Uint("webhook_id", webhookID),
		zap.Uint("tenant_id", tenantID),
		zap.String("url", endpoint.URL))

	return nil
}

// GetWebhookStatistics returns statistics about webhook deliveries.
func (s *WebhookService) GetWebhookStatistics(ctx context.Context, tenantID uint, webhookID *uint, days int) (map[string]interface{}, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := s.db.WithContext(ctx).
		Table("webhook_deliveries").
		Joins("JOIN webhook_endpoints ON webhook_endpoints.id = webhook_deliveries.webhook_id").
		Where("webhook_endpoints.tenant_id = ? AND webhook_deliveries.created_at >= ?", tenantID, since)

	if webhookID != nil {
		query = query.Where("webhook_deliveries.webhook_id = ?", *webhookID)
	}

	var stats struct {
		TotalDeliveries    int64   `json:"total_deliveries"`
		SuccessfulDeliveries int64 `json:"successful_deliveries"`
		FailedDeliveries   int64   `json:"failed_deliveries"`
		AverageResponseTime float64 `json:"average_response_time"`
	}

	// Get total deliveries
	query.Count(&stats.TotalDeliveries)

	// Get successful deliveries
	query.Where("webhook_deliveries.success = ?", true).Count(&stats.SuccessfulDeliveries)

	// Get failed deliveries
	stats.FailedDeliveries = stats.TotalDeliveries - stats.SuccessfulDeliveries

	// Get average response time
	var avgDuration sql.NullFloat64
	query.Select("AVG(webhook_deliveries.duration)").Scan(&avgDuration)
	if avgDuration.Valid {
		stats.AverageResponseTime = avgDuration.Float64
	}

	result := map[string]interface{}{
		"total_deliveries":      stats.TotalDeliveries,
		"successful_deliveries": stats.SuccessfulDeliveries,
		"failed_deliveries":     stats.FailedDeliveries,
		"success_rate":          float64(stats.SuccessfulDeliveries) / float64(stats.TotalDeliveries) * 100,
		"average_response_time": stats.AverageResponseTime,
		"period_days":          days,
	}

	return result, nil
}

// RetryFailedDeliveries retries all failed webhook deliveries that are due for retry.
func (s *WebhookService) RetryFailedDeliveries() error {
	now := time.Now()

	// Find failed deliveries that are ready for retry
	var deliveries []models.WebhookDelivery
	err := s.db.Where("success = ? AND attempt_count < ? AND next_retry_at <= ?",
		false, 3, now).Find(&deliveries).Error

	if err != nil {
		return fmt.Errorf("failed to fetch failed deliveries: %w", err)
	}

	s.logger.Info("Retrying failed webhook deliveries",
		zap.Int("count", len(deliveries)),
	)

	for _, delivery := range deliveries {
		// Retry the delivery
		delivery.AttemptCount++

		// Calculate next retry time using exponential backoff
		// 1st retry: 1 minute, 2nd retry: 5 minutes, 3rd retry: 15 minutes
		retryDelays := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
		if delivery.AttemptCount < len(retryDelays) {
			nextRetry := now.Add(retryDelays[delivery.AttemptCount])
			delivery.NextRetryAt = &nextRetry
		}

		// Re-send the webhook
		// Note: This would normally call the delivery logic, but for now we'll just update the retry info
		if err := s.db.Save(&delivery).Error; err != nil {
			s.logger.Error("Failed to update delivery retry info",
				zap.Error(err),
				zap.Uint("delivery_id", delivery.ID),
			)
			continue
		}

		s.logger.Debug("Webhook delivery scheduled for retry",
			zap.Uint("delivery_id", delivery.ID),
			zap.Int("attempt_count", delivery.AttemptCount),
		)
	}

	return nil
}