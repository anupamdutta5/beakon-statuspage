--
-- Component Service - Initial Schema
-- Database: components
-- Description: Component status management for status pages
-- Based on: internal/models/component.go
--

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Component groups table
CREATE TABLE IF NOT EXISTS component_groups (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    position INTEGER DEFAULT 0,
    is_visible BOOLEAN DEFAULT true
);

CREATE INDEX idx_component_groups_tenant_id ON component_groups(tenant_id);
CREATE INDEX idx_component_groups_deleted_at ON component_groups(deleted_at);
CREATE INDEX idx_component_groups_position ON component_groups(position);

-- Components table
CREATE TABLE IF NOT EXISTS components (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'operational',
    position INTEGER DEFAULT 0,
    is_visible BOOLEAN DEFAULT true,
    group_id BIGINT REFERENCES component_groups(id) ON DELETE SET NULL,
    metadata TEXT
);

CREATE INDEX idx_components_tenant_id ON components(tenant_id);
CREATE INDEX idx_components_deleted_at ON components(deleted_at);
CREATE INDEX idx_components_group_id ON components(group_id);
CREATE INDEX idx_components_status ON components(status);
CREATE INDEX idx_components_position ON components(position);

-- Component statuses table (current status tracking)
CREATE TABLE IF NOT EXISTS component_statuses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    component_id BIGINT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    message TEXT,
    updated_by BIGINT,
    metadata TEXT
);

CREATE INDEX idx_component_statuses_component_id ON component_statuses(component_id);
CREATE INDEX idx_component_statuses_deleted_at ON component_statuses(deleted_at);
CREATE INDEX idx_component_statuses_status ON component_statuses(status);

-- Component history table (status change history)
CREATE TABLE IF NOT EXISTS component_history (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    component_id BIGINT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    message TEXT,
    updated_by BIGINT,
    metadata TEXT
);

CREATE INDEX idx_component_history_component_id ON component_history(component_id);
CREATE INDEX idx_component_history_deleted_at ON component_history(deleted_at);
CREATE INDEX idx_component_history_created_at ON component_history(created_at DESC);

-- Component metrics table
CREATE TABLE IF NOT EXISTS component_metrics (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    component_id BIGINT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    metric_type VARCHAR(100) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50),
    timestamp TIMESTAMP NOT NULL,
    metadata TEXT
);

CREATE INDEX idx_component_metrics_component_id ON component_metrics(component_id);
CREATE INDEX idx_component_metrics_deleted_at ON component_metrics(deleted_at);
CREATE INDEX idx_component_metrics_timestamp ON component_metrics(timestamp DESC);
CREATE INDEX idx_component_metrics_type_time ON component_metrics(component_id, metric_type, timestamp DESC);

-- Component alerts table
CREATE TABLE IF NOT EXISTS component_alerts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    component_id BIGINT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    alert_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    message TEXT,
    threshold DOUBLE PRECISION,
    current_value DOUBLE PRECISION,
    resolved_at TIMESTAMP,
    resolved_by BIGINT,
    metadata TEXT
);

CREATE INDEX idx_component_alerts_component_id ON component_alerts(component_id);
CREATE INDEX idx_component_alerts_deleted_at ON component_alerts(deleted_at);
CREATE INDEX idx_component_alerts_status ON component_alerts(status);
CREATE INDEX idx_component_alerts_type ON component_alerts(alert_type);

-- Component webhooks table
CREATE TABLE IF NOT EXISTS component_webhooks (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    component_id BIGINT NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    url VARCHAR(1000) NOT NULL,
    events TEXT,
    secret VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_triggered TIMESTAMP,
    metadata TEXT
);

CREATE INDEX idx_component_webhooks_component_id ON component_webhooks(component_id);
CREATE INDEX idx_component_webhooks_deleted_at ON component_webhooks(deleted_at);
CREATE INDEX idx_component_webhooks_active ON component_webhooks(is_active);

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for all tables
CREATE TRIGGER update_component_groups_updated_at BEFORE UPDATE ON component_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_components_updated_at BEFORE UPDATE ON components
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_component_statuses_updated_at BEFORE UPDATE ON component_statuses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_component_history_updated_at BEFORE UPDATE ON component_history
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_component_metrics_updated_at BEFORE UPDATE ON component_metrics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_component_alerts_updated_at BEFORE UPDATE ON component_alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_component_webhooks_updated_at BEFORE UPDATE ON component_webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
