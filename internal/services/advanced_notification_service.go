package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/email"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AdvancedNotificationService struct {
	db          *gorm.DB
	emailSender email.Sender
	httpClient  *http.Client
}

func NewAdvancedNotificationService(emailSender email.Sender) *AdvancedNotificationService {
	return &AdvancedNotificationService{
		db:          database.DB,
		emailSender: emailSender,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Email Templates

type EmailTemplate struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
	Type      string `gorm:"not null"` // incident, maintenance, status_update
	Subject   string `gorm:"not null"`
	HTMLBody  string `gorm:"not null"`
	TextBody  string
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EmailNotification struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"not null"`
	TemplateID uint   `gorm:"not null"`
	Recipient  string `gorm:"not null"`
	Subject    string `gorm:"not null"`
	Body       string `gorm:"not null"`
	Status     string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt     *time.Time
	Error      string
	CreatedAt  time.Time
}

// SMS Integration

type SMSProvider struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"not null"`
	Name       string `gorm:"not null"`
	Type       string `gorm:"not null"` // twilio, aws_sns, custom
	APIKey     string `gorm:"not null"`
	APISecret  string
	FromNumber string
	IsActive   bool `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SMSNotification struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"not null"`
	ProviderID uint   `gorm:"not null"`
	ToNumber   string `gorm:"not null"`
	Message    string `gorm:"not null"`
	Status     string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt     *time.Time
	Error      string
	CreatedAt  time.Time
}

// Push Notifications

type PushNotification struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	UserID    uint   `gorm:"not null"`
	Title     string `gorm:"not null"`
	Body      string `gorm:"not null"`
	Data      string // JSON data
	Status    string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt    *time.Time
	Error     string
	CreatedAt time.Time
}

// Webhook Notifications

