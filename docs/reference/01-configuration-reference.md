# EntityDB Configuration System

## Overview

EntityDB v2.32.0 features a comprehensive three-tier configuration hierarchy system with support for database-stored configuration, complete flag coverage, and elimination of hardcoded values.

## Configuration Hierarchy

Configuration settings are loaded in the following order (highest precedence first):

1. **Database Configuration Entities** - Runtime configuration stored as entities (highest priority)
2. **Command Line Flags** - Explicit runtime overrides with `--entitydb-*` format
3. **Environment Variables** - Shell environment with `ENTITYDB_*` prefix (lowest priority)

### Three-Tier Configuration Benefits

- **Database Configuration**: Runtime changes without server restart, cached for performance
- **CLI Flags**: Explicit overrides for deployment and testing scenarios  
- **Environment Variables**: Infrastructure-level configuration for containers and deployments

## Configuration Methods

### Database Configuration (Highest Priority)
- **Storage**: Configuration entities with `type:config` tags
- **Runtime Updates**: Changes apply immediately with 5-minute cache TTL
- **Management**: Via API endpoints or direct entity manipulation
- **Example**: Create entity with `name:default_admin_username` and `value:prodadmin` tags

### Command Line Flags (Medium Priority)
- **Format**: All flags use long format `--entitydb-*` (no conflicts with other tools)
- **Reserved Short Flags**: Only `-h/--help` and `-v/--version` for essential functions
- **Override Behavior**: Only explicitly set flags override environment variables
- **Usage**: `./entitydb --entitydb-port=8080 --entitydb-use-ssl=true`

### Environment Variables (Lowest Priority)
- **Prefix**: All variables use `ENTITYDB_*` naming convention
- **Auto-loading**: Automatically loaded at startup with sensible defaults
- **Container-friendly**: Perfect for Docker, Kubernetes, and other deployment platforms

## Environment Variables

All configuration can be controlled via environment variables with the `ENTITYDB_` prefix:

### Server Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_PORT` | 8085 | HTTP server port |
| `ENTITYDB_SSL_PORT` | 8085 | HTTPS server port |
| `ENTITYDB_USE_SSL` | false | Enable SSL/TLS |
| `ENTITYDB_SSL_CERT` | ./certs/server.pem | SSL certificate path |
| `ENTITYDB_SSL_KEY` | ./certs/server.key | SSL private key path |
| `ENTITYDB_DATA_PATH` | ./var | Data storage directory |
| `ENTITYDB_STATIC_DIR` | ./share/htdocs | Static files directory |

### Security Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_TOKEN_SECRET` | entitydb-secret-key | JWT token signing key |
| `ENTITYDB_SESSION_TTL_HOURS` | 2 | Session timeout in hours |
| `ENTITYDB_DEFAULT_ADMIN_USERNAME` | admin | Default admin username |
| `ENTITYDB_DEFAULT_ADMIN_PASSWORD` | admin | Default admin password ⚠️ |
| `ENTITYDB_DEFAULT_ADMIN_EMAIL` | admin@entitydb.local | Default admin email |
| `ENTITYDB_SYSTEM_USER_ID` | 00000000000000000000000000000001 | System user UUID |
| `ENTITYDB_SYSTEM_USERNAME` | system | System username |
| `ENTITYDB_BCRYPT_COST` | 10 | Password hashing cost (4-31) |

### Logging and Debugging
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_LOG_LEVEL` | info | Log level (trace, debug, info, warn, error) |
| `ENTITYDB_TRACE_SUBSYSTEMS` | "" | Trace subsystems (comma-separated) |
| `ENTITYDB_DEV_MODE` | false | Enable development mode |
| `ENTITYDB_DEBUG_PORT` | 6060 | Debug/profiling port |
| `ENTITYDB_PROFILE_ENABLED` | false | Enable CPU/memory profiling |

### Performance and Timeouts
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_HIGH_PERFORMANCE` | false | Enable high-performance mode |
| `ENTITYDB_HTTP_READ_TIMEOUT` | 15 | HTTP read timeout (seconds) |
| `ENTITYDB_HTTP_WRITE_TIMEOUT` | 15 | HTTP write timeout (seconds) |
| `ENTITYDB_HTTP_IDLE_TIMEOUT` | 60 | HTTP idle timeout (seconds) |
| `ENTITYDB_SHUTDOWN_TIMEOUT` | 30 | Server shutdown timeout (seconds) |

