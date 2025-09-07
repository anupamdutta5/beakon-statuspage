package events

import (
	"context"
	"fmt"
	"time"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// EventPublisher handles publishing events to other services
type EventPublisher struct {
	eventStore *EventStore
	handlers   map[string][]EventHandler
}

// EventHandler represents an event handler function
type EventHandler func(ctx context.Context, event *Event) error

// NewEventPublisher creates a new event publisher
func NewEventPublisher(eventStore *EventStore) *EventPublisher {
	return &EventPublisher{
		eventStore: eventStore,
		handlers:   make(map[string][]EventHandler),
	}
}

// PublishEvent publishes an event to the event store and notifies handlers
func (ep *EventPublisher) PublishEvent(ctx context.Context, event *Event) error {
	// Save event to event store
	if err := ep.eventStore.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	// Notify event handlers
	if err := ep.notifyHandlers(ctx, event); err != nil {
		logger.Log.Error("Failed to notify event handlers",
			zap.String("event_type", event.EventType),
			zap.String("aggregate_id", event.AggregateID),
			zap.Error(err))
		// Don't return error here as the event is already saved
	}

	logger.Log.Info("Event published",
		zap.String("event_type", event.EventType),
		zap.String("aggregate_id", event.AggregateID),
		zap.String("aggregate_type", event.AggregateType),
		zap.Int("version", event.Version))

	return nil
}

// Subscribe registers an event handler for a specific event type
func (ep *EventPublisher) Subscribe(eventType string, handler EventHandler) {
	ep.handlers[eventType] = append(ep.handlers[eventType], handler)

	logger.Log.Info("Event handler subscribed",
		zap.String("event_type", eventType))
}

// Unsubscribe removes an event handler for a specific event type
func (ep *EventPublisher) Unsubscribe(eventType string, handler EventHandler) {
	handlers := ep.handlers[eventType]
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			ep.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	logger.Log.Info("Event handler unsubscribed",
		zap.String("event_type", eventType))
}

// notifyHandlers notifies all registered handlers for an event type
func (ep *EventPublisher) notifyHandlers(ctx context.Context, event *Event) error {
	handlers, exists := ep.handlers[event.EventType]
	if !exists {
		return nil // No handlers registered for this event type
	}

	var lastErr error
	for _, handler := range handlers {
		if err := ep.callHandler(ctx, handler, event); err != nil {
			logger.Log.Error("Event handler failed",
				zap.String("event_type", event.EventType),
				zap.String("aggregate_id", event.AggregateID),
				zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// callHandler calls an event handler with error handling
func (ep *EventPublisher) callHandler(ctx context.Context, handler EventHandler, event *Event) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Event handler panicked",
				zap.String("event_type", event.EventType),
				zap.String("aggregate_id", event.AggregateID),
				zap.Any("panic", r))
		}
	}()

	// Create a context with timeout for the handler
	handlerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return handler(handlerCtx, event)
}

// PublishUserCreated publishes a user created event
func (ep *EventPublisher) PublishUserCreated(ctx context.Context, userID string, userData map[string]interface{}) error {
	event := &Event{
		AggregateID:   userID,
		AggregateType: "User",
		EventType:     "UserCreated",
		EventData:     userData,
		EventMetadata: map[string]interface{}{
			"source":  "user-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishUserUpdated publishes a user updated event
func (ep *EventPublisher) PublishUserUpdated(ctx context.Context, userID string, userData map[string]interface{}) error {
	event := &Event{
		AggregateID:   userID,
		AggregateType: "User",
		EventType:     "UserUpdated",
		EventData:     userData,
		EventMetadata: map[string]interface{}{
			"source":  "user-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishUserDeleted publishes a user deleted event
func (ep *EventPublisher) PublishUserDeleted(ctx context.Context, userID string, userData map[string]interface{}) error {
	event := &Event{
		AggregateID:   userID,
		AggregateType: "User",
		EventType:     "UserDeleted",
		EventData:     userData,
		EventMetadata: map[string]interface{}{
			"source":  "user-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishTenantCreated publishes a tenant created event
func (ep *EventPublisher) PublishTenantCreated(ctx context.Context, tenantID string, tenantData map[string]interface{}) error {
	event := &Event{
		AggregateID:   tenantID,
		AggregateType: "Tenant",
		EventType:     "TenantCreated",
		EventData:     tenantData,
		EventMetadata: map[string]interface{}{
			"source":  "tenant-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishTenantUpdated publishes a tenant updated event
func (ep *EventPublisher) PublishTenantUpdated(ctx context.Context, tenantID string, tenantData map[string]interface{}) error {
	event := &Event{
		AggregateID:   tenantID,
		AggregateType: "Tenant",
		EventType:     "TenantUpdated",
		EventData:     tenantData,
		EventMetadata: map[string]interface{}{
			"source":  "tenant-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishComponentCreated publishes a component created event
func (ep *EventPublisher) PublishComponentCreated(ctx context.Context, componentID string, componentData map[string]interface{}) error {
	event := &Event{
		AggregateID:   componentID,
		AggregateType: "Component",
		EventType:     "ComponentCreated",
		EventData:     componentData,
		EventMetadata: map[string]interface{}{
			"source":  "component-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishComponentStatusChanged publishes a component status changed event
func (ep *EventPublisher) PublishComponentStatusChanged(ctx context.Context, componentID string, statusData map[string]interface{}) error {
	event := &Event{
		AggregateID:   componentID,
		AggregateType: "Component",
		EventType:     "ComponentStatusChanged",
		EventData:     statusData,
		EventMetadata: map[string]interface{}{
			"source":  "component-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishIncidentCreated publishes an incident created event
func (ep *EventPublisher) PublishIncidentCreated(ctx context.Context, incidentID string, incidentData map[string]interface{}) error {
	event := &Event{
		AggregateID:   incidentID,
		AggregateType: "Incident",
		EventType:     "IncidentCreated",
		EventData:     incidentData,
		EventMetadata: map[string]interface{}{
			"source":  "incident-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishIncidentResolved publishes an incident resolved event
func (ep *EventPublisher) PublishIncidentResolved(ctx context.Context, incidentID string, incidentData map[string]interface{}) error {
	event := &Event{
		AggregateID:   incidentID,
		AggregateType: "Incident",
		EventType:     "IncidentResolved",
		EventData:     incidentData,
		EventMetadata: map[string]interface{}{
			"source":  "incident-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishPaymentProcessed publishes a payment processed event
func (ep *EventPublisher) PublishPaymentProcessed(ctx context.Context, paymentID string, paymentData map[string]interface{}) error {
	event := &Event{
		AggregateID:   paymentID,
		AggregateType: "Payment",
		EventType:     "PaymentProcessed",
		EventData:     paymentData,
		EventMetadata: map[string]interface{}{
			"source":  "payment-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishSubscriptionCreated publishes a subscription created event
func (ep *EventPublisher) PublishSubscriptionCreated(ctx context.Context, subscriptionID string, subscriptionData map[string]interface{}) error {
	event := &Event{
		AggregateID:   subscriptionID,
		AggregateType: "Subscription",
		EventType:     "SubscriptionCreated",
		EventData:     subscriptionData,
		EventMetadata: map[string]interface{}{
			"source":  "payment-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// PublishSubscriptionCancelled publishes a subscription cancelled event
func (ep *EventPublisher) PublishSubscriptionCancelled(ctx context.Context, subscriptionID string, subscriptionData map[string]interface{}) error {
	event := &Event{
		AggregateID:   subscriptionID,
		AggregateType: "Subscription",
		EventType:     "SubscriptionCancelled",
		EventData:     subscriptionData,
		EventMetadata: map[string]interface{}{
			"source":  "payment-service",
			"version": "1.0.0",
		},
		Version: 1,
	}

	return ep.PublishEvent(ctx, event)
}

// GetEventHandlers returns all registered event handlers
func (ep *EventPublisher) GetEventHandlers() map[string][]EventHandler {
	return ep.handlers
}

// GetEventHandlerCount returns the number of handlers for an event type
func (ep *EventPublisher) GetEventHandlerCount(eventType string) int {
	return len(ep.handlers[eventType])
}

// GetSupportedEventTypes returns all supported event types
func (ep *EventPublisher) GetSupportedEventTypes() []string {
	var eventTypes []string
	for eventType := range ep.handlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// Close closes the event publisher
func (ep *EventPublisher) Close() error {
	// Clear all handlers
	ep.handlers = make(map[string][]EventHandler)
	return nil
}
