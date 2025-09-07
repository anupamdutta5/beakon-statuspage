package api

import (
	"net/http"
	"strconv"

	"github.com/enterprise-status/statuspage/internal/api/middleware"
	"github.com/enterprise-status/statuspage/internal/services/payment"
	"github.com/gin-gonic/gin"
)

// PaymentHandlersV2 handles payment-related API endpoints with comprehensive payment system
type PaymentHandlersV2 struct {
	paymentService payment.PaymentService
}

// NewPaymentHandlersV2 creates new payment handlers
func NewPaymentHandlersV2(paymentService payment.PaymentService) *PaymentHandlersV2 {
	return &PaymentHandlersV2{
		paymentService: paymentService,
	}
}

// CreatePayment creates a new payment
func (h *PaymentHandlersV2) CreatePayment(c *gin.Context) {
	var req payment.PaymentRequest
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

	req.TenantID = tenant.ID

	// Create payment
	response, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPayment retrieves payment details
func (h *PaymentHandlersV2) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	response, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePayment updates payment details
func (h *PaymentHandlersV2) UpdatePayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.paymentService.UpdatePayment(c.Request.Context(), paymentID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelPayment cancels a payment
func (h *PaymentHandlersV2) CancelPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	response, err := h.paymentService.CancelPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefundPayment processes a refund
func (h *PaymentHandlersV2) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	var req payment.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.PaymentID = paymentID

	response, err := h.paymentService.RefundPayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateSubscription creates a subscription
func (h *PaymentHandlersV2) CreateSubscription(c *gin.Context) {
	var req payment.PaymentRequest
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

	req.TenantID = tenant.ID

	// Create subscription
	response, err := h.paymentService.CreateSubscription(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSubscription retrieves subscription details
func (h *PaymentHandlersV2) GetSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	response, err := h.paymentService.GetSubscription(c.Request.Context(), subscriptionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelSubscription cancels a subscription
func (h *PaymentHandlersV2) CancelSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	response, err := h.paymentService.CancelSubscription(c.Request.Context(), subscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessWebhook processes webhook events
func (h *PaymentHandlersV2) ProcessWebhook(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gateway is required"})
		return
	}

	// Read payload
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read payload"})
		return
	}

	// Get signature from header
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		signature = c.GetHeader("Stripe-Signature")
	}
	if signature == "" {
		signature = c.GetHeader("X-Razorpay-Signature")
	}

	// Process webhook
	event, err := h.paymentService.ProcessWebhook(c.Request.Context(), gateway, payload, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetSupportedMethods returns supported payment methods
func (h *PaymentHandlersV2) GetSupportedMethods(c *gin.Context) {
	methods := h.paymentService.GetSupportedMethods()
	c.JSON(http.StatusOK, gin.H{"methods": methods})
}

// GetGateways returns available gateways
func (h *PaymentHandlersV2) GetGateways(c *gin.Context) {
	gateways := h.paymentService.GetGateways()
	c.JSON(http.StatusOK, gin.H{"gateways": gateways})
}

// GetPaymentHistory returns payment history for a tenant
func (h *PaymentHandlersV2) GetPaymentHistory(c *gin.Context) {
	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get payment history
	payments, err := h.paymentService.GetPaymentHistory(c.Request.Context(), tenant.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// GetInvoiceHistory returns invoice history for a tenant
func (h *PaymentHandlersV2) GetInvoiceHistory(c *gin.Context) {
	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get invoice history
	invoices, err := h.paymentService.GetInvoiceHistory(c.Request.Context(), tenant.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

// GetPaymentMetrics returns payment metrics for a tenant
func (h *PaymentHandlersV2) GetPaymentMetrics(c *gin.Context) {
	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get period parameter
	period := c.DefaultQuery("period", "month")

	// Get payment metrics
	metrics, err := h.paymentService.GetPaymentMetrics(c.Request.Context(), tenant.ID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetConversionFunnel returns conversion funnel data for a tenant
func (h *PaymentHandlersV2) GetConversionFunnel(c *gin.Context) {
	// Get tenant from context
	tenant, exists := middleware.GetTenantFromContext(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get period parameter
	period := c.DefaultQuery("period", "month")

	// Get conversion funnel
	funnel, err := h.paymentService.GetConversionFunnel(c.Request.Context(), tenant.ID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, funnel)
}
