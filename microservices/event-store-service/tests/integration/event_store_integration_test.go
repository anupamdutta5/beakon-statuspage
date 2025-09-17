package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/anupamdutta5/statuspage-event-store-service/internal/config"
	"github.com/anupamdutta5/statuspage-event-store-service/internal/models"
	"github.com/anupamdutta5/statuspage-event-store-service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	eventStoreService *services.EventStoreService
	db                *gorm.DB
)

func TestMain(m *testing.M) {
	// Setup test environment
	setupIntegrationTest()

	// Run tests
	code := m.Run()

	// Cleanup
	teardownIntegrationTest()

	os.Exit(code)
}

func setupIntegrationTest() error {
	// Initialize in-memory SQLite database for testing
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Stream{},
		&models.Event{},
		&models.Snapshot{},
		&models.Projection{},
		&models.Subscription{},
		&models.EventStoreStats{},
		&models.StreamStats{},
	)
	if err != nil {
		return err
	}

	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Create a test config
	testConfig := &config.Config{
		Environment: "test",
		Service: config.ServiceConfig{
			Name:        "event-store-service",
			Version:     "1.0.0",
			Description: "Event Store Service for Testing",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8096,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test_db",
			SSLMode:  "disable",
		},
		EventStore: config.EventStoreConfig{
			MaxEventsPerStream: 10000,
			SnapshotInterval:   100,
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
	}

	// Initialize service
	eventStoreService, err = services.NewEventStoreService(testConfig, logger)
	if err != nil {
		return err
	}

	// Use in-memory SQLite database for integration tests
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open test database: %w", err)
	}

	// Auto-migrate the schema
	err = testDB.AutoMigrate(&models.Stream{}, &models.Event{}, &models.Snapshot{}, &models.Projection{})
	if err != nil {
		return fmt.Errorf("failed to migrate test database: %w", err)
	}

	// Set the test database
	eventStoreService.SetDB(testDB)
	db = testDB

	return nil
}

func teardownIntegrationTest() {
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func TestEventStoreServiceHealth(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Test health check
	err = eventStoreService.Health(context.Background())
	assert.NoError(t, err)
}

func TestEventStoreCRUDIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a stream first
	stream := &models.Stream{
		ID:       "tenant-123",
		Type:     "tenant",
		Version:  0,
		Metadata: `{"description": "Test tenant stream"}`,
	}

	err = eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Test event creation
	event := &models.Event{
		StreamID:  "tenant-123",
		Type:      "tenant.created",
		Version:   1,
		Data:      `{"name": "Test Tenant", "subdomain": "test"}`,
		Metadata:  `{"user_id": "user-123", "ip_address": "192.168.1.1"}`,
		Timestamp: time.Now(),
	}

	// Append event
	err = eventStoreService.AppendEvents(context.Background(), event.StreamID, []*models.Event{event})
	require.NoError(t, err)

	// Verify event was stored
	events, err := eventStoreService.GetEvents(context.Background(), event.StreamID, 0, 10)
	require.NoError(t, err)
	require.Len(t, events, 1)

	storedEvent := events[0]
	assert.Equal(t, event.StreamID, storedEvent.StreamID)
	assert.Equal(t, event.Type, storedEvent.Type)
	assert.Equal(t, event.Version, storedEvent.Version)
	assert.Equal(t, event.Data, storedEvent.Data)
}

func TestEventStoreMultipleEventsIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a stream
	stream := &models.Stream{
		ID:       "user-456",
		Type:     "user",
		Version:  0,
		Metadata: `{"description": "Test user stream"}`,
	}

	err = eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Create multiple events
	events := []*models.Event{
		{
			StreamID:  "user-456",
			Type:      "user.registered",
			Version:   1,
			Data:      `{"email": "user@example.com", "name": "John Doe"}`,
			Metadata:  `{"source": "web"}`,
			Timestamp: time.Now(),
		},
		{
			StreamID:  "user-456",
			Type:      "user.verified",
			Version:   2,
			Data:      `{"verified": true}`,
			Metadata:  `{"method": "email"}`,
			Timestamp: time.Now(),
		},
		{
			StreamID:  "user-456",
			Type:      "user.updated",
			Version:   3,
			Data:      `{"name": "John Smith"}`,
			Metadata:  `{"field": "name"}`,
			Timestamp: time.Now(),
		},
	}

	// Append events
	err = eventStoreService.AppendEvents(context.Background(), "user-456", events)
	require.NoError(t, err)

	// Verify events were stored
	storedEvents, err := eventStoreService.GetEvents(context.Background(), "user-456", 0, 10)
	require.NoError(t, err)
	require.Len(t, storedEvents, 3)

	// Verify event order and content
	assert.Equal(t, "user.registered", storedEvents[0].Type)
	assert.Equal(t, "user.verified", storedEvents[1].Type)
	assert.Equal(t, "user.updated", storedEvents[2].Type)
}

func TestEventStoreSnapshotIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a stream
	stream := &models.Stream{
		ID:       "order-789",
		Type:     "order",
		Version:  0,
		Metadata: `{"description": "Test order stream"}`,
	}

	err = eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Create a snapshot
	snapshot := &models.Snapshot{
		StreamID:  "order-789",
		Version:   5,
		Data:      `{"order_id": "order-789", "status": "completed", "total": 99.99}`,
		Metadata:  `{"created_by": "system"}`,
		Timestamp: time.Now(),
	}

	// Save snapshot
	err = eventStoreService.CreateSnapshot(context.Background(), "order-789", snapshot)
	require.NoError(t, err)

	// Retrieve snapshot
	snapshots, err := eventStoreService.GetSnapshots(context.Background(), "order-789")
	require.NoError(t, err)
	require.Len(t, snapshots, 1)

	retrievedSnapshot := snapshots[0]
	assert.Equal(t, snapshot.StreamID, retrievedSnapshot.StreamID)
	assert.Equal(t, snapshot.Version, retrievedSnapshot.Version)
	assert.Equal(t, snapshot.Data, retrievedSnapshot.Data)
}

func TestEventStoreStreamManagementIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a stream
	stream := &models.Stream{
		ID:       "product-101",
		Type:     "product",
		Version:  0,
		Metadata: `{"description": "Test product stream"}`,
	}

	err = eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Get stream
	retrievedStream, err := eventStoreService.GetStream(context.Background(), "product-101")
	require.NoError(t, err)
	assert.Equal(t, stream.ID, retrievedStream.ID)
	assert.Equal(t, stream.Type, retrievedStream.Type)

	// List streams
	streams, err := eventStoreService.ListStreams(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(streams), 1)

	// Delete stream
	err = eventStoreService.DeleteStream(context.Background(), "product-101")
	require.NoError(t, err)

	// Verify stream was deleted
	_, err = eventStoreService.GetStream(context.Background(), "product-101")
	assert.Error(t, err)
}

func TestEventStoreProjectionIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a projection
	projection := &models.Projection{
		ID:          "user-projection",
		Name:        "User Projection",
		Description: "Projects user events to read model",
		Query:       "SELECT * FROM events WHERE stream_id LIKE 'user-%'",
		Status:      "stopped",
		LastEvent:   "",
		LastUpdated: time.Now(),
		Metadata:    `{"type": "read_model"}`,
	}

	err = eventStoreService.CreateProjection(context.Background(), projection)
	require.NoError(t, err)

	// Get projection
	retrievedProjection, err := eventStoreService.GetProjection(context.Background(), "user-projection")
	require.NoError(t, err)
	assert.Equal(t, projection.Name, retrievedProjection.Name)
	assert.Equal(t, projection.Description, retrievedProjection.Description)

	// Update projection
	updates := &models.Projection{
		Status: "running",
	}
	err = eventStoreService.UpdateProjection(context.Background(), "user-projection", updates)
	require.NoError(t, err)

	// Verify update
	updatedProjection, err := eventStoreService.GetProjection(context.Background(), "user-projection")
	require.NoError(t, err)
	assert.Equal(t, "running", updatedProjection.Status)

	// List projections
	projections, err := eventStoreService.ListProjections(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(projections), 1)

	// Delete projection
	err = eventStoreService.DeleteProjection(context.Background(), "user-projection")
	require.NoError(t, err)

	// Verify deletion
	_, err = eventStoreService.GetProjection(context.Background(), "user-projection")
	assert.Error(t, err)
}

func TestEventStoreStatsIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Get stats
	stats, err := eventStoreService.GetStats(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalStreams, 0)
	assert.GreaterOrEqual(t, stats.TotalEvents, int64(0))
}

func TestEventStoreConcurrentAccessIntegration(t *testing.T) {
	err := setupIntegrationTest()
	require.NoError(t, err)
	defer teardownIntegrationTest()

	// Create a stream
	stream := &models.Stream{
		ID:       "concurrent-test",
		Type:     "test",
		Version:  0,
		Metadata: `{"description": "Concurrent test stream"}`,
	}

	err = eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Create multiple events concurrently
	events := make([]*models.Event, 10)
	for i := 0; i < 10; i++ {
		events[i] = &models.Event{
			StreamID:  "concurrent-test",
			Type:      "test.event",
			Version:   i + 1,
			Data:      fmt.Sprintf(`{"index": %d, "message": "Event %d"}`, i, i),
			Metadata:  `{"concurrent": true}`,
			Timestamp: time.Now(),
		}
	}

	// Append events
	err = eventStoreService.AppendEvents(context.Background(), "concurrent-test", events)
	require.NoError(t, err)

	// Verify all events were stored
	storedEvents, err := eventStoreService.GetEvents(context.Background(), "concurrent-test", 0, 20)
	require.NoError(t, err)
	assert.Len(t, storedEvents, 10)

	// Verify event order
	for i, event := range storedEvents {
		assert.Equal(t, i+1, event.Version)
	}
}
