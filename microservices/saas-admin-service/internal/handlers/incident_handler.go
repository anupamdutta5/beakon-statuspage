// Package handlers provides HTTP handlers for incident management.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anupamdutta5/saas-admin-service/internal/models"
	"github.com/anupamdutta5/saas-admin-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// IncidentHandler handles incident management HTTP requests.
type IncidentHandler struct {
	service *services.SaaSAdminService
	logger  *zap.Logger
}

// NewIncidentHandler creates a new incident handler.
func NewIncidentHandler(service *services.SaaSAdminService, logger *zap.Logger) *IncidentHandler {
	return &IncidentHandler{
		service: service,
		logger:  logger,
	}
}

// CreateIncidentRequest represents the request body for creating an incident.
type CreateIncidentRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	Status          string `json:"status" binding:"required"`
	Impact          string `json:"impact" binding:"required"`
	ComponentStatus string `json:"component_status"`
	TenantID        string `json:"tenant_id" binding:"required"`
	CreatedBy       string `json:"created_by"`
	IsVisible       bool   `json:"is_visible"`
}

// UpdateIncidentRequest represents the request body for updating an incident.
type UpdateIncidentRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	Impact          string `json:"impact"`
	ComponentStatus string `json:"component_status"`
	UpdatedBy       string `json:"updated_by"`
	IsVisible       *bool  `json:"is_visible"`
}

// CreateIncidentUpdateRequest represents the request body for creating an incident update.
type CreateIncidentUpdateRequest struct {
	Status     string `json:"status" binding:"required"`
	Message    string `json:"message" binding:"required"`
	CreatedBy  string `json:"created_by"`
	IsPublic   bool   `json:"is_public"`
	UpdateType string `json:"update_type"`
}

// GetIncidents handles getting all incidents with filtering and pagination.
func (h *IncidentHandler) GetIncidents(c *gin.Context) {
	h.logger.Info("Getting incidents")

	// Get query parameters
	tenantID := c.Query("tenant_id")
	status := c.Query("status")
	impact := c.Query("impact")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Mock incident data for now
	incidents := []models.SaaSIncident{
		{
			ID:              uuid.New(),
			Title:           "API Gateway Performance Issues",
			Description:     "Users may experience slow response times when accessing our API endpoints",
			Status:          "monitoring",
			Impact:          "minor",
			ComponentStatus: "degraded_performance",
			StartedAt:       time.Now().Add(-2 * time.Hour),
			TenantID:        "tenant-1",
			CreatedBy:       "admin@example.com",
			IsVisible:       true,
			CreatedAt:       time.Now().Add(-2 * time.Hour),
			UpdatedAt:       time.Now().Add(-30 * time.Minute),
		},
		{
			ID:              uuid.New(),
			Title:           "Database Connectivity Issues",
			Description:     "We are investigating reports of database connection timeouts",
			Status:          "investigating",
			Impact:          "major",
			ComponentStatus: "partial_outage",
			StartedAt:       time.Now().Add(-45 * time.Minute),
			TenantID:        "tenant-1",
			CreatedBy:       "admin@example.com",
			IsVisible:       true,
			CreatedAt:       time.Now().Add(-45 * time.Minute),
			UpdatedAt:       time.Now().Add(-15 * time.Minute),
		},
		{
			ID:          uuid.New(),
			Title:       "CDN Issues Resolved",
			Description: "CDN performance issues have been fully resolved",
			Status:      "resolved",
			Impact:      "minor",
			ComponentStatus: "operational",
			StartedAt:   time.Now().Add(-24 * time.Hour),
			ResolvedAt:  func() *time.Time { t := time.Now().Add(-22 * time.Hour); return &t }(),
			TenantID:    "tenant-1",
			CreatedBy:   "admin@example.com",
			IsVisible:   true,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-22 * time.Hour),
		},
	}

	// Apply filters
	filteredIncidents := []models.SaaSIncident{}
	for _, incident := range incidents {
		if tenantID != "" && incident.TenantID != tenantID {
			continue
		}
		if status != "" && incident.Status != status {
			continue
		}
		if impact != "" && incident.Impact != impact {
			continue
		}
		filteredIncidents = append(filteredIncidents, incident)
	}

	// Calculate pagination
	total := len(filteredIncidents)
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	paginatedIncidents := filteredIncidents[offset:end]

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"incidents": paginatedIncidents,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_pages": (total + limit - 1) / limit,
			},
		},
	})
}

