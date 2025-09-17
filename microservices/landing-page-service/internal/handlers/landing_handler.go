// Package handlers provides HTTP handlers for the Landing Page Service.
package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/anupamdutta5/statuspage-landing-service/internal/models"
	"github.com/anupamdutta5/statuspage-landing-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LandingHandler handles landing page-related HTTP requests.
type LandingHandler struct {
	service  *services.LandingService
	logger   *zap.Logger
	template *template.Template
}

// NewLandingHandler creates a new landing handler.
func NewLandingHandler(service *services.LandingService, logger *zap.Logger) *LandingHandler {
	// Load templates - try multiple paths for different environments
	var tmpl *template.Template
	var err error

	// Try different template paths
	templatePaths := []string{
		"web/templates/*.html",
		"./web/templates/*.html",
		"../web/templates/*.html",
		"../../web/templates/*.html",
	}

	for _, path := range templatePaths {
		tmpl, err = template.ParseGlob(path)
		if err == nil {
			logger.Info("Templates loaded successfully", zap.String("path", path))
			break
		}
		logger.Debug("Failed to load templates from path", zap.String("path", path), zap.Error(err))
	}

	// If no templates found, create an empty template to prevent panic
	if tmpl == nil {
		logger.Warn("No templates found, creating empty template for testing")
		tmpl = template.New("empty")
	}

	return &LandingHandler{
		service:  service,
		logger:   logger,
		template: tmpl,
	}
}

// HealthCheck handles health check requests.
func (h *LandingHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "landing-page-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "landing-page-service",
		"version": "1.0.0",
	})
}

// LandingPage handles the main landing page.
func (h *LandingHandler) LandingPage(c *gin.Context) {
	h.logger.Info("Serving landing page")

	// Get landing page data
	data, err := h.service.GetLandingPageData(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get landing page data", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Error",
			"message": "Failed to load page data",
		})
		return
	}

	// Render the landing page
	c.HTML(http.StatusOK, "index.html", data)
}

// BlogPage handles the blog listing page.
func (h *LandingHandler) BlogPage(c *gin.Context) {
	h.logger.Info("Serving blog page")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get articles
	articles, err := h.service.GetArticles(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get articles", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Error",
			"message": "Failed to load articles",
		})
		return
	}

	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title":    "Blog",
		"articles": articles,
	})
}

// ArticlePage handles individual article pages.
func (h *LandingHandler) ArticlePage(c *gin.Context) {
	slug := c.Param("slug")
	h.logger.Info("Serving article page", zap.String("slug", slug))

	// Get article
	article, err := h.service.GetArticle(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error("Failed to get article", zap.Error(err))
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Article Not Found",
			"message": "The requested article could not be found",
		})
		return
	}

	c.HTML(http.StatusOK, "article.html", gin.H{
		"title":   article.Title,
		"article": article,
	})
}

// ContactPage handles the contact page.
func (h *LandingHandler) ContactPage(c *gin.Context) {
	h.logger.Info("Serving contact page")

	c.HTML(http.StatusOK, "contact.html", gin.H{
		"title": "Contact Us",
	})
}

// PrivacyPage handles the privacy policy page.
func (h *LandingHandler) PrivacyPage(c *gin.Context) {
	h.logger.Info("Serving privacy page")

	c.HTML(http.StatusOK, "privacy.html", gin.H{
		"title": "Privacy Policy",
	})
}

// TermsPage handles the terms of service page.
func (h *LandingHandler) TermsPage(c *gin.Context) {
	h.logger.Info("Serving terms page")

	c.HTML(http.StatusOK, "terms.html", gin.H{
		"title": "Terms of Service",
	})
}

// Contact Form Handlers

// SubmitContactForm handles contact form submissions.
func (h *LandingHandler) SubmitContactForm(c *gin.Context) {
	var form models.ContactForm
	if err := c.ShouldBindJSON(&form); err != nil {
		h.logger.Error("Invalid contact form data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid form data",
		})
		return
	}

	// Set additional fields
	form.IPAddress = c.ClientIP()
	form.UserAgent = c.GetHeader("User-Agent")

	h.logger.Info("Submitting contact form", zap.String("email", form.Email))

	if err := h.service.SubmitContactForm(c.Request.Context(), &form); err != nil {
		h.logger.Error("Failed to submit contact form", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to submit form",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thank you for your message! We'll get back to you soon.",
	})
}

