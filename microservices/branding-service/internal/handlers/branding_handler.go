// Package handlers provides HTTP handlers for the Branding Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/branding-service/internal/models"
	"github.com/anupamdutta5/branding-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BrandingHandler handles branding-related HTTP requests.
type BrandingHandler struct {
	service *services.BrandingService
	logger  *zap.Logger
}

// NewBrandingHandler creates a new branding handler.
func NewBrandingHandler(service *services.BrandingService, logger *zap.Logger) *BrandingHandler {
	return &BrandingHandler{
		service: service,
		logger:  logger,
	}
}

// HealthCheck handles health check requests.
func (h *BrandingHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "branding-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "branding-service",
		"version": "1.0.0",
	})
}

// Brand Management Handlers

// ListBrands handles listing brands.
func (h *BrandingHandler) ListBrands(c *gin.Context) {
	h.logger.Info("Listing brands")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	brands, err := h.service.ListBrands(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list brands", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list brands",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"brands": brands,
		"count":  len(brands),
		"limit":  limit,
		"offset": offset,
	})
}

// CreateBrand handles creating a new brand.
func (h *BrandingHandler) CreateBrand(c *gin.Context) {
	var brand models.Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating brand", zap.String("brand_name", brand.Name))

	if err := h.service.CreateBrand(c.Request.Context(), &brand); err != nil {
		h.logger.Error("Failed to create brand", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create brand",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"brand":   brand,
		"message": "Brand created successfully",
	})
}

// GetBrand handles retrieving a brand by ID.
func (h *BrandingHandler) GetBrand(c *gin.Context) {
	brandIDStr := c.Param("id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	h.logger.Info("Getting brand", zap.Uint64("brand_id", brandID))

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	brand, err := h.service.GetBrand(c.Request.Context(), uint(brandID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get brand", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Brand not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"brand": brand,
	})
}

// GetBrandBySlug handles retrieving a brand by slug.
func (h *BrandingHandler) GetBrandBySlug(c *gin.Context) {
	slug := c.Param("slug")

	h.logger.Info("Getting brand by slug", zap.String("slug", slug))

	brand, err := h.service.GetBrandBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error("Failed to get brand by slug", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Brand not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"brand": brand,
	})
}

// UpdateBrand handles updating a brand.
func (h *BrandingHandler) UpdateBrand(c *gin.Context) {
	brandIDStr := c.Param("id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	var updates models.Brand
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating brand", zap.Uint64("brand_id", brandID))

	if err := h.service.UpdateBrand(c.Request.Context(), uint(brandID), &updates); err != nil {
		h.logger.Error("Failed to update brand", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update brand",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand updated successfully",
	})
}

// DeleteBrand handles deleting a brand.
func (h *BrandingHandler) DeleteBrand(c *gin.Context) {
	brandIDStr := c.Param("id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	h.logger.Info("Deleting brand", zap.Uint64("brand_id", brandID))

	if err := h.service.DeleteBrand(c.Request.Context(), uint(brandID)); err != nil {
		h.logger.Error("Failed to delete brand", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete brand",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand deleted successfully",
	})
}

// Theme Management Handlers

// ListThemes handles listing themes for a brand.
func (h *BrandingHandler) ListThemes(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	h.logger.Info("Listing themes", zap.Uint64("brand_id", brandID))

	themes, err := h.service.ListThemes(c.Request.Context(), uint(brandID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list themes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list themes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"themes": themes,
		"count":  len(themes),
		"limit":  limit,
		"offset": offset,
	})
}

// CreateTheme handles creating a new theme.
func (h *BrandingHandler) CreateTheme(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	var theme models.Theme
	if err := c.ShouldBindJSON(&theme); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	theme.BrandID = uint(brandID)

	h.logger.Info("Creating theme", zap.String("theme_name", theme.Name))

	if err := h.service.CreateTheme(c.Request.Context(), &theme); err != nil {
		h.logger.Error("Failed to create theme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create theme",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"theme":   theme,
		"message": "Theme created successfully",
	})
}

// GetTheme handles retrieving a theme by ID.
func (h *BrandingHandler) GetTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid theme ID",
		})
		return
	}

	h.logger.Info("Getting theme", zap.Uint64("theme_id", themeID))

	theme, err := h.service.GetTheme(c.Request.Context(), uint(themeID))
	if err != nil {
		h.logger.Error("Failed to get theme", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Theme not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"theme": theme,
	})
}

