# Memory Optimization Migration Guide

**Version**: 2.34.0  
**Audience**: EntityDB Users and Developers  
**Last Updated**: 2025-06-22

## Overview

EntityDB v2.34.0 introduces comprehensive memory optimizations that prevent the high memory utilization issues experienced in previous versions. This guide helps you migrate to the new version and take advantage of these improvements.

## What's Changed

### Automatic Memory Management

Previous versions had unbounded memory growth. Version 2.34.0 introduces:

- **Bounded Caches**: String and entity caches now have size and memory limits
- **Automatic Eviction**: LRU (Least Recently Used) algorithm removes old data
- **Memory Monitoring**: System tracks memory pressure and responds automatically
- **Pressure Relief**: Automatic cleanup when memory usage is high

### Default Behavior Changes

| Feature | Previous | v2.34.0 |
|---------|----------|---------|
| String Cache | Unlimited | 100,000 strings / 100MB |
| Entity Cache | Unlimited | 10,000 entities / 1GB |
| Metrics Retention | Forever | 24 hours raw, 7 days aggregated |
| Memory Monitoring | None | Every 30 seconds |
| Pressure Response | None | Automatic at 80% memory |

## Migration Steps

### 1. Pre-Migration Health Check

Before upgrading, assess your current memory usage:

```bash
# Check current memory usage
curl http://localhost:8085/health | jq '.memory'

# Count entities
curl http://localhost:8085/api/v1/dashboard/stats | jq '.totalEntities'

# Check metric volume
curl http://localhost:8085/api/v1/entities/query?tag=type:metric | jq '. | length'
```

### 2. Backup Your Data

```bash
# Stop the service
systemctl stop entitydb

# Backup the data directory
tar -czf entitydb-backup-$(date +%Y%m%d).tar.gz /opt/entitydb/var/

# Keep the backup safe
mv entitydb-backup-*.tar.gz /backup/location/
```

### 3. Configuration Planning

Based on your current usage, choose appropriate limits:

#### Small Instances (<2GB RAM)
```bash
# /etc/entitydb/entitydb.env
ENTITYDB_STRING_CACHE_SIZE=50000
ENTITYDB_STRING_CACHE_MEMORY_LIMIT=52428800     # 50MB
ENTITYDB_ENTITY_CACHE_SIZE=5000
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=524288000    # 500MB
```

#### Medium Instances (4-8GB RAM)
```bash
# /etc/entitydb/entitydb.env
ENTITYDB_STRING_CACHE_SIZE=100000               # Default
ENTITYDB_STRING_CACHE_MEMORY_LIMIT=104857600    # 100MB
ENTITYDB_ENTITY_CACHE_SIZE=10000                # Default
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=1073741824   # 1GB
```

#### Large Instances (>16GB RAM)
```bash
# /etc/entitydb/entitydb.env
ENTITYDB_STRING_CACHE_SIZE=500000
ENTITYDB_STRING_CACHE_MEMORY_LIMIT=524288000    # 500MB
ENTITYDB_ENTITY_CACHE_SIZE=50000
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=5368709120   # 5GB
```

### 4. Upgrade Process

```bash
# Download new version
wget https://github.com/entitydb/releases/v2.34.0/entitydb-linux-amd64.tar.gz
tar -xzf entitydb-linux-amd64.tar.gz

# Install new binary
sudo mv entitydb /opt/entitydb/bin/entitydb

# Update configuration
sudo vim /etc/entitydb/entitydb.env  # Add memory settings

# Start the service
sudo systemctl start entitydb

# Verify it's running
sudo systemctl status entitydb
```

### 5. Post-Migration Validation

```bash
# Check version
curl http://localhost:8085/health | jq '.version'

# Monitor memory usage (should be stable)
watch -n 5 'curl -s http://localhost:8085/health | jq .memory.heapInUse'

# Check cache performance
curl http://localhost:8085/api/v1/system/metrics | jq '.performance'
```

## Configuration Reference

### Essential Memory Settings

```bash
# String interning cache (for tag deduplication)
ENTITYDB_STRING_CACHE_SIZE=100000              # Number of unique strings
ENTITYDB_STRING_CACHE_MEMORY_LIMIT=104857600   # Bytes (100MB)

# Entity cache (for frequently accessed entities)
ENTITYDB_ENTITY_CACHE_SIZE=10000               # Number of entities
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=1073741824  # Bytes (1GB)

# Memory pressure thresholds
ENTITYDB_MEMORY_HIGH_THRESHOLD=0.8             # 80% - start cleanup
ENTITYDB_MEMORY_CRITICAL_THRESHOLD=0.9         # 90% - emergency measures
```

### Retention Settings

```bash
# Metrics retention (affects memory usage)
ENTITYDB_METRICS_RETENTION_RAW=24h             # Raw metrics
ENTITYDB_METRICS_RETENTION_1MIN=7d             # 1-minute aggregates
ENTITYDB_METRICS_RETENTION_1HOUR=30d           # 1-hour aggregates
ENTITYDB_METRICS_RETENTION_1DAY=365d           # Daily aggregates
```

### Performance Tuning

