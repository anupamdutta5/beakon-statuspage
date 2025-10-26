--
-- Notification Service - Initial Schema
-- Database: notifications
-- Description: Multi-channel notification system with templates and delivery tracking
--

BEGIN;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================================
-- Notification Templates
-- =====================================================================

CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    channel VARCHAR(50) NOT NULL, -- 'email', 'sms', 'webhook', 'slack', 'push'
    subject VARCHAR(500),
    body_template TEXT NOT NULL,
    variables JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_notification_templates_tenant_id ON notification_templates(tenant_id);
CREATE INDEX idx_notification_templates_channel ON notification_templates(channel);
CREATE INDEX idx_notification_templates_deleted_at ON notification_templates(deleted_at);
CREATE UNIQUE INDEX idx_notification_templates_tenant_name ON notification_templates(tenant_id, name) WHERE deleted_at IS NULL;

-- =====================================================================
-- Notification Channels (Configuration)
-- =====================================================================

CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    channel_type VARCHAR(50) NOT NULL, -- 'email', 'sms', 'webhook', 'slack', 'push'
    name VARCHAR(255) NOT NULL,
    configuration JSONB NOT NULL, -- channel-specific config (SMTP, Twilio, etc.)
    is_enabled BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_channels_tenant_id ON notification_channels(tenant_id);
CREATE INDEX idx_notification_channels_type ON notification_channels(channel_type);
CREATE INDEX idx_notification_channels_enabled ON notification_channels(is_enabled);

-- =====================================================================
-- Notifications (Outbound messages)
-- =====================================================================

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    template_id UUID,
    channel VARCHAR(50) NOT NULL,
    recipient VARCHAR(500) NOT NULL, -- email, phone, webhook URL, etc.
    subject VARCHAR(500),
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'sent', 'delivered', 'failed', 'bounced'
    error_message TEXT,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_tenant_id ON notifications(tenant_id);
CREATE INDEX idx_notifications_template_id ON notifications(template_id);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_recipient ON notifications(recipient);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- =====================================================================
-- Notification Logs (Delivery tracking)
-- =====================================================================

CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- 'queued', 'sent', 'delivered', 'opened', 'clicked', 'bounced', 'failed'
    event_data JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_logs_notification_id ON notification_logs(notification_id);
CREATE INDEX idx_notification_logs_event_type ON notification_logs(event_type);
CREATE INDEX idx_notification_logs_created_at ON notification_logs(created_at DESC);

-- =====================================================================
-- Notification Preferences (User preferences)
-- =====================================================================

CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    channel VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- 'incident_created', 'incident_updated', 'maintenance_scheduled', etc.
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_preferences_tenant_id ON notification_preferences(tenant_id);
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX idx_notification_preferences_channel ON notification_preferences(channel);
CREATE UNIQUE INDEX idx_notification_preferences_unique ON notification_preferences(tenant_id, user_id, channel, event_type);

-- =====================================================================
-- Notification Queue (For async processing)
-- =====================================================================

CREATE TABLE IF NOT EXISTS notification_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    notification_id UUID,
    payload JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP,
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_queue_tenant_id ON notification_queue(tenant_id);
CREATE INDEX idx_notification_queue_status ON notification_queue(status);
CREATE INDEX idx_notification_queue_next_retry ON notification_queue(next_retry_at) WHERE status = 'pending';
CREATE INDEX idx_notification_queue_created_at ON notification_queue(created_at DESC);

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

CREATE TRIGGER update_notification_templates_updated_at
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_channels_updated_at
    BEFORE UPDATE ON notification_channels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
