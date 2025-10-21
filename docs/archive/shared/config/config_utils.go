// Package config provides shared configuration utilities for Beakon microservices.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvConfig provides utilities for loading configuration from environment variables
type EnvConfig struct {
	prefix string
}

// NewEnvConfig creates a new environment configuration loader with optional prefix
func NewEnvConfig(prefix string) *EnvConfig {
	return &EnvConfig{prefix: prefix}
}

// GetString gets a string environment variable with a default value
func (e *EnvConfig) GetString(key, defaultValue string) string {
	return GetEnvString(e.prefixKey(key), defaultValue)
}

// GetInt gets an integer environment variable with a default value
func (e *EnvConfig) GetInt(key string, defaultValue int) int {
	return GetEnvInt(e.prefixKey(key), defaultValue)
}

// GetBool gets a boolean environment variable with a default value
func (e *EnvConfig) GetBool(key string, defaultValue bool) bool {
	return GetEnvBool(e.prefixKey(key), defaultValue)
}

// GetDuration gets a duration environment variable with a default value
func (e *EnvConfig) GetDuration(key string, defaultValue time.Duration) time.Duration {
	return GetEnvDuration(e.prefixKey(key), defaultValue)
}

// GetStringSlice gets a comma-separated string slice environment variable with a default value
func (e *EnvConfig) GetStringSlice(key string, defaultValue []string) []string {
	return GetEnvStringSlice(e.prefixKey(key), defaultValue)
}

// prefixKey adds the prefix to a key if prefix is set
func (e *EnvConfig) prefixKey(key string) string {
	if e.prefix == "" {
		return key
	}
	return e.prefix + "_" + key
}

// GetEnvString gets a string environment variable with a default value
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt gets an integer environment variable with a default value
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool gets a boolean environment variable with a default value
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvDuration gets a duration environment variable with a default value
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetEnvStringSlice gets a comma-separated string slice environment variable with a default value
func GetEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// RequiredEnvString gets a required string environment variable or panics
func RequiredEnvString(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("Required environment variable %s is not set", key))
}

// RequiredEnvInt gets a required integer environment variable or panics
func RequiredEnvInt(key string) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		panic(fmt.Sprintf("Environment variable %s is not a valid integer: %s", key, value))
	}
	panic(fmt.Sprintf("Required environment variable %s is not set", key))
}

// ServiceConfig represents common service configuration
type ServiceConfig struct {
	Name        string
	Version     string
	Environment string
	Host        string
	Port        int
	LogLevel    string
	LogFormat   string
}

// LoadServiceConfig loads common service configuration from environment
func LoadServiceConfig(serviceName string) ServiceConfig {
	env := NewEnvConfig("")
	return ServiceConfig{
		Name:        env.GetString("SERVICE_NAME", serviceName),
		Version:     env.GetString("SERVICE_VERSION", "1.0.0"),
		Environment: env.GetString("ENVIRONMENT", "development"),
		Host:        env.GetString("HOST", "0.0.0.0"),
		Port:        env.GetInt("PORT", 8080),
		LogLevel:    env.GetString("LOG_LEVEL", "info"),
		LogFormat:   env.GetString("LOG_FORMAT", "json"),
	}
}

// DatabaseConfig represents common database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// LoadDatabaseConfig loads database configuration from environment
func LoadDatabaseConfig(databaseName string) DatabaseConfig {
	env := NewEnvConfig("DB")
	return DatabaseConfig{
		Host:            env.GetString("HOST", "localhost"),
		Port:            env.GetInt("PORT", 5432),
		User:            env.GetString("USER", "postgres"),
		Password:        env.GetString("PASSWORD", "postgres"),
		Name:            GetEnvString("DB_NAME", databaseName),
		SSLMode:         env.GetString("SSL_MODE", "disable"),
		MaxOpenConns:    env.GetInt("MAX_OPEN_CONNS", 25),
		MaxIdleConns:    env.GetInt("MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: env.GetDuration("CONN_MAX_LIFETIME", time.Hour),
		ConnMaxIdleTime: env.GetDuration("CONN_MAX_IDLE_TIME", time.Minute*30),
	}
}

// GetDSN returns the database DSN string
func (d DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// LoadRedisConfig loads Redis configuration from environment
func LoadRedisConfig() RedisConfig {
	env := NewEnvConfig("REDIS")
	return RedisConfig{
		Host:     env.GetString("HOST", "localhost"),
		Port:     env.GetInt("PORT", 6379),
		Password: env.GetString("PASSWORD", ""),
		DB:       env.GetInt("DB", 0),
		PoolSize: env.GetInt("POOL_SIZE", 10),
	}
}

// GetAddr returns the Redis address string
func (r RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret            string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
	Issuer            string
	Audience          string
}

// LoadJWTConfig loads JWT configuration from environment
func LoadJWTConfig() JWTConfig {
	env := NewEnvConfig("JWT")
	return JWTConfig{
		Secret:            RequiredEnvString("JWT_SECRET"),
		AccessExpiration:  env.GetDuration("ACCESS_EXPIRATION", time.Hour*24),
		RefreshExpiration: env.GetDuration("REFRESH_EXPIRATION", time.Hour*24*7),
		Issuer:            env.GetString("ISSUER", "beakon-statuspage"),
		Audience:          env.GetString("AUDIENCE", "beakon-users"),
	}
}

// Validate validates the configuration
func (c ServiceConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	return nil
}

// Validate validates the database configuration
func (d DatabaseConfig) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if d.Port <= 0 || d.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", d.Port)
	}
	if d.User == "" {
		return fmt.Errorf("database user is required")
	}
	if d.Name == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

// Validate validates the JWT configuration
func (j JWTConfig) Validate() error {
	if j.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(j.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}
	if j.AccessExpiration <= 0 {
		return fmt.Errorf("JWT access expiration must be positive")
	}
	if j.RefreshExpiration <= 0 {
		return fmt.Errorf("JWT refresh expiration must be positive")
	}
	return nil
}