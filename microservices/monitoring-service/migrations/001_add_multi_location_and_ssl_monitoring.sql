-- Migration: Add Multi-Location Monitoring and SSL Certificate Monitoring
-- Date: 2025-10-21
-- Phase: 1 - Critical Missing Features

-- ====================================
-- 1. MONITORING LOCATIONS TABLE
-- ====================================
CREATE TABLE IF NOT EXISTS monitoring_locations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    city VARCHAR(100),
    country VARCHAR(100) NOT NULL,
    region VARCHAR(50) NOT NULL, -- us-east, eu-west, asia-south, etc.
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6),
    is_active BOOLEAN DEFAULT true,
    provider VARCHAR(50), -- aws, gcp, azure, digitalocean
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_monitoring_locations_active ON monitoring_locations(is_active) WHERE is_active = true;
CREATE INDEX idx_monitoring_locations_region ON monitoring_locations(region);

COMMENT ON TABLE monitoring_locations IS 'Global monitoring node locations for multi-location health checks';

-- Seed 10 global monitoring locations
INSERT INTO monitoring_locations (name, city, country, region, latitude, longitude, provider) VALUES
('US East (N. Virginia)', 'Ashburn', 'United States', 'us-east-1', 38.9697, -77.3855, 'aws'),
('US West (Oregon)', 'Portland', 'United States', 'us-west-2', 45.5231, -122.6765, 'aws'),
('EU West (Ireland)', 'Dublin', 'Ireland', 'eu-west-1', 53.3498, -6.2603, 'aws'),
('EU Central (Frankfurt)', 'Frankfurt', 'Germany', 'eu-central-1', 50.1109, 8.6821, 'aws'),
('Asia Pacific (Singapore)', 'Singapore', 'Singapore', 'ap-southeast-1', 1.3521, 103.8198, 'aws'),
('Asia Pacific (Tokyo)', 'Tokyo', 'Japan', 'ap-northeast-1', 35.6762, 139.6503, 'aws'),
('Asia Pacific (Mumbai)', 'Mumbai', 'India', 'ap-south-1', 19.0760, 72.8777, 'aws'),
('South America (São Paulo)', 'São Paulo', 'Brazil', 'sa-east-1', -23.5505, -46.6333, 'aws'),
('Canada (Central)', 'Montreal', 'Canada', 'ca-central-1', 45.5017, -73.5673, 'aws'),
('Asia Pacific (Sydney)', 'Sydney', 'Australia', 'ap-southeast-2', -33.8688, 151.2093, 'aws')
ON CONFLICT DO NOTHING;

-- ====================================
-- 2. SSL CERTIFICATES TABLE
-- ====================================
CREATE TABLE IF NOT EXISTS ssl_certificates (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    domain VARCHAR(255) NOT NULL,
    issuer VARCHAR(255),
    subject VARCHAR(255),
    serial_number VARCHAR(255),
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    days_until_expiry INT GENERATED ALWAYS AS (
        CASE
            WHEN valid_until IS NULL THEN NULL
            ELSE EXTRACT(DAY FROM (valid_until - NOW()))::INT
        END
    ) STORED,
    last_checked TIMESTAMP,
    is_valid BOOLEAN DEFAULT true,
    is_self_signed BOOLEAN DEFAULT false,
    warning_sent_30d BOOLEAN DEFAULT false,
    warning_sent_14d BOOLEAN DEFAULT false,
    warning_sent_7d BOOLEAN DEFAULT false,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, domain)
);

CREATE INDEX idx_ssl_tenant ON ssl_certificates(tenant_id) WHERE is_valid = true;
CREATE INDEX idx_ssl_expiry ON ssl_certificates(days_until_expiry) WHERE days_until_expiry < 30 AND is_valid = true;
CREATE INDEX idx_ssl_expiring_soon ON ssl_certificates(tenant_id, days_until_expiry) WHERE days_until_expiry < 30;

COMMENT ON TABLE ssl_certificates IS 'SSL certificate tracking and expiration monitoring';
COMMENT ON COLUMN ssl_certificates.days_until_expiry IS 'Auto-calculated days until expiration (GENERATED column)';

