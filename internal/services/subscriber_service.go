package services

import (
	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"gorm.io/gorm"
)

// SubscriberService provides methods for managing subscribers.
type SubscriberService struct{}

// NewSubscriberService creates a new SubscriberService.
func NewSubscriberService() *SubscriberService {
	return &SubscriberService{}
}

// Subscribe adds a new subscriber or updates their service subscriptions.
func (s *SubscriberService) Subscribe(email string, serviceIDs []uint) (*models.Subscriber, error) {
	// Find or create the subscriber
	subscriber := &models.Subscriber{Email: email}
	if err := database.DB.Where(models.Subscriber{Email: email}).FirstOrCreate(subscriber).Error; err != nil {
		return nil, err
	}

	var services []*models.Service
	if len(serviceIDs) > 0 {
		// Validate that all service IDs exist
		var count int64
		database.DB.Model(&models.Service{}).Where("id IN ?", serviceIDs).Count(&count)
		if count != int64(len(serviceIDs)) {
			return nil, gorm.ErrRecordNotFound // Or a custom error
		}

		// Fetch the services to subscribe to
		if err := database.DB.Where("id IN ?", serviceIDs).Find(&services).Error; err != nil {
			return nil, err
		}
	}

	// Replace the subscriber's service associations
	if err := database.DB.Model(subscriber).Association("Services").Replace(services); err != nil {
		return nil, err
	}

	return subscriber, nil
}

// Unsubscribe removes a subscriber.
func (s *SubscriberService) Unsubscribe(email string) error {
	var subscriber models.Subscriber
	if err := database.DB.Where("email = ?", email).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Already unsubscribed
		}
		return err
	}

	// Clear associations before deleting
	if err := database.DB.Model(&subscriber).Association("Services").Clear(); err != nil {
		return err
	}

	return database.DB.Delete(&subscriber).Error
}

// GetAllSubscribers returns all subscribers with their service subscriptions.
func (s *SubscriberService) GetAllSubscribers() ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	if err := database.DB.Preload("Services").Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// GetSubscribersForServices returns all subscribers for a given list of service IDs.
func (s *SubscriberService) GetSubscribersForServices(serviceIDs []uint) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	if err := database.DB.
		Joins("JOIN subscriber_services ON subscriber_services.subscriber_id = subscribers.id").
		Where("subscriber_services.service_id IN ?", serviceIDs).
		Distinct().
		Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}
