// Package services provides email service implementation.
package services

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/enterprise-status/statuspage-notification-consumer/internal/config"
	"github.com/enterprise-status/statuspage-notification-consumer/internal/models"
	"go.uber.org/zap"
)

// EmailService handles email notifications.
type EmailService struct {
	config *config.Config
	logger *zap.Logger
}

// NewEmailService creates a new email service.
func NewEmailService(cfg *config.Config, logger *zap.Logger) *EmailService {
	return &EmailService{
		config: cfg,
		logger: logger,
	}
}

// SendEmail sends an email notification.
func (s *EmailService) SendEmail(ctx context.Context, event *models.NotificationEvent) error {
	s.logger.Info("Sending email notification",
		zap.String("event_id", event.ID),
		zap.String("recipient", event.Recipient))

	// Convert notification event to email notification
	emailNotification := s.convertToEmailNotification(event)

	// Send email
	if err := s.sendSMTPEmail(ctx, emailNotification); err != nil {
		s.logger.Error("Failed to send email",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully",
		zap.String("event_id", event.ID),
		zap.String("recipient", event.Recipient))

	return nil
}

// convertToEmailNotification converts a notification event to email notification.
func (s *EmailService) convertToEmailNotification(event *models.NotificationEvent) *models.EmailNotification {
	return &models.EmailNotification{
		ID:             event.ID,
		NotificationID: event.NotificationID,
		TenantID:       event.TenantID,
		UserID:         event.UserID,
		To:             []string{event.Recipient},
		Subject:        event.Subject,
		Body:           event.Content,
		HTMLBody:       event.Content, // For now, use same content for HTML
		TemplateID:     event.TemplateID,
		Variables:      event.Variables,
		Priority:       event.Priority,
		Metadata:       event.Metadata,
	}
}

// sendSMTPEmail sends an email via SMTP.
func (s *EmailService) sendSMTPEmail(ctx context.Context, email *models.EmailNotification) error {
	// For now, we'll simulate sending email
	// In production, you would implement actual SMTP sending

	s.logger.Info("Simulating SMTP email send",
		zap.String("email_id", email.ID),
		zap.Strings("to", email.To),
		zap.String("subject", email.Subject))

	// Simulate processing time
	// time.Sleep(100 * time.Millisecond)

	// In a real implementation, you would:
	// 1. Set up SMTP authentication
	// 2. Create email message with headers
	// 3. Send via SMTP server
	// 4. Handle errors and retries

	return nil
}

// sendSMTPEmailReal sends an email via SMTP (real implementation).
func (s *EmailService) sendSMTPEmailReal(ctx context.Context, email *models.EmailNotification) error {
	// SMTP configuration
	smtpHost := "localhost" // This should be SMTP host from config
	smtpPort := "587"       // This should be from config
	smtpUser := ""          // This should be from config
	smtpPass := ""          // This should be from config

	// Create authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Create message
	message := s.createEmailMessage(email)

	// Send email
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, smtpUser, email.To, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

// createEmailMessage creates an email message.
func (s *EmailService) createEmailMessage(email *models.EmailNotification) string {
	var message strings.Builder

	// Headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", "noreply@statuspage.com"))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")

	// Body
	message.WriteString(email.HTMLBody)

	return message.String()
}

// ValidateEmailAddress validates an email address.
func (s *EmailService) ValidateEmailAddress(email string) bool {
	// Basic email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// GetEmailStats returns email sending statistics.
func (s *EmailService) GetEmailStats() (map[string]interface{}, error) {
	// This would typically query the database for email statistics
	stats := map[string]interface{}{
		"total_sent":   0,
		"total_failed": 0,
		"success_rate": 0.0,
		"average_time": 0.0,
		"last_updated": "2024-01-01T00:00:00Z",
	}

	return stats, nil
}
