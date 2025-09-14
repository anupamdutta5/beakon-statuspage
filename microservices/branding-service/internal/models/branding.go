// Package models provides data models for the Branding Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Brand represents a brand configuration.
type Brand struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"not null;index" json:"name"`
	Slug        string         `gorm:"not null;uniqueIndex" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
	Themes      []Theme        `gorm:"foreignKey:BrandID" json:"themes,omitempty"`
}

// Theme represents a theme configuration.
type Theme struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	BrandID     uint           `gorm:"not null;index" json:"brand_id"`
	Brand       Brand          `gorm:"foreignKey:BrandID" json:"brand"`
	Name        string         `gorm:"not null;index" json:"name"`
	Slug        string         `gorm:"not null;index" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Version     string         `gorm:"default:1.0.0" json:"version"`
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsDefault   bool           `gorm:"default:false" json:"is_default"`
	IsPublic    bool           `gorm:"default:true" json:"is_public"`
	PreviewURL  string         `json:"preview_url"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// ColorScheme represents a color scheme configuration.
type ColorScheme struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ThemeID       uint           `gorm:"not null;index" json:"theme_id"`
	Theme         Theme          `gorm:"foreignKey:ThemeID" json:"theme"`
	Name          string         `gorm:"not null;index" json:"name"`
	Slug          string         `gorm:"not null;index" json:"slug"`
	Primary       string         `gorm:"not null" json:"primary"`    // Primary color hex
	Secondary     string         `gorm:"not null" json:"secondary"`  // Secondary color hex
	Accent        string         `json:"accent"`                     // Accent color hex
	Background    string         `gorm:"not null" json:"background"` // Background color hex
	Surface       string         `gorm:"not null" json:"surface"`    // Surface color hex
	Text          string         `gorm:"not null" json:"text"`       // Text color hex
	TextSecondary string         `json:"text_secondary"`             // Secondary text color hex
	Success       string         `json:"success"`                    // Success color hex
	Warning       string         `json:"warning"`                    // Warning color hex
	Error         string         `json:"error"`                      // Error color hex
	Info          string         `json:"info"`                       // Info color hex
	IsDefault     bool           `gorm:"default:false" json:"is_default"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Typography represents typography configuration.
type Typography struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ThemeID       uint           `gorm:"not null;index" json:"theme_id"`
	Theme         Theme          `gorm:"foreignKey:ThemeID" json:"theme"`
	Name          string         `gorm:"not null;index" json:"name"`
	Slug          string         `gorm:"not null;index" json:"slug"`
	FontFamily    string         `gorm:"not null" json:"font_family"`
	FontSize      string         `gorm:"not null" json:"font_size"`   // Base font size
	LineHeight    string         `gorm:"not null" json:"line_height"` // Line height
	FontWeight    string         `gorm:"not null" json:"font_weight"` // Font weight
	FontStyle     string         `json:"font_style"`                  // Font style
	LetterSpacing string         `json:"letter_spacing"`              // Letter spacing
	TextTransform string         `json:"text_transform"`              // Text transform
	IsDefault     bool           `gorm:"default:false" json:"is_default"`
	Metadata      string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Asset represents a brand asset (logo, favicon, etc.).
type Asset struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	BrandID      uint           `gorm:"not null;index" json:"brand_id"`
	Brand        Brand          `gorm:"foreignKey:BrandID" json:"brand"`
	Name         string         `gorm:"not null;index" json:"name"`
	Type         string         `gorm:"not null;index" json:"type"`     // logo, favicon, background, icon, font, css, js
	Category     string         `gorm:"not null;index" json:"category"` // primary, secondary, accent, utility
	Filename     string         `gorm:"not null" json:"filename"`
	OriginalName string         `gorm:"not null" json:"original_name"`
	MimeType     string         `gorm:"not null" json:"mime_type"`
	Size         int64          `gorm:"not null" json:"size"`
	Width        int            `json:"width"`  // For images
	Height       int            `json:"height"` // For images
	URL          string         `gorm:"not null" json:"url"`
	ThumbnailURL string         `json:"thumbnail_url"` // For images
	AltText      string         `json:"alt_text"`
	Description  string         `gorm:"type:text" json:"description"`
	Status       string         `gorm:"default:active" json:"status"` // active, inactive, processing, error
	IsDefault    bool           `gorm:"default:false" json:"is_default"`
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// CustomCSS represents custom CSS configuration.
type CustomCSS struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	BrandID     uint           `gorm:"not null;index" json:"brand_id"`
	Brand       Brand          `gorm:"foreignKey:BrandID" json:"brand"`
	Name        string         `gorm:"not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CSS         string         `gorm:"type:text;not null" json:"css"`
	Version     string         `gorm:"default:1.0.0" json:"version"`
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsMinified  bool           `gorm:"default:false" json:"is_minified"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// CustomJS represents custom JavaScript configuration.
type CustomJS struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	BrandID     uint           `gorm:"not null;index" json:"brand_id"`
	Brand       Brand          `gorm:"foreignKey:BrandID" json:"brand"`
	Name        string         `gorm:"not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	JavaScript  string         `gorm:"type:text;not null" json:"javascript"`
	Version     string         `gorm:"default:1.0.0" json:"version"`
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsMinified  bool           `gorm:"default:false" json:"is_minified"`
	LoadOrder   int            `gorm:"default:0" json:"load_order"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Layout represents layout configuration.
type Layout struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ThemeID   uint           `gorm:"not null;index" json:"theme_id"`
	Theme     Theme          `gorm:"foreignKey:ThemeID" json:"theme"`
	Name      string         `gorm:"not null;index" json:"name"`
	Slug      string         `gorm:"not null;index" json:"slug"`
	Type      string         `gorm:"not null;index" json:"type"`       // header, footer, sidebar, main, component
	Config    string         `gorm:"type:text;not null" json:"config"` // JSON configuration
	CSS       string         `gorm:"type:text" json:"css"`
	HTML      string         `gorm:"type:text" json:"html"`
	Status    string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsDefault bool           `gorm:"default:false" json:"is_default"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Component represents a UI component configuration.
