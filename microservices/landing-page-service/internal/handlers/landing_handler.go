// Package handlers provides HTTP handlers for the Landing Page Service.
package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/anupamdutta5/landing-page-service/internal/services"
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

	// Create template with custom functions
	funcMap := template.FuncMap{
		"mul": func(a, b int) int {
			return a * b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	// Try different template paths
	templatePaths := []string{
		"web/templates/*.html",
		"./web/templates/*.html",
		"../web/templates/*.html",
		"../../web/templates/*.html",
	}

	for _, path := range templatePaths {
		tmpl, err = template.New("landing").Funcs(funcMap).ParseGlob(path)
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

// ====================== MOCK AUTHENTICATION & PAYMENT ENDPOINTS ======================

// LoginPage displays the login page
func (h *LandingHandler) LoginPage(c *gin.Context) {
	data := gin.H{
		"Title":       "Sign In - StatusPage Pro",
		"Description": "Sign in to your StatusPage Pro account",
	}

	if h.template != nil {
		if err := h.template.ExecuteTemplate(c.Writer, "login.html", data); err != nil {
			h.logger.Error("Failed to render login template", zap.Error(err))
			c.HTML(http.StatusOK, "login.html", data)
		}
	} else {
		c.HTML(http.StatusOK, "login.html", data)
	}
}

// SignupPage displays the signup page
func (h *LandingHandler) SignupPage(c *gin.Context) {
	plan := c.Query("plan")
	if plan == "" {
		plan = "free"
	}

	data := gin.H{
		"Title":       "Get Started - StatusPage Pro",
		"Description": "Create your StatusPage Pro account",
		"Plan":        plan,
	}

	if h.template != nil {
		if err := h.template.ExecuteTemplate(c.Writer, "signup.html", data); err != nil {
			h.logger.Error("Failed to render signup template", zap.Error(err))
			c.HTML(http.StatusOK, "signup.html", data)
		}
	} else {
		c.HTML(http.StatusOK, "signup.html", data)
	}
}

// MockLogin handles login requests (mock)
func (h *LandingHandler) MockLogin(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid login credentials",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Mock login attempt", zap.String("email", loginReq.Email))

	// Mock authentication - in a real system, this would validate credentials
	if loginReq.Email == "" || loginReq.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Mock tenant assignment based on email domain or predefined mapping
	var tenantID string
	var tenantName string
	var redirectURL string

	// Simulate tenant assignment logic
	switch {
	case strings.Contains(loginReq.Email, "demo"):
		tenantID = "demo-tenant-001"
		tenantName = "Demo Company"
	case strings.Contains(loginReq.Email, "acme"):
		tenantID = "acme-corp-002"
		tenantName = "Acme Corporation"
	case strings.Contains(loginReq.Email, "test"):
		tenantID = "test-org-003"
		tenantName = "Test Organization"
	default:
		// Default tenant for any user
		tenantID = "default-tenant-001"
		tenantName = "Default Organization"
	}

	// Generate redirect URL to tenant admin dashboard with tenant information
	tenantAdminURL := "http://localhost:8099"
	redirectURL = fmt.Sprintf("%s/admin?tenant_id=%s&tenant_name=%s", tenantAdminURL, url.QueryEscape(tenantID), url.QueryEscape(tenantName))

	// Generate mock JWT token with tenant information
	mockToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwidGVuYW50IjoiZGVtby10ZW5hbnQtMDAxIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	h.logger.Info("Login successful, redirecting to tenant admin",
		zap.String("tenant_id", tenantID),
		zap.String("tenant_name", tenantName),
		zap.String("redirect_url", redirectURL))

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":     1,
			"email":  loginReq.Email,
			"name":   "Demo User",
			"tenant": gin.H{
				"id":   tenantID,
				"name": tenantName,
			},
		},
		"token": mockToken,
		"redirect": redirectURL,
	})
}

