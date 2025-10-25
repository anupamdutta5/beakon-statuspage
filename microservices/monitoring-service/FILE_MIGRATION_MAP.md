# Monitoring Service - File Migration Mapping

**Date**: 2025-10-25
**Status**: 🟡 In Progress
**Purpose**: Map current layer-based files to new feature-based structure

---

## BEFORE → AFTER STRUCTURE

### CURRENT STRUCTURE (Layer-Based)
```
internal/
├── services/         (38 files)
├── handlers/         (13 files)
├── models/           (12 files)
├── jobs/             (4 files)
├── events/           (1 file)
├── database/         (2 files)
├── middleware/       (1 file)
├── config/           (1 file)
├── utils/            (1 file)
└── shutdown/         (2 files)
```

### NEW STRUCTURE (Feature-Based)
```
internal/
├── features/
│   ├── monitors/
│   ├── alerts/
│   ├── maintenance/
│   ├── integrations/
│   ├── anomaly/
│   ├── locations/
│   ├── sla/
│   ├── heartbeat/
│   ├── escalation/
│   ├── performance/
│   ├── status_automation/
│   ├── docker/
│   ├── kubernetes/
│   └── external_monitoring/
└── core/
    ├── database/
    ├── events/
    ├── middleware/
    ├── errors/
    ├── validation/
    └── config/
```

---

## FILE MIGRATION MAPPINGS

### 1. MONITORS FEATURE

#### HTTP Monitoring
- **From**: `internal/services/monitor_service.go`
- **To**: `internal/features/monitors/http/service.go`
- **Size**: ~500 lines
- **Dependencies**: Monitor model, health_check_service

- **From**: `internal/services/monitoring_service.go`
- **To**: `internal/features/monitors/http/monitoring_service.go`
- **Size**: ~400 lines
- **Dependencies**: Component monitoring, external monitoring

- **From**: `internal/services/health_check_service.go`
- **To**: `internal/features/monitors/http/health_check_service.go`
- **Size**: ~300 lines

- **From**: `internal/models/monitor.go`
- **To**: `internal/features/monitors/models.go`
- **Size**: ~200 lines
- **Note**: May need to decompose this large model

- **From**: `internal/handlers/monitoring_handler.go`
- **To**: `internal/features/monitors/http/handler.go`
- **Size**: ~600 lines

#### TCP Monitoring
- **From**: `internal/services/tcp_port_monitor.go`
- **To**: `internal/features/monitors/tcp/service.go`
- **Size**: ~250 lines

#### Ping Monitoring
- **From**: `internal/services/ping_monitor.go`
- **To**: `internal/features/monitors/ping/service.go`
- **Size**: ~200 lines

#### SSL Monitoring
- **From**: `internal/services/ssl_scanner_service.go`
- **To**: `internal/features/monitors/ssl/service.go`
- **Size**: ~400 lines

- **From**: `internal/models/ssl_certificate.go`
- **To**: `internal/features/monitors/ssl/models.go`
- **Size**: ~150 lines

- **From**: `internal/handlers/ssl_handler.go`
- **To**: `internal/features/monitors/ssl/handler.go`
- **Size**: ~300 lines

- **From**: `internal/jobs/ssl_expiration_checker.go`
- **To**: `internal/features/monitors/ssl/expiration_job.go`
- **Size**: ~400 lines
- **Note**: Contains MaintenanceWindowJob, HeartbeatCheckerJob, EscalationProcessorJob, WebhookRetryJob

#### DNS Monitoring
- **From**: `internal/services/dns_monitor.go`
- **To**: `internal/features/monitors/dns/service.go`
- **Size**: ~200 lines

---

### 2. ALERTS FEATURE

#### Core Alert Management
- **From**: `internal/services/alert_service.go`
- **To**: `internal/features/alerts/core/service.go`
- **Size**: ~310 lines
- **Note**: P1 feature - Auto-resolution and deduplication

- **From**: `internal/models/monitoring.go` (Alert struct)
- **To**: `internal/features/alerts/core/models.go`
- **Size**: Extract Alert-related code

#### Alert Routing
- **From**: `internal/services/alert_routing_service.go`
- **To**: `internal/features/alerts/routing/service.go`
- **Size**: ~400 lines

#### Auto-Resolution
- **From**: `internal/services/alert_service.go` (AutoResolveAlerts method)
- **To**: `internal/features/alerts/auto_resolution/service.go`
- **Size**: Extract auto-resolution logic

#### Deduplication
- **From**: `internal/services/alert_service.go` (CreateAlert with dedup)
- **To**: `internal/features/alerts/deduplication/service.go`
- **Size**: Extract deduplication logic

---

