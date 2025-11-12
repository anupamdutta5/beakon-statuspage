-- Test fixtures for statuspage_incident database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE incident_component_impacts, incident_updates, incident_templates, incidents CASCADE;

-- Insert incident templates
INSERT INTO incident_templates (id, tenant_id, name, title_template, message_template, severity, created_at, updated_at) VALUES
('tmpl-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Database Outage', 'Database Connection Issues', 'We are experiencing connectivity issues with our database. Our team is investigating.', 'critical', NOW(), NOW()),
('tmpl-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'API Slowness', 'API Performance Degradation', 'API response times are higher than normal. We are working to resolve this.', 'major', NOW(), NOW()),
('tmpl-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'Scheduled Maintenance', 'Scheduled Maintenance Window', 'We will be performing scheduled maintenance during the announced window.', 'maintenance', NOW(), NOW()),
('tmpl-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'Service Disruption', 'Service Temporarily Unavailable', 'Our service is currently experiencing issues. We are working on a fix.', 'major', NOW(), NOW());

-- Insert test incidents for Tenant 1
INSERT INTO incidents (id, tenant_id, title, description, status, severity, started_at, resolved_at, created_at, updated_at, created_by) VALUES
-- Active critical incident
('inc-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Database Connection Failure', 'Primary database is not accepting connections', 'investigating', 'critical', NOW() - INTERVAL '30 minutes', NULL, NOW() - INTERVAL '30 minutes', NOW(), 'user-1111-0001-0001-000000000001'),
-- Recent resolved major incident
('inc-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'API Response Time Degradation', 'API endpoints responding slowly', 'resolved', 'major', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '2 hours', 'user-1111-0002-0002-000000000002'),
-- Scheduled maintenance
('inc-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'Scheduled Database Upgrade', 'Upgrading PostgreSQL to version 16', 'scheduled', 'maintenance', NOW() + INTERVAL '2 days', NULL, NOW(), NOW(), 'user-1111-0001-0001-000000000001'),
-- Old resolved incident
('inc-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'Payment Gateway Timeout', 'Payment processing delayed', 'resolved', 'major', NOW() - INTERVAL '7 days', NOW() - INTERVAL '6 days', NOW() - INTERVAL '7 days', NOW() - INTERVAL '6 days', 'user-1111-0002-0002-000000000002'),
-- Minor incident
('inc-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'Email Delivery Delays', 'Notification emails delayed by 5 minutes', 'monitoring', 'minor', NOW() - INTERVAL '1 hour', NULL, NOW() - INTERVAL '1 hour', NOW(), 'user-1111-0001-0001-000000000001');

-- Insert test incidents for Tenant 2
INSERT INTO incidents (id, tenant_id, title, description, status, severity, started_at, resolved_at, created_at, updated_at, created_by) VALUES
('inc-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'CDN Configuration Error', 'Static assets not loading', 'resolved', 'major', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '1 hour', 'user-2222-0001-0001-000000000001');

-- Insert incident updates
INSERT INTO incident_updates (id, incident_id, status, message, created_at, created_by) VALUES
-- Updates for inc-1111-0001 (active critical)
('upd-1111-0001-0001-000000000001', 'inc-1111-0001-0001-000000000001', 'investigating', 'We have identified connectivity issues with the primary database and are investigating the root cause.', NOW() - INTERVAL '30 minutes', 'user-1111-0001-0001-000000000001'),
('upd-1111-0001-0002-000000000002', 'inc-1111-0001-0001-000000000001', 'investigating', 'Database team has been engaged. We are working to restore connectivity.', NOW() - INTERVAL '20 minutes', 'user-1111-0001-0001-000000000001'),
('upd-1111-0001-0003-000000000003', 'inc-1111-0001-0001-000000000001', 'investigating', 'Failover to secondary database in progress. ETA 10 minutes.', NOW() - INTERVAL '10 minutes', 'user-1111-0002-0002-000000000002'),

-- Updates for inc-1111-0002 (resolved major)
('upd-1111-0002-0001-000000000001', 'inc-1111-0002-0002-000000000002', 'investigating', 'We are seeing increased API response times and investigating the cause.', NOW() - INTERVAL '5 hours', 'user-1111-0002-0002-000000000002'),
('upd-1111-0002-0002-000000000002', 'inc-1111-0002-0002-000000000002', 'identified', 'Identified database query optimization issue. Applying fixes.', NOW() - INTERVAL '4 hours', 'user-1111-0002-0002-000000000002'),
('upd-1111-0002-0003-000000000003', 'inc-1111-0002-0002-000000000002', 'monitoring', 'Fixes deployed. Monitoring API response times.', NOW() - INTERVAL '3 hours', 'user-1111-0001-0001-000000000001'),
('upd-1111-0002-0004-000000000004', 'inc-1111-0002-0002-000000000002', 'resolved', 'API response times back to normal. Incident resolved.', NOW() - INTERVAL '2 hours', 'user-1111-0001-0001-000000000001'),

-- Updates for inc-1111-0003 (scheduled maintenance)
('upd-1111-0003-0001-000000000001', 'inc-1111-0003-0003-000000000003', 'scheduled', 'Maintenance window scheduled for 2 days from now. Expected duration: 2 hours.', NOW(), 'user-1111-0001-0001-000000000001'),

-- Updates for inc-1111-0005 (monitoring minor)
('upd-1111-0005-0001-000000000001', 'inc-1111-0005-0005-000000000005', 'investigating', 'Email delivery experiencing delays of approximately 5 minutes.', NOW() - INTERVAL '1 hour', 'user-1111-0001-0001-000000000001'),
('upd-1111-0005-0002-000000000002', 'inc-1111-0005-0005-000000000005', 'monitoring', 'Delay reduced to 2 minutes. Continuing to monitor.', NOW() - INTERVAL '30 minutes', 'user-1111-0001-0001-000000000001'),

-- Updates for inc-2222-0001 (resolved)
('upd-2222-0001-0001-000000000001', 'inc-2222-0001-0001-000000000001', 'investigating', 'CDN configuration issue identified. Correcting now.', NOW() - INTERVAL '3 hours', 'user-2222-0001-0001-000000000001'),
('upd-2222-0001-0002-000000000002', 'inc-2222-0001-0001-000000000001', 'resolved', 'CDN configuration corrected. All assets loading normally.', NOW() - INTERVAL '1 hour', 'user-2222-0001-0001-000000000001');

-- Insert component impacts
INSERT INTO incident_component_impacts (incident_id, component_id, impact_level) VALUES
-- inc-1111-0001 impacts (critical database incident)
('inc-1111-0001-0001-000000000001', 'comp-1111-0001-0001-000000000001', 'major_outage'),
('inc-1111-0001-0001-000000000001', 'comp-1111-0004-0004-000000000004', 'partial_outage'),
('inc-1111-0001-0001-000000000001', 'comp-1111-0006-0006-000000000006', 'degraded_performance'),

-- inc-1111-0002 impacts (API slowness)
('inc-1111-0002-0002-000000000002', 'comp-1111-0006-0006-000000000006', 'degraded_performance'),
('inc-1111-0002-0002-000000000002', 'comp-1111-0007-0007-000000000007', 'degraded_performance'),

-- inc-1111-0003 impacts (scheduled maintenance)
('inc-1111-0003-0003-000000000003', 'comp-1111-0001-0001-000000000001', 'maintenance'),
('inc-1111-0003-0003-000000000003', 'comp-1111-0004-0004-000000000004', 'maintenance'),

-- inc-1111-0005 impacts (email delays)
('inc-1111-0005-0005-000000000005', 'comp-1111-0004-0004-000000000004', 'degraded_performance'),

-- inc-2222-0001 impacts (CDN issue)
('inc-2222-0001-0001-000000000001', 'comp-2222-0001-0001-000000000001', 'partial_outage');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as templates FROM incident_templates;
SELECT COUNT(*) as incidents FROM incidents;
SELECT COUNT(*) as updates FROM incident_updates;
SELECT COUNT(*) as component_impacts FROM incident_component_impacts;
SELECT severity, COUNT(*) as count FROM incidents GROUP BY severity ORDER BY severity;
SELECT status, COUNT(*) as count FROM incidents GROUP BY status ORDER BY status;
