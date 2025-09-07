package gateways

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/services/payment"
)

// RazorpayGateway implements PaymentGateway for Razorpay
type RazorpayGateway struct {
	config     *config.PaymentGateway
	keyID      string
	keySecret  string
	baseURL    string
	httpClient *http.Client
}

// RazorpayPaymentRequest represents Razorpay payment request
type RazorpayPaymentRequest struct {
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Receipt        string            `json:"receipt"`
	PaymentCapture bool              `json:"payment_capture"`
	Notes          map[string]string `json:"notes,omitempty"`
}

// RazorpayPaymentResponse represents Razorpay payment response
type RazorpayPaymentResponse struct {
	ID               string                 `json:"id"`
	Entity           string                 `json:"entity"`
	Amount           int64                  `json:"amount"`
	Currency         string                 `json:"currency"`
	Status           string                 `json:"status"`
	OrderID          string                 `json:"order_id"`
	InvoiceID        string                 `json:"invoice_id"`
	International    bool                   `json:"international"`
	Method           string                 `json:"method"`
	AmountRefunded   int64                  `json:"amount_refunded"`
	RefundStatus     string                 `json:"refund_status"`
	Captured         bool                   `json:"captured"`
	Description      string                 `json:"description"`
	CardID           string                 `json:"card_id"`
	Bank             string                 `json:"bank"`
	Wallet           string                 `json:"wallet"`
	VPA              string                 `json:"vpa"`
	Email            string                 `json:"email"`
	Contact          string                 `json:"contact"`
	Notes            map[string]string      `json:"notes"`
	Fee              int64                  `json:"fee"`
	Tax              int64                  `json:"tax"`
	ErrorCode        string                 `json:"error_code"`
	ErrorDescription string                 `json:"error_description"`
	ErrorSource      string                 `json:"error_source"`
	ErrorStep        string                 `json:"error_step"`
	ErrorReason      string                 `json:"error_reason"`
	AcquirerData     map[string]interface{} `json:"acquirer_data"`
	CreatedAt        int64                  `json:"created_at"`
}

// RazorpayOrderRequest represents Razorpay order request
type RazorpayOrderRequest struct {
	Amount         int64             `json:"amount"`
	Currency       string            `json:"currency"`
	Receipt        string            `json:"receipt"`
	PaymentCapture bool              `json:"payment_capture"`
	Notes          map[string]string `json:"notes,omitempty"`
}

// RazorpayOrderResponse represents Razorpay order response
type RazorpayOrderResponse struct {
	ID         string            `json:"id"`
	Entity     string            `json:"entity"`
	Amount     int64             `json:"amount"`
	AmountPaid int64             `json:"amount_paid"`
	AmountDue  int64             `json:"amount_due"`
	Currency   string            `json:"currency"`
	Receipt    string            `json:"receipt"`
	Status     string            `json:"status"`
	Attempts   int               `json:"attempts"`
	Notes      map[string]string `json:"notes"`
	CreatedAt  int64             `json:"created_at"`
}

// RazorpayRefundRequest represents Razorpay refund request
type RazorpayRefundRequest struct {
	Amount  int64             `json:"amount"`
	Speed   string            `json:"speed"`
	Notes   map[string]string `json:"notes,omitempty"`
	Receipt string            `json:"receipt,omitempty"`
}

// RazorpayRefundResponse represents Razorpay refund response
type RazorpayRefundResponse struct {
	ID             string                 `json:"id"`
	Entity         string                 `json:"entity"`
	Amount         int64                  `json:"amount"`
	Currency       string                 `json:"currency"`
	PaymentID      string                 `json:"payment_id"`
	Notes          map[string]string      `json:"notes"`
	Receipt        string                 `json:"receipt"`
	AcquirerData   map[string]interface{} `json:"acquirer_data"`
	CreatedAt      int64                  `json:"created_at"`
	BatchID        string                 `json:"batch_id"`
	Status         string                 `json:"status"`
	SpeedRequested string                 `json:"speed_requested"`
	SpeedProcessed string                 `json:"speed_processed"`
}

