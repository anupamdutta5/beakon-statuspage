package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// PaymentGateway represents a payment gateway configuration
type PaymentGateway struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // stripe, razorpay, payu, paypal
	IsEnabled    bool              `json:"is_enabled"`
	IsDefault    bool              `json:"is_default"`
	Credentials  map[string]string `json:"credentials"`
	Settings     map[string]string `json:"settings"`
	SupportedMethods []string      `json:"supported_methods"`
	WebhookURL   string            `json:"webhook_url"`
}

// PaymentConfig represents the payment system configuration
type PaymentConfig struct {
	DefaultGateway string            `json:"default_gateway"`
	Gateways       []PaymentGateway  `json:"gateways"`
	Currency       string            `json:"currency"`
	RetrySettings  RetrySettings     `json:"retry_settings"`
	WebhookSettings WebhookSettings  `json:"webhook_settings"`
}

// RetrySettings for payment retry logic
type RetrySettings struct {
	MaxRetries     int     `json:"max_retries"`
	RetryInterval  int     `json:"retry_interval"` // in hours
	BackoffFactor  float64 `json:"backoff_factor"`
	MaxRetryDays   int     `json:"max_retry_days"`
}

// WebhookSettings for webhook configuration
type WebhookSettings struct {
	TimeoutSeconds int    `json:"timeout_seconds"`
	MaxRetries     int    `json:"max_retries"`
	RetryInterval  int    `json:"retry_interval"` // in seconds
	SecretKey      string `json:"secret_key"`
}

// PaymentMethod represents supported payment methods
type PaymentMethod struct {
	Type        string `json:"type"` // card, upi, netbanking, wallet, emi
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	IsEnabled   bool   `json:"is_enabled"`
	Gateway     string `json:"gateway"`
	Config      map[string]interface{} `json:"config"`
}

