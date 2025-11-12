-- Test fixtures for saas_admin database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE plan_features, pricing_tiers, saas_features, saas_plans, platforms CASCADE;

-- Insert test platforms
INSERT INTO platforms (id, name, domain, status, created_at, updated_at) VALUES
('11111111-1111-1111-1111-111111111111', 'Test Platform 1', 'test1.example.com', 'active', NOW(), NOW()),
('22222222-2222-2222-2222-222222222222', 'Test Platform 2', 'test2.example.com', 'active', NOW(), NOW());

-- Insert test features
INSERT INTO saas_features (id, name, description, feature_key, category, created_at, updated_at) VALUES
('33333333-3333-3333-3333-333333333333', 'Basic Monitoring', 'Basic health monitoring', 'monitoring_basic', 'monitoring', NOW(), NOW()),
('44444444-4444-4444-4444-444444444444', 'Advanced Monitoring', 'Advanced monitoring with SSL checks', 'monitoring_advanced', 'monitoring', NOW(), NOW()),
('55555555-5555-5555-5555-555555555555', 'Email Notifications', 'Email notification support', 'notifications_email', 'notifications', NOW(), NOW()),
('66666666-6666-6666-6666-666666666666', 'Slack Notifications', 'Slack integration', 'notifications_slack', 'notifications', NOW(), NOW()),
('77777777-7777-7777-7777-777777777777', 'Custom Branding', 'Custom branding and themes', 'branding_custom', 'branding', NOW(), NOW());

-- Insert test plans
INSERT INTO saas_plans (id, name, description, plan_key, status, created_at, updated_at) VALUES
('88888888-8888-8888-8888-888888888888', 'Free Plan', 'Basic free tier', 'free', 'active', NOW(), NOW()),
('99999999-9999-9999-9999-999999999999', 'Starter Plan', 'Starter plan with basic features', 'starter', 'active', NOW(), NOW()),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Professional Plan', 'Professional plan with advanced features', 'professional', 'active', NOW(), NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Enterprise Plan', 'Enterprise plan with all features', 'enterprise', 'active', NOW(), NOW());

-- Insert pricing tiers
INSERT INTO pricing_tiers (id, plan_id, name, billing_cycle, price_cents, currency, max_users, max_components, max_incidents, created_at, updated_at) VALUES
-- Free plan
('cc000001-0001-0001-0001-000000000001', '88888888-8888-8888-8888-888888888888', 'Free Monthly', 'monthly', 0, 'USD', 1, 3, 10, NOW(), NOW()),
-- Starter plan
('cc000002-0002-0002-0002-000000000002', '99999999-9999-9999-9999-999999999999', 'Starter Monthly', 'monthly', 990, 'USD', 5, 10, 50, NOW(), NOW()),
('cc000003-0003-0003-0003-000000000003', '99999999-9999-9999-9999-999999999999', 'Starter Yearly', 'yearly', 9900, 'USD', 5, 10, 50, NOW(), NOW()),
-- Professional plan
('cc000004-0004-0004-0004-000000000004', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Pro Monthly', 'monthly', 4900, 'USD', 25, 50, 500, NOW(), NOW()),
('cc000005-0005-0005-0005-000000000005', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Pro Yearly', 'yearly', 49000, 'USD', 25, 50, 500, NOW(), NOW()),
-- Enterprise plan
('cc000006-0006-0006-0006-000000000006', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Enterprise Monthly', 'monthly', 19900, 'USD', NULL, NULL, NULL, NOW(), NOW()),
('cc000007-0007-0007-0007-000000000007', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Enterprise Yearly', 'yearly', 199000, 'USD', NULL, NULL, NULL, NOW(), NOW());

-- Associate features with plans
INSERT INTO plan_features (plan_id, feature_id, included, limit_value, created_at, updated_at) VALUES
-- Free plan features
('88888888-8888-8888-8888-888888888888', '33333333-3333-3333-3333-333333333333', true, NULL, NOW(), NOW()),
('88888888-8888-8888-8888-888888888888', '55555555-5555-5555-5555-555555555555', true, '10', NOW(), NOW()),
-- Starter plan features
('99999999-9999-9999-9999-999999999999', '33333333-3333-3333-3333-333333333333', true, NULL, NOW(), NOW()),
('99999999-9999-9999-9999-999999999999', '55555555-5555-5555-5555-555555555555', true, '100', NOW(), NOW()),
('99999999-9999-9999-9999-999999999999', '66666666-6666-6666-6666-666666666666', true, NULL, NOW(), NOW()),
-- Professional plan features
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '33333333-3333-3333-3333-333333333333', true, NULL, NOW(), NOW()),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '44444444-4444-4444-4444-444444444444', true, NULL, NOW(), NOW()),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '55555555-5555-5555-5555-555555555555', true, 'unlimited', NOW(), NOW()),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '66666666-6666-6666-6666-666666666666', true, NULL, NOW(), NOW()),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '77777777-7777-7777-7777-777777777777', true, NULL, NOW(), NOW()),
-- Enterprise plan features (all features)
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333', true, NULL, NOW(), NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444444', true, NULL, NOW(), NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '55555555-5555-5555-5555-555555555555', true, 'unlimited', NOW(), NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '66666666-6666-6666-6666-666666666666', true, NULL, NOW(), NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '77777777-7777-7777-7777-777777777777', true, NULL, NOW(), NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as platforms FROM platforms;
SELECT COUNT(*) as features FROM saas_features;
SELECT COUNT(*) as plans FROM saas_plans;
SELECT COUNT(*) as pricing_tiers FROM pricing_tiers;
SELECT COUNT(*) as plan_features FROM plan_features;
