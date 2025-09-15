// Package models provides data models for the Landing Page Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// LandingPage represents the main landing page configuration.
type LandingPage struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	Slug        string         `gorm:"not null;uniqueIndex" json:"slug"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Content     string         `gorm:"type:text" json:"content"`     // HTML content
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// HeroSection represents the hero section configuration.
type HeroSection struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title           string         `gorm:"not null" json:"title"`
	Subtitle        string         `gorm:"type:text" json:"subtitle"`
	Description     string         `gorm:"type:text" json:"description"`
	ButtonText      string         `json:"button_text"`
	ButtonURL       string         `json:"button_url"`
	ImageURL        string         `json:"image_url"`
	VideoURL        string         `json:"video_url"`
	BackgroundColor string         `json:"background_color"`
	TextColor       string         `json:"text_color"`
	Status          string         `gorm:"default:active" json:"status"` // active, inactive
	Order           int            `gorm:"default:0" json:"order"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// FeatureSection represents a feature section.
type FeatureSection struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `json:"icon"`
	ImageURL    string         `json:"image_url"`
	Status      string         `gorm:"default:active" json:"status"` // active, inactive
	Order       int            `gorm:"default:0" json:"order"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// PricingPlan represents a pricing plan (synced from SaaS Admin Service).
type PricingPlan struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PlanID          uint           `gorm:"not null;uniqueIndex" json:"plan_id"` // Reference to SaaS Admin Service
	Name            string         `gorm:"not null" json:"name"`
	Slug            string         `gorm:"not null;uniqueIndex" json:"slug"`
	Description     string         `gorm:"type:text" json:"description"`
	Price           float64        `gorm:"not null" json:"price"`
	Currency        string         `gorm:"default:USD" json:"currency"`
	BillingInterval string         `gorm:"default:monthly" json:"billing_interval"` // monthly, yearly
	Features        string         `gorm:"type:text" json:"features"`               // JSON array of features
	IsPopular       bool           `gorm:"default:false" json:"is_popular"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	ButtonText      string         `json:"button_text"`
	ButtonURL       string         `json:"button_url"`
	Status          string         `gorm:"default:active" json:"status"` // active, inactive
	Order           int            `gorm:"default:0" json:"order"`
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// SaaSPlan represents a SaaS plan from the SaaS Admin Service (for API communication).
type SaaSPlan struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name            string         `gorm:"not null;uniqueIndex" json:"name"`
	Slug            string         `gorm:"not null;uniqueIndex" json:"slug"`
	Description     string         `gorm:"type:text" json:"description"`
	Price           float64        `gorm:"not null" json:"price"`
	Currency        string         `gorm:"default:USD" json:"currency"`
	BillingInterval string         `gorm:"default:monthly" json:"billing_interval"` // monthly, yearly
	MaxTenants      int            `gorm:"default:1" json:"max_tenants"`
	MaxUsers        int            `gorm:"default:5" json:"max_users"`
	MaxServices     int            `gorm:"default:10" json:"max_services"`
	MaxMonitors     int            `gorm:"default:50" json:"max_monitors"`
	MaxSubscribers  int            `gorm:"default:1000" json:"max_subscribers"`
	MaxIncidents    int            `gorm:"default:100" json:"max_incidents"`
	MaxMaintenance  int            `gorm:"default:50" json:"max_maintenance"`
	CustomDomain    bool           `gorm:"default:false" json:"custom_domain"`
	WhiteLabel      bool           `gorm:"default:false" json:"white_label"`
	API             bool           `gorm:"default:false" json:"api"`
	Integrations    bool           `gorm:"default:false" json:"integrations"`
	Analytics       bool           `gorm:"default:false" json:"analytics"`
	Support         string         `gorm:"default:email" json:"support"` // email, chat, phone
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	IsPublic        bool           `gorm:"default:true" json:"is_public"`
	IsPopular       bool           `gorm:"default:false" json:"is_popular"` // For frontend display
	ButtonText      string         `gorm:"default:Get Started" json:"button_text"`
	ButtonURL       string         `gorm:"default:/signup" json:"button_url"`
	Order           int            `gorm:"default:0" json:"order"`    // Display order
	Features        string         `gorm:"type:text" json:"features"` // JSON array of features
	Limits          string         `gorm:"type:text" json:"limits"`   // JSON object of limits
	Metadata        string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Testimonial represents a customer testimonial.
