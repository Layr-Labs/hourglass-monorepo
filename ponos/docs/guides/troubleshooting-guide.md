# Ponos Troubleshooting Guide

This guide helps diagnose and resolve common issues with Ponos aggregator and executor services.

## Quick Diagnostics

Run this script for initial diagnostics:

```bash
#!/bin/bash
echo "=== Ponos Diagnostics ==="
echo "1. Service Status:"
docker-compose ps

echo -e "\n2. Storage Status:"
ls -la /var/lib/ponos/*/badger/ 2>/dev/null || echo "Storage directories not found"
df -h /var/lib/ponos

echo -e "\n3. Recent Errors:"
docker-compose logs --tail=50 aggregator executor | grep -i error

echo -e "\n4. Resource Usage:"
docker stats --no-stream aggregator executor

echo -e "\n5. Network Connectivity:"
docker-compose exec aggregator ping -c 1 8.8.8.8 >/dev/null 2>&1 && echo "✓ Internet accessible" || echo "✗ No internet"
```

## Common Issues

### 1. Service Won't Start

#### Symptom: Container exits immediately

**Check logs:**
```bash
docker-compose logs --tail=100 aggregator
```

**Common causes and solutions:**

a) **Configuration errors**
```
Error: failed to load config: yaml: line 10: found character that cannot start any token
```
Solution:
```bash
# Validate YAML syntax
yamllint /etc/ponos/aggregator.yaml

# Common issues:
# - Tabs instead of spaces
# - Missing quotes around values
# - Incorrect indentation
```

b) **Missing required fields**
```
Error: validation failed: operator.address is required
```
Solution:
```bash
# Check all required fields are present
# Refer to README.md for complete configuration example
```

c) **Permission denied**
```
Error: open /var/lib/ponos/aggregator/badger: permission denied
```
Solution:
```bash
sudo chown -R ponos:ponos /var/lib/ponos
sudo chmod -R 750 /var/lib/ponos
# If running as different user, adjust accordingly
```

### 2. Storage Issues

#### BadgerDB Won't Open

**Error:**
```
Error: failed to open badger db: Cannot acquire directory lock
```

**Solutions:**

a) **Check for stale lock:**
```bash
# Check if another process is using the directory
lsof /var/lib/ponos/aggregator/badger

# If no process found, remove stale lock
rm -f /var/lib/ponos/aggregator/badger/LOCK
```

b) **Corruption detected:**
```
Error: checksum mismatch
```
```bash
# BadgerDB will attempt auto-recovery
# If it fails, restore from backup
cd /var/lib/ponos/aggregator
mv badger badger.corrupted
tar -xzf /backup/ponos/latest/aggregator-badger.tar.gz
```

#### Disk Space Issues

**Error:**
```
Error: no space left on device
```

**Diagnostics:**
```bash
# Check disk usage
df -h /var/lib/ponos

# Find large files
du -sh /var/lib/ponos/* | sort -h

# Check BadgerDB size
du -sh /var/lib/ponos/*/badger/
```

**Solutions:**
```bash
# 1. Clean up logs
docker-compose logs --tail=0 --follow > /dev/null

# 2. Trigger manual GC (BadgerDB runs GC automatically)
# Restart service to trigger GC cycle

# 3. Move to larger disk
# Stop services, move data, update config, restart
```

### 3. Network/RPC Issues

#### Cannot Connect to Blockchain RPC

**Error:**
```
Error: dial tcp: lookup eth-mainnet.g.alchemy.com: no such host
```

**Diagnostics:**
```bash
# Test DNS resolution
docker-compose exec aggregator nslookup eth-mainnet.g.alchemy.com

# Test RPC endpoint
docker-compose exec aggregator curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://your-rpc-endpoint
```

**Solutions:**
```bash
# 1. Check DNS configuration
cat /etc/resolv.conf

# 2. Test with different RPC endpoint
# Update config with working endpoint

# 3. Add retry logic (built-in, but check settings)
```

#### RPC Rate Limiting

**Error:**
```
Error: 429 Too Many Requests
```

**Solutions:**
```yaml
# Adjust polling interval in config
chains:
  - name: "mainnet"
    pollIntervalSeconds: 15  # Increase from default
```

### 4. Task Processing Issues

#### Tasks Stuck in Pending

**Diagnostics:**
```bash
# Check pending tasks (via logs)
docker-compose logs aggregator | grep -i "pending tasks"

# Check if aggregator is processing blocks
docker-compose logs --tail=50 aggregator | grep "Processing block"
```

**Solutions:**
```bash
# 1. Check if chain polling is working
# 2. Verify operator set configuration
# 3. Check task deadlines haven't expired
```

#### Task Submission Failures

**Error:**
```
Error: rpc error: code = DeadlineExceeded
```

**Solutions:**
```bash
# 1. Check executor is running
grpcurl -plaintext localhost:9090 list

# 2. Check performer is deployed
grpcurl -plaintext localhost:9090 eigenlayer.hourglass.v1.ExecutorService/ListPerformers

# 3. Check network connectivity between services
```

### 5. Memory Issues

#### Out of Memory Errors

**Symptoms:**
- Container gets OOMKilled
- Gradual performance degradation

