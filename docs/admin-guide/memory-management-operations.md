# Memory Management Operations Guide

**Version**: 2.34.0  
**Audience**: System Administrators, SREs  
**Last Updated**: 2025-06-22

## Quick Reference

### Normal Operation Indicators
- Memory usage: <70% of limits
- Cache hit rate: >80%
- Pressure events: <1/hour
- GC pause: <100ms

### Warning Signs
- Memory usage: >80% sustained
- Cache hit rate: <60%
- Pressure events: >10/hour
- GC pause: >500ms

### Critical Issues
- Memory usage: >90%
- Metrics disabled automatically
- OOM errors in logs
- Service restarts

## Configuration Quick Start

### Conservative (Low Memory)
```bash
# 1GB total memory budget
export ENTITYDB_STRING_CACHE_SIZE=10000
export ENTITYDB_STRING_CACHE_MEMORY_LIMIT=52428800      # 50MB
export ENTITYDB_ENTITY_CACHE_SIZE=1000
export ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=104857600     # 100MB
export ENTITYDB_METRICS_RETENTION_RAW=6h
```

### Balanced (Default)
```bash
# 4GB total memory budget
export ENTITYDB_STRING_CACHE_SIZE=100000
export ENTITYDB_STRING_CACHE_MEMORY_LIMIT=104857600     # 100MB
export ENTITYDB_ENTITY_CACHE_SIZE=10000
export ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=1073741824    # 1GB
export ENTITYDB_METRICS_RETENTION_RAW=24h
```

### Performance (High Memory)
```bash
# 16GB total memory budget
export ENTITYDB_STRING_CACHE_SIZE=1000000
export ENTITYDB_STRING_CACHE_MEMORY_LIMIT=524288000     # 500MB
export ENTITYDB_ENTITY_CACHE_SIZE=100000
export ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=5368709120    # 5GB
export ENTITYDB_METRICS_RETENTION_RAW=7d
```

## Monitoring Memory Health

### Key Metrics

1. **System Memory**
```bash
# Check EntityDB memory usage
curl -s http://localhost:8085/health | jq '.memory'

# Output shows:
# - heapInUse: Active heap memory
# - heapAlloc: Allocated heap memory
# - totalAlloc: Cumulative allocated
# - sys: Total system memory reserved
```

2. **Cache Performance**
```bash
# String cache stats
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '.performance.stringCache'

# Entity cache stats  
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '.performance.entityCache'
```

3. **Memory Pressure**
```bash
# Check pressure events
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '.memory.pressureEvents'
```

### Dashboard Monitoring

Access the web dashboard at `https://localhost:8085/` and navigate to:
- **Performance** tab: Real-time memory charts
- **System** tab: Cache statistics
- **Metrics** tab: Pressure event history

### Log Analysis

Watch for memory-related log entries:

```bash
# Normal operation
tail -f var/entitydb.log | grep -E "memory|cache|pressure"

# Warning signs
grep "Memory pressure" var/entitydb.log
grep "High memory pressure detected" var/entitydb.log
grep "CRITICAL memory pressure" var/entitydb.log

# Cache evictions
grep "evictions" var/entitydb.log | tail -20
```

## Troubleshooting Common Issues

### High Memory Usage

**Symptoms**:
- Memory >80% consistently
- Slow response times
- Frequent GC pauses

**Diagnosis**:
```bash
# Check what's using memory
curl -s http://localhost:8085/api/v1/system/metrics | jq '.memory'

# Check cache sizes
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '{string: .performance.stringCache.size, 
       entity: .performance.entityCache.size}'

# Check temporal tag counts
curl -s http://localhost:8085/api/v1/entities/query?tag=type:metric | \
  jq '.[0].tags | length'
```

**Solutions**:
1. Reduce cache sizes:
   ```bash
   export ENTITYDB_STRING_CACHE_SIZE=50000
   export ENTITYDB_ENTITY_CACHE_SIZE=5000
   systemctl restart entitydb
   ```

2. Decrease retention:
   ```bash
   export ENTITYDB_METRICS_RETENTION_RAW=12h
   export ENTITYDB_METRICS_RETENTION_1MIN=3d
   ```

