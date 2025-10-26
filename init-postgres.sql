-- Beakon Platform - PostgreSQL Database Initialization
-- This script creates all required databases for the microservices
-- Executed automatically when PostgreSQL container starts for the first time

-- Create all required databases
CREATE DATABASE saas_admin;
CREATE DATABASE tenant_admin_db;
CREATE DATABASE users;
CREATE DATABASE components;
CREATE DATABASE notifications;
CREATE DATABASE incidents;
CREATE DATABASE payments;
CREATE DATABASE analytics;
CREATE DATABASE monitoring;
CREATE DATABASE events;
CREATE DATABASE branding;
CREATE DATABASE landing_page;

-- Grant all privileges to postgres user (already has them, but being explicit)
GRANT ALL PRIVILEGES ON DATABASE saas_admin TO postgres;
GRANT ALL PRIVILEGES ON DATABASE tenant_admin_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE users TO postgres;
GRANT ALL PRIVILEGES ON DATABASE components TO postgres;
GRANT ALL PRIVILEGES ON DATABASE notifications TO postgres;
GRANT ALL PRIVILEGES ON DATABASE incidents TO postgres;
GRANT ALL PRIVILEGES ON DATABASE payments TO postgres;
GRANT ALL PRIVILEGES ON DATABASE analytics TO postgres;
GRANT ALL PRIVILEGES ON DATABASE monitoring TO postgres;
GRANT ALL PRIVILEGES ON DATABASE events TO postgres;
GRANT ALL PRIVILEGES ON DATABASE branding TO postgres;
GRANT ALL PRIVILEGES ON DATABASE landing_page TO postgres;
