package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-payment-service/internal/handlers"
	"github.com/enterprise-status/statuspage-payment-service/internal/models"
	"github.com/enterprise-status/statuspage-payment-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPaymentHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.GET("/health", handler.Health)

	// Test
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "payment-service", response["service"])
}

func TestPaymentHandler_CreatePayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.POST("/payments", handler.CreatePayment)

	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      29.99,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Monthly subscription",
		Metadata:    `{"plan_id":"pro-monthly","billing_cycle":"monthly"}`,
	}

	jsonData, _ := json.Marshal(payment)
	req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Payment
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, payment.TenantID, response.TenantID)
	assert.Equal(t, payment.UserID, response.UserID)
	assert.Equal(t, payment.Amount, response.Amount)
	assert.Equal(t, payment.Currency, response.Currency)
	assert.Equal(t, payment.Gateway, response.Gateway)
	assert.NotEmpty(t, response.ID)
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	// Create a payment first
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Test payment",
	}
	err := paymentService.CreatePayment(&payment)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/payments/:id", handler.GetPayment)

	// Test
	req, _ := http.NewRequest("GET", "/payments/"+fmt.Sprintf("%d", payment.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Payment
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, payment.ID, response.ID)
	assert.Equal(t, payment.TenantID, response.TenantID)
}

func TestPaymentHandler_UpdatePayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	// Create a payment first
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Test payment",
	}
	err := paymentService.CreatePayment(&payment)
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/payments/:id", handler.UpdatePayment)

	// Update data
	updateData := models.Payment{
		Status:      "completed",
		Description: "Updated payment description",
		GatewayID:   "txn_123456789",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/payments/"+fmt.Sprintf("%d", payment.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Payment
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "completed", response.Status)
	assert.Equal(t, "Updated payment description", response.Description)
	assert.Equal(t, "txn_123456789", response.GatewayID)

	// Verify in database
	updatedPayment, err := paymentService.GetPayment(payment.ID)
	require.NoError(t, err)

	assert.Equal(t, "completed", updatedPayment.Status)
	assert.Equal(t, "Updated payment description", updatedPayment.Description)
}

func TestPaymentHandler_ListPayments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	// Create multiple payments
	payments := []models.Payment{
		{TenantID: 1, UserID: 1, Amount: 29.99, Currency: "USD", Gateway: "stripe", Status: "completed"},
		{TenantID: 1, UserID: 2, Amount: 49.99, Currency: "USD", Gateway: "stripe", Status: "pending"},
		{TenantID: 1, UserID: 3, Amount: 99.99, Currency: "USD", Gateway: "stripe", Status: "failed"},
	}

	for _, payment := range payments {
		err := paymentService.CreatePayment(&payment)
		require.NoError(t, err)
	}

	router := gin.New()
	router.GET("/payments", handler.GetPayments)

	// Test
	req, _ := http.NewRequest("GET", "/payments?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Payments []models.Payment `json:"payments"`
		Total    int              `json:"total"`
		Limit    int              `json:"limit"`
		Offset   int              `json:"offset"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Payments), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestPaymentHandler_ProcessPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	// Create a payment first
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Test payment",
	}
	err := paymentService.CreatePayment(&payment)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/payments/:id/process", handler.UpdatePayment)

	// Process payment
	processData := map[string]interface{}{
		"payment_method_id": "pm_123456789",
		"customer_id":       "cus_123456789",
	}

	jsonData, _ := json.Marshal(processData)
	req, _ := http.NewRequest("POST", "/payments/"+fmt.Sprintf("%d", payment.ID)+"/process", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Payment processed successfully", response["message"])
	assert.Contains(t, response, "transaction_id")
}

func TestPaymentHandler_RefundPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	// Create a completed payment first
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "completed",
		Description: "Test payment",
		GatewayID:   "txn_123456789",
	}
	err := paymentService.CreatePayment(&payment)
	require.NoError(t, err)

	router := gin.New()
	router.POST("/payments/:id/refund", handler.RefundPayment)

	// Refund payment
	refundData := map[string]interface{}{
		"amount": 2999,
		"reason": "customer_request",
		"metadata": map[string]interface{}{
			"refund_reason": "Customer requested refund",
		},
	}

	jsonData, _ := json.Marshal(refundData)
	req, _ := http.NewRequest("POST", "/payments/"+fmt.Sprintf("%d", payment.ID)+"/refund", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Payment refunded successfully", response["message"])
	assert.Contains(t, response, "refund_id")
}

func TestPaymentHandler_GetPaymentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.GET("/payment-methods", handler.GetPayments)

	// Test
	req, _ := http.NewRequest("GET", "/payment-methods?tenant_id=test-tenant-id&user_id=test-user-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Payments []models.Payment `json:"payments"`
		Total          int                    `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Payments)
}

func TestPaymentHandler_CreatePaymentMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.POST("/payment-methods", handler.CreatePayment)

	_ = models.Payment{
		TenantID:      1,
		UserID:        1,
		Amount:        10.00,
		Currency:      "USD",
		PaymentMethod: "card",
		Gateway:       "stripe",
		Status:        "pending",
		Metadata:      `{"last4": "4242", "brand": "visa"}`,
	}

	paymentData := models.Payment{
		TenantID:      1,
		UserID:        1,
		Amount:        10.00,
		Currency:      "USD",
		PaymentMethod: "card",
		Gateway:       "stripe",
		Status:        "pending",
		Metadata:      `{"last4": "4242", "brand": "visa"}`,
	}
	jsonData, _ := json.Marshal(paymentData)
	req, _ := http.NewRequest("POST", "/payment-methods", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Payment
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, paymentData.TenantID, response.TenantID)
	assert.Equal(t, paymentData.UserID, response.UserID)
	assert.Equal(t, paymentData.PaymentMethod, response.PaymentMethod)
	assert.Equal(t, paymentData.Gateway, response.Gateway)
	assert.NotEmpty(t, response.ID)
}

func TestPaymentHandler_GetPaymentStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.GET("/payments/stats", handler.GetPayments)

	// Test
	req, _ := http.NewRequest("GET", "/payments/stats?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_revenue")
	assert.Contains(t, response, "total_payments")
	assert.Contains(t, response, "successful_payments")
	assert.Contains(t, response, "failed_payments")
}

func TestPaymentHandler_GetSubscriptionPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.GET("/subscription-plans", handler.GetPlans)

	// Test
	req, _ := http.NewRequest("GET", "/subscription-plans?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Plans []models.Plan `json:"plans"`
		Total int                       `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Plans)
}

func TestPaymentHandler_CreateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, nil)
	handler := handlers.NewPaymentHandler(paymentService, nil)

	router := gin.New()
	router.POST("/subscriptions", handler.CreateSubscription)

	subscription := models.Subscription{
		TenantID:           1,
		UserID:             1,
		PlanID:             1,
		Status:             "active",
		Amount:             29.99,
		Currency:           "USD",
		BillingCycle:       "monthly",
		StartedAt:          time.Now(),
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		NextBillingDate:    time.Now().AddDate(0, 1, 0),
		Gateway:            "stripe",
	}

	jsonData, _ := json.Marshal(subscription)
	req, _ := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, subscription.TenantID, response.TenantID)
	assert.Equal(t, subscription.UserID, response.UserID)
	assert.Equal(t, subscription.PlanID, response.PlanID)
	assert.Equal(t, subscription.Status, response.Status)
	assert.NotEmpty(t, response.ID)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.Payment{}, &models.Subscription{}, &models.Plan{})
	require.NoError(t, err)

	return db
}