### 3. MAINTENANCE FEATURE

#### Windows Management
- **From**: `internal/services/maintenance_management_service.go`
- **To**: `internal/features/maintenance/windows/service.go`
- **Size**: ~539 lines

- **From**: `internal/models/maintenance_management.go`
- **To**: `internal/features/maintenance/windows/models.go`
- **Size**: ~200 lines

#### Automation
- **From**: `internal/services/maintenance_management_service.go` (Auto methods)
- **To**: `internal/features/maintenance/automation/service.go`
- **Size**: Extract automation logic (SendMaintenanceReminders, AutoStartMaintenanceWindows, AutoCompleteMaintenanceWindows)

- **From**: `internal/jobs/ssl_expiration_checker.go` (MaintenanceWindowJob)
- **To**: `internal/features/maintenance/automation/job.go`
- **Size**: Extract MaintenanceWindowJob

#### Scheduling
- **From**: `internal/services/maintenance_service.go`
- **To**: `internal/features/maintenance/scheduling/service.go`
- **Size**: ~200 lines

---

### 4. INTEGRATIONS FEATURE

#### Slack Integration
- **From**: `internal/services/slack_integration.go`
- **To**: `internal/features/integrations/slack/service.go`
- **Size**: ~400 lines

- **From**: `internal/handlers/slack_handler.go`
- **To**: `internal/features/integrations/slack/handler.go`
- **Size**: ~300 lines

#### PagerDuty Integration
- **From**: `internal/services/pagerduty_integration.go`
- **To**: `internal/features/integrations/pagerduty/service.go`
- **Size**: ~500 lines

- **From**: `internal/handlers/pagerduty_handler.go`
- **To**: `internal/features/integrations/pagerduty/handler.go`
- **Size**: ~600 lines

#### Discord Integration
- **From**: `internal/services/discord_integration.go`
- **To**: `internal/features/integrations/discord/service.go`
- **Size**: ~400 lines

- **From**: `internal/models/discord_integration.go`
- **To**: `internal/features/integrations/discord/models.go`
- **Size**: ~150 lines

- **From**: `internal/handlers/discord_handler.go`
- **To**: `internal/features/integrations/discord/handler.go`
- **Size**: ~400 lines

#### Telegram Integration
- **From**: `internal/services/telegram_integration.go`
- **To**: `internal/features/integrations/telegram/service.go`
- **Size**: ~400 lines

- **From**: `internal/models/telegram_integration.go`
- **To**: `internal/features/integrations/telegram/models.go`
- **Size**: ~150 lines

- **From**: `internal/handlers/telegram_handler.go`
- **To**: `internal/features/integrations/telegram/handler.go`
- **Size**: ~400 lines

#### Teams Integration
- **From**: `internal/services/teams_integration.go`
- **To**: `internal/features/integrations/teams/service.go`
- **Size**: ~300 lines

#### Email Integration
- **From**: `internal/services/email_integration.go`
- **To**: `internal/features/integrations/email/service.go`
- **Size**: ~200 lines

#### Webhook Integration
- **From**: `internal/services/webhook_service.go`
- **To**: `internal/features/integrations/webhook/service.go`
- **Size**: ~300 lines

- **From**: `internal/services/webhook_integration.go`
- **To**: `internal/features/integrations/webhook/integration_service.go`
- **Size**: ~250 lines

- **From**: `internal/models/webhook.go`
- **To**: `internal/features/integrations/webhook/models.go`
- **Size**: ~200 lines

- **From**: `internal/handlers/webhook_handler.go`
- **To**: `internal/features/integrations/webhook/handler.go`
- **Size**: ~400 lines

---

### 5. ANOMALY FEATURE

#### Detection
- **From**: `internal/services/anomaly_detection_service.go`
- **To**: `internal/features/anomaly/detection/service.go`
- **Size**: ~500 lines

- **From**: `internal/handlers/anomaly_handler.go`
- **To**: `internal/features/anomaly/detection/handler.go`
- **Size**: ~300 lines

#### Baselines
- **From**: `internal/services/baseline_calculator.go`
- **To**: `internal/features/anomaly/baselines/calculator.go`
- **Size**: ~400 lines

- **From**: `internal/jobs/baseline_update_job.go`
- **To**: `internal/features/anomaly/baselines/update_job.go`
- **Size**: ~150 lines

- **From**: `internal/models/anomaly.go`
- **To**: `internal/features/anomaly/models.go`
- **Size**: ~200 lines

#### ML Models (Future)
- **To**: `internal/features/anomaly/ml_models/`
- **Note**: Placeholder for future machine learning models

---

### 6. LOCATIONS FEATURE

