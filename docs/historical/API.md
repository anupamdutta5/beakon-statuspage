# Status Page Platform API Documentation

## Overview

The Status Page Platform provides a comprehensive REST API for managing all aspects of your status page, from creating incidents to managing subscriptions and analytics.

## Base URL

```
https://your-domain.com/api/v1
```

## Authentication

The API uses JWT (JSON Web Token) authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Getting an API Token

```bash
curl -X POST https://your-domain.com/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T12:00:00Z"
}
```

## Rate Limiting

API requests are rate limited to prevent abuse:
- **Free Plan**: 100 requests per hour
- **Pro Plan**: 1,000 requests per hour
- **Enterprise Plan**: 10,000 requests per hour

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Error Handling

The API uses standard HTTP status codes and returns errors in the following format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional error details"
  }
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `429` - Too Many Requests
- `500` - Internal Server Error

## Core Endpoints

### Services

#### Get All Services
```http
GET /status
```

Response:
```json
{
  "services": [
    {
      "id": 1,
      "name": "API",
      "description": "Main API service",
      "status": "operational",
      "group": "Backend",
      "show_uptime": true,
      "position": 0,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "status": "operational"
}
```

#### Create Service
```http
POST /admin/status
Content-Type: application/json

{
  "name": "New Service",
  "description": "Service description",
  "group": "Backend",
  "show_uptime": true,
  "position": 1
}
```

#### Update Service
```http
PUT /admin/status/{id}
Content-Type: application/json

{
  "name": "Updated Service",
  "status": "degraded"
}
```

#### Delete Service
```http
DELETE /admin/status/{id}
```

### Incidents

#### Get All Incidents
```http
GET /incidents
```

Query Parameters:
- `status` - Filter by status (investigating, identified, monitoring, resolved)
- `impact` - Filter by impact (minor, major, critical)
- `limit` - Number of incidents to return (default: 50)
- `offset` - Number of incidents to skip (default: 0)

Response:
```json
{
  "incidents": [
    {
      "id": 1,
      "title": "API experiencing issues",
      "description": "We are investigating reports of API issues",
      "status": "investigating",
      "impact": "major",
      "resolved_at": null,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "updates": [
        {
          "id": 1,
          "status": "investigating",
          "message": "We are investigating the issue",
          "created_at": "2024-01-01T00:00:00Z"
        }
      ]
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

#### Create Incident
```http
POST /admin/incidents
Content-Type: application/json

{
  "title": "API experiencing issues",
  "description": "We are investigating reports of API issues",
  "impact": "major",
  "services": [1, 2]
}
```

#### Update Incident
```http
PUT /admin/incidents/{id}
Content-Type: application/json

{
  "status": "identified",
  "message": "We have identified the root cause"
}
```

#### Add Incident Update
```http
POST /admin/incidents/{id}/updates
Content-Type: application/json

{
  "status": "monitoring",
  "message": "We are monitoring the fix"
}
```

### Maintenance

#### Get Maintenance Events
```http
GET /maintenance
```

#### Create Maintenance Event
```http
POST /admin/maintenance
Content-Type: application/json

{
  "title": "Scheduled maintenance",
  "description": "Database maintenance window",
  "scheduled_start": "2024-01-01T02:00:00Z",
  "scheduled_end": "2024-01-01T04:00:00Z",
  "services": [1, 2]
}
```

### Subscribers

#### Subscribe to Updates
```http
POST /subscribers
Content-Type: application/json

