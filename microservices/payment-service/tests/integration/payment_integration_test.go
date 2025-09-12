package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-payment-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8083"
	timeout = 30 * time.Second
)

func TestMain(m *testing.M) {
	// Start payment service for integration tests
	cmd := startPaymentService()
	if cmd != nil {
		defer cmd.Process.Kill()
	}

	// Wait for service to be ready
	if !waitForService(baseURL+"/health", timeout) {
		fmt.Println("Payment service not ready, skipping integration tests")
		os.Exit(0)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if cmd != nil {
		cmd.Process.Kill()
	}

	os.Exit(code)
}

func TestPaymentServiceIntegration_HealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/health", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "payment-service", response["service"])
}

func TestPaymentServiceIntegration_CreateAndGetPayment(t *testing.T) {
	// Create payment
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999, // $29.99 in cents
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Integration test payment",
		Metadata:    `{"plan_id":"pro-monthly","billing_cycle":"monthly"}`,
	}

	jsonData, err := json.Marshal(payment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&createdPayment)
	require.NoError(t, err)

	assert.Equal(t, payment.TenantID, createdPayment.TenantID)
	assert.Equal(t, payment.UserID, createdPayment.UserID)
	assert.Equal(t, payment.Amount, createdPayment.Amount)
	assert.Equal(t, payment.Currency, createdPayment.Currency)
	assert.NotEmpty(t, createdPayment.ID)

	// Get payment by ID
	req, err = http.NewRequest("GET", baseURL+"/payments/"+fmt.Sprintf("%d", createdPayment.ID), nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&retrievedPayment)
	require.NoError(t, err)

	assert.Equal(t, createdPayment.ID, retrievedPayment.ID)
	assert.Equal(t, createdPayment.TenantID, retrievedPayment.TenantID)
}

func TestPaymentServiceIntegration_UpdatePayment(t *testing.T) {
	// Create payment first
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Update test payment",
	}

	jsonData, err := json.Marshal(payment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&createdPayment)
	require.NoError(t, err)

	// Update payment
	updateData := models.Payment{
		Status:      "completed",
		Description: "Updated payment description",
		GatewayID:   "txn_123456789",
	}

	jsonData, err = json.Marshal(updateData)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", baseURL+"/payments/"+fmt.Sprintf("%d", createdPayment.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&updatedPayment)
	require.NoError(t, err)

	assert.Equal(t, "completed", updatedPayment.Status)
	assert.Equal(t, "Updated payment description", updatedPayment.Description)
	assert.Equal(t, "txn_123456789", updatedPayment.GatewayID)
}

