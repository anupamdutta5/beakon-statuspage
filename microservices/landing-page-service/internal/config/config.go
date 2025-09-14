// Package config provides configuration management for the Landing Page Service.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the Landing Page Service configuration.
type Config struct {
	Environment string          `yaml:"environment"`
	Service     ServiceConfig   `yaml:"service"`
	Server      ServerConfig    `yaml:"server"`
	Database    DatabaseConfig  `yaml:"database"`
	Cache       CacheConfig     `yaml:"cache"`
	Templates   TemplatesConfig `yaml:"templates"`
	Static      StaticConfig    `yaml:"static"`
	Services    ServicesConfig  `yaml:"services"`
	Features    FeaturesConfig  `yaml:"features"`
	Landing     LandingConfig   `yaml:"landing"`
	Logging     LoggingConfig   `yaml:"logging"`
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

// LandingConfig represents landing page-specific configuration.
type LandingConfig struct {
	SiteName        string   `yaml:"site_name"`
	SiteURL         string   `yaml:"site_url"`
	SiteDescription string   `yaml:"site_description"`
	SiteKeywords    []string `yaml:"site_keywords"`
	ContactEmail    string   `yaml:"contact_email"`
	SupportEmail    string   `yaml:"support_email"`
	SocialLinks     string   `yaml:"social_links"` // JSON object
	AnalyticsID     string   `yaml:"analytics_id"`
	CDNEnabled      bool     `yaml:"cdn_enabled"`
	CDNURL          string   `yaml:"cdn_url"`
	CacheEnabled    bool     `yaml:"cache_enabled"`
	CacheTTL        int      `yaml:"cache_ttl"`
	SEOEnabled      bool     `yaml:"seo_enabled"`
	OGImage         string   `yaml:"og_image"`
	Favicon         string   `yaml:"favicon"`
	Theme           string   `yaml:"theme"`
	CustomCSS       string   `yaml:"custom_css"`
	CustomJS        string   `yaml:"custom_js"`
}

// CacheConfig represents cache configuration.
type CacheConfig struct {
	Provider string `yaml:"provider"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TTL      int    `yaml:"ttl"`
}

// TemplatesConfig represents templates configuration.
type TemplatesConfig struct {
	Path  string `yaml:"path"`
	Cache bool   `yaml:"cache"`
}

// StaticConfig represents static files configuration.
type StaticConfig struct {
	Path         string `yaml:"path"`
	CacheControl string `yaml:"cache_control"`
}

// ServicesConfig represents external services configuration.
type ServicesConfig struct {
	DatabaseService   ServiceEndpoint `yaml:"database_service"`
	EventStoreService ServiceEndpoint `yaml:"event_store_service"`
	SaaSAdminService  ServiceEndpoint `yaml:"saas_admin_service"`
}

// ServiceEndpoint represents a service endpoint configuration.
type ServiceEndpoint struct {
	BaseURL string `yaml:"base_url"`
	Timeout string `yaml:"timeout"`
}

// FeaturesConfig represents feature flags configuration.
type FeaturesConfig struct {
	EnableAnalytics   bool `yaml:"enable_analytics"`
	EnableContactForm bool `yaml:"enable_contact_form"`
	EnableNewsletter  bool `yaml:"enable_newsletter"`
	EnableBlog        bool `yaml:"enable_blog"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Service: ServiceConfig{
			Name:        getEnv("SERVICE_NAME", "landing-page-service"),
			Version:     getEnv("SERVICE_VERSION", "1.0.0"),
			Description: getEnv("SERVICE_DESCRIPTION", "Beautiful Landing Page Service with Content Management"),
			Tags:        getEnvSlice("SERVICE_TAGS", []string{"landing-page-service", "microservice", "frontend"}),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8100),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnvInt("DB_PORT", 5432),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", "postgres"),
			Name:        getEnv("DB_NAME", "statuspage_landing"),
			SSLMode:     getEnv("DB_SSL_MODE", "disable"),
			MaxConns:    getEnvInt("DB_MAX_CONNS", 100),
			MinConns:    getEnvInt("DB_MIN_CONNS", 10),
			MaxIdle:     getEnvInt("DB_MAX_IDLE", 10),
			MaxLifetime: getEnvInt("DB_MAX_LIFETIME", 3600),
		},
		Cache: CacheConfig{
			Provider: getEnv("CACHE_PROVIDER", "redis"),
			Host:     getEnv("CACHE_HOST", "localhost"),
			Port:     getEnvInt("CACHE_PORT", 6379),
			Password: getEnv("CACHE_PASSWORD", ""),
			DB:       getEnvInt("CACHE_DB", 0),
			TTL:      getEnvInt("CACHE_TTL", 300),
		},
		Templates: TemplatesConfig{
			Path:  getEnv("TEMPLATES_PATH", "web/templates"),
			Cache: getEnvBool("TEMPLATES_CACHE", false),
		},
		Static: StaticConfig{
			Path:         getEnv("STATIC_PATH", "web/static"),
			CacheControl: getEnv("STATIC_CACHE_CONTROL", "public, max-age=3600"),
		},
		Services: ServicesConfig{
			DatabaseService: ServiceEndpoint{
				BaseURL: getEnv("DATABASE_SERVICE_URL", "http://localhost:8090"),
				Timeout: getEnv("DATABASE_SERVICE_TIMEOUT", "5s"),
			},
			EventStoreService: ServiceEndpoint{
				BaseURL: getEnv("EVENT_STORE_SERVICE_URL", "http://localhost:8091"),
				Timeout: getEnv("EVENT_STORE_SERVICE_TIMEOUT", "5s"),
			},
			SaaSAdminService: ServiceEndpoint{
				BaseURL: getEnv("SAAS_ADMIN_SERVICE_URL", "http://localhost:8092"),
				Timeout: getEnv("SAAS_ADMIN_SERVICE_TIMEOUT", "5s"),
			},
		},
		Features: FeaturesConfig{
			EnableAnalytics:   getEnvBool("ENABLE_ANALYTICS", true),
			EnableContactForm: getEnvBool("ENABLE_CONTACT_FORM", true),
			EnableNewsletter:  getEnvBool("ENABLE_NEWSLETTER", true),
			EnableBlog:        getEnvBool("ENABLE_BLOG", true),
		},
		Landing: LandingConfig{
			SiteName:        getEnv("LANDING_SITE_NAME", "StatusPage Pro"),
			SiteURL:         getEnv("LANDING_SITE_URL", "https://statuspage.pro"),
			SiteDescription: getEnv("LANDING_SITE_DESCRIPTION", "Professional status page platform for modern teams"),
			SiteKeywords:    getEnvSlice("LANDING_SITE_KEYWORDS", []string{"status page", "uptime monitoring", "incident management", "team communication"}),
			ContactEmail:    getEnv("LANDING_CONTACT_EMAIL", "hello@statuspage.pro"),
			SupportEmail:    getEnv("LANDING_SUPPORT_EMAIL", "support@statuspage.pro"),
			SocialLinks:     getEnv("LANDING_SOCIAL_LINKS", "{}"),
			AnalyticsID:     getEnv("LANDING_ANALYTICS_ID", ""),
			CDNEnabled:      getEnvBool("LANDING_CDN_ENABLED", false),
			CDNURL:          getEnv("LANDING_CDN_URL", ""),
			CacheEnabled:    getEnvBool("LANDING_CACHE_ENABLED", true),
			CacheTTL:        getEnvInt("LANDING_CACHE_TTL", 3600),
			SEOEnabled:      getEnvBool("LANDING_SEO_ENABLED", true),
			OGImage:         getEnv("LANDING_OG_IMAGE", "/static/images/og-image.png"),
			Favicon:         getEnv("LANDING_FAVICON", "/static/images/favicon.ico"),
			Theme:           getEnv("LANDING_THEME", "modern"),
			CustomCSS:       getEnv("LANDING_CUSTOM_CSS", ""),
			CustomJS:        getEnv("LANDING_CUSTOM_JS", ""),
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

// loadConfigFromFile loads configuration from a YAML file.
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

	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	return nil
}
