// Package config provides configuration management for the Tenant Admin Service.
package config

import (
	"fmt"
	"time"

	"github.com/anupamdutta5/shared-resilience"
)

// Config represents the Tenant Admin Service configuration.
type Config struct {
	Service    ServiceConfig    `yaml:"service"`
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	RabbitMQ   RabbitMQConfig   `yaml:"rabbitmq"`
	JWT        JWTConfig        `yaml:"jwt"`
	CORS       CORSConfig       `yaml:"cors"`
	Session    SessionConfig    `yaml:"session"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	RateLimit  RateLimitConfig  `yaml:"rate_limiting"`
	Security   SecurityConfig   `yaml:"security"`
	Tenant     TenantConfig     `yaml:"tenant"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxConns        int    `yaml:"max_conns"`
	MinConns        int    `yaml:"min_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
}

// RedisConfig represents Redis configuration.
type RedisConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	MaxRetries   int    `yaml:"max_retries"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

// RabbitMQConfig represents RabbitMQ configuration.
type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	VHost    string `yaml:"vhost"`
}

// JWTConfig represents JWT configuration.
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	Expiration string `yaml:"expiration"`
	Issuer     string `yaml:"issuer"`
}

// CORSConfig represents CORS configuration.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// SessionConfig represents session configuration.
type SessionConfig struct {
	CookieName     string `yaml:"cookie_name"`
	CookieSecure   bool   `yaml:"cookie_secure"`
	CookieHTTPOnly bool   `yaml:"cookie_http_only"`
	CookieSameSite string `yaml:"cookie_same_site"`
	SessionTTL     int    `yaml:"session_ttl"`
}

// MonitoringConfig represents monitoring configuration.
type MonitoringConfig struct {
	Enabled     bool   `yaml:"enabled"`
	MetricsPort int    `yaml:"metrics_port"`
	HealthPath  string `yaml:"health_path"`
	MetricsPath string `yaml:"metrics_path"`
	LogLevel    string `yaml:"log_level"`
}

// RateLimitConfig represents rate limiting configuration.
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	Burst             int  `yaml:"burst"`
}

// SecurityConfig represents security configuration.
type SecurityConfig struct {
	SanitizationEnabled bool `yaml:"sanitization_enabled"`
	MaxStringLength     int  `yaml:"max_string_length"`
	StrictMode          bool `yaml:"strict_mode"`
}

// TenantConfig represents tenant-specific limits and features.
type TenantConfig struct {
	DefaultPlan             string   `yaml:"default_plan"`
	AvailablePlans          []string `yaml:"available_plans"`
	MaxUsersPerTenant       int      `yaml:"max_users_per_tenant"`
	MaxServicesPerTenant    int      `yaml:"max_services_per_tenant"`
	MaxMonitorsPerTenant    int      `yaml:"max_monitors_per_tenant"`
	MaxSubscribersPerTenant int      `yaml:"max_subscribers_per_tenant"`
	MaxIncidentsPerTenant   int      `yaml:"max_incidents_per_tenant"`
	MaxMaintenancePerTenant int      `yaml:"max_maintenance_per_tenant"`
	CustomDomainEnabled     bool     `yaml:"custom_domain_enabled"`
	WhiteLabelEnabled       bool     `yaml:"white_label_enabled"`
	APIEnabled              bool     `yaml:"api_enabled"`
	IntegrationsEnabled     bool     `yaml:"integrations_enabled"`
	AnalyticsEnabled        bool     `yaml:"analytics_enabled"`
	SupportLevel            string   `yaml:"support_level"`
	FeatureFlagsEnabled     bool     `yaml:"feature_flags_enabled"`
	AuditLoggingEnabled     bool     `yaml:"audit_logging_enabled"`
	BackupEnabled           bool     `yaml:"backup_enabled"`
	BackupInterval          int      `yaml:"backup_interval"`
	RetentionDays           int      `yaml:"retention_days"`
}

// Load loads configuration from YAML files with .env secret injection.
func Load() (*Config, error) {
	// Create configuration loader pointing to configs/ directory
	loader := resilience.NewConfigLoader("configs")

	// Load configuration from YAML files
	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate JWT secret
	if err := resilience.ValidateSecrets(c.JWT.Secret, c.Database.Password); err != nil {
		return fmt.Errorf("secret validation failed: %w", err)
	}

	// Validate RabbitMQ secret
	if err := resilience.ValidateRabbitMQSecret(c.RabbitMQ.Password); err != nil {
		return fmt.Errorf("rabbitmq secret validation failed: %w", err)
	}

	// Validate Redis secret (only if enabled)
	if err := resilience.ValidateRedisSecret(c.Redis.Password, c.Redis.Enabled); err != nil {
		return fmt.Errorf("redis secret validation failed: %w", err)
	}

	// Validate server port
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}

	// Validate database configuration
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate tenant configuration
	if c.Tenant.DefaultPlan == "" {
		return fmt.Errorf("tenant default plan is required")
	}

	return nil
}

// GetDSN returns the database connection string.
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRabbitMQURL returns the RabbitMQ connection URL.
func (c *Config) GetRabbitMQURL() string {
	vhost := c.RabbitMQ.VHost
	if vhost == "" || vhost == "/" {
		vhost = ""
	} else {
		vhost = "/" + vhost
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.RabbitMQ.User,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		vhost,
	)
}

// GetRedisAddr returns the Redis connection address.
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetReadTimeout parses and returns the read timeout duration.
func (c *Config) GetReadTimeout() time.Duration {
	d, err := time.ParseDuration(c.Server.ReadTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// GetWriteTimeout parses and returns the write timeout duration.
func (c *Config) GetWriteTimeout() time.Duration {
	d, err := time.ParseDuration(c.Server.WriteTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// GetIdleTimeout parses and returns the idle timeout duration.
func (c *Config) GetIdleTimeout() time.Duration {
	d, err := time.ParseDuration(c.Server.IdleTimeout)
	if err != nil {
		return 120 * time.Second
	}
	return d
}

// GetJWTExpiration parses and returns the JWT expiration duration.
func (c *Config) GetJWTExpiration() time.Duration {
	d, err := time.ParseDuration(c.JWT.Expiration)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

// GetConnMaxLifetimeDuration parses and returns the connection max lifetime duration.
func (c *DatabaseConfig) GetConnMaxLifetimeDuration() time.Duration {
	d, err := time.ParseDuration(c.ConnMaxLifetime)
	if err != nil {
		return 3600 * time.Second // 1 hour default
	}
	return d
}

// GetConnMaxIdleTimeDuration parses and returns the connection max idle time duration.
func (c *DatabaseConfig) GetConnMaxIdleTimeDuration() time.Duration {
	d, err := time.ParseDuration(c.ConnMaxIdleTime)
	if err != nil {
		return 600 * time.Second // 10 minutes default
	}
	return d
}

// Helper properties for backward compatibility with old code
// MaxIdle returns MinConns (for backward compatibility)
func (c *DatabaseConfig) MaxIdle() int {
	return c.MinConns
}

// MaxLifetime returns connection max lifetime in seconds (for backward compatibility)
func (c *DatabaseConfig) MaxLifetime() int {
	return int(c.GetConnMaxLifetimeDuration().Seconds())
}

// LoadServiceEndpoints loads service endpoints from service-endpoints.yml.
func LoadServiceEndpoints() (map[string]resilience.ServiceEndpoint, error) {
	loader := resilience.NewConfigLoader("configs")
	endpoints, err := loader.LoadServiceEndpoints()
	if err != nil {
		return nil, fmt.Errorf("failed to load service endpoints: %w", err)
	}
	return endpoints, nil
}
