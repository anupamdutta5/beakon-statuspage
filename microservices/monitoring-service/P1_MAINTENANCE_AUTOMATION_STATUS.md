# P1 Feature: Maintenance Automation - Implementation Status

**Date**: 2025-10-22
**Feature**: Maintenance Automation (Auto-reminder, Auto-start, Auto-complete)
**Service**: monitoring-service
**Status**: ⚠️ **IN PROGRESS** - Schema alignment needed

---

## ✅ Completed Work

### 1. Database Migration
- **File**: [migrations/006_add_maintenance_automation.sql](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service/migrations/006_add_maintenance_automation.sql)
- **Status**: ✅ Applied successfully
- **Changes**:
  ```sql
  -- Added automation tracking columns
  ALTER TABLE maintenance_windows
  ADD COLUMN reminder_sent BOOLEAN DEFAULT false,
  ADD COLUMN auto_started BOOLEAN DEFAULT false,
  ADD COLUMN auto_completed BOOLEAN DEFAULT false,
  ADD COLUMN actual_start_time TIMESTAMP,
  ADD COLUMN actual_end_time TIMESTAMP,
  ADD COLUMN status VARCHAR(50) DEFAULT 'scheduled';

  -- Added indexes for automation queries
  CREATE INDEX idx_maintenance_automation ON maintenance_windows(status, starts_at);
  CREATE INDEX idx_maintenance_reminders ON maintenance_windows(reminder_sent, starts_at);

  -- Added constraint for valid statuses
  ADD CONSTRAINT chk_maintenance_status
  CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled'));
  ```

### 2. Service Logic Implementation
- **File**: [internal/services/maintenance_management_service.go](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service/internal/services/maintenance_management_service.go)
- **Status**: ✅ Methods created, ⚠️ Field names need alignment
- **Methods Added**:
  1. `SendMaintenanceReminders()` - Lines 490-529
     - Sends 60-minute warnings before maintenance starts
     - Queries: `status = 'scheduled' AND reminder_sent = false AND starts_at BETWEEN now AND +60min`
     - Marks reminder_sent = true after sending

  2. `AutoStartMaintenanceWindows()` - Lines 426-456 (existing, updated)
     - Transitions status from 'scheduled' to 'in_progress'
     - Queries: `status = 'scheduled' AND starts_at <= now`

  3. `AutoCompleteMaintenanceWindows()` - Lines 458-488 (existing, updated)
     - Transitions status from 'in_progress' to 'completed'
     - Queries: `status = 'in_progress' AND ends_at <= now`

### 3. Background Job Integration
- **File**: [internal/jobs/ssl_expiration_checker.go](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service/internal/jobs/ssl_expiration_checker.go)
- **Status**: ✅ Job already exists, updated to include reminders
- **Job**: `MaintenanceWindowJob` - Lines 227-290
  - **Interval**: 1 minute (configured in main.go:361)
  - **Methods called**:
    1. `SendMaintenanceReminders()` - NEW
    2. `AutoStartMaintenanceWindows()` - Existing
    3. `AutoCompleteMaintenanceWindows()` - Existing

### 4. Test Suite Created
- **File**: [cmd/test_maintenance_automation.go](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service/cmd/test_maintenance_automation.go)
- **Status**: ✅ Created, ready to run after schema fixes
- **Tests**:
  1. Auto-reminder (60 minutes before start)
  2. Auto-start (transition to in_progress)
  3. Auto-complete (transition to completed)

---

## ⚠️ Blocking Issues

### Schema Mismatch Between Model and Database

**Problem**: The Go model struct uses different field names than the database columns.

| Database Column | Go Model Field (OLD) | Go Model Field (NEW) | Status |
|----------------|---------------------|---------------------|---------|
| `name` | `Title` | `Name` | ⚠️ Partially fixed |
| `starts_at` | `StartTime` | `StartsAt` | ⚠️ Partially fixed |
| `ends_at` | `EndTime` | `EndsAt` | ⚠️ Partially fixed |
| `tenant_id` (UUID) | `TenantID` (uint) | `TenantID` (string) | ⚠️ Type changed |
| `created_by` (UUID) | `CreatedBy` (uint) | `CreatedBy` (string) | ⚠️ Type changed |
| `reminder_sent` | ❌ Not in model | `ReminderSent` | ✅ Added |
| `auto_started` | ❌ Not in model | `AutoStarted` | ✅ Added |
| `auto_completed` | ❌ Not in model | `AutoCompleted` | ✅ Added |

**Files Affected**:
1. ✅ `internal/models/maintenance_management.go` - Model struct updated
2. ⚠️ `internal/services/maintenance_management_service.go` - Has references to old field names in `CreateMaintenanceFromTemplate()` method
3. ⚠️ Other files may reference old fields

