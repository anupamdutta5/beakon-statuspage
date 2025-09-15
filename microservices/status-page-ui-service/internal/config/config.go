// Package config provides configuration management for the Status Page UI Service.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the service configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Services ServicesConfig `yaml:"services"`
	Service  ServiceConfig  `yaml:"service"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	Environment  string `yaml:"environment"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// ServicesConfig represents configuration for downstream services.
type ServicesConfig struct {
	TenantAdminService ServiceEndpoint `yaml:"tenant_admin_service"`
}

// ServiceEndpoint represents a service endpoint configuration.
type ServiceEndpoint struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8093),
			Host:         getEnv("HOST", "0.0.0.0"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "status_page_ui"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 3600),
		},
		Services: ServicesConfig{
			TenantAdminService: ServiceEndpoint{
				URL:     getEnv("TENANT_ADMIN_SERVICE_URL", "http://localhost:8094"),
				Timeout: getEnvInt("TENANT_ADMIN_SERVICE_TIMEOUT", 30),
				Retries: getEnvInt("TENANT_ADMIN_SERVICE_RETRIES", 3),
			},
		},
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "status-page-ui-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Status Page UI Service"),
			Tags:        strings.Split(getEnv("SERVICE_TAGS", "status,ui,frontend"), ","),
			Metadata: map[string]string{
				"environment": getEnv("ENVIRONMENT", "development"),
			},
		},
	}

	// Load from config file if it exists
	if configFile := getEnv("CONFIG_FILE", ""); configFile != "" {
		if err := loadFromFile(config, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	return config, nil
}

// loadFromFile loads configuration from a YAML file.
func loadFromFile(config *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer with a default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
