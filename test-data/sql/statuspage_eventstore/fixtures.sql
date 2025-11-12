-- Test fixtures for statuspage_eventstore database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE event_streams, events, event_snapshots CASCADE;

-- Insert event streams for Tenant 1
INSERT INTO event_streams (id, tenant_id, stream_type, aggregate_id, version, created_at, updated_at) VALUES
-- Incident event streams
('stream-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'incident', 'inc-1111-0001-0001-000000000001', 3, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '10 minutes'),
('stream-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'incident', 'inc-1111-0002-0002-000000000002', 4, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '2 hours'),
('stream-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'incident', 'inc-1111-0003-0003-000000000003', 1, NOW(), NOW()),
('stream-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'incident', 'inc-1111-0005-0005-000000000005', 2, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 minutes'),

-- Component event streams
('stream-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'component', 'comp-1111-0001-0001-000000000001', 5, NOW() - INTERVAL '6 months', NOW() - INTERVAL '30 minutes'),
('stream-1111-0006-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'component', 'comp-1111-0003-0003-000000000003', 3, NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 hour'),
('stream-1111-0007-0007-000000000007', 'tenant-1111-1111-1111-111111111111', 'component', 'comp-1111-0007-0007-000000000007', 2, NOW() - INTERVAL '3 months', NOW() - INTERVAL '1 hour'),

-- Subscription event streams
('stream-1111-0008-0008-000000000008', 'tenant-1111-1111-1111-111111111111', 'subscription', 'sub-1111-0001-0001-000000000001', 7, NOW() - INTERVAL '6 months', NOW() - INTERVAL '15 days'),

-- Tenant event streams
('stream-1111-0009-0009-000000000009', 'tenant-1111-1111-1111-111111111111', 'tenant', 'tenant-1111-1111-1111-111111111111', 2, NOW() - INTERVAL '12 months', NOW() - INTERVAL '3 months'),

-- Insert event streams for Tenant 2
INSERT INTO event_streams (id, tenant_id, stream_type, aggregate_id, version, created_at, updated_at) VALUES
('stream-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'incident', 'inc-2222-0001-0001-000000000001', 2, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '1 hour'),
('stream-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'component', 'comp-2222-0001-0001-000000000001', 4, NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 hours');

-- Insert events for incident inc-1111-0001 (active critical)
INSERT INTO events (id, stream_id, event_type, event_version, aggregate_id, aggregate_type, sequence_number, event_data, metadata, timestamp, created_at) VALUES
('event-1111-0001-0001-000000000001', 'stream-1111-0001-0001-000000000001', 'IncidentCreated', 1, 'inc-1111-0001-0001-000000000001', 'incident', 1,
'{"incident_id": "inc-1111-0001-0001-000000000001", "tenant_id": "tenant-1111-1111-1111-111111111111", "title": "Database Connection Failure", "description": "Primary database is not accepting connections", "severity": "critical", "status": "investigating", "created_by": "user-1111-0001-0001-000000000001"}',
'{"user_id": "user-1111-0001-0001-000000000001", "ip_address": "192.168.1.100", "user_agent": "Mozilla/5.0"}',
NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes'),

('event-1111-0001-0002-000000000002', 'stream-1111-0001-0001-000000000001', 'IncidentUpdated', 1, 'inc-1111-0001-0001-000000000001', 'incident', 2,
'{"incident_id": "inc-1111-0001-0001-000000000001", "status": "investigating", "message": "Database team has been engaged. We are working to restore connectivity.", "updated_by": "user-1111-0001-0001-000000000001"}',
'{"user_id": "user-1111-0001-0001-000000000001", "ip_address": "192.168.1.100"}',
NOW() - INTERVAL '20 minutes', NOW() - INTERVAL '20 minutes'),

('event-1111-0001-0003-000000000003', 'stream-1111-0001-0001-000000000001', 'IncidentUpdated', 1, 'inc-1111-0001-0001-000000000001', 'incident', 3,
'{"incident_id": "inc-1111-0001-0001-000000000001", "status": "investigating", "message": "Failover to secondary database in progress. ETA 10 minutes.", "updated_by": "user-1111-0002-0002-000000000002"}',
'{"user_id": "user-1111-0002-0002-000000000002", "ip_address": "192.168.1.105"}',
NOW() - INTERVAL '10 minutes', NOW() - INTERVAL '10 minutes');

-- Insert events for incident inc-1111-0002 (resolved major)
INSERT INTO events (id, stream_id, event_type, event_version, aggregate_id, aggregate_type, sequence_number, event_data, metadata, timestamp, created_at) VALUES
('event-1111-0002-0001-000000000001', 'stream-1111-0002-0002-000000000002', 'IncidentCreated', 1, 'inc-1111-0002-0002-000000000002', 'incident', 1,
'{"incident_id": "inc-1111-0002-0002-000000000002", "tenant_id": "tenant-1111-1111-1111-111111111111", "title": "API Response Time Degradation", "description": "API endpoints responding slowly", "severity": "major", "status": "investigating", "created_by": "user-1111-0002-0002-000000000002"}',
'{"user_id": "user-1111-0002-0002-000000000002"}',
NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours'),

('event-1111-0002-0002-000000000002', 'stream-1111-0002-0002-000000000002', 'IncidentUpdated', 1, 'inc-1111-0002-0002-000000000002', 'incident', 2,
'{"incident_id": "inc-1111-0002-0002-000000000002", "status": "identified", "message": "Identified database query optimization issue. Applying fixes."}',
'{"user_id": "user-1111-0002-0002-000000000002"}',
NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours'),

('event-1111-0002-0003-000000000003', 'stream-1111-0002-0002-000000000002', 'IncidentUpdated', 1, 'inc-1111-0002-0002-000000000002', 'incident', 3,
'{"incident_id": "inc-1111-0002-0002-000000000002", "status": "monitoring", "message": "Fixes deployed. Monitoring API response times."}',
'{"user_id": "user-1111-0001-0001-000000000001"}',
NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours'),

('event-1111-0002-0004-000000000004', 'stream-1111-0002-0002-000000000002', 'IncidentResolved', 1, 'inc-1111-0002-0002-000000000002', 'incident', 4,
'{"incident_id": "inc-1111-0002-0002-000000000002", "status": "resolved", "message": "API response times back to normal. Incident resolved.", "resolved_by": "user-1111-0001-0001-000000000001"}',
'{"user_id": "user-1111-0001-0001-000000000001"}',
NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours');

-- Insert events for scheduled maintenance inc-1111-0003
INSERT INTO events (id, stream_id, event_type, event_version, aggregate_id, aggregate_type, sequence_number, event_data, metadata, timestamp, created_at) VALUES
('event-1111-0003-0001-000000000001', 'stream-1111-0003-0003-000000000003', 'MaintenanceScheduled', 1, 'inc-1111-0003-0003-000000000003', 'incident', 1,
'{"incident_id": "inc-1111-0003-0003-000000000003", "tenant_id": "tenant-1111-1111-1111-111111111111", "title": "Scheduled Database Upgrade", "description": "Upgrading PostgreSQL to version 16", "severity": "maintenance", "scheduled_start": "' || (NOW() + INTERVAL '2 days')::text || '", "scheduled_duration": "2 hours", "created_by": "user-1111-0001-0001-000000000001"}',
'{"user_id": "user-1111-0001-0001-000000000001"}',
NOW(), NOW());

-- Insert events for component status changes
INSERT INTO events (id, stream_id, event_type, event_version, aggregate_id, aggregate_type, sequence_number, event_data, metadata, timestamp, created_at) VALUES
-- Database component events
('event-1111-0005-0001-000000000001', 'stream-1111-0005-0005-000000000005', 'ComponentCreated', 1, 'comp-1111-0001-0001-000000000001', 'component', 1,
'{"component_id": "comp-1111-0001-0001-000000000001", "tenant_id": "tenant-1111-1111-1111-111111111111", "name": "Database", "status": "operational"}',
'{}', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months'),

('event-1111-0005-0002-000000000002', 'stream-1111-0005-0005-000000000005', 'ComponentStatusChanged', 1, 'comp-1111-0001-0001-000000000001', 'component', 2,
'{"component_id": "comp-1111-0001-0001-000000000001", "previous_status": "operational", "new_status": "degraded_performance", "reason": "High query latency"}',
'{}', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours'),

('event-1111-0005-0003-000000000003', 'stream-1111-0005-0005-000000000005', 'ComponentStatusChanged', 1, 'comp-1111-0001-0001-000000000001', 'component', 3,
'{"component_id": "comp-1111-0001-0001-000000000001", "previous_status": "degraded_performance", "new_status": "operational", "reason": "Query optimization applied"}',
'{}', NOW() - INTERVAL '11 hours', NOW() - INTERVAL '11 hours'),

('event-1111-0005-0004-000000000004', 'stream-1111-0005-0005-000000000005', 'ComponentStatusChanged', 1, 'comp-1111-0001-0001-000000000001', 'component', 4,
'{"component_id": "comp-1111-0001-0001-000000000001", "previous_status": "operational", "new_status": "major_outage", "reason": "Connection failure"}',
'{}', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes'),

('event-1111-0005-0005-000000000005', 'stream-1111-0005-0005-000000000005', 'ComponentMetricsUpdated', 1, 'comp-1111-0001-0001-000000000001', 'component', 5,
'{"component_id": "comp-1111-0001-0001-000000000001", "metrics": {"uptime": 99.95, "response_time_ms": 45.2}}',
'{}', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour'),

-- Message Queue component events
('event-1111-0006-0001-000000000001', 'stream-1111-0006-0006-000000000006', 'ComponentCreated', 1, 'comp-1111-0003-0003-000000000003', 'component', 1,
'{"component_id": "comp-1111-0003-0003-000000000003", "tenant_id": "tenant-1111-1111-1111-111111111111", "name": "Message Queue", "status": "operational"}',
'{}', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months'),

('event-1111-0006-0002-000000000002', 'stream-1111-0006-0006-000000000006', 'ComponentStatusChanged', 1, 'comp-1111-0003-0003-000000000003', 'component', 2,
'{"component_id": "comp-1111-0003-0003-000000000003", "previous_status": "operational", "new_status": "degraded_performance", "reason": "High message queue depth"}',
'{}', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),

('event-1111-0006-0003-000000000003', 'stream-1111-0006-0006-000000000006', 'ComponentMetricsUpdated', 1, 'comp-1111-0003-0003-000000000003', 'component', 3,
'{"component_id": "comp-1111-0003-0003-000000000003", "metrics": {"throughput": 1250.0, "queue_depth": 8500}}',
'{}', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour');

-- Insert subscription events
INSERT INTO events (id, stream_id, event_type, event_version, aggregate_id, aggregate_type, sequence_number, event_data, metadata, timestamp, created_at) VALUES
('event-1111-0008-0001-000000000001', 'stream-1111-0008-0008-000000000008', 'SubscriptionCreated', 1, 'sub-1111-0001-0001-000000000001', 'subscription', 1,
'{"subscription_id": "sub-1111-0001-0001-000000000001", "tenant_id": "tenant-1111-1111-1111-111111111111", "plan_id": "99999999-9999-9999-9999-999999999999", "status": "active"}',
'{}', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months'),

('event-1111-0008-0002-000000000002', 'stream-1111-0008-0008-000000000008', 'PaymentSucceeded', 1, 'sub-1111-0001-0001-000000000001', 'subscription', 2,
'{"subscription_id": "sub-1111-0001-0001-000000000001", "amount": 31.90, "currency": "USD", "invoice_id": "inv-1111-0001-0001-000000000001"}',
'{}', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months'),

('event-1111-0008-0007-000000000007', 'stream-1111-0008-0008-000000000008', 'BillingPeriodRenewed', 1, 'sub-1111-0001-0001-000000000001', 'subscription', 7,
'{"subscription_id": "sub-1111-0001-0001-000000000001", "previous_period_end": "' || (NOW() - INTERVAL '15 days')::text || '", "new_period_end": "' || (NOW() + INTERVAL '15 days')::text || '"}',
'{}', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days');

-- Insert Tenant 2 events
INSERT INTO events (id, stream_id, event_type, event_version, aggregate_id, aggregate_type, sequence_number, event_data, metadata, timestamp, created_at) VALUES
('event-2222-0001-0001-000000000001', 'stream-2222-0001-0001-000000000001', 'IncidentCreated', 1, 'inc-2222-0001-0001-000000000001', 'incident', 1,
'{"incident_id": "inc-2222-0001-0001-000000000001", "tenant_id": "tenant-2222-2222-2222-222222222222", "title": "CDN Configuration Error", "severity": "major", "status": "investigating"}',
'{}', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours'),

('event-2222-0001-0002-000000000002', 'stream-2222-0001-0001-000000000001', 'IncidentResolved', 1, 'inc-2222-0001-0001-000000000001', 'incident', 2,
'{"incident_id": "inc-2222-0001-0001-000000000001", "status": "resolved", "message": "CDN configuration corrected."}',
'{}', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour');

-- Insert event snapshots (for performance optimization)
INSERT INTO event_snapshots (id, stream_id, aggregate_id, aggregate_type, snapshot_version, snapshot_data, created_at) VALUES
-- Snapshot for incident inc-1111-0002 (resolved)
('snap-1111-0001-0001-000000000001', 'stream-1111-0002-0002-000000000002', 'inc-1111-0002-0002-000000000002', 'incident', 4,
'{"incident_id": "inc-1111-0002-0002-000000000002", "tenant_id": "tenant-1111-1111-1111-111111111111", "title": "API Response Time Degradation", "status": "resolved", "severity": "major", "created_at": "' || (NOW() - INTERVAL '5 hours')::text || '", "resolved_at": "' || (NOW() - INTERVAL '2 hours')::text || '"}',
NOW() - INTERVAL '2 hours'),

-- Snapshot for component comp-1111-0001
('snap-1111-0002-0001-000000000001', 'stream-1111-0005-0005-000000000005', 'comp-1111-0001-0001-000000000001', 'component', 5,
'{"component_id": "comp-1111-0001-0001-000000000001", "tenant_id": "tenant-1111-1111-1111-111111111111", "name": "Database", "status": "major_outage", "metrics": {"uptime": 99.95, "response_time_ms": 45.2}}',
NOW() - INTERVAL '1 hour'),

-- Snapshot for subscription
('snap-1111-0003-0001-000000000001', 'stream-1111-0008-0008-000000000008', 'sub-1111-0001-0001-000000000001', 'subscription', 7,
'{"subscription_id": "sub-1111-0001-0001-000000000001", "tenant_id": "tenant-1111-1111-1111-111111111111", "plan_id": "99999999-9999-9999-9999-999999999999", "status": "active", "current_period_end": "' || (NOW() + INTERVAL '15 days')::text || '"}',
NOW() - INTERVAL '15 days');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as event_streams FROM event_streams;
SELECT COUNT(*) as events FROM events;
SELECT COUNT(*) as snapshots FROM event_snapshots;
SELECT stream_type, COUNT(*) as count FROM event_streams GROUP BY stream_type ORDER BY stream_type;
SELECT event_type, COUNT(*) as count FROM events GROUP BY event_type ORDER BY event_type;
SELECT aggregate_type, COUNT(*) as count FROM events GROUP BY aggregate_type ORDER BY aggregate_type;
