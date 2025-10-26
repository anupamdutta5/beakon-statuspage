// Package config provides configuration management for the Notification Service.
package config

import (
	"gopkg.in/yaml.v3"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Notification Service configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Service  ServiceConfig  `yaml:"service"`
	Email    EmailConfig    `yaml:"email"`
	SMS      SMSConfig      `yaml:"sms"`
	Webhook  WebhookConfig  `yaml:"webhook"`
	Queue    QueueConfig    `yaml:"queue"`
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

// EmailConfig represents email configuration.
type EmailConfig struct {
	Enabled    bool   `yaml:"enabled"`
	SMTPHost   string `yaml:"smtp_host"`
	SMTPPort   int    `yaml:"smtp_port"`
	SMTPUser   string `yaml:"smtp_user"`
	SMTPPass   string `yaml:"smtp_pass"`
	FromEmail  string `yaml:"from_email"`
	FromName   string `yaml:"from_name"`
	UseTLS     bool   `yaml:"use_tls"`
	UseSSL     bool   `yaml:"use_ssl"`
	MaxRetries int    `yaml:"max_retries"`
	RetryDelay int    `yaml:"retry_delay"` // in seconds
}

// SMSConfig represents SMS configuration.
type SMSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Provider   string `yaml:"provider"` // twilio, aws_sns, etc.
	AccountSID string `yaml:"account_sid"`
	AuthToken  string `yaml:"auth_token"`
	FromNumber string `yaml:"from_number"`
	MaxRetries int    `yaml:"max_retries"`
	RetryDelay int    `yaml:"retry_delay"` // in seconds
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	Enabled    bool   `yaml:"enabled"`
	MaxRetries int    `yaml:"max_retries"`
	RetryDelay int    `yaml:"retry_delay"` // in seconds
	Timeout    int    `yaml:"timeout"`     // in seconds
	Secret     string `yaml:"secret"`
}

// QueueConfig represents message queue configuration.
type QueueConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Provider   string `yaml:"provider"` // redis, rabbitmq, kafka
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	QueueName  string `yaml:"queue_name"`
	MaxRetries int    `yaml:"max_retries"`
	RetryDelay int    `yaml:"retry_delay"` // in seconds
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8085),
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
			Name:     getEnv("DB_NAME", "notifications"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 100),
			MinConns: getEnvInt("DB_MIN_CONNS", 10),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "development-secret-key"),
			Expiration: getEnvInt("JWT_EXPIRATION", 24),
			Issuer:     getEnv("JWT_ISSUER", "statuspage-notification-service"),
		},
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "notification-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Notification Service for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"notification-service", "microservice"}),
		},
		Email: EmailConfig{
			Enabled:    getEnvBool("EMAIL_ENABLED", true),
			SMTPHost:   getEnv("SMTP_HOST", "localhost"),
			SMTPPort:   getEnvInt("SMTP_PORT", 587),
			SMTPUser:   getEnv("SMTP_USER", ""),
			SMTPPass:   getEnv("SMTP_PASS", ""),
			FromEmail:  getEnv("FROM_EMAIL", "noreply@statuspage.com"),
			FromName:   getEnv("FROM_NAME", "Status Page"),
			UseTLS:     getEnvBool("SMTP_USE_TLS", true),
			UseSSL:     getEnvBool("SMTP_USE_SSL", false),
			MaxRetries: getEnvInt("EMAIL_MAX_RETRIES", 3),
			RetryDelay: getEnvInt("EMAIL_RETRY_DELAY", 60),
		},
		SMS: SMSConfig{
			Enabled:    getEnvBool("SMS_ENABLED", false),
			Provider:   getEnv("SMS_PROVIDER", "twilio"),
			AccountSID: getEnv("SMS_ACCOUNT_SID", ""),
			AuthToken:  getEnv("SMS_AUTH_TOKEN", ""),
			FromNumber: getEnv("SMS_FROM_NUMBER", ""),
			MaxRetries: getEnvInt("SMS_MAX_RETRIES", 3),
			RetryDelay: getEnvInt("SMS_RETRY_DELAY", 60),
		},
		Webhook: WebhookConfig{
			Enabled:    getEnvBool("WEBHOOK_ENABLED", true),
			MaxRetries: getEnvInt("WEBHOOK_MAX_RETRIES", 3),
			RetryDelay: getEnvInt("WEBHOOK_RETRY_DELAY", 60),
			Timeout:    getEnvInt("WEBHOOK_TIMEOUT", 30),
			Secret:     getEnv("WEBHOOK_SECRET", ""),
		},
		Queue: QueueConfig{
			Enabled:    getEnvBool("QUEUE_ENABLED", false),
			Provider:   getEnv("QUEUE_PROVIDER", "redis"),
			Host:       getEnv("QUEUE_HOST", "localhost"),
			Port:       getEnvInt("QUEUE_PORT", 6379),
			Username:   getEnv("QUEUE_USERNAME", ""),
			Password:   getEnv("QUEUE_PASSWORD", ""),
			QueueName:  getEnv("QUEUE_NAME", "notifications"),
			MaxRetries: getEnvInt("QUEUE_MAX_RETRIES", 3),
			RetryDelay: getEnvInt("QUEUE_RETRY_DELAY", 60),
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
