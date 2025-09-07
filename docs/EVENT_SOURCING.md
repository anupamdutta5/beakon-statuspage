# Event Sourcing Implementation

This document describes the event sourcing implementation for the Status Page microservices architecture, including the event store, event publisher, and event subscriber components.

## Overview

Event sourcing is a pattern where the state of an application is determined by a sequence of events. Instead of storing the current state, we store all events that have occurred and reconstruct the state by replaying these events.

## Architecture

### Components

1. **Event Store** - Central repository for all events
2. **Event Publisher** - Publishes events to the event store and notifies handlers
3. **Event Subscriber** - Subscribes to events and processes them
4. **Event Handlers** - Business logic for processing specific event types

### Event Flow

```
Service A → Event Publisher → Event Store → Event Subscriber → Service B
```

## Event Store

The event store is implemented in `internal/events/event_store.go` and provides:

### Core Operations

- **SaveEvent** - Saves an event to the event store
- **GetEvents** - Retrieves events for an aggregate
- **SaveSnapshot** - Saves an aggregate snapshot
- **GetLatestSnapshot** - Retrieves the latest snapshot for an aggregate

### Event Structure

```go
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
```

### Snapshot Structure

```go
type Snapshot struct {
    ID            int64                  `json:"id"`
    AggregateID   string                 `json:"aggregate_id"`
    AggregateType string                 `json:"aggregate_type"`
    SnapshotData  map[string]interface{} `json:"snapshot_data"`
    Version       int                    `json:"version"`
    CreatedAt     time.Time              `json:"created_at"`
}
```

## Event Publisher

The event publisher is implemented in `internal/events/event_publisher.go` and provides:

### Core Operations

- **PublishEvent** - Publishes an event to the event store and notifies handlers
- **Subscribe** - Registers an event handler for a specific event type
- **Unsubscribe** - Removes an event handler for a specific event type

### Predefined Events

The event publisher provides predefined methods for common events:

- **User Events**: `PublishUserCreated`, `PublishUserUpdated`, `PublishUserDeleted`
- **Tenant Events**: `PublishTenantCreated`, `PublishTenantUpdated`
- **Component Events**: `PublishComponentCreated`, `PublishComponentStatusChanged`
- **Incident Events**: `PublishIncidentCreated`, `PublishIncidentResolved`
- **Payment Events**: `PublishPaymentProcessed`, `PublishSubscriptionCreated`, `PublishSubscriptionCancelled`

### Usage Example

```go
// Create event publisher
eventStore := events.NewEventStore(db)
publisher := events.NewEventPublisher(eventStore)

// Publish a user created event
userData := map[string]interface{}{
    "id": "user123",
    "email": "user@example.com",
    "name": "John Doe",
}

err := publisher.PublishUserCreated(ctx, "user123", userData)
if err != nil {
    log.Printf("Failed to publish user created event: %v", err)
}
```

## Event Subscriber

The event subscriber is implemented in `internal/events/event_subscriber.go` and provides:

### Core Operations

- **Subscribe** - Registers an event handler for a specific event type
- **Unsubscribe** - Removes an event handler for a specific event type
- **ProcessEvents** - Processes events for a specific event type
- **ProcessAllEvents** - Processes all events for all subscribed event types
- **StartProcessing** - Starts the event processing loop

### Event Processing

The event subscriber processes events in the following steps:

1. **Get Subscription** - Retrieves the subscription for the event type
2. **Get Unprocessed Events** - Retrieves events that haven't been processed
3. **Process Each Event** - Calls the appropriate event handler
4. **Update Status** - Updates the processing status and last processed event ID

### Usage Example

```go
// Create event subscriber
eventStore := events.NewEventStore(db)
subscriber := events.NewEventSubscriber(eventStore, db)

// Subscribe to user created events
err := subscriber.Subscribe("UserCreated", func(ctx context.Context, event *events.Event) error {
    // Handle user created event
    userID := event.EventData["id"].(string)
    email := event.EventData["email"].(string)
    
    // Send welcome email, create user profile, etc.
    return sendWelcomeEmail(email)
})
if err != nil {
    log.Printf("Failed to subscribe to UserCreated events: %v", err)
}

// Start processing events
go subscriber.StartProcessing(ctx, 5*time.Second)
```

## Database Schema

The event store uses the following database tables:

### Events Table

```sql
CREATE TABLE events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSON NOT NULL,
    event_metadata JSON,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_aggregate (aggregate_id, aggregate_type),
    INDEX idx_event_type (event_type),
    INDEX idx_created_at (created_at)
);
```

### Snapshots Table

```sql
CREATE TABLE snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    snapshot_data JSON NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_aggregate (aggregate_id, aggregate_type),
    INDEX idx_created_at (created_at)
);
```

### Event Subscriptions Table

```sql
CREATE TABLE event_subscriptions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    subscription_name VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    handler_url VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    last_processed_event_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_subscription (subscription_name, event_type),
    INDEX idx_event_type (event_type),
    INDEX idx_is_active (is_active)
);
```

### Event Processing Status Table

