# Dynamic Pricing Management System - Implementation Guide

## Overview

This document describes the complete dynamic pricing management system that allows full control of landing page pricing plans from the SaaS Admin dashboard.

## Architecture

### System Flow

```
┌─────────────────────────┐
│  SaaS Admin Dashboard   │
│  (Port: 8097)           │
│                         │
│  - Create/Edit Plans    │
│  - Set Popular Flag     │
│  - Manage Features      │
│  - Set Discounts        │
│  - Set Display Order    │
└───────────┬─────────────┘
            │
            │ Manual/Auto Sync
            ▼
┌─────────────────────────┐
│  SaaS Admin Service     │
│  Database               │
│  (saas_admin)           │
│                         │
│  - saas_plans           │
│  - pricing_tiers        │
└───────────┬─────────────┘
            │
            │ GET /api/v1/public/pricing-plans
            │ (Service-to-Service)
            ▼
┌─────────────────────────┐
│  Landing Page Service   │
│  (Port: 8100)           │
│                         │
│  POST /api/v1/admin/    │
│       pricing/sync      │
└───────────┬─────────────┘
            │
            │ Atomic Transaction
            ▼
┌─────────────────────────┐
│  Landing Page Service   │
│  Database               │
│  (statuspage_landing)   │
│                         │
│  - pricing_plans        │
└───────────┬─────────────┘
            │
            │ GET /api/v1/pricing
            ▼
┌─────────────────────────┐
│  Landing Page Frontend  │
│  http://localhost:8100  │
│                         │
│  Displays Pricing Cards │
└─────────────────────────┘
```

## Implementation Details

### 1. SaaS Admin Service

#### Models ([saas-admin-service/internal/models/saas_admin.go](microservices/saas-admin-service/internal/models/saas_admin.go))

**SaaSPlan** - Main pricing plan model:
- `IsPopular bool` - Mark as popular/featured plan
- `DisplayOrder int` - Control display order on landing page
- `ButtonText string` - Customize CTA button text
- `ButtonURL string` - Customize CTA button link
- `Features string` - JSON array of feature descriptions
- `IsActive bool` - Enable/disable plan
- `IsPublic bool` - Show on public landing page

**PricingTier** - Supports multiple billing intervals:
- `BillingInterval string` - "monthly" or "yearly"
- `Price float64` - Price for this interval
- `DiscountPercent float64` - Discount for yearly plans
- `IsActive bool` - Enable/disable tier

#### Service Layer ([saas-admin-service/internal/services/saas_admin_service.go](microservices/saas-admin-service/internal/services/saas_admin_service.go#L255-L279))

**`ListPublicPlans()`** - Retrieves active public plans:
```go
// Get all active public plans, ordered by display_order
plans := db.Preload("PricingTiers").
    Where("is_active = ? AND is_public = ?", true, true).
    Order("display_order ASC, created_at ASC").
    Find(&plans)
```

#### API Endpoints ([saas-admin-service/internal/handlers/saas_admin_handler.go](microservices/saas-admin-service/internal/handlers/saas_admin_handler.go#L883-L1015))

**`GET /api/v1/public/pricing-plans`** - Public endpoint for landing page sync:
- Returns all active public plans with pricing tiers
- Calculates monthly vs yearly pricing
- Extracts discount percentages
- Formats features as JSON arrays
- Response format:
```json
{
  "success": true,
  "plans": [
    {
      "plan_id": "uuid",
      "name": "Professional",
      "slug": "professional",
      "description": "Perfect for growing teams",
      "price": 49.00,
      "yearly_price": 490.00,
      "discount_percent": 16.67,
      "currency": "USD",
      "features": ["Unlimited users", "Advanced analytics", "Priority support"],
      "is_popular": true,
      "button_text": "Get Started",
      "button_url": "/signup?plan=professional",
      "display_order": 2
    }
  ],
  "count": 3
}
```

**`POST /api/v1/admin/pricing/sync-to-landing`** - Triggers sync to landing page:
- Calls landing page service `/api/v1/admin/pricing/sync`
- Uses service-to-service authentication
- Returns sync status

### 2. Landing Page Service

#### Pricing Handler ([landing-page-service/internal/handlers/pricing_handler.go](microservices/landing-page-service/internal/handlers/pricing_handler.go))

