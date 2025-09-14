// Package services provides business logic for the Landing Page Service.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-landing-service/internal/config"
	"github.com/enterprise-status/statuspage-landing-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// LandingService handles landing page-related business logic.
type LandingService struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
}

// NewLandingService creates a new landing service.
func NewLandingService(cfg *config.Config, logger *zap.Logger) (*LandingService, error) {
	// Try to initialize database connection, but don't fail if it doesn't work
	db, err := initDatabase(cfg.Database)
	if err != nil {
		logger.Warn("Failed to initialize database, running without database", zap.Error(err))
		db = nil // Set to nil to indicate no database
	}

	return &LandingService{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

// SetDB sets the database connection (for testing)
func (s *LandingService) SetDB(db *gorm.DB) {
	s.db = db
}

// LandingPageData represents the data structure for the landing page template.
type LandingPageData struct {
	SiteName        string
	SiteURL         string
	SiteDescription string
	SiteKeywords    []string
	ContactEmail    string
	SupportEmail    string
	SocialLinks     map[string]string
	AnalyticsID     string
	OGImage         string
	Favicon         string
	CustomCSS       string
	CustomJS        string
	CurrentYear     int
	Hero            *models.HeroSection
	Features        []*models.FeatureSection
	PricingPlans    []*PricingPlanData
	Testimonials    []*models.Testimonial
	FAQs            []*models.FAQ
}

// PricingPlanData represents pricing plan data with parsed features.
type PricingPlanData struct {
	models.PricingPlan
	FeaturesList []string
}

// GetLandingPageData retrieves all data needed for the landing page.
func (s *LandingService) GetLandingPageData(ctx context.Context) (*LandingPageData, error) {
	s.logger.Info("Getting landing page data")

	// Get hero section
	hero, err := s.GetHeroSection(ctx)
	if err != nil {
		s.logger.Warn("Failed to get hero section, using default", zap.Error(err))
		hero = s.getDefaultHero()
	}

	// Get features
	features, err := s.GetFeatures(ctx)
	if err != nil {
		s.logger.Warn("Failed to get features, using default", zap.Error(err))
		features = s.getDefaultFeatures()
	}

	// Get pricing plans
	pricingPlans, err := s.GetPricingPlans(ctx)
	if err != nil {
		s.logger.Warn("Failed to get pricing plans, using default", zap.Error(err))
		pricingPlans = s.getDefaultPricingPlans()
	}

	// Get testimonials
	testimonials, err := s.GetTestimonials(ctx)
	if err != nil {
		s.logger.Warn("Failed to get testimonials, using default", zap.Error(err))
		testimonials = s.getDefaultTestimonials()
	}

	// Get FAQs
	faqs, err := s.GetFAQs(ctx)
	if err != nil {
		s.logger.Warn("Failed to get FAQs, using default", zap.Error(err))
		faqs = s.getDefaultFAQs()
	}

	// Parse social links
	var socialLinks map[string]string
	if s.config.Landing.SocialLinks != "" {
		if err := json.Unmarshal([]byte(s.config.Landing.SocialLinks), &socialLinks); err != nil {
			s.logger.Warn("Failed to parse social links", zap.Error(err))
			socialLinks = make(map[string]string)
		}
	} else {
		socialLinks = make(map[string]string)
	}

	return &LandingPageData{
		SiteName:        s.config.Landing.SiteName,
		SiteURL:         s.config.Landing.SiteURL,
		SiteDescription: s.config.Landing.SiteDescription,
		SiteKeywords:    s.config.Landing.SiteKeywords,
		ContactEmail:    s.config.Landing.ContactEmail,
		SupportEmail:    s.config.Landing.SupportEmail,
		SocialLinks:     socialLinks,
		AnalyticsID:     s.config.Landing.AnalyticsID,
		OGImage:         s.config.Landing.OGImage,
		Favicon:         s.config.Landing.Favicon,
		CustomCSS:       s.config.Landing.CustomCSS,
		CustomJS:        s.config.Landing.CustomJS,
		CurrentYear:     time.Now().Year(),
		Hero:            hero,
		Features:        features,
		PricingPlans:    pricingPlans,
		Testimonials:    testimonials,
		FAQs:            faqs,
	}, nil
}

// Hero Section Management

// GetHeroSection retrieves the active hero section.
func (s *LandingService) GetHeroSection(ctx context.Context) (*models.HeroSection, error) {
	s.logger.Info("Getting hero section")

	// If no database, return default data
	if s.db == nil {
		s.logger.Debug("No database available, returning default hero section")
		return s.getDefaultHero(), nil
	}

	var hero models.HeroSection
	if err := s.db.Where("status = ?", "active").Order("order ASC").First(&hero).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active hero section found")
		}
		s.logger.Error("Failed to get hero section", zap.Error(err))
		return nil, fmt.Errorf("failed to get hero section: %w", err)
	}

	return &hero, nil
}

