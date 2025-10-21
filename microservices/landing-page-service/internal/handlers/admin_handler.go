// Package handlers provides HTTP handlers for the Landing Page Service.
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/anupamdutta5/landing-page-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AdminHandler handles admin dashboard-related HTTP requests.
type AdminHandler struct {
	db            *gorm.DB
	landingService *services.LandingService
	seoService    *services.SEOService
	logger        *zap.Logger
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(db *gorm.DB, landingService *services.LandingService, seoService *services.SEOService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		db:            db,
		landingService: landingService,
		seoService:    seoService,
		logger:        logger,
	}
}

// ============================================================================
// Dashboard & Analytics
// ============================================================================

// GetDashboard handles GET /api/v1/admin/dashboard
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	h.logger.Info("Getting admin dashboard data")

	// Gather dashboard statistics
	stats := h.getDashboardStats(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"stats":      stats,
		"timestamp": time.Now(),
	})
}

// GetAnalytics handles GET /api/v1/admin/analytics
func (h *AdminHandler) GetAnalytics(c *gin.Context) {
	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	h.logger.Info("Getting analytics data",
		zap.String("start_date", startDate),
		zap.String("end_date", endDate))

	// Get analytics data
	analytics := h.getAnalyticsData(startDate, endDate)

	c.JSON(http.StatusOK, gin.H{
		"analytics": analytics,
		"period": gin.H{
			"start": startDate,
			"end":   endDate,
		},
	})
}

// ============================================================================
// Hero Section Management
// ============================================================================

// GetHeroSections handles GET /api/v1/admin/hero
func (h *AdminHandler) GetHeroSections(c *gin.Context) {
	h.logger.Info("Getting hero sections")

	var heroes []models.HeroSection
	query := h.db.Order("`order` ASC, created_at DESC")

	// Optional status filter
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&heroes).Error; err != nil {
		h.logger.Error("Failed to get hero sections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get hero sections",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hero_sections": heroes,
		"count":        len(heroes),
	})
}

// GetHeroSection handles GET /api/v1/admin/hero/:id
func (h *AdminHandler) GetHeroSection(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Getting hero section", zap.Uint("id", id))

	var hero models.HeroSection
	if err := h.db.First(&hero, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hero section not found"})
			return
		}
		h.logger.Error("Failed to get hero section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get hero section"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hero": hero,
	})
}

