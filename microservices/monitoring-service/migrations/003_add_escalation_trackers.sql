-- Migration: Add escalation_trackers table for tracking incident escalations
-- Created: 2025-10-21
-- Purpose: Track escalation state for each incident

-- Create escalation_trackers table
CREATE TABLE IF NOT EXISTS escalation_trackers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    policy_id BIGINT NOT NULL,
    current_level INT DEFAULT 1,
    is_resolved BOOLEAN DEFAULT false,
    last_escalated TIMESTAMP,
    notified_users TEXT,  -- JSON array: ["uuid1", "uuid2", ...]

    CONSTRAINT fk_escalation_policy FOREIGN KEY (policy_id)
        REFERENCES escalation_policies(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_escalation_tracker_tenant ON escalation_trackers(tenant_id);
CREATE INDEX idx_escalation_tracker_incident ON escalation_trackers(incident_id);
CREATE INDEX idx_escalation_tracker_policy ON escalation_trackers(policy_id);
CREATE INDEX idx_escalation_tracker_resolved ON escalation_trackers(is_resolved) WHERE is_resolved = false;

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_escalation_trackers_updated_at
    BEFORE UPDATE ON escalation_trackers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE escalation_trackers IS 'Tracks escalation state for each incident';
COMMENT ON COLUMN escalation_trackers.incident_id IS 'UUID of the incident being escalated';
COMMENT ON COLUMN escalation_trackers.policy_id IS 'Foreign key to escalation_policies table';
COMMENT ON COLUMN escalation_trackers.current_level IS 'Current escalation level (1, 2, 3, ...)';
COMMENT ON COLUMN escalation_trackers.is_resolved IS 'Whether the escalation has been resolved';
COMMENT ON COLUMN escalation_trackers.last_escalated IS 'Timestamp of last level escalation';
COMMENT ON COLUMN escalation_trackers.notified_users IS 'JSON array of user IDs who have been notified';