3. Increase memory limit:
   ```bash
   # If system has available memory
   export ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=2147483648  # 2GB
   ```

### Poor Cache Performance

**Symptoms**:
- Hit rate <60%
- Slow queries
- High disk I/O

**Diagnosis**:
```bash
# Check hit rates
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '.performance | {
    stringHitRate: .stringCache.hitRate,
    entityHitRate: .entityCache.hitRate
  }'

# Check eviction rates
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '.performance | {
    stringEvictions: .stringCache.evictions,
    entityEvictions: .entityCache.evictions
  }'
```

**Solutions**:
1. Increase cache size if memory allows:
   ```bash
   export ENTITYDB_ENTITY_CACHE_SIZE=20000
   ```

2. Adjust workload:
   - Batch similar queries together
   - Use dataset filtering to reduce scope
   - Implement application-level caching

3. Pre-warm cache after restart:
   ```bash
   # Script to pre-load common entities
   for dataset in $(curl -s http://localhost:8085/api/v1/datasets | jq -r '.[].id'); do
     curl -s "http://localhost:8085/api/v1/entities/query?dataset=$dataset&limit=100" > /dev/null
   done
   ```

### Memory Pressure Events

**Symptoms**:
- "Memory pressure detected" in logs
- Automatic feature disabling
- Performance degradation

**Diagnosis**:
```bash
# Check current pressure
curl -s http://localhost:8085/api/v1/system/metrics | \
  jq '.memory.currentPressure'

# Review pressure history
grep "pressure" var/entitydb.log | tail -50 | \
  awk '{print $1, $2, $8}' | sort | uniq -c
```

**Solutions**:
1. Immediate relief:
   ```bash
   # Force garbage collection via API
   curl -X POST http://localhost:8085/api/v1/admin/gc
   
   # Clear caches if critical
   curl -X POST http://localhost:8085/api/v1/admin/clear-caches
   ```

2. Adjust thresholds:
   ```bash
   # Give more headroom
   export ENTITYDB_MEMORY_HIGH_THRESHOLD=0.85  # 85%
   export ENTITYDB_MEMORY_CRITICAL_THRESHOLD=0.95  # 95%
   ```

## Performance Tuning

### Workload Profiles

#### High Write Volume
```bash
# Optimize for writes
export ENTITYDB_BATCH_SIZE=50                # Larger batches
export ENTITYDB_BATCH_FLUSH_INTERVAL=200ms   # Less frequent flushes
export ENTITYDB_STRING_CACHE_SIZE=50000      # Smaller string cache
export ENTITYDB_ENTITY_CACHE_SIZE=5000       # Smaller entity cache
```

#### High Read Volume
```bash
# Optimize for reads
export ENTITYDB_ENTITY_CACHE_SIZE=50000      # Large entity cache
export ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=5368709120  # 5GB
export ENTITYDB_CACHE_TTL=10m                # Longer TTL
```

#### Mixed Workload
```bash
# Balanced configuration
export ENTITYDB_BATCH_SIZE=20
export ENTITYDB_STRING_CACHE_SIZE=100000
export ENTITYDB_ENTITY_CACHE_SIZE=20000
export ENTITYDB_CACHE_TTL=5m
```

### Memory Allocation Patterns

#### Reduce GC Pressure
```bash
# Pre-allocate memory
export GOGC=100                               # Default GC target
export GOMEMLIMIT=8GiB                       # Hard memory limit (Go 1.19+)

# Start with larger heap
export GODEBUG=gctrace=1                    # Enable GC tracing
```

#### NUMA Optimization
```bash
# For multi-socket systems
numactl --interleave=all entitydb
```

## Capacity Planning

### Memory Requirements Formula

```
Total Memory = Base + StringCache + EntityCache + Temporal + Buffer

Where:
- Base: 100MB (binary + OS overhead)
- StringCache: (avg_string_size * cache_size) + 100 bytes overhead per entry
- EntityCache: (avg_entity_size * cache_size) + 200 bytes overhead per entry  
- Temporal: (metrics_per_sec * 86400 * retention_days * 100 bytes)
- Buffer: 20% headroom for GC and spikes
```