**`SyncPricingPlans()`** - Syncs pricing from SaaS Admin:
- Calls SaaS Admin Service public pricing API
- Uses **atomic transactions** for data integrity
- Creates new plans or updates existing ones
- Returns detailed sync statistics
- Rollback on any errors

**`GetPricingPlans()`** - Returns active pricing for frontend:
- Filters by `is_active = true` and `status = 'active'`
- Orders by `sort_order ASC`
- Parses JSON features into array
- Frontend-friendly format

#### Routes ([landing-page-service/internal/server/server.go](microservices/landing-page-service/internal/server/server.go#L198-L205))

**Public API:**
- `GET /api/v1/pricing` - Frontend fetches pricing plans

**Admin API:**
- `POST /api/v1/admin/pricing/sync` - Trigger sync from SaaS Admin

## How to Use

### Option 1: Using Existing Plan Management API

You can manage pricing using the existing SaaS Admin API endpoints:

#### 1. Create a New Plan

```bash
curl -X POST http://localhost:8097/api/v1/admin/plans \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=YOUR_SESSION_ID" \
  -d '{
    "name": "Professional",
    "slug": "professional",
    "description": "Perfect for growing teams",
    "price": 49.00,
    "currency": "USD",
    "billing_interval": "monthly",
    "features": "[\"Unlimited users\", \"Advanced analytics\", \"Priority support\", \"Custom integrations\"]",
    "is_popular": true,
    "is_active": true,
    "is_public": true,
    "button_text": "Get Started",
    "button_url": "/signup?plan=professional",
    "display_order": 2
  }'
```

#### 2. Update an Existing Plan

```bash
curl -X PUT http://localhost:8097/api/v1/admin/plans/{plan_id} \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=YOUR_SESSION_ID" \
  -d '{
    "is_popular": true,
    "price": 59.00,
    "features": "[\"Unlimited users\", \"Advanced analytics\", \"24/7 support\"]"
  }'
```

#### 3. Create Pricing Tier (for Yearly Pricing with Discount)

```bash
curl -X POST http://localhost:8097/api/v1/admin/pricing-tiers \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=YOUR_SESSION_ID" \
  -d '{
    "plan_id": "PLAN_UUID",
    "billing_interval": "yearly",
    "price": 490.00,
    "discount_percent": 16.67,
    "is_active": true
  }'
```

#### 4. Sync to Landing Page

```bash
curl -X POST http://localhost:8100/api/v1/admin/pricing/sync \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer service-token-YOUR_JWT_SECRET"
```

Response:
```json
{
  "success": true,
  "message": "Pricing successfully synced from SaaS Admin Service",
  "total_synced": 3,
  "created": 1,
  "updated": 2,
  "source_count": 3
}
```

### Option 2: Trigger Sync from SaaS Admin Dashboard

```bash
curl -X POST http://localhost:8097/api/v1/admin/pricing/sync-to-landing \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=YOUR_SESSION_ID"
```

## Database Schema

### SaaS Admin Database (saas_admin)

