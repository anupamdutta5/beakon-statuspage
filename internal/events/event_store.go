package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Event represents a domain event
type Event struct {
	ID            int64                  `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     string                 `json:"event_type"`
	EventData     map[string]interface{} `json:"event_data"`
	EventMetadata map[string]interface{} `json:"event_metadata"`
	Version       int                    `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
}

// Snapshot represents an aggregate snapshot
type Snapshot struct {
	ID            int64                  `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	SnapshotData  map[string]interface{} `json:"snapshot_data"`
	Version       int                    `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
}

// EventStore handles event sourcing operations
type EventStore struct {
	db *sql.DB
}

// NewEventStore creates a new event store
func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{
		db: db,
	}
}

// SaveEvent saves an event to the event store
func (es *EventStore) SaveEvent(ctx context.Context, event *Event) error {
	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	eventMetadataJSON, err := json.Marshal(event.EventMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	query := `
		INSERT INTO events (
			aggregate_id, aggregate_type, event_type, 
			event_data, event_metadata, version
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := es.db.ExecContext(ctx, query,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		eventDataJSON,
		eventMetadataJSON,
		event.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get event ID: %w", err)
	}

	event.ID = eventID
	event.CreatedAt = time.Now()

	logger.Log.Debug("Event saved",
		zap.String("aggregate_id", event.AggregateID),
		zap.String("aggregate_type", event.AggregateType),
		zap.String("event_type", event.EventType),
		zap.Int64("event_id", eventID))

	return nil
}

// GetEvents retrieves events for an aggregate
func (es *EventStore) GetEvents(ctx context.Context, aggregateID, aggregateType string, fromVersion int) ([]*Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, 
		       event_data, event_metadata, version, created_at
		FROM events
		WHERE aggregate_id = ? AND aggregate_type = ? AND version > ?
		ORDER BY version ASC
	`

	rows, err := es.db.QueryContext(ctx, query, aggregateID, aggregateType, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		var eventDataJSON, eventMetadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&eventDataJSON,
			&eventMetadataJSON,
			&event.Version,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Unmarshal JSON data
		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		if err := json.Unmarshal(eventMetadataJSON, &event.EventMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// SaveSnapshot saves an aggregate snapshot
func (es *EventStore) SaveSnapshot(ctx context.Context, snapshot *Snapshot) error {
	snapshotDataJSON, err := json.Marshal(snapshot.SnapshotData)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	query := `
		INSERT INTO snapshots (aggregate_id, aggregate_type, snapshot_data, version)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			snapshot_data = VALUES(snapshot_data),
			version = VALUES(version),
			created_at = CURRENT_TIMESTAMP
	`

	_, err = es.db.ExecContext(ctx, query,
		snapshot.AggregateID,
		snapshot.AggregateType,
		snapshotDataJSON,
		snapshot.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	logger.Log.Debug("Snapshot saved",
		zap.String("aggregate_id", snapshot.AggregateID),
		zap.String("aggregate_type", snapshot.AggregateType),
		zap.Int("version", snapshot.Version))

	return nil
}

// GetLatestSnapshot retrieves the latest snapshot for an aggregate
func (es *EventStore) GetLatestSnapshot(ctx context.Context, aggregateID, aggregateType string) (*Snapshot, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, snapshot_data, version, created_at
		FROM snapshots
		WHERE aggregate_id = ? AND aggregate_type = ?
		ORDER BY version DESC
		LIMIT 1
	`

	row := es.db.QueryRowContext(ctx, query, aggregateID, aggregateType)

	snapshot := &Snapshot{}
	var snapshotDataJSON []byte

	err := row.Scan(
		&snapshot.ID,
		&snapshot.AggregateID,
		&snapshot.AggregateType,
		&snapshotDataJSON,
		&snapshot.Version,
		&snapshot.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No snapshot found
		}
		return nil, fmt.Errorf("failed to scan snapshot: %w", err)
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(snapshotDataJSON, &snapshot.SnapshotData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return snapshot, nil
}

// GetEventsByType retrieves events by type
func (es *EventStore) GetEventsByType(ctx context.Context, eventType string, limit int) ([]*Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, 
		       event_data, event_metadata, version, created_at
		FROM events
		WHERE event_type = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := es.db.QueryContext(ctx, query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		var eventDataJSON, eventMetadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&eventDataJSON,
			&eventMetadataJSON,
			&event.Version,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Unmarshal JSON data
		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		if err := json.Unmarshal(eventMetadataJSON, &event.EventMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventCount returns the total number of events
func (es *EventStore) GetEventCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM events`

	var count int64
	err := es.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get event count: %w", err)
	}

	return count, nil
}

// GetEventCountByType returns the number of events by type
func (es *EventStore) GetEventCountByType(ctx context.Context, eventType string) (int64, error) {
	query := `SELECT COUNT(*) FROM events WHERE event_type = ?`

	var count int64
	err := es.db.QueryRowContext(ctx, query, eventType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get event count by type: %w", err)
	}

	return count, nil
}

// GetEventCountByAggregate returns the number of events by aggregate
func (es *EventStore) GetEventCountByAggregate(ctx context.Context, aggregateID, aggregateType string) (int64, error) {
	query := `SELECT COUNT(*) FROM events WHERE aggregate_id = ? AND aggregate_type = ?`

	var count int64
	err := es.db.QueryRowContext(ctx, query, aggregateID, aggregateType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get event count by aggregate: %w", err)
	}

	return count, nil
}

// GetEventSummary returns a summary of events
func (es *EventStore) GetEventSummary(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			aggregate_type,
			event_type,
			COUNT(*) as event_count,
			MIN(created_at) as first_event,
			MAX(created_at) as last_event
		FROM events
		GROUP BY aggregate_type, event_type
		ORDER BY event_count DESC
	`

	rows, err := es.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query event summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[string]interface{})
	var eventTypes []map[string]interface{}

	for rows.Next() {
		var aggregateType, eventType string
		var eventCount int64
		var firstEvent, lastEvent time.Time

		err := rows.Scan(&aggregateType, &eventType, &eventCount, &firstEvent, &lastEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event summary: %w", err)
		}

		eventTypes = append(eventTypes, map[string]interface{}{
			"aggregate_type": aggregateType,
			"event_type":     eventType,
			"event_count":    eventCount,
			"first_event":    firstEvent,
			"last_event":     lastEvent,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event summary: %w", err)
	}

	summary["event_types"] = eventTypes

	// Get total event count
	totalCount, err := es.GetEventCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total event count: %w", err)
	}

	summary["total_events"] = totalCount

	return summary, nil
}

// Close closes the event store connection
func (es *EventStore) Close() error {
	if es.db != nil {
		return es.db.Close()
	}
	return nil
}