// Newsletter Handlers

// SubscribeNewsletter handles newsletter subscriptions.
func (h *LandingHandler) SubscribeNewsletter(c *gin.Context) {
	var subscription models.Newsletter
	if err := c.ShouldBindJSON(&subscription); err != nil {
		h.logger.Error("Invalid newsletter subscription data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription data",
		})
		return
	}

	// Set additional fields
	subscription.IPAddress = c.ClientIP()
	subscription.UserAgent = c.GetHeader("User-Agent")
	subscription.Source = "landing_page"

	h.logger.Info("Subscribing to newsletter", zap.String("email", subscription.Email))

	if err := h.service.SubscribeNewsletter(c.Request.Context(), &subscription); err != nil {
		h.logger.Error("Failed to subscribe to newsletter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to subscribe",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thanks for subscribing! Check your email for confirmation.",
	})
}

// Pricing Plan Handlers

// SyncPricingPlans handles syncing pricing plans from SaaS Admin Service.
func (h *LandingHandler) SyncPricingPlans(c *gin.Context) {
	var plans []*models.PricingPlan
	if err := c.ShouldBindJSON(&plans); err != nil {
		h.logger.Error("Invalid pricing plans data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid pricing plans data",
		})
		return
	}

	h.logger.Info("Syncing pricing plans", zap.Int("count", len(plans)))

	if err := h.service.SyncPricingPlans(c.Request.Context(), plans); err != nil {
		h.logger.Error("Failed to sync pricing plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync pricing plans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing plans synced successfully",
		"count":   len(plans),
	})
}

// Hero Section Handlers

// GetHeroSection handles retrieving the hero section.
func (h *LandingHandler) GetHeroSection(c *gin.Context) {
	h.logger.Info("Getting hero section")

	hero, err := h.service.GetHeroSection(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get hero section", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get hero section",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hero": hero,
	})
}

// CreateHeroSection handles creating a new hero section.
func (h *LandingHandler) CreateHeroSection(c *gin.Context) {
	var hero models.HeroSection
	if err := c.ShouldBindJSON(&hero); err != nil {
		h.logger.Error("Invalid hero section data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hero section data",
		})
		return
	}

	h.logger.Info("Creating hero section", zap.String("title", hero.Title))

	if err := h.service.CreateHeroSection(c.Request.Context(), &hero); err != nil {
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

// UpdateHeroSection handles updating a hero section.
func (h *LandingHandler) UpdateHeroSection(c *gin.Context) {
	heroIDStr := c.Param("id")
	heroID, err := strconv.ParseUint(heroIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hero section ID",
		})
		return
	}

	var updates models.HeroSection
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid hero section update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hero section update data",
		})
		return
	}

	h.logger.Info("Updating hero section", zap.Uint64("hero_id", heroID))

	if err := h.service.UpdateHeroSection(c.Request.Context(), uint(heroID), &updates); err != nil {
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

// Feature Handlers

// GetFeatures handles retrieving features.
func (h *LandingHandler) GetFeatures(c *gin.Context) {
	h.logger.Info("Getting features")

	features, err := h.service.GetFeatures(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get features", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get features",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"features": features,
	})
}

// CreateFeature handles creating a new feature.
func (h *LandingHandler) CreateFeature(c *gin.Context) {
	var feature models.FeatureSection
	if err := c.ShouldBindJSON(&feature); err != nil {
		h.logger.Error("Invalid feature data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature data",
		})
		return
	}

	h.logger.Info("Creating feature", zap.String("title", feature.Title))

	if err := h.service.CreateFeature(c.Request.Context(), &feature); err != nil {
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

// UpdateFeature handles updating a feature.
func (h *LandingHandler) UpdateFeature(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := strconv.ParseUint(featureIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature ID",
		})
		return
	}

	var updates models.FeatureSection
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid feature update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid feature update data",
		})
		return
	}

	h.logger.Info("Updating feature", zap.Uint64("feature_id", featureID))

	if err := h.service.UpdateFeature(c.Request.Context(), uint(featureID), &updates); err != nil {
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

// Testimonial Handlers

// GetTestimonials handles retrieving testimonials.
func (h *LandingHandler) GetTestimonials(c *gin.Context) {
	h.logger.Info("Getting testimonials")

	testimonials, err := h.service.GetTestimonials(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get testimonials", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get testimonials",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"testimonials": testimonials,
	})
}

