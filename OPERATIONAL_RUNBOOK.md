# Beakon - Operational Runbook

**Last Updated**: 2025-10-14
**On-Call Rotation**: 24/7 coverage required
**Escalation**: Engineering → DevOps → Leadership

---

## Emergency Contacts

| Role | Primary | Secondary | Phone | Slack |
|------|---------|-----------|-------|-------|
| On-Call Engineer | [Name] | [Name] | [Phone] | @oncall |
| DevOps Lead | [Name] | [Name] | [Phone] | @devops-lead |
| Engineering Manager | [Name] | [Name] | [Phone] | @eng-manager |
| CTO | [Name] | - | [Phone] | @cto |

**Emergency Slack Channel**: #incident-response
**Status Page**: https://status.beakon.com

---

## Quick Reference

### Service Health Checks

```bash
# Check all services
kubectl get pods -n beakon-prod

# Check specific service
curl http://tenant-admin-service:8099/health

# Check via API Gateway
curl https://api.beakon.com/health
```

### Common Commands

```bash
# View logs
kubectl logs -f deployment/tenant-admin-service --tail=100

# Restart service
kubectl rollout restart deployment/tenant-admin-service

# Scale service
kubectl scale deployment tenant-admin-service --replicas=5

# Check database connections
psql -U postgres -d tenant_admin_db -c "SELECT count(*) FROM pg_stat_activity;"

# Check Redis
redis-cli ping
```

---

## Incident Response Procedures

### Severity Levels

**P0 - Critical** (Response Time: Immediate)
- Complete service outage
- Data loss or corruption
- Security breach
- Payment processing failure

**P1 - High** (Response Time: 15 minutes)
- Partial service outage
- Significant performance degradation
- Customer-impacting bug
- Database replication failure

**P2 - Medium** (Response Time: 1 hour)
- Single service degradation
- Non-critical feature failure
- Monitoring alert
- Slow query

**P3 - Low** (Response Time: Next business day)
- Minor bug
- Cosmetic issue
- Documentation update
- Enhancement request

### Incident Response Workflow

1. **Detect**: Alert fires or customer report
2. **Acknowledge**: On-call engineer acknowledges within SLA
3. **Assess**: Determine severity and impact
4. **Escalate**: If P0/P1, notify relevant stakeholders
5. **Communicate**: Update status page and customers
6. **Mitigate**: Apply temporary fix or workaround
7. **Resolve**: Deploy permanent fix
8. **Verify**: Confirm resolution
9. **Document**: Write post-mortem
10. **Learn**: Update runbook and prevent recurrence

---

## Common Incidents & Resolution

### 1. Service Not Responding

**Symptoms**:
- Health check failing
- Timeout errors
- No response from service

**Diagnosis**:
```bash
# Check if pods are running
kubectl get pods | grep tenant-admin

# Check pod logs
kubectl logs tenant-admin-service-xxx-xxx --tail=50

# Check resource usage
kubectl top pod tenant-admin-service-xxx-xxx

# Check events
kubectl describe pod tenant-admin-service-xxx-xxx
```

**Resolution**:
```bash
# If OOMKilled, increase memory limit
kubectl edit deployment tenant-admin-service
# Update resources.limits.memory

# If CrashLoopBackOff, check logs for error
kubectl logs tenant-admin-service-xxx-xxx --previous

# Restart service
kubectl rollout restart deployment/tenant-admin-service

# If database connection issue
kubectl exec -it postgres-0 -- psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle in transaction';"
```

**Prevention**:
- Set appropriate resource limits
- Implement connection pool management
- Add circuit breakers
- Monitor memory usage trends

---

### 2. Database Connection Pool Exhausted

**Symptoms**:
- "too many connections" errors
- Service unable to connect to database
- Slow response times

**Diagnosis**:
```bash
# Check current connections
psql -U postgres -d tenant_admin_db -c "
SELECT count(*) as total,
       count(*) FILTER (WHERE state = 'active') as active,
       count(*) FILTER (WHERE state = 'idle') as idle
FROM pg_stat_activity;"

# Check max connections
psql -U postgres -c "SHOW max_connections;"

# Find long-running queries
psql -U postgres -d tenant_admin_db -c "
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE state != 'idle' AND query_start < now() - interval '5 minutes'
ORDER BY duration DESC;"
```

