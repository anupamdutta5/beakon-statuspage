package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/gateway/discovery"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Client manages gRPC connections to microservices
type Client struct {
	config    *config.Config
	discovery *discovery.ServiceDiscovery
	conns     map[string]*Connection
}

// Connection represents a gRPC connection
type Connection struct {
	ServiceName string
	URL         string
	Connected   bool
	LastUsed    time.Time
}

// NewClient creates a new gRPC client
func NewClient(cfg *config.Config, discovery *discovery.ServiceDiscovery) *Client {
	return &Client{
		config:    cfg,
		discovery: discovery,
		conns:     make(map[string]*Connection),
	}
}

// GetConnection returns a gRPC connection to the specified service
func (c *Client) GetConnection(serviceName string) (*Connection, error) {
	// Check if connection already exists
	if conn, exists := c.conns[serviceName]; exists {
		conn.LastUsed = time.Now()
		return conn, nil
	}

	// Get service URL
	serviceURL, err := c.discovery.GetServiceURL(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL for %s: %w", serviceName, err)
	}

	// Create gRPC connection
	conn := &Connection{
		ServiceName: serviceName,
		URL:         serviceURL,
		Connected:   true,
		LastUsed:    time.Now(),
	}

	// Store connection
	c.conns[serviceName] = conn

	logger.Log.Info("gRPC connection established",
		zap.String("service", serviceName),
		zap.String("url", serviceURL))

	return conn, nil
}

// Close closes all gRPC connections
func (c *Client) Close() error {
	var lastErr error
	for serviceName, conn := range c.conns {
		conn.Connected = false
		logger.Log.Info("gRPC connection closed",
			zap.String("service", serviceName))
	}
	return lastErr
}

// Call makes a gRPC call to the specified service
func (c *Client) Call(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
	// Get connection
	_, err := c.GetConnection(serviceName)
	if err != nil {
		return nil, err
	}

	// Add tracing context
	ctx = c.addTracingContext(ctx, method)

	// Log request
	logger.Log.Debug("gRPC request started",
		zap.String("method", method),
		zap.String("service", serviceName))

	// Simulate gRPC call
	start := time.Now()

	// In a real implementation, this would make an actual gRPC call
	// For now, we'll just simulate the call
	time.Sleep(10 * time.Millisecond)

	// Log response
	duration := time.Since(start)
	logger.Log.Debug("gRPC request completed",
		zap.String("method", method),
		zap.String("service", serviceName),
		zap.Duration("duration", duration))

	// Return mock response
	return map[string]interface{}{
		"service": serviceName,
		"method":  method,
		"status":  "success",
	}, nil
}

// addTracingContext adds tracing context to the gRPC call
func (c *Client) addTracingContext(ctx context.Context, method string) context.Context {
	// In a real implementation, this would add distributed tracing headers
	// For now, we'll just add the method name
	return context.WithValue(ctx, "grpc_method", method)
}

// StreamCall makes a streaming gRPC call to the specified service
func (c *Client) StreamCall(ctx context.Context, serviceName, method string, request interface{}) (*Stream, error) {
	// Get connection
	_, err := c.GetConnection(serviceName)
	if err != nil {
		return nil, err
	}

	// Add tracing context
	ctx = c.addTracingContext(ctx, method)

	// Log request
	logger.Log.Debug("gRPC stream started",
		zap.String("method", method),
		zap.String("service", serviceName))

	// Create stream
	stream := &Stream{
		serviceName: serviceName,
		method:      method,
		ctx:         ctx,
		start:       time.Now(),
	}

	return stream, nil
}

// Stream represents a gRPC stream
type Stream struct {
	serviceName string
	method      string
	ctx         context.Context
	start       time.Time
	closed      bool
}

// Send sends a message through the stream
func (s *Stream) Send(msg interface{}) error {
	if s.closed {
		return fmt.Errorf("stream is closed")
	}

	logger.Log.Debug("gRPC stream send",
		zap.String("method", s.method),
		zap.String("service", s.serviceName))

	return nil
}

// Recv receives a message from the stream
func (s *Stream) Recv() (interface{}, error) {
	if s.closed {
		return nil, fmt.Errorf("stream is closed")
	}

	logger.Log.Debug("gRPC stream receive",
		zap.String("method", s.method),
		zap.String("service", s.serviceName))

	// Return mock message
	return map[string]interface{}{
		"service": s.serviceName,
		"method":  s.method,
		"data":    "stream data",
	}, nil
}

// Close closes the stream
func (s *Stream) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true
	duration := time.Since(s.start)

	logger.Log.Debug("gRPC stream closed",
		zap.String("method", s.method),
		zap.String("service", s.serviceName),
		zap.Duration("duration", duration))

	return nil
}
