package tracing

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPMiddleware creates HTTP middleware for tracing
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start span
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := GetGlobalTracer().StartSpanWithAttributes(r.Context(), spanName, map[string]interface{}{
				"http.method":  r.Method,
				"http.url":     r.URL.String(),
				"http.host":    r.Host,
				"service.name": serviceName,
			})
			defer func() {
				span.End = time.Now()
				span.Duration = span.End.Sub(span.Start)
			}()

			// Create response writer wrapper
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Add response attributes
			span.Attributes["http.status_code"] = wrapped.statusCode
			span.Attributes["http.status_text"] = http.StatusText(wrapped.statusCode)

			// Set span status based on status code
			if wrapped.statusCode >= 400 {
				span.Status = "error"
			} else {
				span.Status = "success"
			}
		})
	}
}

// HTTPClientMiddleware creates HTTP client middleware for tracing
func HTTPClientMiddleware(serviceName string) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		// Start span
		spanName := fmt.Sprintf("%s %s", req.Method, req.URL.Host)
		ctx, span := GetGlobalTracer().StartSpanWithAttributes(req.Context(), spanName, map[string]interface{}{
			"http.method":  req.Method,
			"http.url":     req.URL.String(),
			"http.host":    req.URL.Host,
			"service.name": serviceName,
		})
		defer func() {
			span.End = time.Now()
			span.Duration = span.End.Sub(span.Start)
		}()

		// Make the request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req.WithContext(ctx))

		// Add response attributes
		if resp != nil {
			span.Attributes["http.status_code"] = resp.StatusCode
			span.Attributes["http.status_text"] = resp.Status

			// Set span status based on status code
			if resp.StatusCode >= 400 {
				span.Status = "error"
			} else {
				span.Status = "success"
			}
		}

		// Record error if any
		if err != nil {
			span.Error = err
			span.Status = "error"
		}

		return resp, err
	}
}

// TraceHTTPRequest traces an HTTP request
func TraceHTTPRequest(ctx context.Context, serviceName string, req *http.Request) (context.Context, *Span) {
	// Start span
	spanName := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	ctx, span := GetGlobalTracer().StartSpanWithAttributes(ctx, spanName, map[string]interface{}{
		"http.method":  req.Method,
		"http.url":     req.URL.String(),
		"http.host":    req.URL.Host,
		"service.name": serviceName,
	})

	return ctx, span
}

// TraceHTTPResponse traces an HTTP response
func TraceHTTPResponse(span *Span, resp *http.Response, err error) {
	if resp != nil {
		span.Attributes["http.status_code"] = resp.StatusCode
		span.Attributes["http.status_text"] = resp.Status

		// Set span status based on status code
		if resp.StatusCode >= 400 {
			span.Status = "error"
		} else {
			span.Status = "success"
		}
	}

	// Record error if any
	if err != nil {
		span.Error = err
		span.Status = "error"
	}

	span.End = time.Now()
	span.Duration = span.End.Sub(span.Start)
}

// TraceDatabaseOperation traces a database operation
func TraceDatabaseOperation(ctx context.Context, serviceName string, operation string, query string) (context.Context, *Span) {
	// Start span
	spanName := fmt.Sprintf("db.%s", operation)
	ctx, span := GetGlobalTracer().StartSpanWithAttributes(ctx, spanName, map[string]interface{}{
		"db.operation": operation,
		"db.statement": query,
		"service.name": serviceName,
	})

	return ctx, span
}

// TraceExternalService traces an external service call
func TraceExternalService(ctx context.Context, serviceName string, externalService string, operation string) (context.Context, *Span) {
	// Start span
	spanName := fmt.Sprintf("external.%s.%s", externalService, operation)
	ctx, span := GetGlobalTracer().StartSpanWithAttributes(ctx, spanName, map[string]interface{}{
		"external.service":   externalService,
		"external.operation": operation,
		"service.name":       serviceName,
	})

	return ctx, span
}

// TraceEventProcessing traces event processing
func TraceEventProcessing(ctx context.Context, serviceName string, eventType string, eventData map[string]interface{}) (context.Context, *Span) {
	// Start span
	spanName := fmt.Sprintf("event.%s", eventType)
	ctx, span := GetGlobalTracer().StartSpanWithAttributes(ctx, spanName, map[string]interface{}{
		"event.type":   eventType,
		"service.name": serviceName,
	})

	// Add event data as attributes
	for k, v := range eventData {
		span.Attributes[fmt.Sprintf("event.data.%s", k)] = v
	}

	return ctx, span
}

// TraceMessageProcessing traces message processing
func TraceMessageProcessing(ctx context.Context, serviceName string, messageType string, messageData map[string]interface{}) (context.Context, *Span) {
	// Start span
	spanName := fmt.Sprintf("message.%s", messageType)
	ctx, span := GetGlobalTracer().StartSpanWithAttributes(ctx, spanName, map[string]interface{}{
		"message.type": messageType,
		"service.name": serviceName,
	})

	// Add message data as attributes
	for k, v := range messageData {
		span.Attributes[fmt.Sprintf("message.data.%s", k)] = v
	}

	return ctx, span
}

// GetTraceIDFromContext returns the trace ID from context
func GetTraceIDFromContext(ctx context.Context) string {
	if span, ok := ctx.Value("span").(*Span); ok {
		return fmt.Sprintf("trace-%d", span.Start.UnixNano())
	}
	return ""
}

// GetSpanIDFromContext returns the span ID from context
func GetSpanIDFromContext(ctx context.Context) string {
	if span, ok := ctx.Value("span").(*Span); ok {
		return fmt.Sprintf("span-%d", span.Start.UnixNano())
	}
	return ""
}

// AddSpanAttributes adds attributes to a span
func AddSpanAttributes(span *Span, attrs map[string]interface{}) {
	if span.Attributes == nil {
		span.Attributes = make(map[string]interface{})
	}
	for k, v := range attrs {
		span.Attributes[k] = v
	}
}

// AddSpanEvent adds an event to a span
func AddSpanEvent(span *Span, name string, attrs map[string]interface{}) {
	event := &Event{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: attrs,
	}
	span.Events = append(span.Events, event)
}

// AddSpanError adds an error to a span
func AddSpanError(span *Span, err error) {
	span.Error = err
	span.Status = "error"
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
