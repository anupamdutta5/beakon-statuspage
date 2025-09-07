package gateways

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/services/payment"
)

// PayUGateway implements PaymentGateway for PayU
type PayUGateway struct {
	config       *config.PaymentGateway
	merchantKey  string
	merchantSalt string
	baseURL      string
	httpClient   *http.Client
}

// PayUPaymentRequest represents PayU payment request
type PayUPaymentRequest struct {
	Key             string `json:"key"`
	TxID            string `json:"txnid"`
	Amount          string `json:"amount"`
	ProductInfo     string `json:"productinfo"`
	FirstName       string `json:"firstname"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Hash            string `json:"hash"`
	ServiceProvider string `json:"service_provider"`
	SuccessURL      string `json:"surl"`
	FailureURL      string `json:"furl"`
	CancelURL       string `json:"curl"`
}

// PayUPaymentResponse represents PayU payment response
type PayUPaymentResponse struct {
	Status        string `json:"status"`
	Message       string `json:"msg"`
	TransactionID string `json:"txnid"`
	Amount        string `json:"amount"`
	ProductInfo   string `json:"productinfo"`
	FirstName     string `json:"firstname"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Hash          string `json:"hash"`
	PaymentID     string `json:"mihpayid"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	BankRefNum    string `json:"bank_ref_num"`
	BankCode      string `json:"bankcode"`
	Error         string `json:"error"`
	ErrorCode     string `json:"error_code"`
}

// PayUVerifyRequest represents PayU verification request
type PayUVerifyRequest struct {
	Key     string `json:"key"`
	Command string `json:"command"`
	Hash    string `json:"hash"`
	Var1    string `json:"var1"`
}

// PayUVerifyResponse represents PayU verification response
type PayUVerifyResponse struct {
	Status        string `json:"status"`
	Message       string `json:"msg"`
	TransactionID string `json:"txnid"`
	Amount        string `json:"amount"`
	ProductInfo   string `json:"productinfo"`
	FirstName     string `json:"firstname"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Hash          string `json:"hash"`
	PaymentID     string `json:"mihpayid"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	BankRefNum    string `json:"bank_ref_num"`
	BankCode      string `json:"bankcode"`
	Error         string `json:"error"`
	ErrorCode     string `json:"error_code"`
}

// NewPayUGateway creates a new PayU gateway
func NewPayUGateway(gatewayConfig *config.PaymentGateway) (*PayUGateway, error) {
	gateway := &PayUGateway{
		config:       gatewayConfig,
		merchantKey:  gatewayConfig.Credentials["merchant_key"],
		merchantSalt: gatewayConfig.Credentials["merchant_salt"],
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// Set base URL based on test mode
	if testMode, ok := gatewayConfig.Settings["test_mode"]; ok && testMode == "true" {
		gateway.baseURL = "https://test.payu.in"
	} else {
		gateway.baseURL = "https://secure.payu.in"
	}

	if gateway.merchantKey == "" || gateway.merchantSalt == "" {
		return nil, fmt.Errorf("payu merchant_key and merchant_salt are required")
	}

	return gateway, nil
}

// Initialize initializes the gateway with configuration
func (g *PayUGateway) Initialize(config map[string]string) error {
	// PayU is already initialized in NewPayUGateway
	return nil
}

// CreatePayment creates a new payment
func (g *PayUGateway) CreatePayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// Generate transaction ID
	txID := fmt.Sprintf("TXN_%d", time.Now().Unix())

	// Create payment request
	payuReq := &PayUPaymentRequest{
		Key:             g.merchantKey,
		TxID:            txID,
		Amount:          fmt.Sprintf("%.2f", req.Amount),
		ProductInfo:     req.Description,
		FirstName:       req.Customer.Name,
		Email:           req.Customer.Email,
		Phone:           req.Customer.Phone,
		ServiceProvider: "payu_paisa",
		SuccessURL:      req.ReturnURL,
		FailureURL:      req.ReturnURL + "?status=failed",
		CancelURL:       req.ReturnURL + "?status=cancelled",
	}

	// Generate hash
	hash := g.generateHash(payuReq)
	payuReq.Hash = hash

	// Create payment response
	response := &payment.PaymentResponse{
		ID:        txID,
		Status:    "pending",
		Amount:    req.Amount,
		Currency:  req.Currency,
		Method:    req.Method,
		Gateway:   "payu",
		GatewayID: txID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add gateway response with payment form data
	response.GatewayResponse = map[string]interface{}{
		"key":              payuReq.Key,
		"txnid":            payuReq.TxID,
		"amount":           payuReq.Amount,
		"productinfo":      payuReq.ProductInfo,
		"firstname":        payuReq.FirstName,
		"email":            payuReq.Email,
		"phone":            payuReq.Phone,
		"hash":             payuReq.Hash,
		"service_provider": payuReq.ServiceProvider,
		"surl":             payuReq.SuccessURL,
		"furl":             payuReq.FailureURL,
		"curl":             payuReq.CancelURL,
		"action_url":       g.baseURL + "/_payment",
	}

	return response, nil
}

// GetPayment retrieves payment details
func (g *PayUGateway) GetPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	// PayU doesn't provide a direct API to get payment details
	// We'll return a basic response with the payment ID
	response := &payment.PaymentResponse{
		ID:        paymentID,
		Status:    "unknown",
		Gateway:   "payu",
		GatewayID: paymentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	response.GatewayResponse = map[string]interface{}{
		"txnid": paymentID,
		"note":  "PayU doesn't provide direct payment status API",
	}

	return response, nil
}

// UpdatePayment updates payment details
func (g *PayUGateway) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*payment.PaymentResponse, error) {
	// PayU doesn't support updating payments
	return g.GetPayment(ctx, paymentID)
}

// CancelPayment cancels a payment
func (g *PayUGateway) CancelPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	// PayU doesn't support canceling payments directly
	return g.GetPayment(ctx, paymentID)
}

// RefundPayment processes a refund
func (g *PayUGateway) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// PayU refund implementation would go here
	// For now, return a basic response
	response := &payment.RefundResponse{
		ID:        fmt.Sprintf("REF_%d", time.Now().Unix()),
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Status:    "pending",
		GatewayID: fmt.Sprintf("REF_%d", time.Now().Unix()),
		CreatedAt: time.Now(),
	}

	response.GatewayResponse = map[string]interface{}{
		"note": "PayU refund implementation pending",
	}

	return response, nil
}

// CreateSubscription creates a subscription
func (g *PayUGateway) CreateSubscription(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// PayU doesn't have native subscription support
	// We'll create a regular payment
	return g.CreatePayment(ctx, req)
}

// GetSubscription retrieves subscription details
func (g *PayUGateway) GetSubscription(ctx context.Context, subscriptionID string) (*payment.PaymentResponse, error) {
	// PayU doesn't have native subscription support
	return g.GetPayment(ctx, subscriptionID)
}

// CancelSubscription cancels a subscription
func (g *PayUGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*payment.PaymentResponse, error) {
	// PayU doesn't have native subscription support
	return g.GetPayment(ctx, subscriptionID)
}

// ProcessWebhook processes webhook events
func (g *PayUGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	// Parse webhook data
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %w", err)
	}

	// Verify webhook signature
	if err := g.VerifyWebhook(payload, signature); err != nil {
		return nil, fmt.Errorf("webhook verification failed: %w", err)
	}

	// Extract event type based on status
	status, ok := webhookData["status"].(string)
	if !ok {
		status = "unknown"
	}

	eventType := "payment." + status

	// Create webhook event
	webhookEvent := &payment.WebhookEvent{
		ID:        fmt.Sprintf("payu_%s_%d", status, time.Now().Unix()),
		Type:      eventType,
		Gateway:   "payu",
		Data:      webhookData,
		CreatedAt: time.Now(),
	}

	return webhookEvent, nil
}

// VerifyWebhook verifies webhook signature
func (g *PayUGateway) VerifyWebhook(payload []byte, signature string) error {
	// Parse the payload to extract hash
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return fmt.Errorf("failed to parse webhook data: %w", err)
	}

	// Extract hash from webhook data
	receivedHash, ok := webhookData["hash"].(string)
	if !ok {
		return fmt.Errorf("hash not found in webhook data")
	}

	// Generate expected hash
	expectedHash := g.generateWebhookHash(webhookData)

	// Compare hashes
	if receivedHash != expectedHash {
		return fmt.Errorf("invalid webhook hash")
	}

	return nil
}

// GetSupportedMethods returns supported payment methods
func (g *PayUGateway) GetSupportedMethods() []string {
	return []string{"card", "upi", "netbanking", "wallet"}
}

// GetGatewayName returns the gateway name
func (g *PayUGateway) GetGatewayName() string {
	return "payu"
}

// IsHealthy checks if the gateway is healthy
func (g *PayUGateway) IsHealthy(ctx context.Context) error {
	// PayU doesn't provide a health check endpoint
	// We'll just verify that we can make a request to their base URL
	req, err := http.NewRequestWithContext(ctx, "GET", g.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("payu health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("payu returned status %d", resp.StatusCode)
	}

	return nil
}

// generateHash generates PayU hash for payment request
func (g *PayUGateway) generateHash(req *PayUPaymentRequest) string {
	// Create hash string
	hashString := fmt.Sprintf("%s|%s|%s|%s|%s|%s|||||||||||%s",
		req.Key,
		req.TxID,
		req.Amount,
		req.ProductInfo,
		req.FirstName,
		req.Email,
		g.merchantSalt,
	)

	// Generate SHA512 hash
	hash := sha512.Sum512([]byte(hashString))
	return hex.EncodeToString(hash[:])
}

// generateWebhookHash generates PayU hash for webhook verification
func (g *PayUGateway) generateWebhookHash(data map[string]interface{}) string {
	// Extract required fields for hash generation
	key := data["key"].(string)
	txnid := data["txnid"].(string)
	amount := data["amount"].(string)
	productinfo := data["productinfo"].(string)
	firstname := data["firstname"].(string)
	email := data["email"].(string)
	status := data["status"].(string)

	// Create hash string
	hashString := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|||||||||||%s",
		key,
		txnid,
		amount,
		productinfo,
		firstname,
		email,
		status,
		g.merchantSalt,
	)

	// Generate SHA512 hash
	hash := sha512.Sum512([]byte(hashString))
	return hex.EncodeToString(hash[:])
}

// generateVerifyHash generates PayU hash for verification request
func (g *PayUGateway) generateVerifyHash(key, command, var1 string) string {
	// Create hash string
	hashString := fmt.Sprintf("%s|%s|%s|%s",
		key,
		command,
		var1,
		g.merchantSalt,
	)

	// Generate SHA512 hash
	hash := sha512.Sum512([]byte(hashString))
	return hex.EncodeToString(hash[:])
}

// verifyPayment verifies a PayU payment
func (g *PayUGateway) verifyPayment(ctx context.Context, paymentID string) (*PayUVerifyResponse, error) {
	// Create verification request
	verifyReq := &PayUVerifyRequest{
		Key:     g.merchantKey,
		Command: "verify_payment",
		Var1:    paymentID,
	}

	// Generate hash
	verifyReq.Hash = g.generateVerifyHash(verifyReq.Key, verifyReq.Command, verifyReq.Var1)

	// Create form data
	formData := url.Values{}
	formData.Set("key", verifyReq.Key)
	formData.Set("command", verifyReq.Command)
	formData.Set("hash", verifyReq.Hash)
	formData.Set("var1", verifyReq.Var1)

	// Make request
	req, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/merchant/postservice.php", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payu verification API error: %s", string(body))
	}

	var verifyResp PayUVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, err
	}

	return &verifyResp, nil
}
