package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	"go.uber.org/zap"
)

// PerformanceService handles performance optimizations
type PerformanceService struct {
	logger      *zap.Logger
	minifier    *minify.M
	cache       gcache.Cache
	imageCache  gcache.Cache
	cdnURL      string
	enableCache bool
	enableGzip  bool
	enableMinify bool
	metrics     *PerformanceMetrics
	mu          sync.RWMutex
}

// PerformanceMetrics tracks performance metrics
type PerformanceMetrics struct {
	CacheHits      int64
	CacheMisses    int64
	BytesSaved     int64
	ResponseTimes  []time.Duration
	TotalRequests  int64
	CachedRequests int64
	mu             sync.RWMutex
}

// PerformanceConfig holds performance configuration
type PerformanceConfig struct {
	EnableCache     bool
	EnableGzip      bool
	EnableMinify    bool
	CacheSize       int
	CacheTTL        time.Duration
	CDNEnabled      bool
	CDNURL          string
	ImageOptimization bool
}

// NewPerformanceService creates a new performance service
func NewPerformanceService(logger *zap.Logger, config *PerformanceConfig) *PerformanceService {
	// Initialize minifier
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("application/json", json.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFunc("text/xml", xml.Minify)
	m.AddFunc("application/xml", xml.Minify)

	// Create caches
	cache := gcache.New(config.CacheSize).
		LRU().
		Expiration(config.CacheTTL).
		Build()

	imageCache := gcache.New(100).
		LRU().
		Expiration(24 * time.Hour).
		Build()

	return &PerformanceService{
		logger:       logger,
		minifier:     m,
		cache:        cache,
		imageCache:   imageCache,
		cdnURL:       config.CDNURL,
		enableCache:  config.EnableCache,
		enableGzip:   config.EnableGzip,
		enableMinify: config.EnableMinify,
		metrics:      &PerformanceMetrics{},
	}
}

// CacheMiddleware provides response caching
func (s *PerformanceService) CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip non-GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip if caching is disabled
		if !s.enableCache {
			c.Next()
			return
		}

		// Generate cache key
		cacheKey := s.generateCacheKey(c.Request)

		// Check cache
		if cached, err := s.cache.Get(cacheKey); err == nil {
			s.recordCacheHit()
			response := cached.(*CachedResponse)

			// Set headers
			for key, values := range response.Headers {
				for _, value := range values {
					c.Header(key, value)
				}
			}

			// Set cache headers
			c.Header("X-Cache", "HIT")
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(ttl.Seconds())))

			// Write response
			c.Data(response.StatusCode, response.ContentType, response.Body)
			c.Abort()
			return
		}

		s.recordCacheMiss()

		// Create response writer wrapper
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Cache successful responses
		if writer.statusCode >= 200 && writer.statusCode < 300 {
			response := &CachedResponse{
				StatusCode:  writer.statusCode,
				Headers:     writer.Header(),
				Body:        writer.body.Bytes(),
				ContentType: writer.Header().Get("Content-Type"),
			}

			s.cache.SetWithExpire(cacheKey, response, ttl)
			c.Header("X-Cache", "MISS")
		}
	}
}

// GzipMiddleware provides gzip compression
func (s *PerformanceService) GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.enableGzip {
			c.Next()
			return
		}

		// Check if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Wrap response writer
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()

		// Record bytes saved
		if gw, ok := c.Writer.(*gzipWriter); ok {
			s.recordBytesSaved(gw.originalSize - gw.compressedSize)
		}
	}
}

// MinifyMiddleware provides HTML/CSS/JS minification
func (s *PerformanceService) MinifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.enableMinify {
			c.Next()
			return
		}

		// Create response writer wrapper
		writer := &minifyWriter{
			ResponseWriter: c.Writer,
			minifier:       s.minifier,
			buffer:         &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// Minify response if applicable
		contentType := c.Writer.Header().Get("Content-Type")
		if s.shouldMinify(contentType) {
			minified, err := s.minifier.Bytes(contentType, writer.buffer.Bytes())
			if err == nil {
				saved := len(writer.buffer.Bytes()) - len(minified)
				s.recordBytesSaved(int64(saved))
				c.Writer.Write(minified)
			} else {
				c.Writer.Write(writer.buffer.Bytes())
			}
		}
	}
}

