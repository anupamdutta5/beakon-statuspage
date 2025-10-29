# RabbitMQ Event Flows

**Purpose**: Document all event-driven communication patterns in the Beakon platform
**Last Updated**: 2025-10-29

---

## Overview

The Beakon platform uses **RabbitMQ** for asynchronous, event-driven communication between services. This enables:
- Loose coupling between services
- Resilient communication (retry, dead-letter queues)
- Scalable async processing
- Eventual consistency patterns

---

## Architecture

### Event Flow Pattern

```
Publisher Service → RabbitMQ Exchange → Queue → Consumer Service → Database
                         ↓
                    Routing Key
```

### Components

**Exchanges**:
- Route messages based on routing keys
- Types: topic, direct, fanout

**Queues**:
- Store messages until consumed
- Durable (survive broker restart)
- Support dead-letter queues for failures

**Routing Keys**:
- Pattern-based message routing
- Format: `entity.action` (e.g., `tenant.created`, `incident.updated`)

---

## Event Flows

### 1. Tenant Creation Flow

**Trigger**: SaaS Admin creates new tenant
**Publisher**: saas-admin-service (8098)
**Consumer**: tenant-admin-service consumer
**Purpose**: Sync tenant data between saas_admin and tenant_admin_db databases

#### Flow Diagram

```
SaaS Admin UI (3001)
    ↓ POST /api/v1/tenants
saas-admin-service
    ↓ Save to saas_admin DB
    ↓ Publish Event
RabbitMQ Exchange: "tenant.events"
    ↓ Routing Key: "tenant.created"
Queue: "tenant_admin_queue"
    ↓
tenant-admin-service (consumer)
    ↓ Create tenant in tenant_admin_db
    ↓ Create admin user (if credentials provided)
    ↓ Create default settings
    ↓ Send acknowledgment
```

#### Event Structure

**Exchange**: `tenant.events` (topic)
**Routing Key**: `tenant.created`
**Queue**: `tenant_admin_queue`

**Message Payload**:
```json
{
  "event_id": "uuid",
  "event_type": "tenant.created",
  "timestamp": "2025-10-29T10:00:00Z",
  "data": {
    "id": "tenant-uuid",
    "name": "Acme Corp",
    "slug": "acme",
    "domain": "acme.com",
    "subdomain": "acme",
    "contact_email": "admin@acme.com",
    "billing_email": "billing@acme.com",
    "status": "active",
    "max_users": 50,
    "admin_email": "admin@acme.com",
    "admin_password": "hashed_password"
  },
  "metadata": {
    "source": "saas-admin-service",
    "correlation_id": "uuid"
  }
}
```

#### HTTP Fallback

If RabbitMQ is unavailable:
```
saas-admin-service → POST /api/v1/public/tenants → tenant-admin-service
```

**Circuit Breaker**: Automatic retries with exponential backoff via `shared-resilience`

---

### 2. Tenant Update Flow

**Trigger**: SaaS Admin updates tenant
**Publisher**: saas-admin-service (8098)
**Consumer**: tenant-admin-service consumer
**Purpose**: Keep tenant metadata in sync

#### Flow Diagram

```
SaaS Admin UI (3001)
    ↓ PUT /api/v1/tenants/:id
saas-admin-service
    ↓ Update saas_admin DB
    ↓ Publish Event
RabbitMQ Exchange: "tenant.events"
    ↓ Routing Key: "tenant.updated"
Queue: "tenant_admin_queue"
    ↓
tenant-admin-service (consumer)
    ↓ Update tenant in tenant_admin_db
    ↓ Preserve admin users (no changes)
    ↓ Send acknowledgment
```

**Event Structure**: Same as tenant.created, routing key: `tenant.updated`

---

### 3. Notification Flow

**Trigger**: Incident created/updated, component status change
**Publisher**: incident-service, monitoring-service, component-service
**Consumer**: notification-consumer
**Purpose**: Send multi-channel notifications (email, SMS, Slack, webhooks)

#### Flow Diagram

```
Incident Service (8086)
    ↓ Incident created/updated
    ↓ Publish Event
RabbitMQ Exchange: "notification.events"
    ↓ Routing Key: "notification.send"
Queue: "notification_queue"
    ↓
notification-consumer
    ↓ Load notification channels (email, SMS, Slack, webhook)
    ↓ Send via notification-service APIs
    ↓ Log delivery status
    ↓ Send acknowledgment
```

#### Event Structure

**Exchange**: `notification.events` (topic)
**Routing Key**: `notification.send`
**Queue**: `notification_queue`

**Message Payload**:
```json
{
  "event_id": "uuid",
  "event_type": "notification.send",
  "timestamp": "2025-10-29T10:00:00Z",
  "data": {
    "tenant_id": "uuid",
    "notification_type": "incident_created",
    "priority": "high",
    "subject": "Incident: Database Service Down",
    "message": "The database service has been detected as down...",
    "channels": ["email", "slack", "webhook"],
    "recipients": {
      "email": ["ops@acme.com"],
      "slack": ["#incidents"],
      "webhook": ["https://hooks.acme.com/incidents"]
    },
    "metadata": {
      "incident_id": "uuid",
      "component_id": "uuid",
      "severity": "critical"
    }
  }
}
```

