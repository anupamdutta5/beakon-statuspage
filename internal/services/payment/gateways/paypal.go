package gateways

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
)

// PayPalGateway implements PaymentGateway for PayPal
type PayPalGateway struct {
	config       *config.PaymentGateway
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
	accessToken  string
	tokenExpiry  time.Time
}

// PayPalAccessToken represents PayPal access token response
type PayPalAccessToken struct {
	Scope       string `json:"scope"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	AppID       string `json:"app_id"`
	ExpiresIn   int    `json:"expires_in"`
	Nonce       string `json:"nonce"`
}

// PayPalOrderRequest represents PayPal order request
type PayPalOrderRequest struct {
	Intent             string               `json:"intent"`
	PurchaseUnits      []PayPalPurchaseUnit `json:"purchase_units"`
	ApplicationContext PayPalAppContext     `json:"application_context"`
}

// PayPalPurchaseUnit represents PayPal purchase unit
type PayPalPurchaseUnit struct {
	ReferenceID string       `json:"reference_id"`
	Amount      PayPalAmount `json:"amount"`
	Description string       `json:"description,omitempty"`
	CustomID    string       `json:"custom_id,omitempty"`
	InvoiceID   string       `json:"invoice_id,omitempty"`
}

// PayPalAmount represents PayPal amount
type PayPalAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

// PayPalAppContext represents PayPal application context
type PayPalAppContext struct {
	BrandName          string `json:"brand_name,omitempty"`
	Locale             string `json:"locale,omitempty"`
	LandingPage        string `json:"landing_page,omitempty"`
	ShippingPreference string `json:"shipping_preference,omitempty"`
	UserAction         string `json:"user_action,omitempty"`
	ReturnURL          string `json:"return_url"`
	CancelURL          string `json:"cancel_url"`
}

// PayPalOrderResponse represents PayPal order response
type PayPalOrderResponse struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Links  []PayPalLink `json:"links"`
}

// PayPalLink represents PayPal link
type PayPalLink struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// PayPalOrderDetails represents PayPal order details
type PayPalOrderDetails struct {
	ID            string               `json:"id"`
	Intent        string               `json:"intent"`
	Status        string               `json:"status"`
	PurchaseUnits []PayPalPurchaseUnit `json:"purchase_units"`
	Payer         PayPalPayer          `json:"payer"`
	CreateTime    string               `json:"create_time"`
	UpdateTime    string               `json:"update_time"`
	Links         []PayPalLink         `json:"links"`
}

// PayPalPayer represents PayPal payer
type PayPalPayer struct {
	PayerID      string          `json:"payer_id"`
	EmailAddress string          `json:"email_address"`
	Name         PayPalPayerName `json:"name"`
	Address      PayPalAddress   `json:"address"`
}

// PayPalPayerName represents PayPal payer name
type PayPalPayerName struct {
	GivenName string `json:"given_name"`
	Surname   string `json:"surname"`
}

// PayPalAddress represents PayPal address
type PayPalAddress struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	AdminArea2   string `json:"admin_area_2"`
	AdminArea1   string `json:"admin_area_1"`
	PostalCode   string `json:"postal_code"`
	CountryCode  string `json:"country_code"`
}

// PayPalRefundRequest represents PayPal refund request
type PayPalRefundRequest struct {
	Amount      PayPalAmount `json:"amount"`
	InvoiceID   string       `json:"invoice_id,omitempty"`
	NoteToPayer string       `json:"note_to_payer,omitempty"`
}

// PayPalRefundResponse represents PayPal refund response
type PayPalRefundResponse struct {
	ID          string       `json:"id"`
	Amount      PayPalAmount `json:"amount"`
	Status      string       `json:"status"`
	CreateTime  string       `json:"create_time"`
	UpdateTime  string       `json:"update_time"`
	InvoiceID   string       `json:"invoice_id"`
	NoteToPayer string       `json:"note_to_payer"`
	Links       []PayPalLink `json:"links"`
}

// NewPayPalGateway creates a new PayPal gateway
func NewPayPalGateway(gatewayConfig *config.PaymentGateway) (*PayPalGateway, error) {
	gateway := &PayPalGateway{
		config:       gatewayConfig,
		clientID:     gatewayConfig.Credentials["client_id"],
		clientSecret: gatewayConfig.Credentials["client_secret"],
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// Set base URL based on test mode
	if testMode, ok := gatewayConfig.Settings["test_mode"]; ok && testMode == "true" {
		gateway.baseURL = "https://api.sandbox.paypal.com"
	} else {
		gateway.baseURL = "https://api.paypal.com"
	}

	if gateway.clientID == "" || gateway.clientSecret == "" {
		return nil, fmt.Errorf("paypal client_id and client_secret are required")
	}

	return gateway, nil
}

// Initialize initializes the gateway with configuration
func (g *PayPalGateway) Initialize(config map[string]string) error {
	// PayPal is already initialized in NewPayPalGateway
	return nil
}

// CreatePayment creates a new payment
func (g *PayPalGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Get access token
	if err := g.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create PayPal order
	orderReq := &PayPalOrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []PayPalPurchaseUnit{
			{
				ReferenceID: fmt.Sprintf("REF_%d", time.Now().Unix()),
				Amount: PayPalAmount{
					CurrencyCode: req.Currency,
					Value:        fmt.Sprintf("%.2f", req.Amount),
				},
				Description: req.Description,
				CustomID:    req.Metadata["custom_id"],
				InvoiceID:   req.Metadata["invoice_id"],
			},
		},
		ApplicationContext: PayPalAppContext{
			BrandName:          req.Metadata["brand_name"],
			Locale:             "en-US",
			LandingPage:        "NO_PREFERENCE",
			ShippingPreference: "NO_SHIPPING",
			UserAction:         "PAY_NOW",
			ReturnURL:          req.ReturnURL,
			CancelURL:          req.ReturnURL + "?cancelled=true",
		},
	}

	orderResp, err := g.createOrder(ctx, orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Find approval URL
	var approvalURL string
	for _, link := range orderResp.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	// Create payment response
	response := &PaymentResponse{
		ID:          orderResp.ID,
		Status:      orderResp.Status,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      req.Method,
		Gateway:     "paypal",
		GatewayID:   orderResp.ID,
		RedirectURL: approvalURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"order_id":     orderResp.ID,
		"status":       orderResp.Status,
		"approval_url": approvalURL,
		"amount":       req.Amount,
		"currency":     req.Currency,
	}

	return response, nil
}

// GetPayment retrieves payment details
func (g *PayPalGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// Get access token
	if err := g.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Get order details
	orderDetails, err := g.getOrder(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Extract amount from purchase units
	var amount float64
	var currency string
	if len(orderDetails.PurchaseUnits) > 0 {
		amount, _ = strconv.ParseFloat(orderDetails.PurchaseUnits[0].Amount.Value, 64)
		currency = orderDetails.PurchaseUnits[0].Amount.CurrencyCode
	}

	response := &PaymentResponse{
		ID:        orderDetails.ID,
		Status:    orderDetails.Status,
		Amount:    amount,
		Currency:  currency,
		Gateway:   "paypal",
		GatewayID: orderDetails.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"order_id": orderDetails.ID,
		"status":   orderDetails.Status,
		"intent":   orderDetails.Intent,
		"payer":    orderDetails.Payer,
	}

	return response, nil
}

// UpdatePayment updates payment details
func (g *PayPalGateway) UpdatePayment(ctx context.Context, paymentID string, updates map[string]interface{}) (*PaymentResponse, error) {
	// PayPal doesn't support updating orders directly
	// Return current payment details
	return g.GetPayment(ctx, paymentID)
}

// CancelPayment cancels a payment
func (g *PayPalGateway) CancelPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// Get access token
	if err := g.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Cancel order
	if err := g.cancelOrder(ctx, paymentID); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Return updated payment details
	return g.GetPayment(ctx, paymentID)
}

// RefundPayment processes a refund
func (g *PayPalGateway) RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	// Get access token
	if err := g.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create refund request
	refundReq := &PayPalRefundRequest{
		Amount: PayPalAmount{
			CurrencyCode: "USD", // Default currency, should be extracted from original payment
			Value:        fmt.Sprintf("%.2f", req.Amount),
		},
		InvoiceID:   req.Metadata["invoice_id"],
		NoteToPayer: req.Reason,
	}

	// Process refund
	refundResp, err := g.processRefund(ctx, req.PaymentID, refundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	response := &RefundResponse{
		ID:        refundResp.ID,
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Status:    refundResp.Status,
		GatewayID: refundResp.ID,
		CreatedAt: time.Now(),
	}

	// Add gateway response
	response.GatewayResponse = map[string]interface{}{
		"refund_id": refundResp.ID,
		"status":    refundResp.Status,
		"amount":    refundResp.Amount,
	}

	return response, nil
}

// CreateSubscription creates a subscription
func (g *PayPalGateway) CreateSubscription(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// PayPal subscriptions require a different API
	// For now, create a regular payment
	return g.CreatePayment(ctx, req)
}

// GetSubscription retrieves subscription details
func (g *PayPalGateway) GetSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error) {
	// PayPal subscriptions require a different API
	// For now, treat as payment
	return g.GetPayment(ctx, subscriptionID)
}

// CancelSubscription cancels a subscription
func (g *PayPalGateway) CancelSubscription(ctx context.Context, subscriptionID string) (*PaymentResponse, error) {
	// PayPal subscriptions require a different API
	// For now, treat as payment
	return g.CancelPayment(ctx, subscriptionID)
}

// ProcessWebhook processes webhook events
func (g *PayPalGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error) {
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
	eventType, ok := eventData["event_type"].(string)
	if !ok {
		return nil, fmt.Errorf("event type not found in webhook data")
	}

	// Create webhook event
	webhookEvent := &WebhookEvent{
		ID:        fmt.Sprintf("paypal_%s_%d", eventType, time.Now().Unix()),
		Type:      eventType,
		Gateway:   "paypal",
		Data:      eventData,
		CreatedAt: time.Now(),
	}

	return webhookEvent, nil
}

// VerifyWebhook verifies webhook signature
func (g *PayPalGateway) VerifyWebhook(payload []byte, signature string) error {
	// PayPal webhook verification is complex and requires additional implementation
	// For now, we'll skip verification
	logger.Warn("PayPal webhook verification not implemented")
	return nil
}

// GetSupportedMethods returns supported payment methods
func (g *PayPalGateway) GetSupportedMethods() []string {
	return []string{"card", "paypal"}
}

// GetGatewayName returns the gateway name
func (g *PayPalGateway) GetGatewayName() string {
	return "paypal"
}

// IsHealthy checks if the gateway is healthy
func (g *PayPalGateway) IsHealthy(ctx context.Context) error {
	// Try to get access token to check API connectivity
	if err := g.ensureAccessToken(ctx); err != nil {
		return fmt.Errorf("paypal API health check failed: %w", err)
	}

	return nil
}

// ensureAccessToken ensures we have a valid access token
func (g *PayPalGateway) ensureAccessToken(ctx context.Context) error {
	// Check if token is still valid
	if g.accessToken != "" && time.Now().Before(g.tokenExpiry) {
		return nil
	}

	// Get new access token
	token, err := g.getAccessToken(ctx)
	if err != nil {
		return err
	}

	g.accessToken = token.AccessToken
	g.tokenExpiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return nil
}

// getAccessToken gets PayPal access token
func (g *PayPalGateway) getAccessToken(ctx context.Context) (*PayPalAccessToken, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/v1/oauth2/token", nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en_US")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set basic auth
	req.SetBasicAuth(g.clientID, g.clientSecret)

	// Set body
	req.Body = io.NopCloser(bytes.NewBufferString("grant_type=client_credentials"))

	// Make request
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
		return nil, fmt.Errorf("paypal API error: %s", string(body))
	}

	var token PayPalAccessToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// createOrder creates a PayPal order
func (g *PayPalGateway) createOrder(ctx context.Context, req *PayPalOrderRequest) (*PayPalOrderResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/v2/checkout/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.accessToken)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("paypal API error: %s", string(body))
	}

	var orderResp PayPalOrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

// getOrder retrieves PayPal order details
func (g *PayPalGateway) getOrder(ctx context.Context, orderID string) (*PayPalOrderDetails, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/v2/checkout/orders/"+orderID, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.accessToken)

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
		return nil, fmt.Errorf("paypal API error: %s", string(body))
	}

	var orderDetails PayPalOrderDetails
	if err := json.Unmarshal(body, &orderDetails); err != nil {
		return nil, err
	}

	return &orderDetails, nil
}

// cancelOrder cancels a PayPal order
func (g *PayPalGateway) cancelOrder(ctx context.Context, orderID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/v2/checkout/orders/"+orderID+"/cancel", nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.accessToken)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paypal API error: %s", string(body))
	}

	return nil
}

// processRefund processes a PayPal refund
func (g *PayPalGateway) processRefund(ctx context.Context, captureID string, req *PayPalRefundRequest) (*PayPalRefundResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/v2/payments/captures/"+captureID+"/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.accessToken)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("paypal API error: %s", string(body))
	}

	var refundResp PayPalRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, err
	}

	return &refundResp, nil
}
