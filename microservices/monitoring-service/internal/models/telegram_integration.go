package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TelegramIntegration represents a Telegram bot configuration for a tenant
type TelegramIntegration struct {
	ID       uint      `gorm:"primarykey" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`

	// Bot configuration
	BotToken   string `gorm:"type:text;not null" json:"bot_token"` // Telegram Bot API token
	BotUsername string `gorm:"size:255" json:"bot_username"`         // Bot username (e.g., @MyStatusBot)
	BotName     string `gorm:"size:255" json:"bot_name"`             // Display name for the bot

	// Default chat configuration
	DefaultChatID   string `gorm:"size:255" json:"default_chat_id"`   // Default Telegram chat ID
	DefaultChatName string `gorm:"size:255" json:"default_chat_name"` // Chat name/title
	DefaultChatType string `gorm:"size:50" json:"default_chat_type"`  // 'private', 'group', 'supergroup', 'channel'

	// Status
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`

	// Notification preferences
	NotifyOnDown        bool `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool `gorm:"default:false" json:"notify_on_maintenance"`

	// Message formatting
	UseMarkdown          bool `gorm:"default:true" json:"use_markdown"`           // Enable Telegram MarkdownV2
	IncludeMonitorURL    bool `gorm:"default:true" json:"include_monitor_url"`    // Include clickable URLs
	IncludeTimestamp     bool `gorm:"default:true" json:"include_timestamp"`      // Include timestamp
	SilentNotifications  bool `gorm:"default:false" json:"silent_notifications"`  // Send silently (no sound)
	DisablePreview       bool `gorm:"default:false" json:"disable_preview"`       // Disable link previews

	// Retry configuration
	RetryCount           int `gorm:"default:3" json:"retry_count"`
	RetryIntervalSeconds int `gorm:"default:5" json:"retry_interval_seconds"`

	// Metadata
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for TelegramIntegration
func (TelegramIntegration) TableName() string {
	return "telegram_integrations"
}

// Validate validates the Telegram integration configuration
func (t *TelegramIntegration) Validate() error {
	if t.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID is required")
	}

	if !isValidTelegramBotToken(t.BotToken) {
		return fmt.Errorf("invalid Telegram bot token format (expected: 123456789:ABCdefGHIjklMNOpqrsTUVwxyz)")
	}

	if t.DefaultChatID != "" && !isValidTelegramChatID(t.DefaultChatID) {
		return fmt.Errorf("invalid Telegram chat ID format")
	}

	if t.RetryCount < 0 || t.RetryCount > 10 {
		return fmt.Errorf("retry_count must be between 0 and 10")
	}

	if t.RetryIntervalSeconds < 1 || t.RetryIntervalSeconds > 300 {
		return fmt.Errorf("retry_interval_seconds must be between 1 and 300")
	}

	return nil
}

// TelegramChatSubscription represents a monitor-to-chat mapping
type TelegramChatSubscription struct {
	ID            uint  `gorm:"primarykey" json:"id"`
	IntegrationID uint  `gorm:"not null;index" json:"integration_id"`
	MonitorID     *uint `gorm:"index" json:"monitor_id,omitempty"` // NULL means all monitors

	// Chat override (if different from default)
	ChatID   string `gorm:"size:255" json:"chat_id,omitempty"`
	ChatName string `gorm:"size:255" json:"chat_name,omitempty"`
	ChatType string `gorm:"size:50" json:"chat_type,omitempty"` // 'private', 'group', 'supergroup', 'channel'

	// Event filtering per subscription
	NotifyOnDown        bool `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp          bool `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded    bool `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool `gorm:"default:false" json:"notify_on_maintenance"`

	// Message customization per subscription
	MessageThreadID     *int   `json:"message_thread_id,omitempty"` // Telegram forum topic/thread ID
	CustomMessagePrefix string `gorm:"type:text" json:"custom_message_prefix,omitempty"`

	// Status
	IsActive bool `gorm:"default:true" json:"is_active"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Integration TelegramIntegration `gorm:"foreignKey:IntegrationID" json:"-"`
}

// TableName specifies the table name for TelegramChatSubscription
func (TelegramChatSubscription) TableName() string {
	return "telegram_chat_subscriptions"
}

