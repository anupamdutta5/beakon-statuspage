# Audit Consumer

## Overview
The Audit Consumer is a security-focused microservice that processes audit events from the message queue system. It consumes events related to security-sensitive operations, compliance activities, and administrative actions, then stores them for compliance reporting and security analysis.

## Key Features

### Audit Event Processing
- **Real-time Audit Logging**: Processes security events as they occur
- **Compliance Tracking**: Tracks events for regulatory compliance (SOC2, GDPR, HIPAA)
- **Security Monitoring**: Detects suspicious patterns and security violations
- **Event Archival**: Long-term storage of audit trails

### Audit Event Types
- **Authentication Events**: Login, logout, password changes, MFA events
- **Authorization Events**: Permission changes, role assignments, access denials
- **Data Access Events**: Read/write operations on sensitive data
- **Administrative Events**: Configuration changes, user management, system settings
- **RBAC Events**: Role and permission modifications
- **Tenant Events**: Tenant creation, modification, deletion

### Compliance Features
- **Tamper-proof Storage**: Immutable audit log storage
- **Event Integrity**: Cryptographic hashing for event verification
- **Retention Policies**: Configurable retention based on compliance requirements
- **Audit Trail Reports**: Generate compliance audit reports

### Security Features
- **Anomaly Detection**: Identifies unusual patterns in audit events
- **Alert Generation**: Triggers alerts for critical security events
- **Data Encryption**: Encrypted storage for sensitive audit data
- **Access Control**: Strict access controls on audit log retrieval

## Architecture

### Event Flow
1. Services publish audit events to message queue
2. Consumer reads events from audit topic
3. Events validated and enriched with context
4. Events stored in immutable audit database
5. Security analysis performed on event patterns
6. Alerts generated for suspicious activities

### Message Queue Integration
- Dedicated audit topic for security events
- Consumer group for reliability
- Event replay capability for compliance

## Configuration

### Environment Variables
- `KAFKA_BROKERS`: Kafka broker addresses
- `KAFKA_TOPIC`: Kafka topic for audit events (default: audit-events)
- `KAFKA_GROUP_ID`: Consumer group ID
- `DB_HOST`: Audit database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Audit database name
- `ENVIRONMENT`: Runtime environment (development/production)
- `RETENTION_DAYS`: Audit log retention period (default: 2555 days / 7 years)

## Dependencies

### External Dependencies
- **Message Queue**: Kafka/RabbitMQ for event streaming
- **PostgreSQL**: Audit log storage with append-only tables
- **Zap Logger**: Structured logging

## Development

### Running the Service
```bash
cd microservices/audit-consumer
go run cmd/main.go
```

### Building
```bash
go build -o audit-consumer cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
audit-consumer/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── consumer/            # Event consumer logic
│   ├── processor/           # Audit event processing
│   ├── analyzer/            # Security analysis
│   └── storage/             # Audit log storage
├── pkg/
│   └── logger/              # Logging utilities
└── go.mod                   # Go module definition
```

## Audit Event Schema

### Audit Event Structure
```json
{
  "event_id": "uuid",
  "event_type": "authentication|authorization|data_access|admin|rbac",
  "action": "login|create|update|delete|read",
  "tenant_id": "tenant_uuid",
  "user_id": "user_uuid",
  "resource_type": "user|role|tenant|component",
  "resource_id": "resource_uuid",
  "timestamp": "ISO8601",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "success": true,
  "metadata": {
    "changes": {},
    "reason": "User initiated action"
  },
  "checksum": "sha256_hash"
}
```

## Compliance Standards

### SOC 2 Compliance
- All security-relevant events logged
- Immutable audit trails
- Regular audit trail reviews
- Access logging for all sensitive data

### GDPR Compliance
- Personal data access logged
- Data modification tracked
- Right to be forgotten events
- Data export events tracked

### HIPAA Compliance
- PHI access logged
- Authentication and authorization tracked
- Security incidents recorded
- Administrative actions logged

## Security Analysis

### Anomaly Detection
- **Failed Login Attempts**: Multiple failed logins from same IP
- **Privilege Escalation**: Unusual permission changes
- **Data Exfiltration**: Large data export operations
- **After-hours Access**: Access outside business hours
- **Unusual Locations**: Access from unexpected geographic locations

### Alert Triggers
- 5+ failed login attempts in 5 minutes
- Role/permission changes to admin accounts
- Bulk data export operations
- Access from blacklisted IPs
- Disabled security features

## Data Retention

### Default Retention Policy
- **Security Events**: 7 years (2555 days)
- **Authentication Events**: 1 year
- **Administrative Events**: 7 years
- **Data Access Events**: 3 years

### Archival Process
- Old logs archived to cold storage
- Compressed and encrypted archives
- Retrievable for compliance audits

## Monitoring

### Key Metrics
- Events processed per second
- Processing latency
- Storage growth rate
- Alert frequency
- Anomaly detection rate

### Health Checks
Background service without HTTP endpoints. Health status logged periodically.

## Error Handling

### Processing Errors
- Validation errors logged to error topic
- Critical events never discarded
- Failed events moved to DLQ for review
- Automatic retry for transient failures

### Data Integrity
- Event checksums verified
- Database constraints prevent tampering
- Append-only log tables
- Regular integrity audits

## Scaling

### Horizontal Scaling
- Multiple consumer instances in same group
- Partitioned by tenant for parallelism
- High availability configuration

### Performance
- Optimized database schema for time-series data
- Batch inserts for high throughput
- Indexed queries for compliance reporting

This consumer service is critical for maintaining security, compliance, and regulatory requirements across the entire platform. It provides an immutable audit trail for all security-relevant operations.
