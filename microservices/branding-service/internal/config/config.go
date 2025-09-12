// Package config provides configuration management for the Branding Service.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the Branding Service configuration.
type Config struct {
	Environment string         `json:"environment"`
	Service     ServiceConfig  `json:"service"`
	Server      ServerConfig   `json:"server"`
	Database    DatabaseConfig `json:"database"`
	Storage     StorageConfig  `json:"storage"`
	Branding    BrandingConfig `json:"branding"`
	Logging     LoggingConfig  `json:"logging"`
}

// ServiceConfig represents service-specific configuration.
type ServiceConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	SSLMode     string `json:"ssl_mode"`
	MaxConns    int    `json:"max_conns"`
	MinConns    int    `json:"min_conns"`
	MaxIdle     int    `json:"max_idle"`
	MaxLifetime int    `json:"max_lifetime"`
}

// StorageConfig represents file storage configuration.
type StorageConfig struct {
	Type         string   `json:"type"` // local, s3, gcs, azure
	LocalPath    string   `json:"local_path"`
	S3Bucket     string   `json:"s3_bucket"`
	S3Region     string   `json:"s3_region"`
	S3AccessKey  string   `json:"s3_access_key"`
	S3SecretKey  string   `json:"s3_secret_key"`
	GCSPath      string   `json:"gcs_path"`
	AzureAccount string   `json:"azure_account"`
	AzureKey     string   `json:"azure_key"`
	MaxFileSize  int64    `json:"max_file_size"` // in bytes
	AllowedTypes []string `json:"allowed_types"`
}

// BrandingConfig represents branding-specific configuration.
type BrandingConfig struct {
	DefaultTheme        string   `json:"default_theme"`
	AvailableThemes     []string `json:"available_themes"`
	CustomCSSEnabled    bool     `json:"custom_css_enabled"`
	CustomJSEnabled     bool     `json:"custom_js_enabled"`
	LogoMaxSize         int64    `json:"logo_max_size"`
	FaviconMaxSize      int64    `json:"favicon_max_size"`
	AllowedImageFormats []string `json:"allowed_image_formats"`
	AllowedFontFormats  []string `json:"allowed_font_formats"`
	CDNEnabled          bool     `json:"cdn_enabled"`
	CDNURL              string   `json:"cdn_url"`
	CacheEnabled        bool     `json:"cache_enabled"`
	CacheTTL            int      `json:"cache_ttl"`
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
			Name:        getEnv("SERVICE_NAME", "branding-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Professional Branding Service for Status Page"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"branding-service", "microservice", "design"}),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8097),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnvInt("DB_PORT", 5432),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			Name:        getEnv("DB_NAME", "statuspage_branding"),
			SSLMode:     getEnv("DB_SSL_MODE", "disable"),
			MaxConns:    getEnvInt("DB_MAX_CONNS", 100),
			MinConns:    getEnvInt("DB_MIN_CONNS", 10),
			MaxIdle:     getEnvInt("DB_MAX_IDLE", 10),
			MaxLifetime: getEnvInt("DB_MAX_LIFETIME", 3600),
		},
		Storage: StorageConfig{
			Type:         getEnv("STORAGE_TYPE", "local"),
			LocalPath:    getEnv("STORAGE_LOCAL_PATH", "./uploads"),
			S3Bucket:     getEnv("STORAGE_S3_BUCKET", ""),
			S3Region:     getEnv("STORAGE_S3_REGION", "us-east-1"),
			S3AccessKey:  getEnv("STORAGE_S3_ACCESS_KEY", ""),
			S3SecretKey:  getEnv("STORAGE_S3_SECRET_KEY", ""),
			GCSPath:      getEnv("STORAGE_GCS_PATH", ""),
			AzureAccount: getEnv("STORAGE_AZURE_ACCOUNT", ""),
			AzureKey:     getEnv("STORAGE_AZURE_KEY", ""),
			MaxFileSize:  getEnvInt64("STORAGE_MAX_FILE_SIZE", 10485760), // 10MB
			AllowedTypes: getEnvSlice("STORAGE_ALLOWED_TYPES", []string{"image/jpeg", "image/png", "image/gif", "image/svg+xml", "image/webp", "font/woff", "font/woff2", "text/css", "application/javascript"}),
		},
		Branding: BrandingConfig{
			DefaultTheme:        getEnv("BRANDING_DEFAULT_THEME", "modern"),
			AvailableThemes:     getEnvSlice("BRANDING_AVAILABLE_THEMES", []string{"modern", "classic", "minimal", "dark", "corporate"}),
			CustomCSSEnabled:    getEnvBool("BRANDING_CUSTOM_CSS_ENABLED", true),
			CustomJSEnabled:     getEnvBool("BRANDING_CUSTOM_JS_ENABLED", true),
			LogoMaxSize:         getEnvInt64("BRANDING_LOGO_MAX_SIZE", 2097152),    // 2MB
			FaviconMaxSize:      getEnvInt64("BRANDING_FAVICON_MAX_SIZE", 1048576), // 1MB
			AllowedImageFormats: getEnvSlice("BRANDING_ALLOWED_IMAGE_FORMATS", []string{"jpg", "jpeg", "png", "gif", "svg", "webp"}),
			AllowedFontFormats:  getEnvSlice("BRANDING_ALLOWED_FONT_FORMATS", []string{"woff", "woff2", "ttf", "otf"}),
			CDNEnabled:          getEnvBool("BRANDING_CDN_ENABLED", false),
			CDNURL:              getEnv("BRANDING_CDN_URL", ""),
			CacheEnabled:        getEnvBool("BRANDING_CACHE_ENABLED", true),
			CacheTTL:            getEnvInt("BRANDING_CACHE_TTL", 3600),
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

// getEnvInt64 gets an environment variable as an int64 with a default value.
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
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

	if c.Storage.MaxFileSize <= 0 {
		return fmt.Errorf("storage max file size must be greater than 0")
	}

	return nil
}

