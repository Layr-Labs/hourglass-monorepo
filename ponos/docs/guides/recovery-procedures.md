# Ponos Recovery Procedures

This document provides detailed procedures for recovering Ponos services after various failure scenarios.

## Quick Reference

| Scenario | Recovery Time | Data Loss | Procedure |
|----------|--------------|-----------|-----------|
| Service Crash | < 1 minute | None | [Automatic Recovery](#automatic-recovery) |
| Storage Corruption | 5-30 minutes | Possible | [Storage Recovery](#storage-corruption-recovery) |
| Complete System Failure | 30-60 minutes | None* | [Full System Recovery](#full-system-recovery) |
| Data Center Loss | 2-4 hours | None* | [Disaster Recovery](#disaster-recovery) |

*Assuming recent backups are available

## Automatic Recovery

When services crash but storage remains intact, Ponos will automatically recover on restart.

### Aggregator Automatic Recovery

```bash
# 1. Service restarts (via systemd, kubernetes, or docker)
docker-compose up -d aggregator

# 2. Aggregator automatically:
#    - Loads last processed block for each chain
#    - Loads pending tasks from storage
#    - Resumes chain polling from last block
#    - Re-queues pending tasks for processing

# 3. Verify recovery
docker-compose logs aggregator | grep -E "(Recovered|Loaded from storage|Resuming from block)"

# Expected output:
# "Loaded last processed block from storage: chain=1 block=18945123"
# "Recovered 3 pending tasks from storage"
# "Resuming chain polling from block 18945123"
```

### Executor Automatic Recovery

```bash
# 1. Service restarts
docker-compose up -d executor

# 2. Executor automatically:
#    - Loads performer states
#    - Verifies container health
#    - Loads inflight tasks
#    - Resumes task processing

# 3. Verify recovery
docker-compose logs executor | grep -E "(Loaded performer|Recovered inflight|Container health)"

# Expected output:
# "Loaded 2 performer states from storage"
# "Verified container health for performer-123: healthy"
# "Recovered 1 inflight tasks from storage"
```

## Manual Recovery Procedures

### Storage Corruption Recovery

When BadgerDB detects corruption or fails to open:

```bash
# 1. Stop the affected service
docker-compose stop aggregator  # or executor

# 2. Check corruption extent
ls -la /var/lib/ponos/aggregator/badger/
# Look for MANIFEST, *.sst, *.vlog files

# 3. Attempt BadgerDB recovery (built-in)
# Simply restart - BadgerDB will attempt self-repair
docker-compose up -d aggregator

# 4. If self-repair fails, check logs
docker-compose logs aggregator | grep -i "badger"

# 5. If corruption persists, restore from backup
cd /var/lib/ponos/aggregator
mv badger badger.corrupted
tar -xzf /backup/ponos/latest/aggregator-badger.tar.gz
chown -R ponos:ponos badger
chmod -R 750 badger

# 6. Start service
docker-compose up -d aggregator

# 7. Verify recovery and check for data gaps
# May need to replay some blocks if backup is old
```

### Partial Data Recovery

When some data is corrupted but service still runs:

```bash
# 1. Identify affected data
docker-compose logs aggregator | grep -E "(ERROR|corrupted|invalid)"

# 2. For task corruption:
# - Mark corrupted tasks as failed
# - They will be retried by AVS if needed

# 3. For block state corruption:
# - Identify last known good block
# - Reset to earlier block and replay
# - Use admin API (future feature) or manual intervention

# 4. For config corruption:
# - Configs will be re-fetched from contracts
# - No manual intervention needed
```

## Full System Recovery

When the entire system needs to be rebuilt:

### Prerequisites
- Access to recent backups
- Configuration files
- Private keys (or remote signer access)
- Network connectivity

### Step-by-Step Recovery

```bash
# 1. Prepare new system
apt-get update && apt-get install -y docker docker-compose
mkdir -p /var/lib/ponos/{aggregator,executor}/badger
mkdir -p /etc/ponos

# 2. Restore configurations
scp backup-server:/backup/ponos/latest/*.yaml /etc/ponos/

# 3. Restore storage data
cd /tmp
scp backup-server:/backup/ponos/latest/*-badger.tar.gz .
tar -xzf aggregator-badger.tar.gz -C /var/lib/ponos/aggregator/
tar -xzf executor-badger.tar.gz -C /var/lib/ponos/executor/

# 4. Set permissions
useradd -r -s /bin/false ponos
chown -R ponos:ponos /var/lib/ponos /etc/ponos
chmod -R 750 /var/lib/ponos
chmod 640 /etc/ponos/*.yaml

# 5. Deploy services
cd /opt/ponos
docker-compose up -d

# 6. Verify recovery
docker-compose ps
docker-compose logs --tail=100 aggregator executor

# 7. Check data continuity
# Compare last processed block with backup metadata
cat /backup/ponos/latest/metadata.json
docker-compose logs aggregator | grep "last processed block"
```

### Post-Recovery Validation

```bash
# 1. Verify aggregator is processing new blocks
watch -n 5 'docker-compose logs --tail=20 aggregator | grep "Processing block"'

# 2. Verify executor has active performers
grpcurl -plaintext localhost:9090 eigenlayer.hourglass.v1.ExecutorService/ListPerformers

# 3. Submit test task (if possible)
grpcurl -plaintext -d '{"avsAddress": "0x...", "taskId": "0xtest...", "payload": "..."}' \
  localhost:9090 eigenlayer.hourglass.v1.ExecutorService/SubmitTask

# 4. Check metrics/monitoring
curl -s localhost:8080/metrics | grep ponos_
```

## Disaster Recovery

For complete data center or region failure:

### Preparation (Before Disaster)

1. **Off-site Backups**:
```bash
# Automated off-site backup script
#!/bin/bash
# After local backup completes
rsync -avz /backup/ponos/ remote-backup-server:/backup/ponos/
# Or use cloud storage
aws s3 sync /backup/ponos/ s3://ponos-backups/
```

2. **Configuration Management**:
- Store configs in version control
- Encrypt sensitive data
- Document all dependencies

3. **Runbook Maintenance**:
- Test DR procedures quarterly
- Update contact information
- Document RTO/RPO requirements

### Disaster Recovery Execution

```bash
# 1. Activate DR site
# - Provision infrastructure (terraform/ansible)
# - Configure network (DNS, load balancers)

# 2. Restore latest off-site backup
aws s3 sync s3://ponos-backups/latest/ /tmp/restore/

# 3. Deploy fresh instances
# Use infrastructure-as-code
terraform apply -var="environment=dr"

# 4. Restore data
ansible-playbook restore-ponos.yml -e "backup_path=/tmp/restore"

# 5. Update DNS/configuration
# Point clients to DR site

# 6. Verify functionality
./dr-validation-tests.sh

# 7. Monitor closely for 24 hours
# Watch for any data inconsistencies
```

## Recovery Scenarios

### Scenario 1: Aggregator Crash During Task Processing

**Symptoms**: Aggregator crashes while processing tasks

**Recovery**:
```bash
# Automatic recovery handles this
# On restart:
# - Pending tasks are reloaded from storage
# - Tasks stuck in "processing" may timeout and retry
# - No manual intervention needed
```

### Scenario 2: Executor Loses Performer Containers

**Symptoms**: Docker containers deleted or corrupted

**Recovery**:
```bash
# 1. Executor detects missing containers on startup
docker-compose logs executor | grep "Container not found"

# 2. Automatic re-deployment triggered
# Executor will redeploy based on saved performer state

# 3. If automatic deployment fails
grpcurl -plaintext -d '{"image": {...}, "avsAddress": "0x..."}' \
  localhost:9090 eigenlayer.hourglass.v1.ExecutorService/DeployArtifact
```

### Scenario 3: Chain Reorganization

**Symptoms**: Blockchain reorg invalidates processed blocks

**Recovery**:
```bash
# 1. Aggregator detects reorg
# "Block reorganization detected"

# 2. Automatic handling:
# - Reverts to last common block
# - Re-processes affected blocks
# - Updates task states if needed

# 3. Monitor for task conflicts
docker-compose logs aggregator | grep -E "(reorg|conflict)"
```

### Scenario 4: Network Partition

**Symptoms**: Services can't reach blockchain RPC

**Recovery**:
```bash
# 1. Services continue from last state
# Tasks remain pending until connection restored

# 2. When network recovers:
# - Automatic reconnection
# - Catch up on missed blocks
# - Process queued tasks

# 3. Check for extended downtime
# May need to increase task deadlines
```

## Recovery Testing

### Monthly Recovery Drill

```bash
#!/bin/bash
# recovery-drill.sh

echo "Starting recovery drill..."

# 1. Create test backup
/usr/local/bin/ponos-backup.sh

# 2. Stop services
docker-compose stop aggregator executor

# 3. Simulate corruption
mv /var/lib/ponos/aggregator/badger /var/lib/ponos/aggregator/badger.drill

# 4. Restore from backup
cd /var/lib/ponos/aggregator
tar -xzf /backup/ponos/latest/aggregator-badger.tar.gz

# 5. Start services
docker-compose up -d aggregator executor

# 6. Validate recovery
sleep 30
if docker-compose ps | grep -q "Up"; then
    echo "Recovery successful"
else
    echo "Recovery failed!"
    exit 1
fi

# 7. Cleanup
rm -rf /var/lib/ponos/aggregator/badger.drill
```

### Recovery Metrics to Track

- **RTO (Recovery Time Objective)**: Target < 30 minutes
- **RPO (Recovery Point Objective)**: Target < 1 hour of data
- **Success Rate**: Track drill success percentage
- **Time to Detect**: How quickly issues are identified
- **Time to Recover**: Actual recovery duration

## Preventive Measures

1. **High Availability Setup**:
   - Run multiple aggregator instances (active/passive)
   - Load balance executor instances
   - Use shared storage or replication

2. **Monitoring and Alerting**:
   - Block processing lag > 5 minutes
   - Storage errors in logs
   - Disk usage > 80%
   - Service health checks failing

3. **Regular Maintenance**:
   - Weekly backup verification
   - Monthly recovery drills
   - Quarterly DR tests
   - Annual runbook review

## Recovery Decision Tree

```
Service Down?
├─ Yes
│  ├─ Storage Accessible?
│  │  ├─ Yes → Restart Service (Automatic Recovery)
│  │  └─ No
│  │     ├─ Corruption Detected?
│  │     │  ├─ Yes → Restore from Backup
│  │     │  └─ No → Check Permissions/Disk Space
│  │     └─ Fix Issue and Restart
│  └─ Monitor Recovery
└─ No
   └─ Check Performance
      ├─ Degraded?
      │  ├─ Yes → Check Logs for Errors
      │  └─ No → Normal Operation
      └─ Continue Monitoring
```

## Contact Information

**Escalation Path**:
1. On-call Engineer
2. Team Lead
3. Infrastructure Team
4. Security Team (if breach suspected)

**Key Resources**:
- Monitoring Dashboard: [URL]
- Backup Location: /backup/ponos/
- Configuration Repo: [Git URL]
- Runbook Updates: [Wiki URL]