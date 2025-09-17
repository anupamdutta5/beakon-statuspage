// Package providers provides notification provider management.
package providers

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Manager manages notification providers.
type Manager struct {
	providers map[string]Provider
	logger    *zap.Logger
	mu        sync.RWMutex
}

// NewManager creates a new provider manager.
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		providers: make(map[string]Provider),
		logger:    logger,
	}
}

// RegisterProvider registers a provider with the manager.
func (m *Manager) RegisterProvider(name string, provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider
	m.logger.Info("Provider registered",
		zap.String("name", name),
		zap.String("type", provider.GetType()),
		zap.Bool("enabled", provider.IsEnabled()))
}

// GetProvider returns a provider by name.
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return provider, nil
}

// GetProvidersByType returns all providers of a specific type.
func (m *Manager) GetProvidersByType(providerType string) []Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var providers []Provider
	for _, provider := range m.providers {
		if provider.GetType() == providerType && provider.IsEnabled() {
			providers = append(providers, provider)
		}
	}

	return providers
}

// ListProviders returns all registered providers.
func (m *Manager) ListProviders() map[string]Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	providers := make(map[string]Provider)
	for name, provider := range m.providers {
		providers[name] = provider
	}

	return providers
}

// Send sends a notification using the specified provider.
func (m *Manager) Send(ctx context.Context, providerName string, request *NotificationRequest) (*NotificationResponse, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	if !provider.IsEnabled() {
		return &NotificationResponse{
			Success: false,
			Status:  "disabled",
			Error:   fmt.Sprintf("Provider '%s' is disabled", providerName),
		}, nil
	}

	return provider.Send(ctx, request)
}

// SendToMultiple sends a notification to multiple providers.
func (m *Manager) SendToMultiple(ctx context.Context, providerNames []string, request *NotificationRequest) map[string]*NotificationResponse {
	responses := make(map[string]*NotificationResponse)
	var wg sync.WaitGroup

	for _, providerName := range providerNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			response, err := m.Send(ctx, name, request)
			if err != nil {
				response = &NotificationResponse{
					Success: false,
					Status:  "error",
					Error:   err.Error(),
				}
			}

			// Thread-safe assignment
			m.mu.Lock()
			responses[name] = response
			m.mu.Unlock()
		}(providerName)
	}

	wg.Wait()
	return responses
}

// SendByType sends a notification to all providers of a specific type.
func (m *Manager) SendByType(ctx context.Context, providerType string, request *NotificationRequest) map[string]*NotificationResponse {
	providers := m.GetProvidersByType(providerType)
	responses := make(map[string]*NotificationResponse)

	if len(providers) == 0 {
		m.logger.Warn("No enabled providers found for type", zap.String("type", providerType))
		return responses
	}

	var wg sync.WaitGroup
	for _, provider := range providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()

			response, err := p.Send(ctx, request)
			if err != nil {
				response = &NotificationResponse{
					Success: false,
					Status:  "error",
					Error:   err.Error(),
				}
			}

			// Thread-safe assignment
			m.mu.Lock()
			responses[p.GetName()] = response
			m.mu.Unlock()
		}(provider)
	}

	wg.Wait()
	return responses
}

// CreateProvider creates a provider instance from configuration.
func (m *Manager) CreateProvider(config *ProviderConfig) (Provider, error) {
	switch config.Type {
	case "sms":
		if config.Name == "twilio" {
			return NewTwilioSMSProvider(config.Settings, m.logger)
		}
		return nil, fmt.Errorf("unsupported SMS provider: %s", config.Name)

	case "slack":
		return NewSlackProvider(config.Settings, m.logger)

	case "teams":
		return NewTeamsProvider(config.Settings, m.logger)

	case "webhook":
		return NewWebhookProvider(config.Settings, m.logger)

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// LoadProvidersFromConfig loads providers from configuration.
func (m *Manager) LoadProvidersFromConfig(configs []ProviderConfig) error {
	for _, config := range configs {
		if !config.Enabled {
			m.logger.Info("Skipping disabled provider",
				zap.String("name", config.Name),
				zap.String("type", config.Type))
			continue
		}

		provider, err := m.CreateProvider(&config)
		if err != nil {
			m.logger.Error("Failed to create provider",
				zap.String("name", config.Name),
				zap.String("type", config.Type),
				zap.Error(err))
			continue
		}

		m.RegisterProvider(config.Name, provider)
	}

	return nil
}

// ValidateProviderConfig validates a provider configuration.
func (m *Manager) ValidateProviderConfig(config *ProviderConfig) error {
	provider, err := m.CreateProvider(config)
	if err != nil {
		return err
	}

	return provider.ValidateConfig(config.Settings)
}

// GetProviderStats returns statistics about registered providers.
func (m *Manager) GetProviderStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_providers": len(m.providers),
		"enabled_providers": 0,
		"providers_by_type": make(map[string]int),
		"provider_details": make([]map[string]interface{}, 0),
	}

	enabledCount := 0
	typeCount := make(map[string]int)

	for name, provider := range m.providers {
		if provider.IsEnabled() {
			enabledCount++
		}

		providerType := provider.GetType()
		typeCount[providerType]++

		details := map[string]interface{}{
			"name":    name,
			"type":    providerType,
			"enabled": provider.IsEnabled(),
		}

		stats["provider_details"] = append(stats["provider_details"].([]map[string]interface{}), details)
	}

	stats["enabled_providers"] = enabledCount
	stats["providers_by_type"] = typeCount

	return stats
}

// TestProvider tests a provider configuration by sending a test notification.
func (m *Manager) TestProvider(ctx context.Context, config *ProviderConfig, testRecipient string) (*NotificationResponse, error) {
	provider, err := m.CreateProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	testRequest := &NotificationRequest{
		Recipient: testRecipient,
		Subject:   "Test Notification from StatusPage",
		Content:   "This is a test notification to verify your notification provider configuration.",
		Type:      config.Type,
		Priority:  "low",
		Metadata: map[string]interface{}{
			"test": true,
			"provider": config.Name,
		},
	}

	return provider.Send(ctx, testRequest)
}

// UpdateProviderConfig updates a provider's configuration.
func (m *Manager) UpdateProviderConfig(name string, config *ProviderConfig) error {
	// Validate the new configuration
	if err := m.ValidateProviderConfig(config); err != nil {
		return fmt.Errorf("invalid provider configuration: %w", err)
	}

	// Create new provider instance
	provider, err := m.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Replace the existing provider
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider
	m.logger.Info("Provider configuration updated",
		zap.String("name", name),
		zap.String("type", config.Type),
		zap.Bool("enabled", config.Enabled))

	return nil
}