#### Multi-Region
- **From**: `internal/services/multi_location_checker.go`
- **To**: `internal/features/locations/multi_region/service.go`
- **Size**: ~300 lines

- **From**: `internal/models/location.go`
- **To**: `internal/features/locations/models.go`
- **Size**: ~100 lines

#### Failover (Future)
- **To**: `internal/features/locations/failover/`
- **Note**: Placeholder for future failover logic

---

### 7. SLA FEATURE

#### Reporting
- **From**: `internal/services/sla_reporting.go`
- **To**: `internal/features/sla/reporting/service.go`
- **Size**: ~400 lines

#### Calculations
- **From**: `internal/services/mttr_mttd_tracking.go`
- **To**: `internal/features/sla/calculations/mttr_mttd.go`
- **Size**: ~300 lines

---

### 8. HEARTBEAT FEATURE

- **From**: `internal/services/heartbeat_service.go`
- **To**: `internal/features/heartbeat/service.go`
- **Size**: ~300 lines

- **From**: `internal/handlers/heartbeat_handler.go`
- **To**: `internal/features/heartbeat/handler.go`
- **Size**: ~200 lines

- **From**: `internal/jobs/ssl_expiration_checker.go` (HeartbeatCheckerJob)
- **To**: `internal/features/heartbeat/checker_job.go`
- **Size**: Extract HeartbeatCheckerJob

---

### 9. ESCALATION FEATURE

- **From**: `internal/services/escalation_service.go`
- **To**: `internal/features/escalation/service.go`
- **Size**: ~400 lines

- **From**: `internal/handlers/escalation_handler.go`
- **To**: `internal/features/escalation/handler.go`
- **Size**: ~300 lines

- **From**: `internal/services/oncall_service.go`
- **To**: `internal/features/escalation/oncall_service.go`
- **Size**: ~350 lines

- **From**: `internal/handlers/oncall_handler.go`
- **To**: `internal/features/escalation/oncall_handler.go`
- **Size**: ~250 lines

- **From**: `internal/services/sms_service.go`
- **To**: `internal/features/escalation/sms_service.go`
- **Size**: ~200 lines

- **From**: `internal/jobs/ssl_expiration_checker.go` (EscalationProcessorJob)
- **To**: `internal/features/escalation/processor_job.go`
- **Size**: Extract EscalationProcessorJob

---

### 10. PERFORMANCE FEATURE

- **From**: `internal/services/performance_metrics.go`
- **To**: `internal/features/performance/metrics_service.go`
- **Size**: ~300 lines

---

### 11. STATUS AUTOMATION FEATURE

- **From**: `internal/services/status_automation_service.go`
- **To**: `internal/features/status_automation/service.go`
- **Size**: ~300 lines

- **From**: `internal/models/status_automation.go`
- **To**: `internal/features/status_automation/models.go`
- **Size**: ~150 lines

---

### 12. DOCKER FEATURE

- **From**: `internal/services/docker_monitoring_service.go`
- **To**: `internal/features/docker/service.go`
- **Size**: ~400 lines

- **From**: `internal/models/docker_monitoring.go`
- **To**: `internal/features/docker/models.go`
- **Size**: ~150 lines

---

### 13. KUBERNETES FEATURE

- **From**: `internal/services/kubernetes_monitoring_service.go`
- **To**: `internal/features/kubernetes/service.go`
- **Size**: ~400 lines

- **From**: `internal/models/kubernetes_monitoring.go`
- **To**: `internal/features/kubernetes/models.go`
- **Size**: ~150 lines

---

### 14. EXTERNAL MONITORING FEATURE

- **From**: `internal/services/external_monitoring_service.go`
- **To**: `internal/features/external_monitoring/service.go`
- **Size**: ~300 lines

- **From**: `internal/models/external_monitoring.go`
- **To**: `internal/features/external_monitoring/models.go`
- **Size**: ~100 lines

---

### 15. CUSTOM METRICS FEATURE

- **From**: `internal/services/custom_metrics_service.go`
- **To**: `internal/features/performance/custom_metrics_service.go`
- **Size**: ~200 lines

---

### 16. COMPONENT MONITORING FEATURE

- **From**: `internal/services/component_monitoring_service.go`
- **To**: `internal/features/monitors/http/component_service.go`
- **Size**: ~300 lines

- **From**: `internal/models/component_monitoring.go`
- **To**: `internal/features/monitors/http/component_models.go`
- **Size**: ~150 lines

---

### 17. INTEGRATION SERVICE

- **From**: `internal/services/integration_service.go`
- **To**: `internal/features/integrations/core_service.go`
- **Size**: ~300 lines

