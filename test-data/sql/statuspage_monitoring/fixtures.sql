-- Test fixtures for statuspage_monitoring database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE monitor_checks, ssl_certificates, alert_rules, alert_notifications, monitors CASCADE;

-- Insert monitors for Tenant 1
INSERT INTO monitors (id, tenant_id, component_id, monitor_type, name, description, target_url, check_interval_seconds, timeout_seconds, is_active, created_at, updated_at) VALUES
-- HTTP monitors
('mon-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0004-0004-000000000004', 'http', 'Web App Health Check', 'Main web application health endpoint', 'https://test1.example.com/health', 60, 30, true, NOW() - INTERVAL '6 months', NOW()),
('mon-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0006-0006-000000000006', 'http', 'REST API Health', 'REST API availability check', 'https://api.test1.example.com/health', 30, 15, true, NOW() - INTERVAL '6 months', NOW()),
('mon-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0007-0007-000000000007', 'http', 'GraphQL API Health', 'GraphQL endpoint health check', 'https://api.test1.example.com/graphql/health', 60, 20, true, NOW() - INTERVAL '3 months', NOW()),

-- Database monitors
('mon-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0001-0001-000000000001', 'database', 'Database Connection Check', 'PostgreSQL database connectivity', 'postgresql://db.test1.example.com:5432/prod', 300, 10, true, NOW() - INTERVAL '6 months', NOW()),

-- TCP port monitors
('mon-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0002-0002-000000000002', 'tcp', 'Redis Port Check', 'Redis cache port connectivity', 'redis://cache.test1.example.com:6379', 120, 5, true, NOW() - INTERVAL '6 months', NOW()),
('mon-1111-0006-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0003-0003-000000000003', 'tcp', 'RabbitMQ Port Check', 'Message queue port connectivity', 'amqp://mq.test1.example.com:5672', 180, 10, true, NOW() - INTERVAL '6 months', NOW()),

