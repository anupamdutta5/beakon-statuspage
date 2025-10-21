// Package handlers provides HTTP handlers for the Landing Page Service.
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MediaHandler handles media upload and management.
type MediaHandler struct {
	db          *gorm.DB
	logger      *zap.Logger
	uploadPath  string
	maxFileSize int64
}

// NewMediaHandler creates a new media handler.
func NewMediaHandler(db *gorm.DB, logger *zap.Logger) *MediaHandler {
	return &MediaHandler{
		db:          db,
		logger:      logger,
		uploadPath:  "web/static/uploads",
		maxFileSize: 10 * 1024 * 1024, // 10MB default
	}
}

// UploadMedia handles POST /api/v1/admin/media/upload
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	h.logger.Info("Processing media upload")

	// Parse multipart form (max 32 MB)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get file from form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File size exceeds maximum allowed (%d MB)", h.maxFileSize/1024/1024),
		})
		return
	}

	// Validate file type
	if !h.isAllowedFileType(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File type not allowed",
		})
		return
	}

	// Get additional metadata from form
	category := c.PostForm("category")
	altText := c.PostForm("alt_text")
	title := c.PostForm("title")
	caption := c.PostForm("caption")

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := h.generateUniqueFilename(ext)

	// Create upload directory if it doesn't exist
	uploadDir := filepath.Join(h.uploadPath, time.Now().Format("2006/01"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.logger.Error("Failed to create upload directory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create upload directory",
		})
		return
	}

	// Save file to disk
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("Failed to create file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		h.logger.Error("Failed to save file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}

	// Get image dimensions if it's an image
	width, height := h.getImageDimensions(filePath)

	// Create database record
	media := models.LandingPageMedia{
		Filename:         filename,
		OriginalFilename: header.Filename,
		MimeType:         header.Header.Get("Content-Type"),
		SizeBytes:        header.Size,
		Width:            width,
		Height:           height,
		StoragePath:      "/" + strings.ReplaceAll(filePath, "\\", "/"),
		AltText:          altText,
		Title:            title,
		Caption:          caption,
		Category:         category,
	}

	if err := h.db.Create(&media).Error; err != nil {
		h.logger.Error("Failed to save media record", zap.Error(err))
		// Clean up uploaded file
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save media record",
		})
		return
	}

	h.logger.Info("Media uploaded successfully",
		zap.String("filename", filename),
		zap.String("category", category))

	// Queue for optimization (WebP/AVIF generation) - async
	go h.optimizeImage(media.ID, filePath)

	c.JSON(http.StatusCreated, gin.H{
		"media":   media,
		"message": "File uploaded successfully",
	})
}

