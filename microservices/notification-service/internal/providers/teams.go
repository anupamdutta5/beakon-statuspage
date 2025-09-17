// Package providers provides Microsoft Teams notification delivery.
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TeamsProvider implements notifications via Microsoft Teams webhooks.
type TeamsProvider struct {
	webhookURL string
	enabled    bool
	httpClient *http.Client
	logger     *zap.Logger
}

// TeamsConfig represents Teams configuration.
type TeamsConfig struct {
	WebhookURL string `json:"webhook_url"`
	Enabled    bool   `json:"enabled"`
}

// TeamsMessage represents a Microsoft Teams message payload.
type TeamsMessage struct {
	Type       string              `json:"@type"`
	Context    string              `json:"@context"`
	ThemeColor string              `json:"themeColor,omitempty"`
	Summary    string              `json:"summary"`
	Sections   []TeamsSection      `json:"sections,omitempty"`
	Actions    []TeamsAction       `json:"potentialAction,omitempty"`
}

// TeamsSection represents a Teams message section.
type TeamsSection struct {
	ActivityTitle    string       `json:"activityTitle,omitempty"`
	ActivitySubtitle string       `json:"activitySubtitle,omitempty"`
	ActivityImage    string       `json:"activityImage,omitempty"`
	Text             string       `json:"text,omitempty"`
	Facts            []TeamsFact  `json:"facts,omitempty"`
	Markdown         bool         `json:"markdown,omitempty"`
}

// TeamsFact represents a Teams message fact.
type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TeamsAction represents a Teams message action.
type TeamsAction struct {
	Type    string           `json:"@type"`
	Name    string           `json:"name"`
	Targets []TeamsTarget    `json:"targets,omitempty"`
}

// TeamsTarget represents a Teams action target.
type TeamsTarget struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

// NewTeamsProvider creates a new Microsoft Teams provider.
func NewTeamsProvider(config map[string]interface{}, logger *zap.Logger) (*TeamsProvider, error) {
	var teamsConfig TeamsConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &teamsConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Teams config: %w", err)
	}

	provider := &TeamsProvider{
		webhookURL: teamsConfig.WebhookURL,
		enabled:    teamsConfig.Enabled,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Teams configuration: %w", err)
	}

	return provider, nil
}

// Send sends a notification to Microsoft Teams.
func (p *TeamsProvider) Send(ctx context.Context, request *NotificationRequest) (*NotificationResponse, error) {
	if !p.enabled {
		return &NotificationResponse{
			Success: false,
			Status:  "disabled",
			Error:   "Teams provider is disabled",
			SentAt:  time.Now(),
		}, nil
	}

	// Create Teams message
	message := p.createTeamsMessage(request)

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

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.webhookURL, strings.NewReader(string(payload)))
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to create request: %v", err),
			SentAt:  time.Now(),
		}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to send Teams message: %v", err),
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
		Metadata:     make(map[string]interface{}),
	}

	if resp.StatusCode >= 400 {
		response.Success = false
		response.Status = "failed"
		response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	p.logger.Info("Message sent to Teams",
		zap.String("status", response.Status),
		zap.Int("response_code", resp.StatusCode),
		zap.Bool("success", response.Success))

	return response, nil
}

// createTeamsMessage creates a Teams message from the notification request.
func (p *TeamsProvider) createTeamsMessage(request *NotificationRequest) *TeamsMessage {
	message := &TeamsMessage{
		Type:    "MessageCard",
		Context: "https://schema.org/extensions",
		Summary: request.Subject,
	}

	// Set theme color based on priority
	switch request.Priority {
	case "urgent":
		message.ThemeColor = "FF0000" // Red
	case "high":
		message.ThemeColor = "FF8C00" // Orange
	case "normal":
		message.ThemeColor = "008000" // Green
	case "low":
		message.ThemeColor = "0078D4" // Blue
	default:
		message.ThemeColor = "0078D4" // Default blue
	}

	// Create main section
	section := TeamsSection{
		ActivityTitle: request.Subject,
		Text:          request.Content,
		Markdown:      true,
	}

	// Add facts from metadata
	if request.Metadata != nil {
		if facts, ok := request.Metadata["facts"].([]interface{}); ok {
			for _, fact := range facts {
				if factMap, ok := fact.(map[string]interface{}); ok {
					if name, hasName := factMap["name"].(string); hasName {
						if value, hasValue := factMap["value"].(string); hasValue {
							section.Facts = append(section.Facts, TeamsFact{
								Name:  name,
								Value: value,
							})
						}
					}
				}
			}
		}

		// Add timestamp fact
		section.Facts = append(section.Facts, TeamsFact{
			Name:  "Timestamp",
			Value: time.Now().Format("2006-01-02 15:04:05 UTC"),
		})

		// Add priority fact
		if request.Priority != "" {
			section.Facts = append(section.Facts, TeamsFact{
				Name:  "Priority",
				Value: strings.Title(request.Priority),
			})
		}

		// Add action links if provided
		if actionURL, ok := request.Metadata["action_url"].(string); ok {
			if actionText, ok := request.Metadata["action_text"].(string); ok {
				action := TeamsAction{
					Type: "OpenUri",
					Name: actionText,
					Targets: []TeamsTarget{
						{
							OS:  "default",
							URI: actionURL,
						},
					},
				}
				message.Actions = []TeamsAction{action}
			}
		}
	}

	message.Sections = []TeamsSection{section}

	return message
}

// ValidateConfig validates the Teams configuration.
func (p *TeamsProvider) ValidateConfig(config map[string]interface{}) error {
	var teamsConfig TeamsConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &teamsConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if teamsConfig.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	if !strings.Contains(teamsConfig.WebhookURL, "outlook.office.com") &&
		!strings.Contains(teamsConfig.WebhookURL, "outlook.office365.com") {
		return fmt.Errorf("webhook_url must be a valid Teams webhook URL")
	}

	return nil
}

// GetType returns the provider type.
func (p *TeamsProvider) GetType() string {
	return "teams"
}

// GetName returns the provider name.
func (p *TeamsProvider) GetName() string {
	return "teams"
}

// IsEnabled returns whether the provider is enabled.
func (p *TeamsProvider) IsEnabled() bool {
	return p.enabled
}