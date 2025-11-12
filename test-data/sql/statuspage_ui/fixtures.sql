-- Test fixtures for statuspage_ui database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE page_widgets, page_subscribers, status_pages CASCADE;

-- Insert status pages for Tenant 1
INSERT INTO status_pages (id, tenant_id, brand_id, page_title, page_slug, description, is_public, show_uptime, show_incident_history, history_days, timezone, language, meta_title, meta_description, is_active, created_at, updated_at) VALUES
-- Main corporate status page
('page-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'brand-1111-0001-0001-000000000001', 'Test1 System Status', 'test1-status', 'Real-time status and uptime monitoring for Test1 services', true, true, true, 90, 'America/New_York', 'en', 'Test1 System Status - Service Health', 'Monitor the current status and uptime of Test1 services and infrastructure', true, NOW() - INTERVAL '6 months', NOW()),

-- Developer portal status page
('page-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'brand-1111-0002-0002-000000000002', 'Test1 Developer API Status', 'test1-dev-api', 'API availability and performance monitoring for developers', true, true, true, 30, 'America/New_York', 'en', 'Test1 Developer API Status', 'Check the real-time status of Test1 developer APIs', true, NOW() - INTERVAL '3 months', NOW()),

-- Internal-only status page
('page-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'brand-1111-0001-0001-000000000001', 'Test1 Internal Monitoring', 'test1-internal', 'Internal infrastructure monitoring dashboard', false, true, true, 180, 'America/New_York', 'en', 'Test1 Internal Status', 'Internal infrastructure health monitoring', true, NOW() - INTERVAL '6 months', NOW());

-- Insert status pages for Tenant 2
INSERT INTO status_pages (id, tenant_id, brand_id, page_title, page_slug, description, is_public, show_uptime, show_incident_history, history_days, timezone, language, meta_title, meta_description, is_active, created_at, updated_at) VALUES
('page-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'brand-2222-0001-0001-000000000001', 'Test2 Service Status', 'test2-status', 'Current status of Test2 platform and services', true, true, true, 60, 'UTC', 'en', 'Test2 Service Status', 'Real-time monitoring of Test2 platform health', true, NOW() - INTERVAL '3 months', NOW());

-- Insert status pages for Tenant 3
INSERT INTO status_pages (id, tenant_id, brand_id, page_title, page_slug, description, is_public, show_uptime, show_incident_history, history_days, timezone, language, meta_title, meta_description, is_active, created_at, updated_at) VALUES
('page-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'brand-3333-0001-0001-000000000001', 'Test3 Status', 'test3-status', 'Test3 service health and status updates', true, true, false, 30, 'UTC', 'en', 'Test3 Status', 'Monitor Test3 service availability', true, NOW(), NOW());

-- Insert page widgets for Tenant 1 main status page
INSERT INTO page_widgets (id, page_id, widget_type, widget_title, widget_config, display_order, is_visible, created_at, updated_at) VALUES
-- Hero/Header widget
('widget-1111-0001-0001-000000000001', 'page-1111-0001-0001-000000000001', 'header', 'System Status Overview',
'{"show_logo": true, "show_overall_status": true, "custom_message": "All systems operational"}',
1, true, NOW() - INTERVAL '6 months', NOW()),

-- Component status widget
('widget-1111-0001-0002-000000000002', 'page-1111-0001-0001-000000000001', 'component_status', 'Service Components',
'{"component_groups": ["group-1111-0001-0001-000000000001", "group-1111-0002-0002-000000000002", "group-1111-0003-0003-000000000003"], "show_description": true, "show_uptime": true}',
2, true, NOW() - INTERVAL '6 months', NOW()),

-- Incident timeline widget
('widget-1111-0001-0003-000000000003', 'page-1111-0001-0001-000000000001', 'incident_timeline', 'Recent Incidents',
'{"show_resolved": true, "max_incidents": 10, "show_updates": true}',
3, true, NOW() - INTERVAL '6 months', NOW()),

-- Uptime chart widget
('widget-1111-0001-0004-000000000004', 'page-1111-0001-0001-000000000001', 'uptime_chart', '90-Day Uptime',
'{"days": 90, "component_ids": ["comp-1111-0001-0001-000000000001", "comp-1111-0002-0002-000000000002", "comp-1111-0004-0004-000000000004", "comp-1111-0006-0006-000000000006"]}',
4, true, NOW() - INTERVAL '6 months', NOW()),

-- Metrics widget
('widget-1111-0001-0005-000000000005', 'page-1111-0001-0001-000000000001', 'metrics', 'Performance Metrics',
'{"metrics": [{"name": "API Response Time", "component_id": "comp-1111-0006-0006-000000000006", "metric_type": "response_time"}, {"name": "Database Uptime", "component_id": "comp-1111-0001-0001-000000000001", "metric_type": "uptime"}]}',
5, true, NOW() - INTERVAL '6 months', NOW()),

-- Subscribe widget
('widget-1111-0001-0006-000000000006', 'page-1111-0001-0001-000000000001', 'subscribe', 'Get Updates',
'{"channels": ["email", "sms", "webhook"], "message": "Subscribe to receive status updates"}',
6, true, NOW() - INTERVAL '6 months', NOW()),

-- Footer widget
('widget-1111-0001-0007-000000000007', 'page-1111-0001-0001-000000000001', 'footer', 'Contact & Links',
'{"show_powered_by": true, "links": [{"text": "Contact Support", "url": "https://test1.com/support"}, {"text": "Documentation", "url": "https://docs.test1.com"}]}',
7, true, NOW() - INTERVAL '6 months', NOW());

-- Insert page widgets for Tenant 1 developer status page
INSERT INTO page_widgets (id, page_id, widget_type, widget_title, widget_config, display_order, is_visible, created_at, updated_at) VALUES
('widget-1111-0002-0001-000000000001', 'page-1111-0002-0002-000000000002', 'header', 'Developer API Status',
'{"show_logo": true, "show_overall_status": true, "custom_message": "All APIs operational"}',
1, true, NOW() - INTERVAL '3 months', NOW()),

('widget-1111-0002-0002-000000000002', 'page-1111-0002-0002-000000000002', 'component_status', 'API Endpoints',
'{"component_groups": ["group-1111-0003-0003-000000000003"], "show_response_time": true}',
2, true, NOW() - INTERVAL '3 months', NOW()),

('widget-1111-0002-0003-000000000003', 'page-1111-0002-0002-000000000002', 'metrics', 'API Performance',
'{"metrics": [{"name": "REST API Latency", "component_id": "comp-1111-0006-0006-000000000006", "metric_type": "response_time"}, {"name": "GraphQL API Latency", "component_id": "comp-1111-0007-0007-000000000007", "metric_type": "response_time"}]}',
3, true, NOW() - INTERVAL '3 months', NOW()),

('widget-1111-0002-0004-000000000004', 'page-1111-0002-0002-000000000002', 'subscribe', 'API Updates',
'{"channels": ["email", "webhook"], "message": "Subscribe for API status notifications"}',
4, true, NOW() - INTERVAL '3 months', NOW());

-- Insert page widgets for Tenant 2 status page
INSERT INTO page_widgets (id, page_id, widget_type, widget_title, widget_config, display_order, is_visible, created_at, updated_at) VALUES
('widget-2222-0001-0001-000000000001', 'page-2222-0001-0001-000000000001', 'header', 'Test2 Status',
'{"show_logo": true, "show_overall_status": true}',
1, true, NOW() - INTERVAL '3 months', NOW()),

('widget-2222-0001-0002-000000000002', 'page-2222-0001-0001-000000000001', 'component_status', 'Services',
'{"component_groups": ["group-2222-0001-0001-000000000001"], "show_uptime": true}',
2, true, NOW() - INTERVAL '3 months', NOW()),

('widget-2222-0001-0003-000000000003', 'page-2222-0001-0001-000000000001', 'incident_timeline', 'Incidents',
'{"show_resolved": true, "max_incidents": 5}',
3, true, NOW() - INTERVAL '3 months', NOW()),

('widget-2222-0001-0004-000000000004', 'page-2222-0001-0001-000000000001', 'subscribe', 'Subscribe',
'{"channels": ["email", "sms"]}',
4, true, NOW() - INTERVAL '3 months', NOW());

-- Insert page widgets for Tenant 3 status page (minimal setup)
INSERT INTO page_widgets (id, page_id, widget_type, widget_title, widget_config, display_order, is_visible, created_at, updated_at) VALUES
('widget-3333-0001-0001-000000000001', 'page-3333-0001-0001-000000000001', 'header', 'Status',
'{"show_overall_status": true}',
1, true, NOW(), NOW()),

('widget-3333-0001-0002-000000000002', 'page-3333-0001-0001-000000000001', 'component_status', 'Components',
'{}',
2, true, NOW(), NOW());

-- Insert page subscribers
INSERT INTO page_subscribers (id, page_id, subscription_type, contact_value, is_verified, verification_token, subscribed_components, subscribed_incidents, is_active, created_at, updated_at, verified_at) VALUES
-- Tenant 1 main page subscribers
('sub-1111-0001-0001-000000000001', 'page-1111-0001-0001-000000000001', 'email', 'user1@example.com', true, NULL, '["comp-1111-0001-0001-000000000001", "comp-1111-0004-0004-000000000004", "comp-1111-0006-0006-000000000006"]', '["critical", "major"]', true, NOW() - INTERVAL '5 months', NOW(), NOW() - INTERVAL '5 months'),

('sub-1111-0001-0002-000000000002', 'page-1111-0001-0001-000000000001', 'email', 'user2@example.com', true, NULL, '[]', '["critical"]', true, NOW() - INTERVAL '4 months', NOW(), NOW() - INTERVAL '4 months'),

('sub-1111-0001-0003-000000000003', 'page-1111-0001-0001-000000000001', 'sms', '+15551234567', true, NULL, '["comp-1111-0001-0001-000000000001"]', '["critical", "major"]', true, NOW() - INTERVAL '3 months', NOW(), NOW() - INTERVAL '3 months'),

('sub-1111-0001-0004-000000000004', 'page-1111-0001-0001-000000000001', 'webhook', 'https://api.customer1.com/status-webhook', true, NULL, '[]', '["critical", "major", "minor"]', true, NOW() - INTERVAL '2 months', NOW(), NOW() - INTERVAL '2 months'),

('sub-1111-0001-0005-000000000005', 'page-1111-0001-0001-000000000001', 'email', 'pending@example.com', false, 'verify-token-1111-abcd1234', '[]', '["critical"]', true, NOW() - INTERVAL '1 day', NOW(), NULL),

-- Tenant 1 developer page subscribers
('sub-1111-0002-0001-000000000001', 'page-1111-0002-0002-000000000002', 'email', 'dev1@example.com', true, NULL, '["comp-1111-0006-0006-000000000006", "comp-1111-0007-0007-000000000007"]', '["critical", "major"]', true, NOW() - INTERVAL '2 months', NOW(), NOW() - INTERVAL '2 months'),

('sub-1111-0002-0002-000000000002', 'page-1111-0002-0002-000000000002', 'webhook', 'https://devtools.customer1.com/api-status', true, NULL, '["comp-1111-0006-0006-000000000006", "comp-1111-0007-0007-000000000007"]', '["critical", "major", "minor"]', true, NOW() - INTERVAL '1 month', NOW(), NOW() - INTERVAL '1 month'),

-- Tenant 2 subscribers
('sub-2222-0001-0001-000000000001', 'page-2222-0001-0001-000000000001', 'email', 'admin@test2.com', true, NULL, '[]', '["critical", "major"]', true, NOW() - INTERVAL '2 months', NOW(), NOW() - INTERVAL '2 months'),

('sub-2222-0001-0002-000000000002', 'page-2222-0001-0001-000000000001', 'email', 'ops@test2.com', true, NULL, '["comp-2222-0001-0001-000000000001", "comp-2222-0002-0002-000000000002"]', '["critical", "major", "minor"]', true, NOW() - INTERVAL '1 month', NOW(), NOW() - INTERVAL '1 month'),

('sub-2222-0001-0003-000000000003', 'page-2222-0001-0001-000000000001', 'sms', '+15559876543', true, NULL, '[]', '["critical"]', true, NOW() - INTERVAL '2 weeks', NOW(), NOW() - INTERVAL '2 weeks'),

-- Tenant 3 subscriber (trial)
('sub-3333-0001-0001-000000000001', 'page-3333-0001-0001-000000000001', 'email', 'admin@test3.com', true, NULL, '[]', '["critical", "major"]', true, NOW(), NOW(), NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as status_pages FROM status_pages;
SELECT COUNT(*) as page_widgets FROM page_widgets;
SELECT COUNT(*) as page_subscribers FROM page_subscribers;
SELECT is_public, COUNT(*) as count FROM status_pages GROUP BY is_public ORDER BY is_public;
SELECT widget_type, COUNT(*) as count FROM page_widgets GROUP BY widget_type ORDER BY widget_type;
SELECT subscription_type, COUNT(*) as count FROM page_subscribers GROUP BY subscription_type ORDER BY subscription_type;
SELECT is_verified, COUNT(*) as count FROM page_subscribers GROUP BY is_verified ORDER BY is_verified;
