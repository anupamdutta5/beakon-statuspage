-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create platform table
CREATE TABLE IF NOT EXISTS platforms (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL UNIQUE,
    url TEXT NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    admin_email TEXT NOT NULL,
    support_email TEXT NOT NULL,
    metadata TEXT
);

-- Create saas_plans table
CREATE TABLE IF NOT EXISTS saas_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_interval VARCHAR(20) DEFAULT 'monthly',
    max_tenants INTEGER DEFAULT 1,
    max_users INTEGER DEFAULT 5,
    max_services INTEGER DEFAULT 10,
    max_monitors INTEGER DEFAULT 50,
    max_subscribers INTEGER DEFAULT 1000,
    max_incidents INTEGER DEFAULT 100,
    max_maintenance INTEGER DEFAULT 50,
    custom_domain BOOLEAN DEFAULT FALSE,
    white_label BOOLEAN DEFAULT FALSE,
    api BOOLEAN DEFAULT FALSE,
    integrations BOOLEAN DEFAULT FALSE,
    analytics BOOLEAN DEFAULT FALSE,
    support VARCHAR(50) DEFAULT 'email',
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT TRUE,
    is_popular BOOLEAN DEFAULT FALSE,
    button_text VARCHAR(100) DEFAULT 'Get Started',
    button_url VARCHAR(255) DEFAULT '/signup',
    display_order INTEGER DEFAULT 0,
    features TEXT,
    limits TEXT,
    metadata TEXT
);

-- Create pricing_features table
CREATE TABLE IF NOT EXISTS pricing_features (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    metadata TEXT
);

-- Create pricing_tiers table
CREATE TABLE IF NOT EXISTS pricing_tiers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    plan_id UUID NOT NULL REFERENCES saas_plans(id) ON DELETE CASCADE,
    billing_interval VARCHAR(20) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    discount_percent DECIMAL(5, 2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    metadata TEXT
);

-- Create plan_features table
CREATE TABLE IF NOT EXISTS plan_features (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    plan_id UUID NOT NULL REFERENCES saas_plans(id) ON DELETE CASCADE,
    feature_id INTEGER NOT NULL REFERENCES pricing_features(id) ON DELETE CASCADE,
    is_enabled BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    metadata TEXT,
    UNIQUE(plan_id, feature_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_saas_plans_is_active ON saas_plans(is_active);
CREATE INDEX IF NOT EXISTS idx_saas_plans_is_public ON saas_plans(is_public);
CREATE INDEX IF NOT EXISTS idx_saas_plans_is_popular ON saas_plans(is_popular);
CREATE INDEX IF NOT EXISTS idx_pricing_features_category ON pricing_features(category);
CREATE INDEX IF NOT EXISTS idx_pricing_features_is_active ON pricing_features(is_active);
CREATE INDEX IF NOT EXISTS idx_pricing_tiers_plan_id ON pricing_tiers(plan_id);
CREATE INDEX IF NOT EXISTS idx_plan_features_plan_id ON plan_features(plan_id);
CREATE INDEX IF NOT EXISTS idx_plan_features_feature_id ON plan_features(feature_id);
