package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// BaseConfig contains configuration common to all microservices
type BaseConfig struct {
	Environment string           `json:"environment" validate:"required,oneof=development staging production"`
	Service     ServiceConfig    `json:"service" validate:"required"`
	Server      ServerConfig     `json:"server" validate:"required"`
	Database    DatabaseConfig   `json:"database" validate:"required"`
	Redis       RedisConfig      `json:"redis" validate:"required"`
	JWT         JWTConfig        `json:"jwt" validate:"required"`
	CORS        CORSConfig       `json:"cors" validate:"required"`
	Logging     LoggingConfig    `json:"logging" validate:"required"`
	Monitoring  MonitoringConfig `json:"monitoring" validate:"required"`
	RateLimit   RateLimitConfig  `json:"rate_limit" validate:"required"`
	Features    FeatureFlags     `json:"features" validate:"required"`
}

// ServiceConfig contains service-specific configuration
type ServiceConfig struct {
	Name        string `json:"name" validate:"required"`
	Version     string `json:"version" validate:"required"`
	Description string `json:"description"`
	Port        int    `json:"port" validate:"required,min=1,max=65535"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host" validate:"required"`
	Port         int           `json:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `json:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `json:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `json:"idle_timeout" validate:"required"`
	GracefulStop time.Duration `json:"graceful_stop" validate:"required"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host            string        `json:"host" validate:"required"`
	Port            int           `json:"port" validate:"required,min=1,max=65535"`
	User            string        `json:"user" validate:"required"`
	Password        string        `json:"password" validate:"required"`
	Name            string        `json:"name" validate:"required"`
	SSLMode         string        `json:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `json:"max_open_conns" validate:"min=1"`
	MaxIdleConns    int           `json:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" validate:"required"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host        string `json:"host" validate:"required"`
	Port        int    `json:"port" validate:"required,min=1,max=65535"`
	Password    string `json:"password"`
	DB          int    `json:"db" validate:"min=0,max=15"`
	MaxRetries  int    `json:"max_retries" validate:"min=0"`
	PoolSize    int    `json:"pool_size" validate:"min=1"`
	MinIdleConns int   `json:"min_idle_conns" validate:"min=0"`
	KeyPrefix   string `json:"key_prefix"`
}

// JWTConfig contains JWT token configuration
type JWTConfig struct {
	Secret            string        `json:"secret" validate:"required,min=32"`
	Expiration        time.Duration `json:"expiration" validate:"required"`
	RefreshExpiration time.Duration `json:"refresh_expiration" validate:"required"`
	Issuer            string        `json:"issuer"`
	Audience          string        `json:"audience"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins" validate:"required"`
	AllowedMethods   []string `json:"allowed_methods" validate:"required"`
	AllowedHeaders   []string `json:"allowed_headers" validate:"required"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age" validate:"min=0"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `json:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format string `json:"format" validate:"required,oneof=json console"`
}

// MonitoringConfig contains monitoring and observability configuration
type MonitoringConfig struct {
	Enabled         bool          `json:"enabled"`
	MetricsPath     string        `json:"metrics_path" validate:"required"`
	MetricsPort     int           `json:"metrics_port" validate:"min=1,max=65535"`
	HealthPath      string        `json:"health_path" validate:"required"`
	JaegerEndpoint  string        `json:"jaeger_endpoint"`
	TracingSampling float64       `json:"tracing_sampling" validate:"min=0,max=1"`
	InfluxURL       string        `json:"influx_url"`
	InfluxToken     string        `json:"influx_token"`
	InfluxOrg       string        `json:"influx_org"`
	InfluxBucket    string        `json:"influx_bucket"`
	ConsulAddr      string        `json:"consul_addr"`
	ServiceRegistry ServiceConfig `json:"service_registry"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled            bool          `json:"enabled"`
	RequestsPerMinute  int           `json:"requests_per_minute" validate:"min=1"`
	Burst              int           `json:"burst" validate:"min=1"`
	RedisKeyPrefix     string        `json:"redis_key_prefix"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	ExcludedPaths      []string      `json:"excluded_paths"`
}

// FeatureFlags contains feature toggle configuration
type FeatureFlags struct {
	MetricsEnabled       bool `json:"metrics_enabled"`
	TracingEnabled       bool `json:"tracing_enabled"`
	CacheEnabled         bool `json:"cache_enabled"`
	RateLimitingEnabled  bool `json:"rate_limiting_enabled"`
	CircuitBreakerEnabled bool `json:"circuit_breaker_enabled"`
	HealthChecksEnabled  bool `json:"health_checks_enabled"`
}