**Resolution**:
```bash
# Kill long-running queries
psql -U postgres -d tenant_admin_db -c "SELECT pg_terminate_backend(12345);"

# Restart service to reset connection pool
kubectl rollout restart deployment/tenant-admin-service

# Temporary: increase max_connections
# Update PostgreSQL configuration
```

**Prevention**:
- Optimize slow queries
- Implement connection pooling (PgBouncer)
- Set appropriate connection timeouts
- Monitor connection usage

---

### 3. High CPU Usage

**Symptoms**:
- Slow response times
- CPU throttling alerts
- Pod eviction

**Diagnosis**:
```bash
# Check CPU usage
kubectl top pods

# Check for CPU throttling
kubectl describe pod tenant-admin-service-xxx-xxx | grep -A 5 "Limits"

# Get pprof CPU profile
kubectl port-forward tenant-admin-service-xxx-xxx 6060:6060
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

**Resolution**:
```bash
# Immediate: Scale horizontally
kubectl scale deployment tenant-admin-service --replicas=5

# Increase CPU limit (if needed)
kubectl edit deployment tenant-admin-service
# Update resources.limits.cpu

# Check for inefficient code
# Review logs for errors or infinite loops
```

**Prevention**:
- Profile code regularly
- Optimize hot paths
- Implement caching
- Use appropriate algorithms

---

### 4. Memory Leak

**Symptoms**:
- Steadily increasing memory usage
- OOMKilled pods
- Pod restarts

**Diagnosis**:
```bash
# Check memory trends
kubectl top pods --sort-by=memory

# Get heap profile
kubectl port-forward tenant-admin-service-xxx-xxx 6060:6060
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof -http=:8080 heap.prof

# Check goroutine leaks
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

**Resolution**:
```bash
# Immediate: Restart affected pods
kubectl delete pod tenant-admin-service-xxx-xxx

# Increase memory limit (temporary)
kubectl edit deployment tenant-admin-service

# Fix code and redeploy
# Review connection closures, slice reallocations, caching
```

**Prevention**:
- Regular memory profiling
- Proper connection management
- Bounded caches
- Context cancellation

---

### 5. Database Slow Queries

**Symptoms**:
- Slow API responses
- Database CPU spike
- Timeout errors

**Diagnosis**:
```sql
-- Enable slow query log
ALTER SYSTEM SET log_min_duration_statement = 1000; -- 1 second
SELECT pg_reload_conf();

-- Check slow queries
SELECT query, calls, mean_exec_time, max_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 1000
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check blocking queries
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

**Resolution**:
```sql
-- Kill blocking query
SELECT pg_terminate_backend(12345);

-- Add missing index
CREATE INDEX CONCURRENTLY idx_table_column ON table(column);

-- Update statistics
ANALYZE table_name;

-- Rewrite inefficient query
EXPLAIN ANALYZE SELECT ...;
```

**Prevention**:
- Regular query optimization
- Proper indexing strategy
- Query plan review
- Set statement timeout

---

### 6. Redis Connection Failure

**Symptoms**:
- Session validation failures
- Cache misses
- Fallback to database

**Diagnosis**:
```bash
# Check Redis health
redis-cli ping

# Check Redis connections
redis-cli info clients

# Check Redis memory
redis-cli info memory

# Check for slow commands
redis-cli slowlog get 10
```

**Resolution**:
```bash
# Restart Redis (graceful)
kubectl rollout restart deployment/redis

# Check Redis configuration
redis-cli config get maxclients
redis-cli config get maxmemory

# Flush cache if corrupted
redis-cli flushdb
```

**Note**: Services will gracefully fallback to database + in-memory cache

**Prevention**:
- Redis cluster for high availability
- Set maxmemory-policy to allkeys-lru
- Monitor Redis memory usage
- Implement cache warming

---

### 7. API Gateway Errors

**Symptoms**:
- 502/503/504 errors
- Requests not reaching services
- Rate limit issues

**Diagnosis**:
```bash
# Check API Gateway logs
kubectl logs -f deployment/api-gateway

# Check downstream service health
for service in tenant-admin user-service component; do
  echo "Checking $service..."
  kubectl get pods -l app=$service
done

# Check rate limit status
curl -v https://api.beakon.com/health | grep X-RateLimit
```

**Resolution**:
```bash
# Restart API Gateway
kubectl rollout restart deployment/api-gateway

# Adjust rate limits if needed
kubectl edit configmap api-gateway-config

