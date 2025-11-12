-- Test fixtures for statuspage_notification database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE notification_delivery_logs, notification_templates, notification_channels, notification_rules, notifications CASCADE;

-- Insert notification channels for Tenant 1
INSERT INTO notification_channels (id, tenant_id, channel_type, channel_name, configuration, is_active, created_at, updated_at) VALUES
('chan-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'email', 'Email Notifications', '{"smtp_host": "smtp.example.com", "from_address": "noreply@test1.com"}', true, NOW(), NOW()),
('chan-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'slack', 'Slack #incidents', '{"webhook_url": "https://hooks.slack.com/services/TEST1", "channel": "#incidents"}', true, NOW(), NOW()),
('chan-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'webhook', 'Custom Webhook', '{"url": "https://api.test1.com/webhooks/incidents", "method": "POST"}', true, NOW(), NOW()),
('chan-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'sms', 'Twilio SMS', '{"account_sid": "ACtest1", "from_number": "+15551234567"}', true, NOW(), NOW()),
('chan-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'discord', 'Discord #alerts', '{"webhook_url": "https://discord.com/api/webhooks/TEST1"}', false, NOW(), NOW());

-- Insert notification channels for Tenant 2
INSERT INTO notification_channels (id, tenant_id, channel_type, channel_name, configuration, is_active, created_at, updated_at) VALUES
('chan-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'email', 'Email Alerts', '{"smtp_host": "smtp.example.com", "from_address": "alerts@test2.com"}', true, NOW(), NOW()),
('chan-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'teams', 'Microsoft Teams', '{"webhook_url": "https://outlook.office.com/webhook/TEST2"}', true, NOW(), NOW());

-- Insert notification templates
INSERT INTO notification_templates (id, tenant_id, template_name, template_type, subject, body, created_at, updated_at) VALUES
('tmpl-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Incident Created', 'incident_created', 'New Incident: {{incident.title}}', 'A new incident has been created:\n\nTitle: {{incident.title}}\nSeverity: {{incident.severity}}\nStatus: {{incident.status}}\n\nDescription: {{incident.description}}', NOW(), NOW()),
('tmpl-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'Incident Updated', 'incident_updated', 'Incident Update: {{incident.title}}', 'Incident has been updated:\n\nTitle: {{incident.title}}\nStatus: {{incident.status}}\n\nUpdate: {{update.message}}', NOW(), NOW()),
('tmpl-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'Incident Resolved', 'incident_resolved', 'Resolved: {{incident.title}}', 'The incident has been resolved:\n\nTitle: {{incident.title}}\nDuration: {{incident.duration}}\n\nResolution: {{update.message}}', NOW(), NOW()),
('tmpl-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'Component Status Change', 'component_status', 'Component Status: {{component.name}}', 'Component status has changed:\n\nComponent: {{component.name}}\nNew Status: {{component.status}}\nPrevious Status: {{component.previous_status}}', NOW(), NOW()),
('tmpl-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'Alert Notification', 'alert', 'Alert: {{alert.title}}', 'Alert triggered:\n\n{{alert.message}}', NOW(), NOW());

-- Insert notification rules
INSERT INTO notification_rules (id, tenant_id, rule_name, trigger_type, conditions, channel_ids, is_active, created_at, updated_at) VALUES
('rule-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Critical Incidents', 'incident_created', '{"severity": "critical"}', '["chan-1111-0001-0001-000000000001", "chan-1111-0002-0002-000000000002", "chan-1111-0004-0004-000000000004"]', true, NOW(), NOW()),
('rule-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'Major Incidents', 'incident_created', '{"severity": "major"}', '["chan-1111-0001-0001-000000000001", "chan-1111-0002-0002-000000000002"]', true, NOW(), NOW()),
('rule-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'All Updates', 'incident_updated', '{}', '["chan-1111-0001-0001-000000000001"]', true, NOW(), NOW()),
('rule-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'Component Outages', 'component_status', '{"status": ["major_outage", "partial_outage"]}', '["chan-1111-0002-0002-000000000002", "chan-1111-0003-0003-000000000003"]', true, NOW(), NOW()),
('rule-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'All Incidents', 'incident_created', '{}', '["chan-2222-0001-0001-000000000001", "chan-2222-0002-0002-000000000002"]', true, NOW(), NOW());

-- Insert notifications
INSERT INTO notifications (id, tenant_id, notification_type, reference_id, reference_type, priority, subject, message, created_at) VALUES
-- Tenant 1 notifications
('notif-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'incident_created', 'inc-1111-0001-0001-000000000001', 'incident', 'high', 'New Incident: Database Connection Failure', 'A new critical incident has been created...', NOW() - INTERVAL '30 minutes'),
('notif-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'incident_updated', 'inc-1111-0001-0001-000000000001', 'incident', 'high', 'Incident Update: Database Connection Failure', 'Database team has been engaged...', NOW() - INTERVAL '20 minutes'),
('notif-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'incident_updated', 'inc-1111-0001-0001-000000000001', 'incident', 'high', 'Incident Update: Database Connection Failure', 'Failover to secondary database in progress...', NOW() - INTERVAL '10 minutes'),
('notif-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'incident_resolved', 'inc-1111-0002-0002-000000000002', 'incident', 'medium', 'Resolved: API Response Time Degradation', 'The incident has been resolved...', NOW() - INTERVAL '2 hours'),
('notif-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'component_status', 'comp-1111-0003-0003-000000000003', 'component', 'medium', 'Component Status: Message Queue', 'Component status changed to degraded_performance...', NOW() - INTERVAL '1 hour'),
-- Tenant 2 notifications
('notif-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'incident_created', 'inc-2222-0001-0001-000000000001', 'incident', 'high', 'New Incident: CDN Configuration Error', 'A new major incident has been created...', NOW() - INTERVAL '3 hours'),
('notif-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'incident_resolved', 'inc-2222-0001-0001-000000000001', 'incident', 'high', 'Resolved: CDN Configuration Error', 'The incident has been resolved...', NOW() - INTERVAL '1 hour');

-- Insert delivery logs
INSERT INTO notification_delivery_logs (id, notification_id, channel_id, delivery_status, attempts, sent_at, delivered_at, error_message, created_at) VALUES
-- Successful deliveries for notif-1111-0001
('log-1111-0001-0001-000000000001', 'notif-1111-0001-0001-000000000001', 'chan-1111-0001-0001-000000000001', 'delivered', 1, NOW() - INTERVAL '29 minutes', NOW() - INTERVAL '29 minutes', NULL, NOW() - INTERVAL '29 minutes'),
('log-1111-0001-0002-000000000002', 'notif-1111-0001-0001-000000000001', 'chan-1111-0002-0002-000000000002', 'delivered', 1, NOW() - INTERVAL '29 minutes', NOW() - INTERVAL '29 minutes', NULL, NOW() - INTERVAL '29 minutes'),
('log-1111-0001-0003-000000000003', 'notif-1111-0001-0001-000000000001', 'chan-1111-0004-0004-000000000004', 'delivered', 2, NOW() - INTERVAL '29 minutes', NOW() - INTERVAL '28 minutes', NULL, NOW() - INTERVAL '28 minutes'),

-- Successful deliveries for notif-1111-0002
('log-1111-0002-0001-000000000001', 'notif-1111-0002-0002-000000000002', 'chan-1111-0001-0001-000000000001', 'delivered', 1, NOW() - INTERVAL '19 minutes', NOW() - INTERVAL '19 minutes', NULL, NOW() - INTERVAL '19 minutes'),
('log-1111-0002-0002-000000000002', 'notif-1111-0002-0002-000000000002', 'chan-1111-0002-0002-000000000002', 'delivered', 1, NOW() - INTERVAL '19 minutes', NOW() - INTERVAL '19 minutes', NULL, NOW() - INTERVAL '19 minutes'),

-- Failed delivery with retry
('log-1111-0003-0001-000000000001', 'notif-1111-0003-0003-000000000003', 'chan-1111-0003-0003-000000000003', 'failed', 3, NOW() - INTERVAL '10 minutes', NULL, 'Connection timeout after 3 attempts', NOW() - INTERVAL '9 minutes'),

-- Successful deliveries for notif-1111-0004
('log-1111-0004-0001-000000000001', 'notif-1111-0004-0004-000000000004', 'chan-1111-0001-0001-000000000001', 'delivered', 1, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NULL, NOW() - INTERVAL '2 hours'),

-- Pending delivery
('log-1111-0005-0001-000000000001', 'notif-1111-0005-0005-000000000005', 'chan-1111-0002-0002-000000000002', 'pending', 0, NULL, NULL, NULL, NOW() - INTERVAL '1 hour'),

-- Tenant 2 deliveries
('log-2222-0001-0001-000000000001', 'notif-2222-0001-0001-000000000001', 'chan-2222-0001-0001-000000000001', 'delivered', 1, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NULL, NOW() - INTERVAL '3 hours'),
('log-2222-0001-0002-000000000002', 'notif-2222-0001-0001-000000000001', 'chan-2222-0002-0002-000000000002', 'delivered', 1, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NULL, NOW() - INTERVAL '3 hours'),
('log-2222-0002-0001-000000000001', 'notif-2222-0002-0002-000000000002', 'chan-2222-0001-0001-000000000001', 'delivered', 1, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NULL, NOW() - INTERVAL '1 hour');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as channels FROM notification_channels;
SELECT COUNT(*) as templates FROM notification_templates;
SELECT COUNT(*) as rules FROM notification_rules;
SELECT COUNT(*) as notifications FROM notifications;
SELECT COUNT(*) as delivery_logs FROM notification_delivery_logs;
SELECT delivery_status, COUNT(*) as count FROM notification_delivery_logs GROUP BY delivery_status ORDER BY delivery_status;
SELECT channel_type, COUNT(*) as count FROM notification_channels GROUP BY channel_type ORDER BY channel_type;
