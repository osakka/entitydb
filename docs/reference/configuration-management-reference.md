# EntityDB Configuration Management Reference

## Overview

EntityDB implements a sophisticated three-tier configuration hierarchy that provides maximum flexibility for production deployments while maintaining development simplicity. This system eliminates all hardcoded values and provides comprehensive runtime configurability.

## Configuration Hierarchy

The configuration system follows a strict priority order:

```
Database Configuration (Highest Priority)
         ↓
Command-Line Flags (Medium Priority)  
         ↓
Environment Variables (Lowest Priority)
```

### 1. Database Configuration (Highest Priority)

Configuration stored as entities with `type:config` tags enables runtime configuration changes without server restarts.

**Storage Format:**
```
Entity ID: config_{namespace}_{key}
Tags: ["type:config", "conf:{namespace}:{key}"]
Content: {"key": "value"}
```

**Example:**
```bash
# Set server port via API
curl -H "Authorization: Bearer $TOKEN" \
  -X POST http://localhost:8085/api/v1/config \
  -d '{"namespace":"server","key":"port","value":"9090"}'
```

**Caching:**
- 5-minute TTL for performance
- Automatic refresh on configuration changes
- Thread-safe access with read-write mutexes

### 2. Command-Line Flags (Medium Priority)

All flags use long-form `--entitydb-*` naming convention. Short flags (`-h`, `-v`) reserved for essential functionality.

**Core Server Flags:**
```bash
--entitydb-port int                    # HTTP server port (default 8085)
--entitydb-ssl-port int               # HTTPS server port (default 8085) 
--entitydb-use-ssl                    # Enable SSL/TLS (default false)
--entitydb-ssl-cert string            # SSL certificate file path
--entitydb-ssl-key string             # SSL private key file path
```

**Database and Storage:**
```bash
--entitydb-data-path string           # Data directory path (default "./var")
--entitydb-database-file string       # Database file path (unified .edb format)
--entitydb-metrics-file string        # Metrics storage file path
--entitydb-backup-path string         # Backup directory path
--entitydb-temp-path string           # Temporary files directory
```

**Authentication and Security:**
```bash
--entitydb-token-secret string        # Secret key for JWT tokens
--entitydb-session-ttl-hours int      # Session timeout in hours (default 2)
--entitydb-bcrypt-cost int            # Bcrypt cost for password hashing (4-31)
--entitydb-default-admin-username string  # Default admin username
--entitydb-default-admin-password string  # Default admin password
--entitydb-default-admin-email string     # Default admin email
```

**Performance and Monitoring:**
```bash
--entitydb-high-performance           # Enable high-performance mode
--entitydb-metrics-interval duration  # Metrics collection interval (default 30s)
--entitydb-metrics-enable-request-tracking    # Enable HTTP request metrics
--entitydb-metrics-enable-storage-tracking    # Enable storage operation metrics
```

**Request Throttling:**
```bash
--entitydb-throttle-enabled           # Enable intelligent request throttling
--entitydb-throttle-requests-per-minute int   # Baseline requests per minute
--entitydb-throttle-polling-threshold int     # Repeated requests threshold
--entitydb-throttle-max-delay duration        # Maximum delay for throttled requests
--entitydb-throttle-cache-duration duration   # Cache duration for repeated requests
```

**Logging and Debugging:**
```bash
--entitydb-log-level string           # Log level (trace, debug, info, warn, error)
--entitydb-log-file string            # Server log file path
--entitydb-trace-subsystems string    # Comma-separated trace subsystems
--entitydb-dev-mode                   # Enable development mode
--entitydb-debug-port int             # Debug/profiling port (default 6060)
--entitydb-profile-enabled            # Enable CPU and memory profiling
```

**HTTP Configuration:**
```bash
--entitydb-http-read-timeout duration     # HTTP read timeout (default 15s)
--entitydb-http-write-timeout duration    # HTTP write timeout (default 15s)
--entitydb-http-idle-timeout duration     # HTTP idle timeout (default 1m)
--entitydb-shutdown-timeout duration      # Server shutdown timeout (default 30s)
```

**Rate Limiting:**
```bash
--entitydb-enable-rate-limit          # Enable rate limiting
--entitydb-rate-limit-requests int    # Requests allowed per window (default 100)
--entitydb-rate-limit-window-minutes int  # Rate limit window in minutes (default 1)
```

### 3. Environment Variables (Lowest Priority)

