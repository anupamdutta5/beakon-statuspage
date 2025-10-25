-- Migration: Add Discord Integration Support
-- This migration adds Discord webhook integration similar to Slack/Teams
-- Discord uses webhook URLs for incoming messages with rich embed support

BEGIN;

-- Discord integration configuration table
CREATE TABLE IF NOT EXISTS discord_integrations (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,

    -- Discord webhook configuration
    webhook_url TEXT NOT NULL,              -- Discord webhook URL
    webhook_name VARCHAR(255),               -- Custom webhook name
    avatar_url TEXT,                         -- Custom avatar for webhook messages

    -- Channel configuration
    default_channel_id VARCHAR(255),         -- Discord channel ID (snowflake)
    default_channel_name VARCHAR(255),       -- Channel name for display

    -- Integration status
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,

    -- Notification preferences (same as Slack/Teams pattern)
    notify_on_down BOOLEAN DEFAULT true,
    notify_on_up BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT true,
    notify_on_maintenance BOOLEAN DEFAULT false,

    -- User mentions configuration
    mention_users TEXT[],                    -- Array of Discord user IDs to mention
    mention_roles TEXT[],                    -- Array of Discord role IDs to mention
    mention_everyone BOOLEAN DEFAULT false,  -- Whether to @everyone

    -- Embed customization
    custom_color VARCHAR(7),                 -- Hex color for embeds (e.g., #5865F2)
    include_monitor_url BOOLEAN DEFAULT true,
    include_timestamp BOOLEAN DEFAULT true,

    -- Rate limiting and retries
    retry_count INT DEFAULT 3,
    retry_interval_seconds INT DEFAULT 5,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP                     -- Soft delete support
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_discord_integrations_tenant_id
    ON discord_integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_discord_integrations_is_active
    ON discord_integrations(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_discord_integrations_deleted_at
    ON discord_integrations(deleted_at);

-- Channel-specific subscriptions (monitor-to-channel mapping)
CREATE TABLE IF NOT EXISTS discord_channel_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL,
    monitor_id BIGINT,                       -- NULL means all monitors

    -- Channel override
    channel_id VARCHAR(255),                 -- Override default channel for this monitor
    channel_name VARCHAR(255),

    -- Event filtering per subscription
    notify_on_down BOOLEAN DEFAULT true,
    notify_on_up BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT true,
    notify_on_maintenance BOOLEAN DEFAULT false,

    -- Subscription status
    is_active BOOLEAN DEFAULT true,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Constraints
    CONSTRAINT discord_channel_subscriptions_integration_id_fkey
        FOREIGN KEY (integration_id)
        REFERENCES discord_integrations(id) ON DELETE CASCADE,
    CONSTRAINT discord_channel_subscriptions_monitor_id_fkey
        FOREIGN KEY (monitor_id)
        REFERENCES monitors(id) ON DELETE CASCADE,

    -- Unique constraint: one subscription per monitor-channel combination
    CONSTRAINT discord_channel_subscriptions_unique
        UNIQUE(integration_id, monitor_id, channel_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_discord_channel_subscriptions_integration_id
    ON discord_channel_subscriptions(integration_id);
CREATE INDEX IF NOT EXISTS idx_discord_channel_subscriptions_monitor_id
    ON discord_channel_subscriptions(monitor_id);
CREATE INDEX IF NOT EXISTS idx_discord_channel_subscriptions_is_active
    ON discord_channel_subscriptions(is_active);

-- Notification delivery tracking (audit log)
CREATE TABLE IF NOT EXISTS discord_notifications (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL,
    monitor_id BIGINT,

    -- Event details
    event_type VARCHAR(50) NOT NULL,         -- down, up, degraded, maintenance
    monitor_name VARCHAR(255),
    monitor_url TEXT,

    -- Delivery status
    status VARCHAR(50) NOT NULL,             -- sent, failed, pending, retrying
    http_status_code INT,

    -- Message content
    message_content TEXT,                    -- The actual message sent
    embed_data JSONB,                        -- Discord embed object

    -- Delivery tracking
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,
    retry_count INT DEFAULT 0,

    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(50),

    -- Response from Discord
    discord_message_id VARCHAR(255),         -- Discord message ID (snowflake)
    discord_channel_id VARCHAR(255),

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Constraints
    CONSTRAINT discord_notifications_integration_id_fkey
        FOREIGN KEY (integration_id)
        REFERENCES discord_integrations(id) ON DELETE CASCADE,
    CONSTRAINT discord_notifications_monitor_id_fkey
        FOREIGN KEY (monitor_id)
        REFERENCES monitors(id) ON DELETE SET NULL
);

-- Indexes for querying notification history
CREATE INDEX IF NOT EXISTS idx_discord_notifications_integration_id
    ON discord_notifications(integration_id);
CREATE INDEX IF NOT EXISTS idx_discord_notifications_monitor_id
    ON discord_notifications(monitor_id);
CREATE INDEX IF NOT EXISTS idx_discord_notifications_status
    ON discord_notifications(status);
CREATE INDEX IF NOT EXISTS idx_discord_notifications_event_type
    ON discord_notifications(event_type);
CREATE INDEX IF NOT EXISTS idx_discord_notifications_created_at
    ON discord_notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_discord_notifications_sent_at
    ON discord_notifications(sent_at DESC);

-- Integration statistics view (for analytics)
CREATE OR REPLACE VIEW discord_integration_stats AS
SELECT
    di.id AS integration_id,
    di.tenant_id,
    di.webhook_name,
    di.is_active,
    COUNT(dn.id) AS total_notifications,
    COUNT(CASE WHEN dn.status = 'sent' THEN 1 END) AS successful_notifications,
    COUNT(CASE WHEN dn.status = 'failed' THEN 1 END) AS failed_notifications,
    COUNT(CASE WHEN dn.event_type = 'down' THEN 1 END) AS down_notifications,
    COUNT(CASE WHEN dn.event_type = 'up' THEN 1 END) AS up_notifications,
    MAX(dn.sent_at) AS last_notification_at,
    di.created_at,
    di.updated_at
FROM discord_integrations di
LEFT JOIN discord_notifications dn ON di.id = dn.integration_id
WHERE di.deleted_at IS NULL
GROUP BY di.id, di.tenant_id, di.webhook_name, di.is_active, di.created_at, di.updated_at;

-- Add comments for documentation
COMMENT ON TABLE discord_integrations IS 'Discord webhook integrations for monitor alerts';
COMMENT ON TABLE discord_channel_subscriptions IS 'Monitor-to-Discord-channel subscription mappings';
COMMENT ON TABLE discord_notifications IS 'Audit log of Discord notification deliveries';
COMMENT ON COLUMN discord_integrations.webhook_url IS 'Discord webhook URL from channel integrations settings';
COMMENT ON COLUMN discord_integrations.mention_users IS 'Array of Discord user IDs (snowflakes) to mention in notifications';
COMMENT ON COLUMN discord_integrations.custom_color IS 'Hex color code for Discord embeds (e.g., #5865F2 for Discord blurple)';
COMMENT ON COLUMN discord_notifications.embed_data IS 'Complete Discord embed object as JSONB for audit purposes';

COMMIT;
