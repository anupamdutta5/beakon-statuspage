// Package providers provides notification delivery providers.
package providers

import (
	"context"
	"time"
)

// Provider defines the interface for notification providers.
type Provider interface {
	// Send sends a notification using the provider
	Send(ctx context.Context, request *NotificationRequest) (*NotificationResponse, error)

	// ValidateConfig validates the provider configuration
	ValidateConfig(config map[string]interface{}) error

	// GetType returns the provider type
	GetType() string

	// GetName returns the provider name
	GetName() string

	// IsEnabled returns whether the provider is enabled
	IsEnabled() bool
}

// NotificationRequest represents a notification request.
type NotificationRequest struct {
	Recipient   string                 `json:"recipient"`
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"`        // email, sms, webhook, slack, teams
	Priority    string                 `json:"priority"`    // low, normal, high, urgent
	Metadata    map[string]interface{} `json:"metadata"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
}

// NotificationResponse represents a notification response.
type NotificationResponse struct {
	Success        bool                   `json:"success"`
	MessageID      string                 `json:"message_id"`
	Status         string                 `json:"status"`
	ResponseCode   int                    `json:"response_code"`
	ResponseBody   string                 `json:"response_body"`
	Error          string                 `json:"error"`
	SentAt         time.Time              `json:"sent_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ProviderConfig represents provider configuration.
type ProviderConfig struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Settings map[string]interface{} `json:"settings"`
}