// CreateHeroSection creates a new hero section.
func (s *LandingService) CreateHeroSection(ctx context.Context, hero *models.HeroSection) error {
	s.logger.Info("Creating hero section", zap.String("title", hero.Title))

	if err := s.db.Create(hero).Error; err != nil {
		s.logger.Error("Failed to create hero section", zap.Error(err))
		return fmt.Errorf("failed to create hero section: %w", err)
	}

	s.logger.Info("Hero section created successfully", zap.Uint("hero_id", hero.ID))
	return nil
}

// UpdateHeroSection updates a hero section.
func (s *LandingService) UpdateHeroSection(ctx context.Context, heroID uint, updates *models.HeroSection) error {
	s.logger.Info("Updating hero section", zap.Uint("hero_id", heroID))

	if err := s.db.Model(&models.HeroSection{}).Where("id = ?", heroID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update hero section", zap.Error(err))
		return fmt.Errorf("failed to update hero section: %w", err)
	}

	s.logger.Info("Hero section updated successfully", zap.Uint("hero_id", heroID))
	return nil
}

// Feature Management

// GetFeatures retrieves all active features.
func (s *LandingService) GetFeatures(ctx context.Context) ([]*models.FeatureSection, error) {
	s.logger.Info("Getting features")

	// If no database, return default data
	if s.db == nil {
		s.logger.Debug("No database available, returning default features")
		return s.getDefaultFeatures(), nil
	}

	var features []*models.FeatureSection
	if err := s.db.Where("status = ?", "active").Order("order ASC").Find(&features).Error; err != nil {
		s.logger.Error("Failed to get features", zap.Error(err))
		return nil, fmt.Errorf("failed to get features: %w", err)
	}

	return features, nil
}

// CreateFeature creates a new feature.
func (s *LandingService) CreateFeature(ctx context.Context, feature *models.FeatureSection) error {
	s.logger.Info("Creating feature", zap.String("title", feature.Title))

	if err := s.db.Create(feature).Error; err != nil {
		s.logger.Error("Failed to create feature", zap.Error(err))
		return fmt.Errorf("failed to create feature: %w", err)
	}

	s.logger.Info("Feature created successfully", zap.Uint("feature_id", feature.ID))
	return nil
}

// UpdateFeature updates a feature.
func (s *LandingService) UpdateFeature(ctx context.Context, featureID uint, updates *models.FeatureSection) error {
	s.logger.Info("Updating feature", zap.Uint("feature_id", featureID))

	if err := s.db.Model(&models.FeatureSection{}).Where("id = ?", featureID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update feature", zap.Error(err))
		return fmt.Errorf("failed to update feature: %w", err)
	}

	s.logger.Info("Feature updated successfully", zap.Uint("feature_id", featureID))
	return nil
}

// Pricing Plan Management

// GetPricingPlans retrieves all active pricing plans.
func (s *LandingService) GetPricingPlans(ctx context.Context) ([]*PricingPlanData, error) {
	s.logger.Info("Getting pricing plans")

	// If no database, return default data
	if s.db == nil {
		s.logger.Debug("No database available, returning default pricing plans")
		return s.getDefaultPricingPlans(), nil
	}

	var plans []*models.PricingPlan
	if err := s.db.Where("status = ? AND is_active = ?", "active", true).Order("order ASC").Find(&plans).Error; err != nil {
		s.logger.Error("Failed to get pricing plans", zap.Error(err))
		return nil, fmt.Errorf("failed to get pricing plans: %w", err)
	}

	// Convert to PricingPlanData with parsed features
	var planData []*PricingPlanData
	for _, plan := range plans {
		var featuresList []string
		if plan.Features != "" {
			if err := json.Unmarshal([]byte(plan.Features), &featuresList); err != nil {
				s.logger.Warn("Failed to parse plan features", zap.Error(err), zap.Uint("plan_id", plan.ID))
				featuresList = []string{}
			}
		}

		planData = append(planData, &PricingPlanData{
			PricingPlan:  *plan,
			FeaturesList: featuresList,
		})
	}

	return planData, nil
}

