package resilience

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains all shared configuration for resilience patterns
type Config struct {
	Environment    string                 `json:"environment"`
	Server         ServerConfig           `json:"server"`
	CircuitBreaker CircuitBreakerConfigs  `json:"circuit_breaker"`
	Cache          CacheConfig            `json:"cache"`
	RateLimit      RateLimitConfig        `json:"rate_limit"`
	Security       SecurityConfig         `json:"security"`
	Database       DatabaseConfig         `json:"database"`
	Redis          RedisConfig            `json:"redis"`
	JWT            JWTConfig              `json:"jwt"`
	CORS           CORSConfig             `json:"cors"`
	Monitoring     MonitoringConfig       `json:"monitoring"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	GracefulStop time.Duration `json:"graceful_stop"`
}

// CircuitBreakerConfigs contains configurations for multiple circuit breakers
type CircuitBreakerConfigs struct {
	Database CircuitBreakerConfig `json:"database"`
	External CircuitBreakerConfig `json:"external"`
	Internal CircuitBreakerConfig `json:"internal"`
}



// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `json:"enabled"`
	RequestsPerMinute int           `json:"requests_per_minute"`
	Burst             int           `json:"burst"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	RedisKeyPrefix    string        `json:"redis_key_prefix"`
	ExcludedPaths     []string      `json:"excluded_paths"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	SanitizationEnabled   bool     `json:"sanitization_enabled"`
	MaxStringLength       int      `json:"max_string_length"`
	StrictMode            bool     `json:"strict_mode"`
	AllowedFileTypes      []string `json:"allowed_file_types"`
	MaxFileSize           int64    `json:"max_file_size"`
	SecurityHeadersEnabled bool    `json:"security_headers_enabled"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Name            string        `json:"name"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	MaxRetries   int    `json:"max_retries"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	KeyPrefix    string `json:"key_prefix"`
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	Secret            string        `json:"secret"`
	Expiration        time.Duration `json:"expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration"`
	Issuer            string        `json:"issuer"`
	Audience          string        `json:"audience"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
	ExposeHeaders    []string `json:"expose_headers"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Enabled         bool    `json:"enabled"`
	MetricsEnabled  bool    `json:"metrics_enabled"`
	TracingEnabled  bool    `json:"tracing_enabled"`
	MetricsPath     string  `json:"metrics_path"`
	HealthPath      string  `json:"health_path"`
	JaegerEndpoint  string  `json:"jaeger_endpoint"`
	TracingSampling float64 `json:"tracing_sampling"`
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),

		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", "30s"),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", "120s"),
			GracefulStop: getEnvDuration("SERVER_GRACEFUL_STOP", "30s"),
		},

		CircuitBreaker: CircuitBreakerConfigs{
			Database: CircuitBreakerConfig{
				Name:         "database",
				MaxRequests:  uint32(getEnvInt("CB_DATABASE_MAX_REQUESTS", 5)),
				Interval:     getEnvDuration("CB_DATABASE_INTERVAL", "10s"),
				Timeout:      getEnvDuration("CB_DATABASE_TIMEOUT", "60s"),
				FailureRatio: getEnvFloat("CB_DATABASE_FAILURE_RATIO", 0.6),
				Enabled:      getEnvBool("CB_DATABASE_ENABLED", true),
			},
			External: CircuitBreakerConfig{
				Name:         "external",
				MaxRequests:  uint32(getEnvInt("CB_EXTERNAL_MAX_REQUESTS", 3)),
				Interval:     getEnvDuration("CB_EXTERNAL_INTERVAL", "10s"),
				Timeout:      getEnvDuration("CB_EXTERNAL_TIMEOUT", "60s"),
				FailureRatio: getEnvFloat("CB_EXTERNAL_FAILURE_RATIO", 0.6),
				Enabled:      getEnvBool("CB_EXTERNAL_ENABLED", true),
			},
			Internal: CircuitBreakerConfig{
				Name:         "internal",
				MaxRequests:  uint32(getEnvInt("CB_INTERNAL_MAX_REQUESTS", 10)),
				Interval:     getEnvDuration("CB_INTERNAL_INTERVAL", "5s"),
				Timeout:      getEnvDuration("CB_INTERNAL_TIMEOUT", "30s"),
				FailureRatio: getEnvFloat("CB_INTERNAL_FAILURE_RATIO", 0.5),
				Enabled:      getEnvBool("CB_INTERNAL_ENABLED", true),
			},
		},

		Cache: CacheConfig{
			Enabled:         getEnvBool("CACHE_ENABLED", true),
			DefaultTTL:      getEnvDuration("CACHE_DEFAULT_TTL", "5m"),
			MaxSize:         getEnvInt("CACHE_MAX_SIZE", 1000),
			CleanupInterval: getEnvDuration("CACHE_CLEANUP_INTERVAL", "10m"),
			Type:            getEnv("CACHE_TYPE", "memory"),
		},

		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
			Burst:             getEnvInt("RATE_LIMIT_BURST", 10),
			CleanupInterval:   getEnvDuration("RATE_LIMIT_CLEANUP_INTERVAL", "5m"),
			RedisKeyPrefix:    getEnv("RATE_LIMIT_REDIS_KEY_PREFIX", "rate_limit:"),
			ExcludedPaths:     getEnvSlice("RATE_LIMIT_EXCLUDED_PATHS", []string{"/health", "/metrics"}),
		},

		Security: SecurityConfig{
			SanitizationEnabled:    getEnvBool("SECURITY_SANITIZATION_ENABLED", true),
			MaxStringLength:        getEnvInt("SECURITY_MAX_STRING_LENGTH", 1000),
			StrictMode:             getEnvBool("SECURITY_STRICT_MODE", false),
			AllowedFileTypes:       getEnvSlice("SECURITY_ALLOWED_FILE_TYPES", []string{"jpg", "jpeg", "png", "gif", "pdf"}),
			MaxFileSize:            int64(getEnvInt("SECURITY_MAX_FILE_SIZE", 10485760)), // 10MB
			SecurityHeadersEnabled: getEnvBool("SECURITY_HEADERS_ENABLED", true),
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
			ExposeHeaders:    getEnvSlice("CORS_EXPOSE_HEADERS", []string{"X-Correlation-ID", "X-Total-Count"}),
		},

		Monitoring: MonitoringConfig{
			Enabled:         getEnvBool("MONITORING_ENABLED", true),
			MetricsEnabled:  getEnvBool("METRICS_ENABLED", true),
			TracingEnabled:  getEnvBool("TRACING_ENABLED", false),
			MetricsPath:     getEnv("METRICS_PATH", "/metrics"),
			HealthPath:      getEnv("HEALTH_PATH", "/health"),
			JaegerEndpoint:  getEnv("JAEGER_ENDPOINT", ""),
			TracingSampling: getEnvFloat("TRACING_SAMPLING", 0.1),
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if c.Environment == "production" {
		return c.validateProduction()
	}

	return nil
}

// validateProduction performs additional validation for production environment
func (c *Config) validateProduction() error {
	// Check for wildcard CORS origins
	for _, origin := range c.CORS.AllowedOrigins {
		if origin == "*" {
			return fmt.Errorf("wildcard CORS origins not allowed in production")
		}
	}

	// Ensure SSL is enabled
	if c.Database.SSLMode == "disable" {
		return fmt.Errorf("SSL must be enabled for database connections in production")
	}

	// Ensure monitoring is enabled
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