```sql
CREATE TABLE event_processing_status (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id BIGINT NOT NULL,
    subscription_name VARCHAR(100) NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    error_message TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_event_subscription (event_id, subscription_name),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

## Event Types

### User Events

- **UserCreated** - Triggered when a new user is created
- **UserUpdated** - Triggered when a user is updated
- **UserDeleted** - Triggered when a user is deleted

### Tenant Events

- **TenantCreated** - Triggered when a new tenant is created
- **TenantUpdated** - Triggered when a tenant is updated
- **TenantDeleted** - Triggered when a tenant is deleted

### Component Events

- **ComponentCreated** - Triggered when a new component is created
- **ComponentUpdated** - Triggered when a component is updated
- **ComponentDeleted** - Triggered when a component is deleted
- **ComponentStatusChanged** - Triggered when a component status changes

### Incident Events

- **IncidentCreated** - Triggered when a new incident is created
- **IncidentUpdated** - Triggered when an incident is updated
- **IncidentResolved** - Triggered when an incident is resolved
- **IncidentClosed** - Triggered when an incident is closed

### Payment Events

- **PaymentProcessed** - Triggered when a payment is processed
- **PaymentFailed** - Triggered when a payment fails
- **SubscriptionCreated** - Triggered when a subscription is created
- **SubscriptionUpdated** - Triggered when a subscription is updated
- **SubscriptionCancelled** - Triggered when a subscription is cancelled

### Notification Events

- **NotificationSent** - Triggered when a notification is sent
- **NotificationFailed** - Triggered when a notification fails
- **NotificationDelivered** - Triggered when a notification is delivered

## Best Practices

### Event Design

1. **Immutable Events** - Events should never be modified after creation
2. **Versioned Events** - Use versioning for event schema evolution
3. **Rich Event Data** - Include all necessary data in the event
4. **Event Metadata** - Include metadata like source, version, timestamp

### Event Handling

1. **Idempotent Handlers** - Event handlers should be idempotent
2. **Error Handling** - Implement proper error handling and retry logic
3. **Async Processing** - Process events asynchronously when possible
4. **Monitoring** - Monitor event processing performance and errors

### Performance

1. **Snapshots** - Use snapshots for aggregates with many events
2. **Indexing** - Properly index event store tables
3. **Partitioning** - Consider partitioning for high-volume event stores
4. **Cleanup** - Implement event cleanup and archiving strategies

## Monitoring and Observability

### Metrics

- **Event Count** - Total number of events by type
- **Processing Time** - Time to process events
- **Error Rate** - Percentage of failed event processing
- **Lag** - Time between event creation and processing

### Logging

- **Event Creation** - Log when events are created
- **Event Processing** - Log event processing start and completion
- **Errors** - Log event processing errors with context
- **Performance** - Log processing times and performance metrics

### Health Checks

- **Event Store Health** - Check event store connectivity
- **Processing Health** - Check event processing status
- **Subscription Health** - Check subscription status
- **Error Health** - Check for failed events

## Error Handling

### Retry Logic

- **Exponential Backoff** - Implement exponential backoff for retries
- **Max Retries** - Set maximum retry attempts
- **Dead Letter Queue** - Move failed events to a dead letter queue
- **Manual Intervention** - Allow manual retry of failed events

### Error Types

- **Transient Errors** - Network issues, temporary service unavailability
- **Permanent Errors** - Invalid event data, business rule violations
- **System Errors** - Database errors, service crashes

## Testing

### Unit Tests

- **Event Creation** - Test event creation and validation
- **Event Processing** - Test event processing logic
- **Error Handling** - Test error handling and retry logic
- **Snapshot Management** - Test snapshot creation and retrieval

### Integration Tests

- **End-to-End Flow** - Test complete event flow
- **Database Operations** - Test database operations
- **Service Integration** - Test service integration
- **Performance Tests** - Test event processing performance

## Deployment

### Database Migration

1. **Create Event Store** - Create event store database and tables
2. **Create Indexes** - Create necessary indexes for performance
3. **Create Views** - Create views for common queries
4. **Create Procedures** - Create stored procedures for common operations

### Service Deployment

1. **Event Publisher** - Deploy event publisher service
2. **Event Subscriber** - Deploy event subscriber service
3. **Event Handlers** - Deploy event handler services
4. **Monitoring** - Deploy monitoring and alerting

## Security

### Access Control

- **Database Access** - Restrict database access to event store
- **Service Authentication** - Authenticate services publishing events
- **Event Validation** - Validate event data before processing
- **Audit Logging** - Log all event store operations

### Data Protection

- **Encryption** - Encrypt sensitive event data
- **Data Retention** - Implement data retention policies
- **Backup** - Regular backup of event store
- **Recovery** - Disaster recovery procedures

## Conclusion

The event sourcing implementation provides a robust foundation for building event-driven microservices. It enables:

- **Loose Coupling** - Services communicate through events
- **Scalability** - Events can be processed asynchronously
- **Auditability** - Complete audit trail of all changes
- **Replayability** - Events can be replayed for debugging or recovery
- **Flexibility** - Easy to add new event handlers and subscribers

This implementation follows industry best practices and provides a solid foundation for building scalable, maintainable, and reliable microservices.
