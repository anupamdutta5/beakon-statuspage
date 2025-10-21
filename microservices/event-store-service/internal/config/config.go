// Package config provides configuration management for the Event Store Service.
package config

import (
	"gopkg.in/yaml.v3"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Event Store Service configuration.
type Config struct {
	Environment string           `yaml:"environment"`
	Service     ServiceConfig    `yaml:"service"`
	Server      ServerConfig     `yaml:"server"`
	Database    DatabaseConfig   `yaml:"database"`
	EventStore  EventStoreConfig `yaml:"event_store"`
	Logging     LoggingConfig    `yaml:"logging"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Metadata    map[string]string `yaml:"metadata"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Name        string `yaml:"name"`
	SSLMode     string `yaml:"ssl_mode"`
	MaxConns    int    `yaml:"max_conns"`
	MinConns    int    `yaml:"min_conns"`
	MaxIdle     int    `yaml:"max_idle"`
	MaxLifetime int    `yaml:"max_lifetime"`
}

// EventStoreConfig represents event store specific configuration.
type EventStoreConfig struct {
	MaxEventsPerStream int    `yaml:"max_events_per_stream"`
	SnapshotInterval   int    `yaml:"snapshot_interval"`
	CompressionEnabled bool   `yaml:"compression_enabled"`
	EncryptionEnabled  bool   `yaml:"encryption_enabled"`
	EncryptionKey      string `yaml:"encryption_key"`
	RetentionDays      int    `yaml:"retention_days"`
	MaxEventSize       int    `yaml:"max_event_size"`
	BatchSize          int    `yaml:"batch_size"`
	FlushInterval      int    `yaml:"flush_interval"`
	ReplicationEnabled bool   `yaml:"replication_enabled"`
	ReplicationFactor  int    `yaml:"replication_factor"`
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
			Name:        getEnv("SERVICE_NAME", "event-store-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Event Store Service for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"event-store-service", "microservice"}),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8096),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnvInt("DB_PORT", 5432),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			Name:        getEnv("DB_NAME", "statuspage_eventstore"),
			SSLMode:     getEnv("DB_SSL_MODE", "disable"),
			MaxConns:    getEnvInt("DB_MAX_CONNS", 100),
			MinConns:    getEnvInt("DB_MIN_CONNS", 10),
			MaxIdle:     getEnvInt("DB_MAX_IDLE", 10),
			MaxLifetime: getEnvInt("DB_MAX_LIFETIME", 3600),
		},
		EventStore: EventStoreConfig{
			MaxEventsPerStream: getEnvInt("MAX_EVENTS_PER_STREAM", 10000),
			SnapshotInterval:   getEnvInt("SNAPSHOT_INTERVAL", 100),
			CompressionEnabled: getEnvBool("COMPRESSION_ENABLED", true),
			EncryptionEnabled:  getEnvBool("ENCRYPTION_ENABLED", true),
			EncryptionKey:      getEnv("ENCRYPTION_KEY", ""),
			RetentionDays:      getEnvInt("RETENTION_DAYS", 2555),    // 7 years
			MaxEventSize:       getEnvInt("MAX_EVENT_SIZE", 1048576), // 1MB
			BatchSize:          getEnvInt("BATCH_SIZE", 100),
			FlushInterval:      getEnvInt("FLUSH_INTERVAL", 1000), // 1 second
			ReplicationEnabled: getEnvBool("REPLICATION_ENABLED", false),
			ReplicationFactor:  getEnvInt("REPLICATION_FACTOR", 3),
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
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be greater than 0")
	}

	return nil
}

