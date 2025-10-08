# Max Users Feature Documentation

## Overview

The Max Users feature allows SaaS administrators to set a maximum limit on the number of users each tenant can create within their tenant admin portal. This provides control over resource usage and enables tiered pricing models.

## Implementation Summary

### Database Schema Changes

#### Tenant Model
- **Field Added**: `max_users` (*int, nullable)
- **Location**: `internal/models/tenant_admin.go:25`
- **Constraints**:
  - NULL = unlimited users
  - Must be > 0 if set
  - Database-level check constraint: `max_users IS NULL OR max_users > 0`
- **Validation**: GORM hooks (`BeforeCreate`, `BeforeUpdate`) ensure data integrity

### API Changes

#### 1. Create Tenant Endpoint
**Endpoint**: `POST /api/v1/public/tenants`

**Request Body** (enhanced):
```json
{
  "name": "Acme Corp",
  "slug": "acme-corp",
  "contact_email": "admin@acme.com",
  "admin_email": "admin@acme.com",
  "admin_password": "securePassword123",
  "max_users": 10
}
```

**Fields**:
- `max_users` (optional): Maximum number of users allowed
  - Type: integer
  - Validation: Must be > 0 if provided
  - Omit or set to `null` for unlimited users

**Response**:
```json
{
  "message": "Tenant created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Acme Corp",
    "slug": "acme-corp",
    "max_users": 10,
    "created_at": "2025-10-01T10:30:00Z",
    ...
  }
}
```

#### 2. Get Tenant User Statistics Endpoint (NEW)
**Endpoint**: `GET /api/v1/tenants/:tenant_id/user-stats`

**Purpose**: Retrieve current user count and limit information for monitoring and dashboards

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

**Response Fields**:
- `current_users`: Number of active users in the tenant
- `max_users`: Maximum allowed users (null if unlimited)
- `is_unlimited`: Boolean indicating if tenant has unlimited users
- `remaining_slots`: Number of remaining user slots (null if unlimited)

### Error Handling

#### User Limit Exceeded
When attempting to create a user that would exceed the limit:

**Status Code**: `400 Bad Request` (validation error) or `500 Internal Server Error` (service layer)

**Response**:
```json
{
  "error": "tenant has reached maximum user limit of 10 users"
}
```

#### Invalid max_users Value
When providing invalid max_users during tenant creation:

**Status Code**: `400 Bad Request`

**Response**:
```json
{
  "error": "max_users must be greater than 0 if specified"
}
```

## Service Layer Implementation

### Key Methods

#### 1. `CanCreateUser(tenantID uuid.UUID) error`
**Location**: `internal/services/tenant_admin_service.go:768`

**Purpose**: Validates whether a tenant can create a new user based on current count and limit

**Logic**:
1. Retrieves tenant from database
2. Returns `nil` if `max_users` is NULL (unlimited)
3. Counts current active users via `users` JOIN `tenant_admins`
4. Compares count against limit
5. Returns error if limit reached

**Features**:
- Comprehensive logging for monitoring
- Proper error handling with wrapped errors
- Accounts for soft-deleted users (excludes them from count)

#### 2. `GetTenantUserStats(tenantID uuid.UUID) (*TenantUserStats, error)`
**Location**: `internal/services/tenant_admin_service.go:724`

**Purpose**: Retrieves user statistics for dashboards and monitoring

**Returns**: `TenantUserStats` struct containing:
- Current user count
- Max users limit
- Unlimited flag
- Remaining slots

#### 3. `CreateAdminUser(tenantID uuid.UUID, email, password string) error`
**Location**: `internal/services/tenant_admin_service.go:788`

**Enhanced**: Now calls `CanCreateUser()` before creating user

**Flow**:
1. Validate max_users limit
2. Hash password
3. Create user record
4. Link user to tenant via tenant_admins
5. Log success

## Validation Layers

The feature implements defense-in-depth with multiple validation layers:

1. **Request Validation** (Handler Layer)
   - Gin binding validation: `binding:"omitempty,gt=0"`
   - Additional explicit check for max_users > 0
   - Location: `internal/handlers/tenant_admin_handler.go:1067`

2. **Model Validation** (Database Layer)
   - GORM hooks: `BeforeCreate`, `BeforeUpdate`
   - Database constraint: `check:max_users IS NULL OR max_users > 0`
   - Location: `internal/models/tenant_admin.go:236-250`

3. **Service Validation** (Business Logic Layer)
   - `CanCreateUser()` method enforces limit during user creation
   - Location: `internal/services/tenant_admin_service.go:768`

## Logging and Monitoring

### Log Levels Used

#### INFO
- Tenant creation with user limit
- User creation allowed
- Successful validation

#### WARN
- User limit exceeded attempts
- Cannot create user due to limit

#### DEBUG
- Detailed validation checks
- Current vs. max user counts
- Remaining slots

#### ERROR
- Database errors
- Tenant not found
- Validation failures

### Example Log Entries

```
INFO  Creating tenant with user limit  {"name": "Acme Corp", "slug": "acme-corp", "max_users": 10}
DEBUG User limit validation  {"tenant_id": "...", "current_users": 7, "max_users": 10}
WARN  Tenant user limit exceeded  {"tenant_id": "...", "current_users": 10, "max_users": 10}
```

## Database Migration

### Automatic Migration
The `max_users` column is automatically added when the service starts via GORM AutoMigrate:
- Column: `max_users`
- Type: INTEGER (nullable)
- Constraint: `max_users IS NULL OR max_users > 0`

### Existing Data
- All existing tenants will have `max_users = NULL` (unlimited users)
- No data migration required
- Backward compatible

## Testing Recommendations

### Unit Tests

1. **Test Unlimited Users** (max_users = NULL)
   - Create tenant without max_users
   - Verify unlimited user creation

