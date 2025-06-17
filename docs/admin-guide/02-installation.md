# Production Deployment Guide

> **Version**: v2.32.2 | **Last Updated**: 2025-06-14 | **Status**: AUTHORITATIVE

This guide provides comprehensive instructions for deploying EntityDB in production environments, covering security hardening, performance optimization, monitoring setup, and operational best practices.

## Prerequisites

### System Requirements
- **OS**: Linux (Ubuntu 20.04+ or CentOS 7+)
- **CPU**: 2+ cores (4+ recommended for high load)
- **RAM**: 4GB minimum (8GB+ recommended)
- **Storage**: 50GB+ available space (SSD recommended)
- **Network**: Firewall configuration capability

### Security Requirements
- SSL/TLS certificate (required for HTTPS)
- Firewall management access
- User account management permissions
- Log monitoring capability

## Pre-Deployment Planning

### 1. Infrastructure Design

#### Single-Node Production Setup
```
┌─────────────────────────────────────────┐
│                EntityDB                 │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │    API      │    │   Web Dashboard │ │
│  │   :8085     │    │     :8443       │ │
│  └─────────────┘    └─────────────────┘ │
│  ┌─────────────────────────────────────┐ │
│  │         Binary Storage (EBF)        │ │
│  │           /opt/entitydb/var/        │ │
│  └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

#### Network Configuration
- **HTTPS**: Port 8443 (recommended for production)
- **HTTP**: Port 8085 (internal/health checks only)
- **Firewall**: Restrict access to necessary ports only

### 2. Security Planning

#### SSL/TLS Certificate
- Obtain valid SSL certificate from trusted CA
- Configure DNS for your domain
- Plan certificate renewal process

#### User Access Control
- Define user roles and permissions
- Plan RBAC structure
- Establish password policies

## Installation Process

### 1. System Preparation

```bash
# Create EntityDB user
sudo useradd -r -s /bin/false entitydb
sudo mkdir -p /opt/entitydb
sudo chown entitydb:entitydb /opt/entitydb

# Install dependencies
sudo apt update
sudo apt install -y curl wget systemd

# Configure firewall
sudo ufw allow 8443/tcp  # HTTPS only for production
sudo ufw enable
```

### 2. EntityDB Installation

```bash
# Download and extract EntityDB
cd /tmp
wget https://git.home.arpa/itdlabs/entitydb/releases/download/v2.32.2/entitydb-v2.32.2-linux-amd64.tar.gz
tar -xzf entitydb-v2.32.2-linux-amd64.tar.gz

