package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"saas-admin-service/internal/config"
)

func TestLoadConfig_Success(t *testing.T) {
	// Arrange - Set environment variables
	os.Setenv("SERVER_PORT", "8098")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "saas_admin")
	os.Setenv("JWT_SECRET", "test-secret-key-minimum-32-characters-long")

	defer func() {
		// Cleanup
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
	}()

	// Act
	cfg, err := config.LoadConfig()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8098", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "saas_admin", cfg.Database.Name)
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Arrange - Clear relevant environment variables
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("ENVIRONMENT")

	// Set minimum required variables
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("JWT_SECRET", "test-secret-key-minimum-32-characters-long")

	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("JWT_SECRET")
	}()

	// Act
	cfg, err := config.LoadConfig()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8098", cfg.Server.Port)          // Default port
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)       // Default host
	assert.Equal(t, "development", cfg.Environment)    // Default environment
}

func TestLoadConfig_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		errorMsg    string
	}{
		{
			name: "Missing JWT secret",
			setupEnv: func() {
				os.Setenv("DB_PASSWORD", "testpass")
				os.Unsetenv("JWT_SECRET")
			},
			errorMsg: "JWT_SECRET is required",
		},
		{
			name: "JWT secret too short",
			setupEnv: func() {
				os.Setenv("DB_PASSWORD", "testpass")
				os.Setenv("JWT_SECRET", "short")
			},
			errorMsg: "JWT_SECRET must be at least 32 characters",
		},
		{
			name: "Missing DB password",
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test-secret-key-minimum-32-characters-long")
				os.Unsetenv("DB_PASSWORD")
			},
			errorMsg: "DB_PASSWORD is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tt.setupEnv()
			defer func() {
				os.Unsetenv("DB_PASSWORD")
				os.Unsetenv("JWT_SECRET")
			}()

			// Act
			cfg, err := config.LoadConfig()

			// Assert
			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestConfig_Validate_Success(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Environment: "development",
		Server: config.ServerConfig{
			Port: "8098",
			Host: "0.0.0.0",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "testpass",
			Name:     "saas_admin",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-minimum-32-characters-long",
			Expiration: "24h",
		},
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Environment: "development",
		Server: config.ServerConfig{
			Port: "invalid",
			Host: "0.0.0.0",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "testpass",
			Name:     "saas_admin",
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-minimum-32-characters-long",
			Expiration: "24h",
		},
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server port")
}

func TestConfig_GetDSN(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "testpass",
			Name:     "saas_admin",
			SSLMode:  "disable",
		},
	}

	// Act
	dsn := cfg.GetDSN()

	// Assert
	expected := "host=localhost port=5432 user=postgres password=testpass dbname=saas_admin sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"Production", "production", true},
		{"Development", "development", false},
		{"Staging", "staging", false},
		{"Test", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &config.Config{
				Environment: tt.environment,
			}

			// Act
			result := cfg.IsProduction()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"Production", "production", false},
		{"Development", "development", true},
		{"Staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &config.Config{
				Environment: tt.environment,
			}

			// Act
			result := cfg.IsDevelopment()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetServerAddress(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: "8098",
		},
	}

	// Act
	address := cfg.GetServerAddress()

	// Assert
	assert.Equal(t, "0.0.0.0:8098", address)
}

func TestConfig_GetJWTExpiration(t *testing.T) {
	tests := []struct {
		name       string
		expiration string
		hasError   bool
	}{
		{"Valid hours", "24h", false},
		{"Valid minutes", "30m", false},
		{"Valid days", "7d", true}, // Invalid Go duration format
		{"Invalid format", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret:     "test-secret-key-minimum-32-characters-long",
					Expiration: tt.expiration,
				},
			}

			// Act
			duration, err := cfg.GetJWTExpiration()

			// Assert
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, duration)
			}
		})
	}
}

func TestConfig_GetRabbitMQURL(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		RabbitMQ: config.RabbitMQConfig{
			Host:     "localhost",
			Port:     "5672",
			User:     "guest",
			Password: "guest",
			VHost:    "/",
		},
	}

	// Act
	url := cfg.GetRabbitMQURL()

	// Assert
	expected := "amqp://guest:guest@localhost:5672/"
	assert.Equal(t, expected, url)
}

func TestConfig_GetRedisAddress(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	// Act
	address := cfg.GetRedisAddress()

	// Assert
	expected := "localhost:6379"
	assert.Equal(t, expected, address)
}

func TestConfig_ShouldEnableCORS(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"Production - CORS disabled", "production", false},
		{"Development - CORS enabled", "development", true},
		{"Staging - CORS enabled", "staging", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &config.Config{
				Environment: tt.environment,
			}

			// Act
			result := cfg.ShouldEnableCORS()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}
