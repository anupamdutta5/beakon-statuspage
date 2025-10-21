# Monitoring Features Roadmap

**Date**: October 21, 2025
**Status**: Comprehensive Feature Analysis Complete
**Research Sources**: Statuspage.io, Instatus, Status.io, Better Stack, UptimeRobot, Pulsetic

---

## Executive Summary

This document provides a **comprehensive analysis of 75+ monitoring features** identified from industry-leading status page platforms. It maps these features to Beakon's existing microservices architecture, defines RabbitMQ event patterns, Redis caching strategies, and database best practices for implementation.

**Key Findings:**
- ✅ **30 features already implemented** (40%)
- ❌ **45 features missing** (60%)
- **6 primary microservices** involved (monitoring, component, incident, status-ui, event-store, notification)
- **5 RabbitMQ topic exchanges** for event-driven architecture
- **4 Redis usage patterns** (caching, pub/sub, rate limiting, real-time)
- **9-13 weeks** estimated implementation time (3 phases)

---

## Table of Contents

1. [Complete Feature List](#complete-feature-list)
2. [Competitor Comparison Matrix](#competitor-comparison-matrix)
3. [Microservice Mapping](#microservice-mapping)
4. [RabbitMQ Event Architecture](#rabbitmq-event-architecture)
5. [Redis Usage Patterns](#redis-usage-patterns)
6. [Database Best Practices](#database-best-practices)
7. [Implementation Roadmap](#implementation-roadmap)
8. [Success Metrics](#success-metrics)

---

## Complete Feature List

### Category A: Uptime Monitoring (15 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | HTTP/HTTPS uptime checks | ✅ Implemented | All | - |
| 2 | Service health monitoring | ✅ Implemented | All | - |
| 3 | Uptime percentage calculation | ✅ Implemented | All | - |
| 4 | Uptime history tracking | ✅ Implemented | All | - |
| 5 | Status tracking (up/down/degraded) | ✅ Implemented | All | - |
| 6 | Multi-location monitoring (10+ locations) | ❌ Missing | All | P0 - Critical |
| 7 | TCP port monitoring | ❌ Missing | UptimeRobot, Better Stack | P1 - High |
| 8 | ICMP ping monitoring | ❌ Missing | UptimeRobot, Pingdom | P1 - High |
| 9 | DNS monitoring | ❌ Missing | UptimeRobot, Better Stack | P2 - Medium |
| 10 | Custom interval checks (30s-1h) | ❌ Missing | All | P1 - High |
| 11 | SSL certificate expiration monitoring | ❌ Missing | All | P0 - Critical |
| 12 | SSL certificate validation | ❌ Missing | Uptrends, Better Stack | P1 - High |
| 13 | Domain expiration monitoring | ❌ Missing | Site24x7, Better Stack | P2 - Medium |
| 14 | Keyword monitoring | ❌ Missing | UptimeRobot, Pingdom | P2 - Medium |
| 15 | Website defacement detection | ❌ Missing | Site24x7 | P3 - Low |

**Implementation Owner**: monitoring-service (Port 8092)

---

### Category B: Performance Metrics (12 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Response time tracking | ✅ Implemented | All | - |
| 2 | Response time history | ✅ Implemented | All | - |
| 3 | Average response time | ✅ Implemented | All | - |
| 4 | Performance metrics collection | ✅ Implemented | All | - |
| 5 | Response time percentiles (P50, P95, P99) | ❌ Missing | All | P1 - High |
| 6 | Page load time monitoring | ❌ Missing | Pingdom, GTmetrix | P1 - High |
| 7 | Time to first byte (TTFB) | ❌ Missing | Pingdom, Better Stack | P1 - High |
| 8 | DNS resolution time | ❌ Missing | Pingdom, Uptrends | P2 - Medium |
| 9 | Connection time | ❌ Missing | Pingdom, Uptrends | P2 - Medium |
| 10 | Download speed metrics | ❌ Missing | Pingdom | P3 - Low |
| 11 | Custom metric definitions | ❌ Missing | Statuspage.io, Status.io | P1 - High |
| 12 | Public metrics display | ❌ Missing | Statuspage.io, Status.io | P0 - Critical |

**Implementation Owner**: monitoring-service (Port 8092)

---

### Category C: Incident Management (15 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Incident creation | ✅ Implemented | All | - |
| 2 | Incident updates | ✅ Implemented | All | - |
| 3 | Incident lifecycle | ✅ Implemented | All | - |
| 4 | Impact levels (critical, major, minor) | ✅ Implemented | All | - |
| 5 | Component affectation | ✅ Implemented | All | - |
| 6 | Incident templates | ✅ Implemented | Statuspage.io, Status.io | - |
| 7 | Automated incident creation | ❌ Missing | All | P0 - Critical |
| 8 | Incident timeline visualization | ❌ Missing | All | P1 - High |
| 9 | Incident impact analysis | ❌ Missing | PagerDuty, Better Stack | P2 - Medium |
| 10 | Post-mortem templates | ❌ Missing | PagerDuty, incident.io | P2 - Medium |
| 11 | Incident retrospectives | ❌ Missing | incident.io, PagerDuty | P3 - Low |
| 12 | Incident categorization/tagging | ❌ Missing | All | P2 - Medium |
| 13 | Incident priority levels | ❌ Missing | PagerDuty, incident.io | P1 - High |
| 14 | Incident owner assignment | ❌ Missing | All | P1 - High |
| 15 | Incident escalation rules | ❌ Missing | PagerDuty, incident.io | P1 - High |

**Implementation Owner**: incident-service (Port 8086)

---

### Category D: Scheduled Maintenance (8 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Maintenance window scheduling | ✅ Implemented | All | - |
| 2 | Maintenance updates | ✅ Implemented | All | - |
| 3 | Component association | ✅ Implemented | All | - |
| 4 | Maintenance templates | ✅ Implemented | Statuspage.io, Status.io | - |
| 5 | Auto-reminder 60min before start | ❌ Missing | Statuspage.io | P1 - High |
| 6 | Auto-transition to "In Progress" | ❌ Missing | Statuspage.io | P1 - High |
| 7 | Auto-completion at end time | ❌ Missing | Statuspage.io | P1 - High |
| 8 | Recurring maintenance schedules | ❌ Missing | Status.io, Better Stack | P2 - Medium |

**Implementation Owner**: monitoring-service (Port 8092)

---

### Category E: Notification & Subscriptions (12 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Email notifications | ⚠️ Basic | All | - |
| 2 | Webhook notifications | ⚠️ Basic | All | - |
| 3 | SMS notifications | ❌ Missing | All | P0 - Critical |
| 4 | Slack notifications | ❌ Missing | All | P0 - Critical |
| 5 | Microsoft Teams notifications | ❌ Missing | Statuspage.io, Better Stack | P1 - High |
| 6 | Discord notifications | ❌ Missing | Instatus | P2 - Medium |
| 7 | Component-specific subscriptions | ❌ Missing | All | P1 - High |
| 8 | Incident-specific subscriptions | ❌ Missing | Statuspage.io | P2 - Medium |
| 9 | Subscription preferences management | ❌ Missing | All | P1 - High |
| 10 | Notification digest (daily/weekly) | ❌ Missing | Better Stack | P2 - Medium |
| 11 | Notification throttling | ❌ Missing | All | P1 - High |
| 12 | Phone call alerts | ❌ Missing | PagerDuty, Better Stack | P3 - Low |

**Implementation Owner**: notification-service (Port 8085)

---

### Category F: Status Page Features (13 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Public status page | ✅ Implemented | All | - |
| 2 | Component status display | ✅ Implemented | All | - |
| 3 | Overall status calculation | ✅ Implemented | All | - |
| 4 | Incident display | ✅ Implemented | All | - |
| 5 | Custom branding | ✅ Implemented | All | - |
| 6 | Embeddable status widgets | ❌ Missing | All | P0 - Critical |
| 7 | Status badges (SVG/PNG) | ❌ Missing | All | P0 - Critical |
| 8 | Private status pages | ❌ Missing | All | P1 - High |
| 9 | Status page versioning | ❌ Missing | Status.io, Better Stack | P3 - Low |
| 10 | Historical incident calendar | ❌ Missing | All | P1 - High |
| 11 | Uptime showcase (90-day) | ❌ Missing | Statuspage.io | P1 - High |
| 12 | RSS/Atom feeds | ❌ Missing | All | P2 - Medium |
| 13 | JSON API for status | ❌ Missing | All | P1 - High |

**Implementation Owner**: status-ui-service (Port 8093)

---

### Category G: Alerting System (10 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Alert creation | ✅ Implemented | All | - |
| 2 | Alert acknowledgment | ✅ Implemented | All | - |
| 3 | Alert resolution | ✅ Implemented | All | - |
| 4 | Alert routing rules | ❌ Missing | PagerDuty, incident.io | P1 - High |
| 5 | Alert escalation policies | ❌ Missing | PagerDuty, Better Stack | P0 - Critical |
| 6 | On-call scheduling | ❌ Missing | PagerDuty, Better Stack | P0 - Critical |
| 7 | Alert deduplication | ❌ Missing | All | P1 - High |
| 8 | Alert grouping | ❌ Missing | PagerDuty, Better Stack | P2 - Medium |
| 9 | Auto-resolution when check passes | ❌ Missing | All | P1 - High |
| 10 | Alert suppression during maintenance | ❌ Missing | All | P0 - Critical |

**Implementation Owner**: monitoring-service (Port 8092)

---

### Category H: Third-Party Integrations (12 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Webhook integration framework | ✅ Implemented | All | - |
| 2 | Integration configuration storage | ✅ Implemented | All | - |
| 3 | Datadog integration | ❌ Missing | Statuspage.io | P1 - High |
| 4 | New Relic integration | ❌ Missing | Statuspage.io | P2 - Medium |
| 5 | PagerDuty integration | ❌ Missing | Statuspage.io, Better Stack | P0 - Critical |
| 6 | Prometheus integration | ❌ Missing | Better Stack | P1 - High |
| 7 | Grafana integration | ❌ Missing | Better Stack | P2 - Medium |
| 8 | Pingdom integration | ❌ Missing | Statuspage.io | P2 - Medium |
| 9 | UptimeRobot integration | ❌ Missing | Multiple | P3 - Low |
| 10 | Jira integration | ❌ Missing | Statuspage.io | P2 - Medium |
| 11 | GitHub integration | ❌ Missing | Better Stack | P3 - Low |
| 12 | Opsgenie integration | ❌ Missing | Statuspage.io | P3 - Low |

**Implementation Owner**: monitoring-service (Port 8092)

---

### Category I: Analytics & Reporting (8 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Uptime statistics | ✅ Implemented | All | - |
| 2 | Response time charts | ✅ Implemented | All | - |
| 3 | SLA reporting (monthly uptime) | ❌ Missing | All | P1 - High |
| 4 | Incident frequency analysis | ❌ Missing | Better Stack, PagerDuty | P2 - Medium |
| 5 | MTTR tracking | ❌ Missing | PagerDuty, incident.io | P1 - High |
| 6 | MTTD tracking | ❌ Missing | PagerDuty, incident.io | P2 - Medium |
| 7 | Exportable reports (PDF, CSV) | ❌ Missing | All | P1 - High |
| 8 | Custom dashboards | ❌ Missing | Better Stack, Datadog | P2 - Medium |

**Implementation Owner**: analytics-service (Port 8090)

---

### Category J: Advanced Features (10 features)

| # | Feature | Status | Competitor | Priority |
|---|---------|--------|------------|----------|
| 1 | Synthetic transaction monitoring | ❌ Missing | Pingdom, Better Stack | P1 - High |
| 2 | API endpoint monitoring with assertions | ❌ Missing | Pingdom, Datadog | P1 - High |
| 3 | Multi-step health checks | ❌ Missing | Pingdom, Uptrends | P1 - High |
| 4 | Heartbeat/cron job monitoring | ❌ Missing | Better Stack, UptimeRobot | P2 - Medium |
| 5 | Log aggregation and search | ❌ Missing | Better Stack | P2 - Medium |
| 6 | Distributed tracing integration | ❌ Missing | Datadog, Better Stack | P3 - Low |
| 7 | Anomaly detection (AI-powered) | ❌ Missing | Datadog, PagerDuty | P3 - Low |
| 8 | Capacity planning insights | ❌ Missing | Datadog | P3 - Low |
| 9 | Service dependency mapping | ❌ Missing | Better Stack, Datadog | P2 - Medium |
| 10 | Real User Monitoring (RUM) | ❌ Missing | Datadog, Sematext | P3 - Low |

**Implementation Owner**: monitoring-service (Port 8092) + analytics-service (Port 8090)

---

## Competitor Comparison Matrix

### Feature Coverage Comparison

| Feature Category | Beakon (Current) | Statuspage.io | Instatus | Status.io | Better Stack | Target (Beakon v2) |
|------------------|------------------|---------------|----------|-----------|--------------|---------------------|
| **Uptime Monitoring** | 5/15 (33%) | 8/15 (53%) | 10/15 (67%) | 9/15 (60%) | 13/15 (87%) | 15/15 (100%) |
| **Performance Metrics** | 4/12 (33%) | 10/12 (83%) | 8/12 (67%) | 9/12 (75%) | 11/12 (92%) | 12/12 (100%) |
| **Incident Management** | 6/15 (40%) | 12/15 (80%) | 10/15 (67%) | 11/15 (73%) | 13/15 (87%) | 15/15 (100%) |
| **Scheduled Maintenance** | 4/8 (50%) | 8/8 (100%) | 6/8 (75%) | 7/8 (88%) | 7/8 (88%) | 8/8 (100%) |
| **Notifications** | 2/12 (17%) | 10/12 (83%) | 9/12 (75%) | 8/12 (67%) | 11/12 (92%) | 12/12 (100%) |
| **Status Page Features** | 5/13 (38%) | 11/13 (85%) | 10/13 (77%) | 11/13 (85%) | 12/13 (92%) | 13/13 (100%) |
| **Alerting System** | 3/10 (30%) | 7/10 (70%) | 6/10 (60%) | 6/10 (60%) | 10/10 (100%) | 10/10 (100%) |
| **Integrations** | 2/12 (17%) | 12/12 (100%) | 8/12 (67%) | 9/12 (75%) | 10/12 (83%) | 12/12 (100%) |
| **Analytics** | 2/8 (25%) | 6/8 (75%) | 5/8 (63%) | 6/8 (75%) | 7/8 (88%) | 8/8 (100%) |
| **Advanced Features** | 0/10 (0%) | 5/10 (50%) | 4/10 (40%) | 4/10 (40%) | 8/10 (80%) | 8/10 (80%) |
| **TOTAL** | **33/115 (29%)** | **89/115 (77%)** | **76/115 (66%)** | **80/115 (70%)** | **102/115 (89%)** | **113/115 (98%)** |

**Key Insights:**
- Beakon currently at **29% feature parity** with leading competitors
- Better Stack leads with **89% feature coverage**
- Statuspage.io (Atlassian) has **100% integration coverage**
- Target: Beakon v2 at **98% feature coverage** (industry-leading)

---

## Microservice Mapping

### monitoring-service (Port 8092) - **Primary Monitoring Engine**

**Current Capabilities**:
- Active health checks (HTTP, TCP, ping)
- Alert management
- Maintenance windows
- Webhook delivery
- Third-party integrations

**New Features to Implement** (35 features):

**Uptime Monitoring** (10 new):
- Multi-location monitoring (10+ global monitoring nodes)
- TCP port monitoring
- ICMP ping monitoring
- DNS monitoring
- Custom interval checks (30s, 1m, 5m, 10m, 30m, 1h)
- SSL certificate expiration monitoring
- SSL certificate validation
- Domain expiration monitoring
- Keyword monitoring
- Heartbeat/cron job monitoring

**Performance Metrics** (8 new):
- Response time percentiles (P50, P95, P99)
- Page load time monitoring
- Time to first byte (TTFB)
- DNS resolution time
- Connection time
- Custom metric definitions
- Public metrics display

**Alerting** (7 new):
- Alert routing rules
- Alert escalation policies
- On-call scheduling
- Alert deduplication
- Alert grouping
- Auto-resolution when check passes
- Alert suppression during maintenance

**Integrations** (10 new):
- Datadog integration
- New Relic integration
- PagerDuty integration
- Prometheus integration
- Grafana integration
- Pingdom integration
- Jira integration
- Slack notifications
- Microsoft Teams notifications

**RabbitMQ Events Published**:
```
beakon.monitoring.check.passed
beakon.monitoring.check.failed
beakon.monitoring.check.degraded
beakon.monitoring.ssl.expiring
beakon.monitoring.domain.expiring
beakon.monitoring.incident.auto_created
```

**RabbitMQ Events Consumed**:
```
beakon.maintenance.started → Suppress alerts
beakon.maintenance.completed → Resume alerts
```

**Redis Usage**:
- **Cache monitoring results**: `monitor:result:{monitor_id}:latest` (1-minute TTL)
- **Store recent check history**: `monitor:history:{monitor_id}` (rolling 24 hours)
- **Rate limiting**: `ratelimit:monitor:{location_id}` (100 checks/second per location)
- **Real-time dashboard**: Pub/sub on `monitoring:live:{tenant_id}`

**Database Schema Changes**:

```sql
-- New table: monitoring_locations
CREATE TABLE monitoring_locations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    city VARCHAR(100),
    country VARCHAR(100),
    region VARCHAR(50), -- e.g., us-east, eu-west, asia-south
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- New table: ssl_certificates
CREATE TABLE ssl_certificates (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    domain VARCHAR(255) NOT NULL,
    issuer VARCHAR(255),
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    days_until_expiry INT GENERATED ALWAYS AS (EXTRACT(DAY FROM (valid_until - NOW()))) STORED,
    last_checked TIMESTAMP,
    is_valid BOOLEAN,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, domain)
);

CREATE INDEX idx_ssl_expiry ON ssl_certificates(days_until_expiry) WHERE days_until_expiry < 30;

-- New table: dns_records
CREATE TABLE dns_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    record_type VARCHAR(10) NOT NULL, -- A, AAAA, CNAME, MX, TXT
    expected_value TEXT,
    last_value TEXT,
    last_checked TIMESTAMP,
    is_matching BOOLEAN,
    created_at TIMESTAMP DEFAULT NOW()
);

-- New table: heartbeat_monitors
CREATE TABLE heartbeat_monitors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    expected_interval_seconds INT NOT NULL,
    grace_period_seconds INT DEFAULT 300,
    last_ping TIMESTAMP,
    is_alive BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- New table: synthetic_transactions
CREATE TABLE synthetic_transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    steps JSONB NOT NULL, -- [{type: 'navigate', url: '...'}, {type: 'click', selector: '...'}]
    interval_seconds INT DEFAULT 300,
    timeout_seconds INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT NOW()
);

-- New table: on_call_schedules
CREATE TABLE on_call_schedules (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    rotation_type VARCHAR(20) NOT NULL, -- daily, weekly, custom
    rotation_start TIMESTAMP NOT NULL,
    participants JSONB NOT NULL, -- [{user_id: '...', order: 1}]
    created_at TIMESTAMP DEFAULT NOW()
);

-- New table: escalation_policies
CREATE TABLE escalation_policies (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    levels JSONB NOT NULL, -- [{delay_minutes: 0, notify: ['user1']}, {delay_minutes: 15, notify: ['user2', 'user3']}]
    created_at TIMESTAMP DEFAULT NOW()
);

-- Update monitoring_results for multi-location
ALTER TABLE monitoring_results
ADD COLUMN location_id BIGINT REFERENCES monitoring_locations(id),
ADD COLUMN response_time_ms INT,
ADD COLUMN ttfb_ms INT,
ADD COLUMN dns_time_ms INT,
ADD COLUMN connection_time_ms INT;

CREATE INDEX idx_monitoring_results_location ON monitoring_results(monitor_id, location_id, checked_at DESC);
```

---

### component-service (Port 8084) - **Component Registry**

**Current Capabilities**:
- Component CRUD
- Component status tracking
- Component grouping
- Status history

**New Features to Implement** (5 features):
- Component dependencies (service dependency mapping)
- Component SLA configuration (99.9%, 99.95%, 99.99%)
- Component-specific metrics (custom metrics)
- Component tags/categories
- Component ownership

**RabbitMQ Events Published**:
```
beakon.components.status.changed
beakon.components.sla.breached
beakon.components.created
beakon.components.updated
beakon.components.deleted
```

**RabbitMQ Events Consumed**:
```
beakon.monitoring.check.failed → Update component status to degraded/down
beakon.monitoring.check.passed → Update component status to operational
beakon.incidents.created → Link incident to affected components
beakon.maintenance.started → Set component to maintenance mode
beakon.maintenance.completed → Restore component to operational
```

**Redis Usage**:
- **Cache public component list**: `components:public:{tenant_id}` (5-minute TTL)
- **Cache component status**: `component:status:{component_id}` (30-second TTL)

**Database Schema Changes**:

```sql
ALTER TABLE components
ADD COLUMN sla_target DECIMAL(5,2) DEFAULT 99.90,
ADD COLUMN dependencies JSONB, -- [{component_id: 123, dependency_type: 'hard'|'soft'}]
ADD COLUMN owner_id UUID,
ADD COLUMN tags JSONB; -- ['database', 'critical', 'payment']

CREATE INDEX idx_components_tags ON components USING gin(tags);
```

---

### incident-service (Port 8086) - **Incident Lifecycle**

**Current Capabilities**:
- Incident CRUD
- Incident updates
- Incident templates
- Workflow automation

**New Features to Implement** (9 features):
- Auto-incident creation from monitoring failures
- Incident impact analysis (affected users, revenue impact)
- Post-mortem templates
- Incident retrospectives
- Incident timeline visualization
- Incident priority/severity matrix
- Incident escalation automation
- Incident owner assignment
- Incident categorization/tagging

**RabbitMQ Events Published**:
```
beakon.incidents.created
beakon.incidents.updated
beakon.incidents.resolved
beakon.incidents.escalated
```

**RabbitMQ Events Consumed**:
```
beakon.monitoring.incident.auto_created → Create incident from monitoring failure
beakon.components.sla.breached → Create high-priority incident
```

**Redis Usage**:
- **Cache active incidents**: `incidents:active:{tenant_id}` (no expiry, invalidate on update)
- **Store incident timeline**: `incident:timeline:{incident_id}` (append-only for fast retrieval)

**Database Schema Changes**:

```sql
ALTER TABLE incidents
ADD COLUMN priority VARCHAR(10) DEFAULT 'P3', -- P0, P1, P2, P3, P4
ADD COLUMN estimated_impact_users INT,
ADD COLUMN estimated_impact_revenue DECIMAL(12,2),
ADD COLUMN post_mortem TEXT,
ADD COLUMN lessons_learned JSONB,
ADD COLUMN owner_id UUID,
ADD COLUMN tags JSONB,
ADD COLUMN auto_created BOOLEAN DEFAULT false,
ADD COLUMN source VARCHAR(50); -- 'manual', 'monitoring', 'integration'

CREATE INDEX idx_incidents_priority ON incidents(priority, status) WHERE status != 'resolved';
```

---

### status-ui-service (Port 8093) - **Public-Facing UI**

**Current Capabilities**:
- Status page display
- Component status
- Incident display
- Subscriptions
- Branding

**New Features to Implement** (8 features):
- Embeddable widgets (iframe, JavaScript snippet)
- Status badges (SVG, PNG)
- Private status pages (password-protected)
- Historical incident calendar
- Uptime showcase (90-day charts)
- RSS/Atom feeds
- JSON API for status
- Multi-language support

**RabbitMQ Events Published**:
```
beakon.subscribers.subscribed
beakon.subscribers.unsubscribed
```

**RabbitMQ Events Consumed**:
```
beakon.components.status.changed → Update real-time status display
beakon.incidents.created → Display new incident banner
beakon.incidents.updated → Push update to subscribers via WebSocket
```

**Redis Usage**:
- **Cache rendered status page**: `statuspage:{tenant_id}:html` (1-minute TTL)
- **WebSocket pub/sub**: Channel `status:updates:{tenant_id}` for real-time updates
- **Rate limiting**: `ratelimit:api:{ip}:{endpoint}` (60 requests/minute)
- **Session storage**: `session:private:{session_id}` for private status pages

**Database Schema Changes**:

```sql
CREATE TABLE status_page_settings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL UNIQUE,
    widget_enabled BOOLEAN DEFAULT true,
    widget_style VARCHAR(20) DEFAULT 'badge', -- badge, banner, inline
    widget_position VARCHAR(20) DEFAULT 'bottom-right',
    badge_type VARCHAR(20) DEFAULT 'svg', -- svg, png
    private_page_enabled BOOLEAN DEFAULT false,
    private_page_password_hash VARCHAR(255),
    uptime_showcase_enabled BOOLEAN DEFAULT true,
    uptime_showcase_days INT DEFAULT 90,
    languages JSONB DEFAULT '["en"]', -- ['en', 'es', 'fr']
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

### notification-service (Port 8085) - **Multi-Channel Notifications**

**Current Capabilities**:
- Email notifications
- Webhook notifications

**New Features to Implement** (10 features):
- SMS notifications (Twilio integration)
- Slack notifications (Slack API)
- Microsoft Teams notifications
- Discord notifications
- Phone call alerts (Twilio voice)
- Notification preferences (per-subscriber)
- Notification throttling/deduplication
- Digest notifications (daily/weekly summaries)
- Component-specific subscriptions
- Incident-specific subscriptions

**RabbitMQ Events Consumed**:
```
beakon.incidents.created → Send notifications to subscribers
beakon.incidents.updated → Send update notifications
beakon.maintenance.scheduled → Send maintenance reminders
beakon.monitoring.ssl.expiring → Send expiration warnings (30, 14, 7 days)
```

**Redis Usage**:
- **Notification queue**: `notifications:queue:{priority}` (high, medium, low)
- **Notification throttling**: `notifications:throttle:{subscriber_id}` (prevent spam)
- **Sent notification cache**: `notifications:sent:{hash}` (deduplication, 1-hour TTL)
- **Digest queue**: `notifications:digest:{subscriber_id}:{frequency}` (batch notifications)

**Database Schema Changes**:

```sql
CREATE TABLE notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    subscriber_id UUID NOT NULL,
    email_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    slack_enabled BOOLEAN DEFAULT false,
    teams_enabled BOOLEAN DEFAULT false,
    discord_enabled BOOLEAN DEFAULT false,
    phone_call_enabled BOOLEAN DEFAULT false,
    component_ids JSONB, -- [123, 456] or null for all
    incident_types JSONB DEFAULT '["critical", "major"]',
    digest_frequency VARCHAR(20) DEFAULT 'none', -- none, daily, weekly
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(subscriber_id)
);

CREATE TABLE notification_digest_queue (
    id BIGSERIAL PRIMARY KEY,
    subscriber_id UUID NOT NULL,
    frequency VARCHAR(20) NOT NULL,
    notifications JSONB NOT NULL, -- [{type: 'incident', id: 123, timestamp: '...'}]
    scheduled_for TIMESTAMP NOT NULL,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_digest_queue_scheduled ON notification_digest_queue(scheduled_for) WHERE sent_at IS NULL;
```

---

### event-store-service (Port 8096) - **Event Sourcing**

**Current Capabilities**:
- Event storage
- Event replay
- Event querying

**New Features to Implement** (3 features):
- Event analytics (incident frequency, MTTR, MTTD)
- Event aggregation for reporting
- Event stream for real-time dashboards

**RabbitMQ Events Consumed** (store ALL events):
```
beakon.monitoring.*
beakon.components.*
beakon.incidents.*
beakon.maintenance.*
beakon.notifications.*
beakon.subscribers.*
```

**Redis Usage**:
- **Cache aggregated metrics**: `events:metrics:{tenant_id}:{period}` (hourly rollups, 1-hour TTL)
- **Real-time event stream**: Redis Streams `events:stream:{tenant_id}`

**Database Schema Changes**:

```sql
-- Add partitioning for scalability
CREATE TABLE events (
    id BIGSERIAL,
    tenant_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create monthly partitions
CREATE TABLE events_2025_10 PARTITION OF events
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE events_2025_11 PARTITION OF events
FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');

-- Indexes for common queries
CREATE INDEX idx_events_tenant_time ON events(tenant_id, timestamp DESC);
CREATE INDEX idx_events_type ON events(event_type, timestamp DESC);
```

---

### analytics-service (Port 8090) - **Reporting & Insights**

**Current Capabilities**:
- Basic metrics collection
- Performance tracking

**New Features to Implement** (6 features):
- SLA reporting (monthly uptime reports)
- MTTR/MTTD tracking
- Incident frequency analysis
- Custom dashboard builder
- Exportable reports (PDF, CSV, JSON)
- Trend analysis

**RabbitMQ Events Consumed**:
```
beakon.components.* → Track component metrics
beakon.incidents.* → Calculate MTTR, MTTD, frequency
```

**Redis Usage**:
- **Cache dashboard queries**: `analytics:dashboard:{tenant_id}:{query_hash}` (5-minute TTL)
- **Pre-aggregated metrics**: `analytics:metrics:{tenant_id}:{period}:{type}` (hourly, daily, monthly)

**Database Schema Changes**:

```sql
CREATE TABLE sla_reports (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    component_id BIGINT REFERENCES components(id),
    month DATE NOT NULL, -- First day of month
    uptime_percentage DECIMAL(5,2),
    total_downtime_seconds INT,
    incident_count INT,
    sla_target DECIMAL(5,2),
    sla_met BOOLEAN,
    report_data JSONB, -- Detailed breakdown
    generated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, component_id, month)
);

CREATE TABLE mttr_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id BIGINT REFERENCES incidents(id),
    time_to_detection_seconds INT, -- Time from failure to incident creation
    time_to_acknowledgment_seconds INT, -- Time from creation to acknowledged
    time_to_resolution_seconds INT, -- Time from creation to resolved
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE custom_dashboards (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    widgets JSONB NOT NULL, -- [{type: 'chart', config: {...}}]
    layout JSONB, -- Grid layout configuration
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## RabbitMQ Event Architecture

### Exchange Strategy

**1. Topic Exchange: `beakon.monitoring`**

Purpose: Monitoring check results and alerts

**Routing Keys**:
- `monitoring.check.passed` - Health check succeeded
- `monitoring.check.failed` - Health check failed
- `monitoring.check.degraded` - Health check degraded (slow response)
- `monitoring.ssl.expiring` - SSL certificate expiring soon
- `monitoring.domain.expiring` - Domain expiring soon
- `monitoring.incident.auto_created` - Auto-created incident from monitoring failure

**Producers**:
- monitoring-service

**Consumers**:
- component-service (update component status)
- incident-service (auto-create incidents)
- notification-service (send alerts)
- event-store-service (store all events)

**Example Message**:
```json
{
  "event_type": "monitoring.check.failed",
  "tenant_id": "72ffbda4-fd0d-4448-854a-935e36d32b90",
  "monitor_id": 123,
  "component_id": 456,
  "location": "us-east-1",
  "status": "down",
  "response_time_ms": null,
  "error": "Connection timeout after 5000ms",
  "timestamp": "2025-10-21T10:30:00Z"
}
```

---

**2. Topic Exchange: `beakon.components`**

Purpose: Component status changes and SLA events

**Routing Keys**:
- `components.status.changed` - Component status updated
- `components.sla.breached` - Component SLA threshold breached
- `components.created` - New component created
- `components.updated` - Component configuration updated
- `components.deleted` - Component deleted

**Producers**:
- component-service

**Consumers**:
- status-ui-service (real-time status page updates)
- incident-service (link components to incidents)
- analytics-service (track component metrics)
- event-store-service (store all events)

**Example Message**:
```json
{
  "event_type": "components.status.changed",
  "tenant_id": "72ffbda4-fd0d-4448-854a-935e36d32b90",
  "component_id": 456,
  "old_status": "operational",
  "new_status": "degraded",
  "reason": "monitoring_check_failed",
  "timestamp": "2025-10-21T10:30:05Z"
}
```

---

**3. Topic Exchange: `beakon.incidents`**

Purpose: Incident lifecycle events

**Routing Keys**:
- `incidents.created` - New incident created
- `incidents.updated` - Incident status or details updated
- `incidents.resolved` - Incident resolved
- `incidents.escalated` - Incident escalated to higher priority

**Producers**:
- incident-service

**Consumers**:
- notification-service (send notifications to subscribers)
- status-ui-service (display incidents on status page)
- analytics-service (calculate MTTR/MTTD)
- event-store-service (store all events)

**Example Message**:
```json
{
  "event_type": "incidents.created",
  "tenant_id": "72ffbda4-fd0d-4448-854a-935e36d32b90",
  "incident_id": 789,
  "title": "Database Connection Pool Exhausted",
  "status": "investigating",
  "severity": "critical",
  "affected_components": [456, 457],
  "auto_created": true,
  "source": "monitoring",
  "timestamp": "2025-10-21T10:30:10Z"
}
```

---

**4. Topic Exchange: `beakon.maintenance`**

Purpose: Scheduled maintenance lifecycle

**Routing Keys**:
- `maintenance.scheduled` - Maintenance window scheduled
- `maintenance.started` - Maintenance window started
- `maintenance.completed` - Maintenance window completed
- `maintenance.cancelled` - Maintenance window cancelled

**Producers**:
- monitoring-service (maintenance window management)

**Consumers**:
- component-service (set components to maintenance mode)
- notification-service (send maintenance reminders)
- monitoring-service (suppress alerts during maintenance)
- event-store-service (store all events)

**Example Message**:
```json
{
  "event_type": "maintenance.started",
  "tenant_id": "72ffbda4-fd0d-4448-854a-935e36d32b90",
  "maintenance_id": 321,
  "title": "Database Upgrade to PostgreSQL 16",
  "affected_components": [456, 457, 458],
  "start_time": "2025-10-21T10:00:00Z",
  "end_time": "2025-10-21T12:00:00Z",
  "timestamp": "2025-10-21T10:00:00Z"
}
```

---

**5. Topic Exchange: `beakon.notifications`**

Purpose: Notification delivery tracking

**Routing Keys**:
- `notifications.sent` - Notification successfully sent
- `notifications.failed` - Notification delivery failed
- `notifications.bounced` - Email bounced or invalid recipient

**Producers**:
- notification-service

**Consumers**:
- analytics-service (track notification metrics)
- event-store-service (store all events)

**Example Message**:
```json
{
  "event_type": "notifications.sent",
  "tenant_id": "72ffbda4-fd0d-4448-854a-935e36d32b90",
  "notification_id": 654,
  "subscriber_id": "98765",
  "channel": "email",
  "incident_id": 789,
  "subject": "Incident Update: Database Connection Pool Exhausted",
  "timestamp": "2025-10-21T10:30:15Z"
}
```

---

### Queue Bindings

**monitoring-service**:
```
Publishes to: beakon.monitoring, beakon.incidents (auto-incidents)
Consumes from:
  - beakon.maintenance (#) → Suppress alerts during maintenance
  Bindings:
    - maintenance.started → Queue: monitoring.maintenance.alerts
    - maintenance.completed → Queue: monitoring.maintenance.alerts
```

**component-service**:
```
Publishes to: beakon.components
Consumes from:
  - beakon.monitoring (#) → Update component status
  - beakon.incidents (#) → Link components to incidents
  Bindings:
    - monitoring.check.* → Queue: components.monitoring.updates
    - incidents.created → Queue: components.incident.links
```

**incident-service**:
```
Publishes to: beakon.incidents
Consumes from:
  - beakon.monitoring (#) → Auto-create incidents
  - beakon.components (#) → Handle SLA breaches
  Bindings:
    - monitoring.incident.auto_created → Queue: incidents.auto.created
    - components.sla.breached → Queue: incidents.sla.breaches
```

**notification-service**:
```
Publishes to: beakon.notifications
Consumes from:
  - beakon.incidents (#) → Send incident notifications
  - beakon.maintenance (#) → Send maintenance reminders
  - beakon.monitoring (#) → Send SSL/domain expiry warnings
  Bindings:
    - incidents.* → Queue: notifications.incidents
    - maintenance.* → Queue: notifications.maintenance
    - monitoring.ssl.expiring → Queue: notifications.ssl
```

**status-ui-service**:
```
Publishes to: None (read-only consumer)
Consumes from:
  - beakon.components (#) → Real-time status updates
  - beakon.incidents (#) → Display incident banner
  Bindings:
    - components.status.changed → Queue: statusui.components
    - incidents.* → Queue: statusui.incidents
```

**event-store-service**:
```
Publishes to: None (storage only)
Consumes from: ALL exchanges (fanout for complete event history)
  Bindings:
    - beakon.monitoring (#) → Queue: eventstore.monitoring
    - beakon.components (#) → Queue: eventstore.components
    - beakon.incidents (#) → Queue: eventstore.incidents
    - beakon.maintenance (#) → Queue: eventstore.maintenance
    - beakon.notifications (#) → Queue: eventstore.notifications
```

**analytics-service**:
```
Publishes to: None (read-only consumer)
Consumes from:
  - beakon.components (#) → Track component metrics
  - beakon.incidents (#) → Calculate MTTR/MTTD
  Bindings:
    - components.* → Queue: analytics.components
    - incidents.* → Queue: analytics.incidents
```

---

### Dead Letter Exchange (DLX) Pattern

**Purpose**: Handle failed message processing

**Configuration**:
```go
// Example: Create queue with DLX
args := amqp.Table{
    "x-dead-letter-exchange": "beakon.dlx",
    "x-dead-letter-routing-key": "failed.monitoring.check",
    "x-message-ttl": 300000, // 5 minutes
}
ch.QueueDeclare("monitoring.check.failed", true, false, false, false, args)
```

**DLX Queue**: `beakon.dlx.queue`
- Store failed messages for manual inspection
- Alert operations team on DLX message arrival
- Retry mechanism with exponential backoff

---

## Redis Usage Patterns

### Pattern 1: Caching (Read-Heavy Optimization)

**Use Case**: Reduce database load for frequently accessed data

**Implementation Examples**:

**1. Monitoring Results Cache**
```go
// Key: monitor:result:{monitor_id}:latest
// TTL: 1 minute
// Type: Hash

// Write
redis.HSet(ctx, "monitor:result:123:latest", map[string]interface{}{
    "status": "operational",
    "response_time_ms": 150,
    "location": "us-east-1",
    "timestamp": time.Now().Unix(),
})
redis.Expire(ctx, "monitor:result:123:latest", 1*time.Minute)

// Read
result := redis.HGetAll(ctx, "monitor:result:123:latest").Val()
```

**2. Component Status Cache**
```go
// Key: component:status:{component_id}
// TTL: 30 seconds
// Type: Hash

redis.HSet(ctx, "component:status:456", map[string]interface{}{
    "status": "degraded",
    "last_checked": time.Now().Unix(),
    "uptime_percentage": 99.87,
})
redis.Expire(ctx, "component:status:456", 30*time.Second)
```

**3. Public Status Page Cache**
```go
// Key: statuspage:{tenant_id}:html
// TTL: 1 minute
// Type: String

renderedHTML := renderStatusPage(tenantID)
redis.Set(ctx, fmt.Sprintf("statuspage:%s:html", tenantID), renderedHTML, 1*time.Minute)
```

**4. Active Incidents Cache**
```go
// Key: incidents:active:{tenant_id}
// TTL: No expiry (invalidate on update)
// Type: List

redis.RPush(ctx, "incidents:active:tenant123", incidentID)

// Invalidate on update
redis.Del(ctx, "incidents:active:tenant123")
```

---

### Pattern 2: Pub/Sub (Real-Time Updates)

**Use Case**: Push real-time updates to WebSocket clients

**Implementation Examples**:

**1. Real-Time Status Updates**
```go
// Publisher (component-service)
channel := fmt.Sprintf("status:updates:%s", tenantID)
message := map[string]interface{}{
    "component_id": 456,
    "status": "degraded",
    "timestamp": time.Now().Unix(),
}
redis.Publish(ctx, channel, json.Marshal(message))

// Subscriber (status-ui-service WebSocket handler)
pubsub := redis.Subscribe(ctx, fmt.Sprintf("status:updates:%s", tenantID))
ch := pubsub.Channel()
for msg := range ch {
    // Push to WebSocket clients
    broadcastToClients(msg.Payload)
}
```

**2. Real-Time Monitoring Dashboard**
```go
// Publisher (monitoring-service)
channel := fmt.Sprintf("monitoring:live:%s", tenantID)
message := map[string]interface{}{
    "monitor_id": 123,
    "status": "down",
    "location": "us-west-1",
    "timestamp": time.Now().Unix(),
}
redis.Publish(ctx, channel, json.Marshal(message))
```

---

### Pattern 3: Rate Limiting (Fair Usage)

**Use Case**: Prevent abuse and ensure fair resource allocation

**Implementation Examples**:

**1. Public API Rate Limiting**
```go
// Key: ratelimit:api:{ip_address}:{endpoint}
// TTL: 1 minute
// Type: Counter
// Limit: 60 requests/minute

key := fmt.Sprintf("ratelimit:api:%s:%s", ipAddress, endpoint)
count := redis.Incr(ctx, key).Val()

if count == 1 {
    redis.Expire(ctx, key, 1*time.Minute)
}

if count > 60 {
    return errors.New("Rate limit exceeded: 60 requests per minute")
}
```

**2. Monitoring Check Rate Limiting**
```go
// Key: ratelimit:monitor:{location_id}
// TTL: 1 second
// Type: Counter
// Limit: 100 checks/second per location

key := fmt.Sprintf("ratelimit:monitor:%s", locationID)
count := redis.Incr(ctx, key).Val()

if count == 1 {
    redis.Expire(ctx, key, 1*time.Second)
}

if count > 100 {
    return errors.New("Location rate limit exceeded")
}
```

---

### Pattern 4: Distributed Locks (Concurrency Control)

**Use Case**: Prevent duplicate processing

**Implementation Example**:

```go
// Prevent duplicate incident creation
lockKey := fmt.Sprintf("lock:incident:create:%s:%d", tenantID, componentID)
lockValue := uuid.New().String()

// Try to acquire lock
acquired := redis.SetNX(ctx, lockKey, lockValue, 10*time.Second).Val()
if !acquired {
    return errors.New("Incident creation already in progress")
}

defer func() {
    // Release lock only if we still own it
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
    redis.Eval(ctx, script, []string{lockKey}, lockValue)
}()

// Create incident
createIncident()
```

---

### Pattern 5: Leaderboards (Sorted Sets)

**Use Case**: Track component uptime rankings

**Implementation Example**:

```go
// Key: leaderboard:uptime:{tenant_id}
// Type: Sorted Set
// Score: Uptime percentage

redis.ZAdd(ctx, "leaderboard:uptime:tenant123", &redis.Z{
    Score: 99.95,
    Member: "component:456",
})

// Get top 10 components by uptime
top10 := redis.ZRevRange(ctx, "leaderboard:uptime:tenant123", 0, 9).Val()
```

---

### Pattern 6: Time Series (Redis Streams)

**Use Case**: Store real-time monitoring data for dashboards

**Implementation Example**:

```go
// Add monitoring data point
redis.XAdd(ctx, &redis.XAddArgs{
    Stream: fmt.Sprintf("monitoring:timeseries:%d", monitorID),
    MaxLen: 1000, // Keep last 1000 data points
    Approx: true,
    Values: map[string]interface{}{
        "response_time": 150,
        "status": "operational",
        "location": "us-east-1",
    },
})

// Read last 100 data points
entries := redis.XRevRange(ctx, fmt.Sprintf("monitoring:timeseries:%d", monitorID), "+", "-").Val()
```

---

### Redis Configuration Best Practices

**Development Environment**:
```conf
maxmemory 512mb
maxmemory-policy allkeys-lru
timeout 300
tcp-keepalive 60
```

**Production Environment**:
```conf
maxmemory 4gb
maxmemory-policy volatile-lru
timeout 300
tcp-keepalive 60
appendonly yes
appendfsync everysec
save 900 1
save 300 10
save 60 10000
```

**Connection Pooling (Go)**:
```go
redisClient := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     100,
    MinIdleConns: 10,
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})
```

---

## Database Best Practices

### 1. Time-Series Data Strategy

**Challenge**: Monitoring generates millions of data points daily

**Solution**: Hybrid time-series approach

**Hot Data (Last 7 Days)**: PostgreSQL with daily partitioning
```sql
CREATE TABLE monitoring_results (
    id BIGSERIAL,
    monitor_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    checked_at TIMESTAMP NOT NULL,
    status VARCHAR(20),
    response_time_ms INT,
    ttfb_ms INT,
    dns_time_ms INT,
    connection_time_ms INT,
    error_message TEXT,
    PRIMARY KEY (id, checked_at)
) PARTITION BY RANGE (checked_at);

-- Create daily partitions
CREATE TABLE monitoring_results_2025_10_21
PARTITION OF monitoring_results
FOR VALUES FROM ('2025-10-21 00:00:00') TO ('2025-10-22 00:00:00');

CREATE TABLE monitoring_results_2025_10_22
PARTITION OF monitoring_results
FOR VALUES FROM ('2025-10-22 00:00:00') TO ('2025-10-23 00:00:00');

-- Auto-create partitions using pg_partman extension
CREATE EXTENSION pg_partman;
SELECT create_parent('public.monitoring_results', 'checked_at', 'native', 'daily');
UPDATE part_config SET retention = '7 days', retention_keep_table = false WHERE parent_table = 'public.monitoring_results';
```

**Warm Data (8-90 Days)**: Aggregated hourly
```sql
CREATE TABLE monitoring_results_hourly (
    monitor_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    hour TIMESTAMP NOT NULL,
    avg_response_time_ms INT,
    min_response_time_ms INT,
    max_response_time_ms INT,
    p95_response_time_ms INT,
    p99_response_time_ms INT,
    uptime_percentage DECIMAL(5,2),
    check_count INT,
    failure_count INT,
    PRIMARY KEY (monitor_id, location_id, hour)
);

CREATE INDEX idx_monitoring_hourly_time ON monitoring_results_hourly(hour DESC);

-- Aggregation job (run hourly via cron)
INSERT INTO monitoring_results_hourly
SELECT
    monitor_id,
    location_id,
    date_trunc('hour', checked_at) AS hour,
    AVG(response_time_ms)::INT AS avg_response_time_ms,
    MIN(response_time_ms) AS min_response_time_ms,
    MAX(response_time_ms) AS max_response_time_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time_ms)::INT AS p95_response_time_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY response_time_ms)::INT AS p99_response_time_ms,
    (SUM(CASE WHEN status = 'operational' THEN 1 ELSE 0 END)::DECIMAL / COUNT(*) * 100) AS uptime_percentage,
    COUNT(*) AS check_count,
    SUM(CASE WHEN status != 'operational' THEN 1 ELSE 0 END) AS failure_count
FROM monitoring_results
WHERE checked_at >= NOW() - INTERVAL '2 hours'
  AND checked_at < NOW() - INTERVAL '1 hour'
GROUP BY monitor_id, location_id, date_trunc('hour', checked_at)
ON CONFLICT (monitor_id, location_id, hour) DO NOTHING;
```

**Cold Data (90+ Days)**: Daily/monthly aggregates
```sql
CREATE TABLE monitoring_results_daily (
    monitor_id BIGINT NOT NULL,
    day DATE NOT NULL,
    avg_response_time_ms INT,
    p95_response_time_ms INT,
    uptime_percentage DECIMAL(5,2),
    total_checks INT,
    total_failures INT,
    PRIMARY KEY (monitor_id, day)
);

CREATE TABLE monitoring_results_monthly (
    monitor_id BIGINT NOT NULL,
    month DATE NOT NULL, -- First day of month
    avg_response_time_ms INT,
    p95_response_time_ms INT,
    uptime_percentage DECIMAL(5,2),
    total_checks INT,
    total_failures INT,
    sla_met BOOLEAN,
    PRIMARY KEY (monitor_id, month)
);
```

---

### 2. Indexing Strategy

**Monitoring Tables**:
```sql
-- Compound index for common query pattern (monitor + time)
CREATE INDEX idx_monitoring_results_monitor_time
ON monitoring_results(monitor_id, checked_at DESC);

-- Index for filtering failures
CREATE INDEX idx_monitoring_results_status
ON monitoring_results(status, checked_at DESC)
WHERE status != 'operational';

-- Index for multi-location queries
CREATE INDEX idx_monitoring_results_location
ON monitoring_results(location_id, monitor_id, checked_at DESC);

-- Partial index for recent data (last 24 hours)
CREATE INDEX idx_monitoring_results_recent
ON monitoring_results(monitor_id, checked_at DESC)
WHERE checked_at > NOW() - INTERVAL '24 hours';
```

**Incident Tables**:
```sql
-- Compound index for tenant + creation time
CREATE INDEX idx_incidents_tenant_created
ON incidents(tenant_id, created_at DESC);

-- Partial index for active incidents
CREATE INDEX idx_incidents_active
ON incidents(tenant_id, status, priority)
WHERE status IN ('investigating', 'identified', 'monitoring');

-- Full-text search index for incident search
CREATE INDEX idx_incidents_search
ON incidents USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));
```

**Component Tables**:
```sql
-- Index for public status page queries
CREATE INDEX idx_components_public
ON components(tenant_id, is_public, position)
WHERE is_public = true AND deleted_at IS NULL;

-- GIN index for tags search
CREATE INDEX idx_components_tags
ON components USING gin(tags);

-- Index for dependency lookups
CREATE INDEX idx_components_dependencies
ON components USING gin(dependencies);
```

---

### 3. Connection Pooling (from shared-resilience)

**Configuration**:
- **Development**: 80 max connections (CPU × 10)
- **Production**: 200 max connections (CPU × 25)
- **Max idle**: 40% of max connections
- **Connection lifetime**: 1 hour
- **Idle timeout**: 10 minutes

**Implementation**:
```go
// shared-resilience library
db, err := resilience.NewDatabaseConnection(resilience.DatabaseConfig{
    Host:            os.Getenv("DB_HOST"),
    Port:            5432,
    Database:        os.Getenv("DB_NAME"),
    User:            os.Getenv("DB_USER"),
    Password:        os.Getenv("DB_PASSWORD"),
    SSLMode:         os.Getenv("DB_SSLMODE"),
    MaxOpenConns:    200,
    MaxIdleConns:    80,
    ConnMaxLifetime: 1 * time.Hour,
    ConnMaxIdleTime: 10 * time.Minute,
})
```

---

### 4. Data Retention Policy

**Monitoring Results**:
- Raw data: 7 days (daily partitions, auto-drop)
- Hourly aggregates: 90 days
- Daily aggregates: 2 years
- Monthly aggregates: Forever

**Incidents**:
- All incidents: Forever (compliance requirement)
- Incident updates: Forever
- Incident attachments: 2 years (archive to S3 after)

**Event Store**:
- All events: Forever (event sourcing pattern)
- Partitioned by month for query performance

**Logs**:
- Application logs: 30 days
- Audit logs: 7 years (compliance - SOC 2, GDPR)
- Error logs: 90 days

**Implementation (pg_partman)**:
```sql
-- Configure retention
UPDATE part_config
SET retention = '7 days',
    retention_keep_table = false
WHERE parent_table = 'public.monitoring_results';

-- Run maintenance daily (cron)
SELECT run_maintenance('public.monitoring_results');
```

---

### 5. Query Optimization

**Use EXPLAIN ANALYZE**:
```sql
EXPLAIN ANALYZE
SELECT monitor_id, AVG(response_time_ms)
FROM monitoring_results
WHERE checked_at > NOW() - INTERVAL '24 hours'
GROUP BY monitor_id;
```

**Covering Indexes** (include frequently selected columns):
```sql
CREATE INDEX idx_monitoring_results_covering
ON monitoring_results(monitor_id, checked_at DESC)
INCLUDE (status, response_time_ms);
```

**Materialized Views** (for expensive aggregations):
```sql
CREATE MATERIALIZED VIEW mv_component_uptime_24h AS
SELECT
    c.id AS component_id,
    c.name,
    (SUM(CASE WHEN mr.status = 'operational' THEN 1 ELSE 0 END)::DECIMAL / COUNT(*) * 100) AS uptime_24h
FROM components c
LEFT JOIN monitors m ON m.component_id = c.id
LEFT JOIN monitoring_results mr ON mr.monitor_id = m.id
WHERE mr.checked_at > NOW() - INTERVAL '24 hours'
GROUP BY c.id, c.name;

CREATE UNIQUE INDEX ON mv_component_uptime_24h(component_id);

-- Refresh hourly
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_component_uptime_24h;
```

---

### 6. Database Maintenance

**Vacuum Strategy**:
```sql
-- Auto-vacuum settings (postgresql.conf)
autovacuum = on
autovacuum_max_workers = 4
autovacuum_naptime = 30s
autovacuum_vacuum_threshold = 50
autovacuum_analyze_threshold = 50
autovacuum_vacuum_scale_factor = 0.1
autovacuum_analyze_scale_factor = 0.05

-- Manual vacuum for heavily updated tables
VACUUM ANALYZE monitoring_results;
```

**ANALYZE Statistics**:
```sql
-- Update statistics after bulk inserts
ANALYZE monitoring_results;
ANALYZE incidents;
```

---

## Implementation Roadmap

### Phase 1: Critical Missing Features (Weeks 1-3)

**Goal**: Implement highest-priority features with immediate customer value

#### Week 1: Multi-Location Monitoring + SSL Monitoring

**Tasks**:
1. **Multi-Location Monitoring** (monitoring-service)
   - Add `monitoring_locations` table (10 global locations)
   - Update `monitoring_results` schema (add `location_id` column)
   - Implement location-based health check distribution
   - Add location selection UI in monitoring configuration
   - Test: Run health checks from 3+ locations simultaneously

2. **SSL Certificate Monitoring** (monitoring-service)
   - Add `ssl_certificates` table
   - Implement SSL certificate discovery (scan tenant domains)
   - Add expiration tracking (30, 14, 7 days warnings)
   - Publish `monitoring.ssl.expiring` RabbitMQ events
   - Add SSL certificate dashboard page
   - Test: Detect expiring certificate and send notification

**Deliverables**:
- Multi-location monitoring feature (100% functional)
- SSL certificate monitoring (100% functional)
- RabbitMQ events for SSL expiry
- Admin dashboard showing SSL certificate status

**Success Metrics**:
- ✅ Health checks run from ≥3 locations per monitor
- ✅ SSL expiration warnings sent 30, 14, 7 days before expiry
- ✅ SSL dashboard shows all tenant certificates

---

#### Week 2: Embeddable Widgets + Auto-Incident Creation

**Tasks**:
1. **Embeddable Widgets** (status-ui-service)
   - Add `status_page_settings` table (widget config)
   - Implement status badge endpoint (SVG, PNG)
   - Implement embeddable widget (iframe, JavaScript snippet)
   - Add widget customization API (style, position, colors)
   - Add widget documentation page
   - Test: Embed widget on external site

2. **Auto-Incident Creation** (monitoring-service + incident-service)
   - Publish `monitoring.incident.auto_created` RabbitMQ event on repeated failures
   - Consume event in incident-service
   - Create incident with details (affected components, error message)
   - Auto-resolve incident when checks pass
   - Add configuration UI (enable/disable auto-incidents, failure threshold)
   - Test: Fail health check 3 times → incident auto-created

**Deliverables**:
- Status badge (SVG/PNG) endpoint
- Embeddable widget (iframe + JS snippet)
- Auto-incident creation (functional)
- Auto-incident resolution (functional)

**Success Metrics**:
- ✅ Widget embeds successfully on 3rd-party sites
- ✅ Auto-incident created after 3 consecutive failures
- ✅ Auto-incident resolved when check passes

---

#### Week 3: SMS Notifications + Alert Suppression

**Tasks**:
1. **SMS Notifications** (notification-service)
   - Integrate Twilio SDK
   - Add `notification_preferences` table (SMS opt-in)
   - Implement SMS sending logic (incident created, incident updated)
   - Add SMS subscription preferences UI
   - Add SMS delivery tracking
   - Test: Send SMS notification on incident creation

2. **Alert Suppression During Maintenance** (monitoring-service)
   - Consume `maintenance.started` RabbitMQ event
   - Suppress alerts for affected components during maintenance window
   - Resume alerts on `maintenance.completed` event
   - Add suppression indicator in monitoring dashboard
   - Test: Start maintenance → verify no alerts sent for affected components

**Deliverables**:
- SMS notifications (functional via Twilio)
- Notification preferences UI (email, SMS, Slack toggles)
- Alert suppression during maintenance

**Success Metrics**:
- ✅ SMS notifications sent within 30 seconds of incident
- ✅ No alerts sent for components in maintenance mode
- ✅ Alerts resume after maintenance completion

---

### Phase 2: High-Value Features (Weeks 4-7)

**Goal**: Add competitive differentiators and advanced features

#### Week 4-5: Performance Metrics + Third-Party Integrations

**Tasks**:
1. **Performance Metrics** (monitoring-service)
   - Update schema: Add P50, P95, P99 columns to `monitoring_results_hourly`
   - Implement percentile calculation in hourly aggregation job
   - Add TTFB, DNS time, connection time tracking
   - Add performance metrics charts (line chart, histogram)
   - Test: Verify P95 and P99 calculated correctly

2. **Third-Party Integrations** (monitoring-service)
   - **Datadog Integration**: Send monitoring data to Datadog Metrics API
   - **PagerDuty Integration**: Trigger PagerDuty incidents on critical alerts
   - **Slack Notifications**: Post to Slack channels on incident updates
   - Add integration configuration UI (API keys, webhook URLs)
   - Test: Create incident → PagerDuty alert triggered + Slack message sent

**Deliverables**:
- Response time percentiles (P50, P95, P99) in dashboard
- TTFB, DNS time, connection time metrics
- Datadog integration (functional)
- PagerDuty integration (functional)
- Slack notifications (functional)

**Success Metrics**:
- ✅ P95 and P99 response times displayed in charts
- ✅ Datadog receives monitoring metrics every minute
- ✅ PagerDuty incidents created for critical alerts
- ✅ Slack notifications sent within 10 seconds

---

#### Week 6-7: SLA Reporting + Private Status Pages

**Tasks**:
1. **SLA Reporting** (analytics-service)
   - Add `sla_reports` table
   - Implement monthly uptime calculation job (cron)
   - Add SLA breach detection (uptime < target)
   - Generate PDF reports using wkhtmltopdf
   - Add SLA dashboard page (monthly reports, breach alerts)
   - Test: Generate monthly report for component with 99.5% uptime target

2. **Private Status Pages** (status-ui-service)
   - Add password protection for status pages
   - Implement session management for private pages
   - Add private page configuration UI (enable/disable, password)
   - Add team-only visibility toggle
   - Test: Access private status page → redirects to login

**Deliverables**:
- Monthly SLA reports (auto-generated)
- SLA breach alerts
- Exportable PDF reports
- Password-protected status pages
- Team-only status pages

**Success Metrics**:
- ✅ Monthly SLA reports generated on 1st of each month
- ✅ SLA breach alerts sent when uptime < target
- ✅ PDF reports downloadable from dashboard
- ✅ Private status page requires password

---

### Phase 3: Advanced Features (Weeks 8-13)

**Goal**: Differentiate with AI-powered and advanced monitoring

#### Week 8-10: Synthetic Transaction Monitoring + On-Call Scheduling

**Tasks**:
1. **Synthetic Transaction Monitoring** (monitoring-service)
   - Add `synthetic_transactions` table
   - Implement multi-step health check executor (using Playwright)
   - Support steps: navigate, click, fill, assert
   - Add transaction editor UI (visual workflow builder)
   - Add transaction execution dashboard
   - Test: Create 3-step transaction (login → dashboard → logout)

2. **On-Call Scheduling** (monitoring-service)
   - Add `on_call_schedules` and `escalation_policies` tables
   - Implement on-call rotation logic (daily, weekly, custom)
   - Add escalation workflow (level 1 → level 2 → level 3)
   - Add on-call calendar UI (who's on call this week)
   - Add override feature (swap on-call shifts)
   - Test: Create weekly rotation → verify correct person alerted

**Deliverables**:
- Synthetic transaction monitoring (functional)
- Multi-step health check editor
- On-call scheduling system
- Escalation policies
- On-call calendar UI

**Success Metrics**:
- ✅ Synthetic transactions execute multi-step workflows
- ✅ Correct on-call person alerted based on schedule
- ✅ Escalation occurs if no response within timeout

---

#### Week 11-12: Service Dependency Mapping + Advanced Analytics

**Tasks**:
1. **Service Dependency Mapping** (component-service)
   - Update `components` table (add `dependencies` JSON column)
   - Implement dependency graph visualization (D3.js)
   - Add dependency impact analysis (what breaks if X fails)
   - Add dependency health scoring (composite health)
   - Test: Create dependency chain → verify impact visualization

2. **Advanced Analytics** (analytics-service)
   - Add `custom_dashboards` table
   - Implement custom dashboard builder (drag-drop widgets)
   - Add MTTR/MTTD tracking
   - Add incident frequency analysis
   - Add trend analysis (week-over-week, month-over-month)
   - Test: Create custom dashboard with 5 widgets

**Deliverables**:
- Service dependency graph (interactive visualization)
- Dependency impact analysis
- Custom dashboard builder
- MTTR/MTTD metrics
- Incident frequency trends

**Success Metrics**:
- ✅ Dependency graph shows service relationships
- ✅ Impact analysis identifies downstream affected services
- ✅ MTTR calculated for all resolved incidents
- ✅ Custom dashboards save and load correctly

---

#### Week 13: Anomaly Detection (AI-Powered)

**Tasks**:
1. **Anomaly Detection** (analytics-service)
   - Integrate anomaly detection library (e.g., Prophet, statsmodels)
   - Implement baseline calculation (7-day, 30-day averages)
   - Detect anomalies: response time spikes, unexpected downtime
   - Add anomaly alerts (send notification when detected)
   - Add anomaly dashboard (visualize detected anomalies)
   - Test: Simulate response time spike → anomaly detected

**Deliverables**:
- Anomaly detection engine (functional)
- Baseline calculation (auto-updated)
- Anomaly alerts
- Anomaly visualization dashboard

**Success Metrics**:
- ✅ Response time anomalies detected within 5 minutes
- ✅ Anomaly alerts sent to operations team
- ✅ False positive rate < 5%

---

## Success Metrics

### Customer Success Metrics

**1. Feature Adoption**
- **Target**: 80% of tenants use ≥5 new monitoring features within 3 months of launch
- **Measurement**: Track feature usage via analytics events
- **Priority**: High

**2. Customer Satisfaction (CSAT)**
- **Target**: CSAT score ≥ 4.5/5.0 for monitoring features
- **Measurement**: In-app surveys after feature usage
- **Priority**: High

**3. Churn Reduction**
- **Target**: Reduce churn by 15% due to improved monitoring capabilities
- **Measurement**: Compare churn rate before/after implementation
- **Priority**: Critical

**4. Upsell Opportunities**
- **Target**: 25% of free-tier users upgrade to paid plans for advanced monitoring
- **Measurement**: Track upgrade conversions from free → paid
- **Priority**: Medium

---

### Technical Performance Metrics

**5. API Response Time**
- **Target**: P95 response time < 500ms for all monitoring APIs
- **Measurement**: Prometheus metrics
- **Priority**: High

**6. Monitoring Check Frequency**
- **Target**: 1M+ health checks per day across all tenants
- **Measurement**: Database query (count checks per day)
- **Priority**: Medium

**7. Event Processing Latency**
- **Target**: RabbitMQ events processed within 100ms (P95)
- **Measurement**: RabbitMQ monitoring
- **Priority**: High

**8. Database Query Performance**
- **Target**: P95 query time < 50ms for time-series data
- **Measurement**: PostgreSQL slow query log
- **Priority**: Medium

**9. Cache Hit Rate**
- **Target**: Redis cache hit rate > 85%
- **Measurement**: Redis INFO stats
- **Priority**: Medium

**10. Uptime SLA**
- **Target**: 99.95% uptime for monitoring-service
- **Measurement**: Internal monitoring + third-party (StatusCake)
- **Priority**: Critical

---

### Business Metrics

**11. Feature Parity with Competitors**
- **Target**: Achieve 95%+ feature parity with Statuspage.io
- **Measurement**: Feature comparison matrix (this document)
- **Baseline**: 29% (current)
- **Priority**: Critical

**12. Time to Market**
- **Target**: Launch Phase 1 features within 3 weeks
- **Measurement**: Project timeline tracking
- **Priority**: High

**13. Cost Efficiency**
- **Target**: Infrastructure cost < $0.10 per tenant per month
- **Measurement**: AWS/cloud cost analysis
- **Priority**: Medium

---

## Appendix A: Feature Priority Definitions

**P0 - Critical** (Must-Have for Competitive Parity):
- Multi-location monitoring
- SSL certificate monitoring
- Embeddable widgets
- Status badges
- Auto-incident creation
- SMS notifications
- Slack notifications
- PagerDuty integration
- Alert suppression during maintenance
- On-call scheduling
- Alert escalation policies

**P1 - High** (Strong Competitive Advantage):
- Performance metrics (P95, P99)
- Private status pages
- SLA reporting
- MTTR tracking
- Component-specific subscriptions
- Synthetic transaction monitoring
- Third-party integrations (Datadog, Prometheus)

**P2 - Medium** (Nice-to-Have):
- Incident retrospectives
- Notification digests
- Custom dashboards
- Service dependency mapping
- Heartbeat monitoring

**P3 - Low** (Future Enhancements):
- Anomaly detection (AI)
- Real User Monitoring (RUM)
- Website defacement detection
- Distributed tracing integration

---

## Appendix B: Existing Documentation References

**Related Documents**:
- [docs/historical/MONITORING_ANALYTICS.md](../historical/MONITORING_ANALYTICS.md) - Existing analytics features
- [docs/historical/MONITORING_OBSERVABILITY.md](../historical/MONITORING_OBSERVABILITY.md) - Observability stack
- [microservices/monitoring-service/README.md](../../microservices/monitoring-service/README.md) - Service documentation

**Service Catalogs**:
- [SERVICE_CATALOG.md](../../SERVICE_CATALOG.md) - Complete service reference
- [DATABASE_ARCHITECTURE.md](../../DATABASE_ARCHITECTURE.md) - Database schemas

**Development Guides**:
- [CLAUDE.md](../../CLAUDE.md) - Developer onboarding
- [microservices/QUICK_START.md](../../microservices/QUICK_START.md) - Setup guide

---

## Appendix C: Technology Stack

**Backend**:
- Go 1.21+ (all microservices)
- Gin (HTTP framework)
- GORM (ORM)
- PostgreSQL 16 (primary database)
- Redis 7 (caching, pub/sub)
- RabbitMQ 4.1 (message broker)

**Frontend**:
- Next.js 14 (React 18, TypeScript 5)
- TanStack Query (data fetching)
- Chart.js (visualizations)
- Tailwind CSS + shadcn/ui (styling)

**Third-Party Services**:
- Twilio (SMS, voice calls)
- Datadog (monitoring integration)
- PagerDuty (incident management)
- Slack API (notifications)
- Microsoft Teams API (notifications)

**Tools**:
- Playwright (synthetic monitoring)
- wkhtmltopdf (PDF report generation)
- Prophet/statsmodels (anomaly detection)

---

**Document Version**: 1.0
**Last Updated**: October 21, 2025
**Next Review**: November 21, 2025 (after Phase 1 completion)

---

**Status**: ✅ Ready for Implementation