**Diagnostics:**
```bash
# Check memory usage
docker stats aggregator executor

# Check for OOM kills
dmesg | grep -i "killed process"
journalctl -u docker --since "1 hour ago" | grep -i oom
```

**Solutions:**

a) **Increase memory limits:**
```yaml
# docker-compose.yml
services:
  aggregator:
    mem_limit: 4g
    environment:
      - GOGC=100
      - GOMEMLIMIT=3500MiB
```

b) **Tune garbage collection:**
```bash
# More aggressive GC
export GOGC=50  # Default is 100
```

c) **Check for memory leaks:**
```bash
# Monitor memory growth over time
while true; do
  docker stats --no-stream aggregator
  sleep 60
done
```

### 6. Performance Issues

#### Slow Block Processing

**Diagnostics:**
```bash
# Check processing lag
# Compare latest block in logs vs current chain height

# Check processing time per block
docker-compose logs aggregator | grep -E "Processing block.*took"
```

**Solutions:**

a) **Optimize RPC calls:**
- Use batch requests where possible
- Add caching layer
- Use dedicated nodes

b) **Tune BadgerDB:**
```yaml
storage:
  badger:
    valueLogFileSize: 2147483648  # Increase for better performance
    numLevelZeroTables: 10        # Increase for write-heavy loads
```

#### High CPU Usage

**Diagnostics:**
```bash
# Check CPU usage by container
docker stats aggregator executor

# Profile the application (if debug enabled)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
```

**Solutions:**
- Check for infinite loops in logs
- Reduce polling frequency
- Optimize task processing logic

### 7. Docker/Container Issues

#### Container Network Issues

**Error:**
```
Error: cannot connect to performer container
```

**Diagnostics:**
```bash
# List networks
docker network ls

# Inspect network
docker network inspect ponos_default

# Test connectivity
docker-compose exec aggregator ping executor
```

**Solutions:**
```bash
# Recreate network
docker-compose down
docker network prune
docker-compose up -d
```

#### Volume Mount Issues

**Error:**
```
Error: mount denied: the source path doesn't exist
```

**Solutions:**
```bash
# Create missing directories
sudo mkdir -p /var/lib/ponos/{aggregator,executor}/badger

# Fix permissions
sudo chown -R $USER:$USER /var/lib/ponos  # or appropriate user
```

## Advanced Debugging

### Enable Debug Logging

```yaml
# In config files
debug: true

# Or via environment
environment:
  - LOG_LEVEL=debug
```

### Trace Specific Operations

```bash
# Follow specific task
docker-compose logs -f aggregator | grep -i "task-id-123"

# Trace block processing
docker-compose logs -f aggregator | grep -E "block (18945123|18945124)"
```

### BadgerDB Inspection

```bash
# Get BadgerDB info (future tool)
badger info --dir /var/lib/ponos/aggregator/badger

# List keys (be careful, can be large)
badger keys --dir /var/lib/ponos/aggregator/badger | head -20
```

### Network Traffic Analysis

```bash
# Capture RPC traffic
tcpdump -i any -w ponos.pcap host your-rpc-endpoint

# Analyze with Wireshark
wireshark ponos.pcap
```

## Performance Tuning Checklist

1. **Storage Optimization**:
   - [ ] Using SSD for BadgerDB
   - [ ] Appropriate valueLogFileSize
   - [ ] Regular garbage collection working

2. **Memory Management**:
   - [ ] Adequate memory limits set
   - [ ] GOGC tuned for workload
   - [ ] No memory leaks detected

3. **Network Optimization**:
   - [ ] Using local/dedicated RPC nodes
   - [ ] Batch operations where possible
   - [ ] Connection pooling enabled

4. **Resource Allocation**:
   - [ ] CPU limits appropriate
   - [ ] I/O limits not constraining
   - [ ] Network bandwidth sufficient

## Monitoring Checklist

Essential metrics to monitor:

1. **Service Health**:
   - [ ] Container status (up/down)
   - [ ] Restart count
   - [ ] Health check status

2. **Performance Metrics**:
   - [ ] Block processing lag
   - [ ] Task processing time
   - [ ] Queue depths

3. **Resource Metrics**:
   - [ ] CPU usage
   - [ ] Memory usage
   - [ ] Disk I/O
   - [ ] Network I/O

4. **Application Metrics**:
   - [ ] Tasks processed/failed
   - [ ] RPC call rates
   - [ ] Error rates

## Getting Help

When reporting issues, include:

1. **Environment details**:
```bash
docker version
docker-compose version
uname -a
df -h
free -m
```

2. **Configuration** (sanitized):
```bash
# Remove private keys!
cat /etc/ponos/*.yaml | grep -v -E "(private|key|secret)"
```

3. **Recent logs**:
```bash
docker-compose logs --tail=500 aggregator executor > ponos-logs.txt
```

4. **Diagnostic output**:
```bash
# Run the diagnostic script from top of this guide
./ponos-diagnostics.sh > diagnostics.txt
```

5. **Steps to reproduce**:
- What were you doing when the issue occurred?
- Can you reproduce it consistently?
- What changed recently?

## Emergency Contacts

- On-call: [Phone/Slack]
- Team Lead: [Contact]
- Infrastructure: [Contact]
- Security: [Contact] (for security issues only)