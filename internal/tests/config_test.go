package tests

import (
	"os"
	"testing"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	DatabaseURL string
	JWTSecret   string
	AdminEmail  string
	AdminPass   string
}

// GetTestConfig returns test configuration
func GetTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: getEnv("TEST_DATABASE_URL", "postgres://postgres:password@localhost:5432/statuspage_test?sslmode=disable"),
		JWTSecret:   getEnv("TEST_JWT_SECRET", "test-secret-key"),
		AdminEmail:  getEnv("TEST_ADMIN_EMAIL", "admin@test.com"),
		AdminPass:   getEnv("TEST_ADMIN_PASSWORD", "testpassword"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// SetupTestEnvironment sets up test environment
func SetupTestEnvironment(t *testing.T) {
	// Set test environment variables
	os.Setenv("ENV", "test")
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("ADMIN_EMAIL", "admin@test.com")
	os.Setenv("ADMIN_PASSWORD", "testpassword")
}

// CleanupTestEnvironment cleans up test environment
func CleanupTestEnvironment(t *testing.T) {
	// Clean up environment variables
	os.Unsetenv("ENV")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ADMIN_EMAIL")
	os.Unsetenv("ADMIN_PASSWORD")
}
