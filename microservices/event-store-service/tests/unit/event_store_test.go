// Package unit provides unit tests for the Event Store Service.
package unit

import (
	"context"
	"testing"
	"time"

	"github.com/enterprise-status/statuspage-event-store-service/internal/config"
	"github.com/enterprise-status/statuspage-event-store-service/internal/models"
	"github.com/enterprise-status/statuspage-event-store-service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEventStoreService_ListStreams(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	eventStoreService, _ := services.NewEventStoreService(cfg, logger)
	// Override the database connection with our test database
	eventStoreService.SetDB(db)

	// Test
	streams, err := eventStoreService.ListStreams(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, streams)
	assert.IsType(t, []*models.Stream{}, streams)
}

func TestEventStoreService_CreateStream(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
		},
	}
	eventStoreService, _ := services.NewEventStoreService(cfg, logger)
	// Override the database connection with our test database
	eventStoreService.SetDB(db)

	// Test data
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}

	// Test
	err := eventStoreService.CreateStream(context.Background(), stream)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, stream.ID)
}

func TestEventStoreService_GetStream(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	eventStoreService, _ := services.NewEventStoreService(cfg, nil)

	// Create a stream first
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}
	err := eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Test
	retrievedStream, err := eventStoreService.GetStream(context.Background(), stream.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, retrievedStream)
	assert.Equal(t, stream.ID, retrievedStream.ID)
	assert.Equal(t, stream.Type, retrievedStream.Type)
	assert.Equal(t, stream.Version, retrievedStream.Version)
}

func TestEventStoreService_AppendEvent(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	eventStoreService, _ := services.NewEventStoreService(cfg, nil)

	// Create a stream first
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}
	err := eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Test data
	event := &models.Event{
		StreamID:  stream.ID,
		Type:      "UserCreated",
		Version:   1,
		Data:      `{"name": "John Doe", "email": "john@example.com"}`,
		Metadata:  `{"source": "test"}`,
		Timestamp: time.Now(),
	}

	// Test
	err = eventStoreService.AppendEvents(context.Background(), event.StreamID, []*models.Event{event})

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID)
}

func TestEventStoreService_GetEvents(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	eventStoreService, _ := services.NewEventStoreService(cfg, nil)

	// Create a stream first
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}
	err := eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Create multiple events
	events := []*models.Event{
		{
			StreamID:  stream.ID,
			Type:      "UserCreated",
			Version:   1,
			Data:      `{"name": "John Doe", "email": "john@example.com"}`,
			Metadata:  `{"source": "test"}`,
			Timestamp: time.Now(),
		},
		{
			StreamID:  stream.ID,
			Type:      "UserUpdated",
			Version:   2,
			Data:      `{"name": "John Smith", "email": "john.smith@example.com"}`,
			Metadata:  `{"source": "test"}`,
			Timestamp: time.Now(),
		},
	}

	err = eventStoreService.AppendEvents(context.Background(), stream.ID, events)
	require.NoError(t, err)

	// Test
	retrievedEvents, err := eventStoreService.GetEvents(context.Background(), stream.ID, 0, 10)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, retrievedEvents)
	assert.Len(t, retrievedEvents, 2)
	assert.Equal(t, "UserCreated", retrievedEvents[0].Type)
	assert.Equal(t, "UserUpdated", retrievedEvents[1].Type)
}

func TestEventStoreService_CreateSnapshot(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	eventStoreService, _ := services.NewEventStoreService(cfg, nil)

	// Create a stream first
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}
	err := eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Test data
	snapshot := &models.Snapshot{
		StreamID:  stream.ID,
		Version:   10,
		Data:      `{"name": "John Doe", "email": "john@example.com", "version": 10}`,
		Metadata:  `{"source": "test"}`,
		Timestamp: time.Now(),
	}

	// Test
	err = eventStoreService.CreateSnapshot(context.Background(), snapshot.StreamID, snapshot)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, snapshot.ID)
}

func TestEventStoreService_GetSnapshot(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	eventStoreService, _ := services.NewEventStoreService(cfg, nil)

	// Create a stream first
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}
	err := eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Create a snapshot
	snapshot := &models.Snapshot{
		StreamID:  stream.ID,
		Version:   10,
		Data:      `{"name": "John Doe", "email": "john@example.com", "version": 10}`,
		Metadata:  `{"source": "test"}`,
		Timestamp: time.Now(),
	}
	err = eventStoreService.CreateSnapshot(context.Background(), snapshot.StreamID, snapshot)
	require.NoError(t, err)

	// Test
	retrievedSnapshot, err := eventStoreService.GetSnapshot(context.Background(), stream.ID, "10")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSnapshot)
	assert.Equal(t, snapshot.StreamID, retrievedSnapshot.StreamID)
	assert.Equal(t, snapshot.Version, retrievedSnapshot.Version)
	assert.Equal(t, snapshot.Data, retrievedSnapshot.Data)
}

func TestEventStoreService_DeleteStream(t *testing.T) {
	// Setup
	_ = setupTestDB(t)
	cfg := &config.Config{}
	eventStoreService, _ := services.NewEventStoreService(cfg, nil)

	// Create a stream first
	stream := &models.Stream{
		ID:       "test-stream-1",
		Type:     "User",
		Version:  0,
		Metadata: `{"source": "test"}`,
	}
	err := eventStoreService.CreateStream(context.Background(), stream)
	require.NoError(t, err)

	// Test
	err = eventStoreService.DeleteStream(context.Background(), stream.ID)

	// Assertions
	assert.NoError(t, err)

	// Verify deletion
	_, err = eventStoreService.GetStream(context.Background(), stream.ID)
	assert.Error(t, err)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Stream{},
		&models.Event{},
		&models.Snapshot{},
		&models.Projection{},
	)
	require.NoError(t, err)

	return db
}
