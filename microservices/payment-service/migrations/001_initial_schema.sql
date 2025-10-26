--
-- Payment Service - Initial Schema
-- Database: payments
-- Description: Billing, subscriptions, payments, and invoicing
--

BEGIN;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================================
-- Payment and Subscription Types
-- =====================================================================

CREATE TYPE subscription_status AS ENUM ('trial', 'active', 'past_due', 'cancelled', 'expired');
CREATE TYPE payment_status AS ENUM ('pending', 'processing', 'succeeded', 'failed', 'refunded', 'cancelled');
CREATE TYPE payment_method_type AS ENUM ('card', 'bank_account', 'paypal', 'stripe', 'manual');
CREATE TYPE billing_interval AS ENUM ('monthly', 'yearly', 'quarterly');

-- =====================================================================
-- Subscriptions
-- =====================================================================

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    plan_id UUID NOT NULL, -- references subscription_plans in saas_admin
    status subscription_status DEFAULT 'trial',
    billing_interval billing_interval DEFAULT 'monthly',
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT false,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method_id UUID,
    stripe_subscription_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_current_period_end ON subscriptions(current_period_end);
CREATE INDEX idx_subscriptions_stripe_id ON subscriptions(stripe_subscription_id);

-- =====================================================================
-- Payment Methods
-- =====================================================================

CREATE TABLE IF NOT EXISTS payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    type payment_method_type NOT NULL,
    is_default BOOLEAN DEFAULT false,
    card_last_four VARCHAR(4),
    card_brand VARCHAR(50),
    card_exp_month INTEGER,
    card_exp_year INTEGER,
    bank_account_last_four VARCHAR(4),
    bank_account_routing VARCHAR(50),
    paypal_email VARCHAR(255),
    stripe_payment_method_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_payment_methods_tenant_id ON payment_methods(tenant_id);
CREATE INDEX idx_payment_methods_type ON payment_methods(type);
CREATE INDEX idx_payment_methods_is_default ON payment_methods(is_default);
CREATE INDEX idx_payment_methods_stripe_id ON payment_methods(stripe_payment_method_id);
CREATE INDEX idx_payment_methods_deleted_at ON payment_methods(deleted_at);

-- =====================================================================
-- Payments
-- =====================================================================

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    subscription_id UUID,
    invoice_id UUID,
    payment_method_id UUID,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status payment_status DEFAULT 'pending',
    description TEXT,
    stripe_payment_intent_id VARCHAR(255),
    stripe_charge_id VARCHAR(255),
    failure_code VARCHAR(100),
    failure_message TEXT,
    paid_at TIMESTAMP,
    refunded_at TIMESTAMP,
    refund_amount DECIMAL(10,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_tenant_id ON payments(tenant_id);
CREATE INDEX idx_payments_subscription_id ON payments(subscription_id);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_stripe_intent_id ON payments(stripe_payment_intent_id);
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

-- =====================================================================
-- Invoices
-- =====================================================================

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    subscription_id UUID,
    invoice_number VARCHAR(100) UNIQUE NOT NULL,
    status payment_status DEFAULT 'pending',
    amount_due DECIMAL(10,2) NOT NULL,
    amount_paid DECIMAL(10,2) DEFAULT 0,
    amount_remaining DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    description TEXT,
    due_date TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    stripe_invoice_id VARCHAR(255),
    pdf_url TEXT,
    hosted_invoice_url TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_stripe_id ON invoices(stripe_invoice_id);
CREATE INDEX idx_invoices_created_at ON invoices(created_at DESC);

-- =====================================================================
-- Invoice Line Items
-- =====================================================================

CREATE TABLE IF NOT EXISTS invoice_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL,
    description TEXT NOT NULL,
    quantity INTEGER DEFAULT 1,
    unit_amount DECIMAL(10,2) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoice_line_items_invoice_id ON invoice_line_items(invoice_id);

-- =====================================================================
-- Billing History (Audit trail)
-- =====================================================================

CREATE TABLE IF NOT EXISTS billing_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- 'subscription_created', 'payment_succeeded', 'invoice_sent', etc.
    related_id UUID, -- subscription_id, payment_id, invoice_id
    amount DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_billing_history_tenant_id ON billing_history(tenant_id);
CREATE INDEX idx_billing_history_event_type ON billing_history(event_type);
CREATE INDEX idx_billing_history_related_id ON billing_history(related_id);
CREATE INDEX idx_billing_history_created_at ON billing_history(created_at DESC);

-- =====================================================================
-- Usage Records (For usage-based billing)
-- =====================================================================

CREATE TABLE IF NOT EXISTS usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    subscription_id UUID NOT NULL,
    metric VARCHAR(100) NOT NULL, -- 'api_calls', 'storage_gb', 'bandwidth_gb', etc.
    quantity DECIMAL(15,5) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_usage_records_tenant_id ON usage_records(tenant_id);
CREATE INDEX idx_usage_records_subscription_id ON usage_records(subscription_id);
CREATE INDEX idx_usage_records_metric ON usage_records(metric);
CREATE INDEX idx_usage_records_timestamp ON usage_records(timestamp DESC);

-- =====================================================================
-- Triggers
-- =====================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_methods_updated_at
    BEFORE UPDATE ON payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