### Rate Limiting
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_ENABLE_RATE_LIMIT` | false | Enable rate limiting |
| `ENTITYDB_RATE_LIMIT_REQUESTS` | 100 | Requests per window |
| `ENTITYDB_RATE_LIMIT_WINDOW_MINUTES` | 1 | Rate limit window size |

### Metrics Collection
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_METRICS_INTERVAL` | 30 | Metrics collection interval (seconds) |
| `ENTITYDB_METRICS_AGGREGATION_INTERVAL` | 30 | Metrics aggregation interval (seconds) |
| `ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING` | true | Enable HTTP request metrics |
| `ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING` | true | Enable storage metrics |

### File and Path Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_DATABASE_FILENAME` | entities.ebf | Main database filename |
| `ENTITYDB_WAL_SUFFIX` | .wal | Write-Ahead Log file suffix |
| `ENTITYDB_INDEX_SUFFIX` | .idx | Index file suffix |
| `ENTITYDB_BACKUP_PATH` | ./backup | Backup directory path |
| `ENTITYDB_TEMP_PATH` | ./tmp | Temporary files directory |
| `ENTITYDB_PID_FILE` | ./var/entitydb.pid | Process ID file path |
| `ENTITYDB_LOG_FILE` | ./var/entitydb.log | Server log file path |

## Command Line Flags

All configuration options are available as command-line flags using the `--entitydb-*` format:

```bash
# Server configuration
./entitydb --entitydb-port=8080 --entitydb-use-ssl=true

# Security configuration  
./entitydb --entitydb-default-admin-username=prodadmin \
           --entitydb-default-admin-password=secure123 \
           --entitydb-bcrypt-cost=12

# Performance tuning
./entitydb --entitydb-high-performance=true \
           --entitydb-http-read-timeout=30s \
           --entitydb-metrics-interval=10s

# Development mode
./entitydb --entitydb-dev-mode=true \
           --entitydb-log-level=debug \
           --entitydb-trace-subsystems=auth,storage
```

### Flag Naming Convention

- **Long Format Only**: All configuration flags use `--entitydb-*` format to avoid conflicts
- **Reserved Short Flags**: Only `-h/--help` and `-v/--version` for essential functions  
- **Consistency**: Flag names match environment variable names (lowercase, hyphens vs underscores)

### Override Behavior

- Flags only override environment variables if explicitly provided
- Use `./entitydb --help` to see current values including environment overrides
- Explicit flag detection prevents accidental overrides of environment configuration

## Database Configuration (Runtime Updates)

EntityDB supports runtime configuration changes through database entities with `type:config` tags:

### Configuration Entity Format

```json
{
  "id": "config_default_admin_username",
  "tags": [
    "type:config",
    "name:default_admin_username", 
    "value:prodadmin",
    "category:security",
    "description:Default admin username for new installations"
  ]
}
```

### Configuration Cache

- **Cache TTL**: 5 minutes for performance optimization
- **Immediate Updates**: Changes apply on next cache refresh or explicit refresh
- **Thread Safety**: All configuration access is thread-safe with read-write mutexes

## Security Considerations

### Production Security Checklist

⚠️ **Critical**: Change these defaults in production environments:

1. **Admin Credentials**: Set `ENTITYDB_DEFAULT_ADMIN_PASSWORD` to a strong password
2. **Token Secret**: Use a cryptographically secure `ENTITYDB_TOKEN_SECRET` (minimum 32 characters)
3. **SSL/TLS**: Enable `ENTITYDB_USE_SSL=true` with proper certificates
4. **Bcrypt Cost**: Consider increasing `ENTITYDB_BCRYPT_COST` to 12+ for enhanced security
5. **System User ID**: Only change `ENTITYDB_SYSTEM_USER_ID` during initial setup

### Security Best Practices

```bash
# Production-ready security configuration
export ENTITYDB_DEFAULT_ADMIN_USERNAME="admin"
export ENTITYDB_DEFAULT_ADMIN_PASSWORD="$(openssl rand -base64 32)"
export ENTITYDB_TOKEN_SECRET="$(openssl rand -base64 32)"  
export ENTITYDB_BCRYPT_COST=12
export ENTITYDB_USE_SSL=true
export ENTITYDB_SESSION_TTL_HOURS=2
```

## Migration from Hardcoded Values

### V2.32.0 Configuration Migration

Previous versions had hardcoded values that are now configurable:

