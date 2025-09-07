package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Config represents the application configuration
type Config struct {
	// Server Configuration
	Server ServerConfig `json:"server" yaml:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Redis Configuration
	Redis RedisConfig `json:"redis" yaml:"redis"`

	// JWT Configuration
	JWT JWTConfig `json:"jwt" yaml:"jwt"`

	// Email Configuration
	Email EmailConfig `json:"email" yaml:"email"`

	// Payment Configuration
	Payment PaymentConfig `json:"payment" yaml:"payment"`

	// Monitoring Configuration
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`

	// Security Configuration
	Security SecurityConfig `json:"security" yaml:"security"`

	// Service Discovery Configuration
	ServiceDiscovery ServiceDiscoveryConfig `json:"service_discovery" yaml:"service_discovery"`

	// Kafka Configuration
	Kafka KafkaConfig `json:"kafka" yaml:"kafka"`

	// Tracing Configuration
	Tracing TracingConfig `json:"tracing" yaml:"tracing"`

	// Environment
	Environment string `json:"environment" yaml:"environment"`

	// Logging Configuration
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	TLS          TLSConfig     `json:"tls" yaml:"tls"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`
	MaxConns int    `json:"max_conns" yaml:"max_conns"`
	MinConns int    `json:"min_conns" yaml:"min_conns"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret    string        `json:"secret" yaml:"secret"`
	ExpiresIn time.Duration `json:"expires_in" yaml:"expires_in"`
	RefreshIn time.Duration `json:"refresh_in" yaml:"refresh_in"`
	Issuer    string        `json:"issuer" yaml:"issuer"`
	Audience  string        `json:"audience" yaml:"audience"`
}

// EmailConfig represents email configuration
type EmailConfig struct {
	Provider string         `json:"provider" yaml:"provider"`
	SMTP     SMTPConfig     `json:"smtp" yaml:"smtp"`
	SendGrid SendGridConfig `json:"sendgrid" yaml:"sendgrid"`
}

// SMTPConfig represents SMTP configuration
type SMTPConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	From     string `json:"from" yaml:"from"`
	UseTLS   bool   `json:"use_tls" yaml:"use_tls"`
}

// SendGridConfig represents SendGrid configuration
type SendGridConfig struct {
	APIKey string `json:"api_key" yaml:"api_key"`
	From   string `json:"from" yaml:"from"`
}

// PaymentConfig represents payment configuration
type PaymentConfig struct {
	DefaultProvider string                 `json:"default_provider" yaml:"default_provider"`
	Providers       map[string]interface{} `json:"providers" yaml:"providers"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Prometheus PrometheusConfig `json:"prometheus" yaml:"prometheus"`
	Jaeger     JaegerConfig     `json:"jaeger" yaml:"jaeger"`
	InfluxDB   InfluxDBConfig   `json:"influxdb" yaml:"influxdb"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Port    int    `json:"port" yaml:"port"`
	Path    string `json:"path" yaml:"path"`
}

// JaegerConfig represents Jaeger configuration
type JaegerConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Endpoint    string `json:"endpoint" yaml:"endpoint"`
	ServiceName string `json:"service_name" yaml:"service_name"`
}

// InfluxDBConfig represents InfluxDB configuration
type InfluxDBConfig struct {
	URL    string `json:"url" yaml:"url"`
	Token  string `json:"token" yaml:"token"`
	Org    string `json:"org" yaml:"org"`
	Bucket string `json:"bucket" yaml:"bucket"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	RateLimit  RateLimitConfig  `json:"rate_limit" yaml:"rate_limit"`
	CORS       CORSConfig       `json:"cors" yaml:"cors"`
	Encryption EncryptionConfig `json:"encryption" yaml:"encryption"`
	Session    SessionConfig    `json:"session" yaml:"session"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	Rate    int           `json:"rate" yaml:"rate"`
	Burst   int           `json:"burst" yaml:"burst"`
	Window  time.Duration `json:"window" yaml:"window"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Key       string `json:"key" yaml:"key"`
	Algorithm string `json:"algorithm" yaml:"algorithm"`
}

// SessionConfig represents session configuration
type SessionConfig struct {
	Secret   string        `json:"secret" yaml:"secret"`
	MaxAge   time.Duration `json:"max_age" yaml:"max_age"`
	Secure   bool          `json:"secure" yaml:"secure"`
	HTTPOnly bool          `json:"http_only" yaml:"http_only"`
	SameSite string        `json:"same_site" yaml:"same_site"`
}

