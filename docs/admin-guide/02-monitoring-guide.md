# Monitoring and Metrics Guide

> **Version**: v2.31.0 | **Last Updated**: 2025-06-14 | **Status**: AUTHORITATIVE

This guide covers comprehensive monitoring and metrics collection for EntityDB, including system health monitoring, performance metrics, alerting setup, and observability best practices.

## Overview

EntityDB v2.31.0 provides extensive monitoring capabilities through multiple endpoints and metric collection systems designed for production observability and performance optimization.

## Monitoring Endpoints

### 1. Health Check Endpoint

The `/health` endpoint provides immediate system status:

```bash
# Basic health check
curl http://localhost:8085/health
```

**Response Format:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-14T10:00:00Z",
  "version": "2.31.0",
  "uptime": "24h15m30s",
  "checks": {
    "database": "healthy",
    "memory": "healthy",
    "storage": "healthy"
  }
}
```

**Health Status Codes:**
- `200`: System healthy
- `503`: System unhealthy or degraded
- `500`: Critical system failure

### 2. Prometheus Metrics Endpoint

The `/metrics` endpoint provides Prometheus-compatible metrics:

```bash
# Get Prometheus metrics
curl http://localhost:8085/metrics
```

**Key Metrics Available:**
- `entitydb_entities_total` - Total entity count
- `entitydb_http_requests_total` - HTTP request count
- `entitydb_http_request_duration_seconds` - Request latency
- `entitydb_memory_usage_bytes` - Memory consumption
- `entitydb_storage_size_bytes` - Storage utilization

### 3. System Metrics Endpoint

The `/api/v1/system/metrics` endpoint provides detailed EntityDB metrics:

```bash
# Get detailed system metrics (requires authentication)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics
```

### 4. RBAC Metrics Endpoint

The `/api/v1/rbac/metrics` endpoint provides authentication and authorization metrics:

```bash
# Get RBAC metrics (requires admin permissions)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/rbac/metrics
```

## Metrics Categories

### 1. System Metrics

#### Memory and Performance
```bash
# Memory usage monitoring
curl -s http://localhost:8085/api/v1/system/metrics | jq '.data.system'
```

**Key System Metrics:**
- `memory_usage_mb` - Current memory usage in MB
- `memory_allocated_mb` - Total allocated memory
- `gc_cycles` - Garbage collection cycles
- `goroutines` - Active goroutines count
- `cpu_usage_percent` - CPU utilization

#### Storage Metrics
```bash
# Storage utilization
curl -s http://localhost:8085/api/v1/system/metrics | jq '.data.storage'
```

**Key Storage Metrics:**
- `database_size_mb` - Total database size
- `wal_size_mb` - Write-Ahead Log size
- `entities_count` - Total entities stored
- `chunks_count` - Content chunks count
- `index_size_mb` - Index file sizes

### 2. Performance Metrics (v2.31.0 Enhancements)

#### Tag Operations
```bash
# Tag performance metrics
curl -s http://localhost:8085/api/v1/system/metrics | jq '.data.performance.tags'
```

**Key Performance Metrics:**
- `tag_cache_hits` - O(1) tag cache hit rate
- `tag_cache_misses` - Cache miss count
- `index_build_time_ms` - Parallel index building time
- `temporal_cache_hits` - Temporal tag variant cache hits

#### JSON Processing
```bash
# JSON encoder metrics
curl -s http://localhost:8085/api/v1/system/metrics | jq '.data.performance.json'
```

**JSON Performance Metrics:**
- `encoder_pool_size` - JSON encoder pool utilization
- `encoding_time_ms` - Average encoding time
- `pool_hits` - Encoder pool hit rate

#### Batch Operations
```bash
# Batch operation metrics
curl -s http://localhost:8085/api/v1/system/metrics | jq '.data.performance.batch'
```

**Batch Performance Metrics:**
- `batch_write_count` - Batch write operations
- `batch_size_avg` - Average batch size
- `batch_timeout_count` - Batch timeout occurrences

### 3. HTTP Metrics

#### Request Analytics
```bash
# HTTP request metrics
curl -s http://localhost:8085/api/v1/system/metrics | jq '.data.http'
```

**HTTP Metrics:**
- `requests_total` - Total HTTP requests
- `requests_per_second` - Current request rate
- `response_time_avg_ms` - Average response time
- `status_codes` - Response status distribution
- `active_connections` - Current active connections

### 4. Authentication Metrics

#### RBAC Analytics
```bash
# Authentication and authorization metrics
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/rbac/metrics
```

**RBAC Metrics:**
- `active_sessions` - Current active sessions
- `login_attempts` - Authentication attempts
- `permission_checks` - Authorization checks
- `failed_auths` - Failed authentication count

## Monitoring Setup

### 1. Prometheus Configuration

#### Prometheus Config
```yaml
# prometheus.yml
global:
  scrape_interval: 30s
  evaluation_interval: 30s

