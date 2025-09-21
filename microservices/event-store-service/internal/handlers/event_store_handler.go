// Package handlers provides HTTP handlers for the Event Store Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/anupamdutta5/event-store-service/internal/models"
	"github.com/anupamdutta5/event-store-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EventStoreHandler handles event store-related HTTP requests.
type EventStoreHandler struct {
	service *services.EventStoreService
	logger  *zap.Logger
}

// NewEventStoreHandler creates a new event store handler.
func NewEventStoreHandler(service *services.EventStoreService, logger *zap.Logger) *EventStoreHandler {
	return &EventStoreHandler{
		service: service,
		logger:  logger,
	}
}

// HealthCheck handles health check requests.
func (h *EventStoreHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "event-store-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "event-store-service",
		"version": "1.0.0",
	})
}

// ListStreams handles listing event streams.
func (h *EventStoreHandler) ListStreams(c *gin.Context) {
	h.logger.Info("Listing event streams")

	streams, err := h.service.ListStreams(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list streams", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list streams",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"streams": streams,
		"count":   len(streams),
	})
}

// CreateStream handles creating a new event stream.
func (h *EventStoreHandler) CreateStream(c *gin.Context) {
	var stream models.Stream
	if err := c.ShouldBindJSON(&stream); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating event stream", zap.String("stream_id", stream.ID))

	if err := h.service.CreateStream(c.Request.Context(), &stream); err != nil {
		h.logger.Error("Failed to create stream", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create stream",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"stream":  stream,
		"message": "Event stream created successfully",
	})
}

// GetStream handles retrieving a stream by ID.
func (h *EventStoreHandler) GetStream(c *gin.Context) {
	streamID := c.Param("id")

	h.logger.Info("Getting event stream", zap.String("stream_id", streamID))

	stream, err := h.service.GetStream(c.Request.Context(), streamID)
	if err != nil {
		h.logger.Error("Failed to get stream", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Stream not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream": stream,
	})
}

// DeleteStream handles deleting a stream.
func (h *EventStoreHandler) DeleteStream(c *gin.Context) {
	streamID := c.Param("id")

	h.logger.Info("Deleting event stream", zap.String("stream_id", streamID))

	if err := h.service.DeleteStream(c.Request.Context(), streamID); err != nil {
		h.logger.Error("Failed to delete stream", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete stream",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event stream deleted successfully",
	})
}

