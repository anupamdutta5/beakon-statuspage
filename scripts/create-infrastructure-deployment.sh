#!/bin/bash
#
# Create docker-deployment structure for infrastructure services
#

set -e

cd /Users/anuoamdutta/Desktop/statuspage/Beakon

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Creating Infrastructure Deployment${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Create directories for infrastructure
mkdir -p docker-deployment/postgres/configs
mkdir -p docker-deployment/redis/configs
mkdir -p docker-deployment/rabbitmq/configs

echo -e "${GREEN}✓${NC} Created infrastructure directories"
echo ""

# Generate PostgreSQL docker-compose.yml
cat > docker-deployment/postgres/docker-compose.yml << 'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    container_name: statuspage-postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./configs/init-scripts:/docker-entrypoint-initdb.d:ro
    environment:
      - POSTGRES_USER=${POSTGRES_USER:-postgres}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
      - POSTGRES_DB=${POSTGRES_DB:-postgres}
    networks:
      - beakon-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
    name: beakon_postgres_data

networks:
  beakon-network:
    name: beakon-network
    external: true
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/postgres/docker-compose.yml"

# Generate Redis docker-compose.yml
cat > docker-deployment/redis/docker-compose.yml << 'EOF'
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: beakon-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./configs/redis.conf:/usr/local/etc/redis/redis.conf:ro
    command: redis-server /usr/local/etc/redis/redis.conf
    networks:
      - beakon-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

volumes:
  redis_data:
    name: beakon_redis_data

networks:
  beakon-network:
    name: beakon-network
    external: true
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/redis/docker-compose.yml"

# Generate RabbitMQ docker-compose.yml
cat > docker-deployment/rabbitmq/docker-compose.yml << 'EOF'
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: beakon-rabbitmq
    ports:
      - "5672:5672"    # AMQP port
      - "15672:15672"  # Management UI
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
      - ./configs/rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf:ro
    environment:
      - RABBITMQ_DEFAULT_USER=${RABBITMQ_USER:-guest}
      - RABBITMQ_DEFAULT_PASS=${RABBITMQ_PASSWORD:-guest}
      - RABBITMQ_DEFAULT_VHOST=${RABBITMQ_VHOST:-/}
    networks:
      - beakon-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  rabbitmq_data:
    name: beakon_rabbitmq_data

networks:
  beakon-network:
    name: beakon-network
    external: true
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/rabbitmq/docker-compose.yml"
echo ""

# Generate PostgreSQL .env
cat > docker-deployment/postgres/.env << 'EOF'
# PostgreSQL Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/postgres/.env"

# Generate Redis .env
cat > docker-deployment/redis/.env << 'EOF'
# Redis Configuration
REDIS_PASSWORD=
REDIS_MAXMEMORY=256mb
REDIS_MAXMEMORY_POLICY=allkeys-lru
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/redis/.env"

# Generate RabbitMQ .env
cat > docker-deployment/rabbitmq/.env << 'EOF'
# RabbitMQ Configuration
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/rabbitmq/.env"
echo ""

# Generate PostgreSQL config
cat > docker-deployment/postgres/configs/postgresql.conf << 'EOF'
# PostgreSQL Configuration for Beakon Platform
# This file provides custom PostgreSQL settings

# Connection Settings
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB
maintenance_work_mem = 64MB

# Write Ahead Log
wal_buffers = 16MB
checkpoint_completion_target = 0.9

# Query Planning
random_page_cost = 1.1
effective_io_concurrency = 200

# Logging
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_rotation_age = 1d
log_rotation_size = 100MB
log_line_prefix = '%m [%p] %u@%d '
log_timezone = 'UTC'

# Locale
datestyle = 'iso, mdy'
timezone = 'UTC'
lc_messages = 'en_US.utf8'
lc_monetary = 'en_US.utf8'
lc_numeric = 'en_US.utf8'
lc_time = 'en_US.utf8'
default_text_search_config = 'pg_catalog.english'
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/postgres/configs/postgresql.conf"

# Create init-scripts directory
mkdir -p docker-deployment/postgres/configs/init-scripts

cat > docker-deployment/postgres/configs/init-scripts/README.md << 'EOF'
# PostgreSQL Initialization Scripts

Place SQL scripts here to be executed when PostgreSQL container first starts.

Scripts are executed in alphabetical order.

Example:
```
01_create_databases.sql
02_create_users.sql
03_grant_permissions.sql
```

Note: These scripts only run on first container startup (when data volume is empty).
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/postgres/configs/init-scripts/README.md"

# Generate Redis config
cat > docker-deployment/redis/configs/redis.conf << 'EOF'
# Redis Configuration for Beakon Platform

# Network
bind 0.0.0.0
protected-mode no
port 6379

# General
daemonize no
supervised no
pidfile /var/run/redis.pid
loglevel notice
logfile ""

# Snapshotting
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data

# Replication
replica-serve-stale-data yes
replica-read-only yes

# Security
# requirepass your_password_here

# Limits
maxclients 10000
maxmemory 256mb
maxmemory-policy allkeys-lru

# Append Only File
appendonly no
appendfilename "appendonly.aof"
appendfsync everysec

# Slow Log
slowlog-log-slower-than 10000
slowlog-max-len 128
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/redis/configs/redis.conf"

# Generate RabbitMQ config
cat > docker-deployment/rabbitmq/configs/rabbitmq.conf << 'EOF'
# RabbitMQ Configuration for Beakon Platform

# Network
listeners.tcp.default = 5672

# Management Plugin
management.tcp.port = 15672
management.load_definitions = /etc/rabbitmq/definitions.json

# Logging
log.console = true
log.console.level = info
log.file.level = info

# Connection Settings
channel_max = 2048
heartbeat = 60

# Memory
vm_memory_high_watermark.relative = 0.6

# Disk Space
disk_free_limit.absolute = 2GB

# Queue Settings
queue_master_locator = min-masters
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/rabbitmq/configs/rabbitmq.conf"

# Generate RabbitMQ definitions (exchanges, queues)
cat > docker-deployment/rabbitmq/configs/definitions.json << 'EOF'
{
  "rabbit_version": "3.11.0",
  "rabbitmq_version": "3.11.0",
  "exchanges": [
    {
      "name": "beakon_events",
      "vhost": "/",
      "type": "topic",
      "durable": true,
      "auto_delete": false,
      "internal": false,
      "arguments": {}
    }
  ],
  "queues": [
    {
      "name": "analytics-consumer_queue",
      "vhost": "/",
      "durable": true,
      "auto_delete": false,
      "arguments": {}
    },
    {
      "name": "notification-consumer_queue",
      "vhost": "/",
      "durable": true,
      "auto_delete": false,
      "arguments": {}
    },
    {
      "name": "audit-consumer_queue",
      "vhost": "/",
      "durable": true,
      "auto_delete": false,
      "arguments": {}
    },
    {
      "name": "billing-consumer_queue",
      "vhost": "/",
      "durable": true,
      "auto_delete": false,
      "arguments": {}
    }
  ],
  "bindings": []
}
EOF

echo -e "${GREEN}✓${NC} Generated docker-deployment/rabbitmq/configs/definitions.json"
echo ""

echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo "Infrastructure services configured: 3"
echo "  - postgres (Port 5432)"
echo "  - redis (Port 6379)"
echo "  - rabbitmq (Ports 5672, 15672)"
echo ""
echo "Files generated per service:"
echo "  - docker-compose.yml"
echo "  - .env"
echo "  - configs/*.conf"
echo ""
echo -e "${GREEN}✅ Infrastructure deployment structure created!${NC}"
echo ""
echo "To start infrastructure:"
echo "  cd docker-deployment/postgres && docker-compose up -d"
echo "  cd docker-deployment/redis && docker-compose up -d"
echo "  cd docker-deployment/rabbitmq && docker-compose up -d"
EOF

chmod +x /Users/anuoamdutta/Desktop/statuspage/Beakon/scripts/create-infrastructure-deployment.sh
