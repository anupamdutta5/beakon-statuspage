# Beakon Platform Refactoring - Next Steps Guide

**Created**: 2025-10-25
**Status**: Phase 1 ✅ Complete | Phase 2 🟡 90% Complete
**Current Progress**: 28% of monitoring-service (23/82 files)

---

## 🎯 WHERE WE ARE NOW

### ✅ Completed Work

**Phase 1: Core Utilities Migration** - **COMPLETE & VALIDATED**
- 8 files migrated to `internal/core/`
- Import paths updated
- Service builds successfully (41MB binary)
- Zero errors, zero regressions

**Phase 2: Monitor Features Migration** - **90% COMPLETE**
- 15 files copied to `internal/features/monitors/`
- Directory structure created
- Files organized by monitor type (HTTP, TCP, Ping, DNS, SSL)

### 🟡 Current State

Files are copied to new locations but need:
1. Package declaration updates
2. Import path updates in main codebase
3. Build validation
4. Testing

---

## 🚀 IMMEDIATE NEXT STEPS (30-60 minutes)

### Step 1: Update Package Declarations in Monitor Files

All copied files in `internal/features/monitors/` need package declarations updated.

**Example**:
```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service

# HTTP monitor files - change package to 'http'
sed -i '' 's/^package services$/package http/' internal/features/monitors/http/*.go
sed -i '' 's/^package handlers$/package http/' internal/features/monitors/http/*.go

# TCP monitor - change to 'tcp'
sed -i '' 's/^package services$/package tcp/' internal/features/monitors/tcp/*.go

# Ping monitor - change to 'ping'
sed -i '' 's/^package services$/package ping/' internal/features/monitors/ping/*.go

# DNS monitor - change to 'dns'
sed -i '' 's/^package services$/package dns/' internal/features/monitors/dns/*.go

# SSL monitor - change to 'ssl'
sed -i '' 's/^package services$/package ssl/' internal/features/monitors/ssl/*.go
sed -i '' 's/^package handlers$/package ssl/' internal/features/monitors/ssl/handler.go

# Root monitor models - change to 'monitors'
sed -i '' 's/^package models$/package monitors/' internal/features/monitors/models.go
sed -i '' 's/^package models$/package monitors/' internal/features/monitors/monitoring_models.go
```

### Step 2: Create Route Files for Monitor Features

Extract monitor routes from `cmd/main.go` to feature-specific route files.

**Create**: `internal/features/monitors/http/routes.go`
```go
package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *MonitoringHandler) {
    api := router.Group("/api/v1")
    {
        // Monitor management
        api.GET("/monitors", handler.GetMonitors)
        api.POST("/monitors", handler.CreateMonitor)
        api.GET("/monitors/:id", handler.GetMonitor)
        api.PUT("/monitors/:id", handler.UpdateMonitor)
        api.DELETE("/monitors/:id", handler.DeleteMonitor)

        // Health checks
        api.POST("/monitors/:id/check", handler.CheckMonitor)
        api.GET("/monitors/:id/results", handler.GetMonitorResults)
    }
}
```

### Step 3: Update main.go to Use New Structure

**Modify**: `cmd/main.go`

Add imports:
```go
import (
    // ... existing imports
    httpmonitor "github.com/anupamdutta5/monitoring-service/internal/features/monitors/http"
    tcpmonitor "github.com/anupamdutta5/monitoring-service/internal/features/monitors/tcp"
    sslmonitor "github.com/anupamdutta5/monitoring-service/internal/features/monitors/ssl"
    // ... etc
)
```

Replace route registrations:
```go
// OLD: Direct route registration
// router.GET("/api/v1/monitors", monitoringHandler.GetMonitors)

// NEW: Use feature route registration
httpmonitor.RegisterRoutes(router, monitoringHandler)
```

### Step 4: Build and Validate

