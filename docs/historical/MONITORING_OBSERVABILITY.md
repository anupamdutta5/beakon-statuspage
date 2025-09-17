# Monitoring and Observability Guide

This document describes the comprehensive monitoring and observability system for the Status Page microservices architecture.

## Overview

The monitoring and observability system provides comprehensive insights into the health, performance, and behavior of the microservices. It includes metrics collection, health checks, structured logging, and distributed tracing.

## Architecture

### Components

1. **Metrics Collection** - Prometheus-compatible metrics
2. **Health Checks** - Service health monitoring
3. **Structured Logging** - Centralized logging with context
4. **Distributed Tracing** - Request tracing across services
5. **Alerting** - Automated alerting based on metrics
6. **Dashboards** - Visualization of metrics and logs

### Monitoring Stack

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Microservices │    │   API Gateway   │    │   Load Balancer │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Metrics       │    │   Health Checks │    │   Logging       │
│   Collection    │    │   & Monitoring  │    │   & Tracing     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Prometheus    │    │   Grafana       │    │   Jaeger        │
│   (Metrics)     │    │   (Dashboards)  │    │   (Tracing)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AlertManager  │    │   Elasticsearch │    │   Kibana        │
│   (Alerts)      │    │   (Logs)        │    │   (Log Analysis)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Metrics Collection

### Prometheus Metrics

The system exposes Prometheus-compatible metrics for:

#### HTTP Metrics
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - HTTP request duration
- `http_request_size_bytes` - HTTP request size
- `http_response_size_bytes` - HTTP response size

#### gRPC Metrics
- `grpc_requests_total` - Total gRPC requests
- `grpc_request_duration_seconds` - gRPC request duration

#### Database Metrics
- `db_connections_active` - Active database connections
- `db_connections_idle` - Idle database connections
- `db_queries_total` - Total database queries
- `db_query_duration_seconds` - Database query duration

#### Business Metrics
- `users_total` - Total number of users
- `tenants_total` - Total number of tenants
- `components_total` - Total number of components
- `incidents_total` - Total number of incidents
- `notifications_total` - Total notifications sent

#### System Metrics
- `memory_usage_bytes` - Memory usage
- `cpu_usage_percent` - CPU usage
- `goroutines_total` - Number of goroutines
- `gc_duration_seconds` - Garbage collection duration

### Usage Example

```go
// Initialize metrics collector
collector := monitoring.NewMetricsCollector("user-service")

// Record HTTP request
collector.RecordHTTPRequest("GET", "/users", "200", duration, requestSize, responseSize)

// Record gRPC request
collector.RecordGRPCRequest("GetUser", "OK", duration)

// Record database query
collector.RecordDatabaseQuery("SELECT", "users", "success", duration)

// Set business metrics
collector.SetBusinessMetrics("users", 1000, "active")

// Record notification
collector.RecordNotification("email", "success")
```

## Health Checks

### Health Check Types

#### 1. Database Health Check
```go
// Create database health check
dbCheck := monitoring.NewDatabaseHealthCheck("postgresql", func(ctx context.Context) error {
    return db.PingContext(ctx)
})

// Register health check
registry.Register("database", dbCheck)
```

#### 2. External Service Health Check
```go
// Create external service health check
externalCheck := monitoring.NewExternalServiceHealthCheck("payment-gateway", "https://api.stripe.com/health")

// Register health check
registry.Register("payment-gateway", externalCheck)
```

#### 3. Custom Health Check
```go
// Create custom health check
customCheck := monitoring.NewCustomHealthCheck("business-logic", func(ctx context.Context) (monitoring.HealthStatus, string, map[string]interface{}) {
    // Perform custom health check
    if isHealthy {
        return monitoring.HealthStatusHealthy, "All systems operational", map[string]interface{}{
            "active_connections": 10,
            "queue_size": 0,
        }
    }
    return monitoring.HealthStatusUnhealthy, "System degraded", nil
})

// Register health check
registry.Register("business-logic", customCheck)
```

### Health Check Endpoints