// UpdateTheme handles updating a theme.
func (h *BrandingHandler) UpdateTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid theme ID",
		})
		return
	}

	var updates models.Theme
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating theme", zap.Uint64("theme_id", themeID))

	if err := h.service.UpdateTheme(c.Request.Context(), uint(themeID), &updates); err != nil {
		h.logger.Error("Failed to update theme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update theme",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Theme updated successfully",
	})
}

// DeleteTheme handles deleting a theme.
func (h *BrandingHandler) DeleteTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid theme ID",
		})
		return
	}

	h.logger.Info("Deleting theme", zap.Uint64("theme_id", themeID))

	if err := h.service.DeleteTheme(c.Request.Context(), uint(themeID)); err != nil {
		h.logger.Error("Failed to delete theme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete theme",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Theme deleted successfully",
	})
}

// Asset Management Handlers

// ListAssets handles listing assets for a brand.
func (h *BrandingHandler) ListAssets(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	assetType := c.Query("type")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	h.logger.Info("Listing assets", zap.Uint64("brand_id", brandID), zap.String("asset_type", assetType))

	assets, err := h.service.ListAssets(c.Request.Context(), uint(brandID), assetType, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list assets",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
		"count":  len(assets),
		"limit":  limit,
		"offset": offset,
	})
}

// CreateAsset handles creating a new asset.
func (h *BrandingHandler) CreateAsset(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	var asset models.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	asset.BrandID = uint(brandID)

	h.logger.Info("Creating asset", zap.String("asset_name", asset.Name))

	if err := h.service.CreateAsset(c.Request.Context(), &asset); err != nil {
		h.logger.Error("Failed to create asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create asset",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"asset":   asset,
		"message": "Asset created successfully",
	})
}

// GetAsset handles retrieving an asset by ID.
func (h *BrandingHandler) GetAsset(c *gin.Context) {
	assetIDStr := c.Param("id")
	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid asset ID",
		})
		return
	}

	h.logger.Info("Getting asset", zap.Uint64("asset_id", assetID))

	asset, err := h.service.GetAsset(c.Request.Context(), uint(assetID))
	if err != nil {
		h.logger.Error("Failed to get asset", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Asset not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"asset": asset,
	})
}

// UpdateAsset handles updating an asset.
func (h *BrandingHandler) UpdateAsset(c *gin.Context) {
	assetIDStr := c.Param("id")
	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid asset ID",
		})
		return
	}

	var updates models.Asset
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating asset", zap.Uint64("asset_id", assetID))

	if err := h.service.UpdateAsset(c.Request.Context(), uint(assetID), &updates); err != nil {
		h.logger.Error("Failed to update asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update asset",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset updated successfully",
	})
}

// DeleteAsset handles deleting an asset.
func (h *BrandingHandler) DeleteAsset(c *gin.Context) {
	assetIDStr := c.Param("id")
	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid asset ID",
		})
		return
	}

	h.logger.Info("Deleting asset", zap.Uint64("asset_id", assetID))

	if err := h.service.DeleteAsset(c.Request.Context(), uint(assetID)); err != nil {
		h.logger.Error("Failed to delete asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete asset",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset deleted successfully",
	})
}