// SyncPricingPlans syncs pricing plans from the SaaS Admin Service.
func (s *LandingService) SyncPricingPlans(ctx context.Context, plans []*models.PricingPlan) error {
	s.logger.Info("Syncing pricing plans", zap.Int("count", len(plans)))

	// Clear existing plans
	if err := s.db.Where("1 = 1").Delete(&models.PricingPlan{}).Error; err != nil {
		s.logger.Error("Failed to clear existing pricing plans", zap.Error(err))
		return fmt.Errorf("failed to clear existing pricing plans: %w", err)
	}

	// Insert new plans
	for _, plan := range plans {
		if err := s.db.Create(plan).Error; err != nil {
			s.logger.Error("Failed to create pricing plan", zap.Error(err), zap.String("plan_name", plan.Name))
			return fmt.Errorf("failed to create pricing plan: %w", err)
		}
	}

	s.logger.Info("Pricing plans synced successfully", zap.Int("count", len(plans)))
	return nil
}

// Testimonial Management

// GetTestimonials retrieves all active testimonials.
func (s *LandingService) GetTestimonials(ctx context.Context) ([]*models.Testimonial, error) {
	s.logger.Info("Getting testimonials")

	// If no database, return default data
	if s.db == nil {
		s.logger.Debug("No database available, returning default testimonials")
		return s.getDefaultTestimonials(), nil
	}

	var testimonials []*models.Testimonial
	if err := s.db.Where("status = ?", "active").Order("`order` ASC").Find(&testimonials).Error; err != nil {
		s.logger.Error("Failed to get testimonials", zap.Error(err))
		return nil, fmt.Errorf("failed to get testimonials: %w", err)
	}

	return testimonials, nil
}

// CreateTestimonial creates a new testimonial.
func (s *LandingService) CreateTestimonial(ctx context.Context, testimonial *models.Testimonial) error {
	s.logger.Info("Creating testimonial", zap.String("name", testimonial.Name))

	if s.db == nil {
		s.logger.Warn("No database available, cannot create testimonial")
		return fmt.Errorf("database not available")
	}

	if err := s.db.Create(testimonial).Error; err != nil {
		s.logger.Error("Failed to create testimonial", zap.Error(err))
		return fmt.Errorf("failed to create testimonial: %w", err)
	}

	s.logger.Info("Testimonial created successfully", zap.Uint("testimonial_id", testimonial.ID))
	return nil
}

// UpdateTestimonial updates a testimonial.
func (s *LandingService) UpdateTestimonial(ctx context.Context, testimonialID uint, updates *models.Testimonial) error {
	s.logger.Info("Updating testimonial", zap.Uint("testimonial_id", testimonialID))

	if err := s.db.Model(&models.Testimonial{}).Where("id = ?", testimonialID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update testimonial", zap.Error(err))
		return fmt.Errorf("failed to update testimonial: %w", err)
	}

	s.logger.Info("Testimonial updated successfully", zap.Uint("testimonial_id", testimonialID))
	return nil
}

// FAQ Management

// GetFAQs retrieves all active FAQs.
func (s *LandingService) GetFAQs(ctx context.Context) ([]*models.FAQ, error) {
	s.logger.Info("Getting FAQs")

	// If no database, return default data
	if s.db == nil {
		s.logger.Debug("No database available, returning default FAQs")
		return s.getDefaultFAQs(), nil
	}

	var faqs []*models.FAQ
	if err := s.db.Where("status = ?", "active").Order("`order` ASC").Find(&faqs).Error; err != nil {
		s.logger.Error("Failed to get FAQs", zap.Error(err))
		return nil, fmt.Errorf("failed to get FAQs: %w", err)
	}

	return faqs, nil
}

