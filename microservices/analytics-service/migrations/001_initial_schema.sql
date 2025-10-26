--
-- Analytics Service - Initial Schema
-- Database: analytics
-- Description: Page views, visitor sessions, uptime metrics, and analytics data
--

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS page_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    page_url VARCHAR(1000) NOT NULL,
    referrer VARCHAR(1000),
    user_agent TEXT,
    ip_address VARCHAR(45),
    visitor_id UUID,
    session_id UUID,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_views_tenant_id ON page_views(tenant_id);
CREATE INDEX idx_page_views_viewed_at ON page_views(viewed_at DESC);
CREATE INDEX idx_page_views_session_id ON page_views(session_id);

CREATE TABLE IF NOT EXISTS visitor_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    visitor_id UUID NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    country VARCHAR(100),
    city VARCHAR(100),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    page_count INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_visitor_sessions_tenant_id ON visitor_sessions(tenant_id);
CREATE INDEX idx_visitor_sessions_started_at ON visitor_sessions(started_at DESC);

CREATE TABLE IF NOT EXISTS uptime_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    component_id UUID,
    date DATE NOT NULL,
    uptime_percentage DECIMAL(5,2) DEFAULT 100.00,
    downtime_minutes INTEGER DEFAULT 0,
    incident_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_uptime_metrics_tenant_id ON uptime_metrics(tenant_id);
CREATE INDEX idx_uptime_metrics_date ON uptime_metrics(date DESC);
CREATE UNIQUE INDEX idx_uptime_metrics_tenant_component_date ON uptime_metrics(tenant_id, component_id, date);

CREATE TABLE IF NOT EXISTS response_times (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    component_id UUID,
    response_time_ms INTEGER NOT NULL,
    status_code INTEGER,
    measured_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_response_times_tenant_id ON response_times(tenant_id);
CREATE INDEX idx_response_times_component_id ON response_times(component_id);
CREATE INDEX idx_response_times_measured_at ON response_times(measured_at DESC);

COMMIT;
