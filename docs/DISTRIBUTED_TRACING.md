# Distributed Tracing Implementation

This document describes the distributed tracing implementation for the Status Page microservices architecture using OpenTelemetry and Jaeger.

## Overview

Distributed tracing allows us to track requests as they flow through multiple microservices, providing visibility into the system's behavior and performance. This implementation uses OpenTelemetry for instrumentation and Jaeger for trace collection and visualization.

## Architecture

### Components

1. **Tracer** - Main tracing component that manages spans and trace context
2. **HTTP Middleware** - HTTP request/response tracing
3. **gRPC Interceptors** - gRPC request/response tracing
4. **Event Tracing** - Event processing tracing
5. **Database Tracing** - Database operation tracing

### Trace Flow

```
Client Request → API Gateway → Microservice A → Microservice B → Database
     ↓              ↓              ↓              ↓              ↓
   Span 1        Span 2        Span 3        Span 4        Span 5
```

## Implementation

### Tracer Component

The main tracer component is implemented in `internal/tracing/tracer.go`:

```go
type Tracer struct {
    tracerProvider *sdktrace.TracerProvider
    tracer         trace.Tracer
    serviceName    string
    serviceVersion string
}
```

#### Key Features

- **Jaeger Integration** - Exports traces to Jaeger collector
- **Resource Management** - Manages service metadata and versioning
- **Sampling** - Configurable sampling ratio for performance
- **Context Propagation** - Propagates trace context across services

#### Usage Example

```go
// Create tracer
tracer, err := tracing.NewTracer(cfg, "user-service", "1.0.0")
if err != nil {
    log.Fatal("Failed to create tracer:", err)
}
defer tracer.Close()

// Start span
ctx, span := tracer.StartSpan(ctx, "user.create")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("user.id", "123"),
    attribute.String("user.email", "user@example.com"),
)

// Add event
span.AddEvent("user.created", trace.WithAttributes(
    attribute.String("user.id", "123"),
))
```

### HTTP Middleware

HTTP middleware is implemented in `internal/tracing/http_middleware.go`:

#### Server Middleware

```go
// Create HTTP middleware
middleware := tracing.HTTPMiddleware("user-service")

// Use with HTTP server
http.Handle("/api/users", middleware(http.HandlerFunc(handleUsers)))
```

#### Client Middleware

```go
// Create HTTP client middleware
clientMiddleware := tracing.HTTPClientMiddleware("user-service")

// Use with HTTP client
req, _ := http.NewRequest("GET", "http://api.example.com/users", nil)
resp, err := clientMiddleware(req)
```

#### Features

- **Automatic Span Creation** - Creates spans for all HTTP requests
- **Context Propagation** - Propagates trace context via headers
- **Status Code Tracking** - Tracks HTTP status codes and errors
- **Performance Metrics** - Measures request duration and performance

### gRPC Interceptors

gRPC interceptors are implemented in `internal/tracing/grpc_interceptors.go`:

#### Server Interceptor

```go
// Create gRPC server with tracing
server := grpc.NewServer(
    grpc.UnaryInterceptor(tracing.UnaryServerInterceptor("user-service")),
    grpc.StreamInterceptor(tracing.StreamServerInterceptor("user-service")),
)
```

#### Client Interceptor

```go
// Create gRPC client with tracing
conn, err := grpc.Dial("user-service:8080",
    grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor("gateway")),
    grpc.WithStreamInterceptor(tracing.StreamClientInterceptor("gateway")),
)
```

#### Features

- **Automatic Span Creation** - Creates spans for all gRPC calls
- **Context Propagation** - Propagates trace context via metadata
- **Error Tracking** - Tracks gRPC errors and status codes
- **Stream Support** - Supports both unary and streaming RPCs

### Event Tracing

Event tracing is implemented for event sourcing:

```go
// Trace event processing
ctx, span := tracing.TraceEventProcessing(ctx, "user-service", "UserCreated", eventData)
defer span.End()

// Add event attributes
span.SetAttributes(
    attribute.String("event.aggregate_id", event.AggregateID),
    attribute.String("event.aggregate_type", event.AggregateType),
)
```

