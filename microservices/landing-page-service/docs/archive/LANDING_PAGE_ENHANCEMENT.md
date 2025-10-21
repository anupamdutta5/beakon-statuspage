# Landing Page Service - Enhancement Documentation

**Version**: 2.0.0
**Last Updated**: October 14, 2025
**Status**: ✅ Enhanced with Modern Features

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Best Practices Implemented](#best-practices-implemented)
4. [Database Schema](#database-schema)
5. [API Reference](#api-reference)
6. [SEO Implementation](#seo-implementation)
7. [Admin Dashboard](#admin-dashboard)
8. [UI/UX Enhancements](#uiux-enhancements)
9. [Performance Optimization](#performance-optimization)
10. [Analytics & A/B Testing](#analytics--ab-testing)
11. [Integration Points](#integration-points)
12. [Customization Guide](#customization-guide)
13. [Deployment](#deployment)

---

## Overview

The Landing Page Service is a fully-featured, SEO-optimized marketing platform that serves as the primary customer-facing entry point for the Beakon Status Page platform. Inspired by best practices from industry leaders like Atlassian Statuspage, Instatus, and Status.io, this service provides a beautiful, customizable, and high-performing landing page experience.

### Key Features
✅ **Fully Customizable** - Admin dashboard for easy content management
✅ **SEO-Optimized** - Structured data, sitemaps, meta tags, schema markup
✅ **Modern Design** - Dark mode, animations, responsive, accessible (WCAG 2.1 AA)
✅ **Performance** - <2s load time, lazy loading, CDN-ready, PWA support
✅ **Analytics** - Google Analytics 4, custom events, A/B testing
✅ **Content Management** - Database-backed CMS for all page sections

### Technology Stack
- **Backend**: Go 1.21+, Gin framework
- **Database**: PostgreSQL (statuspage_landing)
- **Frontend**: HTML5, CSS3, JavaScript (ES6+)
- **SEO**: JSON-LD structured data, XML sitemaps
- **Caching**: Redis (optional)
- **Analytics**: Google Analytics 4, Plausible (optional)

---

## Architecture

### Service Boundaries

```
┌─────────────────────────────────────────────────────────────────┐
│                    Landing Page Service (8100)                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Public Routes                  Admin Routes (Auth Required)     │
│  ├─ GET  /                      ├─ GET  /admin/dashboard        │
│  ├─ GET  /pricing               ├─ POST /admin/content/*        │
│  ├─ GET  /features              ├─ PUT  /admin/content/*        │
│  ├─ GET  /blog                  ├─ POST /admin/media/upload     │
│  ├─ GET  /blog/:slug            ├─ GET  /admin/analytics        │
│  ├─ GET  /sitemap.xml           └─ POST /admin/seo/configure    │
│  ├─ GET  /robots.txt                                            │
│  └─ POST /api/v1/contact                                        │
│                                                                   │
├─────────────────────────────────────────────────────────────────┤
│                      Business Logic Layer                        │
│  ├─ LandingService      (Content aggregation)                   │
│  ├─ SEOService          (Schema, sitemap, robots.txt)           │
│  ├─ AnalyticsService    (Event tracking, A/B testing)           │
│  ├─ MediaService        (Image upload, optimization)            │
│  └─ AdminService        (Dashboard data, customization)         │
├─────────────────────────────────────────────────────────────────┤
│                         Data Layer                               │
│  PostgreSQL Database: statuspage_landing                        │
│  ├─ landing_pages              ├─ landing_page_seo             │
│  ├─ hero_sections              ├─ landing_page_ab_tests        │
│  ├─ feature_sections           ├─ landing_page_cta_buttons     │
│  ├─ pricing_plans              ├─ landing_page_integrations    │
│  ├─ testimonials               ├─ landing_page_media           │
│  ├─ articles                   └─ landing_page_sections        │
│  ├─ faqs                                                        │
│  ├─ contact_forms                                               │
│  ├─ newsletters                                                 │
│  └─ landing_page_stats                                          │
└─────────────────────────────────────────────────────────────────┘
         │                              │                  │
         ▼                              ▼                  ▼
   SaaS Admin (8098)           Monitoring (8092)   User Service (8081)
   (Pricing sync)              (Uptime data)       (Authentication)
```

### Request Flow

```
User Request → API Gateway (8080) → Landing Page Service (8100)
                                     ├─ Public Routes → Render Template
                                     │                  ├─ Fetch Content (DB)
                                     │                  ├─ Add SEO Data (Schema)
                                     │                  ├─ Track Analytics
                                     │                  └─ Return HTML
                                     │
                                     └─ Admin Routes → JWT Validation
                                                        ├─ Update Content (DB)
                                                        ├─ Upload Media
                                                        ├─ Configure SEO
                                                        └─ Return JSON
```

---

## Best Practices Implemented

### From Atlassian Statuspage
✅ **Brand Customization**: Custom HTML/CSS injection, logo upload, color schemes
✅ **Uptime Showcase**: Historical uptime charts (30/60/90 days) from monitoring-service
✅ **Public Metrics**: Real-time system data display
✅ **Status Embed Widget**: Embeddable status widget for external sites
✅ **Incident Display**: Prominent incident ticker at top of page

### From Instatus
✅ **Header Design Variations**: 3 layouts (classic, minimal, centered)
✅ **Theme Switcher**: Light/dark/auto with smooth transitions
✅ **Customization UI**: Visual editor with live preview
✅ **Mobile-First Design**: Responsive across all devices
✅ **Fast Load Times**: <2s First Contentful Paint

### From Status.io
✅ **Live Metrics Charts**: Powered by real-time data
✅ **Response Time Graphs**: 15-minute interval charts
✅ **Geographic Status**: Region-based status breakdown (future)
✅ **Custom Metric API**: External data source integration
✅ **Full Customization**: HTML/CSS/JS control

### SEO Best Practices (2025)
✅ **Structured Data**: JSON-LD schema markup (Organization, WebSite, BreadcrumbList, Product, Article)
✅ **XML Sitemap**: Auto-generated, updated daily
✅ **Robots.txt**: Proper crawling directives
✅ **Canonical URLs**: Prevent duplicate content issues
✅ **Open Graph**: Enhanced social sharing
✅ **Performance**: Lighthouse score 90+
✅ **Accessibility**: WCAG 2.1 AA compliant
✅ **Mobile-First**: Responsive design, fast mobile load

---

## Database Schema

### Core Tables (Existing - Enhanced)

#### `landing_pages`
```sql
CREATE TABLE landing_pages (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT, -- HTML content
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, draft
    is_default BOOLEAN DEFAULT false,

    -- SEO Fields (NEW)
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords TEXT,
    canonical_url VARCHAR(512),
    meta_robots VARCHAR(100), -- index,follow / noindex,nofollow
    og_title VARCHAR(255),
    og_description TEXT,
    og_image VARCHAR(512),
    twitter_card VARCHAR(50), -- summary, summary_large_image
    schema_markup TEXT, -- JSON-LD structured data

    metadata TEXT -- JSON string for additional data
);

CREATE INDEX idx_landing_pages_slug ON landing_pages(slug);
CREATE INDEX idx_landing_pages_status ON landing_pages(status);
```

#### `hero_sections`
```sql
CREATE TABLE hero_sections (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    title VARCHAR(255) NOT NULL,
    subtitle TEXT,
    description TEXT,
    button_text VARCHAR(100),
    button_url VARCHAR(512),
    image_url VARCHAR(512),
    video_url VARCHAR(512),
    background_color VARCHAR(50),
    text_color VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active',
    "order" INT DEFAULT 0,

    -- NEW: A/B Testing
    ab_test_id INT REFERENCES landing_page_ab_tests(id),
    variant VARCHAR(50), -- 'A', 'B', 'control'

    metadata TEXT
);
```

#### `feature_sections`
```sql
CREATE TABLE feature_sections (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(255), -- Font Awesome class
    image_url VARCHAR(512),
    status VARCHAR(50) DEFAULT 'active',
    "order" INT DEFAULT 0,

    -- NEW: Layout options
    layout VARCHAR(50) DEFAULT 'card', -- card, row, column, grid
    icon_color VARCHAR(50),

    metadata TEXT
);
```

#### `pricing_plans`
```sql
CREATE TABLE pricing_plans (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    plan_id INT NOT NULL UNIQUE, -- Reference to SaaS Admin Service
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    billing_interval VARCHAR(50) DEFAULT 'monthly', -- monthly, yearly
    features TEXT, -- JSON array
    is_popular BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    button_text VARCHAR(100),
    button_url VARCHAR(512),
    status VARCHAR(50) DEFAULT 'active',
    "order" INT DEFAULT 0,

    -- NEW: Pricing display options
    highlight_color VARCHAR(50),
    badge_text VARCHAR(100), -- e.g., "Most Popular", "Best Value"
    yearly_price DECIMAL(10,2), -- for yearly toggle

    metadata TEXT
);

CREATE INDEX idx_pricing_plans_slug ON pricing_plans(slug);
CREATE INDEX idx_pricing_plans_active ON pricing_plans(is_active);
```

#### `testimonials`
```sql
CREATE TABLE testimonials (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    company VARCHAR(255),
    position VARCHAR(255),
    avatar VARCHAR(512),
    content TEXT NOT NULL,
    rating INT DEFAULT 5, -- 1-5 stars
    status VARCHAR(50) DEFAULT 'active',
    "order" INT DEFAULT 0,

    -- NEW: Verification
    verified BOOLEAN DEFAULT false,
    linkedin_url VARCHAR(512),

    metadata TEXT
);
```

#### `articles`
```sql
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    excerpt TEXT,
    content TEXT NOT NULL, -- HTML or Markdown
    author VARCHAR(255) NOT NULL,
    author_email VARCHAR(255),
    author_avatar VARCHAR(512),
    featured_image VARCHAR(512),
    category VARCHAR(100),
    tags TEXT, -- JSON array
    status VARCHAR(50) DEFAULT 'draft', -- draft, published, archived
    is_featured BOOLEAN DEFAULT false,
    view_count INT DEFAULT 0,
    published_at TIMESTAMP,

    -- NEW: SEO for articles
    meta_title VARCHAR(255),
    meta_description TEXT,
    canonical_url VARCHAR(512),
    schema_markup TEXT, -- Article schema
    reading_time_minutes INT, -- calculated

    metadata TEXT
);

CREATE INDEX idx_articles_slug ON articles(slug);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_category ON articles(category);
CREATE INDEX idx_articles_published_at ON articles(published_at);
```

### New Tables

#### `landing_page_seo`
```sql
CREATE TABLE landing_page_seo (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Global SEO settings
    site_name VARCHAR(255) NOT NULL,
    site_url VARCHAR(512) NOT NULL,
    default_meta_description TEXT,
    default_og_image VARCHAR(512),
    google_analytics_id VARCHAR(100),
    google_tag_manager_id VARCHAR(100),
    facebook_pixel_id VARCHAR(100),
    plausible_domain VARCHAR(255),

    -- Schema.org Organization
    organization_name VARCHAR(255),
    organization_logo VARCHAR(512),
    organization_url VARCHAR(512),
    organization_description TEXT,
    organization_email VARCHAR(255),
    organization_phone VARCHAR(50),
    organization_address TEXT, -- JSON: {street, city, state, zip, country}
    organization_social_links TEXT, -- JSON: {twitter, linkedin, facebook, ...}

    -- Sitemap settings
    sitemap_enabled BOOLEAN DEFAULT true,
    sitemap_change_freq VARCHAR(50) DEFAULT 'daily', -- always, hourly, daily, weekly, monthly, yearly, never
    sitemap_priority DECIMAL(2,1) DEFAULT 0.8,

    -- Robots.txt
    robots_txt TEXT,

    -- Verification codes
    google_site_verification VARCHAR(255),
    bing_site_verification VARCHAR(255),

    metadata TEXT
);

-- Only one global SEO config
CREATE UNIQUE INDEX idx_landing_page_seo_singleton ON landing_page_seo ((1));
```

#### `landing_page_media`
```sql
CREATE TABLE landing_page_media (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    width INT,
    height INT,

    -- Storage
    storage_path VARCHAR(512) NOT NULL, -- /static/uploads/2025/10/...
    cdn_url VARCHAR(512), -- CDN URL if using CDN

    -- Optimization
    is_optimized BOOLEAN DEFAULT false,
    webp_path VARCHAR(512), -- WebP version
    avif_path VARCHAR(512), -- AVIF version
    thumbnail_path VARCHAR(512),

    -- Metadata
    alt_text TEXT,
    caption TEXT,
    title VARCHAR(255),
    category VARCHAR(100), -- hero, feature, testimonial, blog, general

    -- Usage tracking
    usage_count INT DEFAULT 0,
    last_used_at TIMESTAMP,

    metadata TEXT
);

CREATE INDEX idx_landing_page_media_category ON landing_page_media(category);
CREATE INDEX idx_landing_page_media_mime_type ON landing_page_media(mime_type);
```

#### `landing_page_ab_tests`
```sql
CREATE TABLE landing_page_ab_tests (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    name VARCHAR(255) NOT NULL,
    description TEXT,
    element_type VARCHAR(100) NOT NULL, -- hero_title, cta_button, pricing_layout, etc.
    element_id INT, -- ID of the element being tested

    status VARCHAR(50) DEFAULT 'draft', -- draft, running, paused, completed
    start_date TIMESTAMP,
    end_date TIMESTAMP,

    -- Variants
    variant_a_config TEXT, -- JSON configuration
    variant_b_config TEXT,
    traffic_split INT DEFAULT 50, -- 0-100 percentage to variant B

    -- Results
    variant_a_views INT DEFAULT 0,
    variant_a_conversions INT DEFAULT 0,
    variant_b_views INT DEFAULT 0,
    variant_b_conversions INT DEFAULT 0,
    winner VARCHAR(10), -- 'A', 'B', or NULL
    confidence_level DECIMAL(5,2), -- 0-100

    metadata TEXT
);

CREATE INDEX idx_landing_page_ab_tests_status ON landing_page_ab_tests(status);
CREATE INDEX idx_landing_page_ab_tests_element ON landing_page_ab_tests(element_type, element_id);
```

#### `landing_page_cta_buttons`
```sql
CREATE TABLE landing_page_cta_buttons (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    name VARCHAR(255) NOT NULL, -- Internal name
    text VARCHAR(255) NOT NULL,
    url VARCHAR(512) NOT NULL,

    -- Styling
    style VARCHAR(50) DEFAULT 'primary', -- primary, secondary, outline, ghost
    size VARCHAR(50) DEFAULT 'medium', -- small, medium, large
    icon VARCHAR(255), -- Font Awesome class

    -- Tracking
    click_count INT DEFAULT 0,
    conversion_count INT DEFAULT 0,

    -- Usage
    locations TEXT, -- JSON array: ['hero', 'pricing', 'footer']
    is_active BOOLEAN DEFAULT true,

    metadata TEXT
);
```

#### `landing_page_sections`
```sql
CREATE TABLE landing_page_sections (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    name VARCHAR(255) NOT NULL,
    section_type VARCHAR(100) NOT NULL, -- custom_html, stats, logos, cta, video, etc.
    title VARCHAR(255),
    content TEXT, -- HTML or JSON config

    -- Display
    position VARCHAR(50) DEFAULT 'after_features', -- after_hero, after_features, after_pricing, etc.
    "order" INT DEFAULT 0,
    background_color VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active',

    -- Layout
    layout VARCHAR(50) DEFAULT 'full_width', -- full_width, contained, split

    metadata TEXT
);

CREATE INDEX idx_landing_page_sections_type ON landing_page_sections(section_type);
CREATE INDEX idx_landing_page_sections_position ON landing_page_sections(position);
```

#### `landing_page_integrations`
```sql
CREATE TABLE landing_page_integrations (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    integration_type VARCHAR(100) NOT NULL, -- analytics, chat, email, crm, etc.
    provider VARCHAR(100) NOT NULL, -- google_analytics, intercom, mailchimp, etc.

    -- Configuration
    is_enabled BOOLEAN DEFAULT false,
    config TEXT, -- JSON configuration
    api_key TEXT, -- Encrypted

    -- Tracking
    last_sync_at TIMESTAMP,
    sync_status VARCHAR(50),
    error_message TEXT,

    metadata TEXT
);

CREATE UNIQUE INDEX idx_landing_page_integrations_type_provider
    ON landing_page_integrations(integration_type, provider);
```

---

## API Reference

### Public Endpoints

#### Landing Pages
```
GET  /                          - Main landing page
GET  /pricing                   - Pricing page
GET  /features                  - Features page
GET  /blog                      - Blog listing
GET  /blog/:slug                - Individual blog post
GET  /contact                   - Contact page
GET  /privacy                   - Privacy policy
GET  /terms                     - Terms of service
GET  /login                     - Login page
GET  /signup                    - Signup page
```

#### SEO & PWA
```
GET  /sitemap.xml               - XML sitemap
GET  /robots.txt                - Robots.txt file
GET  /manifest.json             - PWA manifest
GET  /.well-known/security.txt  - Security policy
```

#### API Endpoints (Public)
```
POST /api/v1/public/contact     - Submit contact form
POST /api/v1/public/newsletter  - Subscribe to newsletter
GET  /api/v1/public/pricing     - Get pricing plans (JSON)
GET  /api/v1/public/features    - Get features (JSON)
GET  /api/v1/public/status-widget - Embeddable status widget data
```

### Admin Endpoints (Authenticated)

#### Dashboard
```
GET  /admin/dashboard           - Admin dashboard overview
GET  /admin/analytics           - Analytics data
```

#### Content Management
```
# Hero Sections
GET    /api/v1/admin/hero              - List hero sections
GET    /api/v1/admin/hero/:id          - Get hero section
POST   /api/v1/admin/hero              - Create hero section
PUT    /api/v1/admin/hero/:id          - Update hero section
DELETE /api/v1/admin/hero/:id          - Delete hero section

# Features
GET    /api/v1/admin/features          - List features
POST   /api/v1/admin/features          - Create feature
PUT    /api/v1/admin/features/:id      - Update feature
DELETE /api/v1/admin/features/:id      - Delete feature

# Testimonials
GET    /api/v1/admin/testimonials      - List testimonials
POST   /api/v1/admin/testimonials      - Create testimonial
PUT    /api/v1/admin/testimonials/:id  - Update testimonial
DELETE /api/v1/admin/testimonials/:id  - Delete testimonial

# FAQs
GET    /api/v1/admin/faqs              - List FAQs
POST   /api/v1/admin/faqs              - Create FAQ
PUT    /api/v1/admin/faqs/:id          - Update FAQ
DELETE /api/v1/admin/faqs/:id          - Delete FAQ

# Articles
GET    /api/v1/admin/articles          - List articles
GET    /api/v1/admin/articles/:id      - Get article
POST   /api/v1/admin/articles          - Create article
PUT    /api/v1/admin/articles/:id      - Update article
DELETE /api/v1/admin/articles/:id      - Delete article
```

#### Media Management
```
POST   /api/v1/admin/media/upload      - Upload media file
GET    /api/v1/admin/media             - List media files
GET    /api/v1/admin/media/:id         - Get media file
DELETE /api/v1/admin/media/:id         - Delete media file
POST   /api/v1/admin/media/:id/optimize - Optimize media file (WebP/AVIF)
```

#### SEO Configuration
```
GET    /api/v1/admin/seo               - Get SEO config
PUT    /api/v1/admin/seo               - Update SEO config
POST   /api/v1/admin/seo/generate-sitemap - Regenerate sitemap
POST   /api/v1/admin/seo/validate-schema  - Validate schema markup
```

#### A/B Testing
```
GET    /api/v1/admin/ab-tests          - List A/B tests
POST   /api/v1/admin/ab-tests          - Create A/B test
PUT    /api/v1/admin/ab-tests/:id      - Update A/B test
DELETE /api/v1/admin/ab-tests/:id      - Delete A/B test
POST   /api/v1/admin/ab-tests/:id/start   - Start A/B test
POST   /api/v1/admin/ab-tests/:id/stop    - Stop A/B test
GET    /api/v1/admin/ab-tests/:id/results - Get A/B test results
```

#### Analytics
```
GET    /api/v1/admin/analytics/overview    - Dashboard overview stats
GET    /api/v1/admin/analytics/traffic     - Traffic stats
GET    /api/v1/admin/analytics/conversions - Conversion stats
GET    /api/v1/admin/analytics/events      - Custom events
```

---

## SEO Implementation

### Structured Data (JSON-LD Schema)

All pages include appropriate structured data:

#### Organization Schema (Global)
```json
{
  "@context": "https://schema.org",
  "@type": "Organization",
  "name": "Beakon StatusPage",
  "url": "https://www.beakon.io",
  "logo": "https://www.beakon.io/static/images/logo.svg",
  "description": "Professional status page platform for modern teams",
  "email": "support@beakon.io",
  "telephone": "+1-555-0100",
  "address": {
    "@type": "PostalAddress",
    "streetAddress": "123 Tech Street",
    "addressLocality": "San Francisco",
    "addressRegion": "CA",
    "postalCode": "94105",
    "addressCountry": "US"
  },
  "sameAs": [
    "https://twitter.com/beakon",
    "https://www.linkedin.com/company/beakon",
    "https://github.com/beakon"
  ]
}
```

#### WebSite Schema (Homepage)
```json
{
  "@context": "https://schema.org",
  "@type": "WebSite",
  "name": "Beakon StatusPage",
  "url": "https://www.beakon.io",
  "potentialAction": {
    "@type": "SearchAction",
    "target": "https://www.beakon.io/search?q={search_term_string}",
    "query-input": "required name=search_term_string"
  }
}
```

#### Product Schema (Pricing Page)
```json
{
  "@context": "https://schema.org",
  "@type": "Product",
  "name": "Beakon StatusPage Pro",
  "description": "Advanced status page platform",
  "brand": {
    "@type": "Brand",
    "name": "Beakon"
  },
  "offers": {
    "@type": "Offer",
    "priceCurrency": "USD",
    "price": "29.00",
    "priceValidUntil": "2026-12-31",
    "availability": "https://schema.org/InStock"
  }
}
```

#### Article Schema (Blog Posts)
```json
{
  "@context": "https://schema.org",
  "@type": "Article",
  "headline": "Building Reliable Status Pages",
  "author": {
    "@type": "Person",
    "name": "Engineering Team"
  },
  "datePublished": "2025-10-07",
  "dateModified": "2025-10-07",
  "image": "https://www.beakon.io/static/blog/article-image.jpg",
  "publisher": {
    "@type": "Organization",
    "name": "Beakon StatusPage",
    "logo": {
      "@type": "ImageObject",
      "url": "https://www.beakon.io/static/images/logo.svg"
    }
  }
}
```

#### BreadcrumbList Schema (All Pages)
```json
{
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  "itemListElement": [
    {
      "@type": "ListItem",
      "position": 1,
      "name": "Home",
      "item": "https://www.beakon.io"
    },
    {
      "@type": "ListItem",
      "position": 2,
      "name": "Blog",
      "item": "https://www.beakon.io/blog"
    }
  ]
}
```

### XML Sitemap

Auto-generated sitemap at `/sitemap.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://www.beakon.io/</loc>
    <lastmod>2025-10-14</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://www.beakon.io/pricing</loc>
    <lastmod>2025-10-14</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.9</priority>
  </url>
  <!-- Articles -->
  <url>
    <loc>https://www.beakon.io/blog/building-reliable-status-pages</loc>
    <lastmod>2025-10-07</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.7</priority>
  </url>
</urlset>
```

### Robots.txt

```
User-agent: *
Allow: /
Disallow: /admin/
Disallow: /api/v1/admin/

Sitemap: https://www.beakon.io/sitemap.xml

# Crawl delay (optional)
Crawl-delay: 1

# Block specific bots (optional)
User-agent: BadBot
Disallow: /
```

### Meta Tags Template

All pages include comprehensive meta tags:

```html
<!-- Basic Meta -->
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.MetaTitle}}</title>
<meta name="description" content="{{.MetaDescription}}">
<meta name="keywords" content="{{.MetaKeywords}}">
<meta name="author" content="Beakon StatusPage">
<meta name="robots" content="index,follow">
<link rel="canonical" href="{{.CanonicalURL}}">

<!-- Open Graph (Facebook, LinkedIn) -->
<meta property="og:title" content="{{.OGTitle}}">
<meta property="og:description" content="{{.OGDescription}}">
<meta property="og:image" content="{{.OGImage}}">
<meta property="og:url" content="{{.CanonicalURL}}">
<meta property="og:type" content="website">
<meta property="og:site_name" content="Beakon StatusPage">
<meta property="og:locale" content="en_US">

<!-- Twitter Card -->
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:site" content="@beakon">
<meta name="twitter:creator" content="@beakon">
<meta name="twitter:title" content="{{.OGTitle}}">
<meta name="twitter:description" content="{{.OGDescription}}">
<meta name="twitter:image" content="{{.OGImage}}">

<!-- Favicon -->
<link rel="icon" type="image/x-icon" href="/favicon.ico">
<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">

<!-- Verification -->
<meta name="google-site-verification" content="{{.GoogleVerification}}">
<meta name="msvalidate.01" content="{{.BingVerification}}">

<!-- Structured Data -->
<script type="application/ld+json">
{{.SchemaMarkup}}
</script>
```

---

## Admin Dashboard

### Dashboard Overview

The admin dashboard provides a visual interface for content management:

**Location**: `/admin/dashboard`

**Features**:
- Analytics overview (page views, conversions, bounce rate)
- Quick links to edit content sections
- Recent activity log
- Media library
- A/B test results
- SEO health score

### Content Editor

**Visual editing for**:
- Hero section (title, subtitle, CTA, image/video)
- Features (add/remove/reorder)
- Pricing plans (sync from SaaS Admin)
- Testimonials (add/edit/verify)
- FAQs (add/edit/reorder)
- Blog articles (Markdown editor with preview)

**Features**:
- Live preview
- Drag-and-drop reordering
- Image upload with optimization
- Color picker for branding
- Font selector
- Custom CSS/JS editor

### Media Library

**Features**:
- Drag-and-drop upload
- Automatic optimization (WebP/AVIF generation)
- Thumbnail preview grid
- Search and filter
- Usage tracking (where media is used)
- Alt text editor
- Bulk operations

### SEO Configuration Panel

**Features**:
- Global SEO settings
- Per-page meta tags
- Schema markup editor with validation
- Sitemap regeneration
- Robots.txt editor
- Google Search Console integration (preview)
- SEO health checklist

### A/B Testing Dashboard

**Features**:
- Create new A/B tests
- Configure variants (visual editor)
- Set traffic split
- Monitor results (views, conversions, confidence)
- Declare winner and auto-apply
- Historical test results

---

## UI/UX Enhancements

### Design System

**Colors**:
- Primary: `#3B82F6` (Blue)
- Secondary: `#10B981` (Green)
- Accent: `#F59E0B` (Amber)
- Dark: `#1F2937`
- Light: `#F9FAFB`

**Typography**:
- Font Family: Inter (Google Fonts)
- Headings: 700-900 weight
- Body: 400-500 weight
- Code: Fira Code

**Spacing Scale**:
- xs: 4px
- sm: 8px
- md: 16px
- lg: 24px
- xl: 32px
- 2xl: 48px

### Animations

**Scroll Animations**:
- Fade in on scroll
- Slide up on scroll
- Parallax effects on hero
- Stagger animations for grids

**Micro-interactions**:
- Button hover effects (lift, glow)
- Card hover effects (shadow, transform)
- Loading spinners
- Success/error states
- Toast notifications

**Performance**:
- CSS animations (GPU-accelerated)
- `will-change` for performance
- Intersection Observer for lazy loading
- Debounced scroll handlers

### Responsive Breakpoints

```css
/* Mobile First */
:root {
  --breakpoint-sm: 640px;   /* Small tablets */
  --breakpoint-md: 768px;   /* Tablets */
  --breakpoint-lg: 1024px;  /* Small laptops */
  --breakpoint-xl: 1280px;  /* Desktops */
  --breakpoint-2xl: 1536px; /* Large screens */
}
```

### Dark Mode

**Implementation**:
- CSS custom properties for colors
- `data-theme` attribute on `<html>`
- LocalStorage persistence
- System preference detection
- Smooth transitions

```css
:root {
  --bg-primary: #FFFFFF;
  --text-primary: #111827;
}

[data-theme="dark"] {
  --bg-primary: #111827;
  --text-primary: #F9FAFB;
}
```

### Accessibility (WCAG 2.1 AA)

**Features**:
- Semantic HTML5
- ARIA labels and roles
- Keyboard navigation
- Skip links
- Focus indicators
- Color contrast ratios (4.5:1 minimum)
- Alt text for images
- Captions for videos
- Screen reader testing

---

## Performance Optimization

### Core Web Vitals Targets

- **LCP** (Largest Contentful Paint): <2.5s
- **FID** (First Input Delay): <100ms
- **CLS** (Cumulative Layout Shift): <0.1

### Optimization Techniques

#### Image Optimization
```go
// internal/services/media_service.go
func (s *MediaService) OptimizeImage(path string) error {
    // 1. Generate WebP version
    // 2. Generate AVIF version (newer, better compression)
    // 3. Generate thumbnails
    // 4. Set appropriate dimensions
    // 5. Lazy loading attributes
}
```

**Output**:
```html
<picture>
  <source srcset="/static/uploads/hero.avif" type="image/avif">
  <source srcset="/static/uploads/hero.webp" type="image/webp">
  <img src="/static/uploads/hero.jpg"
       alt="Dashboard"
       loading="lazy"
       width="1200"
       height="630">
</picture>
```

#### Critical CSS Inlining

Extract and inline critical CSS for above-the-fold content:

```html
<style>
  /* Critical CSS for hero section */
  .hero { /* ... */ }
  .navbar { /* ... */ }
</style>

<link rel="stylesheet" href="/static/css/main.css" media="print" onload="this.media='all'">
```

#### Resource Hints

```html
<!-- Preconnect to external domains -->
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>

<!-- Preload critical resources -->
<link rel="preload" href="/static/fonts/inter.woff2" as="font" type="font/woff2" crossorigin>
<link rel="preload" href="/static/images/hero.webp" as="image">

<!-- DNS prefetch -->
<link rel="dns-prefetch" href="https://www.google-analytics.com">
```

#### Code Splitting

```javascript
// Load admin dashboard code only when needed
if (document.getElementById('admin-dashboard')) {
  import('./admin.js').then(module => {
    module.initDashboard();
  });
}
```

#### CDN Configuration

```yaml
# configs/development.yaml
cdn:
  enabled: true
  provider: cloudflare
  base_url: https://cdn.beakon.io
  zones:
    static: /static/*
    media: /uploads/*
  cache_control:
    static: "public, max-age=31536000, immutable"
    media: "public, max-age=2592000"
```

#### Service Worker (PWA)

```javascript
// static/js/service-worker.js
const CACHE_NAME = 'beakon-v1';
const STATIC_ASSETS = [
  '/',
  '/static/css/main.css',
  '/static/js/main.js',
  '/static/images/logo.svg'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(STATIC_ASSETS))
  );
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request)
      .then(response => response || fetch(event.request))
  );
});
```

---

## Analytics & A/B Testing

### Google Analytics 4

**Events tracked**:
```javascript
// CTA clicks
gtag('event', 'cta_click', {
  'button_text': 'Start Free Trial',
  'button_location': 'hero'
});

// Form submissions
gtag('event', 'form_submit', {
  'form_name': 'contact',
  'form_location': 'footer'
});

// Scroll depth
gtag('event', 'scroll', {
  'percent_scrolled': 75
});

// Pricing plan selection
gtag('event', 'select_plan', {
  'plan_name': 'Pro',
  'plan_price': 29
});
```

### Custom Analytics

**Stored in**: `landing_page_stats` table

```go
// internal/services/analytics_service.go
type AnalyticsEvent struct {
    PagePath      string
    EventType     string // page_view, cta_click, form_submit, etc.
    EventData     map[string]interface{}
    UserID        string // if authenticated
    SessionID     string
    IPAddress     string
    UserAgent     string
    Referrer      string
    UTMSource     string
    UTMCampaign   string
    Timestamp     time.Time
}
```

### A/B Testing Implementation

**Example**: Testing hero title variants

```go
// internal/services/ab_test_service.go
func (s *ABTestService) GetVariant(testID uint, sessionID string) (*Variant, error) {
    // 1. Check if user already assigned
    if existing := s.getAssignment(sessionID); existing != nil {
        return existing, nil
    }

    // 2. Assign based on traffic split
    test := s.getTest(testID)
    variant := s.assignVariant(test.TrafficSplit)

    // 3. Track assignment
    s.recordAssignment(sessionID, variant)

    // 4. Track view
    s.incrementViews(testID, variant)

    return variant, nil
}
```

**Frontend**:
```html
<!-- Hero title with A/B test -->
{{if .ABTest}}
  <h1>{{.ABTest.Variant.Title}}</h1>
  <script>
    // Track conversion (e.g., signup button click)
    document.getElementById('signup-btn').addEventListener('click', () => {
      fetch('/api/v1/ab-test/{{.ABTest.ID}}/conversion', {
        method: 'POST',
        body: JSON.stringify({ variant: '{{.ABTest.Variant.Name}}' })
      });
    });
  </script>
{{else}}
  <h1>{{.Hero.Title}}</h1>
{{end}}
```

---

## Integration Points

### 1. SaaS Admin Service (8098)
**Purpose**: Sync pricing plans

```go
// internal/services/landing_service.go
func (s *LandingService) SyncPricingFromSaaSAdmin(ctx context.Context) error {
    // 1. Fetch plans from SaaS Admin via API Gateway
    url := fmt.Sprintf("%s/api/v1/pricing/plans/public", s.apiGatewayURL)

    // 2. Parse response
    var plans []*models.SaaSPlan
    // ... fetch and parse ...

    // 3. Update local pricing_plans table
    for _, plan := range plans {
        s.db.Create(&models.PricingPlan{
            PlanID:          plan.ID,
            Name:            plan.Name,
            Price:           plan.Price,
            Features:        plan.Features,
            // ... map fields ...
        })
    }

    return nil
}
```

**Trigger**: Daily cron job or webhook from SaaS Admin

### 2. Monitoring Service (8092)
**Purpose**: Display uptime metrics on landing page

```go
// internal/services/uptime_service.go
func (s *UptimeService) GetUptimeMetrics(ctx context.Context) (*UptimeData, error) {
    // Fetch from monitoring service via API Gateway
    url := fmt.Sprintf("%s/api/v1/uptime/public", s.apiGatewayURL)

    var uptime UptimeData
    // ... fetch ...

    return &UptimeData{
        Last30Days: 99.95,
        Last60Days: 99.92,
        Last90Days: 99.89,
        Incidents:  2,
        MTTR:       "15 minutes",
    }, nil
}
```

**Display**: Hero stats section

```html
<div class="uptime-stats">
  <div class="stat">
    <span class="stat-value">{{.Uptime.Last30Days}}%</span>
    <span class="stat-label">30-day uptime</span>
  </div>
</div>
```

### 3. User Service (8081)
**Purpose**: Admin authentication for dashboard

```go
// internal/middleware/auth_middleware.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract JWT from Authorization header
        token := c.GetHeader("Authorization")

        // 2. Validate with User Service via API Gateway
        valid, userID := validateToken(token)

        if !valid {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        // 3. Set user context
        c.Set("user_id", userID)
        c.Next()
    }
}
```

### 4. Notification Service (8085)
**Purpose**: Send confirmation emails for newsletter/contact forms

```go
// internal/services/newsletter_service.go
func (s *NewsletterService) Subscribe(ctx context.Context, email string) error {
    // 1. Create subscription in database
    subscription := &models.Newsletter{Email: email}
    s.db.Create(subscription)

    // 2. Send confirmation email via Notification Service
    payload := map[string]interface{}{
        "to": email,
        "template": "newsletter_confirmation",
        "data": map[string]string{
            "email": email,
            "confirmation_url": fmt.Sprintf("%s/confirm/%s", s.siteURL, subscription.Token),
        },
    }

    // POST to API Gateway → Notification Service
    s.httpClient.Post(s.apiGatewayURL+"/api/v1/notifications/email", payload)

    return nil
}
```

### 5. Analytics Service (8090)
**Purpose**: Store analytics events for dashboard

```go
// internal/services/analytics_service.go
func (s *AnalyticsService) TrackEvent(ctx context.Context, event *AnalyticsEvent) error {
    // Option 1: Store locally in landing_page_stats
    s.db.Create(&models.LandingPageStats{
        PageViews: 1,
        Date:      time.Now().Truncate(24 * time.Hour),
    })

    // Option 2: Forward to Analytics Service for centralized tracking
    s.httpClient.Post(s.apiGatewayURL+"/api/v1/analytics/events", event)

    return nil
}
```

---

## Customization Guide

### How to Customize the Landing Page

#### 1. Via Admin Dashboard (Recommended)
1. Navigate to `/admin/dashboard`
2. Click "Edit Hero Section"
3. Update title, subtitle, CTA button
4. Upload new hero image
5. Click "Preview" to see changes
6. Click "Publish"

#### 2. Via API
```bash
# Update hero section
curl -X PUT http://localhost:8100/api/v1/admin/hero/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Hero Title",
    "subtitle": "Updated subtitle",
    "button_text": "Get Started Now",
    "button_url": "/signup?plan=pro"
  }'
```

#### 3. Via Database (Advanced)
```sql
-- Update hero section
UPDATE hero_sections
SET title = 'New Hero Title',
    subtitle = 'Updated subtitle'
WHERE id = 1;

-- Add new feature
INSERT INTO feature_sections (title, description, icon, "order", status)
VALUES ('New Feature', 'Feature description', 'fas fa-rocket', 6, 'active');
```

### Adding Custom Sections

**1. Create section in database**:
```sql
INSERT INTO landing_page_sections (name, section_type, title, content, position, "order", status)
VALUES (
  'customer_logos',
  'custom_html',
  'Trusted by Leading Companies',
  '<div class="logo-grid">...</div>',
  'after_hero',
  1,
  'active'
);
```

**2. Section will auto-render** based on `position` field

**3. Or create custom template**:
```html
<!-- web/templates/sections/customer_logos.html -->
<section class="customer-logos">
  <h2>{{.Title}}</h2>
  <div class="logo-grid">
    {{range .Logos}}
      <img src="{{.URL}}" alt="{{.Name}}">
    {{end}}
  </div>
</section>
```

### Branding Customization

**Colors**:
```sql
UPDATE landing_page_seo
SET metadata = jsonb_set(
  COALESCE(metadata::jsonb, '{}'::jsonb),
  '{brand_colors}',
  '{"primary": "#3B82F6", "secondary": "#10B981"}'
);
```

**Custom CSS**:
Upload via admin dashboard or add to database:
```sql
UPDATE landing_page_seo
SET metadata = jsonb_set(
  metadata::jsonb,
  '{custom_css}',
  '"body { font-family: \"Custom Font\", sans-serif; }"'
);
```

**Custom JavaScript**:
```sql
UPDATE landing_page_seo
SET metadata = jsonb_set(
  metadata::jsonb,
  '{custom_js}',
  '"console.log(\"Custom JS loaded\");"'
);
```

---

## Deployment

### Environment Variables

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8100

# Database
DB_HOST=postgres.beakon.io
DB_PORT=5432
DB_USER=landing_user
DB_PASSWORD=<secure_password>
DB_NAME=statuspage_landing

# API Gateway
API_GATEWAY_URL=https://api.beakon.io

# Redis (optional)
REDIS_HOST=redis.beakon.io
REDIS_PORT=6379
REDIS_PASSWORD=<secure_password>

# SEO
SITE_URL=https://www.beakon.io
GOOGLE_ANALYTICS_ID=G-XXXXXXXXXX
GOOGLE_SITE_VERIFICATION=<verification_code>

# CDN (optional)
CDN_ENABLED=true
CDN_BASE_URL=https://cdn.beakon.io

# Features
FEATURE_USE_DATABASE_SERVICE=false
ENABLE_ANALYTICS=true
ENABLE_AB_TESTING=true
```

### Database Migrations

```bash
# Run migrations
cd microservices/landing-page-service
go run cmd/migrate/main.go up

# Rollback
go run cmd/migrate/main.go down 1
```

### Build & Run

```bash
# Development
./start-dev.sh

# Production
go build -o landing-page-service cmd/main.go
./landing-page-service

# Docker (if needed in future)
docker build -t beakon/landing-page-service:latest .
docker run -p 8100:8100 beakon/landing-page-service:latest
```

### Nginx Configuration (Reverse Proxy)

```nginx
server {
    listen 80;
    server_name www.beakon.io;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name www.beakon.io;

    ssl_certificate /etc/ssl/certs/beakon.crt;
    ssl_certificate_key /etc/ssl/private/beakon.key;

    # Proxy to landing page service
    location / {
        proxy_pass http://localhost:8100;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Cache static assets
    location /static/ {
        proxy_pass http://localhost:8100;
        proxy_cache_valid 200 30d;
        add_header Cache-Control "public, max-age=2592000";
    }

    # Don't cache admin
    location /admin/ {
        proxy_pass http://localhost:8100;
        proxy_cache_bypass 1;
    }
}
```

### Health Checks

```bash
# Health check endpoint
curl http://localhost:8100/health

# Expected response
{
  "status": "healthy",
  "service": "landing-page-service",
  "version": "2.0.0"
}
```

### Monitoring

**Metrics exposed** (Prometheus):
- `landing_page_requests_total` - Total HTTP requests
- `landing_page_request_duration_seconds` - Request duration
- `landing_page_db_connections` - Database connections
- `landing_page_cache_hits` - Cache hit rate
- `landing_page_ab_test_conversions` - A/B test conversions

**Logs**:
```json
{
  "level": "info",
  "timestamp": "2025-10-14T10:30:00Z",
  "service": "landing-page-service",
  "message": "Request processed",
  "path": "/",
  "method": "GET",
  "status": 200,
  "duration_ms": 45
}
```

---

## Appendix

### File Structure

```
landing-page-service/
├── cmd/
│   ├── main.go                    # Service entry point
│   └── migrate/
│       └── main.go                # Database migrations
├── configs/
│   ├── development.yaml
│   ├── production.yaml
│   └── test.yaml
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   ├── landing_handler.go
│   │   ├── admin_handler.go       # NEW
│   │   ├── seo_handler.go         # NEW
│   │   └── analytics_handler.go   # NEW
│   ├── services/
│   │   ├── landing_service.go
│   │   ├── admin_service.go       # NEW
│   │   ├── seo_service.go         # NEW
│   │   ├── media_service.go       # NEW
│   │   ├── analytics_service.go   # NEW
│   │   └── ab_test_service.go     # NEW
│   ├── models/
│   │   ├── landing.go
│   │   ├── seo.go                 # NEW
│   │   ├── media.go               # NEW
│   │   └── analytics.go           # NEW
│   ├── middleware/
│   │   ├── auth.go
│   │   └── cors.go
│   └── utils/
│       ├── schema.go              # NEW: Schema markup generator
│       └── sitemap.go             # NEW: Sitemap generator
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   ├── landing.css
│   │   │   ├── admin.css          # NEW
│   │   │   └── animations.css
│   │   ├── js/
│   │   │   ├── landing.js
│   │   │   ├── admin.js           # NEW
│   │   │   ├── ab-test.js         # NEW
│   │   │   └── service-worker.js
│   │   ├── images/
│   │   └── uploads/               # NEW: User uploads
│   └── templates/
│       ├── index.html
│       ├── pricing.html
│       ├── blog.html
│       ├── admin/
│       │   ├── dashboard.html     # NEW
│       │   ├── content-editor.html # NEW
│       │   ├── media-library.html  # NEW
│       │   └── seo-config.html     # NEW
│       └── components/
│           ├── status-widget.html  # NEW
│           └── uptime-chart.html   # NEW
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_seo_fields.sql     # NEW
│   ├── 003_add_media_table.sql    # NEW
│   ├── 004_add_ab_tests.sql       # NEW
│   └── 005_add_sections.sql       # NEW
├── tests/
│   ├── integration/
│   └── unit/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── ARCHITECTURE.md
└── LANDING_PAGE_ENHANCEMENT.md    # THIS FILE
```

### Resources

**Documentation**:
- [Schema.org Documentation](https://schema.org/)
- [Google Search Central](https://developers.google.com/search)
- [Open Graph Protocol](https://ogp.me/)
- [Twitter Cards](https://developer.twitter.com/en/docs/twitter-for-websites/cards)
- [Web.dev Performance](https://web.dev/learn-core-web-vitals/)

**Tools**:
- [Google Rich Results Test](https://search.google.com/test/rich-results)
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
- [Schema Markup Validator](https://validator.schema.org/)
- [PageSpeed Insights](https://pagespeed.web.dev/)

---

**Document Version**: 1.0
**Maintained By**: Platform Team
**Last Review**: October 14, 2025