// GetMediaList handles GET /api/v1/admin/media
func (h *MediaHandler) GetMediaList(c *gin.Context) {
	h.logger.Info("Getting media list")

	var media []models.LandingPageMedia
	query := h.db.Order("created_at DESC")

	// Apply filters
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if mimeType := c.Query("mime_type"); mimeType != "" {
		if strings.Contains(mimeType, "image") {
			query = query.Where("mime_type LIKE ?", "image/%")
		} else if strings.Contains(mimeType, "video") {
			query = query.Where("mime_type LIKE ?", "video/%")
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query = query.Limit(limit).Offset(offset)

	if err := query.Find(&media).Error; err != nil {
		h.logger.Error("Failed to get media list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get media list",
		})
		return
	}

	// Get total count
	var total int64
	h.db.Model(&models.LandingPageMedia{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"media": media,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetMedia handles GET /api/v1/admin/media/:id
func (h *MediaHandler) GetMedia(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Getting media", zap.Uint("id", id))

	var media models.LandingPageMedia
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}
		h.logger.Error("Failed to get media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media": media,
	})
}

// UpdateMedia handles PUT /api/v1/admin/media/:id
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updates struct {
		AltText  string `json:"alt_text"`
		Title    string `json:"title"`
		Caption  string `json:"caption"`
		Category string `json:"category"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Updating media", zap.Uint("id", id))

	if err := h.db.Model(&models.LandingPageMedia{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		h.logger.Error("Failed to update media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update media",
		})
		return
	}

	// Update usage timestamp
	now := time.Now()
	h.db.Model(&models.LandingPageMedia{}).Where("id = ?", id).Update("last_used_at", now)

	c.JSON(http.StatusOK, gin.H{
		"message": "Media updated successfully",
	})
}

// DeleteMedia handles DELETE /api/v1/admin/media/:id
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Deleting media", zap.Uint("id", id))

	// Get media record
	var media models.LandingPageMedia
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}
		h.logger.Error("Failed to get media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media"})
		return
	}

	// Delete physical files
	filesToDelete := []string{
		media.StoragePath,
		media.WebPPath,
		media.AVIFPath,
		media.ThumbnailPath,
	}

	for _, file := range filesToDelete {
		if file != "" {
			// Remove leading slash and convert to proper path
			filePath := strings.TrimPrefix(file, "/")
			if err := os.Remove(filePath); err != nil {
				h.logger.Warn("Failed to delete file",
					zap.String("file", filePath),
					zap.Error(err))
			}
		}
	}

	// Delete database record
	if err := h.db.Delete(&media).Error; err != nil {
		h.logger.Error("Failed to delete media record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete media",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Media deleted successfully",
	})
}

// OptimizeMedia handles POST /api/v1/admin/media/:id/optimize
func (h *MediaHandler) OptimizeMedia(c *gin.Context) {
	id := h.parseID(c)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	h.logger.Info("Optimizing media", zap.Uint("id", id))

	// Get media record
	var media models.LandingPageMedia
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
			return
		}
		h.logger.Error("Failed to get media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media"})
		return
	}

	// Check if it's an image
	if !strings.HasPrefix(media.MimeType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only images can be optimized",
		})
		return
	}

	// Trigger optimization
	go h.optimizeImage(media.ID, strings.TrimPrefix(media.StoragePath, "/"))

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Optimization started",
	})
}

// BulkDelete handles POST /api/v1/admin/media/bulk-delete
func (h *MediaHandler) BulkDelete(c *gin.Context) {
	var request struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid bulk delete data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data",
		})
		return
	}

	h.logger.Info("Bulk deleting media", zap.Int("count", len(request.IDs)))

	// Get media records
	var mediaList []models.LandingPageMedia
	if err := h.db.Where("id IN ?", request.IDs).Find(&mediaList).Error; err != nil {
		h.logger.Error("Failed to get media records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get media records",
		})
		return
	}

	// Delete physical files
	for _, media := range mediaList {
		filesToDelete := []string{
			media.StoragePath,
			media.WebPPath,
			media.AVIFPath,
			media.ThumbnailPath,
		}

		for _, file := range filesToDelete {
			if file != "" {
				filePath := strings.TrimPrefix(file, "/")
				if err := os.Remove(filePath); err != nil {
					h.logger.Warn("Failed to delete file",
						zap.String("file", filePath),
						zap.Error(err))
				}
			}
		}
	}

	// Delete database records
	if err := h.db.Where("id IN ?", request.IDs).Delete(&models.LandingPageMedia{}).Error; err != nil {
		h.logger.Error("Failed to delete media records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete media",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d media files deleted successfully", len(request.IDs)),
	})
}

// GetMediaStats handles GET /api/v1/admin/media/stats
func (h *MediaHandler) GetMediaStats(c *gin.Context) {
	h.logger.Info("Getting media statistics")

	stats := gin.H{}

	// Total count and size
	var totalCount int64
	var totalSize int64
	h.db.Model(&models.LandingPageMedia{}).Count(&totalCount)
	h.db.Model(&models.LandingPageMedia{}).Select("SUM(size_bytes)").Scan(&totalSize)

	stats["total_files"] = totalCount
	stats["total_size_mb"] = float64(totalSize) / 1024 / 1024

	// Count by type
	var imageCount, videoCount, otherCount int64
	h.db.Model(&models.LandingPageMedia{}).Where("mime_type LIKE ?", "image/%").Count(&imageCount)
	h.db.Model(&models.LandingPageMedia{}).Where("mime_type LIKE ?", "video/%").Count(&videoCount)
	otherCount = totalCount - imageCount - videoCount

	stats["by_type"] = gin.H{
		"images": imageCount,
		"videos": videoCount,
		"other":  otherCount,
	}

	// Count by category
	type CategoryCount struct {
		Category string
		Count    int64
	}
	var categoryCounts []CategoryCount
	h.db.Model(&models.LandingPageMedia{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryCounts)

	categoryMap := gin.H{}
	for _, cc := range categoryCounts {
		if cc.Category == "" {
			categoryMap["uncategorized"] = cc.Count
		} else {
			categoryMap[cc.Category] = cc.Count
		}
	}
	stats["by_category"] = categoryMap

	// Optimization stats
	var optimizedCount int64
	h.db.Model(&models.LandingPageMedia{}).Where("is_optimized = ?", true).Count(&optimizedCount)
	stats["optimized"] = optimizedCount
	stats["optimization_rate"] = float64(optimizedCount) / float64(totalCount) * 100

	// Recent uploads
	var recentUploads []models.LandingPageMedia
	h.db.Order("created_at DESC").Limit(5).Find(&recentUploads)
	stats["recent_uploads"] = recentUploads

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ============================================================================
// Helper Methods
// ============================================================================

// parseID parses ID from URL parameter
func (h *MediaHandler) parseID(c *gin.Context) uint {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}

// generateUniqueFilename generates a unique filename
func (h *MediaHandler) generateUniqueFilename(ext string) string {
	timestamp := time.Now().UnixNano()
	randomStr := strconv.FormatInt(timestamp, 36)
	return fmt.Sprintf("%s%s", randomStr, ext)
}

// isAllowedFileType checks if the file type is allowed
func (h *MediaHandler) isAllowedFileType(filename string) bool {
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
		".mp4":  true,
		".webm": true,
		".pdf":  true,
		".ico":  true,
	}

	ext := strings.ToLower(filepath.Ext(filename))
	return allowedExtensions[ext]
}

// getImageDimensions gets the width and height of an image
func (h *MediaHandler) getImageDimensions(filePath string) (width, height int) {
	// This is a placeholder - in production, use an image processing library
	// like github.com/disintegration/imaging or github.com/h2non/bimg

	// For now, return default values
	return 0, 0
}

// optimizeImage optimizes an image by creating WebP and AVIF versions
func (h *MediaHandler) optimizeImage(mediaID uint, filePath string) {
	h.logger.Info("Starting image optimization",
		zap.Uint("media_id", mediaID),
		zap.String("file", filePath))

	// This is a placeholder for actual image optimization
	// In production, you would:
	// 1. Use an image processing library (e.g., github.com/disintegration/imaging)
	// 2. Generate WebP version
	// 3. Generate AVIF version (if supported)
	// 4. Generate thumbnail
	// 5. Update database with new paths

	// For now, just mark as optimized
	time.Sleep(2 * time.Second) // Simulate processing

	updates := map[string]interface{}{
		"is_optimized": true,
		// "webp_path": "/path/to/webp",
		// "avif_path": "/path/to/avif",
		// "thumbnail_path": "/path/to/thumbnail",
	}

	if err := h.db.Model(&models.LandingPageMedia{}).Where("id = ?", mediaID).Updates(updates).Error; err != nil {
		h.logger.Error("Failed to update optimization status",
			zap.Uint("media_id", mediaID),
			zap.Error(err))
	} else {
		h.logger.Info("Image optimization completed",
			zap.Uint("media_id", mediaID))
	}
}