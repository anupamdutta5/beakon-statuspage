-- Create the statuspage_user with password
CREATE USER statuspage_user WITH PASSWORD 'statuspage_password';

-- Create the statuspage_users database
CREATE DATABASE statuspage_users;

-- Grant all privileges on the database to the user
GRANT ALL PRIVILEGES ON DATABASE statuspage_users TO statuspage_user;

-- Connect to the new database
\c statuspage_users;

-- Grant all necessary privileges to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO statuspage_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO statuspage_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO statuspage_user;
GRANT ALL PRIVILEGES ON SCHEMA public TO statuspage_user;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO statuspage_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO statuspage_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON FUNCTIONS TO statuspage_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TYPES TO statuspage_user;
