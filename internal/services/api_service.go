package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type APIService struct {
	db *gorm.DB
}

func NewAPIService() *APIService {
	return &APIService{
		db: database.DB,
	}
}

// API Key Management

type APIKey struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    uint   `gorm:"not null"`
	Name        string `gorm:"not null"`
	Key         string `gorm:"uniqueIndex;not null"`
	Secret      string `gorm:"not null"`
	Permissions string `gorm:"not null"` // JSON array of permissions
	IsActive    bool   `gorm:"default:true"`
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *APIService) CreateAPIKey(tenantID uint, name string, permissions []string) (*APIKey, error) {
	// Generate API key and secret
	key, err := s.generateAPIKey()
	if err != nil {
		return nil, err
	}

	secret, err := s.generateAPISecret()
	if err != nil {
		return nil, err
	}

	// Convert permissions to JSON
	permissionsJSON, err := s.permissionsToJSON(permissions)
	if err != nil {
		return nil, err
	}

	apiKey := &APIKey{
		TenantID:    tenantID,
		Name:        name,
		Key:         key,
		Secret:      secret,
		Permissions: permissionsJSON,
		IsActive:    true,
	}

	if err := s.db.Create(apiKey).Error; err != nil {
		logger.Error("Failed to create API key", zap.Error(err))
		return nil, err
	}

	logger.Info("API key created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("name", name))

	return apiKey, nil
}

func (s *APIService) ValidateAPIKey(key, secret string) (*APIKey, error) {
	var apiKey APIKey
	if err := s.db.Where("key = ? AND secret = ? AND is_active = ?", key, secret, true).First(&apiKey).Error; err != nil {
		return nil, err
	}

	// Check if key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key expired")
	}

	// Update last used timestamp
	now := time.Now()
	apiKey.LastUsedAt = &now
	s.db.Save(&apiKey)

	return &apiKey, nil
}

func (s *APIService) GetAPIKeys(tenantID uint) ([]APIKey, error) {
	var apiKeys []APIKey
	err := s.db.Where("tenant_id = ?", tenantID).Find(&apiKeys).Error
	return apiKeys, err
}

func (s *APIService) RevokeAPIKey(keyID uint) error {
	if err := s.db.Model(&APIKey{}).Where("id = ?", keyID).Update("is_active", false).Error; err != nil {
		logger.Error("Failed to revoke API key", zap.Error(err))
		return err
	}

	logger.Info("API key revoked successfully", zap.Uint("key_id", keyID))
	return nil
}

// API Rate Limiting

type APIRateLimit struct {
	ID          uint      `gorm:"primaryKey"`
	TenantID    uint      `gorm:"not null"`
	APIKeyID    uint      `gorm:"not null"`
	Endpoint    string    `gorm:"not null"`
	Requests    int64     `gorm:"default:0"`
	WindowStart time.Time `gorm:"not null"`
	WindowEnd   time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *APIService) CheckRateLimit(apiKeyID uint, endpoint string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Truncate(window)

	var rateLimit APIRateLimit
	err := s.db.Where("api_key_id = ? AND endpoint = ? AND window_start = ?",
		apiKeyID, endpoint, windowStart).First(&rateLimit).Error

	if err == gorm.ErrRecordNotFound {
		// Create new rate limit record
		rateLimit = APIRateLimit{
			TenantID:    0, // Will be set from API key
			APIKeyID:    apiKeyID,
			Endpoint:    endpoint,
			Requests:    1,
			WindowStart: windowStart,
			WindowEnd:   windowStart.Add(window),
		}
		return s.db.Create(&rateLimit).Error == nil, nil
	} else if err != nil {
		return false, err
	}

	// Check if limit exceeded
	if rateLimit.Requests >= int64(limit) {
		return false, nil
	}

	// Increment request count
	rateLimit.Requests++
	return s.db.Save(&rateLimit).Error == nil, nil
}

// API Usage Analytics

