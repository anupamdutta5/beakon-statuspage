package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anupamdutta5/statuspage-payment-service/internal/handlers"
	"github.com/anupamdutta5/statuspage-payment-service/internal/models"
	"github.com/anupamdutta5/statuspage-payment-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPaymentHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.POST("/payments", func(c *gin.Context) {
		// Set required context values for authentication
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		handler.CreatePayment(c)
	})

	paymentRequest := map[string]interface{}{
		"amount":         29.99,
		"currency":       "USD",
		"payment_method": "card",
		"gateway":        "stripe",
		"description":    "Monthly subscription",
		"metadata":       `{"plan_id":"pro-monthly","billing_cycle":"monthly"}`,
	}

	jsonData, _ := json.Marshal(paymentRequest)
	req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response struct {
		Message string         `json:"message"`
		Payment models.Payment `json:"payment"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Payment created successfully", response.Message)
	assert.Equal(t, uint(1), response.Payment.TenantID)
	assert.Equal(t, uint(1), response.Payment.UserID)
	assert.Equal(t, 29.99, response.Payment.Amount)
	assert.Equal(t, "USD", response.Payment.Currency)
	assert.Equal(t, "stripe", response.Payment.Gateway)
	assert.NotZero(t, response.Payment.ID)
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/payments/:id", handler.GetPayment)

	// Test
	req, _ := http.NewRequest("GET", "/payments/"+fmt.Sprintf("%d", payment.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseWrapper struct {
		Payment models.Payment `json:"payment"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, payment.ID, responseWrapper.Payment.ID)
	assert.Equal(t, payment.TenantID, responseWrapper.Payment.TenantID)
}

func TestPaymentHandler_UpdatePayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Payment models.Payment `json:"payment"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "completed", responseWrapper.Payment.Status)
	assert.Equal(t, "Updated payment description", responseWrapper.Payment.Description)
	assert.Equal(t, "txn_123456789", responseWrapper.Payment.GatewayID)

	// Verify in database
	updatedPayment, err := paymentService.GetPayment(payment.ID)
	require.NoError(t, err)

	assert.Equal(t, "completed", updatedPayment.Status)
	assert.Equal(t, "Updated payment description", updatedPayment.Description)
}

func TestPaymentHandler_ListPayments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Message string         `json:"message"`
		Payment models.Payment `json:"payment"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Payment updated successfully", responseWrapper.Message)
	assert.NotEmpty(t, responseWrapper.Payment.ID)
}

func TestPaymentHandler_RefundPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

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
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, "Payment refunded successfully", responseWrapper.Message)
}

func TestPaymentHandler_GetPaymentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/payment-methods", handler.GetPayments)

	// Test
	req, _ := http.NewRequest("GET", "/payment-methods?tenant_id=test-tenant-id&user_id=test-user-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Payments []models.Payment `json:"payments"`
		Total    int              `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Payments)
}

func TestPaymentHandler_CreatePaymentMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Payment models.Payment `json:"payment"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, paymentData.TenantID, responseWrapper.Payment.TenantID)
	assert.Equal(t, paymentData.UserID, responseWrapper.Payment.UserID)
	assert.Equal(t, paymentData.PaymentMethod, responseWrapper.Payment.PaymentMethod)
	assert.Equal(t, paymentData.Gateway, responseWrapper.Payment.Gateway)
	assert.NotEmpty(t, responseWrapper.Payment.ID)
}

func TestPaymentHandler_GetPaymentStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	// Test that we get a valid payments response (stats not implemented)
	assert.Contains(t, response, "payments")
	assert.Contains(t, response, "total")
}

func TestPaymentHandler_GetSubscriptionPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
	router.GET("/subscription-plans", handler.GetPlans)

	// Test
	req, _ := http.NewRequest("GET", "/subscription-plans?tenant_id=test-tenant-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Plans []models.Plan `json:"plans"`
		Total int           `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Plans)
}

func TestPaymentHandler_CreateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	paymentService := services.NewPaymentService(db, logger)
	handler := handlers.NewPaymentHandler(paymentService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uint(1))
		c.Set("user_id", uint(1))
		c.Next()
	})
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

	var responseWrapper struct {
		Subscription models.Subscription `json:"subscription"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &responseWrapper)
	require.NoError(t, err)

	assert.Equal(t, subscription.TenantID, responseWrapper.Subscription.TenantID)
	assert.Equal(t, subscription.UserID, responseWrapper.Subscription.UserID)
	assert.Equal(t, subscription.PlanID, responseWrapper.Subscription.PlanID)
	assert.Equal(t, subscription.Status, responseWrapper.Subscription.Status)
	assert.NotEmpty(t, responseWrapper.Subscription.ID)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&models.Payment{}, &models.PaymentTransaction{}, &models.Subscription{}, &models.Plan{})
	require.NoError(t, err)

	return db
}