type WebhookNotification struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"not null"`
	URL       string `gorm:"not null"`
	EventType string `gorm:"not null"`
	Payload   string `gorm:"not null"`          // JSON payload
	Status    string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt    *time.Time
	Error     string
	CreatedAt time.Time
}

// Main notification methods

func (s *AdvancedNotificationService) SendIncidentNotification(tenantID uint, incident *models.Incident, notificationType string) error {
	// Get subscribers for the tenant
	var subscribers []models.Subscriber
	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&subscribers).Error; err != nil {
		return err
	}

	// Get notification preferences
	for _, subscriber := range subscribers {
		preferences, err := s.getNotificationPreferences(subscriber.ID)
		if err != nil {
			logger.Error("Failed to get notification preferences", zap.Error(err))
			continue
		}

		// Send email notification
		if preferences.EmailEnabled && notificationType == "email" {
			if err := s.sendEmailNotification(tenantID, &subscriber, incident, "incident"); err != nil {
				logger.Error("Failed to send email notification", zap.Error(err))
			}
		}

		// Send SMS notification
		if preferences.SMSEnabled && notificationType == "sms" {
			if err := s.sendSMSNotification(tenantID, &subscriber, incident); err != nil {
				logger.Error("Failed to send SMS notification", zap.Error(err))
			}
		}

		// Send push notification
		if preferences.PushEnabled && notificationType == "push" {
			if err := s.sendPushNotification(tenantID, &subscriber, incident); err != nil {
				logger.Error("Failed to send push notification", zap.Error(err))
			}
		}
	}

	return nil
}

func (s *AdvancedNotificationService) SendMaintenanceNotification(tenantID uint, maintenance *models.Maintenance, notificationType string) error {
	// Similar to incident notification but for maintenance events
	var subscribers []models.Subscriber
	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&subscribers).Error; err != nil {
		return err
	}

	for _, subscriber := range subscribers {
		preferences, err := s.getNotificationPreferences(subscriber.ID)
		if err != nil {
			continue
		}

		if preferences.EmailEnabled && notificationType == "email" {
			if err := s.sendEmailNotification(tenantID, &subscriber, maintenance, "maintenance"); err != nil {
				logger.Error("Failed to send email notification", zap.Error(err))
			}
		}
	}

	return nil
}

// Email notification methods

func (s *AdvancedNotificationService) sendEmailNotification(tenantID uint, subscriber *models.Subscriber, data interface{}, templateType string) error {
	// Get email template
	var template EmailTemplate
	if err := s.db.Where("tenant_id = ? AND type = ? AND is_active = ?",
		tenantID, templateType, true).First(&template).Error; err != nil {
		// Use default template if none found
		template = s.getDefaultEmailTemplate(templateType)
	}

	// Render template
	subject, body, err := s.renderEmailTemplate(&template, data, subscriber)
	if err != nil {
		return err
	}

	// Send email
	if err := s.emailSender.Send(subscriber.Email, subject, body); err != nil {
		// Log failed email
		s.logEmailNotification(tenantID, template.ID, subscriber.Email, subject, body, "failed", err.Error())
		return err
	}

	// Log successful email
	s.logEmailNotification(tenantID, template.ID, subscriber.Email, subject, body, "sent", "")
	return nil
}

func (s *AdvancedNotificationService) renderEmailTemplate(emailTemplate *EmailTemplate, data interface{}, subscriber *models.Subscriber) (string, string, error) {
	// Create template data
	templateData := map[string]interface{}{
		"Subscriber": subscriber,
		"Data":       data,
		"Timestamp":  time.Now(),
	}

	// Render subject
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return "", "", err
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, templateData); err != nil {
		return "", "", err
	}

	// Render body
	bodyTmpl, err := template.New("body").Parse(emailTemplate.HTMLBody)
	if err != nil {
		return "", "", err
	}

	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, templateData); err != nil {
		return "", "", err
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// SMS notification methods

func (s *AdvancedNotificationService) sendSMSNotification(tenantID uint, subscriber *models.Subscriber, incident *models.Incident) error {
	// Get SMS provider
	var provider SMSProvider
	if err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).First(&provider).Error; err != nil {
		return err
	}

	// Create SMS message
	message := fmt.Sprintf("Alert: %s - %s", incident.Title, incident.Description)
	if len(message) > 160 {
		message = message[:157] + "..."
	}

	// Send SMS based on provider type
	switch provider.Type {
	case "twilio":
		return s.sendTwilioSMS(&provider, subscriber.Phone, message)
	case "aws_sns":
		return s.sendAWSSNS(&provider, subscriber.Phone, message)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", provider.Type)
	}
}

func (s *AdvancedNotificationService) sendTwilioSMS(provider *SMSProvider, toNumber, message string) error {
	// Twilio SMS implementation
	// This would use the Twilio Go SDK
	_ = provider // TODO: Use provider configuration for Twilio API credentials
	logger.Info("Sending Twilio SMS", zap.String("to", toNumber), zap.String("message", message))
	return nil
}

func (s *AdvancedNotificationService) sendAWSSNS(provider *SMSProvider, toNumber, message string) error {
	// AWS SNS SMS implementation
	// This would use the AWS Go SDK
	logger.Info("Sending AWS SNS SMS", zap.String("to", toNumber), zap.String("message", message))
	return nil
}

// Push notification methods

func (s *AdvancedNotificationService) sendPushNotification(tenantID uint, subscriber *models.Subscriber, incident *models.Incident) error {
	// Get user's push tokens
	pushTokens := []string{} // This would come from a user device tokens table

	for _, token := range pushTokens {
		notification := PushNotification{
			TenantID: tenantID,
			UserID:   subscriber.ID,
			Title:    incident.Title,
			Body:     incident.Description,
			Data:     fmt.Sprintf(`{"incident_id": %d, "type": "incident"}`, incident.ID),
			Status:   "pending",
		}

		if err := s.db.Create(&notification).Error; err != nil {
			logger.Error("Failed to create push notification", zap.Error(err))
			continue
		}

		// Send push notification (FCM, APNS, etc.)
		if err := s.sendFCMPushNotification(token, &notification); err != nil {
			logger.Error("Failed to send FCM push notification", zap.Error(err))
			notification.Status = "failed"
			notification.Error = err.Error()
		} else {
			notification.Status = "sent"
			now := time.Now()
			notification.SentAt = &now
		}

		s.db.Save(&notification)
	}

	return nil
}

func (s *AdvancedNotificationService) sendFCMPushNotification(token string, notification *PushNotification) error {
	// Firebase Cloud Messaging implementation
	// This would use the Firebase Go SDK
	logger.Info("Sending FCM push notification", zap.String("token", token))
	return nil
}

// Webhook notification methods

func (s *AdvancedNotificationService) SendWebhookNotification(tenantID uint, webhookURL string, eventType string, data interface{}) error {
	payload := map[string]interface{}{
		"event":     eventType,
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	webhookNotification := WebhookNotification{
		TenantID:  tenantID,
		URL:       webhookURL,
		EventType: eventType,
		Payload:   string(jsonData),
		Status:    "pending",
	}

	if err := s.db.Create(&webhookNotification).Error; err != nil {
		return err
	}

	// Send webhook
	resp, err := s.httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		webhookNotification.Status = "failed"
		webhookNotification.Error = err.Error()
		s.db.Save(&webhookNotification)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		webhookNotification.Status = "failed"
		webhookNotification.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		webhookNotification.Status = "sent"
		now := time.Now()
		webhookNotification.SentAt = &now
	}

	s.db.Save(&webhookNotification)
	return nil
}

// Template management

func (s *AdvancedNotificationService) CreateEmailTemplate(template *EmailTemplate) error {
	return s.db.Create(template).Error
}

func (s *AdvancedNotificationService) GetEmailTemplates(tenantID uint) ([]EmailTemplate, error) {
	var templates []EmailTemplate
	err := s.db.Where("tenant_id = ?", tenantID).Find(&templates).Error
	return templates, err
}

func (s *AdvancedNotificationService) UpdateEmailTemplate(template *EmailTemplate) error {
	return s.db.Save(template).Error
}

func (s *AdvancedNotificationService) DeleteEmailTemplate(templateID uint) error {
	return s.db.Delete(&EmailTemplate{}, templateID).Error
}

// Notification preferences

type NotificationPreferences struct {
	ID             uint `gorm:"primaryKey"`
	SubscriberID   uint `gorm:"not null"`
	EmailEnabled   bool `gorm:"default:true"`
	SMSEnabled     bool `gorm:"default:false"`
	PushEnabled    bool `gorm:"default:true"`
	WebhookEnabled bool `gorm:"default:false"`
	WebhookURL     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *AdvancedNotificationService) getNotificationPreferences(subscriberID uint) (*NotificationPreferences, error) {
	var preferences NotificationPreferences
	err := s.db.Where("subscriber_id = ?", subscriberID).First(&preferences).Error
	if err == gorm.ErrRecordNotFound {
		// Create default preferences
		preferences = NotificationPreferences{
			SubscriberID:   subscriberID,
			EmailEnabled:   true,
			SMSEnabled:     false,
			PushEnabled:    true,
			WebhookEnabled: false,
		}
		s.db.Create(&preferences)
		return &preferences, nil
	}
	return &preferences, err
}

// Helper methods

func (s *AdvancedNotificationService) logEmailNotification(tenantID, templateID uint, recipient, subject, body, status, errorMsg string) {
	notification := EmailNotification{
		TenantID:   tenantID,
		TemplateID: templateID,
		Recipient:  recipient,
		Subject:    subject,
		Body:       body,
		Status:     status,
		Error:      errorMsg,
	}

	if status == "sent" {
		now := time.Now()
		notification.SentAt = &now
	}

	s.db.Create(&notification)
}

func (s *AdvancedNotificationService) getDefaultEmailTemplate(templateType string) EmailTemplate {
	switch templateType {
	case "incident":
		return EmailTemplate{
			Subject: "🚨 Incident Alert: {{.Data.Title}}",
			HTMLBody: `
				<h2>Incident Alert</h2>
				<p><strong>Title:</strong> {{.Data.Title}}</p>
				<p><strong>Description:</strong> {{.Data.Description}}</p>
				<p><strong>Status:</strong> {{.Data.Status}}</p>
				<p><strong>Severity:</strong> {{.Data.Severity}}</p>
				<p><strong>Started:</strong> {{.Data.StartedAt}}</p>
				<p>Visit your status page for more details.</p>
			`,
		}
	case "maintenance":
		return EmailTemplate{
			Subject: "🔧 Scheduled Maintenance: {{.Data.Title}}",
			HTMLBody: `
				<h2>Scheduled Maintenance</h2>
				<p><strong>Title:</strong> {{.Data.Title}}</p>
				<p><strong>Description:</strong> {{.Data.Description}}</p>
				<p><strong>Scheduled Start:</strong> {{.Data.ScheduledStart}}</p>
				<p><strong>Scheduled End:</strong> {{.Data.ScheduledEnd}}</p>
				<p>Visit your status page for more details.</p>
			`,
		}
	default:
		return EmailTemplate{
			Subject:  "Status Page Notification",
			HTMLBody: "<p>You have received a notification from your status page.</p>",
		}
	}
}
