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

type SubscriptionService struct {
	db *gorm.DB
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		db: database.DB,
	}
}

// CreateSubscription creates a new subscription for a tenant
func (s *SubscriptionService) CreateSubscription(tenantID uint, planSlug string) (*models.Subscription, error) {
	// Get the plan
	var plan models.SubscriptionPlan
	if err := s.db.Where("slug = ? AND is_active = ?", planSlug, true).First(&plan).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Check if tenant already has an active subscription
	var existingSubscription models.Subscription
	if err := s.db.Where("tenant_id = ? AND status IN ?", tenantID, []string{"active", "trialing"}).First(&existingSubscription).Error; err == nil {
		return nil, fmt.Errorf("tenant already has an active subscription")
	}

	// Create new subscription
	now := time.Now()
	subscription := &models.Subscription{
		TenantID:           tenantID,
		PlanID:             plan.ID,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0), // 1 month from now
	}

	if err := s.db.Create(subscription).Error; err != nil {
		logger.Error("Failed to create subscription", zap.Error(err))
		return nil, err
	}

	// Initialize feature flags for the tenant based on the plan
	if err := s.initializeFeatureFlags(tenantID, &plan); err != nil {
		logger.Error("Failed to initialize feature flags", zap.Error(err))
		// Don't fail the subscription creation, just log the error
	}

	logger.Info("Subscription created successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("plan", planSlug))

	return subscription, nil
}

// UpgradeSubscription upgrades a tenant's subscription to a higher plan
func (s *SubscriptionService) UpgradeSubscription(tenantID uint, newPlanSlug string) error {
	// Get current subscription
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Get new plan
	var newPlan models.SubscriptionPlan
	if err := s.db.Where("slug = ? AND is_active = ?", newPlanSlug, true).First(&newPlan).Error; err != nil {
		return fmt.Errorf("new plan not found: %w", err)
	}

	// Check if it's actually an upgrade
	if newPlan.Price <= subscription.Plan.Price {
		return fmt.Errorf("new plan is not an upgrade")
	}

	// Update subscription
	subscription.PlanID = newPlan.ID
	subscription.UpdatedAt = time.Now()

	if err := s.db.Save(subscription).Error; err != nil {
		logger.Error("Failed to upgrade subscription", zap.Error(err))
		return err
	}

	// Update feature flags
	if err := s.initializeFeatureFlags(tenantID, &newPlan); err != nil {
		logger.Error("Failed to update feature flags after upgrade", zap.Error(err))
	}

	logger.Info("Subscription upgraded successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("new_plan", newPlanSlug))

	return nil
}

// DowngradeSubscription downgrades a tenant's subscription to a lower plan
func (s *SubscriptionService) DowngradeSubscription(tenantID uint, newPlanSlug string) error {
	// Get current subscription
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Get new plan
	var newPlan models.SubscriptionPlan
	if err := s.db.Where("slug = ? AND is_active = ?", newPlanSlug, true).First(&newPlan).Error; err != nil {
		return fmt.Errorf("new plan not found: %w", err)
	}

	// Check if it's actually a downgrade
	if newPlan.Price >= subscription.Plan.Price {
		return fmt.Errorf("new plan is not a downgrade")
	}

	// Check if tenant is within limits of new plan
	if err := s.validatePlanLimits(tenantID, &newPlan); err != nil {
		return fmt.Errorf("cannot downgrade: %w", err)
	}

	// Update subscription
	subscription.PlanID = newPlan.ID
	subscription.UpdatedAt = time.Now()

	if err := s.db.Save(subscription).Error; err != nil {
		logger.Error("Failed to downgrade subscription", zap.Error(err))
		return err
	}

	// Update feature flags
	if err := s.initializeFeatureFlags(tenantID, &newPlan); err != nil {
		logger.Error("Failed to update feature flags after downgrade", zap.Error(err))
	}

	logger.Info("Subscription downgraded successfully",
		zap.Uint("tenant_id", tenantID),
		zap.String("new_plan", newPlanSlug))

	return nil
}

// CancelSubscription cancels a tenant's subscription
func (s *SubscriptionService) CancelSubscription(tenantID uint) error {
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	now := time.Now()
	subscription.Status = "cancelled"
	subscription.CancelledAt = &now
	subscription.UpdatedAt = now

	if err := s.db.Save(subscription).Error; err != nil {
		logger.Error("Failed to cancel subscription", zap.Error(err))
		return err
	}

	logger.Info("Subscription cancelled successfully", zap.Uint("tenant_id", tenantID))
	return nil
}

// GetSubscriptionByTenantID gets the active subscription for a tenant
func (s *SubscriptionService) GetSubscriptionByTenantID(tenantID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.Preload("Plan").Where("tenant_id = ? AND status = ?", tenantID, "active").First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

// CheckFeatureAccess checks if a tenant has access to a specific feature
func (s *SubscriptionService) CheckFeatureAccess(tenantID uint, feature string) (bool, error) {
	// Get subscription
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		// If no subscription, check if it's a free feature
		return s.isFreeFeature(feature), nil
	}

	// Get feature flag
	var flag models.FeatureFlag
	if err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).First(&flag).Error; err != nil {
		// If no feature flag, check if it's included in the plan
		return s.isFeatureIncludedInPlan(subscription.PlanID, feature), nil
	}

	return flag.IsEnabled, nil
}