// NewRazorpayGateway creates a new Razorpay gateway
func NewRazorpayGateway(gatewayConfig *config.PaymentGateway) (*RazorpayGateway, error) {
	gateway := &RazorpayGateway{
		config:     gatewayConfig,
		keyID:      gatewayConfig.Credentials["key_id"],
		keySecret:  gatewayConfig.Credentials["key_secret"],
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Set base URL based on test mode
	if testMode, ok := gatewayConfig.Settings["test_mode"]; ok && testMode == "true" {
		gateway.baseURL = "https://api.razorpay.com/v1"
	} else {
		gateway.baseURL = "https://api.razorpay.com/v1"
	}

	if gateway.keyID == "" || gateway.keySecret == "" {
		return nil, fmt.Errorf("razorpay key_id and key_secret are required")
	}

	return gateway, nil
}

// Initialize initializes the gateway with configuration
func (g *RazorpayGateway) Initialize(config map[string]string) error {
	// Razorpay is already initialized in NewRazorpayGateway
	return nil
}

// CreatePayment creates a new payment
func (g *RazorpayGateway) CreatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// First create an order
	orderReq := &RazorpayOrderRequest{
		Amount:         int64(req.Amount * 100), // Convert to paise
		Currency:       req.Currency,
		Receipt:        fmt.Sprintf("receipt_%d", time.Now().Unix()),
		PaymentCapture: true,
		Notes:          req.Metadata,
	}

	orderResp, err := g.createOrder(ctx, orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create payment response
	response := &payment.PaymentResponse{
		ID:        orderResp.ID,
		Status:    orderResp.Status,
		Amount:    float64(orderResp.Amount) / 100, // Convert from paise
		Currency:  orderResp.Currency,
		Method:    req.Method,
		Gateway:   "razorpay",
		GatewayID: orderResp.ID,
		CreatedAt: time.Unix(orderResp.CreatedAt, 0),
		UpdatedAt: time.Unix(orderResp.CreatedAt, 0),
	}

	// Add gateway response with order details
	response.GatewayResponse = map[string]interface{}{
		"order_id":     orderResp.ID,
		"amount":       orderResp.Amount,
		"currency":     orderResp.Currency,
		"receipt":      orderResp.Receipt,
		"status":       orderResp.Status,
		"key_id":       g.keyID,
		"callback_url": req.ReturnURL,
		"prefill": map[string]string{
			"name":    req.Customer.Name,
			"email":   req.Customer.Email,
			"contact": req.Customer.Phone,
		},
	}

	return response, nil
}

// GetPayment retrieves payment details
func (g *RazorpayGateway) GetPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	// Try to get as payment first
	paymentResp, err := g.getPayment(ctx, paymentID)
	if err == nil {
		response := &payment.PaymentResponse{
			ID:        paymentResp.ID,
			Status:    paymentResp.Status,
			Amount:    float64(paymentResp.Amount) / 100,
			Currency:  paymentResp.Currency,
			Method:    paymentResp.Method,
			Gateway:   "razorpay",
			GatewayID: paymentResp.ID,
			CreatedAt: time.Unix(paymentResp.CreatedAt, 0),
			UpdatedAt: time.Unix(paymentResp.CreatedAt, 0),
		}

		response.GatewayResponse = map[string]interface{}{
			"payment_id": paymentResp.ID,
			"status":     paymentResp.Status,
			"method":     paymentResp.Method,
			"amount":     paymentResp.Amount,
			"currency":   paymentResp.Currency,
		}

		return response, nil
	}

	// If not found as payment, try as order
	orderResp, err := g.getOrder(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment/order %s not found", paymentID)
	}

	response := &payment.PaymentResponse{
		ID:        orderResp.ID,
		Status:    orderResp.Status,
		Amount:    float64(orderResp.Amount) / 100,
		Currency:  orderResp.Currency,
		Gateway:   "razorpay",
		GatewayID: orderResp.ID,
		CreatedAt: time.Unix(orderResp.CreatedAt, 0),
		UpdatedAt: time.Unix(orderResp.CreatedAt, 0),
	}

	response.GatewayResponse = map[string]interface{}{
		"order_id": orderResp.ID,
		"status":   orderResp.Status,
		"amount":   orderResp.Amount,
		"currency": orderResp.Currency,
	}

	return response, nil
}

// UpdatePayment updates payment details
func (g *RazorpayGateway) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*payment.PaymentResponse, error) {
	// Razorpay doesn't support updating payments directly
	// Return current payment details
	return g.GetPayment(ctx, paymentID)
}

// CancelPayment cancels a payment
func (g *RazorpayGateway) CancelPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	// Razorpay doesn't support canceling payments directly
	// Return current payment details
	return g.GetPayment(ctx, paymentID)
}

