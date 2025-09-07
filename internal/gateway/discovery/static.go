package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// StaticDiscoverer provides static service discovery for development
type StaticDiscoverer struct {
	config   *config.Config
	services map[string][]*Service
	watchers map[string][]func([]*Service)
	mu       sync.RWMutex
	ctx      context.Context
}

// NewStaticDiscoverer creates a new static service discoverer
func NewStaticDiscoverer(cfg *config.Config) *StaticDiscoverer {
	return &StaticDiscoverer{
		config:   cfg,
		services: make(map[string][]*Service),
		watchers: make(map[string][]func([]*Service)),
	}
}

// Start initializes the static discoverer
func (sd *StaticDiscoverer) Start(ctx context.Context) error {
	sd.ctx = ctx
	logger.Log.Info("Static service discovery started")
	return nil
}

// Stop stops the static discoverer
func (sd *StaticDiscoverer) Stop() error {
	logger.Log.Info("Static service discovery stopped")
	return nil
}

// Register registers a service with static discovery
func (sd *StaticDiscoverer) Register(service *Service) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	serviceName := service.Name
	services := sd.services[serviceName]

	// Check if service already exists
	for i, existingService := range services {
		if existingService.Address == service.Address && existingService.Port == service.Port {
			// Update existing service
			services[i] = service
			sd.services[serviceName] = services
			sd.notifyWatchers(serviceName, services)
			return nil
		}
	}

	// Add new service
	services = append(services, service)
	sd.services[serviceName] = services
	sd.notifyWatchers(serviceName, services)

	logger.Log.Info("Service registered with static discovery",
		zap.String("service", service.Name),
		zap.String("address", service.Address),
		zap.Int("port", service.Port))

	return nil
}

// Deregister removes a service from static discovery
func (sd *StaticDiscoverer) Deregister(serviceName string) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	delete(sd.services, serviceName)
	sd.notifyWatchers(serviceName, []*Service{})

	logger.Log.Info("Service deregistered from static discovery",
		zap.String("service", serviceName))

	return nil
}

// Discover returns services for a given service name
func (sd *StaticDiscoverer) Discover(serviceName string) ([]*Service, error) {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	services, exists := sd.services[serviceName]
	if !exists {
		return []*Service{}, nil
	}

	// Return a copy of the services
	result := make([]*Service, len(services))
	copy(result, services)
	return result, nil
}

// Watch sets up a watcher for service changes
func (sd *StaticDiscoverer) Watch(serviceName string, callback func([]*Service)) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	// Add callback to watchers
	sd.watchers[serviceName] = append(sd.watchers[serviceName], callback)

	// Immediately call callback with current services
	if services, exists := sd.services[serviceName]; exists {
		go callback(services)
	} else {
		go callback([]*Service{})
	}

	logger.Log.Info("Watcher added for service",
		zap.String("service", serviceName))

	return nil
}

// notifyWatchers notifies all watchers of a service about changes
func (sd *StaticDiscoverer) notifyWatchers(serviceName string, services []*Service) {
	watchers, exists := sd.watchers[serviceName]
	if !exists {
		return
	}

	for _, watcher := range watchers {
		go watcher(services)
	}
}

// GetServiceHealth checks the health of a service
func (sd *StaticDiscoverer) GetServiceHealth(service *Service) string {
	// For static discovery, we assume services are healthy
	// In a real implementation, this would make HTTP health checks
	return "healthy"
}

// UpdateServiceHealth updates the health status of a service
func (sd *StaticDiscoverer) UpdateServiceHealth(serviceName string, address string, port int, health string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	services, exists := sd.services[serviceName]
	if !exists {
		return
	}

	for _, service := range services {
		if service.Address == address && service.Port == port {
			service.Health = health
			service.LastSeen = time.Now()
			sd.notifyWatchers(serviceName, services)
			break
		}
	}
}
