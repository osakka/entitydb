# Performance Comparison Quick Reference

## At-a-Glance Performance Comparison

### Basic Operations (milliseconds)

| Operation | EntityDB | MySQL | InfluxDB | Redis |
|-----------|----------|--------|----------|-------|
| Simple Write | 4.78 | 182 | 8.2 | 0.15 |
| Simple Read | 0.92 | 89 | 15 | 0.8 |
| Complex Query | 12.1 | 524 | 87 | 15 |
| Temporal Query | 1.4 | 95 | 2.8 | 12* |

*Custom implementation required

### Large-Scale Performance (100k entities)

| Metric | EntityDB | MySQL | InfluxDB | Redis |
|--------|----------|--------|----------|-------|
| Bulk Load Time | 8.5 min | 5.2 hrs | 13.7 min | 45 sec |
| Query Response | 42ms | 4.2s | 678ms | 89ms |
| Memory Usage | 512MB | 2.8GB | 1.4GB | 890MB |
| Disk Usage | 378MB | 1.6GB | 892MB | 567MB |

### Best Use Cases

**EntityDB**
- Temporal data with complex relationships
- Event sourcing and audit trails
- Historical analytics
- Time-travel queries

**MySQL**
- Traditional OLTP workloads
- Strong ACID requirements
- Complex JOINs and transactions
- Well-known SQL queries

**InfluxDB**
- Time-series metrics
- IoT sensor data
- Monitoring and alerting
- Simple time-based aggregations

**Redis**
- High-speed caching
- Session storage
- Real-time leaderboards
- Pub/sub messaging