// AppendEvents handles appending events to a stream.
func (h *EventStoreHandler) AppendEvents(c *gin.Context) {
	streamID := c.Param("id")

	var request struct {
		Events []*models.Event `json:"events" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Appending events to stream",
		zap.String("stream_id", streamID),
		zap.Int("event_count", len(request.Events)))

	if err := h.service.AppendEvents(c.Request.Context(), streamID, request.Events); err != nil {
		h.logger.Error("Failed to append events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to append events",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Events appended successfully",
		"event_count": len(request.Events),
	})
}

// GetEvents handles retrieving events from a stream.
func (h *EventStoreHandler) GetEvents(c *gin.Context) {
	streamID := c.Param("id")

	// Parse query parameters
	fromVersionStr := c.DefaultQuery("from_version", "0")
	fromVersion, err := strconv.Atoi(fromVersionStr)
	if err != nil {
		fromVersion = 0
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	h.logger.Info("Getting events from stream",
		zap.String("stream_id", streamID),
		zap.Int("from_version", fromVersion),
		zap.Int("limit", limit))

	events, err := h.service.GetEvents(c.Request.Context(), streamID, fromVersion, limit)
	if err != nil {
		h.logger.Error("Failed to get events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// GetEvent handles retrieving a specific event.
func (h *EventStoreHandler) GetEvent(c *gin.Context) {
	streamID := c.Param("id")
	eventID := c.Param("event_id")

	h.logger.Info("Getting event",
		zap.String("stream_id", streamID),
		zap.String("event_id", eventID))

	event, err := h.service.GetEvent(c.Request.Context(), streamID, eventID)
	if err != nil {
		h.logger.Error("Failed to get event", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event": event,
	})
}

// CreateSnapshot handles creating a snapshot for a stream.
func (h *EventStoreHandler) CreateSnapshot(c *gin.Context) {
	streamID := c.Param("id")

	var snapshot models.Snapshot
	if err := c.ShouldBindJSON(&snapshot); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating snapshot",
		zap.String("stream_id", streamID),
		zap.String("snapshot_id", snapshot.ID))

	if err := h.service.CreateSnapshot(c.Request.Context(), streamID, &snapshot); err != nil {
		h.logger.Error("Failed to create snapshot", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create snapshot",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"snapshot": snapshot,
		"message":  "Snapshot created successfully",
	})
}

// GetSnapshots handles retrieving snapshots for a stream.
func (h *EventStoreHandler) GetSnapshots(c *gin.Context) {
	streamID := c.Param("id")

	h.logger.Info("Getting snapshots", zap.String("stream_id", streamID))

	snapshots, err := h.service.GetSnapshots(c.Request.Context(), streamID)
	if err != nil {
		h.logger.Error("Failed to get snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get snapshots",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshots": snapshots,
		"count":     len(snapshots),
	})
}

// GetSnapshot handles retrieving a specific snapshot.
func (h *EventStoreHandler) GetSnapshot(c *gin.Context) {
	streamID := c.Param("id")
	snapshotID := c.Param("snapshot_id")

	h.logger.Info("Getting snapshot",
		zap.String("stream_id", streamID),
		zap.String("snapshot_id", snapshotID))

	snapshot, err := h.service.GetSnapshot(c.Request.Context(), streamID, snapshotID)
	if err != nil {
		h.logger.Error("Failed to get snapshot", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Snapshot not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snapshot": snapshot,
	})
}

// ListProjections handles listing projections.
func (h *EventStoreHandler) ListProjections(c *gin.Context) {
	h.logger.Info("Listing projections")

	projections, err := h.service.ListProjections(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list projections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list projections",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projections": projections,
		"count":       len(projections),
	})
}

// CreateProjection handles creating a new projection.
func (h *EventStoreHandler) CreateProjection(c *gin.Context) {
	var projection models.Projection
	if err := c.ShouldBindJSON(&projection); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating projection", zap.String("projection_id", projection.ID))

	if err := h.service.CreateProjection(c.Request.Context(), &projection); err != nil {
		h.logger.Error("Failed to create projection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create projection",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"projection": projection,
		"message":    "Projection created successfully",
	})
}

// GetProjection handles retrieving a projection by ID.
func (h *EventStoreHandler) GetProjection(c *gin.Context) {
	projectionID := c.Param("id")

	h.logger.Info("Getting projection", zap.String("projection_id", projectionID))

	projection, err := h.service.GetProjection(c.Request.Context(), projectionID)
	if err != nil {
		h.logger.Error("Failed to get projection", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Projection not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projection": projection,
	})
}

// UpdateProjection handles updating a projection.
func (h *EventStoreHandler) UpdateProjection(c *gin.Context) {
	projectionID := c.Param("id")

	var updates models.Projection
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating projection", zap.String("projection_id", projectionID))

	if err := h.service.UpdateProjection(c.Request.Context(), projectionID, &updates); err != nil {
		h.logger.Error("Failed to update projection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update projection",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Projection updated successfully",
	})
}

// DeleteProjection handles deleting a projection.
func (h *EventStoreHandler) DeleteProjection(c *gin.Context) {
	projectionID := c.Param("id")

	h.logger.Info("Deleting projection", zap.String("projection_id", projectionID))

	if err := h.service.DeleteProjection(c.Request.Context(), projectionID); err != nil {
		h.logger.Error("Failed to delete projection", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete projection",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Projection deleted successfully",
	})
}

// GetStats handles getting event store statistics.
func (h *EventStoreHandler) GetStats(c *gin.Context) {
	h.logger.Info("Getting event store statistics")

	stats, err := h.service.GetStats(c.Request.Context())
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

// GetStreamStats handles getting stream statistics.
func (h *EventStoreHandler) GetStreamStats(c *gin.Context) {
	streamID := c.Param("id")

	h.logger.Info("Getting stream statistics", zap.String("stream_id", streamID))

	stats, err := h.service.GetStreamStats(c.Request.Context(), streamID)
	if err != nil {
		h.logger.Error("Failed to get stream statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get stream statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Placeholder handlers for remaining endpoints
func (h *EventStoreHandler) StartProjection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) StopProjection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) ResetProjection(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) ListSubscriptions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) CreateSubscription(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) GetSubscription(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) UpdateSubscription(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *EventStoreHandler) DeleteSubscription(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