2. **Test User Limit Enforcement**
   - Create tenant with max_users = 2
   - Create 2 users successfully
   - Attempt 3rd user creation → should fail with proper error

3. **Test Validation**
   - Attempt to set max_users = 0 → should fail
   - Attempt to set max_users = -5 → should fail
   - Set max_users = 100 → should succeed

4. **Test Edge Cases**
   - max_users = 1 (single user tenant)
   - Updating max_users after tenant creation
   - Soft-deleted users not counted toward limit

### Integration Tests

1. **End-to-End Tenant Creation**
   ```bash
   # Create tenant with limit
   curl -X POST http://localhost:8000/api/v1/public/tenants \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Test Corp",
       "slug": "test-corp",
       "contact_email": "test@test.com",
       "admin_email": "admin@test.com",
       "admin_password": "Password123!",
       "max_users": 5
     }'
   ```

2. **Verify User Statistics**
   ```bash
   # Get user stats
   curl -X GET http://localhost:8000/api/v1/tenants/{tenant_id}/user-stats \
     -H "Authorization: Bearer {token}"
   ```

3. **Test Limit Enforcement**
   - Create users up to limit
   - Verify error on exceeding limit
   - Check error message clarity

## Performance Considerations

### Database Queries
- User count query uses indexed JOIN on `tenant_id`
- Query optimized with proper indexes on:
  - `tenant_admins.tenant_id`
  - `users.deleted_at`
  - `tenant_admins.deleted_at`

### Caching Opportunities
For high-volume deployments, consider caching:
- Current user count (with TTL)
- Max users limit
- Invalidate cache on user creation/deletion

### Race Conditions
- Current implementation uses standard SELECT
- For high-concurrency scenarios, consider:
  - Database-level locks (`FOR UPDATE`)
  - Transaction-based user creation
  - Optimistic locking with version fields

## Security Considerations

1. **Multi-Tenant Isolation**
   - User counts are scoped to tenant_id
   - Prevents cross-tenant data leakage

2. **Input Validation**
   - Multiple validation layers prevent invalid data
   - SQL injection protected via parameterized queries

3. **Authorization**
   - Only SaaS admin can set max_users during tenant creation
   - Tenant admins cannot modify their own limit

## Backward Compatibility

### API Compatibility
- `max_users` field is optional in all requests
- Omitting field maintains current behavior (unlimited)
- Existing API consumers unaffected

### Database Compatibility
- New column is nullable
- Existing records have NULL value (unlimited)
- No breaking changes to existing functionality

### Service Compatibility
- Validation only applied if max_users is set
- NULL value bypasses all limit checks
- Existing user creation flows unchanged

## Future Enhancements

### Potential Improvements

1. **Audit Trail**
   - Log limit changes to tenant_activities table
   - Track failed user creation attempts

2. **Notification System**
   - Alert tenant when approaching limit (e.g., 80% capacity)
   - Email notification on limit reached

3. **Grace Period**
   - Allow temporary over-limit with warning
   - Configurable grace period per tenant

4. **Dynamic Limit Updates**
   - API endpoint to update max_users for existing tenants
   - Validation to prevent setting limit below current count

5. **Usage Analytics**
   - Dashboard showing user count trends
   - Capacity planning insights
   - Tenant growth metrics

## Troubleshooting

### Common Issues

#### Issue: Users can't be created despite available slots
**Cause**: Soft-deleted users may still exist
**Solution**: Verify user count with:
```sql
SELECT COUNT(*) FROM users u
INNER JOIN tenant_admins ta ON u.id = ta.user_id
WHERE ta.tenant_id = '{tenant_id}'
AND u.deleted_at IS NULL
AND ta.deleted_at IS NULL;
```

#### Issue: max_users not enforced
**Cause**: Validation bypassed or database migration incomplete
**Solution**:
1. Verify column exists: `\d tenants` in psql
2. Check service logs for validation errors
3. Ensure database migration ran successfully

#### Issue: Error "tenant has reached maximum user limit" but count seems wrong
**Cause**: Multiple active user records or stale data
**Solution**:
1. Query tenant_admins for duplicate entries
2. Check for orphaned user records
3. Verify tenant_id matches

## Configuration

### Environment Variables
No new environment variables required. The feature uses existing database configuration.

### Feature Flags
Consider adding feature flag for gradual rollout:
```go
// In tenant settings or feature flags
"max_users_enforcement_enabled": true
```

## API Summary

### Modified Endpoints

| Endpoint | Method | Changes |
|----------|--------|---------|
| `/api/v1/public/tenants` | POST | Added optional `max_users` field |
| `/api/v1/tenants/:id` | GET | Returns `max_users` in response |
| `/api/v1/tenants/:id` | PUT | Allows updating `max_users` (if implemented) |

### New Endpoints

| Endpoint | Method | Purpose | Auth |
|----------|--------|---------|------|
| `/api/v1/tenants/:tenant_id/user-stats` | GET | Get user count and limit stats | Required |

## Support and Maintenance

### Monitoring Metrics
Track these metrics in production:
- `tenant_user_limit_exceeded_total` (counter)
- `tenant_user_creation_attempts_total` (counter)
- `tenant_users_current` (gauge per tenant)
- `tenant_users_remaining_slots` (gauge per tenant)

### Health Checks
Existing `/health` endpoint covers this feature.

### Rollback Plan
If issues occur:
1. Remove validation call from `CreateAdminUser()`
2. Redeploy service
3. Column can remain (NULL values = unlimited)
4. No data loss or corruption risk

## Conclusion

The Max Users feature provides robust, secure, and backward-compatible user limit enforcement for multi-tenant environments. The implementation follows best practices with multiple validation layers, comprehensive logging, and proper error handling.
