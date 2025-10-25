-- Migration: Add Telegram Bot Integration Support
-- Description: Enables Telegram bot notifications for monitor status changes
-- Author: Claude Code
-- Date: 2025-10-25

-- =====================================================
-- Table: telegram_integrations
-- Purpose: Stores Telegram bot configuration per tenant
-- =====================================================
CREATE TABLE IF NOT EXISTS telegram_integrations (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,

    -- Bot configuration
    bot_token TEXT NOT NULL,  -- Telegram Bot API token (from @BotFather)
    bot_username VARCHAR(255), -- Bot username (e.g., @MyStatusBot)
    bot_name VARCHAR(255),     -- Display name for the bot

    -- Default chat configuration
    default_chat_id VARCHAR(255), -- Default Telegram chat ID to send notifications
    default_chat_name VARCHAR(255), -- Chat name/title (for display purposes)
    default_chat_type VARCHAR(50), -- 'private', 'group', 'supergroup', 'channel'

    -- Status
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,

    -- Notification preferences
    notify_on_down BOOLEAN DEFAULT true,
    notify_on_up BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT true,
    notify_on_maintenance BOOLEAN DEFAULT false,

    -- Message formatting
    use_markdown BOOLEAN DEFAULT true,       -- Enable Telegram MarkdownV2 formatting
    include_monitor_url BOOLEAN DEFAULT true, -- Include clickable monitor URLs
    include_timestamp BOOLEAN DEFAULT true,   -- Include timestamp in messages
    silent_notifications BOOLEAN DEFAULT false, -- Send notifications silently (no sound)
    disable_preview BOOLEAN DEFAULT false,    -- Disable link previews

    -- Retry configuration
    retry_count INT DEFAULT 3,
    retry_interval_seconds INT DEFAULT 5,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP -- Soft delete support
);

-- Indexes for telegram_integrations
CREATE INDEX IF NOT EXISTS idx_telegram_integrations_tenant_id ON telegram_integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_telegram_integrations_is_active ON telegram_integrations(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_telegram_integrations_deleted_at ON telegram_integrations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_telegram_integrations_bot_username ON telegram_integrations(bot_username) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE telegram_integrations IS 'Telegram bot integration configuration for monitor notifications';
COMMENT ON COLUMN telegram_integrations.bot_token IS 'Telegram Bot API token from @BotFather (format: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11)';
COMMENT ON COLUMN telegram_integrations.default_chat_id IS 'Telegram chat ID (positive for users, negative for groups/channels)';
COMMENT ON COLUMN telegram_integrations.use_markdown IS 'Whether to use Telegram MarkdownV2 formatting for rich messages';
COMMENT ON COLUMN telegram_integrations.silent_notifications IS 'Send notifications without sound/vibration';

-- =====================================================
-- Table: telegram_chat_subscriptions
-- Purpose: Maps specific monitors to specific Telegram chats
-- Allows granular routing (e.g., critical monitors to ops chat)
-- =====================================================
CREATE TABLE IF NOT EXISTS telegram_chat_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL,
    monitor_id BIGINT, -- NULL means all monitors use this chat

    -- Chat override (if different from default)
    chat_id VARCHAR(255), -- Telegram chat ID
    chat_name VARCHAR(255), -- Chat name/title
    chat_type VARCHAR(50), -- 'private', 'group', 'supergroup', 'channel'

    -- Event filtering per subscription
    notify_on_down BOOLEAN DEFAULT true,
    notify_on_up BOOLEAN DEFAULT true,
    notify_on_degraded BOOLEAN DEFAULT true,
    notify_on_maintenance BOOLEAN DEFAULT false,

    -- Message customization per subscription
    message_thread_id INT, -- Telegram forum topic/thread ID
    custom_message_prefix TEXT, -- Custom prefix for messages in this chat

    -- Status
    is_active BOOLEAN DEFAULT true,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Foreign keys
    FOREIGN KEY (integration_id) REFERENCES telegram_integrations(id) ON DELETE CASCADE,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,

    -- Ensure unique subscription per monitor per chat
    UNIQUE(integration_id, monitor_id, chat_id)
);

