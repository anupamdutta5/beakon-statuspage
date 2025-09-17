// Package services provides comprehensive monitoring management and orchestration.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/anupamdutta5/statuspage-monitoring-service/internal/models"
	"go.uber.org/zap"
)

// MonitoringManager orchestrates all monitoring services and provides unified monitoring capabilities.
type MonitoringManager struct {
	monitoringService *MonitoringService
	logger            *zap.Logger
}

// NewMonitoringManager creates a new monitoring manager.
func NewMonitoringManager(monitoringService *MonitoringService, logger *zap.Logger) *MonitoringManager {
	return &MonitoringManager{
		monitoringService: monitoringService,
		logger:            logger,
	}
}

// ComprehensiveMonitoringOverview provides a complete overview of all monitoring systems.
func (m *MonitoringManager) ComprehensiveMonitoringOverview(tenantID uint) (map[string]interface{}, error) {
	overview := make(map[string]interface{})

	// Get core monitoring overview
	coreOverview, err := m.monitoringService.GetMonitoringOverview(tenantID)
	if err != nil {
		m.logger.Error("Failed to get core monitoring overview", zap.Error(err))
		return nil, fmt.Errorf("failed to get core monitoring overview: %w", err)
	}
	overview["core_monitoring"] = coreOverview

	// Get component monitoring overview
	componentOverview, err := m.getComponentMonitoringOverview(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get component monitoring overview", zap.Error(err))
		componentOverview = map[string]interface{}{"error": err.Error()}
	}
	overview["component_monitoring"] = componentOverview

	// Get external monitoring overview
	externalOverview, err := m.getExternalMonitoringOverview(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get external monitoring overview", zap.Error(err))
		externalOverview = map[string]interface{}{"error": err.Error()}
	}
	overview["external_monitoring"] = externalOverview

	// Get custom metrics overview
	metricsOverview, err := m.getCustomMetricsOverview(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get custom metrics overview", zap.Error(err))
		metricsOverview = map[string]interface{}{"error": err.Error()}
	}
	overview["custom_metrics"] = metricsOverview

	// Get Kubernetes monitoring overview
	k8sOverview, err := m.getKubernetesMonitoringOverview(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get Kubernetes monitoring overview", zap.Error(err))
		k8sOverview = map[string]interface{}{"error": err.Error()}
	}
	overview["kubernetes_monitoring"] = k8sOverview

	// Get Docker monitoring overview
	dockerOverview, err := m.getDockerMonitoringOverview(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get Docker monitoring overview", zap.Error(err))
		dockerOverview = map[string]interface{}{"error": err.Error()}
	}
	overview["docker_monitoring"] = dockerOverview

	// Get status automation overview
	automationOverview, err := m.getStatusAutomationOverview(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get status automation overview", zap.Error(err))
		automationOverview = map[string]interface{}{"error": err.Error()}
	}
	overview["status_automation"] = automationOverview

	overview["timestamp"] = time.Now().UTC()
	overview["tenant_id"] = tenantID

	return overview, nil
}

// PerformComprehensiveHealthCheck performs health checks across all monitoring systems.
func (m *MonitoringManager) PerformComprehensiveHealthCheck(tenantID uint) (map[string]interface{}, error) {
	healthCheck := make(map[string]interface{})

	// Check core monitoring health
	coreHealth, err := m.monitoringService.GetPublicHealth(tenantID)
	if err != nil {
		m.logger.Error("Failed to get core monitoring health", zap.Error(err))
		coreHealth = map[string]interface{}{"error": err.Error()}
	}
	healthCheck["core_monitoring"] = coreHealth

	// Check component monitoring health
	componentHealth, err := m.getComponentHealthSummary(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get component health summary", zap.Error(err))
		componentHealth = map[string]interface{}{"error": err.Error()}
	}
	healthCheck["component_monitoring"] = componentHealth

	// Check external monitoring health
	externalHealth, err := m.getExternalHealthSummary(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get external health summary", zap.Error(err))
		externalHealth = map[string]interface{}{"error": err.Error()}
	}
	healthCheck["external_monitoring"] = externalHealth

	// Check Kubernetes monitoring health
	k8sHealth, err := m.getKubernetesHealthSummary(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get Kubernetes health summary", zap.Error(err))
		k8sHealth = map[string]interface{}{"error": err.Error()}
	}
	healthCheck["kubernetes_monitoring"] = k8sHealth

	// Check Docker monitoring health
	dockerHealth, err := m.getDockerHealthSummary(tenantID)
	if err != nil {
		m.logger.Warn("Failed to get Docker health summary", zap.Error(err))
		dockerHealth = map[string]interface{}{"error": err.Error()}
	}
	healthCheck["docker_monitoring"] = dockerHealth

	healthCheck["timestamp"] = time.Now().UTC()
	healthCheck["tenant_id"] = tenantID

	return healthCheck, nil
}

