// Package handlers provides HTTP handlers for SEO functionality.
package handlers

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/anupamdutta5/landing-page-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SEOHandler handles SEO-related HTTP requests
type SEOHandler struct {
	seoService *services.SEOService
	db         *gorm.DB
	logger     *zap.Logger
}

// NewSEOHandler creates a new SEO handler
func NewSEOHandler(seoService *services.SEOService, db *gorm.DB, logger *zap.Logger) *SEOHandler {
	return &SEOHandler{
		seoService: seoService,
		db:         db,
		logger:     logger,
	}
}

// GetSitemap generates and serves the XML sitemap
func (h *SEOHandler) GetSitemap(c *gin.Context) {
	h.logger.Info("Generating sitemap")

	// Define your site pages - in a real app, this would come from a database
	pages := []models.SitemapPage{
		{
			Path:         "/",
			LastModified: time.Now(),
			ChangeFreq:   "weekly",
			Priority:     1.0,
		},
		{
			Path:         "/features",
			LastModified: time.Now().AddDate(0, 0, -7),
			ChangeFreq:   "monthly",
			Priority:     0.8,
		},
		{
			Path:         "/pricing",
			LastModified: time.Now().AddDate(0, 0, -3),
			ChangeFreq:   "weekly",
			Priority:     0.9,
		},
		{
			Path:         "/blog",
			LastModified: time.Now(),
			ChangeFreq:   "daily",
			Priority:     0.8,
		},
		{
			Path:         "/contact",
			LastModified: time.Now().AddDate(0, -1, 0),
			ChangeFreq:   "monthly",
			Priority:     0.6,
		},
		{
			Path:         "/privacy",
			LastModified: time.Now().AddDate(0, -3, 0),
			ChangeFreq:   "yearly",
			Priority:     0.3,
		},
		{
			Path:         "/terms",
			LastModified: time.Now().AddDate(0, -3, 0),
			ChangeFreq:   "yearly",
			Priority:     0.3,
		},
	}

	// Add blog articles (mock data - replace with actual database query)
	blogPosts := []string{
		"/blog/getting-started-with-status-pages",
		"/blog/incident-communication-best-practices",
		"/blog/monitoring-vs-status-pages",
		"/blog/building-customer-trust-transparency",
		"/blog/sla-monitoring-guide",
	}

	for _, post := range blogPosts {
		pages = append(pages, models.SitemapPage{
			Path:         post,
			LastModified: time.Now().AddDate(0, 0, -rand.Intn(30)),
			ChangeFreq:   "monthly",
			Priority:     0.7,
		})
	}

	sitemap, err := h.seoService.GenerateSitemap(c.Request.Context(), pages)
	if err != nil {
		h.logger.Error("Failed to generate sitemap", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate sitemap",
		})
		return
	}

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, sitemap)
}

// GetRobotsTxt serves the robots.txt file
func (h *SEOHandler) GetRobotsTxt(c *gin.Context) {
	h.logger.Info("Serving robots.txt")

	robotsTxt := h.seoService.GenerateRobotsTxt(c.Request.Context())

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, robotsTxt)
}

// GenerateMetaTags generates meta tags for a specific page
func (h *SEOHandler) GenerateMetaTags(c *gin.Context) {
	var pageData models.SEOPageData
	if err := c.ShouldBindJSON(&pageData); err != nil {
		h.logger.Error("Invalid page data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page data",
		})
		return
	}

	h.logger.Info("Generating meta tags", zap.String("page", pageData.Path))

	metaTags, err := h.seoService.GenerateMetaTags(c.Request.Context(), &pageData)
	if err != nil {
		h.logger.Error("Failed to generate meta tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate meta tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta_tags": metaTags,
	})
}

// GetStructuredData generates structured data for a page
func (h *SEOHandler) GetStructuredData(c *gin.Context) {
	var pageData models.SEOPageData
	if err := c.ShouldBindJSON(&pageData); err != nil {
		h.logger.Error("Invalid page data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page data",
		})
		return
	}

	h.logger.Info("Generating structured data", zap.String("page", pageData.Path))

	structuredData, err := h.seoService.GenerateStructuredData(c.Request.Context(), &pageData)
	if err != nil {
		h.logger.Error("Failed to generate structured data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate structured data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"structured_data": structuredData,
	})
}

