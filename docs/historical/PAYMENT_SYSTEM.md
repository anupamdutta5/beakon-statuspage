# Comprehensive Payment System Documentation

## Overview

This document provides comprehensive documentation for the full-fledged payment system implemented in the SaaS StatusPage platform. The system supports multiple payment gateways, UPI payments, card payments, invoicing, and follows industry best practices.

## Features

### ✅ Payment Gateways
- **Stripe**: Full support for cards, UPI, netbanking, and wallets
- **Razorpay**: Indian payment gateway with UPI, cards, netbanking, wallets, and EMI
- **PayU**: Indian payment gateway with comprehensive payment methods
- **PayPal**: International payment gateway for cards and PayPal payments

### ✅ Payment Methods
- **Card Payments**: Credit/Debit cards via all gateways
- **UPI Payments**: Unified Payments Interface for Indian customers
- **Net Banking**: Direct bank transfers
- **Digital Wallets**: Paytm, PhonePe, Google Pay, etc.
- **EMI**: Equated Monthly Installments (Razorpay)

### ✅ Advanced Features
- **Payment Retry Logic**: Automatic retry with exponential backoff
- **Webhook Processing**: Real-time payment event handling
- **Invoice Generation**: Automated invoice creation and PDF generation
- **Payment Analytics**: Comprehensive metrics and conversion funnels
- **Notification System**: Email and in-app notifications
- **Multi-currency Support**: USD, INR, and other currencies

## Architecture

### Core Components

1. **Payment Configuration System** (`internal/config/payment_config.go`)
   - Centralized configuration management
   - Environment-based gateway selection
   - Plug-and-play gateway integration

2. **Payment Service** (`internal/services/payment/service.go`)
   - Main payment orchestration
   - Gateway abstraction layer
   - Payment lifecycle management

3. **Gateway Implementations** (`internal/services/payment/gateways/`)
   - Stripe gateway (`stripe.go`)
   - Razorpay gateway (`razorpay.go`)
   - PayU gateway (`payu.go`)
   - PayPal gateway (`paypal.go`)

4. **Supporting Services**
   - Retry Service (`retry_service.go`)
   - Notification Service (`notification_service.go`)
   - Analytics Service (`analytics_service.go`)

5. **Invoicing System** (`internal/services/invoicing_service.go`)
   - Invoice generation and management
   - PDF generation
   - Email delivery

## Configuration

### Environment Variables

Create a `.env` file with the following variables:

```bash
# Default Payment Gateway
PAYMENT_DEFAULT_GATEWAY=stripe

# Default Currency
PAYMENT_CURRENCY=USD

# Stripe Configuration
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key_here
STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_stripe_webhook_secret_here
STRIPE_TEST_MODE=true

# Razorpay Configuration (for UPI and Indian payments)
RAZORPAY_KEY_ID=rzp_test_your_razorpay_key_id_here
RAZORPAY_KEY_SECRET=your_razorpay_key_secret_here
RAZORPAY_WEBHOOK_SECRET=your_razorpay_webhook_secret_here
RAZORPAY_TEST_MODE=true

# PayU Configuration (for Indian payments)
PAYU_MERCHANT_KEY=your_payu_merchant_key_here
PAYU_MERCHANT_SALT=your_payu_merchant_salt_here
PAYU_WEBHOOK_SECRET=your_payu_webhook_secret_here
PAYU_TEST_MODE=true

# PayPal Configuration
PAYPAL_CLIENT_ID=your_paypal_client_id_here
PAYPAL_CLIENT_SECRET=your_paypal_client_secret_here
PAYPAL_WEBHOOK_ID=your_paypal_webhook_id_here
PAYPAL_TEST_MODE=true
```

### Payment Configuration File

The system also supports a JSON configuration file (`configs/payment.json`):

```json
{
  "default_gateway": "stripe",
  "currency": "USD",
  "retry_settings": {
    "max_retries": 5,
    "retry_interval": 24,
    "backoff_factor": 1.5,
    "max_retry_days": 7
  },
  "gateways": [
    {
      "name": "stripe",
      "type": "stripe",
      "is_enabled": true,
      "is_default": true,
      "supported_methods": ["card", "upi", "netbanking", "wallet"]
    }
  ]
}
```

## API Endpoints

### Payment Management

#### Create Payment
```http
POST /api/v1/payments
Content-Type: application/json

{
  "amount": 29.99,
  "currency": "USD",
  "method": "card",
  "description": "Monthly subscription",
  "customer": {
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1234567890"
  },
  "return_url": "https://example.com/success",
  "webhook_url": "https://example.com/webhook"
}
```

#### Get Payment
```http
GET /api/v1/payments/{payment_id}
```

#### Cancel Payment
```http
POST /api/v1/payments/{payment_id}/cancel
```

#### Refund Payment
```http
POST /api/v1/payments/{payment_id}/refund
Content-Type: application/json

{
  "amount": 29.99,
  "reason": "Customer request"
}
```

### Subscription Management

#### Create Subscription
```http
POST /api/v1/subscriptions
Content-Type: application/json

{
  "amount": 29.99,
  "currency": "USD",
  "method": "card",
  "subscription": {
    "plan_slug": "pro",
    "billing_interval": "monthly",
    "trial_days": 14
  }
}
```

#### Get Subscription
```http
GET /api/v1/subscriptions/{subscription_id}
```

#### Cancel Subscription
```http
POST /api/v1/subscriptions/{subscription_id}/cancel
```

### Webhook Processing

