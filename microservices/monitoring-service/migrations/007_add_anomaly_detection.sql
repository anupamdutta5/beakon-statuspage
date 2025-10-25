-- Migration: Add anomaly detection tables
-- Created: 2025-01-25
-- Updated: 2025-01-25 (Changed tenant_id from UUID to BIGINT to match monitoring service schema)
-- Description: Adds tables for metric collection, baseline calculation, and anomaly detection

-- Table 1: Metric Snapshots (time-series metric storage)
CREATE TABLE IF NOT EXISTS metric_snapshots (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    monitor_id BIGINT,  -- NULL for tenant-wide metrics, references uptime_checks
    metric_type VARCHAR(50) NOT NULL,  -- 'response_time', 'error_rate', 'request_volume', 'status_changes'
    metric_value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Foreign key to uptime_checks table
    CONSTRAINT fk_metric_monitor FOREIGN KEY (monitor_id)
        REFERENCES uptime_checks(id) ON DELETE CASCADE
);

-- Indexes for metric_snapshots
CREATE INDEX idx_metric_snapshots_tenant ON metric_snapshots(tenant_id);
CREATE INDEX idx_metric_snapshots_monitor ON metric_snapshots(monitor_id);
CREATE INDEX idx_metric_snapshots_type ON metric_snapshots(metric_type);
CREATE INDEX idx_metric_snapshots_timestamp ON metric_snapshots(timestamp DESC);
CREATE INDEX idx_metric_snapshots_tenant_monitor_time ON metric_snapshots(tenant_id, monitor_id, timestamp DESC);
CREATE INDEX idx_metric_snapshots_tenant_type_time ON metric_snapshots(tenant_id, metric_type, timestamp DESC);

-- Partition by month for better performance (optional, for high-volume deployments)
-- This can be enabled later if metric_snapshots grows beyond 10M rows

COMMENT ON TABLE metric_snapshots IS 'Time-series storage for metrics used in anomaly detection and analytics';
COMMENT ON COLUMN metric_snapshots.metric_type IS 'Type of metric: response_time (ms), error_rate (%), request_volume (count), status_changes (count)';
COMMENT ON COLUMN metric_snapshots.metric_value IS 'Numeric value of the metric at this timestamp';

---

-- Table 2: Anomaly Baselines (calculated baseline statistics)
CREATE TABLE IF NOT EXISTS anomaly_baselines (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    monitor_id BIGINT,  -- NULL for tenant-wide baselines
    metric_type VARCHAR(50) NOT NULL,
    baseline_type VARCHAR(20) NOT NULL,  -- '7day', '30day', 'hourly', 'daily', 'weekly'
    hour_of_day INT CHECK (hour_of_day >= 0 AND hour_of_day <= 23),  -- For hourly patterns
    day_of_week INT CHECK (day_of_week >= 0 AND day_of_week <= 6),   -- For weekly patterns (0=Sunday)

    -- Statistical measures
    mean_value DOUBLE PRECISION NOT NULL,
    std_dev DOUBLE PRECISION NOT NULL,
    min_value DOUBLE PRECISION,
    max_value DOUBLE PRECISION,
    p50 DOUBLE PRECISION,  -- Median
    p95 DOUBLE PRECISION,
    p99 DOUBLE PRECISION,

    -- Metadata
    sample_count INT NOT NULL,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,  -- When to recalculate

    -- Ensure unique baselines per configuration
    UNIQUE (tenant_id, monitor_id, metric_type, baseline_type, hour_of_day, day_of_week),

    -- Foreign key
    CONSTRAINT fk_baseline_monitor FOREIGN KEY (monitor_id)
        REFERENCES uptime_checks(id) ON DELETE CASCADE
);

-- Indexes for anomaly_baselines
CREATE INDEX idx_anomaly_baselines_tenant ON anomaly_baselines(tenant_id);
CREATE INDEX idx_anomaly_baselines_monitor ON anomaly_baselines(monitor_id);
CREATE INDEX idx_anomaly_baselines_type ON anomaly_baselines(metric_type);
CREATE INDEX idx_anomaly_baselines_expires ON anomaly_baselines(expires_at);
CREATE INDEX idx_anomaly_baselines_tenant_monitor_type ON anomaly_baselines(tenant_id, monitor_id, metric_type);

COMMENT ON TABLE anomaly_baselines IS 'Pre-calculated baseline statistics for anomaly detection';
COMMENT ON COLUMN anomaly_baselines.baseline_type IS 'Timeframe for baseline: 7day, 30day, hourly (by hour), daily (by day of month), weekly (by day of week)';
COMMENT ON COLUMN anomaly_baselines.hour_of_day IS 'Hour 0-23 for hourly baselines, NULL for non-hourly baselines';
COMMENT ON COLUMN anomaly_baselines.day_of_week IS 'Day 0-6 (0=Sunday) for weekly baselines, NULL for non-weekly baselines';

---

-- Table 3: Detected Anomalies (anomaly records)
CREATE TABLE IF NOT EXISTS detected_anomalies (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    monitor_id BIGINT,
    metric_type VARCHAR(50) NOT NULL,

    -- Timestamp and values
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    actual_value DOUBLE PRECISION NOT NULL,
    expected_value DOUBLE PRECISION NOT NULL,
    deviation_score DOUBLE PRECISION NOT NULL,  -- Z-score or similar

    -- Classification
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('minor', 'major', 'critical')),
    detection_method VARCHAR(50) DEFAULT 'z_score',  -- 'z_score', 'ewma', 'percentile', 'seasonal'

    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved', 'false_positive')),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by BIGINT,  -- User ID
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Foreign keys
    CONSTRAINT fk_anomaly_monitor FOREIGN KEY (monitor_id)
        REFERENCES uptime_checks(id) ON DELETE CASCADE
);

