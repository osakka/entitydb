# EntityDB Configuration System

## Overview

EntityDB v2.13.0 introduces a streamlined environment-based configuration system that provides flexibility while maintaining simplicity.

## Configuration Hierarchy

Configuration settings are loaded in the following order (highest precedence first):

1. **Command Line Flags** - Runtime overrides for any setting
2. **Environment Variables** - Set in shell or configuration files
3. **Instance Configuration** - `var/entitydb.env` for instance-specific settings
4. **Default Configuration** - `share/config/entitydb_server.env` with all defaults
5. **Hardcoded Defaults** - Built into the application code

## Configuration Files

### Default Configuration
- **Location**: `share/config/entitydb_server.env`
- **Purpose**: Contains all available settings with sensible defaults
- **Usage**: Reference for available options, rarely needs modification

### Instance Configuration
- **Location**: `var/entitydb.env`
- **Purpose**: Override specific settings for this EntityDB instance
- **Usage**: Copy from default config and modify only needed values

## Environment Variables

All configuration can be controlled via environment variables with the `ENTITYDB_` prefix:

| Variable | Default | Description |
|----------|---------|-------------|
| `ENTITYDB_PORT` | 8085 | HTTP server port |
| `ENTITYDB_SSL_PORT` | 8443 | HTTPS server port |
| `ENTITYDB_USE_SSL` | false | Enable SSL/TLS |
| `ENTITYDB_SSL_CERT` | /etc/ssl/certs/server.pem | SSL certificate path |
| `ENTITYDB_SSL_KEY` | /etc/ssl/private/server.key | SSL private key path |
| `ENTITYDB_DATA_PATH` | /opt/entitydb/var | Data storage directory |
| `ENTITYDB_LOG_LEVEL` | info | Logging level (debug, info, warn, error) |
| `ENTITYDB_TOKEN_SECRET` | entitydb-secret-key | JWT token signing key |
| `ENTITYDB_SESSION_TTL_HOURS` | 2 | Session timeout in hours |
| `ENTITYDB_ENABLE_RATE_LIMIT` | false | Enable rate limiting |
| `ENTITYDB_RATE_LIMIT_REQUESTS` | 100 | Requests per window |
| `ENTITYDB_RATE_LIMIT_WINDOW_MINUTES` | 1 | Rate limit window size |

## Dynamic Configuration

In addition to static configuration, EntityDB supports dynamic configuration through entities:

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