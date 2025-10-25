# Phase 3, Week 13: Anomaly Detection System - Design Document

**Created**: January 2025
**Status**: 🎯 Design Phase
**Complexity**: High (ML/AI Integration)
**Estimated Implementation**: 5-7 days

---

## 📋 Executive Summary

Anomaly Detection is the final feature of Phase 3, providing AI-powered monitoring to detect unusual patterns in system behavior before they become critical incidents. This document outlines the architecture and implementation approach.

**Goal**: Automatically detect anomalies in metrics (response time, error rate, traffic) and alert teams proactively.

**Approach**: Statistical anomaly detection using Go-native algorithms (no external ML dependencies initially).

---

## 🎯 Feature Requirements

### Core Requirements

**1. Metric Collection**:
- Response time (ms)
- Error rate (%)
- Request volume (requests/minute)
- Component status changes
- Incident frequency

**2. Baseline Calculation**:
- Rolling 7-day baseline
- Rolling 30-day baseline
- Hour-of-day patterns (weekday vs weekend)
- Seasonal adjustments

**3. Anomaly Detection**:
- **Statistical Methods**:
  - Z-score (standard deviation)
  - Moving average
  - Exponential smoothing
  - Percentile-based (P95, P99)
- **Thresholds**:
  - Minor: 2 standard deviations
  - Major: 3 standard deviations
  - Critical: 4 standard deviations

**4. Alerting**:
- Anomaly detected → notification
- Integration with existing escalation policies
- Configurable sensitivity levels

**5. Visualization**:
- Anomaly timeline
- Baseline vs actual metrics
- Anomaly severity heatmap
- Historical anomaly trends

---

## 🏗️ System Architecture

### Architecture Decision: Go-Native vs External ML

**Option 1: Go-Native Statistical Algorithms** ⭐ (Recommended)
- **Pros**:
  - No external dependencies (Python, R)
  - Simpler deployment (single binary)
  - Lower latency (in-process calculations)
  - Easier to maintain
  - Good enough for most use cases
- **Cons**:
  - Less sophisticated than ML models
  - May have more false positives
  - Limited to statistical methods

**Option 2: Python ML Integration**
- **Pros**:
  - Advanced algorithms (Prophet, ARIMA, LSTM)
  - Better accuracy for complex patterns
  - Rich ecosystem (statsmodels, scikit-learn)
- **Cons**:
  - Requires Python runtime
  - Inter-process communication overhead
  - Deployment complexity
  - Maintenance burden (two languages)

**Decision**: Start with **Option 1 (Go-Native)**, with architecture allowing future upgrade to Option 2.

---

## 🗄️ Database Schema

### 1. Metric Snapshots Table

```sql
CREATE TABLE metric_snapshots (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    monitor_id BIGINT,  -- NULL for tenant-wide metrics
    metric_type VARCHAR(50) NOT NULL,  -- 'response_time', 'error_rate', 'request_volume'
    metric_value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    INDEX idx_metric_snapshots_tenant (tenant_id),
    INDEX idx_metric_snapshots_monitor (monitor_id),
    INDEX idx_metric_snapshots_type (metric_type),
    INDEX idx_metric_snapshots_timestamp (timestamp),
    INDEX idx_metric_snapshots_tenant_monitor_time (tenant_id, monitor_id, timestamp)
);
```

### 2. Anomaly Baselines Table

```sql
CREATE TABLE anomaly_baselines (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    monitor_id BIGINT,
    metric_type VARCHAR(50) NOT NULL,
    baseline_type VARCHAR(20) NOT NULL,  -- '7day', '30day', 'hourly'
    hour_of_day INT,  -- 0-23 for hourly baselines, NULL otherwise
    day_of_week INT,  -- 0-6 for weekly patterns, NULL otherwise
    mean_value DOUBLE PRECISION NOT NULL,
    std_dev DOUBLE PRECISION NOT NULL,
    p50 DOUBLE PRECISION,
    p95 DOUBLE PRECISION,
    p99 DOUBLE PRECISION,
    sample_count INT NOT NULL,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    INDEX idx_anomaly_baselines_tenant (tenant_id),
    INDEX idx_anomaly_baselines_monitor (monitor_id),
    INDEX idx_anomaly_baselines_type (metric_type),
    INDEX idx_anomaly_baselines_expires (expires_at),
    UNIQUE (tenant_id, monitor_id, metric_type, baseline_type, hour_of_day, day_of_week)
);
```