- **From**: `internal/models/integration.go`
- **To**: `internal/features/integrations/models.go`
- **Size**: ~150 lines

- **From**: `internal/handlers/integration_handler.go`
- **To**: `internal/features/integrations/handler.go`
- **Size**: ~300 lines

---

### 18. NOTIFICATION THROTTLING

- **From**: `internal/services/notification_throttling_service.go`
- **To**: `internal/features/alerts/core/throttling_service.go`
- **Size**: ~200 lines

---

### 19. MONITORING MANAGER

- **From**: `internal/services/monitoring_manager.go`
- **To**: `internal/features/monitors/manager.go`
- **Size**: ~300 lines
- **Note**: May need refactoring to determine if needed

---

### 20. PUBLIC METRICS HANDLER

- **From**: `internal/handlers/public_metrics_handler.go`
- **To**: `internal/features/performance/public_handler.go`
- **Size**: ~200 lines

---

## CORE UTILITIES MIGRATION

### Database
- **From**: `internal/database/manager.go`
- **To**: `internal/core/database/manager.go`
- **Size**: ~200 lines

- **From**: `internal/database/logger.go`
- **To**: `internal/core/database/logger.go`
- **Size**: ~100 lines

### Events
- **From**: `internal/events/publisher.go`
- **To**: `internal/core/events/publisher.go`
- **Size**: ~300 lines

### Middleware
- **From**: `internal/middleware/middleware.go`
- **To**: `internal/core/middleware/middleware.go`
- **Size**: ~150 lines

### Config
- **From**: `internal/config/config.go`
- **To**: `internal/core/config/config.go`
- **Size**: ~200 lines

### Validation
- **From**: `internal/utils/validation.go`
- **To**: `internal/core/validation/validation.go`
- **Size**: ~100 lines

### Shutdown
- **From**: `internal/shutdown/manager.go`
- **To**: `internal/core/shutdown/manager.go`
- **Size**: ~150 lines

- **From**: `internal/shutdown/example_integration.go`
- **To**: `internal/core/shutdown/example_integration.go`
- **Size**: ~100 lines

---

## JOBS MIGRATION

### Metric Collection
- **From**: `internal/jobs/metric_collection_job.go`
- **To**: `internal/features/anomaly/detection/metric_collection_job.go`
- **Size**: ~200 lines

### Cleanup
- **From**: `internal/jobs/cleanup_job.go`
- **To**: `internal/core/jobs/cleanup_job.go`
- **Size**: ~250 lines

### SSL Expiration Checker
- **From**: `internal/jobs/ssl_expiration_checker.go`
- **To**: Split into multiple feature jobs:
  - `internal/features/monitors/ssl/expiration_job.go` (SSL expiration)
  - `internal/features/maintenance/automation/job.go` (Maintenance window)
  - `internal/features/heartbeat/checker_job.go` (Heartbeat)
  - `internal/features/escalation/processor_job.go` (Escalation)
  - `internal/features/integrations/webhook/retry_job.go` (Webhook retry)
- **Size**: ~400 lines total, split into 5 files

---

## SUMMARY STATISTICS

| Category | Files to Move | Estimated Lines | Priority |
|----------|---------------|-----------------|----------|
| Monitors | 15 files | ~4,000 lines | HIGH |
| Alerts | 5 files | ~1,200 lines | HIGH |
| Maintenance | 5 files | ~900 lines | HIGH |
| Integrations | 20 files | ~6,500 lines | MEDIUM |
| Anomaly | 5 files | ~1,300 lines | MEDIUM |
| Locations | 2 files | ~400 lines | LOW |
| SLA | 2 files | ~700 lines | MEDIUM |
| Heartbeat | 3 files | ~500 lines | MEDIUM |
| Escalation | 6 files | ~1,500 lines | MEDIUM |
| Performance | 3 files | ~700 lines | LOW |
| Status Automation | 2 files | ~450 lines | LOW |
| Docker | 2 files | ~550 lines | LOW |
| Kubernetes | 2 files | ~550 lines | LOW |
| External Monitoring | 2 files | ~400 lines | LOW |
| Core Utilities | 8 files | ~1,300 lines | HIGH |
| **TOTAL** | **82 files** | **~21,450 lines** | - |

---

## VALIDATION CHECKPOINTS

After each feature migration:
- [ ] All files moved to correct locations
- [ ] Import paths updated
- [ ] Service compiles without errors
- [ ] Existing tests pass
- [ ] No duplicate code
- [ ] Feature routes extracted from main.go
- [ ] README updated

---

**Status**: 🟡 In Progress - Directory structure created
**Next Step**: Begin systematic file migration starting with Core Utilities
**Updated**: 2025-10-25
