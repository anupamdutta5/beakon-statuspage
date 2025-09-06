package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MobileService struct {
	db *gorm.DB
}

func NewMobileService() *MobileService {
	return &MobileService{
		db: database.DB,
	}
}

// Mobile Device Management

type MobileDevice struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null"`
	DeviceToken string `gorm:"uniqueIndex;not null"`
	DeviceType  string `gorm:"not null"` // ios, android
	AppVersion  string
	OSVersion   string
	DeviceModel string
	IsActive    bool `gorm:"default:true"`
	LastSeenAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *MobileService) RegisterDevice(userID uint, deviceToken, deviceType, appVersion, osVersion, deviceModel string) (*MobileDevice, error) {
	// Check if device already exists
	var existingDevice MobileDevice
	if err := s.db.Where("device_token = ?", deviceToken).First(&existingDevice).Error; err == nil {
		// Update existing device
		existingDevice.UserID = userID
		existingDevice.DeviceType = deviceType
		existingDevice.AppVersion = appVersion
		existingDevice.OSVersion = osVersion
		existingDevice.DeviceModel = deviceModel
		existingDevice.IsActive = true
		existingDevice.LastSeenAt = time.Now()

		if err := s.db.Save(&existingDevice).Error; err != nil {
			logger.Error("Failed to update mobile device", zap.Error(err))
			return nil, err
		}

		return &existingDevice, nil
	}

	// Create new device
	device := &MobileDevice{
		UserID:      userID,
		DeviceToken: deviceToken,
		DeviceType:  deviceType,
		AppVersion:  appVersion,
		OSVersion:   osVersion,
		DeviceModel: deviceModel,
		IsActive:    true,
		LastSeenAt:  time.Now(),
	}

	if err := s.db.Create(device).Error; err != nil {
		logger.Error("Failed to register mobile device", zap.Error(err))
		return nil, err
	}

	logger.Info("Mobile device registered successfully",
		zap.Uint("user_id", userID),
		zap.String("device_type", deviceType))

	return device, nil
}

func (s *MobileService) GetUserDevices(userID uint) ([]MobileDevice, error) {
	var devices []MobileDevice
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&devices).Error
	return devices, err
}

func (s *MobileService) UnregisterDevice(deviceToken string) error {
	if err := s.db.Model(&MobileDevice{}).Where("device_token = ?", deviceToken).Update("is_active", false).Error; err != nil {
		logger.Error("Failed to unregister mobile device", zap.Error(err))
		return err
	}

	logger.Info("Mobile device unregistered successfully", zap.String("device_token", deviceToken))
	return nil
}

// Mobile Push Notifications

type MobilePushNotification struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null"`
	DeviceToken string `gorm:"not null"`
	Title       string `gorm:"not null"`
	Body        string `gorm:"not null"`
	Data        string // JSON data
	Type        string `gorm:"not null"`          // incident, maintenance, status_update
	Status      string `gorm:"default:'pending'"` // pending, sent, failed
	SentAt      *time.Time
	Error       string
	CreatedAt   time.Time
}

func (s *MobileService) SendPushNotification(userID uint, title, body, data, notificationType string) error {
	// Get user's active devices
	devices, err := s.GetUserDevices(userID)
	if err != nil {
		return err
	}

	for _, device := range devices {
		notification := &MobilePushNotification{
			UserID:      userID,
			DeviceToken: device.DeviceToken,
			Title:       title,
			Body:        body,
			Data:        data,
			Type:        notificationType,
			Status:      "pending",
		}

		if err := s.db.Create(notification).Error; err != nil {
			logger.Error("Failed to create push notification", zap.Error(err))
			continue
		}

		// Send push notification (this would integrate with FCM/APNS)
		if err := s.sendPushToDevice(notification, &device); err != nil {
			logger.Error("Failed to send push notification", zap.Error(err))
			notification.Status = "failed"
			notification.Error = err.Error()
		} else {
			notification.Status = "sent"
			now := time.Now()
			notification.SentAt = &now
		}

		s.db.Save(notification)
	}

	return nil
}

