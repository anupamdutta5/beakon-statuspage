// Package config provides configuration management for the Monitoring Service.
package config

import (
	"gopkg.in/yaml.v3"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Monitoring Service configuration.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	JWT        JWTConfig        `yaml:"jwt"`
	Service    ServiceConfig    `yaml:"service"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
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
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_conns"`
	MinConns int    `yaml:"min_conns"`
}

// JWTConfig represents JWT configuration.
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	Expiration int    `yaml:"expiration"` // in hours
	Issuer     string `yaml:"issuer"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

// MonitoringConfig represents monitoring-specific configuration.
type MonitoringConfig struct {
	Enabled             bool   `yaml:"enabled"`
	MetricsPort         int    `yaml:"metrics_port"`
	HealthPort          int    `yaml:"health_port"`
	LogLevel            string `yaml:"log_level"`
	CheckInterval       int    `yaml:"check_interval"` // in seconds
	AlertCooldown       int    `yaml:"alert_cooldown"` // in minutes
	RetentionDays       int    `yaml:"retention_days"`
	MaxConcurrentChecks int    `yaml:"max_concurrent_checks"`
	TimeoutSeconds      int    `yaml:"timeout_seconds"`
}

// PrometheusConfig represents Prometheus configuration.
type PrometheusConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Port      int               `yaml:"port"`
	Path      string            `yaml:"path"`
	Namespace string            `yaml:"namespace"`
	Subsystem string            `yaml:"subsystem"`
	Labels    map[string]string `yaml:"labels"`
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8092),
			Host:         getEnv("HOST", "0.0.0.0"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "statuspage_monitoring"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 100),
			MinConns: getEnvInt("DB_MIN_CONNS", 10),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "development-secret-key"),
			Expiration: getEnvInt("JWT_EXPIRATION", 24),
			Issuer:     getEnv("JWT_ISSUER", "statuspage-monitoring-service"),
		},
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "monitoring-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Monitoring and Observability Service for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"monitoring-service", "microservice"}),
		},
		Monitoring: MonitoringConfig{
			Enabled:             getEnvBool("MONITORING_ENABLED", true),
			MetricsPort:         getEnvInt("METRICS_PORT", 9102),
			HealthPort:          getEnvInt("HEALTH_PORT", 8093),
			LogLevel:            getEnv("LOG_LEVEL", "info"),
			CheckInterval:       getEnvInt("CHECK_INTERVAL", 60),
			AlertCooldown:       getEnvInt("ALERT_COOLDOWN", 15),
			RetentionDays:       getEnvInt("RETENTION_DAYS", 30),
			MaxConcurrentChecks: getEnvInt("MAX_CONCURRENT_CHECKS", 100),
			TimeoutSeconds:      getEnvInt("TIMEOUT_SECONDS", 30),
		},
		Prometheus: PrometheusConfig{
			Enabled:   getEnvBool("PROMETHEUS_ENABLED", true),
			Port:      getEnvInt("PROMETHEUS_PORT", 9102),
			Path:      getEnv("PROMETHEUS_PATH", "/metrics"),
			Namespace: getEnv("PROMETHEUS_NAMESPACE", "statuspage"),
			Subsystem: getEnv("PROMETHEUS_SUBSYSTEM", "monitoring"),
			Labels: map[string]string{
				"service": getEnv("SERVICE_NAME", "monitoring-service"),
				"version": getEnv("SERVICE_VERSION", "1.0.0"),
			},
		},
	}

	// Load from config file if provided
	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		if err := loadConfigFromFile(config, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	return config, nil
}

// loadConfigFromFile loads configuration from a JSON file.
func loadConfigFromFile(config *Config, configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
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

// getEnvBool gets an environment variable as a boolean with a default value.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvSlice gets an environment variable as a slice with a default value.
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	return nil
}
