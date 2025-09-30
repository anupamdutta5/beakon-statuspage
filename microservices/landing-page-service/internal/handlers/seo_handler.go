// Package handlers provides HTTP handlers for SEO functionality.
package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/anupamdutta5/landing-page-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SEOHandler handles SEO-related HTTP requests
type SEOHandler struct {
	seoService *services.SEOService
	logger     *zap.Logger
}

// NewSEOHandler creates a new SEO handler
func NewSEOHandler(seoService *services.SEOService, logger *zap.Logger) *SEOHandler {
	return &SEOHandler{
		seoService: seoService,
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