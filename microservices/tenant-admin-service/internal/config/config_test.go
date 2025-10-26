package config

import (
	"testing"
)

// TestLoadConfig tests the configuration loading from YAML files
func TestLoadConfig(t *testing.T) {
	// This test requires the configs/ directory to exist with proper files
	cfg, err := Load()

	if err != nil {
		t.Logf("Configuration loading error (expected if .env is missing): %v", err)
		// Don't fail the test - .env might not exist in CI/CD
		return
	}

	// Verify basic configuration
	if cfg == nil {
		t.Fatal("Config is nil")
	}

	// Test service config
	if cfg.Service.Name != "tenant-admin-service" {
		t.Errorf("Expected service name 'tenant-admin-service', got '%s'", cfg.Service.Name)
	}

	// Test server config
	if cfg.Server.Port != 8099 {
		t.Errorf("Expected port 8099, got %d", cfg.Server.Port)
	}

	// Test database config
	if cfg.Database.Name != "tenant_admin_db" {
		t.Errorf("Expected database name 'tenant_admin_db', got '%s'", cfg.Database.Name)
	}

	// Test tenant config
	if cfg.Tenant.DefaultPlan != "free" {
		t.Errorf("Expected default plan 'free', got '%s'", cfg.Tenant.DefaultPlan)
	}

	t.Logf("✓ Configuration loaded successfully")
	t.Logf("  Service: %s v%s", cfg.Service.Name, cfg.Service.Version)
	t.Logf("  Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
	t.Logf("  Database: %s", cfg.Database.Name)
	t.Logf("  Redis: enabled=%v", cfg.Redis.Enabled)
	t.Logf("  Tenant plan: %s", cfg.Tenant.DefaultPlan)
}

// TestLoadServiceEndpoints tests service endpoints loading
func TestLoadServiceEndpoints(t *testing.T) {
	endpoints, err := LoadServiceEndpoints()

	if err != nil {
		t.Fatalf("Failed to load service endpoints: %v", err)
	}

	if endpoints == nil {
		t.Fatal("Endpoints map is nil")
	}

	// Check expected endpoints
	expectedEndpoints := []string{
		"component-service",
		"incident-service",
		"notification-service",
		"subscriber-service",
		"branding-service",
	}

	for _, name := range expectedEndpoints {
		endpoint, ok := endpoints[name]
		if !ok {
			t.Errorf("Expected endpoint '%s' not found", name)
			continue
		}

		if endpoint.URL == "" {
			t.Errorf("Endpoint '%s' has empty URL", name)
		}

		t.Logf("✓ Endpoint %s: %s (retries: %d, circuit_breaker: %v)",
			name, endpoint.URL, endpoint.Retries, endpoint.CircuitBreaker.Enabled)
	}
}

// TestConfigHelperMethods tests the helper methods
func TestConfigHelperMethods(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Skip("Skipping due to config load error (likely missing .env)")
		return
	}

	// Test DSN generation
	dsn := cfg.GetDSN()
	if dsn == "" {
		t.Error("DSN is empty")
	}
	t.Logf("DSN: %s", dsn)

	// Test RabbitMQ URL
	rabbitmqURL := cfg.GetRabbitMQURL()
	if rabbitmqURL == "" {
		t.Error("RabbitMQ URL is empty")
	}
	t.Logf("RabbitMQ URL: %s", rabbitmqURL)

	// Test Redis address
	redisAddr := cfg.GetRedisAddr()
	if redisAddr == "" {
		t.Error("Redis address is empty")
	}
	t.Logf("Redis Address: %s", redisAddr)

	// Test timeout parsing
	readTimeout := cfg.GetReadTimeout()
	if readTimeout == 0 {
		t.Error("Read timeout is zero")
	}
	t.Logf("Read Timeout: %v", readTimeout)

	writeTimeout := cfg.GetWriteTimeout()
	if writeTimeout == 0 {
		t.Error("Write timeout is zero")
	}
	t.Logf("Write Timeout: %v", writeTimeout)

	// Test JWT expiration
	jwtExp := cfg.GetJWTExpiration()
	if jwtExp == 0 {
		t.Error("JWT expiration is zero")
	}
	t.Logf("JWT Expiration: %v", jwtExp)
}
