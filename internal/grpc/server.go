package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Context key type for gRPC server
type grpcServerContextKey string

const (
	grpcServerMethodKey grpcServerContextKey = "grpc_method"
)

// Server represents a gRPC server
type Server struct {
	config     *config.Config
	grpcServer *MockGRPCServer
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
}

// MockGRPCServer represents a mock gRPC server
type MockGRPCServer struct {
	services map[string]interface{}
	running  bool
}

// NewServer creates a new gRPC server
func NewServer(cfg *config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		config: cfg,
		grpcServer: &MockGRPCServer{
			services: make(map[string]interface{}),
			running:  false,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port+1000))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	// Start server in goroutine
	go func() {
		s.grpcServer.running = true
		logger.Log.Info("gRPC server started",
			zap.String("address", listener.Addr().String()))

		// Simulate server running
		for s.grpcServer.running {
			select {
			case <-s.ctx.Done():
				s.grpcServer.running = false
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (s *Server) Stop() error {
	s.cancel()
	s.grpcServer.running = false

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	logger.Log.Info("gRPC server stopped")
	return nil
}

// RegisterService registers a service with the gRPC server
func (s *Server) RegisterService(serviceName string, service interface{}) {
	s.grpcServer.services[serviceName] = service
	logger.Log.Info("gRPC service registered",
		zap.String("service", serviceName))
}

// GetService returns a registered service
func (s *Server) GetService(serviceName string) (interface{}, error) {
	service, exists := s.grpcServer.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	return service, nil
}

// UnaryInterceptor provides logging and tracing for unary RPCs
func (s *Server) UnaryInterceptor() func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Add tracing context
		ctx = context.WithValue(ctx, grpcServerMethodKey, info.FullMethod)

		// Log request
		logger.Log.Debug("gRPC request started",
			zap.String("method", info.FullMethod))

		// Call handler
		resp, err := handler(ctx, req)

		// Log response
		duration := time.Since(start)
		if err != nil {
			logger.Log.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			logger.Log.Debug("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration))
		}

		return resp, err
	}
}

// StreamInterceptor provides logging and tracing for streaming RPCs
func (s *Server) StreamInterceptor() func(srv interface{}, ss StreamServer, info *StreamServerInfo, handler StreamHandler) error {
	return func(srv interface{}, ss StreamServer, info *StreamServerInfo, handler StreamHandler) error {
		start := time.Now()

		// Add tracing context
		ctx := context.WithValue(ss.Context(), grpcServerMethodKey, info.FullMethod)

		// Wrap stream with logging
		loggingStream := &loggingServerStream{
			StreamServer: ss,
			ctx:          ctx,
			start:        start,
		}

		// Log request
		logger.Log.Debug("gRPC stream started",
			zap.String("method", info.FullMethod))

		// Call handler
		err := handler(srv, loggingStream)

		// Log response
		duration := time.Since(start)
		if err != nil {
			logger.Log.Error("gRPC stream failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			logger.Log.Debug("gRPC stream completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration))
		}

		return err
	}
}

// UnaryServerInfo represents unary server info
type UnaryServerInfo struct {
	FullMethod string
	Server     interface{}
}

// StreamServerInfo represents stream server info
type StreamServerInfo struct {
	FullMethod     string
	IsClientStream bool
	IsServerStream bool
}

// UnaryHandler represents a unary handler
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// StreamHandler represents a stream handler
type StreamHandler func(srv interface{}, stream StreamServer) error

// StreamServer represents a stream server
type StreamServer interface {
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

// loggingServerStream wraps a StreamServer with logging
type loggingServerStream struct {
	StreamServer
	ctx   context.Context
	start time.Time
}

// Context returns the context
func (s *loggingServerStream) Context() context.Context {
	return s.ctx
}

// SendMsg logs and forwards the message
func (s *loggingServerStream) SendMsg(m interface{}) error {
	err := s.StreamServer.SendMsg(m)
	if err != nil {
		method := "unknown"
		if methodValue := s.ctx.Value("grpc_method"); methodValue != nil {
			if methodStr, ok := methodValue.(string); ok {
				method = methodStr
			}
		}
		logger.Log.Error("gRPC stream send failed",
			zap.String("method", method),
			zap.Error(err))
	}
	return err
}

// RecvMsg logs and forwards the message
func (s *loggingServerStream) RecvMsg(m interface{}) error {
	err := s.StreamServer.RecvMsg(m)
	if err != nil {
		method := "unknown"
		if methodValue := s.ctx.Value("grpc_method"); methodValue != nil {
			if methodStr, ok := methodValue.(string); ok {
				method = methodStr
			}
		}
		logger.Log.Error("gRPC stream receive failed",
			zap.String("method", method),
			zap.Error(err))
	}
	return err
}
