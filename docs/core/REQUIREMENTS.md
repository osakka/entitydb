# EntityDB Requirements

This document outlines the system requirements, dependencies, and compatibility requirements for deploying and operating EntityDB.

## System Requirements

### Minimum Hardware Requirements

| Component     | Minimum                 | Recommended              |
|---------------|-------------------------|--------------------------|
| CPU           | 2 cores                 | 4+ cores                 |
| RAM           | 4GB                     | 8GB+                     |
| Disk Space    | 1GB + storage           | 10GB + storage           |
| Disk Type     | SSD                     | NVMe SSD                 |
| Network       | 100 Mbps                | 1 Gbps                   |

### Operating System Compatibility

EntityDB is tested and supported on the following operating systems:

- **Linux**
  - Ubuntu 20.04 LTS or newer
  - CentOS/RHEL 8 or newer
  - Debian 10 or newer
- **macOS**
  - macOS Monterey (12) or newer
- **Windows**
  - Windows 10/11 with WSL2
  - Windows Server 2019 or newer

### Software Dependencies

- **Runtime Dependencies**
  - Go 1.18 or newer (for building from source)
  - OpenSSL 1.1.1 or newer (for SSL connections)
  - POSIX-compliant shell (for scripts)

- **Optional Dependencies**
  - jq (for API script examples)
  - curl (for API examples)
  - nginx (for production proxy)

## Performance Characteristics

### Scaling Guidelines

EntityDB scaling characteristics based on dataset size:

| Entity Count  | Disk Usage    | RAM Usage*    | Query Time    |
|---------------|---------------|---------------|---------------|
| 100K          | ~50MB         | ~100MB        | <0.1ms        |
| 1M            | ~500MB        | ~250MB        | ~0.5ms        |
| 5M            | ~2.5GB        | ~500MB        | ~1.5ms        |
| 10M           | ~5GB          | ~1GB          | ~3ms          |
| 50M           | ~25GB         | ~3GB          | ~10ms         |

*RAM usage is significantly affected by OS disk caching and active query patterns

### Performance Factors

- **Temporal Indexing:** Enables 100x performance improvement for historical queries
- **Memory-Mapped Files:** Zero-copy reads with OS-level caching
- **B-tree Indexes:** O(log n) lookup performance regardless of dataset size
- **Bloom Filters:** Fast negative lookups, reducing unnecessary disk reads
- **Skip-List Indexes:** Efficient range queries on temporal data

## Storage Requirements

### Disk Space Calculation

Storage requirements can be estimated as follows:

- **Base Size:** ~5 bytes per entity (fixed overhead)
- **Tag Size:** ~20 bytes per tag (including timestamp)
- **Content Size:** Variable based on entity content
- **Index Size:** ~10% of total entity size for indexes

Example calculation for 1M entities with 10 tags each and 1KB of content on average:
```
1,000,000 × (5 + 20 × 10 + 1,024) × 1.1 = ~1.16GB
```

### Autochunking

Large files are automatically chunked with the following characteristics:
- **Default Chunk Size:** 4MB
- **Chunking Threshold:** Files larger than 4MB are automatically chunked
- **Storage Overhead:** ~100 bytes per chunk for metadata
- **Parent-Child Relationship:** Transparent retrieval via parent entity

## Network Requirements

### Ports

The following network ports are used by default:

- **HTTP:** 8085 (default, configurable)
- **HTTPS:** 8443 (when SSL is enabled, configurable)

### Firewall Considerations

For basic deployment, ensure the following:

- Allow inbound connections to HTTP/HTTPS ports
- Allow outbound connections for updates (if applicable)
- Consider restricting access to administration endpoints

### SSL Configuration

- **Default Mode:** SSL-enabled (since v2.10.0)
- **Certificate Paths:**
  - Certificate: `/opt/entitydb/var/ssl/cert.pem`
  - Key: `/opt/entitydb/var/ssl/key.pem`
- **Self-signed Certificates:** Generated automatically if not provided

## Database Compatibility

EntityDB uses a custom binary format (EBF) with the following characteristics:

- **Format Version:** v2.12.0 (current)
- **Backward Compatibility:** Compatible with v2.8.0+
- **Upgrade Path:** In-place upgrades supported from v2.8.0+

*Note: Data migration required for versions prior to v2.8.0*

## Production Deployment Recommendations

For production environments, we recommend:

1. **Dedicated Server:** Isolated hardware for predictable performance
2. **Backup Strategy:** Regular database backups (binary data + WAL)
3. **Monitoring:** System metrics collection for CPU, RAM, and disk usage
4. **Reverse Proxy:** Use nginx/HAProxy for SSL termination and load balancing
5. **Security:** Proper RBAC implementation with strong passwords
6. **High Availability:** Consider multiple instances with shared storage for critical deployments

## Development Environment

For development purposes, you can run with reduced requirements:

- **Minimum RAM:** 2GB
- **Disk Space:** 500MB
- **Dependencies:** Go 1.18+ for compilation

## Known Limitations

- **Maximum Entity Size:** No hard limit (but practical limit ~1GB per entity due to memory constraints during operations)
- **Maximum Tags per Entity:** No hard limit (but performance degrades over ~1000 tags)
- **Concurrent Connections:** Limited by OS file descriptor limits
- **Rate Limiting:** Not implemented in current version
- **Distributed Deployment:** Not supported in current version