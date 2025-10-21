-- Migration: Add Monitors and Auto-Incident Creation
-- Date: 2025-10-21
-- Phase: 1, Week 2 - Auto-Incident Creation

-- ====================================
-- 1. MONITORS TABLE
-- ====================================
-- Stores health check/monitoring configuration for components
CREATE TABLE IF NOT EXISTS monitors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    component_id UUID NOT NULL,  -- References components.id in tenant_admin_db
    name VARCHAR(255) NOT NULL,
    monitor_type VARCHAR(50) NOT NULL, -- http, ping, tcp, ssl, heartbeat
    check_url VARCHAR(1000),  -- For HTTP/ping monitors
    check_interval_seconds INT NOT NULL DEFAULT 60,  -- How often to check (seconds)
    timeout_seconds INT NOT NULL DEFAULT 30,

    -- HTTP-specific settings
    http_method VARCHAR(10) DEFAULT 'GET',  -- GET, POST, HEAD
    http_headers TEXT,  -- JSON: {"Authorization": "Bearer token"}
    http_body TEXT,
    expected_status_codes TEXT DEFAULT '200,201,204',  -- Comma-separated

    -- Advanced settings
    follow_redirects BOOLEAN DEFAULT true,
    verify_ssl BOOLEAN DEFAULT true,

    -- Multi-location monitoring
    enabled_locations TEXT,  -- JSON array: [1, 2, 3] (location IDs), NULL = all locations

    -- Auto-incident creation settings
    auto_create_incidents BOOLEAN DEFAULT true,
    failure_threshold INT DEFAULT 3,  -- Number of consecutive failures before creating incident
    consecutive_failures INT DEFAULT 0,  -- Current count of consecutive failures
    last_incident_id UUID,  -- Last auto-created incident (for auto-resolution)

    -- Status tracking
    is_active BOOLEAN DEFAULT true,
    current_status VARCHAR(20) DEFAULT 'unknown',  -- operational, degraded, down, unknown
    last_check_at TIMESTAMP,
    last_success_at TIMESTAMP,
    last_failure_at TIMESTAMP,
    uptime_percentage DECIMAL(5,2) DEFAULT 100.00,

    -- Maintenance windows
    in_maintenance BOOLEAN DEFAULT false,
    maintenance_until TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_monitors_tenant ON monitors(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monitors_component ON monitors(component_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monitors_active ON monitors(is_active) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX idx_monitors_status ON monitors(current_status);
CREATE INDEX idx_monitors_failures ON monitors(consecutive_failures) WHERE consecutive_failures >= 3;
CREATE UNIQUE INDEX idx_monitors_tenant_component ON monitors(tenant_id, component_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE monitors IS 'Health check and monitoring configuration for components';
COMMENT ON COLUMN monitors.failure_threshold IS 'Number of consecutive failures before auto-creating an incident';
COMMENT ON COLUMN monitors.consecutive_failures IS 'Current consecutive failure count (resets on success)';
COMMENT ON COLUMN monitors.last_incident_id IS 'UUID of last auto-created incident for auto-resolution';

-- ====================================
-- 2. UPDATE MONITORING_RESULTS TABLE
-- ====================================
-- Add monitor_id column to link results to monitors
DO $$
BEGIN
    -- Check if monitor_id column doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'monitoring_results' AND column_name = 'monitor_id'
    ) THEN
        -- For partitioned tables, we need to add the column to the parent table
        ALTER TABLE monitoring_results ADD COLUMN IF NOT EXISTS monitor_id BIGINT;
        CREATE INDEX IF NOT EXISTS idx_monitoring_results_monitor ON monitoring_results(monitor_id);
    END IF;
END $$;

-- ====================================
-- 3. AUTO-INCIDENT TRACKING TABLE
-- ====================================
-- Tracks which incidents were auto-created and their resolution status
CREATE TABLE IF NOT EXISTS auto_incidents (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    monitor_id BIGINT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    incident_id UUID NOT NULL,  -- References incidents.id in statuspage_incident database
    component_id UUID NOT NULL,

    -- Creation details
    triggered_by_result_id BIGINT,  -- The monitoring result that triggered incident creation
    failure_count INT NOT NULL,  -- How many consecutive failures when created
    created_at TIMESTAMP DEFAULT NOW(),

    -- Resolution tracking
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP,
    resolved_by_result_id BIGINT,  -- The monitoring result that resolved the incident
    auto_resolved BOOLEAN DEFAULT false,  -- true if auto-resolved, false if manually resolved

    -- Metadata
    error_message TEXT,
    affected_locations TEXT  -- JSON array of location IDs where failures occurred
);

CREATE INDEX idx_auto_incidents_tenant ON auto_incidents(tenant_id);
CREATE INDEX idx_auto_incidents_monitor ON auto_incidents(monitor_id);
CREATE INDEX idx_auto_incidents_incident ON auto_incidents(incident_id);
CREATE INDEX idx_auto_incidents_unresolved ON auto_incidents(resolved) WHERE resolved = false;

COMMENT ON TABLE auto_incidents IS 'Tracks auto-created incidents and their resolution status';
COMMENT ON COLUMN auto_incidents.auto_resolved IS 'True if incident was auto-resolved when checks passed';

-- ====================================
-- 4. MONITOR NOTIFICATION SETTINGS
-- ====================================
-- Stores notification preferences for monitors
CREATE TABLE IF NOT EXISTS monitor_notifications (
    id BIGSERIAL PRIMARY KEY,
    monitor_id BIGINT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,

    -- Notification channels
    notify_email BOOLEAN DEFAULT true,
    notify_sms BOOLEAN DEFAULT false,
    notify_slack BOOLEAN DEFAULT false,
    notify_webhook BOOLEAN DEFAULT false,

    -- Notification triggers
    notify_on_failure BOOLEAN DEFAULT true,
    notify_on_recovery BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT false,

    -- Escalation
    escalation_policy_id BIGINT REFERENCES escalation_policies(id),
    on_call_schedule_id BIGINT REFERENCES on_call_schedules(id),

    -- Custom recipients (overrides default tenant notification settings)
    custom_email_recipients TEXT,  -- JSON array: ["user1@example.com", "user2@example.com"]
    custom_webhook_url VARCHAR(1000),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_monitor_notifications_monitor ON monitor_notifications(monitor_id);

COMMENT ON TABLE monitor_notifications IS 'Notification preferences for individual monitors';

-- ====================================
-- 5. MAINTENANCE WINDOWS TABLE
-- ====================================
-- Scheduled maintenance windows (suppress alerts during maintenance)
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Schedule
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,

    -- Affected monitors
    affected_monitors TEXT,  -- JSON array: [1, 2, 3] (monitor IDs), NULL = all monitors

    -- Status
    is_active BOOLEAN DEFAULT false,

    -- Options
    suppress_notifications BOOLEAN DEFAULT true,
    auto_update_status_page BOOLEAN DEFAULT true,  -- Automatically show maintenance banner

    created_by UUID,  -- User ID who created the window
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_maintenance_tenant ON maintenance_windows(tenant_id);
CREATE INDEX idx_maintenance_active ON maintenance_windows(is_active) WHERE is_active = true;
CREATE INDEX idx_maintenance_schedule ON maintenance_windows(starts_at, ends_at);

COMMENT ON TABLE maintenance_windows IS 'Scheduled maintenance windows for alert suppression';

-- ====================================
-- 6. MONITOR STATUS HISTORY
-- ====================================
-- Historical status changes for monitors (for uptime calculations and graphs)
CREATE TABLE IF NOT EXISTS monitor_status_history (
    id BIGSERIAL PRIMARY KEY,
    monitor_id BIGINT NOT NULL,
    tenant_id UUID NOT NULL,

    -- Status change
    previous_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,

    -- Timing
    changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    duration_seconds BIGINT,  -- How long previous status lasted

    -- Context
    triggered_by VARCHAR(50),  -- 'health_check', 'manual', 'maintenance', 'auto_recovery'
    error_message TEXT,

    -- Metadata
    metadata TEXT  -- JSON: additional context
);

-- Partition by month for efficient querying
CREATE INDEX idx_status_history_monitor ON monitor_status_history(monitor_id, changed_at DESC);
CREATE INDEX idx_status_history_tenant ON monitor_status_history(tenant_id, changed_at DESC);

COMMENT ON TABLE monitor_status_history IS 'Historical record of status changes for uptime tracking';

-- ====================================
-- 7. SAMPLE DATA (for testing)
-- ====================================
-- Uncomment to insert test data
/*
-- Create a test monitor
INSERT INTO monitors (tenant_id, component_id, name, monitor_type, check_url, check_interval_seconds, auto_create_incidents, failure_threshold)
VALUES (
    '00000000-0000-0000-0000-000000000001',  -- Test tenant ID
    '00000000-0000-0000-0000-000000000001',  -- Test component ID
    'API Health Check',
    'http',
    'https://api.example.com/health',
    60,  -- Check every 60 seconds
    true,  -- Auto-create incidents
    3  -- After 3 consecutive failures
);
*/
