// Package config provides service configuration for the SaaS Admin Service.
package config

import "os"

// ServiceURLs contains URLs for downstream services that SaaS Admin proxies to.
type ServiceURLs struct {
	TenantAdminService string
	ComponentService   string
	IncidentService    string
	MonitoringService  string
	AnalyticsService   string
	NotificationService string
}

// GetDefaultServiceURLs returns default service URLs for development.
func GetDefaultServiceURLs() ServiceURLs {
	return ServiceURLs{
		TenantAdminService:  "http://127.0.0.1:8093",
		ComponentService:    "http://127.0.0.1:8093",
		IncidentService:     "http://127.0.0.1:8094",
		MonitoringService:   "http://127.0.0.1:8095",
		AnalyticsService:    "http://127.0.0.1:8096",
		NotificationService: "http://127.0.0.1:8097",
	}
}

// GetServiceURLsFromEnv returns service URLs from environment variables.
func GetServiceURLsFromEnv() ServiceURLs {
	return ServiceURLs{
		TenantAdminService:  getEnvOrDefault("TENANT_ADMIN_SERVICE_URL", "http://localhost:8099"),
		ComponentService:    getEnvOrDefault("COMPONENT_SERVICE_URL", "http://localhost:8093"),
		IncidentService:     getEnvOrDefault("INCIDENT_SERVICE_URL", "http://localhost:8094"),
		MonitoringService:   getEnvOrDefault("MONITORING_SERVICE_URL", "http://localhost:8095"),
		AnalyticsService:    getEnvOrDefault("ANALYTICS_SERVICE_URL", "http://localhost:8096"),
		NotificationService: getEnvOrDefault("NOTIFICATION_SERVICE_URL", "http://localhost:8097"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
