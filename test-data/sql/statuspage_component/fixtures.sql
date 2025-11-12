-- Test fixtures for statuspage_component database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE component_metrics, components, component_groups CASCADE;

-- Insert test component groups for Tenant 1
INSERT INTO component_groups (id, tenant_id, name, description, display_order, created_at, updated_at) VALUES
('group-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Infrastructure', 'Core infrastructure components', 1, NOW(), NOW()),
('group-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'Applications', 'Application services', 2, NOW(), NOW()),
('group-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'APIs', 'API endpoints', 3, NOW(), NOW());

-- Insert test component groups for Tenant 2
INSERT INTO component_groups (id, tenant_id, name, description, display_order, created_at, updated_at) VALUES
('group-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'Services', 'Microservices', 1, NOW(), NOW());

-- Insert test components for Tenant 1
INSERT INTO components (id, tenant_id, group_id, name, description, status, display_order, created_at, updated_at) VALUES
-- Infrastructure group
('comp-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'group-1111-0001-0001-000000000001', 'Database', 'PostgreSQL Database', 'operational', 1, NOW(), NOW()),
('comp-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'group-1111-0001-0001-000000000001', 'Redis Cache', 'Redis caching layer', 'operational', 2, NOW(), NOW()),
('comp-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'group-1111-0001-0001-000000000001', 'Message Queue', 'RabbitMQ', 'degraded_performance', 3, NOW(), NOW()),
-- Applications group
('comp-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'group-1111-0002-0002-000000000002', 'Web Application', 'Main web app', 'operational', 1, NOW(), NOW()),
('comp-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'group-1111-0002-0002-000000000002', 'Mobile App', 'iOS/Android app', 'partial_outage', 2, NOW(), NOW()),
-- APIs group
('comp-1111-0006-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'group-1111-0003-0003-000000000003', 'REST API', 'RESTful API', 'operational', 1, NOW(), NOW()),
('comp-1111-0007-0007-000000000007', 'tenant-1111-1111-1111-111111111111', 'group-1111-0003-0003-000000000003', 'GraphQL API', 'GraphQL endpoint', 'major_outage', 2, NOW(), NOW());

-- Insert test components for Tenant 2
INSERT INTO components (id, tenant_id, group_id, name, description, status, display_order, created_at, updated_at) VALUES
('comp-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'group-2222-0001-0001-000000000001', 'Payment Service', 'Payment processing', 'operational', 1, NOW(), NOW()),
('comp-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'group-2222-0001-0001-000000000001', 'User Service', 'User management', 'operational', 2, NOW(), NOW());

-- Insert component metrics
INSERT INTO component_metrics (id, component_id, metric_type, value, unit, timestamp, created_at) VALUES
-- Tenant 1 metrics
('metric-1111-0001-0001-000000000001', 'comp-1111-0001-0001-000000000001', 'uptime', 99.95, 'percent', NOW() - INTERVAL '1 hour', NOW()),
('metric-1111-0002-0002-000000000002', 'comp-1111-0001-0001-000000000001', 'response_time', 45.2, 'ms', NOW() - INTERVAL '1 hour', NOW()),
('metric-1111-0003-0003-000000000003', 'comp-1111-0002-0002-000000000002', 'uptime', 99.99, 'percent', NOW() - INTERVAL '1 hour', NOW()),
('metric-1111-0004-0004-000000000004', 'comp-1111-0002-0002-000000000002', 'hit_rate', 92.5, 'percent', NOW() - INTERVAL '1 hour', NOW()),
('metric-1111-0005-0005-000000000005', 'comp-1111-0003-0003-000000000003', 'throughput', 1250.0, 'msg/s', NOW() - INTERVAL '1 hour', NOW()),
-- Tenant 2 metrics
('metric-2222-0001-0001-000000000001', 'comp-2222-0001-0001-000000000001', 'uptime', 99.9, 'percent', NOW() - INTERVAL '1 hour', NOW()),
('metric-2222-0002-0002-000000000002', 'comp-2222-0002-0002-000000000002', 'response_time', 123.4, 'ms', NOW() - INTERVAL '1 hour', NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as component_groups FROM component_groups;
SELECT COUNT(*) as components FROM components;
SELECT COUNT(*) as metrics FROM component_metrics;
SELECT status, COUNT(*) as count FROM components GROUP BY status ORDER BY status;
