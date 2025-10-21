--
-- Migration: Complete UUID conversion for all remaining tables
-- Date: 2025-10-21
-- Description: Converts all bigint/integer IDs to UUID for full UUID consistency
-- Rollback: Restore from backup
--

BEGIN;

-- =====================================================================
-- PART 1: Components Table
-- =====================================================================

-- Add UUID column for components.id
ALTER TABLE components
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid();

-- Update foreign key references in dependent tables
ALTER TABLE component_alerts
  ADD COLUMN IF NOT EXISTS component_id_uuid UUID;

ALTER TABLE component_history
  ADD COLUMN IF NOT EXISTS component_id_uuid UUID;

ALTER TABLE component_metrics
  ADD COLUMN IF NOT EXISTS component_id_uuid UUID;

ALTER TABLE component_statuses
  ADD COLUMN IF NOT EXISTS component_id_uuid UUID;

ALTER TABLE component_webhooks
  ADD COLUMN IF NOT EXISTS component_id_uuid UUID;

-- Populate UUID references
UPDATE component_alerts ca
SET component_id_uuid = (
  SELECT c.id_uuid FROM components c WHERE c.id = ca.component_id
)
WHERE ca.component_id_uuid IS NULL;

UPDATE component_history ch
SET component_id_uuid = (
  SELECT c.id_uuid FROM components c WHERE c.id = ch.component_id
)
WHERE ch.component_id_uuid IS NULL;

UPDATE component_metrics cm
SET component_id_uuid = (
  SELECT c.id_uuid FROM components c WHERE c.id = cm.component_id
)
WHERE cm.component_id_uuid IS NULL;

UPDATE component_statuses cs
SET component_id_uuid = (
  SELECT c.id_uuid FROM components c WHERE c.id = cs.component_id
)
WHERE cs.component_id_uuid IS NULL;

UPDATE component_webhooks cw
SET component_id_uuid = (
  SELECT c.id_uuid FROM components c WHERE c.id = cw.component_id
)
WHERE cw.component_id_uuid IS NULL;

-- Drop old primary key and columns for components
ALTER TABLE components DROP CONSTRAINT IF EXISTS components_pkey CASCADE;
ALTER TABLE components DROP COLUMN IF EXISTS id;

-- Drop old foreign key columns
ALTER TABLE component_alerts DROP COLUMN IF EXISTS component_id;
ALTER TABLE component_history DROP COLUMN IF EXISTS component_id;
ALTER TABLE component_metrics DROP COLUMN IF EXISTS component_id;
ALTER TABLE component_statuses DROP COLUMN IF EXISTS component_id;
ALTER TABLE component_webhooks DROP COLUMN IF EXISTS component_id;

-- Rename UUID columns
ALTER TABLE components RENAME COLUMN id_uuid TO id;
ALTER TABLE component_alerts RENAME COLUMN component_id_uuid TO component_id;
ALTER TABLE component_history RENAME COLUMN component_id_uuid TO component_id;
ALTER TABLE component_metrics RENAME COLUMN component_id_uuid TO component_id;
ALTER TABLE component_statuses RENAME COLUMN component_id_uuid TO component_id;
ALTER TABLE component_webhooks RENAME COLUMN component_id_uuid TO component_id;

-- Set NOT NULL and add primary key
ALTER TABLE components ALTER COLUMN id SET NOT NULL;
ALTER TABLE components ADD PRIMARY KEY (id);

-- Set NOT NULL for foreign keys
ALTER TABLE component_alerts ALTER COLUMN component_id SET NOT NULL;
ALTER TABLE component_history ALTER COLUMN component_id SET NOT NULL;
ALTER TABLE component_metrics ALTER COLUMN component_id SET NOT NULL;
ALTER TABLE component_statuses ALTER COLUMN component_id SET NOT NULL;
ALTER TABLE component_webhooks ALTER COLUMN component_id SET NOT NULL;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_components_tenant_id ON components(tenant_id);
CREATE INDEX IF NOT EXISTS idx_components_deleted_at ON components(deleted_at);
CREATE INDEX IF NOT EXISTS idx_component_alerts_component_id ON component_alerts(component_id);
CREATE INDEX IF NOT EXISTS idx_component_history_component_id ON component_history(component_id);
CREATE INDEX IF NOT EXISTS idx_component_metrics_component_id ON component_metrics(component_id);
CREATE INDEX IF NOT EXISTS idx_component_statuses_component_id ON component_statuses(component_id);
CREATE INDEX IF NOT EXISTS idx_component_webhooks_component_id ON component_webhooks(component_id);

-- =====================================================================
-- PART 2: Team Tables (teams, team_members, team_roles)
-- =====================================================================

-- Convert teams table
ALTER TABLE teams
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS created_by_uuid UUID;

-- Populate tenant_id for teams - if it's bigint, we need to handle empty tables
-- For empty tables, we'll just set up the structure
UPDATE teams SET tenant_id_uuid = NULL WHERE tenant_id_uuid IS NULL;
UPDATE teams SET created_by_uuid = NULL WHERE created_by_uuid IS NULL;

