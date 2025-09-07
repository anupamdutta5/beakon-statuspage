-- Database initialization script for microservices
-- This script creates separate databases for each microservice

-- Create databases for each microservice
CREATE DATABASE IF NOT EXISTS statuspage_users;
CREATE DATABASE IF NOT EXISTS statuspage_tenants;
CREATE DATABASE IF NOT EXISTS statuspage_components;
CREATE DATABASE IF NOT EXISTS statuspage_incidents;
CREATE DATABASE IF NOT EXISTS statuspage_notifications;
CREATE DATABASE IF NOT EXISTS statuspage_payments;
CREATE DATABASE IF NOT EXISTS statuspage_analytics;
CREATE DATABASE IF NOT EXISTS statuspage_monitoring;

-- Create users for each database
CREATE USER IF NOT EXISTS 'user_service'@'%' IDENTIFIED BY 'user_service_password';
CREATE USER IF NOT EXISTS 'tenant_service'@'%' IDENTIFIED BY 'tenant_service_password';
CREATE USER IF NOT EXISTS 'component_service'@'%' IDENTIFIED BY 'component_service_password';
CREATE USER IF NOT EXISTS 'incident_service'@'%' IDENTIFIED BY 'incident_service_password';
CREATE USER IF NOT EXISTS 'notification_service'@'%' IDENTIFIED BY 'notification_service_password';
CREATE USER IF NOT EXISTS 'payment_service'@'%' IDENTIFIED BY 'payment_service_password';
CREATE USER IF NOT EXISTS 'analytics_service'@'%' IDENTIFIED BY 'analytics_service_password';
CREATE USER IF NOT EXISTS 'monitoring_service'@'%' IDENTIFIED BY 'monitoring_service_password';

-- Grant permissions
GRANT ALL PRIVILEGES ON statuspage_users.* TO 'user_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_tenants.* TO 'tenant_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_components.* TO 'component_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_incidents.* TO 'incident_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_notifications.* TO 'notification_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_payments.* TO 'payment_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_analytics.* TO 'analytics_service'@'%';
GRANT ALL PRIVILEGES ON statuspage_monitoring.* TO 'monitoring_service'@'%';

-- Flush privileges
FLUSH PRIVILEGES;

-- Create event store database for event sourcing
CREATE DATABASE IF NOT EXISTS statuspage_events;
CREATE USER IF NOT EXISTS 'event_service'@'%' IDENTIFIED BY 'event_service_password';
GRANT ALL PRIVILEGES ON statuspage_events.* TO 'event_service'@'%';
FLUSH PRIVILEGES;

-- Create event store tables
USE statuspage_events;

CREATE TABLE IF NOT EXISTS events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSON NOT NULL,
    event_metadata JSON,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_aggregate (aggregate_id, aggregate_type),
    INDEX idx_event_type (event_type),
    INDEX idx_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    snapshot_data JSON NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_aggregate (aggregate_id, aggregate_type),
    INDEX idx_created_at (created_at)
);

-- Create event projections table
CREATE TABLE IF NOT EXISTS event_projections (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    projection_name VARCHAR(100) NOT NULL,
    last_processed_event_id BIGINT NOT NULL,
    last_processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_projection (projection_name),
    INDEX idx_last_processed (last_processed_event_id)
);

-- Create event subscriptions table
CREATE TABLE IF NOT EXISTS event_subscriptions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    subscription_name VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    handler_url VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    last_processed_event_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_subscription (subscription_name, event_type),
    INDEX idx_event_type (event_type),
    INDEX idx_is_active (is_active)
);

-- Create event processing status table
CREATE TABLE IF NOT EXISTS event_processing_status (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id BIGINT NOT NULL,
    subscription_name VARCHAR(100) NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    error_message TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_event_subscription (event_id, subscription_name),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Create event store indexes for performance
CREATE INDEX idx_events_aggregate_version ON events(aggregate_id, aggregate_type, version);
CREATE INDEX idx_events_type_created ON events(event_type, created_at);
CREATE INDEX idx_snapshots_aggregate ON snapshots(aggregate_id, aggregate_type);

-- Create event store views for common queries
CREATE VIEW IF NOT EXISTS event_summary AS
SELECT 
    aggregate_type,
    event_type,
    COUNT(*) as event_count,
    MIN(created_at) as first_event,
    MAX(created_at) as last_event
FROM events
GROUP BY aggregate_type, event_type;

CREATE VIEW IF NOT EXISTS projection_status AS
SELECT 
    ep.projection_name,
    ep.last_processed_event_id,
    ep.last_processed_at,
    (SELECT COUNT(*) FROM events WHERE id > ep.last_processed_event_id) as pending_events
FROM event_projections ep;

-- Create event store initialization complete
SELECT 'Event store initialization completed successfully' as status;