func (s *MobileService) SendIncidentNotification(tenantID uint, incident *models.Incident) error {
	// Get all users for the tenant who have mobile devices
	var devices []MobileDevice
	if err := s.db.Joins("JOIN users ON mobile_devices.user_id = users.id").
		Where("users.tenant_id = ? AND mobile_devices.is_active = ?", tenantID, true).
		Find(&devices).Error; err != nil {
		return err
	}

	title := fmt.Sprintf("🚨 %s", incident.Title)
	body := incident.Description
	if len(body) > 100 {
		body = body[:97] + "..."
	}

	data := fmt.Sprintf(`{"incident_id": %d, "type": "incident", "status": "%s"}`, incident.ID, incident.Status)

	for _, device := range devices {
		notification := &MobilePushNotification{
			UserID:      device.UserID,
			DeviceToken: device.DeviceToken,
			Title:       title,
			Body:        body,
			Data:        data,
			Type:        "incident",
			Status:      "pending",
		}

		if err := s.db.Create(notification).Error; err != nil {
			continue
		}

		if err := s.sendPushToDevice(notification, &device); err != nil {
			notification.Status = "failed"
			notification.Error = err.Error()
		} else {
			notification.Status = "sent"
			now := time.Now()
			notification.SentAt = &now
		}

		s.db.Save(notification)
	}

	return nil
}

// Mobile API Authentication

type MobileAPIToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	DeviceID  string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsActive  bool      `gorm:"default:true"`
	CreatedAt time.Time
}

func (s *MobileService) GenerateMobileAPIToken(userID uint, deviceID string) (*MobileAPIToken, error) {
	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := "mobile_" + hex.EncodeToString(tokenBytes)

	// Create token record
	apiToken := &MobileAPIToken{
		UserID:    userID,
		Token:     token,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		IsActive:  true,
	}

	if err := s.db.Create(apiToken).Error; err != nil {
		logger.Error("Failed to create mobile API token", zap.Error(err))
		return nil, err
	}

	logger.Info("Mobile API token generated successfully",
		zap.Uint("user_id", userID),
		zap.String("device_id", deviceID))

	return apiToken, nil
}

func (s *MobileService) ValidateMobileAPIToken(token string) (*MobileAPIToken, error) {
	var apiToken MobileAPIToken
	if err := s.db.Where("token = ? AND is_active = ? AND expires_at > ?",
		token, true, time.Now()).First(&apiToken).Error; err != nil {
		return nil, err
	}

	return &apiToken, nil
}

func (s *MobileService) RevokeMobileAPIToken(token string) error {
	return s.db.Model(&MobileAPIToken{}).Where("token = ?", token).Update("is_active", false).Error
}

// Mobile App Analytics

type MobileAppAnalytics struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    uint   `gorm:"not null"`
	UserID      uint   `gorm:"not null"`
	DeviceID    string `gorm:"not null"`
	EventType   string `gorm:"not null"` // app_open, incident_view, notification_received, etc.
	EventData   string // JSON data
	AppVersion  string
	OSVersion   string
	DeviceModel string
	CreatedAt   time.Time
}

func (s *MobileService) RecordAppEvent(tenantID, userID uint, deviceID, eventType, eventData, appVersion, osVersion, deviceModel string) error {
	analytics := &MobileAppAnalytics{
		TenantID:    tenantID,
		UserID:      userID,
		DeviceID:    deviceID,
		EventType:   eventType,
		EventData:   eventData,
		AppVersion:  appVersion,
		OSVersion:   osVersion,
		DeviceModel: deviceModel,
	}

	return s.db.Create(analytics).Error
}