#### Basic Health Check
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "statuspage"
}
```

#### Detailed Health Check
```bash
curl http://localhost:8080/health/detailed
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "statuspage",
  "checks": {
    "database": {
      "name": "database",
      "status": "healthy",
      "message": "Database connection is healthy",
      "duration": "2ms",
      "last_checked": "2024-01-15T10:30:00Z"
    },
    "payment-gateway": {
      "name": "payment-gateway",
      "status": "healthy",
      "message": "External service is healthy",
      "duration": "150ms",
      "last_checked": "2024-01-15T10:30:00Z"
    }
  }
}
```

#### Readiness Check
```bash
curl http://localhost:8080/ready
```

Response:
```json
{
  "ready": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "statuspage"
}
```

#### Liveness Check
```bash
curl http://localhost:8080/live
```

Response:
```json
{
  "alive": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "statuspage"
}
```

## Structured Logging

### Log Configuration

```go
// Log configuration
config := monitoring.LogConfig{
    Level:      monitoring.LogLevelInfo,
    Format:     "json",
    Output:     "stdout",
    FilePath:   "/var/log/statuspage.log",
    MaxSize:    100, // MB
    MaxBackups: 3,
    MaxAge:     7, // days
    Compress:   true,
}

// Initialize logger
logger, err := monitoring.NewLogger(config, "user-service", "1.0.0", "production")
if err != nil {
    log.Fatal("Failed to initialize logger:", err)
}
```

### Log Levels

- **DEBUG** - Detailed information for debugging
- **INFO** - General information about application flow
- **WARN** - Warning messages for potential issues
- **ERROR** - Error messages for recoverable errors
- **FATAL** - Fatal errors that cause application termination

### Logging Examples

#### Basic Logging
```go
// Debug logging
logger.Debug("Processing user request", zap.String("user_id", "123"))

// Info logging
logger.Info("User created successfully", zap.String("user_id", "123"))

// Warning logging
logger.Warn("Rate limit approaching", zap.Int("current_rate", 95))

// Error logging
logger.Error("Failed to process payment", zap.Error(err))

// Fatal logging
logger.Fatal("Database connection failed", zap.Error(err))
```

#### Context-Aware Logging
```go
// Create context with trace ID
ctx := context.WithValue(context.Background(), "trace_id", "abc123")

// Create logger with context
ctxLogger := logger.WithContext(ctx)

// Log with context
ctxLogger.Info("Processing request", zap.String("endpoint", "/users"))
```

#### Business Event Logging
```go
// Log business events
logger.LogBusinessEvent("user_created", "user", "123",
    zap.String("email", "user@example.com"),
    zap.String("plan", "premium"),
)

logger.LogBusinessEvent("payment_processed", "payment", "pay_123",
    zap.Float64("amount", 99.99),
    zap.String("currency", "USD"),
)
```

#### Security Event Logging
```go
// Log security events
logger.LogSecurityEvent("failed_login", "high",
    zap.String("ip_address", "192.168.1.1"),
    zap.String("user_agent", "Mozilla/5.0..."),
    zap.String("reason", "invalid_password"),
)
```

#### Performance Event Logging
```go
// Log performance events
logger.LogPerformanceEvent("database_query", duration,
    zap.String("query", "SELECT * FROM users"),
    zap.Int64("rows_affected", 100),
)
```

## Distributed Tracing

### Trace Context Propagation

```go
// Start span
span := tracer.StartSpan("user-service:get-user")
defer span.End()

// Add span attributes
span.SetAttribute("user.id", "123")
span.SetAttribute("user.email", "user@example.com")

// Inject trace context into HTTP request
headers := make(map[string]string)
tracer.Inject(span, headers)

// Make HTTP request with trace context
req.Header.Set("X-Trace-ID", headers["trace-id"])
req.Header.Set("X-Span-ID", headers["span-id"])
```

### gRPC Tracing

```go
// gRPC server interceptor
func TracingServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        span := tracer.StartSpan(info.FullMethod)
        defer span.End()
        
        ctx = tracer.ContextWithSpan(ctx, span)
        return handler(ctx, req)
    }
}

// gRPC client interceptor
func TracingClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        span := tracer.StartSpan(method)
        defer span.End()
        
        ctx = tracer.ContextWithSpan(ctx, span)
        return invoker(ctx, method, req, reply, cc, opts...)
    }
}
```

## Alerting

### Alert Rules

#### High Error Rate
```yaml
groups:
- name: statuspage.rules
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status_code=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors per second"
```

#### High Response Time
```yaml
  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High response time detected"
      description: "95th percentile response time is {{ $value }} seconds"
```

#### Database Connection Issues
```yaml
  - alert: DatabaseConnectionIssues
    expr: db_connections_active < 1
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Database connection issues"
      description: "No active database connections"
```

#### High Memory Usage
```yaml
  - alert: HighMemoryUsage
    expr: memory_usage_bytes / (1024*1024*1024) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High memory usage"
      description: "Memory usage is {{ $value }} GB"
