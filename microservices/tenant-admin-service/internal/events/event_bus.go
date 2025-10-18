package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EventType defines different event types
type EventType string

const (
	// Tenant events
	TenantCreated EventType = "tenant.created"
	TenantUpdated EventType = "tenant.updated"
	TenantDeleted EventType = "tenant.deleted"

	// User events
	UserCreated    EventType = "user.created"
	UserUpdated    EventType = "user.updated"
	UserDeleted    EventType = "user.deleted"
	UserLoggedIn   EventType = "user.logged_in"
	UserLoggedOut  EventType = "user.logged_out"

	// Session events
	SessionCreated EventType = "session.created"
	SessionExpired EventType = "session.expired"
	SessionDeleted EventType = "session.deleted"

	// System events
	SystemHealthCheck EventType = "system.health_check"
	SystemError       EventType = "system.error"
	SystemWarning     EventType = "system.warning"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata"`
}

// EventHandler processes events
type EventHandler func(ctx context.Context, event Event) error

// EventBus manages event publishing and subscription
type EventBus struct {
	handlers       map[EventType][]EventHandler
	asyncHandlers  map[EventType][]EventHandler
	mu             sync.RWMutex
	logger         *zap.Logger
	buffer         chan Event
	workers        int
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	errorHandler   func(error, Event)
	metrics        *EventMetrics
}

// EventMetrics tracks event processing metrics
type EventMetrics struct {
	mu               sync.RWMutex
	eventsPublished  map[EventType]int64
	eventsProcessed  map[EventType]int64
	eventsFailed     map[EventType]int64
	processingTime   map[EventType]time.Duration
	lastEventTime    map[EventType]time.Time
}

// NewEventBus creates a new event bus
func NewEventBus(logger *zap.Logger, bufferSize int, workers int) *EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	eb := &EventBus{
		handlers:      make(map[EventType][]EventHandler),
		asyncHandlers: make(map[EventType][]EventHandler),
		logger:        logger,
		buffer:        make(chan Event, bufferSize),
		workers:       workers,
		ctx:           ctx,
		cancel:        cancel,
		errorHandler:  defaultErrorHandler(logger),
		metrics:       newEventMetrics(),
	}

	// Start worker pool
	eb.startWorkers()

	return eb
}

// Subscribe registers a synchronous handler for an event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Debug("Subscribed handler to event",
		zap.String("event_type", string(eventType)))
}

// SubscribeAsync registers an asynchronous handler for an event type
func (eb *EventBus) SubscribeAsync(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.asyncHandlers[eventType] = append(eb.asyncHandlers[eventType], handler)
	eb.logger.Debug("Subscribed async handler to event",
		zap.String("event_type", string(eventType)))
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	// Set event ID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Update metrics
	eb.metrics.recordPublished(event.Type)

	// Process synchronous handlers first
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		if err := eb.executeHandler(ctx, handler, event); err != nil {
			eb.logger.Error("Synchronous handler failed",
				zap.String("event_id", event.ID),
				zap.String("event_type", string(event.Type)),
				zap.Error(err))
			return err // Fail fast for sync handlers
		}
	}

	// Queue async handlers
	select {
	case eb.buffer <- event:
		eb.logger.Debug("Event queued for async processing",
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)))
	case <-time.After(100 * time.Millisecond):
		eb.logger.Warn("Event buffer full, dropping event",
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)))
		return fmt.Errorf("event buffer full")
	}

	return nil
}

// PublishAsync publishes an event asynchronously without waiting
func (eb *EventBus) PublishAsync(event Event) {
	go func() {
		if err := eb.Publish(context.Background(), event); err != nil {
			eb.logger.Error("Failed to publish async event",
				zap.String("event_id", event.ID),
				zap.String("event_type", string(event.Type)),
				zap.Error(err))
		}
	}()
}

// startWorkers starts the worker pool for async event processing
func (eb *EventBus) startWorkers() {
	for i := 0; i < eb.workers; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}
}

// worker processes events from the buffer
func (eb *EventBus) worker(id int) {
	defer eb.wg.Done()

	eb.logger.Info("Event worker started",
		zap.Int("worker_id", id))

	for {
		select {
		case event := <-eb.buffer:
			eb.processAsyncEvent(event)
		case <-eb.ctx.Done():
			eb.logger.Info("Event worker stopping",
				zap.Int("worker_id", id))
			return
		}
	}
}

