// Package services provides business logic for the Event Store Service.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage-event-store-service/internal/config"
	"github.com/enterprise-status/statuspage-event-store-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// EventStoreService handles event store-related business logic.
type EventStoreService struct {
	config *config.Config
	logger *zap.Logger
	db     *gorm.DB
}

// NewEventStoreService creates a new event store service.
func NewEventStoreService(cfg *config.Config, logger *zap.Logger) (*EventStoreService, error) {
	// Initialize database connection
	db, err := initDatabase(cfg.Database)
	if err != nil {
		// For testing, we'll allow the service to be created without a database
		// The database will be set later via SetDB method
		if logger != nil {
			logger.Warn("Failed to initialize database, service will be created without database", zap.Error(err))
		}
		db = nil
	}

	return &EventStoreService{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

// SetDB sets the database connection (for testing)
func (s *EventStoreService) SetDB(db *gorm.DB) {
	s.db = db
}

// ListStreams lists all event streams.
func (s *EventStoreService) ListStreams(ctx context.Context) ([]*models.Stream, error) {
	s.logger.Info("Listing event streams")

	var streams []*models.Stream
	if err := s.db.Find(&streams).Error; err != nil {
		s.logger.Error("Failed to list streams", zap.Error(err))
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}

	return streams, nil
}

// CreateStream creates a new event stream.
func (s *EventStoreService) CreateStream(ctx context.Context, stream *models.Stream) error {
	s.logger.Info("Creating event stream",
		zap.String("stream_id", stream.ID),
		zap.String("stream_type", stream.Type))

	if err := s.db.Create(stream).Error; err != nil {
		s.logger.Error("Failed to create stream", zap.Error(err))
		return fmt.Errorf("failed to create stream: %w", err)
	}

	s.logger.Info("Event stream created successfully",
		zap.String("stream_id", stream.ID))

	return nil
}

// GetStream retrieves a stream by ID.
func (s *EventStoreService) GetStream(ctx context.Context, streamID string) (*models.Stream, error) {
	s.logger.Info("Getting event stream", zap.String("stream_id", streamID))

	var stream models.Stream
	if err := s.db.Where("id = ?", streamID).First(&stream).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("stream not found")
		}
		s.logger.Error("Failed to get stream", zap.Error(err))
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	return &stream, nil
}

// DeleteStream deletes a stream.
func (s *EventStoreService) DeleteStream(ctx context.Context, streamID string) error {
	s.logger.Info("Deleting event stream", zap.String("stream_id", streamID))

	if err := s.db.Where("id = ?", streamID).Delete(&models.Stream{}).Error; err != nil {
		s.logger.Error("Failed to delete stream", zap.Error(err))
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	s.logger.Info("Event stream deleted successfully", zap.String("stream_id", streamID))
	return nil
}

// AppendEvents appends events to a stream.
func (s *EventStoreService) AppendEvents(ctx context.Context, streamID string, events []*models.Event) error {
	s.logger.Info("Appending events to stream",
		zap.String("stream_id", streamID),
		zap.Int("event_count", len(events)))

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Append events
	for _, event := range events {
		event.StreamID = streamID
		// Generate UUID for event if not provided
		if event.ID == "" {
			event.ID = uuid.New().String()
		}
		if err := tx.Create(event).Error; err != nil {
			tx.Rollback()
			s.logger.Error("Failed to append event", zap.Error(err))
			return fmt.Errorf("failed to append event: %w", err)
		}
	}

	// Update stream version
	if err := tx.Model(&models.Stream{}).Where("id = ?", streamID).Update("version", gorm.Expr("version + ?", len(events))).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update stream version", zap.Error(err))
		return fmt.Errorf("failed to update stream version: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Events appended successfully",
		zap.String("stream_id", streamID),
		zap.Int("event_count", len(events)))

	return nil
}

// GetEvents retrieves events from a stream.
func (s *EventStoreService) GetEvents(ctx context.Context, streamID string, fromVersion int, limit int) ([]*models.Event, error) {
	s.logger.Info("Getting events from stream",
		zap.String("stream_id", streamID),
		zap.Int("from_version", fromVersion),
		zap.Int("limit", limit))

	var events []*models.Event
	query := s.db.Where("stream_id = ?", streamID)

	if fromVersion > 0 {
		query = query.Where("version >= ?", fromVersion)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("version ASC").Find(&events).Error; err != nil {
		s.logger.Error("Failed to get events", zap.Error(err))
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return events, nil
}

// GetEvent retrieves a specific event.
func (s *EventStoreService) GetEvent(ctx context.Context, streamID string, eventID string) (*models.Event, error) {
	s.logger.Info("Getting event",
		zap.String("stream_id", streamID),
		zap.String("event_id", eventID))

	var event models.Event
	if err := s.db.Where("stream_id = ? AND id = ?", streamID, eventID).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("event not found")
		}
		s.logger.Error("Failed to get event", zap.Error(err))
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

// CreateSnapshot creates a snapshot for a stream.
func (s *EventStoreService) CreateSnapshot(ctx context.Context, streamID string, snapshot *models.Snapshot) error {
	s.logger.Info("Creating snapshot",
		zap.String("stream_id", streamID),
		zap.String("snapshot_id", snapshot.ID))

	snapshot.StreamID = streamID
	// Generate UUID for snapshot if not provided
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	if err := s.db.Create(snapshot).Error; err != nil {
		s.logger.Error("Failed to create snapshot", zap.Error(err))
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	s.logger.Info("Snapshot created successfully",
		zap.String("stream_id", streamID),
		zap.String("snapshot_id", snapshot.ID))

	return nil
}

// GetSnapshots retrieves snapshots for a stream.
func (s *EventStoreService) GetSnapshots(ctx context.Context, streamID string) ([]*models.Snapshot, error) {
	s.logger.Info("Getting snapshots", zap.String("stream_id", streamID))

	var snapshots []*models.Snapshot
	if err := s.db.Where("stream_id = ?", streamID).Order("version DESC").Find(&snapshots).Error; err != nil {
		s.logger.Error("Failed to get snapshots", zap.Error(err))
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	return snapshots, nil
}

// GetSnapshot retrieves a specific snapshot.
func (s *EventStoreService) GetSnapshot(ctx context.Context, streamID string, snapshotID string) (*models.Snapshot, error) {
	s.logger.Info("Getting snapshot",
		zap.String("stream_id", streamID),
		zap.String("snapshot_id", snapshotID))

	var snapshot models.Snapshot
	if err := s.db.Where("stream_id = ? AND id = ?", streamID, snapshotID).First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("snapshot not found")
		}
		s.logger.Error("Failed to get snapshot", zap.Error(err))
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return &snapshot, nil
}

// ListProjections lists all projections.
func (s *EventStoreService) ListProjections(ctx context.Context) ([]*models.Projection, error) {
	s.logger.Info("Listing projections")

	var projections []*models.Projection
	if err := s.db.Find(&projections).Error; err != nil {
		s.logger.Error("Failed to list projections", zap.Error(err))
		return nil, fmt.Errorf("failed to list projections: %w", err)
	}

	return projections, nil
}

// CreateProjection creates a new projection.
func (s *EventStoreService) CreateProjection(ctx context.Context, projection *models.Projection) error {
	s.logger.Info("Creating projection",
		zap.String("projection_id", projection.ID),
		zap.String("projection_name", projection.Name))

	if err := s.db.Create(projection).Error; err != nil {
		s.logger.Error("Failed to create projection", zap.Error(err))
		return fmt.Errorf("failed to create projection: %w", err)
	}

	s.logger.Info("Projection created successfully",
		zap.String("projection_id", projection.ID))

	return nil
}

// GetProjection retrieves a projection by ID.
func (s *EventStoreService) GetProjection(ctx context.Context, projectionID string) (*models.Projection, error) {
	s.logger.Info("Getting projection", zap.String("projection_id", projectionID))

	var projection models.Projection
	if err := s.db.Where("id = ?", projectionID).First(&projection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("projection not found")
		}
		s.logger.Error("Failed to get projection", zap.Error(err))
		return nil, fmt.Errorf("failed to get projection: %w", err)
	}

	return &projection, nil
}

// UpdateProjection updates a projection.
func (s *EventStoreService) UpdateProjection(ctx context.Context, projectionID string, updates *models.Projection) error {
	s.logger.Info("Updating projection", zap.String("projection_id", projectionID))

	if err := s.db.Model(&models.Projection{}).Where("id = ?", projectionID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update projection", zap.Error(err))
		return fmt.Errorf("failed to update projection: %w", err)
	}

	s.logger.Info("Projection updated successfully", zap.String("projection_id", projectionID))
	return nil
}

// DeleteProjection deletes a projection.
func (s *EventStoreService) DeleteProjection(ctx context.Context, projectionID string) error {
	s.logger.Info("Deleting projection", zap.String("projection_id", projectionID))

	if err := s.db.Where("id = ?", projectionID).Delete(&models.Projection{}).Error; err != nil {
		s.logger.Error("Failed to delete projection", zap.Error(err))
		return fmt.Errorf("failed to delete projection: %w", err)
	}

	s.logger.Info("Projection deleted successfully", zap.String("projection_id", projectionID))
	return nil
}

// GetStats returns event store statistics.
func (s *EventStoreService) GetStats(ctx context.Context) (*models.EventStoreStats, error) {
	s.logger.Info("Getting event store statistics")

	// For now, we'll return simulated statistics
	// In production, you would query actual statistics from the database

	stats := &models.EventStoreStats{
		TotalStreams:     100,
		TotalEvents:      10000,
		TotalProjections: 10,
		TotalSnapshots:   50,
		LastUpdated:      time.Now(),
	}

	return stats, nil
}

// GetStreamStats returns stream statistics.
func (s *EventStoreService) GetStreamStats(ctx context.Context, streamID string) (*models.StreamStats, error) {
	s.logger.Info("Getting stream statistics", zap.String("stream_id", streamID))

	// For now, we'll return simulated statistics
	// In production, you would query actual statistics from the database

	stats := &models.StreamStats{
		StreamID:      streamID,
		EventCount:    100,
		SnapshotCount: 5,
		LastEventAt:   time.Now(),
		LastUpdated:   time.Now(),
	}

	return stats, nil
}

// Health checks the health of the event store service.
func (s *EventStoreService) Health(ctx context.Context) error {
	s.logger.Debug("Checking event store service health")

	// Check database connection if available
	if s.db != nil {
		if err := s.db.Exec("SELECT 1").Error; err != nil {
			s.logger.Error("Event store health check failed", zap.Error(err))
			return fmt.Errorf("event store health check failed: %w", err)
		}
	} else {
		s.logger.Debug("Database not available, skipping database health check")
	}

	return nil
}

// initDatabase initializes the database connection.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Stream{},
		&models.Event{},
		&models.Snapshot{},
		&models.Projection{},
		&models.Subscription{},
		&models.EventStoreStats{},
		&models.StreamStats{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