// CheckResourceLimits checks if a tenant is within their plan limits
func (s *SubscriptionService) CheckResourceLimits(tenantID uint, resource string, currentCount int) (bool, error) {
	subscription, err := s.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		// If no subscription, use free plan limits
		return s.checkFreePlanLimits(resource, currentCount), nil
	}

	return s.checkPlanLimits(subscription.PlanID, resource, currentCount)
}

// GetUsageMetrics gets current usage metrics for a tenant
func (s *SubscriptionService) GetUsageMetrics(tenantID uint) (map[string]int, error) {
	metrics := make(map[string]int)

	// Count services
	var servicesCount int64
	if err := s.db.Model(&models.Service{}).Where("tenant_id = ?", tenantID).Count(&servicesCount).Error; err != nil {
		return nil, err
	}
	metrics["services"] = int(servicesCount)

	// Count monitors
	var monitorsCount int64
	if err := s.db.Model(&models.Monitor{}).Where("tenant_id = ?", tenantID).Count(&monitorsCount).Error; err != nil {
		return nil, err
	}
	metrics["monitors"] = int(monitorsCount)

	// Count subscribers
	var subscribersCount int64
	if err := s.db.Model(&models.Subscriber{}).Where("tenant_id = ?", tenantID).Count(&subscribersCount).Error; err != nil {
		return nil, err
	}
	metrics["subscribers"] = int(subscribersCount)

	// Count incidents
	var incidentsCount int64
	if err := s.db.Model(&models.Incident{}).Where("tenant_id = ?", tenantID).Count(&incidentsCount).Error; err != nil {
		return nil, err
	}
	metrics["incidents"] = int(incidentsCount)

	// Count maintenance
	var maintenanceCount int64
	if err := s.db.Model(&models.Maintenance{}).Where("tenant_id = ?", tenantID).Count(&maintenanceCount).Error; err != nil {
		return nil, err
	}
	metrics["maintenance"] = int(maintenanceCount)

	return metrics, nil
}

// Private helper methods

func (s *SubscriptionService) initializeFeatureFlags(tenantID uint, plan *models.SubscriptionPlan) error {
	// Define feature flags based on plan
	features := map[string]bool{
		"custom_domain":       plan.CustomDomain,
		"api_access":          plan.API,
		"integrations":        plan.Integrations,
		"analytics":           plan.Analytics,
		"white_label":         plan.WhiteLabel,
		"priority_support":    plan.Support == "phone" || plan.Support == "chat",
		"sla_guarantee":       plan.Support == "phone",
		"advanced_monitoring": plan.Integrations,
	}

	for feature, enabled := range features {
		flag := &models.FeatureFlag{
			TenantID:  tenantID,
			Feature:   feature,
			IsEnabled: enabled,
			Config:    "",
		}

		// Use upsert to create or update
		if err := s.db.Where("tenant_id = ? AND feature = ?", tenantID, feature).
			Assign(models.FeatureFlag{IsEnabled: enabled}).
			FirstOrCreate(flag).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *SubscriptionService) validatePlanLimits(tenantID uint, plan *models.SubscriptionPlan) error {
	metrics, err := s.GetUsageMetrics(tenantID)
	if err != nil {
		return err
	}

	// Check services limit
	if metrics["services"] > plan.MaxServices {
		return fmt.Errorf("too many services (%d > %d)", metrics["services"], plan.MaxServices)
	}

	// Check monitors limit
	if metrics["monitors"] > plan.MaxMonitors {
		return fmt.Errorf("too many monitors (%d > %d)", metrics["monitors"], plan.MaxMonitors)
	}

	// Check subscribers limit
	if metrics["subscribers"] > plan.MaxSubscribers {
		return fmt.Errorf("too many subscribers (%d > %d)", metrics["subscribers"], plan.MaxSubscribers)
	}

	return nil
}

func (s *SubscriptionService) isFreeFeature(feature string) bool {
	// Define free features
	freeFeatures := map[string]bool{
		"basic_monitoring":       true,
		"email_notifications":    true,
		"incident_management":    true,
		"maintenance_scheduling": true,
	}
	return freeFeatures[feature]
}

func (s *SubscriptionService) isFeatureIncludedInPlan(planID uint, feature string) bool {
	var plan models.SubscriptionPlan
	if err := s.db.First(&plan, planID).Error; err != nil {
		return false
	}

	features := map[string]bool{
		"custom_domain":       plan.CustomDomain,
		"api_access":          plan.API,
		"integrations":        plan.Integrations,
		"analytics":           plan.Analytics,
		"white_label":         plan.WhiteLabel,
		"priority_support":    plan.Support == "phone" || plan.Support == "chat",
		"sla_guarantee":       plan.Support == "phone",
		"advanced_monitoring": plan.Integrations,
	}

	return features[feature]
}

func (s *SubscriptionService) checkFreePlanLimits(resource string, currentCount int) bool {
	limits := map[string]int{
		"services":    5,
		"monitors":    10,
		"subscribers": 100,
	}
	return currentCount < limits[resource]
}

func (s *SubscriptionService) checkPlanLimits(planID uint, resource string, currentCount int) (bool, error) {
	var plan models.SubscriptionPlan
	if err := s.db.First(&plan, planID).Error; err != nil {
		return false, err
	}

	limits := map[string]int{
		"services":    plan.MaxServices,
		"monitors":    plan.MaxMonitors,
		"subscribers": plan.MaxSubscribers,
	}

	limit, exists := limits[resource]
	if !exists {
		return true, nil // No limit for this resource
	}

	return currentCount < limit, nil
}