// CreateTestimonial handles creating a new testimonial.
func (h *LandingHandler) CreateTestimonial(c *gin.Context) {
	var testimonial models.Testimonial
	if err := c.ShouldBindJSON(&testimonial); err != nil {
		h.logger.Error("Invalid testimonial data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid testimonial data",
		})
		return
	}

	h.logger.Info("Creating testimonial", zap.String("name", testimonial.Name))

	if err := h.service.CreateTestimonial(c.Request.Context(), &testimonial); err != nil {
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

// UpdateTestimonial handles updating a testimonial.
func (h *LandingHandler) UpdateTestimonial(c *gin.Context) {
	testimonialIDStr := c.Param("id")
	testimonialID, err := strconv.ParseUint(testimonialIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid testimonial ID",
		})
		return
	}

	var updates models.Testimonial
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid testimonial update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid testimonial update data",
		})
		return
	}

	h.logger.Info("Updating testimonial", zap.Uint64("testimonial_id", testimonialID))

	if err := h.service.UpdateTestimonial(c.Request.Context(), uint(testimonialID), &updates); err != nil {
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

// FAQ Handlers

// GetFAQs handles retrieving FAQs.
func (h *LandingHandler) GetFAQs(c *gin.Context) {
	h.logger.Info("Getting FAQs")

	faqs, err := h.service.GetFAQs(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get FAQs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get FAQs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"faqs": faqs,
	})
}

// CreateFAQ handles creating a new FAQ.
func (h *LandingHandler) CreateFAQ(c *gin.Context) {
	var faq models.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		h.logger.Error("Invalid FAQ data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid FAQ data",
		})
		return
	}

	h.logger.Info("Creating FAQ", zap.String("question", faq.Question))

	if err := h.service.CreateFAQ(c.Request.Context(), &faq); err != nil {
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

// UpdateFAQ handles updating a FAQ.
func (h *LandingHandler) UpdateFAQ(c *gin.Context) {
	faqIDStr := c.Param("id")
	faqID, err := strconv.ParseUint(faqIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid FAQ ID",
		})
		return
	}

	var updates models.FAQ
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid FAQ update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid FAQ update data",
		})
		return
	}

	h.logger.Info("Updating FAQ", zap.Uint64("faq_id", faqID))

	if err := h.service.UpdateFAQ(c.Request.Context(), uint(faqID), &updates); err != nil {
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

// Article Handlers

// GetArticles handles retrieving articles.
func (h *LandingHandler) GetArticles(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	h.logger.Info("Getting articles", zap.Int("limit", limit), zap.Int("offset", offset))

	articles, err := h.service.GetArticles(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get articles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get articles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"count":    len(articles),
		"limit":    limit,
		"offset":   offset,
	})
}

// CreateArticle handles creating a new article.
func (h *LandingHandler) CreateArticle(c *gin.Context) {
	var article models.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		h.logger.Error("Invalid article data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid article data",
		})
		return
	}

	h.logger.Info("Creating article", zap.String("title", article.Title))

	if err := h.service.CreateArticle(c.Request.Context(), &article); err != nil {
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

// UpdateArticle handles updating an article.
func (h *LandingHandler) UpdateArticle(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid article ID",
		})
		return
	}

	var updates models.Article
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid article update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid article update data",
		})
		return
	}

	h.logger.Info("Updating article", zap.Uint64("article_id", articleID))

	if err := h.service.UpdateArticle(c.Request.Context(), uint(articleID), &updates); err != nil {
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

// DeleteArticle handles deleting an article.
func (h *LandingHandler) DeleteArticle(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid article ID",
		})
		return
	}

	h.logger.Info("Deleting article", zap.Uint64("article_id", articleID))

	if err := h.service.DeleteArticle(c.Request.Context(), uint(articleID)); err != nil {
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