# Install to production directory
sudo cp -r entitydb-v2.32.2/* /opt/entitydb/
sudo chown -R entitydb:entitydb /opt/entitydb
sudo chmod +x /opt/entitydb/bin/entitydb
sudo chmod +x /opt/entitydb/bin/entitydbd.sh
```

### 3. Production Configuration

#### Environment Configuration
```bash
# Create production environment file
sudo tee /opt/entitydb/var/entitydb.env << 'EOF'
# EntityDB Production Configuration v2.32.2

# Server Configuration
ENTITYDB_BIND_ADDRESS="0.0.0.0"
ENTITYDB_HTTP_PORT="8085"
ENTITYDB_HTTPS_PORT="8443"

# SSL Configuration (REQUIRED for production)
ENTITYDB_USE_SSL="true"
ENTITYDB_SSL_CERT_FILE="/opt/entitydb/var/ssl/server.crt"
ENTITYDB_SSL_KEY_FILE="/opt/entitydb/var/ssl/server.key"

# Database Configuration
ENTITYDB_DATASET_NAME="production"
ENTITYDB_DATASET_PATH="/opt/entitydb/var/db"

# Security Configuration
ENTITYDB_JWT_SECRET="$(openssl rand -hex 32)"
ENTITYDB_JWT_EXPIRY="24h"
ENTITYDB_ADMIN_PASSWORD_HASH="$(python3 -c 'import bcrypt; print(bcrypt.hashpw(b"CHANGE_THIS_PASSWORD", bcrypt.gensalt()).decode())')"

# Performance Configuration
ENTITYDB_MAX_CONTENT_SIZE="104857600"  # 100MB
ENTITYDB_CHUNK_SIZE="4194304"          # 4MB
ENTITYDB_WAL_CHECKPOINT_INTERVAL="1000"
ENTITYDB_WAL_CHECKPOINT_SIZE="104857600"  # 100MB

# Logging Configuration
ENTITYDB_LOG_LEVEL="INFO"
ENTITYDB_LOG_FILE="/opt/entitydb/var/log/entitydb.log"
ENTITYDB_LOG_MAX_SIZE="100MB"
ENTITYDB_LOG_MAX_BACKUPS="10"

# Metrics Configuration
ENTITYDB_METRICS_ENABLED="true"
ENTITYDB_METRICS_INTERVAL="30s"
EOF

# Secure the environment file
sudo chmod 600 /opt/entitydb/var/entitydb.env
```

#### SSL Certificate Setup
```bash
# Create SSL directory
sudo mkdir -p /opt/entitydb/var/ssl
sudo chown entitydb:entitydb /opt/entitydb/var/ssl
sudo chmod 700 /opt/entitydb/var/ssl

# Install your SSL certificate and key
sudo cp your-domain.crt /opt/entitydb/var/ssl/server.crt
sudo cp your-domain.key /opt/entitydb/var/ssl/server.key
sudo chown entitydb:entitydb /opt/entitydb/var/ssl/*
sudo chmod 600 /opt/entitydb/var/ssl/*
```

### 4. Systemd Service Configuration

```bash
# Create systemd service file
sudo tee /etc/systemd/system/entitydb.service << 'EOF'
[Unit]
Description=EntityDB - High-performance temporal database
Documentation=https://git.home.arpa/itdlabs/entitydb
After=network.target
Wants=network.target

[Service]
Type=simple
User=entitydb
Group=entitydb
WorkingDirectory=/opt/entitydb
ExecStart=/opt/entitydb/bin/entitydb
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=entitydb

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/entitydb/var
PrivateTmp=true

# Environment
EnvironmentFile=/opt/entitydb/var/entitydb.env

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable entitydb
sudo systemctl start entitydb
```

## Security Hardening

### 1. Network Security

#### Firewall Configuration
```bash
# Production firewall rules
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (adjust port as needed)
sudo ufw allow 22/tcp

# Allow HTTPS only (HTTP disabled in production)
sudo ufw allow 8443/tcp

# Allow health check from specific monitoring IPs
sudo ufw allow from MONITORING_IP to any port 8085

sudo ufw enable
```

#### Reverse Proxy Setup (Optional)
```nginx
# /etc/nginx/sites-available/entitydb
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/ssl/cert.pem;
    ssl_private_key /path/to/ssl/private.key;
    
    location / {
        proxy_pass https://localhost:8443;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /health {
        proxy_pass http://localhost:8085/health;
        access_log off;
    }
}
```

### 2. User Security

#### Change Default Admin Password
```bash
# First login to get token
RESPONSE=$(curl -s -X POST https://your-domain.com:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo $RESPONSE | jq -r '.data.token')

# Generate new password hash
NEW_PASSWORD="your-secure-password-here"
NEW_SALT=$(openssl rand -hex 16)
NEW_HASH=$(python3 -c "
import bcrypt
password = '$NEW_PASSWORD'.encode()
hash = bcrypt.hashpw(password, bcrypt.gensalt())
print(hash.decode())
")

# Update admin user
curl -X PUT https://your-domain.com:8443/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "admin-user-id",
    "content": "'$NEW_SALT'|'$NEW_HASH'"
  }'
```

### 3. File System Security

```bash
# Set proper ownership and permissions
sudo chown -R entitydb:entitydb /opt/entitydb
sudo chmod -R 750 /opt/entitydb
sudo chmod 600 /opt/entitydb/var/entitydb.env
sudo chmod 700 /opt/entitydb/var/ssl
sudo chmod 600 /opt/entitydb/var/ssl/*

# Secure log files
sudo mkdir -p /opt/entitydb/var/log
sudo chown entitydb:entitydb /opt/entitydb/var/log
sudo chmod 750 /opt/entitydb/var/log
```

## Performance Optimization

### 1. System-Level Optimization

#### File System Configuration
```bash
# Add to /etc/fstab for EntityDB data partition
/dev/sdX /opt/entitydb/var ext4 defaults,noatime,barrier=0 0 2

# Configure system limits
sudo tee -a /etc/security/limits.conf << 'EOF'
entitydb soft nofile 65536
entitydb hard nofile 65536
entitydb soft nproc 32768
entitydb hard nproc 32768
EOF
```

#### Kernel Parameters
```bash
# Add to /etc/sysctl.conf
sudo tee -a /etc/sysctl.conf << 'EOF'
# EntityDB optimizations
vm.swappiness=10
vm.dirty_ratio=15
vm.dirty_background_ratio=5
net.core.somaxconn=4096
net.ipv4.tcp_max_syn_backlog=4096
EOF

sudo sysctl -p
```

### 2. EntityDB Performance Configuration

#### High-Performance Settings
```bash
# Update production environment
sudo tee -a /opt/entitydb/var/entitydb.env << 'EOF'

# Performance Optimization (v2.32.2)
ENTITYDB_TAG_CACHE_SIZE="10000"           # O(1) tag value caching
ENTITYDB_PARALLEL_INDEX_WORKERS="4"       # Parallel index building
ENTITYDB_JSON_ENCODER_POOL_SIZE="100"     # JSON encoder pooling
ENTITYDB_BATCH_WRITE_SIZE="10"            # Batch write operations
ENTITYDB_BATCH_WRITE_TIMEOUT="100ms"      # Batch timeout
ENTITYDB_TEMPORAL_CACHE_SIZE="5000"       # Temporal tag variant caching

# Memory Management
ENTITYDB_MEMORY_LIMIT="4GB"
ENTITYDB_GC_TARGET_PERCENTAGE="80"

# Connection Management
ENTITYDB_MAX_CONNECTIONS="1000"
ENTITYDB_CONNECTION_TIMEOUT="30s"
ENTITYDB_READ_TIMEOUT="30s"
ENTITYDB_WRITE_TIMEOUT="30s"
EOF
```

## Monitoring and Observability

### 1. Health Monitoring Setup

#### Health Check Script
```bash
# Create health check script
sudo tee /opt/entitydb/bin/health-check.sh << 'EOF'
#!/bin/bash
# EntityDB Health Check Script

HEALTH_URL="http://localhost:8085/health"
LOG_FILE="/opt/entitydb/var/log/health-check.log"

# Check health endpoint
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/health-response $HEALTH_URL)
HTTP_CODE="${RESPONSE: -3}"

TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "[$TIMESTAMP] HEALTHY - HTTP $HTTP_CODE" >> $LOG_FILE
    exit 0
else
    echo "[$TIMESTAMP] UNHEALTHY - HTTP $HTTP_CODE" >> $LOG_FILE
    cat /tmp/health-response >> $LOG_FILE
    exit 1
fi
EOF

sudo chmod +x /opt/entitydb/bin/health-check.sh
```

#### Monitoring Cron Job
```bash
# Add health check to crontab
sudo crontab -e
# Add this line:
# */1 * * * * /opt/entitydb/bin/health-check.sh
```

### 2. Metrics Collection

#### Prometheus Integration
```yaml
# prometheus.yml configuration
scrape_configs:
  - job_name: 'entitydb'
    static_configs:
      - targets: ['localhost:8085']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