type Component struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ThemeID    uint           `gorm:"not null;index" json:"theme_id"`
	Theme      Theme          `gorm:"foreignKey:ThemeID" json:"theme"`
	Name       string         `gorm:"not null;index" json:"name"`
	Slug       string         `gorm:"not null;index" json:"slug"`
	Type       string         `gorm:"not null;index" json:"type"`       // button, card, modal, form, navigation, etc.
	Category   string         `gorm:"not null;index" json:"category"`   // ui, layout, interactive, utility
	Config     string         `gorm:"type:text;not null" json:"config"` // JSON configuration
	CSS        string         `gorm:"type:text" json:"css"`
	HTML       string         `gorm:"type:text" json:"html"`
	JavaScript string         `gorm:"type:text" json:"javascript"`
	Status     string         `gorm:"default:active" json:"status"` // active, inactive, draft
	IsDefault  bool           `gorm:"default:false" json:"is_default"`
	IsPublic   bool           `gorm:"default:true" json:"is_public"`
	Metadata   string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// BrandingStats represents branding statistics.
type BrandingStats struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	TotalBrands     int       `json:"total_brands"`
	TotalThemes     int       `json:"total_themes"`
	TotalAssets     int       `json:"total_assets"`
	TotalCustomCSS  int       `json:"total_custom_css"`
	TotalCustomJS   int       `json:"total_custom_js"`
	TotalLayouts    int       `json:"total_layouts"`
	TotalComponents int       `json:"total_components"`
	LastUpdated     time.Time `json:"last_updated"`
}

// TableName returns the table name for Brand.
func (Brand) TableName() string {
	return "brands"
}

// TableName returns the table name for Theme.
func (Theme) TableName() string {
	return "themes"
}

// TableName returns the table name for ColorScheme.
func (ColorScheme) TableName() string {
	return "color_schemes"
}

// TableName returns the table name for Typography.
func (Typography) TableName() string {
	return "typographies"
}

// TableName returns the table name for Asset.
func (Asset) TableName() string {
	return "assets"
}

// TableName returns the table name for CustomCSS.
func (CustomCSS) TableName() string {
	return "custom_css"
}

// TableName returns the table name for CustomJS.
func (CustomJS) TableName() string {
	return "custom_js"
}

// TableName returns the table name for Layout.
func (Layout) TableName() string {
	return "layouts"
}

// TableName returns the table name for Component.
func (Component) TableName() string {
	return "components"
}

// TableName returns the table name for BrandingStats.
func (BrandingStats) TableName() string {
	return "branding_stats"
}