// ETaGMiddleware provides ETag support
func (s *PerformanceService) ETaGMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip non-GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Create response writer wrapper
		writer := &etagWriter{
			ResponseWriter: c.Writer,
			buffer:         &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// Generate ETag
		if writer.statusCode >= 200 && writer.statusCode < 300 {
			etag := s.generateETag(writer.buffer.Bytes())
			c.Header("ETag", etag)

			// Check If-None-Match
			if c.GetHeader("If-None-Match") == etag {
				c.Status(http.StatusNotModified)
				return
			}

			// Write response
			c.Writer.Write(writer.buffer.Bytes())
		}
	}
}

// PreloadHeaders adds resource hints for faster loading
func (s *PerformanceService) PreloadHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add preload headers for critical resources
		c.Header("Link", "</static/css/modern-landing.css>; rel=preload; as=style")
		c.Header("Link", "</static/js/modern-landing.js>; rel=preload; as=script")

		// DNS prefetch for external domains
		c.Header("Link", "<https://fonts.googleapis.com>; rel=dns-prefetch")
		c.Header("Link", "<https://cdnjs.cloudflare.com>; rel=dns-prefetch")

		c.Next()
	}
}

// SecurityHeaders adds security headers
func (s *PerformanceService) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self' https://api.analytics.com"
		c.Header("Content-Security-Policy", csp)

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// OptimizeImage optimizes images on the fly
func (s *PerformanceService) OptimizeImage(ctx context.Context, imagePath string, quality int) ([]byte, error) {
	// Check image cache first
	cacheKey := fmt.Sprintf("%s:%d", imagePath, quality)
	if cached, err := s.imageCache.Get(cacheKey); err == nil {
		return cached.([]byte), nil
	}

	// Read original image
	// In production, use image processing library like imaging or vips
	// For now, return a placeholder implementation

	s.logger.Info("Optimizing image",
		zap.String("path", imagePath),
		zap.Int("quality", quality))

	// Cache result
	// s.imageCache.Set(cacheKey, optimizedImage)

	return nil, fmt.Errorf("image optimization not implemented")
}

// GenerateSrcSet generates responsive image srcset
func (s *PerformanceService) GenerateSrcSet(imagePath string, sizes []int) string {
	var srcset []string

	for _, size := range sizes {
		// Generate resized image URL
		url := fmt.Sprintf("%s?w=%d %dw", s.getCDNURL(imagePath), size, size)
		srcset = append(srcset, url)
	}

	return strings.Join(srcset, ", ")
}

// GetMetrics returns performance metrics
func (s *PerformanceService) GetMetrics() map[string]interface{} {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	var avgResponseTime time.Duration
	if len(s.metrics.ResponseTimes) > 0 {
		var total time.Duration
		for _, t := range s.metrics.ResponseTimes {
			total += t
		}
		avgResponseTime = total / time.Duration(len(s.metrics.ResponseTimes))
	}

	cacheHitRate := float64(0)
	if s.metrics.CacheHits+s.metrics.CacheMisses > 0 {
		cacheHitRate = float64(s.metrics.CacheHits) / float64(s.metrics.CacheHits+s.metrics.CacheMisses) * 100
	}

	return map[string]interface{}{
		"cache_hits":        s.metrics.CacheHits,
		"cache_misses":      s.metrics.CacheMisses,
		"cache_hit_rate":    fmt.Sprintf("%.2f%%", cacheHitRate),
		"bytes_saved":       s.humanReadableSize(s.metrics.BytesSaved),
		"avg_response_time": avgResponseTime.String(),
		"total_requests":    s.metrics.TotalRequests,
		"cached_requests":   s.metrics.CachedRequests,
	}
}

// Helper methods

func (s *PerformanceService) generateCacheKey(r *http.Request) string {
	// Include query parameters in cache key
	return fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, r.URL.RawQuery)
}

func (s *PerformanceService) generateETag(content []byte) string {
	hash := md5.Sum(content)
	return `"` + hex.EncodeToString(hash[:]) + `"`
}

