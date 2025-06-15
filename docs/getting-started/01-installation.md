---
title: Installation Guide
category: Getting Started
tags: [installation, setup, quick-start]
last_updated: 2025-06-13
version: v2.31.0
---

# EntityDB Installation Guide

This guide walks you through installing and running EntityDB on your system.

## Prerequisites

- Linux, macOS, or Windows with WSL2
- Go 1.19+ (for building from source)
- 512MB+ RAM
- 1GB+ disk space

## Quick Installation

### Binary Release (Recommended)

```bash
# Download the latest release
wget https://git.home.arpa/itdlabs/entitydb/releases/download/v2.31.0/entitydb-v2.31.0-linux-amd64.tar.gz

# Extract and install
tar -xzf entitydb-v2.31.0-linux-amd64.tar.gz
sudo cp entitydb /usr/local/bin/
```

### Build from Source

```bash
# Clone the repository
git clone https://git.home.arpa/itdlabs/entitydb.git
cd entitydb/src

# Build the server
make

# Install scripts
make install
```

## Configuration

EntityDB uses a 3-tier configuration system:
1. Database configuration (highest priority)
2. Command-line flags
3. Environment variables (lowest priority)

### Basic Environment Setup

```bash
export ENTITYDB_HOST=0.0.0.0
export ENTITYDB_PORT=8085
export ENTITYDB_DATA_DIR=/opt/entitydb/var
export ENTITYDB_USE_SSL=false  # Enable for production
```

### SSL Configuration (Production)

```bash
export ENTITYDB_USE_SSL=true
export ENTITYDB_PORT=8443
export ENTITYDB_SSL_CERT=/path/to/cert.pem
export ENTITYDB_SSL_KEY=/path/to/key.pem
```

## First Run

Start EntityDB:

```bash
# Start the server
./bin/entitydbd.sh start

# Check status
./bin/entitydbd.sh status

# View logs
tail -f /opt/entitydb/var/entitydb.log
```

On first start, EntityDB automatically creates:
- Data directory structure
- Default admin user (username: `admin`, password: `admin`)
- System configuration

## Verification

Test your installation:

```bash
# Check server health
curl http://localhost:8085/health

# Access the dashboard
open http://localhost:8085

# Test API
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'
```

## Next Steps

- [Quick Start Tutorial](./02-quick-start.md) - Learn basic operations
- [Core Concepts](./03-core-concepts.md) - Understand EntityDB fundamentals
- [Security Configuration](../50-admin-guides/01-security-configuration.md) - Secure your installation

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Change the port
export ENTITYDB_PORT=8086
```

**Permission denied:**
```bash
# Check data directory permissions
sudo chown -R $USER:$USER /opt/entitydb/var
```

**SSL certificate errors:**
- Verify certificate paths in configuration
- Ensure certificates are valid and readable
- Check [SSL Setup Guide](../70-deployment/03-ssl-setup.md)

For more issues, see [Troubleshooting Guide](../80-troubleshooting/).