// CreateHeroSection handles POST /api/v1/admin/hero
func (h *AdminHandler) CreateHeroSection(c *gin.Context) {
	var hero models.HeroSection
	if err := c.ShouldBindJSON(&hero); err != nil {
		h.logger.Error("Invalid hero section data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Creating hero section", zap.String("title", hero.Title))

	if err := h.db.Create(&hero).Error; err != nil {
		h.logger.Error("Failed to create hero section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create hero section",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"hero":    hero,
		"message": "Hero section created successfully",
	})
}

// UpdateHeroSection handles PUT /api/v1/admin/hero/:id
func (h *AdminHandler) UpdateHeroSection(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates models.HeroSection
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating hero section", zap.Uint("id", id))

	if err := h.db.Model(&models.HeroSection{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		h.logger.Error("Failed to update hero section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update hero section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hero section updated successfully",
	})
}

// DeleteHeroSection handles DELETE /api/v1/admin/hero/:id
func (h *AdminHandler) DeleteHeroSection(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting hero section", zap.Uint("id", id))

	if err := h.db.Delete(&models.HeroSection{}, id).Error; err != nil {
		h.logger.Error("Failed to delete hero section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete hero section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hero section deleted successfully",
	})
}

// ============================================================================
// Feature Section Management
// ============================================================================

// GetFeatures handles GET /api/v1/admin/features
func (h *AdminHandler) GetFeatures(c *gin.Context) {
	h.logger.Info("Getting features")

	var features []models.FeatureSection
	query := h.db.Order("`order` ASC, created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&features).Error; err != nil {
		h.logger.Error("Failed to get features", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get features",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"count":    len(features),
	})
}

// CreateFeature handles POST /api/v1/admin/features
func (h *AdminHandler) CreateFeature(c *gin.Context) {
	var feature models.FeatureSection
	if err := c.ShouldBindJSON(&feature); err != nil {
		h.logger.Error("Invalid feature data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Creating feature", zap.String("title", feature.Title))

	if err := h.db.Create(&feature).Error; err != nil {
		h.logger.Error("Failed to create feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create feature",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"feature": feature,
		"message": "Feature created successfully",
	})
}

// UpdateFeature handles PUT /api/v1/admin/features/:id
func (h *AdminHandler) UpdateFeature(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates models.FeatureSection
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating feature", zap.Uint("id", id))

	if err := h.db.Model(&models.FeatureSection{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		h.logger.Error("Failed to update feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature updated successfully",
	})
}

// DeleteFeature handles DELETE /api/v1/admin/features/:id
func (h *AdminHandler) DeleteFeature(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting feature", zap.Uint("id", id))

	if err := h.db.Delete(&models.FeatureSection{}, id).Error; err != nil {
		h.logger.Error("Failed to delete feature", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete feature",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature deleted successfully",
	})
}

// ============================================================================
// Testimonial Management
// ============================================================================

// GetTestimonials handles GET /api/v1/admin/testimonials
func (h *AdminHandler) GetTestimonials(c *gin.Context) {
	h.logger.Info("Getting testimonials")

	var testimonials []models.Testimonial
	query := h.db.Order("`order` ASC, created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&testimonials).Error; err != nil {
		h.logger.Error("Failed to get testimonials", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get testimonials",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"testimonials": testimonials,
		"count":       len(testimonials),
	})
}

// CreateTestimonial handles POST /api/v1/admin/testimonials
func (h *AdminHandler) CreateTestimonial(c *gin.Context) {
	var testimonial models.Testimonial
	if err := c.ShouldBindJSON(&testimonial); err != nil {
		h.logger.Error("Invalid testimonial data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Creating testimonial", zap.String("name", testimonial.Name))

	if err := h.db.Create(&testimonial).Error; err != nil {
		h.logger.Error("Failed to create testimonial", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create testimonial",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"testimonial": testimonial,
		"message":     "Testimonial created successfully",
	})
}

// UpdateTestimonial handles PUT /api/v1/admin/testimonials/:id
func (h *AdminHandler) UpdateTestimonial(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates models.Testimonial
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating testimonial", zap.Uint("id", id))

	if err := h.db.Model(&models.Testimonial{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		h.logger.Error("Failed to update testimonial", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update testimonial",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonial updated successfully",
	})
}

// DeleteTestimonial handles DELETE /api/v1/admin/testimonials/:id
func (h *AdminHandler) DeleteTestimonial(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting testimonial", zap.Uint("id", id))

	if err := h.db.Delete(&models.Testimonial{}, id).Error; err != nil {
		h.logger.Error("Failed to delete testimonial", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete testimonial",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Testimonial deleted successfully",
	})
}

// ============================================================================
// FAQ Management
// ============================================================================

// GetFAQs handles GET /api/v1/admin/faqs
func (h *AdminHandler) GetFAQs(c *gin.Context) {
	h.logger.Info("Getting FAQs")

	var faqs []models.FAQ
	query := h.db.Order("`order` ASC, created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&faqs).Error; err != nil {
		h.logger.Error("Failed to get FAQs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get FAQs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"faqs":  faqs,
		"count": len(faqs),
	})
}

// CreateFAQ handles POST /api/v1/admin/faqs
func (h *AdminHandler) CreateFAQ(c *gin.Context) {
	var faq models.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		h.logger.Error("Invalid FAQ data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Creating FAQ", zap.String("question", faq.Question))

	if err := h.db.Create(&faq).Error; err != nil {
		h.logger.Error("Failed to create FAQ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create FAQ",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"faq":     faq,
		"message": "FAQ created successfully",
	})
}

// UpdateFAQ handles PUT /api/v1/admin/faqs/:id
func (h *AdminHandler) UpdateFAQ(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates models.FAQ
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating FAQ", zap.Uint("id", id))

	if err := h.db.Model(&models.FAQ{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		h.logger.Error("Failed to update FAQ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update FAQ",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ updated successfully",
	})
}

// DeleteFAQ handles DELETE /api/v1/admin/faqs/:id
func (h *AdminHandler) DeleteFAQ(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting FAQ", zap.Uint("id", id))

	if err := h.db.Delete(&models.FAQ{}, id).Error; err != nil {
		h.logger.Error("Failed to delete FAQ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete FAQ",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ deleted successfully",
	})
}

// ============================================================================
// Article Management
// ============================================================================

// GetArticles handles GET /api/v1/admin/articles
func (h *AdminHandler) GetArticles(c *gin.Context) {
	h.logger.Info("Getting articles")

	var articles []models.Article
	query := h.db.Order("published_at DESC, created_at DESC")

	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if featured := c.Query("featured"); featured == "true" {
		query = query.Where("is_featured = ?", true)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query = query.Limit(limit).Offset(offset)

	if err := query.Find(&articles).Error; err != nil {
		h.logger.Error("Failed to get articles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get articles",
		})
		return
	}

	// Get total count
	var total int64
	h.db.Model(&models.Article{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetArticle handles GET /api/v1/admin/articles/:id
func (h *AdminHandler) GetArticle(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Getting article", zap.Uint("id", id))

	var article models.Article
	if err := h.db.First(&article, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		h.logger.Error("Failed to get article", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"article": article,
	})
}

// CreateArticle handles POST /api/v1/admin/articles
func (h *AdminHandler) CreateArticle(c *gin.Context) {
	var article models.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		h.logger.Error("Invalid article data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Creating article", zap.String("title", article.Title))

	// Calculate reading time
	article.ReadingTimeMinutes = h.calculateReadingTime(article.Content)

	// Generate slug if not provided
	if article.Slug == "" {
		article.Slug = h.generateSlug(article.Title)
	}

	if err := h.db.Create(&article).Error; err != nil {
		h.logger.Error("Failed to create article", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create article",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"article": article,
		"message": "Article created successfully",
	})
}

// UpdateArticle handles PUT /api/v1/admin/articles/:id
func (h *AdminHandler) UpdateArticle(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates models.Article
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating article", zap.Uint("id", id))

	// Recalculate reading time if content changed
	if updates.Content != "" {
		updates.ReadingTimeMinutes = h.calculateReadingTime(updates.Content)
	}

	if err := h.db.Model(&models.Article{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		h.logger.Error("Failed to update article", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update article",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Article updated successfully",
	})
}

// DeleteArticle handles DELETE /api/v1/admin/articles/:id
func (h *AdminHandler) DeleteArticle(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting article", zap.Uint("id", id))

	if err := h.db.Delete(&models.Article{}, id).Error; err != nil {
		h.logger.Error("Failed to delete article", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete article",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Article deleted successfully",
	})
}

// ============================================================================
// Custom Sections Management
// ============================================================================

// GetCustomSections handles GET /api/v1/admin/sections
func (h *AdminHandler) GetCustomSections(c *gin.Context) {
	h.logger.Info("Getting custom sections")

	var sections []models.LandingPageSection
	query := h.db.Order("`position`, `order` ASC")

	if sectionType := c.Query("type"); sectionType != "" {
		query = query.Where("section_type = ?", sectionType)
	}

	if position := c.Query("position"); position != "" {
		query = query.Where("position = ?", position)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&sections).Error; err != nil {
		h.logger.Error("Failed to get sections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sections",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sections": sections,
		"count":    len(sections),
	})
}

// CreateCustomSection handles POST /api/v1/admin/sections
func (h *AdminHandler) CreateCustomSection(c *gin.Context) {
	var section models.LandingPageSection
	if err := c.ShouldBindJSON(&section); err != nil {
		h.logger.Error("Invalid section data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Creating custom section", zap.String("name", section.Name))

	if err := h.db.Create(&section).Error; err != nil {
		h.logger.Error("Failed to create section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create section",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"section": section,
		"message": "Section created successfully",
	})
}

// UpdateCustomSection handles PUT /api/v1/admin/sections/:id
func (h *AdminHandler) UpdateCustomSection(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates models.LandingPageSection
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating custom section", zap.Uint("id", id))

	if err := h.db.Model(&models.LandingPageSection{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		h.logger.Error("Failed to update section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Section updated successfully",
	})
}

// DeleteCustomSection handles DELETE /api/v1/admin/sections/:id
func (h *AdminHandler) DeleteCustomSection(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting custom section", zap.Uint("id", id))

	if err := h.db.Delete(&models.LandingPageSection{}, id).Error; err != nil {
		h.logger.Error("Failed to delete section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Section deleted successfully",
	})
}

// ============================================================================
// Content Preview
// ============================================================================

// GetContentPreview handles GET /api/v1/admin/content/preview
func (h *AdminHandler) GetContentPreview(c *gin.Context) {
	pageType := c.DefaultQuery("type", "landing")

	h.logger.Info("Getting content preview", zap.String("type", pageType))

	// Get all active content for preview
	preview := gin.H{
		"type": pageType,
	}

	// Get hero section
	var hero models.HeroSection
	if err := h.db.Where("status = ?", "active").First(&hero).Error; err == nil {
		preview["hero"] = hero
	}

	// Get features
	var features []models.FeatureSection
	if err := h.db.Where("status = ?", "active").Order("`order` ASC").Find(&features).Error; err == nil {
		preview["features"] = features
	}

	// Get testimonials
	var testimonials []models.Testimonial
	if err := h.db.Where("status = ?", "active").Order("`order` ASC").Limit(3).Find(&testimonials).Error; err == nil {
		preview["testimonials"] = testimonials
	}

	// Get FAQs
	var faqs []models.FAQ
	if err := h.db.Where("status = ?", "active").Order("`order` ASC").Limit(5).Find(&faqs).Error; err == nil {
		preview["faqs"] = faqs
	}

	c.JSON(http.StatusOK, preview)
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkUpdateOrder handles POST /api/v1/admin/bulk/update-order
func (h *AdminHandler) BulkUpdateOrder(c *gin.Context) {
	var request struct {
		Type  string `json:"type" binding:"required"` // hero, feature, testimonial, faq
		Items []struct {
			ID    uint `json:"id" binding:"required"`
			Order int  `json:"order" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid bulk update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Bulk updating order",
		zap.String("type", request.Type),
		zap.Int("count", len(request.Items)))

	// Start transaction
	tx := h.db.Begin()

	for _, item := range request.Items {
		var err error
		switch request.Type {
		case "hero":
			err = tx.Model(&models.HeroSection{}).Where("id = ?", item.ID).Update("order", item.Order).Error
		case "feature":
			err = tx.Model(&models.FeatureSection{}).Where("id = ?", item.ID).Update("order", item.Order).Error
		case "testimonial":
			err = tx.Model(&models.Testimonial{}).Where("id = ?", item.ID).Update("order", item.Order).Error
		case "faq":
			err = tx.Model(&models.FAQ{}).Where("id = ?", item.ID).Update("order", item.Order).Error
		default:
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
			return
		}

		if err != nil {
			tx.Rollback()
			h.logger.Error("Failed to update order", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update order",
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
		"count":   len(request.Items),
	})
}

// ============================================================================
// Helper Methods
// ============================================================================

// parseID parses ID from URL parameter
func (h *AdminHandler) parseID(c *gin.Context) uint {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}

// getDashboardStats gathers dashboard statistics
func (h *AdminHandler) getDashboardStats(ctx context.Context) gin.H {
	stats := gin.H{}

	// Content counts
	var heroCount, featureCount, testimonialCount, faqCount, articleCount int64
	h.db.Model(&models.HeroSection{}).Where("status = ?", "active").Count(&heroCount)
	h.db.Model(&models.FeatureSection{}).Where("status = ?", "active").Count(&featureCount)
	h.db.Model(&models.Testimonial{}).Where("status = ?", "active").Count(&testimonialCount)
	h.db.Model(&models.FAQ{}).Where("status = ?", "active").Count(&faqCount)
	h.db.Model(&models.Article{}).Where("status = ?", "published").Count(&articleCount)

	stats["content"] = gin.H{
		"hero_sections": heroCount,
		"features":     featureCount,
		"testimonials": testimonialCount,
		"faqs":        faqCount,
		"articles":    articleCount,
	}

	// Recent activity
	var recentArticles []models.Article
	h.db.Order("created_at DESC").Limit(5).Find(&recentArticles)
	stats["recent_articles"] = recentArticles

	// A/B test stats
	var runningTests int64
	h.db.Model(&models.LandingPageABTest{}).Where("status = ?", "running").Count(&runningTests)
	stats["ab_tests"] = gin.H{
		"running": runningTests,
	}

	// Today's stats
	today := time.Now().Truncate(24 * time.Hour)
	var todayViews int64
	h.db.Model(&models.LandingPageStats{}).Where("date = ?", today).Select("SUM(page_views)").Scan(&todayViews)
	stats["today"] = gin.H{
		"page_views": todayViews,
	}

	return stats
}

// getAnalyticsData retrieves analytics data for the specified period
func (h *AdminHandler) getAnalyticsData(startDate, endDate string) gin.H {
	analytics := gin.H{
		"page_views":    []gin.H{},
		"conversions":   []gin.H{},
		"top_pages":     []gin.H{},
		"traffic_sources": []gin.H{},
	}

	// Parse dates
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	// Get daily stats
	var stats []models.LandingPageStats
	h.db.Where("date >= ? AND date <= ?", start, end).Order("date ASC").Find(&stats)

	for _, stat := range stats {
		analytics["page_views"] = append(analytics["page_views"].([]gin.H), gin.H{
			"date":  stat.Date.Format("2006-01-02"),
			"value": stat.PageViews,
		})
	}

	return analytics
}

// calculateReadingTime calculates reading time for an article
func (h *AdminHandler) calculateReadingTime(content string) int {
	// Average reading speed: 200-250 words per minute
	// Using 225 as average
	wordsPerMinute := 225

	// Simple word count (can be improved)
	words := len(content) / 5 // Rough estimate: average word length is 5 characters

	readingTime := words / wordsPerMinute
	if readingTime < 1 {
		readingTime = 1
	}

	return readingTime
}

// generateSlug generates a URL-friendly slug from a title
func (h *AdminHandler) generateSlug(title string) string {
	// Simple slug generation (can be improved with a proper slug library)
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "!", "")
	slug = strings.ReplaceAll(slug, "&", "and")

	// Add timestamp to ensure uniqueness
	slug = slug + "-" + strconv.FormatInt(time.Now().Unix(), 10)

	return slug
}