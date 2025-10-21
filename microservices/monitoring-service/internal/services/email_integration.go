package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EmailIntegration represents an email notification configuration
type EmailIntegration struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	TenantID          uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name              string    `gorm:"size:255;not null" json:"name"`
	SMTPHost          string    `gorm:"size:255;not null" json:"smtp_host"`
	SMTPPort          int       `gorm:"not null" json:"smtp_port"`
	SMTPUsername      string    `gorm:"size:255;not null" json:"smtp_username"`
	SMTPPassword      string    `gorm:"size:255;not null" json:"smtp_password"` // Should be encrypted in production
	FromEmail         string    `gorm:"size:255;not null" json:"from_email"`
	FromName          string    `gorm:"size:255" json:"from_name"`
	UseTLS            bool      `gorm:"default:true" json:"use_tls"`
	IsActive          bool      `gorm:"default:true" json:"is_active"`
	NotifyOnDown      bool      `gorm:"default:true" json:"notify_on_down"`
	NotifyOnUp        bool      `gorm:"default:true" json:"notify_on_up"`
	NotifyOnDegraded  bool      `gorm:"default:true" json:"notify_on_degraded"`
	NotifyOnMaintenance bool    `gorm:"default:false" json:"notify_on_maintenance"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// EmailSubscriber represents an email subscriber for notifications
type EmailSubscriber struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	IntegrationID   uint      `gorm:"not null;index" json:"integration_id"`
	Email           string    `gorm:"size:255;not null" json:"email"`
	Name            string    `gorm:"size:255" json:"name"`
	MonitorIDs      string    `gorm:"type:text" json:"monitor_ids"` // Comma-separated monitor IDs, empty for all
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	VerifiedAt      *time.Time `json:"verified_at"`
	UnsubscribedAt  *time.Time `json:"unsubscribed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// EmailNotification represents a sent email notification
