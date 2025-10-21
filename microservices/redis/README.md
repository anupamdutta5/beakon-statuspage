# Redis Infrastructure for Beakon

**Purpose**: High-performance session and cache storage for Beakon microservices
**Pattern**: Infrastructure service (NOT a microservice wrapper)
**Deployment**: Docker-based with easy migration to managed services

---

## Quick Start

### Development (Local)

```bash
cd microservices/redis

# Start Redis
docker-compose -f docker-compose.redis.yml up -d

# Verify
docker ps | grep redis
redis-cli ping  # Should return: PONG

# Access Web UI (Redis Commander)
open http://localhost:8081  # Login: admin/admin
```

### Production

```bash
cd microservices/redis

# Deploy with high availability
docker-compose -f docker-compose.production.yml up -d

# Verify all services
docker ps | grep beakon-redis
```

---

## Files

| File | Purpose |
|------|---------|
| `docker-compose.redis.yml` | Development setup (simple, single instance) |
| `docker-compose.production.yml` | Production setup (HA with Sentinel) |
| `redis.conf` | Development configuration |
| `redis-production.conf` | Production configuration (with security) |
| `sentinel.conf` | High availability monitoring config |

---

## Configuration

### Connect from Microservices

```bash
# Development
export REDIS_ENABLED=true
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=  # Empty for dev

# Production
export REDIS_ENABLED=true
export REDIS_HOST=<your-redis-server-ip>
export REDIS_PORT=6379
export REDIS_PASSWORD=<strong-password>
```

### Security (Production)

**Edit `redis-production.conf`**:
```conf
# Generate password: openssl rand -base64 32
requirepass your-strong-password-min-32-chars
```

**Edit `sentinel.conf`**:
```conf
sentinel auth-pass beakon-master your-strong-password-min-32-chars
```

---

## Migration to Managed Service

When ready to scale, simply change environment variables:

```bash
# From Docker Redis
export REDIS_HOST=localhost

# To AWS ElastiCache
export REDIS_HOST=your-cluster.abc123.cache.amazonaws.com

# To DigitalOcean Managed Redis
export REDIS_HOST=your-redis-cluster-do.ondigitalocean.com
export REDIS_PORT=25061

# NO CODE CHANGES NEEDED!
```

---

## Monitoring

### Metrics (Prometheus)

```bash
# Redis Exporter exposes metrics on port 9121
curl http://localhost:9121/metrics
```

### Web UI (Redis Commander)

- **URL**: http://localhost:8081
- **Login**: admin / admin
- **Features**: Browse keys, execute commands, monitor stats

### CLI Commands

```bash
# Connect
redis-cli

# Basic commands
PING                    # Test connection
INFO                    # Server info
DBSIZE                  # Number of keys
KEYS refresh:*          # List refresh tokens
GET refresh:<token>     # Get token data
TTL refresh:<token>     # Time to live
```

---

## Management

```bash
# Start Redis
docker-compose -f docker-compose.redis.yml up -d

# Stop Redis
docker-compose -f docker-compose.redis.yml down

# View logs
docker logs beakon-redis -f

# Restart Redis
docker-compose -f docker-compose.redis.yml restart

# Remove data (CAREFUL!)
docker-compose -f docker-compose.redis.yml down -v
```

---

## Backup & Restore

### Automatic Backups

Redis is configured with dual persistence:
- **RDB snapshots**: Every 15 minutes
- **AOF append-only**: Every second

### Manual Backup

```bash
# Trigger snapshot
docker exec -it beakon-redis-master redis-cli -a <password> BGSAVE

# Copy backup files
docker cp beakon-redis-master:/data/dump.rdb ./backup-$(date +%Y%m%d).rdb
docker cp beakon-redis-master:/data/appendonly.aof ./backup-$(date +%Y%m%d).aof
```

### Restore

```bash
# Stop Redis
docker-compose -f docker-compose.production.yml down

# Copy backup files
docker cp ./backup.rdb beakon-redis-master:/data/dump.rdb
docker cp ./backup.aof beakon-redis-master:/data/appendonly.aof

# Start Redis
docker-compose -f docker-compose.production.yml up -d
```

---

## Troubleshooting

### Redis won't start

```bash
# Check logs
docker logs beakon-redis

# Common issues:
# - Port 6379 in use: Change port in docker-compose
# - Permission denied: Check volume permissions
# - Config error: Validate redis.conf
```

### Services can't connect

```bash
# Test from host
redis-cli -h localhost -p 6379 ping

# Test from microservice
telnet <redis-host> 6379
```

### Out of memory

```bash
# Check memory usage
docker exec -it beakon-redis redis-cli INFO memory

# Increase maxmemory in redis.conf (default: 256MB dev, 1GB prod)
maxmemory 512mb
```

---

## Performance

| Operation | Latency | Throughput |
|-----------|---------|------------|
| SET | 0.1-0.5ms | 100k ops/s |
| GET | 0.1-0.3ms | 150k ops/s |
| DEL | 0.1-0.4ms | 100k ops/s |

**Comparison to PostgreSQL**:
- 10-30x faster for session operations
- Sub-millisecond latency vs 5-15ms

---

## Documentation

For detailed guides, see:
- **[../../REDIS_SETUP.md](../../REDIS_SETUP.md)** - Quick setup guide
- **[../../REDIS_PRODUCTION_DEPLOYMENT.md](../../REDIS_PRODUCTION_DEPLOYMENT.md)** - Production deployment
- **[../../AUTHENTICATION_GUIDE.md](../../AUTHENTICATION_GUIDE.md)** - Redis usage in authentication

---

## Summary

✅ **Simple**: One command to start
✅ **Production-ready**: HA with Sentinel, persistence, monitoring
✅ **Flexible**: Easy migration to managed services
✅ **Fast**: 10-30x faster than PostgreSQL
✅ **Reliable**: Automatic PostgreSQL fallback if Redis fails

**Not a microservice** - This is infrastructure that microservices connect to directly.