### Example Calculations

#### Small Deployment
- 100 entities, 10 metrics/sec, 1 day retention
```
Base:         100MB
StringCache:  (50 * 10,000) + (100 * 10,000) = 1.5MB
EntityCache:  (1KB * 1,000) + (200 * 1,000) = 1.2MB
Temporal:     (10 * 86,400 * 1 * 100) = 86MB
Buffer:       20% = 38MB
Total:        ~227MB (recommend 512MB)
```

#### Medium Deployment
- 10k entities, 100 metrics/sec, 7 days retention
```
Base:         100MB
StringCache:  (50 * 100,000) + (100 * 100,000) = 15MB
EntityCache:  (2KB * 10,000) + (200 * 10,000) = 22MB
Temporal:     (100 * 86,400 * 7 * 100) = 6GB
Buffer:       20% = 1.2GB
Total:        ~7.3GB (recommend 8GB)
```

#### Large Deployment
- 1M entities, 1000 metrics/sec, 30 days retention
```
Base:         100MB
StringCache:  (50 * 1,000,000) + (100 * 1,000,000) = 150MB
EntityCache:  (5KB * 100,000) + (200 * 100,000) = 520MB
Temporal:     (1000 * 86,400 * 30 * 100) = 260GB
Buffer:       20% = 52GB
Total:        ~313GB (recommend 512GB with retention adjustment)
```

## Emergency Procedures

### Memory Crisis Response

1. **Immediate Actions** (Memory >95%):
   ```bash
   # Disable metrics collection
   curl -X POST http://localhost:8085/api/v1/admin/disable-metrics
   
   # Force aggressive GC
   curl -X POST http://localhost:8085/api/v1/admin/gc?aggressive=true
   
   # Clear non-essential caches
   curl -X POST http://localhost:8085/api/v1/admin/clear-caches?level=aggressive
   ```

2. **Stabilization** (5 minutes):
   ```bash
   # Reduce limits drastically
   export ENTITYDB_STRING_CACHE_SIZE=1000
   export ENTITYDB_ENTITY_CACHE_SIZE=100
   systemctl restart entitydb
   ```

3. **Recovery** (when stable):
   ```bash
   # Gradually increase limits
   # Monitor memory usage closely
   # Re-enable features one by one
   ```

### Preventive Measures

1. **Monitoring Alerts**:
   ```yaml
   # Prometheus alert rules
   - alert: EntityDBHighMemory
     expr: entitydb_memory_heap_inuse_bytes / entitydb_memory_sys_bytes > 0.8
     for: 5m
     
   - alert: EntityDBMemoryPressure
     expr: rate(entitydb_memory_pressure_events[5m]) > 1
     for: 10m
   ```

2. **Automated Response**:
   ```bash
   #!/bin/bash
   # Auto-scale script
   PRESSURE=$(curl -s http://localhost:8085/api/v1/system/metrics | jq -r '.memory.currentPressure')
   if (( $(echo "$PRESSURE > 0.8" | bc -l) )); then
     # Reduce cache sizes
     systemctl set-environment ENTITYDB_ENTITY_CACHE_SIZE=5000
     systemctl restart entitydb
   fi
   ```

## Best Practices

1. **Regular Maintenance**:
   - Review memory metrics weekly
   - Adjust retention policies monthly
   - Plan capacity quarterly

2. **Testing Changes**:
   - Test configuration changes in staging
   - Monitor for 24 hours after changes
   - Keep previous configuration documented

3. **Documentation**:
   - Document your specific tuning parameters
   - Record rationale for settings
   - Track changes over time

4. **Gradual Adjustments**:
   - Change one parameter at a time
   - Increase by 20% maximum per change
   - Monitor impact before next change

## Summary

EntityDB's memory management system provides robust protection against memory exhaustion while maintaining performance. Success requires:

1. Understanding your workload characteristics
2. Monitoring key metrics continuously  
3. Tuning configuration appropriately
4. Responding quickly to pressure events

With proper configuration and monitoring, EntityDB can handle demanding workloads within predictable memory boundaries.