#### Custom Monitoring Script
```bash
# Create metrics monitoring script
sudo tee /opt/entitydb/bin/monitor-metrics.sh << 'EOF'
#!/bin/bash
# EntityDB Metrics Monitoring

METRICS_URL="http://localhost:8085/api/v1/system/metrics"
LOG_FILE="/opt/entitydb/var/log/metrics.log"

# Collect metrics
METRICS=$(curl -s $METRICS_URL)
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Extract key metrics
ENTITY_COUNT=$(echo $METRICS | jq -r '.data.entities.total_count // 0')
MEMORY_USAGE=$(echo $METRICS | jq -r '.data.system.memory_usage_mb // 0')
GOROUTINES=$(echo $METRICS | jq -r '.data.system.goroutines // 0')

echo "[$TIMESTAMP] Entities: $ENTITY_COUNT, Memory: ${MEMORY_USAGE}MB, Goroutines: $GOROUTINES" >> $LOG_FILE
EOF

sudo chmod +x /opt/entitydb/bin/monitor-metrics.sh
```

## Backup and Recovery

### 1. Backup Strategy

#### Database Backup Script
```bash
# Create backup script
sudo tee /opt/entitydb/bin/backup.sh << 'EOF'
#!/bin/bash
# EntityDB Backup Script

BACKUP_DIR="/opt/entitydb/backups"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BACKUP_NAME="entitydb_backup_$TIMESTAMP"

# Create backup directory
mkdir -p $BACKUP_DIR

# Stop EntityDB temporarily (or use snapshot if available)
systemctl stop entitydb

# Create backup
tar -czf "$BACKUP_DIR/$BACKUP_NAME.tar.gz" \
  -C /opt/entitydb/var \
  db/ log/ ssl/

# Restart EntityDB
systemctl start entitydb

# Keep only last 7 backups
find $BACKUP_DIR -name "entitydb_backup_*.tar.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_NAME.tar.gz"
EOF

sudo chmod +x /opt/entitydb/bin/backup.sh
```

