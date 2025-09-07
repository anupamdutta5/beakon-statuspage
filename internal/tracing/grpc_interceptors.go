package tracing

import (
	"context"
	"fmt"
	"time"
)

// UnaryServerInterceptor creates a gRPC unary server interceptor for tracing
func UnaryServerInterceptor(serviceName string) func(ctx context.Context, req interface{}, info interface{}, handler func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info interface{}, handler func(context.Context, interface{}) (interface{}, error)) (interface{}, error) {
		// Start span
		ctx, span := GetGlobalTracer().StartSpan(ctx, fmt.Sprintf("grpc.server.%s", serviceName))
		defer func() {
			span.End = time.Now()
			span.Duration = span.End.Sub(span.Start)
		}()

		// Process request
		resp, err := handler(ctx, req)

		// Add error if any
		if err != nil {
			span.Error = err
			span.Status = "error"
		} else {
			span.Status = "success"
		}

		return resp, err
	}
}

// StreamServerInterceptor creates a gRPC stream server interceptor for tracing
func StreamServerInterceptor(serviceName string) func(srv interface{}, ss interface{}, info interface{}, handler func(interface{}, interface{}) error) error {
	return func(srv interface{}, ss interface{}, info interface{}, handler func(interface{}, interface{}) error) error {
		// Start span
		_, span := GetGlobalTracer().StartSpan(context.Background(), fmt.Sprintf("grpc.stream.%s", serviceName))
		defer func() {
			span.End = time.Now()
			span.Duration = span.End.Sub(span.Start)
		}()

		// Process stream
		err := handler(srv, ss)

		// Add error if any
		if err != nil {
			span.Error = err
			span.Status = "error"
		} else {
			span.Status = "success"
		}

		return err
	}
}

// UnaryClientInterceptor creates a gRPC unary client interceptor for tracing
func UnaryClientInterceptor(serviceName string) func(ctx context.Context, method string, req, reply interface{}, cc interface{}, invoker func(context.Context, string, interface{}, interface{}, interface{}) error) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc interface{}, invoker func(context.Context, string, interface{}, interface{}, interface{}) error) error {
		// Start span
		ctx, span := GetGlobalTracer().StartSpan(ctx, fmt.Sprintf("grpc.client.%s", serviceName))
		defer func() {
			span.End = time.Now()
			span.Duration = span.End.Sub(span.Start)
		}()

		// Make the call
		err := invoker(ctx, method, req, reply, cc)

		// Add error if any
		if err != nil {
			span.Error = err
			span.Status = "error"
		} else {
			span.Status = "success"
		}

		return err
	}
}

// StreamClientInterceptor creates a gRPC stream client interceptor for tracing
func StreamClientInterceptor(serviceName string) func(ctx context.Context, desc interface{}, cc interface{}, method string, streamer func(context.Context, interface{}, interface{}, string) (interface{}, error)) (interface{}, error) {
	return func(ctx context.Context, desc interface{}, cc interface{}, method string, streamer func(context.Context, interface{}, interface{}, string) (interface{}, error)) (interface{}, error) {
		// Start span
		ctx, span := GetGlobalTracer().StartSpan(ctx, fmt.Sprintf("grpc.stream.client.%s", serviceName))
		defer func() {
			span.End = time.Now()
			span.Duration = span.End.Sub(span.Start)
		}()

		// Create stream
		stream, err := streamer(ctx, desc, cc, method)
		if err != nil {
			span.Error = err
			span.Status = "error"
			return nil, err
		}

		span.Status = "success"
		return stream, nil
	}
}
