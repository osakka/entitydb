# EntityDB Example Applications

This directory contains comprehensive examples showcasing EntityDB's temporal database capabilities in real-world scenarios.

## 🚀 Monitoring System Demo

**Location**: `monitoring_system.py`

A complete monitoring system demonstrating EntityDB's temporal awesomeness with:

### Features Demonstrated

🕰️ **Temporal Database Capabilities**
- Nanosecond-precision timestamp storage
- Historical trend analysis using temporal queries
- Point-in-time recovery with `as-of` queries
- Time-series data analysis and aggregation

📊 **Real-World Monitoring**
- Server metrics collection (CPU, memory, disk, network)
- Service health monitoring (response times, error rates, availability)
- Intelligent alerting based on historical patterns
- Dashboard with real-time and historical views

🧠 **Smart Analytics**
- Baseline establishment from historical data
- Anomaly detection using statistical analysis
- Trend direction analysis (increasing/decreasing/stable)
- Context-aware alerting (not just threshold-based)

### Architecture Highlights

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Servers       │    │   Services      │    │   Metrics       │
│   (entities)    │    │   (entities)    │    │   (temporal     │
│                 │    │                 │    │    tags)        │
│ • web-01        │    │ • web-frontend  │    │ • cpu_usage     │
│ • api-01        │    │ • user-api      │    │ • response_time │
│ • db-01         │    │ • payment-api   │    │ • error_rate    │
│ • cache-01      │    │ • database      │    │ • availability  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   EntityDB      │
                    │   Temporal      │
                    │   Storage       │
                    │                 │
                    │ • Nanosecond    │
                    │   precision     │
                    │ • Point-in-time │
                    │   queries       │
                    │ • Historical    │
                    │   analysis      │
                    └─────────────────┘
```

### Usage

```bash
# Install dependencies
pip3 install requests

# Run the monitoring system
cd /opt/entitydb/examples
python3 monitoring_system.py
```

### Sample Output

```
🎯 EntityDB Temporal Monitoring System Demo
==================================================
This demonstration showcases EntityDB's temporal database capabilities
with a real-world monitoring system that tracks servers and services.

Features demonstrated:
• Nanosecond-precision temporal data storage
• Historical trend analysis using temporal queries  
• Point-in-time queries (as-of functionality)
• Intelligent alerting based on historical patterns
• Real-time metric collection and analysis

🏗️  Setting up monitoring infrastructure...
  ✓ Created server: web-01
  ✓ Created server: web-02
  ✓ Created server: api-01
  ✓ Created server: api-02
  ✓ Created server: db-01
  ✓ Created server: cache-01
  ✓ Created service: web-frontend
  ✓ Created service: user-api
  ✓ Created service: product-api
  ✓ Created service: payment-api
  ✓ Created service: database
  ✓ Created service: redis-cache
🎯 Infrastructure ready: 6 servers, 6 services

📊 MONITORING DASHBOARD - 14:30:15
------------------------------------------------------------
🚨 Checking for alerts...
  🚨 WARNING: CPU usage 82.5% (trend: increasing)
  🚨 CRITICAL: Error rate 5.2% (trend: increasing)
  ✅ All other systems nominal

🎯 SYSTEM OVERVIEW:
  Servers: 5/6 healthy
  Services: 5/6 healthy
  Active Alerts: 2

🖥️  SERVER TRENDS (sample):
  web-01: CPU 45.2%, Memory 52.1% [healthy]
  web-02: CPU 82.5%, Memory 48.9% [warning]
  api-01: CPU 38.7%, Memory 61.3% [healthy]

⏱️  Demonstrating EntityDB temporal capabilities:
  - Collecting metrics every 10 seconds with nanosecond precision
  - Historical trend analysis using temporal queries
  - Intelligent alerting based on historical patterns
  - Point-in-time recovery and as-of queries available
```

## 🔍 What Makes This Special

### EntityDB Temporal Advantages

1. **Nanosecond Precision**: Unlike traditional monitoring systems that store timestamps as integers or lose precision, EntityDB maintains nanosecond accuracy for all temporal data.

2. **Native Temporal Queries**: Built-in support for:
   - `history`: Full change history of any entity
   - `as-of`: Point-in-time snapshots
   - `diff`: Compare states between timestamps
   - `changes`: Track specific changes over time

3. **No Data Loss**: Every metric value is preserved with its exact timestamp, enabling perfect historical reconstruction.

4. **Intelligent Analysis**: Historical pattern recognition enables smart alerting that considers trends, not just current values.

### Real-World Benefits

- **Root Cause Analysis**: Trace back through exact system states when issues occurred
- **Capacity Planning**: Analyze historical growth patterns with nanosecond precision
- **Performance Baselining**: Establish normal operating ranges from historical data
- **Incident Response**: Quickly identify when problems started and what changed
- **Compliance**: Perfect audit trails with immutable historical records

## 🚀 Try It Yourself

The monitoring system is fully functional and will:

1. Create realistic server and service entities
2. Generate realistic metrics with daily patterns and anomalies
3. Store all data with nanosecond timestamps in EntityDB
4. Perform real temporal queries for trend analysis
5. Generate intelligent alerts based on historical patterns
6. Demonstrate the power of temporal database capabilities

**Ready to see EntityDB's temporal awesomeness in action? Run the demo!**