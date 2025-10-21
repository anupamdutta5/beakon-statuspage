// Package models contains data structures for the Landing Page Service.
package models

import (
	"encoding/xml"
	"time"
)

// SEOPageData contains all data needed for SEO optimization
type SEOPageData struct {
	// Basic page info
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    []string `json:"keywords"`
	Path        string `json:"path"`
	Type        string `json:"type"` // homepage, article, product, faq, contact, etc.
	Language    string `json:"language"`

	// Content metadata
	Author      string     `json:"author"`
	PublishedAt *time.Time `json:"published_at"`
	ModifiedAt  *time.Time `json:"modified_at"`
	Category    string     `json:"category"`

	// Images and media
	Image       string   `json:"image"`
	Images      []string `json:"images"`
	Video       string   `json:"video"`

	// Social media overrides
	OGTitle       string `json:"og_title"`
	OGDescription string `json:"og_description"`
	OGImage       string `json:"og_image"`

	// SEO settings
	NoIndex    bool `json:"no_index"`
	NoFollow   bool `json:"no_follow"`
	Canonical  string `json:"canonical"`

	// Structured data
	FAQs []FAQ `json:"faqs"`

	// Article specific
	Content     string `json:"content"`
	ReadingTime int    `json:"reading_time"` // in minutes

	// Product specific
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	SKU      string  `json:"sku"`

	// Rating/Reviews
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"review_count"`
}

// SEOMetaTags contains all meta tags for a page
type SEOMetaTags struct {
	// Basic HTML meta tags
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	Canonical   string `json:"canonical"`
	Language    string `json:"language"`

	// Open Graph tags
	OGTitle       string `json:"og_title"`
	OGDescription string `json:"og_description"`
	OGImage       string `json:"og_image"`
	OGType        string `json:"og_type"`
	OGURL         string `json:"og_url"`
	OGSiteName    string `json:"og_site_name"`
	OGLocale      string `json:"og_locale"`

	// Twitter Card tags
	TwitterCard        string `json:"twitter_card"`
	TwitterTitle       string `json:"twitter_title"`
	TwitterDescription string `json:"twitter_description"`
	TwitterImage       string `json:"twitter_image"`
	TwitterSite        string `json:"twitter_site"`
	TwitterCreator     string `json:"twitter_creator"`

	// Additional meta tags
	Robots                 string `json:"robots"`
	Viewport               string `json:"viewport"`
	ThemeColor             string `json:"theme_color"`
	MSApplicationTileColor string `json:"ms_application_tile_color"`

	// Article-specific meta tags
	ArticleAuthor       string     `json:"article_author"`
	ArticlePublishedTime *time.Time `json:"article_published_time"`
	ArticleModifiedTime  *time.Time `json:"article_modified_time"`
	ArticleSection       string     `json:"article_section"`
	ArticleTags          []string   `json:"article_tags"`

	// Structured data
	StructuredData string `json:"structured_data"`
}

// SitemapPage represents a page in the sitemap
type SitemapPage struct {
	Path         string    `json:"path"`
	LastModified time.Time `json:"last_modified"`
	ChangeFreq   string    `json:"change_freq"` // always, hourly, daily, weekly, monthly, yearly, never
	Priority     float64   `json:"priority"`    // 0.0 to 1.0
}

// Sitemap represents the XML sitemap structure
type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapURL represents a URL in the sitemap
type SitemapURL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod"`
	ChangeFreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

// WebVitalsData contains Core Web Vitals metrics
type WebVitalsData struct {
	LCP  float64 `json:"lcp"`  // Largest Contentful Paint
	FID  float64 `json:"fid"`  // First Input Delay
	CLS  float64 `json:"cls"`  // Cumulative Layout Shift
	FCP  float64 `json:"fcp"`  // First Contentful Paint
	TTI  float64 `json:"tti"`  // Time to Interactive
	TBT  float64 `json:"tbt"`  // Total Blocking Time
	SI   float64 `json:"si"`   // Speed Index
}

// SEOAnalytics contains SEO performance metrics
type SEOAnalytics struct {
	PageURL          string        `json:"page_url"`
	Title            string        `json:"title"`
	MetaDescription  string        `json:"meta_description"`
	WordCount        int           `json:"word_count"`
	ReadabilityScore float64       `json:"readability_score"`
	WebVitals        WebVitalsData `json:"web_vitals"`

	// SEO scores
	TitleScore       int `json:"title_score"`        // 0-100
	DescriptionScore int `json:"description_score"`  // 0-100
	KeywordDensity   map[string]float64 `json:"keyword_density"`

	// Technical SEO
	HasH1            bool   `json:"has_h1"`
	H1Count          int    `json:"h1_count"`
	HasMetaDesc      bool   `json:"has_meta_desc"`
	MetaDescLength   int    `json:"meta_desc_length"`
	HasCanonical     bool   `json:"has_canonical"`
	HasOGTags        bool   `json:"has_og_tags"`
	HasStructuredData bool  `json:"has_structured_data"`
	ImageAltMissing  int    `json:"image_alt_missing"`

	// Performance
	PageSize         int64 `json:"page_size"`
	LoadTime         float64 `json:"load_time"`

	// Links
	InternalLinks    int `json:"internal_links"`
	ExternalLinks    int `json:"external_links"`

	Timestamp        time.Time `json:"timestamp"`
}

// SEORecommendation represents an SEO improvement suggestion
type SEORecommendation struct {
	Type        string `json:"type"`        // title, description, content, technical, performance
	Priority    string `json:"priority"`    // high, medium, low
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`      // high, medium, low
	Effort      string `json:"effort"`      // high, medium, low
	Category    string `json:"category"`    // on-page, technical, content, performance
}

// SEOAuditResult contains the complete SEO audit for a page
type SEOAuditResult struct {
	PageURL       string              `json:"page_url"`
	OverallScore  int                 `json:"overall_score"` // 0-100
	Analytics     SEOAnalytics        `json:"analytics"`
	Recommendations []SEORecommendation `json:"recommendations"`
	Timestamp     time.Time           `json:"timestamp"`

	// Scores by category
	ContentScore    int `json:"content_score"`
	TechnicalScore  int `json:"technical_score"`
	PerformanceScore int `json:"performance_score"`
	UserExperienceScore int `json:"user_experience_score"`
}

// RichSnippet represents structured data for rich snippets
type RichSnippet struct {
	Type string                 `json:"type"` // Article, Product, FAQ, Organization, etc.
	Data map[string]interface{} `json:"data"`
}

// SEOContent represents optimized content with SEO metadata
type SEOContent struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Type        string    `json:"type"`        // page, article, product
	Slug        string    `json:"slug" gorm:"uniqueIndex"`
	Title       string    `json:"title"`
	Content     string    `json:"content" gorm:"type:text"`
	Excerpt     string    `json:"excerpt"`

	// SEO fields
	SEOTitle       string   `json:"seo_title"`
	SEODescription string   `json:"seo_description"`
	SEOKeywords    []string `json:"seo_keywords" gorm:"serializer:json"`
	FocusKeyword   string   `json:"focus_keyword"`

	// Meta fields
	Author         string    `json:"author"`
	PublishedAt    *time.Time `json:"published_at"`
	LastModified   time.Time  `json:"last_modified"`
	Status         string     `json:"status"` // draft, published, archived

	// Performance
	ViewCount      int     `json:"view_count"`
	SEOScore       int     `json:"seo_score"`
	ReadingTime    int     `json:"reading_time"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SEORedirect represents URL redirects for SEO
type SEORedirect struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FromURL     string    `json:"from_url" gorm:"uniqueIndex"`
	ToURL       string    `json:"to_url"`
	StatusCode  int       `json:"status_code"` // 301, 302, 307, 308
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================================================
// Database-Backed Models for Landing Page Enhancement
// ============================================================================

// LandingPageSEOConfig represents global SEO configuration (singleton).
type LandingPageSEOConfig struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Global SEO settings
	SiteName                string `gorm:"not null" json:"site_name"`
	SiteURL                 string `gorm:"not null" json:"site_url"`
	DefaultMetaDescription  string `gorm:"type:text" json:"default_meta_description"`
	DefaultOGImage          string `json:"default_og_image"`
	GoogleAnalyticsID       string `json:"google_analytics_id"`
	GoogleTagManagerID      string `json:"google_tag_manager_id"`
	FacebookPixelID         string `json:"facebook_pixel_id"`
	PlausibleDomain         string `json:"plausible_domain"`

	// Schema.org Organization
	OrganizationName        string `json:"organization_name"`
	OrganizationLogo        string `json:"organization_logo"`
	OrganizationURL         string `json:"organization_url"`
	OrganizationDescription string `gorm:"type:text" json:"organization_description"`
	OrganizationEmail       string `json:"organization_email"`
	OrganizationPhone       string `json:"organization_phone"`
	OrganizationAddress     string `gorm:"type:text" json:"organization_address"`     // JSON: {street, city, state, zip, country}
	OrganizationSocialLinks string `gorm:"type:text" json:"organization_social_links"` // JSON: {twitter, linkedin, facebook, ...}

	// Sitemap settings
	SitemapEnabled    bool    `gorm:"default:true" json:"sitemap_enabled"`
	SitemapChangeFreq string  `gorm:"default:'daily'" json:"sitemap_change_freq"` // always, hourly, daily, weekly, monthly, yearly, never
	SitemapPriority   float32 `gorm:"default:0.8" json:"sitemap_priority"`

	// Robots.txt
	RobotsTxt     string `gorm:"type:text" json:"robots_txt"`
	RobotsEnabled bool   `gorm:"default:true" json:"robots_enabled"`

	// Schema
	SchemaEnabled bool `gorm:"default:true" json:"schema_enabled"`

	// Twitter Card
	TwitterCard string `json:"twitter_card"`

	// Default SEO values
	DefaultTitle         string `json:"default_title"`
	DefaultKeywords      string `gorm:"type:text" json:"default_keywords"`
	OGDefaultTitle       string `json:"og_default_title"`
	OGDefaultDescription string `gorm:"type:text" json:"og_default_description"`

	// Verification codes
	GoogleSiteVerification string `json:"google_site_verification"`
	BingSiteVerification   string `json:"bing_site_verification"`

	Metadata string `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPageSEOConfig.
func (LandingPageSEOConfig) TableName() string {
	return "landing_page_seo_config"
}

// LandingPageMedia represents uploaded media files.
type LandingPageMedia struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Filename         string `gorm:"not null" json:"filename"`
	OriginalFilename string `gorm:"not null" json:"original_filename"`
	MimeType         string `gorm:"not null" json:"mime_type"`
	SizeBytes        int64  `gorm:"not null" json:"size_bytes"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`

	// Storage
	StoragePath string `gorm:"not null" json:"storage_path"` // /static/uploads/2025/10/...
	CDNURL      string `json:"cdn_url"`                      // CDN URL if using CDN

	// Optimization
	IsOptimized   bool   `gorm:"default:false" json:"is_optimized"`
	WebPPath      string `json:"webp_path"`      // WebP version
	AVIFPath      string `json:"avif_path"`      // AVIF version
	ThumbnailPath string `json:"thumbnail_path"` // Thumbnail

	// Metadata
	AltText  string `gorm:"type:text" json:"alt_text"`
	Caption  string `gorm:"type:text" json:"caption"`
	Title    string `json:"title"`
	Category string `gorm:"index" json:"category"` // hero, feature, testimonial, blog, general

	// Usage tracking
	UsageCount int        `gorm:"default:0" json:"usage_count"`
	LastUsedAt *time.Time `json:"last_used_at"`

	Metadata string `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPageMedia.
func (LandingPageMedia) TableName() string {
	return "landing_page_media"
}

// LandingPageABTest represents A/B testing configuration.
type LandingPageABTest struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ElementType string `gorm:"not null;index" json:"element_type"` // hero_title, cta_button, pricing_layout, etc.
	ElementID   *uint  `gorm:"index" json:"element_id"`            // ID of the element being tested

	Status    string     `gorm:"default:'draft';index" json:"status"` // draft, running, paused, completed
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`

	// Variants
	VariantAConfig string `gorm:"type:text" json:"variant_a_config"` // JSON configuration
	VariantBConfig string `gorm:"type:text" json:"variant_b_config"` // JSON configuration
	TrafficSplit   int    `gorm:"default:50" json:"traffic_split"`   // 0-100 percentage to variant B

	// Results
	VariantAViews       int     `gorm:"default:0" json:"variant_a_views"`
	VariantAConversions int     `gorm:"default:0" json:"variant_a_conversions"`
	VariantBViews       int     `gorm:"default:0" json:"variant_b_views"`
	VariantBConversions int     `gorm:"default:0" json:"variant_b_conversions"`
	Winner              string  `json:"winner"`           // 'A', 'B', or NULL
	ConfidenceLevel     float32 `json:"confidence_level"` // 0-100

	Metadata string `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPageABTest.
func (LandingPageABTest) TableName() string {
	return "landing_page_ab_tests"
}

// ConversionRateA calculates the conversion rate for variant A.
func (t *LandingPageABTest) ConversionRateA() float64 {
	if t.VariantAViews == 0 {
		return 0
	}
	return float64(t.VariantAConversions) / float64(t.VariantAViews) * 100
}

// ConversionRateB calculates the conversion rate for variant B.
func (t *LandingPageABTest) ConversionRateB() float64 {
	if t.VariantBViews == 0 {
		return 0
	}
	return float64(t.VariantBConversions) / float64(t.VariantBViews) * 100
}

// LandingPageCTAButton represents reusable CTA buttons.
type LandingPageCTAButton struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Name string `gorm:"not null" json:"name"` // Internal name
	Text string `gorm:"not null" json:"text"`
	URL  string `gorm:"not null" json:"url"`

	// Styling
	Style string `gorm:"default:'primary'" json:"style"` // primary, secondary, outline, ghost
	Size  string `gorm:"default:'medium'" json:"size"`   // small, medium, large
	Icon  string `json:"icon"`                         // Font Awesome class

	// Tracking
	ClickCount      int `gorm:"default:0" json:"click_count"`
	ConversionCount int `gorm:"default:0" json:"conversion_count"`

	// Usage
	Locations string `gorm:"type:text" json:"locations"` // JSON array: ['hero', 'pricing', 'footer']
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	Metadata string `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPageCTAButton.
func (LandingPageCTAButton) TableName() string {
	return "landing_page_cta_buttons"
}

// LandingPageSection represents custom page sections.
type LandingPageSection struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	Name        string `gorm:"not null" json:"name"`
	SectionType string `gorm:"not null;index" json:"section_type"` // custom_html, stats, logos, cta, video, etc.
	Title       string `json:"title"`
	Content     string `gorm:"type:text" json:"content"` // HTML or JSON config

	// Display
	Position        string `gorm:"index" json:"position"` // after_hero, after_features, after_pricing, etc.
	Order           int    `gorm:"default:0" json:"order"`
	BackgroundColor string `json:"background_color"`
	Status          string `gorm:"default:'active'" json:"status"` // active, inactive

	// Layout
	Layout string `json:"layout"` // full_width, contained, split

	Metadata string `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPageSection.
func (LandingPageSection) TableName() string {
	return "landing_page_sections"
}

// LandingPageIntegration represents third-party integrations.
type LandingPageIntegration struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	IntegrationType string `gorm:"not null;uniqueIndex:idx_integration_type_provider" json:"integration_type"` // analytics, chat, email, crm, etc.
	Provider        string `gorm:"not null;uniqueIndex:idx_integration_type_provider" json:"provider"`         // google_analytics, intercom, mailchimp, etc.

	// Configuration
	IsEnabled bool   `gorm:"default:false" json:"is_enabled"`
	Config    string `gorm:"type:text" json:"config"`  // JSON configuration
	APIKey    string `gorm:"type:text" json:"api_key"` // Encrypted

	// Tracking
	LastSyncAt   *time.Time `json:"last_sync_at"`
	SyncStatus   string     `json:"sync_status"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`

	Metadata string `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// TableName returns the table name for LandingPageIntegration.
func (LandingPageIntegration) TableName() string {
	return "landing_page_integrations"
}

// ABTestAssignment tracks user assignments to A/B test variants.
type ABTestAssignment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	TestID    uint   `gorm:"not null;index:idx_test_session" json:"test_id"`
	SessionID string `gorm:"not null;index:idx_test_session" json:"session_id"`
	Variant   string `gorm:"not null" json:"variant"` // 'A' or 'B'
	UserID    string `gorm:"index" json:"user_id"`    // If authenticated
	Converted bool   `gorm:"default:false" json:"converted"`
}

// TableName returns the table name for ABTestAssignment.
func (ABTestAssignment) TableName() string {
	return "ab_test_assignments"
}