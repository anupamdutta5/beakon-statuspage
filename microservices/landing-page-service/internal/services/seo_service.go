// Package services provides business logic for the Landing Page Service.
package services

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"go.uber.org/zap"
)

// SEOService handles all SEO-related functionality
type SEOService struct {
	logger *zap.Logger
	config *SEOConfig
}

// SEOConfig contains SEO configuration
type SEOConfig struct {
	SiteURL     string
	SiteName    string
	DefaultLang string
	Analytics   AnalyticsConfig
}

// AnalyticsConfig contains analytics configuration
type AnalyticsConfig struct {
	GoogleAnalyticsID string
	GoogleTagManager  string
	FacebookPixelID   string
	LinkedInPartnerID string
}

// NewSEOService creates a new SEO service
func NewSEOService(logger *zap.Logger, config *SEOConfig) *SEOService {
	return &SEOService{
		logger: logger,
		config: config,
	}
}

// GenerateMetaTags generates comprehensive meta tags for a page
func (s *SEOService) GenerateMetaTags(ctx context.Context, pageData *models.SEOPageData) (*models.SEOMetaTags, error) {
	meta := &models.SEOMetaTags{
		Title:       s.generateTitle(pageData),
		Description: s.generateDescription(pageData),
		Keywords:    strings.Join(pageData.Keywords, ", "),
		Canonical:   s.generateCanonicalURL(pageData),
		Language:    s.getLanguage(pageData),

		// Open Graph tags
		OGTitle:       s.generateOGTitle(pageData),
		OGDescription: s.generateOGDescription(pageData),
		OGImage:       s.generateOGImage(pageData),
		OGType:        s.getOGType(pageData),
		OGURL:         s.generateCanonicalURL(pageData),
		OGSiteName:    s.config.SiteName,
		OGLocale:      s.getOGLocale(pageData),

		// Twitter Card tags
		TwitterCard:        s.getTwitterCardType(pageData),
		TwitterTitle:       s.generateTwitterTitle(pageData),
		TwitterDescription: s.generateTwitterDescription(pageData),
		TwitterImage:       s.generateTwitterImage(pageData),
		TwitterSite:        "@" + strings.ToLower(s.config.SiteName),
		TwitterCreator:     pageData.Author,

		// Additional meta tags
		Robots:           s.generateRobotsTag(pageData),
		Viewport:         "width=device-width, initial-scale=1.0",
		ThemeColor:       "#0066cc",
		MSApplicationTileColor: "#0066cc",

		// Article-specific tags
		ArticleAuthor:      pageData.Author,
		ArticlePublishedTime: pageData.PublishedAt,
		ArticleModifiedTime:  pageData.ModifiedAt,
		ArticleSection:       pageData.Category,
		ArticleTags:          pageData.Keywords,
	}

	// Add JSON-LD structured data
	jsonLD, err := s.GenerateStructuredData(ctx, pageData)
	if err != nil {
		s.logger.Error("Failed to generate structured data", zap.Error(err))
	} else {
		meta.StructuredData = jsonLD
	}

	return meta, nil
}

// GenerateStructuredData creates JSON-LD structured data
func (s *SEOService) GenerateStructuredData(ctx context.Context, pageData *models.SEOPageData) (string, error) {
	var structuredData interface{}

	switch pageData.Type {
	case "homepage":
		structuredData = s.generateOrganizationSchema(pageData)
	case "article":
		structuredData = s.generateArticleSchema(pageData)
	case "product":
		structuredData = s.generateProductSchema(pageData)
	case "faq":
		structuredData = s.generateFAQSchema(pageData)
	case "contact":
		structuredData = s.generateContactSchema(pageData)
	default:
		structuredData = s.generateWebPageSchema(pageData)
	}

	jsonBytes, err := json.MarshalIndent(structuredData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal structured data: %w", err)
	}

	return string(jsonBytes), nil
}

// generateOrganizationSchema creates Organization schema for homepage
func (s *SEOService) generateOrganizationSchema(pageData *models.SEOPageData) map[string]interface{} {
	return map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "Organization",
		"name":     s.config.SiteName,
		"url":      s.config.SiteURL,
		"logo": map[string]interface{}{
			"@type": "ImageObject",
			"url":   s.config.SiteURL + "/static/images/logo.png",
		},
		"description": pageData.Description,
		"foundingDate": "2024",
		"contactPoint": map[string]interface{}{
			"@type":      "ContactPoint",
			"telephone":  "+1-555-123-4567",
			"contactType": "customer service",
			"email":      "support@statuspage.pro",
			"areaServed": "Worldwide",
			"availableLanguage": []string{"English"},
		},
		"sameAs": []string{
			"https://twitter.com/statuspage",
			"https://linkedin.com/company/statuspage",
			"https://github.com/statuspage",
		},
		"address": map[string]interface{}{
			"@type":           "PostalAddress",
			"streetAddress":   "123 Tech Street",
			"addressLocality": "San Francisco",
			"addressRegion":   "CA",
			"postalCode":      "94105",
			"addressCountry":  "US",
		},
	}
}

