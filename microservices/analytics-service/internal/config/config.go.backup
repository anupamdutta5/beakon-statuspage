// Package config provides configuration management for the Analytics Service.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Analytics Service configuration.
type Config struct {
	Server     ServerConfig     `json:"server"`
	Database   DatabaseConfig   `json:"database"`
	JWT        JWTConfig        `json:"jwt"`
	Service    ServiceConfig    `json:"service"`
	Monitoring MonitoringConfig `json:"monitoring"`
	Analytics  AnalyticsConfig  `json:"analytics"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	Environment  string `json:"environment"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
	MaxConns int    `json:"max_conns"`
	MinConns int    `json:"min_conns"`
}

// JWTConfig represents JWT configuration.
type JWTConfig struct {
	Secret     string `json:"secret"`
	Expiration int    `json:"expiration"` // in hours
	Issuer     string `json:"issuer"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

// MonitoringConfig represents monitoring configuration.
type MonitoringConfig struct {
	Enabled     bool   `json:"enabled"`
	MetricsPort int    `json:"metrics_port"`
	HealthPort  int    `json:"health_port"`
	LogLevel    string `json:"log_level"`
}

// AnalyticsConfig represents analytics-specific configuration.
type AnalyticsConfig struct {
	DataRetentionDays   int      `json:"data_retention_days"`
	AggregationInterval string   `json:"aggregation_interval"` // hourly, daily, weekly, monthly
	CacheEnabled        bool     `json:"cache_enabled"`
	CacheTTL            int      `json:"cache_ttl"`      // in seconds
	ExportFormats       []string `json:"export_formats"` // csv, json, excel
	MaxDataPoints       int      `json:"max_data_points"`
	RealTimeEnabled     bool     `json:"real_time_enabled"`
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8090),
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
			Name:     getEnv("DB_NAME", "statuspage_analytics"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 100),
			MinConns: getEnvInt("DB_MIN_CONNS", 10),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "development-secret-key"),
			Expiration: getEnvInt("JWT_EXPIRATION", 24),
			Issuer:     getEnv("JWT_ISSUER", "statuspage-analytics-service"),
		},
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "analytics-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Analytics and Reporting Service for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"analytics-service", "microservice"}),
		},
		Monitoring: MonitoringConfig{
			Enabled:     getEnvBool("MONITORING_ENABLED", true),
			MetricsPort: getEnvInt("METRICS_PORT", 9100),
			HealthPort:  getEnvInt("HEALTH_PORT", 8091),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		Analytics: AnalyticsConfig{
			DataRetentionDays:   getEnvInt("DATA_RETENTION_DAYS", 365),
			AggregationInterval: getEnv("AGGREGATION_INTERVAL", "daily"),
			CacheEnabled:        getEnvBool("CACHE_ENABLED", true),
			CacheTTL:            getEnvInt("CACHE_TTL", 300), // 5 minutes
			ExportFormats:       getEnvSlice("EXPORT_FORMATS", []string{"csv", "json", "excel"}),
			MaxDataPoints:       getEnvInt("MAX_DATA_POINTS", 10000),
			RealTimeEnabled:     getEnvBool("REAL_TIME_ENABLED", true),
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

	return json.Unmarshal(data, config)
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
