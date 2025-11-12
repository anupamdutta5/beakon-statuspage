-- Test fixtures for statuspage_payment database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE payment_methods, invoices, invoice_items, subscriptions, payment_transactions CASCADE;

-- Insert subscriptions for Tenant 1
INSERT INTO subscriptions (id, tenant_id, plan_id, status, billing_cycle, current_period_start, current_period_end, trial_ends_at, canceled_at, created_at, updated_at) VALUES
('sub-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', '99999999-9999-9999-9999-999999999999', 'active', 'monthly', NOW() - INTERVAL '15 days', NOW() + INTERVAL '15 days', NULL, NULL, NOW() - INTERVAL '6 months', NOW()),
('sub-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'canceled', 'monthly', NOW() - INTERVAL '45 days', NOW() - INTERVAL '15 days', NULL, NOW() - INTERVAL '15 days', NOW() - INTERVAL '12 months', NOW() - INTERVAL '15 days');

-- Insert subscriptions for Tenant 2
INSERT INTO subscriptions (id, tenant_id, plan_id, status, billing_cycle, current_period_start, current_period_end, trial_ends_at, canceled_at, created_at, updated_at) VALUES
('sub-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'active', 'yearly', NOW() - INTERVAL '3 months', NOW() + INTERVAL '9 months', NULL, NULL, NOW() - INTERVAL '3 months', NOW());

-- Insert subscriptions for Tenant 3 (trial)
INSERT INTO subscriptions (id, tenant_id, plan_id, status, billing_cycle, current_period_start, current_period_end, trial_ends_at, canceled_at, created_at, updated_at) VALUES
('sub-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'trialing', 'monthly', NOW(), NOW() + INTERVAL '30 days', NOW() + INTERVAL '14 days', NULL, NOW(), NOW());

-- Insert payment methods
INSERT INTO payment_methods (id, tenant_id, method_type, provider, provider_payment_method_id, is_default, card_last4, card_brand, card_exp_month, card_exp_year, billing_email, created_at, updated_at) VALUES
-- Tenant 1 payment methods
('pm-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'card', 'stripe', 'pm_test_visa_1111', true, '4242', 'visa', 12, 2025, 'billing@test1.com', NOW() - INTERVAL '6 months', NOW()),
('pm-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'card', 'stripe', 'pm_test_mastercard_1111', false, '5555', 'mastercard', 6, 2026, 'billing@test1.com', NOW() - INTERVAL '3 months', NOW()),

-- Tenant 2 payment method
('pm-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'card', 'stripe', 'pm_test_amex_2222', true, '1234', 'amex', 9, 2025, 'payments@test2.com', NOW() - INTERVAL '3 months', NOW()),

-- Tenant 3 payment method (trial, no charges yet)
('pm-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'card', 'stripe', 'pm_test_visa_3333', true, '7890', 'visa', 3, 2026, 'finance@test3.com', NOW(), NOW());

-- Insert invoices
INSERT INTO invoices (id, tenant_id, subscription_id, invoice_number, status, subtotal, tax, total, currency, due_date, paid_at, created_at, updated_at) VALUES
-- Tenant 1 invoices (current active subscription)
('inv-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-001', 'paid', 29.00, 2.90, 31.90, 'USD', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months'),
('inv-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-002', 'paid', 29.00, 2.90, 31.90, 'USD', NOW() - INTERVAL '5 months', NOW() - INTERVAL '5 months', NOW() - INTERVAL '5 months', NOW() - INTERVAL '5 months'),
('inv-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-003', 'paid', 29.00, 2.90, 31.90, 'USD', NOW() - INTERVAL '4 months', NOW() - INTERVAL '4 months', NOW() - INTERVAL '4 months', NOW() - INTERVAL '4 months'),
('inv-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-004', 'paid', 29.00, 2.90, 31.90, 'USD', NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months'),
('inv-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-005', 'paid', 29.00, 2.90, 31.90, 'USD', NOW() - INTERVAL '2 months', NOW() - INTERVAL '2 months', NOW() - INTERVAL '2 months', NOW() - INTERVAL '2 months'),
('inv-1111-0006-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-006', 'paid', 29.00, 2.90, 31.90, 'USD', NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month'),
('inv-1111-0007-0007-000000000007', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0001-0001-000000000001', 'INV-2024-007', 'open', 29.00, 2.90, 31.90, 'USD', NOW() + INTERVAL '15 days', NULL, NOW(), NOW()),

-- Tenant 1 old subscription invoices (canceled)
('inv-1111-0008-0008-000000000008', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0002-0002-000000000002', 'INV-2023-001', 'paid', 99.00, 9.90, 108.90, 'USD', NOW() - INTERVAL '12 months', NOW() - INTERVAL '12 months', NOW() - INTERVAL '12 months', NOW() - INTERVAL '12 months'),
('inv-1111-0009-0009-000000000009', 'tenant-1111-1111-1111-111111111111', 'sub-1111-0002-0002-000000000002', 'INV-2023-002', 'void', 99.00, 9.90, 108.90, 'USD', NOW() - INTERVAL '11 months', NULL, NOW() - INTERVAL '11 months', NOW() - INTERVAL '10 months'),

-- Tenant 2 invoices (yearly subscription)
('inv-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'sub-2222-0001-0001-000000000001', 'INV-2024-100', 'paid', 990.00, 99.00, 1089.00, 'USD', NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months'),

-- Tenant 3 invoice (trial, upcoming)
('inv-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'sub-3333-0001-0001-000000000001', 'INV-2024-200', 'draft', 199.00, 19.90, 218.90, 'USD', NOW() + INTERVAL '14 days', NULL, NOW(), NOW());

-- Insert invoice items
INSERT INTO invoice_items (id, invoice_id, description, quantity, unit_price, amount, created_at) VALUES
-- Items for Tenant 1 current invoices
('item-1111-0001-0001-000000000001', 'inv-1111-0001-0001-000000000001', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW() - INTERVAL '6 months'),
('item-1111-0002-0002-000000000002', 'inv-1111-0002-0002-000000000002', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW() - INTERVAL '5 months'),
('item-1111-0003-0003-000000000003', 'inv-1111-0003-0003-000000000003', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW() - INTERVAL '4 months'),
('item-1111-0004-0004-000000000004', 'inv-1111-0004-0004-000000000004', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW() - INTERVAL '3 months'),
('item-1111-0005-0005-000000000005', 'inv-1111-0005-0005-000000000005', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW() - INTERVAL '2 months'),
('item-1111-0006-0006-000000000006', 'inv-1111-0006-0006-000000000006', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW() - INTERVAL '1 month'),
('item-1111-0007-0007-000000000007', 'inv-1111-0007-0007-000000000007', 'Starter Plan - Monthly', 1, 29.00, 29.00, NOW()),

-- Items for Tenant 1 old subscription
('item-1111-0008-0008-000000000008', 'inv-1111-0008-0008-000000000008', 'Professional Plan - Monthly', 1, 99.00, 99.00, NOW() - INTERVAL '12 months'),

-- Items for Tenant 2 yearly
('item-2222-0001-0001-000000000001', 'inv-2222-0001-0001-000000000001', 'Professional Plan - Yearly', 1, 990.00, 990.00, NOW() - INTERVAL '3 months'),

-- Items for Tenant 3 trial
('item-3333-0001-0001-000000000001', 'inv-3333-0001-0001-000000000001', 'Enterprise Plan - Monthly', 1, 199.00, 199.00, NOW());

-- Insert payment transactions
INSERT INTO payment_transactions (id, tenant_id, invoice_id, payment_method_id, transaction_type, amount, currency, status, provider, provider_transaction_id, error_message, processed_at, created_at) VALUES
-- Successful payments for Tenant 1
('txn-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0001-0001-000000000001', 'pm-1111-0001-0001-000000000001', 'payment', 31.90, 'USD', 'succeeded', 'stripe', 'ch_test_1111_001', NULL, NOW() - INTERVAL '6 months', NOW() - INTERVAL '6 months'),
('txn-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0002-0002-000000000002', 'pm-1111-0001-0001-000000000001', 'payment', 31.90, 'USD', 'succeeded', 'stripe', 'ch_test_1111_002', NULL, NOW() - INTERVAL '5 months', NOW() - INTERVAL '5 months'),
('txn-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0003-0003-000000000003', 'pm-1111-0001-0001-000000000001', 'payment', 31.90, 'USD', 'succeeded', 'stripe', 'ch_test_1111_003', NULL, NOW() - INTERVAL '4 months', NOW() - INTERVAL '4 months'),
('txn-1111-0004-0004-000000000004', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0004-0004-000000000004', 'pm-1111-0001-0001-000000000001', 'payment', 31.90, 'USD', 'succeeded', 'stripe', 'ch_test_1111_004', NULL, NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months'),
('txn-1111-0005-0005-000000000005', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0005-0005-000000000005', 'pm-1111-0001-0001-000000000001', 'payment', 31.90, 'USD', 'succeeded', 'stripe', 'ch_test_1111_005', NULL, NOW() - INTERVAL '2 months', NOW() - INTERVAL '2 months'),
('txn-1111-0006-0006-000000000006', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0006-0006-000000000006', 'pm-1111-0001-0001-000000000001', 'payment', 31.90, 'USD', 'succeeded', 'stripe', 'ch_test_1111_006', NULL, NOW() - INTERVAL '1 month', NOW() - INTERVAL '1 month'),

-- Failed payment attempt
('txn-1111-0007-0007-000000000007', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0009-0009-000000000009', 'pm-1111-0001-0001-000000000001', 'payment', 108.90, 'USD', 'failed', 'stripe', NULL, 'Card declined - insufficient funds', NOW() - INTERVAL '10 months', NOW() - INTERVAL '10 months'),

-- Refund transaction
('txn-1111-0008-0008-000000000008', 'tenant-1111-1111-1111-111111111111', 'inv-1111-0008-0008-000000000008', 'pm-1111-0001-0001-000000000001', 'refund', 108.90, 'USD', 'succeeded', 'stripe', 're_test_1111_001', NULL, NOW() - INTERVAL '10 months', NOW() - INTERVAL '10 months'),

-- Successful payment for Tenant 2
('txn-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'inv-2222-0001-0001-000000000001', 'pm-2222-0001-0001-000000000001', 'payment', 1089.00, 'USD', 'succeeded', 'stripe', 'ch_test_2222_001', NULL, NOW() - INTERVAL '3 months', NOW() - INTERVAL '3 months');

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as subscriptions FROM subscriptions;
SELECT COUNT(*) as payment_methods FROM payment_methods;
SELECT COUNT(*) as invoices FROM invoices;
SELECT COUNT(*) as invoice_items FROM invoice_items;
SELECT COUNT(*) as transactions FROM payment_transactions;
SELECT status, COUNT(*) as count FROM subscriptions GROUP BY status ORDER BY status;
SELECT status, COUNT(*) as count FROM invoices GROUP BY status ORDER BY status;
SELECT status, COUNT(*) as count FROM payment_transactions GROUP BY status ORDER BY status;