-- ====================================
-- 3. UPDATE EXISTING TABLES
-- ====================================

-- Add location_id to health_checks table if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='health_checks' AND column_name='location_id') THEN
        ALTER TABLE health_checks ADD COLUMN location_id BIGINT REFERENCES monitoring_locations(id);
        CREATE INDEX idx_health_checks_location ON health_checks(location_id);
    END IF;
END $$;

-- Add performance metric columns to health_checks if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='health_checks' AND column_name='response_time_ms') THEN
        ALTER TABLE health_checks
        ADD COLUMN response_time_ms INT,
        ADD COLUMN ttfb_ms INT,
        ADD COLUMN dns_time_ms INT,
        ADD COLUMN connection_time_ms INT,
        ADD COLUMN ssl_handshake_ms INT;
    END IF;
END $$;

-- ====================================
-- 4. MONITORING RESULTS PARTITIONING
-- ====================================

-- Create monitoring_results table with partitioning (if not exists)
CREATE TABLE IF NOT EXISTS monitoring_results (
    id BIGSERIAL,
    monitor_id BIGINT NOT NULL,
    location_id BIGINT REFERENCES monitoring_locations(id),
    checked_at TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL, -- operational, degraded, down
    response_time_ms INT,
    ttfb_ms INT,
    dns_time_ms INT,
    connection_time_ms INT,
    ssl_handshake_ms INT,
    status_code INT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (id, checked_at)
) PARTITION BY RANGE (checked_at);

-- Create index on partitioned table
CREATE INDEX IF NOT EXISTS idx_monitoring_results_monitor_time
ON monitoring_results(monitor_id, checked_at DESC);

CREATE INDEX IF NOT EXISTS idx_monitoring_results_location
ON monitoring_results(location_id, monitor_id, checked_at DESC);

CREATE INDEX IF NOT EXISTS idx_monitoring_results_status
ON monitoring_results(status, checked_at DESC) WHERE status != 'operational';

COMMENT ON TABLE monitoring_results IS 'Time-series monitoring check results (partitioned by day)';

-- Create initial partitions for current and next 7 days
DO $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
    start_date TEXT;
    end_date TEXT;
BEGIN
    FOR i IN 0..7 LOOP
        partition_date := CURRENT_DATE + (i || ' days')::INTERVAL;
        partition_name := 'monitoring_results_' || TO_CHAR(partition_date, 'YYYY_MM_DD');
        start_date := TO_CHAR(partition_date, 'YYYY-MM-DD');
        end_date := TO_CHAR(partition_date + INTERVAL '1 day', 'YYYY-MM-DD');

        -- Check if partition exists
        IF NOT EXISTS (
            SELECT 1 FROM pg_class WHERE relname = partition_name
        ) THEN
            EXECUTE format(
                'CREATE TABLE %I PARTITION OF monitoring_results FOR VALUES FROM (%L) TO (%L)',
                partition_name,
                start_date,
                end_date
            );
            RAISE NOTICE 'Created partition: %', partition_name;
        END IF;
    END LOOP;
END $$;

-- ====================================
-- 5. AGGREGATED METRICS TABLES
-- ====================================

-- Hourly aggregates (warm data: 8-90 days)
CREATE TABLE IF NOT EXISTS monitoring_results_hourly (
    monitor_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    hour TIMESTAMP NOT NULL,
    avg_response_time_ms INT,
    min_response_time_ms INT,
    max_response_time_ms INT,
    p50_response_time_ms INT,
    p95_response_time_ms INT,
    p99_response_time_ms INT,
    uptime_percentage DECIMAL(5,2),
    check_count INT,
    failure_count INT,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (monitor_id, location_id, hour)
);

CREATE INDEX idx_monitoring_hourly_time ON monitoring_results_hourly(hour DESC);

COMMENT ON TABLE monitoring_results_hourly IS 'Hourly aggregated monitoring metrics for historical analysis';

-- Daily aggregates (cold data: 90+ days)
CREATE TABLE IF NOT EXISTS monitoring_results_daily (
    monitor_id BIGINT NOT NULL,
    day DATE NOT NULL,
    avg_response_time_ms INT,
    p95_response_time_ms INT,
    p99_response_time_ms INT,
    uptime_percentage DECIMAL(5,2),
    total_checks INT,
    total_failures INT,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (monitor_id, day)
);