# Check circuit breaker status (in logs)
kubectl logs api-gateway-xxx-xxx | grep "circuit"

# Bypass rate limit for specific IP (temporary)
# Update rate limit whitelist
```

**Prevention**:
- Monitor circuit breaker metrics
- Set appropriate timeouts
- Implement health checks
- Load test regularly

---

### 8. Certificate Expiration

**Symptoms**:
- SSL/TLS errors
- Browser warnings
- Connection refused

**Diagnosis**:
```bash
# Check certificate expiration
echo | openssl s_client -servername api.beakon.com -connect api.beakon.com:443 2>/dev/null | openssl x509 -noout -dates

# Check cert-manager status
kubectl get certificates -n beakon-prod

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager
```

**Resolution**:
```bash
# Manually renew certificate
kubectl delete certificate beakon-tls
kubectl apply -f beakon-tls-certificate.yaml

# Check renewal
kubectl describe certificate beakon-tls
```

**Prevention**:
- Set up certificate expiration alerts (30 days before)
- Use cert-manager for automatic renewal
- Monitor certificate status daily

---

### 9. Disk Space Full

**Symptoms**:
- Pod eviction
- Database write failures
- Log write failures

**Diagnosis**:
```bash
# Check disk usage
kubectl exec -it postgres-0 -- df -h

# Check PVC usage
kubectl get pvc