// GetIncident handles getting a specific incident by ID.
func (h *IncidentHandler) GetIncident(c *gin.Context) {
	h.logger.Info("Getting incident")

	idStr := c.Param("id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	// Mock incident data
	incident := models.SaaSIncident{
		ID:              incidentID,
		Title:           "API Gateway Performance Issues",
		Description:     "Users may experience slow response times when accessing our API endpoints",
		Status:          "monitoring",
		Impact:          "minor",
		ComponentStatus: "degraded_performance",
		StartedAt:       time.Now().Add(-2 * time.Hour),
		TenantID:        "tenant-1",
		CreatedBy:       "admin@example.com",
		IsVisible:       true,
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		UpdatedAt:       time.Now().Add(-30 * time.Minute),
		Updates: []models.SaaSIncidentUpdate{
			{
				ID:         uuid.New(),
				IncidentID: incidentID,
				Status:     "investigating",
				Message:    "We are investigating reports of slow API response times",
				CreatedBy:  "admin@example.com",
				IsPublic:   true,
				UpdateType: "status_update",
				CreatedAt:  time.Now().Add(-2 * time.Hour),
			},
			{
				ID:         uuid.New(),
				IncidentID: incidentID,
				Status:     "identified",
				Message:    "We have identified the issue and are working on a fix",
				CreatedBy:  "admin@example.com",
				IsPublic:   true,
				UpdateType: "status_update",
				CreatedAt:  time.Now().Add(-1 * time.Hour),
			},
			{
				ID:         uuid.New(),
				IncidentID: incidentID,
				Status:     "monitoring",
				Message:    "A fix has been implemented and we are monitoring the situation",
				CreatedBy:  "admin@example.com",
				IsPublic:   true,
				UpdateType: "status_update",
				CreatedAt:  time.Now().Add(-30 * time.Minute),
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"incident": incident,
	})
}

// CreateIncident handles creating a new incident.
func (h *IncidentHandler) CreateIncident(c *gin.Context) {
	h.logger.Info("Creating incident")

	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind incident data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident data: " + err.Error(),
		})
		return
	}

	// Validate status
	validStatuses := []string{"investigating", "identified", "monitoring", "resolved"}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Must be one of: investigating, identified, monitoring, resolved",
		})
		return
	}

	// Validate impact
	validImpacts := []string{"none", "minor", "major", "critical"}
	isValidImpact := false
	for _, impact := range validImpacts {
		if req.Impact == impact {
			isValidImpact = true
			break
		}
	}
	if !isValidImpact {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid impact. Must be one of: none, minor, major, critical",
		})
		return
	}

	// Create incident
	incident := models.SaaSIncident{
		ID:              uuid.New(),
		Title:           req.Title,
		Description:     req.Description,
		Status:          req.Status,
		Impact:          req.Impact,
		ComponentStatus: req.ComponentStatus,
		StartedAt:       time.Now(),
		TenantID:        req.TenantID,
		CreatedBy:       req.CreatedBy,
		IsVisible:       req.IsVisible,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// TODO: Save to database

	h.logger.Info("Incident created successfully", zap.String("incident_id", incident.ID.String()))

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"incident": incident,
	})
}

// UpdateIncident handles updating an existing incident.
func (h *IncidentHandler) UpdateIncident(c *gin.Context) {
	h.logger.Info("Updating incident")

	idStr := c.Param("id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	var req UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind incident update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident update data: " + err.Error(),
		})
		return
	}

	// TODO: Get incident from database and update it

	// Mock updated incident
	incident := models.SaaSIncident{
		ID:              incidentID,
		Title:           req.Title,
		Description:     req.Description,
		Status:          req.Status,
		Impact:          req.Impact,
		ComponentStatus: req.ComponentStatus,
		StartedAt:       time.Now().Add(-2 * time.Hour),
		TenantID:        "tenant-1",
		UpdatedBy:       req.UpdatedBy,
		IsVisible:       true,
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		UpdatedAt:       time.Now(),
	}

	if req.IsVisible != nil {
		incident.IsVisible = *req.IsVisible
	}

	// Set resolved time if status is resolved
	if req.Status == "resolved" {
		now := time.Now()
		incident.ResolvedAt = &now
	}

	h.logger.Info("Incident updated successfully", zap.String("incident_id", incident.ID.String()))

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"incident": incident,
	})
}