type Testimonial struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null" json:"name"`
	Company   string         `json:"company"`
	Position  string         `json:"position"`
	Avatar    string         `json:"avatar"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Rating    int            `gorm:"default:5" json:"rating"`      // 1-5 stars
	Status    string         `gorm:"default:active" json:"status"` // active, inactive
	Order     int            `gorm:"default:0" json:"order"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Article represents a blog/article.
type Article struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title         string         `gorm:"not null" json:"title"`
	Slug          string         `gorm:"not null;uniqueIndex" json:"slug"`
	Excerpt       string         `gorm:"type:text" json:"excerpt"`
	Content       string         `gorm:"type:text;not null" json:"content"` // HTML content
	Author        string         `gorm:"not null" json:"author"`
	AuthorEmail   string         `json:"author_email"`
	AuthorAvatar  string         `json:"author_avatar"`
	FeaturedImage string         `json:"featured_image"`
	Category      string         `gorm:"index" json:"category"`
	Tags          string         `gorm:"type:text" json:"tags"`       // JSON array of tags
	Status        string         `gorm:"default:draft" json:"status"` // draft, published, archived
	IsFeatured    bool           `gorm:"default:false" json:"is_featured"`
	ViewCount     int            `gorm:"default:0" json:"view_count"`
	PublishedAt   *time.Time     `json:"published_at"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// FAQ represents a frequently asked question.
type FAQ struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Question  string         `gorm:"not null" json:"question"`
	Answer    string         `gorm:"type:text;not null" json:"answer"`
	Category  string         `gorm:"index" json:"category"`
	Status    string         `gorm:"default:active" json:"status"` // active, inactive
	Order     int            `gorm:"default:0" json:"order"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ContactForm represents a contact form submission.
type ContactForm struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"not null" json:"email"`
	Company   string         `json:"company"`
	Subject   string         `json:"subject"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Status    string         `gorm:"default:pending" json:"status"` // pending, read, replied, archived
	IPAddress string         `json:"ip_address"`
	UserAgent string         `gorm:"type:text" json:"user_agent"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Newsletter represents a newsletter subscription.
type Newsletter struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Email     string         `gorm:"not null;uniqueIndex" json:"email"`
	Name      string         `json:"name"`
	Status    string         `gorm:"default:active" json:"status"` // active, unsubscribed, bounced
	Source    string         `json:"source"`                       // landing page, footer, popup, etc.
	IPAddress string         `json:"ip_address"`
	UserAgent string         `gorm:"type:text" json:"user_agent"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// LandingPageStats represents landing page statistics.
type LandingPageStats struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	Date              time.Time `gorm:"not null;index" json:"date"`
	PageViews         int       `gorm:"default:0" json:"page_views"`
	UniqueVisitors    int       `gorm:"default:0" json:"unique_visitors"`
	ContactForms      int       `gorm:"default:0" json:"contact_forms"`
	NewsletterSignups int       `gorm:"default:0" json:"newsletter_signups"`
	PlanViews         int       `gorm:"default:0" json:"plan_views"`
	PlanClicks        int       `gorm:"default:0" json:"plan_clicks"`
	ArticleViews      int       `gorm:"default:0" json:"article_views"`
	BounceRate        float64   `gorm:"default:0" json:"bounce_rate"`
	AvgSessionTime    float64   `gorm:"default:0" json:"avg_session_time"`
	Metadata          string    `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPage.
func (LandingPage) TableName() string {
	return "landing_pages"
}

// TableName returns the table name for HeroSection.
func (HeroSection) TableName() string {
	return "hero_sections"
}

// TableName returns the table name for FeatureSection.
func (FeatureSection) TableName() string {
	return "feature_sections"
}

// TableName returns the table name for PricingPlan.
func (PricingPlan) TableName() string {
	return "pricing_plans"
}

// TableName returns the table name for Testimonial.
func (Testimonial) TableName() string {
	return "testimonials"
}

// TableName returns the table name for Article.
func (Article) TableName() string {
	return "articles"
}

// TableName returns the table name for FAQ.
func (FAQ) TableName() string {
	return "faqs"
}

// TableName returns the table name for ContactForm.
func (ContactForm) TableName() string {
	return "contact_forms"
}

// TableName returns the table name for Newsletter.
func (Newsletter) TableName() string {
	return "newsletters"
}

// TableName returns the table name for LandingPageStats.
func (LandingPageStats) TableName() string {
	return "landing_page_stats"
}
