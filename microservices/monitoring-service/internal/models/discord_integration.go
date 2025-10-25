// Package models provides Discord integration data models for the Monitoring Service.
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// DiscordIntegration represents a Discord webhook integration configuration
type DiscordIntegration struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`

	// Discord webhook configuration
	WebhookURL  string `gorm:"type:text;not null" json:"webhook_url"` // Discord webhook URL
	WebhookName string `json:"webhook_name,omitempty"`                 // Custom webhook name
	AvatarURL   string `gorm:"type:text" json:"avatar_url,omitempty"`  // Custom avatar for webhook messages

	// Channel configuration
	DefaultChannelID   string `json:"default_channel_id,omitempty"`   // Discord channel ID (snowflake)
	DefaultChannelName string `json:"default_channel_name,omitempty"` // Channel name for display

	// Integration status
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`

	// Notification preferences
	NotifyOnDown        bool `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool `gorm:"default:false" json:"notify_on_maintenance"`

	// User mentions configuration
	MentionUsers    pq.StringArray `gorm:"type:text[]" json:"mention_users,omitempty"`    // Array of Discord user IDs to mention
	MentionRoles    pq.StringArray `gorm:"type:text[]" json:"mention_roles,omitempty"`    // Array of Discord role IDs to mention
	MentionEveryone bool           `gorm:"default:false" json:"mention_everyone"`         // Whether to @everyone

	// Embed customization
	CustomColor       string `json:"custom_color,omitempty"`                  // Hex color for embeds (e.g., #5865F2)
	IncludeMonitorURL bool   `gorm:"default:true" json:"include_monitor_url"` // Include link to monitor
	IncludeTimestamp  bool   `gorm:"default:true" json:"include_timestamp"`   // Include timestamp in embeds

	// Rate limiting and retries
	RetryCount           int `gorm:"default:3" json:"retry_count"`
	RetryIntervalSeconds int `gorm:"default:5" json:"retry_interval_seconds"`

	// Related entities
	ChannelSubscriptions []DiscordChannelSubscription `gorm:"foreignKey:IntegrationID" json:"channel_subscriptions,omitempty"`
	Notifications        []DiscordNotification         `gorm:"foreignKey:IntegrationID" json:"notifications,omitempty"`
}

// DiscordChannelSubscription represents a monitor-to-Discord-channel subscription mapping
type DiscordChannelSubscription struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	IntegrationID uint                `gorm:"not null;index" json:"integration_id"`
	Integration   *DiscordIntegration `gorm:"foreignKey:IntegrationID" json:"integration,omitempty"`

	MonitorID *uint `gorm:"index" json:"monitor_id,omitempty"` // NULL means all monitors

	// Channel override
	ChannelID   string `json:"channel_id,omitempty"`   // Override default channel for this monitor
	ChannelName string `json:"channel_name,omitempty"` // Channel name for display

	// Event filtering per subscription
	NotifyOnDown        bool `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool `gorm:"default:false" json:"notify_on_maintenance"`

	// Subscription status
	IsActive bool `gorm:"default:true" json:"is_active"`
}

// DiscordNotification represents a Discord notification delivery log entry
type DiscordNotification struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	IntegrationID uint                `gorm:"not null;index" json:"integration_id"`
	Integration   *DiscordIntegration `gorm:"foreignKey:IntegrationID" json:"integration,omitempty"`

	MonitorID *uint `gorm:"index" json:"monitor_id,omitempty"`

	// Event details
	EventType   string `gorm:"size:50;not null;index" json:"event_type"` // down, up, degraded, maintenance
	MonitorName string `json:"monitor_name,omitempty"`
	MonitorURL  string `gorm:"type:text" json:"monitor_url,omitempty"`

	// Delivery status
	Status         string `gorm:"size:50;not null;index" json:"status"` // sent, failed, pending, retrying
	HTTPStatusCode *int   `json:"http_status_code,omitempty"`

	// Message content
	MessageContent string          `gorm:"type:text" json:"message_content,omitempty"` // The actual message sent
	EmbedData      json.RawMessage `gorm:"type:jsonb" json:"embed_data,omitempty"`     // Discord embed object

	// Delivery tracking
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`

	// Error tracking
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`

	// Response from Discord
	DiscordMessageID string `json:"discord_message_id,omitempty"` // Discord message ID (snowflake)
	DiscordChannelID string `json:"discord_channel_id,omitempty"`
}

