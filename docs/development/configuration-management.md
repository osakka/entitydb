# EntityDB Configuration Management

## Overview

EntityDB uses a three-tier configuration hierarchy with the following priority order (highest to lowest):

1. **Database Configuration Entities** (highest priority)
2. **Command-Line Flags** 
3. **Environment Variables** (lowest priority)

This design ensures that configuration can be managed at runtime through the database while still allowing environment-based deployment and command-line overrides for testing.

## Configuration Sources

### 1. Environment Variables

All configuration options can be set via environment variables with the prefix `ENTITYDB_`. These are loaded from:

- System environment
- Default configuration file: `share/config/entitydb.env`
- Instance configuration file: `var/entitydb.env` (overrides defaults)

Example:
```bash
export ENTITYDB_PORT=8085
export ENTITYDB_USE_SSL=true
export ENTITYDB_LOG_LEVEL=debug
```

### 2. Command-Line Flags

All configuration options use long flag names following the pattern `--entitydb-<component>-<setting>`. Short flags are reserved only for essential functions:

- `-h`, `--help` - Show help
- `-v`, `--version` - Show version

Example:
```bash
entitydb --entitydb-port 8085 --entitydb-use-ssl --entitydb-log-level debug
```

### 3. Database Configuration

Configuration can be stored as entities with `type:config` and `conf:namespace:key` tags. These take the highest priority and can be modified at runtime.

Example entity:
```json
{
  "id": "config_server_port",
  "tags": ["type:config", "conf:server:port"],
  "content": {"port": 8443}
}
```

## Configuration Options

### Server Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_PORT | --entitydb-port | server.port | 8085 | HTTP server port |
| ENTITYDB_SSL_PORT | --entitydb-ssl-port | server.ssl_port | 8085 | HTTPS server port |
| ENTITYDB_USE_SSL | --entitydb-use-ssl | server.use_ssl | false | Enable SSL/TLS |
| ENTITYDB_SSL_CERT | --entitydb-ssl-cert | server.ssl_cert | ./certs/server.pem | SSL certificate path |
| ENTITYDB_SSL_KEY | --entitydb-ssl-key | server.ssl_key | ./certs/server.key | SSL private key path |

### Storage Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_DATA_PATH | --entitydb-data-path | paths.data | ./var | Data directory path |
| ENTITYDB_STATIC_DIR | --entitydb-static-dir | paths.static | ./share/htdocs | Static files directory |

### Security Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_TOKEN_SECRET | --entitydb-token-secret | security.token_secret | entitydb-secret-key | JWT token secret |
| ENTITYDB_SESSION_TTL_HOURS | --entitydb-session-ttl-hours | security.session_ttl_hours | 2 | Session timeout in hours |

### Logging Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_LOG_LEVEL | --entitydb-log-level | logging.level | info | Log level (trace, debug, info, warn, error) |
| ENTITYDB_TRACE_SUBSYSTEMS | --entitydb-trace-subsystems | logging.trace_subsystems | - | Comma-separated list of trace subsystems |
| ENTITYDB_HTTP_TRACE | N/A | N/A | false | Enable HTTP request tracing |

Available trace subsystems:
- `auth` - Authentication and authorization flow
- `storage` - Storage operations and transactions  
- `cache` - Cache operations
- `temporal` - Temporal operations and indexing
- `lock` - Lock acquisition and contention
- `query` - Query execution and optimization
- `metrics` - Metrics collection
- `dataspace` - Dataspace operations
- `relationship` - Entity relationships
- `chunking` - Content chunking operations

### Performance Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_HIGH_PERFORMANCE | --entitydb-high-performance | performance.high_performance | false | Enable memory-mapped indexing |

### HTTP Timeouts

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_HTTP_READ_TIMEOUT | --entitydb-http-read-timeout | timeouts.http_read | 15s | HTTP read timeout |
| ENTITYDB_HTTP_WRITE_TIMEOUT | --entitydb-http-write-timeout | timeouts.http_write | 15s | HTTP write timeout |
| ENTITYDB_HTTP_IDLE_TIMEOUT | --entitydb-http-idle-timeout | timeouts.http_idle | 60s | HTTP idle timeout |
| ENTITYDB_SHUTDOWN_TIMEOUT | --entitydb-shutdown-timeout | timeouts.shutdown | 30s | Server shutdown timeout |

