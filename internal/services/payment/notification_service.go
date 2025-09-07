package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotificationService implements the NotificationService interface
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		db: db,
	}
}

// SendPaymentSuccessNotification sends payment success notification
func (s *NotificationService) SendPaymentSuccessNotification(ctx context.Context, payment *PaymentResponse) error {
	// Create notification record
	notification := &models.Notification{
		Type:          "payment_success",
		Title:         "Payment Successful",
		Message:       fmt.Sprintf("Your payment of %.2f %s has been processed successfully.", payment.Amount, payment.Currency),
		RecipientID:   "0", // Will be extracted from payment data
		RecipientType: "tenant",
		Data:          `{"payment_id":"` + payment.ID + `","amount":` + fmt.Sprintf("%.2f", payment.Amount) + `,"currency":"` + payment.Currency + `","method":"` + payment.Method + `","gateway":"` + payment.Gateway + `"}`,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(notification).Error; err != nil {
		logger.Error("Failed to create payment success notification", zap.Error(err))
		return err
	}

	// Send email notification
	if err := s.sendEmailNotification(ctx, notification); err != nil {
		logger.Error("Failed to send payment success email", zap.Error(err))
		// Don't fail the entire operation if email fails
	}

	// Send in-app notification
	if err := s.sendInAppNotification(ctx, notification); err != nil {
		logger.Error("Failed to send payment success in-app notification", zap.Error(err))
		// Don't fail the entire operation if in-app notification fails
	}

	logger.Info("Payment success notification sent",
		zap.String("payment_id", payment.ID),
		zap.String("recipient", "0"))

	return nil
}

// SendPaymentFailureNotification sends payment failure notification
func (s *NotificationService) SendPaymentFailureNotification(ctx context.Context, payment *PaymentResponse, reason string) error {
	// Create notification record
	notification := &models.Notification{
		Type:          "payment_failure",
		Title:         "Payment Failed",
		Message:       fmt.Sprintf("Your payment of %.2f %s failed. Reason: %s", payment.Amount, payment.Currency, reason),
		RecipientID:   "0", // Will be extracted from payment data
		RecipientType: "tenant",
		Data:          `{"payment_id":"` + payment.ID + `","amount":` + fmt.Sprintf("%.2f", payment.Amount) + `,"currency":"` + payment.Currency + `","method":"` + payment.Method + `","gateway":"` + payment.Gateway + `","reason":"` + reason + `"}`,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(notification).Error; err != nil {
		logger.Error("Failed to create payment failure notification", zap.Error(err))
		return err
	}

	// Send email notification
	if err := s.sendEmailNotification(ctx, notification); err != nil {
		logger.Error("Failed to send payment failure email", zap.Error(err))
		// Don't fail the entire operation if email fails
	}

	// Send in-app notification
	if err := s.sendInAppNotification(ctx, notification); err != nil {
		logger.Error("Failed to send payment failure in-app notification", zap.Error(err))
		// Don't fail the entire operation if in-app notification fails
	}

	logger.Info("Payment failure notification sent",
		zap.String("payment_id", payment.ID),
		zap.String("recipient", "0"),
		zap.String("reason", reason))

	return nil
}

// SendRefundNotification sends refund notification
func (s *NotificationService) SendRefundNotification(ctx context.Context, refund *RefundResponse) error {
	// Create notification record
	notification := &models.Notification{
		Type:          "refund_processed",
		Title:         "Refund Processed",
		Message:       fmt.Sprintf("Your refund of %.2f has been processed successfully.", refund.Amount),
		RecipientID:   "0", // Will be extracted from refund data
		RecipientType: "tenant",
		Data:          `{"refund_id":"` + refund.ID + `","payment_id":"` + refund.PaymentID + `","amount":` + fmt.Sprintf("%.2f", refund.Amount) + `,"status":"` + refund.Status + `"}`,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(notification).Error; err != nil {
		logger.Error("Failed to create refund notification", zap.Error(err))
		return err
	}

	// Send email notification
	if err := s.sendEmailNotification(ctx, notification); err != nil {
		logger.Error("Failed to send refund email", zap.Error(err))
		// Don't fail the entire operation if email fails
	}

	// Send in-app notification
	if err := s.sendInAppNotification(ctx, notification); err != nil {
		logger.Error("Failed to send refund in-app notification", zap.Error(err))
		// Don't fail the entire operation if in-app notification fails
	}

	logger.Info("Refund notification sent",
		zap.String("refund_id", refund.ID),
		zap.String("recipient", "0"))

	return nil
}

// SendSubscriptionNotification sends subscription notification
func (s *NotificationService) SendSubscriptionNotification(ctx context.Context, subscription *PaymentResponse, eventType string) error {
	var title, message string

	switch eventType {
	case "created":
		title = "Subscription Created"
		message = "Your subscription has been created successfully."
	case "updated":
		title = "Subscription Updated"
		message = "Your subscription has been updated."
	case "cancelled":
		title = "Subscription Cancelled"
		message = "Your subscription has been cancelled."
	default:
		title = "Subscription Update"
		message = fmt.Sprintf("Your subscription has been %s.", eventType)
	}

	// Create notification record
	notification := &models.Notification{
		Type:          "subscription_" + eventType,
		Title:         title,
		Message:       message,
		RecipientID:   "0", // Will be extracted from subscription data
		RecipientType: "tenant",
		Data:          `{"subscription_id":"` + subscription.ID + `","status":"` + subscription.Status + `","event_type":"` + eventType + `","gateway":"` + subscription.Gateway + `"}`,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(notification).Error; err != nil {
		logger.Error("Failed to create subscription notification", zap.Error(err))
		return err
	}

	// Send email notification
	if err := s.sendEmailNotification(ctx, notification); err != nil {
		logger.Error("Failed to send subscription email", zap.Error(err))
		// Don't fail the entire operation if email fails
	}

	// Send in-app notification
	if err := s.sendInAppNotification(ctx, notification); err != nil {
		logger.Error("Failed to send subscription in-app notification", zap.Error(err))
		// Don't fail the entire operation if in-app notification fails
	}

	logger.Info("Subscription notification sent",
		zap.String("subscription_id", subscription.ID),
		zap.String("recipient", "0"),
		zap.String("event_type", eventType))

	return nil
}

// sendEmailNotification sends email notification
func (s *NotificationService) sendEmailNotification(ctx context.Context, notification *models.Notification) error {
	// Get tenant details
	var tenant models.Tenant
	if err := s.db.First(&tenant, notification.RecipientID).Error; err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Create email notification record
	emailNotification := &models.EmailNotification{
		NotificationID: notification.ID,
		RecipientEmail: "noreply@example.com", // Placeholder email
		Subject:        notification.Title,
		Body:           notification.Message,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(emailNotification).Error; err != nil {
		return fmt.Errorf("failed to create email notification: %w", err)
	}

	// Here you would integrate with your email service
	// For now, we'll just mark it as sent
	emailNotification.Status = "sent"
	emailNotification.SentAt = &[]time.Time{time.Now()}[0]
	emailNotification.UpdatedAt = time.Now()
	s.db.Save(emailNotification)

	logger.Info("Email notification sent",
		zap.Uint("notification_id", notification.ID),
		zap.String("recipient_email", "noreply@example.com"))

	return nil
}

// sendInAppNotification sends in-app notification
func (s *NotificationService) sendInAppNotification(ctx context.Context, notification *models.Notification) error {
	// Mark notification as sent
	notification.Status = "sent"
	notification.SentAt = &[]time.Time{time.Now()}[0]
	notification.UpdatedAt = time.Now()
	s.db.Save(notification)

	logger.Info("In-app notification sent",
		zap.Uint("notification_id", notification.ID),
		zap.String("recipient", notification.RecipientID))

	return nil
}
