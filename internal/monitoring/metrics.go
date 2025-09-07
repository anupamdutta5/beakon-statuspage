package monitoring

import (
	"net/http"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric represents a single metric
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description,omitempty"`
}

// MetricsCollector handles all metrics collection for the microservices
type MetricsCollector struct {
	serviceName string
	metrics     map[string]*Metric
	mu          sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(serviceName string) *MetricsCollector {
	return &MetricsCollector{
		serviceName: serviceName,
		metrics:     make(map[string]*Metric),
	}
}

// RecordHTTPRequest records HTTP request metrics
func (mc *MetricsCollector) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration, requestSize, responseSize int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Record request count
	key := "http_requests_total"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value++
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    map[string]string{"method": method, "endpoint": endpoint, "status_code": statusCode, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}

	// Record request duration
	key = "http_request_duration_seconds"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value = duration.Seconds()
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeHistogram,
			Value:     duration.Seconds(),
			Labels:    map[string]string{"method": method, "endpoint": endpoint, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}
}

// RecordGRPCRequest records gRPC request metrics
func (mc *MetricsCollector) RecordGRPCRequest(method, status string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Record request count
	key := "grpc_requests_total"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value++
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    map[string]string{"method": method, "status": status, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}

	// Record request duration
	key = "grpc_request_duration_seconds"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value = duration.Seconds()
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeHistogram,
			Value:     duration.Seconds(),
			Labels:    map[string]string{"method": method, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}
}

// RecordDatabaseQuery records database query metrics
func (mc *MetricsCollector) RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Record query count
	key := "db_queries_total"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value++
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    map[string]string{"operation": operation, "table": table, "status": status, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}

	// Record query duration
	key = "db_query_duration_seconds"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value = duration.Seconds()
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeHistogram,
			Value:     duration.Seconds(),
			Labels:    map[string]string{"operation": operation, "table": table, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}
}

// SetDatabaseConnections sets database connection metrics
func (mc *MetricsCollector) SetDatabaseConnections(database string, active, idle int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Set active connections
	key := "db_connections_active"
	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeGauge,
		Value:     float64(active),
		Labels:    map[string]string{"database": database, "service": mc.serviceName},
		Timestamp: time.Now(),
	}

	// Set idle connections
	key = "db_connections_idle"
	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeGauge,
		Value:     float64(idle),
		Labels:    map[string]string{"database": database, "service": mc.serviceName},
		Timestamp: time.Now(),
	}
}

// SetBusinessMetrics sets business metrics
func (mc *MetricsCollector) SetBusinessMetrics(metricType string, value float64, labels ...string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := metricType + "_total"
	labelMap := map[string]string{"service": mc.serviceName}

	// Add additional labels
	for i := 0; i < len(labels); i += 2 {
		if i+1 < len(labels) {
			labelMap[labels[i]] = labels[i+1]
		}
	}

	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeGauge,
		Value:     value,
		Labels:    labelMap,
		Timestamp: time.Now(),
	}
}

// RecordNotification records notification metrics
func (mc *MetricsCollector) RecordNotification(notificationType, status string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := "notifications_total"
	if metric, exists := mc.metrics[key]; exists {
		metric.Value++
	} else {
		mc.metrics[key] = &Metric{
			Name:      key,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    map[string]string{"type": notificationType, "status": status, "service": mc.serviceName},
			Timestamp: time.Now(),
		}
	}
}

// SetSystemMetrics sets system metrics
func (mc *MetricsCollector) SetSystemMetrics(memoryUsage, cpuUsage float64, goroutines int, gcDuration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Set memory usage
	key := "memory_usage_bytes"
	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeGauge,
		Value:     memoryUsage,
		Labels:    map[string]string{"type": "heap", "service": mc.serviceName},
		Timestamp: time.Now(),
	}

	// Set CPU usage
	key = "cpu_usage_percent"
	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeGauge,
		Value:     cpuUsage,
		Labels:    map[string]string{"service": mc.serviceName},
		Timestamp: time.Now(),
	}

	// Set goroutines count
	key = "goroutines_total"
	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeGauge,
		Value:     float64(goroutines),
		Labels:    map[string]string{"service": mc.serviceName},
		Timestamp: time.Now(),
	}

	// Set GC duration
	key = "gc_duration_seconds"
	mc.metrics[key] = &Metric{
		Name:      key,
		Type:      MetricTypeHistogram,
		Value:     gcDuration.Seconds(),
		Labels:    map[string]string{"service": mc.serviceName},
		Timestamp: time.Now(),
	}
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := make(map[string]*Metric)
	for key, metric := range mc.metrics {
		metrics[key] = metric
	}
	return metrics
}

// MetricsMiddleware provides HTTP middleware for metrics collection
type MetricsMiddleware struct {
	collector *MetricsCollector
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(collector *MetricsCollector) *MetricsMiddleware {
	return &MetricsMiddleware{
		collector: collector,
	}
}

// HTTPMiddleware returns HTTP middleware for metrics collection
func (mm *MetricsMiddleware) HTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture response size
			wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: 200}

			// Get request size
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start)
			mm.collector.RecordHTTPRequest(
				r.Method,
				r.URL.Path,
				http.StatusText(wrapped.statusCode),
				duration,
				requestSize,
				wrapped.responseSize,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture response size and status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(n)
	return n, err
}

// MetricsCollectorInterface defines the interface for metrics collection
type MetricsCollectorInterface interface {
	RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration, requestSize, responseSize int64)
	RecordGRPCRequest(method, status string, duration time.Duration)
	RecordDatabaseQuery(operation, table, status string, duration time.Duration)
	SetDatabaseConnections(database string, active, idle int)
	SetBusinessMetrics(metricType string, value float64, labels ...string)
	RecordNotification(notificationType, status string)
	SetSystemMetrics(memoryUsage, cpuUsage float64, goroutines int, gcDuration time.Duration)
	GetMetrics() map[string]*Metric
}

// Global metrics collector instance
var globalMetricsCollector *MetricsCollector

// InitGlobalMetrics initializes the global metrics collector
func InitGlobalMetrics(serviceName string) {
	globalMetricsCollector = NewMetricsCollector(serviceName)
}

// GetGlobalMetrics returns the global metrics collector
func GetGlobalMetrics() *MetricsCollector {
	return globalMetricsCollector
}

// RecordHTTPRequest is a convenience function for recording HTTP requests
func RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration, requestSize, responseSize int64) {
	if globalMetricsCollector != nil {
		globalMetricsCollector.RecordHTTPRequest(method, endpoint, statusCode, duration, requestSize, responseSize)
	}
}

// RecordGRPCRequest is a convenience function for recording gRPC requests
func RecordGRPCRequest(method, status string, duration time.Duration) {
	if globalMetricsCollector != nil {
		globalMetricsCollector.RecordGRPCRequest(method, status, duration)
	}
}

// RecordDatabaseQuery is a convenience function for recording database queries
func RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	if globalMetricsCollector != nil {
		globalMetricsCollector.RecordDatabaseQuery(operation, table, status, duration)
	}
}

// SetBusinessMetrics is a convenience function for setting business metrics
func SetBusinessMetrics(metricType string, value float64, labels ...string) {
	if globalMetricsCollector != nil {
		globalMetricsCollector.SetBusinessMetrics(metricType, value, labels...)
	}
}

// RecordNotification is a convenience function for recording notifications
func RecordNotification(notificationType, status string) {
	if globalMetricsCollector != nil {
		globalMetricsCollector.RecordNotification(notificationType, status)
	}
}
