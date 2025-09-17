-- Status Pages Schema Migration
-- This migration adds support for multi-tenant status pages with custom domains

-- Status page table
CREATE TABLE IF NOT EXISTS status_pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    timezone VARCHAR(50) DEFAULT 'UTC',
    custom_css TEXT,
    custom_js TEXT,
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords TEXT[],
    google_analytics_id VARCHAR(50),
    is_public BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, slug)
);

-- Status page domains table for custom domains
CREATE TABLE IF NOT EXISTS status_page_domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status_page_id UUID NOT NULL REFERENCES status_pages(id) ON DELETE CASCADE,
    domain_name VARCHAR(255) NOT NULL,
    ssl_certificate TEXT,
    ssl_private_key TEXT,
    ssl_issuer TEXT,
    ssl_expires_at TIMESTAMP WITH TIME ZONE,
    is_verified BOOLEAN DEFAULT FALSE,
    verification_code VARCHAR(100),
    is_primary BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(domain_name)
);

-- Status page settings
CREATE TABLE IF NOT EXISTS status_page_settings (
    status_page_id UUID PRIMARY KEY REFERENCES status_pages(id) ON DELETE CASCADE,
    logo_url TEXT,
    favicon_url TEXT,
    theme_color VARCHAR(7) DEFAULT '#2563eb',
    background_color VARCHAR(7) DEFAULT '#ffffff',
    text_color VARCHAR(7) DEFAULT '#1f2937',
    show_incidents BOOLEAN DEFAULT TRUE,
    show_metrics BOOLEAN DEFAULT TRUE,
    show_metrics_history BOOLEAN DEFAULT TRUE,
    show_uptime BOOLEAN DEFAULT TRUE,
    show_response_time BOOLEAN DEFAULT TRUE,
    show_incident_history BOOLEAN DEFAULT TRUE,
    show_subscribe BOOLEAN DEFAULT TRUE,
    show_powered_by BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Monitor types table
CREATE TYPE monitor_type AS ENUM (
    'http', 'https', 'tcp', 'udp', 'ping', 'dns', 'docker', 'kubernetes', 'api', 'custom'
);

-- Monitors table
CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type monitor_type NOT NULL,
    target TEXT NOT NULL,
    interval_seconds INTEGER DEFAULT 60,
    timeout_seconds INTEGER DEFAULT 30,
    retries INTEGER DEFAULT 1,
    method VARCHAR(10) DEFAULT 'GET',
    expected_status_codes INTEGER[],
    expected_body TEXT,
    headers JSONB,
    body TEXT,
    auth_username VARCHAR(255),
    auth_password TEXT,
    verify_ssl BOOLEAN DEFAULT TRUE,
    follow_redirects BOOLEAN DEFAULT TRUE,
    allow_insecure BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Status page monitors (many-to-many relationship)
CREATE TABLE IF NOT EXISTS status_page_monitors (
    status_page_id UUID NOT NULL REFERENCES status_pages(id) ON DELETE CASCADE,
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    display_name VARCHAR(255),
    display_order INTEGER DEFAULT 0,
    show_on_status_page BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (status_page_id, monitor_id)
);

-- Monitor status history
CREATE TABLE IF NOT EXISTS monitor_status_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    monitor_id UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    response_body TEXT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX idx_status_pages_tenant_id ON status_pages(tenant_id);
CREATE INDEX idx_status_page_domains_status_page_id ON status_page_domains(status_page_id);
CREATE INDEX idx_monitors_tenant_id ON monitors(tenant_id);
CREATE INDEX idx_monitor_status_history_monitor_id ON monitor_status_history(monitor_id);
CREATE INDEX idx_monitor_status_history_created_at ON monitor_status_history(created_at);

-- Add status page limits to saas_plans table
ALTER TABLE saas_plans 
ADD COLUMN IF NOT EXISTS max_status_pages INTEGER DEFAULT 3,
ADD COLUMN IF NOT EXISTS max_custom_domains INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS max_monitors_per_page INTEGER DEFAULT 20;

-- Add status page feature flags to saas_plans table
ALTER TABLE saas_plans
ADD COLUMN IF NOT EXISTS custom_status_pages BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS custom_domains BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS ssl_certificates BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS advanced_analytics BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS team_collaboration BOOLEAN DEFAULT FALSE;

-- Create a function to update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic updated_at updates
CREATE TRIGGER update_status_pages_updated_at
BEFORE UPDATE ON status_pages
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_status_page_domains_updated_at
BEFORE UPDATE ON status_page_domains
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_status_page_settings_updated_at
BEFORE UPDATE ON status_page_settings
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_monitors_updated_at
BEFORE UPDATE ON monitors
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create a view for status page public data
CREATE OR REPLACE VIEW public_status_page_data AS
SELECT 
    sp.id,
    sp.name,
    sp.slug,
    sp.description,
    sp.timezone,
    sp.custom_css,
    sp.custom_js,
    sp.meta_title,
    sp.meta_description,
    sp.meta_keywords,
    sp.google_analytics_id,
    sps.logo_url,
    sps.favicon_url,
    sps.theme_color,
    sps.background_color,
    sps.text_color,
    sps.show_incidents,
    sps.show_metrics,
    sps.show_metrics_history,
    sps.show_uptime,
    sps.show_response_time,
    sps.show_incident_history,
    sps.show_subscribe,
    sps.show_powered_by,
    spd.domain_name as primary_domain
FROM 
    status_pages sp
LEFT JOIN 
    status_page_settings sps ON sp.id = sps.status_page_id
LEFT JOIN 
    status_page_domains spd ON sp.id = spd.status_page_id AND spd.is_primary = TRUE
WHERE 
    sp.is_public = TRUE 
    AND sp.deleted_at IS NULL;
