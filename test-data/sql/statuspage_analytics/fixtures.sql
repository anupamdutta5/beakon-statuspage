-- Test fixtures for statuspage_analytics database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE uptime_metrics, response_time_metrics, sla_tracking, page_view_analytics CASCADE;

-- Insert uptime metrics for Tenant 1 components
-- Last 24 hours of data (hourly intervals)
INSERT INTO uptime_metrics (id, tenant_id, component_id, timestamp, status, uptime_percentage, downtime_seconds, created_at) VALUES
-- Database component (99.95% uptime)
('metric-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '24 hours', 'operational', 100.0, 0, NOW() - INTERVAL '24 hours'),
('metric-1111-0001-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '23 hours', 'operational', 100.0, 0, NOW() - INTERVAL '23 hours'),
('metric-1111-0001-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '22 hours', 'operational', 100.0, 0, NOW() - INTERVAL '22 hours'),
('metric-1111-0001-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '12 hours', 'degraded_performance', 95.0, 180, NOW() - INTERVAL '12 hours'),
('metric-1111-0001-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '6 hours', 'operational', 100.0, 0, NOW() - INTERVAL '6 hours'),
('metric-1111-0001-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '1 hour', 'operational', 100.0, 0, NOW() - INTERVAL '1 hour'),

-- Redis Cache component (99.99% uptime)
('metric-1111-0002-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '24 hours', 'operational', 100.0, 0, NOW() - INTERVAL '24 hours'),
('metric-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '12 hours', 'operational', 100.0, 0, NOW() - INTERVAL '12 hours'),
('metric-1111-0002-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '6 hours', 'operational', 100.0, 0, NOW() - INTERVAL '6 hours'),
('metric-1111-0002-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '1 hour', 'operational', 100.0, 0, NOW() - INTERVAL '1 hour'),

-- Message Queue component (currently degraded)
('metric-1111-0003-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', NOW() - INTERVAL '24 hours', 'operational', 100.0, 0, NOW() - INTERVAL '24 hours'),
('metric-1111-0003-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', NOW() - INTERVAL '12 hours', 'operational', 100.0, 0, NOW() - INTERVAL '12 hours'),
('metric-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', NOW() - INTERVAL '2 hours', 'degraded_performance', 85.0, 540, NOW() - INTERVAL '2 hours'),
('metric-1111-0003-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', NOW() - INTERVAL '1 hour', 'degraded_performance', 80.0, 720, NOW() - INTERVAL '1 hour'),

-- Web Application component
('metric-1111-0004-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0004-0004-000000000004', NOW() - INTERVAL '24 hours', 'operational', 100.0, 0, NOW() - INTERVAL '24 hours'),
('metric-1111-0004-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0004-0004-000000000004', NOW() - INTERVAL '12 hours', 'operational', 100.0, 0, NOW() - INTERVAL '12 hours'),
('metric-1111-0004-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0004-0004-000000000004', NOW() - INTERVAL '1 hour', 'operational', 100.0, 0, NOW() - INTERVAL '1 hour');

-- Insert response time metrics
INSERT INTO response_time_metrics (id, tenant_id, component_id, timestamp, avg_response_time_ms, min_response_time_ms, max_response_time_ms, p50_ms, p95_ms, p99_ms, request_count, created_at) VALUES
-- Database response times
('resp-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '24 hours', 42.5, 15.2, 125.8, 38.0, 85.0, 120.0, 15280, NOW() - INTERVAL '24 hours'),
('resp-1111-0001-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '12 hours', 45.2, 18.5, 142.3, 41.0, 92.0, 135.0, 16432, NOW() - INTERVAL '12 hours'),
('resp-1111-0001-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '6 hours', 38.7, 12.8, 98.5, 35.0, 78.0, 95.0, 14560, NOW() - INTERVAL '6 hours'),
('resp-1111-0001-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '1 hour', 41.2, 16.4, 108.9, 38.5, 82.0, 105.0, 15789, NOW() - INTERVAL '1 hour'),

-- Redis Cache response times
('resp-1111-0002-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '24 hours', 2.3, 0.8, 15.2, 2.0, 5.5, 12.0, 45820, NOW() - INTERVAL '24 hours'),
('resp-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '12 hours', 2.1, 0.9, 12.8, 1.9, 4.8, 10.5, 48320, NOW() - INTERVAL '12 hours'),
('resp-1111-0002-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', NOW() - INTERVAL '1 hour', 2.5, 1.0, 18.5, 2.2, 6.2, 15.0, 46789, NOW() - INTERVAL '1 hour'),

-- REST API response times
('resp-1111-0006-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0006-0006-000000000006', NOW() - INTERVAL '24 hours', 125.8, 45.2, 852.3, 98.0, 285.0, 650.0, 8540, NOW() - INTERVAL '24 hours'),
('resp-1111-0006-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0006-0006-000000000006', NOW() - INTERVAL '12 hours', 132.4, 52.1, 945.7, 105.0, 312.0, 720.0, 9120, NOW() - INTERVAL '12 hours'),
('resp-1111-0006-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0006-0006-000000000006', NOW() - INTERVAL '1 hour', 118.9, 48.5, 725.4, 92.0, 265.0, 580.0, 8890, NOW() - INTERVAL '1 hour');

-- Insert SLA tracking (monthly data)
INSERT INTO sla_tracking (id, tenant_id, component_id, period_start, period_end, target_uptime_percentage, actual_uptime_percentage, total_downtime_seconds, total_incidents, sla_met, created_at, updated_at) VALUES
-- Current month data for Tenant 1 components
('sla-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 99.95, 216, 1, true, NOW(), NOW()),
('sla-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 99.99, 43, 0, true, NOW(), NOW()),
('sla-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 98.5, 6480, 2, false, NOW(), NOW()),
('sla-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0004-0004-000000000004', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 100.0, 0, 0, true, NOW(), NOW()),
('sla-1111-0006-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0006-0006-000000000006', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 99.97, 129, 0, true, NOW(), NOW()),

-- Previous month data for Tenant 1 components
('sla-1111-0001-0001-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', DATE_TRUNC('month', NOW() - INTERVAL '1 month'), DATE_TRUNC('month', NOW()), 99.9, 99.98, 86, 0, true, NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month'),
('sla-1111-0002-0002-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', DATE_TRUNC('month', NOW() - INTERVAL '1 month'), DATE_TRUNC('month', NOW()), 99.9, 100.0, 0, 0, true, NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month'),
('sla-1111-0003-0003-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', DATE_TRUNC('month', NOW() - INTERVAL '1 month'), DATE_TRUNC('month', NOW()), 99.9, 99.92, 345, 1, true, NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month'),

-- Current month data for Tenant 2 components
('sla-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'comp-2222-0001-0001-000000000001', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 99.9, 432, 1, true, NOW(), NOW()),
('sla-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'comp-2222-0002-0002-000000000002', DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month', 99.9, 99.95, 216, 0, true, NOW(), NOW());

-- Insert page view analytics
INSERT INTO page_view_analytics (id, tenant_id, page_type, page_id, timestamp, unique_visitors, total_views, avg_time_on_page_seconds, bounce_rate, created_at) VALUES
-- Status page views for Tenant 1
('view-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'status_page', 'tenant-1111-1111-1111-111111111111', NOW() - INTERVAL '24 hours', 1250, 2840, 45.2, 35.5, NOW() - INTERVAL '24 hours'),
('view-1111-0001-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'status_page', 'tenant-1111-1111-1111-111111111111', NOW() - INTERVAL '12 hours', 1580, 3420, 52.8, 28.3, NOW() - INTERVAL '12 hours'),
('view-1111-0001-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'status_page', 'tenant-1111-1111-1111-111111111111', NOW() - INTERVAL '6 hours', 980, 1890, 38.5, 42.1, NOW() - INTERVAL '6 hours'),
('view-1111-0001-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'status_page', 'tenant-1111-1111-1111-111111111111', NOW() - INTERVAL '1 hour', 340, 625, 35.8, 45.2, NOW() - INTERVAL '1 hour'),

-- Incident page views for Tenant 1
('view-1111-0002-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'incident_page', 'inc-1111-0001-0001-000000000001', NOW() - INTERVAL '1 hour', 520, 1240, 125.4, 15.2, NOW() - INTERVAL '1 hour'),
('view-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'incident_page', 'inc-1111-0001-0001-000000000001', NOW() - INTERVAL '30 minutes', 680, 1580, 142.8, 12.5, NOW() - INTERVAL '30 minutes'),

-- Component page views for Tenant 1
('view-1111-0003-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'component_page', 'comp-1111-0001-0001-000000000001', NOW() - INTERVAL '24 hours', 185, 320, 78.5, 55.2, NOW() - INTERVAL '24 hours'),
('view-1111-0003-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'component_page', 'comp-1111-0006-0006-000000000006', NOW() - INTERVAL '24 hours', 245, 485, 92.3, 48.7, NOW() - INTERVAL '24 hours'),

-- Status page views for Tenant 2
('view-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'status_page', 'tenant-2222-2222-2222-222222222222', NOW() - INTERVAL '24 hours', 850, 1640, 42.8, 38.5, NOW() - INTERVAL '24 hours'),
('view-2222-0001-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'status_page', 'tenant-2222-2222-2222-222222222222', NOW() - INTERVAL '1 hour', 220, 380, 38.2, 42.5, NOW() - INTERVAL '1 hour');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as uptime_metrics FROM uptime_metrics;
SELECT COUNT(*) as response_time_metrics FROM response_time_metrics;
SELECT COUNT(*) as sla_tracking FROM sla_tracking;
SELECT COUNT(*) as page_views FROM page_view_analytics;
SELECT sla_met, COUNT(*) as count FROM sla_tracking GROUP BY sla_met ORDER BY sla_met;
SELECT page_type, COUNT(*) as count FROM page_view_analytics GROUP BY page_type ORDER BY page_type;