func (s *PerformanceService) shouldMinify(contentType string) bool {
	minifiableTypes := []string{
		"text/html",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"image/svg+xml",
	}

	for _, t := range minifiableTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

func (s *PerformanceService) getCDNURL(path string) string {
	if s.cdnURL != "" {
		return s.cdnURL + path
	}
	return path
}

func (s *PerformanceService) recordCacheHit() {
	s.metrics.mu.Lock()
	s.metrics.CacheHits++
	s.metrics.mu.Unlock()
}

func (s *PerformanceService) recordCacheMiss() {
	s.metrics.mu.Lock()
	s.metrics.CacheMisses++
	s.metrics.mu.Unlock()
}

func (s *PerformanceService) recordBytesSaved(bytes int64) {
	s.metrics.mu.Lock()
	s.metrics.BytesSaved += bytes
	s.metrics.mu.Unlock()
}

func (s *PerformanceService) humanReadableSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Response writer wrappers

type CachedResponse struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ContentType string
}

type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

type gzipWriter struct {
	gin.ResponseWriter
	writer         *gzip.Writer
	originalSize   int64
	compressedSize int64
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	w.originalSize += int64(len(b))
	n, err := w.writer.Write(b)
	w.compressedSize += int64(n)
	return n, err
}

type minifyWriter struct {
	gin.ResponseWriter
	minifier *minify.M
	buffer   *bytes.Buffer
}

func (w *minifyWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

type etagWriter struct {
	gin.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (w *etagWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w *etagWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// LazyLoadImages adds lazy loading attributes to images
func (s *PerformanceService) LazyLoadImages(html string) string {
	// Simple implementation - in production use a proper HTML parser
	html = strings.ReplaceAll(html, "<img ", `<img loading="lazy" `)
	return html
}

// GenerateCSPNonce generates a nonce for Content Security Policy
func (s *PerformanceService) GenerateCSPNonce() string {
	// Generate random nonce
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// PreconnectDomains returns domains to preconnect
func (s *PerformanceService) PreconnectDomains() []string {
	return []string{
		"https://fonts.googleapis.com",
		"https://fonts.gstatic.com",
		"https://cdnjs.cloudflare.com",
		"https://www.google-analytics.com",
	}
}

// CriticalCSS extracts critical CSS for above-the-fold content
func (s *PerformanceService) CriticalCSS(ctx context.Context, pageURL string) (string, error) {
	// This would typically use a tool like critical or penthouse
	// For now, return a placeholder
	return "", fmt.Errorf("critical CSS extraction not implemented")
}

// GenerateWebPImage converts images to WebP format
func (s *PerformanceService) GenerateWebPImage(ctx context.Context, imagePath string) (string, error) {
	// This would use an image processing library
	// For now, return a placeholder
	webpPath := strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + ".webp"
	return webpPath, fmt.Errorf("WebP generation not implemented")
}

// ResourceHints generates resource hints for faster page loads
func (s *PerformanceService) ResourceHints(pageType string) []ResourceHint {
	hints := []ResourceHint{
		{
			Rel:  "preload",
			Href: "/static/css/modern-landing.css",
			As:   "style",
		},
		{
			Rel:  "preload",
			Href: "/static/js/modern-landing.js",
			As:   "script",
		},
		{
			Rel:  "preconnect",
			Href: "https://fonts.googleapis.com",
		},
		{
			Rel:  "dns-prefetch",
			Href: "https://www.google-analytics.com",
		},
	}

	// Add page-specific hints
	switch pageType {
	case "home":
		hints = append(hints, ResourceHint{
			Rel:  "preload",
			Href: "/static/images/hero-bg.jpg",
			As:   "image",
		})
	case "pricing":
		hints = append(hints, ResourceHint{
			Rel:  "prefetch",
			Href: "/api/v1/pricing",
		})
	}

	return hints
}

// ResourceHint represents a resource hint
type ResourceHint struct {
	Rel        string
	Href       string
	As         string
	Type       string
	Crossorigin string
}

// BundleAssets combines and minifies CSS/JS files
func (s *PerformanceService) BundleAssets(files []string, assetType string) (string, error) {
	var combined bytes.Buffer

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		combined.Write(content)
		combined.WriteString("\n")
	}

	// Minify combined content
	contentType := "text/css"
	if assetType == "js" {
		contentType = "text/javascript"
	}

	minified, err := s.minifier.Bytes(contentType, combined.Bytes())
	if err != nil {
		return "", err
	}

	// Generate hash for cache busting
	hash := md5.Sum(minified)
	filename := fmt.Sprintf("bundle.%s.%s", hex.EncodeToString(hash[:])[:8], assetType)

	return filename, nil
}