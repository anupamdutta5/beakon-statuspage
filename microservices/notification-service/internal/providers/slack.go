// Package providers provides Slack notification delivery.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anupamdutta5/shared-resilience"
	"go.uber.org/zap"
)

// SlackProvider implements notifications via Slack webhooks.
type SlackProvider struct {
	webhookURL    string
	channel       string
	username      string
	iconEmoji     string
	enabled       bool
	serviceClient *resilience.ServiceClient
	logger        *zap.Logger
}

// SlackConfig represents Slack configuration.
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
	Enabled    bool   `json:"enabled"`
}

// SlackMessage represents a Slack message payload.
type SlackMessage struct {
	Text        string             `json:"text"`
	Channel     string             `json:"channel,omitempty"`
	Username    string             `json:"username,omitempty"`
	IconEmoji   string             `json:"icon_emoji,omitempty"`
	Attachments []SlackAttachment  `json:"attachments,omitempty"`
	Blocks      []SlackBlock       `json:"blocks,omitempty"`
}

// SlackAttachment represents a Slack message attachment.
type SlackAttachment struct {
	Color      string       `json:"color,omitempty"`
	Pretext    string       `json:"pretext,omitempty"`
	AuthorName string       `json:"author_name,omitempty"`
	AuthorLink string       `json:"author_link,omitempty"`
	AuthorIcon string       `json:"author_icon,omitempty"`
	Title      string       `json:"title,omitempty"`
	TitleLink  string       `json:"title_link,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	ImageURL   string       `json:"image_url,omitempty"`
	ThumbURL   string       `json:"thumb_url,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	FooterIcon string       `json:"footer_icon,omitempty"`
	Timestamp  int64        `json:"ts,omitempty"`
}

// SlackField represents a Slack attachment field.
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SlackBlock represents a Slack block element.
type SlackBlock struct {
	Type string      `json:"type"`
	Text interface{} `json:"text,omitempty"`
}

// NewSlackProvider creates a new Slack provider.
func NewSlackProvider(config map[string]interface{}, serviceClient *resilience.ServiceClient, logger *zap.Logger) (*SlackProvider, error) {
	var slackConfig SlackConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &slackConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Slack config: %w", err)
	}

	provider := &SlackProvider{
		webhookURL:    slackConfig.WebhookURL,
		channel:       slackConfig.Channel,
		username:      slackConfig.Username,
		iconEmoji:     slackConfig.IconEmoji,
		enabled:       slackConfig.Enabled,
		serviceClient: serviceClient,
		logger:        logger,
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Slack configuration: %w", err)
	}

	return provider, nil
}

// Send sends a notification to Slack.
func (p *SlackProvider) Send(ctx context.Context, request *NotificationRequest) (*NotificationResponse, error) {
	if !p.enabled {
		return &NotificationResponse{
			Success: false,
			Status:  "disabled",
			Error:   "Slack provider is disabled",
			SentAt:  time.Now(),
		}, nil
	}

	// Create Slack message
	message := p.createSlackMessage(request)

	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to marshal message: %v", err),
			SentAt:  time.Now(),
		}, err
	}

	// Send via ServiceClient with circuit breaker and retries
	resp, err := p.serviceClient.Call(ctx, resilience.ServiceRequest{
		ServiceName: "slack-api",
		Method:      "POST",
		URL:         p.webhookURL,
		Body:        payload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})

	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to send Slack message: %v", err),
			SentAt:  time.Now(),
		}, err
	}

	body := resp.Body

	response := &NotificationResponse{
		Success:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		Status:       "sent",
		ResponseCode: resp.StatusCode,
		ResponseBody: string(body),
		SentAt:       time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	if resp.StatusCode >= 400 {
		response.Success = false
		response.Status = "failed"
		response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	p.logger.Info("Message sent to Slack",
		zap.String("channel", p.channel),
		zap.String("status", response.Status),
		zap.Int("response_code", resp.StatusCode),
		zap.Bool("success", response.Success))

	return response, nil
}

// createSlackMessage creates a Slack message from the notification request.
func (p *SlackProvider) createSlackMessage(request *NotificationRequest) *SlackMessage {
	message := &SlackMessage{
		Text:     request.Content,
		Channel:  p.channel,
		Username: p.username,
	}

	if p.iconEmoji != "" {
		message.IconEmoji = p.iconEmoji
	}

	// Add rich formatting if metadata contains structured data
	if request.Metadata != nil {
		if title, ok := request.Metadata["title"].(string); ok && title != "" {
			// Create attachment for better formatting
			attachment := SlackAttachment{
				Title: title,
				Text:  request.Content,
			}

			// Set color based on priority
			switch request.Priority {
			case "urgent":
				attachment.Color = "danger"
			case "high":
				attachment.Color = "warning"
			case "normal":
				attachment.Color = "good"
			case "low":
				attachment.Color = "#439FE0"
			}

			// Add fields from metadata
			if fields, ok := request.Metadata["fields"].([]interface{}); ok {
				for _, field := range fields {
					if fieldMap, ok := field.(map[string]interface{}); ok {
						if title, hasTitle := fieldMap["title"].(string); hasTitle {
							if value, hasValue := fieldMap["value"].(string); hasValue {
								short := false
								if shortVal, hasShort := fieldMap["short"].(bool); hasShort {
									short = shortVal
								}
								attachment.Fields = append(attachment.Fields, SlackField{
									Title: title,
									Value: value,
									Short: short,
								})
							}
						}
					}
				}
			}

			// Add timestamp
			attachment.Timestamp = time.Now().Unix()

			message.Attachments = []SlackAttachment{attachment}
			message.Text = "" // Clear text since we're using attachment
		}
	}

	return message
}

// ValidateConfig validates the Slack configuration.
func (p *SlackProvider) ValidateConfig(config map[string]interface{}) error {
	var slackConfig SlackConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &slackConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if slackConfig.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	if !strings.HasPrefix(slackConfig.WebhookURL, "https://hooks.slack.com/") {
		return fmt.Errorf("webhook_url must be a valid Slack webhook URL")
	}

	return nil
}

// GetType returns the provider type.
func (p *SlackProvider) GetType() string {
	return "slack"
}

// GetName returns the provider name.
func (p *SlackProvider) GetName() string {
	return "slack"
}

// IsEnabled returns whether the provider is enabled.
func (p *SlackProvider) IsEnabled() bool {
	return p.enabled
}