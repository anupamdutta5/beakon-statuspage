# P1 (High Priority) Features - Implementation Plan

**Date**: 2025-10-22
**Status**: Ready to Start
**Prerequisites**: ✅ All P0 features complete (12/12 - 100%)

---

## 📋 Executive Summary

After completing all 12 P0 (Critical) features, we're now moving to **P1 (High Priority) features** that provide strong competitive advantage and differentiation.

**P1 Features Overview**:
- 17 total P1 features identified
- Estimated time: 4-6 weeks
- Focus: Performance metrics, integrations, analytics, advanced monitoring

---

## 🎯 P1 Features by Category

### Category A: Uptime Monitoring (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | TCP port monitoring | Medium | 2 days | P1 - High |
| 2 | ICMP ping monitoring | Medium | 2 days | P1 - High |
| 3 | Custom interval checks (30s-1h) | Low | 1 day | P1 - High |

**Total Time**: 5 days (1 week)

---

### Category B: Performance Metrics (4 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Response time percentiles (P50, P95, P99) | Medium | 3 days | P1 - High |
| 2 | Page load time monitoring | High | 4 days | P1 - High |
| 3 | Time to first byte (TTFB) | Medium | 2 days | P1 - High |
| 4 | Custom metric definitions | High | 5 days | P1 - High |

**Total Time**: 14 days (2.8 weeks)

---

### Category C: Incident Management (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Incident timeline visualization | Medium | 3 days | P1 - High |
| 2 | Incident priority levels | Low | 2 days | P1 - High |
| 3 | Incident owner assignment | Medium | 2 days | P1 - High |

**Total Time**: 7 days (1.4 weeks)

---

### Category D: Scheduled Maintenance (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Auto-reminder 60min before start | Low | 1 day | P1 - High |
| 2 | Auto-transition to "In Progress" | Low | 1 day | P1 - High |
| 3 | Auto-completion at end time | Low | 1 day | P1 - High |

**Total Time**: 3 days (0.6 weeks)

---

### Category E: Notifications (2 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Microsoft Teams notifications | Medium | 3 days | P1 - High |
| 2 | Component-specific subscriptions | Medium | 3 days | P1 - High |
| 3 | Subscription preferences management | Medium | 3 days | P1 - High |
| 4 | Notification throttling | Low | 2 days | P1 - High |

**Total Time**: 11 days (2.2 weeks)

---

### Category F: Status Page Features (4 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Private status pages | Medium | 3 days | P1 - High |
| 2 | Historical incident calendar | Medium | 3 days | P1 - High |
| 3 | Uptime showcase (90-day) | Medium | 3 days | P1 - High |
| 4 | JSON API for status | Low | 2 days | P1 - High |

**Total Time**: 11 days (2.2 weeks)

---

### Category G: Alerting System (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Alert routing rules | Medium | 3 days | P1 - High |
| 2 | Alert deduplication | Low | 2 days | P1 - High |
| 3 | Auto-resolution when check passes | Low | 2 days | P1 - High |

**Total Time**: 7 days (1.4 weeks)

---

### Category H: Third-Party Integrations (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Datadog integration | High | 4 days | P1 - High |
| 2 | Prometheus integration | High | 4 days | P1 - High |
| 3 | SSL certificate validation | Medium | 2 days | P1 - High |

**Total Time**: 10 days (2 weeks)

---

### Category I: Analytics & Reporting (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | SLA reporting (monthly uptime) | High | 5 days | P1 - High |
| 2 | MTTR tracking | Medium | 3 days | P1 - High |
| 3 | Exportable reports (PDF, CSV) | High | 4 days | P1 - High |

**Total Time**: 12 days (2.4 weeks)

---

### Category J: Advanced Features (3 features)

| # | Feature | Complexity | Time | Priority |
|---|---------|------------|------|----------|
| 1 | Synthetic transaction monitoring | Very High | 7 days | P1 - High |
| 2 | API endpoint monitoring with assertions | High | 5 days | P1 - High |
| 3 | Multi-step health checks | High | 5 days | P1 - High |

**Total Time**: 17 days (3.4 weeks)

---

## 📊 Overall Summary

| Category | Features | Days | Weeks |
|----------|----------|------|-------|
| Uptime Monitoring | 3 | 5 | 1.0 |
| Performance Metrics | 4 | 14 | 2.8 |
| Incident Management | 3 | 7 | 1.4 |
| Scheduled Maintenance | 3 | 3 | 0.6 |
| Notifications | 4 | 11 | 2.2 |
| Status Page Features | 4 | 11 | 2.2 |
| Alerting System | 3 | 7 | 1.4 |
| Integrations | 3 | 10 | 2.0 |
| Analytics & Reporting | 3 | 12 | 2.4 |
| Advanced Features | 3 | 17 | 3.4 |
| **TOTAL** | **33** | **97** | **19.4** |

