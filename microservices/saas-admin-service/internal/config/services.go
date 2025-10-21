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
		TenantAdminService:  "http://127.0.0.1:8099",
		ComponentService:    "http://127.0.0.1:8084",
		IncidentService:     "http://127.0.0.1:8086",
		MonitoringService:   "http://127.0.0.1:8092",
		AnalyticsService:    "http://127.0.0.1:8090",
		NotificationService: "http://127.0.0.1:8085",
	}
}

// GetServiceURLsFromEnv returns service URLs from environment variables.
// All services should route through API Gateway unless overridden
func GetServiceURLsFromEnv() ServiceURLs {
	defaults := GetDefaultServiceURLs()

	return ServiceURLs{
		TenantAdminService:  getEnvOrDefault("TENANT_ADMIN_SERVICE_URL", defaults.TenantAdminService),
		ComponentService:    getEnvOrDefault("COMPONENT_SERVICE_URL", defaults.ComponentService),
		IncidentService:     getEnvOrDefault("INCIDENT_SERVICE_URL", defaults.IncidentService),
		MonitoringService:   getEnvOrDefault("MONITORING_SERVICE_URL", defaults.MonitoringService),
		AnalyticsService:    getEnvOrDefault("ANALYTICS_SERVICE_URL", defaults.AnalyticsService),
		NotificationService: getEnvOrDefault("NOTIFICATION_SERVICE_URL", defaults.NotificationService),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