// LoadPaymentConfig loads payment configuration from environment and config files
func LoadPaymentConfig() (*PaymentConfig, error) {
	config := &PaymentConfig{
		DefaultGateway: "stripe",
		Currency:       "USD",
		RetrySettings: RetrySettings{
			MaxRetries:    5,
			RetryInterval: 24, // 24 hours
			BackoffFactor: 1.5,
			MaxRetryDays:  7,
		},
		WebhookSettings: WebhookSettings{
			TimeoutSeconds: 30,
			MaxRetries:     3,
			RetryInterval:  60, // 60 seconds
		},
	}

	// Load from environment variables
	if err := config.loadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load payment config from environment: %w", err)
	}

	// Load from config file if exists
	if err := config.loadFromFile(); err != nil {
		// Config file is optional, so we don't fail if it doesn't exist
		fmt.Printf("Warning: Could not load payment config file: %v\n", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid payment configuration: %w", err)
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables
func (pc *PaymentConfig) loadFromEnv() error {
	// Load default gateway
	if gateway := os.Getenv("PAYMENT_DEFAULT_GATEWAY"); gateway != "" {
		pc.DefaultGateway = gateway
	}

	// Load currency
	if currency := os.Getenv("PAYMENT_CURRENCY"); currency != "" {
		pc.Currency = currency
	}

	// Load Stripe configuration
	if os.Getenv("STRIPE_SECRET_KEY") != "" {
		stripeGateway := PaymentGateway{
			Name:      "stripe",
			Type:      "stripe",
			IsEnabled: true,
			Credentials: map[string]string{
				"secret_key":      os.Getenv("STRIPE_SECRET_KEY"),
				"publishable_key": os.Getenv("STRIPE_PUBLISHABLE_KEY"),
				"webhook_secret":  os.Getenv("STRIPE_WEBHOOK_SECRET"),
			},
			Settings: map[string]string{
				"api_version": "2023-10-16",
				"test_mode":   os.Getenv("STRIPE_TEST_MODE"),
			},
			SupportedMethods: []string{"card", "upi", "netbanking", "wallet"},
		}
		pc.Gateways = append(pc.Gateways, stripeGateway)
	}

	// Load Razorpay configuration
	if os.Getenv("RAZORPAY_KEY_ID") != "" {
		razorpayGateway := PaymentGateway{
			Name:      "razorpay",
			Type:      "razorpay",
			IsEnabled: true,
			Credentials: map[string]string{
				"key_id":     os.Getenv("RAZORPAY_KEY_ID"),
				"key_secret": os.Getenv("RAZORPAY_KEY_SECRET"),
				"webhook_secret": os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
			},
			Settings: map[string]string{
				"api_version": "v1",
				"test_mode":   os.Getenv("RAZORPAY_TEST_MODE"),
			},
			SupportedMethods: []string{"card", "upi", "netbanking", "wallet", "emi"},
		}
		pc.Gateways = append(pc.Gateways, razorpayGateway)
	}

	// Load PayU configuration
	if os.Getenv("PAYU_MERCHANT_KEY") != "" {
		payuGateway := PaymentGateway{
			Name:      "payu",
			Type:      "payu",
			IsEnabled: true,
			Credentials: map[string]string{
				"merchant_key":    os.Getenv("PAYU_MERCHANT_KEY"),
				"merchant_salt":   os.Getenv("PAYU_MERCHANT_SALT"),
				"webhook_secret":  os.Getenv("PAYU_WEBHOOK_SECRET"),
			},
			Settings: map[string]string{
				"api_version": "v1",
				"test_mode":   os.Getenv("PAYU_TEST_MODE"),
			},
			SupportedMethods: []string{"card", "upi", "netbanking", "wallet"},
		}
		pc.Gateways = append(pc.Gateways, payuGateway)
	}

	// Load PayPal configuration
	if os.Getenv("PAYPAL_CLIENT_ID") != "" {
		paypalGateway := PaymentGateway{
			Name:      "paypal",
			Type:      "paypal",
			IsEnabled: true,
			Credentials: map[string]string{
				"client_id":     os.Getenv("PAYPAL_CLIENT_ID"),
				"client_secret": os.Getenv("PAYPAL_CLIENT_SECRET"),
				"webhook_id":    os.Getenv("PAYPAL_WEBHOOK_ID"),
			},
			Settings: map[string]string{
				"api_version": "v2",
				"test_mode":   os.Getenv("PAYPAL_TEST_MODE"),
			},
			SupportedMethods: []string{"card", "paypal"},
		}
		pc.Gateways = append(pc.Gateways, paypalGateway)
	}

	return nil
}

// loadFromFile loads configuration from a JSON file
func (pc *PaymentConfig) loadFromFile() error {
	configPath := os.Getenv("PAYMENT_CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs/payment.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	fileConfig := &PaymentConfig{}
	if err := json.Unmarshal(data, fileConfig); err != nil {
		return err
	}

	// Merge file config with environment config
	pc.mergeConfig(fileConfig)

	return nil
}

// mergeConfig merges file configuration with environment configuration
func (pc *PaymentConfig) mergeConfig(fileConfig *PaymentConfig) {
	if fileConfig.DefaultGateway != "" {
		pc.DefaultGateway = fileConfig.DefaultGateway
	}
	if fileConfig.Currency != "" {
		pc.Currency = fileConfig.Currency
	}
	if fileConfig.RetrySettings.MaxRetries > 0 {
		pc.RetrySettings = fileConfig.RetrySettings
	}
	if fileConfig.WebhookSettings.TimeoutSeconds > 0 {
		pc.WebhookSettings = fileConfig.WebhookSettings
	}

	// Merge gateways (file config takes precedence)
	for _, fileGateway := range fileConfig.Gateways {
		found := false
		for i, envGateway := range pc.Gateways {
			if envGateway.Name == fileGateway.Name {
				pc.Gateways[i] = fileGateway
				found = true
				break
			}
		}
		if !found {
			pc.Gateways = append(pc.Gateways, fileGateway)
		}
	}
}

// Validate validates the payment configuration
func (pc *PaymentConfig) Validate() error {
	if pc.DefaultGateway == "" {
		return fmt.Errorf("default gateway not specified")
	}

	if len(pc.Gateways) == 0 {
		return fmt.Errorf("no payment gateways configured")
	}

	// Check if default gateway exists and is enabled
	defaultGatewayFound := false
	for _, gateway := range pc.Gateways {
		if gateway.Name == pc.DefaultGateway {
			if !gateway.IsEnabled {
				return fmt.Errorf("default gateway %s is disabled", pc.DefaultGateway)
			}
			defaultGatewayFound = true
			break
		}
	}

	if !defaultGatewayFound {
		return fmt.Errorf("default gateway %s not found in configured gateways", pc.DefaultGateway)
	}

	// Validate each gateway
	for _, gateway := range pc.Gateways {
		if err := gateway.Validate(); err != nil {
			return fmt.Errorf("invalid gateway %s: %w", gateway.Name, err)
		}
	}

	return nil
}

// Validate validates a payment gateway configuration
func (pg *PaymentGateway) Validate() error {
	if pg.Name == "" {
		return fmt.Errorf("gateway name is required")
	}

	if pg.Type == "" {
		return fmt.Errorf("gateway type is required")
	}

	// Validate required credentials based on gateway type
	switch pg.Type {
	case "stripe":
		if pg.Credentials["secret_key"] == "" {
			return fmt.Errorf("stripe secret_key is required")
		}
	case "razorpay":
		if pg.Credentials["key_id"] == "" || pg.Credentials["key_secret"] == "" {
			return fmt.Errorf("razorpay key_id and key_secret are required")
		}
	case "payu":
		if pg.Credentials["merchant_key"] == "" || pg.Credentials["merchant_salt"] == "" {
			return fmt.Errorf("payu merchant_key and merchant_salt are required")
		}
	case "paypal":
		if pg.Credentials["client_id"] == "" || pg.Credentials["client_secret"] == "" {
			return fmt.Errorf("paypal client_id and client_secret are required")
		}
	default:
		return fmt.Errorf("unsupported gateway type: %s", pg.Type)
	}

	return nil
}

// GetGateway returns a gateway by name
func (pc *PaymentConfig) GetGateway(name string) (*PaymentGateway, error) {
	for _, gateway := range pc.Gateways {
		if gateway.Name == name {
			return &gateway, nil
		}
	}
	return nil, fmt.Errorf("gateway %s not found", name)
}

// GetDefaultGateway returns the default gateway
func (pc *PaymentConfig) GetDefaultGateway() (*PaymentGateway, error) {
	return pc.GetGateway(pc.DefaultGateway)
}

// GetEnabledGateways returns all enabled gateways
func (pc *PaymentConfig) GetEnabledGateways() []PaymentGateway {
	var enabled []PaymentGateway
	for _, gateway := range pc.Gateways {
		if gateway.IsEnabled {
			enabled = append(enabled, gateway)
		}
	}
	return enabled
}

// GetSupportedMethods returns all supported payment methods across gateways
func (pc *PaymentConfig) GetSupportedMethods() []string {
	methodMap := make(map[string]bool)
	for _, gateway := range pc.Gateways {
		if gateway.IsEnabled {
			for _, method := range gateway.SupportedMethods {
				methodMap[method] = true
			}
		}
	}

	var methods []string
	for method := range methodMap {
		methods = append(methods, method)
	}
	return methods
}

// IsMethodSupported checks if a payment method is supported
func (pc *PaymentConfig) IsMethodSupported(method string) bool {
	for _, gateway := range pc.Gateways {
		if gateway.IsEnabled {
			for _, supportedMethod := range gateway.SupportedMethods {
				if strings.EqualFold(supportedMethod, method) {
					return true
				}
			}
		}
	}
	return false
}

// GetGatewayForMethod returns the best gateway for a payment method
func (pc *PaymentConfig) GetGatewayForMethod(method string) (*PaymentGateway, error) {
	// First try the default gateway
	defaultGateway, err := pc.GetDefaultGateway()
	if err == nil && pc.isMethodSupportedByGateway(defaultGateway, method) {
		return defaultGateway, nil
	}

	// Find any enabled gateway that supports this method
	for _, gateway := range pc.Gateways {
		if gateway.IsEnabled && pc.isMethodSupportedByGateway(&gateway, method) {
			return &gateway, nil
		}
	}

	return nil, fmt.Errorf("no gateway found that supports payment method: %s", method)
}

// isMethodSupportedByGateway checks if a gateway supports a specific method
func (pc *PaymentConfig) isMethodSupportedByGateway(gateway *PaymentGateway, method string) bool {
	for _, supportedMethod := range gateway.SupportedMethods {
		if strings.EqualFold(supportedMethod, method) {
			return true
		}
	}
	return false
}
