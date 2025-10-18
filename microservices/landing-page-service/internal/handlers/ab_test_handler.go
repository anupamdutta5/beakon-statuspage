// Package handlers provides HTTP handlers for the Landing Page Service.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/landing-page-service/internal/models"
	"github.com/anupamdutta5/landing-page-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ABTestHandler handles A/B testing-related HTTP requests.
type ABTestHandler struct {
	service *services.ABTestService
	logger  *zap.Logger
}

// NewABTestHandler creates a new A/B test handler.
func NewABTestHandler(service *services.ABTestService, logger *zap.Logger) *ABTestHandler {
	return &ABTestHandler{
		service: service,
		logger:  logger,
	}
}

// CreateABTest handles POST /api/v1/admin/ab-tests
func (h *ABTestHandler) CreateABTest(c *gin.Context) {
	var test models.LandingPageABTest
	if err := c.ShouldBindJSON(&test); err != nil {
		h.logger.Error("Invalid A/B test data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid test data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Creating A/B test", zap.String("name", test.Name))

	if err := h.service.CreateABTest(c.Request.Context(), &test); err != nil {
		h.logger.Error("Failed to create A/B test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create A/B test",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"test":    test,
		"message": "A/B test created successfully",
	})
}

// GetAllABTests handles GET /api/v1/admin/ab-tests
func (h *ABTestHandler) GetAllABTests(c *gin.Context) {
	status := c.Query("status") // Optional filter by status
	h.logger.Info("Getting all A/B tests", zap.String("status_filter", status))

	tests, err := h.service.GetAllTests(c.Request.Context(), status)
	if err != nil {
		h.logger.Error("Failed to get A/B tests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get A/B tests",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tests": tests,
		"count": len(tests),
	})
}

// GetABTest handles GET /api/v1/admin/ab-tests/:id
func (h *ABTestHandler) GetABTest(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	h.logger.Info("Getting A/B test", zap.Uint64("test_id", testID))

	test, err := h.service.GetTestByID(c.Request.Context(), uint(testID))
	if err != nil {
		h.logger.Error("Failed to get A/B test", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Test not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"test": test,
	})
}

// UpdateABTest handles PUT /api/v1/admin/ab-tests/:id
func (h *ABTestHandler) UpdateABTest(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid update data",
		})
		return
	}

	h.logger.Info("Updating A/B test", zap.Uint64("test_id", testID))

	if err := h.service.UpdateABTest(c.Request.Context(), uint(testID), updates); err != nil {
		h.logger.Error("Failed to update A/B test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A/B test updated successfully",
	})
}

// DeleteABTest handles DELETE /api/v1/admin/ab-tests/:id
func (h *ABTestHandler) DeleteABTest(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	h.logger.Info("Deleting A/B test", zap.Uint64("test_id", testID))

	if err := h.service.DeleteABTest(c.Request.Context(), uint(testID)); err != nil {
		h.logger.Error("Failed to delete A/B test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A/B test deleted successfully",
	})
}

// StartABTest handles POST /api/v1/admin/ab-tests/:id/start
func (h *ABTestHandler) StartABTest(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	h.logger.Info("Starting A/B test", zap.Uint64("test_id", testID))

	if err := h.service.StartABTest(c.Request.Context(), uint(testID)); err != nil {
		h.logger.Error("Failed to start A/B test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A/B test started successfully",
	})
}

// StopABTest handles POST /api/v1/admin/ab-tests/:id/stop
func (h *ABTestHandler) StopABTest(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	h.logger.Info("Stopping A/B test", zap.Uint64("test_id", testID))

	if err := h.service.StopABTest(c.Request.Context(), uint(testID)); err != nil {
		h.logger.Error("Failed to stop A/B test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A/B test stopped successfully",
	})
}

// GetABTestResults handles GET /api/v1/admin/ab-tests/:id/results
func (h *ABTestHandler) GetABTestResults(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	h.logger.Info("Getting A/B test results", zap.Uint64("test_id", testID))

	results, err := h.service.CalculateResults(c.Request.Context(), uint(testID))
	if err != nil {
		h.logger.Error("Failed to get A/B test results", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get results",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// GetABTestReport handles GET /api/v1/admin/ab-tests/:id/report
func (h *ABTestHandler) GetABTestReport(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	h.logger.Info("Generating A/B test report", zap.Uint64("test_id", testID))

	report, err := h.service.GenerateTestReport(c.Request.Context(), uint(testID))
	if err != nil {
		h.logger.Error("Failed to generate report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate report",
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Public endpoints for variant assignment and tracking

// GetVariant handles GET /api/v1/public/ab-test/:id/variant
// This is called by the frontend to get variant assignment
func (h *ABTestHandler) GetVariant(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	// Get session ID from cookie or generate new one
	sessionID, err := c.Cookie("ab_session_id")
	if err != nil || sessionID == "" {
		// Generate new session ID
		sessionID = generateSessionID()
		c.SetCookie("ab_session_id", sessionID, 60*60*24*30, "/", "", false, true)
	}

	h.logger.Debug("Getting variant for session",
		zap.Uint64("test_id", testID),
		zap.String("session_id", sessionID))

	variant, err := h.service.GetVariant(c.Request.Context(), uint(testID), sessionID)
	if err != nil {
		h.logger.Error("Failed to get variant", zap.Error(err))
		// Return control variant on error
		c.JSON(http.StatusOK, gin.H{
			"variant": "A",
			"config":  map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"variant": variant.Name,
		"config":  variant.Config,
	})
}

// TrackConversion handles POST /api/v1/public/ab-test/:id/conversion
// This is called when a conversion event occurs
func (h *ABTestHandler) TrackConversion(c *gin.Context) {
	testIDStr := c.Param("id")
	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid test ID",
		})
		return
	}

	// Get session ID from cookie
	sessionID, err := c.Cookie("ab_session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No session found",
		})
		return
	}

	h.logger.Info("Tracking conversion",
		zap.Uint64("test_id", testID),
		zap.String("session_id", sessionID))

	if err := h.service.TrackConversion(c.Request.Context(), uint(testID), sessionID); err != nil {
		h.logger.Error("Failed to track conversion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to track conversion",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversion tracked",
	})
}

// Helper function to generate session ID
func generateSessionID() string {
	// Simple session ID generation - in production, use a more robust method
	return "sess_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// RegisterRoutes registers all A/B testing routes
func (h *ABTestHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Admin routes (require authentication)
	admin := router.Group("/admin/ab-tests")
	{
		admin.GET("", h.GetAllABTests)
		admin.POST("", h.CreateABTest)
		admin.GET("/:id", h.GetABTest)
		admin.PUT("/:id", h.UpdateABTest)
		admin.DELETE("/:id", h.DeleteABTest)
		admin.POST("/:id/start", h.StartABTest)
		admin.POST("/:id/stop", h.StopABTest)
		admin.GET("/:id/results", h.GetABTestResults)
		admin.GET("/:id/report", h.GetABTestReport)
	}

	// Public routes (no authentication required)
	public := router.Group("/public/ab-test")
	{
		public.GET("/:id/variant", h.GetVariant)
		public.POST("/:id/conversion", h.TrackConversion)
	}
}