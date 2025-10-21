package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
)

type PricingPlanRepository interface {
	Create(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.SaaSPlan, error)
	FindAll(ctx context.Context) ([]models.SaaSPlan, error)
	FindByStatus(ctx context.Context, isActive bool) ([]models.SaaSPlan, error)
	Update(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pricingPlanRepository struct {
	db *gorm.DB
}

func NewPricingPlanRepository(db *gorm.DB) PricingPlanRepository {
	return &pricingPlanRepository{db: db}
}

func (r *pricingPlanRepository) Create(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error) {
	if plan == nil {
		return nil, errors.New("plan cannot be nil")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the plan
		if err := tx.Create(plan).Error; err != nil {
			return fmt.Errorf("failed to create plan: %w", err)
		}

		// Create associated pricing tiers if any
		if len(plan.PricingTiers) > 0 {
			for i := range plan.PricingTiers {
				plan.PricingTiers[i].PlanID = plan.ID
			}
			if err := tx.Create(&plan.PricingTiers).Error; err != nil {
				return fmt.Errorf("failed to create pricing tiers: %w", err)
			}
		}

		// Create associated plan features if any
		if len(plan.PlanFeatures) > 0 {
			for i := range plan.PlanFeatures {
				plan.PlanFeatures[i].PlanID = plan.ID
			}
			if err := tx.Create(&plan.PlanFeatures).Error; err != nil {
				return fmt.Errorf("failed to create plan features: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the plan with all associations
	return r.FindByID(ctx, plan.ID)
}

func (r *pricingPlanRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.SaaSPlan, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid plan ID")
	}

	var plan models.SaaSPlan
	err := r.db.WithContext(ctx).
		Preload("PricingTiers").
		Preload("PlanFeatures").
		First(&plan, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plan not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find plan: %w", err)
	}

	return &plan, nil
}

func (r *pricingPlanRepository) FindAll(ctx context.Context) ([]models.SaaSPlan, error) {
	var plans []models.SaaSPlan
	err := r.db.WithContext(ctx).
		Preload("PricingTiers").
		Preload("PlanFeatures").
		Order("display_order ASC").
		Find(&plans).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find all plans: %w", err)
	}

	return plans, nil
}

func (r *pricingPlanRepository) FindByStatus(ctx context.Context, isActive bool) ([]models.SaaSPlan, error) {
	var plans []models.SaaSPlan
	err := r.db.WithContext(ctx).
		Preload("PricingTiers").
		Preload("PlanFeatures").
		Where("is_active = ?", isActive).
		Order("display_order ASC").
		Find(&plans).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find plans by status: %w", err)
	}

	return plans, nil
}

func (r *pricingPlanRepository) Update(ctx context.Context, plan *models.SaaSPlan) (*models.SaaSPlan, error) {
	if plan == nil {
		return nil, errors.New("plan cannot be nil")
	}

	if plan.ID == uuid.Nil {
		return nil, errors.New("invalid plan ID")
	}

		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the plan
		if err := tx.Save(plan).Error; err != nil {
			return fmt.Errorf("failed to update plan: %w", err)
		}

		// Delete existing pricing tiers
		if err := tx.Where("plan_id = ?", plan.ID).Delete(&models.PricingTier{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing pricing tiers: %w", err)
		}

		// Create new pricing tiers if any exist
		if len(plan.PricingTiers) > 0 {
			for i := range plan.PricingTiers {
				plan.PricingTiers[i].PlanID = plan.ID
			}
			if err := tx.Create(&plan.PricingTiers).Error; err != nil {
				return fmt.Errorf("failed to create pricing tiers: %w", err)
			}
		}

		// Delete existing plan features
		if err := tx.Where("plan_id = ?", plan.ID).Delete(&models.PlanFeature{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing plan features: %w", err)
		}

		// Create new plan features if any exist
		if len(plan.PlanFeatures) > 0 {
			for i := range plan.PlanFeatures {
				plan.PlanFeatures[i].PlanID = plan.ID
			}
			if err := tx.Create(&plan.PlanFeatures).Error; err != nil {
				return fmt.Errorf("failed to create plan features: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the plan with all associations
	return r.FindByID(ctx, plan.ID)
}

func (r *pricingPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid plan ID")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete related pricing tiers
		if err := tx.Where("plan_id = ?", id).Delete(&models.PricingTier{}).Error; err != nil {
			return fmt.Errorf("failed to delete pricing tiers: %w", err)
		}

		// Delete related plan features
		if err := tx.Where("plan_id = ?", id).Delete(&models.PlanFeature{}).Error; err != nil {
			return fmt.Errorf("failed to delete plan features: %w", err)
		}

		// Delete the plan
		if err := tx.Delete(&models.SaaSPlan{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete plan: %w", err)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("plan not found: %w", err)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}
