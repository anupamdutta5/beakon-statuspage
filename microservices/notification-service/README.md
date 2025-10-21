# Notification Service

## Overview
The Notification Service manages multi-channel notification delivery for the Beakon status page platform. It handles sending notifications via email, SMS, webhooks, Slack, and other channels, with support for templates, subscriptions, and delivery tracking.

## Key Features

### Multi-Channel Delivery
- **Email Notifications**: SMTP and transactional email providers (SendGrid, Mailgun)
- **SMS Notifications**: Twilio, AWS SNS integration
- **Webhook Notifications**: HTTP POST to custom endpoints
- **Slack Integration**: Post to Slack channels and DMs
- **Microsoft Teams**: Post to Teams channels
- **PagerDuty**: Alert on-call engineers for critical incidents

### Notification Types
- **Incident Notifications**: Incident creation, updates, resolution
- **Maintenance Notifications**: Scheduled maintenance alerts
- **System Notifications**: Account changes, billing updates
- **Status Subscriptions**: Component status change notifications
- **Custom Notifications**: Arbitrary notification delivery

### Template Management
- **Template CRUD**: Create, read, update, delete notification templates
- **Variable Substitution**: Dynamic content via template variables
- **Multi-language Support**: Templates in multiple languages
- **HTML & Plain Text**: Support for both email formats

### Subscription Management
- **User Subscriptions**: Users subscribe to status updates
- **Component Subscriptions**: Subscribe to specific components
- **Incident Subscriptions**: Subscribe to specific incidents
- **Subscription Preferences**: Channel preferences per subscription
- **Unsubscribe Management**: One-click unsubscribe links

### Provider Management
- **Provider Configuration**: Configure multiple notification providers
- **Provider Testing**: Test provider configuration before use
- **Failover Support**: Automatic failover to backup providers
- **Provider Health Monitoring**: Track provider availability

### Delivery Tracking
- **Delivery Status**: Track sent, delivered, failed, bounced
- **Retry Logic**: Automatic retry with exponential backoff
- **Failed Notification Queue**: DLQ for failed deliveries
- **Delivery Metrics**: Success rate, latency, failure reasons

## API Endpoints

### Health & Monitoring
- `GET /health` - Service health check

### Notification Sending
- `POST /api/v1/notifications/send` - Send notification
- `POST /api/v1/notifications/maintenance` - Send maintenance notification
- `POST /api/v1/notifications/incident` - Send incident notification
- `GET /api/v1/notifications` - List notifications (legacy)

### Provider Management
- `GET /api/v1/providers` - List configured providers
- `POST /api/v1/providers/configure` - Configure provider
- `POST /api/v1/providers/test` - Test provider configuration

### Channel Management
- `GET /api/v1/channels` - List notification channels
- `POST /api/v1/channels` - Create notification channel

### Template Management
- `GET /api/v1/templates` - List notification templates
- `POST /api/v1/templates` - Create template
- `GET /api/v1/templates/:id` - Get template
- `PUT /api/v1/templates/:id` - Update template
- `DELETE /api/v1/templates/:id` - Delete template

### Subscription Management
- `GET /api/v1/subscriptions` - List subscriptions
- `POST /api/v1/subscriptions` - Create subscription
- `GET /api/v1/subscriptions/:id` - Get subscription
- `PUT /api/v1/subscriptions/:id` - Update subscription
- `DELETE /api/v1/subscriptions/:id` - Delete subscription (unsubscribe)

### Webhooks
- `POST /api/v1/webhook` - Handle provider webhooks (delivery status, bounces)

## Dependencies

### Internal Services
- **incident-service**: Incident notifications
- **tenant-admin-service**: Tenant notification settings
- **user-service**: User contact information

### External Dependencies
- **shared-resilience**: Common patterns, database, middleware
- **PostgreSQL**: Notification data storage
- **Email Providers**: SendGrid, Mailgun, SMTP
- **SMS Providers**: Twilio, AWS SNS
- **Slack API**: Slack integration
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

### Email Provider Configuration
- `SMTP_HOST`: SMTP server host
- `SMTP_PORT`: SMTP server port
- `SMTP_USER`: SMTP username
- `SMTP_PASSWORD`: SMTP password
- `SENDGRID_API_KEY`: SendGrid API key
- `MAILGUN_API_KEY`: Mailgun API key
- `MAILGUN_DOMAIN`: Mailgun domain

### SMS Provider Configuration
- `TWILIO_ACCOUNT_SID`: Twilio account SID
- `TWILIO_AUTH_TOKEN`: Twilio auth token
- `TWILIO_PHONE_NUMBER`: Twilio phone number
- `AWS_SNS_REGION`: AWS SNS region

### Integration Configuration
- `SLACK_BOT_TOKEN`: Slack bot token
- `PAGERDUTY_API_KEY`: PagerDuty API key

## Development