### 3. Detected Anomalies Table

```sql
CREATE TABLE detected_anomalies (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    monitor_id BIGINT,
    metric_type VARCHAR(50) NOT NULL,
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL,
    actual_value DOUBLE PRECISION NOT NULL,
    expected_value DOUBLE PRECISION NOT NULL,
    deviation_score DOUBLE PRECISION NOT NULL,  -- Z-score
    severity VARCHAR(20) NOT NULL,  -- 'minor', 'major', 'critical'
    status VARCHAR(20) NOT NULL DEFAULT 'open',  -- 'open', 'acknowledged', 'resolved', 'false_positive'
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    INDEX idx_detected_anomalies_tenant (tenant_id),
    INDEX idx_detected_anomalies_monitor (monitor_id),
    INDEX idx_detected_anomalies_detected_at (detected_at),
    INDEX idx_detected_anomalies_status (status),
    INDEX idx_detected_anomalies_severity (severity)
);
```

### 4. Anomaly Detection Config Table

```sql
CREATE TABLE anomaly_detection_config (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    monitor_id BIGINT,  -- NULL for tenant-wide defaults
    metric_type VARCHAR(50),  -- NULL for all metrics
    enabled BOOLEAN DEFAULT TRUE,
    sensitivity VARCHAR(20) DEFAULT 'medium',  -- 'low', 'medium', 'high'
    min_baseline_samples INT DEFAULT 50,
    z_score_threshold_minor DOUBLE PRECISION DEFAULT 2.0,
    z_score_threshold_major DOUBLE PRECISION DEFAULT 3.0,
    z_score_threshold_critical DOUBLE PRECISION DEFAULT 4.0,
    notification_enabled BOOLEAN DEFAULT TRUE,
    escalation_policy_id BIGINT,  -- Link to existing escalation policies
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (tenant_id, monitor_id, metric_type)
);
```

---

## 🔧 Service Implementation (Go)

### Package Structure

```
monitoring-service/
├── internal/
│   ├── models/
│   │   └── anomaly.go                  # Data models
│   ├── services/
│   │   ├── anomaly_detection.go        # Main detection service
│   │   ├── baseline_calculator.go      # Baseline calculation
│   │   ├── metric_collector.go         # Metric collection
│   │   └── anomaly_notifier.go         # Alert notifications
│   ├── handlers/
│   │   └── anomaly_handler.go          # HTTP API endpoints
│   └── jobs/
│       ├── baseline_update_job.go      # Periodic baseline updates
│       ├── anomaly_detection_job.go    # Continuous anomaly detection
│       └── metric_collection_job.go    # Periodic metric collection
```

### Core Algorithms

#### 1. Z-Score Anomaly Detection

```go
// Z-score formula: z = (x - μ) / σ
// where x = actual value, μ = mean, σ = standard deviation

type ZScoreDetector struct {
    mean   float64
    stdDev float64
}

func (d *ZScoreDetector) Detect(value float64) (isAnomaly bool, score float64, severity string) {
    if d.stdDev == 0 {
        return false, 0, ""
    }

    zScore := math.Abs((value - d.mean) / d.stdDev)

    if zScore >= 4.0 {
        return true, zScore, "critical"
    } else if zScore >= 3.0 {
        return true, zScore, "major"
    } else if zScore >= 2.0 {
        return true, zScore, "minor"
    }

    return false, zScore, ""
}
```

#### 2. Exponential Weighted Moving Average (EWMA)