# Find large files
kubectl exec -it postgres-0 -- du -sh /* | sort -h
```

**Resolution**:
```bash
# Clean up old logs
kubectl exec -it postgres-0 -- find /var/log -name "*.log" -mtime +7 -delete

# Clean up old WAL files
kubectl exec -it postgres-0 -- pg_archivecleanup /var/lib/postgresql/data/pg_wal latest

# Resize PVC
kubectl patch pvc postgres-storage -p '{"spec":{"resources":{"requests":{"storage":"200Gi"}}}}'

# Run VACUUM to reclaim space
kubectl exec -it postgres-0 -- psql -U postgres -d tenant_admin_db -c "VACUUM FULL;"
```

**Prevention**:
- Set up disk space alerts (> 80%)
- Implement log rotation
- Regular database maintenance
- Monitor growth trends

---

### 10. Failed Deployment

**Symptoms**:
- Pods stuck in ImagePullBackOff
- Pods stuck in CrashLoopBackOff
- Rollout stuck

**Diagnosis**:
```bash
# Check rollout status
kubectl rollout status deployment/tenant-admin-service

# Check deployment history
kubectl rollout history deployment/tenant-admin-service

# Check events
kubectl get events --sort-by='.lastTimestamp' | tail -20

# Check pod details
kubectl describe pod tenant-admin-service-xxx-xxx
```

**Resolution**:
```bash
# Rollback to previous version
kubectl rollout undo deployment/tenant-admin-service

# Or rollback to specific revision
kubectl rollout undo deployment/tenant-admin-service --to-revision=3

# Fix and redeploy
# 1. Fix the issue (image, config, etc.)
# 2. Build and push new image
# 3. Update deployment
kubectl set image deployment/tenant-admin-service tenant-admin-service=beakon/tenant-admin-service:v1.2.1
```

**Prevention**:
- Test deployments in staging first
- Use gradual rollouts
- Implement smoke tests
- Keep previous image versions

---

## Maintenance Procedures

### Database Backup

**Schedule**: Daily at 2 AM UTC

```bash
# Manual backup
kubectl exec -it postgres-0 -- pg_dump -U postgres -F c -b -v tenant_admin_db > backup_$(date +%Y%m%d).dump

# Verify backup
pg_restore --list backup_20251014.dump | head -20

# Automated backup (CronJob already configured)
kubectl get cronjob backup-databases
```

### Database Restore

```bash
# Stop services
kubectl scale deployment tenant-admin-service --replicas=0
kubectl scale deployment saas-admin-service --replicas=0

# Drop and recreate database
kubectl exec -it postgres-0 -- psql -U postgres -c "DROP DATABASE tenant_admin_db;"
kubectl exec -it postgres-0 -- psql -U postgres -c "CREATE DATABASE tenant_admin_db;"

# Restore backup
kubectl exec -i postgres-0 -- pg_restore -U postgres -d tenant_admin_db < backup_20251014.dump

# Start services
kubectl scale deployment tenant-admin-service --replicas=3
kubectl scale deployment saas-admin-service --replicas=3

# Verify health
curl http://tenant-admin-service:8099/health
```

### Database Migration

```bash
# 1. Backup before migration
kubectl exec -it postgres-0 -- pg_dump -U postgres tenant_admin_db > pre_migration_backup.dump

# 2. Test migration in staging first

# 3. Run migration in production
kubectl exec -it tenant-admin-service-xxx-xxx -- ./tenant-admin-service --migrate-only

# 4. Verify migration
kubectl exec -it postgres-0 -- psql -U postgres -d tenant_admin_db -c "\dt"

# 5. Restart services
kubectl rollout restart deployment/tenant-admin-service
```

### Certificate Renewal

```bash
# Check certificate status
kubectl get certificate beakon-tls -o yaml

# Manual renewal (if needed)
kubectl delete secret beakon-tls
kubectl delete certificate beakon-tls
kubectl apply -f certificates/beakon-tls.yaml

# Verify
kubectl describe certificate beakon-tls
```

### Scaling for High Traffic

```bash
# Scale up before anticipated traffic spike
kubectl scale deployment tenant-admin-service --replicas=10
kubectl scale deployment api-gateway --replicas=5

# Monitor during event
watch kubectl get pods
watch kubectl top pods

# Scale down after traffic subsides
kubectl scale deployment tenant-admin-service --replicas=3
kubectl scale deployment api-gateway --replicas=2
```

---

## Monitoring & Alerts

### Key Metrics

1. **Service Health**
   - HTTP 2xx/4xx/5xx rates
   - Response time (p50, p95, p99)
   - Error rate

2. **Database**
   - Connection pool usage
   - Query execution time
   - Replication lag
   - Disk usage

3. **Resource Usage**
   - CPU usage
   - Memory usage
   - Network I/O
   - Disk I/O

4. **Business Metrics**
   - Active users
   - API calls per minute
   - Tenant creation rate
   - Incident creation rate

### Alert Thresholds

| Alert | Threshold | Severity | Action |
|-------|-----------|----------|--------|
| Service down | Health check fails | P0 | Restart service |
| High error rate | > 5% | P1 | Investigate logs |
| High latency | p95 > 2s | P1 | Check database |
| DB connections | > 90% | P1 | Investigate queries |
| CPU usage | > 80% | P2 | Scale or optimize |
| Memory usage | > 85% | P2 | Check for leaks |
| Disk space | > 85% | P2 | Clean up or expand |
| Certificate expiry | < 30 days | P2 | Renew certificate |

---

## Post-Incident Review

### Post-Mortem Template

```markdown
# Incident Post-Mortem: [Title]

**Date**: YYYY-MM-DD
**Severity**: P0/P1/P2/P3
**Duration**: X hours Y minutes
**Affected Users**: X (or X%)

## Summary
Brief description of what happened.

## Impact
- Number of affected users
- Services impacted
- Revenue impact (if applicable)
- Customer complaints received

## Timeline
- **HH:MM** - Incident began
- **HH:MM** - Alert fired
- **HH:MM** - Incident acknowledged
- **HH:MM** - Root cause identified
- **HH:MM** - Mitigation applied
- **HH:MM** - Incident resolved
- **HH:MM** - Post-incident verification complete

## Root Cause
Detailed explanation of what caused the incident.

## Resolution
Steps taken to resolve the incident.

## Lessons Learned
What went well:
- ...

What didn't go well:
- ...

## Action Items
- [ ] Item 1 (Owner: Name, Due: Date)
- [ ] Item 2 (Owner: Name, Due: Date)
- [ ] Item 3 (Owner: Name, Due: Date)

## Prevention
How we'll prevent this from happening again.
```

---

## Useful Links

- **Grafana Dashboards**: https://grafana.beakon.com
- **Prometheus**: https://prometheus.beakon.com
- **Kubernetes Dashboard**: https://k8s.beakon.com
- **Status Page**: https://status.beakon.com
- **Documentation**: https://docs.beakon.com
- **GitHub**: https://github.com/beakon/beakon
- **Slack**: #beakon-engineering

---

## Runbook Updates

This runbook should be updated:
- After every incident (add new scenarios)
- When procedures change
- When new services are added
- When contact information changes

**Last Review**: 2025-10-14
**Next Review**: 2026-01-14 (Quarterly)
**Maintained By**: DevOps Team

---

**Remember**: When in doubt, escalate early. It's better to escalate unnecessarily than to delay resolution.