// RefundPayment processes a refund
func (g *RazorpayGateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	refundReq := &RazorpayRefundRequest{
		Amount:  int64(req.Amount * 100), // Convert to paise
		Speed:   "normal",
		Notes:   req.Metadata,
		Receipt: fmt.Sprintf("refund_receipt_%d", time.Now().Unix()),
	}

	refundResp, err := g.createRefund(ctx, req.PaymentID, refundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	response := &payment.RefundResponse{
		ID:        refundResp.ID,
		PaymentID: refundResp.PaymentID,
		Amount:    float64(refundResp.Amount) / 100,
		Status:    refundResp.Status,
		GatewayID: refundResp.ID,
		CreatedAt: time.Unix(refundResp.CreatedAt, 0),
	}

	response.GatewayResponse = map[string]interface{}{
		"refund_id": refundResp.ID,
		"status":    refundResp.Status,
		"amount":    refundResp.Amount,
		"currency":  refundResp.Currency,
	}

	return response, nil
}

// CreateSubscription creates a subscription
func (g *RazorpayGateway) CreateSubscription(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// For subscriptions, we'll create a payment link or use Razorpay's subscription API
	// This is a simplified implementation
	return g.CreatePayment(ctx, req)
}

// GetSubscription retrieves subscription details
func (g *RazorpayGateway) GetSubscription(ctx context.Context, subscriptionID string) (*payment.PaymentResponse, error) {
	// For now, treat subscription as payment
	return g.GetPayment(ctx, subscriptionID)
}

// CancelSubscription cancels a subscription
func (g *RazorpayGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*payment.PaymentResponse, error) {
	// For now, treat subscription as payment
	return g.GetPayment(ctx, subscriptionID)
}

// ProcessWebhook processes webhook events
func (g *RazorpayGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	// Verify webhook signature
	if err := g.VerifyWebhook(payload, signature); err != nil {
		return nil, fmt.Errorf("webhook verification failed: %w", err)
	}

	// Parse webhook event
	var eventData map[string]interface{}
	if err := json.Unmarshal(payload, &eventData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}

	// Extract event type
	eventType, ok := eventData["event"].(string)
	if !ok {
		return nil, fmt.Errorf("event type not found in webhook data")
	}

	// Extract entity
	entity, ok := eventData["entity"].(string)
	if !ok {
		return nil, fmt.Errorf("entity not found in webhook data")
	}

	// Create webhook event
	webhookEvent := &payment.WebhookEvent{
		ID:        fmt.Sprintf("%s_%s", eventType, entity),
		Type:      eventType,
		Gateway:   "razorpay",
		Data:      eventData,
		CreatedAt: time.Now(),
	}

	return webhookEvent, nil
}

// VerifyWebhook verifies webhook signature
func (g *RazorpayGateway) VerifyWebhook(payload []byte, signature string) error {
	webhookSecret := g.config.Credentials["webhook_secret"]
	if webhookSecret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	if signature != expectedSignature {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// GetSupportedMethods returns supported payment methods
func (g *RazorpayGateway) GetSupportedMethods() []string {
	return []string{"card", "upi", "netbanking", "wallet", "emi"}
}

// GetGatewayName returns the gateway name
func (g *RazorpayGateway) GetGatewayName() string {
	return "razorpay"
}

// IsHealthy checks if the gateway is healthy
func (g *RazorpayGateway) IsHealthy(ctx context.Context) error {
	// Try to make a simple API call to check connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/orders", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.SetBasicAuth(g.keyID, g.keySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("razorpay API health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("razorpay API returned status %d", resp.StatusCode)
	}

	return nil
}

// createOrder creates a Razorpay order
func (g *RazorpayGateway) createOrder(ctx context.Context, req *RazorpayOrderRequest) (*RazorpayOrderResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.SetBasicAuth(g.keyID, g.keySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var orderResp RazorpayOrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

// getOrder retrieves a Razorpay order
func (g *RazorpayGateway) getOrder(ctx context.Context, orderID string) (*RazorpayOrderResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/orders/"+orderID, nil)
	if err != nil {
		return nil, err
	}

	httpReq.SetBasicAuth(g.keyID, g.keySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var orderResp RazorpayOrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

// getPayment retrieves a Razorpay payment
func (g *RazorpayGateway) getPayment(ctx context.Context, paymentID string) (*RazorpayPaymentResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, err
	}

	httpReq.SetBasicAuth(g.keyID, g.keySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var paymentResp RazorpayPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}

// createRefund creates a Razorpay refund
func (g *RazorpayGateway) createRefund(ctx context.Context, paymentID string, req *RazorpayRefundRequest) (*RazorpayRefundResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/payments/"+paymentID+"/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.SetBasicAuth(g.keyID, g.keySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API error: %s", string(body))
	}

	var refundResp RazorpayRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, err
	}

	return &refundResp, nil
}
