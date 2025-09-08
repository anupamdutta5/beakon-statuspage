package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Service represents a discovered service
type Service struct {
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Health   string            `json:"health"`
	Tags     []string          `json:"tags"`
	Meta     map[string]string `json:"meta"`
	LastSeen time.Time         `json:"last_seen"`
}

// ServiceDiscovery handles service discovery and registration
type ServiceDiscovery struct {
	config     *config.Config
	services   map[string][]*Service
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	discoverer ServiceDiscoverer
}

// ServiceDiscoverer interface for different discovery backends
type ServiceDiscoverer interface {
	Start(ctx context.Context) error
	Stop() error
	Register(service *Service) error
	Deregister(serviceName string) error
	Discover(serviceName string) ([]*Service, error)
	Watch(serviceName string, callback func([]*Service)) error
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(cfg *config.Config) *ServiceDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	// Choose discovery backend based on configuration
	var discoverer ServiceDiscoverer
	switch cfg.ServiceDiscovery.Provider {
	case "consul":
		discoverer = NewStaticDiscoverer(cfg) // TODO: Implement Consul discoverer
	case "etcd":
		discoverer = NewStaticDiscoverer(cfg) // TODO: Implement etcd discoverer
	default:
		// Default to static discovery for development
		discoverer = NewStaticDiscoverer(cfg)
	}

	return &ServiceDiscovery{
		config:     cfg,
		services:   make(map[string][]*Service),
		ctx:        ctx,
		cancel:     cancel,
		discoverer: discoverer,
	}
}

// Start begins the service discovery process
func (sd *ServiceDiscovery) Start() {
	logger.Log.Info("Starting service discovery",
		zap.String("provider", sd.config.ServiceDiscovery.Provider))

	// Start the discovery backend
	if err := sd.discoverer.Start(sd.ctx); err != nil {
		logger.Log.Error("Failed to start service discovery", zap.Error(err))
		return
	}

	// Register default services
	sd.registerDefaultServices()

	// Start health checking
	go sd.startHealthChecking()

	// Start service watching
	go sd.startServiceWatching()
}

// Stop stops the service discovery
func (sd *ServiceDiscovery) Stop() {
	logger.Log.Info("Stopping service discovery")
	sd.cancel()
	if err := sd.discoverer.Stop(); err != nil {
		logger.Log.Error("Failed to stop discoverer", zap.Error(err))
	}
}

// GetService returns available instances of a service
func (sd *ServiceDiscovery) GetService(serviceName string) ([]*Service, error) {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	services, exists := sd.services[serviceName]
	if !exists || len(services) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", serviceName)
	}

	// Filter healthy services
	var healthyServices []*Service
	for _, service := range services {
		if service.Health == "healthy" {
			healthyServices = append(healthyServices, service)
		}
	}

	if len(healthyServices) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	return healthyServices, nil
}

// GetServiceURL returns a URL for a service instance
func (sd *ServiceDiscovery) GetServiceURL(serviceName string) (string, error) {
	services, err := sd.GetService(serviceName)
	if err != nil {
		return "", err
	}

	// Use the first healthy service (could implement load balancing here)
	service := services[0]
	return fmt.Sprintf("http://%s:%d", service.Address, service.Port), nil
}

// RegisterService registers a service with the discovery system
func (sd *ServiceDiscovery) RegisterService(service *Service) error {
	return sd.discoverer.Register(service)
}

// DeregisterService removes a service from the discovery system
func (sd *ServiceDiscovery) DeregisterService(serviceName string) error {
	return sd.discoverer.Deregister(serviceName)
}

// registerDefaultServices registers the default microservices
func (sd *ServiceDiscovery) registerDefaultServices() {
	defaultServices := []*Service{
		{
			Name:    "user-service",
			Address: "user-service",
			Port:    8081,
			Health:  "healthy",
			Tags:    []string{"api", "user"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "tenant-service",
			Address: "tenant-service",
			Port:    8082,
			Health:  "healthy",
			Tags:    []string{"api", "tenant"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "component-service",
			Address: "component-service",
			Port:    8083,
			Health:  "healthy",
			Tags:    []string{"api", "component"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "incident-service",
			Address: "incident-service",
			Port:    8084,
			Health:  "healthy",
			Tags:    []string{"api", "incident"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "notification-service",
			Address: "notification-service",
			Port:    8085,
			Health:  "healthy",
			Tags:    []string{"api", "notification"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "payment-service",
			Address: "payment-service",
			Port:    8086,
			Health:  "healthy",
			Tags:    []string{"api", "payment"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "analytics-service",
			Address: "analytics-service",
			Port:    8087,
			Health:  "healthy",
			Tags:    []string{"api", "analytics"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
		{
			Name:    "monitoring-service",
			Address: "monitoring-service",
			Port:    8088,
			Health:  "healthy",
			Tags:    []string{"api", "monitoring"},
			Meta: map[string]string{
				"version": "1.0.0",
				"env":     sd.config.Environment,
			},
			LastSeen: time.Now(),
		},
	}

	for _, service := range defaultServices {
		if err := sd.RegisterService(service); err != nil {
			logger.Log.Error("Failed to register service",
				zap.String("service", service.Name),
				zap.Error(err))
		} else {
			logger.Log.Info("Registered service", zap.String("service", service.Name))
		}
	}
}

// startHealthChecking periodically checks service health
func (sd *ServiceDiscovery) startHealthChecking() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sd.ctx.Done():
			return
		case <-ticker.C:
			sd.checkServiceHealth()
		}
	}
}

// checkServiceHealth checks the health of all registered services
func (sd *ServiceDiscovery) checkServiceHealth() {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	for serviceName, services := range sd.services {
		for _, service := range services {
			// Check if service is still healthy
			if time.Since(service.LastSeen) > 2*time.Minute {
				service.Health = "unhealthy"
				logger.Log.Warn("Service marked as unhealthy",
					zap.String("service", serviceName),
					zap.String("address", service.Address))
			}
		}
	}
}

// startServiceWatching watches for service changes
func (sd *ServiceDiscovery) startServiceWatching() {
	serviceNames := []string{
		"user-service", "tenant-service", "component-service",
		"incident-service", "notification-service", "payment-service",
		"analytics-service", "monitoring-service",
	}

	for _, serviceName := range serviceNames {
		go func(name string) {
			if err := sd.discoverer.Watch(name, func(services []*Service) {
				sd.mu.Lock()
				sd.services[name] = services
				sd.mu.Unlock()

				logger.Log.Info("Service instances updated",
					zap.String("service", name),
					zap.Int("count", len(services)))
			}); err != nil {
				logger.Log.Error("Failed to watch service",
					zap.String("service", name),
					zap.Error(err))
			}
		}(serviceName)
	}
}