// CreateFAQ creates a new FAQ.
func (s *LandingService) CreateFAQ(ctx context.Context, faq *models.FAQ) error {
	s.logger.Info("Creating FAQ", zap.String("question", faq.Question))

	if s.db == nil {
		s.logger.Warn("No database available, cannot create FAQ")
		return fmt.Errorf("database not available")
	}

	if err := s.db.Create(faq).Error; err != nil {
		s.logger.Error("Failed to create FAQ", zap.Error(err))
		return fmt.Errorf("failed to create FAQ: %w", err)
	}

	s.logger.Info("FAQ created successfully", zap.Uint("faq_id", faq.ID))
	return nil
}

// UpdateFAQ updates a FAQ.
func (s *LandingService) UpdateFAQ(ctx context.Context, faqID uint, updates *models.FAQ) error {
	s.logger.Info("Updating FAQ", zap.Uint("faq_id", faqID))

	if err := s.db.Model(&models.FAQ{}).Where("id = ?", faqID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update FAQ", zap.Error(err))
		return fmt.Errorf("failed to update FAQ: %w", err)
	}

	s.logger.Info("FAQ updated successfully", zap.Uint("faq_id", faqID))
	return nil
}

// Article Management

// GetArticles retrieves published articles.
func (s *LandingService) GetArticles(ctx context.Context, limit, offset int) ([]*models.Article, error) {
	s.logger.Info("Getting articles", zap.Int("limit", limit), zap.Int("offset", offset))

	var articles []*models.Article
	query := s.db.Where("status = ?", "published").Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&articles).Error; err != nil {
		s.logger.Error("Failed to get articles", zap.Error(err))
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	return articles, nil
}

// GetArticle retrieves an article by slug.
func (s *LandingService) GetArticle(ctx context.Context, slug string) (*models.Article, error) {
	s.logger.Info("Getting article", zap.String("slug", slug))

	var article models.Article
	if err := s.db.Where("slug = ? AND status = ?", slug, "published").First(&article).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("article not found")
		}
		s.logger.Error("Failed to get article", zap.Error(err))
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	// Increment view count
	s.db.Model(&article).Update("view_count", gorm.Expr("view_count + 1"))

	return &article, nil
}

// CreateArticle creates a new article.
func (s *LandingService) CreateArticle(ctx context.Context, article *models.Article) error {
	s.logger.Info("Creating article", zap.String("title", article.Title))

	if s.db == nil {
		s.logger.Warn("No database available, cannot create article")
		return fmt.Errorf("database not available")
	}

	if err := s.db.Create(article).Error; err != nil {
		s.logger.Error("Failed to create article", zap.Error(err))
		return fmt.Errorf("failed to create article: %w", err)
	}

	s.logger.Info("Article created successfully", zap.Uint("article_id", article.ID))
	return nil
}

// UpdateArticle updates an article.
func (s *LandingService) UpdateArticle(ctx context.Context, articleID uint, updates *models.Article) error {
	s.logger.Info("Updating article", zap.Uint("article_id", articleID))

	if err := s.db.Model(&models.Article{}).Where("id = ?", articleID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update article", zap.Error(err))
		return fmt.Errorf("failed to update article: %w", err)
	}

	s.logger.Info("Article updated successfully", zap.Uint("article_id", articleID))
	return nil
}

// DeleteArticle deletes an article.
func (s *LandingService) DeleteArticle(ctx context.Context, articleID uint) error {
	s.logger.Info("Deleting article", zap.Uint("article_id", articleID))

	if err := s.db.Delete(&models.Article{}, articleID).Error; err != nil {
		s.logger.Error("Failed to delete article", zap.Error(err))
		return fmt.Errorf("failed to delete article: %w", err)
	}

	s.logger.Info("Article deleted successfully", zap.Uint("article_id", articleID))
	return nil
}

// Contact Form Management

