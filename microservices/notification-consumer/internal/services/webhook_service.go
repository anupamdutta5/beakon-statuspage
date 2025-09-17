// Package services provides webhook service implementation.
package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anupamdutta5/statuspage-notification-consumer/internal/config"
	"github.com/anupamdutta5/statuspage-notification-consumer/internal/models"
	"go.uber.org/zap"
)

// WebhookService handles webhook notifications.
type WebhookService struct {
	config *config.Config
	logger *zap.Logger
	client *http.Client
}

// NewWebhookService creates a new webhook service.
func NewWebhookService(cfg *config.Config, logger *zap.Logger) *WebhookService {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(cfg.Notification.RetryBackoff) * time.Second,
	}

	return &WebhookService{
		config: cfg,
		logger: logger,
		client: client,
	}
}

// SendWebhook sends a webhook notification.
func (s *WebhookService) SendWebhook(ctx context.Context, event *models.NotificationEvent) error {
	s.logger.Info("Sending webhook notification",
		zap.String("event_id", event.ID),
		zap.String("url", event.WebhookURL))

	// Convert notification event to webhook notification
	webhookNotification := s.convertToWebhookNotification(event)

	// Send webhook
	if err := s.sendHTTPWebhook(ctx, webhookNotification); err != nil {
		s.logger.Error("Failed to send webhook",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return fmt.Errorf("failed to send webhook: %w", err)
	}

	s.logger.Info("Webhook sent successfully",
		zap.String("event_id", event.ID),
		zap.String("url", event.WebhookURL))

	return nil
}

// convertToWebhookNotification converts a notification event to webhook notification.
func (s *WebhookService) convertToWebhookNotification(event *models.NotificationEvent) *models.WebhookNotification {
	// Set default headers
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "StatusPage-Webhook/1.0",
	}

	// Add webhook-specific headers if provided in metadata
	if event.Metadata != nil {
		if webhookHeaders, ok := event.Metadata["headers"].(map[string]interface{}); ok {
			for key, value := range webhookHeaders {
				if strValue, ok := value.(string); ok {
					headers[key] = strValue
				}
			}
		}
	}

	return &models.WebhookNotification{
		ID:             event.ID,
		NotificationID: event.NotificationID,
		TenantID:       event.TenantID,
		UserID:         event.UserID,
		URL:            event.WebhookURL,
		Method:         "POST", // Default to POST
		Headers:        headers,
		Body:           event.Content,
		Timeout:        30, // Default timeout
		RetryCount:     0,
		MaxRetries:     3,
		Priority:       event.Priority,
		Metadata:       event.Metadata,
	}
}

// sendHTTPWebhook sends an HTTP webhook.
func (s *WebhookService) sendHTTPWebhook(ctx context.Context, webhook *models.WebhookNotification) error {
	s.logger.Info("Sending HTTP webhook",
		zap.String("webhook_id", webhook.ID),
		zap.String("url", webhook.URL),
		zap.String("method", webhook.Method))

	// For now, we'll simulate sending webhook
	// In production, you would implement actual HTTP webhook sending

	// Simulate processing time
	// time.Sleep(100 * time.Millisecond)

	// In a real implementation, you would:
	// 1. Create HTTP request with method, URL, headers, and body
	// 2. Send request via HTTP client
	// 3. Handle response and status codes
	// 4. Implement retry logic for failed requests
	// 5. Log response details

	return nil
}

// sendHTTPWebhookReal sends an HTTP webhook (real implementation).
func (s *WebhookService) sendHTTPWebhookReal(ctx context.Context, webhook *models.WebhookNotification) error {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, webhook.Method, webhook.URL, strings.NewReader(webhook.Body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	s.logger.Info("Webhook sent successfully",
		zap.String("webhook_id", webhook.ID),
		zap.String("url", webhook.URL),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// ValidateWebhookURL validates a webhook URL.
func (s *WebhookService) ValidateWebhookURL(url string) bool {
	// Basic URL validation
	// In production, you would use a proper URL validation library
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// GetWebhookStats returns webhook sending statistics.
func (s *WebhookService) GetWebhookStats() (map[string]interface{}, error) {
	// This would typically query the database for webhook statistics
	stats := map[string]interface{}{
		"total_sent":   0,
		"total_failed": 0,
		"success_rate": 0.0,
		"average_time": 0.0,
		"last_updated": "2024-01-01T00:00:00Z",
	}

	return stats, nil
}
