-- Test fixtures for statuspage_audit database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE audit_logs CASCADE;

-- Insert audit logs for Tenant 1
INSERT INTO audit_logs (id, tenant_id, user_id, action, resource_type, resource_id, ip_address, user_agent, changes, metadata, created_at) VALUES
-- Recent incident creation and updates
('audit-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'incident.created', 'incident', 'inc-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"id": "inc-1111-0001-0001-000000000001", "title": "Database Connection Failure", "status": "investigating", "severity": "critical"}}',
'{"correlation_id": "corr-1111-0001", "request_id": "req-1111-0001"}',
NOW() - INTERVAL '30 minutes'),

('audit-1111-0001-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'incident.updated', 'incident', 'inc-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"status": "investigating"}, "after": {"status": "investigating", "update_message": "Database team has been engaged"}}',
'{"correlation_id": "corr-1111-0002", "request_id": "req-1111-0002"}',
NOW() - INTERVAL '20 minutes'),

('audit-1111-0001-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'user-1111-0002-0002-000000000002', 'incident.updated', 'incident', 'inc-1111-0001-0001-000000000001', '192.168.1.105', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
'{"before": {"status": "investigating"}, "after": {"status": "investigating", "update_message": "Failover to secondary database in progress"}}',
'{"correlation_id": "corr-1111-0003", "request_id": "req-1111-0003"}',
NOW() - INTERVAL '10 minutes'),

-- Component status changes
('audit-1111-0002-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'component.status_changed', 'component', 'comp-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"status": "operational"}, "after": {"status": "major_outage", "reason": "Connection failure"}}',
'{"correlation_id": "corr-1111-0004", "automated": true}',
NOW() - INTERVAL '30 minutes'),

('audit-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0002-0002-000000000002', 'component.status_changed', 'component', 'comp-1111-0003-0003-000000000003', '192.168.1.105', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
'{"before": {"status": "operational"}, "after": {"status": "degraded_performance", "reason": "High message queue depth"}}',
'{"correlation_id": "corr-1111-0005"}',
NOW() - INTERVAL '2 hours'),

-- User management actions
('audit-1111-0003-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'user.created', 'user', 'user-1111-0003-0003-000000000003', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"id": "user-1111-0003-0003-000000000003", "email": "viewer@test1.com", "role": "viewer", "status": "active"}}',
'{"correlation_id": "corr-1111-0010"}',
NOW() - INTERVAL '7 days'),

('audit-1111-0003-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'user.role_changed', 'user', 'user-1111-0002-0002-000000000002', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"role": "viewer"}, "after": {"role": "admin"}}',
'{"correlation_id": "corr-1111-0011"}',
NOW() - INTERVAL '14 days'),

('audit-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'user.deleted', 'user', 'user-1111-0099-0099-000000000099', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"id": "user-1111-0099-0099-000000000099", "email": "old@test1.com", "status": "active"}, "after": null}',
'{"correlation_id": "corr-1111-0012", "reason": "User account closure requested"}',
NOW() - INTERVAL '30 days'),

-- Configuration changes
('audit-1111-0004-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'notification_channel.created', 'notification_channel', 'chan-1111-0002-0002-000000000002', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"id": "chan-1111-0002-0002-000000000002", "type": "slack", "name": "Slack #incidents"}}',
'{"correlation_id": "corr-1111-0020"}',
NOW() - INTERVAL '3 months'),

('audit-1111-0004-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'notification_rule.updated', 'notification_rule', 'rule-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"channels": ["chan-1111-0001-0001-000000000001"]}, "after": {"channels": ["chan-1111-0001-0001-000000000001", "chan-1111-0002-0002-000000000002", "chan-1111-0004-0004-000000000004"]}}',
'{"correlation_id": "corr-1111-0021"}',
NOW() - INTERVAL '3 months'),

-- Subscription and billing actions
('audit-1111-0005-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'subscription.upgraded', 'subscription', 'sub-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"plan": "free"}, "after": {"plan": "starter", "effective_date": "' || (NOW() - INTERVAL '6 months')::text || '"}}',
'{"correlation_id": "corr-1111-0030"}',
NOW() - INTERVAL '6 months'),

('audit-1111-0005-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'payment.succeeded', 'payment', 'txn-1111-0006-0006-000000000006', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"transaction_id": "txn-1111-0006-0006-000000000006", "amount": 31.90, "currency": "USD", "status": "succeeded"}}',
'{"correlation_id": "corr-1111-0031", "invoice_id": "inv-1111-0006-0006-000000000006"}',
NOW() - INTERVAL '1 month'),

-- Authentication events
('audit-1111-0006-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'auth.login', 'session', 'session-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{}',
'{"correlation_id": "corr-1111-0040", "login_method": "email_password"}',
NOW() - INTERVAL '2 hours'),

('audit-1111-0006-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0002-0002-000000000002', 'auth.login_failed', 'session', NULL, '192.168.1.105', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
'{}',
'{"correlation_id": "corr-1111-0041", "reason": "invalid_password", "attempts": 3}',
NOW() - INTERVAL '1 day'),

('audit-1111-0006-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'auth.password_changed', 'user', 'user-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{}',
'{"correlation_id": "corr-1111-0042"}',
NOW() - INTERVAL '15 days'),