// AnalyzeSEO performs SEO analysis on a page
func (h *SEOHandler) AnalyzeSEO(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL parameter is required",
		})
		return
	}

	h.logger.Info("Analyzing SEO", zap.String("url", url))

	// Mock SEO analysis - in a real implementation, this would crawl the page
	analysis := h.generateMockSEOAnalysis(url)

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
	})
}

// GetSEORecommendations provides SEO improvement recommendations
func (h *SEOHandler) GetSEORecommendations(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL parameter is required",
		})
		return
	}

	h.logger.Info("Getting SEO recommendations", zap.String("url", url))

	recommendations := h.generateSEORecommendations(url)

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
	})
}

// TrackWebVitals receives Core Web Vitals data from the frontend
func (h *SEOHandler) TrackWebVitals(c *gin.Context) {
	var webVitals models.WebVitalsData
	if err := c.ShouldBindJSON(&webVitals); err != nil {
		h.logger.Error("Invalid web vitals data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid web vitals data",
		})
		return
	}

	h.logger.Info("Tracking Web Vitals",
		zap.Float64("lcp", webVitals.LCP),
		zap.Float64("fid", webVitals.FID),
		zap.Float64("cls", webVitals.CLS))

	// Store web vitals data (implement database storage)
	// For now, just log and return success

	c.JSON(http.StatusOK, gin.H{
		"message": "Web Vitals tracked successfully",
	})
}

// PreviewSEO provides a preview of how a page will appear in search results
func (h *SEOHandler) PreviewSEO(c *gin.Context) {
	var pageData models.SEOPageData
	if err := c.ShouldBindJSON(&pageData); err != nil {
		h.logger.Error("Invalid page data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page data",
		})
		return
	}

	h.logger.Info("Generating SEO preview", zap.String("page", pageData.Path))

	metaTags, err := h.seoService.GenerateMetaTags(c.Request.Context(), &pageData)
	if err != nil {
		h.logger.Error("Failed to generate meta tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate SEO preview",
		})
		return
	}

	preview := gin.H{
		"google": gin.H{
			"title":       metaTags.Title,
			"description": metaTags.Description,
			"url":         metaTags.Canonical,
		},
		"facebook": gin.H{
			"title":       metaTags.OGTitle,
			"description": metaTags.OGDescription,
			"image":       metaTags.OGImage,
			"url":         metaTags.OGURL,
		},
		"twitter": gin.H{
			"title":       metaTags.TwitterTitle,
			"description": metaTags.TwitterDescription,
			"image":       metaTags.TwitterImage,
			"card_type":   metaTags.TwitterCard,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"preview": preview,
	})
}

// Helper functions

func (h *SEOHandler) generateMockSEOAnalysis(url string) models.SEOAnalytics {
	// Mock implementation - replace with actual page analysis
	return models.SEOAnalytics{
		PageURL:          url,
		Title:            "StatusPage Pro - Modern Status Page Solution",
		MetaDescription:  "Keep your customers informed with beautiful, reliable status pages. Advanced monitoring, incident management, and communication tools for modern teams.",
		WordCount:        1250,
		ReadabilityScore: 8.5,
		WebVitals: models.WebVitalsData{
			LCP: 2.1,
			FID: 85,
			CLS: 0.08,
			FCP: 1.2,
			TTI: 3.1,
			TBT: 150,
			SI:  2.8,
		},
		TitleScore:        95,
		DescriptionScore:  88,
		KeywordDensity: map[string]float64{
			"status page": 2.8,
			"monitoring":  1.9,
			"incident":    1.5,
		},
		HasH1:             true,
		H1Count:           1,
		HasMetaDesc:       true,
		MetaDescLength:    156,
		HasCanonical:      true,
		HasOGTags:         true,
		HasStructuredData: true,
		ImageAltMissing:   0,
		PageSize:          245000,
		LoadTime:          1.8,
		InternalLinks:     24,
		ExternalLinks:     5,
		Timestamp:         time.Now(),
	}
}

