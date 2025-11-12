-- Test fixtures for tenant_admin_db database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE user_sessions, user_roles, user_teams, team_members, teams, roles, users, tenants CASCADE;

-- Insert test tenants
INSERT INTO tenants (id, name, subdomain, plan_id, status, max_users, created_at, updated_at) VALUES
('tenant-1111-1111-1111-111111111111', 'Test Tenant 1', 'test1', '99999999-9999-9999-9999-999999999999', 'active', 5, NOW(), NOW()),
('tenant-2222-2222-2222-222222222222', 'Test Tenant 2', 'test2', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'active', 25, NOW(), NOW()),
('tenant-3333-3333-3333-333333333333', 'Test Tenant 3', 'test3', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'active', NULL, NOW(), NOW());

-- Insert test roles
INSERT INTO roles (id, tenant_id, name, description, permissions, created_at, updated_at) VALUES
-- Tenant 1 roles
('role-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'owner', 'Tenant Owner', '{"all": true}', NOW(), NOW()),
('role-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'admin', 'Administrator', '{"components": "write", "incidents": "write", "users": "write"}', NOW(), NOW()),
('role-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'viewer', 'Viewer', '{"components": "read", "incidents": "read"}', NOW(), NOW()),
-- Tenant 2 roles
('role-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'owner', 'Tenant Owner', '{"all": true}', NOW(), NOW()),
('role-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'admin', 'Administrator', '{"components": "write", "incidents": "write", "users": "write"}', NOW(), NOW()),
-- Tenant 3 roles
('role-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'owner', 'Tenant Owner', '{"all": true}', NOW(), NOW());

-- Insert test users
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status, email_verified, created_at, updated_at) VALUES
-- Tenant 1 users (using bcrypt hash for 'password123')
('user-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'owner@test1.com', '$2a$10$YourHashHere', 'John', 'Owner', 'active', true, NOW(), NOW()),
('user-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'admin@test1.com', '$2a$10$YourHashHere', 'Jane', 'Admin', 'active', true, NOW(), NOW()),
('user-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'viewer@test1.com', '$2a$10$YourHashHere', 'Bob', 'Viewer', 'active', true, NOW(), NOW()),
-- Tenant 2 users
('user-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'owner@test2.com', '$2a$10$YourHashHere', 'Alice', 'Owner', 'active', true, NOW(), NOW()),
('user-2222-0002-0002-000000000002', 'tenant-2222-2222-2222-222222222222', 'admin@test2.com', '$2a$10$YourHashHere', 'Charlie', 'Admin', 'active', true, NOW(), NOW()),
-- Tenant 3 users
('user-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'owner@test3.com', '$2a$10$YourHashHere', 'David', 'Owner', 'active', true, NOW(), NOW());

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id, created_at) VALUES
-- Tenant 1
('user-1111-0001-0001-000000000001', 'role-1111-0001-0001-000000000001', NOW()),
('user-1111-0002-0002-000000000002', 'role-1111-0002-0002-000000000002', NOW()),
('user-1111-0003-0003-000000000003', 'role-1111-0003-0003-000000000003', NOW()),
-- Tenant 2
('user-2222-0001-0001-000000000001', 'role-2222-0001-0001-000000000001', NOW()),
('user-2222-0002-0002-000000000002', 'role-2222-0002-0002-000000000002', NOW()),
-- Tenant 3
('user-3333-0001-0001-000000000001', 'role-3333-0001-0001-000000000001', NOW());

-- Insert test teams
INSERT INTO teams (id, tenant_id, name, description, created_at, updated_at) VALUES
('team-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Operations Team', 'Operations and monitoring', NOW(), NOW()),
('team-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'Development Team', 'Development and deployments', NOW(), NOW()),
('team-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'Platform Team', 'Platform operations', NOW(), NOW());

-- Assign users to teams
INSERT INTO team_members (team_id, user_id, created_at) VALUES
-- Tenant 1 teams
('team-1111-0001-0001-000000000001', 'user-1111-0001-0001-000000000001', NOW()),
('team-1111-0001-0001-000000000001', 'user-1111-0002-0002-000000000002', NOW()),
('team-1111-0002-0002-000000000002', 'user-1111-0002-0002-000000000002', NOW()),
('team-1111-0002-0002-000000000002', 'user-1111-0003-0003-000000000003', NOW()),
-- Tenant 2 teams
('team-2222-0001-0001-000000000001', 'user-2222-0001-0001-000000000001', NOW()),
('team-2222-0001-0001-000000000001', 'user-2222-0002-0002-000000000002', NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as tenants FROM tenants;
SELECT COUNT(*) as users FROM users;
SELECT COUNT(*) as roles FROM roles;
SELECT COUNT(*) as teams FROM teams;
SELECT COUNT(*) as user_roles FROM user_roles;
SELECT COUNT(*) as team_members FROM team_members;