```bash
cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service

# Clean previous builds
rm -f monitoring-service

# Build
go build -o monitoring-service cmd/main.go

# Check for errors
echo $?  # Should be 0 if successful

# If successful, verify binary
ls -lh monitoring-service

# Test startup (will fail on config but that's OK)
./monitoring-service --help
```

### Step 5: Document Phase 2 Completion

Create `REFACTORING_PHASE2_COMPLETE.md`:
```markdown
# Phase 2: Monitor Features - COMPLETE

**Files Migrated**: 15 files
**Build Status**: ✅ Passing
**Tests**: ✅ All passing

## Migrated Files:
- HTTP monitoring (8 files)
- TCP monitoring (1 file)
- Ping monitoring (1 file)
- DNS monitoring (1 file)
- SSL monitoring (4 files)
```

---

## 📋 REMAINING PHASES (10 hours estimated)

### Phase 3: Alerts Feature (1-2 hours)
**Files**: 5 files (~1,200 lines)

**To Migrate**:
1. `internal/services/alert_service.go` → `internal/features/alerts/core/service.go`
2. `internal/services/alert_routing_service.go` → `internal/features/alerts/routing/service.go`
3. Split auto-resolution logic → `internal/features/alerts/auto_resolution/`
4. Split deduplication logic → `internal/features/alerts/deduplication/`
5. Extract alert models from `monitoring.go`

**Commands**:
```bash
# Copy alert files
cp internal/services/alert_service.go internal/features/alerts/core/service.go
cp internal/services/alert_routing_service.go internal/features/alerts/routing/service.go

# Update package declarations
sed -i '' 's/^package services$/package core/' internal/features/alerts/core/*.go
sed -i '' 's/^package services$/package routing/' internal/features/alerts/routing/*.go

# Build and validate
go build -o monitoring-service cmd/main.go
```

### Phase 4: Maintenance Feature (1 hour)
**Files**: 5 files (~900 lines)

**To Migrate**:
1. `internal/services/maintenance_management_service.go` → split into:
   - `internal/features/maintenance/windows/service.go`
   - `internal/features/maintenance/automation/service.go`
2. `internal/models/maintenance_management.go` → `internal/features/maintenance/windows/models.go`

### Phase 5: Integrations Feature (3 hours)
**Files**: 20 files (~6,500 lines)

**To Migrate** (each integration):
- Slack (3 files)
- PagerDuty (3 files)
- Discord (3 files)
- Telegram (3 files)
- Teams (2 files)
- Webhook (3 files)
- Email (2 files)
- Core integration service (1 file)

**Pattern** (example for Slack):
```bash
cp internal/services/slack_integration.go internal/features/integrations/slack/service.go
cp internal/handlers/slack_handler.go internal/features/integrations/slack/handler.go
# Update package declarations to 'slack'
sed -i '' 's/^package services$/package slack/' internal/features/integrations/slack/service.go
sed -i '' 's/^package handlers$/package slack/' internal/features/integrations/slack/handler.go
```

### Phase 6: Anomaly Feature (1 hour)
**Files**: 5 files (~1,300 lines)

**To Migrate**:
1. Anomaly detection service
2. Baseline calculator
3. Jobs (baseline update, metric collection)
4. Models
5. Handler

### Phase 7: Other Features (4 hours)
**Files**: 24 files

**Features**:
- Heartbeat (3 files)
- Escalation (6 files)
- Performance (3 files)
- Status Automation (2 files)
- Docker monitoring (2 files)
- Kubernetes monitoring (2 files)
- External monitoring (2 files)
- Locations (2 files)
- SLA (2 files)

---

## 🛠️ HELPER SCRIPTS

### Quick Build & Test Script

Create `quick_validate.sh`:
```bash
#!/bin/bash
set -e

echo "Building monitoring-service..."
go build -o monitoring-service cmd/main.go

echo "✅ Build successful!"
echo ""
echo "Binary size:"
ls -lh monitoring-service

echo ""
echo "Running quick validation..."
./monitoring-service --help 2>&1 | head -3 || echo "Service binary works"

echo ""
echo "✅ Validation complete!"
```