All flags have corresponding environment variables. Convert flag names by:
1. Remove `--entitydb-` prefix
2. Convert to uppercase
3. Replace hyphens with underscores
4. Add `ENTITYDB_` prefix

**Examples:**
```bash
# Flag: --entitydb-port
# Env:  ENTITYDB_PORT

# Flag: --entitydb-use-ssl  
# Env:  ENTITYDB_USE_SSL

# Flag: --entitydb-metrics-enable-request-tracking
# Env:  ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING
```

**Common Environment Variables:**
```bash
export ENTITYDB_PORT=8085
export ENTITYDB_USE_SSL=true
export ENTITYDB_SSL_CERT="/path/to/cert.pem"
export ENTITYDB_SSL_KEY="/path/to/key.pem"
export ENTITYDB_DATABASE_FILE="/data/entities.edb"
export ENTITYDB_LOG_LEVEL="debug"
export ENTITYDB_TOKEN_SECRET="your-secret-key"
export ENTITYDB_DEFAULT_ADMIN_USERNAME="admin"
export ENTITYDB_DEFAULT_ADMIN_PASSWORD="secure-password"
```

## Usage Patterns

### Development Environment

Use environment variables for development convenience:

```bash
# .env file
ENTITYDB_PORT=8085
ENTITYDB_USE_SSL=false
ENTITYDB_LOG_LEVEL=debug
ENTITYDB_DEV_MODE=true
ENTITYDB_DATABASE_FILE=./dev/entities.edb
```

### Production Deployment

Combine environment variables and command-line flags:

```bash
# Environment variables for secure values
export ENTITYDB_TOKEN_SECRET="$(cat /secrets/token-secret)"
export ENTITYDB_DEFAULT_ADMIN_PASSWORD="$(cat /secrets/admin-password)"

# Command-line flags for deployment-specific values
./entitydb \
  --entitydb-port 8085 \
  --entitydb-use-ssl true \
  --entitydb-ssl-cert /certs/server.pem \
  --entitydb-ssl-key /certs/server.key \
  --entitydb-database-file /data/production.edb \
  --entitydb-log-level info
```

### Runtime Configuration Updates

Use database configuration for runtime changes:

```bash
# Change log level without restart
curl -H "Authorization: Bearer $TOKEN" \
  -X POST http://localhost:8085/api/v1/config \
  -d '{"namespace":"logging","key":"level","value":"debug"}'

# Enable high-performance mode
curl -H "Authorization: Bearer $TOKEN" \
  -X POST http://localhost:8085/api/v1/config \
  -d '{"namespace":"performance","key":"high_performance","value":"true"}'
```

## File and Path Configuration

EntityDB follows a unified file architecture with single source of truth:

### Unified Database Format
- **Single File:** All data stored in unified `.edb` format
- **Embedded Components:** WAL, indexes, and data sections embedded in single file
- **No Legacy Files:** No separate `.db`, `.wal`, or `.idx` files
- **Configuration:** Use `--entitydb-database-file` for database location

### Directory Structure
```
ENTITYDB_DATA_PATH/
├── entities.edb          # Unified database file (configurable)
├── metrics.json          # Metrics storage (configurable)
├── entitydb.log          # Server logs (configurable)
├── entitydb.pid          # Process ID file (configurable)
└── backup/               # Backup directory (configurable)
    └── entities_backup.edb
```

### Path Configuration Examples
```bash
# Custom data directory
--entitydb-data-path /var/lib/entitydb

# Custom database file location
--entitydb-database-file /data/production.edb

# Custom log file location  
--entitydb-log-file /var/log/entitydb/server.log

# Custom backup directory
--entitydb-backup-path /backups/entitydb
```

## Configuration Manager API

### Programmatic Access

```go
// Create configuration manager
configManager := config.NewConfigManager(entityRepo)

// Register all flags
configManager.RegisterFlags()

// Parse command line
flag.Parse()

// Initialize with hierarchy
cfg, err := configManager.Initialize()

// Access configuration
port := cfg.Port
useSSL := cfg.UseSSL
dbFile := cfg.DatabaseFilename
```

### Runtime Configuration Management

```go
// Set database configuration
err := configManager.SetDatabaseConfig("server", "port", "9090")

// Get database configuration  
value, err := configManager.GetDatabaseConfig("server", "port")

// Refresh configuration cache
err := configManager.RefreshConfig()

// Get current effective configuration
cfg := configManager.GetConfig()
```

## Security Considerations

