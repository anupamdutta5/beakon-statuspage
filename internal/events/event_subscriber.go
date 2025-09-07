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

// EventSubscriber handles subscribing to and processing events
type EventSubscriber struct {
	eventStore *EventStore
	db         *sql.DB
	handlers   map[string]EventHandler
}

// EventSubscription represents a subscription to events
type EventSubscription struct {
	ID                   int64     `json:"id"`
	SubscriptionName     string    `json:"subscription_name"`
	EventType            string    `json:"event_type"`
	HandlerURL           string    `json:"handler_url"`
	IsActive             bool      `json:"is_active"`
	RetryCount           int       `json:"retry_count"`
	MaxRetries           int       `json:"max_retries"`
	LastProcessedEventID int64     `json:"last_processed_event_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// EventProcessingStatus represents the status of event processing
type EventProcessingStatus struct {
	ID               int64     `json:"id"`
	EventID          int64     `json:"event_id"`
	SubscriptionName string    `json:"subscription_name"`
	Status           string    `json:"status"`
	ErrorMessage     string    `json:"error_message"`
	RetryCount       int       `json:"retry_count"`
	MaxRetries       int       `json:"max_retries"`
	ProcessedAt      time.Time `json:"processed_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(eventStore *EventStore, db *sql.DB) *EventSubscriber {
	return &EventSubscriber{
		eventStore: eventStore,
		db:         db,
		handlers:   make(map[string]EventHandler),
	}
}

// Subscribe registers an event handler for a specific event type
func (es *EventSubscriber) Subscribe(eventType string, handler EventHandler) error {
	es.handlers[eventType] = handler

	// Register subscription in database
	subscription := &EventSubscription{
		SubscriptionName:     fmt.Sprintf("handler_%s_%d", eventType, time.Now().Unix()),
		EventType:            eventType,
		HandlerURL:           "internal", // Internal handler
		IsActive:             true,
		RetryCount:           0,
		MaxRetries:           3,
		LastProcessedEventID: 0,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := es.saveSubscription(subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	logger.Log.Info("Event subscription created",
		zap.String("event_type", eventType),
		zap.String("subscription_name", subscription.SubscriptionName))

	return nil
}

// Unsubscribe removes an event handler for a specific event type
func (es *EventSubscriber) Unsubscribe(eventType string) error {
	delete(es.handlers, eventType)

	// Deactivate subscription in database
	query := `UPDATE event_subscriptions SET is_active = false WHERE event_type = ?`
	_, err := es.db.Exec(query, eventType)
	if err != nil {
		return fmt.Errorf("failed to deactivate subscription: %w", err)
	}

	logger.Log.Info("Event subscription removed",
		zap.String("event_type", eventType))

	return nil
}

// ProcessEvents processes events for a specific event type
func (es *EventSubscriber) ProcessEvents(ctx context.Context, eventType string) error {
	// Get subscription
	subscription, err := es.getSubscription(eventType)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("no subscription found for event type: %s", eventType)
	}

	// Get unprocessed events
	events, err := es.getUnprocessedEvents(ctx, eventType, subscription.LastProcessedEventID)
	if err != nil {
		return fmt.Errorf("failed to get unprocessed events: %w", err)
	}

	// Process each event
	for _, event := range events {
		if err := es.processEvent(ctx, event, subscription); err != nil {
			logger.Log.Error("Failed to process event",
				zap.String("event_type", eventType),
				zap.Int64("event_id", event.ID),
				zap.Error(err))
			continue
		}

		// Update last processed event ID
		if err := es.updateLastProcessedEventID(subscription.SubscriptionName, event.ID); err != nil {
			logger.Log.Error("Failed to update last processed event ID",
				zap.String("subscription_name", subscription.SubscriptionName),
				zap.Int64("event_id", event.ID),
				zap.Error(err))
		}
	}

	return nil
}

// ProcessAllEvents processes all events for all subscribed event types
func (es *EventSubscriber) ProcessAllEvents(ctx context.Context) error {
	for eventType := range es.handlers {
		if err := es.ProcessEvents(ctx, eventType); err != nil {
			logger.Log.Error("Failed to process events",
				zap.String("event_type", eventType),
				zap.Error(err))
		}
	}
	return nil
}

// StartProcessing starts the event processing loop
func (es *EventSubscriber) StartProcessing(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Log.Info("Event processing started",
		zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Event processing stopped")
			return
		case <-ticker.C:
			if err := es.ProcessAllEvents(ctx); err != nil {
				logger.Log.Error("Failed to process events",
					zap.Error(err))
			}
		}
	}
}

