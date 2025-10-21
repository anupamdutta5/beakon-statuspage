# Billing Consumer

## Overview
The Billing Consumer is a financial event processing microservice that handles billing-related events from the message queue. It processes usage events, subscription changes, payment notifications, and invoice generation triggers for the platform's billing system.

## Key Features

### Event Processing
- **Usage Event Processing**: Track usage for metered billing
- **Subscription Events**: Handle subscription lifecycle events
- **Payment Events**: Process payment confirmations and failures
- **Invoice Generation**: Trigger invoice creation based on billing cycles

### Billing Event Types
- **Usage Events**: API calls, storage usage, bandwidth consumption
- **Subscription Events**: Created, upgraded, downgraded, cancelled, renewed
- **Payment Events**: Payment success, payment failure, refund
- **Invoice Events**: Invoice generated, invoice sent, invoice paid
- **Credit Events**: Credits added, credits used, credits expired

### Billing Operations
- **Usage Aggregation**: Aggregate usage metrics for billing calculations
- **Proration Calculations**: Calculate prorated charges for plan changes
- **Billing Cycle Management**: Track billing periods and due dates
- **Revenue Recognition**: Track revenue by accounting periods

### Integration Features
- **Payment Gateway Integration**: Process events from Stripe/PayPal
- **Accounting System Sync**: Send billing data to accounting systems
- **Tax Calculation**: Calculate taxes based on jurisdiction
- **Multi-currency Support**: Handle billing in different currencies

## Architecture

### Event Flow
1. Services publish billing events to message queue
2. Consumer reads events from billing topic
3. Events validated and processed
4. Usage aggregated and stored
5. Billing calculations performed
6. Invoices generated and sent to payment service

### Message Queue Integration
- Dedicated billing topic
- Consumer group for reliability
- Event ordering for consistent billing

## Configuration

### Environment Variables
- `KAFKA_BROKERS`: Kafka broker addresses
- `KAFKA_TOPIC`: Kafka topic for billing events
- `KAFKA_GROUP_ID`: Consumer group ID
- `DB_HOST`: Billing database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `ENVIRONMENT`: Runtime environment (development/production)
- `STRIPE_API_KEY`: Stripe API key for payment processing
- `PAYMENT_SERVICE_URL`: Payment service endpoint

## Dependencies

### Internal Services
- **payment-service**: For payment processing and invoice management
- **tenant-admin-service**: For tenant billing information

### External Dependencies
- **Message Queue**: Kafka/RabbitMQ for event streaming
- **PostgreSQL**: Billing data storage
- **Stripe/PayPal**: Payment gateway integration
- **Zap Logger**: Structured logging

## Development

### Running the Service
```bash
cd microservices/billing-consumer
go run cmd/main.go
```

### Building
```bash
go build -o billing-consumer cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
billing-consumer/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── consumer/            # Event consumer logic
│   ├── processor/           # Billing event processing
│   ├── calculator/          # Billing calculations
│   └── storage/             # Billing data storage
├── pkg/
│   └── logger/              # Logging utilities
└── go.mod                   # Go module definition
```

## Billing Event Schema

### Usage Event
```json
{
  "event_id": "uuid",
  "event_type": "usage",
  "tenant_id": "tenant_uuid",
  "timestamp": "ISO8601",
  "usage_type": "api_calls|storage|bandwidth",
  "quantity": 1000,
  "unit": "requests|GB|GB",
  "metadata": {
    "service": "analytics-service",
    "endpoint": "/api/v1/metrics"
  }
}
```

### Subscription Event
```json
{
  "event_id": "uuid",
  "event_type": "subscription",
  "action": "created|upgraded|downgraded|cancelled|renewed",
  "tenant_id": "tenant_uuid",
  "timestamp": "ISO8601",
  "subscription_id": "sub_uuid",
  "plan_id": "plan_uuid",
  "old_plan_id": "plan_uuid",
  "effective_date": "ISO8601",
  "metadata": {
    "proration": true,
    "reason": "User requested upgrade"
  }
}
```

### Payment Event
```json
{
  "event_id": "uuid",
  "event_type": "payment",
  "action": "success|failure|refund",
  "tenant_id": "tenant_uuid",
  "timestamp": "ISO8601",
  "payment_id": "pay_uuid",
  "invoice_id": "inv_uuid",
  "amount": 99.99,
  "currency": "USD",
  "payment_method": "stripe",
  "metadata": {
    "stripe_payment_intent": "pi_xxxxx"
  }
}
```

## Billing Calculations

### Usage-based Billing
```
Cost = (Usage Quantity / Unit Price) * Tier Rate
```

### Subscription Billing
```
Monthly Cost = Base Plan Price + Overages
Proration = (Days Remaining / Days in Month) * Plan Price
```

### Tax Calculation
```
Tax = Subtotal * Tax Rate (based on jurisdiction)
Total = Subtotal + Tax
```

## Monitoring

### Key Metrics
- Events processed per second
- Processing latency
- Failed billing calculations
- Invoice generation rate
- Revenue tracked
- Payment success rate

### Health Checks
Background service without HTTP endpoints. Health status logged periodically.

### Alerts
- Payment processing failures
- Invoice generation errors
- Usage aggregation delays
- Failed payment gateway calls

## Error Handling

### Processing Errors
- Invalid events sent to DLQ
- Failed payment events retry with backoff
- Critical billing events never discarded
- Duplicate event detection

### Data Integrity
- Transaction-based billing updates
- Idempotency for duplicate events
- Audit trail for all billing operations
- Reconciliation reports

## Compliance

### Financial Compliance
- **PCI DSS**: No credit card data stored
- **Revenue Recognition**: GAAP/IFRS compliant
- **Tax Compliance**: Automatic tax calculations
- **Audit Trail**: Complete billing history

### Data Retention
- Billing events retained for 7 years
- Usage data archived after billing cycle
- Payment records retained per regulations

## Scaling

### Horizontal Scaling
- Multiple consumer instances
- Partitioned by tenant for parallelism
- Distributed usage aggregation

### Performance
- Batch processing for usage aggregation
- Cached pricing information
- Optimized database queries
- Background invoice generation

## Integration Points

### Payment Gateway
- Webhook event processing
- Payment intent creation
- Subscription management
- Refund processing

### Accounting Systems
- Revenue journal entries
- Account receivable updates
- Tax reporting
- Financial statements

This consumer service is essential for accurate billing, revenue tracking, and financial operations across the platform. It ensures all usage is properly metered and billed according to subscription plans.