| Previously Hardcoded | Now Configurable | Environment Variable |
|---------------------|------------------|---------------------|
| Admin username: "admin" | ✅ Configurable | `ENTITYDB_DEFAULT_ADMIN_USERNAME` |
| Admin password: "admin" | ✅ Configurable | `ENTITYDB_DEFAULT_ADMIN_PASSWORD` |
| Admin email: "admin@entitydb.local" | ✅ Configurable | `ENTITYDB_DEFAULT_ADMIN_EMAIL` |
| System user ID: "00000000000000000000000000000001" | ✅ Configurable | `ENTITYDB_SYSTEM_USER_ID` |
| System username: "system" | ✅ Configurable | `ENTITYDB_SYSTEM_USERNAME` |
| Bcrypt cost: 10 | ✅ Configurable | `ENTITYDB_BCRYPT_COST` |

### Backward Compatibility

- **Default Values**: All new configuration options maintain backward-compatible defaults
- **No Breaking Changes**: Existing installations continue to work without modification
- **Gradual Migration**: Update configuration at your own pace using environment variables or flags

## Troubleshooting Configuration

### Common Configuration Issues

1. **Environment Variables Not Loading**
   ```bash
   # Check if environment variables are set
   env | grep ENTITYDB_
   
   # Verify configuration loading
   ./entitydb --help | grep -A5 "Default:"
   ```

2. **Flag Override Not Working**
   ```bash
   # Flags must be explicitly set to override environment
   ./entitydb --entitydb-port=9090  # ✅ Correct
   ENTITYDB_PORT=8080 ./entitydb    # ✅ Also correct
   ```

3. **Database Configuration Cache Issues**
   ```bash
   # Configuration cache refreshes every 5 minutes
   # For immediate updates, restart the server or wait for cache expiry
   ```

### Configuration Validation

The server validates configuration at startup and logs any issues:

```bash
# Check startup logs for configuration errors
tail -f var/entitydb.log | grep -i config
```

## Dynamic Configuration

Legacy dynamic configuration through entities remains supported for backward compatibility, but the new three-tier system is recommended for all new configurations.

### Configuration Entities
- **Type**: `type:config`
- **Tags**: `conf:namespace:key`
- **Content**: JSON value

### Feature Flags
- **Type**: `type:feature_flag`
- **Tags**: `feat:stage:flag`
- **Status**: `status:enabled` or `status:disabled`

## Migration from v2.12.0

The main change in v2.13.0 is the removal of the `--config` flag, which was unused. All configuration is now handled through environment variables and configuration files.

## SSL Configuration

EntityDB v2.13.0 includes comprehensive SSL/TLS support for securing your server:

### Basic SSL Configuration
```bash
# Enable SSL
ENTITYDB_USE_SSL=true

# Specify certificate and key paths
ENTITYDB_SSL_CERT=/etc/ssl/certs/server.pem
ENTITYDB_SSL_KEY=/etc/ssl/private/server.key

# Use standard HTTPS port
ENTITYDB_SSL_PORT=443
```

### SSL on Same Port
Starting with v2.13.0, you can run SSL on the same port as HTTP:

```bash
# Enable SSL
ENTITYDB_USE_SSL=true

# Use same port for both HTTP and HTTPS
ENTITYDB_PORT=8085
ENTITYDB_SSL_PORT=8085
```

### SSL Certificate Verification
On startup, EntityDB will verify that your SSL certificates:
- Exist and are readable
- Are valid X.509 certificates
- Have not expired

The server will display certificate information on startup.

## Examples

### Development Setup
```bash
# var/entitydb.env
ENTITYDB_USE_SSL=false
ENTITYDB_LOG_LEVEL=debug
ENTITYDB_PORT=8080
```

### Production Setup
```bash
# var/entitydb.env
ENTITYDB_USE_SSL=true
ENTITYDB_SSL_CERT=/path/to/cert.pem
ENTITYDB_SSL_KEY=/path/to/key.pem
ENTITYDB_LOG_LEVEL=warn
ENTITYDB_TOKEN_SECRET=$(openssl rand -hex 32)
```

### Secure Production with SSL on Same Port
```bash
# var/entitydb.env
ENTITYDB_USE_SSL=true
ENTITYDB_SSL_CERT=/etc/ssl/certs/server.pem
ENTITYDB_SSL_KEY=/etc/ssl/private/server.key
ENTITYDB_PORT=443
ENTITYDB_SSL_PORT=443
ENTITYDB_LOG_LEVEL=warn
ENTITYDB_TOKEN_SECRET=$(openssl rand -hex 32)
```

### Running with Overrides
```bash
ENTITYDB_LOG_LEVEL=debug ./bin/entitydbd.sh start
```

## Best Practices

1. **Never commit** instance configuration files with secrets
2. **Use environment variables** for sensitive values in production
3. **Document all changes** to configuration in your instance config
4. **Test configuration changes** in development before production
5. **Keep defaults secure** - SSL disabled by default for development only