# Beakon Application Startup Guide

## Prerequisites

### 1. Database Setup
- **PostgreSQL** must be running on your system
- **Database Name**: `statuspage`
- **Username**: `postgres`
- **Password**: `postgres` (default PostgreSQL password)
- **Host**: `localhost`
- **Port**: `5432`

### 2. Environment Setup
- **Go** version 1.19+ installed
- **PostgreSQL** client tools available

## Database Configuration

### Check Database Connection
```bash
# Test if PostgreSQL is running
psql -h localhost -U postgres -d statuspage -c "SELECT 1;"
```

If this fails, you need to:
1. Start PostgreSQL service
2. Create the `statuspage` database
3. Ensure the `postgres` user has the correct password

### Create Database (if needed)
```bash
# Connect to PostgreSQL as superuser
psql -h localhost -U postgres

# Create the database
CREATE DATABASE statuspage;

# Exit psql
\q
```

## Application Startup

### Method 1: Using start.sh Script (Recommended)
```bash
# Make the script executable
chmod +x start.sh

# Start the application
./start.sh
```

### Method 2: Manual Environment Variables
```bash
# Set all required environment variables
export JWT_SECRET="development-secret-key-change-in-production"
export STRIPE_SECRET_KEY="sk_test_dev_key_for_development_only"
export STRIPE_WEBHOOK_SECRET="whsec_test_dev_key_for_development_only"
export STRIPE_PUBLISHABLE_KEY="pk_test_dev_key_for_development_only"
export DB_PASSWORD="postgres"
export ENCRYPTION_KEY="dev_encryption_key_32_chars_long"
export SESSION_SECRET="dev_session_secret"
export ENVIRONMENT="development"

# Start the application
go run cmd/api/main.go
```

## Environment Variables Reference

| Variable | Value | Description |
|----------|-------|-------------|
| `JWT_SECRET` | `development-secret-key-change-in-production` | JWT signing secret |
| `STRIPE_SECRET_KEY` | `sk_test_dev_key_for_development_only` | Stripe secret key |
| `STRIPE_WEBHOOK_SECRET` | `whsec_test_dev_key_for_development_only` | Stripe webhook secret |
| `STRIPE_PUBLISHABLE_KEY` | `pk_test_dev_key_for_development_only` | Stripe publishable key |
| `DB_PASSWORD` | `postgres` | Database password |
| `ENCRYPTION_KEY` | `dev_encryption_key_32_chars_long` | Data encryption key |
| `SESSION_SECRET` | `dev_session_secret` | Session encryption secret |
| `ENVIRONMENT` | `development` | Application environment |

## Verification

### 1. Check Server Status
```bash
# Test if the server is running
curl http://localhost:8080/api/test-admin
```

Expected response:
```json
{"message":"Admin API is working without auth"}
```

### 2. Check Database Connection
```bash
# Test database connectivity
curl http://localhost:8080/api/v1/admin/saas/admin/tenants
```

### 3. Access the Application
- **Main Dashboard**: http://localhost:8080/admin/saas
- **API Base**: http://localhost:8080/api/v1/

## Troubleshooting

### Database Connection Issues

**Error**: `FATAL: password authentication failed for user "postgres"`

**Solutions**:
1. **Check PostgreSQL is running**:
   ```bash
   brew services list | grep postgresql
   # or
   sudo systemctl status postgresql
   ```

2. **Reset PostgreSQL password**:
   ```bash
   # Connect as superuser
   sudo -u postgres psql
   
   # Change password
   ALTER USER postgres PASSWORD 'postgres';
   \q
   ```

3. **Check database exists**:
   ```bash
   psql -h localhost -U postgres -l | grep statuspage
   ```

### Port Already in Use

**Error**: `listen tcp :8080: bind: address already in use`

**Solution**:
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill <PID>

# Or kill all Go processes
pkill -f "go run"
```

### Environment Variables Not Loading

**Issue**: Environment variables not being picked up

**Solution**:
1. Check if `.env` file exists and has correct values
2. Use explicit export commands before starting
3. Use the `start.sh` script which sets all variables

## Development Workflow

### 1. Start the Application
```bash
./start.sh
```

### 2. Make Changes
- Edit Go files in `internal/`
- Edit frontend files in `web/`
- Update templates in `web/templates/`

### 3. Test Changes
- The server auto-reloads on Go file changes
- For frontend changes, update cache buster in templates
- Test API endpoints with curl or browser

### 4. Stop the Application
```bash
# Find the process
ps aux | grep "go run"

# Kill the process
kill <PID>

# Or use Ctrl+C if running in foreground
```

## Common Commands

```bash
# Start application
./start.sh

# Check server status
curl http://localhost:8080/api/test-admin

# View logs (if running in background)
tail -f /tmp/beakon.log

# Test database connection
psql -h localhost -U postgres -d statuspage -c "SELECT 1;"

# Kill all Go processes
pkill -f "go run"

# Check what's running on port 8080
lsof -i :8080
```

## File Structure

```
Beakon/
├── start.sh                    # Startup script
├── cmd/api/main.go            # Application entry point
├── internal/api/              # API handlers and routes
├── web/                       # Frontend files
│   ├── templates/             # HTML templates
│   └── static/                # CSS, JS, images
├── .env                       # Environment variables (if exists)
└── STARTUP_GUIDE.md          # This file
```

## Notes

- The application runs on port **8080** by default
- Database migrations are handled automatically on startup
- All API endpoints are prefixed with `/api/v1/`
- The admin dashboard is available at `/admin/saas`
- Environment variables in `start.sh` take precedence over `.env` file

