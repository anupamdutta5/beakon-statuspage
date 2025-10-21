# Event Store Service

## Overview

The Event Store Service is an immutable event log and audit trail system that implements the Event Sourcing pattern. It serves as the single source of truth for all events that occur across the Beakon platform, providing a complete, chronological history of state changes, user actions, and system events. Every change is captured as an immutable event, enabling time-travel debugging, audit compliance, and event replay capabilities.

**Service Port**: 8093

## Core Responsibilities

### 1. Event Storage & Persistence
- Store immutable events in chronological order
- Organize events into logical streams (aggregate-based)
- Maintain event ordering and versioning
- Ensure event immutability (append-only architecture)
- Provide durable, reliable event persistence

### 2. Event Retrieval & Querying
- Query events by stream
- Filter events by event type, timestamp, or metadata
- Support pagination for large event sets
- Retrieve events in chronological order
- Provide snapshot support for efficient state reconstruction

### 3. Audit Trail & Compliance
- Maintain complete audit log of all system activities
- Track who did what, when, and why
- Support compliance requirements (SOC 2, GDPR, HIPAA)
- Enable forensic analysis and debugging
- Provide tamper-proof event history

### 4. Event Replay & Time Travel
- Replay events to reconstruct past states
- Support point-in-time state reconstruction
- Enable debugging by replaying specific event sequences
- Facilitate data migration and system recovery
- Allow "what-if" scenario analysis

### 5. Event Subscriptions & Notifications
- Support event subscriptions for real-time notifications
- Enable event-driven architecture across services
- Provide event streaming capabilities
- Support multiple concurrent subscribers
- Enable fan-out event distribution

## What is Event Sourcing?

**Event Sourcing** is an architectural pattern where state changes are stored as a sequence of events rather than just storing the current state. Instead of updating a record in place, you append a new event describing what happened.

### Traditional CRUD vs Event Sourcing

**Traditional Approach**:
```
User balance = $100
User makes $20 purchase
User balance = $80  (overwrites previous value)
```
You only know the current balance ($80), not the history of how it got there.

**Event Sourcing Approach**:
```
Event 1: AccountCreated { balance: $100 }
Event 2: PurchaseMade { amount: $20, balance: $80 }
Event 3: RefundIssued { amount: $10, balance: $90 }
```
You have complete history and can reconstruct the state at any point in time.

## Architecture

### Data Models