```

## Dashboards

### Grafana Dashboard Configuration

#### Service Overview Dashboard
```json
{
  "dashboard": {
    "title": "Status Page Services Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{service}} - {{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile - {{service}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status_code=~\"5..\"}[5m])",
            "legendFormat": "{{service}} - {{endpoint}}"
          }
        ]
      }
    ]
  }
}
```

#### Business Metrics Dashboard
```json
{
  "dashboard": {
    "title": "Business Metrics",
    "panels": [
      {
        "title": "Total Users",
        "type": "stat",
        "targets": [
          {
            "expr": "users_total",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "Total Tenants",
        "type": "stat",
        "targets": [
          {
            "expr": "tenants_total",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "Active Incidents",
        "type": "stat",
        "targets": [
          {
            "expr": "incidents_total{status=\"active\"}",
            "legendFormat": "Active"
          }
        ]
      }
    ]
  }
}
```

## Best Practices

### Metrics

1. **Use Consistent Labels** - Use consistent label names across all metrics
2. **Avoid High Cardinality** - Avoid labels with high cardinality (e.g., user IDs)
3. **Use Histograms for Latency** - Use histograms for measuring latency distributions
4. **Include Service Name** - Always include service name in metrics
5. **Use Appropriate Buckets** - Use appropriate histogram buckets for your use case

### Logging

1. **Use Structured Logging** - Use structured logging with consistent field names
2. **Include Context** - Include relevant context in log messages
3. **Use Appropriate Log Levels** - Use appropriate log levels for different types of messages
4. **Avoid Logging Sensitive Data** - Never log sensitive data like passwords or tokens
5. **Use Correlation IDs** - Use correlation IDs to trace requests across services

### Health Checks

1. **Check Dependencies** - Include health checks for all external dependencies
2. **Use Timeouts** - Use appropriate timeouts for health checks
3. **Provide Meaningful Messages** - Provide meaningful error messages in health checks
4. **Check Business Logic** - Include health checks for critical business logic
5. **Use Appropriate Status Codes** - Use appropriate HTTP status codes for health endpoints

### Tracing

1. **Propagate Trace Context** - Always propagate trace context across service boundaries
2. **Use Meaningful Span Names** - Use meaningful span names that describe the operation
3. **Add Span Attributes** - Add relevant attributes to spans
4. **Handle Trace Context** - Handle trace context in both HTTP and gRPC requests
5. **Use Sampling** - Use appropriate sampling rates for production environments

## Troubleshooting

### Common Issues

#### 1. Metrics Not Appearing
```bash
# Check if metrics endpoint is accessible
curl http://localhost:8080/metrics

# Check Prometheus configuration
kubectl get configmap prometheus-config -o yaml
```

#### 2. Health Checks Failing
```bash
# Check health check logs
kubectl logs -f deployment/user-service -n statuspage

# Check health check endpoints
curl http://localhost:8080/health/detailed
```

#### 3. Logs Not Appearing
```bash
# Check log configuration
kubectl get configmap logging-config -o yaml

# Check log files
kubectl exec -it deployment/user-service -n statuspage -- tail -f /var/log/statuspage.log
```

#### 4. Tracing Issues
```bash
# Check Jaeger configuration
kubectl get configmap jaeger-config -o yaml

# Check trace context propagation
curl -H "X-Trace-ID: abc123" http://localhost:8080/users
```

### Debugging Commands

```bash
# Check metrics
curl http://localhost:8080/metrics | grep http_requests_total

# Check health
curl http://localhost:8080/health

# Check logs
kubectl logs -f deployment/user-service -n statuspage

# Check traces
curl http://localhost:16686/api/traces?service=user-service
```

## Conclusion

The monitoring and observability system provides comprehensive insights into the health, performance, and behavior of the microservices. It includes:

- **Metrics Collection** - Prometheus-compatible metrics for all services
- **Health Checks** - Comprehensive health monitoring for all dependencies
- **Structured Logging** - Centralized logging with context and correlation
- **Distributed Tracing** - Request tracing across all services
- **Alerting** - Automated alerting based on metrics and thresholds
- **Dashboards** - Visualization of metrics, logs, and traces

This system enables:

- **Proactive Monitoring** - Early detection of issues before they impact users
- **Performance Optimization** - Identification of performance bottlenecks
- **Debugging** - Easy debugging of issues across distributed services
- **Capacity Planning** - Understanding of resource usage and scaling needs
- **Compliance** - Audit trails and security event logging

The monitoring system follows industry best practices and provides a solid foundation for operating microservices in production.