// generateArticleSchema creates Article schema for blog posts
func (s *SEOService) generateArticleSchema(pageData *models.SEOPageData) map[string]interface{} {
	return map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "Article",
		"headline": pageData.Title,
		"description": pageData.Description,
		"image": []string{pageData.Image},
		"author": map[string]interface{}{
			"@type": "Person",
			"name":  pageData.Author,
		},
		"publisher": map[string]interface{}{
			"@type": "Organization",
			"name":  s.config.SiteName,
			"logo": map[string]interface{}{
				"@type": "ImageObject",
				"url":   s.config.SiteURL + "/static/images/logo.png",
			},
		},
		"datePublished": pageData.PublishedAt,
		"dateModified":  pageData.ModifiedAt,
		"mainEntityOfPage": map[string]interface{}{
			"@type": "WebPage",
			"@id":   s.generateCanonicalURL(pageData),
		},
	}
}

// generateProductSchema creates Product schema for service pages
func (s *SEOService) generateProductSchema(pageData *models.SEOPageData) map[string]interface{} {
	return map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "SoftwareApplication",
		"name":     pageData.Title,
		"description": pageData.Description,
		"applicationCategory": "BusinessApplication",
		"operatingSystem": "Web Browser",
		"offers": map[string]interface{}{
			"@type": "Offer",
			"price": "29.99",
			"priceCurrency": "USD",
			"priceValidUntil": time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
			"availability": "https://schema.org/InStock",
			"seller": map[string]interface{}{
				"@type": "Organization",
				"name":  s.config.SiteName,
			},
		},
		"aggregateRating": map[string]interface{}{
			"@type": "AggregateRating",
			"ratingValue": "4.8",
			"reviewCount": "150",
			"bestRating": "5",
			"worstRating": "1",
		},
	}
}

// generateFAQSchema creates FAQPage schema
func (s *SEOService) generateFAQSchema(pageData *models.SEOPageData) map[string]interface{} {
	mainEntity := make([]map[string]interface{}, 0)

	for _, faq := range pageData.FAQs {
		mainEntity = append(mainEntity, map[string]interface{}{
			"@type": "Question",
			"name":  faq.Question,
			"acceptedAnswer": map[string]interface{}{
				"@type": "Answer",
				"text":  faq.Answer,
			},
		})
	}

	return map[string]interface{}{
		"@context":   "https://schema.org",
		"@type":      "FAQPage",
		"mainEntity": mainEntity,
	}
}

// generateContactSchema creates ContactPage schema
func (s *SEOService) generateContactSchema(pageData *models.SEOPageData) map[string]interface{} {
	return map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "ContactPage",
		"name":     pageData.Title,
		"description": pageData.Description,
		"url":      s.generateCanonicalURL(pageData),
		"mainEntity": map[string]interface{}{
			"@type": "Organization",
			"name":  s.config.SiteName,
			"contactPoint": map[string]interface{}{
				"@type":      "ContactPoint",
				"telephone":  "+1-555-123-4567",
				"contactType": "customer service",
				"email":      "support@statuspage.pro",
			},
		},
	}
}

// generateWebPageSchema creates generic WebPage schema
func (s *SEOService) generateWebPageSchema(pageData *models.SEOPageData) map[string]interface{} {
	return map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "WebPage",
		"name":     pageData.Title,
		"description": pageData.Description,
		"url":      s.generateCanonicalURL(pageData),
		"isPartOf": map[string]interface{}{
			"@type": "WebSite",
			"name":  s.config.SiteName,
			"url":   s.config.SiteURL,
		},
		"about": map[string]interface{}{
			"@type": "Organization",
			"name":  s.config.SiteName,
		},
	}
}