#### Scheduled Backups
```bash
# Add backup to crontab (daily at 2 AM)
sudo crontab -e
# Add this line:
# 0 2 * * * /opt/entitydb/bin/backup.sh
```

### 2. Recovery Procedures

#### Restore from Backup
```bash
# Stop EntityDB
sudo systemctl stop entitydb

# Restore backup
cd /opt/entitydb/var
sudo tar -xzf /opt/entitydb/backups/entitydb_backup_TIMESTAMP.tar.gz

# Fix permissions
sudo chown -R entitydb:entitydb /opt/entitydb/var

# Start EntityDB
sudo systemctl start entitydb
```

## Troubleshooting

### 1. Common Issues

#### Service Won't Start
```bash
# Check service status
sudo systemctl status entitydb

# Check logs
sudo journalctl -u entitydb -n 50

# Check configuration
sudo -u entitydb /opt/entitydb/bin/entitydb --validate-config
```

#### Performance Issues
```bash
# Check system resources
top -p $(pgrep entitydb)
df -h /opt/entitydb/var

# Check EntityDB metrics
curl http://localhost:8085/api/v1/system/metrics
```

#### SSL Certificate Issues
```bash
# Test SSL configuration
openssl s_client -connect your-domain.com:8443 -servername your-domain.com

# Check certificate expiry
openssl x509 -in /opt/entitydb/var/ssl/server.crt -noout -dates
```

### 2. Log Analysis

#### Key Log Patterns
```bash
# Authentication failures
grep "auth.*failed" /opt/entitydb/var/log/entitydb.log

# Performance warnings
grep "slow.*query" /opt/entitydb/var/log/entitydb.log

# Error patterns
grep "ERROR" /opt/entitydb/var/log/entitydb.log
```

## Maintenance Procedures

### 1. Regular Maintenance

#### Weekly Tasks
- Check system resources and performance metrics
- Review authentication and error logs
- Verify backup integrity
- Update SSL certificates if needed

#### Monthly Tasks
- Review user access and permissions
- Analyze storage usage and growth trends
- Test backup and recovery procedures
- Update monitoring thresholds if needed

### 2. Updates and Upgrades

#### Upgrade Process
```bash
# 1. Backup current installation
/opt/entitydb/bin/backup.sh

# 2. Download new version
wget https://git.home.arpa/itdlabs/entitydb/releases/download/vX.Y.Z/entitydb-vX.Y.Z-linux-amd64.tar.gz

# 3. Stop service
sudo systemctl stop entitydb

# 4. Install new version
tar -xzf entitydb-vX.Y.Z-linux-amd64.tar.gz
sudo cp entitydb-vX.Y.Z/bin/* /opt/entitydb/bin/
sudo chown entitydb:entitydb /opt/entitydb/bin/*
sudo chmod +x /opt/entitydb/bin/*

# 5. Start service
sudo systemctl start entitydb

# 6. Verify upgrade
curl http://localhost:8085/health
```

## Security Checklist

### Pre-Production Checklist
- [ ] SSL/TLS certificate installed and validated
- [ ] Default admin password changed
- [ ] Firewall configured (only HTTPS port open)
- [ ] Environment file secured (600 permissions)
- [ ] Service running as non-root user
- [ ] Log monitoring configured
- [ ] Backup system tested

### Ongoing Security Tasks
- [ ] Regular password rotation
- [ ] User access review
- [ ] SSL certificate renewal
- [ ] Security patch management
- [ ] Log analysis for suspicious activity

---

*This production deployment guide ensures secure, performant, and maintainable EntityDB installations. For additional security configuration, see [Security Configuration](../admin-guide/01-security-configuration.md).*