scrape_configs:
  - job_name: 'entitydb'
    static_configs:
      - targets: ['localhost:8085']
    metrics_path: '/metrics'
    scrape_interval: 30s
    scrape_timeout: 10s
    scheme: http
```

#### Docker Prometheus Setup
```bash
# Run Prometheus with EntityDB monitoring
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v /path/to/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus:latest
```

### 2. Grafana Dashboard

#### EntityDB Dashboard JSON
```json
{
  "dashboard": {
    "title": "EntityDB Monitoring",
    "panels": [
      {
        "title": "Entity Count",
        "type": "stat",
        "targets": [
          {
            "expr": "entitydb_entities_total",
            "legendFormat": "Total Entities"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(entitydb_http_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "entitydb_memory_usage_bytes / 1024 / 1024",
            "legendFormat": "Memory (MB)"
          }
        ]
      }
    ]
  }
}
```

### 3. Custom Monitoring Scripts

#### System Health Monitor
```bash
#!/bin/bash
# EntityDB Health Monitor Script

HEALTH_URL="http://localhost:8085/health"
METRICS_URL="http://localhost:8085/api/v1/system/metrics"
LOG_FILE="/opt/entitydb/var/log/monitor.log"
ALERT_THRESHOLD_MEMORY=4096  # MB
ALERT_THRESHOLD_STORAGE=80   # Percent

# Check health status
health_check() {
    local response=$(curl -s -w "%{http_code}" -o /tmp/health.json $HEALTH_URL)
    local http_code="${response: -3}"
    
    if [ "$http_code" -ne 200 ]; then
        echo "$(date): ALERT - Health check failed (HTTP $http_code)" >> $LOG_FILE
        return 1
    fi
    
    local status=$(jq -r '.status' /tmp/health.json)
    if [ "$status" != "healthy" ]; then
        echo "$(date): ALERT - System status: $status" >> $LOG_FILE
        return 1
    fi
    
    return 0
}

# Check memory usage
memory_check() {
    local token="$1"
    local memory_usage=$(curl -s -H "Authorization: Bearer $token" $METRICS_URL | \
                        jq -r '.data.system.memory_usage_mb // 0')
    
    if [ "$memory_usage" -gt "$ALERT_THRESHOLD_MEMORY" ]; then
        echo "$(date): ALERT - High memory usage: ${memory_usage}MB" >> $LOG_FILE
        return 1
    fi
    
    return 0
}

# Check storage usage
storage_check() {
    local usage_percent=$(df /opt/entitydb/var | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$usage_percent" -gt "$ALERT_THRESHOLD_STORAGE" ]; then
        echo "$(date): ALERT - High storage usage: ${usage_percent}%" >> $LOG_FILE
        return 1
    fi
    
    return 0
}

# Main monitoring function
main() {
    echo "$(date): Starting health check" >> $LOG_FILE
    
    # Basic health check
    if ! health_check; then
        # Send alert notification here
        exit 1
    fi
    
    # Storage check
    if ! storage_check; then
        # Send alert notification here
        exit 1
    fi
    
    echo "$(date): All checks passed" >> $LOG_FILE
}

main "$@"
```

#### Performance Monitor
```bash
#!/bin/bash
# EntityDB Performance Monitor

METRICS_URL="http://localhost:8085/api/v1/system/metrics"
PERFORMANCE_LOG="/opt/entitydb/var/log/performance.log"

# Collect performance metrics
collect_metrics() {
    local token="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Get metrics
    local metrics=$(curl -s -H "Authorization: Bearer $token" $METRICS_URL)
    
    # Extract key performance indicators
    local entity_count=$(echo "$metrics" | jq -r '.data.entities.total_count // 0')
    local memory_usage=$(echo "$metrics" | jq -r '.data.system.memory_usage_mb // 0')
    local request_rate=$(echo "$metrics" | jq -r '.data.http.requests_per_second // 0')
    local avg_response_time=$(echo "$metrics" | jq -r '.data.http.response_time_avg_ms // 0')
    local tag_cache_hits=$(echo "$metrics" | jq -r '.data.performance.tags.cache_hits // 0')
    local goroutines=$(echo "$metrics" | jq -r '.data.system.goroutines // 0')
    
    # Log performance data
    echo "$timestamp,$entity_count,$memory_usage,$request_rate,$avg_response_time,$tag_cache_hits,$goroutines" >> $PERFORMANCE_LOG
}

# Initialize performance log
if [ ! -f "$PERFORMANCE_LOG" ]; then
    echo "timestamp,entities,memory_mb,req_per_sec,avg_response_ms,cache_hits,goroutines" > $PERFORMANCE_LOG
fi

collect_metrics "$@"
```

## Alerting Configuration

### 1. Prometheus Alerting Rules

```yaml
# entitydb-alerts.yml
groups:
  - name: entitydb
    rules:
      - alert: EntityDBDown
        expr: up{job="entitydb"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "EntityDB is down"
          description: "EntityDB has been down for more than 1 minute"

      - alert: HighMemoryUsage
        expr: entitydb_memory_usage_bytes / 1024 / 1024 > 4096
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is above 4GB"

      - alert: HighResponseTime
        expr: entitydb_http_request_duration_seconds{quantile="0.95"} > 1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High response time"
          description: "95th percentile response time is above 1 second"

      - alert: LowCacheHitRate
        expr: rate(entitydb_tag_cache_hits[5m]) / rate(entitydb_tag_operations_total[5m]) < 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate"
          description: "Tag cache hit rate is below 80%"
```

### 2. Notification Channels

#### Slack Integration
```bash
# Slack webhook notification script
send_slack_alert() {
    local message="$1"
    local webhook_url="YOUR_SLACK_WEBHOOK_URL"
    
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"EntityDB Alert: $message\"}" \
        $webhook_url
}
```

#### Email Alerts
```bash
# Email notification script
send_email_alert() {
    local subject="$1"
    local message="$2"
    local recipient="admin@yourcompany.com"
    
    echo "$message" | mail -s "EntityDB Alert: $subject" $recipient
}
```

## Dashboard Configuration

### 1. EntityDB Web Dashboard

The built-in dashboard provides real-time metrics at `https://your-domain:8443/`:

**Dashboard Features:**
- Real-time entity counts and growth
- Memory usage charts with health scoring
- HTTP request metrics and response times
- Tag performance and cache hit rates
- System health indicators
- Authentication and session analytics

### 2. Custom Dashboard Setup

#### Health Score Calculation
```javascript
// Health scoring algorithm (0-100%)
function calculateHealthScore(metrics) {
    let score = 100;
    
    // Memory usage impact (0-30 point penalty)
    const memoryUsagePercent = metrics.memory_usage_mb / metrics.memory_limit_mb;
    if (memoryUsagePercent > 0.9) score -= 30;
    else if (memoryUsagePercent > 0.8) score -= 20;
    else if (memoryUsagePercent > 0.7) score -= 10;
    
    // Response time impact (0-25 point penalty)
    if (metrics.avg_response_time_ms > 1000) score -= 25;
    else if (metrics.avg_response_time_ms > 500) score -= 15;
    else if (metrics.avg_response_time_ms > 200) score -= 5;
    
    // Cache hit rate impact (0-20 point penalty)
    if (metrics.cache_hit_rate < 0.7) score -= 20;
    else if (metrics.cache_hit_rate < 0.8) score -= 10;
    else if (metrics.cache_hit_rate < 0.9) score -= 5;
    
    // Error rate impact (0-25 point penalty)
    if (metrics.error_rate > 0.05) score -= 25;
    else if (metrics.error_rate > 0.02) score -= 15;
    else if (metrics.error_rate > 0.01) score -= 5;
    
    return Math.max(0, score);
}
```

## Performance Monitoring (v2.31.0)

### 1. O(1) Tag Value Caching

Monitor tag cache performance:
```bash
# Check tag cache efficiency
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics | \
  jq '.data.performance.tags | {
    cache_hits: .cache_hits,
    cache_misses: .cache_misses,
    hit_rate: (.cache_hits / (.cache_hits + .cache_misses))
  }'
```

### 2. Parallel Index Building

Monitor index build performance:
```bash
# Check index building metrics
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics | \
  jq '.data.performance.indexing'
```

### 3. JSON Encoder Pool Monitoring

Track JSON processing efficiency:
```bash
# Monitor JSON encoder pool
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics | \
  jq '.data.performance.json'
```

## Troubleshooting Monitoring Issues

### 1. Metrics Collection Problems

#### Missing Metrics
```bash
# Check if metrics endpoint is accessible
curl -v http://localhost:8085/metrics

# Verify authentication for detailed metrics
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics
```

#### Performance Degradation
```bash
# Check for slow queries
grep "slow.*query" /opt/entitydb/var/log/entitydb.log

# Monitor resource usage
top -p $(pgrep entitydb)
iotop -p $(pgrep entitydb)
```

### 2. Dashboard Issues

#### Dashboard Not Loading
```bash
# Check HTTPS configuration
curl -k https://localhost:8443/

# Verify static file serving
ls -la /opt/entitydb/share/htdocs/
```

#### Real-time Updates Not Working
```bash
# Check WebSocket connectivity
curl -v -H "Upgrade: websocket" \
  -H "Connection: Upgrade" \
  https://localhost:8443/ws
```

## Best Practices

### 1. Monitoring Strategy

#### Baseline Establishment
- Monitor for 1-2 weeks to establish normal operating ranges
- Document typical memory usage, response times, and cache hit rates
- Identify daily/weekly usage patterns

#### Alert Thresholds
- Set memory alerts at 80% of available RAM
- Configure response time alerts at 2x baseline average
- Set storage alerts at 80% capacity
- Monitor cache hit rates below 70%

### 2. Performance Optimization

#### Cache Optimization
- Monitor tag cache hit rates and increase cache size if needed
- Optimize query patterns to improve temporal cache usage
- Balance memory allocation between different cache types

#### Resource Management
- Monitor memory usage and configure appropriate limits
- Track goroutine counts to identify potential leaks
- Monitor file descriptor usage in high-concurrency scenarios

### 3. Maintenance Windows

#### Regular Health Checks
- Weekly performance metric reviews
- Monthly capacity planning assessments
- Quarterly monitoring system updates

#### Metric Retention
- Keep detailed metrics for 30 days
- Aggregate hourly metrics for 6 months
- Store daily summaries for 2 years

---

*This monitoring guide provides comprehensive observability for EntityDB v2.31.0. For additional operational procedures, see [Production Deployment](../70-deployment/01-production-deployment.md).*