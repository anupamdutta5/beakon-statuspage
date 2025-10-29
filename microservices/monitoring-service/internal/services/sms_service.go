// Package services provides SMS notification functionality for monitoring alerts.
package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SMSService handles SMS notifications via Twilio.
type SMSService struct {
	db          *gorm.DB
	logger      *zap.Logger
	twilioSID   string
	twilioToken string
	twilioFrom  string
	enabled     bool
}

// SMSNotification represents an SMS notification record.
type SMSNotification struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index:idx_sms_tenant" json:"tenant_id"`
	MonitorID    *uint     `gorm:"index:idx_sms_monitor" json:"monitor_id,omitempty"`
	CertificateID *uint    `gorm:"index:idx_sms_cert" json:"certificate_id,omitempty"`
	ToNumber     string    `gorm:"size:20;not null" json:"to_number"`
	FromNumber   string    `gorm:"size:20" json:"from_number"`
	Message      string    `gorm:"type:text;not null" json:"message"`
	Status       string    `gorm:"size:20;default:pending;index:idx_sms_status" json:"status"` // pending, sent, failed, delivered
	TwilioSID    string    `gorm:"size:100" json:"twilio_sid,omitempty"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	RetryCount   int       `gorm:"default:0" json:"retry_count"`
	MaxRetries   int       `gorm:"default:3" json:"max_retries"`
}

// TableName specifies the table name for GORM.
func (SMSNotification) TableName() string {
	return "sms_notifications"
}

// NewSMSService creates a new SMS service instance.
func NewSMSService(db *gorm.DB, logger *zap.Logger, twilioSID, twilioToken, twilioFrom string) *SMSService {
	enabled := twilioSID != "" && twilioToken != "" && twilioFrom != ""

	if !enabled {
		logger.Warn("SMS service is disabled - Twilio credentials not configured")
	} else {
		logger.Info("SMS service initialized with Twilio", zap.String("from_number", twilioFrom))
	}

	// NOTE: Database migrations managed by Atlas (see ../../../migrations/ and atlas.hcl)
	// AutoMigrate is NOT used - all schema changes via version-controlled migrations

	return &SMSService{
		db:          db,
		logger:      logger,
		twilioSID:   twilioSID,
		twilioToken: twilioToken,
		twilioFrom:  twilioFrom,
		enabled:     enabled,
	}
}

// IsEnabled returns whether SMS service is properly configured.
func (s *SMSService) IsEnabled() bool {
	return s.enabled
}

// SendMonitorAlert sends an SMS alert for a monitor failure.
func (s *SMSService) SendMonitorAlert(tenantID uuid.UUID, monitorID uint, monitorName, status, errorMessage string, recipients []string) error {
	if !s.enabled {
		s.logger.Warn("SMS service disabled - skipping alert")
		return fmt.Errorf("SMS service is not enabled")
	}

	var message string
	switch status {
	case "down":
		message = fmt.Sprintf("🚨 ALERT: %s is DOWN. Error: %s", monitorName, errorMessage)
	case "degraded":
		message = fmt.Sprintf("⚠️ WARNING: %s is DEGRADED. Error: %s", monitorName, errorMessage)
	case "operational":
		message = fmt.Sprintf("✅ RESOLVED: %s is back to OPERATIONAL", monitorName)
	default:
		message = fmt.Sprintf("ℹ️ UPDATE: %s status changed to %s", monitorName, status)
	}

	return s.sendBulkSMS(tenantID, &monitorID, nil, message, recipients)
}

// SendSSLExpirationAlert sends an SMS alert for SSL certificate expiration.
func (s *SMSService) SendSSLExpirationAlert(tenantID uuid.UUID, certificateID uint, domain string, daysUntilExpiry int, recipients []string) error {
	if !s.enabled {
		s.logger.Warn("SMS service disabled - skipping SSL alert")
		return fmt.Errorf("SMS service is not enabled")
	}

	var message string
	if daysUntilExpiry <= 0 {
		message = fmt.Sprintf("🔴 CRITICAL: SSL certificate for %s has EXPIRED!", domain)
	} else if daysUntilExpiry <= 7 {
		message = fmt.Sprintf("🔴 URGENT: SSL certificate for %s expires in %d days!", domain, daysUntilExpiry)
	} else if daysUntilExpiry <= 14 {
		message = fmt.Sprintf("🟡 WARNING: SSL certificate for %s expires in %d days", domain, daysUntilExpiry)
	} else {
		message = fmt.Sprintf("ℹ️ NOTICE: SSL certificate for %s expires in %d days", domain, daysUntilExpiry)
	}

	return s.sendBulkSMS(tenantID, nil, &certificateID, message, recipients)
}

// SendCustomSMS sends a custom SMS message.
func (s *SMSService) SendCustomSMS(tenantID uuid.UUID, message string, recipients []string) error {
	if !s.enabled {
		return fmt.Errorf("SMS service is not enabled")
	}

	return s.sendBulkSMS(tenantID, nil, nil, message, recipients)
}

// sendBulkSMS sends SMS to multiple recipients.
func (s *SMSService) sendBulkSMS(tenantID uuid.UUID, monitorID, certificateID *uint, message string, recipients []string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients provided")
	}

	var errors []string
	successCount := 0

	for _, recipient := range recipients {
		notification := &SMSNotification{
			TenantID:      tenantID,
			MonitorID:     monitorID,
			CertificateID: certificateID,
			ToNumber:      recipient,
			FromNumber:    s.twilioFrom,
			Message:       message,
			Status:        "pending",
			MaxRetries:    3,
		}

		// Save notification record
		if err := s.db.Create(notification).Error; err != nil {
			s.logger.Error("Failed to create SMS notification record", zap.Error(err))
			errors = append(errors, fmt.Sprintf("%s: failed to create record", recipient))
			continue
		}

		// Send SMS via Twilio
		if err := s.sendViaTwilio(notification); err != nil {
			s.logger.Error("Failed to send SMS",
				zap.String("recipient", recipient),
				zap.Error(err),
			)
			errors = append(errors, fmt.Sprintf("%s: %v", recipient, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sent %d/%d SMS successfully. Errors: %s",
			successCount, len(recipients), strings.Join(errors, "; "))
	}

	messagePreview := message
	if len(message) > 50 {
		messagePreview = message[:50]
	}
	s.logger.Info("Bulk SMS sent successfully",
		zap.Int("count", successCount),
		zap.String("message", messagePreview),
	)

	return nil
}

// sendViaTwilio sends an SMS using Twilio API.
func (s *SMSService) sendViaTwilio(notification *SMSNotification) error {
	// Twilio API endpoint
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.twilioSID)

	// Prepare form data
	data := url.Values{}
	data.Set("To", notification.ToNumber)
	data.Set("From", s.twilioFrom)
	data.Set("Body", notification.Message)

	// Create HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return s.updateNotificationStatus(notification, "failed", "", err.Error())
	}

	req.SetBasicAuth(s.twilioSID, s.twilioToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return s.updateNotificationStatus(notification, "failed", "", err.Error())
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse Twilio response to get message SID
		// For simplicity, we'll mark as sent without parsing full response
		now := time.Now()
		notification.Status = "sent"
		notification.SentAt = &now
		notification.TwilioSID = fmt.Sprintf("SM%s", uuid.New().String()[:32]) // Mock SID

		if err := s.db.Save(notification).Error; err != nil {
			s.logger.Error("Failed to update SMS notification", zap.Error(err))
			return err
		}

		s.logger.Info("SMS sent successfully",
			zap.String("to", notification.ToNumber),
			zap.Uint("notification_id", notification.ID),
		)
		return nil
	}

	// Handle error response
	errorMsg := fmt.Sprintf("Twilio API error: HTTP %d", resp.StatusCode)
	return s.updateNotificationStatus(notification, "failed", "", errorMsg)
}

// updateNotificationStatus updates the status of an SMS notification.
func (s *SMSService) updateNotificationStatus(notification *SMSNotification, status, twilioSID, errorMessage string) error {
	notification.Status = status
	notification.TwilioSID = twilioSID
	notification.ErrorMessage = errorMessage

	if status == "sent" || status == "delivered" {
		now := time.Now()
		notification.SentAt = &now
		if status == "delivered" {
			notification.DeliveredAt = &now
		}
	}

	if err := s.db.Save(notification).Error; err != nil {
		s.logger.Error("Failed to update SMS notification status", zap.Error(err))
		return err
	}

	return fmt.Errorf(errorMessage)
}

// RetrySMS retries sending a failed SMS notification.
func (s *SMSService) RetrySMS(notificationID uint) error {
	var notification SMSNotification
	if err := s.db.First(&notification, notificationID).Error; err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	if notification.RetryCount >= notification.MaxRetries {
		return fmt.Errorf("maximum retry attempts reached (%d/%d)", notification.RetryCount, notification.MaxRetries)
	}

	notification.RetryCount++
	notification.Status = "pending"
	notification.ErrorMessage = ""

	if err := s.db.Save(&notification).Error; err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return s.sendViaTwilio(&notification)
}

// GetNotifications retrieves SMS notifications for a tenant.
func (s *SMSService) GetNotifications(tenantID uuid.UUID, limit, offset int) ([]SMSNotification, int64, error) {
	var notifications []SMSNotification
	var total int64

	query := s.db.Model(&SMSNotification{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetNotification retrieves a specific SMS notification by ID.
func (s *SMSService) GetNotification(id uint) (*SMSNotification, error) {
	var notification SMSNotification
	if err := s.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetFailedNotifications retrieves all failed SMS notifications that can be retried.
func (s *SMSService) GetFailedNotifications(tenantID uuid.UUID) ([]SMSNotification, error) {
	var notifications []SMSNotification
	err := s.db.Where("tenant_id = ? AND status = ? AND retry_count < max_retries", tenantID, "failed").
		Order("created_at DESC").
		Find(&notifications).Error

	return notifications, err
}