#### Event Stream
```go
type EventStream struct {
    ID          string    `json:"id"`           // Unique stream identifier (e.g., "user-123")
    Type        string    `json:"type"`         // Stream type (e.g., "user", "incident", "payment")
    TenantID    string    `json:"tenant_id"`    // Multi-tenancy support
    Version     int64     `json:"version"`      // Current version (event count)
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### Event
```go
type Event struct {
    ID              string            `json:"id"`              // Unique event ID (UUID)
    StreamID        string            `json:"stream_id"`       // Which stream this belongs to
    EventType       string            `json:"event_type"`      // Type of event (e.g., "UserCreated")
    EventVersion    int64             `json:"event_version"`   // Version within the stream
    Data            json.RawMessage   `json:"data"`            // Event payload (JSON)
    Metadata        map[string]string `json:"metadata"`        // Additional metadata
    Timestamp       time.Time         `json:"timestamp"`       // When event occurred
    CausationID     string            `json:"causation_id"`    // ID of command that caused this
    CorrelationID   string            `json:"correlation_id"`  // ID linking related events
    UserID          string            `json:"user_id"`         // Who performed the action
}
```

#### Snapshot
```go
type Snapshot struct {
    ID              string          `json:"id"`
    StreamID        string          `json:"stream_id"`
    Version         int64           `json:"version"`         // Stream version at snapshot time
    State           json.RawMessage `json:"state"`           // Current state at this version
    Timestamp       time.Time       `json:"timestamp"`
}
```

#### Subscription
```go
type Subscription struct {
    ID              string   `json:"id"`
    SubscriberID    string   `json:"subscriber_id"`    // Service subscribing to events
    StreamPatterns  []string `json:"stream_patterns"`  // Stream patterns to match
    EventTypes      []string `json:"event_types"`      // Event types to receive
    LastEventID     string   `json:"last_event_id"`    // Last processed event
    IsActive        bool     `json:"is_active"`
}
```

## API Endpoints

### Health & Status
- `GET /health` - Service health check
- `GET /metrics` - Service metrics (Prometheus format)

### Event Stream Management
- `GET /api/v1/streams` - List all event streams
- `POST /api/v1/streams` - Create new event stream
- `GET /api/v1/streams/:id` - Get stream details and metadata

### Event Management
- `GET /api/v1/streams/:stream_id/events` - Get all events in a stream
- `POST /api/v1/streams/:stream_id/events` - Append new event to stream
- `GET /api/v1/streams/:stream_id/events/:version` - Get event at specific version
- `GET /api/v1/streams/:stream_id/events/range` - Get events in version range

### Event Querying
- `GET /api/v1/events` - Query events across all streams
- `GET /api/v1/events/by-type/:event_type` - Get events by type
- `GET /api/v1/events/by-correlation/:correlation_id` - Get related events
- `GET /api/v1/events/by-user/:user_id` - Get events by user

### Snapshots
- `GET /api/v1/streams/:stream_id/snapshots` - List snapshots for stream
- `POST /api/v1/streams/:stream_id/snapshots` - Create snapshot
- `GET /api/v1/streams/:stream_id/snapshots/latest` - Get latest snapshot
- `GET /api/v1/streams/:stream_id/state` - Get current state (from snapshot + events)

### Subscriptions
- `GET /api/v1/subscriptions` - List subscriptions
- `POST /api/v1/subscriptions` - Create subscription
- `GET /api/v1/subscriptions/:id` - Get subscription details
- `PUT /api/v1/subscriptions/:id` - Update subscription
- `DELETE /api/v1/subscriptions/:id` - Delete subscription
- `POST /api/v1/subscriptions/:id/acknowledge` - Acknowledge processed events

### Event Replay
- `POST /api/v1/streams/:stream_id/replay` - Replay events from stream
- `POST /api/v1/replay/from-version` - Replay from specific version
- `POST /api/v1/replay/to-service` - Replay events to specific service

### Projections (Read Models)
- `GET /api/v1/projections` - List projections
- `POST /api/v1/projections` - Create projection
- `POST /api/v1/projections/:id/rebuild` - Rebuild projection from events

## Dependencies

### Internal Services
- **All Services**: Publishes events to Event Store for audit trail
- **Analytics Service**: Consumes events for analytics and reporting
- **Notification Service**: Subscribes to events for triggering notifications
- **Incident Service**: Publishes incident lifecycle events
- **User Service**: Publishes user action events
- **Payment Service**: Publishes payment transaction events

### External Dependencies
- PostgreSQL database for event persistence (append-only tables)
- Redis (optional) for caching and rate limiting

## Environment Variables

```bash
# Server Configuration
SERVER_PORT=8093
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=60s
SERVER_GRACEFUL_STOP=10s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=event_store_db
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT Authentication
JWT_SECRET=your_jwt_secret_key

# Cache Configuration (Optional)
CACHE_ENABLED=true
CACHE_TYPE=redis  # redis or memory
CACHE_TTL=5m

# Redis Configuration (if using Redis cache)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=1000

# Circuit Breaker
CIRCUIT_BREAKER_DATABASE_ENABLED=true
CIRCUIT_BREAKER_EXTERNAL_ENABLED=true

# Monitoring
MONITORING_ENABLED=true
METRICS_ENABLED=true

