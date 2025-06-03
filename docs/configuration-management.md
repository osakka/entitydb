# EntityDB Configuration Management

This document describes the comprehensive configuration management system for EntityDB, which follows a three-tier hierarchy to ensure consistent behavior across all components.

## Configuration Hierarchy

EntityDB uses a three-tier configuration system with the following precedence (highest to lowest):

1. **Database Configuration Entities** (highest priority)
2. **CLI Flags** 
3. **Environment Variables** (lowest priority)

### 1. Database Configuration Entities

Configuration stored as entities in the EntityDB database itself:

- **Config Entities**: `type:config` with `conf:namespace:key` tags
- **Feature Flags**: `type:feature_flag` with `feat:stage:flag` tags

Example:
```bash
# Set database timeout via API
curl -X POST /api/v1/config/set \
  -d '{"namespace": "database", "key": "timeout", "value": "30"}'
```

### 2. CLI Flags

Runtime command-line flags override environment variables:

```bash
./bin/entitydb \
  --port 8080 \
  --ssl-port 8443 \
  --use-ssl true \
  --data /custom/data/path \
  --log-level debug \
  --http-read-timeout 30s \
  --metrics-interval 60s
```

### 3. Environment Variables

Environment variables provide the base configuration layer:

#### Server Configuration
```bash
ENTITYDB_PORT=8085                    # HTTP server port
ENTITYDB_SSL_PORT=8085               # HTTPS server port  
ENTITYDB_USE_SSL=true                # Enable SSL/TLS
ENTITYDB_SSL_CERT=/path/to/cert.pem  # SSL certificate path
ENTITYDB_SSL_KEY=/path/to/key.pem    # SSL private key path
```

#### Paths Configuration
```bash
ENTITYDB_DATA_PATH=/opt/entitydb/var           # Database storage path
ENTITYDB_STATIC_DIR=/opt/entitydb/share/htdocs # Static web files path
```

#### Security Configuration
```bash
ENTITYDB_TOKEN_SECRET=your-secret-key    # JWT token secret
ENTITYDB_SESSION_TTL_HOURS=2            # Session timeout in hours
```

#### HTTP Timeout Configuration
```bash
ENTITYDB_HTTP_READ_TIMEOUT=15           # HTTP read timeout (seconds)
ENTITYDB_HTTP_WRITE_TIMEOUT=15          # HTTP write timeout (seconds)
ENTITYDB_HTTP_IDLE_TIMEOUT=60           # HTTP idle timeout (seconds)
ENTITYDB_SHUTDOWN_TIMEOUT=30            # Server shutdown timeout (seconds)
```

#### Metrics Configuration
```bash
ENTITYDB_METRICS_INTERVAL=30                    # Collection interval (seconds)
ENTITYDB_METRICS_AGGREGATION_INTERVAL=30       # Aggregation interval (seconds)
```

#### API Configuration
```bash
ENTITYDB_SWAGGER_HOST=localhost:8085    # Swagger documentation host
```

#### Logging Configuration
```bash
ENTITYDB_LOG_LEVEL=info                 # Log level: debug, info, warn, error
```

#### Performance Configuration
```bash
ENTITYDB_HIGH_PERFORMANCE=false         # Enable memory-mapped indexing
```

#### Rate Limiting Configuration
```bash
ENTITYDB_ENABLE_RATE_LIMIT=false        # Enable rate limiting
ENTITYDB_RATE_LIMIT_REQUESTS=100        # Requests per window
ENTITYDB_RATE_LIMIT_WINDOW_MINUTES=1    # Rate limit window
```

#### Application Information
```bash
ENTITYDB_APP_NAME="EntityDB Server"     # Application name
ENTITYDB_APP_VERSION="2.24.0"           # Application version
```

## Configuration Files

### Default Configuration
Location: `/opt/entitydb/share/config/entitydb.env`

Contains all default values and serves as the configuration template.

### Instance Configuration  
Location: `/opt/entitydb/var/entitydb.env`

Instance-specific overrides. This file is loaded after defaults and takes precedence.

### Runtime Script Configuration
Location: `/opt/entitydb/bin/entitydbd.sh`

The daemon script loads configuration in this order:
1. Default config from `share/config/entitydb.env`
2. Instance config from `var/entitydb.env`
3. Command-line arguments passed to the script

