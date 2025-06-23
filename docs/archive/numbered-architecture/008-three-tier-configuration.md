# ADR-008: Three-Tier Configuration Hierarchy

## Status
Accepted (2025-06-07)

## Context
EntityDB v2.27.0 through v2.32.0 underwent a comprehensive configuration management overhaul. The previous system had numerous hardcoded values throughout the codebase, making production deployment and customization difficult.

### Problems with Hardcoded Configuration
- **Security Issues**: Hardcoded admin credentials (admin/admin) in production
- **Deployment Inflexibility**: Cannot customize for different environments
- **Maintenance Overhead**: Configuration changes required code modifications
- **Poor Security Practices**: Sensitive values embedded in source code
- **Limited Customization**: No way to override system parameters

### Configuration Sources Evaluated
1. **Configuration Files Only**: Traditional config file approach
2. **Environment Variables Only**: 12-factor app methodology
3. **Command Line Arguments**: Explicit parameter passing
4. **Database Configuration**: Runtime configuration storage
5. **Hybrid Approach**: Multiple sources with precedence hierarchy

## Decision
We decided to implement a **three-tier configuration hierarchy** with clear precedence rules:

```
1. Database Configuration Entities (HIGHEST PRIORITY)
2. CLI Flags  
3. Environment Variables (LOWEST PRIORITY)
```

### Hierarchy Rationale
- **Database**: Runtime configuration changes without restart
- **CLI Flags**: Explicit deployment-time configuration
- **Environment**: Infrastructure-level defaults

### Implementation Architecture
```go
type ConfigManager struct {
    dbConfig    map[string]string // Database entities
    cliFlags    map[string]string // Command line flags
    envVars     map[string]string // Environment variables
    cache       map[string]string // Resolved values with TTL
    cacheTTL    time.Duration     // 5-minute cache expiry
}

func (cm *ConfigManager) Get(key string) string {
    // Check cache first
    if value, exists := cm.cache[key]; exists {
        return value
    }
    
    // Apply hierarchy: Database > CLI > Environment
    if value, exists := cm.dbConfig[key]; exists {
        cm.cache[key] = value
        return value
    }
    
    if value, exists := cm.cliFlags[key]; exists {
        cm.cache[key] = value
        return value
    }
    
    if value, exists := cm.envVars[key]; exists {
        cm.cache[key] = value
        return value
    }
    
    return "" // No configuration found
}
```

## Consequences

### Positive
- **Production Security**: Configurable admin credentials and system parameters
- **Deployment Flexibility**: Environment-specific configuration without code changes  
- **Runtime Updates**: Database configuration changes without restart
- **Hierarchical Override**: Clear precedence for different deployment scenarios
- **Zero Hardcoded Values**: Complete elimination of hardcoded configuration
- **Cache Performance**: 5-minute TTL reduces database load for config access

### Negative
- **Complexity**: More complex configuration resolution logic
- **Debugging**: Configuration source may not be immediately obvious
- **Migration Effort**: Required updating all hardcoded values in codebase
- **Cache Inconsistency**: Potential 5-minute delay for configuration updates

### Security Improvements
- **Configurable Admin**: `ENTITYDB_DEFAULT_ADMIN_USERNAME/PASSWORD/EMAIL`
- **System User Config**: `ENTITYDB_SYSTEM_USER_ID/USERNAME`
- **Bcrypt Cost**: `ENTITYDB_BCRYPT_COST` for security vs performance tuning
- **SSL Configuration**: Runtime SSL certificate and key configuration

## Configuration Categories

### Server Configuration
```bash
ENTITYDB_PORT=8085                    # HTTP port
ENTITYDB_SSL_PORT=8085               # HTTPS port  
ENTITYDB_USE_SSL=true                # Enable SSL/TLS
ENTITYDB_DATA_PATH=/opt/entitydb/var # Database storage
```

### Security Configuration
```bash
ENTITYDB_DEFAULT_ADMIN_USERNAME=admin
ENTITYDB_DEFAULT_ADMIN_PASSWORD=admin
ENTITYDB_DEFAULT_ADMIN_EMAIL=admin@entitydb.local
ENTITYDB_BCRYPT_COST=12
ENTITYDB_TOKEN_SECRET=entitydb-secret-key
```

### Performance Configuration
```bash
ENTITYDB_HTTP_READ_TIMEOUT=60        # HTTP timeouts
ENTITYDB_METRICS_INTERVAL=30         # Metrics collection
ENTITYDB_HIGH_PERFORMANCE=true       # Enable optimizations
```

### CLI Flag Standardization
All flags follow `--entitydb-*` format:
```bash
--entitydb-port=8085
--entitydb-use-ssl=true
--entitydb-data-path=/opt/entitydb/var
--entitydb-log-level=info
```

## Implementation History
- v2.27.0: Initial three-tier configuration system (June 7, 2025)
- v2.32.0: Complete hardcoded value elimination (June 16, 2025)

### Hardcoded Values Eliminated
- Admin user credentials (username, password, email)
- System user parameters (ID, username)
- Bcrypt cost factor
- HTTP timeout values
- File paths and directory locations
- Default ports and SSL settings
- Logging levels and subsystems

## Configuration File Structure
```ini
# /opt/entitydb/share/config/entitydb.env (defaults)
ENTITYDB_PORT=8085
ENTITYDB_USE_SSL=true
ENTITYDB_LOG_LEVEL=info

# /opt/entitydb/var/entitydb.env (instance overrides)
ENTITYDB_DEFAULT_ADMIN_PASSWORD=secure_production_password
ENTITYDB_TOKEN_SECRET=production_secret_key
```

## Database Configuration Entities
Runtime configuration stored as entities:
```json
{
  "id": "config-http-timeout",
  "tags": ["type:config", "config:http_read_timeout"],
  "content": "60"
}
```

## Runtime Configuration API
```bash
# Get configuration value
GET /api/v1/admin/config?key=http_read_timeout

# Set configuration value (requires admin)
POST /api/v1/admin/config
{
  "key": "http_read_timeout",
  "value": "120"
}
```

## Migration Strategy
1. **Audit Phase**: Identify all hardcoded values using comprehensive grep
2. **Configuration Phase**: Create environment variables and CLI flags
3. **Default Phase**: Establish sensible defaults in config files
4. **Testing Phase**: Verify all configuration sources work correctly
5. **Documentation Phase**: Update all documentation with new patterns

## Related Decisions
- [ADR-006: Credential Storage in Entities](./006-credential-storage-in-entities.md) - Configurable admin credentials
- [ADR-002: Binary Storage Format](./002-binary-storage-format.md) - Configurable storage paths