# Environment
ENVIRONMENT=development  # development, staging, production
```

## How to Run

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 14 or higher
- Redis (optional, for caching)

### Local Development

1. **Set up database**:
```bash
createdb event_store_db
```

2. **Run migrations** (create event tables):
```sql
-- Create streams table
CREATE TABLE event_streams (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(255),
    version BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create events table (append-only)
CREATE TABLE events (
    id VARCHAR(255) PRIMARY KEY,
    stream_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_version BIGINT NOT NULL,
    data JSONB NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP DEFAULT NOW(),
    causation_id VARCHAR(255),
    correlation_id VARCHAR(255),
    user_id VARCHAR(255),
    FOREIGN KEY (stream_id) REFERENCES event_streams(id),
    UNIQUE (stream_id, event_version)
);

-- Create indexes for fast querying
CREATE INDEX idx_events_stream_id ON events(stream_id);
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_correlation_id ON events(correlation_id);
CREATE INDEX idx_events_user_id ON events(user_id);

-- Create snapshots table
CREATE TABLE snapshots (
    id VARCHAR(255) PRIMARY KEY,
    stream_id VARCHAR(255) NOT NULL,
    version BIGINT NOT NULL,
    state JSONB NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (stream_id) REFERENCES event_streams(id),
    UNIQUE (stream_id, version)
);
```

3. **Set environment variables**:
```bash
export DB_PASSWORD="postgres"
export DB_NAME="event_store_db"
export SERVER_PORT="8093"
export JWT_SECRET="your_jwt_secret"
```

4. **Build the service**:
```bash
go build -o event-store-service cmd/main.go
```

5. **Run the service**:
```bash
./event-store-service
```

### Using Docker

```bash
docker build -t event-store-service .
docker run -p 8093:8093 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=event_store_db \
  -e JWT_SECRET=your_jwt_secret \
  event-store-service
```

## Testing

### Unit Tests
```bash
go test ./tests/unit/...
```

### Integration Tests
```bash
go test ./tests/integration/...
```

## Key Features

### 1. Immutable Event Log
Events are never updated or deleted, only appended. This ensures:
- Complete audit trail
- No data loss
- Ability to reconstruct any past state
- Tamper-proof history

### 2. Event Versioning
Each event has a version number within its stream, ensuring:
- Optimistic concurrency control
- Conflict detection
- Ordered event processing

### 3. Event Correlation
Events can be linked together using:
- **Causation ID**: The command/event that caused this event
- **Correlation ID**: Groups related events across multiple streams
- **User ID**: Tracks which user initiated the action

### 4. Snapshot Support
Snapshots optimize state reconstruction:
- Instead of replaying 10,000 events, load latest snapshot + new events
- Configurable snapshot frequency (every N events)
- Automatic snapshot creation on high-volume streams

### 5. Multi-Tenancy Support
- Events are isolated by tenant
- Stream IDs include tenant context
- Queries are automatically scoped to tenant

### 6. Event Subscriptions
Services can subscribe to events for real-time processing:
- Push-based notifications
- At-least-once delivery guarantee
- Automatic retry on failure
- Checkpoint management

## Use Cases

### 1. User Action Audit Trail
Track every user action for compliance:
```json
POST /api/v1/streams/user-123/events
{
  "event_type": "UserLoggedIn",
  "data": {
    "user_id": "123",
    "ip_address": "192.168.1.1",
    "device": "Chrome on MacOS"
  },
  "metadata": {
    "source": "auth-service",
    "request_id": "req-456"
  }
}
```

### 2. Incident Lifecycle Tracking
Record complete incident history:
```json
POST /api/v1/streams/incident-789/events
{
  "event_type": "IncidentCreated",
  "data": {
    "incident_id": "789",
    "title": "Database unavailable",
    "severity": "critical"
  },
  "user_id": "admin-123"
}

POST /api/v1/streams/incident-789/events
{
  "event_type": "IncidentStatusUpdated",
  "data": {
    "incident_id": "789",
    "old_status": "investigating",
    "new_status": "resolved",
    "message": "Database restored from backup"
  },
  "user_id": "admin-123"
}
```

### 3. Payment Transaction History
Maintain immutable payment records:
```json
POST /api/v1/streams/payment-456/events
{
  "event_type": "PaymentInitiated",
  "data": {
    "payment_id": "456",
    "amount": 49.99,
    "currency": "USD",
    "customer_id": "cust-123"
  }
}

POST /api/v1/streams/payment-456/events
{
  "event_type": "PaymentCompleted",
  "data": {
    "payment_id": "456",
    "transaction_id": "txn-789",
    "status": "success"
  }
}
```

### 4. Time-Travel Debugging
Reconstruct state at any point in time:
```bash
# Get state as of event version 50
GET /api/v1/streams/user-123/events/range?from=0&to=50

# Replay events to recreate bug
POST /api/v1/streams/user-123/replay
{
  "from_version": 45,
  "to_version": 55
}
```

### 5. Event-Driven Analytics
Subscribe to events for real-time analytics:
```json
POST /api/v1/subscriptions
{
  "subscriber_id": "analytics-service",
  "stream_patterns": ["payment-*", "subscription-*"],
  "event_types": ["PaymentCompleted", "SubscriptionCreated"]
}
```

## Event Sourcing Benefits

### 1. Complete Audit Trail
- Every change is recorded with who, what, when, why
- Meets compliance requirements (SOC 2, GDPR, HIPAA)
- Forensic analysis capabilities

### 2. Time Travel
- Reconstruct state at any point in history
- Debug issues by replaying events
- "What if" scenario analysis

### 3. Event-Driven Architecture
- Services react to events in real-time
- Loose coupling between services
- Scalable, distributed processing

### 4. No Data Loss
- Append-only architecture prevents data loss
- Even "deleted" entities have full history
- Can always undo/redo operations

### 5. Business Intelligence
- Complete business event history
- Derive new insights from historical data
- Build new projections without data migration

## Event Sourcing Challenges

### 1. Event Schema Evolution
Events are immutable, but schemas need to evolve:
- **Solution**: Use event versioning and schema migration strategies
- Support multiple event versions simultaneously
- Transform old events to new format when reading

### 2. Storage Growth
Events accumulate over time:
- **Solution**: Use snapshots to reduce replay costs
- Archive old events to cold storage
- Implement event retention policies

### 3. Eventual Consistency
Event processing is asynchronous:
- **Solution**: Design UI for eventual consistency
- Provide optimistic UI updates
- Handle idempotency in event handlers

## Performance Considerations

- Events are stored in append-only tables (no updates/deletes)
- Database writes are sequential (fast)
- Indexes on stream_id, event_type, timestamp for fast queries
- Snapshots reduce state reconstruction time
- Connection pooling (25 max connections)
- Optional Redis caching for frequently accessed streams

## Security

- JWT authentication required for all write operations
- Multi-tenant isolation at database query level
- Events include user_id for access control
- Audit logs are tamper-proof (append-only)
- Database connections use SSL in production

## Troubleshooting

### Service Not Starting
- Check database connection parameters
- Verify PostgreSQL is running and accessible
- Ensure event tables exist (run migrations)

### Events Not Being Stored
- Verify JWT token is valid
- Check stream exists before appending events
- Review version conflicts (optimistic locking)

### Slow Event Queries
- Check database indexes
- Consider using snapshots for large streams
- Enable Redis caching

### Subscription Not Receiving Events
- Verify subscription is active
- Check stream patterns match event streams
- Review last_event_id cursor position

## Future Enhancements

- Event archival to S3/cold storage
- Compression for old events
- Event transformation pipelines
- Multi-region event replication
- GraphQL interface for event queries
- Event schema registry
- Automated snapshot management
- Event stream partitioning for scale