// processEvent processes a single event
func (es *EventSubscriber) processEvent(ctx context.Context, event *Event, subscription *EventSubscription) error {
	// Create processing status record
	status := &EventProcessingStatus{
		EventID:          event.ID,
		SubscriptionName: subscription.SubscriptionName,
		Status:           "processing",
		RetryCount:       0,
		MaxRetries:       subscription.MaxRetries,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := es.saveProcessingStatus(status); err != nil {
		return fmt.Errorf("failed to save processing status: %w", err)
	}

	// Get handler
	handler, exists := es.handlers[event.EventType]
	if !exists {
		status.Status = "failed"
		status.ErrorMessage = "No handler found for event type"
		es.updateProcessingStatus(status)
		return fmt.Errorf("no handler found for event type: %s", event.EventType)
	}

	// Process event
	if err := handler(ctx, event); err != nil {
		status.Status = "failed"
		status.ErrorMessage = err.Error()
		status.RetryCount++
		es.updateProcessingStatus(status)
		return fmt.Errorf("failed to process event: %w", err)
	}

	// Mark as completed
	status.Status = "completed"
	status.ProcessedAt = time.Now()
	es.updateProcessingStatus(status)

	logger.Log.Debug("Event processed successfully",
		zap.String("event_type", event.EventType),
		zap.Int64("event_id", event.ID),
		zap.String("subscription_name", subscription.SubscriptionName))

	return nil
}

// saveSubscription saves a subscription to the database
func (es *EventSubscriber) saveSubscription(subscription *EventSubscription) error {
	query := `
		INSERT INTO event_subscriptions (
			subscription_name, event_type, handler_url, is_active,
			retry_count, max_retries, last_processed_event_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := es.db.Exec(query,
		subscription.SubscriptionName,
		subscription.EventType,
		subscription.HandlerURL,
		subscription.IsActive,
		subscription.RetryCount,
		subscription.MaxRetries,
		subscription.LastProcessedEventID,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	return err
}

// getSubscription retrieves a subscription by event type
func (es *EventSubscriber) getSubscription(eventType string) (*EventSubscription, error) {
	query := `
		SELECT id, subscription_name, event_type, handler_url, is_active,
		       retry_count, max_retries, last_processed_event_id, created_at, updated_at
		FROM event_subscriptions
		WHERE event_type = ? AND is_active = true
		LIMIT 1
	`

	row := es.db.QueryRow(query, eventType)

	subscription := &EventSubscription{}
	err := row.Scan(
		&subscription.ID,
		&subscription.SubscriptionName,
		&subscription.EventType,
		&subscription.HandlerURL,
		&subscription.IsActive,
		&subscription.RetryCount,
		&subscription.MaxRetries,
		&subscription.LastProcessedEventID,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return subscription, nil
}

// getUnprocessedEvents retrieves unprocessed events for a subscription
func (es *EventSubscriber) getUnprocessedEvents(ctx context.Context, eventType string, lastProcessedEventID int64) ([]*Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, 
		       event_data, event_metadata, version, created_at
		FROM events
		WHERE event_type = ? AND id > ?
		ORDER BY id ASC
		LIMIT 100
	`

	rows, err := es.db.QueryContext(ctx, query, eventType, lastProcessedEventID)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// Unmarshal JSON data
		if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(eventMetadataJSON, &event.EventMetadata); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

// saveProcessingStatus saves event processing status
func (es *EventSubscriber) saveProcessingStatus(status *EventProcessingStatus) error {
	query := `
		INSERT INTO event_processing_status (
			event_id, subscription_name, status, error_message,
			retry_count, max_retries, processed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := es.db.Exec(query,
		status.EventID,
		status.SubscriptionName,
		status.Status,
		status.ErrorMessage,
		status.RetryCount,
		status.MaxRetries,
		status.ProcessedAt,
		status.CreatedAt,
		status.UpdatedAt,
	)

	return err
}

// updateProcessingStatus updates event processing status
func (es *EventSubscriber) updateProcessingStatus(status *EventProcessingStatus) error {
	query := `
		UPDATE event_processing_status
		SET status = ?, error_message = ?, retry_count = ?, processed_at = ?, updated_at = ?
		WHERE event_id = ? AND subscription_name = ?
	`

	_, err := es.db.Exec(query,
		status.Status,
		status.ErrorMessage,
		status.RetryCount,
		status.ProcessedAt,
		status.UpdatedAt,
		status.EventID,
		status.SubscriptionName,
	)

	return err
}

// updateLastProcessedEventID updates the last processed event ID for a subscription
func (es *EventSubscriber) updateLastProcessedEventID(subscriptionName string, eventID int64) error {
	query := `
		UPDATE event_subscriptions
		SET last_processed_event_id = ?, updated_at = ?
		WHERE subscription_name = ?
	`

	_, err := es.db.Exec(query, eventID, time.Now(), subscriptionName)
	return err
}

// GetProcessingStatus returns the processing status for an event
func (es *EventSubscriber) GetProcessingStatus(eventID int64, subscriptionName string) (*EventProcessingStatus, error) {
	query := `
		SELECT id, event_id, subscription_name, status, error_message,
		       retry_count, max_retries, processed_at, created_at, updated_at
		FROM event_processing_status
		WHERE event_id = ? AND subscription_name = ?
	`

	row := es.db.QueryRow(query, eventID, subscriptionName)

	status := &EventProcessingStatus{}
	err := row.Scan(
		&status.ID,
		&status.EventID,
		&status.SubscriptionName,
		&status.Status,
		&status.ErrorMessage,
		&status.RetryCount,
		&status.MaxRetries,
		&status.ProcessedAt,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return status, nil
}

// GetFailedEvents returns events that failed processing
func (es *EventSubscriber) GetFailedEvents(limit int) ([]*EventProcessingStatus, error) {
	query := `
		SELECT id, event_id, subscription_name, status, error_message,
		       retry_count, max_retries, processed_at, created_at, updated_at
		FROM event_processing_status
		WHERE status = 'failed'
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := es.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*EventProcessingStatus
	for rows.Next() {
		status := &EventProcessingStatus{}
		err := rows.Scan(
			&status.ID,
			&status.EventID,
			&status.SubscriptionName,
			&status.Status,
			&status.ErrorMessage,
			&status.RetryCount,
			&status.MaxRetries,
			&status.ProcessedAt,
			&status.CreatedAt,
			&status.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// RetryFailedEvents retries failed events
func (es *EventSubscriber) RetryFailedEvents(ctx context.Context) error {
	failedEvents, err := es.GetFailedEvents(100)
	if err != nil {
		return fmt.Errorf("failed to get failed events: %w", err)
	}

	for _, status := range failedEvents {
		if status.RetryCount >= status.MaxRetries {
			continue // Skip events that have exceeded max retries
		}

		// Get the event
		event, err := es.eventStore.GetEvents(ctx, "", "", 0)
		if err != nil {
			logger.Log.Error("Failed to get event for retry",
				zap.Int64("event_id", status.EventID),
				zap.Error(err))
			continue
		}

		if len(event) == 0 {
			continue
		}

		// Get subscription
		subscription, err := es.getSubscription(event[0].EventType)
		if err != nil {
			logger.Log.Error("Failed to get subscription for retry",
				zap.String("event_type", event[0].EventType),
				zap.Error(err))
			continue
		}

		if subscription == nil {
			continue
		}

		// Retry processing
		if err := es.processEvent(ctx, event[0], subscription); err != nil {
			logger.Log.Error("Failed to retry event",
				zap.Int64("event_id", status.EventID),
				zap.Error(err))
		}
	}

	return nil
}

// Close closes the event subscriber
func (es *EventSubscriber) Close() error {
	// Clear all handlers
	es.handlers = make(map[string]EventHandler)
	return nil
}