-- Integration actions
('audit-1111-0007-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'integration.connected', 'integration', 'int-slack-1111-0001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"type": "slack", "workspace": "test1-workspace", "channel": "#incidents"}}',
'{"correlation_id": "corr-1111-0050"}',
NOW() - INTERVAL '3 months'),

('audit-1111-0007-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'integration.disconnected', 'integration', 'int-webhook-old', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"type": "webhook", "url": "https://old-webhook.test1.com"}, "after": null}',
'{"correlation_id": "corr-1111-0051"}',
NOW() - INTERVAL '2 months'),

-- Status page configuration
('audit-1111-0008-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'status_page.published', 'status_page', 'page-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"is_published": false}, "after": {"is_published": true, "published_at": "' || (NOW() - INTERVAL '6 months')::text || '"}}',
'{"correlation_id": "corr-1111-0060"}',
NOW() - INTERVAL '6 months'),

('audit-1111-0008-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'user-1111-0001-0001-000000000001', 'status_page.updated', 'status_page', 'page-1111-0001-0001-000000000001', '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"show_uptime": false}, "after": {"show_uptime": true, "show_incident_history": true}}',
'{"correlation_id": "corr-1111-0061"}',
NOW() - INTERVAL '1 week');

-- Insert audit logs for Tenant 2
INSERT INTO audit_logs (id, tenant_id, user_id, action, resource_type, resource_id, ip_address, user_agent, changes, metadata, created_at) VALUES
-- Recent incident
('audit-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'user-2222-0001-0001-000000000001', 'incident.created', 'incident', 'inc-2222-0001-0001-000000000001', '10.0.1.50', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"id": "inc-2222-0001-0001-000000000001", "title": "CDN Configuration Error", "severity": "major", "status": "investigating"}}',
'{"correlation_id": "corr-2222-0001"}',
NOW() - INTERVAL '3 hours'),

('audit-2222-0001-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'user-2222-0001-0001-000000000001', 'incident.resolved', 'incident', 'inc-2222-0001-0001-000000000001', '10.0.1.50', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": {"status": "investigating"}, "after": {"status": "resolved", "resolved_at": "' || (NOW() - INTERVAL '1 hour')::text || '"}}',
'{"correlation_id": "corr-2222-0002"}',
NOW() - INTERVAL '1 hour'),

-- User management
('audit-2222-0002-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'user-2222-0001-0001-000000000001', 'user.created', 'user', 'user-2222-0002-0002-000000000002', '10.0.1.50', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"id": "user-2222-0002-0002-000000000002", "email": "admin@test2.com", "role": "admin"}}',
'{"correlation_id": "corr-2222-0010"}',
NOW() - INTERVAL '2 months'),

-- Subscription
('audit-2222-0003-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'user-2222-0001-0001-000000000001', 'subscription.created', 'subscription', 'sub-2222-0001-0001-000000000001', '10.0.1.50', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{"before": null, "after": {"plan": "professional", "billing_cycle": "yearly", "status": "active"}}',
'{"correlation_id": "corr-2222-0020"}',
NOW() - INTERVAL '3 months'),

-- Auth
('audit-2222-0004-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'user-2222-0001-0001-000000000001', 'auth.login', 'session', 'session-2222-0001', '10.0.1.50', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
'{}',
'{"correlation_id": "corr-2222-0030", "login_method": "email_password"}',
NOW() - INTERVAL '1 hour');

-- Insert audit logs for Tenant 3 (trial tenant - minimal activity)
INSERT INTO audit_logs (id, tenant_id, user_id, action, resource_type, resource_id, ip_address, user_agent, changes, metadata, created_at) VALUES
('audit-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'user-3333-0001-0001-000000000001', 'tenant.created', 'tenant', 'tenant-3333-3333-3333-333333333333', '203.0.113.45', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
'{"before": null, "after": {"id": "tenant-3333-3333-3333-333333333333", "name": "Test Tenant 3", "status": "active"}}',
'{"correlation_id": "corr-3333-0001", "signup_source": "website"}',
NOW()),

('audit-3333-0001-0002-000000000002', 'tenant-3333-3333-3333-333333333333', 'user-3333-0001-0001-000000000001', 'user.created', 'user', 'user-3333-0001-0001-000000000001', '203.0.113.45', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
'{"before": null, "after": {"id": "user-3333-0001-0001-000000000001", "email": "owner@test3.com", "role": "owner"}}',
'{"correlation_id": "corr-3333-0002"}',
NOW()),

('audit-3333-0001-0003-000000000003', 'tenant-3333-3333-3333-333333333333', 'user-3333-0001-0001-000000000001', 'auth.login', 'session', 'session-3333-0001', '203.0.113.45', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
'{}',
'{"correlation_id": "corr-3333-0003", "login_method": "email_password", "first_login": true}',
NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as total_audit_logs FROM audit_logs;
SELECT tenant_id, COUNT(*) as log_count FROM audit_logs GROUP BY tenant_id ORDER BY tenant_id;
SELECT action, COUNT(*) as count FROM audit_logs GROUP BY action ORDER BY count DESC LIMIT 10;
SELECT resource_type, COUNT(*) as count FROM audit_logs GROUP BY resource_type ORDER BY count DESC;
SELECT DATE(created_at) as date, COUNT(*) as count FROM audit_logs WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY DATE(created_at) ORDER BY date DESC;
