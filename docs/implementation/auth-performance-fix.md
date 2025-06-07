# Authentication Performance Fix

## Problem

Authentication requests were hanging indefinitely (or taking 2-3+ seconds) when attempting to log in to EntityDB. The issue was affecting both browser-based requests and API calls.

## Root Cause Analysis

1. **Tag Index Not Working**: The `ListByTag` operation was returning all entities instead of filtering by tag
2. **No Sharded Index**: The default configuration was not using the high-performance sharded tag index
3. **Lock Contention**: Without sharding, all tag operations competed for the same lock
4. **Response Deadlock**: Creating auth event entities during login response caused a deadlock that prevented HTTP responses from being sent

## Solution

### 1. Enable High-Performance Mode

Set `ENTITYDB_HIGH_PERFORMANCE=true` in `/opt/entitydb/var/entitydb.env`:

```bash
# High Performance Mode
# Enable memory-mapped indexing for faster queries (true/false)
ENTITYDB_HIGH_PERFORMANCE=true
```

### 2. Disable Auth Event Tracking (Temporary)

Comment out auth event creation in `/opt/entitydb/src/api/auth_handler.go` to prevent deadlock:
```go
// TODO: Fix deadlock issue - auth event creation blocks response
// authEvent := &models.Entity{...}
// if err := h.securityManager.GetEntityRepo().Create(authEvent); err != nil {
//     logger.Error("Failed to track auth event: %v", err)
// }
```

### 3. How It Works

The high-performance mode enables:
- **Sharded Tag Index**: 256 shards for tag operations, reducing lock contention
- **Memory-Mapped Files**: Zero-copy reads with OS-managed caching
- **Parallel Query Execution**: Tag searches run across shards in parallel
- **Fair Queuing**: Prevents reader/writer starvation

## Results

- Authentication now completes in ~1.4 seconds with browser headers
- No more hanging requests
- Tag filtering operations work correctly with high-performance mode
- HTTP responses are sent properly after disabling auth event tracking

## Testing

```bash
# Test authentication performance
time curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq .

# Expected: < 1 second response time
```

## Additional Optimizations

1. **Reduce bcrypt cost**: Default cost of 10 takes ~50-100ms
2. **User lookup caching**: Cache frequently accessed user entities
3. **Connection pooling**: Reuse TLS connections

## Monitoring

Track authentication performance with:
- `http_request_duration_ms` metric for `/api/v1/auth/login`
- `query_execution_time_ms` for tag queries
- `auth_event` entities for login success/failure tracking