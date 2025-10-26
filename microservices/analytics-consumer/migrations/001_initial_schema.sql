--
-- Analytics Consumer - Initial Schema
-- Database: analytics
-- Description: Consumer service for processing analytics events from queue
-- Based on: internal/models/analytics.go
--

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Metric data table (processed metrics from analytics events)
CREATE TABLE IF NOT EXISTS metric_data (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    user_id BIGINT,
    session_id VARCHAR(255),
    metadata TEXT
);

CREATE INDEX idx_metric_data_tenant_id ON metric_data(tenant_id);
CREATE INDEX idx_metric_data_metric_name ON metric_data(metric_name);
CREATE INDEX idx_metric_data_timestamp ON metric_data(timestamp DESC);
CREATE INDEX idx_metric_data_user_id ON metric_data(user_id);
CREATE INDEX idx_metric_data_session_id ON metric_data(session_id);
CREATE INDEX idx_metric_data_deleted_at ON metric_data(deleted_at);

-- Aggregated data table (pre-calculated aggregations)
CREATE TABLE IF NOT EXISTS aggregated_data (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    aggregation_type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    count BIGINT NOT NULL,
    period VARCHAR(50) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    metadata TEXT
);

CREATE INDEX idx_aggregated_data_tenant_id ON aggregated_data(tenant_id);
CREATE INDEX idx_aggregated_data_metric_name ON aggregated_data(metric_name);
CREATE INDEX idx_aggregated_data_aggregation_type ON aggregated_data(aggregation_type);
CREATE INDEX idx_aggregated_data_period ON aggregated_data(period);
CREATE INDEX idx_aggregated_data_start_time ON aggregated_data(start_time DESC);
CREATE INDEX idx_aggregated_data_end_time ON aggregated_data(end_time DESC);
CREATE INDEX idx_aggregated_data_deleted_at ON aggregated_data(deleted_at);

-- Page view data table
CREATE TABLE IF NOT EXISTS page_view_data (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    page VARCHAR(500) NOT NULL,
    user_id BIGINT,
    session_id VARCHAR(255),
    user_agent TEXT,
    ip_address VARCHAR(50),
    referrer VARCHAR(1000),
    duration BIGINT,
    timestamp TIMESTAMP NOT NULL,
    metadata TEXT
);

CREATE INDEX idx_page_view_data_tenant_id ON page_view_data(tenant_id);
CREATE INDEX idx_page_view_data_page ON page_view_data(page);
CREATE INDEX idx_page_view_data_user_id ON page_view_data(user_id);
CREATE INDEX idx_page_view_data_session_id ON page_view_data(session_id);
CREATE INDEX idx_page_view_data_timestamp ON page_view_data(timestamp DESC);
CREATE INDEX idx_page_view_data_deleted_at ON page_view_data(deleted_at);

-- User action data table
CREATE TABLE IF NOT EXISTS user_action_data (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    action VARCHAR(255) NOT NULL,
    user_id BIGINT,
    session_id VARCHAR(255),
    page VARCHAR(500),
    element VARCHAR(255),
    value VARCHAR(1000),
    timestamp TIMESTAMP NOT NULL,
    metadata TEXT
);

CREATE INDEX idx_user_action_data_tenant_id ON user_action_data(tenant_id);
CREATE INDEX idx_user_action_data_action ON user_action_data(action);
CREATE INDEX idx_user_action_data_user_id ON user_action_data(user_id);
CREATE INDEX idx_user_action_data_session_id ON user_action_data(session_id);
CREATE INDEX idx_user_action_data_page ON user_action_data(page);
CREATE INDEX idx_user_action_data_timestamp ON user_action_data(timestamp DESC);
CREATE INDEX idx_user_action_data_deleted_at ON user_action_data(deleted_at);

-- Performance data table
CREATE TABLE IF NOT EXISTS performance_data (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    user_id BIGINT,
    session_id VARCHAR(255),
    page VARCHAR(500),
    timestamp TIMESTAMP NOT NULL,
    metadata TEXT
);

CREATE INDEX idx_performance_data_tenant_id ON performance_data(tenant_id);
CREATE INDEX idx_performance_data_metric_name ON performance_data(metric_name);
CREATE INDEX idx_performance_data_user_id ON performance_data(user_id);
CREATE INDEX idx_performance_data_session_id ON performance_data(session_id);
CREATE INDEX idx_performance_data_page ON performance_data(page);
CREATE INDEX idx_performance_data_timestamp ON performance_data(timestamp DESC);
CREATE INDEX idx_performance_data_deleted_at ON performance_data(deleted_at);

-- Error data table
CREATE TABLE IF NOT EXISTS error_data (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    tenant_id BIGINT NOT NULL,
    error_type VARCHAR(255) NOT NULL,
    error_message TEXT,
    user_id BIGINT,
    session_id VARCHAR(255),
    page VARCHAR(500),
    stack TEXT,
    timestamp TIMESTAMP NOT NULL,
    metadata TEXT
);

CREATE INDEX idx_error_data_tenant_id ON error_data(tenant_id);
CREATE INDEX idx_error_data_error_type ON error_data(error_type);
CREATE INDEX idx_error_data_user_id ON error_data(user_id);
CREATE INDEX idx_error_data_session_id ON error_data(session_id);
CREATE INDEX idx_error_data_page ON error_data(page);
CREATE INDEX idx_error_data_timestamp ON error_data(timestamp DESC);
CREATE INDEX idx_error_data_deleted_at ON error_data(deleted_at);

-- Processing logs table (tracks message processing)
CREATE TABLE IF NOT EXISTS processing_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    message_id VARCHAR(255) NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    processing_time BIGINT,
    error TEXT,
    retry_count INTEGER DEFAULT 0,
    metadata TEXT
);

CREATE INDEX idx_processing_logs_message_id ON processing_logs(message_id);
CREATE INDEX idx_processing_logs_event_id ON processing_logs(event_id);
CREATE INDEX idx_processing_logs_event_type ON processing_logs(event_type);
CREATE INDEX idx_processing_logs_status ON processing_logs(status);
CREATE INDEX idx_processing_logs_deleted_at ON processing_logs(deleted_at);

-- Queue messages table (persistent queue for reliability)
CREATE TABLE IF NOT EXISTS queue_messages (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    queue_name VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    data TEXT NOT NULL,
    priority INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    processed_at TIMESTAMP,
    failed_at TIMESTAMP,
    error TEXT,
    metadata TEXT
);

CREATE INDEX idx_queue_messages_queue_name ON queue_messages(queue_name);
CREATE INDEX idx_queue_messages_event_type ON queue_messages(event_type);
CREATE INDEX idx_queue_messages_status ON queue_messages(status);
CREATE INDEX idx_queue_messages_priority ON queue_messages(priority DESC);
CREATE INDEX idx_queue_messages_deleted_at ON queue_messages(deleted_at);
CREATE INDEX idx_queue_messages_pending ON queue_messages(status, priority DESC) WHERE status = 'pending';

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for all tables
CREATE TRIGGER update_metric_data_updated_at BEFORE UPDATE ON metric_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_aggregated_data_updated_at BEFORE UPDATE ON aggregated_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_page_view_data_updated_at BEFORE UPDATE ON page_view_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_action_data_updated_at BEFORE UPDATE ON user_action_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_performance_data_updated_at BEFORE UPDATE ON performance_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_error_data_updated_at BEFORE UPDATE ON error_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processing_logs_updated_at BEFORE UPDATE ON processing_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_queue_messages_updated_at BEFORE UPDATE ON queue_messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
