#!/bin/bash
set -e

# Check if database exists
db_exists=$(psql -U "$POSTGRES_USER" -tAc "SELECT 1 FROM pg_database WHERE datname='saas_admin'")

if [ -z "$db_exists" ]; then
    # Create database if it doesn't exist
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        CREATE DATABASE saas_admin;
    EOSQL
    echo "Created database saas_admin"
else
    echo "Database saas_admin already exists, skipping creation"
fi

# Connect to the database and set up extensions
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "saas_admin" <<-EOSQL
    -- Create necessary extensions
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    
    -- Add any other initialization SQL here
    
    -- List all databases for debugging
    \l
    
    -- List all extensions in the database
    \dx
EOSQL