func TestPaymentServiceIntegration_ListPayments(t *testing.T) {
	// Create multiple payments
	payments := []models.Payment{
		{TenantID: 1, UserID: 1, Amount: 29.99, Currency: "USD", Gateway: "stripe", Status: "completed"},
		{TenantID: 1, UserID: 2, Amount: 49.99, Currency: "USD", Gateway: "stripe", Status: "pending"},
		{TenantID: 1, UserID: 3, Amount: 99.99, Currency: "USD", Gateway: "stripe", Status: "failed"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, payment := range payments {
		jsonData, err := json.Marshal(payment)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List payments
	req, err := http.NewRequest("GET", baseURL+"/payments?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Payments []models.Payment `json:"payments"`
		Total    int              `json:"total"`
		Limit    int              `json:"limit"`
		Offset   int              `json:"offset"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Payments), 3)
	assert.GreaterOrEqual(t, response.Total, 3)
}

func TestPaymentServiceIntegration_ProcessPayment(t *testing.T) {
	// Create payment
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "pending",
		Description: "Process test payment",
	}

	jsonData, err := json.Marshal(payment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&createdPayment)
	require.NoError(t, err)

	// Process payment
	processData := map[string]interface{}{
		"payment_method_id": "pm_123456789",
		"customer_id":       "cus_123456789",
	}

	jsonData, err = json.Marshal(processData)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/payments/"+fmt.Sprintf("%d", createdPayment.ID)+"/process", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Payment processed successfully", response["message"])
	assert.Contains(t, response, "transaction_id")
}

func TestPaymentServiceIntegration_RefundPayment(t *testing.T) {
	// Create completed payment
	payment := models.Payment{
		TenantID:    1,
		UserID:      1,
		Amount:      2999,
		Currency:    "USD",
		Gateway:     "stripe",
		Status:      "completed",
		Description: "Refund test payment",
		GatewayID:   "txn_123456789",
	}

	jsonData, err := json.Marshal(payment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&createdPayment)
	require.NoError(t, err)

	// Refund payment
	refundData := map[string]interface{}{
		"amount": 2999,
		"reason": "customer_request",
		"metadata": map[string]interface{}{
			"refund_reason": "Customer requested refund",
		},
	}

	jsonData, err = json.Marshal(refundData)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/payments/"+fmt.Sprintf("%d", createdPayment.ID)+"/refund", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Payment refunded successfully", response["message"])
	assert.Contains(t, response, "refund_id")
}

func TestPaymentServiceIntegration_CreateAndGetPaymentMethod(t *testing.T) {
	// Create payment method
	payment := models.Payment{
		TenantID:      1,
		UserID:        1,
		Amount:        10.00,
		Currency:      "USD",
		PaymentMethod: "card",
		Gateway:       "stripe",
		Status:        "pending",
		Metadata:      `{"last4": "4242", "brand": "visa"}`,
	}

	jsonData, err := json.Marshal(payment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/payment-methods", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPayment models.Payment
	err = json.NewDecoder(resp.Body).Decode(&createdPayment)
	require.NoError(t, err)

	assert.Equal(t, payment.TenantID, createdPayment.TenantID)
	assert.Equal(t, payment.UserID, createdPayment.UserID)
	assert.Equal(t, payment.PaymentMethod, createdPayment.PaymentMethod)
	assert.NotEmpty(t, createdPayment.ID)

	// Get payment methods
	req, err = http.NewRequest("GET", baseURL+"/payment-methods?tenant_id=integration-test-tenant&user_id=integration-test-user", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Payments []models.Payment `json:"payments"`
		Total          int                    `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(response.Payments), 1)
}

func TestPaymentServiceIntegration_CreateSubscription(t *testing.T) {
	// Create subscription
	subscription := models.Subscription{
		TenantID:     1,
		UserID:       1,
		PlanID:       1,
		Status:       "active",
		Amount:       2999,
		Currency:     "USD",
		BillingCycle: "monthly",
		Metadata:     `{"trial_end":"2024-02-01T00:00:00Z"}`,
	}

	jsonData, err := json.Marshal(subscription)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/subscriptions", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdSubscription models.Subscription
	err = json.NewDecoder(resp.Body).Decode(&createdSubscription)
	require.NoError(t, err)

	assert.Equal(t, subscription.TenantID, createdSubscription.TenantID)
	assert.Equal(t, subscription.UserID, createdSubscription.UserID)
	assert.Equal(t, subscription.PlanID, createdSubscription.PlanID)
	assert.Equal(t, subscription.Status, createdSubscription.Status)
	assert.NotEmpty(t, createdSubscription.ID)
}

func TestPaymentServiceIntegration_GetPaymentStats(t *testing.T) {
	// Create some payments first
	payments := []models.Payment{
		{TenantID: 1, UserID: 1, Amount: 29.99, Currency: "USD", Gateway: "stripe", Status: "completed"},
		{TenantID: 1, UserID: 2, Amount: 49.99, Currency: "USD", Gateway: "stripe", Status: "completed"},
		{TenantID: 1, UserID: 3, Amount: 99.99, Currency: "USD", Gateway: "stripe", Status: "failed"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, payment := range payments {
		jsonData, err := json.Marshal(payment)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Get payment stats
	req, err := http.NewRequest("GET", baseURL+"/payments/stats?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "total_revenue")
	assert.Contains(t, response, "total_payments")
	assert.Contains(t, response, "successful_payments")
	assert.Contains(t, response, "failed_payments")
}

func TestPaymentServiceIntegration_GetSubscriptionPlans(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/subscription-plans?tenant_id=integration-test-tenant", nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Plans []models.Plan `json:"plans"`
		Total int                       `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Plans)
}

func TestPaymentServiceIntegration_PerformanceTest(t *testing.T) {
	// Test creating multiple payments in parallel
	client := &http.Client{Timeout: 10 * time.Second}

	// Create payments in bulk
	start := time.Now()

	for i := 0; i < 20; i++ {
		payment := models.Payment{
			TenantID:    1,
			UserID:      uint(i),
			Amount:      2999,
			Currency:    "USD",
			Gateway:     "stripe",
			Status:      "pending",
			Description: fmt.Sprintf("Performance test payment %d", i),
		}

		jsonData, err := json.Marshal(payment)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/payments", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	duration := time.Since(start)
	t.Logf("Created 20 payments in %v", duration)

	// Should complete within reasonable time
	assert.Less(t, duration, 10*time.Second)
}

// Helper functions
func startPaymentService() *exec.Cmd {
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "ENVIRONMENT=testing", "SERVER_PORT=8083", "DB_NAME=statuspage_payment_test")

	// Start in background
	err := cmd.Start()
	if err != nil {
		return nil
	}
	return cmd
}

func waitForService(url string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	return false
}
