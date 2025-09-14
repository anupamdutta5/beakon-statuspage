// Package config provides configuration management for the Tenant Admin Service.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Tenant Admin Service configuration.
type Config struct {
	Environment string         `json:"environment"`
	Service     ServiceConfig  `json:"service"`
	Server      ServerConfig   `json:"server"`
	Database    DatabaseConfig `json:"database"`
	Tenant      TenantConfig   `json:"tenant"`
	Logging     LoggingConfig  `json:"logging"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	SSLMode     string `json:"ssl_mode"`
	MaxConns    int    `json:"max_conns"`
	MinConns    int    `json:"min_conns"`
	MaxIdle     int    `json:"max_idle"`
	MaxLifetime int    `json:"max_lifetime"`
}

// TenantConfig represents tenant-specific configuration.
type TenantConfig struct {
	DefaultPlan             string   `json:"default_plan"`
	AvailablePlans          []string `json:"available_plans"`
	MaxUsersPerTenant       int      `json:"max_users_per_tenant"`
	MaxServicesPerTenant    int      `json:"max_services_per_tenant"`
	MaxMonitorsPerTenant    int      `json:"max_monitors_per_tenant"`
	MaxSubscribersPerTenant int      `json:"max_subscribers_per_tenant"`
	MaxIncidentsPerTenant   int      `json:"max_incidents_per_tenant"`
	MaxMaintenancePerTenant int      `json:"max_maintenance_per_tenant"`
	CustomDomainEnabled     bool     `json:"custom_domain_enabled"`
	WhiteLabelEnabled       bool     `json:"white_label_enabled"`
	APIEnabled              bool     `json:"api_enabled"`
	IntegrationsEnabled     bool     `json:"integrations_enabled"`
	AnalyticsEnabled        bool     `json:"analytics_enabled"`
	SupportLevel            string   `json:"support_level"` // email, chat, phone
	FeatureFlagsEnabled     bool     `json:"feature_flags_enabled"`
	AuditLoggingEnabled     bool     `json:"audit_logging_enabled"`
	BackupEnabled           bool     `json:"backup_enabled"`
	BackupInterval          int      `json:"backup_interval"`
	RetentionDays           int      `json:"retention_days"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // json, console
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "tenant-admin-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Tenant Admin Service for Tenant-Specific Administration"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"tenant-admin-service", "microservice", "admin"}),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8099),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnvInt("DB_PORT", 5432),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			Name:        getEnv("DB_NAME", "statuspage_tenant_admin"),
			SSLMode:     getEnv("DB_SSL_MODE", "disable"),
			MaxConns:    getEnvInt("DB_MAX_CONNS", 100),
			MinConns:    getEnvInt("DB_MIN_CONNS", 10),
			MaxIdle:     getEnvInt("DB_MAX_IDLE", 10),
			MaxLifetime: getEnvInt("DB_MAX_LIFETIME", 3600),
		},
		Tenant: TenantConfig{
			DefaultPlan:             getEnv("TENANT_DEFAULT_PLAN", "free"),
			AvailablePlans:          getEnvSlice("TENANT_AVAILABLE_PLANS", []string{"free", "pro", "enterprise"}),
			MaxUsersPerTenant:       getEnvInt("TENANT_MAX_USERS_PER_TENANT", 10),
			MaxServicesPerTenant:    getEnvInt("TENANT_MAX_SERVICES_PER_TENANT", 25),
			MaxMonitorsPerTenant:    getEnvInt("TENANT_MAX_MONITORS_PER_TENANT", 100),
			MaxSubscribersPerTenant: getEnvInt("TENANT_MAX_SUBSCRIBERS_PER_TENANT", 1000),
			MaxIncidentsPerTenant:   getEnvInt("TENANT_MAX_INCIDENTS_PER_TENANT", 100),
			MaxMaintenancePerTenant: getEnvInt("TENANT_MAX_MAINTENANCE_PER_TENANT", 50),
			CustomDomainEnabled:     getEnvBool("TENANT_CUSTOM_DOMAIN_ENABLED", true),
			WhiteLabelEnabled:       getEnvBool("TENANT_WHITE_LABEL_ENABLED", false),
			APIEnabled:              getEnvBool("TENANT_API_ENABLED", true),
			IntegrationsEnabled:     getEnvBool("TENANT_INTEGRATIONS_ENABLED", true),
			AnalyticsEnabled:        getEnvBool("TENANT_ANALYTICS_ENABLED", true),
			SupportLevel:            getEnv("TENANT_SUPPORT_LEVEL", "email"),
			FeatureFlagsEnabled:     getEnvBool("TENANT_FEATURE_FLAGS_ENABLED", true),
			AuditLoggingEnabled:     getEnvBool("TENANT_AUDIT_LOGGING_ENABLED", true),
			BackupEnabled:           getEnvBool("TENANT_BACKUP_ENABLED", true),
			BackupInterval:          getEnvInt("TENANT_BACKUP_INTERVAL", 24), // hours
			RetentionDays:           getEnvInt("TENANT_RETENTION_DAYS", 365), // 1 year
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
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
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be greater than 0")
	}

	if c.Tenant.DefaultPlan == "" {
		return fmt.Errorf("default plan is required")
	}

	return nil
}

