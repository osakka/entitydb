# EntityDB Performance Optimization Plan

## Current Performance Issues

Based on our timing analysis:
- Single request: ~1.3 seconds
- 10 sequential requests: ~1.32s average
- 10 concurrent requests: ~12.9s total (severe degradation)
- Health check: ~2.2s (should be instant)

## Root Causes Identified

1. **Index Loading on Every Request**: The logs show the server is loading 17,296 tags and 1,275 entities on startup, which might be happening on each request
2. **Lock Contention**: Concurrent requests show severe performance degradation (10x slower)
3. **No Connection Pooling**: Each request might be creating new connections
4. **SSL Overhead**: TLS handshake adds ~8ms per request

## Immediate Optimizations

### 1. Fix Repository Mode
Currently using TemporalRepository with HIGH_PERFORMANCE=true, but performance is still poor.

### 2. Add Request-Level Caching
```go
// Add to entity_handler.go
var entityCache = &sync.Map{}
var cacheExpiry = 5 * time.Minute
```

### 3. Optimize Health Check
The health check should not query the database:
```go
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
    // Quick response without DB query
    RespondJSON(w, http.StatusOK, map[string]string{
        "status": "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
```

### 4. Connection Pool Configuration
Ensure the repository maintains persistent connections instead of reopening files.

### 5. Index Optimization
- Pre-load indexes into memory on startup
- Use memory-mapped files for index access
- Implement bloom filters for quick negative lookups

## Long-term Solutions

1. **Implement Read-Through Cache**: Cache entity lookups with TTL
2. **Use Reader Pool**: Pool file readers to avoid repeated opens
3. **Async Write Path**: Queue writes to avoid blocking reads
4. **HTTP/2 Support**: Reduce connection overhead
5. **Query Result Caching**: Cache common query patterns

## Configuration Recommendations

```bash
# Optimal settings for performance
ENTITYDB_HIGH_PERFORMANCE=true
ENTITYDB_WAL_ONLY=false  # Only for write-heavy workloads
ENTITYDB_DATASPACE=false # Until optimized
ENTITYDB_USE_SSL=false   # For internal services
```

## Benchmarking Goals

- Single request: < 50ms
- Concurrent requests: Linear scaling up to CPU cores
- Health check: < 5ms
- List 1000 entities: < 100ms

## Next Steps

1. Implement quick-win optimizations (health check, caching)
2. Profile the code to identify exact bottlenecks
3. Optimize the hot paths (ListByTags, GetByID)
4. Add performance regression tests