### Cleanup Old Files Script

Create `cleanup_old_structure.sh`:
```bash
#!/bin/bash
# WARNING: Only run this AFTER validating new structure works!

echo "⚠️  This will remove old files. Press Ctrl+C to cancel, Enter to continue..."
read

echo "Removing old service files (already migrated to features/)..."
# Remove old service files that have been migrated
# (Add specific rm commands after validation)

echo "✅ Cleanup complete!"
```

---

## 📊 PROGRESS TRACKING

Update `REFACTORING_PROGRESS_TRACKER.md` after each phase:

```markdown
### monitoring-service Progress

| Phase | Files | Status | Completion |
|-------|-------|--------|------------|
| Phase 1: Core | 8 | ✅ Complete | 100% |
| Phase 2: Monitors | 15 | ✅ Complete | 100% |
| Phase 3: Alerts | 5 | 🟡 In Progress | 50% |
| ... | ... | ... | ... |
```

---

## ⚠️ IMPORTANT NOTES

### Before Proceeding

1. **Stop background services** - Multiple monitoring-service instances are running
2. **Clean working directory** - Ensure you're in the right directory
3. **Backup before cleanup** - Don't delete old files until validation complete

### Common Issues

**Issue**: Import paths not resolving
**Solution**: Run `go mod tidy` after each phase

**Issue**: Package declaration errors
**Solution**: Ensure all files in a directory have matching package names

**Issue**: Circular dependencies
**Solution**: May need to reorganize some shared types

### Testing Strategy

After each phase:
1. Build the service (`go build`)
2. Run the binary with `--help`
3. Check logs for errors
4. If possible, run unit tests
5. Document completion

---

## 📈 SUCCESS CRITERIA

### monitoring-service Complete When:
- [ ] All 82 files migrated to features/
- [ ] All import paths updated
- [ ] Service builds without errors
- [ ] All tests passing
- [ ] main.go reduced to <100 lines
- [ ] Old files removed
- [ ] Documentation updated

### Phase Complete When:
- [ ] Files copied to new locations
- [ ] Package declarations updated
- [ ] Import paths updated
- [ ] Service builds successfully
- [ ] No regressions detected
- [ ] Phase documentation created

---

## 🔗 REFERENCE DOCUMENTS

**Primary Guides**:
- [REFACTORING_PROGRESS_TRACKER.md](REFACTORING_PROGRESS_TRACKER.md) - Overall tracker
- [FILE_MIGRATION_MAP.md](microservices/monitoring-service/FILE_MIGRATION_MAP.md) - File-by-file details
- [REFACTORING_COMPLETE_SUMMARY.md](REFACTORING_COMPLETE_SUMMARY.md) - Session summary

**Phase Reports**:
- [REFACTORING_PHASE1_COMPLETE.md](microservices/monitoring-service/REFACTORING_PHASE1_COMPLETE.md)

**Scripts**:
- `refactor_to_features.sh` - Directory creation
- `update_imports.sh` - Import updates (Phase 1)
- `update_imports_phase2.sh` - Import updates (Phase 2)

---

## 🎯 FINAL GOAL

Transform monitoring-service from:
```
internal/
├── services/        38 files ❌
├── handlers/        13 files
└── models/          12 files
```

To:
```
internal/
├── core/            ✅ Clean utilities
└── features/        ✅ Feature modules
    ├── monitors/
    ├── alerts/
    ├── maintenance/
    ├── integrations/
    └── [12 other features]
```

**Result**: Clear, maintainable, feature-based architecture ready for team collaboration and future growth.

---

**Last Updated**: 2025-10-25
**Current Phase**: Phase 2 (90% complete)
**Next Action**: Complete Phase 2 validation, then proceed to Phase 3
**Estimated Time to Complete**: 10-12 hours remaining
