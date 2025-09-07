package api

import (
	"net/http"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/services"
	"github.com/gin-gonic/gin"
)

// PaymentHandlers handles payment-related API endpoints
type PaymentHandlers struct {
	paymentService *services.PaymentService
}

func NewPaymentHandlers(paymentService *services.PaymentService) *PaymentHandlers {
	return &PaymentHandlers{
		paymentService: paymentService,
	}
}

// CreateCheckoutSession creates a Stripe checkout session
func (h *PaymentHandlers) CreateCheckoutSession(c *gin.Context) {
	var req struct {
		PlanSlug   string `json:"plan_slug" binding:"required"`
		SuccessURL string `json:"success_url" binding:"required"`
		CancelURL  string `json:"cancel_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Create checkout session
	session, err := h.paymentService.CreateCheckoutSession(
		tenant.ID,
		req.PlanSlug,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": session.URL,
		"session_id":   session.ID,
	})
}

// HandleWebhook handles Stripe webhooks
func (h *PaymentHandlers) HandleWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	if err := h.paymentService.HandleWebhook(payload, signature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetCustomerPortalURL creates a Stripe customer portal session
func (h *PaymentHandlers) GetCustomerPortalURL(c *gin.Context) {
	var req struct {
		ReturnURL string `json:"return_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get customer portal URL
	portalURL, err := h.paymentService.GetCustomerPortalURL(tenant.ID, req.ReturnURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"portal_url": portalURL,
	})
}

// CancelSubscription cancels a subscription
func (h *PaymentHandlers) CancelSubscription(c *gin.Context) {
	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get subscription for tenant
	subscriptionService := services.NewSubscriptionService()
	subscription, err := subscriptionService.GetSubscriptionByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Cancel subscription
	if err := h.paymentService.CancelSubscription(subscription.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled successfully"})
}

// GetSubscriptionStatus gets the current subscription status
func (h *PaymentHandlers) GetSubscriptionStatus(c *gin.Context) {
	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get subscription for tenant
	subscriptionService := services.NewSubscriptionService()
	subscription, err := subscriptionService.GetSubscriptionByTenantID(tenant.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"has_subscription": false,
			"plan":             "free",
			"status":           "active",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_subscription":     true,
		"plan":                 subscription.Plan.Slug,
		"status":               subscription.Status,
		"current_period_end":   subscription.CurrentPeriodEnd,
		"cancel_at_period_end": subscription.CancelAtPeriodEnd,
	})
}