// DeleteIncident handles deleting an incident.
func (h *IncidentHandler) DeleteIncident(c *gin.Context) {
	h.logger.Info("Deleting incident")

	idStr := c.Param("id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	// TODO: Delete incident from database

	h.logger.Info("Incident deleted successfully", zap.String("incident_id", incidentID.String()))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Incident deleted successfully",
	})
}

// CreateIncidentUpdate handles creating an update for an incident.
func (h *IncidentHandler) CreateIncidentUpdate(c *gin.Context) {
	h.logger.Info("Creating incident update")

	idStr := c.Param("id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	var req CreateIncidentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind incident update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident update data: " + err.Error(),
		})
		return
	}

	// Create incident update
	update := models.SaaSIncidentUpdate{
		ID:         uuid.New(),
		IncidentID: incidentID,
		Status:     req.Status,
		Message:    req.Message,
		CreatedBy:  req.CreatedBy,
		IsPublic:   req.IsPublic,
		UpdateType: req.UpdateType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// TODO: Save to database and update incident status

	h.logger.Info("Incident update created successfully",
		zap.String("incident_id", incidentID.String()),
		zap.String("update_id", update.ID.String()))

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"update":  update,
	})
}

// GetIncidentUpdates handles getting all updates for an incident.
func (h *IncidentHandler) GetIncidentUpdates(c *gin.Context) {
	h.logger.Info("Getting incident updates")

	idStr := c.Param("id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	// Mock updates data
	updates := []models.SaaSIncidentUpdate{
		{
			ID:         uuid.New(),
			IncidentID: incidentID,
			Status:     "investigating",
			Message:    "We are investigating reports of slow API response times",
			CreatedBy:  "admin@example.com",
			IsPublic:   true,
			UpdateType: "status_update",
			CreatedAt:  time.Now().Add(-2 * time.Hour),
		},
		{
			ID:         uuid.New(),
			IncidentID: incidentID,
			Status:     "identified",
			Message:    "We have identified the issue and are working on a fix",
			CreatedBy:  "admin@example.com",
			IsPublic:   true,
			UpdateType: "status_update",
			CreatedAt:  time.Now().Add(-1 * time.Hour),
		},
		{
			ID:         uuid.New(),
			IncidentID: incidentID,
			Status:     "monitoring",
			Message:    "A fix has been implemented and we are monitoring the situation",
			CreatedBy:  "admin@example.com",
			IsPublic:   true,
			UpdateType: "status_update",
			CreatedAt:  time.Now().Add(-30 * time.Minute),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updates": updates,
	})
}

// GetIncidentStats handles getting incident statistics.
func (h *IncidentHandler) GetIncidentStats(c *gin.Context) {
	h.logger.Info("Getting incident statistics")

	tenantID := c.Query("tenant_id")
	timeRange := c.DefaultQuery("range", "30d")

	// Mock stats data
	stats := gin.H{
		"total_incidents": 45,
		"active_incidents": 2,
		"resolved_incidents": 43,
		"by_status": gin.H{
			"investigating": 1,
			"identified":    0,
			"monitoring":    1,
			"resolved":      43,
		},
		"by_impact": gin.H{
			"none":     5,
			"minor":    25,
			"major":    12,
			"critical": 3,
		},
		"avg_resolution_time": "2h 45m",
		"uptime_percentage": 99.85,
		"last_30_days": []gin.H{
			{"date": "2025-09-23", "incidents": 1, "resolved": 0},
			{"date": "2025-09-22", "incidents": 0, "resolved": 1},
			{"date": "2025-09-21", "incidents": 2, "resolved": 2},
			{"date": "2025-09-20", "incidents": 1, "resolved": 1},
			{"date": "2025-09-19", "incidents": 0, "resolved": 0},
		},
	}

	if tenantID != "" {
		h.logger.Debug("Filtering stats by tenant", zap.String("tenant_id", tenantID))
	}

	if timeRange != "30d" {
		h.logger.Debug("Custom time range requested", zap.String("range", timeRange))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}