func (s *MobileService) GetMobileAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	analytics := make(map[string]interface{})

	// Total app opens
	var totalOpens int64
	if err := s.db.Model(&MobileAppAnalytics{}).
		Where("tenant_id = ? AND event_type = ? AND created_at >= ?", tenantID, "app_open", startDate).
		Count(&totalOpens).Error; err != nil {
		return nil, err
	}
	analytics["total_app_opens"] = totalOpens

	// Unique users
	var uniqueUsers int64
	if err := s.db.Model(&MobileAppAnalytics{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Distinct("user_id").Count(&uniqueUsers).Error; err != nil {
		return nil, err
	}
	analytics["unique_users"] = uniqueUsers

	// Device breakdown
	var deviceBreakdown []struct {
		DeviceType string `json:"device_type"`
		Count      int64  `json:"count"`
	}
	if err := s.db.Model(&MobileAppAnalytics{}).
		Select("device_model as device_type, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("device_model").
		Order("count DESC").
		Limit(10).
		Scan(&deviceBreakdown).Error; err != nil {
		return nil, err
	}
	analytics["device_breakdown"] = deviceBreakdown

	// Daily usage
	var dailyUsage []struct {
		Date  string `json:"date"`
		Opens int64  `json:"opens"`
	}
	if err := s.db.Model(&MobileAppAnalytics{}).
		Select("DATE(created_at) as date, COUNT(*) as opens").
		Where("tenant_id = ? AND event_type = ? AND created_at >= ?", tenantID, "app_open", startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyUsage).Error; err != nil {
		return nil, err
	}
	analytics["daily_usage"] = dailyUsage

	return analytics, nil
}

// Mobile App Configuration

type MobileAppConfig struct {
	ID                      uint   `gorm:"primaryKey"`
	TenantID                uint   `gorm:"not null"`
	AppName                 string `gorm:"not null"`
	AppIcon                 string
	PrimaryColor            string `gorm:"default:'#0052cc'"`
	SecondaryColor          string `gorm:"default:'#f4f5f7'"`
	EnablePushNotifications bool   `gorm:"default:true"`
	EnableBiometricAuth     bool   `gorm:"default:false"`
	RequireAppUpdate        bool   `gorm:"default:false"`
	MinAppVersion           string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (s *MobileService) GetMobileAppConfig(tenantID uint) (*MobileAppConfig, error) {
	var config MobileAppConfig
	err := s.db.Where("tenant_id = ?", tenantID).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		// Return default config
		return &MobileAppConfig{
			TenantID:                tenantID,
			AppName:                 "Status Page",
			PrimaryColor:            "#0052cc",
			SecondaryColor:          "#f4f5f7",
			EnablePushNotifications: true,
			EnableBiometricAuth:     false,
			RequireAppUpdate:        false,
		}, nil
	}
	return &config, err
}

func (s *MobileService) UpdateMobileAppConfig(tenantID uint, config *MobileAppConfig) error {
	config.TenantID = tenantID

	// Use upsert to create or update
	if err := s.db.Where("tenant_id = ?", tenantID).
		Assign(*config).
		FirstOrCreate(config).Error; err != nil {
		logger.Error("Failed to update mobile app config", zap.Error(err))
		return err
	}

	logger.Info("Mobile app config updated successfully", zap.Uint("tenant_id", tenantID))
	return nil
}

// Private helper methods

func (s *MobileService) sendPushToDevice(notification *MobilePushNotification, device *MobileDevice) error {
	// This would integrate with Firebase Cloud Messaging (FCM) for Android
	// and Apple Push Notification Service (APNS) for iOS

	switch device.DeviceType {
	case "ios":
		return s.sendAPNSPush(notification, device)
	case "android":
		return s.sendFCMPush(notification, device)
	default:
		return fmt.Errorf("unsupported device type: %s", device.DeviceType)
	}
}

func (s *MobileService) sendAPNSPush(notification *MobilePushNotification, device *MobileDevice) error {
	// APNS implementation would go here
	// This would use the Apple Push Notification service
	logger.Info("Sending APNS push notification",
		zap.String("device_token", device.DeviceToken),
		zap.String("title", notification.Title))
	return nil
}

func (s *MobileService) sendFCMPush(notification *MobilePushNotification, device *MobileDevice) error {
	// FCM implementation would go here
	// This would use Firebase Cloud Messaging
	logger.Info("Sending FCM push notification",
		zap.String("device_token", device.DeviceToken),
		zap.String("title", notification.Title))
	return nil
}