// TelegramNotification represents a Telegram notification delivery record
type TelegramNotification struct {
	ID            uint  `gorm:"primarykey" json:"id"`
	IntegrationID uint  `gorm:"not null;index" json:"integration_id"`
	MonitorID     *uint `gorm:"index" json:"monitor_id,omitempty"`

	// Event details
	EventType  string `gorm:"size:50;not null;index" json:"event_type"` // 'down', 'up', 'degraded', 'maintenance'
	MonitorName string `gorm:"size:255" json:"monitor_name,omitempty"`
	MonitorURL  string `gorm:"type:text" json:"monitor_url,omitempty"`

	// Delivery status
	Status         string `gorm:"size:50;not null;index" json:"status"` // 'sent', 'failed', 'pending', 'retrying'
	HTTPStatusCode *int   `json:"http_status_code,omitempty"`

	// Message content
	MessageText string `gorm:"type:text" json:"message_text,omitempty"`
	ParseMode   string `gorm:"size:50" json:"parse_mode,omitempty"` // 'MarkdownV2', 'HTML', 'Markdown', NULL

	// Delivery tracking
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`

	// Error tracking
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`
	ErrorCode    string `gorm:"size:50" json:"error_code,omitempty"`

	// Telegram response
	TelegramMessageID int64  `json:"telegram_message_id,omitempty"` // Telegram's message ID
	TelegramChatID    string `gorm:"size:255;index" json:"telegram_chat_id,omitempty"`
	TelegramChatType  string `gorm:"size:50" json:"telegram_chat_type,omitempty"`

	// Metadata
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Integration TelegramIntegration `gorm:"foreignKey:IntegrationID" json:"-"`
}

// TableName specifies the table name for TelegramNotification
func (TelegramNotification) TableName() string {
	return "telegram_notifications"
}

// TelegramSendMessageRequest represents a Telegram sendMessage API request
type TelegramSendMessageRequest struct {
	ChatID                string  `json:"chat_id"`                            // Unique identifier for the target chat
	Text                  string  `json:"text"`                               // Text of the message to be sent
	ParseMode             string  `json:"parse_mode,omitempty"`               // 'MarkdownV2', 'HTML', or 'Markdown'
	DisableWebPagePreview bool    `json:"disable_web_page_preview,omitempty"` // Disables link previews
	DisableNotification   bool    `json:"disable_notification,omitempty"`     // Sends silently
	MessageThreadID       *int    `json:"message_thread_id,omitempty"`        // Forum topic ID
	ReplyMarkup           *string `json:"reply_markup,omitempty"`             // Inline keyboard markup
}

// TelegramSendMessageResponse represents a Telegram sendMessage API response
type TelegramSendMessageResponse struct {
	OK          bool                   `json:"ok"`
	Result      *TelegramMessage       `json:"result,omitempty"`
	Description string                 `json:"description,omitempty"`
	ErrorCode   int                    `json:"error_code,omitempty"`
	Parameters  *TelegramResponseParams `json:"parameters,omitempty"`
}

// TelegramMessage represents a Telegram message object
type TelegramMessage struct {
	MessageID int64        `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      TelegramChat `json:"chat"`
	Date      int64        `json:"date"`
	Text      string       `json:"text,omitempty"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // 'private', 'group', 'supergroup', 'channel'
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// TelegramResponseParams contains information about why a request was unsuccessful
type TelegramResponseParams struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int   `json:"retry_after,omitempty"`
}

// TelegramGetMeResponse represents the response from the getMe API call
type TelegramGetMeResponse struct {
	OK          bool          `json:"ok"`
	Result      *TelegramUser `json:"result,omitempty"`
	Description string        `json:"description,omitempty"`
	ErrorCode   int           `json:"error_code,omitempty"`
}

// Helper functions

// isValidTelegramBotToken validates Telegram bot token format
// Format: {bot_id}:{bot_token} (e.g., 123456789:ABCdefGHIjklMNOpqrsTUVwxyz)
func isValidTelegramBotToken(token string) bool {
	if token == "" {
		return false
	}
	// Telegram bot token format: digits:alphanumeric (usually 46 chars total)
	pattern := `^\d{8,10}:[A-Za-z0-9_-]{35}$`
	matched, _ := regexp.MatchString(pattern, token)
	return matched
}

