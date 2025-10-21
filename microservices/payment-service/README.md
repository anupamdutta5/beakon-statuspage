# Payment Service

## Overview
The Payment Service handles all payment processing, subscription management, invoicing, and billing operations for the Beakon status page platform. It integrates with payment gateways (Stripe, PayPal) to process payments and manage the complete billing lifecycle.

## Key Features

### Payment Processing
- **Payment Intent Creation**: Initiate payment transactions
- **Payment Execution**: Process credit card, ACH, and alternative payments
- **Payment Confirmation**: Verify and confirm successful payments
- **Payment Refunds**: Process full and partial refunds
- **Payment History**: Track all payment transactions

### Subscription Management
- **Subscription Creation**: Initialize subscriptions for tenants
- **Plan Changes**: Upgrade, downgrade subscription plans
- **Subscription Cancellation**: Cancel subscriptions with proration
- **Trial Management**: Free trial activation and conversion
- **Renewal Processing**: Automatic subscription renewals

### Invoice Management
- **Invoice Generation**: Automatically generate invoices
- **Invoice Delivery**: Send invoices via email
- **Invoice Payment**: Track invoice payment status
- **Prorated Invoices**: Handle mid-cycle plan changes
- **Invoice History**: Access to all past invoices

### Payment Method Management
- **Add Payment Method**: Store credit cards, bank accounts
- **Update Payment Method**: Change default payment method
- **Remove Payment Method**: Delete stored payment methods
- **Payment Method Validation**: Verify payment method validity

### Billing Integration
- **Stripe Integration**: Primary payment gateway
- **PayPal Integration**: Alternative payment option
- **Webhook Processing**: Handle payment gateway webhooks
- **Tax Calculation**: Automatic tax calculations by jurisdiction
- **Multi-currency Support**: Process payments in different currencies

### Revenue Tracking
- **Revenue Recognition**: Track revenue by accounting period
- **Churn Analysis**: Monitor subscription cancellations
- **MRR/ARR Tracking**: Monthly and annual recurring revenue
- **Payment Analytics**: Success rates, failure analysis

## API Endpoints

### Health & Monitoring
- `GET /health` - Service health check

### Payment Management
- `GET /api/v1/payments` - List payments
- `POST /api/v1/payments` - Create payment
- `GET /api/v1/payments/:id` - Get payment details
- `POST /api/v1/payments/:id/refund` - Refund payment
- `POST /api/v1/payments/:id/capture` - Capture authorized payment

### Subscription Management
- `GET /api/v1/subscriptions` - List subscriptions
- `POST /api/v1/subscriptions` - Create subscription
- `GET /api/v1/subscriptions/:id` - Get subscription details
- `PUT /api/v1/subscriptions/:id` - Update subscription
- `DELETE /api/v1/subscriptions/:id` - Cancel subscription
- `POST /api/v1/subscriptions/:id/pause` - Pause subscription
- `POST /api/v1/subscriptions/:id/resume` - Resume subscription

### Invoice Management
- `GET /api/v1/invoices` - List invoices
- `GET /api/v1/invoices/:id` - Get invoice details
- `POST /api/v1/invoices/:id/pay` - Pay invoice
- `GET /api/v1/invoices/:id/download` - Download invoice PDF

### Payment Method Management
- `GET /api/v1/payment-methods` - List payment methods
- `POST /api/v1/payment-methods` - Add payment method
- `PUT /api/v1/payment-methods/:id` - Update payment method
- `DELETE /api/v1/payment-methods/:id` - Remove payment method
- `POST /api/v1/payment-methods/:id/set-default` - Set default payment method

### Pricing & Plans
- `GET /api/v1/plans` - List available plans
- `GET /api/v1/plans/:id` - Get plan details
- `POST /api/v1/pricing/calculate` - Calculate pricing for plan

### Webhooks
- `POST /api/v1/webhooks/stripe` - Stripe webhook endpoint
- `POST /api/v1/webhooks/paypal` - PayPal webhook endpoint

## Dependencies

### Internal Services
- **tenant-admin-service**: Tenant billing information
- **billing-consumer**: Consumes billing events

### External Dependencies
- **shared-resilience**: Common patterns, database, middleware
- **PostgreSQL**: Payment data storage
- **Stripe**: Primary payment gateway
- **PayPal**: Alternative payment gateway
- **Gin**: HTTP web framework

## Configuration

### Environment Variables
- `SERVER_PORT`: HTTP server port (default: 8080)
- `SERVER_HOST`: HTTP server host
- `ENVIRONMENT`: Runtime environment (development/production)
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name

### Payment Gateway Configuration
- `STRIPE_SECRET_KEY`: Stripe secret API key
- `STRIPE_PUBLISHABLE_KEY`: Stripe publishable key
- `STRIPE_WEBHOOK_SECRET`: Stripe webhook signing secret
- `PAYPAL_CLIENT_ID`: PayPal client ID
- `PAYPAL_SECRET`: PayPal secret
- `PAYPAL_WEBHOOK_ID`: PayPal webhook ID

