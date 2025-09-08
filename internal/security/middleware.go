package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Context keys for security middleware
type contextKey string

const (
	userIDKey     contextKey = "user_id"
	tenantIDKey   contextKey = "tenant_id"
	emailKey      contextKey = "email"
	rolesKey      contextKey = "roles"
	scopesKey     contextKey = "scopes"
	sessionIDKey  contextKey = "session_id"
	clientCertKey contextKey = "client_cert"
	clientIDKey   contextKey = "client_id"
)

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWTConfig     JWTConfig               `json:"jwt_config"`
	OAuth2Configs map[string]OAuth2Config `json:"oauth2_configs"`
	MTLSConfig    mTLSConfig              `json:"mtls_config"`
	RateLimit     RateLimitConfig         `json:"rate_limit"`
	CORS          CORSConfig              `json:"cors"`
	Headers       HeadersConfig           `json:"headers"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	Requests    int           `json:"requests"`
	Window      time.Duration `json:"window"`
	Burst       int           `json:"burst"`
	SkipSuccess bool          `json:"skip_success"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// HeadersConfig represents security headers configuration
type HeadersConfig struct {
	Enabled                 bool   `json:"enabled"`
	ContentTypeNosniff      bool   `json:"content_type_nosniff"`
	FrameOptions            string `json:"frame_options"`
	XSSProtection           string `json:"xss_protection"`
	ContentSecurityPolicy   string `json:"content_security_policy"`
	ReferrerPolicy          string `json:"referrer_policy"`
	PermissionsPolicy       string `json:"permissions_policy"`
	StrictTransportSecurity string `json:"strict_transport_security"`
}

// SecurityMiddleware provides comprehensive security middleware
type SecurityMiddleware struct {
	config        SecurityConfig
	jwtManager    *JWTManager
	oauth2Manager *OAuth2Manager
	mTLSManager   *mTLSManager
	rateLimiter   *RateLimiter
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig) (*SecurityMiddleware, error) {
	// Initialize JWT manager
	jwtManager, err := NewJWTManager(config.JWTConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT manager: %w", err)
	}

	// Initialize OAuth2 manager
	oauth2Manager := NewOAuth2Manager()
	for name, oauth2Config := range config.OAuth2Configs {
		oauth2Manager.RegisterProvider(name, oauth2Config)
	}

	// Initialize mTLS manager
	mTLSManager, err := NewmTLSManager(config.MTLSConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mTLS manager: %w", err)
	}

	// Initialize rate limiter
	rateLimiter := NewRateLimiter(config.RateLimit)

	return &SecurityMiddleware{
		config:        config,
		jwtManager:    jwtManager,
		oauth2Manager: oauth2Manager,
		mTLSManager:   mTLSManager,
		rateLimiter:   rateLimiter,
	}, nil
}