// ServiceDiscoveryConfig represents service discovery configuration
type ServiceDiscoveryConfig struct {
	Provider string       `json:"provider" yaml:"provider"`
	Consul   ConsulConfig `json:"consul" yaml:"consul"`
}

// ConsulConfig represents Consul configuration
type ConsulConfig struct {
	Address string `json:"address" yaml:"address"`
	Token   string `json:"token" yaml:"token"`
	DC      string `json:"dc" yaml:"dc"`
}

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers []string          `json:"brokers" yaml:"brokers"`
	Topics  map[string]string `json:"topics" yaml:"topics"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	ServiceName string  `json:"service_name" yaml:"service_name"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Load from config file if provided
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			// Log warning if logger is available, otherwise use fmt
			if logger.Log != nil {
				logger.Log.Warn("Failed to load config file, using defaults", zap.String("path", configPath), zap.Error(err))
			}
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Set defaults for any missing values
	setDefaults(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromFile loads configuration from a file
func loadFromFile(config *Config, path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".json":
		return loadFromJSON(config, path)
	case ".yaml", ".yml":
		return loadFromYAML(config, path)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// loadFromJSON loads configuration from JSON file
func loadFromJSON(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

// loadFromYAML loads configuration from YAML file
func loadFromYAML(config *Config, path string) error {
	// For now, we'll use JSON unmarshaling
	// In production, you'd use a YAML library like gopkg.in/yaml.v2
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Server Configuration
	config.Server.Host = getEnv("SERVER_HOST", config.Server.Host)
	config.Server.Port = getEnvInt("SERVER_PORT", config.Server.Port)
	config.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", config.Server.ReadTimeout)
	config.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", config.Server.WriteTimeout)
	config.Server.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", config.Server.IdleTimeout)

	// Database Configuration
	config.Database.Host = getEnv("DB_HOST", config.Database.Host)
	config.Database.Port = getEnvInt("DB_PORT", config.Database.Port)
	config.Database.User = getEnv("DB_USER", config.Database.User)
	config.Database.Password = getEnv("DB_PASSWORD", config.Database.Password)
	config.Database.Name = getEnv("DB_NAME", config.Database.Name)
	config.Database.SSLMode = getEnv("DB_SSLMODE", config.Database.SSLMode)
	config.Database.MaxConns = getEnvInt("DB_MAX_CONNS", config.Database.MaxConns)
	config.Database.MinConns = getEnvInt("DB_MIN_CONNS", config.Database.MinConns)

	// Redis Configuration
	config.Redis.Host = getEnv("REDIS_HOST", config.Redis.Host)
	config.Redis.Port = getEnvInt("REDIS_PORT", config.Redis.Port)
	config.Redis.Password = getEnv("REDIS_PASSWORD", config.Redis.Password)
	config.Redis.DB = getEnvInt("REDIS_DB", config.Redis.DB)
	config.Redis.PoolSize = getEnvInt("REDIS_POOL_SIZE", config.Redis.PoolSize)

	// JWT Configuration
	config.JWT.Secret = getEnv("JWT_SECRET", config.JWT.Secret)
	config.JWT.ExpiresIn = getEnvDuration("JWT_EXPIRES_IN", config.JWT.ExpiresIn)
	config.JWT.RefreshIn = getEnvDuration("JWT_REFRESH_IN", config.JWT.RefreshIn)
	config.JWT.Issuer = getEnv("JWT_ISSUER", config.JWT.Issuer)
	config.JWT.Audience = getEnv("JWT_AUDIENCE", config.JWT.Audience)

	// Email Configuration
	config.Email.Provider = getEnv("EMAIL_PROVIDER", config.Email.Provider)
	config.Email.SMTP.Host = getEnv("SMTP_HOST", config.Email.SMTP.Host)
	config.Email.SMTP.Port = getEnvInt("SMTP_PORT", config.Email.SMTP.Port)
	config.Email.SMTP.Username = getEnv("SMTP_USERNAME", config.Email.SMTP.Username)
	config.Email.SMTP.Password = getEnv("SMTP_PASSWORD", config.Email.SMTP.Password)
	config.Email.SMTP.From = getEnv("SMTP_FROM", config.Email.SMTP.From)
	config.Email.SMTP.UseTLS = getEnvBool("SMTP_USE_TLS", config.Email.SMTP.UseTLS)
	config.Email.SendGrid.APIKey = getEnv("SENDGRID_API_KEY", config.Email.SendGrid.APIKey)
	config.Email.SendGrid.From = getEnv("SENDGRID_FROM", config.Email.SendGrid.From)

	// Payment Configuration
	config.Payment.DefaultProvider = getEnv("PAYMENT_DEFAULT_PROVIDER", config.Payment.DefaultProvider)

	// Monitoring Configuration
	config.Monitoring.Prometheus.Enabled = getEnvBool("PROMETHEUS_ENABLED", config.Monitoring.Prometheus.Enabled)
	config.Monitoring.Prometheus.Port = getEnvInt("PROMETHEUS_PORT", config.Monitoring.Prometheus.Port)
	config.Monitoring.Prometheus.Path = getEnv("PROMETHEUS_PATH", config.Monitoring.Prometheus.Path)
	config.Monitoring.Jaeger.Enabled = getEnvBool("JAEGER_ENABLED", config.Monitoring.Jaeger.Enabled)
	config.Monitoring.Jaeger.Endpoint = getEnv("JAEGER_ENDPOINT", config.Monitoring.Jaeger.Endpoint)
	config.Monitoring.Jaeger.ServiceName = getEnv("JAEGER_SERVICE_NAME", config.Monitoring.Jaeger.ServiceName)
	config.Monitoring.InfluxDB.URL = getEnv("INFLUX_URL", config.Monitoring.InfluxDB.URL)
	config.Monitoring.InfluxDB.Token = getEnv("INFLUX_TOKEN", config.Monitoring.InfluxDB.Token)
	config.Monitoring.InfluxDB.Org = getEnv("INFLUX_ORG", config.Monitoring.InfluxDB.Org)
	config.Monitoring.InfluxDB.Bucket = getEnv("INFLUX_BUCKET", config.Monitoring.InfluxDB.Bucket)

	// Security Configuration
	config.Security.RateLimit.Enabled = getEnvBool("RATE_LIMIT_ENABLED", config.Security.RateLimit.Enabled)
	config.Security.RateLimit.Rate = getEnvInt("RATE_LIMIT_RATE", config.Security.RateLimit.Rate)
	config.Security.RateLimit.Burst = getEnvInt("RATE_LIMIT_BURST", config.Security.RateLimit.Burst)
	config.Security.RateLimit.Window = getEnvDuration("RATE_LIMIT_WINDOW", config.Security.RateLimit.Window)
	config.Security.Encryption.Key = getEnv("ENCRYPTION_KEY", config.Security.Encryption.Key)
	config.Security.Encryption.Algorithm = getEnv("ENCRYPTION_ALGORITHM", config.Security.Encryption.Algorithm)
	config.Security.Session.Secret = getEnv("SESSION_SECRET", config.Security.Session.Secret)
	config.Security.Session.MaxAge = getEnvDuration("SESSION_MAX_AGE", config.Security.Session.MaxAge)
	config.Security.Session.Secure = getEnvBool("SESSION_SECURE", config.Security.Session.Secure)
	config.Security.Session.HTTPOnly = getEnvBool("SESSION_HTTP_ONLY", config.Security.Session.HTTPOnly)
	config.Security.Session.SameSite = getEnv("SESSION_SAME_SITE", config.Security.Session.SameSite)

	// Service Discovery Configuration
	config.ServiceDiscovery.Provider = getEnv("SERVICE_DISCOVERY_PROVIDER", config.ServiceDiscovery.Provider)
	config.ServiceDiscovery.Consul.Address = getEnv("CONSUL_ADDR", config.ServiceDiscovery.Consul.Address)
	config.ServiceDiscovery.Consul.Token = getEnv("CONSUL_TOKEN", config.ServiceDiscovery.Consul.Token)
	config.ServiceDiscovery.Consul.DC = getEnv("CONSUL_DC", config.ServiceDiscovery.Consul.DC)

	// Kafka Configuration
	config.Kafka.Brokers = getEnvSlice("KAFKA_BROKERS", config.Kafka.Brokers)

	// Tracing Configuration
	config.Tracing.Enabled = getEnvBool("TRACING_ENABLED", config.Tracing.Enabled)
	config.Tracing.ServiceName = getEnv("TRACING_SERVICE_NAME", config.Tracing.ServiceName)
	config.Tracing.Endpoint = getEnv("TRACING_ENDPOINT", config.Tracing.Endpoint)
	config.Tracing.SampleRate = getEnvFloat("TRACING_SAMPLE_RATE", config.Tracing.SampleRate)

	// Logging Configuration
	config.Logging.Level = getEnv("LOG_LEVEL", config.Logging.Level)
	config.Logging.Format = getEnv("LOG_FORMAT", config.Logging.Format)
	config.Logging.Output = getEnv("LOG_OUTPUT", config.Logging.Output)
	config.Logging.Filename = getEnv("LOG_FILENAME", config.Logging.Filename)
	config.Logging.MaxSize = getEnvInt("LOG_MAX_SIZE", config.Logging.MaxSize)
	config.Logging.MaxBackups = getEnvInt("LOG_MAX_BACKUPS", config.Logging.MaxBackups)
	config.Logging.MaxAge = getEnvInt("LOG_MAX_AGE", config.Logging.MaxAge)
	config.Logging.Compress = getEnvBool("LOG_COMPRESS", config.Logging.Compress)

	// Environment
	config.Environment = getEnv("ENVIRONMENT", config.Environment)
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 120 * time.Second
	}

	// Database defaults
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.User == "" {
		config.Database.User = "postgres"
	}
	if config.Database.Name == "" {
		config.Database.Name = "statuspage"
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxConns == 0 {
		config.Database.MaxConns = 25
	}
	if config.Database.MinConns == 0 {
		config.Database.MinConns = 5
	}

	// Redis defaults
	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}
	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}

	// JWT defaults
	if config.JWT.ExpiresIn == 0 {
		config.JWT.ExpiresIn = 24 * time.Hour
	}
	if config.JWT.RefreshIn == 0 {
		config.JWT.RefreshIn = 7 * 24 * time.Hour
	}
	if config.JWT.Issuer == "" {
		config.JWT.Issuer = "statuspage"
	}
	if config.JWT.Audience == "" {
		config.JWT.Audience = "statuspage-api"
	}

	// Email defaults
	if config.Email.Provider == "" {
		config.Email.Provider = "smtp"
	}
	if config.Email.SMTP.Host == "" {
		config.Email.SMTP.Host = "smtp.gmail.com"
	}
	if config.Email.SMTP.Port == 0 {
		config.Email.SMTP.Port = 587
	}
	if config.Email.SMTP.From == "" {
		config.Email.SMTP.From = "noreply@statuspage.com"
	}

	// Monitoring defaults
	if config.Monitoring.Prometheus.Port == 0 {
		config.Monitoring.Prometheus.Port = 9090
	}
	if config.Monitoring.Prometheus.Path == "" {
		config.Monitoring.Prometheus.Path = "/metrics"
	}
	if config.Monitoring.Jaeger.ServiceName == "" {
		config.Monitoring.Jaeger.ServiceName = "statuspage"
	}

	// Security defaults
	if config.Security.RateLimit.Rate == 0 {
		config.Security.RateLimit.Rate = 100
	}
	if config.Security.RateLimit.Burst == 0 {
		config.Security.RateLimit.Burst = 200
	}
	if config.Security.RateLimit.Window == 0 {
		config.Security.RateLimit.Window = time.Minute
	}
	if config.Security.Encryption.Algorithm == "" {
		config.Security.Encryption.Algorithm = "AES-256-GCM"
	}
	if config.Security.Session.MaxAge == 0 {
		config.Security.Session.MaxAge = 24 * time.Hour
	}
	if config.Security.Session.SameSite == "" {
		config.Security.Session.SameSite = "Lax"
	}

	// Service Discovery defaults
	if config.ServiceDiscovery.Provider == "" {
		config.ServiceDiscovery.Provider = "consul"
	}
	if config.ServiceDiscovery.Consul.Address == "" {
		config.ServiceDiscovery.Consul.Address = "localhost:8500"
	}

	// Kafka defaults
	if len(config.Kafka.Brokers) == 0 {
		config.Kafka.Brokers = []string{"localhost:9092"}
	}

	// Tracing defaults
	if config.Tracing.ServiceName == "" {
		config.Tracing.ServiceName = "statuspage"
	}
	if config.Tracing.SampleRate == 0 {
		config.Tracing.SampleRate = 0.1
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	// Environment default
	if config.Environment == "" {
		config.Environment = "development"
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate required fields
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if config.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if config.Security.Encryption.Key == "" {
		return fmt.Errorf("encryption key is required")
	}
	if config.Security.Session.Secret == "" {
		return fmt.Errorf("session secret is required")
	}

	// Validate ranges
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}
	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", config.Redis.Port)
	}

	return nil
}

// Helper functions for environment variable parsing
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

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