CREATE INDEX idx_monitoring_daily_day ON monitoring_results_daily(day DESC);

COMMENT ON TABLE monitoring_results_daily IS 'Daily aggregated monitoring metrics for long-term storage';

-- ====================================
-- 6. HEARTBEAT MONITORS TABLE
-- ====================================

CREATE TABLE IF NOT EXISTS heartbeat_monitors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unique_key VARCHAR(255) NOT NULL, -- Used in heartbeat ping URL
    expected_interval_seconds INT NOT NULL, -- Expected time between pings
    grace_period_seconds INT DEFAULT 300, -- Grace period before marking as down
    last_ping TIMESTAMP,
    is_alive BOOLEAN DEFAULT false,
    consecutive_misses INT DEFAULT 0,
    alert_sent BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, unique_key)
);

CREATE INDEX idx_heartbeat_tenant ON heartbeat_monitors(tenant_id);
CREATE INDEX idx_heartbeat_status ON heartbeat_monitors(is_alive, last_ping);

COMMENT ON TABLE heartbeat_monitors IS 'Cron job and heartbeat monitoring (expects regular pings)';

-- ====================================
-- 7. ON-CALL SCHEDULES TABLE
-- ====================================

CREATE TABLE IF NOT EXISTS on_call_schedules (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    rotation_type VARCHAR(20) NOT NULL, -- daily, weekly, custom
    rotation_start TIMESTAMP NOT NULL,
    rotation_interval_hours INT DEFAULT 168, -- Default: weekly (168 hours)
    participants JSONB NOT NULL, -- [{user_id: 'uuid', order: 1}, ...]
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_on_call_tenant ON on_call_schedules(tenant_id);
CREATE INDEX idx_on_call_active ON on_call_schedules(is_active) WHERE is_active = true;

COMMENT ON TABLE on_call_schedules IS 'On-call rotation schedules for alert escalation';

-- ====================================
-- 8. ESCALATION POLICIES TABLE
-- ====================================

CREATE TABLE IF NOT EXISTS escalation_policies (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    levels JSONB NOT NULL, -- [{delay_minutes: 0, notify: ['user1', 'user2']}, {delay_minutes: 15, notify: ['user3']}]
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_escalation_tenant ON escalation_policies(tenant_id);
CREATE UNIQUE INDEX idx_escalation_default ON escalation_policies(tenant_id) WHERE is_default = true;

COMMENT ON TABLE escalation_policies IS 'Alert escalation policies defining who gets notified and when';

-- ====================================
-- 9. FUNCTIONS AND TRIGGERS
-- ====================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to tables
DO $$
DECLARE
    table_name TEXT;
BEGIN
    FOR table_name IN
        SELECT unnest(ARRAY[
            'monitoring_locations',
            'ssl_certificates',
            'heartbeat_monitors',
            'on_call_schedules',
            'escalation_policies'
        ])
    LOOP
        EXECUTE format('
            DROP TRIGGER IF EXISTS update_%I_updated_at ON %I;
            CREATE TRIGGER update_%I_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        ', table_name, table_name, table_name, table_name);
    END LOOP;
END $$;

-- ====================================
-- MIGRATION COMPLETE
-- ====================================

-- Insert migration record
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO schema_migrations (version) VALUES ('001_add_multi_location_and_ssl_monitoring')
ON CONFLICT (version) DO NOTHING;

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ Migration 001 completed successfully!';
    RAISE NOTICE '   - Created monitoring_locations table with 10 global locations';
    RAISE NOTICE '   - Created ssl_certificates table for SSL monitoring';
    RAISE NOTICE '   - Created monitoring_results partitioned table';
    RAISE NOTICE '   - Created monitoring_results_hourly aggregation table';
    RAISE NOTICE '   - Created monitoring_results_daily aggregation table';
    RAISE NOTICE '   - Created heartbeat_monitors table';
    RAISE NOTICE '   - Created on_call_schedules table';
    RAISE NOTICE '   - Created escalation_policies table';
    RAISE NOTICE '   - Added triggers for updated_at columns';
END $$;