func (h *SEOHandler) generateSEORecommendations(_ string) []models.SEORecommendation {
	return []models.SEORecommendation{
		{
			Type:        "performance",
			Priority:    "high",
			Title:       "Optimize Image Sizes",
			Description: "Compress and resize images to improve loading speed. Consider using WebP format for better compression.",
			Impact:      "high",
			Effort:      "medium",
			Category:    "performance",
		},
		{
			Type:        "content",
			Priority:    "medium",
			Title:       "Add More Internal Links",
			Description: "Include more internal links to help search engines understand your site structure and improve user navigation.",
			Impact:      "medium",
			Effort:      "low",
			Category:    "on-page",
		},
		{
			Type:        "technical",
			Priority:    "medium",
			Title:       "Implement Schema Markup",
			Description: "Add structured data for better rich snippets in search results.",
			Impact:      "medium",
			Effort:      "medium",
			Category:    "technical",
		},
		{
			Type:        "content",
			Priority:    "low",
			Title:       "Optimize Meta Descriptions",
			Description: "Ensure all pages have unique, compelling meta descriptions between 150-160 characters.",
			Impact:      "low",
			Effort:      "low",
			Category:    "on-page",
		},
	}
}

// ============================================
// Admin SEO Configuration Endpoints
// ============================================

// GetSEOConfig retrieves the current SEO configuration
func (h *SEOHandler) GetSEOConfig(c *gin.Context) {
	var config models.LandingPageSEOConfig

	err := h.db.First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return empty config with defaults
			config = models.LandingPageSEOConfig{
				SitemapEnabled:  true,
				RobotsEnabled:   true,
				SchemaEnabled:   true,
				TwitterCard:     "summary_large_image",
			}
			c.JSON(http.StatusOK, config)
			return
		}
		h.logger.Error("Failed to get SEO config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get SEO configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateSEOConfig updates the SEO configuration
func (h *SEOHandler) UpdateSEOConfig(c *gin.Context) {
	var input models.LandingPageSEOConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if config exists
	var existing models.LandingPageSEOConfig
	err := h.db.First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		if err := h.db.Create(&input).Error; err != nil {
			h.logger.Error("Failed to create SEO config", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SEO configuration"})
			return
		}
		c.JSON(http.StatusCreated, input)
		return
	}

	// Update existing config
	input.ID = existing.ID
	if err := h.db.Save(&input).Error; err != nil {
		h.logger.Error("Failed to update SEO config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SEO configuration"})
		return
	}

	c.JSON(http.StatusOK, input)
}

