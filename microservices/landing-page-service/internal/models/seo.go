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