```bash
# Batch writing (reduces memory pressure)
ENTITYDB_BATCH_SIZE=10                         # Entities per batch
ENTITYDB_BATCH_FLUSH_INTERVAL=100ms            # Flush interval

# Cache TTL for temporal queries
ENTITYDB_CACHE_TTL=5m                          # Query result cache
```

## Common Scenarios

### Scenario 1: High Metric Volume

If you generate >100 metrics/second:

```bash
# Reduce retention for raw metrics
ENTITYDB_METRICS_RETENTION_RAW=6h              # Keep only 6 hours

# Increase string cache for metric tags
ENTITYDB_STRING_CACHE_SIZE=200000              # More unique tags

# Consider disabling some metrics
ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING=false  # Reduce internal metrics
```

### Scenario 2: Large Entity Database

If you have >100k entities:

```bash
# Increase entity cache size
ENTITYDB_ENTITY_CACHE_SIZE=50000               # Cache more entities

# But watch memory limit
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=2147483648  # 2GB max

# Enable high-performance mode
ENTITYDB_HIGH_PERFORMANCE=true                  # Memory-mapped files
```

### Scenario 3: Memory Constrained Environment

If running on <2GB RAM:

```bash
# Minimal configuration
ENTITYDB_STRING_CACHE_SIZE=10000
ENTITYDB_STRING_CACHE_MEMORY_LIMIT=10485760    # 10MB
ENTITYDB_ENTITY_CACHE_SIZE=1000
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=104857600   # 100MB

# Aggressive retention
ENTITYDB_METRICS_RETENTION_RAW=1h              # Very short
ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING=false  # Disable non-essential
```

## Monitoring Your Migration

### Key Metrics to Watch

1. **Memory Stability**
   - Should plateau after initial growth
   - No continuous upward trend
   - Stays below configured limits

2. **Cache Performance**
   ```bash
   # String cache hit rate (should be >80%)
   curl http://localhost:8085/api/v1/system/metrics | \
     jq '.performance.stringCache.hitRate'
   
   # Entity cache hit rate (should be >70%)  
   curl http://localhost:8085/api/v1/system/metrics | \
     jq '.performance.entityCache.hitRate'
   ```

3. **Pressure Events**
   ```bash
   # Check for memory pressure (should be rare)
   grep "memory pressure" /var/log/entitydb/entitydb.log | tail -10
   ```

### Performance Impact

Expected changes after migration:

- **Memory Usage**: 50-80% reduction in steady state
- **Query Performance**: 
  - Cache hits: Same performance
  - Cache misses: ~10ms additional latency
- **Write Performance**: Minimal impact (<5%)
- **GC Pauses**: More frequent but shorter

## Troubleshooting

### Issue: High Cache Miss Rate

**Symptoms**: Slow queries, hit rate <60%

**Solution**:
```bash
# Increase cache size
export ENTITYDB_ENTITY_CACHE_SIZE=20000
systemctl restart entitydb
```

### Issue: Frequent Memory Pressure

**Symptoms**: "Memory pressure detected" logs

**Solution**:
```bash
# Reduce retention
export ENTITYDB_METRICS_RETENTION_RAW=12h
# Or increase memory limits
export ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=2147483648
```

### Issue: Metrics Collection Stopped

**Symptoms**: No new metrics, "metrics disabled" log

**Solution**:
This is emergency protection. Fix memory pressure first:
```bash
# Reduce cache sizes
export ENTITYDB_STRING_CACHE_SIZE=50000
export ENTITYDB_ENTITY_CACHE_SIZE=5000
systemctl restart entitydb
```

## Rollback Procedure

If you need to rollback to a previous version:

```bash
# Stop service
systemctl stop entitydb

# Restore old binary
cp /backup/entitydb-old /opt/entitydb/bin/entitydb

# Remove new configuration
# Comment out new environment variables
vim /etc/entitydb/entitydb.env

# Start service
systemctl start entitydb
```

**Note**: Data format is backward compatible, but you'll lose memory optimization benefits.

## Getting Help

### Documentation
- [Memory Architecture](../architecture/memory-optimization-architecture.md)
- [Operations Guide](../admin-guide/memory-management-operations.md)
- [ADR-029](../architecture/decisions/ADR-029-memory-optimization-strategy.md)

### Support Channels
- GitHub Issues: https://github.com/entitydb/entitydb/issues
- Community Forum: https://forum.entitydb.io
- Slack: entitydb.slack.com

### Diagnostic Information

When reporting issues, include:

```bash
# Version info
curl http://localhost:8085/health

# Memory metrics
curl http://localhost:8085/api/v1/system/metrics | jq '.memory'

# Cache stats
curl http://localhost:8085/api/v1/system/metrics | jq '.performance'

# Recent logs
tail -1000 /var/log/entitydb/entitydb.log | grep -E "memory|pressure|cache"
```

## Summary

The v2.34.0 memory optimizations make EntityDB production-ready for long-running deployments. Key points:

1. **Automatic Management**: No manual intervention required
2. **Configurable Limits**: Tune based on your resources
3. **Graceful Degradation**: System slows but doesn't crash
4. **Backward Compatible**: Easy upgrade path

With proper configuration, EntityDB now provides predictable memory usage and stable operation under all workload conditions.