{
  "email": "user@example.com",
  "services": [1, 2, 3]
}
```

#### Unsubscribe
```http
DELETE /subscribers/{id}
```

## SaaS Management API

### Tenants

#### Get All Tenants
```http
GET /admin/saas/tenants
```

Response:
```json
{
  "tenants": [
    {
      "id": 1,
      "name": "Acme Corp",
      "slug": "acme-corp",
      "domain": "status.acme.com",
      "status": "active",
      "plan": "pro",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

#### Create Tenant
```http
POST /admin/saas/tenants
Content-Type: application/json

{
  "name": "New Company",
  "slug": "new-company",
  "domain": "status.newcompany.com",
  "plan": "free"
}
```

#### Get Tenant Details
```http
GET /admin/saas/tenants/{id}
```

#### Update Tenant
```http
PUT /admin/saas/tenants/{id}
Content-Type: application/json

{
  "name": "Updated Company Name",
  "status": "suspended"
}
```

### Subscription Plans

#### Get All Plans
```http
GET /admin/saas/plans
```

Response:
```json
{
  "plans": [
    {
      "id": 1,
      "name": "Free",
      "slug": "free",
      "description": "Basic status page",
      "price": 0,
      "currency": "USD",
      "max_services": 5,
      "max_monitors": 10,
      "max_subscribers": 100,
      "custom_domain": false,
      "api_access": false,
      "integrations": false,
      "analytics": false,
      "white_label": false,
      "is_active": true
    }
  ]
}
```

#### Create Plan
```http
POST /admin/saas/plans
Content-Type: application/json

{
  "name": "Premium",
  "slug": "premium",
  "description": "Premium status page with advanced features",
  "price": 99,
  "currency": "USD",
  "max_services": 100,
  "max_monitors": 500,
  "max_subscribers": 10000,
  "custom_domain": true,
  "api_access": true,
  "integrations": true,
  "analytics": true,
  "white_label": true
}
```

### Subscriptions

#### Get Tenant Subscription
```http
GET /admin/saas/tenants/{id}/subscription
```

#### Create Subscription
```http
POST /admin/saas/tenants/{id}/subscription
Content-Type: application/json

{
  "plan_slug": "pro"
}
```

#### Upgrade Subscription
```http
PUT /admin/saas/tenants/{id}/subscription/upgrade
Content-Type: application/json

{
  "new_plan_slug": "enterprise"
}
```

#### Cancel Subscription
```http
POST /admin/saas/tenants/{id}/subscription/cancel
```

### Usage Analytics

#### Get Tenant Usage
```http
GET /admin/saas/tenants/{id}/usage
```

Response:
```json
{
  "usage": {
    "services": 3,
    "monitors": 8,
    "subscribers": 45,
    "incidents": 2,
    "maintenance": 1
  },
  "limits": {
    "services": 25,
    "monitors": 100,
    "subscribers": 1000
  }
}
```

## Monitoring API

### Health Checks

#### Get Service Health
```http
GET /admin/monitors/{id}/health
```

#### Create Health Check
```http
POST /admin/monitors
Content-Type: application/json

{
  "name": "API Health Check",
  "url": "https://api.example.com/health",
  "interval": 60,
  "timeout": 30,
  "service_id": 1
}
```

### Alerts

#### Get Alerts
```http
GET /admin/alerts
```

#### Create Alert
```http
POST /admin/alerts
Content-Type: application/json

{
  "name": "High Error Rate",
  "description": "Alert when error rate exceeds 5%",
  "query": "rate(http_requests_total{status=~\"5..\"}[5m])",
  "condition": "greater_than",
  "threshold": 0.05,
  "severity": "high",
  "create_incident": true
}
```

## Integrations API

### Webhooks

#### Get Webhooks
```http
GET /admin/integrations/webhooks
```

#### Create Webhook
```http
POST /admin/integrations/webhooks
Content-Type: application/json

{
  "name": "Slack Notifications",
  "url": "https://hooks.slack.com/services/...",
  "events": ["incident.created", "incident.resolved"],
  "secret": "webhook-secret"
}
```

### Third-Party Integrations

#### Get Integrations
```http
GET /admin/integrations
```

#### Create Integration
```http
POST /admin/integrations
Content-Type: application/json

{
  "type": "slack",
  "name": "Slack Integration",
  "webhook_url": "https://hooks.slack.com/services/...",
  "is_active": true
}
```

## Analytics API

### Page Analytics

#### Get Page Analytics
```http
GET /admin/analytics/pages?days=30
```

Response:
```json
{
  "total_views": 15420,
  "unique_visitors": 3240,
  "top_pages": [
    {
      "page": "/",
      "views": 12000
    }
  ],
  "daily_views": [
    {
      "date": "2024-01-01",
      "views": 500
    }
  ]
}
```

### Uptime Analytics

#### Get Uptime Analytics
```http
GET /admin/analytics/uptime?days=30
```

Response:
```json
{
  "overall_uptime": 99.9,
  "service_uptime": [
    {
      "service_id": 1,
      "service_name": "API",
      "uptime": 99.95
    }
  ],
  "daily_uptime": [
    {
      "date": "2024-01-01",
      "uptime": 100.0
    }
  ]
}
```

### Incident Analytics

#### Get Incident Analytics
```http
GET /admin/analytics/incidents?days=30
```

Response:
```json
{
  "total_incidents": 5,
  "avg_duration": 45.5,
  "incidents_by_severity": [
    {
      "severity": "major",
      "count": 3
    }
  ],
  "monthly_incidents": [
    {
      "month": "2024-01",
      "count": 5
    }
  ]
}
```

## Webhooks

### Webhook Events

The platform sends webhooks for the following events:

#### Incident Events
- `incident.created` - New incident created
- `incident.updated` - Incident status updated
- `incident.resolved` - Incident resolved

#### Maintenance Events
- `maintenance.scheduled` - Maintenance scheduled
- `maintenance.started` - Maintenance started
- `maintenance.completed` - Maintenance completed

#### Service Events
- `service.created` - New service created
- `service.updated` - Service updated
- `service.deleted` - Service deleted

### Webhook Payload

```json
{
  "event": "incident.created",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "incident": {
      "id": 1,
      "title": "API experiencing issues",
      "description": "We are investigating reports of API issues",
      "status": "investigating",
      "impact": "major",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### Webhook Security

Webhooks include a signature header for verification:

```
X-Webhook-Signature: sha256=abc123...
```

Verify the signature using your webhook secret:

```python
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

## SDKs and Libraries

### Official SDKs
- **Go**: `github.com/statuspage/go-sdk`
- **Python**: `pip install statuspage-sdk`
- **Node.js**: `npm install @statuspage/sdk`
- **PHP**: `composer require statuspage/sdk`

### Example Usage (Go)

```go
package main

import (
    "fmt"
    "github.com/statuspage/go-sdk"
)

func main() {
    client := statuspage.NewClient("your-api-token")
    
    // Create an incident
    incident, err := client.Incidents.Create(&statuspage.Incident{
        Title: "API experiencing issues",
        Description: "We are investigating reports of API issues",
        Impact: "major",
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Created incident: %s\n", incident.Title)
}
```

### Example Usage (Python)

```python
from statuspage import StatusPage

client = StatusPage(api_token="your-api-token")

# Create an incident
incident = client.incidents.create(
    title="API experiencing issues",
    description="We are investigating reports of API issues",
    impact="major"
)

print(f"Created incident: {incident.title}")
```

## Best Practices

### Authentication
- Store API tokens securely
- Rotate tokens regularly
- Use different tokens for different environments

### Rate Limiting
- Implement exponential backoff for rate limit errors
- Cache responses when appropriate
- Monitor your API usage

### Error Handling
- Always check HTTP status codes
- Implement retry logic for transient errors
- Log errors for debugging

### Webhooks
- Verify webhook signatures
- Implement idempotency for webhook handlers
- Handle duplicate webhook deliveries

## Support

For API support and questions:
- **Documentation**: [docs.your-domain.com](https://docs.your-domain.com)
- **Email**: [api-support@your-domain.com](mailto:api-support@your-domain.com)
- **GitHub Issues**: [github.com/your-org/statuspage-platform/issues](https://github.com/your-org/statuspage-platform/issues)
