-- Initialize all databases for Beakon Status Page platform
-- This script creates all 14 required databases

-- SaaS Admin database
CREATE DATABASE saas_admin;

-- Tenant Admin database (shared by saas-admin and tenant-admin services)
CREATE DATABASE tenant_admin_db;

-- User Service database
CREATE DATABASE statuspage_user;

-- Component Service database
CREATE DATABASE statuspage_component;

-- Incident Service database
CREATE DATABASE statuspage_incident;

-- Payment Service database
CREATE DATABASE statuspage_payment;

-- Landing Page Service database
CREATE DATABASE statuspage_landing;

-- Notification Service database
CREATE DATABASE statuspage_notification;

-- Analytics Service database
CREATE DATABASE statuspage_analytics;

-- Monitoring Service database
CREATE DATABASE statuspage_monitoring;

-- Event Store Service database
CREATE DATABASE statuspage_eventstore;

-- Branding Service database
CREATE DATABASE statuspage_branding;

-- Status UI Service database
CREATE DATABASE statuspage_ui;

-- Audit Service database
CREATE DATABASE statuspage_audit;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE saas_admin TO postgres;
GRANT ALL PRIVILEGES ON DATABASE tenant_admin_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_user TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_component TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_incident TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_payment TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_landing TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_notification TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_analytics TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_monitoring TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_eventstore TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_branding TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_ui TO postgres;
GRANT ALL PRIVILEGES ON DATABASE statuspage_audit TO postgres;