// UploadAsset handles uploading an asset file with multi-tenant support.
func (h *BrandingHandler) UploadAsset(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	// Get tenant ID from context (set by auth middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant ID not found in context",
		})
		return
	}

	// Parse multipart form
	err = c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse multipart form",
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	defer file.Close()

	// Get asset metadata from form
	name := c.PostForm("name")
	assetType := c.PostForm("type")
	category := c.PostForm("category")
	description := c.PostForm("description")
	altText := c.PostForm("alt_text")

	if name == "" || assetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name and type are required",
		})
		return
	}

	// Create asset model
	asset := &models.Asset{
		BrandID:     uint(brandID),
		Name:        name,
		Type:        assetType,
		Category:    category,
		Description: description,
		AltText:     altText,
		Status:      "active",
	}

	h.logger.Info("Uploading asset file",
		zap.String("filename", header.Filename),
		zap.String("asset_type", assetType),
		zap.Uint64("brand_id", brandID))

	// Use enhanced service method with tenant validation
	if err := h.service.CreateAssetWithFile(c.Request.Context(), asset, file, header, tenantID.(uint)); err != nil {
		h.logger.Error("Failed to upload asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload asset: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"asset":   asset,
		"message": "Asset uploaded successfully",
	})
}

// Custom CSS Management Handlers

// ListCustomCSS handles listing custom CSS for a brand.
func (h *BrandingHandler) ListCustomCSS(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	h.logger.Info("Listing custom CSS", zap.Uint64("brand_id", brandID))

	customCSS, err := h.service.ListCustomCSS(c.Request.Context(), uint(brandID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list custom CSS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list custom CSS",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"custom_css": customCSS,
		"count":      len(customCSS),
		"limit":      limit,
		"offset":     offset,
	})
}

// CreateCustomCSS handles creating custom CSS.
func (h *BrandingHandler) CreateCustomCSS(c *gin.Context) {
	brandIDStr := c.Param("brand_id")
	brandID, err := strconv.ParseUint(brandIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid brand ID",
		})
		return
	}

	var customCSS models.CustomCSS
	if err := c.ShouldBindJSON(&customCSS); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	customCSS.BrandID = uint(brandID)

	h.logger.Info("Creating custom CSS", zap.String("css_name", customCSS.Name))

	if err := h.service.CreateCustomCSS(c.Request.Context(), &customCSS); err != nil {
		h.logger.Error("Failed to create custom CSS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create custom CSS",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"custom_css": customCSS,
		"message":    "Custom CSS created successfully",
	})
}

// GetCustomCSS handles retrieving custom CSS by ID.
func (h *BrandingHandler) GetCustomCSS(c *gin.Context) {
	cssIDStr := c.Param("id")
	cssID, err := strconv.ParseUint(cssIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid CSS ID",
		})
		return
	}

	h.logger.Info("Getting custom CSS", zap.Uint64("css_id", cssID))

	customCSS, err := h.service.GetCustomCSS(c.Request.Context(), uint(cssID))
	if err != nil {
		h.logger.Error("Failed to get custom CSS", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Custom CSS not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"custom_css": customCSS,
	})
}

// UpdateCustomCSS handles updating custom CSS.
func (h *BrandingHandler) UpdateCustomCSS(c *gin.Context) {
	cssIDStr := c.Param("id")
	cssID, err := strconv.ParseUint(cssIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid CSS ID",
		})
		return
	}

	var updates models.CustomCSS
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating custom CSS", zap.Uint64("css_id", cssID))

	if err := h.service.UpdateCustomCSS(c.Request.Context(), uint(cssID), &updates); err != nil {
		h.logger.Error("Failed to update custom CSS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update custom CSS",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom CSS updated successfully",
	})
}

// DeleteCustomCSS handles deleting custom CSS.
func (h *BrandingHandler) DeleteCustomCSS(c *gin.Context) {
	cssIDStr := c.Param("id")
	cssID, err := strconv.ParseUint(cssIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid CSS ID",
		})
		return
	}

	h.logger.Info("Deleting custom CSS", zap.Uint64("css_id", cssID))

	if err := h.service.DeleteCustomCSS(c.Request.Context(), uint(cssID)); err != nil {
		h.logger.Error("Failed to delete custom CSS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete custom CSS",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom CSS deleted successfully",
	})
}

// Statistics Handlers

