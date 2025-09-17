package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// DomainService handles domain resolution and routing for custom domains
type DomainService struct {
	logger       *zap.Logger
	redisClient  *redis.Client
	tokenCache   map[string]string
	cacheMutex   sync.RWMutex
	config       *DomainConfig
}

// DomainConfig contains configuration for the domain service
type DomainConfig struct {
	RedisURL          string
	CacheTTL          time.Duration
	DefaultTenantID   uint
	DefaultStatusPage string
}

// DomainInfo contains information about a custom domain
type DomainInfo struct {
	TenantID   uint   `json:"tenant_id"`
	StatusPage string `json:"status_page"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
}

// NewDomainService creates a new domain service
func NewDomainService(logger *zap.Logger, config *DomainConfig) (*DomainService, error) {
	// Initialize Redis client
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	service := &DomainService{
		logger:      logger,
		redisClient: redis.NewClient(opt),
		tokenCache:  make(map[string]string),
		config:      config,
	}

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := service.redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return service, nil
}

// ResolveDomain resolves a custom domain to a tenant ID and status page slug
func (s *DomainService) ResolveDomain(ctx context.Context, hostname string) (uint, string, error) {
	// First check in-memory cache
	s.cacheMutex.RLock()
	if cached, ok := s.tokenCache[hostname]; ok {
		s.cacheMutex.RUnlock()
		// Parse tenantID and status page from cache (format: "tenantID:statusPage:token")
		parts := strings.SplitN(cached, ":", 3)
		if len(parts) >= 2 {
			if tenantID := parseTenantID(parts[0]); tenantID > 0 {
				return tenantID, parts[1], nil
			}
		}
		return s.config.DefaultTenantID, s.config.DefaultStatusPage, nil
	}
	s.cacheMutex.RUnlock()

	// Check Redis cache
	cacheKey := fmt.Sprintf("domain:%s", hostname)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		// Parse tenantID and status page from cached value (format: "tenantID:statusPage:token")
		parts := strings.SplitN(cached, ":", 3)
		if len(parts) >= 2 {
			// Add to in-memory cache
			s.cacheMutex.Lock()
			s.tokenCache[hostname] = cached
			s.cacheMutex.Unlock()

			// Parse tenantID
			if tenantIDStr := parts[0]; tenantIDStr != "" {
				if tenantID := parseTenantID(tenantIDStr); tenantID > 0 {
					return tenantID, parts[1], nil
				}
			}
		}
		return s.config.DefaultTenantID, s.config.DefaultStatusPage, nil
	}

	// If not in cache, query the tenant admin service
	// TODO: Implement tenant admin service lookup
	// For now, return defaults
	return s.config.DefaultTenantID, s.config.DefaultStatusPage, nil
}

// VerifyDomainOwnership verifies domain ownership by checking DNS records
func (s *DomainService) VerifyDomainOwnership(ctx context.Context, domain string, verificationToken string) (bool, error) {
	// TODO: Implement DNS verification
	// This would typically check for a TXT record with the verification token
	return true, nil
}

// AddCustomDomain adds a new custom domain for a status page
func (s *DomainService) AddCustomDomain(ctx context.Context, tenantID uint, statusPage, domain string) (string, error) {
	// Generate verification token
	verificationToken := generateVerificationToken()

	// Store in Redis with TTL
	cacheKey := fmt.Sprintf("domain:verify:%s", domain)
	err := s.redisClient.Set(ctx, cacheKey, fmt.Sprintf("%d:%s:%s", tenantID, statusPage, verificationToken), 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store verification token: %w", err)
	}

	return verificationToken, nil
}

// CompleteDomainVerification completes the domain verification process
func (s *DomainService) CompleteDomainVerification(ctx context.Context, domain string) error {
	// Look up verification record
	cacheKey := fmt.Sprintf("domain:verify:%s", domain)
	val, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("verification token not found or expired")
		}
		return fmt.Errorf("failed to get verification token: %w", err)
	}

	// Parse the stored value
	parts := strings.SplitN(val, ":", 3)
	if len(parts) != 3 {
		return errors.New("invalid verification token format")
	}

	// TODO: Store the verified domain in the tenant admin service

	// Add to cache
	domainKey := fmt.Sprintf("domain:%s", domain)
	err = s.redisClient.Set(ctx, domainKey, val, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to cache domain mapping: %w", err)
	}

	// Update in-memory cache
	s.cacheMutex.Lock()
	s.tokenCache[domain] = val
	s.cacheMutex.Unlock()

	// Clean up verification token
	s.redisClient.Del(ctx, cacheKey)

	return nil
}

// generateVerificationToken generates a random verification token
func generateVerificationToken() string {
	// Generate a more secure verification token using current timestamp and random component
	timestamp := time.Now().Unix()
	return fmt.Sprintf("beakon_verify_%d_%d", timestamp, timestamp%10000)
}

// Middleware returns a middleware function that resolves custom domains
func (s *DomainService) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for health checks and internal routes
			if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/internal") {
				next.ServeHTTP(w, r)
				return
			}

			// Get the hostname
			hostname := r.Host
			if hostname == "" {
				http.Error(w, "Host header required", http.StatusBadRequest)
				return
			}

			// Resolve the domain
			ctx := r.Context()
			tenantID, statusPage, err := s.ResolveDomain(ctx, hostname)
			if err != nil {
				s.logger.Error("Failed to resolve domain", zap.String("hostname", hostname), zap.Error(err))
				http.Error(w, "Domain not found", http.StatusNotFound)
				return
			}

			// Add tenant ID and status page to the request context
			ctx = context.WithValue(ctx, "tenantID", tenantID)
			ctx = context.WithValue(ctx, "statusPage", statusPage)

			// Update the request context
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// parseTenantID safely parses a tenant ID string to uint
func parseTenantID(tenantIDStr string) uint {
	if id, err := strconv.ParseUint(tenantIDStr, 10, 32); err == nil {
		return uint(id)
	}
	return 0
}