// LoadBaseConfig loads the base configuration from environment variables
func LoadBaseConfig() (*BaseConfig, error) {
	config := &BaseConfig{
		Environment: getEnv("ENVIRONMENT", "development"),
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "beakon-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Beakon microservice"),
			Port:        getEnvInt("SERVICE_PORT", 8080),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", "60s"),
			GracefulStop: getEnvDuration("SERVER_GRACEFUL_STOP", "10s"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "beakon"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", "300s"),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", "300s"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			KeyPrefix:    getEnv("REDIS_KEY_PREFIX", "beakon:"),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", ""),
			Expiration:        getEnvDuration("JWT_EXPIRATION", "24h"),
			RefreshExpiration: getEnvDuration("JWT_REFRESH_EXPIRATION", "168h"),
			Issuer:            getEnv("JWT_ISSUER", "beakon"),
			Audience:          getEnv("JWT_AUDIENCE", "beakon-users"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods:   getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
			AllowedHeaders:   getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With", "X-Correlation-ID"}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvInt("CORS_MAX_AGE", 3600),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Monitoring: MonitoringConfig{
			Enabled:         getEnvBool("MONITORING_ENABLED", true),
			MetricsPath:     getEnv("METRICS_PATH", "/metrics"),
			MetricsPort:     getEnvInt("METRICS_PORT", 9090),
			HealthPath:      getEnv("HEALTH_PATH", "/health"),
			JaegerEndpoint:  getEnv("JAEGER_ENDPOINT", ""),
			TracingSampling: getEnvFloat("TRACING_SAMPLING", 0.1),
			InfluxURL:       getEnv("INFLUX_URL", ""),
			InfluxToken:     getEnv("INFLUX_TOKEN", ""),
			InfluxOrg:       getEnv("INFLUX_ORG", "beakon"),
			InfluxBucket:    getEnv("INFLUX_BUCKET", "metrics"),
			ConsulAddr:      getEnv("CONSUL_ADDR", "localhost:8500"),
		},
		RateLimit: RateLimitConfig{
			Enabled:            getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute:  getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
			Burst:              getEnvInt("RATE_LIMIT_BURST", 10),
			RedisKeyPrefix:     getEnv("RATE_LIMIT_REDIS_KEY_PREFIX", "rate_limit:"),
			CleanupInterval:    getEnvDuration("RATE_LIMIT_CLEANUP_INTERVAL", "5m"),
			ExcludedPaths:      getEnvSlice("RATE_LIMIT_EXCLUDED_PATHS", []string{"/health", "/metrics"}),
		},
		Features: FeatureFlags{
			MetricsEnabled:        getEnvBool("FEATURE_METRICS_ENABLED", true),
			TracingEnabled:        getEnvBool("FEATURE_TRACING_ENABLED", true),
			CacheEnabled:          getEnvBool("FEATURE_CACHE_ENABLED", true),
			RateLimitingEnabled:   getEnvBool("FEATURE_RATE_LIMITING_ENABLED", true),
			CircuitBreakerEnabled: getEnvBool("FEATURE_CIRCUIT_BREAKER_ENABLED", true),
			HealthChecksEnabled:   getEnvBool("FEATURE_HEALTH_CHECKS_ENABLED", true),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *BaseConfig) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required and cannot be empty")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required and cannot be empty")
	}

	if c.Environment == "production" {
		if err := c.validateProduction(); err != nil {
			return fmt.Errorf("production validation failed: %w", err)
		}
	}

	return nil
}

// validateProduction performs additional validation for production environment
func (c *BaseConfig) validateProduction() error {
	// Ensure no wildcard CORS origins in production
	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			return fmt.Errorf("wildcard CORS origins not allowed in production")
		}
	}

	// Ensure SSL is enabled in production
	if c.Database.SSLMode == "disable" {
		return fmt.Errorf("SSL must be enabled for database connections in production")
	}

	// Ensure monitoring is enabled in production
	if !c.Monitoring.Enabled {
		return fmt.Errorf("monitoring must be enabled in production")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// GetRedisAddr returns the Redis connection address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Environment variable helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 0
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}