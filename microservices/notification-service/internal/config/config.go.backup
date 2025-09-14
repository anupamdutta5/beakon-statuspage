// Package config provides configuration management for the Notification Service.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Notification Service configuration.
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Service  ServiceConfig  `json:"service"`
	Email    EmailConfig    `json:"email"`
	SMS      SMSConfig      `json:"sms"`
	Webhook  WebhookConfig  `json:"webhook"`
	Queue    QueueConfig    `json:"queue"`
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

// EmailConfig represents email configuration.
type EmailConfig struct {
	Enabled    bool   `json:"enabled"`
	SMTPHost   string `json:"smtp_host"`
	SMTPPort   int    `json:"smtp_port"`
	SMTPUser   string `json:"smtp_user"`
	SMTPPass   string `json:"smtp_pass"`
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
	UseTLS     bool   `json:"use_tls"`
	UseSSL     bool   `json:"use_ssl"`
	MaxRetries int    `json:"max_retries"`
	RetryDelay int    `json:"retry_delay"` // in seconds
}

// SMSConfig represents SMS configuration.
type SMSConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"` // twilio, aws_sns, etc.
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	FromNumber string `json:"from_number"`
	MaxRetries int    `json:"max_retries"`
	RetryDelay int    `json:"retry_delay"` // in seconds
}

// WebhookConfig represents webhook configuration.
type WebhookConfig struct {
	Enabled    bool   `json:"enabled"`
	MaxRetries int    `json:"max_retries"`
	RetryDelay int    `json:"retry_delay"` // in seconds
	Timeout    int    `json:"timeout"`     // in seconds
	Secret     string `json:"secret"`
}

// QueueConfig represents message queue configuration.
type QueueConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"` // redis, rabbitmq, kafka
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	QueueName  string `json:"queue_name"`
	MaxRetries int    `json:"max_retries"`
	RetryDelay int    `json:"retry_delay"` // in seconds
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
			Name:     getEnv("DB_NAME", "statuspage_notifications"),
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