// DiscordEmbed represents a Discord embed message structure
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`        // Decimal color value
	Timestamp   string              `json:"timestamp,omitempty"`    // ISO8601 timestamp
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
}

// DiscordEmbedFooter represents the footer section of a Discord embed
type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordWebhookPayload represents the complete payload sent to Discord webhook
type DiscordWebhookPayload struct {
	Content   string          `json:"content,omitempty"`    // Plain text message
	Username  string          `json:"username,omitempty"`   // Override webhook username
	AvatarURL string          `json:"avatar_url,omitempty"` // Override webhook avatar
	Embeds    []DiscordEmbed  `json:"embeds,omitempty"`     // Rich embeds
}

// DiscordWebhookResponse represents the response from Discord API
type DiscordWebhookResponse struct {
	ID              string `json:"id"`                         // Message ID (snowflake)
	Type            int    `json:"type"`                       // Message type
	ChannelID       string `json:"channel_id"`                 // Channel ID (snowflake)
	Timestamp       string `json:"timestamp"`                  // ISO8601 timestamp
	EditedTimestamp string `json:"edited_timestamp,omitempty"` // ISO8601 timestamp if edited
}

// Constants for Discord integration
const (
	// Event types
	DiscordEventDown        = "down"
	DiscordEventUp          = "up"
	DiscordEventDegraded    = "degraded"
	DiscordEventMaintenance = "maintenance"

	// Notification status
	DiscordStatusPending  = "pending"
	DiscordStatusSent     = "sent"
	DiscordStatusFailed   = "failed"
	DiscordStatusRetrying = "retrying"

	// Discord embed colors (decimal values)
	DiscordColorGreen  = 3066993  // #2ECC71 - Success/Up
	DiscordColorRed    = 15158332 // #E74C3C - Error/Down
	DiscordColorOrange = 15105570 // #E67E22 - Warning/Degraded
	DiscordColorBlue   = 3447003  // #3498DB - Info/Maintenance
	DiscordColorPurple = 10181046 // #9B59B6 - Discord branding
	DiscordColorBlurple = 5793266 // #5865F2 - Discord blurple

	// Discord API limits
	DiscordMaxEmbeds        = 10   // Max embeds per message
	DiscordMaxFields        = 25   // Max fields per embed
	DiscordMaxFieldNameLen  = 256  // Max field name length
	DiscordMaxFieldValueLen = 1024 // Max field value length
	DiscordMaxDescLen       = 4096 // Max embed description length
	DiscordMaxTitleLen      = 256  // Max embed title length
	DiscordMaxContentLen    = 2000 // Max content length
)

// Table name functions
func (DiscordIntegration) TableName() string {
	return "discord_integrations"
}

func (DiscordChannelSubscription) TableName() string {
	return "discord_channel_subscriptions"
}

func (DiscordNotification) TableName() string {
	return "discord_notifications"
}

// Validation methods
func (d *DiscordIntegration) Validate() error {
	if d.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}
	if d.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if !isValidDiscordWebhookURL(d.WebhookURL) {
		return fmt.Errorf("invalid Discord webhook URL format")
	}
	if d.RetryCount < 0 || d.RetryCount > 10 {
		return fmt.Errorf("retry count must be between 0 and 10")
	}
	if d.RetryIntervalSeconds < 1 || d.RetryIntervalSeconds > 60 {
		return fmt.Errorf("retry interval must be between 1 and 60 seconds")
	}
	if d.CustomColor != "" && !isValidHexColor(d.CustomColor) {
		return fmt.Errorf("custom color must be a valid hex color (e.g., #5865F2)")
	}
	return nil
}

func (s *DiscordChannelSubscription) Validate() error {
	if s.IntegrationID == 0 {
		return fmt.Errorf("integration ID is required")
	}
	return nil
}

func (n *DiscordNotification) Validate() error {
	if n.IntegrationID == 0 {
		return fmt.Errorf("integration ID is required")
	}
	if n.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	validEvents := []string{DiscordEventDown, DiscordEventUp, DiscordEventDegraded, DiscordEventMaintenance}
	isValid := false
	for _, validEvent := range validEvents {
		if n.EventType == validEvent {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid event type: %s", n.EventType)
	}
	return nil
}

// Helper functions
func isValidDiscordWebhookURL(url string) bool {
	// Discord webhooks follow pattern: https://discord.com/api/webhooks/{webhook.id}/{webhook.token}
	// or https://discordapp.com/api/webhooks/{webhook.id}/{webhook.token}
	if len(url) == 0 {
		return false
	}
	return len(url) > 40 && (
		len(url) > 7 && url[:7] == "https://" || len(url) > 6 && url[:6] == "http://") &&
		(contains(url, "discord.com/api/webhooks/") || contains(url, "discordapp.com/api/webhooks/"))
}

func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper methods for building Discord messages
func (d *DiscordIntegration) GetColorForEventType(eventType string) int {
	switch eventType {
	case DiscordEventUp:
		return DiscordColorGreen
	case DiscordEventDown:
		return DiscordColorRed
	case DiscordEventDegraded:
		return DiscordColorOrange
	case DiscordEventMaintenance:
		return DiscordColorBlue
	default:
		return DiscordColorBlurple
	}
}

// GetCustomColorDecimal converts hex color to decimal if set, otherwise returns default for event type
func (d *DiscordIntegration) GetCustomColorDecimal(eventType string) int {
	if d.CustomColor != "" && isValidHexColor(d.CustomColor) {
		// Convert hex to decimal
		var r, g, b int
		fmt.Sscanf(d.CustomColor, "#%02x%02x%02x", &r, &g, &b)
		return (r << 16) + (g << 8) + b
	}
	return d.GetColorForEventType(eventType)
}

// ShouldNotify checks if notification should be sent for given event type
func (d *DiscordIntegration) ShouldNotify(eventType string) bool {
	switch eventType {
	case DiscordEventDown:
		return d.NotifyOnDown
	case DiscordEventUp:
		return d.NotifyOnUp
	case DiscordEventDegraded:
		return d.NotifyOnDegraded
	case DiscordEventMaintenance:
		return d.NotifyOnMaintenance
	default:
		return false
	}
}

// ShouldNotify checks if subscription should trigger notification for given event type
func (s *DiscordChannelSubscription) ShouldNotify(eventType string) bool {
	if !s.IsActive {
		return false
	}
	switch eventType {
	case DiscordEventDown:
		return s.NotifyOnDown
	case DiscordEventUp:
		return s.NotifyOnUp
	case DiscordEventDegraded:
		return s.NotifyOnDegraded
	case DiscordEventMaintenance:
		return s.NotifyOnMaintenance
	default:
		return false
	}
}
