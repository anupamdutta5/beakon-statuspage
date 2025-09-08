package services

import (
	"time"

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
func (s *SubscriberService) Subscribe(email string, serviceIDs []uint, tenantID uint) (*models.Subscriber, error) {
	// Find or create the subscriber
	subscriber := &models.Subscriber{Email: email, TenantID: tenantID}
	if err := database.DB.Where("email = ? AND tenant_id = ?", email, tenantID).FirstOrCreate(subscriber).Error; err != nil {
		return nil, err
	}

	var services []*models.Service
	if len(serviceIDs) > 0 {
		// Validate that all service IDs exist and belong to the tenant
		var count int64
		database.DB.Model(&models.Service{}).Where("id IN ? AND tenant_id = ?", serviceIDs, tenantID).Count(&count)
		if count != int64(len(serviceIDs)) {
			return nil, gorm.ErrRecordNotFound
		}

		// Fetch the services to subscribe to
		if err := database.DB.Where("id IN ? AND tenant_id = ?", serviceIDs, tenantID).Find(&services).Error; err != nil {
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
func (s *SubscriberService) Unsubscribe(email string, tenantID uint) error {
	var subscriber models.Subscriber
	if err := database.DB.Where("email = ? AND tenant_id = ?", email, tenantID).First(&subscriber).Error; err != nil {
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

// GetSubscribersByTenantID retrieves all subscribers for a specific tenant
func (s *SubscriberService) GetSubscribersByTenantID(tenantID uint) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	if err := database.DB.Where("tenant_id = ?", tenantID).Preload("Services").Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// CreateSubscriber creates a new subscriber
func (s *SubscriberService) CreateSubscriber(subscriber *models.Subscriber) error {
	if err := database.DB.Create(subscriber).Error; err != nil {
		return err
	}
	return nil
}

// UpdateSubscriber updates an existing subscriber
func (s *SubscriberService) UpdateSubscriber(subscriber *models.Subscriber) error {
	return database.DB.Save(subscriber).Error
}

// DeleteSubscriber deletes a subscriber by ID
func (s *SubscriberService) DeleteSubscriber(id uint) error {
	var subscriber models.Subscriber
	if err := database.DB.First(&subscriber, id).Error; err != nil {
		return err
	}

	// Clear associations before deleting
	if err := database.DB.Model(&subscriber).Association("Services").Clear(); err != nil {
		return err
	}

	return database.DB.Delete(&subscriber).Error
}

// GetSubscriberByID retrieves a subscriber by ID
func (s *SubscriberService) GetSubscriberByID(id uint) (*models.Subscriber, error) {
	var subscriber models.Subscriber
	if err := database.DB.Preload("Services").First(&subscriber, id).Error; err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// GetSubscriberByEmail retrieves a subscriber by email and tenant
func (s *SubscriberService) GetSubscriberByEmail(email string, tenantID uint) (*models.Subscriber, error) {
	var subscriber models.Subscriber
	if err := database.DB.Where("email = ? AND tenant_id = ?", email, tenantID).Preload("Services").First(&subscriber).Error; err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// GetSubscribersForServices returns all subscribers for a given list of service IDs.
func (s *SubscriberService) GetSubscribersForServices(serviceIDs []uint) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	if err := database.DB.
		Joins("JOIN subscriber_services ON subscriber_services.subscriber_id = subscribers.id").
		Where("subscriber_services.service_id IN ?", serviceIDs).
		Distinct().
		Preload("Services").
		Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// GetSubscribersForService returns all subscribers for a specific service
func (s *SubscriberService) GetSubscribersForService(serviceID uint) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	if err := database.DB.
		Joins("JOIN subscriber_services ON subscriber_services.subscriber_id = subscribers.id").
		Where("subscriber_services.service_id = ?", serviceID).
		Distinct().
		Preload("Services").
		Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// AddServiceToSubscriber adds a service subscription to a subscriber
func (s *SubscriberService) AddServiceToSubscriber(subscriberID uint, serviceID uint) error {
	var subscriber models.Subscriber
	if err := database.DB.First(&subscriber, subscriberID).Error; err != nil {
		return err
	}

	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil {
		return err
	}

	return database.DB.Model(&subscriber).Association("Services").Append(&service)
}

// RemoveServiceFromSubscriber removes a service subscription from a subscriber
func (s *SubscriberService) RemoveServiceFromSubscriber(subscriberID uint, serviceID uint) error {
	var subscriber models.Subscriber
	if err := database.DB.First(&subscriber, subscriberID).Error; err != nil {
		return err
	}

	var service models.Service
	if err := database.DB.First(&service, serviceID).Error; err != nil {
		return err
	}

	return database.DB.Model(&subscriber).Association("Services").Delete(&service)
}

// GetSubscriberStatistics returns subscriber statistics for a tenant
func (s *SubscriberService) GetSubscriberStatistics(tenantID uint) (*SubscriberStatistics, error) {
	var stats SubscriberStatistics

	// Total subscribers
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ?", tenantID).
		Count(&stats.TotalSubscribers).Error; err != nil {
		return nil, err
	}

	// Subscribers with phone numbers (SMS capable)
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ? AND phone != ''", tenantID).
		Count(&stats.SMSSubscribers).Error; err != nil {
		return nil, err
	}

	// Recent subscribers (last 30 days)
	since := time.Now().Add(-30 * 24 * time.Hour)
	if err := database.DB.Model(&models.Subscriber{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&stats.RecentSubscribers).Error; err != nil {
		return nil, err
	}

	// Subscribers by service
	var serviceStats []struct {
		ServiceID uint
		Count     int64
	}
	if err := database.DB.Table("subscriber_services").
		Select("service_id, COUNT(*) as count").
		Joins("JOIN subscribers ON subscriber_services.subscriber_id = subscribers.id").
		Where("subscribers.tenant_id = ?", tenantID).
		Group("service_id").
		Scan(&serviceStats).Error; err != nil {
		return nil, err
	}

	stats.SubscribersByService = make(map[uint]int64)
	for _, stat := range serviceStats {
		stats.SubscribersByService[stat.ServiceID] = stat.Count
	}

	return &stats, nil
}

// SearchSubscribers searches subscribers by email
func (s *SubscriberService) SearchSubscribers(tenantID uint, query string) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	searchQuery := "%" + query + "%"
	if err := database.DB.Where("tenant_id = ? AND email ILIKE ?", tenantID, searchQuery).
		Preload("Services").
		Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// GetSubscribersWithPagination returns paginated subscribers
func (s *SubscriberService) GetSubscribersWithPagination(tenantID uint, page, pageSize int) ([]models.Subscriber, int64, error) {
	var subscribers []models.Subscriber
	var total int64

	// Get total count
	if err := database.DB.Model(&models.Subscriber{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := database.DB.Where("tenant_id = ?", tenantID).
		Preload("Services").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&subscribers).Error; err != nil {
		return nil, 0, err
	}

	return subscribers, total, nil
}

// BulkSubscribe adds multiple subscribers at once
func (s *SubscriberService) BulkSubscribe(emails []string, serviceIDs []uint, tenantID uint) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	var services []*models.Service

	// Validate services
	if len(serviceIDs) > 0 {
		if err := database.DB.Where("id IN ? AND tenant_id = ?", serviceIDs, tenantID).Find(&services).Error; err != nil {
			return nil, err
		}
	}

	// Create subscribers
	for _, email := range emails {
		subscriber := &models.Subscriber{
			Email:    email,
			TenantID: tenantID,
		}

		if err := database.DB.Where("email = ? AND tenant_id = ?", email, tenantID).FirstOrCreate(subscriber).Error; err != nil {
			return nil, err
		}

		// Associate services
		if len(services) > 0 {
			if err := database.DB.Model(subscriber).Association("Services").Replace(services); err != nil {
				return nil, err
			}
		}

		subscribers = append(subscribers, *subscriber)
	}

	return subscribers, nil
}

// ExportSubscribers exports subscribers to CSV format
func (s *SubscriberService) ExportSubscribers(tenantID uint) ([]SubscriberExport, error) {
	var subscribers []models.Subscriber
	if err := database.DB.Where("tenant_id = ?", tenantID).Preload("Services").Find(&subscribers).Error; err != nil {
		return nil, err
	}

	var exports []SubscriberExport
	for _, subscriber := range subscribers {
		var serviceNames []string
		for _, service := range subscriber.Services {
			serviceNames = append(serviceNames, service.Name)
		}

		exports = append(exports, SubscriberExport{
			Email:        subscriber.Email,
			Phone:        subscriber.Phone,
			Services:     serviceNames,
			SubscribedAt: subscriber.CreatedAt,
		})
	}

	return exports, nil
}

// SubscriberStatistics represents subscriber statistics
type SubscriberStatistics struct {
	TotalSubscribers     int64          `json:"total_subscribers"`
	SMSSubscribers       int64          `json:"sms_subscribers"`
	RecentSubscribers    int64          `json:"recent_subscribers"`
	SubscribersByService map[uint]int64 `json:"subscribers_by_service"`
}

// SubscriberExport represents subscriber data for export
type SubscriberExport struct {
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Services     []string  `json:"services"`
	SubscribedAt time.Time `json:"subscribed_at"`
}
