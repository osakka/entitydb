# EntityDB Quick Start Guide

> **Version**: v2.30.0 | **Last Updated**: 2025-06-12 | **Status**: GUIDANCE

Welcome to EntityDB! This guide will get you up and running with EntityDB v2.30.0 in just a few minutes.

> **⚠️ Critical**: v2.29.0+ includes major authentication changes. User credentials are now embedded in user entities. All users from previous versions must be recreated.

## Prerequisites

- Git
- Go 1.19+ (for building from source)
- Linux/macOS environment
- `jq` (for JSON parsing in examples)

## Installation

### 1. Clone Repository
```bash
git clone https://git.home.arpa/itdlabs/entitydb.git
cd entitydb
```

### 2. Build Server
```bash
cd src
make
cd ..
```

### 3. Verify Installation
```bash
./bin/entitydb --version
# Should output: EntityDB v2.30.0
```

## Starting EntityDB

### Start the Server
```bash
# Start server daemon with SSL enabled (default)
./bin/entitydbd.sh start

# Check server status
./bin/entitydbd.sh status

# View server logs
./bin/entitydbd.sh logs

# Stop server
./bin/entitydbd.sh stop
```

### Server Configuration
- **URL**: https://localhost:8085 (SSL enabled by default)
- **Data**: Stored in `/opt/entitydb/var/`
- **Config**: `/opt/entitydb/share/config/entitydb.env`

> **SSL Note**: EntityDB uses SSL by default. The `-k` flag in curl commands bypasses certificate verification for development.

## Default Admin Access

EntityDB automatically creates a default admin user on first startup:
- **Username**: `admin`
- **Password**: `admin`
- **Roles**: `admin`, `user`

> **Security**: Change the default password immediately in production!

## Your First API Calls

### 1. Login and Get Token
```bash
TOKEN=$(curl -s -k -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Token: $TOKEN"
```

### 2. Create Your First Entity
```bash
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my_first_document",
    "tags": ["type:document", "status:draft", "category:tutorial"],
    "content": "VGhpcyBpcyBteSBmaXJzdCBFbnRpdHlEQiBkb2N1bWVudCE="
  }'
```

### 3. List All Entities
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/list" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Query by Tag
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"
```

### 5. Get Specific Entity
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/get?id=my_first_document" \
  -H "Authorization: Bearer $TOKEN"
```

### 6. Update Entity
```bash
curl -k -X PUT "https://localhost:8085/api/v1/entities/update?id=my_first_document" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "status:published", "category:tutorial"],
    "content": "VXBkYXRlZCBmaXJzdCBFbnRpdHlEQiBkb2N1bWVudCE="
  }'
```

## Web Dashboard

Access the EntityDB dashboard at: **https://localhost:8085**

The dashboard provides:
- Entity browser and search
- Real-time metrics and monitoring  
- System health status
- Administrative tools

## Core Concepts

### Entities
Everything in EntityDB is an entity with:
- **ID**: Unique identifier
- **Tags**: Timestamped metadata
- **Content**: Binary data (auto-chunked for large files)

### Tags
Hierarchical namespace system:
```
type:document          # Entity classification
status:published       # Entity state  
category:tutorial      # Custom metadata
rbac:role:admin       # Access control
```

### Temporal Storage
All tags are timestamped with nanosecond precision, enabling:
- Time-travel queries
- Full audit trails
- Historical data analysis

## Next Steps

### Essential Reading
- [Core Concepts](./03-core-concepts.md) - Understand EntityDB fundamentals
- [Authentication](../30-api-reference/02-authentication.md) - Session management
- [Entity API](../30-api-reference/03-entities.md) - Complete API reference

### Architecture Deep-Dive
- [System Overview](../20-architecture/01-system-overview.md) - High-level architecture
- [Temporal Storage](../20-architecture/02-temporal-architecture.md) - Time-based features
- [RBAC System](../20-architecture/03-rbac.md) - Access control

### Advanced Features
- [Temporal Queries](../40-user-guides/01-temporal-queries.md) - Time-travel operations
- [Entity Relationships](../40-user-guides/03-entity-relationships.md) - Connect entities
- [Performance Tuning](../performance/performance-optimization-results.md) - Optimize for scale

### Administration
- [Production Deployment](../70-deployment/02-production-checklist.md) - Deploy to production
- [Security Configuration](../50-admin-guides/01-security-configuration.md) - Secure your installation
- [Monitoring Setup](../50-admin-guides/03-monitoring.md) - Set up monitoring

## Troubleshooting

**Server won't start?**
- Check logs: `./bin/entitydbd.sh logs`
- Verify ports aren't in use: `netstat -tlnp | grep 8085`

**SSL certificate errors?**
- Use `-k` flag for development: `curl -k https://localhost:8085`
- Configure proper certificates for production

**Authentication issues?**
- Verify token: `curl -k -X GET https://localhost:8085/api/v1/auth/whoami -H "Authorization: Bearer $TOKEN"`
- Check user permissions in dashboard

## Getting Help

- **Documentation**: [EntityDB Docs](../README.md)
- **Issues**: https://git.home.arpa/itdlabs/entitydb/issues
- **Troubleshooting**: [Common Issues](../80-troubleshooting/README.md)

---

**Congratulations!** You now have EntityDB running and understand the basics. Ready to build something amazing? 🚀