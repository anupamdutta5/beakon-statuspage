# Max Users Feature - Quick Reference Guide

## TL;DR

**What**: Set maximum user limits per tenant for tiered pricing
**Where**: Configured during tenant creation via SaaS admin
**How**: Add `max_users` field to tenant creation request
**Default**: NULL (unlimited users) - backward compatible

---

## Quick Start

### Create Tenant with User Limit

```bash
curl -X POST http://localhost:8000/api/v1/public/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme-corp",
    "contact_email": "admin@acme.com",
    "admin_email": "admin@acme.com",
    "admin_password": "SecurePass123!",
    "max_users": 10
  }'
```

### Create Tenant with Unlimited Users

```bash
curl -X POST http://localhost:8000/api/v1/public/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme-corp",
    "contact_email": "admin@acme.com",
    "admin_email": "admin@acme.com",
    "admin_password": "SecurePass123!",
    "max_users": null
  }'
```

Or simply omit the field:
```bash
# Omitting max_users = unlimited
curl -X POST http://localhost:8000/api/v1/public/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme-corp",
    "contact_email": "admin@acme.com",
    "admin_email": "admin@acme.com",
    "admin_password": "SecurePass123!"
  }'
```

### Check User Statistics

```bash
curl -X GET http://localhost:8000/api/v1/tenants/{tenant_id}/user-stats \
  -H "Authorization: Bearer {your_jwt_token}"
```

**Response**:
```json
{
  "data": {
    "current_users": 7,
    "max_users": 10,
    "is_unlimited": false,
    "remaining_slots": 3
  }
}
```

---

## Field Specification

| Field | Type | Required | Validation | Default |
|-------|------|----------|------------|---------|
| `max_users` | integer or null | No | Must be > 0 if set | `null` (unlimited) |

---

## Validation Rules

✅ **Valid**:
- `max_users: null` → Unlimited users
- `max_users: 1` → Single user
- `max_users: 100` → 100 users
- Field omitted → Unlimited users

❌ **Invalid**:
- `max_users: 0` → Rejected (must be positive)
- `max_users: -5` → Rejected (must be positive)

---

## Error Responses

### Limit Exceeded
```json
{
  "error": "tenant has reached maximum user limit of 10 users"
}
```
**Status**: 400 Bad Request (handler) or 500 Internal Server Error (service)

### Invalid Value
```json
{
  "error": "max_users must be greater than 0 if specified"
}
```
**Status**: 400 Bad Request

---

## Database Schema

```sql
-- Tenant table
CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR NOT NULL,
  slug VARCHAR UNIQUE NOT NULL,
  max_users INTEGER CHECK (max_users IS NULL OR max_users > 0),
  -- ... other fields
);
```

**Query current users**:
```sql
SELECT COUNT(*)
FROM users u
INNER JOIN tenant_admins ta ON u.id = ta.user_id
WHERE ta.tenant_id = '{tenant_id}'
  AND u.deleted_at IS NULL
  AND ta.deleted_at IS NULL;
```

---

## Implementation Files

| File | Purpose |
|------|---------|
| `internal/models/tenant_admin.go:25` | Model definition & validation |
| `internal/handlers/tenant_admin_handler.go:1055` | API request handling |
| `internal/services/tenant_admin_service.go:768` | Business logic validation |
| `cmd/main.go:489` | Route configuration |

---

## Common Use Cases

### Tiered Pricing
```javascript
const pricingTiers = {
  free: { max_users: 5 },
  starter: { max_users: 10 },
  professional: { max_users: 50 },
  enterprise: { max_users: null } // unlimited
};
```

### SaaS Admin Dashboard
```javascript
// Display user limit status
GET /api/v1/tenants/{tenant_id}/user-stats

// Show progress bar
{
  "current_users": 7,
  "max_users": 10,
  "usage_percentage": 70
}
```

### User Creation Flow
```go
// In application code
if err := service.CanCreateUser(tenantID); err != nil {
    // Show error: "Upgrade plan to add more users"
    return err
}
// Proceed with user creation
```

---

## Testing Checklist

- [ ] Create tenant with `max_users = 5`
- [ ] Add 5 users successfully
- [ ] Attempt to add 6th user → should fail
- [ ] Query `/user-stats` → verify correct counts
- [ ] Create tenant without `max_users` → verify unlimited
- [ ] Attempt `max_users = 0` → should fail with validation error
- [ ] Soft-delete a user → verify count decreases
- [ ] Verify existing tenants still work (NULL = unlimited)

---

## Monitoring

### Key Log Messages

```
INFO  Creating tenant with user limit  {"max_users": 10}
DEBUG User limit validation  {"current_users": 7, "max_users": 10}
WARN  Tenant user limit exceeded  {"current_users": 10, "max_users": 10}
```

### Recommended Metrics

```prometheus
# Counter: limit violations
tenant_user_limit_exceeded_total{tenant_id="..."}

# Gauge: current users
tenant_users_current{tenant_id="..."}

# Gauge: remaining slots
tenant_users_remaining_slots{tenant_id="..."}
```

---

## Troubleshooting

**Problem**: Users can't be created despite available slots

**Solution**: Check for soft-deleted users:
```sql
SELECT id, email, deleted_at
FROM users u
INNER JOIN tenant_admins ta ON u.id = ta.user_id
WHERE ta.tenant_id = '{tenant_id}';
```

**Problem**: max_users not enforced

**Solution**: Verify database column exists:
```sql
\d tenants  -- in psql
-- Check for max_users column
```

**Problem**: Validation error even though value is valid

**Solution**: Check logs for detailed error:
```bash
grep "max_users" /var/log/tenant-admin-service.log
```

---

## API Endpoints Summary

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/v1/public/tenants` | POST | No | Create tenant (with limit) |
| `/api/v1/tenants/:id` | GET | Yes | Get tenant (includes limit) |
| `/api/v1/tenants/:tenant_id/user-stats` | GET | Yes | Get user statistics |

---

## Migration Path

### For Existing Tenants
1. No action needed
2. Existing tenants have `max_users = NULL`
3. Unlimited users by default
4. Update later via API (if implemented)

### For New Tenants
1. Set `max_users` during creation
2. Based on pricing tier
3. Enforced automatically

---

## Best Practices

1. **Set Realistic Limits**: Base on pricing tier and expected usage
2. **Monitor Usage**: Use `/user-stats` endpoint for dashboards
3. **Notify Users**: Alert at 80% capacity
4. **Grace Period**: Consider soft warnings before hard enforcement
5. **Document Limits**: Show in tenant portal
6. **Upgrade Path**: Provide clear upgrade options when limit reached

---

## Support

**Full Documentation**: See `MAX_USERS_FEATURE.md` for comprehensive details

**Implementation Details**: See `IMPLEMENTATION_SUMMARY.md` for technical details

**Code Location**: `internal/models/tenant_admin.go:25`

**Questions?**: Check logs, review documentation, verify database schema