// JWTAuthMiddleware provides JWT authentication middleware
func (sm *SecurityMiddleware) JWTAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			claims, err := sm.jwtManager.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, emailKey, claims.Email)
			ctx = context.WithValue(ctx, rolesKey, claims.Roles)
			ctx = context.WithValue(ctx, scopesKey, claims.Scopes)
			ctx = context.WithValue(ctx, sessionIDKey, claims.SessionID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// RoleAuthMiddleware provides role-based authorization middleware
func (sm *SecurityMiddleware) RoleAuthMiddleware(requiredRoles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get roles from context
			roles, ok := r.Context().Value(rolesKey).([]string)
			if !ok {
				http.Error(w, "No roles found in context", http.StatusUnauthorized)
				return
			}

			// Check if user has required role
			hasRole := false
			for _, requiredRole := range requiredRoles {
				for _, role := range roles {
					if role == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ScopeAuthMiddleware provides scope-based authorization middleware
func (sm *SecurityMiddleware) ScopeAuthMiddleware(requiredScopes ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get scopes from context
			scopes, ok := r.Context().Value(scopesKey).([]string)
			if !ok {
				http.Error(w, "No scopes found in context", http.StatusUnauthorized)
				return
			}

			// Check if user has required scope
			hasScope := false
			for _, requiredScope := range requiredScopes {
				for _, scope := range scopes {
					if scope == requiredScope {
						hasScope = true
						break
					}
				}
				if hasScope {
					break
				}
			}

			if !hasScope {
				http.Error(w, "Insufficient scopes", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantAuthMiddleware provides tenant-based authorization middleware
func (sm *SecurityMiddleware) TenantAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get tenant ID from context
			tenantID, ok := r.Context().Value(tenantIDKey).(string)
			if !ok || tenantID == "" {
				http.Error(w, "Tenant ID required", http.StatusUnauthorized)
				return
			}

			// Get tenant ID from URL path
			pathTenantID := r.URL.Path
			if strings.HasPrefix(pathTenantID, "/tenants/") {
				parts := strings.Split(pathTenantID, "/")
				if len(parts) >= 3 {
					pathTenantID = parts[2]
				}
			}

			// Check if user has access to the requested tenant
			if pathTenantID != "" && pathTenantID != tenantID {
				http.Error(w, "Access denied to tenant", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware provides rate limiting middleware
func (sm *SecurityMiddleware) RateLimitMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !sm.config.RateLimit.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get client identifier
			clientID := sm.getClientID(r)

			// Check rate limit
			allowed, err := sm.rateLimiter.Allow(clientID)
			if err != nil {
				http.Error(w, "Rate limit error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware provides CORS middleware
func (sm *SecurityMiddleware) CORSMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !sm.config.CORS.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if sm.isOriginAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set CORS headers
			if len(sm.config.CORS.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(sm.config.CORS.AllowedMethods, ", "))
			}

			if len(sm.config.CORS.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(sm.config.CORS.AllowedHeaders, ", "))
			}

			if len(sm.config.CORS.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(sm.config.CORS.ExposedHeaders, ", "))
			}

			if sm.config.CORS.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if sm.config.CORS.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", sm.config.CORS.MaxAge))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware provides security headers middleware
func (sm *SecurityMiddleware) SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !sm.config.Headers.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Set security headers
			if sm.config.Headers.ContentTypeNosniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			if sm.config.Headers.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", sm.config.Headers.FrameOptions)
			}

			if sm.config.Headers.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", sm.config.Headers.XSSProtection)
			}

			if sm.config.Headers.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", sm.config.Headers.ContentSecurityPolicy)
			}

			if sm.config.Headers.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", sm.config.Headers.ReferrerPolicy)
			}

			if sm.config.Headers.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", sm.config.Headers.PermissionsPolicy)
			}

			if sm.config.Headers.StrictTransportSecurity != "" {
				w.Header().Set("Strict-Transport-Security", sm.config.Headers.StrictTransportSecurity)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions
func (sm *SecurityMiddleware) getClientID(r *http.Request) string {
	// Try to get client ID from various sources
	if clientID := r.Header.Get("X-Client-ID"); clientID != "" {
		return clientID
	}

	if userID := r.Context().Value(userIDKey); userID != nil {
		return fmt.Sprintf("%v", userID)
	}

	// Fall back to IP address
	return r.RemoteAddr
}

func (sm *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range sm.config.CORS.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}

	return false
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	config RateLimitConfig
	// In a real implementation, you would use Redis or in-memory store
	// For now, we'll use a simple map
	clients map[string]*ClientLimit
}

// ClientLimit represents rate limit for a client
type ClientLimit struct {
	Requests []time.Time
	LastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config:  config,
		clients: make(map[string]*ClientLimit),
	}
}

// Allow checks if a request is allowed for a client
func (rl *RateLimiter) Allow(clientID string) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-rl.config.Window)

	// Get or create client limit
	clientLimit, exists := rl.clients[clientID]
	if !exists {
		clientLimit = &ClientLimit{
			Requests: make([]time.Time, 0),
			LastSeen: now,
		}
		rl.clients[clientID] = clientLimit
	}

	// Clean up old requests
	var validRequests []time.Time
	for _, requestTime := range clientLimit.Requests {
		if requestTime.After(windowStart) {
			validRequests = append(validRequests, requestTime)
		}
	}
	clientLimit.Requests = validRequests

	// Check if limit is exceeded
	if len(clientLimit.Requests) >= rl.config.Requests {
		return false, nil
	}

	// Add current request
	clientLimit.Requests = append(clientLimit.Requests, now)
	clientLimit.LastSeen = now

	return true, nil
}

// Global security middleware instance
var globalSecurityMiddleware *SecurityMiddleware

// InitGlobalSecurity initializes the global security middleware
func InitGlobalSecurity(config SecurityConfig) error {
	middleware, err := NewSecurityMiddleware(config)
	if err != nil {
		return err
	}
	globalSecurityMiddleware = middleware
	return nil
}

// GetGlobalSecurity returns the global security middleware
func GetGlobalSecurity() *SecurityMiddleware {
	return globalSecurityMiddleware
}

// JWTAuthMiddleware is a convenience function for JWT authentication
func JWTAuthMiddleware() func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.JWTAuthMiddleware()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}

// RoleAuthMiddleware is a convenience function for role-based authorization
func RoleAuthMiddleware(requiredRoles ...string) func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.RoleAuthMiddleware(requiredRoles...)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}

// ScopeAuthMiddleware is a convenience function for scope-based authorization
func ScopeAuthMiddleware(requiredScopes ...string) func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.ScopeAuthMiddleware(requiredScopes...)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}

// TenantAuthMiddleware is a convenience function for tenant-based authorization
func TenantAuthMiddleware() func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.TenantAuthMiddleware()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}

// RateLimitMiddleware is a convenience function for rate limiting
func RateLimitMiddleware() func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.RateLimitMiddleware()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}

// CORSMiddleware is a convenience function for CORS
func CORSMiddleware() func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.CORSMiddleware()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}

// SecurityHeadersMiddleware is a convenience function for security headers
func SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	if globalSecurityMiddleware != nil {
		return globalSecurityMiddleware.SecurityHeadersMiddleware()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Security middleware not initialized", http.StatusInternalServerError)
		})
	}
}
