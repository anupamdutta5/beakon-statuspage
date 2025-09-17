-- Drop triggers and functions first to avoid dependency issues
DROP TRIGGER IF EXISTS ensure_single_primary_domain_trigger ON status_page_domains;
DROP TRIGGER IF EXISTS update_status_page_domains_modtime ON status_page_domains;

-- Drop functions
DROP FUNCTION IF EXISTS ensure_single_primary_domain();
DROP FUNCTION IF EXISTS get_primary_domain(BIGINT);
DROP FUNCTION IF EXISTS update_modified_column();

-- Drop the table
DROP TABLE IF EXISTS status_page_domains;
