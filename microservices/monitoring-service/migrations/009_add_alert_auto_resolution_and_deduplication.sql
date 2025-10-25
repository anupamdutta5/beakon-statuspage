-- Migration: Add Alert Auto-Resolution and Deduplication
-- P1 Features: Alert Auto-Resolution + Alert Deduplication
-- Date: 2025-10-25

-- Add auto-resolution fields to alerts table
ALTER TABLE alerts
ADD COLUMN IF NOT EXISTS resolution_type VARCHAR(20),  -- 'auto' or 'manual'
ADD COLUMN IF NOT EXISTS resolution_note TEXT;

-- Add deduplication fields to alerts table
ALTER TABLE alerts
ADD COLUMN IF NOT EXISTS dedup_key VARCHAR(255),
ADD COLUMN IF NOT EXISTS error_type VARCHAR(50);

-- Create index for deduplication queries (find recent duplicates)
CREATE INDEX IF NOT EXISTS idx_alerts_dedup
ON alerts(dedup_key, created_at DESC)
WHERE status IN ('active', 'acknowledged');

-- Create index for auto-resolution queries (find alerts to auto-resolve)
CREATE INDEX IF NOT EXISTS idx_alerts_auto_resolve
ON alerts(service_id, status, updated_at DESC)
WHERE status IN ('active', 'acknowledged');

-- Add comments for documentation
COMMENT ON COLUMN alerts.resolution_type IS 'How the alert was resolved: auto (automatic) or manual (by user)';
COMMENT ON COLUMN alerts.resolution_note IS 'Additional notes about the resolution';
COMMENT ON COLUMN alerts.dedup_key IS 'Deduplication key: monitor-{id}-{error_type} to prevent duplicate alerts';
COMMENT ON COLUMN alerts.error_type IS 'Type of error that triggered the alert (timeout, 5xx, ssl_error, etc)';
