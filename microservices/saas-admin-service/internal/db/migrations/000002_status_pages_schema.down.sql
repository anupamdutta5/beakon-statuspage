-- Drop the public status page view
DROP VIEW IF EXISTS public_status_page_data;

-- Drop triggers
DROP TRIGGER IF EXISTS update_status_pages_updated_at ON status_pages;
DROP TRIGGER IF EXISTS update_status_page_domains_updated_at ON status_page_domains;
DROP TRIGGER IF EXISTS update_status_page_settings_updated_at ON status_page_settings;
DROP TRIGGER IF EXISTS update_monitors_updated_at ON monitors;

-- Drop the update function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the status page monitors table
DROP TABLE IF EXISTS status_page_monitors;

-- Drop monitor status history
DROP TABLE IF EXISTS monitor_status_history;

-- Drop monitors table
DROP TABLE IF EXISTS monitors;

-- Drop monitor type enum
DROP TYPE IF EXISTS monitor_type;

-- Drop status page settings
DROP TABLE IF EXISTS status_page_settings;

-- Drop status page domains
DROP TABLE IF EXISTS status_page_domains;

-- Drop status pages table
DROP TABLE IF EXISTS status_pages;

-- Remove added columns from saas_plans
ALTER TABLE saas_plans 
DROP COLUMN IF EXISTS max_status_pages,
DROP COLUMN IF EXISTS max_custom_domains,
DROP COLUMN IF EXISTS max_monitors_per_page,
DROP COLUMN IF EXISTS custom_status_pages,
DROP COLUMN IF EXISTS custom_domains,
DROP COLUMN IF EXISTS ssl_certificates,
DROP COLUMN IF EXISTS advanced_analytics,
DROP COLUMN IF EXISTS team_collaboration;
