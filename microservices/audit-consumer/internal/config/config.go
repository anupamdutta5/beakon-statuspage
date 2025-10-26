// Package config provides configuration management for the Audit Consumer.
package config

import (
	"gopkg.in/yaml.v3"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Audit Consumer configuration.
type Config struct {
	Environment string         `yaml:"environment"`
	Service     ServiceConfig  `yaml:"service"`
	Queue       QueueConfig    `yaml:"queue"`
	Database    DatabaseConfig `yaml:"database"`
	Audit       AuditConfig    `yaml:"audit"`
	Logging     LoggingConfig  `yaml:"logging"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

// QueueConfig represents message queue configuration.
type QueueConfig struct {
	Provider    string `yaml:"provider"` // redis, rabbitmq, kafka
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	QueueName   string `yaml:"queue_name"`
	MaxRetries  int    `yaml:"max_retries"`
	RetryDelay  int    `yaml:"retry_delay"` // in seconds
	BatchSize   int    `yaml:"batch_size"`
	PollTimeout int    `yaml:"poll_timeout"` // in seconds
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

// AuditConfig represents audit-specific configuration.
type AuditConfig struct {
	ProcessingEnabled bool   `yaml:"processing_enabled"`
	ComplianceEnabled bool   `yaml:"compliance_enabled"`
	RetentionDays     int    `yaml:"retention_days"`
	MaxConcurrency    int    `yaml:"max_concurrency"`
	ProcessingDelay   int    `yaml:"processing_delay"` // in milliseconds
	RetryBackoff      int    `yaml:"retry_backoff"`    // in seconds
	DeadLetterQueue   string `yaml:"dead_letter_queue"`
	EncryptionEnabled bool   `yaml:"encryption_enabled"`
	EncryptionKey     string `yaml:"encryption_key"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json, console
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "audit-consumer"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Audit Consumer for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"audit-consumer", "microservice"}),
		},
		Queue: QueueConfig{
			Provider:    getEnv("QUEUE_PROVIDER", "redis"),
			Host:        getEnv("QUEUE_HOST", "localhost"),
			Port:        getEnvInt("QUEUE_PORT", 6379),
			Username:    getEnv("QUEUE_USERNAME", ""),
			Password:    getEnv("QUEUE_PASSWORD", ""),
			QueueName:   getEnv("QUEUE_NAME", "audit"),
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
			Name:     getEnv("DB_NAME", "audit"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 100),
			MinConns: getEnvInt("DB_MIN_CONNS", 10),
		},
		Audit: AuditConfig{
			ProcessingEnabled: getEnvBool("PROCESSING_ENABLED", true),
			ComplianceEnabled: getEnvBool("COMPLIANCE_ENABLED", true),
			RetentionDays:     getEnvInt("RETENTION_DAYS", 2555), // 7 years
			MaxConcurrency:    getEnvInt("MAX_CONCURRENCY", 10),
			ProcessingDelay:   getEnvInt("PROCESSING_DELAY", 100),
			RetryBackoff:      getEnvInt("RETRY_BACKOFF", 5),
			DeadLetterQueue:   getEnv("DEAD_LETTER_QUEUE", "audit-dlq"),
			EncryptionEnabled: getEnvBool("ENCRYPTION_ENABLED", true),
			EncryptionKey:     getEnv("ENCRYPTION_KEY", ""),
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
	if c.Queue.Host == "" {
		return fmt.Errorf("queue host is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	return nil
}