## Shared Configuration Library

For tools and utilities, EntityDB provides a shared configuration library:

```go
package main

import "entitydb/config"

func main() {
    // Load configuration from environment
    cfg := config.Load()
    
    // Access configuration values
    dbPath := cfg.DatabasePath()
    dataPath := cfg.DataPath
    logLevel := cfg.LogLevel
}
```

### Available Configuration Methods

```go
type Config struct {
    // Server Configuration
    Port             int
    SSLPort          int
    UseSSL           bool
    SSLCert          string
    SSLKey           string
    
    // Paths
    DataPath         string
    StaticDir        string
    
    // Security
    TokenSecret      string
    SessionTTLHours  int
    
    // Timeouts
    HTTPReadTimeout  time.Duration
    HTTPWriteTimeout time.Duration
    HTTPIdleTimeout  time.Duration
    ShutdownTimeout  time.Duration
    
    // Metrics
    MetricsInterval  time.Duration
    AggregationInterval time.Duration
    
    // And more...
}

// Helper methods
func (c *Config) DatabasePath() string  // Full database file path
```

## Migration from Hardcoded Values

### Before (Hardcoded)
```go
// ❌ Hardcoded values
timeout := 15 * time.Second
port := 8085
dataPath := "/opt/entitydb/var"
```

### After (Configurable)
```go
// ✅ Configuration-driven
cfg := config.Load()
timeout := cfg.HTTPReadTimeout
port := cfg.Port  
dataPath := cfg.DataPath
```

## Configuration Validation

EntityDB validates configuration on startup:

1. **Path Validation**: Ensures data and static directories exist
2. **SSL Validation**: Validates certificate and key files if SSL enabled
3. **Port Validation**: Checks port availability
4. **Timeout Validation**: Ensures reasonable timeout values
5. **Secret Validation**: Warns about default/weak secrets

## Tools Configuration

All EntityDB tools now use the shared configuration library:

### Tool Usage Examples

```bash
# Use default database path from environment
./tools/cleanup_old_metrics --dry-run

# Override with custom database path
./tools/cleanup_old_metrics --db /custom/path/entities.db

# Use environment variable
ENTITYDB_DATA_PATH=/custom/path ./tools/cleanup_old_metrics
```

### Tool Configuration Priority

1. Command-line flags (highest)
2. Environment variables
3. Default values from config library (lowest)

## Security Considerations

### Secret Management
- Default token secrets are for development only
- Production deployments should use strong, randomly generated secrets
- Secrets can be provided via environment variables or external secret management

### File Permissions
- Configuration files should have restricted permissions (600 or 640)
- SSL certificate and key files should be readable only by the EntityDB user
- Data directory should be owned by the EntityDB user

### Environment Isolation
- Use instance-specific configuration files for different environments
- Environment variables can be set per deployment environment
- Database configuration entities provide runtime configuration changes

## Best Practices

### Development
1. Use default configuration files for local development
2. Override specific values with environment variables as needed
3. Test configuration changes with `--dry-run` flags where available

### Production
1. Create instance-specific configuration in `var/entitydb.env`
2. Use strong secrets and appropriate timeouts
3. Enable SSL and configure proper certificates
4. Set appropriate log levels for monitoring
5. Configure rate limiting based on expected load

### Tools and Automation
1. Use the shared configuration library for consistency
2. Respect the configuration hierarchy in scripts
3. Provide reasonable defaults for all configuration values
4. Validate configuration before performing operations

## Troubleshooting

### Configuration Issues
- Check configuration loading order and precedence
- Verify file permissions and paths
- Validate environment variable formats
- Review logs for configuration validation errors

### Common Problems
1. **SSL Issues**: Check certificate paths and permissions
2. **Port Conflicts**: Verify ports are not in use
3. **Path Issues**: Ensure directories exist and are writable
4. **Timeout Issues**: Check if values are reasonable for environment
5. **Tool Issues**: Verify tools can access configuration files

## Version History

- **v2.24.0**: Comprehensive configuration management system
  - Added HTTP timeout configuration
  - Added metrics interval configuration  
  - Added shared configuration library
  - Fixed configuration inconsistencies
  - Added configuration validation