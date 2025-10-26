--
-- Incident Service - Initial Schema
-- Database: incidents
-- Description: Incident management, updates, and component impact tracking
--

BEGIN;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================================
-- Incident Severity and Status Types
-- =====================================================================

CREATE TYPE incident_severity AS ENUM ('minor', 'major', 'critical');
CREATE TYPE incident_status AS ENUM ('investigating', 'identified', 'monitoring', 'resolved', 'postmortem');
CREATE TYPE component_status AS ENUM ('operational', 'degraded_performance', 'partial_outage', 'major_outage', 'under_maintenance');

-- =====================================================================
-- Incidents
-- =====================================================================

CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    severity incident_severity DEFAULT 'minor',
    status incident_status DEFAULT 'investigating',
    impact TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    detected_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    closed_at TIMESTAMP,
    created_by UUID,
    assigned_to UUID,
    is_scheduled BOOLEAN DEFAULT false, -- for maintenance windows
    scheduled_start TIMESTAMP,
    scheduled_end TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_incidents_tenant_id ON incidents(tenant_id);
CREATE INDEX idx_incidents_status ON incidents(status);
CREATE INDEX idx_incidents_severity ON incidents(severity);
CREATE INDEX idx_incidents_created_by ON incidents(created_by);
CREATE INDEX idx_incidents_assigned_to ON incidents(assigned_to);
CREATE INDEX idx_incidents_started_at ON incidents(started_at DESC);
CREATE INDEX idx_incidents_resolved_at ON incidents(resolved_at DESC);
CREATE INDEX idx_incidents_deleted_at ON incidents(deleted_at);

-- =====================================================================
-- Incident Updates (Timeline)
-- =====================================================================

CREATE TABLE IF NOT EXISTS incident_updates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    status incident_status NOT NULL,
    message TEXT NOT NULL,
    created_by UUID,
    is_customer_visible BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_incident_updates_incident_id ON incident_updates(incident_id);
CREATE INDEX idx_incident_updates_tenant_id ON incident_updates(tenant_id);
CREATE INDEX idx_incident_updates_created_at ON incident_updates(created_at DESC);

-- =====================================================================
-- Incident Components (Affected components)
-- =====================================================================

CREATE TABLE IF NOT EXISTS incident_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL,
    component_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    component_status component_status NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_incident_components_incident_id ON incident_components(incident_id);
CREATE INDEX idx_incident_components_component_id ON incident_components(component_id);
CREATE INDEX idx_incident_components_tenant_id ON incident_components(tenant_id);
CREATE UNIQUE INDEX idx_incident_components_unique ON incident_components(incident_id, component_id);

-- =====================================================================
-- Incident Subscribers (Notifications)
-- =====================================================================

CREATE TABLE IF NOT EXISTS incident_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    subscriber_id UUID NOT NULL, -- references subscribers table in other service
    notification_sent BOOLEAN DEFAULT false,
    notification_sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_incident_subscribers_incident_id ON incident_subscribers(incident_id);
CREATE INDEX idx_incident_subscribers_subscriber_id ON incident_subscribers(subscriber_id);
CREATE INDEX idx_incident_subscribers_tenant_id ON incident_subscribers(tenant_id);
CREATE UNIQUE INDEX idx_incident_subscribers_unique ON incident_subscribers(incident_id, subscriber_id);

-- =====================================================================
-- Incident Templates (Pre-defined incident templates)
-- =====================================================================

CREATE TABLE IF NOT EXISTS incident_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    title_template VARCHAR(500) NOT NULL,
    description_template TEXT,
    default_severity incident_severity DEFAULT 'minor',
    component_ids JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_incident_templates_tenant_id ON incident_templates(tenant_id);
CREATE INDEX idx_incident_templates_deleted_at ON incident_templates(deleted_at);
CREATE UNIQUE INDEX idx_incident_templates_tenant_name ON incident_templates(tenant_id, name) WHERE deleted_at IS NULL;

-- =====================================================================
-- Incident Metrics (For analytics)
-- =====================================================================

CREATE TABLE IF NOT EXISTS incident_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    time_to_detect INTEGER, -- seconds
    time_to_acknowledge INTEGER, -- seconds
    time_to_resolve INTEGER, -- seconds
    affected_users INTEGER DEFAULT 0,
    notification_count INTEGER DEFAULT 0,
    update_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_incident_metrics_incident_id ON incident_metrics(incident_id);
CREATE INDEX idx_incident_metrics_tenant_id ON incident_metrics(tenant_id);
CREATE INDEX idx_incident_metrics_created_at ON incident_metrics(created_at DESC);

-- =====================================================================
-- Triggers
-- =====================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_incidents_updated_at
    BEFORE UPDATE ON incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_incident_updates_updated_at
    BEFORE UPDATE ON incident_updates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_incident_components_updated_at
    BEFORE UPDATE ON incident_components
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_incident_templates_updated_at
    BEFORE UPDATE ON incident_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
