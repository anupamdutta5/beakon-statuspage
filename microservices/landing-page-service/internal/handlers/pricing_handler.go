// Package handlers provides HTTP handlers for the Landing Page Service.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PricingHandler handles pricing-related HTTP requests.
type PricingHandler struct {
	db         *gorm.DB
	logger     *zap.Logger
	httpClient *http.Client
}

// NewPricingHandler creates a new pricing handler.
func NewPricingHandler(db *gorm.DB, logger *zap.Logger) *PricingHandler {
	return &PricingHandler{
		db:         db,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SyncPricingPlans syncs pricing plans from SaaS Admin Service.
func (h *PricingHandler) SyncPricingPlans(c *gin.Context) {
	h.logger.Info("Starting pricing sync from SaaS Admin Service")

	// Get SaaS Admin Service URL from environment
	saasAdminURL := os.Getenv("SAAS_ADMIN_SERVICE_URL")
	if saasAdminURL == "" {
		saasAdminURL = "http://localhost:8098" // Default (SaaS Admin runs on 8098)
	}

	// Call SaaS Admin Service to get public pricing plans
	pricingURL := saasAdminURL + "/api/v1/pricing/plans/public"
	req, err := http.NewRequestWithContext(context.Background(), "GET", pricingURL, nil)
	if err != nil {
		h.logger.Error("Failed to create pricing sync request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initiate pricing sync",
		})
		return
	}

	// Add service-to-service authentication
	req.Header.Set("Authorization", "Bearer service-token-"+os.Getenv("JWT_SECRET"))
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("Pricing sync request failed",
			zap.String("target", pricingURL),
			zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "SaaS Admin Service unavailable",
		})
		return
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		h.logger.Error("Pricing sync failed",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(bodyBytes)))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch pricing from SaaS Admin Service",
			"details": string(bodyBytes),
		})
		return
	}

	// Parse response
	var pricingResp struct {
		Success bool `json:"success"`
		Plans   []struct {
			PlanID          string   `json:"plan_id"`
			Name            string   `json:"name"`
			Slug            string   `json:"slug"`
			Description     string   `json:"description"`
			Price           float64  `json:"price"`
			YearlyPrice     float64  `json:"yearly_price"`
			DiscountPercent float64  `json:"discount_percent"`
			Currency        string   `json:"currency"`
			BillingInterval string   `json:"billing_interval"`
			Features        []string `json:"features"`
			IsPopular       bool     `json:"is_popular"`
			IsActive        bool     `json:"is_active"`
			ButtonText      string   `json:"button_text"`
			ButtonURL       string   `json:"button_url"`
			DisplayOrder    int      `json:"display_order"`
		} `json:"plans"`
		Count int `json:"count"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read pricing response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read pricing data",
		})
		return
	}

	if err := json.Unmarshal(bodyBytes, &pricingResp); err != nil {
		h.logger.Error("Failed to parse pricing response",
			zap.Error(err),
			zap.String("body", string(bodyBytes)))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse pricing data",
		})
		return
	}

	if !pricingResp.Success {
		h.logger.Error("Pricing sync returned unsuccessful response")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Pricing sync was not successful",
		})
		return
	}

	// Start transaction for atomic updates
	tx := h.db.Begin()
	if tx.Error != nil {
		h.logger.Error("Failed to start transaction", zap.Error(tx.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start database transaction",
		})
		return
	}

	syncedCount := 0
	updatedCount := 0
	createdCount := 0

	// Sync each plan
	for _, saasplan := range pricingResp.Plans {
		// Convert features slice to JSON string
		featuresJSON, err := json.Marshal(saasplan.Features)
		if err != nil {
			h.logger.Warn("Failed to marshal features",
				zap.String("plan_id", saasplan.PlanID),
				zap.Error(err))
			featuresJSON = []byte("[]")
		}

		// Check if plan already exists in landing page database
		var existingPlan models.PricingPlan
		err = tx.Where("slug = ?", saasplan.Slug).First(&existingPlan).Error

		if err == gorm.ErrRecordNotFound {
			// Create new plan
			newPlan := models.PricingPlan{
				PlanID:          saasplan.PlanID,
				Name:            saasplan.Name,
				Slug:            saasplan.Slug,
				Description:     saasplan.Description,
				Price:           saasplan.Price,
				Currency:        saasplan.Currency,
				BillingInterval: saasplan.BillingInterval,
				Features:        string(featuresJSON),
				IsPopular:       saasplan.IsPopular,
				IsActive:        saasplan.IsActive,
				ButtonText:      saasplan.ButtonText,
				ButtonURL:       saasplan.ButtonURL,
				SortOrder:       saasplan.DisplayOrder,
				Status:          "active",
			}

			if err := tx.Create(&newPlan).Error; err != nil {
				h.logger.Error("Failed to create pricing plan",
					zap.String("slug", saasplan.Slug),
					zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to create plan: %s", saasplan.Name),
				})
				return
			}

			createdCount++
			h.logger.Info("Created new pricing plan",
				zap.String("slug", saasplan.Slug),
				zap.String("name", saasplan.Name))

		} else if err != nil {
			h.logger.Error("Database error checking existing plan",
				zap.String("slug", saasplan.Slug),
				zap.Error(err))
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error during sync",
			})
			return
		} else {
			// Update existing plan
			existingPlan.PlanID = saasplan.PlanID
			existingPlan.Name = saasplan.Name
			existingPlan.Description = saasplan.Description
			existingPlan.Price = saasplan.Price
			existingPlan.Currency = saasplan.Currency
			existingPlan.BillingInterval = saasplan.BillingInterval
			existingPlan.Features = string(featuresJSON)
			existingPlan.IsPopular = saasplan.IsPopular
			existingPlan.IsActive = saasplan.IsActive
			existingPlan.ButtonText = saasplan.ButtonText
			existingPlan.ButtonURL = saasplan.ButtonURL
			existingPlan.SortOrder = saasplan.DisplayOrder

			if err := tx.Save(&existingPlan).Error; err != nil {
				h.logger.Error("Failed to update pricing plan",
					zap.String("slug", saasplan.Slug),
					zap.Error(err))
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to update plan: %s", saasplan.Name),
				})
				return
			}

			updatedCount++
			h.logger.Info("Updated existing pricing plan",
				zap.String("slug", saasplan.Slug),
				zap.String("name", saasplan.Name))
		}

		syncedCount++
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		h.logger.Error("Failed to commit pricing sync transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit pricing sync",
		})
		return
	}

	h.logger.Info("Pricing sync completed successfully",
		zap.Int("total_synced", syncedCount),
		zap.Int("created", createdCount),
		zap.Int("updated", updatedCount))

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Pricing successfully synced from SaaS Admin Service",
		"total_synced":  syncedCount,
		"created":       createdCount,
		"updated":       updatedCount,
		"source_count":  pricingResp.Count,
	})
}

// GetPricingPlans returns all active pricing plans for the landing page.
func (h *PricingHandler) GetPricingPlans(c *gin.Context) {
	h.logger.Info("Getting pricing plans for landing page")

	var plans []models.PricingPlan
	if err := h.db.Where("is_active = ? AND status = ?", true, "active").
		Order("sort_order ASC, created_at ASC").
		Find(&plans).Error; err != nil {
		h.logger.Error("Failed to get pricing plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve pricing plans",
		})
		return
	}

	// Transform plans for frontend
	frontendPlans := make([]gin.H, 0, len(plans))
	for _, plan := range plans {
		// Parse features from JSON
		var features []string
		if plan.Features != "" {
			if err := json.Unmarshal([]byte(plan.Features), &features); err != nil {
				h.logger.Warn("Failed to parse features for plan",
					zap.Uint("plan_id", plan.ID),
					zap.Error(err))
				features = []string{}
			}
		}

		frontendPlans = append(frontendPlans, gin.H{
			"id":               plan.ID,
			"name":             plan.Name,
			"slug":             plan.Slug,
			"description":      plan.Description,
			"price":            plan.Price,
			"currency":         plan.Currency,
			"billing_interval": plan.BillingInterval,
			"features":         features,
			"is_popular":       plan.IsPopular,
			"button_text":      plan.ButtonText,
			"button_url":       plan.ButtonURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plans":   frontendPlans,
		"count":   len(frontendPlans),
	})
}