```go
type EWMADetector struct {
    alpha    float64  // Smoothing factor (0-1)
    ewma     float64  // Current EWMA value
    variance float64  // Current variance
}

func (d *EWMADetector) Update(value float64) {
    if d.ewma == 0 {
        d.ewma = value
        return
    }

    // Update EWMA
    d.ewma = d.alpha*value + (1-d.alpha)*d.ewma

    // Update variance
    delta := value - d.ewma
    d.variance = (1-d.alpha)*d.variance + d.alpha*delta*delta
}

func (d *EWMADetector) Detect(value float64) (isAnomaly bool, score float64) {
    d.Update(value)

    stdDev := math.Sqrt(d.variance)
    if stdDev == 0 {
        return false, 0
    }

    zScore := math.Abs((value - d.ewma) / stdDev)
    return zScore >= 3.0, zScore
}
```

#### 3. Percentile-Based Detection

```go
type PercentileDetector struct {
    p50 float64
    p95 float64
    p99 float64
}

func (d *PercentileDetector) Detect(value float64) (isAnomaly bool, severity string) {
    if value > d.p99 {
        return true, "critical"
    } else if value > d.p95 {
        return true, "major"
    }
    return false, ""
}
```

#### 4. Seasonal Decomposition (Simplified)

```go
type SeasonalDetector struct {
    hourlyBaselines [24]float64  // Average for each hour
    hourlyStdDevs   [24]float64  // StdDev for each hour
}

func (d *SeasonalDetector) Detect(value float64, hour int) (isAnomaly bool, score float64) {
    mean := d.hourlyBaselines[hour]
    stdDev := d.hourlyStdDevs[hour]

    if stdDev == 0 {
        return false, 0
    }

    zScore := math.Abs((value - mean) / stdDev)
    return zScore >= 3.0, zScore
}
```

---

## 🔄 Operational Flow

### 1. Metric Collection (Every 1 Minute)

```
┌─────────────────┐
│  Monitor Check  │ → Response time: 250ms
│  Completes      │   Error rate: 0.5%
└────────┬────────┘   Request count: 120
         │
         ↓
┌─────────────────┐
│ Store Metrics   │ → INSERT INTO metric_snapshots
│ in Database     │   (tenant_id, monitor_id, metric_type, metric_value, timestamp)
└─────────────────┘
```

### 2. Baseline Calculation (Every 6 Hours)

```
┌──────────────────┐
│  Fetch Last 7    │ → SELECT * FROM metric_snapshots
│  Days of Data    │   WHERE timestamp >= NOW() - INTERVAL '7 days'
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Calculate       │ → mean = AVG(metric_value)
│  Statistics      │   std_dev = STDDEV(metric_value)
│                  │   p50, p95, p99 = PERCENTILE_CONT(...)
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Store Baseline  │ → UPSERT INTO anomaly_baselines
│  in Database     │   (tenant_id, monitor_id, metric_type, mean_value, std_dev, ...)
└──────────────────┘
```

### 3. Anomaly Detection (Every 1 Minute)

```
┌──────────────────┐
│  New Metric      │ → Response time: 1500ms
│  Value Arrives   │
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Fetch Baseline  │ → SELECT * FROM anomaly_baselines
│  for Monitor     │   WHERE monitor_id = ? AND metric_type = 'response_time'
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Run Detection   │ → Z-score = (1500 - 250) / 100 = 12.5
│  Algorithm       │   Severity = critical (z > 4.0)
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Store Anomaly   │ → INSERT INTO detected_anomalies
│  Record          │   (tenant_id, monitor_id, actual_value, expected_value,
│                  │    deviation_score, severity, status='open')
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│  Send            │ → Call escalation policy
│  Notification    │   Send Slack/PagerDuty/Email
└──────────────────┘
```

---

## 📡 API Endpoints

### 1. Get Anomalies

```
GET /api/v1/anomalies
Query params:
  - monitor_id (optional)
  - status (optional): open, acknowledged, resolved
  - severity (optional): minor, major, critical
  - from (optional): timestamp
  - to (optional): timestamp
  - limit (optional): default 100

Response 200:
{
  "anomalies": [
    {
      "id": 1,
      "monitor_id": 5,
      "monitor_name": "API Gateway",
      "metric_type": "response_time",
      "detected_at": "2025-01-25T10:15:00Z",
      "actual_value": 1500,
      "expected_value": 250,
      "deviation_score": 12.5,
      "severity": "critical",
      "status": "open"
    }
  ],
  "total": 25,
  "page": 1
}
```