type EmailNotification struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IntegrationID  uint      `gorm:"not null;index" json:"integration_id"`
	MonitorID      uint      `gorm:"not null;index" json:"monitor_id"`
	RecipientEmail string    `gorm:"size:255;not null" json:"recipient_email"`
	Subject        string    `gorm:"type:text" json:"subject"`
	Body           string    `gorm:"type:text" json:"body"`
	EventType      string    `gorm:"size:50;not null" json:"event_type"` // down, up, degraded, maintenance
	Status         string    `gorm:"size:50" json:"status"` // sent, failed, pending
	ErrorMessage   string    `gorm:"type:text" json:"error_message"`
	SentAt         time.Time `json:"sent_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// EmailTemplate represents email template data
type EmailTemplate struct {
	MonitorName   string
	EventType     string
	StatusText    string
	StatusColor   string
	ResponseTime  int
	StatusCode    int
	Error         string
	Location      string
	Timestamp     string
	MonitorURL    string
	TenantName    string
	UnsubscribeURL string
}

// EmailIntegrationService manages email integrations and notifications
type EmailIntegrationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewEmailIntegrationService creates a new email integration service
func NewEmailIntegrationService(db *gorm.DB, logger *zap.Logger) *EmailIntegrationService {
	return &EmailIntegrationService{
		db:     db,
		logger: logger,
	}
}

// CreateIntegration creates a new email integration
func (s *EmailIntegrationService) CreateIntegration(integration *EmailIntegration) error {
	// Test SMTP connection
	err := s.TestSMTPConnection(integration)
	if err != nil {
		return fmt.Errorf("SMTP connection test failed: %w", err)
	}

	err = s.db.Create(integration).Error
	if err != nil {
		s.logger.Error("Failed to create email integration",
			zap.Error(err),
			zap.String("name", integration.Name))
		return fmt.Errorf("failed to create integration: %w", err)
	}

	s.logger.Info("Email integration created",
		zap.Uint("id", integration.ID),
		zap.String("name", integration.Name))

	return nil
}

// UpdateIntegration updates an existing email integration
func (s *EmailIntegrationService) UpdateIntegration(integration *EmailIntegration) error {
	err := s.db.Save(integration).Error
	if err != nil {
		s.logger.Error("Failed to update email integration", zap.Error(err))
		return fmt.Errorf("failed to update integration: %w", err)
	}

	s.logger.Info("Email integration updated", zap.Uint("id", integration.ID))
	return nil
}

// GetIntegration retrieves an email integration by ID
func (s *EmailIntegrationService) GetIntegration(id uint, tenantID uuid.UUID) (*EmailIntegration, error) {
	var integration EmailIntegration
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&integration).Error
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}
	return &integration, nil
}

// GetIntegrationsByTenant retrieves all integrations for a tenant
func (s *EmailIntegrationService) GetIntegrationsByTenant(tenantID uuid.UUID) ([]EmailIntegration, error) {
	var integrations []EmailIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}
	return integrations, nil
}

// DeleteIntegration deletes an email integration
func (s *EmailIntegrationService) DeleteIntegration(id uint, tenantID uuid.UUID) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&EmailIntegration{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete integration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	s.logger.Info("Email integration deleted", zap.Uint("id", id))
	return nil
}

// AddSubscriber adds an email subscriber
func (s *EmailIntegrationService) AddSubscriber(subscriber *EmailSubscriber) error {
	err := s.db.Create(subscriber).Error
	if err != nil {
		s.logger.Error("Failed to add subscriber", zap.Error(err))
		return fmt.Errorf("failed to add subscriber: %w", err)
	}

	s.logger.Info("Subscriber added",
		zap.Uint("integration_id", subscriber.IntegrationID),
		zap.String("email", subscriber.Email))

	// TODO: Send verification email
	return nil
}

// RemoveSubscriber removes an email subscriber
func (s *EmailIntegrationService) RemoveSubscriber(id uint) error {
	now := time.Now()
	result := s.db.Model(&EmailSubscriber{}).Where("id = ?", id).Updates(map[string]interface{}{
		"unsubscribed_at": now,
		"is_active":       false,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to remove subscriber: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscriber not found")
	}

	s.logger.Info("Subscriber removed", zap.Uint("id", id))
	return nil
}

// GetSubscribers retrieves all subscribers for an integration
func (s *EmailIntegrationService) GetSubscribers(integrationID uint) ([]EmailSubscriber, error) {
	var subscribers []EmailSubscriber
	err := s.db.Where("integration_id = ? AND is_active = ?", integrationID, true).
		Find(&subscribers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers: %w", err)
	}
	return subscribers, nil
}

// SendMonitorAlert sends a monitor alert email
func (s *EmailIntegrationService) SendMonitorAlert(monitorID uint, tenantID uuid.UUID, eventType string, monitorName string, monitorURL string, details map[string]interface{}) error {
	// Get active integrations
	integrations, err := s.GetIntegrationsByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	// Filter active integrations based on event type
	var activeIntegrations []EmailIntegration
	for _, integration := range integrations {
		if !integration.IsActive {
			continue
		}

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

		if shouldNotify {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}

	if len(activeIntegrations) == 0 {
		s.logger.Info("No active email integrations for event type",
			zap.String("event_type", eventType))
		return nil
	}

	// Send emails for each active integration
	for _, integration := range activeIntegrations {
		// Get subscribers
		subscribers, err := s.GetSubscribers(integration.ID)
		if err != nil {
			s.logger.Error("Failed to get subscribers", zap.Error(err))
			continue
		}

		// Send to each subscriber
		for _, subscriber := range subscribers {
			// Check if subscriber is subscribed to this monitor
			if subscriber.MonitorIDs != "" {
				// TODO: Parse comma-separated monitor IDs and check if monitorID is in the list
			}

			err := s.sendEmail(&integration, subscriber.Email, subscriber.Name, monitorName, eventType, monitorURL, details)
			if err != nil {
				s.logger.Error("Failed to send email",
					zap.Error(err),
					zap.String("recipient", subscriber.Email))
			}
		}
	}

	return nil
}

// sendEmail sends an email using SMTP
func (s *EmailIntegrationService) sendEmail(integration *EmailIntegration, recipientEmail, recipientName, monitorName, eventType, monitorURL string, details map[string]interface{}) error {
	// Build email content
	subject := s.buildSubject(monitorName, eventType)
	body := s.buildHTMLBody(monitorName, eventType, monitorURL, details)

	// Prepare notification record
	notification := EmailNotification{
		IntegrationID:  integration.ID,
		MonitorID:      0, // Should be passed as parameter
		RecipientEmail: recipientEmail,
		Subject:        subject,
		Body:           body,
		EventType:      eventType,
		Status:         "pending",
	}

	// Build email message
	from := integration.FromEmail
	if integration.FromName != "" {
		from = fmt.Sprintf("%s <%s>", integration.FromName, integration.FromEmail)
	}

	to := recipientEmail
	if recipientName != "" {
		to = fmt.Sprintf("%s <%s>", recipientName, recipientEmail)
	}

	// Construct email headers and body
	message := s.constructEmail(from, to, subject, body)

	// Send via SMTP
	auth := smtp.PlainAuth("", integration.SMTPUsername, integration.SMTPPassword, integration.SMTPHost)
	addr := fmt.Sprintf("%s:%d", integration.SMTPHost, integration.SMTPPort)

	var err error
	if integration.UseTLS {
		// Use TLS
		tlsConfig := &tls.Config{
			ServerName:         integration.SMTPHost,
			InsecureSkipVerify: false,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, integration.SMTPHost)
		if err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("SMTP client creation failed: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("SMTP auth failed: %w", err)
		}

		if err = client.Mail(integration.FromEmail); err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("MAIL command failed: %w", err)
		}

		if err = client.Rcpt(recipientEmail); err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("RCPT command failed: %w", err)
		}

		writer, err := client.Data()
		if err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("DATA command failed: %w", err)
		}

		_, err = writer.Write([]byte(message))
		if err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("Write failed: %w", err)
		}

		err = writer.Close()
		if err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("Close failed: %w", err)
		}
	} else {
		// Use standard SMTP
		err = smtp.SendMail(addr, auth, integration.FromEmail, []string{recipientEmail}, []byte(message))
		if err != nil {
			notification.Status = "failed"
			notification.ErrorMessage = err.Error()
			s.db.Create(&notification)
			return fmt.Errorf("send mail failed: %w", err)
		}
	}

	// Mark as sent
	notification.Status = "sent"
	notification.SentAt = time.Now()
	s.db.Create(&notification)

	s.logger.Info("Email sent successfully",
		zap.Uint("integration_id", integration.ID),
		zap.String("recipient", recipientEmail),
		zap.String("event_type", eventType))

	return nil
}

// buildSubject builds the email subject line
func (s *EmailIntegrationService) buildSubject(monitorName, eventType string) string {
	switch eventType {
	case "down":
		return fmt.Sprintf("🔴 ALERT: %s is DOWN", monitorName)
	case "up":
		return fmt.Sprintf("✅ RECOVERED: %s is back UP", monitorName)
	case "degraded":
		return fmt.Sprintf("⚠️ WARNING: %s is DEGRADED", monitorName)
	case "maintenance":
		return fmt.Sprintf("🔧 MAINTENANCE: %s maintenance scheduled", monitorName)
	default:
		return fmt.Sprintf("Notification: %s", monitorName)
	}
}

// buildHTMLBody builds the email HTML body
func (s *EmailIntegrationService) buildHTMLBody(monitorName, eventType, monitorURL string, details map[string]interface{}) string {
	// Prepare template data
	data := EmailTemplate{
		MonitorName: monitorName,
		EventType:   eventType,
		MonitorURL:  monitorURL,
		Timestamp:   time.Now().Format("January 2, 2006 at 3:04 PM MST"),
	}

	// Set status text and color
	switch eventType {
	case "down":
		data.StatusText = "DOWN"
		data.StatusColor = "#dc2626"
	case "up":
		data.StatusText = "OPERATIONAL"
		data.StatusColor = "#16a34a"
	case "degraded":
		data.StatusText = "DEGRADED"
		data.StatusColor = "#f59e0b"
	case "maintenance":
		data.StatusText = "MAINTENANCE"
		data.StatusColor = "#6b7280"
	}

	// Extract details
	if rt, ok := details["response_time"].(int); ok {
		data.ResponseTime = rt
	}
	if sc, ok := details["status_code"].(int); ok {
		data.StatusCode = sc
	}
	if err, ok := details["error"].(string); ok {
		data.Error = err
	}
	if loc, ok := details["location"].(string); ok {
		data.Location = loc
	}

	// HTML template
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: {{.StatusColor}}; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background-color: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .status-badge { display: inline-block; padding: 8px 16px; background-color: {{.StatusColor}}; color: white; border-radius: 4px; font-weight: bold; }
        .detail-row { margin: 15px 0; padding: 10px; background-color: white; border-radius: 4px; }
        .detail-label { font-weight: bold; color: #6b7280; }
        .button { display: inline-block; padding: 12px 24px; background-color: #3b82f6; color: white; text-decoration: none; border-radius: 4px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">Monitor Status Alert</h1>
        </div>
        <div class="content">
            <p><strong>Monitor:</strong> {{.MonitorName}}</p>
            <p><strong>Status:</strong> <span class="status-badge">{{.StatusText}}</span></p>

            <div class="detail-row">
                <div class="detail-label">Timestamp</div>
                <div>{{.Timestamp}}</div>
            </div>

            {{if .Location}}
            <div class="detail-row">
                <div class="detail-label">Location</div>
                <div>{{.Location}}</div>
            </div>
            {{end}}

            {{if .ResponseTime}}
            <div class="detail-row">
                <div class="detail-label">Response Time</div>
                <div>{{.ResponseTime}} ms</div>
            </div>
            {{end}}

            {{if .StatusCode}}
            <div class="detail-row">
                <div class="detail-label">Status Code</div>
                <div>{{.StatusCode}}</div>
            </div>
            {{end}}

            {{if .Error}}
            <div class="detail-row">
                <div class="detail-label">Error</div>
                <div>{{.Error}}</div>
            </div>
            {{end}}

            {{if .MonitorURL}}
            <a href="{{.MonitorURL}}" class="button">View Monitor Details</a>
            {{end}}
        </div>
        <div class="footer">
            <p>This is an automated notification from Beakon Status Page</p>
            <p><a href="{{.UnsubscribeURL}}" style="color: #6b7280;">Unsubscribe from these notifications</a></p>
        </div>
    </div>
</body>
</html>
`

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		s.logger.Error("Failed to parse email template", zap.Error(err))
		return "Failed to generate email"
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		s.logger.Error("Failed to execute email template", zap.Error(err))
		return "Failed to generate email"
	}

	return buf.String()
}