**Adjusted for parallel work**: ~4-6 weeks with efficient prioritization

---

## 🚀 Recommended Implementation Order

### **Quick Wins First** (Week 1-2): Low-hanging fruit with high impact

1. **Maintenance Automation** (3 days)
   - Auto-reminder 60min before start
   - Auto-transition to "In Progress"
   - Auto-completion at end time
   - **Why First**: Simple, high value, builds on existing maintenance system

2. **Alerting Improvements** (4 days)
   - Auto-resolution when check passes
   - Alert deduplication
   - **Why Next**: Reduces alert fatigue, improves UX

3. **Incident Enhancements** (4 days)
   - Incident priority levels
   - Incident owner assignment
   - **Why Next**: Essential for team collaboration

4. **Custom Intervals** (1 day)
   - Custom interval checks (30s-1h)
   - **Why Next**: Frequently requested, easy to implement

**Total Week 1-2**: 12 days → Can be done in 2 weeks

---

### **High-Impact Features** (Week 3-4): Competitive differentiators

5. **Performance Metrics** (7 days)
   - Response time percentiles (P50, P95, P99)
   - Time to first byte (TTFB)
   - **Why**: Critical for performance-focused customers

6. **Status Page Enhancements** (8 days)
   - Private status pages
   - JSON API for status
   - Uptime showcase (90-day)
   - **Why**: Expands target market (private pages), enables integrations (API)

**Total Week 3-4**: 15 days → Can be done in 2 weeks with focus

---

### **Advanced Capabilities** (Week 5-6): Power features

7. **Analytics & SLA** (9 days)
   - SLA reporting (monthly uptime)
   - MTTR tracking
   - Exportable reports (PDF)
   - **Why**: Enterprise requirement, revenue driver

8. **Monitoring Enhancements** (7 days)
   - TCP port monitoring
   - ICMP ping monitoring
   - SSL certificate validation
   - **Why**: Expands monitoring capabilities

9. **Alert Routing** (3 days)
   - Alert routing rules
   - **Why**: Advanced alerting for complex teams

**Total Week 5-6**: 19 days → Can be done in 3 weeks

---

## 📝 Detailed Implementation: Quick Wins (Recommended Start)

### 1. Maintenance Automation (3 days)

**Goal**: Automate maintenance window lifecycle

#### Day 1: Auto-Reminder Implementation

**Backend (monitoring-service)**:
```go
// internal/jobs/maintenance_reminder_job.go (NEW)
type MaintenanceReminderJob struct {
    db     *gorm.DB
    logger *zap.Logger
}

func (j *MaintenanceReminderJob) Start() {
    ticker := time.NewTicker(5 * time.Minute)

    for range ticker.C {
        j.checkUpcomingMaintenance()
    }
}

func (j *MaintenanceReminderJob) checkUpcomingMaintenance() {
    // Find maintenance windows starting in 60 minutes
    now := time.Now()
    reminderTime := now.Add(60 * time.Minute)

    var windows []MaintenanceWindow
    j.db.Where("scheduled_start BETWEEN ? AND ?", reminderTime, reminderTime.Add(5*time.Minute)).
        Where("status = ?", "scheduled").
        Where("reminder_sent = ?", false).
        Find(&windows)

    for _, window := range windows {
        // Publish reminder event
        publishEvent("beakon.maintenance.reminder", window)

        // Mark as sent
        j.db.Model(&window).Update("reminder_sent", true)
    }
}
```

**Database Migration**:
```sql
ALTER TABLE maintenance_windows
ADD COLUMN reminder_sent BOOLEAN DEFAULT false,
ADD COLUMN auto_started BOOLEAN DEFAULT false,
ADD COLUMN auto_completed BOOLEAN DEFAULT false;
```

**RabbitMQ Event**:
```json
{
  "event_type": "maintenance.reminder",
  "tenant_id": "...",
  "maintenance_id": 123,
  "title": "Database Upgrade",
  "start_time": "2025-10-22T14:00:00Z",
  "minutes_until_start": 60
}
```

**Test**:
```bash
# Create maintenance window starting in 60 minutes
# Verify reminder sent to subscribers
```

#### Day 2: Auto-Transition Implementation