// SubmitContactForm submits a contact form.
func (s *LandingService) SubmitContactForm(ctx context.Context, form *models.ContactForm) error {
	s.logger.Info("Submitting contact form", zap.String("email", form.Email))

	if s.db == nil {
		s.logger.Warn("No database available, cannot submit contact form")
		return fmt.Errorf("database not available")
	}

	if err := s.db.Create(form).Error; err != nil {
		s.logger.Error("Failed to submit contact form", zap.Error(err))
		return fmt.Errorf("failed to submit contact form: %w", err)
	}

	s.logger.Info("Contact form submitted successfully", zap.Uint("form_id", form.ID))
	return nil
}

// Newsletter Management

// SubscribeNewsletter subscribes to the newsletter.
func (s *LandingService) SubscribeNewsletter(ctx context.Context, subscription *models.Newsletter) error {
	s.logger.Info("Subscribing to newsletter", zap.String("email", subscription.Email))

	if s.db == nil {
		s.logger.Warn("No database available, cannot subscribe to newsletter")
		return fmt.Errorf("database not available")
	}

	if err := s.db.Create(subscription).Error; err != nil {
		s.logger.Error("Failed to subscribe to newsletter", zap.Error(err))
		return fmt.Errorf("failed to subscribe to newsletter: %w", err)
	}

	s.logger.Info("Newsletter subscription successful", zap.Uint("subscription_id", subscription.ID))
	return nil
}

// Health checks the health of the landing service.
func (s *LandingService) Health(ctx context.Context) error {
	s.logger.Debug("Checking landing service health")

	// Check database connection if available
	if s.db != nil {
		if err := s.db.Exec("SELECT 1").Error; err != nil {
			s.logger.Error("Landing service health check failed", zap.Error(err))
			return fmt.Errorf("landing service health check failed: %w", err)
		}
	} else {
		s.logger.Debug("Database not available, skipping database health check")
	}

	return nil
}

// Default data methods

func (s *LandingService) getDefaultHero() *models.HeroSection {
	return &models.HeroSection{
		Title:       "Keep your users informed with beautiful status pages",
		Subtitle:    "Professional status page platform for modern teams",
		Description: "Create stunning status pages, manage incidents, and keep your users updated with real-time notifications. Trusted by thousands of teams worldwide.",
		ButtonText:  "Start Free Trial",
		ButtonURL:   "/signup",
		ImageURL:    "/static/images/hero-dashboard.png",
		Status:      "active",
		Order:       0,
	}
}

func (s *LandingService) getDefaultFeatures() []*models.FeatureSection {
	return []*models.FeatureSection{
		{
			Title:       "Real-time Status Updates",
			Description: "Keep your users informed with instant status updates and incident notifications.",
			Icon:        "fas fa-broadcast-tower",
			Status:      "active",
			Order:       0,
		},
		{
			Title:       "Beautiful Custom Pages",
			Description: "Create stunning status pages that match your brand with our powerful customization tools.",
			Icon:        "fas fa-palette",
			Status:      "active",
			Order:       1,
		},
		{
			Title:       "Advanced Analytics",
			Description: "Get insights into your uptime, incident patterns, and user engagement with detailed analytics.",
			Icon:        "fas fa-chart-line",
			Status:      "active",
			Order:       2,
		},
		{
			Title:       "Team Collaboration",
			Description: "Work together seamlessly with your team to manage incidents and communicate updates.",
			Icon:        "fas fa-users",
			Status:      "active",
			Order:       3,
		},
		{
			Title:       "API Integration",
			Description: "Integrate with your existing tools and workflows using our comprehensive REST API.",
			Icon:        "fas fa-code",
			Status:      "active",
			Order:       4,
		},
		{
			Title:       "24/7 Support",
			Description: "Get help when you need it with our dedicated support team available around the clock.",
			Icon:        "fas fa-headset",
			Status:      "active",
			Order:       5,
		},
	}
}