-- Update team_members references
ALTER TABLE team_members
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS team_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS user_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS added_by_uuid UUID;

-- Update team_roles references
ALTER TABLE team_roles
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS team_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS role_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS assigned_by_uuid UUID;

-- For empty tables, just prepare structure
UPDATE team_members SET team_id_uuid = NULL WHERE team_id_uuid IS NULL;
UPDATE team_members SET tenant_id_uuid = NULL WHERE tenant_id_uuid IS NULL;
UPDATE team_members SET added_by_uuid = NULL WHERE added_by_uuid IS NULL;

UPDATE team_roles SET team_id_uuid = NULL WHERE team_id_uuid IS NULL;
UPDATE team_roles SET tenant_id_uuid = NULL WHERE tenant_id_uuid IS NULL;
UPDATE team_roles SET assigned_by_uuid = NULL WHERE assigned_by_uuid IS NULL;

-- Drop old constraints and columns
ALTER TABLE teams DROP CONSTRAINT IF EXISTS teams_pkey CASCADE;
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_pkey CASCADE;
ALTER TABLE team_roles DROP CONSTRAINT IF EXISTS team_roles_pkey CASCADE;

ALTER TABLE teams DROP COLUMN IF EXISTS id;
ALTER TABLE teams DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE teams DROP COLUMN IF EXISTS created_by;

ALTER TABLE team_members DROP COLUMN IF EXISTS id;
ALTER TABLE team_members DROP COLUMN IF EXISTS team_id;
ALTER TABLE team_members DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE team_members DROP COLUMN IF EXISTS added_by;

ALTER TABLE team_roles DROP COLUMN IF EXISTS id;
ALTER TABLE team_roles DROP COLUMN IF EXISTS team_id;
ALTER TABLE team_roles DROP COLUMN IF EXISTS role_id;
ALTER TABLE team_roles DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE team_roles DROP COLUMN IF EXISTS assigned_by;

-- Rename UUID columns
ALTER TABLE teams RENAME COLUMN id_uuid TO id;
ALTER TABLE teams RENAME COLUMN tenant_id_uuid TO tenant_id;
ALTER TABLE teams RENAME COLUMN created_by_uuid TO created_by;

ALTER TABLE team_members RENAME COLUMN id_uuid TO id;
ALTER TABLE team_members RENAME COLUMN team_id_uuid TO team_id;
ALTER TABLE team_members RENAME COLUMN tenant_id_uuid TO tenant_id;
ALTER TABLE team_members RENAME COLUMN added_by_uuid TO added_by;

ALTER TABLE team_roles RENAME COLUMN id_uuid TO id;
ALTER TABLE team_roles RENAME COLUMN team_id_uuid TO team_id;
ALTER TABLE team_roles RENAME COLUMN role_id_uuid TO role_id;
ALTER TABLE team_roles RENAME COLUMN tenant_id_uuid TO tenant_id;
ALTER TABLE team_roles RENAME COLUMN assigned_by_uuid TO assigned_by;

-- Add NOT NULL constraints and primary keys
ALTER TABLE teams ALTER COLUMN id SET NOT NULL;
ALTER TABLE teams ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE teams ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE teams ADD PRIMARY KEY (id);

ALTER TABLE team_members ALTER COLUMN id SET NOT NULL;
ALTER TABLE team_members ALTER COLUMN team_id SET NOT NULL;
ALTER TABLE team_members ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE team_members ALTER COLUMN added_by SET NOT NULL;
ALTER TABLE team_members ADD PRIMARY KEY (id);

ALTER TABLE team_roles ALTER COLUMN id SET NOT NULL;
ALTER TABLE team_roles ALTER COLUMN team_id SET NOT NULL;
ALTER TABLE team_roles ALTER COLUMN role_id SET NOT NULL;
ALTER TABLE team_roles ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE team_roles ALTER COLUMN assigned_by SET NOT NULL;
ALTER TABLE team_roles ADD PRIMARY KEY (id);

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_teams_tenant_id ON teams(tenant_id);
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_tenant_id ON team_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_team_roles_team_id ON team_roles(team_id);
CREATE INDEX IF NOT EXISTS idx_team_roles_role_id ON team_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_team_roles_tenant_id ON team_roles(tenant_id);

-- =====================================================================
-- PART 3: Audit Logs Table
-- =====================================================================

-- Convert audit_logs table
ALTER TABLE audit_logs
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS user_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS resource_id_uuid UUID;

-- For empty tables, prepare structure
UPDATE audit_logs SET tenant_id_uuid = NULL WHERE tenant_id_uuid IS NULL;
UPDATE audit_logs SET user_id_uuid = NULL WHERE user_id_uuid IS NULL;
UPDATE audit_logs SET resource_id_uuid = NULL WHERE resource_id_uuid IS NULL;

-- Drop old constraints and columns
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pkey CASCADE;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS user_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS resource_id;