### Credential Management
- Never hardcode secrets in configuration files
- Use environment variables for sensitive values
- Consider external secret management systems
- Rotate credentials regularly

### Production Security
```bash
# Secure defaults for production
export ENTITYDB_TOKEN_SECRET="$(openssl rand -base64 32)"
export ENTITYDB_DEFAULT_ADMIN_PASSWORD="$(openssl rand -base64 24)"
export ENTITYDB_BCRYPT_COST=12  # Higher cost for production
```

### File Permissions
```bash
# Secure file permissions
chmod 600 /secrets/token-secret
chmod 600 /certs/server.key
chmod 644 /certs/server.pem
chmod 755 /data/entitydb/
chmod 644 /data/entitydb/entities.edb
```

## Troubleshooting

### Configuration Debugging

1. **Check Effective Configuration:**
   ```bash
   # View help to see current defaults (affected by env vars)
   ./entitydb --help | grep -A1 "entitydb-port"
   ```

2. **Test Environment Variables:**
   ```bash
   # Test specific environment variable
   ENTITYDB_PORT=9999 ./entitydb --help | grep "entitydb-port"
   ```

3. **Verify Database Configuration:**
   ```bash
   # Query configuration entities
   curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8085/api/v1/entities/query?tag=type:config
   ```

### Common Issues

1. **Flag Not Taking Effect:**
   - Verify database configuration isn't overriding
   - Check environment variable naming
   - Ensure flag is explicitly set (not just default)

2. **Environment Variable Ignored:**
   - Check variable name matches pattern
   - Verify no command-line flag override
   - Check for typos in variable names

3. **Database Configuration Issues:**
   - Verify entity repository is available
   - Check authentication for config endpoints
   - Confirm configuration entity format

## Migration Guide

### From Hardcoded Values

1. **Identify Hardcoded Values:**
   ```bash
   # Search for potential hardcoded values
   grep -r "8085\|admin\|password" . --exclude-dir=.git
   ```

2. **Convert to Environment Variables:**
   ```bash
   # Before (hardcoded)
   port := 8085
   
   # After (configurable)
   port := cfg.Port  // Uses configuration hierarchy
   ```

3. **Update Tools and Scripts:**
   ```bash
   # Before (hardcoded path)
   reader, err := binary.NewReader("./var/entities.edb")
   
   # After (configuration-based)
   cfg := config.Load()
   reader, err := binary.NewReader(cfg.DatabaseFilename)
   ```

### From Legacy Configuration

1. **Update Environment Variables:**
   ```bash
   # Legacy variable names
   export DB_PORT=8085
   
   # New standard names
   export ENTITYDB_PORT=8085
   ```

2. **Convert Configuration Files:**
   ```bash
   # Convert existing config to database entities
   curl -H "Authorization: Bearer $TOKEN" \
     -X POST http://localhost:8085/api/v1/config \
     -d '{"namespace":"server","key":"port","value":"8085"}'
   ```

## Best Practices

### Configuration Organization

1. **Use Consistent Namespaces:**
   - `server.*` - Server configuration
   - `security.*` - Authentication and security
   - `logging.*` - Logging configuration
   - `performance.*` - Performance settings
   - `paths.*` - File and directory paths

2. **Environment-Specific Values:**
   ```bash
   # Development
   export ENTITYDB_LOG_LEVEL=debug
   export ENTITYDB_DEV_MODE=true
   
   # Production
   export ENTITYDB_LOG_LEVEL=info
   export ENTITYDB_HIGH_PERFORMANCE=true
   ```

3. **Documentation:**
   - Document all configuration changes
   - Maintain environment-specific `.env` files
   - Use descriptive variable names
   - Include security considerations

### Performance Optimization

1. **Database Configuration Caching:**
   - 5-minute cache TTL balances performance and responsiveness
   - Use `RefreshConfig()` for immediate updates
   - Monitor cache hit rates in production

2. **Configuration Loading:**
   - Initialize configuration once at startup
   - Cache frequently accessed values
   - Use read locks for concurrent access

3. **Memory Management:**
   - Configuration manager uses minimal memory overhead
   - Automatic cleanup of expired cache entries
   - Thread-safe operations prevent data races

## Related Documentation

- [EntityDB Architecture Overview](../architecture/README.md)
- [Environment Variables Reference](environment-variables.md)
- [Security Configuration Guide](../admin-guide/security-configuration.md)
- [Production Deployment Guide](../admin-guide/production-deployment.md)
- [Troubleshooting Guide](../admin-guide/troubleshooting.md)