// isValidTelegramChatID validates Telegram chat ID format
// Chat IDs can be:
// - Positive integers for private chats (user IDs)
// - Negative integers for groups/supergroups/channels
// - Format: -100{channel_id} for channels/supergroups
func isValidTelegramChatID(chatID string) bool {
	if chatID == "" {
		return false
	}
	// Match positive or negative integers
	pattern := `^-?\d+$`
	matched, _ := regexp.MatchString(pattern, chatID)
	return matched
}

// FormatTelegramMarkdownV2 escapes special characters for Telegram MarkdownV2
// MarkdownV2 requires escaping: _ * [ ] ( ) ~ ` > # + - = | { } . !
func FormatTelegramMarkdownV2(text string) string {
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
	}

	result := text
	for _, char := range specialChars {
		result = regexp.MustCompile(regexp.QuoteMeta(char)).ReplaceAllString(result, "\\"+char)
	}
	return result
}

// GetEventEmoji returns the emoji for a given event type
func GetEventEmoji(eventType string) string {
	switch eventType {
	case "down":
		return "🔴"
	case "up":
		return "🟢"
	case "degraded":
		return "🟡"
	case "maintenance":
		return "🔵"
	default:
		return "ℹ️"
	}
}

// BuildTelegramMessage formats a monitor alert message for Telegram
func BuildTelegramMessage(
	eventType string,
	monitorName string,
	monitorURL string,
	details map[string]interface{},
	useMarkdown bool,
	includeURL bool,
	includeTimestamp bool,
) string {
	emoji := GetEventEmoji(eventType)
	title := fmt.Sprintf("%s Monitor %s: %s", emoji, getEventAction(eventType), monitorName)

	var message string
	if useMarkdown {
		// Use MarkdownV2 formatting
		title = FormatTelegramMarkdownV2(title)
		message = fmt.Sprintf("*%s*\n\n", title)

		// Add description
		description := getEventDescription(eventType)
		message += FormatTelegramMarkdownV2(description) + "\n\n"

		// Add details
		if len(details) > 0 {
			message += "*Details:*\n"
			for key, value := range details {
				keyFormatted := FormatTelegramMarkdownV2(fmt.Sprintf("%s", key))
				valueFormatted := FormatTelegramMarkdownV2(fmt.Sprintf("%v", value))
				message += fmt.Sprintf("• %s: %s\n", keyFormatted, valueFormatted)
			}
			message += "\n"
		}

		// Add monitor URL
		if includeURL && monitorURL != "" {
			message += fmt.Sprintf("[View Dashboard](%s)\n\n", monitorURL)
		}

		// Add timestamp
		if includeTimestamp {
			timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
			message += FormatTelegramMarkdownV2(fmt.Sprintf("🕐 %s", timestamp))
		}
	} else {
		// Plain text formatting
		message = fmt.Sprintf("%s\n\n", title)
		message += getEventDescription(eventType) + "\n\n"

		if len(details) > 0 {
			message += "Details:\n"
			for key, value := range details {
				message += fmt.Sprintf("• %s: %v\n", key, value)
			}
			message += "\n"
		}

		if includeURL && monitorURL != "" {
			message += fmt.Sprintf("View Dashboard: %s\n\n", monitorURL)
		}

		if includeTimestamp {
			timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
			message += fmt.Sprintf("🕐 %s", timestamp)
		}
	}

	return message
}

// getEventAction returns the action verb for event type
func getEventAction(eventType string) string {
	switch eventType {
	case "down":
		return "Down"
	case "up":
		return "Recovered"
	case "degraded":
		return "Degraded"
	case "maintenance":
		return "Maintenance"
	default:
		return "Alert"
	}
}

// getEventDescription returns the description for event type
func getEventDescription(eventType string) string {
	switch eventType {
	case "down":
		return "The monitor has gone down and is currently unreachable."
	case "up":
		return "The monitor has recovered and is now operational."
	case "degraded":
		return "The monitor is experiencing performance degradation."
	case "maintenance":
		return "Scheduled maintenance has started for this monitor."
	default:
		return "Monitor status has changed."
	}
}
