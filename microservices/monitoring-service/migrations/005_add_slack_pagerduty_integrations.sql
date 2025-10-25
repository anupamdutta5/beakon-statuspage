-- Migration 005: Add Slack and PagerDuty Integration Tables
-- Created: 2025-10-22
-- Description: Creates tables for Slack and PagerDuty integrations

-- ========================================
-- SLACK INTEGRATION TABLES
-- ========================================

CREATE TABLE IF NOT EXISTS slack_integrations (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    workspace_name VARCHAR(255) NOT NULL,
    webhook_url TEXT NOT NULL,
    default_channel VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    notify_on_down BOOLEAN DEFAULT true,
    notify_on_up BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT true,
    notify_on_maintenance BOOLEAN DEFAULT false,
    mention_users TEXT,
    mention_channel BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_slack_integrations_tenant ON slack_integrations(tenant_id);
CREATE INDEX idx_slack_integrations_active ON slack_integrations(is_active) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS slack_channel_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL REFERENCES slack_integrations(id) ON DELETE CASCADE,
    monitor_id BIGINT,
    channel_name VARCHAR(255) NOT NULL,
    channel_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_slack_subscriptions_integration ON slack_channel_subscriptions(integration_id);
CREATE INDEX idx_slack_subscriptions_monitor ON slack_channel_subscriptions(monitor_id);

CREATE TABLE IF NOT EXISTS slack_notifications (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL REFERENCES slack_integrations(id) ON DELETE CASCADE,
    monitor_id BIGINT,
    event_type VARCHAR(50) NOT NULL,
    event_data TEXT,
    status VARCHAR(50) NOT NULL,
    sent_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_slack_notifications_integration ON slack_notifications(integration_id);
CREATE INDEX idx_slack_notifications_status ON slack_notifications(status);
CREATE INDEX idx_slack_notifications_sent ON slack_notifications(sent_at DESC);

-- ========================================
-- PAGERDUTY INTEGRATION TABLES
-- ========================================

CREATE TABLE IF NOT EXISTS pagerduty_integrations (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    integration_name VARCHAR(255) NOT NULL,
    integration_key VARCHAR(255) NOT NULL,
    api_key VARCHAR(255),
    service_id VARCHAR(255),
    service_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    auto_resolve BOOLEAN DEFAULT true,
    severity VARCHAR(50) DEFAULT 'error',
    notify_on_down BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT true,
    notify_on_maintenance BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_pagerduty_integrations_tenant ON pagerduty_integrations(tenant_id);
CREATE INDEX idx_pagerduty_integrations_active ON pagerduty_integrations(is_active) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS pagerduty_monitor_mappings (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL REFERENCES pagerduty_integrations(id) ON DELETE CASCADE,
    monitor_id BIGINT NOT NULL,
    severity VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_pagerduty_mappings_integration ON pagerduty_monitor_mappings(integration_id);
CREATE INDEX idx_pagerduty_mappings_monitor ON pagerduty_monitor_mappings(monitor_id);

CREATE TABLE IF NOT EXISTS pagerduty_incidents (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL REFERENCES pagerduty_integrations(id) ON DELETE CASCADE,
    monitor_id BIGINT NOT NULL,
    incident_key VARCHAR(255) NOT NULL UNIQUE,
    pd_incident_id VARCHAR(255),
    event_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    severity VARCHAR(50),
    summary TEXT,
    source VARCHAR(255),
    component VARCHAR(255),
    details JSONB,
    triggered_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    response_code INT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_pagerduty_incidents_integration ON pagerduty_incidents(integration_id);
CREATE INDEX idx_pagerduty_incidents_monitor ON pagerduty_incidents(monitor_id);
CREATE INDEX idx_pagerduty_incidents_key ON pagerduty_incidents(incident_key);
CREATE INDEX idx_pagerduty_incidents_status ON pagerduty_incidents(status);

-- ========================================
-- TRIGGERS FOR UPDATED_AT
-- ========================================

-- Slack integrations trigger
CREATE OR REPLACE FUNCTION update_slack_integrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_slack_integrations_updated_at
    BEFORE UPDATE ON slack_integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_slack_integrations_updated_at();

-- Slack channel subscriptions trigger
CREATE OR REPLACE FUNCTION update_slack_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_slack_subscriptions_updated_at
    BEFORE UPDATE ON slack_channel_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_slack_subscriptions_updated_at();

-- PagerDuty integrations trigger
CREATE OR REPLACE FUNCTION update_pagerduty_integrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_pagerduty_integrations_updated_at
    BEFORE UPDATE ON pagerduty_integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_pagerduty_integrations_updated_at();

-- PagerDuty monitor mappings trigger
CREATE OR REPLACE FUNCTION update_pagerduty_mappings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_pagerduty_mappings_updated_at
    BEFORE UPDATE ON pagerduty_monitor_mappings
    FOR EACH ROW
    EXECUTE FUNCTION update_pagerduty_mappings_updated_at();

-- PagerDuty incidents trigger
CREATE OR REPLACE FUNCTION update_pagerduty_incidents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_pagerduty_incidents_updated_at
    BEFORE UPDATE ON pagerduty_incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_pagerduty_incidents_updated_at();

-- ========================================
-- VERIFICATION QUERIES
-- ========================================

-- Uncomment to verify tables were created
-- SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%slack%' OR table_name LIKE '%pagerduty%' ORDER BY table_name;

-- Uncomment to verify indexes
-- SELECT indexname FROM pg_indexes WHERE tablename LIKE '%slack%' OR tablename LIKE '%pagerduty%' ORDER BY indexname;
