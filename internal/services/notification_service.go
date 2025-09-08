package services

import (
	"fmt"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/email"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// NotificationService is responsible for sending notifications to subscribers.
type NotificationService struct {
	subscriberService *SubscriberService
	emailSender       email.Sender
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(subscriberService *SubscriberService, emailSender email.Sender) *NotificationService {
	return &NotificationService{
		subscriberService: subscriberService,
		emailSender:       emailSender,
	}
}

// NotifyNewIncident sends a notification to subscribers of affected services.
func (s *NotificationService) NotifyNewIncident(incident *models.Incident) error {
	if len(incident.Services) == 0 {
		return nil // No services affected, no one to notify
	}

	var serviceIDs []uint
	for _, service := range incident.Services {
		serviceIDs = append(serviceIDs, service.ID)
	}

	subscribers, err := s.subscriberService.GetSubscribersForServices(serviceIDs)
	if err != nil {
		return fmt.Errorf("failed to get subscribers for services: %w", err)
	}

	subject := fmt.Sprintf("New Incident: %s", incident.Title)
	body := fmt.Sprintf("A new incident has been reported: %s\n\nDescription: %s\nStatus: %s", incident.Title, incident.Description, incident.Status)

	for _, subscriber := range subscribers {
		if err := s.emailSender.Send(subscriber.Email, subject, body); err != nil {
			logger.Log.Error("Failed to send incident notification email",
				zap.String("email", subscriber.Email),
				zap.Error(err))
		}
	}

	return nil
}

// NotifyNewMaintenance sends a notification to subscribers of affected services.
func (s *NotificationService) NotifyNewMaintenance(event *models.Maintenance) error {
	if len(event.Services) == 0 {
		return nil // No services affected, no one to notify
	}

	var serviceIDs []uint
	for _, service := range event.Services {
		serviceIDs = append(serviceIDs, service.ID)
	}

	subscribers, err := s.subscriberService.GetSubscribersForServices(serviceIDs)
	if err != nil {
		return fmt.Errorf("failed to get subscribers for services: %w", err)
	}

	subject := fmt.Sprintf("Scheduled Maintenance: %s", event.Title)
	body := fmt.Sprintf("A new maintenance event has been scheduled: %s\n\nDescription: %s\nStatus: %s\nFrom: %s To: %s", event.Title, event.Description, event.Status, event.StartAt, event.EndAt)

	for _, subscriber := range subscribers {
		if err := s.emailSender.Send(subscriber.Email, subject, body); err != nil {
			logger.Log.Error("Failed to send maintenance notification email",
				zap.String("email", subscriber.Email),
				zap.Error(err))
		}
	}

	return nil
}