// processAsyncEvent processes an event asynchronously
func (eb *EventBus) processAsyncEvent(event Event) {
	ctx, cancel := context.WithTimeout(eb.ctx, 30*time.Second)
	defer cancel()

	eb.mu.RLock()
	handlers := eb.asyncHandlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		if err := eb.executeHandler(ctx, handler, event); err != nil {
			eb.logger.Error("Async handler failed",
				zap.String("event_id", event.ID),
				zap.String("event_type", string(event.Type)),
				zap.Error(err))
			eb.errorHandler(err, event)
			eb.metrics.recordFailed(event.Type)
		}
	}
}

// executeHandler executes a handler with timing and error handling
func (eb *EventBus) executeHandler(ctx context.Context, handler EventHandler, event Event) error {
	start := time.Now()

	// Execute with panic recovery
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("handler panic: %v", r)
				eb.logger.Error("Handler panicked",
					zap.String("event_id", event.ID),
					zap.String("event_type", string(event.Type)),
					zap.Any("panic", r))
			}
		}()
		err = handler(ctx, event)
	}()

	// Record metrics
	duration := time.Since(start)
	eb.metrics.recordProcessed(event.Type, duration)

	return err
}

// SetErrorHandler sets a custom error handler
func (eb *EventBus) SetErrorHandler(handler func(error, Event)) {
	eb.errorHandler = handler
}

// Stop gracefully stops the event bus
func (eb *EventBus) Stop(timeout time.Duration) error {
	eb.logger.Info("Stopping event bus")

	// Signal workers to stop
	eb.cancel()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		eb.logger.Info("Event bus stopped gracefully")
		return nil
	case <-time.After(timeout):
		eb.logger.Warn("Event bus stop timeout exceeded")
		return fmt.Errorf("stop timeout exceeded")
	}
}

// GetMetrics returns event processing metrics
func (eb *EventBus) GetMetrics() map[string]interface{} {
	return eb.metrics.toMap()
}

// EventStore provides event persistence
type EventStore interface {
	Store(ctx context.Context, event Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	GetByType(ctx context.Context, eventType EventType, limit int) ([]Event, error)
	GetByTenant(ctx context.Context, tenantID string, limit int) ([]Event, error)
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]Event, error)
}

// OutboxPattern implements transactional outbox pattern for reliable event publishing
type OutboxPattern struct {
	bus    *EventBus
	store  EventStore
	logger *zap.Logger
	ticker *time.Ticker
	stop   chan struct{}
}

// NewOutboxPattern creates a new outbox pattern processor
func NewOutboxPattern(bus *EventBus, store EventStore, logger *zap.Logger) *OutboxPattern {
	op := &OutboxPattern{
		bus:    bus,
		store:  store,
		logger: logger,
		ticker: time.NewTicker(5 * time.Second),
		stop:   make(chan struct{}),
	}

	go op.processOutbox()

	return op
}

// processOutbox periodically processes pending events
func (op *OutboxPattern) processOutbox() {
	for {
		select {
		case <-op.ticker.C:
			// Fetch and process pending events from store
			// Implementation depends on your storage backend
		case <-op.stop:
			op.ticker.Stop()
			return
		}
	}
}

// Helper functions

func defaultErrorHandler(logger *zap.Logger) func(error, Event) {
	return func(err error, event Event) {
		logger.Error("Event processing failed",
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)),
			zap.Error(err))
	}
}

func newEventMetrics() *EventMetrics {
	return &EventMetrics{
		eventsPublished: make(map[EventType]int64),
		eventsProcessed: make(map[EventType]int64),
		eventsFailed:    make(map[EventType]int64),
		processingTime:  make(map[EventType]time.Duration),
		lastEventTime:   make(map[EventType]time.Time),
	}
}

func (m *EventMetrics) recordPublished(eventType EventType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsPublished[eventType]++
	m.lastEventTime[eventType] = time.Now()
}

func (m *EventMetrics) recordProcessed(eventType EventType, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsProcessed[eventType]++
	m.processingTime[eventType] += duration
}

func (m *EventMetrics) recordFailed(eventType EventType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsFailed[eventType]++
}

func (m *EventMetrics) toMap() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})
	for eventType, count := range m.eventsPublished {
		key := string(eventType)
		result[key] = map[string]interface{}{
			"published":     count,
			"processed":     m.eventsProcessed[eventType],
			"failed":        m.eventsFailed[eventType],
			"avg_duration":  m.processingTime[eventType] / time.Duration(m.eventsProcessed[eventType]+1),
			"last_event":    m.lastEventTime[eventType],
		}
	}

	return result
}

// CreateEvent is a helper function to create events
func CreateEvent(eventType EventType, tenantID, userID string, data map[string]interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		TenantID:  tenantID,
		UserID:    userID,
		Data:      data,
		Metadata: map[string]string{
			"source": "tenant-admin-service",
		},
	}
}