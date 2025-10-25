-- Migration: Add Incident Owner Assignment
-- P1 Feature: Incident Owner Assignment
-- Date: 2025-10-25

-- Add owner assignment fields to saas_incidents table
ALTER TABLE saas_incidents
ADD COLUMN IF NOT EXISTS owner_id UUID,
ADD COLUMN IF NOT EXISTS assigned_at TIMESTAMP;

-- Create index for owner queries
CREATE INDEX IF NOT EXISTS idx_saas_incidents_owner_id
ON saas_incidents(owner_id)
WHERE owner_id IS NOT NULL;

-- Create index for assigned incidents queries
CREATE INDEX IF NOT EXISTS idx_saas_incidents_assigned
ON saas_incidents(owner_id, assigned_at DESC)
WHERE owner_id IS NOT NULL AND assigned_at IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN saas_incidents.owner_id IS 'P1: User ID of the incident owner (person responsible for resolving)';
COMMENT ON COLUMN saas_incidents.assigned_at IS 'P1: Timestamp when the incident was assigned to owner';