**Backend (monitoring-service)**:
```go
// Update maintenance_reminder_job.go
func (j *MaintenanceReminderJob) checkStartTransitions() {
    now := time.Now()

    var windows []MaintenanceWindow
    j.db.Where("scheduled_start <= ?", now).
        Where("status = ?", "scheduled").
        Where("auto_started = ?", false).
        Find(&windows)

    for _, window := range windows {
        // Transition to in_progress
        j.db.Model(&window).Updates(map[string]interface{}{
            "status": "in_progress",
            "actual_start": now,
            "auto_started": true,
        })

        // Publish event
        publishEvent("beakon.maintenance.started", window)

        j.logger.Info("Auto-started maintenance", zap.Uint("id", window.ID))
    }
}
```

**Test**:
```bash
# Create maintenance scheduled for now
# Wait 1 minute
# Verify status changed to "in_progress"
```

#### Day 3: Auto-Completion Implementation

**Backend (monitoring-service)**:
```go
func (j *MaintenanceReminderJob) checkCompletionTransitions() {
    now := time.Now()

    var windows []MaintenanceWindow
    j.db.Where("scheduled_end <= ?", now).
        Where("status = ?", "in_progress").
        Where("auto_completed = ?", false).
        Find(&windows)

    for _, window := range windows {
        // Transition to completed
        j.db.Model(&window).Updates(map[string]interface{}{
            "status": "completed",
            "actual_end": now,
            "auto_completed": true,
        })

        // Publish event
        publishEvent("beakon.maintenance.completed", window)

        j.logger.Info("Auto-completed maintenance", zap.Uint("id", window.ID))
    }
}
```

**Integration**: Wire up in main.go
```go
maintenanceReminderJob := jobs.NewMaintenanceReminderJob(dbManager.GetDB(), logger)
go maintenanceReminderJob.Start()
```

**Frontend UI** (tenant-admin-frontend):
```typescript
// Add toggle in maintenance creation form
<Checkbox
  label="Enable automatic transitions"
  checked={autoTransitions}
  onChange={(e) => setAutoTransitions(e.target.checked)}
/>
<p className="text-sm text-gray-500">
  Automatically send reminders, start, and complete maintenance windows
</p>
```

**Deliverables**:
- ✅ Auto-reminder 60 minutes before start
- ✅ Auto-transition to "in_progress" at start time
- ✅ Auto-completion at end time
- ✅ Admin UI toggle to enable/disable
- ✅ RabbitMQ events published for each transition

---

### 2. Alert Auto-Resolution (2 days)

**Goal**: Auto-resolve alerts when health checks pass

#### Day 1: Backend Implementation

**Backend (monitoring-service)**:
```go
// internal/services/alert_service.go
func (s *AlertService) AutoResolveAlerts(monitorID uint, location string) error {
    // Find active alerts for this monitor
    var alerts []Alert
    s.db.Where("monitor_id = ?", monitorID).
        Where("status IN ?", []string{"triggered", "acknowledged"}).
        Find(&alerts)

    for _, alert := range alerts {
        // Check if monitor is now operational
        if s.isMonitorOperational(monitorID, location) {
            // Auto-resolve
            s.db.Model(&alert).Updates(map[string]interface{}{
                "status": "resolved",
                "resolved_at": time.Now(),
                "resolution_type": "auto",
                "resolution_note": "Monitor returned to operational status",
            })

            // Publish event
            publishEvent("beakon.alerts.auto_resolved", alert)

            s.logger.Info("Auto-resolved alert", zap.Uint("alert_id", alert.ID))
        }
    }

    return nil
}

func (s *AlertService) isMonitorOperational(monitorID uint, location string) bool {
    // Check last 3 consecutive checks are operational
    var results []MonitoringResult
    s.db.Where("monitor_id = ?", monitorID).
        Where("location = ?", location).
        Order("checked_at DESC").
        Limit(3).
        Find(&results)

    if len(results) < 3 {
        return false
    }

    for _, r := range results {
        if r.Status != "operational" {
            return false
        }
    }

    return true
}
```

**Integration**: Call after successful health check
```go
// After recording successful check
if checkResult.Status == "operational" {
    alertService.AutoResolveAlerts(monitor.ID, location)
}
```

**Database Migration**:
```sql
ALTER TABLE alerts
ADD COLUMN resolution_type VARCHAR(20), -- auto, manual
ADD COLUMN resolution_note TEXT;
```

#### Day 2: Testing & UI

**Test**:
```bash
# 1. Trigger alert (fail health check 3 times)
# 2. Pass health check 3 times
# 3. Verify alert auto-resolved
```