func (s *LandingService) getDefaultPricingPlans() []*PricingPlanData {
	return []*PricingPlanData{
		{
			PricingPlan: models.PricingPlan{
				Name:            "Free",
				Slug:            "free",
				Description:     "Perfect for small teams and personal projects",
				Price:           0,
				Currency:        "USD",
				BillingInterval: "monthly",
				ButtonText:      "Get Started",
				ButtonURL:       "/signup?plan=free",
				IsActive:        true,
				Status:          "active",
				Order:           0,
			},
			FeaturesList: []string{
				"Up to 5 services",
				"Basic status page",
				"Email notifications",
				"Community support",
			},
		},
		{
			PricingPlan: models.PricingPlan{
				Name:            "Pro",
				Slug:            "pro",
				Description:     "Advanced features for growing businesses",
				Price:           29,
				Currency:        "USD",
				BillingInterval: "monthly",
				ButtonText:      "Start Free Trial",
				ButtonURL:       "/signup?plan=pro",
				IsPopular:       true,
				IsActive:        true,
				Status:          "active",
				Order:           1,
			},
			FeaturesList: []string{
				"Up to 25 services",
				"Custom domains",
				"Advanced analytics",
				"API access",
				"Priority support",
				"Team collaboration",
			},
		},
		{
			PricingPlan: models.PricingPlan{
				Name:            "Enterprise",
				Slug:            "enterprise",
				Description:     "Complete solution for large organizations",
				Price:           99,
				Currency:        "USD",
				BillingInterval: "monthly",
				ButtonText:      "Contact Sales",
				ButtonURL:       "/contact?plan=enterprise",
				IsActive:        true,
				Status:          "active",
				Order:           2,
			},
			FeaturesList: []string{
				"Unlimited services",
				"White-label solution",
				"Advanced integrations",
				"Custom SLA monitoring",
				"Dedicated support",
				"SSO integration",
			},
		},
	}
}

func (s *LandingService) getDefaultTestimonials() []*models.Testimonial {
	return []*models.Testimonial{
		{
			Name:     "Sarah Johnson",
			Company:  "TechCorp",
			Position: "CTO",
			Avatar:   "/static/images/avatars/sarah.jpg",
			Content:  "StatusPage Pro has transformed how we communicate with our users. The beautiful interface and real-time updates have significantly improved our user satisfaction.",
			Rating:   5,
			Status:   "active",
			Order:    0,
		},
		{
			Name:     "Michael Chen",
			Company:  "StartupXYZ",
			Position: "Founder",
			Avatar:   "/static/images/avatars/michael.jpg",
			Content:  "The analytics and incident management features are incredible. We can now proactively address issues before they become major problems.",
			Rating:   5,
			Status:   "active",
			Order:    1,
		},
		{
			Name:     "Emily Rodriguez",
			Company:  "GlobalTech",
			Position: "DevOps Lead",
			Avatar:   "/static/images/avatars/emily.jpg",
			Content:  "The API integration and team collaboration features make it easy to keep everyone in sync. Highly recommended for any growing team.",
			Rating:   5,
			Status:   "active",
			Order:    2,
		},
	}
}

func (s *LandingService) getDefaultFAQs() []*models.FAQ {
	return []*models.FAQ{
		{
			Question: "How quickly can I set up my status page?",
			Answer:   "You can have your status page up and running in under 5 minutes. Simply sign up, add your services, and customize your page to match your brand.",
			Status:   "active",
			Order:    0,
		},
		{
			Question: "Can I use my own domain?",
			Answer:   "Yes! Pro and Enterprise plans include custom domain support. You can use your own domain like status.yourcompany.com.",
			Status:   "active",
			Order:    1,
		},
		{
			Question: "What integrations are available?",
			Answer:   "We integrate with popular tools like Slack, PagerDuty, Datadog, New Relic, and many more. We also provide a comprehensive REST API.",
			Status:   "active",
			Order:    2,
		},
		{
			Question: "Is there a free trial?",
			Answer:   "Yes! All paid plans come with a 14-day free trial. No credit card required to get started.",
			Status:   "active",
			Order:    3,
		},
		{
			Question: "What kind of support do you offer?",
			Answer:   "Free plan includes community support. Pro plan includes priority email support. Enterprise plan includes dedicated support with phone and chat options.",
			Status:   "active",
			Order:    4,
		},
	}
}

// initDatabase initializes the database connection.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.LandingPage{},
		&models.HeroSection{},
		&models.FeatureSection{},
		&models.PricingPlan{},
		&models.Testimonial{},
		&models.Article{},
		&models.FAQ{},
		&models.ContactForm{},
		&models.Newsletter{},
		&models.LandingPageStats{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