// MockSignup handles signup requests (mock)
func (h *LandingHandler) MockSignup(c *gin.Context) {
	var signupReq struct {
		Name     string `json:"name" binding:"required,min=2"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Plan     string `json:"plan"`
		Company  string `json:"company"`
	}

	if err := c.ShouldBindJSON(&signupReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid signup data",
			"details": err.Error(),
		})
		return
	}

	if signupReq.Plan == "" {
		signupReq.Plan = "free"
	}

	h.logger.Info("Mock signup attempt",
		zap.String("email", signupReq.Email),
		zap.String("plan", signupReq.Plan))

	// Mock user creation
	userID := 1

	// If it's a paid plan, initiate payment flow
	if signupReq.Plan != "free" {
		c.JSON(http.StatusOK, gin.H{
			"message": "Account created successfully",
			"user": gin.H{
				"id":      userID,
				"email":   signupReq.Email,
				"name":    signupReq.Name,
				"company": signupReq.Company,
				"plan":    signupReq.Plan,
			},
			"payment_required": true,
			"payment_url":      "/payment",
			"redirect":         "/payment?plan=" + signupReq.Plan,
		})
		return
	}

	// Free plan - direct activation with tenant assignment
	// Generate a new tenant for the user
	tenantID := fmt.Sprintf("tenant-%s-%d", strings.ToLower(strings.ReplaceAll(signupReq.Company, " ", "-")), userID)
	if signupReq.Company == "" {
		tenantID = fmt.Sprintf("user-tenant-%d", userID)
	}

	tenantAdminURL := "http://localhost:8099"
	redirectURL := fmt.Sprintf("%s/admin?tenant=%s", tenantAdminURL, tenantID)

	h.logger.Info("Created new tenant for user",
		zap.String("tenant_id", tenantID),
		zap.String("user_email", signupReq.Email))

	c.JSON(http.StatusOK, gin.H{
		"message": "Account created successfully",
		"user": gin.H{
			"id":      userID,
			"email":   signupReq.Email,
			"name":    signupReq.Name,
			"company": signupReq.Company,
			"plan":    signupReq.Plan,
			"tenant": gin.H{
				"id":   tenantID,
				"name": signupReq.Company,
			},
		},
		"redirect": redirectURL,
		"tenant_admin_url": tenantAdminURL,
		"next_steps": []string{
			"Complete your tenant setup",
			"Configure your status page",
			"Add your first service",
		},
	})
}

// PaymentPage displays the payment page
func (h *LandingHandler) PaymentPage(c *gin.Context) {
	plan := c.Query("plan")
	if plan == "" {
		plan = "pro"
	}

	// Mock plan pricing
	planPricing := map[string]gin.H{
		"pro": {
			"name":  "Pro",
			"price": 29.99,
			"features": []string{
				"Up to 50 services",
				"Custom domain",
				"Email support",
				"Advanced analytics",
			},
		},
		"enterprise": {
			"name":  "Enterprise",
			"price": 99.99,
			"features": []string{
				"Unlimited services",
				"White-label branding",
				"Priority support",
				"Advanced integrations",
				"SLA guarantees",
			},
		},
	}

	planData, exists := planPricing[plan]
	if !exists {
		planData = planPricing["pro"]
	}

	data := gin.H{
		"Title":       "Complete Payment - StatusPage Pro",
		"Description": "Complete your subscription payment",
		"Plan":        planData,
		"PlanName":    plan,
	}

	if h.template != nil {
		if err := h.template.ExecuteTemplate(c.Writer, "payment.html", data); err != nil {
			h.logger.Error("Failed to render payment template", zap.Error(err))
			c.HTML(http.StatusOK, "payment.html", data)
		}
	} else {
		c.HTML(http.StatusOK, "payment.html", data)
	}
}

// MockPayment handles payment processing (mock)
func (h *LandingHandler) MockPayment(c *gin.Context) {
	var paymentReq struct {
		Plan           string `json:"plan" binding:"required"`
		PaymentMethod  string `json:"payment_method" binding:"required"`
		CardNumber     string `json:"card_number"`
		ExpiryMonth    string `json:"expiry_month"`
		ExpiryYear     string `json:"expiry_year"`
		CVV           string `json:"cvv"`
		BillingEmail   string `json:"billing_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&paymentReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Mock payment processing",
		zap.String("plan", paymentReq.Plan),
		zap.String("email", paymentReq.BillingEmail))

	// Mock payment processing delay
	// time.Sleep(2 * time.Second)

	// Simulate payment success
	paymentID := "pay_mock_" + strconv.FormatInt(time.Now().Unix(), 10)

	// Extract user name from email for display (temporary solution for demo)
	emailParts := strings.Split(paymentReq.BillingEmail, "@")
	userName := emailParts[0]
	if len(userName) < 2 {
		userName = "User"
	}
	// Capitalize first letter
	userName = strings.ToUpper(userName[:1]) + userName[1:]

	// Generate company name from email domain if not provided
	companyName := "Your Company"
	if len(emailParts) > 1 {
		domain := emailParts[1]
		domainParts := strings.Split(domain, ".")
		if len(domainParts) > 0 {
			companyName = strings.ToUpper(domainParts[0][:1]) + domainParts[0][1:]
		}
	}

	// Generate tenant URL based on company
	tenantSlug := strings.ToLower(strings.ReplaceAll(companyName, " ", "-"))
	tenantURL := fmt.Sprintf("https://%s.statuspage.pro", tenantSlug)

	// Build tenant admin redirect URL with user data
	redirectURL := fmt.Sprintf("http://localhost:8099/admin?user_name=%s&user_email=%s&plan=%s&company=%s&tenant_url=%s",
		url.QueryEscape(userName), url.QueryEscape(paymentReq.BillingEmail),
		url.QueryEscape(paymentReq.Plan), url.QueryEscape(companyName), url.QueryEscape(tenantURL))

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment processed successfully",
		"payment": gin.H{
			"id":     paymentID,
			"status": "completed",
			"amount": getPlanPrice(paymentReq.Plan),
			"plan":   paymentReq.Plan,
		},
		"tenant_created": true,
		"redirect": redirectURL,
		"access_url": tenantURL,
	})
}

// Helper function to get plan pricing
func getPlanPrice(plan string) float64 {
	switch plan {
	case "pro":
		return 29.99
	case "enterprise":
		return 99.99
	default:
		return 0.0
	}
}