### Database Tracing

Database tracing is implemented for database operations:

```go
// Trace database operation
ctx, span := tracing.TraceDatabaseOperation(ctx, "user-service", "SELECT", query)
defer span.End()

// Execute database operation
rows, err := db.QueryContext(ctx, query)
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

## Configuration

### Environment Variables

```bash
# Jaeger configuration
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
JAEGER_SAMPLING_RATIO=0.1

# Service configuration
SERVICE_NAME=user-service
SERVICE_VERSION=1.0.0
```

### Configuration Structure

```go
type TracingConfig struct {
    JaegerEndpoint  string  `json:"jaeger_endpoint"`
    SamplingRatio   float64 `json:"sampling_ratio"`
    ServiceName     string  `json:"service_name"`
    ServiceVersion  string  `json:"service_version"`
}
```

## Trace Attributes

### Standard Attributes

- **service.name** - Name of the service
- **service.version** - Version of the service
- **http.method** - HTTP method (GET, POST, etc.)
- **http.url** - HTTP URL
- **http.status_code** - HTTP status code
- **rpc.system** - RPC system (grpc, http, etc.)
- **rpc.service** - RPC service name
- **rpc.method** - RPC method name
- **db.system** - Database system (postgresql, mysql, etc.)
- **db.operation** - Database operation (SELECT, INSERT, etc.)
- **db.statement** - Database statement
- **event.type** - Event type
- **event.aggregate_id** - Event aggregate ID
- **event.aggregate_type** - Event aggregate type

### Custom Attributes

- **user.id** - User ID
- **tenant.id** - Tenant ID
- **request.id** - Request ID
- **component.id** - Component ID
- **incident.id** - Incident ID
- **payment.id** - Payment ID

## Span Types

### Server Spans

- **HTTP Server** - Incoming HTTP requests
- **gRPC Server** - Incoming gRPC requests
- **Event Consumer** - Event processing
- **Message Consumer** - Message processing

### Client Spans

- **HTTP Client** - Outgoing HTTP requests
- **gRPC Client** - Outgoing gRPC requests
- **Database Client** - Database operations
- **External Service** - External service calls

### Internal Spans

- **Business Logic** - Business logic operations
- **Data Processing** - Data processing operations
- **Validation** - Input validation
- **Transformation** - Data transformation

## Error Handling

### Error Tracking

```go
// Record error in span
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
```

### Error Attributes

- **error.type** - Error type
- **error.message** - Error message
- **error.stack** - Error stack trace
- **error.code** - Error code

## Performance Monitoring

### Metrics

- **Trace Duration** - Time taken to complete traces
- **Span Count** - Number of spans per trace
- **Error Rate** - Percentage of failed traces
- **Throughput** - Number of traces per second

### Sampling

```go
// Configure sampling ratio
tracerProvider := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // 10% sampling
)
```

## Integration with Services

### API Gateway

```go
// Add tracing middleware to API Gateway
gateway := &Gateway{
    tracer: tracing.NewTracer(cfg, "api-gateway", "1.0.0"),
}

// Use HTTP middleware
gateway.router.Use(tracing.HTTPMiddleware("api-gateway"))
```

### Microservices

```go
// Add tracing to microservices
service := &UserService{
    tracer: tracing.NewTracer(cfg, "user-service", "1.0.0"),
}

// Use gRPC interceptors
server := grpc.NewServer(
    grpc.UnaryInterceptor(tracing.UnaryServerInterceptor("user-service")),
)
```

### Event Processing

```go
// Add tracing to event processing
subscriber := &EventSubscriber{
    tracer: tracing.NewTracer(cfg, "event-processor", "1.0.0"),
}

