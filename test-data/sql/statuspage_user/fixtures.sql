-- Test fixtures for statuspage_user database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE user_sessions, users CASCADE;

-- Insert test users
-- Password hash is bcrypt hash for 'password123'
INSERT INTO users (id, email, password_hash, first_name, last_name, is_verified, is_active, created_at, updated_at) VALUES
('11111111-1111-1111-1111-111111111111', 'test1@example.com', '$2a$10$YourBcryptHashHereForTesting123456789012345678901234567890', 'Test', 'User1', true, true, NOW(), NOW()),
('22222222-2222-2222-2222-222222222222', 'test2@example.com', '$2a$10$YourBcryptHashHereForTesting123456789012345678901234567890', 'Test', 'User2', true, true, NOW(), NOW()),
('33333333-3333-3333-3333-333333333333', 'test3@example.com', '$2a$10$YourBcryptHashHereForTesting123456789012345678901234567890', 'Test', 'User3', true, true, NOW(), NOW()),
('44444444-4444-4444-4444-444444444444', 'unverified@example.com', '$2a$10$YourBcryptHashHereForTesting123456789012345678901234567890', 'Unverified', 'User', false, true, NOW(), NOW()),
('55555555-5555-5555-5555-555555555555', 'inactive@example.com', '$2a$10$YourBcryptHashHereForTesting123456789012345678901234567890', 'Inactive', 'User', true, false, NOW(), NOW());

-- Insert test sessions
INSERT INTO user_sessions (id, user_id, token, expires_at, created_at, last_accessed) VALUES
('session-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'token-valid-user1', NOW() + INTERVAL '24 hours', NOW(), NOW()),
('session-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', 'token-valid-user2', NOW() + INTERVAL '24 hours', NOW(), NOW()),
('session-3333-3333-3333-333333333333', '33333333-3333-3333-3333-333333333333', 'token-expired', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '25 hours', NOW() - INTERVAL '25 hours');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as users FROM users;
SELECT COUNT(*) as sessions FROM user_sessions;
SELECT COUNT(*) as active_users FROM users WHERE is_active = true;
SELECT COUNT(*) as verified_users FROM users WHERE is_verified = true;