-- Indexes for telegram_chat_subscriptions
CREATE INDEX IF NOT EXISTS idx_telegram_chat_subscriptions_integration_id ON telegram_chat_subscriptions(integration_id);
CREATE INDEX IF NOT EXISTS idx_telegram_chat_subscriptions_monitor_id ON telegram_chat_subscriptions(monitor_id);
CREATE INDEX IF NOT EXISTS idx_telegram_chat_subscriptions_is_active ON telegram_chat_subscriptions(is_active);
CREATE INDEX IF NOT EXISTS idx_telegram_chat_subscriptions_chat_id ON telegram_chat_subscriptions(chat_id);

-- Comments
COMMENT ON TABLE telegram_chat_subscriptions IS 'Maps monitors to specific Telegram chats for targeted notifications';
COMMENT ON COLUMN telegram_chat_subscriptions.message_thread_id IS 'Telegram forum topic ID for group chats with topics enabled';
COMMENT ON COLUMN telegram_chat_subscriptions.custom_message_prefix IS 'Custom prefix prepended to messages (e.g., "[PRODUCTION]")';

-- =====================================================
-- Table: telegram_notifications
-- Purpose: Audit log of all Telegram notification deliveries
-- =====================================================
CREATE TABLE IF NOT EXISTS telegram_notifications (
    id BIGSERIAL PRIMARY KEY,
    integration_id BIGINT NOT NULL,
    monitor_id BIGINT,

    -- Event details
    event_type VARCHAR(50) NOT NULL, -- 'down', 'up', 'degraded', 'maintenance'
    monitor_name VARCHAR(255),
    monitor_url TEXT,

    -- Delivery status
    status VARCHAR(50) NOT NULL, -- 'sent', 'failed', 'pending', 'retrying'
    http_status_code INT,

    -- Message content
    message_text TEXT, -- Full message text sent
    parse_mode VARCHAR(50), -- 'MarkdownV2', 'HTML', 'Markdown', NULL

    -- Delivery tracking
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,
    retry_count INT DEFAULT 0,

    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(50),

    -- Telegram response
    telegram_message_id BIGINT, -- Telegram's message ID
    telegram_chat_id VARCHAR(255), -- Chat where message was sent
    telegram_chat_type VARCHAR(50), -- 'private', 'group', 'supergroup', 'channel'

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Foreign keys
    FOREIGN KEY (integration_id) REFERENCES telegram_integrations(id) ON DELETE CASCADE,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE SET NULL
);

