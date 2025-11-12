// Package providers provides SMS notification delivery via Twilio.
package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	resilience "github.com/anupamdutta5/shared-resilience"
	"go.uber.org/zap"
)

// TwilioSMSProvider implements SMS notifications using Twilio.
type TwilioSMSProvider struct {
	accountSID    string
	authToken     string
	fromNumber    string
	enabled       bool
	serviceClient *resilience.ServiceClient
	logger        *zap.Logger
}

// TwilioConfig represents Twilio configuration.
type TwilioConfig struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	FromNumber string `json:"from_number"`
	Enabled    bool   `json:"enabled"`
}

// NewTwilioSMSProvider creates a new Twilio SMS provider.
func NewTwilioSMSProvider(config map[string]interface{}, serviceClient *resilience.ServiceClient, logger *zap.Logger) (*TwilioSMSProvider, error) {
	var twilioConfig TwilioConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &twilioConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Twilio config: %w", err)
	}

	provider := &TwilioSMSProvider{
		accountSID:    twilioConfig.AccountSID,
		authToken:     twilioConfig.AuthToken,
		fromNumber:    twilioConfig.FromNumber,
		enabled:       twilioConfig.Enabled,
		serviceClient: serviceClient,
		logger:        logger,
	}

	if err := provider.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Twilio configuration: %w", err)
	}

	return provider, nil
}

// Send sends an SMS using Twilio.
func (p *TwilioSMSProvider) Send(ctx context.Context, request *NotificationRequest) (*NotificationResponse, error) {
	if !p.enabled {
		return &NotificationResponse{
			Success: false,
			Status:  "disabled",
			Error:   "Twilio SMS provider is disabled",
			SentAt:  time.Now(),
		}, nil
	}

	// Prepare request data
	data := url.Values{}
	data.Set("From", p.fromNumber)
	data.Set("To", request.Recipient)
	data.Set("Body", request.Content)

	// Build Twilio API URL
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", p.accountSID)

	// Prepare auth header (Basic Auth)
	auth := p.accountSID + ":" + p.authToken
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	// Create HTTP request for Twilio API
	req, err := http.NewRequestWithContext(ctx, "POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to create request: %v", err),
			SentAt:  time.Now(),
		}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	// Send via HTTP client (external API)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &NotificationResponse{
			Success: false,
			Status:  "failed",
			Error:   fmt.Sprintf("Failed to send SMS: %v", err),
			SentAt:  time.Now(),
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Warn("Failed to read Twilio response body", zap.Error(err))
		body = []byte{}
	}

	// Parse response
	var twilioResp map[string]interface{}
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		p.logger.Warn("Failed to parse Twilio response", zap.Error(err), zap.String("body", string(body)))
	}

	response := &NotificationResponse{
		Success:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		Status:       "sent",
		ResponseCode: resp.StatusCode,
		ResponseBody: string(body),
		SentAt:       time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	if twilioResp != nil {
		if sid, ok := twilioResp["sid"].(string); ok {
			response.MessageID = sid
		}
		if status, ok := twilioResp["status"].(string); ok {
			response.Status = status
		}
		response.Metadata["twilio_response"] = twilioResp
	}

	if resp.StatusCode >= 400 {
		response.Success = false
		response.Status = "failed"
		if twilioResp != nil {
			if errorCode, ok := twilioResp["code"].(float64); ok {
				response.Metadata["error_code"] = int(errorCode)
			}
			if errorMessage, ok := twilioResp["message"].(string); ok {
				response.Error = errorMessage
			}
		}
		if response.Error == "" {
			response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
	}

	p.logger.Info("SMS sent via Twilio",
		zap.String("recipient", request.Recipient),
		zap.String("message_id", response.MessageID),
		zap.String("status", response.Status),
		zap.Int("response_code", resp.StatusCode),
		zap.Bool("success", response.Success))

	return response, nil
}

// ValidateConfig validates the Twilio configuration.
func (p *TwilioSMSProvider) ValidateConfig(config map[string]interface{}) error {
	var twilioConfig TwilioConfig
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &twilioConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if twilioConfig.AccountSID == "" {
		return fmt.Errorf("account_sid is required")
	}
	if twilioConfig.AuthToken == "" {
		return fmt.Errorf("auth_token is required")
	}
	if twilioConfig.FromNumber == "" {
		return fmt.Errorf("from_number is required")
	}

	return nil
}

// GetType returns the provider type.
func (p *TwilioSMSProvider) GetType() string {
	return "sms"
}

// GetName returns the provider name.
func (p *TwilioSMSProvider) GetName() string {
	return "twilio"
}

// IsEnabled returns whether the provider is enabled.
func (p *TwilioSMSProvider) IsEnabled() bool {
	return p.enabled
}