**Compilation Errors**:
```
internal/services/maintenance_management_service.go:316:16: cannot use template.TenantID (variable of type uint) as string
internal/services/maintenance_management_service.go:317:3: unknown field Title in struct literal
internal/services/maintenance_management_service.go:320:3: unknown field Type in struct literal
internal/services/maintenance_management_service.go:321:3: unknown field Impact in struct literal
internal/services/maintenance_management_service.go:322:3: unknown field StartTime in struct literal
internal/services/maintenance_management_service.go:323:3: unknown field EndTime in struct literal
```

---

## 📋 Next Steps

### Immediate (Required to Complete Feature)

1. **Fix Remaining Field References** in `maintenance_management_service.go`:
   - Update `CreateMaintenanceFromTemplate()` method (lines 304-362)
   - Change `Title` → `Name`
   - Change `StartTime` → `StartsAt`
   - Change `EndTime` → `EndsAt`
   - Handle `TenantID` and `CreatedBy` type conversion (uint → string/UUID)

2. **Remove or Comment Out Template-Related Fields** (if not in database):
   - Model has `Type`, `Impact` fields but database doesn't
   - Either:
     - Option A: Remove from model entirely
     - Option B: Add these columns to database via migration
     - **Recommendation**: Check database schema first with:
       ```sql
       SELECT column_name FROM information_schema.columns
       WHERE table_name = 'maintenance_windows'
       ORDER BY ordinal_position;
       ```

3. **Fix Related Model Methods**:
   - Update all methods in `maintenance_management_service.go` that reference old fields
   - Search for: `Title`, `StartTime`, `EndTime`, `Type`, `Impact`

4. **Rebuild and Test**:
   ```bash
   go build -o monitoring-service cmd/main.go
   pkill -f monitoring-service
   ./monitoring-service > /tmp/monitoring-service.log 2>&1 &
   go run cmd/test_maintenance_automation.go
   ```

### Future Enhancements (After Basic Automation Works)

1. **Integrate with Notification Service**:
   - Currently `SendMaintenanceReminders()` only marks `reminder_sent = true`
   - TODO (line 510): Send actual notifications via notification-service
   - Could use RabbitMQ events or direct HTTP calls

2. **Track Actual Times**:
   - Set `actual_start_time` when auto-starting
   - Set `actual_end_time` when auto-completing
   - Currently these fields exist but aren't populated

3. **Add Metrics**:
   - Track how many maintenances auto-start vs manual start
   - Track reminder delivery success rate
   - Monitor auto-completion accuracy

4. **Frontend Integration**:
   - Add UI to show automation status (reminder sent, auto-started, etc.)
   - Display actual vs scheduled times
   - Show automation history

---

## 🔍 Database Schema Reference

**Current `maintenance_windows` table** (verified 2025-10-22):
```sql
CREATE TABLE maintenance_windows (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,
    affected_monitors TEXT,
    is_active BOOLEAN DEFAULT false,
    suppress_notifications BOOLEAN DEFAULT true,
    auto_update_status_page BOOLEAN DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),

    -- Automation fields (added in migration 006)
    reminder_sent BOOLEAN DEFAULT false,
    auto_started BOOLEAN DEFAULT false,
    auto_completed BOOLEAN DEFAULT false,
    actual_start_time TIMESTAMP,
    actual_end_time TIMESTAMP,
    status VARCHAR(50) DEFAULT 'scheduled',

    CONSTRAINT chk_maintenance_status
    CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled'))
);

CREATE INDEX idx_maintenance_automation ON maintenance_windows(status, starts_at) WHERE is_active = true;
CREATE INDEX idx_maintenance_reminders ON maintenance_windows(reminder_sent, starts_at) WHERE is_active = true AND reminder_sent = false;
```

---

## 📊 Implementation Progress

| Component | Status | Notes |
|-----------|--------|-------|
| Database Migration | ✅ Complete | Migration 006 applied |
| Model Fields | ⚠️ 80% | Automation fields added, field name alignment in progress |
| Service Methods | ⚠️ 90% | 3 automation methods created, field references need fixes |
| Background Job | ✅ Complete | MaintenanceWindowJob updated, runs every 1 minute |
| Test Suite | ✅ Complete | Ready to run after fixes |
| **Overall** | **⚠️ 85%** | **Blocked by field name alignment** |

---

## 🎯 Definition of Done

- [ ] All Go files compile successfully
- [ ] Test suite runs and all 3 tests pass
- [ ] Background job logs show automation working:
  - Reminders sent for windows starting in 60 minutes
  - Windows auto-start when start time arrives
  - Windows auto-complete when end time arrives
- [ ] No errors in `/tmp/monitoring-service.log`
- [ ] Database records show correct `reminder_sent`, `status` updates

---

## 📝 Related Documentation

- [P1 Features Implementation Plan](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/P1_FEATURES_IMPLEMENTATION_PLAN.md)
- [Monitoring Features Roadmap](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/docs/features/MONITORING_FEATURES_ROADMAP.md)
- [P0 Integration Test Results](file:///Users/anuoamdutta/Desktop/statuspage/Beakon/P0_INTEGRATION_TEST_RESULTS.md)

---

**Last Updated**: 2025-10-22
**Next Action**: Fix field name references in `maintenance_management_service.go` lines 304-362
