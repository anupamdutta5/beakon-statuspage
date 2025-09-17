package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/models"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/services"
)

type PricingPlanHandler struct {
	service services.PricingPlanService
}

func NewPricingPlanHandler(service services.PricingPlanService) *PricingPlanHandler {
	return &PricingPlanHandler{service: service}
}

// CreatePricingPlan creates a new pricing plan
// @Summary Create a new pricing plan
// @Description Create a new pricing plan with the provided details
// @Tags pricing-plans
// @Accept json
// @Produce json
// @Param plan body models.SaaSPlan true "Pricing plan details"
// @Success 201 {object} models.SaaSPlan
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/plans [post]
func (h *PricingPlanHandler) CreatePricingPlan(c *gin.Context) {
	var plan models.SaaSPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	createdPlan, err := h.service.CreatePlan(c.Request.Context(), &plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdPlan)
}

// GetPricingPlan retrieves a pricing plan by ID
// @Summary Get a pricing plan
// @Description Get a pricing plan by its ID
// @Tags pricing-plans
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} models.SaaSPlan
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/plans/{id} [get]
func (h *PricingPlanHandler) GetPricingPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid plan ID"})
		return
	}

	plan, err := h.service.GetPlanByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// ListPricingPlans retrieves all pricing plans
// @Summary List all pricing plans
// @Description Get a list of all pricing plans
// @Tags pricing-plans
// @Produce json
// @Param is_active query bool false "Filter by active status"
// @Success 200 {array} models.SaaSPlan
// @Failure 500 {object} ErrorResponse
// @Router /admin/plans [get]
func (h *PricingPlanHandler) ListPricingPlans(c *gin.Context) {
	isActive := c.DefaultQuery("is_active", "")
	var active *bool
	if isActive != "" {
		val, err := strconv.ParseBool(isActive)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid is_active parameter"})
			return
		}
		active = &val
	}

	plans, err := h.service.ListPlans(c.Request.Context(), active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, plans)
}

// UpdatePricingPlan updates an existing pricing plan
// @Summary Update a pricing plan
// @Description Update an existing pricing plan
// @Tags pricing-plans
// @Accept json
// @Produce json
// @Param id path string true "Plan ID"
// @Param plan body models.SaaSPlan true "Updated plan details"
// @Success 200 {object} models.SaaSPlan
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/plans/{id} [put]
func (h *PricingPlanHandler) UpdatePricingPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid plan ID"})
		return
	}

	var plan models.SaaSPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Set the ID from the URL parameter
	plan.ID = id

	updatedPlan, err := h.service.UpdatePlan(c.Request.Context(), &plan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update plan"})
		return
	}

	c.JSON(http.StatusOK, updatedPlan)
}

// DeletePricingPlan deletes a pricing plan
// @Summary Delete a pricing plan
// @Description Delete a pricing plan by its ID
// @Tags pricing-plans
// @Produce json
// @Param id path string true "Plan ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/plans/{id} [delete]
func (h *PricingPlanHandler) DeletePricingPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid plan ID"})
		return
	}

	if err := h.service.DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ErrorResponse represents an error response
// swagger:model
// @name ErrorResponse
type ErrorResponse struct {
	Error string `json:"error"`
}
