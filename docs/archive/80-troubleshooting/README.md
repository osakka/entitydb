# EntityDB Troubleshooting

> **Category**: Problem Resolution | **Target Audience**: All Users | **Technical Level**: Varies

This section provides solutions to common EntityDB issues, error diagnosis guides, and problem resolution strategies. Find quick solutions to common problems and detailed debugging guidance.

## 📋 Contents

### [Content Format Issues](./01-content-format.md)
**Data format and content problems**
- Binary format (EBF) corruption issues
- Content encoding and decoding errors
- Data integrity validation failures
- File format compatibility problems

### [Content Wrapping Problems](./02-content-wrapping.md)
**Content processing and display issues**
- API response formatting problems
- JSON parsing and serialization errors
- Content size and chunking issues
- Character encoding problems

### [Tag Index Persistence](./03-tag-index-persistence.md)
**Tag indexing and search problems**
- Tag index corruption and rebuilding
- Search performance degradation
- Index synchronization issues
- Memory mapping problems

### [Temporal Tags Issues](./04-temporal-tags.md)
**Time-series and temporal data problems**
- Timestamp parsing and format errors
- Temporal query performance issues
- History retrieval problems
- Time-travel query failures

## 🚨 Quick Problem Resolution

### Common Startup Issues
```bash
# Server won't start
./bin/entitydbd.sh status              # Check if already running
sudo netstat -tlnp | grep 8085         # Check port availability
tail -f var/log/entitydb.log           # Check error logs

# Permission denied errors
sudo chown -R $USER:$USER /opt/entitydb/var/
chmod +x bin/entitydb bin/entitydbd.sh
```

### Authentication Problems
```bash
# Default admin login not working
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Check user creation
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/entities/list?tags=type:user
```

### Performance Issues
```bash
# Check system metrics
curl http://localhost:8085/api/v1/system/metrics
curl http://localhost:8085/health

# Monitor resource usage
top -p $(pgrep entitydb)
df -h /opt/entitydb/var/
```

## 🔍 Diagnostic Steps

### 1. Check System Health
1. **Server Status**: Verify EntityDB is running
2. **Port Availability**: Confirm ports 8085/8443 are accessible
3. **Log Analysis**: Review application logs for errors
4. **Resource Usage**: Check CPU, memory, and disk usage

### 2. Validate Configuration
1. **Environment Variables**: Verify configuration settings
2. **File Permissions**: Ensure proper read/write access
3. **SSL Configuration**: Check certificate validity
4. **Database Files**: Verify EBF file integrity

### 3. Test API Connectivity
1. **Health Check**: `GET /health` should return 200
2. **Authentication**: Test login with admin credentials
3. **Basic Operations**: Create and retrieve a test entity
4. **Metrics Access**: Verify monitoring endpoints work

## 🛠️ Common Error Patterns

### Authentication Errors
- **401 Unauthorized**: Invalid or expired token
- **403 Forbidden**: Insufficient RBAC permissions
- **Invalid credentials**: Check username/password format

### Data Access Errors
- **404 Not Found**: Entity ID doesn't exist
- **400 Bad Request**: Invalid query parameters or data format
- **500 Internal Server Error**: Database corruption or system issues

### Performance Degradation
- **Slow queries**: Tag index needs rebuilding
- **High memory usage**: Large entity content without chunking
- **Disk space issues**: WAL files growing without checkpointing

## 🔗 Quick Navigation

- **Installation Issues**: [Getting Started](../10-getting-started/) - Setup and installation guidance
- **API Problems**: [API Reference](../30-api-reference/) - Endpoint documentation
- **Configuration Issues**: [Configuration Reference](../90-reference/01-configuration-reference.md) - All config options
- **Performance Optimization**: [Architecture](../20-architecture/) - System design insights

## 📞 Getting Help

### Community Support
1. **Check existing issues** in the repository
2. **Search documentation** for similar problems
3. **Review error logs** and include relevant details
4. **Provide reproduction steps** when reporting issues

### Information to Include
- EntityDB version (`curl http://localhost:8085/health`)
- Operating system and version
- Error messages and log excerpts
- Configuration details (sanitized)
- Steps to reproduce the issue

---

*This troubleshooting guide provides systematic problem resolution for EntityDB. Start with the quick diagnostic steps and escalate to detailed guides as needed.*