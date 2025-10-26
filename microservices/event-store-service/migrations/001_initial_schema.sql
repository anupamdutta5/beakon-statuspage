--
-- Event Store Service - Initial Schema
-- Database: event_store
-- Description: Event sourcing for the platform
--

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_version INTEGER DEFAULT 1,
    event_data JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    user_id UUID,
    correlation_id UUID,
    causation_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_tenant_id ON events(tenant_id);
CREATE INDEX idx_events_aggregate ON events(aggregate_type, aggregate_id);
CREATE INDEX idx_events_created_at ON events(created_at DESC);
CREATE INDEX idx_events_correlation_id ON events(correlation_id);

CREATE TABLE IF NOT EXISTS event_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    version INTEGER NOT NULL,
    snapshot_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_event_snapshots_aggregate ON event_snapshots(aggregate_type, aggregate_id, version DESC);

CREATE TABLE IF NOT EXISTS event_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_name VARCHAR(255) UNIQUE NOT NULL,
    event_types TEXT[] NOT NULL,
    last_processed_event_id UUID,
    last_processed_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