---

### 4. Analytics Event Flow

**Trigger**: Service events (uptime checks, incidents, user actions)
**Publisher**: monitoring-service, incident-service, user-service, etc.
**Consumer**: analytics-consumer
**Purpose**: Aggregate metrics, calculate SLA, generate reports

#### Flow Diagram

```
Monitoring Service (8092)
    ↓ Health check completed
    ↓ Publish Event
RabbitMQ Exchange: "analytics.events"
    ↓ Routing Key: "analytics.uptime"
Queue: "analytics_queue"
    ↓
analytics-consumer
    ↓ Calculate uptime percentage
    ↓ Update SLA metrics
    ↓ Store in analytics DB
    ↓ Send acknowledgment
```

#### Event Structure

**Exchange**: `analytics.events` (topic)
**Routing Keys**:
- `analytics.uptime` - Uptime/health check events
- `analytics.incident` - Incident lifecycle events
- `analytics.usage` - User activity events

**Message Payload**:
```json
{
  "event_id": "uuid",
  "event_type": "analytics.uptime",
  "timestamp": "2025-10-29T10:00:00Z",
  "data": {
    "tenant_id": "uuid",
    "component_id": "uuid",
    "monitor_id": "uuid",
    "status": "up",
    "response_time_ms": 45,
    "location": "us-east-1",
    "check_type": "http"
  }
}
```

---

### 5. Audit Event Flow

**Trigger**: User actions, admin operations, security events
**Publisher**: All services
**Consumer**: audit-consumer
**Purpose**: Complete audit trail for compliance

#### Flow Diagram

```
Any Service
    ↓ User performs action
    ↓ Publish Event
RabbitMQ Exchange: "audit.events"
    ↓ Routing Key: "audit.user_action"
Queue: "audit_queue"
    ↓
audit-consumer
    ↓ Store in audit_logs table
    ↓ Check for suspicious activity
    ↓ Alert if needed
    ↓ Send acknowledgment
```

#### Event Structure

**Exchange**: `audit.events` (topic)
**Routing Keys**:
- `audit.user_action` - User-initiated actions
- `audit.admin_action` - Admin operations
- `audit.security_event` - Security-related events

**Message Payload**:
```json
{
  "event_id": "uuid",
  "event_type": "audit.user_action",
  "timestamp": "2025-10-29T10:00:00Z",
  "data": {
    "tenant_id": "uuid",
    "user_id": "uuid",
    "action": "component.delete",
    "resource_type": "component",
    "resource_id": "uuid",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "result": "success",
    "metadata": {
      "component_name": "API Gateway",
      "previous_status": "operational"
    }
  }
}
```

---

### 6. Billing Event Flow

**Trigger**: Subscription changes, usage events, payment processing
**Publisher**: payment-service, saas-admin-service
**Consumer**: billing-consumer
**Purpose**: Calculate bills, track usage, process payments

#### Flow Diagram

```
Payment Service (8088)
    ↓ Subscription created/updated
    ↓ Publish Event
RabbitMQ Exchange: "billing.events"
    ↓ Routing Key: "billing.subscription_changed"
Queue: "billing_queue"
    ↓
billing-consumer
    ↓ Calculate prorated charges
    ↓ Update invoice
    ↓ Process payment if needed
    ↓ Send acknowledgment
```

#### Event Structure

**Exchange**: `billing.events` (topic)
**Routing Keys**:
- `billing.subscription_changed` - Plan upgrades/downgrades
- `billing.usage_recorded` - Usage-based billing events
- `billing.payment_processed` - Payment confirmations

**Message Payload**:
```json
{
  "event_id": "uuid",
  "event_type": "billing.subscription_changed",
  "timestamp": "2025-10-29T10:00:00Z",
  "data": {
    "tenant_id": "uuid",
    "old_plan_id": "uuid",
    "new_plan_id": "uuid",
    "change_type": "upgrade",
    "effective_date": "2025-11-01",
    "prorated_amount": 125.50,
    "billing_cycle": "monthly"
  }
}
```

---

## Exchange & Queue Configuration

### Exchanges

| Exchange Name | Type | Durable | Purpose |
|--------------|------|---------|---------|
| tenant.events | topic | Yes | Tenant lifecycle events |
| notification.events | topic | Yes | Notification dispatch |
| analytics.events | topic | Yes | Metrics and analytics |
| audit.events | topic | Yes | Audit trail logging |
| billing.events | topic | Yes | Billing and payments |

### Queues

