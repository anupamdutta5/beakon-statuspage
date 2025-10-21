# Analytics Consumer

## Overview
The Analytics Consumer is an event-driven microservice that processes analytics events from the message queue system. It consumes events related to user activity, system metrics, and status page interactions, then processes and stores them for analysis and reporting.

## Key Features

### Event Processing
- **Real-time Analytics**: Processes analytics events as they occur
- **Event Aggregation**: Aggregates raw events into meaningful metrics
- **Data Enrichment**: Enriches events with additional context and metadata
- **Batch Processing**: Supports batch processing for high-volume events

### Analytics Events
- **User Activity Events**: Login, logout, page views, interactions
- **Status Page Events**: Incident views, component views, subscription events
- **System Metrics**: Performance metrics, API usage, error rates
- **Custom Events**: Extensible event schema for custom analytics

### Data Storage
- **Time-series Data**: Stores metrics in time-series format for analysis
- **Event Archive**: Archives raw events for historical analysis
- **Metric Rollup**: Automatically rolls up data for long-term storage

### Resilience Features
- **Graceful Shutdown**: Properly handles shutdown signals
- **Message Acknowledgment**: Ensures no message loss during processing
- **Error Recovery**: Handles processing errors with retry logic
- **Dead Letter Queue**: Failed messages moved to DLQ for manual review

## Architecture

### Event Flow
1. Analytics events published to message queue by other services
2. Consumer reads events from queue
3. Events validated and enriched with additional data
4. Processed events stored in analytics database
5. Metrics aggregated for reporting

### Message Queue Integration
- Uses Kafka/RabbitMQ for event streaming
- Consumer group for horizontal scaling
- Topic subscription with filtering

## Configuration

### Environment Variables
- `KAFKA_BROKERS`: Kafka broker addresses
- `KAFKA_TOPIC`: Kafka topic for analytics events
- `KAFKA_GROUP_ID`: Consumer group ID
- `DB_HOST`: Analytics database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `ENVIRONMENT`: Runtime environment (development/production)

## Dependencies

### Internal Services
- **analytics-service**: Shares analytics database and data models
- **event-store-service**: Source of analytics events

### External Dependencies
- **shared-resilience**: Common resilience patterns and utilities
- **Message Queue**: Kafka/RabbitMQ for event streaming
- **PostgreSQL**: Analytics data storage
- **Zap Logger**: Structured logging

## Development

### Running the Service
```bash
cd microservices/analytics-consumer
go run cmd/main.go
```

### Building
```bash
go build -o analytics-consumer cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
analytics-consumer/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── consumer/            # Event consumer logic
│   ├── processor/           # Event processing logic
│   └── storage/             # Data storage layer
├── pkg/
│   └── logger/              # Logging utilities
└── go.mod                   # Go module definition
```

## Event Schema

### Analytics Event Structure
```json
{
  "event_id": "uuid",
  "event_type": "page_view|user_action|system_metric",
  "tenant_id": "tenant_uuid",
  "user_id": "user_uuid",
  "timestamp": "ISO8601",
  "data": {
    "page": "/status",
    "duration_ms": 150,
    "metadata": {}
  }
}
```

## Monitoring

### Key Metrics
- Events processed per second
- Processing latency
- Error rate
- Queue lag
- Database write performance

### Health Checks
The consumer logs health status but doesn't expose HTTP endpoints as it's a background service.

## Scaling

### Horizontal Scaling
- Deploy multiple instances in same consumer group
- Kafka/RabbitMQ automatically distributes partitions
- Each instance processes a subset of events

### Performance Tuning
- Adjust batch size for database writes
- Configure consumer fetch size
- Tune database connection pool

## Error Handling

### Processing Errors
- Validation errors logged and sent to DLQ
- Database errors trigger retry with exponential backoff
- Unrecoverable errors logged and message acknowledged

### Recovery
- Consumer maintains offset for crash recovery
- Dead letter queue for failed messages
- Monitoring alerts for high error rates

This consumer service is critical for collecting analytics data across the platform, enabling data-driven insights and reporting for all tenants.