### 2. Get Anomaly Details

```
GET /api/v1/anomalies/:id

Response 200:
{
  "id": 1,
  "monitor_id": 5,
  "monitor_name": "API Gateway",
  "metric_type": "response_time",
  "detected_at": "2025-01-25T10:15:00Z",
  "actual_value": 1500,
  "expected_value": 250,
  "deviation_score": 12.5,
  "severity": "critical",
  "status": "open",
  "context": {
    "7day_mean": 250,
    "7day_std_dev": 50,
    "30day_mean": 240,
    "historical_anomalies_count": 3
  }
}
```

### 3. Acknowledge Anomaly

```
POST /api/v1/anomalies/:id/acknowledge
Request body:
{
  "notes": "Investigating database slow query"
}

Response 200:
{
  "message": "Anomaly acknowledged",
  "acknowledged_at": "2025-01-25T10:20:00Z"
}
```

### 4. Resolve Anomaly

```
POST /api/v1/anomalies/:id/resolve
Request body:
{
  "resolution": "false_positive" | "resolved",
  "notes": "Database query optimized"
}

Response 200:
{
  "message": "Anomaly resolved",
  "resolved_at": "2025-01-25T10:30:00Z"
}
```

### 5. Get Baselines

```
GET /api/v1/anomalies/baselines/:monitor_id

Response 200:
{
  "monitor_id": 5,
  "monitor_name": "API Gateway",
  "baselines": [
    {
      "metric_type": "response_time",
      "baseline_type": "7day",
      "mean_value": 250,
      "std_dev": 50,
      "p50": 220,
      "p95": 350,
      "p99": 450,
      "calculated_at": "2025-01-25T09:00:00Z"
    }
  ]
}
```

### 6. Update Detection Config

```
PUT /api/v1/anomalies/config
Request body:
{
  "monitor_id": 5,
  "metric_type": "response_time",
  "sensitivity": "high",  // low, medium, high
  "z_score_threshold_minor": 2.0,
  "z_score_threshold_major": 2.5,
  "z_score_threshold_critical": 3.0,
  "notification_enabled": true,
  "escalation_policy_id": 3
}

Response 200:
{
  "message": "Configuration updated successfully"
}
```

### 7. Get Anomaly Statistics

```
GET /api/v1/anomalies/stats
Query params:
  - period: '24h', '7d', '30d'

Response 200:
{
  "period": "7d",
  "total_anomalies": 45,
  "by_severity": {
    "minor": 20,
    "major": 18,
    "critical": 7
  },
  "by_status": {
    "open": 5,
    "acknowledged": 10,
    "resolved": 25,
    "false_positive": 5
  },
  "false_positive_rate": 11.1,
  "most_affected_monitors": [...]
}
```

---

## 🎨 UI Components

### 1. Anomaly Dashboard

**Location**: `app/admin/anomalies/page.tsx`

**Features**:
- Anomaly timeline (line chart with markers)
- Active anomalies list (table)
- Anomaly severity distribution (pie chart)
- Recent anomalies feed (list with actions)
- Quick filters (status, severity, monitor)

### 2. Anomaly Detail View

**Features**:
- Metric chart with anomaly highlighted
- Baseline comparison
- Context (7-day and 30-day baselines)
- Historical anomalies for same monitor
- Acknowledge/resolve actions
- Notes and resolution tracking

### 3. Baseline Visualization

**Features**:
- Expected range (mean ± 2σ)
- Actual values overlaid
- Hour-of-day pattern chart
- Day-of-week pattern chart

### 4. Configuration UI

**Features**:
- Sensitivity slider (low/medium/high)
- Custom Z-score thresholds
- Enable/disable per monitor or metric
- Link to escalation policies

---

## 🧪 Testing Strategy

### Unit Tests

**Baseline Calculation**:
- Test mean and stddev calculations
- Test percentile calculations (P50, P95, P99)
- Test with edge cases (all same values, outliers)

**Anomaly Detection**:
- Test Z-score calculation
- Test severity classification
- Test false positive scenarios
- Test with different sensitivities

### Integration Tests