// PerformBulkHealthChecks performs health checks for multiple services concurrently.
func (m *MonitoringManager) PerformBulkHealthChecks(ctx context.Context, requests []*HealthCheckRequest) ([]*HealthCheckResponse, error) {
	return m.monitoringService.HealthCheck.PerformBulkHealthCheck(ctx, requests)
}

// CreateMonitoringDashboard creates a comprehensive monitoring dashboard configuration.
func (m *MonitoringManager) CreateMonitoringDashboard(tenantID uint, config map[string]interface{}) (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// Get all monitoring data
	overview, err := m.ComprehensiveMonitoringOverview(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitoring overview: %w", err)
	}

	// Get health status
	health, err := m.PerformComprehensiveHealthCheck(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health status: %w", err)
	}

	// Get recent alerts
	alerts, _, err := m.monitoringService.GetAlerts(tenantID, 10, 0)
	if err != nil {
		m.logger.Warn("Failed to get recent alerts", zap.Error(err))
		alerts = []*models.Alert{}
	}

	// Get active services
	services, _, err := m.monitoringService.GetServices(tenantID, 20, 0)
	if err != nil {
		m.logger.Warn("Failed to get active services", zap.Error(err))
		services = []*models.MonitoredService{}
	}

	dashboard["overview"] = overview
	dashboard["health"] = health
	dashboard["recent_alerts"] = alerts
	dashboard["active_services"] = services
	dashboard["config"] = config
	dashboard["timestamp"] = time.Now().UTC()
	dashboard["tenant_id"] = tenantID

	return dashboard, nil
}

// Helper methods for getting overviews