### Running the Service
```bash
cd microservices/notification-service
go run cmd/main.go
```

### Building
```bash
go build -o notification-service cmd/main.go
```

### Testing
```bash
go test ./...
```

### Project Structure
```
notification-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   │   ├── notification_handler.go
│   │   └── enhanced_notification_handler.go
│   ├── services/            # Business logic layer
│   │   ├── notification_service.go
│   │   └── enhanced_notification_service.go
│   ├── providers/           # Notification provider integrations
│   │   ├── email/
│   │   ├── sms/
│   │   ├── slack/
│   │   └── webhook/
│   ├── models/              # Data models
│   └── templates/           # Notification templates
└── go.mod                   # Go module definition
```

## Notification Schema

### Notification Request
```json
{
  "tenant_id": "uuid",
  "type": "incident|maintenance|system",
  "recipients": [
    {
      "email": "user@example.com",
      "phone": "+1234567890",
      "channels": ["email", "sms"]
    }
  ],
  "template_id": "uuid",
  "variables": {
    "incident_title": "Database Outage",
    "severity": "critical"
  },
  "channels": ["email", "sms", "webhook"]
}
```

### Template Schema
```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "Incident Created",
  "type": "incident",
  "channels": {
    "email": {
      "subject": "New Incident: {{incident_title}}",
      "html_body": "<p>Incident: {{incident_title}}</p>",
      "text_body": "Incident: {{incident_title}}"
    },
    "sms": {
      "body": "New incident: {{incident_title}}"
    }
  }
}
```

## Notification Channels

### Email
- SMTP for self-hosted
- SendGrid for transactional email
- Mailgun for high-volume
- HTML and plain text support
- Attachment support

### SMS
- Twilio for global SMS
- AWS SNS for AWS-native
- Character limits enforced
- Link shortening for URLs

### Webhook
- HTTP POST to custom endpoints
- Configurable headers and authentication
- Retry with exponential backoff
- Signature verification

### Slack
- Post to public/private channels
- Direct messages to users
- Rich message formatting
- Interactive buttons (future)

### PagerDuty
- Trigger incidents for critical alerts
- Auto-resolve when issue fixed
- On-call rotation integration

## Subscription System

### Subscription Types
- **Component Subscriptions**: Notify on component status changes
- **Incident Subscriptions**: Notify on incident updates
- **Maintenance Subscriptions**: Notify on scheduled maintenance
- **Status Page Subscriptions**: All updates for a status page

### Subscription Management
- Users subscribe via status page
- Email verification required
- Manage preferences (channels, frequency)
- One-click unsubscribe
- GDPR compliance

## Template Variables

### Common Variables
- `{{tenant_name}}` - Tenant company name
- `{{incident_title}}` - Incident title
- `{{incident_status}}` - Current incident status
- `{{component_name}}` - Component name
- `{{severity}}` - Incident severity
- `{{start_time}}` - Incident start time
- `{{end_time}}` - Incident resolution time
- `{{message}}` - Custom message

## Delivery Tracking

### Status Tracking
- **Queued**: Notification queued for sending
- **Sending**: Actively being sent
- **Sent**: Successfully handed to provider
- **Delivered**: Confirmed delivery (email opened, SMS received)
- **Failed**: Failed to send
- **Bounced**: Email bounced
- **Unsubscribed**: Recipient unsubscribed

### Retry Logic
- 3 retry attempts with exponential backoff
- Failed notifications moved to DLQ
- Manual retry option
- Alternative provider failover

## Monitoring

### Key Metrics
- Notifications sent per minute
- Delivery success rate
- Provider latency
- Failed delivery rate
- Bounce rate
- Unsubscribe rate

### Health Checks
- Database connectivity
- Provider connectivity
- Email delivery test
- SMS delivery test

### Alerting
- High failure rate
- Provider unavailable
- Queue backup
- Delivery delays

## Security

### Email Security
- SPF, DKIM, DMARC configuration
- Prevent email spoofing
- Sanitize HTML content
- Block malicious links

### Webhook Security
- Signature verification
- TLS/HTTPS required
- IP whitelisting
- Request timeouts

### Privacy
- Unsubscribe honored immediately
- Personal data encryption
- GDPR compliance
- Data retention policies

## Rate Limiting

### Provider Limits
- Respect provider rate limits
- Queuing for high volume
- Batch sending when possible
- Distributed rate limiting

### Anti-spam
- Limit notifications per user
- Cooldown periods
- Duplicate detection

## Best Practices

### Notification Design
- Clear, concise messages
- Mobile-friendly HTML
- Plain text fallback
- Actionable links

### Template Management
- Version control templates
- Test before deployment
- A/B test subject lines
- Monitor click-through rates

### Provider Management
- Configure multiple providers
- Test failover regularly
- Monitor provider status
- Review costs periodically

This service is essential for keeping users informed about incidents, maintenance, and status changes, ensuring transparent communication across the platform.
