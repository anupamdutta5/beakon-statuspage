package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// MiddlewareFunc represents a middleware function
type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// Middleware provides middleware functionality for the API Gateway
type Middleware struct {
	config *config.Config
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{
		config: cfg,
	}
}

// AuthRequired middleware checks for valid authentication
func (m *Middleware) AuthRequired() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// Check for Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Token required")
				return
			}

			// Validate token (simplified - in production, use proper JWT validation)
			userID, tenantID, err := m.validateToken(token)
			if err != nil {
				logger.Log.Warn("Invalid token", zap.Error(err))
				m.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add user and tenant context
			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "tenant_id", tenantID)
			r = r.WithContext(ctx)

			// Continue to next handler
			next(w, r)
		}
	}
}

// TenantRequired middleware ensures tenant context is available
func (m *Middleware) TenantRequired() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if tenant ID is in context
			tenantID := r.Context().Value("tenant_id")
			if tenantID == nil {
				m.writeErrorResponse(w, http.StatusBadRequest, "Tenant context required")
				return
			}

			// Validate tenant exists and user has access
			if !m.validateTenantAccess(tenantID.(string), r.Context().Value("user_id").(string)) {
				m.writeErrorResponse(w, http.StatusForbidden, "Access denied to tenant")
				return
			}

			// Continue to next handler
			next(w, r)
		}
	}
}

// RateLimit middleware implements rate limiting
func (m *Middleware) RateLimit() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !m.config.Security.RateLimit.Enabled {
				next(w, r)
				return
			}

			// Get client identifier
			clientID := m.getClientID(r)

			// Check rate limit
			if !m.checkRateLimit(clientID) {
				m.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}

			// Continue to next handler
			next(w, r)
		}
	}
}

// Logging middleware logs requests
func (m *Middleware) Logging() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next(wrapped, r)

			// Log request
			duration := time.Since(start)
			logger.Log.Info("Request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.String("request_id", r.Context().Value("request_id").(string)),
			)
		}
	}
}

// CORS middleware handles CORS headers
func (m *Middleware) CORS() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cors := m.config.Security.CORS

			// Set CORS headers
			if len(cors.AllowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", strings.Join(cors.AllowedOrigins, ", "))
			}
			if len(cors.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cors.AllowedMethods, ", "))
			}
			if len(cors.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cors.AllowedHeaders, ", "))
			}
			if cors.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue to next handler
			next(w, r)
		}
	}
}

// SecurityHeaders middleware adds security headers
func (m *Middleware) SecurityHeaders() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			// Continue to next handler
			next(w, r)
		}
	}
}

// RequestID middleware adds request ID to requests
func (m *Middleware) RequestID() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Generate or extract request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			}

			// Add to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

			// Add to response headers
			w.Header().Set("X-Request-ID", requestID)

			// Continue to next handler
			next(w, r)
		}
	}
}

// Timeout middleware adds request timeout
func (m *Middleware) Timeout(timeout time.Duration) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			go func() {
				next(w, r)
				close(done)
			}()

			select {
			case <-done:
				// Request completed successfully
			case <-ctx.Done():
				// Request timed out
				m.writeErrorResponse(w, http.StatusRequestTimeout, "Request timeout")
			}
		}
	}
}

// validateToken validates a JWT token (simplified implementation)
func (m *Middleware) validateToken(token string) (userID, tenantID string, err error) {
	// In a real implementation, this would:
	// 1. Parse the JWT token
	// 2. Verify the signature
	// 3. Check expiration
	// 4. Extract user and tenant information

	// For now, return mock data
	return "user123", "tenant456", nil
}

// validateTenantAccess validates that a user has access to a tenant
func (m *Middleware) validateTenantAccess(tenantID, userID string) bool {
	// In a real implementation, this would check the database
	// to verify the user has access to the tenant

	// For now, return true
	return true
}

// getClientID extracts a client identifier for rate limiting
func (m *Middleware) getClientID(r *http.Request) string {
	// Try to get from X-Forwarded-For header first
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// checkRateLimit checks if a client has exceeded the rate limit
func (m *Middleware) checkRateLimit(clientID string) bool {
	// In a real implementation, this would use Redis or similar
	// to track request counts per client

	// For now, always return true (no rate limiting)
	return true
}

// writeErrorResponse writes an error response
func (m *Middleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	}

	json.NewEncoder(w).Encode(response)
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

// CircuitBreaker middleware provides circuit breaker functionality
func (m *Middleware) CircuitBreaker(serviceName string) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// In a real implementation, this would check circuit breaker state
			// and prevent requests if the circuit is open

			// For now, always allow requests
			next(w, r)
		}
	}
}

// Metrics middleware collects metrics
func (m *Middleware) Metrics() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next(wrapped, r)

			// Collect metrics
			duration := time.Since(start)

			// In a real implementation, this would send metrics to Prometheus
			logger.Log.Debug("Request metrics",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
			)
		}
	}
}