| Queue Name | Exchange | Routing Pattern | Consumer | Durable | DLQ |
|-----------|----------|----------------|----------|---------|-----|
| tenant_admin_queue | tenant.events | tenant.* | tenant-admin-service | Yes | tenant_admin_dlq |
| notification_queue | notification.events | notification.* | notification-consumer | Yes | notification_dlq |
| analytics_queue | analytics.events | analytics.* | analytics-consumer | Yes | analytics_dlq |
| audit_queue | audit.events | audit.* | audit-consumer | Yes | audit_dlq |
| billing_queue | billing.events | billing.* | billing-consumer | Yes | billing_dlq |

### Dead Letter Queues (DLQ)

**Purpose**: Store failed messages for manual review

**Configuration**:
- Messages moved to DLQ after 3 retry attempts
- TTL: 7 days (then auto-deleted)
- Manual reprocessing via admin tool

---

## Message Guarantees

### Publisher Confirms

All publishers use **confirm mode** to ensure messages are accepted by RabbitMQ:

```go
// Publisher waits for confirmation (2-second timeout)
if !publisher.WaitForConfirmation(2 * time.Second) {
    // Fall back to HTTP call
    fallbackToHTTP()
}
```

### Consumer Acknowledgments

All consumers use **manual acknowledgment**:

```go
// Process message
if err := processMessage(msg); err != nil {
    msg.Nack(false, true) // Requeue for retry
} else {
    msg.Ack(false) // Success
}
```

### Retry Strategy

**Retry Policy**:
1. Immediate retry (consumer requeue)
2. Retry after 1 minute (x-message-ttl)
3. Retry after 5 minutes
4. Move to DLQ after 3 failures

---

## Monitoring

### Metrics

**Per Queue**:
- `rabbitmq_queue_messages_ready` - Messages waiting
- `rabbitmq_queue_messages_unacknowledged` - Processing
- `rabbitmq_queue_messages_rate` - Throughput

**Per Consumer**:
- `consumer_messages_processed_total` - Success count
- `consumer_messages_failed_total` - Failure count
- `consumer_processing_duration_seconds` - Latency

### Alerts

**Critical**:
- Queue depth > 1000 messages (consumer lag)
- Consumer failure rate > 5%
- Message age > 5 minutes

**Warning**:
- Queue depth > 500 messages
- Consumer failure rate > 2%
- Processing latency > 2 seconds

---

## Best Practices

### Publisher Side

1. **Always use publisher confirms** - Wait for RabbitMQ acknowledgment
2. **Implement HTTP fallback** - Direct service calls if RabbitMQ unavailable
3. **Set correlation IDs** - Track events across services
4. **Include timestamps** - Detect stale events
5. **Validate payload** - Schema validation before publishing

### Consumer Side

1. **Use manual acknowledgment** - Only ack after successful processing
2. **Implement idempotency** - Handle duplicate messages gracefully
3. **Set processing timeout** - Prevent stuck consumers
4. **Log failures** - Include message ID and error details
5. **Monitor queue depth** - Alert on backlog buildup

### Message Design

1. **Keep payloads small** - < 10KB ideal
2. **Use JSON format** - Easy to debug and evolve
3. **Include event metadata** - Source, timestamp, correlation ID
4. **Version events** - Support schema evolution
5. **Avoid sensitive data** - Encrypt if necessary

---

## Troubleshooting

### Common Issues

**1. Messages stuck in queue**
```bash
# Check consumer status
rabbitmqctl list_consumers

# Check queue bindings
rabbitmqctl list_bindings

# Purge queue if needed (dev only)
rabbitmqctl purge_queue tenant_admin_queue
```

**2. Consumer not processing**
```bash
# Check consumer logs
docker logs billing-consumer

# Restart consumer
docker restart billing-consumer
```

**3. Dead Letter Queue growing**
```bash
# View DLQ messages
rabbitmqctl list_queues name messages | grep dlq

# Inspect failed message
# Use RabbitMQ Management UI to view message content
```

---

## Development

### Local Setup

```bash
# Start RabbitMQ with management UI
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=admin \
  -e RABBITMQ_DEFAULT_PASS=SecureP@ssw0rd2024! \
  rabbitmq:3-management

# Access UI: http://localhost:15672
# Login: admin / SecureP@ssw0rd2024!
```

### Testing Events

**Publish test event**:
```bash
# Using RabbitMQ Management API
curl -u admin:SecureP@ssw0rd2024! \
  -H "Content-Type: application/json" \
  -X POST http://localhost:15672/api/exchanges/%2F/tenant.events/publish \
  -d '{
    "properties": {},
    "routing_key": "tenant.created",
    "payload": "{\"event_id\":\"test-123\",\"event_type\":\"tenant.created\"}",
    "payload_encoding": "string"
  }'
```

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - Overall system architecture
- [SERVICE_CATALOG.md](SERVICE_CATALOG.md) - Service communication patterns
- [OPERATIONAL_RUNBOOK.md](../guides/OPERATIONAL_RUNBOOK.md) - Operations guide

---

**Last Updated**: 2025-10-29
**Maintained By**: Platform Team
