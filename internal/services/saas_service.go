package services

import (
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/models"
	"github.com/enterprise-status/statuspage/pkg/database"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SaaSService struct {
	db *gorm.DB
}

func NewSaaSService() *SaaSService {
	return &SaaSService{
		db: database.DB,
	}
}

// Tenant Management

func (s *SaaSService) CreateTenant(tenant *models.Tenant) error {
	if err := s.db.Create(tenant).Error; err != nil {
		logger.Error("Failed to create tenant", zap.Error(err))
		return err
	}

	// Create default branding for the tenant
	branding := &models.Branding{
		TenantID:       tenant.ID,
		CompanyName:    tenant.Name,
		PrimaryColor:   tenant.PrimaryColor,
		SecondaryColor: tenant.SecondaryColor,
		FooterText:     tenant.FooterText,
	}

	if err := s.db.Create(branding).Error; err != nil {
		logger.Error("Failed to create default branding for tenant", zap.Error(err))
		return err
	}

	logger.Info("Tenant created successfully", zap.Uint("tenant_id", tenant.ID))
	return nil
}

func (s *SaaSService) GetTenantByID(id uint) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.Preload("CreatedByUser").First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *SaaSService) GetTenantBySlug(slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.Where("slug = ?", slug).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *SaaSService) GetTenantByDomain(domain string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *SaaSService) GetTenantBySubdomain(subdomain string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.Where("subdomain = ?", subdomain).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *SaaSService) GetAllTenants(limit, offset int) ([]models.Tenant, error) {
	var tenants []models.Tenant
	query := s.db.Preload("CreatedByUser")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *SaaSService) UpdateTenant(tenant *models.Tenant) error {
	if err := s.db.Save(tenant).Error; err != nil {
		logger.Error("Failed to update tenant", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) DeleteTenant(id uint) error {
	// Start a transaction to ensure data consistency
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all related data
	if err := tx.Where("tenant_id = ?", id).Delete(&models.Service{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tenant_id = ?", id).Delete(&models.Incident{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tenant_id = ?", id).Delete(&models.Maintenance{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tenant_id = ?", id).Delete(&models.Subscriber{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tenant_id = ?", id).Delete(&models.Integration{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tenant_id = ?", id).Delete(&models.Branding{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tenant_id = ?", id).Delete(&models.PrivatePage{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Finally delete the tenant
	if err := tx.Delete(&models.Tenant{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Subscription Plan Management

func (s *SaaSService) CreateSubscriptionPlan(plan *models.SubscriptionPlan) error {
	if err := s.db.Create(plan).Error; err != nil {
		logger.Error("Failed to create subscription plan", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetSubscriptionPlanByID(id uint) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	if err := s.db.First(&plan, id).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *SaaSService) GetSubscriptionPlanBySlug(slug string) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	if err := s.db.Where("slug = ? AND is_active = ?", slug, true).First(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *SaaSService) GetAllSubscriptionPlans() ([]models.SubscriptionPlan, error) {
	var plans []models.SubscriptionPlan
	if err := s.db.Where("is_active = ?", true).Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *SaaSService) UpdateSubscriptionPlan(plan *models.SubscriptionPlan) error {
	if err := s.db.Save(plan).Error; err != nil {
		logger.Error("Failed to update subscription plan", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) DeleteSubscriptionPlan(id uint) error {
	if err := s.db.Delete(&models.SubscriptionPlan{}, id).Error; err != nil {
		logger.Error("Failed to delete subscription plan", zap.Error(err))
		return err
	}
	return nil
}

// Subscription Management

func (s *SaaSService) CreateSubscription(subscription *models.Subscription) error {
	if err := s.db.Create(subscription).Error; err != nil {
		logger.Error("Failed to create subscription", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetSubscriptionByTenantID(tenantID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.Preload("Plan").Where("tenant_id = ?", tenantID).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (s *SaaSService) UpdateSubscription(subscription *models.Subscription) error {
	if err := s.db.Save(subscription).Error; err != nil {
		logger.Error("Failed to update subscription", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) CancelSubscription(tenantID uint) error {
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return err
	}

	now := time.Now()
	subscription.Status = "cancelled"
	subscription.CancelledAt = &now

	return s.UpdateSubscription(subscription)
}

// Feature Flag Management

func (s *SaaSService) GetFeatureFlag(tenantID uint, feature string) (*models.FeatureFlag, error) {
	var flag models.FeatureFlag
	if err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).First(&flag).Error; err != nil {
		return nil, err
	}
	return &flag, nil
}

func (s *SaaSService) SetFeatureFlag(tenantID uint, feature string, isEnabled bool, config string) error {
	flag := &models.FeatureFlag{
		TenantID:  tenantID,
		Feature:   feature,
		IsEnabled: isEnabled,
		Config:    config,
	}

	// Use upsert to create or update
	if err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).
		Assign(models.FeatureFlag{IsEnabled: isEnabled, Config: config}).
		FirstOrCreate(flag).Error; err != nil {
		logger.Error("Failed to set feature flag", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetTenantFeatureFlags(tenantID uint) ([]models.FeatureFlag, error) {
	var flags []models.FeatureFlag
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&flags).Error; err != nil {
		return nil, err
	}
	return flags, nil
}

// Usage Metrics

func (s *SaaSService) RecordUsageMetrics(tenantID uint, date time.Time, metrics map[string]int) error {
	usage := &models.UsageMetrics{
		TenantID:         tenantID,
		Date:             date,
		ServicesCount:    metrics["services"],
		MonitorsCount:    metrics["monitors"],
		SubscribersCount: metrics["subscribers"],
		IncidentsCount:   metrics["incidents"],
		MaintenanceCount: metrics["maintenance"],
		APIRequests:      metrics["api_requests"],
		PageViews:        metrics["page_views"],
	}

	// Use upsert to create or update daily metrics
	if err := s.db.Where("tenant_id = ? AND date = ?", tenantID, date).
		Assign(usage).
		FirstOrCreate(usage).Error; err != nil {
		logger.Error("Failed to record usage metrics", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetUsageMetrics(tenantID uint, startDate, endDate time.Time) ([]models.UsageMetrics, error) {
	var metrics []models.UsageMetrics
	if err := s.db.Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("date ASC").Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

// Billing Events

func (s *SaaSService) RecordBillingEvent(tenantID uint, eventType string, amount float64, currency, externalID, metadata string) error {
	event := &models.BillingEvent{
		TenantID:    tenantID,
		EventType:   eventType,
		Amount:      amount,
		Currency:    currency,
		ExternalID:  externalID,
		Metadata:    metadata,
		ProcessedAt: time.Now(),
	}

	if err := s.db.Create(event).Error; err != nil {
		logger.Error("Failed to record billing event", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetBillingEvents(tenantID uint, limit, offset int) ([]models.BillingEvent, error) {
	var events []models.BillingEvent
	query := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// Admin Settings

func (s *SaaSService) GetAdminSetting(key string) (*models.AdminSettings, error) {
	var setting models.AdminSettings
	if err := s.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (s *SaaSService) SetAdminSetting(key, value, description, settingType string, isPublic bool) error {
	setting := &models.AdminSettings{
		Key:         key,
		Value:       value,
		Description: description,
		Type:        settingType,
		IsPublic:    isPublic,
	}

	// Use upsert to create or update
	if err := s.db.Where("key = ?", key).
		Assign(models.AdminSettings{Value: value, Description: description, Type: settingType, IsPublic: isPublic}).
		FirstOrCreate(setting).Error; err != nil {
		logger.Error("Failed to set admin setting", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetAllAdminSettings() ([]models.AdminSettings, error) {
	var settings []models.AdminSettings
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// System Notifications

func (s *SaaSService) CreateSystemNotification(notification *models.SystemNotification) error {
	if err := s.db.Create(notification).Error; err != nil {
		logger.Error("Failed to create system notification", zap.Error(err))
		return err
	}
	return nil
}

func (s *SaaSService) GetActiveSystemNotifications(tenantID *uint) ([]models.SystemNotification, error) {
	var notifications []models.SystemNotification
	now := time.Now()

	query := s.db.Where("is_active = ? AND start_at <= ?", true, now)

	query = query.Where("end_at IS NULL OR end_at >= ?", now)

	// If tenantID is provided, filter for that tenant or global notifications
	if tenantID != nil {
		query = query.Where("target_tenants = '' OR target_tenants LIKE ?", fmt.Sprintf("%%\"%d\"%%", *tenantID))
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// Analytics and Reporting

func (s *SaaSService) GetTenantAnalytics(tenantID uint, days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Get usage metrics
	metrics, err := s.GetUsageMetrics(tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totals := map[string]int{
		"services":     0,
		"monitors":     0,
		"subscribers":  0,
		"incidents":    0,
		"maintenance":  0,
		"api_requests": 0,
		"page_views":   0,
	}

	for _, metric := range metrics {
		totals["services"] += metric.ServicesCount
		totals["monitors"] += metric.MonitorsCount
		totals["subscribers"] += metric.SubscribersCount
		totals["incidents"] += metric.IncidentsCount
		totals["maintenance"] += metric.MaintenanceCount
		totals["api_requests"] += metric.APIRequests
		totals["page_views"] += metric.PageViews
	}

	// Get subscription info
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		// Subscription might not exist yet
		subscription = nil
	}

	analytics := map[string]interface{}{
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
			"days":       days,
		},
		"usage":        totals,
		"subscription": subscription,
		"metrics":      metrics,
	}

	return analytics, nil
}

// Initialize default subscription plans
func (s *SaaSService) InitializeDefaultPlans() error {
	plans := []models.SubscriptionPlan{
		{
			Name:            "Free",
			Slug:            "free",
			Description:     "Perfect for small teams and personal projects",
			Price:           0,
			Currency:        "USD",
			BillingInterval: "monthly",
			MaxServices:     5,
			MaxMonitors:     10,
			MaxSubscribers:  100,
			MaxIncidents:    10,
			MaxMaintenance:  5,
			CustomDomain:    false,
			WhiteLabel:      false,
			API:             false,
			Integrations:    false,
			Analytics:       false,
			Support:         "email",
			IsActive:        true,
			Features:        `["basic_monitoring", "email_notifications"]`,
		},
		{
			Name:            "Pro",
			Slug:            "pro",
			Description:     "Advanced features for growing businesses",
			Price:           29,
			Currency:        "USD",
			BillingInterval: "monthly",
			MaxServices:     25,
			MaxMonitors:     100,
			MaxSubscribers:  1000,
			MaxIncidents:    100,
			MaxMaintenance:  50,
			CustomDomain:    true,
			WhiteLabel:      false,
			API:             true,
			Integrations:    true,
			Analytics:       true,
			Support:         "chat",
			IsActive:        true,
			Features:        `["advanced_monitoring", "custom_domains", "api_access", "integrations", "analytics", "priority_support"]`,
		},
		{
			Name:            "Enterprise",
			Slug:            "enterprise",
			Description:     "Full-featured solution for large organizations",
			Price:           99,
			Currency:        "USD",
			BillingInterval: "monthly",
			MaxServices:     -1, // Unlimited
			MaxMonitors:     -1, // Unlimited
			MaxSubscribers:  -1, // Unlimited
			MaxIncidents:    -1, // Unlimited
			MaxMaintenance:  -1, // Unlimited
			CustomDomain:    true,
			WhiteLabel:      true,
			API:             true,
			Integrations:    true,
			Analytics:       true,
			Support:         "phone",
			IsActive:        true,
			Features:        `["unlimited_monitoring", "white_labeling", "custom_domains", "api_access", "all_integrations", "advanced_analytics", "phone_support", "sla_guarantee"]`,
		},
	}

	for _, plan := range plans {
		// Check if plan already exists
		var existingPlan models.SubscriptionPlan
		if err := s.db.Where("slug = ?", plan.Slug).First(&existingPlan).Error; err != nil {
			// Plan doesn't exist, create it
			if err := s.db.Create(&plan).Error; err != nil {
				logger.Error("Failed to create default plan", zap.String("plan", plan.Slug), zap.Error(err))
				return err
			}
		}
	}

	logger.Info("Default subscription plans initialized")
	return nil
}
