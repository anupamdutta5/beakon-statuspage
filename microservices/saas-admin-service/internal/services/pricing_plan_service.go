package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/google/uuid"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/models"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/repositories"
	"gorm.io/gorm"
)

type PricingPlanService interface {
	CreatePlan(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*models.SaaSPlan, error)
	ListPlans(ctx context.Context, isActive *bool) ([]models.SaaSPlan, error)
	UpdatePlan(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error
}

type pricingPlanService struct {
	repo repositories.PricingPlanRepository
}

func NewPricingPlanService(repo repositories.PricingPlanRepository) PricingPlanService {
	return &pricingPlanService{repo: repo}
}

func (s *pricingPlanService) CreatePlan(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error) {
	// Validate required fields
	if plan.Name == "" {
		return nil, errors.New("plan name is required")
	}

	// Set default values if not provided
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.BillingInterval == "" {
		plan.BillingInterval = "monthly"
	}

	// Generate slug from name if not provided
	if plan.Slug == "" {
		plan.Slug = slug.Make(plan.Name)
	}

	// Ensure ID is generated
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}

	return s.repo.Create(ctx, plan)
}

func (s *pricingPlanService) GetPlanByID(ctx context.Context, id uuid.UUID) (*models.SaaSPlan, error) {
	plan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *pricingPlanService) ListPlans(ctx context.Context, isActive *bool) ([]models.SaaSPlan, error) {
	var plans []models.SaaSPlan
	var err error

	if isActive != nil {
		plans, err = s.repo.FindByStatus(ctx, *isActive)
	} else {
		plans, err = s.repo.FindAll(ctx)
	}

	if err != nil {
		return nil, err
	}

	return plans, nil
}

func (s *pricingPlanService) UpdatePlan(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error) {
	// Validate ID
	if plan.ID == uuid.Nil {
		return nil, errors.New("invalid plan ID")
	}

	// Check if plan exists
	existingPlan, err := s.repo.FindByID(ctx, plan.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, fmt.Errorf("failed to find plan: %w", err)
	}

	// Update fields
	existingPlan.Name = plan.Name
	existingPlan.Slug = plan.Slug
	existingPlan.Description = plan.Description
	existingPlan.Price = plan.Price
	existingPlan.Currency = plan.Currency
	existingPlan.BillingInterval = plan.BillingInterval
	existingPlan.MaxTenants = plan.MaxTenants
	existingPlan.MaxUsers = plan.MaxUsers
	existingPlan.MaxServices = plan.MaxServices
	existingPlan.MaxMonitors = plan.MaxMonitors
	existingPlan.MaxSubscribers = plan.MaxSubscribers
	existingPlan.MaxIncidents = plan.MaxIncidents
	existingPlan.MaxMaintenance = plan.MaxMaintenance
	existingPlan.CustomDomain = plan.CustomDomain
	existingPlan.WhiteLabel = plan.WhiteLabel
	existingPlan.API = plan.API
	existingPlan.Integrations = plan.Integrations
	existingPlan.Analytics = plan.Analytics
	existingPlan.Support = plan.Support
	existingPlan.Metadata = plan.Metadata

	return s.repo.Update(ctx, existingPlan)
}

func (s *pricingPlanService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid plan ID")
	}

	// Check if plan exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("plan not found")
		}
		return fmt.Errorf("failed to find plan: %w", err)
	}

	// Check if plan is in use by any tenants
	// This would require a call to the tenant service or a join query
	// For now, we'll assume a method exists to check this
	// if inUse, err := s.repo.IsPlanInUse(ctx, id); err != nil || inUse {
	// 	return errors.New("cannot delete plan that is in use")
	// }

	return s.repo.Delete(ctx, id)
}