**End-to-End Flow**:
1. Collect metrics → Store in DB
2. Calculate baseline → Verify correct values
3. Detect anomaly → Verify detection
4. Send notification → Verify escalation

### Performance Tests

**Load Testing**:
- 1000 monitors × 3 metrics = 3000 checks/minute
- Baseline calculation for 1000 monitors
- Query performance with 1M+ metric snapshots

---

## 📊 Expected Performance

**Metric Collection**:
- 1000 monitors, 3 metrics each
- 3000 inserts/minute
- Estimated DB size: 4.3M rows/month

**Baseline Calculation**:
- 1000 monitors, 3 metrics each
- 3000 baseline calculations every 6 hours
- Estimated time: 5-10 minutes

**Anomaly Detection**:
- Real-time (within 1 minute of metric collection)
- 3000 detection checks/minute
- Estimated latency: <10ms per check

---

## 🚨 Challenges and Mitigations

### Challenge 1: False Positives

**Problem**: Too many false alarms lead to alert fatigue

**Mitigation**:
- Adjustable sensitivity levels
- Mark anomalies as false positives
- Learn from feedback (reduce sensitivity for specific monitors)
- Require multiple consecutive anomalies before alerting

### Challenge 2: Cold Start (Insufficient Data)

**Problem**: New monitors don't have enough data for baselines

**Mitigation**:
- Minimum 50 samples required before enabling detection
- Use tenant-wide baseline as fallback
- Gradually reduce threshold as more data is collected

### Challenge 3: Seasonal Patterns

**Problem**: Daily/weekly patterns cause false anomalies (e.g., traffic spike at 9 AM)

**Mitigation**:
- Hour-of-day baselines
- Day-of-week baselines
- Weekend vs weekday patterns

### Challenge 4: Storage Growth

**Problem**: Metric snapshots table grows rapidly

**Mitigation**:
- Retention policy (90 days)
- Downsample old data (e.g., hourly aggregates after 30 days)
- Partition table by month

---

## 🎯 Success Criteria

**Technical**:
- ✅ False positive rate <15%
- ✅ Detection latency <1 minute
- ✅ Baseline calculation <10 minutes
- ✅ Support 1000+ monitors per tenant

**Business**:
- ✅ Catch 80%+ of incidents before they become critical
- ✅ Reduce MTTR by 20% (early detection)
- ✅ User satisfaction: 4+ stars

---

## 📅 Implementation Timeline

**Day 1-2: Database & Models**
- Create database schema
- Implement Go models
- Write migration scripts

**Day 3-4: Core Services**
- Implement baseline calculator
- Implement anomaly detection algorithms
- Implement metric collector

**Day 5: API & Integration**
- Implement HTTP handlers
- Integrate with existing escalation policies
- Add background jobs

**Day 6: UI Implementation**
- Create anomaly dashboard
- Create baseline visualization
- Create configuration UI

**Day 7: Testing & Documentation**
- Write unit tests
- Perform integration testing
- Create user documentation

---

## 💡 Future Enhancements

**Phase 2: Machine Learning**
- Integrate Prophet for time-series forecasting
- LSTM for complex patterns
- Auto-tuning sensitivity based on feedback

**Phase 2: Advanced Features**
- Multi-metric correlation (e.g., high response time + high error rate)
- Anomaly clustering (group related anomalies)
- Root cause analysis suggestions
- Predictive anomalies (forecast future issues)

---

## ✅ Deliverables

**Backend**:
- [ ] Database migrations (4 tables)
- [ ] Go models (anomaly.go)
- [ ] Services (4 services)
- [ ] API handlers (7 endpoints)
- [ ] Background jobs (3 jobs)

**Frontend**:
- [ ] Anomaly dashboard page
- [ ] Anomaly detail view
- [ ] Baseline visualization
- [ ] Configuration UI
- [ ] API client (anomaly.ts)

**Documentation**:
- [x] Architecture design (this document)
- [ ] API documentation
- [ ] User guide
- [ ] Deployment guide

---

**Status**: 🎯 Ready to Implement
**Estimated Effort**: 5-7 days
**Complexity**: High
**Business Value**: Very High (proactive incident prevention)

Generated: January 2025
Sprint: Phase 3, Week 13