**saas_plans:**
```sql
CREATE TABLE saas_plans (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE,
    slug VARCHAR NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL NOT NULL,
    currency VARCHAR DEFAULT 'USD',
    billing_interval VARCHAR DEFAULT 'monthly',
    features TEXT,  -- JSON array
    is_popular BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT TRUE,
    button_text VARCHAR DEFAULT 'Get Started',
    button_url VARCHAR DEFAULT '/signup',
    display_order INT DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**pricing_tiers:**
```sql
CREATE TABLE pricing_tiers (
    id SERIAL PRIMARY KEY,
    plan_id UUID REFERENCES saas_plans(id),
    billing_interval VARCHAR NOT NULL,  -- monthly, yearly
    price DECIMAL NOT NULL,
    currency VARCHAR DEFAULT 'USD',
    discount_percent DECIMAL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Landing Page Database (statuspage_landing)

**pricing_plans:**
```sql
CREATE TABLE pricing_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    slug VARCHAR NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL NOT NULL,
    currency VARCHAR DEFAULT 'USD',
    billing_interval VARCHAR DEFAULT 'monthly',
    features TEXT,  -- JSON array
    is_popular BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    button_text VARCHAR,
    button_url VARCHAR,
    sort_order INT DEFAULT 0,
    status VARCHAR DEFAULT 'active',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Features Implemented

### ✅ Complete Backend Infrastructure
- [x] Public pricing API in SaaS Admin Service
- [x] Service-to-service authentication
- [x] Pricing sync handler in Landing Page Service
- [x] Atomic transactions for data integrity
- [x] Error handling and rollback
- [x] Detailed sync statistics

### ✅ Pricing Customization
- [x] Mark plans as popular/featured
- [x] Control display order
- [x] Customize button text and URL
- [x] Manage feature lists (JSON arrays)
- [x] Enable/disable plans
- [x] Public/private visibility

### ✅ Discount Support
- [x] Pricing tiers model
- [x] Monthly vs yearly pricing
- [x] Discount percentage calculation
- [x] Multiple billing intervals

### ✅ Data Synchronization
- [x] Manual sync trigger
- [x] Create/update logic
- [x] Slug-based matching
- [x] Transaction safety

## Testing the Implementation

### 1. Check SaaS Admin Plans

```bash
curl http://localhost:8097/api/v1/public/pricing-plans
```

### 2. Trigger Sync

```bash
curl -X POST http://localhost:8100/api/v1/admin/pricing/sync \
  -H "Authorization: Bearer service-token-development-secret-key-statuspage-2024"
```

### 3. Verify Landing Page Pricing

```bash
curl http://localhost:8100/api/v1/pricing
```

### 4. Check Landing Page

Open: http://localhost:8100

The pricing section should now display your customized plans from the SaaS Admin database!

## Next Steps (Optional Enhancements)

### 1. UI Dashboard for Pricing Management
Create a dedicated pricing management page in SaaS Admin dashboard with:
- Visual plan editor
- Drag-and-drop ordering
- Real-time preview
- One-click sync button

### 2. Automated Sync
Implement webhook or event-driven sync:
- Trigger sync automatically when plans are updated
- Use Redis pub/sub or event queue
- Background job processing

### 3. Version Control
Track pricing history:
- Plan version history
- Audit log for pricing changes
- Rollback capability

### 4. A/B Testing
Test different pricing strategies:
- Multiple active pricing versions
- Traffic splitting
- Conversion tracking

### 5. Multi-Currency Support
Extend for global pricing:
- Currency conversion
- Regional pricing
- Tax calculations

## Security Considerations

1. **Service-to-Service Auth**: Uses JWT secret for inter-service communication
2. **Admin-Only Endpoints**: Sync endpoint requires admin authentication
3. **Transaction Safety**: Atomic updates prevent partial sync failures
4. **Validation**: Input validation on all pricing fields
5. **Audit Logging**: All pricing changes should be logged (TODO)

## Performance Optimizations

1. **Preloading**: Eager loading of pricing tiers
2. **Caching**: Consider Redis cache for frequently accessed pricing
3. **Indexes**: Database indexes on slug, is_active, is_public
4. **Pagination**: Limit query results for large datasets

## Troubleshooting

### Sync Fails with "Service Unavailable"
- Check if SaaS Admin Service is running on port 8097
- Verify JWT_SECRET environment variable matches
- Check network connectivity between services

### Plans Not Showing on Landing Page
- Verify plans have `is_active = true` and `is_public = true`
- Check pricing sync was successful
- Verify database connection in Landing Page Service

### Features Not Displaying
- Ensure features field contains valid JSON array
- Check JSON escaping in the database
- Verify frontend parsing logic

## Files Modified/Created

### Created:
- `microservices/landing-page-service/internal/handlers/pricing_handler.go`
- `microservices/PRICING_MANAGEMENT_IMPLEMENTATION.md` (this file)

### Modified:
- `microservices/saas-admin-service/internal/handlers/saas_admin_handler.go`
- `microservices/saas-admin-service/internal/services/saas_admin_service.go`
- `microservices/landing-page-service/internal/server/server.go`

## Conclusion

The dynamic pricing management system is now fully operational! You can manage all pricing plans from the SaaS Admin Service and sync them to the Landing Page Service with a simple API call. The system is robust, resilient, and fail-proof with atomic transactions and comprehensive error handling.

For questions or issues, please refer to the troubleshooting section or check the implementation files directly.