// GetStats handles getting branding statistics.
func (h *BrandingHandler) GetStats(c *gin.Context) {
	h.logger.Info("Getting branding statistics")

	stats, err := h.service.GetBrandingStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Color Scheme Management Handlers

func (h *BrandingHandler) ListColorSchemes(c *gin.Context) {
	themeIDStr := c.Param("theme_id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme ID"})
		return
	}

	// For now, return empty list - would implement database query
	c.JSON(http.StatusOK, gin.H{
		"color_schemes": []interface{}{},
		"theme_id":      themeID,
	})
}

func (h *BrandingHandler) CreateColorScheme(c *gin.Context) {
	themeIDStr := c.Param("theme_id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme ID"})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var colorScheme models.ColorScheme
	if err := c.ShouldBindJSON(&colorScheme); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	colorScheme.ThemeID = uint(themeID)

	if err := h.service.CreateColorScheme(c.Request.Context(), &colorScheme, tenantID.(uint)); err != nil {
		h.logger.Error("Failed to create color scheme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"color_scheme": colorScheme,
		"message":      "Color scheme created successfully",
	})
}

func (h *BrandingHandler) GetColorScheme(c *gin.Context) {
	colorSchemeIDStr := c.Param("id")
	colorSchemeID, err := strconv.ParseUint(colorSchemeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid color scheme ID"})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	colorScheme, err := h.service.GetColorScheme(c.Request.Context(), uint(colorSchemeID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get color scheme", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Color scheme not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"color_scheme": colorScheme})
}

func (h *BrandingHandler) UpdateColorScheme(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update color scheme not implemented"})
}

func (h *BrandingHandler) DeleteColorScheme(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete color scheme not implemented"})
}

// Typography Management Handlers

func (h *BrandingHandler) ListTypographies(c *gin.Context) {
	themeIDStr := c.Param("theme_id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"typographies": []interface{}{},
		"theme_id":     themeID,
	})
}

func (h *BrandingHandler) CreateTypography(c *gin.Context) {
	themeIDStr := c.Param("theme_id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme ID"})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	var typography models.Typography
	if err := c.ShouldBindJSON(&typography); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	typography.ThemeID = uint(themeID)

	if err := h.service.CreateTypography(c.Request.Context(), &typography, tenantID.(uint)); err != nil {
		h.logger.Error("Failed to create typography", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"typography": typography,
		"message":    "Typography created successfully",
	})
}

func (h *BrandingHandler) GetTypography(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get typography not implemented"})
}

func (h *BrandingHandler) UpdateTypography(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update typography not implemented"})
}

func (h *BrandingHandler) DeleteTypography(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete typography not implemented"})
}

func (h *BrandingHandler) ListCustomJS(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) CreateCustomJS(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetCustomJS(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) UpdateCustomJS(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) DeleteCustomJS(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) ListLayouts(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) CreateLayout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetLayout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) UpdateLayout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) DeleteLayout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) ListComponents(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) CreateComponent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetComponent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) UpdateComponent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) DeleteComponent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetBrandStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetThemeStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetPublicBrand(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetPublicTheme(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *BrandingHandler) GetPublicAsset(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// CompileTheme handles theme compilation for preview/export
func (h *BrandingHandler) CompileTheme(c *gin.Context) {
	themeIDStr := c.Param("id")
	themeID, err := strconv.ParseUint(themeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme ID"})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	h.logger.Info("Compiling theme", zap.Uint64("theme_id", themeID))

	compiledTheme, err := h.service.CompileTheme(c.Request.Context(), uint(themeID), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to compile theme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compiled_theme": compiledTheme,
		"message":        "Theme compiled successfully",
	})
}

// GetTenantStats handles getting tenant-specific branding statistics
func (h *BrandingHandler) GetTenantStats(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		return
	}

	h.logger.Info("Getting tenant branding statistics", zap.Uint("tenant_id", tenantID.(uint)))

	stats, err := h.service.GetTenantBrandingStats(c.Request.Context(), tenantID.(uint))
	if err != nil {
		h.logger.Error("Failed to get tenant statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