// constructEmail constructs the full email message with headers
func (s *EmailIntegrationService) constructEmail(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return message
}

// TestSMTPConnection tests the SMTP connection
func (s *EmailIntegrationService) TestSMTPConnection(integration *EmailIntegration) error {
	auth := smtp.PlainAuth("", integration.SMTPUsername, integration.SMTPPassword, integration.SMTPHost)
	addr := fmt.Sprintf("%s:%d", integration.SMTPHost, integration.SMTPPort)

	if integration.UseTLS {
		tlsConfig := &tls.Config{
			ServerName:         integration.SMTPHost,
			InsecureSkipVerify: false,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, integration.SMTPHost)
		if err != nil {
			return fmt.Errorf("SMTP client creation failed: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}

		return nil
	}

	// Test connection without TLS
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP dial failed: %w", err)
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	return nil
}

// GetNotificationHistory retrieves notification history
func (s *EmailIntegrationService) GetNotificationHistory(monitorID uint, limit int) ([]EmailNotification, error) {
	var notifications []EmailNotification
	err := s.db.Where("monitor_id = ?", monitorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get notification history: %w", err)
	}

	return notifications, nil
}

// GetNotificationStats retrieves notification statistics
func (s *EmailIntegrationService) GetNotificationStats(tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var integrations []EmailIntegration
	err := s.db.Where("tenant_id = ?", tenantID).Find(&integrations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(integrations) == 0 {
		return map[string]interface{}{
			"total_sent":   0,
			"total_failed": 0,
			"success_rate": 0.0,
		}, nil
	}

	integrationIDs := make([]uint, len(integrations))
	for i, integration := range integrations {
		integrationIDs[i] = integration.ID
	}

	// Count sent
	var totalSent int64
	s.db.Model(&EmailNotification{}).
		Where("integration_id IN ? AND status = ? AND sent_at BETWEEN ? AND ?", integrationIDs, "sent", startDate, endDate).
		Count(&totalSent)

	// Count failed
	var totalFailed int64
	s.db.Model(&EmailNotification{}).
		Where("integration_id IN ? AND status = ? AND created_at BETWEEN ? AND ?", integrationIDs, "failed", startDate, endDate).
		Count(&totalFailed)

	successRate := 0.0
	total := totalSent + totalFailed
	if total > 0 {
		successRate = float64(totalSent) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_sent":   totalSent,
		"total_failed": totalFailed,
		"success_rate": successRate,
	}, nil
}