-- Indexes for telegram_notifications
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_integration_id ON telegram_notifications(integration_id);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_monitor_id ON telegram_notifications(monitor_id);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_status ON telegram_notifications(status);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_event_type ON telegram_notifications(event_type);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_created_at ON telegram_notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_sent_at ON telegram_notifications(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_telegram_notifications_chat_id ON telegram_notifications(telegram_chat_id);

-- Comments
COMMENT ON TABLE telegram_notifications IS 'Complete audit trail of Telegram notification deliveries';
COMMENT ON COLUMN telegram_notifications.telegram_message_id IS 'Telegram message ID returned by sendMessage API';
COMMENT ON COLUMN telegram_notifications.parse_mode IS 'Telegram message parsing mode (MarkdownV2, HTML, or plain text)';
COMMENT ON COLUMN telegram_notifications.status IS 'Delivery status: sent (200 OK), failed (error), pending (not yet sent), retrying (in retry queue)';

-- =====================================================
-- View: telegram_integration_stats
-- Purpose: Real-time statistics per integration
-- =====================================================
CREATE OR REPLACE VIEW telegram_integration_stats AS
SELECT
    ti.id AS integration_id,
    ti.tenant_id,
    ti.bot_username,
    ti.bot_name,
    ti.is_active,

    -- Notification counts
    COUNT(tn.id) AS total_notifications,
    COUNT(CASE WHEN tn.status = 'sent' THEN 1 END) AS successful_notifications,
    COUNT(CASE WHEN tn.status = 'failed' THEN 1 END) AS failed_notifications,
    COUNT(CASE WHEN tn.status = 'pending' THEN 1 END) AS pending_notifications,
    COUNT(CASE WHEN tn.status = 'retrying' THEN 1 END) AS retrying_notifications,

    -- Event type breakdown
    COUNT(CASE WHEN tn.event_type = 'down' THEN 1 END) AS down_notifications,
    COUNT(CASE WHEN tn.event_type = 'up' THEN 1 END) AS up_notifications,
    COUNT(CASE WHEN tn.event_type = 'degraded' THEN 1 END) AS degraded_notifications,
    COUNT(CASE WHEN tn.event_type = 'maintenance' THEN 1 END) AS maintenance_notifications,

    -- Success rate
    CASE
        WHEN COUNT(tn.id) > 0 THEN
            ROUND((COUNT(CASE WHEN tn.status = 'sent' THEN 1 END)::NUMERIC / COUNT(tn.id)::NUMERIC) * 100, 2)
        ELSE 0
    END AS success_rate,

    -- Timestamps
    MAX(tn.sent_at) AS last_notification_at,
    ti.created_at,
    ti.updated_at
FROM telegram_integrations ti
LEFT JOIN telegram_notifications tn ON ti.id = tn.integration_id
WHERE ti.deleted_at IS NULL
GROUP BY ti.id, ti.tenant_id, ti.bot_username, ti.bot_name, ti.is_active, ti.created_at, ti.updated_at;

-- Comments
COMMENT ON VIEW telegram_integration_stats IS 'Real-time statistics for Telegram integrations including success rates and event breakdowns';

-- =====================================================
-- Triggers: Auto-update updated_at timestamps
-- =====================================================

-- telegram_integrations update trigger
CREATE OR REPLACE FUNCTION update_telegram_integrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_telegram_integrations_updated_at ON telegram_integrations;
CREATE TRIGGER trigger_telegram_integrations_updated_at
    BEFORE UPDATE ON telegram_integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_telegram_integrations_updated_at();

-- telegram_chat_subscriptions update trigger
CREATE OR REPLACE FUNCTION update_telegram_chat_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_telegram_chat_subscriptions_updated_at ON telegram_chat_subscriptions;
CREATE TRIGGER trigger_telegram_chat_subscriptions_updated_at
    BEFORE UPDATE ON telegram_chat_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_telegram_chat_subscriptions_updated_at();

-- telegram_notifications update trigger
CREATE OR REPLACE FUNCTION update_telegram_notifications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_telegram_notifications_updated_at ON telegram_notifications;
CREATE TRIGGER trigger_telegram_notifications_updated_at
    BEFORE UPDATE ON telegram_notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_telegram_notifications_updated_at();

-- =====================================================
-- Sample Data (Optional - for development/testing)
-- =====================================================

-- Uncomment to insert sample data for testing
/*
INSERT INTO telegram_integrations (
    tenant_id,
    bot_token,
    bot_username,
    bot_name,
    default_chat_id,
    default_chat_name,
    default_chat_type,
    is_active,
    notify_on_down,
    notify_on_up,
    notify_on_degraded,
    notify_on_maintenance
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    '123456789:ABCdefGHIjklMNOpqrsTUVwxyz',
    'BeakonStatusBot',
    'Beakon Status Alerts',
    '-1001234567890',
    'Status Alerts Channel',
    'channel',
    true,
    true,
    true,
    true,
    false
);
*/

-- =====================================================
-- Migration Complete
-- =====================================================
-- Tables created: 3 (telegram_integrations, telegram_chat_subscriptions, telegram_notifications)
-- Views created: 1 (telegram_integration_stats)
-- Indexes created: 15
-- Triggers created: 3
-- Foreign keys: 4