#### Stripe Webhook
```http
POST /api/v1/webhooks/stripe
X-Stripe-Signature: t=1234567890,v1=signature
```

#### Razorpay Webhook
```http
POST /api/v1/webhooks/razorpay
X-Razorpay-Signature: signature
```

#### PayU Webhook
```http
POST /api/v1/webhooks/payu
```

#### PayPal Webhook
```http
POST /api/v1/webhooks/paypal
```

### Analytics and Reporting

#### Get Payment Metrics
```http
GET /api/v1/payments/metrics?period=month
```

#### Get Conversion Funnel
```http
GET /api/v1/payments/funnel?period=month
```

#### Get Payment History
```http
GET /api/v1/payments/history?limit=20&offset=0
```

#### Get Invoice History
```http
GET /api/v1/invoices/history?limit=20&offset=0
```

## Payment Flow

### 1. Payment Creation
1. Client sends payment request to API
2. System selects appropriate gateway based on payment method
3. Gateway creates payment session/order
4. Client redirects to gateway payment page
5. Customer completes payment

### 2. Webhook Processing
1. Gateway sends webhook to our system
2. System verifies webhook signature
3. System processes payment event
4. System updates payment status
5. System sends notifications
6. System records analytics

### 3. Retry Logic
1. Failed payment triggers retry check
2. System determines if payment is retryable
3. System schedules retry with exponential backoff
4. System processes retry at scheduled time
5. System updates retry status

## Security Features

### 1. Webhook Verification
- All webhooks are verified using gateway-specific signatures
- Invalid signatures are rejected immediately
- Webhook secrets are stored securely

### 2. Payment Data Protection
- No sensitive payment data is stored locally
- All payment data is handled by certified gateways
- PCI DSS compliance through gateway providers

### 3. API Security
- JWT-based authentication for all API endpoints
- Rate limiting on payment endpoints
- Input validation and sanitization

## Error Handling

### 1. Payment Failures
- Automatic retry for temporary failures
- Different handling for hard vs soft declines
- Comprehensive error logging and monitoring

### 2. Gateway Failures
- Fallback to alternative gateways
- Circuit breaker pattern for gateway health
- Graceful degradation of payment methods

### 3. Webhook Failures
- Retry mechanism for failed webhook processing
- Dead letter queue for unprocessable events
- Manual intervention capabilities

## Monitoring and Analytics

### 1. Payment Metrics
- Total payments and amounts
- Success/failure rates
- Conversion rates by payment method
- Revenue analytics

### 2. Performance Monitoring
- Payment processing times
- Gateway response times
- Webhook processing latency
- Error rates and patterns

### 3. Business Intelligence
- Payment method preferences
- Geographic payment patterns
- Seasonal payment trends
- Customer payment behavior

## Testing

### 1. Unit Tests
- Gateway implementation tests
- Payment service tests
- Webhook processing tests
- Retry logic tests

### 2. Integration Tests
- End-to-end payment flows
- Webhook processing flows
- Multi-gateway scenarios
- Error handling scenarios

### 3. Load Testing
- High-volume payment processing
- Webhook processing under load
- Gateway timeout scenarios
- Database performance under load

## Deployment

### 1. Environment Setup
1. Copy `env.payment.example` to `.env`
2. Configure gateway credentials
3. Set up webhook endpoints
4. Configure database connections

### 2. Database Migration
```bash
# Run database migrations
go run cmd/migrate/main.go up
```

### 3. Webhook Configuration
1. Configure webhook URLs in gateway dashboards
2. Set up SSL certificates for webhook endpoints
3. Test webhook delivery and processing

### 4. Monitoring Setup
1. Configure payment monitoring dashboards
2. Set up alerting for payment failures
3. Configure log aggregation
4. Set up performance monitoring

## Best Practices

### 1. Payment Processing
- Always verify webhook signatures
- Implement idempotency for payment operations
- Use proper error handling and logging
- Implement comprehensive retry logic

### 2. Security
- Never store sensitive payment data
- Use HTTPS for all payment communications
- Implement proper authentication and authorization
- Regular security audits and updates

### 3. Performance
- Implement caching for frequently accessed data
- Use connection pooling for database operations
- Implement proper timeout handling
- Monitor and optimize slow queries

### 4. Reliability
- Implement circuit breakers for external services
- Use proper error handling and recovery
- Implement comprehensive logging
- Regular backup and disaster recovery testing

## Troubleshooting

### Common Issues

1. **Payment Failures**
   - Check gateway credentials
   - Verify webhook configuration
   - Check payment method support
   - Review error logs

2. **Webhook Issues**
   - Verify webhook signatures
   - Check webhook URL accessibility
   - Review webhook processing logs
   - Test webhook delivery

3. **Gateway Connectivity**
   - Check network connectivity
   - Verify API credentials
   - Review gateway status pages
   - Check rate limiting

4. **Database Issues**
   - Check database connectivity
   - Review query performance
   - Check for deadlocks
   - Verify data consistency

### Debug Mode

Enable debug mode for detailed logging:

```bash
ENVIRONMENT=development
LOG_LEVEL=debug
```

## Support

For technical support and questions:

1. Check the logs for error details
2. Review the API documentation
3. Test with gateway test credentials
4. Contact the development team

## Changelog

### Version 1.0.0
- Initial implementation of comprehensive payment system
- Support for Stripe, Razorpay, PayU, and PayPal
- UPI payment support
- Invoice generation and management
- Payment analytics and reporting
- Webhook processing and retry logic
- Multi-currency support
