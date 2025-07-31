# Ponos Operations Guide

This guide covers operational procedures for running Ponos aggregator and executor services in production with persistent storage.

## Table of Contents

1. [Pre-deployment Checklist](#pre-deployment-checklist)
2. [Deployment Procedures](#deployment-procedures)
3. [Monitoring and Health Checks](#monitoring-and-health-checks)
4. [Backup Procedures](#backup-procedures)
5. [Recovery Procedures](#recovery-procedures)
6. [Maintenance Operations](#maintenance-operations)
7. [Troubleshooting](#troubleshooting)

## Pre-deployment Checklist

### System Requirements
- [ ] Sufficient disk space (minimum 10GB for BadgerDB)
- [ ] SSD storage recommended for BadgerDB
- [ ] Appropriate file system permissions
- [ ] Docker/Kubernetes environment ready
- [ ] Network connectivity to blockchain RPC endpoints

### Configuration Validation
- [ ] Valid operator keys and addresses
- [ ] Correct RPC endpoints for all chains
- [ ] Storage directories configured
- [ ] Environment variables set
- [ ] Contract addresses verified

### Security Checklist
- [ ] Private keys secured (use remote signer in production)
- [ ] Storage directories have restrictive permissions (750)
- [ ] Network policies configured
- [ ] Firewall rules in place
- [ ] TLS certificates for gRPC if exposed

## Deployment Procedures

### Initial Deployment

1. **Create storage directories**:
```bash
# As root or with sudo
mkdir -p /var/lib/ponos/{aggregator,executor}/badger
useradd -r -s /bin/false ponos
chown -R ponos:ponos /var/lib/ponos
chmod -R 750 /var/lib/ponos
```

2. **Deploy with Docker Compose**:
```bash
# Update docker-compose.yml with persistent volumes
docker-compose up -d aggregator executor

# Verify services are running
docker-compose ps
docker-compose logs -f aggregator executor
```

3. **Verify storage initialization**:
```bash
# Check that BadgerDB directories are created
ls -la /var/lib/ponos/*/badger/
# Should see MANIFEST, 000001.vlog, etc.
```

### Rolling Updates

For zero-downtime updates with persistent storage:

1. **Aggregator Update** (can have brief downtime):
```bash
# Stop aggregator gracefully
docker-compose stop aggregator

# Update image
docker-compose pull aggregator

# Start with new version
docker-compose up -d aggregator

# Verify recovery from storage
docker-compose logs aggregator | grep -E "(Recovered|Loaded|storage)"
```

2. **Executor Update** (requires careful coordination):
```bash
# List current performers
grpcurl -plaintext localhost:9090 eigenlayer.hourglass.v1.ExecutorService/ListPerformers

# Stop executor gracefully
docker-compose stop executor

# Update and restart
docker-compose pull executor
docker-compose up -d executor

# Verify performers recovered
grpcurl -plaintext localhost:9090 eigenlayer.hourglass.v1.ExecutorService/ListPerformers
```

## Monitoring and Health Checks

### Service Health Checks

1. **Aggregator Health**:
```bash
# Check if aggregator is processing blocks
docker-compose logs aggregator | tail -20 | grep "Processing block"

# Check last processed block from storage
# (Future: Add admin API endpoint)
```

2. **Executor Health**:
```bash
# Check gRPC endpoint
grpcurl -plaintext localhost:9090 list

# Check performer status
grpcurl -plaintext localhost:9090 eigenlayer.hourglass.v1.ExecutorService/ListPerformers
```

### Storage Monitoring

1. **Disk Usage**:
```bash
# Monitor storage growth
df -h /var/lib/ponos

# Check BadgerDB size
du -sh /var/lib/ponos/*/badger

# Set up alerts for > 80% disk usage
```

2. **BadgerDB Health**:
```bash
# Check for corruption (logs)
docker-compose logs aggregator executor | grep -i "badger.*error"

# Monitor garbage collection
docker-compose logs aggregator executor | grep "value log GC"
```

### Metrics to Monitor

- Block processing lag (current block vs chain head)
- Task processing rate and latency
- Storage size growth rate
- Memory usage
- CPU usage
- Network connectivity to RPC endpoints

## Backup Procedures

### Automated Backups

Create a backup script `/usr/local/bin/ponos-backup.sh`:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/ponos/$(date +%Y%m%d_%H%M%S)"
RETENTION_DAYS=7

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Stop services for consistent backup
docker-compose stop aggregator executor

# Backup BadgerDB directories
tar -czf "$BACKUP_DIR/aggregator-badger.tar.gz" -C /var/lib/ponos/aggregator badger
tar -czf "$BACKUP_DIR/executor-badger.tar.gz" -C /var/lib/ponos/executor badger

# Backup configurations
cp /etc/ponos/*.yaml "$BACKUP_DIR/"

# Restart services
docker-compose up -d aggregator executor

# Clean old backups
find /backup/ponos -type d -mtime +$RETENTION_DAYS -exec rm -rf {} \;

echo "Backup completed: $BACKUP_DIR"
```

Schedule with cron:
```bash
# Daily backup at 2 AM
0 2 * * * /usr/local/bin/ponos-backup.sh >> /var/log/ponos-backup.log 2>&1
```

### Manual Backup

For immediate backup without downtime:

```bash
# Create snapshot using BadgerDB backup (future feature)
# For now, use filesystem snapshots if available (LVM, ZFS, etc.)
```

## Recovery Procedures

### Service Recovery After Crash

1. **Automatic Recovery** (with persistent storage):
```bash
# Services should recover automatically on restart
docker-compose up -d aggregator executor

# Verify recovery
docker-compose logs aggregator | grep -A5 "Recovering from storage"
docker-compose logs executor | grep -A5 "Loading performer states"
```

2. **Manual Recovery Steps**:
```bash
# If automatic recovery fails

# 1. Check storage integrity
ls -la /var/lib/ponos/*/badger/

# 2. Check logs for errors
docker-compose logs aggregator executor | grep -i error

# 3. If corrupted, restore from backup
systemctl stop ponos-aggregator ponos-executor
rm -rf /var/lib/ponos/*/badger
tar -xzf /backup/ponos/latest/aggregator-badger.tar.gz -C /var/lib/ponos/aggregator/
tar -xzf /backup/ponos/latest/executor-badger.tar.gz -C /var/lib/ponos/executor/
systemctl start ponos-aggregator ponos-executor
```

### Disaster Recovery

Complete system recovery from backup:

1. **Prepare new system** with same configuration
2. **Restore data**:
```bash
# Copy backup to new system
scp -r backup-server:/backup/ponos/latest/* /tmp/ponos-restore/

# Stop services
docker-compose down

# Restore storage
mkdir -p /var/lib/ponos/{aggregator,executor}
tar -xzf /tmp/ponos-restore/aggregator-badger.tar.gz -C /var/lib/ponos/aggregator/
tar -xzf /tmp/ponos-restore/executor-badger.tar.gz -C /var/lib/ponos/executor/

# Restore configs
cp /tmp/ponos-restore/*.yaml /etc/ponos/

# Set permissions
chown -R ponos:ponos /var/lib/ponos
chmod -R 750 /var/lib/ponos

# Start services
docker-compose up -d
```

### Data Corruption Recovery

If BadgerDB corruption is detected:

1. **Try automatic recovery** (BadgerDB has self-healing):
```bash
# BadgerDB will attempt recovery on startup
docker-compose restart aggregator executor
```

2. **Manual corruption fix**:
```bash
# Stop service
docker-compose stop aggregator

# Use BadgerDB tools (if available)
# badger info --dir /var/lib/ponos/aggregator/badger
# badger flatten --dir /var/lib/ponos/aggregator/badger

# If unfixable, restore from backup
```

## Maintenance Operations

### Storage Maintenance

1. **Garbage Collection** (automatic every 5 minutes)
   - Monitor logs for GC activity
   - No manual intervention needed

2. **Compaction** (if needed):
```bash
# Stop service for full compaction
docker-compose stop aggregator
# Future: Use BadgerDB compaction tool
docker-compose start aggregator
```

3. **Storage Migration**:
```bash
# To migrate to new storage location
# 1. Stop service
# 2. Copy BadgerDB directory to new location
# 3. Update configuration
# 4. Start service
```

### Log Rotation

Configure log rotation for Ponos logs:

```bash
# /etc/logrotate.d/ponos
/var/log/ponos/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 ponos ponos
    sharedscripts
    postrotate
        docker-compose kill -s USR1 aggregator executor
    endscript
}
```

### Performance Tuning

1. **BadgerDB Tuning**:
```yaml
storage:
  type: "badger"
  badger:
    dir: "/var/lib/ponos/aggregator/badger"
    valueLogFileSize: 2147483648  # 2GB for large datasets
    numVersionsToKeep: 1          # Only latest version
    numLevelZeroTables: 10        # Increase for write-heavy loads
    numLevelZeroTablesStall: 20   # Increase stall threshold
```

2. **Memory Allocation**:
```yaml
# docker-compose.yml
services:
  aggregator:
    mem_limit: 4g
    environment:
      - GOGC=100  # Default GC target
      - GOMEMLIMIT=3500MiB  # Leave headroom
```

## Troubleshooting

### Common Issues

1. **Storage Permission Errors**:
```
Error: failed to open badger db: cannot create directory: permission denied
```
Solution:
```bash
chown -R ponos:ponos /var/lib/ponos
chmod -R 750 /var/lib/ponos
```

2. **Disk Space Issues**:
```
Error: no space left on device
```
Solution:
- Check disk usage: `df -h`
- Clean old logs: `docker-compose logs --tail=0 --follow`
- Increase disk space or move to larger volume

3. **BadgerDB Lock Issues**:
```
Error: resource temporarily unavailable
```
Solution:
```bash
# Ensure no other process is using the directory
lsof /var/lib/ponos/*/badger
# Remove lock file if stale
rm -f /var/lib/ponos/*/badger/LOCK
```

4. **Memory Issues**:
```
Error: out of memory
```
Solution:
- Increase container memory limits
- Tune GOGC and GOMEMLIMIT
- Check for memory leaks in logs

5. **Recovery Failures**:
```
Error: failed to recover from storage
```
Solution:
- Check storage integrity
- Verify configuration matches
- Restore from backup if needed

### Debug Commands

```bash
# Check service status
docker-compose ps
docker-compose logs --tail=100 aggregator executor

# Inspect storage
ls -la /var/lib/ponos/*/badger/
file /var/lib/ponos/*/badger/MANIFEST

# Check resource usage
docker stats aggregator executor

# Network connectivity
docker-compose exec aggregator ping -c 3 8.8.8.8
docker-compose exec aggregator curl -s http://eth-mainnet.g.alchemy.com/v2/demo

# Detailed logs
docker-compose logs aggregator executor | grep -E "(ERROR|WARN|storage|recovery)"
```

### Emergency Procedures

1. **Complete Service Failure**:
   - Switch to backup system (if available)
   - Restore from latest backup
   - Investigate root cause from logs

2. **Data Loss**:
   - Stop all writes immediately
   - Assess extent of data loss
   - Restore from backup
   - Replay missing blocks if possible

3. **Security Breach**:
   - Rotate all keys immediately
   - Audit access logs
   - Restore from known-good backup
   - Implement additional security measures

## Best Practices

1. **Regular Backups**: Daily automated backups with 7-day retention
2. **Monitoring**: Set up alerts for disk usage, memory, and service health
3. **Testing**: Regularly test recovery procedures
4. **Documentation**: Keep runbooks updated with lessons learned
5. **Security**: Use remote signers, restrict permissions, enable audit logging
6. **Capacity Planning**: Monitor growth trends and plan for scaling

## Support

For additional support:
- Check logs thoroughly before escalating
- Gather system information (configs, logs, metrics)
- Document steps to reproduce issues
- Contact team with detailed information