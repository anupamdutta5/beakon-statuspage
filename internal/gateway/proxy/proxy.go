package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/gateway/discovery"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// ReverseProxy handles reverse proxy functionality
type ReverseProxy struct {
	config    *config.Config
	discovery *discovery.ServiceDiscovery
	client    *http.Client
}

// NewReverseProxy creates a new reverse proxy instance
func NewReverseProxy(cfg *config.Config, discovery *discovery.ServiceDiscovery) *ReverseProxy {
	return &ReverseProxy{
		config:    cfg,
		discovery: discovery,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ServeHTTP proxies requests to the appropriate microservice
func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request, serviceName, path string) {
	// Get service URL
	serviceURL, err := rp.discovery.GetServiceURL(serviceName)
	if err != nil {
		logger.Log.Error("Failed to get service URL",
			zap.String("service", serviceName),
			zap.Error(err))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Parse the target URL
	targetURL, err := url.Parse(serviceURL)
	if err != nil {
		logger.Log.Error("Failed to parse service URL",
			zap.String("service", serviceName),
			zap.String("url", serviceURL),
			zap.Error(err))
		http.Error(w, "Invalid service URL", http.StatusInternalServerError)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Customize the proxy director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Update the request path
		req.URL.Path = path
		req.URL.RawPath = path

		// Preserve original host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		if r.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		}

		// Add service name header
		req.Header.Set("X-Service-Name", serviceName)

		// Preserve request ID
		if requestID := r.Context().Value("request_id"); requestID != nil {
			req.Header.Set("X-Request-ID", requestID.(string))
		}

		// Preserve tenant context
		if tenantID := r.Context().Value("tenant_id"); tenantID != nil {
			req.Header.Set("X-Tenant-ID", tenantID.(string))
		}

		// Preserve user context
		if userID := r.Context().Value("user_id"); userID != nil {
			req.Header.Set("X-User-ID", userID.(string))
		}
	}

	// Customize error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Log.Error("Proxy error",
			zap.String("service", serviceName),
			zap.String("path", path),
			zap.Error(err))

		// Try to get another instance of the service
		if rp.tryFailover(w, r, serviceName, path) {
			return
		}

		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}

	// Add response modification
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Add CORS headers
		resp.Header.Set("Access-Control-Allow-Origin", "*")
		resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		// Add service name to response headers
		resp.Header.Set("X-Served-By", serviceName)

		// Preserve request ID
		if requestID := r.Context().Value("request_id"); requestID != nil {
			resp.Header.Set("X-Request-ID", requestID.(string))
		}

		return nil
	}

	// Start timing
	start := time.Now()

	// Serve the request
	proxy.ServeHTTP(w, r)

	// Log the request
	duration := time.Since(start)
	logger.Log.Info("Request proxied",
		zap.String("service", serviceName),
		zap.String("path", path),
		zap.String("method", r.Method),
		zap.Duration("duration", duration),
		zap.Int("status", w.(*responseWriter).statusCode))
}

// tryFailover attempts to failover to another service instance
func (rp *ReverseProxy) tryFailover(w http.ResponseWriter, r *http.Request, serviceName, path string) bool {
	services, err := rp.discovery.GetService(serviceName)
	if err != nil || len(services) <= 1 {
		return false
	}

	// Try the next available service
	for _, service := range services[1:] {
		serviceURL := fmt.Sprintf("http://%s:%d", service.Address, service.Port)
		if rp.tryService(w, r, serviceURL, path) {
			logger.Log.Info("Failover successful",
				zap.String("service", serviceName),
				zap.String("address", service.Address))
			return true
		}
	}

	return false
}

// tryService attempts to proxy to a specific service instance
func (rp *ReverseProxy) tryService(w http.ResponseWriter, r *http.Request, serviceURL, path string) bool {
	targetURL, err := url.Parse(serviceURL)
	if err != nil {
		return false
	}

	// Create a new request
	req, err := http.NewRequest(r.Method, targetURL.String()+path, r.Body)
	if err != nil {
		return false
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	resp, err := rp.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	return err == nil
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoadBalancer provides load balancing strategies
type LoadBalancer struct {
	strategy string
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy string) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
	}
}

// SelectService selects a service instance based on the load balancing strategy
func (lb *LoadBalancer) SelectService(services []*discovery.Service) *discovery.Service {
	if len(services) == 0 {
		return nil
	}

	switch lb.strategy {
	case "round-robin":
		return lb.roundRobin(services)
	case "random":
		return lb.random(services)
	case "least-connections":
		return lb.leastConnections(services)
	default:
		return services[0] // Default to first service
	}
}

// roundRobin implements round-robin load balancing
func (lb *LoadBalancer) roundRobin(services []*discovery.Service) *discovery.Service {
	// Simple round-robin implementation
	// In a real implementation, this would maintain state
	return services[0]
}

// random implements random load balancing
func (lb *LoadBalancer) random(services []*discovery.Service) *discovery.Service {
	// Simple random selection
	// In a real implementation, this would use proper random selection
	return services[0]
}

// leastConnections implements least connections load balancing
func (lb *LoadBalancer) leastConnections(services []*discovery.Service) *discovery.Service {
	// Simple least connections implementation
	// In a real implementation, this would track connection counts
	return services[0]
}

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker struct {
	failureThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
	failureCount     int
	state            string // "closed", "open", "half-open"
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		timeout:          timeout,
		state:            "closed",
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	if cb.state == "open" {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = "half-open"
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()
	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.failureThreshold {
			cb.state = "open"
		}

		return err
	}

	// Success - reset circuit breaker
	cb.failureCount = 0
	cb.state = "closed"
	return nil
}