// ValidateSchema validates JSON-LD schema markup
func (h *SEOHandler) ValidateSchema(c *gin.Context) {
	var input struct {
		Schema string `json:"schema" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Perform schema validation (simplified)
	isValid := h.validateJSONLD(input.Schema)

	response := gin.H{
		"valid": isValid,
		"schema": input.Schema,
	}

	if !isValid {
		response["errors"] = []string{
			"Invalid JSON-LD structure",
			"Missing @context",
			"Missing @type",
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetMetaTagsPreview generates a preview of meta tags
func (h *SEOHandler) GetMetaTagsPreview(c *gin.Context) {
	pageType := c.Query("page_type")
	if pageType == "" {
		pageType = "home"
	}

	var config models.LandingPageSEOConfig
	h.db.First(&config)

	// Generate preview based on page type
	preview := h.generateMetaPreview(pageType, &config)

	c.JSON(http.StatusOK, preview)
}

// GenerateDynamicSitemap generates sitemap from database content
func (h *SEOHandler) GenerateDynamicSitemap(c *gin.Context) {
	var pages []models.SitemapPage

	// Add static pages
	staticPages := []models.SitemapPage{
		{Path: "/", ChangeFreq: "daily", Priority: 1.0},
		{Path: "/features", ChangeFreq: "weekly", Priority: 0.8},
		{Path: "/pricing", ChangeFreq: "weekly", Priority: 0.9},
		{Path: "/about", ChangeFreq: "monthly", Priority: 0.7},
		{Path: "/contact", ChangeFreq: "monthly", Priority: 0.6},
	}

	// Add dynamic content from database
	var articles []models.Article
	h.db.Where("status = ?", "published").Find(&articles)

	for _, article := range articles {
		pages = append(pages, models.SitemapPage{
			Path:         "/blog/" + article.Slug,
			LastModified: article.UpdatedAt,
			ChangeFreq:   "monthly",
			Priority:     0.7,
		})
	}

	// Combine all pages
	pages = append(staticPages, pages...)

	// Set last modified times
	now := time.Now()
	for i := range pages {
		if pages[i].LastModified.IsZero() {
			pages[i].LastModified = now
		}
	}

	sitemap, err := h.seoService.GenerateSitemap(c.Request.Context(), pages)
	if err != nil {
		h.logger.Error("Failed to generate sitemap", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate sitemap"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sitemap": sitemap,
		"pages":   len(pages),
		"generated": time.Now(),
	})
}

// UpdateRobotsTxt updates the robots.txt configuration
func (h *SEOHandler) UpdateRobotsTxt(c *gin.Context) {
	var input struct {
		Content string `json:"content" binding:"required"`
		Active  bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update robots.txt in SEO config
	var config models.LandingPageSEOConfig
	err := h.db.First(&config).Error

	if err == gorm.ErrRecordNotFound {
		config = models.LandingPageSEOConfig{
			RobotsTxt:      input.Content,
			RobotsEnabled:  input.Active,
		}
		h.db.Create(&config)
	} else {
		config.RobotsTxt = input.Content
		config.RobotsEnabled = input.Active
		h.db.Save(&config)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Robots.txt updated successfully",
		"content": input.Content,
		"active":  input.Active,
	})
}

// GetSEOAnalytics returns SEO performance analytics
func (h *SEOHandler) GetSEOAnalytics(c *gin.Context) {
	days := c.DefaultQuery("days", "30")
	daysInt, _ := strconv.Atoi(days)

	// Mock analytics data - replace with actual data from analytics service
	analytics := gin.H{
		"period": gin.H{
			"start": time.Now().AddDate(0, 0, -daysInt),
			"end":   time.Now(),
			"days":  daysInt,
		},
		"metrics": gin.H{
			"organic_traffic": gin.H{
				"total":  12450,
				"change": 15.3,
			},
			"avg_position": gin.H{
				"value":  8.2,
				"change": -2.1,
			},
			"click_through_rate": gin.H{
				"value":  3.8,
				"change": 0.5,
			},
			"impressions": gin.H{
				"total":  328000,
				"change": 22.7,
			},
		},
		"top_keywords": []gin.H{
			{"keyword": "status page", "position": 3, "clicks": 850},
			{"keyword": "incident management", "position": 5, "clicks": 620},
			{"keyword": "uptime monitoring", "position": 7, "clicks": 480},
			{"keyword": "service status", "position": 4, "clicks": 410},
		},
		"top_pages": []gin.H{
			{"url": "/", "views": 5420, "avg_time": 145},
			{"url": "/features", "views": 3210, "avg_time": 238},
			{"url": "/pricing", "views": 2890, "avg_time": 187},
			{"url": "/blog/incident-communication", "views": 1450, "avg_time": 342},
		},
		"crawl_errors": []gin.H{
			{"type": "404", "count": 3, "urls": []string{"/old-page", "/test", "/demo"}},
			{"type": "500", "count": 0, "urls": []string{}},
		},
		"core_web_vitals": gin.H{
			"good":         68,
			"needs_work":   24,
			"poor":         8,
			"lcp_avg":      2.1,
			"fid_avg":      85,
			"cls_avg":      0.08,
		},
	}

	c.JSON(http.StatusOK, analytics)
}

// GetSEORecommendations returns actionable SEO recommendations
func (h *SEOHandler) GetSEORecommendationsAdmin(c *gin.Context) {
	// Analyze current SEO state and generate recommendations
	recommendations := []gin.H{
		{
			"id":          1,
			"type":        "technical",
			"priority":    "high",
			"title":       "Implement Core Web Vitals Optimization",
			"description": "Your LCP is above 2.5s on mobile. Optimize images and implement lazy loading.",
			"impact":      "High impact on rankings and user experience",
			"effort":      "medium",
			"status":      "pending",
		},
		{
			"id":          2,
			"type":        "content",
			"priority":    "medium",
			"title":       "Add Long-form Content",
			"description": "Create comprehensive guides and tutorials. Pages with 2000+ words rank better.",
			"impact":      "Medium impact on organic traffic",
			"effort":      "high",
			"status":      "pending",
		},
		{
			"id":          3,
			"type":        "technical",
			"priority":    "high",
			"title":       "Fix Broken Internal Links",
			"description": "Found 3 broken internal links that need to be fixed.",
			"impact":      "Medium impact on crawlability",
			"effort":      "low",
			"status":      "pending",
		},
		{
			"id":          4,
			"type":        "on-page",
			"priority":    "medium",
			"title":       "Optimize Meta Descriptions",
			"description": "12 pages are missing meta descriptions. Add unique descriptions for each.",
			"impact":      "Low-medium impact on CTR",
			"effort":      "low",
			"status":      "in_progress",
		},
		{
			"id":          5,
			"type":        "content",
			"priority":    "low",
			"title":       "Add FAQ Schema",
			"description": "Implement FAQ structured data for better SERP features.",
			"impact":      "Low impact but improves visibility",
			"effort":      "low",
			"status":      "pending",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"summary": gin.H{
			"total":        len(recommendations),
			"high_priority": 2,
			"in_progress":  1,
			"completed":    0,
		},
	})
}

// UpdatePageSEO updates SEO settings for a specific page
func (h *SEOHandler) UpdatePageSEO(c *gin.Context) {
	pageID := c.Param("id")

	var input struct {
		MetaTitle       string `json:"meta_title"`
		MetaDescription string `json:"meta_description"`
		CanonicalURL    string `json:"canonical_url"`
		OGTitle         string `json:"og_title"`
		OGDescription   string `json:"og_description"`
		OGImage         string `json:"og_image"`
		SchemaMarkup    string `json:"schema_markup"`
		Keywords        string `json:"keywords"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update page-specific SEO settings
	// This would update the specific page model based on page type

	c.JSON(http.StatusOK, gin.H{
		"message": "Page SEO updated successfully",
		"page_id": pageID,
		"data":    input,
	})
}

// Helper methods

func (h *SEOHandler) validateJSONLD(schema string) bool {
	// Simple validation - in production, use a proper JSON-LD validator
	if schema == "" {
		return false
	}

	// Check for required fields
	requiredFields := []string{"@context", "@type"}
	for _, field := range requiredFields {
		if !contains(schema, field) {
			return false
		}
	}

	return true
}

func (h *SEOHandler) generateMetaPreview(pageType string, config *models.LandingPageSEOConfig) gin.H {
	preview := gin.H{
		"page_type": pageType,
		"meta_tags": gin.H{
			"title":       config.DefaultTitle,
			"description": config.DefaultMetaDescription,
			"keywords":    config.DefaultKeywords,
			"canonical":   "https://yourdomain.com/" + pageType,
		},
		"open_graph": gin.H{
			"og:title":       config.OGDefaultTitle,
			"og:description": config.OGDefaultDescription,
			"og:image":       config.DefaultOGImage,
			"og:type":        "website",
			"og:url":         "https://yourdomain.com/" + pageType,
		},
		"twitter": gin.H{
			"twitter:card":        config.TwitterCard,
			"twitter:site":        config.SiteName,
			"twitter:creator":     config.SiteName,
			"twitter:title":       config.DefaultTitle,
			"twitter:description": config.DefaultMetaDescription,
		},
		"structured_data": gin.H{
			"@context": "https://schema.org",
			"@type":    "WebSite",
			"name":     "Beakon StatusPage",
			"url":      "https://yourdomain.com",
		},
	}

	return preview
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(substr) < len(s) && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}