### Metrics Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_METRICS_INTERVAL | --entitydb-metrics-interval | metrics.collection_interval | 30s | Metrics collection interval |
| ENTITYDB_METRICS_AGGREGATION_INTERVAL | --entitydb-metrics-aggregation-interval | metrics.aggregation_interval | 30s | Metrics aggregation interval |

### Rate Limiting

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_ENABLE_RATE_LIMIT | --entitydb-enable-rate-limit | rate_limit.enabled | false | Enable rate limiting |
| ENTITYDB_RATE_LIMIT_REQUESTS | --entitydb-rate-limit-requests | rate_limit.requests | 100 | Requests per window |
| ENTITYDB_RATE_LIMIT_WINDOW_MINUTES | --entitydb-rate-limit-window-minutes | rate_limit.window_minutes | 1 | Rate limit window in minutes |

### API Configuration

| Environment Variable | Command-Line Flag | Database Key | Default | Description |
|---------------------|-------------------|--------------|---------|-------------|
| ENTITYDB_SWAGGER_HOST | --entitydb-swagger-host | api.swagger_host | localhost:8085 | Swagger documentation host |

## Configuration Management API

### Get Configuration

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://localhost:8085/api/v1/config?namespace=server&key=port"
```

### Set Configuration

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"namespace": "server", "key": "port", "value": "8443"}' \
  "https://localhost:8085/api/v1/config/set"
```

## Implementation Details

### Configuration Manager

The `ConfigManager` (located in `src/config/manager.go`) handles the three-tier hierarchy:

1. Loads environment variables on startup
2. Applies command-line flags if explicitly set
3. Queries database for configuration entities
4. Caches database configuration with 5-minute expiry
5. Provides refresh mechanism for runtime updates

### Configuration Initialization

```go
// Create configuration manager
configManager := config.NewConfigManager(entityRepo)

// Register all configuration flags
configManager.RegisterFlags()

// Parse command line
flag.Parse()

// Initialize with proper hierarchy
cfg, err := configManager.Initialize()
```

### Runtime Configuration Updates

Configuration stored in the database can be updated at runtime:

```go
// Set configuration value
err := configManager.SetDatabaseConfig("server", "port", "8443")

// Refresh configuration from database
err := configManager.RefreshConfig()

// Get current configuration
cfg := configManager.GetConfig()
```

## Migration Guide

### For Existing Deployments

1. **Update Environment Variables**: Ensure all environment variables use the `ENTITYDB_` prefix
2. **Update Command-Line Flags**: Replace short flags with long flags:
   - `-port` → `--entitydb-port`
   - `-data` → `--entitydb-data-path`
   - `-static-dir` → `--entitydb-static-dir`
   - `-log-level` → `--entitydb-log-level`
3. **Update Scripts**: Modify startup scripts to use new flag names
4. **Test Configuration**: Verify configuration hierarchy works as expected

### For Tools and Utilities

Tools should use the common configuration system:

```go
import "entitydb/tools/config"

cfg := config.GetToolConfig()
config.RegisterToolFlags(cfg)
flag.Parse()

// Use cfg.DataPath, cfg.APIEndpoint, etc.
```

## Best Practices

1. **Use Environment Variables** for deployment-specific settings
2. **Use Command-Line Flags** for testing and development overrides
3. **Use Database Configuration** for runtime-adjustable settings
4. **Avoid Hardcoded Values** - always use configuration system
5. **Document New Options** - update this guide when adding configuration

## Security Considerations

1. **Sensitive Values**: Store secrets (tokens, passwords) in environment variables, not database
2. **File Permissions**: Ensure configuration files have appropriate permissions (600)
3. **SSL Certificates**: Use absolute paths or paths relative to EntityDB root
4. **Token Rotation**: Regularly rotate JWT token secrets in production

## Troubleshooting

### Configuration Not Taking Effect

1. Check priority hierarchy - database overrides flags, flags override environment
2. Verify environment variables are exported
3. Check for typos in flag names (all use `--entitydb-` prefix)
4. Review logs for configuration loading errors

### Database Configuration Not Working

1. Ensure entity has correct tags: `type:config` and `conf:namespace:key`
2. Check entity content is valid JSON
3. Verify configuration cache refresh (5-minute default)
4. Check RBAC permissions for configuration access

### Path Resolution Issues

1. All paths are resolved relative to EntityDB installation directory
2. Use absolute paths for production deployments
3. Check working directory when running from different locations