### Billing Configuration
- `DEFAULT_CURRENCY`: Default currency (USD)
- `TAX_RATE`: Default tax rate
- `INVOICE_DUE_DAYS`: Invoice payment due days (default: 30)

## Development

### Running the Service
```bash
cd microservices/payment-service
go run cmd/main.go
```

### Building
```bash
go build -o payment-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
payment-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic layer
│   ├── gateways/            # Payment gateway integrations
│   │   ├── stripe/
│   │   └── paypal/
│   ├── models/              # Data models
│   └── webhooks/            # Webhook processors
└── go.mod                   # Go module definition
```

## Payment Flow

### New Subscription Flow
1. User selects plan on landing page
2. User enters payment information
3. Frontend calls `POST /api/v1/payment-methods` to tokenize card
4. Frontend calls `POST /api/v1/subscriptions` with payment method
5. Service creates Stripe subscription
6. Service stores subscription in database
7. Service publishes subscription event
8. Billing consumer processes event
9. Invoice generated and sent

### Payment Webhook Flow
1. Payment gateway sends webhook to `/api/v1/webhooks/stripe`
2. Service verifies webhook signature
3. Service processes event (payment success, failure, refund)
4. Service updates payment/subscription status
5. Service publishes event to message queue
6. Notification sent to tenant

## Subscription Schema

### Subscription Model
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "plan_id": "plan_uuid",
  "status": "active|paused|cancelled|past_due",
  "payment_method_id": "pm_uuid",
  "current_period_start": "ISO8601",
  "current_period_end": "ISO8601",
  "cancel_at_period_end": false,
  "trial_end": "ISO8601",
  "stripe_subscription_id": "sub_xxxxx",
  "created_at": "ISO8601"
}
```

### Payment Model
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "invoice_id": "inv_uuid",
  "amount": 99.99,
  "currency": "USD",
  "status": "pending|succeeded|failed|refunded",
  "payment_method": "card|ach|paypal",
  "stripe_payment_intent_id": "pi_xxxxx",
  "created_at": "ISO8601"
}
```

## Pricing Plans

### Plan Tiers
- **Free**: Limited features, community support
- **Starter**: $29/month, basic features
- **Professional**: $99/month, advanced features
- **Enterprise**: Custom pricing, all features

### Pricing Structure
- Base plan price
- Per-user pricing (if applicable)
- Usage-based pricing (API calls, storage)
- Add-on pricing (premium features)

## Tax Calculation

### Tax Configuration
- Automatic tax calculation by jurisdiction
- Support for VAT, GST, sales tax
- Tax exemption handling
- Tax ID validation

### Tax Integration
- Stripe Tax for automatic calculation
- TaxJar integration (optional)
- Manual tax rate configuration

## Webhook Events

### Stripe Webhooks
- `payment_intent.succeeded` - Payment successful
- `payment_intent.payment_failed` - Payment failed
- `customer.subscription.created` - Subscription created
- `customer.subscription.updated` - Subscription updated
- `customer.subscription.deleted` - Subscription cancelled
- `invoice.payment_succeeded` - Invoice paid
- `invoice.payment_failed` - Invoice payment failed

### Event Processing
- Webhook signature verification
- Idempotent event processing
- Event retry handling
- Dead letter queue for failed events

## Security

### PCI Compliance
- **No card storage**: Tokenization via Stripe/PayPal
- **PCI DSS Level 1**: Stripe handles card data
- **Secure communication**: TLS/HTTPS only
- **Access logging**: All payment access logged

### Payment Security
- Webhook signature verification
- API key rotation
- Encrypted sensitive data
- Fraud detection integration

## Monitoring

### Key Metrics
- Payment success rate
- Payment failure rate
- Revenue per day/month
- Subscription churn rate
- Average revenue per user (ARPU)
- Failed payment recovery rate

### Health Checks
- Database connectivity
- Stripe API connectivity
- PayPal API connectivity
- Webhook processing status

### Alerting
- Failed payment spike
- Subscription cancellation spike
- Gateway API errors
- Webhook processing delays

## Error Handling

### Payment Failures
- Insufficient funds
- Expired card
- Card declined
- Authentication required (3D Secure)

### Retry Logic
- Failed payments retry after 3, 7, 14 days
- Subscription status updated to past_due
- Notifications sent to tenant
- Subscription cancelled after final retry

## Refund Policy

### Refund Processing
- Full refunds within 30 days
- Partial refunds for partial service
- Prorated refunds on cancellation
- Refund to original payment method

## Revenue Recognition

### Accounting
- Deferred revenue for prepaid subscriptions
- Revenue recognized over subscription period
- Prorated revenue for plan changes
- Refund accounting

This service is critical for monetizing the platform, ensuring reliable payment processing, and maintaining accurate billing and financial records.
