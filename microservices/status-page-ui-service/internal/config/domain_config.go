package config

import (
	"strings"
	"time"
)

// DomainConfig represents the domain service configuration
type DomainConfig struct {
	// RedisURL is the connection string for Redis
	RedisURL string `yaml:"redis_url" envconfig:"REDIS_URL"`
	
	// CacheTTL is the time-to-live for domain cache entries
	CacheTTL time.Duration `yaml:"cache_ttl" envconfig:"CACHE_TTL" default:"5m"`
	
	// DefaultTenantID is the fallback tenant ID when domain resolution fails
	DefaultTenantID uint `yaml:"default_tenant_id" envconfig:"DEFAULT_TENANT_ID"`
	
	// DefaultStatusPage is the fallback status page slug when domain resolution fails
	DefaultStatusPage string `yaml:"default_status_page" envconfig:"DEFAULT_STATUS_PAGE" default:"default"`
	
	// VerificationTTL is the time-to-live for domain verification tokens
	VerificationTTL time.Duration `yaml:"verification_ttl" envconfig:"VERIFICATION_TTL" default:"24h"`
	
	// DNSResolvers is a list of DNS servers to use for verification
	DNSResolvers []string `yaml:"dns_resolvers" envconfig:"DNS_RESOLVERS" default:"8.8.8.8:53,1.1.1.1:53"`
	
	// DNSTimeout is the timeout for DNS lookups
	DNSTimeout time.Duration `yaml:"dns_timeout" envconfig:"DNS_TIMEOUT" default:"5s"`
}

// LoadDomainConfig loads the domain service configuration from environment variables
func LoadDomainConfig() (*DomainConfig, error) {
	config := &DomainConfig{
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		CacheTTL:          getEnvDuration("CACHE_TTL", 5*time.Minute),
		DefaultTenantID:   uint(getEnvInt("DEFAULT_TENANT_ID", 0)),
		DefaultStatusPage: getEnv("DEFAULT_STATUS_PAGE", "default"),
		VerificationTTL:   getEnvDuration("VERIFICATION_TTL", 24*time.Hour),
		DNSResolvers:      getEnvSlice("DNS_RESOLVERS", []string{"8.8.8.8:53", "1.1.1.1:53"}),
		DNSTimeout:        getEnvDuration("DNS_TIMEOUT", 5*time.Second),
	}

	return config, nil
}

// getEnvDuration gets an environment variable as a time.Duration with a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	val := getEnv(key, "")
	if val == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return defaultValue
	}

	return duration
}

// getEnvSlice gets an environment variable as a string slice with a default value
func getEnvSlice(key string, defaultValue []string) []string {
	val := getEnv(key, "")
	if val == "" {
		return defaultValue
	}

	return strings.Split(val, ",")
}
