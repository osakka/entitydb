# EntityDB Temporal Metrics Collection

## Concept

EntityDB's temporal storage makes it perfect for metrics collection. Instead of creating thousands of individual metric data points, we use ONE entity per metric and leverage temporal tags.

## How It Works

### Traditional Approach (What We DON'T Do)
```
entity_1: cpu_usage @ 10:00:00 = 45%
entity_2: cpu_usage @ 10:00:05 = 47%
entity_3: cpu_usage @ 10:00:10 = 43%
... thousands of entities ...
```

### EntityDB Temporal Approach (What We DO)
```
entity_cpu_usage_server1:
  tags:
    - type:metric
    - metric:name:cpu_usage
    - 2025-05-24T10:00:00Z|metric:value:45.0:percent
    - 2025-05-24T10:00:05Z|metric:value:47.0:percent
    - 2025-05-24T10:00:10Z|metric:value:43.0:percent
    - metric:current:cpu_usage:43.0:percent (snapshot)
```

## API Design

### Collect a Metric
```bash
POST /api/v1/metrics/collect
{
  "metric_name": "cpu_usage",
  "value": 45.5,
  "unit": "percent",
  "instance": "server1",
  "labels": {
    "datacenter": "us-west",
    "env": "production"
  }
}
```

### Query Metric History
```bash
GET /api/v1/metrics/history?metric=cpu_usage&instance=server1&since=2025-05-24T10:00:00Z
```

### Get Current Values
```bash
GET /api/v1/metrics/current
```

## Benefits

1. **Storage Efficiency**: One entity per metric, not per data point
2. **Query Performance**: All history in one entity
3. **Temporal Queries**: Built-in time-range filtering
4. **Atomic Updates**: Each metric update is a simple tag addition
5. **No Data Loss**: Full history preserved automatically

## Implementation Details

### Entity Structure
```json
{
  "id": "metric_cpu_usage_server1",
  "tags": [
    "type:metric",
    "metric:name:cpu_usage",
    "metric:instance:server1",
    "metric:label:datacenter:us-west",
    "metric:label:env:production",
    "TIMESTAMP|metric:value:VALUE:UNIT",  // Repeated for each value
    "metric:current:cpu_usage:45.5:percent" // Latest snapshot
  ],
  "content": {
    "metric": "cpu_usage",
    "instance": "server1",
    "value": 45.5,
    "unit": "percent",
    "labels": {...},
    "updated_at": "2025-05-24T10:00:10Z"
  }
}
```

### Tag Format
- **Value Tags**: `TIMESTAMP|metric:value:VALUE:UNIT`
- **Snapshot Tag**: `metric:current:NAME:VALUE:UNIT`
- **Metadata Tags**: `metric:name:NAME`, `metric:instance:INSTANCE`
- **Label Tags**: `metric:label:KEY:VALUE`

## Use Cases

1. **System Monitoring**: CPU, memory, disk, network metrics
2. **Application Metrics**: Request rates, error counts, latencies
3. **Business Metrics**: Revenue, user counts, conversion rates
4. **IoT Sensors**: Temperature, humidity, pressure readings
5. **Performance Tracking**: Query times, cache hit rates

## Query Examples

### Get Last Hour of CPU Usage
```bash
GET /api/v1/entities/as-of?id=metric_cpu_usage_server1&timestamp=1hour_ago
```

### Find All Metrics for an Instance
```bash
GET /api/v1/entities/query?tags=type:metric,metric:instance:server1
```

### Get Metrics Above Threshold
```bash
# Would need custom query logic to parse value tags
GET /api/v1/entities/query?tags=type:metric,metric:current:cpu_usage:*
```

## Future Enhancements

1. **Aggregation Functions**: Min/max/avg over time ranges
2. **Downsampling**: Automatic rollup of old data
3. **Alerting**: Threshold-based notifications
4. **Grafana Plugin**: Direct visualization support
5. **Metric Math**: Calculations across metrics

## Conclusion

EntityDB's temporal storage transforms metrics collection from a data storage challenge into an elegant solution. By storing time-series data as temporal tags on a single entity, we achieve both efficiency and powerful query capabilities.