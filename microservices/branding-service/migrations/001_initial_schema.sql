--
-- Branding Service - Initial Schema
-- Database: branding
-- Description: Tenant branding, logos, themes, and customization
--

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tenant_branding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE,
    logo_url VARCHAR(1000),
    favicon_url VARCHAR(1000),
    primary_color VARCHAR(7) DEFAULT '#0066CC',
    secondary_color VARCHAR(7) DEFAULT '#333333',
    accent_color VARCHAR(7) DEFAULT '#00AA00',
    background_color VARCHAR(7) DEFAULT '#FFFFFF',
    text_color VARCHAR(7) DEFAULT '#333333',
    font_family VARCHAR(100) DEFAULT 'Inter',
    custom_css TEXT,
    custom_js TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenant_branding_tenant_id ON tenant_branding(tenant_id);

CREATE TABLE IF NOT EXISTS email_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    template_type VARCHAR(100) NOT NULL,
    subject VARCHAR(500),
    body_html TEXT NOT NULL,
    body_text TEXT,
    from_name VARCHAR(255),
    from_email VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_templates_tenant_id ON email_templates(tenant_id);
CREATE UNIQUE INDEX idx_email_templates_tenant_type ON email_templates(tenant_id, template_type);

CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    domain VARCHAR(255) NOT NULL UNIQUE,
    is_verified BOOLEAN DEFAULT false,
    verification_token VARCHAR(255),
    verified_at TIMESTAMP,
    ssl_enabled BOOLEAN DEFAULT false,
    ssl_certificate TEXT,
    ssl_private_key TEXT,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_custom_domains_tenant_id ON custom_domains(tenant_id);
CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);

COMMIT;