// Trace event processing
ctx, span := subscriber.tracer.TraceEventProcessing(ctx, "UserCreated", eventData)
defer span.End()
```

## Jaeger Integration

### Jaeger Configuration

```yaml
# docker-compose.yml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # HTTP collector
      - "14250:14250"  # gRPC collector
    environment:
      - COLLECTOR_OTLP_ENABLED=true
```

### Trace Export

```go
// Create Jaeger exporter
exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
    jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
))
```

### Trace Visualization

- **Trace Timeline** - Visual representation of trace execution
- **Span Details** - Detailed information about each span
- **Service Map** - Service dependency visualization
- **Performance Metrics** - Performance analysis and optimization

## Best Practices

### Span Naming

- Use descriptive names: `user.create`, `payment.process`, `incident.resolve`
- Include service name: `user-service.create`, `payment-service.process`
- Use consistent naming conventions

### Attribute Naming

- Use lowercase with dots: `user.id`, `tenant.name`, `payment.amount`
- Use standard attribute names when possible
- Include relevant business context

### Error Handling

- Always record errors in spans
- Set appropriate span status
- Include error context and stack traces
- Don't expose sensitive information

### Performance

- Use appropriate sampling ratios
- Avoid high-cardinality attributes
- Limit span duration and size
- Monitor trace performance impact

## Monitoring and Alerting

### Metrics

- **Trace Success Rate** - Percentage of successful traces
- **Trace Duration** - Average and P95 trace duration
- **Error Rate** - Percentage of traces with errors
- **Throughput** - Number of traces per second

### Alerts

- **High Error Rate** - Alert when error rate exceeds threshold
- **Slow Traces** - Alert when trace duration exceeds threshold
- **Missing Traces** - Alert when expected traces are missing
- **Service Down** - Alert when service stops producing traces

## Testing

### Unit Tests

```go
func TestTracer(t *testing.T) {
    // Create test tracer
    tracer, err := tracing.NewTracer(cfg, "test-service", "1.0.0")
    assert.NoError(t, err)
    defer tracer.Close()

    // Test span creation
    ctx, span := tracer.StartSpan(context.Background(), "test.operation")
    assert.NotNil(t, span)
    span.End()
}
```

### Integration Tests

```go
func TestHTTPTracing(t *testing.T) {
    // Create test server with tracing
    server := httptest.NewServer(tracing.HTTPMiddleware("test-service")(handler))
    defer server.Close()

    // Make request
    resp, err := http.Get(server.URL + "/test")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Deployment

### Docker

```dockerfile
# Add OpenTelemetry dependencies
RUN go mod download

# Set environment variables
ENV JAEGER_ENDPOINT=http://jaeger:14268/api/traces
ENV SERVICE_NAME=user-service
ENV SERVICE_VERSION=1.0.0
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  template:
    spec:
      containers:
      - name: user-service
        env:
        - name: JAEGER_ENDPOINT
          value: "http://jaeger:14268/api/traces"
        - name: SERVICE_NAME
          value: "user-service"
        - name: SERVICE_VERSION
          value: "1.0.0"
```

## Troubleshooting

### Common Issues

1. **Traces Not Appearing** - Check Jaeger endpoint configuration
2. **High Memory Usage** - Reduce sampling ratio or span size
3. **Slow Performance** - Optimize span creation and attribute setting
4. **Missing Context** - Ensure proper context propagation

### Debugging

```go
// Enable debug logging
import "go.opentelemetry.io/otel/sdk/trace"

// Set debug level
trace.SetLogger(logger)

// Check trace context
traceID := tracing.GetTraceIDFromContext(ctx)
spanID := tracing.GetSpanIDFromContext(ctx)
```

## Conclusion

The distributed tracing implementation provides comprehensive visibility into the microservices architecture, enabling:

- **Request Tracking** - Track requests across multiple services
- **Performance Analysis** - Identify bottlenecks and optimize performance
- **Error Debugging** - Quickly identify and debug errors
- **Service Dependencies** - Understand service relationships
- **Business Insights** - Analyze user behavior and system usage

This implementation follows OpenTelemetry standards and provides a solid foundation for observability in the microservices architecture.
