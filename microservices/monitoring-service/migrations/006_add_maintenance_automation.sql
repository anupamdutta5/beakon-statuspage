-- Migration: Add Maintenance Automation Fields
-- Date: 2025-10-22
-- Description: Adds automation tracking fields for maintenance windows

-- Add automation tracking columns
ALTER TABLE maintenance_windows
ADD COLUMN IF NOT EXISTS reminder_sent BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS auto_started BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS auto_completed BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS actual_start_time TIMESTAMP,
ADD COLUMN IF NOT EXISTS actual_end_time TIMESTAMP,
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'scheduled';

-- Add index for automation job queries
CREATE INDEX IF NOT EXISTS idx_maintenance_automation
ON maintenance_windows(status, starts_at)
WHERE is_active = true;

-- Add index for reminder queries
CREATE INDEX IF NOT EXISTS idx_maintenance_reminders
ON maintenance_windows(reminder_sent, starts_at)
WHERE is_active = true AND reminder_sent = false;

-- Update existing records to have status
UPDATE maintenance_windows
SET status = 'scheduled'
WHERE status IS NULL;

-- Add check constraint for valid statuses
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_maintenance_status'
        AND conrelid = 'maintenance_windows'::regclass
    ) THEN
        ALTER TABLE maintenance_windows
        ADD CONSTRAINT chk_maintenance_status
        CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled'));
    END IF;
END $$;

COMMENT ON COLUMN maintenance_windows.reminder_sent IS 'Whether 60-minute reminder has been sent';
COMMENT ON COLUMN maintenance_windows.auto_started IS 'Whether maintenance was auto-started';
COMMENT ON COLUMN maintenance_windows.auto_completed IS 'Whether maintenance was auto-completed';
COMMENT ON COLUMN maintenance_windows.actual_start_time IS 'Actual time maintenance started (for tracking)';
COMMENT ON COLUMN maintenance_windows.actual_end_time IS 'Actual time maintenance completed (for tracking)';
COMMENT ON COLUMN maintenance_windows.status IS 'Current status: scheduled, in_progress, completed, cancelled';
