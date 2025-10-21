-- Create the management database if it doesn't exist
SELECT 'CREATE DATABASE statuspage_management' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'statuspage_management')\gexec

-- Connect to the management database
\c statuspage_management

-- Create extension for UUID support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tables for database service
CREATE TABLE IF NOT EXISTS databases (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    ssl_mode VARCHAR(20) DEFAULT 'disable',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on database name
CREATE UNIQUE INDEX IF NOT EXISTS idx_databases_name ON databases (name) WHERE deleted_at IS NULL;

-- Create table for database users
CREATE TABLE IF NOT EXISTS database_users (
    id SERIAL PRIMARY KEY,
    database_id INTEGER REFERENCES databases(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_superuser BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on database users
CREATE INDEX IF NOT EXISTS idx_database_users_database_id ON database_users(database_id);

-- Create table for database backups
CREATE TABLE IF NOT EXISTS database_backups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    database_id INTEGER REFERENCES databases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

-- Create index on database backups
CREATE INDEX IF NOT EXISTS idx_database_backups_database_id ON database_backups(database_id);

-- Create table for database migrations
CREATE TABLE IF NOT EXISTS database_migrations (
    id SERIAL PRIMARY KEY,
    database_id INTEGER REFERENCES databases(id) ON DELETE CASCADE,
    version BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    execution_time_ms BIGINT,
    error_message TEXT
);

-- Create index on database migrations
CREATE INDEX IF NOT EXISTS idx_database_migrations_database_id ON database_migrations(database_id);

-- Create initial admin user for the management database
INSERT INTO database_users (database_id, username, password, is_superuser)
SELECT 1, 'statuspage_admin', crypt('statuspage_password', gen_salt('bf')), true
WHERE NOT EXISTS (SELECT 1 FROM database_users WHERE username = 'statuspage_admin');

-- Create function to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to update updated_at column
DO $$
DECLARE
    t record;
BEGIN
    FOR t IN 
        SELECT table_name 
        FROM information_schema.columns 
        WHERE column_name = 'updated_at' 
        AND table_schema = 'public'
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS update_%s_updated_at ON %I', t.table_name, t.table_name);
        EXECUTE format('CREATE TRIGGER update_%s_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', 
                      t.table_name, t.table_name);
    END LOOP;
END;
$$ LANGUAGE plpgsql;