-- Ping monitors
('mon-1111-0007-0007-000000000007', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0004-0004-000000000004', 'ping', 'Web Server Ping', 'ICMP ping to web server', 'test1.example.com', 60, 5, true, NOW() - INTERVAL '6 months', NOW()),

-- Disabled monitor
('mon-1111-0008-0008-000000000008', 'tenant-1111-1111-1111-111111111111', 'comp-1111-0005-0005-000000000005', 'http', 'Mobile App API (Disabled)', 'Mobile app backend API', 'https://mobile-api.test1.example.com/health', 60, 30, false, NOW() - INTERVAL '3 months', NOW());

-- Insert monitors for Tenant 2
INSERT INTO monitors (id, tenant_id, component_id, monitor_type, name, description, target_url, check_interval_seconds, timeout_seconds, is_active, created_at, updated_at) VALUES
('mon-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'comp-2222-0001-0001-000000000001', 'http', 'Payment Service Health', 'Payment processing service', 'https://payments.test2.example.com/health', 30, 15, true, NOW() - INTERVAL '3 months', NOW()),
('mon-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'comp-2222-0002-0002-000000000002', 'http', 'User Service Health', 'User management service', 'https://users.test2.example.com/health', 60, 20, true, NOW() - INTERVAL '3 months', NOW());

-- Insert monitor checks (last 24 hours)
INSERT INTO monitor_checks (id, monitor_id, check_timestamp, status, response_time_ms, status_code, error_message, created_at) VALUES
-- Successful checks for mon-1111-0001 (Web App)
('check-1111-0001-0001-000000000001', 'mon-1111-0001-0001-000000000001', NOW() - INTERVAL '24 hours', 'success', 125.4, 200, NULL, NOW() - INTERVAL '24 hours'),
('check-1111-0001-0002-000000000002', 'mon-1111-0001-0001-000000000001', NOW() - INTERVAL '23 hours', 'success', 132.8, 200, NULL, NOW() - INTERVAL '23 hours'),
('check-1111-0001-0003-000000000003', 'mon-1111-0001-0001-000000000001', NOW() - INTERVAL '22 hours', 'success', 118.5, 200, NULL, NOW() - INTERVAL '22 hours'),
('check-1111-0001-0004-000000000004', 'mon-1111-0001-0001-000000000001', NOW() - INTERVAL '12 hours', 'success', 145.2, 200, NULL, NOW() - INTERVAL '12 hours'),
('check-1111-0001-0005-000000000005', 'mon-1111-0001-0001-000000000001', NOW() - INTERVAL '6 hours', 'success', 108.9, 200, NULL, NOW() - INTERVAL '6 hours'),
('check-1111-0001-0006-000000000006', 'mon-1111-0001-0001-000000000001', NOW() - INTERVAL '1 hour', 'success', 142.3, 200, NULL, NOW() - INTERVAL '1 hour'),

-- REST API checks with some failures
('check-1111-0002-0001-000000000001', 'mon-1111-0002-0002-000000000002', NOW() - INTERVAL '24 hours', 'success', 85.2, 200, NULL, NOW() - INTERVAL '24 hours'),
('check-1111-0002-0002-000000000002', 'mon-1111-0002-0002-000000000002', NOW() - INTERVAL '23 hours', 'success', 92.5, 200, NULL, NOW() - INTERVAL '23 hours'),
('check-1111-0002-0003-000000000003', 'mon-1111-0002-0002-000000000002', NOW() - INTERVAL '18 hours', 'failure', 0, 0, 'Connection timeout', NOW() - INTERVAL '18 hours'),
('check-1111-0002-0004-000000000004', 'mon-1111-0002-0002-000000000002', NOW() - INTERVAL '12 hours', 'success', 98.7, 200, NULL, NOW() - INTERVAL '12 hours'),
('check-1111-0002-0005-000000000005', 'mon-1111-0002-0002-000000000002', NOW() - INTERVAL '6 hours', 'success', 88.4, 200, NULL, NOW() - INTERVAL '6 hours'),
('check-1111-0002-0006-000000000006', 'mon-1111-0002-0002-000000000002', NOW() - INTERVAL '1 hour', 'success', 95.8, 200, NULL, NOW() - INTERVAL '1 hour'),

-- GraphQL API checks with degraded performance
('check-1111-0003-0001-000000000001', 'mon-1111-0003-0003-000000000003', NOW() - INTERVAL '24 hours', 'success', 152.3, 200, NULL, NOW() - INTERVAL '24 hours'),
('check-1111-0003-0002-000000000002', 'mon-1111-0003-0003-000000000003', NOW() - INTERVAL '12 hours', 'success', 185.6, 200, NULL, NOW() - INTERVAL '12 hours'),
('check-1111-0003-0003-000000000003', 'mon-1111-0003-0003-000000000003', NOW() - INTERVAL '6 hours', 'success', 425.8, 200, NULL, NOW() - INTERVAL '6 hours'),
('check-1111-0003-0004-000000000004', 'mon-1111-0003-0003-000000000003', NOW() - INTERVAL '1 hour', 'failure', 0, 502, 'Bad Gateway', NOW() - INTERVAL '1 hour'),

-- Database checks
('check-1111-0004-0001-000000000001', 'mon-1111-0004-0004-000000000004', NOW() - INTERVAL '24 hours', 'success', 15.2, NULL, NULL, NOW() - INTERVAL '24 hours'),
('check-1111-0004-0002-000000000002', 'mon-1111-0004-0004-000000000004', NOW() - INTERVAL '12 hours', 'success', 18.5, NULL, NULL, NOW() - INTERVAL '12 hours'),
('check-1111-0004-0003-000000000003', 'mon-1111-0004-0004-000000000004', NOW() - INTERVAL '6 hours', 'success', 12.8, NULL, NULL, NOW() - INTERVAL '6 hours'),

-- Redis checks
('check-1111-0005-0001-000000000001', 'mon-1111-0005-0005-000000000005', NOW() - INTERVAL '24 hours', 'success', 2.5, NULL, NULL, NOW() - INTERVAL '24 hours'),
('check-1111-0005-0002-000000000002', 'mon-1111-0005-0005-000000000005', NOW() - INTERVAL '12 hours', 'success', 3.1, NULL, NULL, NOW() - INTERVAL '12 hours'),
('check-1111-0005-0003-000000000003', 'mon-1111-0005-0005-000000000005', NOW() - INTERVAL '6 hours', 'success', 2.8, NULL, NULL, NOW() - INTERVAL '6 hours'),

-- Tenant 2 checks
('check-2222-0001-0001-000000000001', 'mon-2222-0001-0001-000000000001', NOW() - INTERVAL '24 hours', 'success', 98.5, 200, NULL, NOW() - INTERVAL '24 hours'),
('check-2222-0001-0002-000000000002', 'mon-2222-0001-0001-000000000001', NOW() - INTERVAL '12 hours', 'success', 105.2, 200, NULL, NOW() - INTERVAL '12 hours'),
('check-2222-0001-0003-000000000003', 'mon-2222-0001-0001-000000000001', NOW() - INTERVAL '1 hour', 'success', 92.8, 200, NULL, NOW() - INTERVAL '1 hour');

-- Insert SSL certificates
INSERT INTO ssl_certificates (id, tenant_id, domain, issuer, valid_from, valid_until, days_until_expiry, status, last_checked_at, created_at, updated_at) VALUES
-- Tenant 1 certificates
('ssl-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'test1.example.com', 'Let''s Encrypt Authority X3', NOW() - INTERVAL '3 months', NOW() + INTERVAL '3 months', 90, 'valid', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '3 months', NOW()),
('ssl-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'api.test1.example.com', 'Let''s Encrypt Authority X3', NOW() - INTERVAL '2 months', NOW() + INTERVAL '4 months', 120, 'valid', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '2 months', NOW()),
('ssl-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'old.test1.example.com', 'DigiCert Inc', NOW() - INTERVAL '12 months', NOW() + INTERVAL '15 days', 15, 'expiring_soon', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '12 months', NOW()),

-- Tenant 2 certificates
('ssl-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'test2.example.com', 'Let''s Encrypt Authority X3', NOW() - INTERVAL '1 month', NOW() + INTERVAL '5 months', 150, 'valid', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 month', NOW()),
('ssl-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'payments.test2.example.com', 'DigiCert Inc', NOW() - INTERVAL '6 months', NOW() + INTERVAL '6 months', 180, 'valid', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '6 months', NOW());

-- Insert alert rules
INSERT INTO alert_rules (id, tenant_id, monitor_id, rule_name, condition_type, threshold_value, consecutive_failures, alert_channels, is_active, created_at, updated_at) VALUES
-- Tenant 1 alert rules
('rule-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'mon-1111-0001-0001-000000000001', 'Web App Down Alert', 'failure', NULL, 3, '["chan-1111-0002-0002-000000000002", "chan-1111-0004-0004-000000000004"]', true, NOW() - INTERVAL '6 months', NOW()),
('rule-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'mon-1111-0001-0001-000000000001', 'Web App Slow Response', 'response_time', 500.0, 2, '["chan-1111-0002-0002-000000000002"]', true, NOW() - INTERVAL '6 months', NOW()),
('rule-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'mon-1111-0002-0002-000000000002', 'REST API Down Alert', 'failure', NULL, 2, '["chan-1111-0001-0001-000000000001", "chan-1111-0002-0002-000000000002"]', true, NOW() - INTERVAL '6 months', NOW()),
('rule-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'mon-1111-0003-0003-000000000003', 'GraphQL API Alert', 'failure', NULL, 2, '["chan-1111-0002-0002-000000000002"]', true, NOW() - INTERVAL '3 months', NOW()),
('rule-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'mon-1111-0004-0004-000000000004', 'Database Down Alert', 'failure', NULL, 1, '["chan-1111-0002-0002-000000000002", "chan-1111-0004-0004-000000000004"]', true, NOW() - INTERVAL '6 months', NOW()),

-- Tenant 2 alert rules
('rule-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'mon-2222-0001-0001-000000000001', 'Payment Service Alert', 'failure', NULL, 2, '["chan-2222-0001-0001-000000000001", "chan-2222-0002-0002-000000000002"]', true, NOW() - INTERVAL '3 months', NOW()),
('rule-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'mon-2222-0002-0002-000000000002', 'User Service Alert', 'failure', NULL, 3, '["chan-2222-0001-0001-000000000001"]', true, NOW() - INTERVAL '3 months', NOW());

-- Insert alert notifications
INSERT INTO alert_notifications (id, alert_rule_id, monitor_id, notification_type, status, message, triggered_at, resolved_at, created_at) VALUES
-- Recent GraphQL API failure alert (active)
('alert-1111-0001-0001-000000000001', 'rule-1111-0004-0004-000000000004', 'mon-1111-0003-0003-000000000003', 'failure', 'active', 'GraphQL API health check failed: Bad Gateway (502)', NOW() - INTERVAL '1 hour', NULL, NOW() - INTERVAL '1 hour'),

-- REST API failure from 18 hours ago (resolved)
('alert-1111-0002-0001-000000000001', 'rule-1111-0003-0003-000000000003', 'mon-1111-0002-0002-000000000002', 'failure', 'resolved', 'REST API health check failed: Connection timeout', NOW() - INTERVAL '18 hours', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '18 hours'),

-- SSL certificate expiring soon warning (active)
('alert-1111-0003-0001-000000000001', NULL, NULL, 'ssl_expiring', 'active', 'SSL certificate for old.test1.example.com expires in 15 days', NOW() - INTERVAL '12 hours', NULL, NOW() - INTERVAL '12 hours'),

-- Old resolved alerts from previous incidents
('alert-1111-0004-0001-000000000001', 'rule-1111-0001-0001-000000000001', 'mon-1111-0001-0001-000000000001', 'failure', 'resolved', 'Web App health check failed after 3 consecutive failures', NOW() - INTERVAL '7 days', NOW() - INTERVAL '6 days', NOW() - INTERVAL '7 days'),
('alert-1111-0005-0001-000000000001', 'rule-1111-0005-0005-000000000005', 'mon-1111-0004-0004-000000000004', 'failure', 'resolved', 'Database connection check failed', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as monitors FROM monitors;
SELECT COUNT(*) as monitor_checks FROM monitor_checks;
SELECT COUNT(*) as ssl_certificates FROM ssl_certificates;
SELECT COUNT(*) as alert_rules FROM alert_rules;
SELECT COUNT(*) as alert_notifications FROM alert_notifications;
SELECT monitor_type, COUNT(*) as count FROM monitors GROUP BY monitor_type ORDER BY monitor_type;
SELECT status, COUNT(*) as count FROM monitor_checks GROUP BY status ORDER BY status;
SELECT status, COUNT(*) as count FROM ssl_certificates GROUP BY status ORDER BY status;
SELECT status, COUNT(*) as count FROM alert_notifications GROUP BY status ORDER BY status;