-- Rename UUID columns
ALTER TABLE audit_logs RENAME COLUMN id_uuid TO id;
ALTER TABLE audit_logs RENAME COLUMN tenant_id_uuid TO tenant_id;
ALTER TABLE audit_logs RENAME COLUMN user_id_uuid TO user_id;
ALTER TABLE audit_logs RENAME COLUMN resource_id_uuid TO resource_id;

-- Add NOT NULL constraints and primary key
ALTER TABLE audit_logs ALTER COLUMN id SET NOT NULL;
ALTER TABLE audit_logs ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE audit_logs ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE audit_logs ADD PRIMARY KEY (id);

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);

-- =====================================================================
-- PART 4: User Roles Table (ID column)
-- =====================================================================

-- Convert user_roles table ID to UUID (other columns already UUID from migration 007)
ALTER TABLE user_roles
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid();

-- Drop old constraints and columns
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_pkey CASCADE;
ALTER TABLE user_roles DROP COLUMN IF EXISTS id;

-- Rename UUID column
ALTER TABLE user_roles RENAME COLUMN id_uuid TO id;

-- Add NOT NULL constraint and primary key
ALTER TABLE user_roles ALTER COLUMN id SET NOT NULL;
ALTER TABLE user_roles ADD PRIMARY KEY (id);

-- =====================================================================
-- PART 5: User Sessions Table
-- =====================================================================

-- Convert user_sessions table
ALTER TABLE user_sessions
  ADD COLUMN IF NOT EXISTS id_uuid UUID DEFAULT gen_random_uuid(),
  ADD COLUMN IF NOT EXISTS user_id_uuid UUID,
  ADD COLUMN IF NOT EXISTS tenant_id_uuid UUID;

-- For empty tables, prepare structure
UPDATE user_sessions SET user_id_uuid = NULL WHERE user_id_uuid IS NULL;
UPDATE user_sessions SET tenant_id_uuid = NULL WHERE tenant_id_uuid IS NULL;

-- Drop old constraints and columns
ALTER TABLE user_sessions DROP CONSTRAINT IF EXISTS user_sessions_pkey CASCADE;
ALTER TABLE user_sessions DROP COLUMN IF EXISTS id;
ALTER TABLE user_sessions DROP COLUMN IF EXISTS user_id;
-- Note: tenant_id might already be UUID or text, handle accordingly
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'user_sessions'
    AND column_name = 'tenant_id'
    AND data_type != 'uuid'
  ) THEN
    ALTER TABLE user_sessions DROP COLUMN IF EXISTS tenant_id;
  END IF;
END $$;

-- Rename UUID columns
ALTER TABLE user_sessions RENAME COLUMN id_uuid TO id;
ALTER TABLE user_sessions RENAME COLUMN user_id_uuid TO user_id;
ALTER TABLE user_sessions RENAME COLUMN tenant_id_uuid TO tenant_id;

-- Add NOT NULL constraints and primary key
ALTER TABLE user_sessions ALTER COLUMN id SET NOT NULL;
ALTER TABLE user_sessions ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE user_sessions ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE user_sessions ADD PRIMARY KEY (id);

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_tenant_id ON user_sessions(tenant_id);

-- =====================================================================
-- PART 6: Incidents Table - Convert created_by and updated_by
-- =====================================================================

-- Add UUID columns for created_by and updated_by
ALTER TABLE incidents
  ADD COLUMN IF NOT EXISTS created_by_uuid UUID,
  ADD COLUMN IF NOT EXISTS updated_by_uuid UUID;

-- Populate from users table if data exists
UPDATE incidents i
SET created_by_uuid = (
  SELECT u.id FROM users u WHERE u.id::text = i.created_by::text
)
WHERE i.created_by IS NOT NULL AND i.created_by_uuid IS NULL;

UPDATE incidents i
SET updated_by_uuid = (
  SELECT u.id FROM users u WHERE u.id::text = i.updated_by::text
)
WHERE i.updated_by IS NOT NULL AND i.updated_by_uuid IS NULL;

-- Drop old columns
ALTER TABLE incidents DROP COLUMN IF EXISTS created_by;
ALTER TABLE incidents DROP COLUMN IF EXISTS updated_by;

-- Rename UUID columns
ALTER TABLE incidents RENAME COLUMN created_by_uuid TO created_by;
ALTER TABLE incidents RENAME COLUMN updated_by_uuid TO updated_by;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_incidents_created_by ON incidents(created_by);
CREATE INDEX IF NOT EXISTS idx_incidents_updated_by ON incidents(updated_by);

COMMIT;

-- Verification queries (run separately after migration):
-- SELECT id, tenant_id, name FROM components LIMIT 3;
-- SELECT id, tenant_id, name FROM teams LIMIT 3;
-- SELECT id, team_id, user_id FROM team_members LIMIT 3;
-- SELECT id, tenant_id, user_id, action FROM audit_logs LIMIT 3;
-- SELECT id, user_id, role_id FROM user_roles LIMIT 3;
-- SELECT id, user_id, tenant_id FROM user_sessions LIMIT 3;
-- SELECT id, tenant_id, title, created_by, updated_by FROM incidents LIMIT 3;