**Frontend UI**: Show auto-resolution indicator
```typescript
{alert.resolution_type === 'auto' && (
  <Badge variant="success">
    <CheckCircle2 className="h-3 w-3" />
    Auto-Resolved
  </Badge>
)}
```

**Deliverables**:
- ✅ Auto-resolve alerts after 3 consecutive successes
- ✅ RabbitMQ event published
- ✅ UI indicator for auto-resolved alerts
- ✅ Configuration option (enable/disable per monitor)

---

### 3. Alert Deduplication (2 days)

**Goal**: Prevent duplicate alerts for the same issue

#### Day 1: Backend Implementation

**Backend (monitoring-service)**:
```go
func (s *AlertService) CreateAlert(monitor *Monitor, checkResult *MonitoringResult) error {
    // Generate deduplication key
    dedupKey := s.generateDedupKey(monitor.ID, checkResult.ErrorType)

    // Check if alert already exists (within last 15 minutes)
    var existingAlert Alert
    err := s.db.Where("dedup_key = ?", dedupKey).
        Where("created_at > ?", time.Now().Add(-15*time.Minute)).
        Where("status IN ?", []string{"triggered", "acknowledged"}).
        First(&existingAlert).Error

    if err == nil {
        // Alert already exists, skip
        s.logger.Info("Skipping duplicate alert",
            zap.String("dedup_key", dedupKey),
            zap.Uint("existing_id", existingAlert.ID))
        return nil
    }

    // Create new alert
    alert := &Alert{
        MonitorID:   monitor.ID,
        DedupKey:    dedupKey,
        Status:      "triggered",
        Severity:    s.calculateSeverity(checkResult),
        Message:     checkResult.ErrorMessage,
        TriggeredAt: time.Now(),
    }

    return s.db.Create(alert).Error
}

func (s *AlertService) generateDedupKey(monitorID uint, errorType string) string {
    return fmt.Sprintf("monitor-%d-%s", monitorID, errorType)
}
```

**Database Migration**:
```sql
ALTER TABLE alerts
ADD COLUMN dedup_key VARCHAR(255),
ADD COLUMN error_type VARCHAR(50);

CREATE INDEX idx_alerts_dedup ON alerts(dedup_key, created_at DESC)
WHERE status IN ('triggered', 'acknowledged');
```

#### Day 2: Configuration UI

**Frontend**: Add deduplication settings
```typescript
<FormField>
  <Label>Alert Deduplication Window</Label>
  <Select value={dedupWindow} onValueChange={setDedupWindow}>
    <SelectItem value="5">5 minutes</SelectItem>
    <SelectItem value="15">15 minutes</SelectItem>
    <SelectItem value="30">30 minutes</SelectItem>
    <SelectItem value="60">1 hour</SelectItem>
  </Select>
  <p className="text-sm text-gray-500">
    Suppress duplicate alerts within this time window
  </p>
</FormField>
```

**Deliverables**:
- ✅ Alert deduplication based on monitor + error type
- ✅ Configurable deduplication window
- ✅ Admin UI for configuration
- ✅ Metrics: track deduplicated alert count

---

## 🎯 Success Criteria

### Week 1-2 Success Metrics

**Maintenance Automation**:
- ✅ 100% of scheduled maintenance receive reminders
- ✅ Auto-transitions occur within 1 minute of scheduled time
- ✅ No manual intervention required for routine maintenance

**Alert Improvements**:
- ✅ 80% of alerts auto-resolve when monitors recover
- ✅ 50% reduction in duplicate alerts
- ✅ Average alert noise reduced by 40%

**Incident Management**:
- ✅ 100% of incidents have assigned priority
- ✅ 100% of critical incidents have assigned owner
- ✅ Average incident assignment time < 5 minutes

---

## 📋 Next Steps

**After reviewing this plan, we can**:

1. **Start with Quick Wins** (Recommended)
   - Implement Maintenance Automation (Week 1)
   - High impact, low complexity
   - Builds momentum

2. **Jump to High-Impact Features**
   - Start with Performance Metrics (P50, P95, P99)
   - Customer-requested feature
   - Competitive advantage

3. **Focus on Specific Category**
   - Choose one category (e.g., Analytics)
   - Complete all P1 features in that category
   - Achieve depth before breadth

**Which approach would you like to take?**

---

**Document Status**: Ready for Implementation
**Estimated Duration**: 4-6 weeks (all P1 features)
**Dependencies**: All P0 features complete ✅
**Next Action**: Choose implementation approach and start coding!