type APIUsage struct {
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"not null"`
	APIKeyID     uint   `gorm:"not null"`
	Endpoint     string `gorm:"not null"`
	Method       string `gorm:"not null"`
	StatusCode   int    `gorm:"not null"`
	ResponseTime int64  `gorm:"not null"` // in milliseconds
	UserAgent    string
	IPAddress    string
	CreatedAt    time.Time
}

func (s *APIService) LogAPIUsage(tenantID, apiKeyID uint, endpoint, method string, statusCode int, responseTime int64, userAgent, ipAddress string) error {
	usage := &APIUsage{
		TenantID:     tenantID,
		APIKeyID:     apiKeyID,
		Endpoint:     endpoint,
		Method:       method,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
	}

	return s.db.Create(usage).Error
}

func (s *APIService) GetAPIUsageStats(tenantID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	stats := make(map[string]interface{})

	// Total requests
	var totalRequests int64
	if err := s.db.Model(&APIUsage{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).Count(&totalRequests).Error; err != nil {
		return nil, err
	}
	stats["total_requests"] = totalRequests

	// Average response time
	var avgResponseTime float64
	if err := s.db.Model(&APIUsage{}).Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Select("AVG(response_time)").Scan(&avgResponseTime).Error; err != nil {
		return nil, err
	}
	stats["avg_response_time"] = avgResponseTime

	// Error rate
	var errorCount int64
	if err := s.db.Model(&APIUsage{}).Where("tenant_id = ? AND created_at >= ? AND status_code >= 400", tenantID, startDate).Count(&errorCount).Error; err != nil {
		return nil, err
	}
	errorRate := float64(errorCount) / float64(totalRequests) * 100
	stats["error_rate"] = errorRate

	// Top endpoints
	var topEndpoints []struct {
		Endpoint string `json:"endpoint"`
		Count    int64  `json:"count"`
	}
	if err := s.db.Model(&APIUsage{}).
		Select("endpoint, COUNT(*) as count").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startDate).
		Group("endpoint").
		Order("count DESC").
		Limit(10).
		Scan(&topEndpoints).Error; err != nil {
		return nil, err
	}
	stats["top_endpoints"] = topEndpoints

	return stats, nil
}

// Webhook Management

type Webhook struct {
	ID              uint   `gorm:"primaryKey"`
	TenantID        uint   `gorm:"not null"`
	Name            string `gorm:"not null"`
	URL             string `gorm:"not null"`
	Events          string `gorm:"not null"` // JSON array of events
	Secret          string `gorm:"not null"`
	IsActive        bool   `gorm:"default:true"`
	LastTriggeredAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (s *APIService) CreateWebhook(tenantID uint, name, url string, events []string) (*Webhook, error) {
	// Generate webhook secret
	secret, err := s.generateWebhookSecret()
	if err != nil {
		return nil, err
	}

	// Convert events to JSON
	eventsJSON, err := s.eventsToJSON(events)
	if err != nil {
		return nil, err
	}

	webhook := &Webhook{
		TenantID: tenantID,
		Name:     name,
		URL:      url,
		Events:   eventsJSON,
		Secret:   secret,
		IsActive: true,
	}

	if err := s.db.Create(webhook).Error; err != nil {
		logger.Error("Failed to create webhook", zap.Error(err))
		return nil, err
	}

	logger.Info("Webhook created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("name", name))

	return webhook, nil
}

func (s *APIService) GetWebhooks(tenantID uint) ([]Webhook, error) {
	var webhooks []Webhook
	err := s.db.Where("tenant_id = ?", tenantID).Find(&webhooks).Error
	return webhooks, err
}

func (s *APIService) TriggerWebhook(webhookID uint, eventType string, data interface{}) error {
	var webhook Webhook
	if err := s.db.First(&webhook, webhookID).Error; err != nil {
		return err
	}

	// Check if webhook is active and subscribes to this event
	if !webhook.IsActive {
		return fmt.Errorf("webhook is not active")
	}

	events, err := s.eventsFromJSON(webhook.Events)
	if err != nil {
		return err
	}

	// Check if webhook subscribes to this event
	if !s.containsEvent(events, eventType) {
		return fmt.Errorf("webhook does not subscribe to event: %s", eventType)
	}

	// Trigger webhook (this would be implemented by the webhook service)
	// For now, we'll just update the last triggered timestamp
	now := time.Now()
	webhook.LastTriggeredAt = &now
	s.db.Save(&webhook)

	logger.Info("Webhook triggered",
		zap.Uint("webhook_id", webhookID),
		zap.String("event_type", eventType))

	return nil
}

// API Documentation

type APIDocumentation struct {
	ID          uint   `gorm:"primaryKey"`
	TenantID    uint   `gorm:"not null"`
	Title       string `gorm:"not null"`
	Description string
	Version     string `gorm:"not null"`
	BaseURL     string `gorm:"not null"`
	Endpoints   string `gorm:"not null"` // JSON array of endpoints
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *APIService) CreateAPIDocumentation(tenantID uint, doc *APIDocumentation) error {
	doc.TenantID = tenantID

	if err := s.db.Create(doc).Error; err != nil {
		logger.Error("Failed to create API documentation", zap.Error(err))
		return err
	}

	logger.Info("API documentation created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("title", doc.Title))

	return nil
}

func (s *APIService) GetAPIDocumentation(tenantID uint) (*APIDocumentation, error) {
	var doc APIDocumentation
	err := s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).First(&doc).Error
	return &doc, err
}

// Private helper methods

func (s *APIService) generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(bytes), nil
}

func (s *APIService) generateAPISecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *APIService) generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(bytes), nil
}

func (s *APIService) permissionsToJSON(permissions []string) (string, error) {
	// Simple JSON conversion for permissions array
	json := "["
	for i, perm := range permissions {
		if i > 0 {
			json += ","
		}
		json += `"` + perm + `"`
	}
	json += "]"
	return json, nil
}

func (s *APIService) eventsToJSON(events []string) (string, error) {
	// Simple JSON conversion for events array
	json := "["
	for i, event := range events {
		if i > 0 {
			json += ","
		}
		json += `"` + event + `"`
	}
	json += "]"
	return json, nil
}

func (s *APIService) eventsFromJSON(eventsJSON string) ([]string, error) {
	// Simple JSON parsing for events array
	// In a real implementation, you'd use encoding/json
	// For now, we'll return a mock implementation
	_ = eventsJSON // TODO: Parse JSON string to extract events
	return []string{"incident.created", "incident.updated", "maintenance.scheduled"}, nil
}

func (s *APIService) containsEvent(events []string, event string) bool {
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}