-- Indexes for detected_anomalies
CREATE INDEX idx_detected_anomalies_tenant ON detected_anomalies(tenant_id);
CREATE INDEX idx_detected_anomalies_monitor ON detected_anomalies(monitor_id);
CREATE INDEX idx_detected_anomalies_detected_at ON detected_anomalies(detected_at DESC);
CREATE INDEX idx_detected_anomalies_status ON detected_anomalies(status);
CREATE INDEX idx_detected_anomalies_severity ON detected_anomalies(severity);
CREATE INDEX idx_detected_anomalies_tenant_status ON detected_anomalies(tenant_id, status);
CREATE INDEX idx_detected_anomalies_tenant_monitor_status ON detected_anomalies(tenant_id, monitor_id, status);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_detected_anomalies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_detected_anomalies_updated_at
    BEFORE UPDATE ON detected_anomalies
    FOR EACH ROW
    EXECUTE FUNCTION update_detected_anomalies_updated_at();

COMMENT ON TABLE detected_anomalies IS 'Records of detected anomalies with tracking and resolution';
COMMENT ON COLUMN detected_anomalies.deviation_score IS 'Z-score or normalized deviation score indicating how far the value is from expected';
COMMENT ON COLUMN detected_anomalies.detection_method IS 'Algorithm used to detect this anomaly: z_score, ewma, percentile, seasonal';
COMMENT ON COLUMN detected_anomalies.status IS 'Lifecycle status: open (new), acknowledged (team notified), resolved (fixed), false_positive (ignored)';

---

-- Table 4: Anomaly Detection Configuration (per-monitor settings)
CREATE TABLE IF NOT EXISTS anomaly_detection_config (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    monitor_id BIGINT,  -- NULL for tenant-wide defaults
    metric_type VARCHAR(50),  -- NULL for all metrics

    -- Enable/disable
    enabled BOOLEAN DEFAULT TRUE,

    -- Sensitivity settings
    sensitivity VARCHAR(20) DEFAULT 'medium' CHECK (sensitivity IN ('low', 'medium', 'high')),
    min_baseline_samples INT DEFAULT 50,  -- Minimum samples required before detection

    -- Z-score thresholds
    z_score_threshold_minor DOUBLE PRECISION DEFAULT 2.0,
    z_score_threshold_major DOUBLE PRECISION DEFAULT 3.0,
    z_score_threshold_critical DOUBLE PRECISION DEFAULT 4.0,

    -- Notification settings
    notification_enabled BOOLEAN DEFAULT TRUE,
    notification_cooldown_minutes INT DEFAULT 30,  -- Avoid alert spam
    require_consecutive_anomalies INT DEFAULT 1,  -- How many in a row before alerting

    -- Integration
    escalation_policy_id BIGINT,  -- Link to escalation_policies table (nullable, may not exist)

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Ensure unique configs
    UNIQUE (tenant_id, monitor_id, metric_type),

    -- Foreign keys
    CONSTRAINT fk_config_monitor FOREIGN KEY (monitor_id)
        REFERENCES uptime_checks(id) ON DELETE CASCADE
    -- Note: escalation_policy_id FK not added since escalation_policies table may not exist
);

-- Indexes for anomaly_detection_config
CREATE INDEX idx_anomaly_detection_config_tenant ON anomaly_detection_config(tenant_id);
CREATE INDEX idx_anomaly_detection_config_monitor ON anomaly_detection_config(monitor_id);
CREATE INDEX idx_anomaly_detection_config_enabled ON anomaly_detection_config(enabled);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_anomaly_detection_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_anomaly_detection_config_updated_at
    BEFORE UPDATE ON anomaly_detection_config
    FOR EACH ROW
    EXECUTE FUNCTION update_anomaly_detection_config_updated_at();

COMMENT ON TABLE anomaly_detection_config IS 'Configuration for anomaly detection per monitor or tenant-wide';
COMMENT ON COLUMN anomaly_detection_config.sensitivity IS 'Detection sensitivity: low (fewer alerts), medium (balanced), high (more alerts)';
COMMENT ON COLUMN anomaly_detection_config.require_consecutive_anomalies IS 'Number of consecutive anomalies required before triggering alert (reduces false positives)';
COMMENT ON COLUMN anomaly_detection_config.notification_cooldown_minutes IS 'Minimum minutes between notifications for same monitor to avoid alert fatigue';

---

-- View: Enriched anomalies with monitor details
CREATE OR REPLACE VIEW anomaly_details AS
SELECT
    da.id,
    da.tenant_id,
    da.monitor_id,
    uc.name AS monitor_name,
    uc.url AS monitor_url,
    da.metric_type,
    da.detected_at,
    da.actual_value,
    da.expected_value,
    da.deviation_score,
    da.severity,
    da.detection_method,
    da.status,
    da.acknowledged_at,
    da.acknowledged_by,
    da.resolved_at,
    da.resolution_notes,
    da.created_at,
    da.updated_at
FROM detected_anomalies da
LEFT JOIN uptime_checks uc ON da.monitor_id = uc.id;

COMMENT ON VIEW anomaly_details IS 'Enriched view of anomalies with uptime check (monitor) information';

---

-- Insert default tenant-wide configuration for existing tenants
INSERT INTO anomaly_detection_config (tenant_id, monitor_id, metric_type, enabled, sensitivity)
SELECT DISTINCT tenant_id, NULL::BIGINT, NULL::VARCHAR(50), TRUE, 'medium'
FROM uptime_checks
ON CONFLICT (tenant_id, monitor_id, metric_type) DO NOTHING;

COMMENT ON TABLE anomaly_detection_config IS 'Default configuration inserted for all existing tenants';