func (m *MonitoringManager) getComponentMonitoringOverview(tenantID uint) (map[string]interface{}, error) {
	components, total, err := m.monitoringService.ComponentMonitoring.GetComponents(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	healthyComponents := 0
	for _, component := range components {
		if component.Status == "operational" {
			healthyComponents++
		}
	}

	return map[string]interface{}{
		"total_components":     total,
		"healthy_components":   healthyComponents,
		"unhealthy_components": total - int64(healthyComponents),
		"components":           components,
	}, nil
}

func (m *MonitoringManager) getExternalMonitoringOverview(tenantID uint) (map[string]interface{}, error) {
	services, total, err := m.monitoringService.ExternalMonitoring.GetExternalServices(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	healthyServices := 0
	for _, service := range services {
		if service.Status == "operational" {
			healthyServices++
		}
	}

	return map[string]interface{}{
		"total_services":     total,
		"healthy_services":   healthyServices,
		"unhealthy_services": total - int64(healthyServices),
		"services":           services,
	}, nil
}

func (m *MonitoringManager) getCustomMetricsOverview(tenantID uint) (map[string]interface{}, error) {
	metrics, total, err := m.monitoringService.CustomMetrics.GetCustomMetrics(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	activeMetrics := 0
	for _, metric := range metrics {
		if metric.IsActive {
			activeMetrics++
		}
	}

	return map[string]interface{}{
		"total_metrics":    total,
		"active_metrics":   activeMetrics,
		"inactive_metrics": total - int64(activeMetrics),
		"metrics":          metrics,
	}, nil
}

func (m *MonitoringManager) getKubernetesMonitoringOverview(tenantID uint) (map[string]interface{}, error) {
	resources, total, err := m.monitoringService.KubernetesMonitoring.GetKubernetesResources(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	healthyResources := 0
	for _, resource := range resources {
		if resource.Status == "running" {
			healthyResources++
		}
	}

	return map[string]interface{}{
		"total_resources":     total,
		"healthy_resources":   healthyResources,
		"unhealthy_resources": total - int64(healthyResources),
		"resources":           resources,
	}, nil
}

func (m *MonitoringManager) getDockerMonitoringOverview(tenantID uint) (map[string]interface{}, error) {
	containers, total, err := m.monitoringService.DockerMonitoring.GetDockerContainers(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	runningContainers := 0
	for _, container := range containers {
		if container.Status == "running" {
			runningContainers++
		}
	}

	return map[string]interface{}{
		"total_containers":   total,
		"running_containers": runningContainers,
		"stopped_containers": total - int64(runningContainers),
		"containers":         containers,
	}, nil
}

func (m *MonitoringManager) getStatusAutomationOverview(tenantID uint) (map[string]interface{}, error) {
	automations, total, err := m.monitoringService.StatusAutomation.GetStatusAutomations(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	activeAutomations := 0
	for _, automation := range automations {
		if automation.IsActive {
			activeAutomations++
		}
	}

	return map[string]interface{}{
		"total_automations":    total,
		"active_automations":   activeAutomations,
		"inactive_automations": total - int64(activeAutomations),
		"automations":          automations,
	}, nil
}

// Helper methods for getting health summaries

func (m *MonitoringManager) getComponentHealthSummary(tenantID uint) (map[string]interface{}, error) {
	components, _, err := m.monitoringService.ComponentMonitoring.GetComponents(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	healthSummary := make(map[string]interface{})
	healthSummary["total_components"] = len(components)
	healthSummary["component_health"] = make([]map[string]interface{}, 0)

	for _, component := range components {
		health, err := m.monitoringService.ComponentMonitoring.GetComponentHealth(component.ID)
		if err != nil {
			m.logger.Warn("Failed to get component health", zap.Uint("component_id", component.ID), zap.Error(err))
			continue
		}
		healthSummary["component_health"] = append(healthSummary["component_health"].([]map[string]interface{}), health)
	}

	return healthSummary, nil
}

func (m *MonitoringManager) getExternalHealthSummary(tenantID uint) (map[string]interface{}, error) {
	services, _, err := m.monitoringService.ExternalMonitoring.GetExternalServices(tenantID, 100, 0)
	if err != nil {
		return nil, err
	}

	healthSummary := make(map[string]interface{})
	healthSummary["total_services"] = len(services)
	healthSummary["service_health"] = make([]map[string]interface{}, 0)

	for _, service := range services {
		health, err := m.monitoringService.ExternalMonitoring.GetExternalServiceHealth(service.ID)
		if err != nil {
			m.logger.Warn("Failed to get external service health", zap.Uint("service_id", service.ID), zap.Error(err))
			continue
		}
		healthSummary["service_health"] = append(healthSummary["service_health"].([]map[string]interface{}), health)
	}

	return healthSummary, nil
}

func (m *MonitoringManager) getKubernetesHealthSummary(tenantID uint) (map[string]interface{}, error) {
	clusterHealth, err := m.monitoringService.KubernetesMonitoring.GetKubernetesClusterHealth(tenantID)
	if err != nil {
		return nil, err
	}

	return clusterHealth, nil
}

func (m *MonitoringManager) getDockerHealthSummary(tenantID uint) (map[string]interface{}, error) {
	hostHealth, err := m.monitoringService.DockerMonitoring.GetDockerHostHealth(tenantID)
	if err != nil {
		return nil, err
	}

	return hostHealth, nil
}