// GenerateSitemap creates XML sitemap
func (s *SEOService) GenerateSitemap(ctx context.Context, pages []models.SitemapPage) (string, error) {
	sitemap := models.Sitemap{
		XMLName: xml.Name{Local: "urlset"},
		Xmlns:   "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:    make([]models.SitemapURL, 0, len(pages)),
	}

	for _, page := range pages {
		sitemapURL := models.SitemapURL{
			Loc:        s.config.SiteURL + page.Path,
			LastMod:    page.LastModified.Format("2006-01-02T15:04:05-07:00"),
			ChangeFreq: page.ChangeFreq,
			Priority:   page.Priority,
		}
		sitemap.URLs = append(sitemap.URLs, sitemapURL)
	}

	xmlBytes, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sitemap: %w", err)
	}

	return xml.Header + string(xmlBytes), nil
}

// GenerateRobotsTxt creates robots.txt content
func (s *SEOService) GenerateRobotsTxt(ctx context.Context) string {
	return fmt.Sprintf(`User-agent: *
Allow: /

# Sitemap
Sitemap: %s/sitemap.xml

# High-crawl rate pages
Crawl-delay: 1

# Block admin areas
Disallow: /admin/
Disallow: /api/
Disallow: /dashboard/
Disallow: /*.json$
Disallow: /*?*

# Allow important pages
Allow: /blog/
Allow: /features/
Allow: /pricing/
Allow: /contact/

# Block search engines from private content
User-agent: *
Disallow: /private/
Disallow: /temp/
`, s.config.SiteURL)
}

// Helper methods

func (s *SEOService) generateTitle(pageData *models.SEOPageData) string {
	if pageData.Title == "" {
		return s.config.SiteName
	}

	if pageData.Type == "homepage" {
		return fmt.Sprintf("%s - %s", pageData.Title, pageData.Description)
	}

	return fmt.Sprintf("%s | %s", pageData.Title, s.config.SiteName)
}

func (s *SEOService) generateDescription(pageData *models.SEOPageData) string {
	if pageData.Description != "" {
		return pageData.Description
	}
	return "StatusPage Pro - Modern status page solution for teams who care about transparency and communication."
}

func (s *SEOService) generateCanonicalURL(pageData *models.SEOPageData) string {
	baseURL, _ := url.Parse(s.config.SiteURL)
	pageURL, _ := url.Parse(pageData.Path)
	return baseURL.ResolveReference(pageURL).String()
}

func (s *SEOService) getLanguage(pageData *models.SEOPageData) string {
	if pageData.Language != "" {
		return pageData.Language
	}
	return s.config.DefaultLang
}

func (s *SEOService) generateOGTitle(pageData *models.SEOPageData) string {
	if pageData.OGTitle != "" {
		return pageData.OGTitle
	}
	return s.generateTitle(pageData)
}

func (s *SEOService) generateOGDescription(pageData *models.SEOPageData) string {
	if pageData.OGDescription != "" {
		return pageData.OGDescription
	}
	return s.generateDescription(pageData)
}

func (s *SEOService) generateOGImage(pageData *models.SEOPageData) string {
	if pageData.Image != "" {
		if strings.HasPrefix(pageData.Image, "http") {
			return pageData.Image
		}
		return s.config.SiteURL + pageData.Image
	}
	return s.config.SiteURL + "/static/images/og-default.png"
}

func (s *SEOService) getOGType(pageData *models.SEOPageData) string {
	switch pageData.Type {
	case "article":
		return "article"
	case "product":
		return "product"
	default:
		return "website"
	}
}

func (s *SEOService) getOGLocale(pageData *models.SEOPageData) string {
	lang := s.getLanguage(pageData)
	switch lang {
	case "en":
		return "en_US"
	case "es":
		return "es_ES"
	case "fr":
		return "fr_FR"
	default:
		return "en_US"
	}
}

func (s *SEOService) getTwitterCardType(pageData *models.SEOPageData) string {
	if pageData.Image != "" {
		return "summary_large_image"
	}
	return "summary"
}

func (s *SEOService) generateTwitterTitle(pageData *models.SEOPageData) string {
	title := s.generateTitle(pageData)
	if len(title) > 70 {
		return title[:67] + "..."
	}
	return title
}

func (s *SEOService) generateTwitterDescription(pageData *models.SEOPageData) string {
	desc := s.generateDescription(pageData)
	if len(desc) > 200 {
		return desc[:197] + "..."
	}
	return desc
}

func (s *SEOService) generateTwitterImage(pageData *models.SEOPageData) string {
	return s.generateOGImage(pageData)
}

func (s *SEOService) generateRobotsTag(pageData *models.SEOPageData) string {
	if pageData.NoIndex {
		return "noindex, nofollow"
	}

	if pageData.Type == "admin" || pageData.Type == "private" {
		return "noindex, nofollow"
	}

	return "index, follow"
}