// Package providers provides webhook notification delivery.
package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"

	"go.uber.org/zap"
)

// WebhookProvider implements notifications via HTTP webhooks.
type WebhookProvider struct {
	url         string
	method      string
	headers     map[string]string
	secret      string
	timeout     time.Duration
	enabled     bool
	httpClient  *http.Client
	logger      *zap.Logger
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
	Secret   string            `json:"secret"`
	Timeout  int               `json:"timeout"` // in seconds
	Enabled  bool              `json:"enabled"`
}

// WebhookPayload represents the webhook payload structure.
type WebhookPayload struct {
	Event       string                 `json:"event"`
	Timestamp   string                 `json:"timestamp"`
	ID          string                 `json:"id"`
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"`
	Recipient   string                 `json:"recipient"`
	Metadata    map[string]interface{} `json:"metadata"`
	Signature   string                 `json:"signature,omitempty"`
}

// NewWebhookProvider creates a new webhook provider.
func NewWebhookProvider(config map[string]interface{}, logger *zap.Logger) (*WebhookProvider, error) {
	var webhookConfig WebhookConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &webhookConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook config: %w", err)
	}

	// Set defaults
	if webhookConfig.Method == "" {
		webhookConfig.Method = "POST"
	}
	if webhookConfig.Timeout == 0 {
		webhookConfig.Timeout = 30
	}
	if webhookConfig.Headers == nil {
		webhookConfig.Headers = make(map[string]string)
	}

	// Ensure Content-Type is set
	if _, hasContentType := webhookConfig.Headers["Content-Type"]; !hasContentType {
		webhookConfig.Headers["Content-Type"] = "application/json"
	}

	provider := &WebhookProvider{
		url:     webhookConfig.URL,
		method:  webhookConfig.Method,
		headers: webhookConfig.Headers,
		secret:  webhookConfig.Secret,
		timeout: time.Duration(webhookConfig.Timeout) * time.Second,
		enabled: webhookConfig.Enabled,
		httpClient: &http.Client{
			Timeout: time.Duration(webhookConfig.Timeout) * time.Second,
		},
		logger: logger,
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid webhook configuration: %w", err)
	}

	return provider, nil
}

// Send sends a notification via webhook.
func (p *WebhookProvider) Send(ctx context.Context, request *NotificationRequest) (*NotificationResponse, error) {
	if !p.enabled {
		return &NotificationResponse{
			Success: false,
			Status:  "disabled",
			Error:   "Webhook provider is disabled",
			SentAt:  time.Now(),
		}, nil
	}

	// Create webhook payload
	// Generate secure notification ID
	secureID, err := resilience.GenerateSecureToken(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure notification ID: %w", err)
	}

	payload := &WebhookPayload{
		Event:     "notification.sent",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		ID:        fmt.Sprintf("notif_%s", secureID),
		Subject:   request.Subject,
		Content:   request.Content,
		Type:      request.Type,
		Priority:  request.Priority,
		Recipient: request.Recipient,
		Metadata:  request.Metadata,
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to marshal payload: %v", err),
			SentAt:  time.Now(),
		}, err
	}

	// Add signature if secret is provided
	if p.secret != "" {
		signature := p.generateSignature(payloadBytes)
		payload.Signature = signature

		// Re-marshal with signature
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return &NotificationResponse{
				Success: false,
				Status:  "failed",
				Error:   fmt.Sprintf("Failed to marshal payload with signature: %v", err),
				SentAt:  time.Now(),
			}, err
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, p.method, p.url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to create request: %v", err),
			SentAt:  time.Now(),
		}, err
	}

	// Set headers
	for key, value := range p.headers {
		req.Header.Set(key, value)
	}

	// Add signature header if secret is provided
	if p.secret != "" {
		signature := p.generateSignature(payloadBytes)
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Signature-256", "sha256="+signature)
	}

	// Add user agent
	req.Header.Set("User-Agent", "StatusPage-Webhook/1.0")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to send webhook: %v", err),
			SentAt:  time.Now(),
		}, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &NotificationResponse{
			Success:      false,
			Status:       "failed",
			ResponseCode: resp.StatusCode,
			Error:        fmt.Sprintf("Failed to read response: %v", err),
			SentAt:       time.Now(),
		}, err
	}

	response := &NotificationResponse{
		Success:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		Status:       "sent",
		ResponseCode: resp.StatusCode,
		ResponseBody: string(body),
		SentAt:       time.Now(),
		Metadata: map[string]interface{}{
			"webhook_url":    p.url,
			"webhook_method": p.method,
			"payload_size":   len(payloadBytes),
		},
	}

	if resp.StatusCode >= 400 {
		response.Success = false
		response.Status = "failed"
		response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	p.logger.Info("Webhook notification sent",
		zap.String("url", p.url),
		zap.String("method", p.method),
		zap.String("status", response.Status),
		zap.Int("response_code", resp.StatusCode),
		zap.Bool("success", response.Success),
		zap.Int("payload_size", len(payloadBytes)))

	return response, nil
}

// generateSignature generates HMAC-SHA256 signature for the payload.
func (p *WebhookProvider) generateSignature(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(p.secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateConfig validates the webhook configuration.
func (p *WebhookProvider) ValidateConfig(config map[string]interface{}) error {
	var webhookConfig WebhookConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &webhookConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if webhookConfig.URL == "" {
		return fmt.Errorf("url is required")
	}

	if !strings.HasPrefix(webhookConfig.URL, "http://") && !strings.HasPrefix(webhookConfig.URL, "https://") {
		return fmt.Errorf("url must be a valid HTTP or HTTPS URL")
	}

	// Validate method
	validMethods := []string{"GET", "POST", "PUT", "PATCH"}
	methodValid := false
	for _, method := range validMethods {
		if strings.ToUpper(webhookConfig.Method) == method {
			methodValid = true
			break
		}
	}
	if !methodValid {
		return fmt.Errorf("method must be one of: %s", strings.Join(validMethods, ", "))
	}

	// Validate timeout
	if webhookConfig.Timeout < 1 || webhookConfig.Timeout > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}

	return nil
}

// GetType returns the provider type.
func (p *WebhookProvider) GetType() string {
	return "webhook"
}

// GetName returns the provider name.
func (p *WebhookProvider) GetName() string {
	return "webhook"
}

// IsEnabled returns whether the provider is enabled.
func (p *WebhookProvider) IsEnabled() bool {
	return p.enabled
}