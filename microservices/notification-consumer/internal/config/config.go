// Package config provides configuration management for the Notification Consumer.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Notification Consumer configuration.
type Config struct {
	Environment  string             `json:"environment"`
	Service      ServiceConfig      `json:"service"`
	Queue        QueueConfig        `json:"queue"`
	Database     DatabaseConfig     `json:"database"`
	Notification NotificationConfig `json:"notification"`
	Logging      LoggingConfig      `json:"logging"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

// QueueConfig represents message queue configuration.
type QueueConfig struct {
	Provider    string `json:"provider"` // redis, rabbitmq, kafka
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	QueueName   string `json:"queue_name"`
	MaxRetries  int    `json:"max_retries"`
	RetryDelay  int    `json:"retry_delay"` // in seconds
	BatchSize   int    `json:"batch_size"`
	PollTimeout int    `json:"poll_timeout"` // in seconds
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

// NotificationConfig represents notification-specific configuration.
type NotificationConfig struct {
	EmailEnabled    bool   `json:"email_enabled"`
	SMSEnabled      bool   `json:"sms_enabled"`
	WebhookEnabled  bool   `json:"webhook_enabled"`
	PushEnabled     bool   `json:"push_enabled"`
	MaxConcurrency  int    `json:"max_concurrency"`
	ProcessingDelay int    `json:"processing_delay"` // in milliseconds
	RetryBackoff    int    `json:"retry_backoff"`    // in seconds
	DeadLetterQueue string `json:"dead_letter_queue"`
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
			Name:        getEnv("SERVICE_NAME", "notification-consumer"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Notification Consumer for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"notification-consumer", "microservice"}),
		},
		Queue: QueueConfig{
			Provider:    getEnv("QUEUE_PROVIDER", "redis"),
			Host:        getEnv("QUEUE_HOST", "localhost"),
			Port:        getEnvInt("QUEUE_PORT", 6379),
			Username:    getEnv("QUEUE_USERNAME", ""),
			Password:    getEnv("QUEUE_PASSWORD", ""),
			QueueName:   getEnv("QUEUE_NAME", "notifications"),
			MaxRetries:  getEnvInt("QUEUE_MAX_RETRIES", 3),
			RetryDelay:  getEnvInt("QUEUE_RETRY_DELAY", 60),
			BatchSize:   getEnvInt("QUEUE_BATCH_SIZE", 10),
			PollTimeout: getEnvInt("QUEUE_POLL_TIMEOUT", 30),
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
		Notification: NotificationConfig{
			EmailEnabled:    getEnvBool("EMAIL_ENABLED", true),
			SMSEnabled:      getEnvBool("SMS_ENABLED", false),
			WebhookEnabled:  getEnvBool("WEBHOOK_ENABLED", true),
			PushEnabled:     getEnvBool("PUSH_ENABLED", false),
			MaxConcurrency:  getEnvInt("MAX_CONCURRENCY", 10),
			ProcessingDelay: getEnvInt("PROCESSING_DELAY", 100),
			RetryBackoff:    getEnvInt("RETRY_BACKOFF", 5),
			DeadLetterQueue: getEnv("DEAD_LETTER_QUEUE", "notifications-dlq"),
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
	if c.Queue.Host == "" {
		return fmt.Errorf("queue host is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	return nil
}

