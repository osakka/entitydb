# MetHub - Metrics Monitoring Hub

MetHub is a lightweight, real-time metrics monitoring system built on EntityDB's temporal storage capabilities.

## Features

- **Minimal Dependencies**: Shell script agent only requires curl
- **Real-time Monitoring**: Configurable collection intervals
- **Pluggable Widgets**: Easy to add/modify dashboard widgets
- **Custom Metrics**: Send any metric data to EntityDB
- **Time-series Storage**: Leverages EntityDB's temporal features
- **Multi-host Support**: Monitor multiple servers from one dashboard

## Quick Start

### 1. Start the Agent

```bash
# On each server you want to monitor:
export ENTITYDB_URL=https://your-entitydb-server:8085
export ENTITYDB_USER=admin
export ENTITYDB_PASS=admin
export METHUB_INTERVAL=30  # seconds

/opt/entitydb/bin/methub-agent.sh
```

### 2. Access the Dashboard

Open your browser to: `https://your-entitydb-server:8085/methub/`

## Custom Metrics

Create `/etc/methub/custom-metrics.sh` to send custom metrics:

```bash
# Example: Monitor service status
service_status=$(systemctl is-active nginx | grep -q active && echo 1 || echo 0)
send_custom_metric "nginx_status" "$service_status" "status" '"service:nginx"'

# Example: Count active connections
connections=$(ss -tn | grep -c ESTABLISHED)
send_custom_metric "tcp_connections" "$connections" "count"
```

## Widget Types

- **Gauge**: Perfect for percentages (CPU, Memory, Disk)
- **Line Chart**: Time-series data visualization
- **Bar Chart**: Compare values across dimensions
- **Single Value**: Display latest metric with trend
- **Table**: Show multiple metrics in tabular format
- **Heatmap**: Visualize patterns across hosts/time

## Architecture

MetHub uses EntityDB's hub architecture with the `metrics` hub:
- Each metric is stored as an entity with nanosecond timestamp
- Tags enable efficient querying by host, type, and name
- Temporal queries power historical views and aggregations

## Configuration

Agent environment variables:
- `ENTITYDB_URL`: EntityDB server URL
- `ENTITYDB_USER`: Username for authentication
- `ENTITYDB_PASS`: Password for authentication
- `METHUB_INTERVAL`: Collection interval in seconds
- `METHUB_HOSTNAME`: Override hostname (optional)

## Why MetHub?

- **Lightweight**: No heavy dependencies like Prometheus or InfluxDB
- **Flexible**: Send any metric from any source
- **Fast**: Leverages EntityDB's high-performance storage
- **Simple**: Easy to understand and extend
- **Integrated**: Part of the EntityDB ecosystem